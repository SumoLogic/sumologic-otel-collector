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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
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

func (er *countingErrorConsumer) ConsumeLogs(context.Context, pdata.Logs) error {
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
		componenttest.NewNopReceiverCreateSettings(),
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
		componenttest.NewNopReceiverCreateSettings(),
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

func TestProcessEvent(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	sink := new(consumertest.LogsSink)
	r, err := newRawK8sEventsReceiver(
		componenttest.NewNopReceiverCreateSettings(),
		rCfg,
		sink,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	r.ctx = context.Background()
	k8sEvent := getEvent()
	r.processEvent(context.Background(), k8sEvent)

	assert.Equal(t, sink.LogRecordCount(), 1)
}

func TestConsumeRetryOnRecoverableError(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	core, observedLogs := observer.New(zapcore.InfoLevel)
	consumerError := errors.New("recoverable error")
	consumer := newCountingErrorConsumer(consumerError)
	settings := componenttest.NewNopReceiverCreateSettings()
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

	logs := pdata.Logs{}
	err = receiver.consumeWithRetry(ctx, logs)
	assert.Error(t, err)
	assert.Equal(t, int(rCfg.ConsumeMaxRetries+1), consumer.CallCount)
	assert.Equal(t, int(rCfg.ConsumeMaxRetries), observedLogs.Len())
	assert.Equal(t, int(rCfg.ConsumeMaxRetries), observedLogs.FilterMessage("ConsumeMetrics() recoverable error, will retry").Len())
}

func TestConsumeNoRetryOnPermanentError(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	client := fake.NewSimpleClientset()
	ctx := context.Background()
	core, observedLogs := observer.New(zapcore.InfoLevel)
	consumerError := consumererror.NewPermanent(errors.New("permanent error"))
	consumer := newCountingErrorConsumer(consumerError)
	settings := componenttest.NewNopReceiverCreateSettings()
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

	logs := pdata.Logs{}
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
		componenttest.NewNopReceiverCreateSettings(),
		rCfg,
		sink,
		client,
		fakeListWatchFactory,
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	r.ctx = context.Background()
	k8sEvent := getEvent()
	logs := r.convertToLog(k8sEvent)
	assert.Equal(t, logs.LogRecordCount(), 1)
	// TODO: Add some real tests about the data format
}

func TestEventFilterByTime(t *testing.T) {
	// TODO: add this
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
			ClusterName:       "testCluster",
			CreationTimestamp: v1.Now(),
			ManagedFields: []v1.ManagedFieldsEntry{
				{
					Manager:    "kubelite",
					Operation:  "Update",
					APIVersion: "v1",
					Time:       &time,
					FieldsType: "FieldsV1",
					FieldsV1: &v1.FieldsV1{
						Raw: []byte(`{"f:count": {}}`),
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
