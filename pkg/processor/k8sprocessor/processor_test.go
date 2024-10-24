// Copyright 2020 OpenTelemetry Authors
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

package k8sprocessor

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	conventions "go.opentelemetry.io/collector/semconv/v1.18.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

func newTracesProcessor(cfg component.Config, next consumer.Traces, options ...Option) (processor.Traces, error) {
	opts := append(options, withKubeClientProvider(newFakeClient))
	return createTracesProcessorWithOptions(
		context.Background(),
		processor.Settings{
			TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		},
		cfg,
		next,
		opts...,
	)
}

func newMetricsProcessor(cfg component.Config, nextMetricsConsumer consumer.Metrics, options ...Option) (processor.Metrics, error) {
	opts := append(options, withKubeClientProvider(newFakeClient))
	return createMetricsProcessorWithOptions(
		context.Background(),
		processor.Settings{
			TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		},
		cfg,
		nextMetricsConsumer,
		opts...,
	)
}

func newLogsProcessor(cfg component.Config, nextLogsConsumer consumer.Logs, options ...Option) (processor.Logs, error) {
	opts := append(options, withKubeClientProvider(newFakeClient))
	return createLogsProcessorWithOptions(
		context.Background(),
		processor.Settings{
			TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		},
		cfg,
		nextLogsConsumer,
		opts...,
	)
}

// withKubeClientProvider sets the specific implementation for getting K8s Client instances
func withKubeClientProvider(kcp kube.ClientProvider) Option {
	return func(p *kubernetesprocessor) error {
		return p.initKubeClient(p.logger, kcp)
	}
}

// withExtractKubernetesProcessorInto allows to pull the internal model easily even when processorhelper factory is used
func withExtractKubernetesProcessorInto(kp **kubernetesprocessor) Option {
	return func(p *kubernetesprocessor) error {
		*kp = p
		return nil
	}
}

type multiTest struct {
	t *testing.T

	tp processor.Traces
	mp processor.Metrics
	lp processor.Logs

	nextTrace   *consumertest.TracesSink
	nextMetrics *consumertest.MetricsSink
	nextLogs    *consumertest.LogsSink

	kpMetrics *kubernetesprocessor
	kpTrace   *kubernetesprocessor
	kpLogs    *kubernetesprocessor
}

func newMultiTest(
	t *testing.T,
	cfg component.Config,
	errFunc func(err error),
	options ...Option,
) *multiTest {
	m := &multiTest{
		t:           t,
		nextTrace:   new(consumertest.TracesSink),
		nextMetrics: new(consumertest.MetricsSink),
		nextLogs:    new(consumertest.LogsSink),
	}

	tp, err := newTracesProcessor(cfg, m.nextTrace, append(options, withExtractKubernetesProcessorInto(&m.kpTrace))...)
	if errFunc == nil {
		assert.NotNil(t, tp)
		require.NoError(t, err)
	} else {
		assert.Nil(t, tp)
		errFunc(err)
	}

	mp, err := newMetricsProcessor(cfg, m.nextMetrics, append(options, withExtractKubernetesProcessorInto(&m.kpMetrics))...)
	if errFunc == nil {
		assert.NotNil(t, mp)
		require.NoError(t, err)
	} else {
		assert.Nil(t, mp)
		errFunc(err)
	}

	lp, err := newLogsProcessor(cfg, m.nextLogs, append(options, withExtractKubernetesProcessorInto(&m.kpLogs))...)
	if errFunc == nil {
		assert.NotNil(t, lp)
		require.NoError(t, err)
	} else {
		assert.Nil(t, lp)
		errFunc(err)
	}

	m.tp = tp
	m.mp = mp
	m.lp = lp
	return m
}

func (m *multiTest) testConsume(
	ctx context.Context,
	traces ptrace.Traces,
	metrics pmetric.Metrics,
	logs plog.Logs,
	errFunc func(err error),
) {
	errs := []error{
		m.tp.ConsumeTraces(ctx, traces),
		m.mp.ConsumeMetrics(ctx, metrics),
		m.lp.ConsumeLogs(ctx, logs),
	}

	for _, err := range errs {
		if errFunc != nil {
			errFunc(err)
		}
	}
}

func (m *multiTest) kubernetesProcessorOperation(kpOp func(kp *kubernetesprocessor)) {
	kpOp(m.kpTrace)
	kpOp(m.kpMetrics)
	kpOp(m.kpLogs)
}

func (m *multiTest) assertBatchesLen(batchesLen int) {
	require.Len(m.t, m.nextTrace.AllTraces(), batchesLen)
	require.Len(m.t, m.nextMetrics.AllMetrics(), batchesLen)
	require.Len(m.t, m.nextLogs.AllLogs(), batchesLen)
}

func (m *multiTest) assertResourceObjectLen(batchNo int) {
	assert.Equal(m.t, m.nextTrace.AllTraces()[batchNo].ResourceSpans().Len(), 1)
	assert.Equal(m.t, m.nextMetrics.AllMetrics()[batchNo].ResourceMetrics().Len(), 1)
	assert.Equal(m.t, m.nextLogs.AllLogs()[batchNo].ResourceLogs().Len(), 1)
}

func (m *multiTest) assertResourceAttributesLen(batchNo int, attrsLen int) {
	assert.Equal(m.t, m.nextTrace.AllTraces()[batchNo].ResourceSpans().At(0).Resource().Attributes().Len(), attrsLen)
	assert.Equal(m.t, m.nextMetrics.AllMetrics()[batchNo].ResourceMetrics().At(0).Resource().Attributes().Len(), attrsLen)
	assert.Equal(m.t, m.nextLogs.AllLogs()[batchNo].ResourceLogs().At(0).Resource().Attributes().Len(), attrsLen)
}

func (m *multiTest) assertResource(batchNum int, resourceFunc func(res pcommon.Resource)) {
	rss := m.nextTrace.AllTraces()[batchNum].ResourceSpans()
	r := rss.At(0).Resource()

	if resourceFunc != nil {
		resourceFunc(r)
	}
}

func TestNewProcessor(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()

	newMultiTest(t, cfg, nil)
}

func TestProcessorBadConfig(t *testing.T) {
	cfg := NewFactory().CreateDefaultConfig()
	oCfg := cfg.(*Config)
	oCfg.Extract.Metadata = []string{"bad-attribute"}

	newMultiTest(t, cfg, func(err error) {
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "\"bad-attribute\" is not a supported metadata field")
	})
}

func TestProcessorBadClientProvider(t *testing.T) {
	clientProvider := func(
		_ *zap.Logger,
		_ k8sconfig.APIConfig,
		_ kube.ExtractionRules,
		_ kube.Filters,
		_ []kube.Association,
		_ kube.Excludes,
		_ kube.APIClientsetProvider,
		_ kube.InformerProvider,
		_ kube.OwnerProvider,
		_ string,
		_ int,
		_ time.Duration,
		_ time.Duration,
	) (kube.Client, error) {
		return nil, fmt.Errorf("bad client error")
	}

	newMultiTest(t, NewFactory().CreateDefaultConfig(), func(err error) {
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "bad client error")
	}, withKubeClientProvider(clientProvider))
}

type generateResourceFunc func(res pcommon.Resource)

func generateTraces(resourceFunc ...generateResourceFunc) ptrace.Traces {
	t := ptrace.NewTraces()
	rs := t.ResourceSpans().AppendEmpty()
	for _, resFun := range resourceFunc {
		res := rs.Resource()
		resFun(res)
	}
	span := rs.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.SetName("foobar")
	return t
}

func generateMetrics(resourceFunc ...generateResourceFunc) pmetric.Metrics {
	m := pmetric.NewMetrics()
	ms := m.ResourceMetrics().AppendEmpty()
	for _, resFun := range resourceFunc {
		res := ms.Resource()
		resFun(res)
	}
	metric := ms.ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	metric.SetName("foobar")
	return m
}

func generateLogs(resourceFunc ...generateResourceFunc) plog.Logs {
	l := plog.NewLogs()
	ls := l.ResourceLogs().AppendEmpty()
	for _, resFun := range resourceFunc {
		res := ls.Resource()
		resFun(res)
	}
	ls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	return l
}

func withPassthroughIP(passthroughIP string) generateResourceFunc {
	return func(res pcommon.Resource) {
		res.Attributes().PutStr(k8sIPLabelName, passthroughIP)
	}
}

func withHostname(hostname string) generateResourceFunc {
	return func(res pcommon.Resource) {
		res.Attributes().PutStr(conventions.AttributeHostName, hostname)
	}
}

func withPodUID(uid string) generateResourceFunc {
	return func(res pcommon.Resource) {
		res.Attributes().PutStr("k8s.pod.uid", uid)
	}
}

func withPodAndNamespace(pod string, namespace string) generateResourceFunc {
	return func(res pcommon.Resource) {
		res.Attributes().PutStr("k8s.pod.name", pod)
		res.Attributes().PutStr("k8s.namespace.name", namespace)
	}
}

type strAddr string

func (s strAddr) String() string {
	return "1.1.1.1:3200"
}

func (strAddr) Network() string {
	return "tcp"
}
func TestIPDetectionFromContext(t *testing.T) {
	addresses := []net.Addr{
		&net.IPAddr{
			IP: net.IPv4(1, 1, 1, 1),
		},
		&net.TCPAddr{
			IP:   net.IPv4(1, 1, 1, 1),
			Port: 3200,
		},
		&net.UDPAddr{
			IP:   net.IPv4(1, 1, 1, 1),
			Port: 3200,
		},
		strAddr("1.1.1.1:3200"),
	}
	for _, addr := range addresses {
		m := newMultiTest(t, NewFactory().CreateDefaultConfig(), nil)
		ctx := client.NewContext(context.Background(), client.Info{
			Addr: addr,
		})
		m.testConsume(
			ctx,
			generateTraces(),
			generateMetrics(),
			generateLogs(),
			func(err error) {
				assert.NoError(t, err)
			})

		m.assertBatchesLen(1)
		m.assertResourceObjectLen(0)
		m.assertResource(0, func(r pcommon.Resource) {
			require.Greater(t, r.Attributes().Len(), 0)
			assertResourceHasStringAttribute(t, r, "k8s.pod.ip", "1.1.1.1")
		})
	}
}

func TestNilBatch(t *testing.T) {
	m := newMultiTest(t, NewFactory().CreateDefaultConfig(), nil)
	m.testConsume(
		context.Background(),
		ptrace.NewTraces(),
		pmetric.NewMetrics(),
		generateLogs(),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(1)
}

func TestProcessorNoAttrs(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
		WithExtractMetadata(metadataPodName),
	)

	addr, err := net.ResolveIPAddr("ip", "1.1.1.1")
	require.NoError(t, err)
	ctx := client.NewContext(context.Background(),
		client.Info{
			Addr: addr,
		})

	// pod doesn't have attrs to add
	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.kc.(*fakeClient).Pods["1.1.1.1"] = &kube.Pod{Name: "PodA"}
	})

	m.testConsume(
		ctx,
		generateTraces(),
		generateMetrics(),
		generateLogs(),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(1)
	m.assertResourceObjectLen(0)
	m.assertResourceAttributesLen(0, 1)

	// attrs should be added now
	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.kc.(*fakeClient).Pods["1.1.1.1"] = &kube.Pod{
			Name: "PodA",
			Attributes: map[string]string{
				"k":  "v",
				"1":  "2",
				"aa": "b",
			},
		}
	})

	m.testConsume(
		ctx,
		generateTraces(),
		generateMetrics(),
		generateLogs(),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(2)
	m.assertResourceObjectLen(1)
	m.assertResourceAttributesLen(1, 4)

	// passthrough doesn't add attrs
	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.passthroughMode = true
	})
	m.testConsume(
		ctx,
		generateTraces(),
		generateMetrics(),
		generateLogs(),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(3)
	m.assertResourceObjectLen(2)
	m.assertResourceAttributesLen(2, 1)
}

func TestNoIP(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	m.testConsume(context.Background(), generateTraces(), generateMetrics(), generateLogs(), nil)

	m.assertBatchesLen(1)
	m.assertResourceObjectLen(0)
	m.assertResource(0, func(res pcommon.Resource) {
		assert.Equal(t, 0, res.Attributes().Len())
	})
}

func TestIPSourceWithoutPodAssociation(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	type testCase struct {
		name, resourceIP, resourceK8SIP, contextIP, out string
	}

	testCases := []testCase{
		{
			name:          "k8sIP",
			resourceIP:    "1.1.1.1",
			resourceK8SIP: "2.2.2.2",
			contextIP:     "3.3.3.3",
			out:           "2.2.2.2",
		},
		{
			name:       "clientIP",
			resourceIP: "1.1.1.1",
			contextIP:  "3.3.3.3",
			out:        "1.1.1.1",
		},
		{
			name:      "contextIP",
			contextIP: "3.3.3.3",
			out:       "3.3.3.3",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.contextIP != "" {
				addr, err := net.ResolveIPAddr("ip", tc.contextIP)
				require.NoError(t, err)
				ctx = client.NewContext(context.Background(),
					client.Info{
						Addr: addr,
					})
			}

			traces := generateTraces()
			metrics := generateMetrics()
			logs := generateLogs()

			resources := []pcommon.Resource{
				traces.ResourceSpans().At(0).Resource(),
				metrics.ResourceMetrics().At(0).Resource(),
			}

			for _, res := range resources {
				if tc.resourceK8SIP != "" {
					res.Attributes().PutStr(k8sIPLabelName, tc.resourceK8SIP)
				}
				if tc.resourceIP != "" {
					res.Attributes().PutStr(clientIPLabelName, tc.resourceIP)
				}
			}

			m.testConsume(ctx, traces, metrics, logs, nil)
			m.assertBatchesLen(i + 1)
			m.assertResource(i, func(res pcommon.Resource) {
				require.Greater(t, res.Attributes().Len(), 0)
				assertResourceHasStringAttribute(t, res, "k8s.pod.ip", tc.out)
			})
		})
	}
}

func TestIPSourceWithPodAssociation(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	type testCase struct {
		name, contextIP, labelName, labelValue, outLabel, outValue string
	}

	testCases := []testCase{
		{
			name:       "k8sIP",
			contextIP:  "",
			labelName:  "k8s.pod.ip",
			labelValue: "1.1.1.1",
			outLabel:   "k8s.pod.ip",
			outValue:   "1.1.1.1",
		},
		{
			name:       "client IP",
			contextIP:  "",
			labelName:  "ip",
			labelValue: "2.2.2.2",
			outLabel:   "ip",
			outValue:   "2.2.2.2",
		},
		{
			name:       "Hostname",
			contextIP:  "",
			labelName:  "host.name",
			labelValue: "1.1.1.1",
			outLabel:   "k8s.pod.ip",
			outValue:   "1.1.1.1",
		},
	}
	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.podAssociations = []kube.Association{
			{
				From: "resource_attribute",
				Name: "k8s.pod.ip",
			},
			{
				From: "resource_attribute",
				Name: "ip",
			},
			{
				From: "resource_attribute",
				Name: "host.name",
			},
		}
	})

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.contextIP != "" {
				addr, err := net.ResolveIPAddr("ip", tc.contextIP)
				require.NoError(t, err)
				ctx = client.NewContext(context.Background(),
					client.Info{
						Addr: addr,
					})
			}

			traces := generateTraces()
			metrics := generateMetrics()
			logs := generateLogs()

			resources := []pcommon.Resource{
				traces.ResourceSpans().At(0).Resource(),
				metrics.ResourceMetrics().At(0).Resource(),
				logs.ResourceLogs().At(0).Resource(),
			}

			for _, res := range resources {
				res.Attributes().PutStr(tc.labelName, tc.labelValue)
			}

			m.testConsume(ctx, traces, metrics, logs, nil)
			m.assertBatchesLen(i + 1)
			m.assertResource(i, func(res pcommon.Resource) {
				require.Greater(t, res.Attributes().Len(), 0)
				assertResourceHasStringAttribute(t, res, tc.outLabel, tc.outValue)
			})
		})
	}
}

func TestPodUID(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)
	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.podAssociations = []kube.Association{
			{
				From: "resource_attribute",
				Name: "k8s.pod.uid",
			},
		}
		kp.kc.(*fakeClient).Pods["ef10d10b-2da5-4030-812e-5f45c1531227"] = &kube.Pod{
			Name: "PodA",
			Attributes: map[string]string{
				"k":  "v",
				"1":  "2",
				"aa": "b",
			},
		}
	})

	m.testConsume(context.Background(),
		generateTraces(withPodUID("ef10d10b-2da5-4030-812e-5f45c1531227")),
		generateMetrics(withPodUID("ef10d10b-2da5-4030-812e-5f45c1531227")),
		generateLogs(withPodUID("ef10d10b-2da5-4030-812e-5f45c1531227")),
		nil)

	m.assertBatchesLen(1)
	m.assertResourceObjectLen(0)
	m.assertResource(0, func(r pcommon.Resource) {
		require.Greater(t, r.Attributes().Len(), 0)
		assertResourceHasStringAttribute(t, r, "k8s.pod.uid", "ef10d10b-2da5-4030-812e-5f45c1531227")
	})
}

func TestProcessorAddLabels(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	tests := map[string]map[string]string{
		"1.1.1.1": {
			"pod":         "test-2323",
			"ns":          "default",
			"another tag": "value",
		},
		"2.2.2.2": {
			"pod": "test-12",
		},
	}
	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.podAssociations = []kube.Association{
			{
				From: "connection",
				Name: "ip",
			},
		}
	})

	for ip, attrs := range tests {
		m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
			kp.kc.(*fakeClient).Pods[kube.PodIdentifier(ip)] = &kube.Pod{Attributes: attrs}
		})
	}

	var i int
	for ip, attrs := range tests {
		addr, err := net.ResolveIPAddr("ip", ip)
		require.NoError(t, err)
		ctx := client.NewContext(context.Background(),
			client.Info{
				Addr: addr,
			})

		m.testConsume(
			ctx,
			generateTraces(),
			generateMetrics(),
			generateLogs(),
			func(err error) {
				assert.NoError(t, err)
			})

		m.assertBatchesLen(i + 1)
		m.assertResourceObjectLen(i)
		m.assertResource(i, func(res pcommon.Resource) {
			require.Greater(t, res.Attributes().Len(), 0)
			assertResourceHasStringAttribute(t, res, "k8s.pod.ip", ip)
			for k, v := range attrs {
				assertResourceHasStringAttribute(t, res, k, v)
			}
		})

		i++
	}
}

func TestProcessorPicksUpPassthoughPodIp(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.podAssociations = []kube.Association{
			{
				From: "resource_attribute",
				Name: "k8s.pod.ip",
			},
		}
		kp.kc.(*fakeClient).Pods["2.2.2.2"] = &kube.Pod{
			Name: "PodA",
			Attributes: map[string]string{
				"k": "v",
				"1": "2",
			},
		}
	})

	m.testConsume(
		context.Background(),
		generateTraces(withPassthroughIP("2.2.2.2")),
		generateMetrics(withPassthroughIP("2.2.2.2")),
		generateLogs(withPassthroughIP("2.2.2.2")),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(1)
	m.assertResourceObjectLen(0)
	m.assertResourceAttributesLen(0, 3)

	m.assertResource(0, func(res pcommon.Resource) {
		assertResourceHasStringAttribute(t, res, k8sIPLabelName, "2.2.2.2")
		assertResourceHasStringAttribute(t, res, "k", "v")
		assertResourceHasStringAttribute(t, res, "1", "2")
	})
}

func TestProcessorByPodNameAndNamespace(t *testing.T) {
	m := newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		nil,
	)

	m.kubernetesProcessorOperation(func(kp *kubernetesprocessor) {
		kp.podAssociations = []kube.Association{
			{
				From: "build_hostname",
				Name: "_hostname",
			},
		}
		kp.kc.(*fakeClient).Pods["PodA.test"] = &kube.Pod{
			Name: "PodA",
			Attributes: map[string]string{
				"k": "v",
				"1": "2",
			},
		}
	})

	m.testConsume(
		context.Background(),
		generateTraces(withPodAndNamespace("PodA", "test")),
		generateMetrics(withPodAndNamespace("PodA", "test")),
		generateLogs(withPodAndNamespace("PodA", "test")),
		func(err error) {
			assert.NoError(t, err)
		})

	m.assertBatchesLen(1)
	m.assertResourceObjectLen(0)
	m.assertResourceAttributesLen(0, 5)

	m.assertResource(0, func(res pcommon.Resource) {
		assertResourceHasStringAttribute(t, res, "_hostname", "PodA.test")
		assertResourceHasStringAttribute(t, res, "k", "v")
		assertResourceHasStringAttribute(t, res, "1", "2")
	})
}

func TestMetricsProcessorHostname(t *testing.T) {
	next := new(consumertest.MetricsSink)
	var kp *kubernetesprocessor
	p, err := newMetricsProcessor(
		NewFactory().CreateDefaultConfig(),
		next,
		WithExtractMetadata(metadataPodName),
		withExtractKubernetesProcessorInto(&kp),
	)
	require.NoError(t, err)
	kc := kp.kc.(*fakeClient)

	// invalid ip should not be used to lookup k8s pod
	kc.Pods["invalid-ip"] = &kube.Pod{
		Name: "PodA",
		Attributes: map[string]string{
			"k":  "v",
			"1":  "2",
			"aa": "b",
		},
	}
	kc.Pods["3.3.3.3"] = &kube.Pod{
		Name: "PodA",
		Attributes: map[string]string{
			"kk": "vv",
		},
	}

	type testCase struct {
		name, hostname string
		expectedAttrs  map[string]string
	}

	testCases := []testCase{
		{
			name:     "invalid IP in hostname",
			hostname: "invalid-ip",
			expectedAttrs: map[string]string{
				conventions.AttributeHostName: "invalid-ip",
			},
		},
		{
			name:     "valid IP in hostname",
			hostname: "3.3.3.3",
			expectedAttrs: map[string]string{
				conventions.AttributeHostName: "3.3.3.3",
				k8sIPLabelName:                "3.3.3.3",
				"kk":                          "vv",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := generateMetrics(withHostname(tc.hostname))
			assert.NoError(t, p.ConsumeMetrics(context.Background(), metrics))
			require.Len(t, next.AllMetrics(), i+1)

			md := next.AllMetrics()[i]
			require.Equal(t, 1, md.ResourceMetrics().Len())
			res := md.ResourceMetrics().At(0).Resource()
			assert.Equal(t, len(tc.expectedAttrs), res.Attributes().Len())
			for k, v := range tc.expectedAttrs {
				assertResourceHasStringAttribute(t, res, k, v)
			}
		})
	}
}

func TestMetricsProcessorHostnameWithPodAssociation(t *testing.T) {
	next := new(consumertest.MetricsSink)
	var kp *kubernetesprocessor
	p, err := newMetricsProcessor(
		NewFactory().CreateDefaultConfig(),
		next,
		WithExtractMetadata(metadataPodName),
		withExtractKubernetesProcessorInto(&kp),
	)
	require.NoError(t, err)
	kc := kp.kc.(*fakeClient)
	kp.podAssociations = []kube.Association{
		{
			From: "resource_attribute",
			Name: "host.name",
		},
	}

	// invalid ip should not be used to lookup k8s pod
	kc.Pods["invalid-ip"] = &kube.Pod{
		Name: "PodA",
		Attributes: map[string]string{
			"k":  "v",
			"1":  "2",
			"aa": "b",
		},
	}
	kc.Pods["3.3.3.3"] = &kube.Pod{
		Name: "PodA",
		Attributes: map[string]string{
			"kk": "vv",
		},
	}

	type testCase struct {
		name, hostname string
		expectedAttrs  map[string]string
	}

	testCases := []testCase{
		{
			name:     "invalid IP in hostname",
			hostname: "invalid-ip",
			expectedAttrs: map[string]string{
				conventions.AttributeHostName: "invalid-ip",
			},
		},
		{
			name:     "valid IP in hostname",
			hostname: "3.3.3.3",
			expectedAttrs: map[string]string{
				conventions.AttributeHostName: "3.3.3.3",
				k8sIPLabelName:                "3.3.3.3",
				"kk":                          "vv",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := generateMetrics(withHostname(tc.hostname))
			assert.NoError(t, p.ConsumeMetrics(context.Background(), metrics))
			require.Len(t, next.AllMetrics(), i+1)

			md := next.AllMetrics()[i]
			require.Equal(t, 1, md.ResourceMetrics().Len())
			res := md.ResourceMetrics().At(0).Resource()
			assert.Equal(t, len(tc.expectedAttrs), res.Attributes().Len())
			for k, v := range tc.expectedAttrs {
				assertResourceHasStringAttribute(t, res, k, v)
			}
		})
	}
}

func TestPassthroughStart(t *testing.T) {
	next := new(consumertest.TracesSink)
	opts := []Option{WithPassthrough()}

	p, err := newTracesProcessor(
		NewFactory().CreateDefaultConfig(),
		next,
		opts...,
	)
	require.NoError(t, err)

	// Just make sure this doesn't fail when Passthrough is enabled
	assert.NoError(t, p.Start(context.Background(), componenttest.NewNopHost()))
	assert.NoError(t, p.Shutdown(context.Background()))
}

func TestRealClient(t *testing.T) {
	newMultiTest(
		t,
		NewFactory().CreateDefaultConfig(),
		func(err error) {
			assert.Error(t, err)
			assert.Equal(t, err.Error(), "unable to load k8s config, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
		},
		withKubeClientProvider(kubeClientProvider),
		WithAPIConfig(k8sconfig.APIConfig{AuthType: "none"}),
	)
}

func TestCapabilities(t *testing.T) {
	p, err := newTracesProcessor(
		NewFactory().CreateDefaultConfig(),
		consumertest.NewNop(),
	)
	assert.NoError(t, err)
	caps := p.Capabilities()
	assert.True(t, caps.MutatesData)
}

func TestStartStop(t *testing.T) {
	var kp *kubernetesprocessor
	p, err := newTracesProcessor(
		NewFactory().CreateDefaultConfig(),
		consumertest.NewNop(),
		withExtractKubernetesProcessorInto(&kp),
	)
	require.NoError(t, err)

	assert.NoError(t, p.Start(context.Background(), componenttest.NewNopHost()))
	assert.NoError(t, p.Start(context.Background(), componenttest.NewNopHost()))

	assert.NotNil(t, kp)
	kc := kp.kc.(*fakeClient)
	controller := kc.Informer.GetController().(*kube.FakeController)

	assert.False(t, controller.HasStopped())
	assert.NoError(t, p.Shutdown(context.Background()))
	time.Sleep(time.Millisecond * 500)
	assert.True(t, controller.HasStopped())
}

func assertResourceHasStringAttribute(t *testing.T, r pcommon.Resource, k, v string) {
	got, ok := r.Attributes().Get(k)
	assert.True(t, ok, fmt.Sprintf("resource does not contain attribute %s", k))
	assert.EqualValues(t, pcommon.ValueTypeStr, got.Type(), "attribute %s is not of type string", k)
	assert.EqualValues(t, v, got.Str(), "attribute %s is not equal to %s", k, v)
}
