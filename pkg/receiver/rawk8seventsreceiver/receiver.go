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
	"strconv"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/extension/experimental/storage"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
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
	"normal":  plog.SeverityNumberInfo,
	"warning": plog.SeverityNumberWarn,
}

const latestResourceVersionStorageKey string = "latestResourceVersion"

type rawK8sEventsReceiver struct {
	cfg                   *Config
	client                k8s.Interface
	eventControllers      []cache.Controller
	eventCh               chan *eventChange
	ctx                   context.Context
	cancel                context.CancelFunc
	startTime             time.Time
	storage               storage.Client
	latestResourceVersion uint64

	consumer consumer.Logs
	logger   *zap.Logger
	id       component.ID
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
	params receiver.Settings,
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

		informerOptions := cache.InformerOptions{
			ListerWatcher: namespaceListWatch,
			ObjectType:    &corev1.Event{},
			Handler: cache.ResourceEventHandlerFuncs{
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
			},
			ResyncPeriod: 0, // Same as before, no resync period
		}

		_, namespaceController = cache.NewInformerWithOptions(informerOptions)
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
		id:               params.ID,
	}
	return receiver, nil
}

// Start tells the receiver to start.
func (r *rawK8sEventsReceiver) Start(ctx context.Context, host component.Host) error {
	var err error
	r.storage, err = r.getStorage(ctx, host)
	if err != nil {
		return fmt.Errorf("error when getting storage: %s", err)
	}

	r.latestResourceVersion, err = r.getLatestResourceVersion(ctx)
	if err != nil {
		return fmt.Errorf("error when getting latest resource version: %s", err)
	}

	r.ctx, r.cancel = context.WithCancel(ctx)

	go r.processEventChangeLoop()

	for _, eventController := range r.eventControllers {
		go eventController.Run(r.ctx.Done())
	}

	return nil
}

// Shutdown is invoked during service shutdown.
func (r *rawK8sEventsReceiver) Shutdown(ctx context.Context) error {
	r.cancel()
	var err error
	if r.storage != nil {
		err = r.storage.Close(ctx)
	}
	return err
}

func (r *rawK8sEventsReceiver) getStorage(ctx context.Context, host component.Host) (storage.Client, error) {
	if host == nil {
		r.logger.Debug("Storage not initialized: host is not available")
		return nil, nil
	}

	var storageExtension storage.Extension
	var storageExtensionId component.ID
	for extentionId, extension := range host.GetExtensions() {
		if se, ok := extension.(storage.Extension); ok {
			if storageExtension != nil {
				return nil, fmt.Errorf("multiple storage extensions found: '%s', '%s'", storageExtensionId, extentionId)
			}
			storageExtension = se
			storageExtensionId = extentionId
		}
	}

	if storageExtension == nil {
		r.logger.Debug("Storage not initialized: no storage extension found")
		return nil, nil
	}

	storageClient, err := storageExtension.GetClient(ctx, component.KindReceiver, r.id, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get storage client for extension '%s': %s", storageExtensionId, err)
	}

	r.logger.Info("Initialized storage", zap.Any("storage_extension_id", storageExtensionId))
	return storageClient, nil
}

func (r *rawK8sEventsReceiver) getLatestResourceVersion(ctx context.Context) (uint64, error) {
	if r.storage == nil {
		r.logger.Info("Did not find latest resource version, as there is no storage.")
		return 0, nil
	}

	latestResourceVersionBytes, err := r.storage.Get(ctx, latestResourceVersionStorageKey)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve latest resource version from storage: %s", err)
	}

	if latestResourceVersionBytes == nil {
		r.logger.Info("Latest resource version not found in storage")
		return 0, nil
	}

	latestResourceVersionString := string(latestResourceVersionBytes)
	latestResourceVersion, err := strconv.ParseUint(latestResourceVersionString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse latest resource version '%s' to number: %s", latestResourceVersionString, err)
	}

	r.logger.Info("Found latest resource version in storage", zap.Any("latest_resource_version", latestResourceVersion))
	return latestResourceVersion, nil
}

// Consume logs and retry on recoverable errors
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
			r.logger.Warn("ConsumeLogs() recoverable error, will retry",
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
	r.recordEventReceived(eventChange.event)
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
		r.logger.Error("ConsumeLogs() error",
			zap.String("error", err.Error()),
		)
	}
}

func (r *rawK8sEventsReceiver) recordEventReceived(event *corev1.Event) {
	if r.storage == nil {
		return
	}

	err := r.storage.Set(r.ctx, latestResourceVersionStorageKey, []byte(event.ResourceVersion))
	if err != nil {
		r.logger.Warn("failed to record event received", zap.Error(err), zap.String("incoming_resource_version", event.ResourceVersion))
	}
}

// Check if we should process the event.
// If a latest resource version was retrieved from storage, compare that to the incoming event's resource version.
// Otherwise, check event time and compare it to collector's start time.
func (r *rawK8sEventsReceiver) isEventAccepted(event *corev1.Event) bool {
	if r.latestResourceVersion > 0 {
		incomingEventResourceVersion, err := strconv.ParseUint(event.ResourceVersion, 10, 64)
		if err != nil {
			r.logger.Debug("Failed checking if event is accepted, cannot convert incoming resource version to a number. Accepting the incoming event.",
				zap.Error(err),
				zap.Any("incoming_event_version", event.ResourceVersion),
				zap.Any("latest_resource_version", r.latestResourceVersion),
			)
			return true
		}

		incomingEventIsNewer := incomingEventResourceVersion > r.latestResourceVersion
		if incomingEventIsNewer {
			r.logger.Debug("Incoming event is accepted as it is newer.",
				zap.Any("incoming_event_version", incomingEventResourceVersion),
				zap.Any("latest_resource_version", r.latestResourceVersion),
			)
			return true
		} else {
			r.logger.Debug("Incoming event is NOT accepted, as it is older.",
				zap.Any("incoming_event_version", incomingEventResourceVersion),
				zap.Any("latest_resource_version", r.latestResourceVersion),
			)
			return false
		}
	}

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
	pdataObjectMap := pcommon.NewMap()
	err = pdataObjectMap.FromRaw(map[string]interface{}{"object": eventMap})
	if err != nil {
		return ld, err
	}

	lr.SetTimestamp(pcommon.NewTimestampFromTime(getEventTimestamp(event)))

	// The Message field contains description about the event,
	// which is best suited for the "Body" of the LogRecordSlice.
	lr.Body().SetStr(event.Message)

	// Set the "SeverityNumber" and "SeverityText" if a known type of severity is found.
	if severityNumber, ok := severityMap[strings.ToLower(event.Type)]; ok {
		lr.SetSeverityNumber(severityNumber)
		lr.SetSeverityText(event.Type)
	} else {
		r.logger.Debug("unknown severity type", zap.String("type", event.Type))
	}

	pdataObjectMap.CopyTo(lr.Attributes())

	// for compatibility with the FluentD plugin's data format, we need to put the change type under "type"
	if _, ok := lr.Attributes().Get("type"); !ok {
		lr.Attributes().PutStr("type", string(eventChange.changeType))
	}
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
