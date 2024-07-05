// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/storagetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	cachetest "k8s.io/client-go/tools/cache/testing"
)

type countingErrorConsumer struct {
	err       error
	CallCount int
}

func (er *countingErrorConsumer) ConsumeLogs(context.Context, plog.Logs) error {
	er.CallCount++
	return er.err
}

func (er *countingErrorConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

// newCountingErrorConsumer returns a Consumer that drops all received data and returns the specified error to Consume* callers
// It also counts the number of time Consume* was called
func newCountingErrorConsumer(err error) *countingErrorConsumer {
	return &countingErrorConsumer{err: err, CallCount: 0}
}

func TestNewRawK8sEventsReceiver(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset(&corev1.Event{})
	r, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		rCfg,
		consumertest.NewNop(),
		client,
		fakeListWatchFactory,
	)

	require.NoError(t, err)
	require.NotNil(t, r)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	assert.NoError(t, r.Shutdown(context.Background()))

	rCfg.Namespaces = []string{"test", "another_test"}
	r1, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		rCfg,
		consumertest.NewNop(),
		client,
		fakeListWatchFactory,
	)

	require.NoError(t, err)
	require.NotNil(t, r1)
	require.NoError(t, r1.Start(context.Background(), componenttest.NewNopHost()))
	assert.NoError(t, r1.Shutdown(context.Background()))
}

func TestProcessEventE2E(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	sink := new(consumertest.LogsSink)
	listWatch := cachetest.NewFakeControllerSource()
	listWatchFactory := func(
		c cache.Getter,
		resource string,
		namespace string,
		fieldSelector fields.Selector,
	) cache.ListerWatcher {
		return listWatch
	}

	r, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		rCfg,
		sink,
		client,
		listWatchFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, r)

	ctx := context.Background()
	err = r.Start(ctx, componenttest.NewNopHost())
	assert.NoError(t, err)
	listWatch.Add(getEvent())
	assert.Eventually(t, func() bool {
		return sink.LogRecordCount() == 1
	}, time.Second, time.Millisecond, "expected one event, got 0")

	err = r.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestProcessEvent(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	sink := new(consumertest.LogsSink)
	r, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		rCfg,
		sink,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	r.ctx = context.Background()
	eventChange := eventChange{getEvent(), eventChangeTypeAdded}
	r.processEventChange(context.Background(), &eventChange)

	assert.Equal(t, 1, sink.LogRecordCount())
}

func TestConsumeRetryOnRecoverableError(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	core, observedLogs := observer.New(zapcore.InfoLevel)
	consumerError := errors.New("recoverable error")
	consumer := newCountingErrorConsumer(consumerError)
	settings := receivertest.NewNopSettings()
	settings.Logger = zap.New(core)
	rCfg.ConsumeMaxRetries = 1
	rCfg.ConsumeRetryDelay = time.Nanosecond

	receiver, err := newRawK8sEventsReceiver(
		settings,
		rCfg,
		consumer,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	receiver.ctx = ctx

	logs := plog.Logs{}
	err = receiver.consumeWithRetry(ctx, logs)
	assert.Error(t, err)
	assert.Equal(t, int(rCfg.ConsumeMaxRetries+1), consumer.CallCount)
	assert.Equal(t, int(rCfg.ConsumeMaxRetries), observedLogs.Len())
	assert.Equal(t, int(rCfg.ConsumeMaxRetries), observedLogs.FilterMessage("ConsumeLogs() recoverable error, will retry").Len())
}

func TestConsumeNoRetryOnPermanentError(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	core, observedLogs := observer.New(zapcore.InfoLevel)
	consumerError := consumererror.NewPermanent(errors.New("permanent error"))
	consumer := newCountingErrorConsumer(consumerError)
	settings := receivertest.NewNopSettings()
	settings.Logger = zap.New(core)
	rCfg.ConsumeRetryDelay = time.Nanosecond

	receiver, err := newRawK8sEventsReceiver(
		settings,
		rCfg,
		consumer,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	receiver.ctx = context.Background()

	logs := plog.Logs{}
	err = receiver.consumeWithRetry(ctx, logs)
	assert.Error(t, err)
	assert.Equal(t, 1, consumer.CallCount)
	assert.Equal(t, 0, observedLogs.Len())
}

func TestConvertEventToLog(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	sink := new(consumertest.LogsSink)
	r, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		rCfg,
		sink,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	r.ctx = context.Background()
	k8sEvent := getEvent()
	eventChange := &eventChange{k8sEvent, eventChangeTypeAdded}
	logs, err := r.convertToLog(eventChange)
	assert.NoError(t, err)
	assert.Equal(t, 1, logs.LogRecordCount())

	// check the standard log record fields: body, severity and timestamp
	logRecord := logs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	assert.Equal(t, eventChange.event.Message, logRecord.Body().AsString())
	assert.Equal(t, plog.SeverityNumberInfo, logRecord.SeverityNumber())
	assert.Equal(t, eventChange.event.FirstTimestamp.Time.UTC(), logRecord.Timestamp().AsTime())

	// check the top-level attributes: `object` and `type`
	logAttributes := logRecord.Attributes()
	typeAttributeValue, typeExists := logAttributes.Get("type")
	objectAttributeValue, objectExists := logAttributes.Get("object")
	assert.True(t, typeExists)
	assert.Equal(t, string(eventChange.changeType), typeAttributeValue.AsString())
	assert.True(t, objectExists)

	// check the event fields inside `object`
	expectedObjectKeys := []string{
		"count",
		"eventTime",
		"firstTimestamp",
		"involvedObject",
		"lastTimestamp",
		"message",
		"metadata",
		"reason",
		"reportingComponent",
		"reportingInstance",
		"source",
		"type",
	}
	objectMap := objectAttributeValue.Map()
	for _, objectKey := range expectedObjectKeys {
		_, keyExists := objectMap.Get(objectKey)
		assert.True(t, keyExists, objectKey)
	}

}

func TestEventFilterByTime(t *testing.T) {
	maxEventAge := time.Minute * 5
	rCfg := createDefaultConfig().(*Config)
	rCfg.MaxEventAge = maxEventAge
	client := fake.NewSimpleClientset()
	sink := new(consumertest.LogsSink)
	r, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		rCfg,
		sink,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	k8sEvent := getEvent()

	k8sEvent.FirstTimestamp = v1.NewTime(r.startTime.Add(-maxEventAge))
	assert.True(t, r.isEventAccepted(k8sEvent))

	k8sEvent.FirstTimestamp = v1.NewTime(r.startTime.Add(-maxEventAge).Add(-time.Nanosecond))
	assert.False(t, r.isEventAccepted(k8sEvent))
}

func TestGetEventTimestamp(t *testing.T) {
	k8sEvent := getEvent()
	eventTimestamp := getEventTimestamp(k8sEvent)
	assert.Equal(t, k8sEvent.FirstTimestamp.Time, eventTimestamp)

	k8sEvent.FirstTimestamp = v1.Time{Time: time.Now().Add(-time.Hour)}
	k8sEvent.LastTimestamp = v1.Now()
	eventTimestamp = getEventTimestamp(k8sEvent)
	assert.Equal(t, k8sEvent.LastTimestamp.Time, eventTimestamp)

	k8sEvent.FirstTimestamp = v1.Time{}
	k8sEvent.LastTimestamp = v1.Time{}
	k8sEvent.EventTime = v1.MicroTime(v1.Now())
	eventTimestamp = getEventTimestamp(k8sEvent)
	assert.Equal(t, k8sEvent.EventTime.Time, eventTimestamp)
}

func TestNoStorage(t *testing.T) {
	receiverConfig := createDefaultConfig().(*Config)
	logsSink := new(consumertest.LogsSink)
	listWatch := cachetest.NewFakeControllerSource()
	listWatchFactory := func(
		c cache.Getter,
		resource string,
		namespace string,
		fieldSelector fields.Selector,
	) cache.ListerWatcher {
		return listWatch
	}

	receiver, err := newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		receiverConfig,
		logsSink,
		fake.NewSimpleClientset(),
		listWatchFactory,
	)
	require.NoError(t, err)

	// Create the first k8s event.
	firstEvent := getEvent()
	firstEvent.UID = types.UID("ec279341-e2d8-4b2a-b17d-6e0566481001")
	listWatch.Add(firstEvent)

	// Start the receiver without storage extension.
	ctx := context.Background()
	host := componenttest.NewNopHost()
	assert.NoError(t, receiver.Start(ctx, host))

	// Create the second k8s event.
	secondEvent := getEvent()
	secondEvent.UID = types.UID("ec279341-e2d8-4b2a-b17d-6e0566481002")
	listWatch.Add(secondEvent)

	// Both events should be picked up by the receiver.
	assert.Eventually(t, func() bool {
		return logsSink.LogRecordCount() == 2
	}, 100*time.Millisecond, 10*time.Millisecond, "expected two events")

	// Shutdown the receiver.
	assert.NoError(t, receiver.Shutdown(ctx))
	for _, extension := range host.GetExtensions() {
		require.NoError(t, extension.Shutdown(ctx))
	}
	logsSink.Reset()

	// Create the third k8s event.
	thirdEvent := getEvent()
	thirdEvent.UID = types.UID("ec279341-e2d8-4b2a-b17d-6e0566481003")
	listWatch.Add(thirdEvent)

	// Start the receiver again.
	receiver, err = newRawK8sEventsReceiver(
		receivertest.NewNopSettings(),
		receiverConfig,
		logsSink,
		fake.NewSimpleClientset(),
		listWatchFactory,
	)
	require.NoError(t, err)
	assert.NoError(t, receiver.Start(ctx, componenttest.NewNopHost()))

	// Since the receiver has no storage, it should pick up events from last minute on start
	// which means it should get all three events.
	assert.Eventually(t, func() bool {
		return logsSink.LogRecordCount() == 3
	}, 100*time.Millisecond, 10*time.Millisecond, "expected 3 events")
}

func TestStorage(t *testing.T) {
	receiverConfig := createDefaultConfig().(*Config)
	logsSink := new(consumertest.LogsSink)
	listWatch := cachetest.NewFakeControllerSource()
	listWatchFactory := func(
		c cache.Getter,
		resource string,
		namespace string,
		fieldSelector fields.Selector,
	) cache.ListerWatcher {
		return listWatch
	}
	settings := receivertest.NewNopSettings()

	receiver, err := newRawK8sEventsReceiver(
		settings,
		receiverConfig,
		logsSink,
		fake.NewSimpleClientset(),
		listWatchFactory,
	)
	require.NoError(t, err)

	// Create the first k8s event.
	firstEvent := getEvent()
	firstEvent.UID = types.UID("ec279341-e2d8-4b2a-b17d-6e0566481001")
	listWatch.Add(firstEvent)

	// Start the receiver with storage extension.
	ctx := context.Background()
	storageDir := t.TempDir()
	host := storagetest.NewStorageHost().WithFileBackedStorageExtension("test", storageDir)
	assert.NoError(t, receiver.Start(ctx, host))
	t.Cleanup(func() {
		assert.NoError(t, receiver.Shutdown(ctx))
	})

	time.Sleep(10 * time.Millisecond)

	// Create the second k8s event.
	secondEvent := getEvent()
	secondEvent.UID = "ec279341-e2d8-4b2a-b17d-6e0566481002"
	listWatch.Add(secondEvent)

	// Both events should be picked up by the receiver.
	// The last resource version processed should be saved in storage.
	assert.Eventually(t, func() bool {
		return logsSink.LogRecordCount() == 2
	}, 100*time.Minute, 10*time.Millisecond, "expected 2 events")

	// Shutdown the receiver.
	require.NoError(t, receiver.Shutdown(ctx))
	for _, extension := range host.GetExtensions() {
		require.NoError(t, extension.Shutdown(ctx))
	}
	logsSink.Reset()

	// Create the third k8s event.
	thirdEvent := getEvent()
	thirdEvent.UID = types.UID("ec279341-e2d8-4b2a-b17d-6e0566481003")
	listWatch.Add(thirdEvent)

	// Start the receiver again.
	receiver, err = newRawK8sEventsReceiver(
		settings,
		receiverConfig,
		logsSink,
		fake.NewSimpleClientset(),
		listWatchFactory,
	)
	require.NoError(t, err)
	host = storagetest.NewStorageHost().WithFileBackedStorageExtension("test", storageDir)
	require.NoError(t, receiver.Start(ctx, host))

	// The receiver should only pick up the third event,
	// as it is the only one with newer resource version.
	assert.Eventually(t, func() bool {
		return logsSink.LogRecordCount() == 1
	}, 100*time.Millisecond, 10*time.Millisecond, "expected one event")
}

func getEvent() *corev1.Event {
	time := v1.Now()
	return &corev1.Event{
		TypeMeta: v1.TypeMeta{
			Kind:       "Event",
			APIVersion: "v1",
		},
		InvolvedObject: corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "test-34bcd-rn54",
			Namespace:  "test",
			UID:        types.UID("059f3edc-b5a9"),
		},
		Reason:         "testing_event_1",
		Count:          2,
		FirstTimestamp: v1.Now(),
		Type:           "Normal",
		Message:        "testing event message",
		ObjectMeta: v1.ObjectMeta{
			UID:               types.UID("289686f9-a5c0"),
			Name:              "1",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
			ManagedFields: []v1.ManagedFieldsEntry{
				{
					Manager:    "kubelite",
					Operation:  "Update",
					APIVersion: "v1",
					Time:       &time,
					FieldsType: "FieldsV1",
					FieldsV1: &v1.FieldsV1{
						Raw: []byte(`{"f:count":{}}`),
					},
				},
			},
		},
		Source: corev1.EventSource{
			Component: "testComponent",
			Host:      "testHost",
		},
	}
}

func fakeListWatchFactory(
	c cache.Getter,
	resource string,
	namespace string,
	fieldSelector fields.Selector,
) cache.ListerWatcher {
	return cachetest.NewFakeControllerSource()
}
