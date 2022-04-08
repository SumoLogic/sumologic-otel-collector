// Copyright 2022, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rawk8seventsreceiver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	k8s_scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
)

// Only two types of events are created as of now.
// For more info: https://docs.openshift.com/container-platform/4.9/rest_api/metadata_apis/event-core-v1.html
var severityMap = map[string]plog.SeverityNumber{
	"normal":  plog.SeverityNumberINFO,
	"warning": plog.SeverityNumberWARN,
}

type rawK8sEventsReceiver struct {
	cfg              *Config
	client           k8s.Interface
	eventControllers []cache.Controller
	eventCh          chan *eventChange
	ctx              context.Context
	cancel           context.CancelFunc
	startTime        time.Time

	consumer consumer.Logs
	logger   *zap.Logger
}

// Function type for creating ListerWatcher objects. Used for injecting mocks into k8s informers.
type ListerWatcherFactory func(c cache.Getter, resource string, namespace string, fieldSelector fields.Selector) cache.ListerWatcher

// We care about event creation and updates. The eventChange struct carries information about these changes.
type eventChangeType string // can be ADDED or MODIFIED
const (
	eventChangeTypeAdded    = "ADDED"
	eventChangeTypeModified = "MODIFIED"
)

type eventChange struct {
	event      *corev1.Event
	changeType eventChangeType
}

// create a new receiver
func newRawK8sEventsReceiver(
	params component.ReceiverCreateSettings,
	cfg *Config,
	consumer consumer.Logs,
	client k8s.Interface,
	listerWatcherFactory ListerWatcherFactory,
) (*rawK8sEventsReceiver, error) {
	var namespaceController cache.Controller
	var namespaces []string

	// if no namespaces are specified, watch all of them
	if len(cfg.Namespaces) == 0 {
		namespaces = []string{corev1.NamespaceAll}
	} else {
		namespaces = cfg.Namespaces
	}

	eventCh := make(chan *eventChange)
	eventControllers := []cache.Controller{}

	restClient := client.CoreV1().RESTClient()
	for _, namespace := range namespaces {
		namespaceListWatch := listerWatcherFactory(restClient, "events", namespace, fields.Everything())
		_, namespaceController = cache.NewInformer(namespaceListWatch, &corev1.Event{}, 0, cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*corev1.Event)
				eventCh <- &eventChange{
					changeType: eventChangeTypeAdded,
					event:      event,
				}
			},
			UpdateFunc: func(_, obj interface{}) {
				event := obj.(*corev1.Event)
				eventCh <- &eventChange{
					changeType: eventChangeTypeModified,
					event:      event,
				}
			},
		})
		eventControllers = append(eventControllers, namespaceController)
	}
	receiver := &rawK8sEventsReceiver{
		cfg:              cfg,
		client:           client,
		eventControllers: eventControllers,
		eventCh:          eventCh,
		consumer:         consumer,
		logger:           params.Logger,
		startTime:        time.Now(),
	}
	return receiver, nil
}

// Start tells the receiver to start.
func (r *rawK8sEventsReceiver) Start(ctx context.Context, host component.Host) error {
	r.ctx, r.cancel = context.WithCancel(ctx)

	go r.processEventChangeLoop()

	for _, eventController := range r.eventControllers {
		go eventController.Run(r.ctx.Done())
	}

	return nil
}

// Shutdown is invoked during service shutdown.
func (r *rawK8sEventsReceiver) Shutdown(context.Context) error {
	r.cancel()
	return nil
}

// Consume metrics and retry on recoverable errors
func (r *rawK8sEventsReceiver) consumeWithRetry(ctx context.Context, logs plog.Logs) error {
	constantBackoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(r.cfg.ConsumeRetryDelay), r.cfg.ConsumeMaxRetries)

	// retry handling according to https://github.com/open-telemetry/opentelemetry-collector/blob/ac9eb92edc4a2e16cf721ffe40c2cdfc2fb76ab9/component/receiver.go#L45
	err := backoff.RetryNotify(
		func() error {
			// we need to check for context cancellation here
			select {
			case <-r.ctx.Done():
				return backoff.Permanent(errors.New("closing"))
			default:
			}
			err := r.consumer.ConsumeLogs(ctx, logs)
			if consumererror.IsPermanent(err) {
				return backoff.Permanent(err)
			} else {
				return err
			}
		},
		constantBackoff,
		func(err error, delay time.Duration) {
			r.logger.Warn("ConsumeMetrics() recoverable error, will retry",
				zap.Error(err), zap.Duration("delay", delay),
			)
		},
	)

	return err
}

// Get event changes from a channel and process them
// we have a separate loop for this to serialize the changes and avoid doing
// expensive processing in informer handler functions
func (r *rawK8sEventsReceiver) processEventChangeLoop() {
	for eventChange := range r.eventCh {
		r.processEventChange(context.Background(), eventChange)
	}
}

// Process a single event change
// this includes: checking if we should process the event, converting it into a plog.Logs
// and sending it to the next consumer in the pipeline
func (r *rawK8sEventsReceiver) processEventChange(ctx context.Context, eventChange *eventChange) {
	if !r.isEventAccepted(eventChange.event) {
		r.logger.Debug("skipping event, too old", zap.Any("event", eventChange.event))
		return
	}
	r.logger.Debug("processing event", zap.Any("event", eventChange.event), zap.String("type", string(eventChange.changeType)))

	logs, err := r.convertToLog(eventChange)
	if err != nil {
		r.logger.Error("failed to convert event", zap.Error(err), zap.Any("event", eventChange.event))
		return
	}
	err = r.consumeWithRetry(ctx, logs)
	if err != nil {
		r.logger.Error("ConsumeMetrics() error",
			zap.String("error", err.Error()),
		)
	}
}

// Check if we should process the event
// Currently this only checks if the event isn't too old
func (r *rawK8sEventsReceiver) isEventAccepted(event *corev1.Event) bool {
	eventTime := getEventTimestamp(event)
	minAcceptableTime := r.startTime.Add(-r.cfg.MaxEventAge)
	return eventTime.After(minAcceptableTime) || eventTime.Equal(minAcceptableTime)
}

// Convert an eventChange record to an opentelemetry Logs record in a format compatible
// with Sumo Logic's FluentD plugin
func (r *rawK8sEventsReceiver) convertToLog(eventChange *eventChange) (plog.Logs, error) {
	event := eventChange.event
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	lr := sl.LogRecords().AppendEmpty()

	// Convert the event into a map[string][interface{}]
	// informers return objects without Kind information, add it
	// see: https://github.com/kubernetes/client-go/issues/308
	gvks, _, err := k8s_scheme.Scheme.ObjectKinds(event)
	if err != nil {
		return ld, fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}

	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		event.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	// Convert the event into a map[string]interface{}
	eventMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(event)
	if err != nil {
		return ld, err
	}

	// for compatibility with the FluentD plugin's data format, we need to put the event data under the "object" key
	pdataObjectMap := pcommon.NewMapFromRaw(map[string]interface{}{"object": eventMap})

	lr.SetTimestamp(pcommon.NewTimestampFromTime(getEventTimestamp(event)))

	// The Message field contains description about the event,
	// which is best suited for the "Body" of the LogRecordSlice.
	lr.Body().SetStringVal(event.Message)

	// Set the "SeverityNumber" and "SeverityText" if a known type of severity is found.
	if severityNumber, ok := severityMap[strings.ToLower(event.Type)]; ok {
		lr.SetSeverityNumber(severityNumber)
		lr.SetSeverityText(event.Type)
	} else {
		r.logger.Debug("unknown severity type", zap.String("type", event.Type))
	}

	pdataObjectMap.CopyTo(lr.Attributes())

	// for compatibility with the FluentD plugin's data format, we need to put the change type under "type"
	lr.Attributes().InsertString("type", string(eventChange.changeType))
	return ld, nil
}

// Return the EventTimestamp based on the populated k8s event timestamps.
// Priority: EventTime > LastTimestamp > FirstTimestamp.
func getEventTimestamp(ev *corev1.Event) time.Time {
	var eventTimestamp time.Time

	switch {
	case ev.EventTime.Time != time.Time{}:
		eventTimestamp = ev.EventTime.Time
	case ev.LastTimestamp.Time != time.Time{}:
		eventTimestamp = ev.LastTimestamp.Time
	case ev.FirstTimestamp.Time != time.Time{}:
		eventTimestamp = ev.FirstTimestamp.Time
	}

	return eventTimestamp
}
