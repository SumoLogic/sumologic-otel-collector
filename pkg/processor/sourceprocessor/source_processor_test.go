// Copyright 2019 OpenTelemetry Authors
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

package sourceprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
)

func createConfig() *Config {
	factory := NewFactory()
	config := factory.CreateDefaultConfig().(*Config)
	config.Collector = "foocollector"
	config.SourceCategoryPrefix = "prefix/"
	config.SourceCategoryReplaceDash = "#"
	config.PodNameKey = "pod_name"
	config.PodKey = "pod"
	config.NamespaceKey = "namespace"
	config.ContainerKey = "container"
	config.PodTemplateHashKey = "pod_labels_pod-template-hash"
	config.AnnotationPrefix = "pod_annotation_"
	config.PodIDKey = "pod_id"
	return config
}

func createK8sLabels() map[string]string {
	return map[string]string{
		"namespace":                    "namespace-1",
		"pod_id":                       "pod-1234",
		"pod":                          "pod-5db86d8867-sdqlj",
		"pod_labels_pod-template-hash": "5db86d8867",
		"container":                    "container-1",
	}
}

var (
	cfg = createConfig()

	k8sLabels = createK8sLabels()

	mergedK8sLabels = map[string]string{
		"container":                    "container-1",
		"namespace":                    "namespace-1",
		"pod_id":                       "pod-1234",
		"pod":                          "pod-5db86d8867-sdqlj",
		"pod_name":                     "pod",
		"pod_labels_pod-template-hash": "5db86d8867",
		"_sourceName":                  "namespace-1.pod-5db86d8867-sdqlj.container-1",
		"_sourceCategory":              "prefix/namespace#1/pod",
	}

	mergedK8sLabelsWithMeta = map[string]string{
		"container":                    "container-1",
		"namespace":                    "namespace-1",
		"pod_id":                       "pod-1234",
		"pod":                          "pod-5db86d8867-sdqlj",
		"pod_name":                     "pod",
		"pod_labels_pod-template-hash": "5db86d8867",
		"_collector":                   "foocollector",
		"_sourceName":                  "namespace-1.pod-5db86d8867-sdqlj.container-1",
		"_sourceCategory":              "prefix/namespace#1/pod",
	}

	k8sNewLabels = map[string]string{
		"k8s.namespace.name":              "namespace-1",
		"k8s.pod.id":                      "pod-1234",
		"k8s.pod.name":                    "pod-5db86d8867-sdqlj",
		"k8s.pod.label.pod-template-hash": "5db86d8867",
		"k8s.container.name":              "container-1",
	}

	mergedK8sNewLabelsWithMeta = map[string]string{
		"k8s.namespace.name":              "namespace-1",
		"k8s.pod.id":                      "pod-1234",
		"k8s.pod.name":                    "pod-5db86d8867-sdqlj",
		"k8s.pod.label.pod-template-hash": "5db86d8867",
		"k8s.container.name":              "container-1",
		"k8s.pod.pod_name":                "pod",
		"_sourceName":                     "namespace-1.pod-5db86d8867-sdqlj.container-1",
		"_sourceCategory":                 "prefix/namespace#1/pod",
		"_collector":                      "foocollector",
	}

	limitedLabels = map[string]string{
		"pod_id": "pod-1234",
	}

	limitedLabelsWithMeta = map[string]string{
		"pod_id":     "pod-1234",
		"_collector": "foocollector",
	}
)

func newLogsDataWithLogs(resourceAttrs map[string]string, logAttrs map[string]string) pdata.Logs {
	ld := pdata.NewLogs()
	rs := ld.ResourceLogs().AppendEmpty()
	attrs := rs.Resource().Attributes()
	for k, v := range resourceAttrs {
		attrs.UpsertString(k, v)
	}

	ills := rs.InstrumentationLibraryLogs().AppendEmpty()
	log := ills.Logs().AppendEmpty()
	log.Body().SetStringVal("dummy log")
	for k, v := range logAttrs {
		log.Attributes().InsertString(k, v)
	}

	return ld
}

func newTraceData(labels map[string]string) pdata.Traces {
	td := pdata.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	attrs := rs.Resource().Attributes()
	for k, v := range labels {
		attrs.UpsertString(k, v)
	}
	return td
}

func newTraceDataWithSpans(_resourceLabels map[string]string, _spanLabels map[string]string) pdata.Traces {
	// This will be very small attribute set, the actual data will be at span level
	td := newTraceData(_resourceLabels)
	ils := td.ResourceSpans().At(0).InstrumentationLibrarySpans().AppendEmpty()
	span := ils.Spans().AppendEmpty()
	span.SetName("foo")
	spanAttrs := span.Attributes()
	for k, v := range _spanLabels {
		spanAttrs.UpsertString(k, v)
	}
	return td
}

func prepareAttributesForAssert(t pdata.Traces) {
	for i := 0; i < t.ResourceSpans().Len(); i++ {
		rss := t.ResourceSpans().At(i)
		rss.Resource().Attributes().Sort()
		for j := 0; j < rss.InstrumentationLibrarySpans().Len(); j++ {
			ss := rss.InstrumentationLibrarySpans().At(j).Spans()
			for k := 0; k < ss.Len(); k++ {
				ss.At(k).Attributes().Sort()
			}
		}
	}
}

func assertTracesEqual(t *testing.T, t1 pdata.Traces, t2 pdata.Traces) {
	prepareAttributesForAssert(t1)
	prepareAttributesForAssert(t2)
	assert.Equal(t, t1, t2)
}

func assertSpansEqual(t *testing.T, t1 pdata.Traces, t2 pdata.Traces) {
	prepareAttributesForAssert(t1)
	prepareAttributesForAssert(t2)
	assert.Equal(t, t1.ResourceSpans().Len(), t2.ResourceSpans().Len())
	for i := 0; i < t1.ResourceSpans().Len(); i++ {
		rss1 := t1.ResourceSpans().At(i)
		rss2 := t2.ResourceSpans().At(i)
		assert.Equal(t, rss1.InstrumentationLibrarySpans(), rss2.InstrumentationLibrarySpans())
	}
}

func TestLogsSourceHostKey(t *testing.T) {
	resourceAttrs := map[string]string{
		"_SYSTEMD_UNIT": `docker.service`,
		"_HOSTNAME":     `sumologic-kubernetes-collection-hostname`,
	}
	logAttrs := map[string]string{
		"PRIORITY":               `6`,
		"_BOOT_ID":               `7878e140730d4ee89fadd300c929a892`,
		"_MACHINE_ID":            `1e54ca203e554cda9c944fd7f00e94e1`,
		"MESSAGE":                `time="2021-09-15T17:31:49.523983251+02:00" level=info msg="ignoring event"module=libcontainerd namespace=moby`,
		"_TRANSPORT":             `stdout`,
		"SYSLOG_FACILITY":        `3`,
		"_UID":                   `0`,
		"_GID":                   `0`,
		"_CAP_EFFECTIVE":         `3fffffffff`,
		"_SELINUX_CONTEXT":       `unconfined`,
		"_SYSTEMD_SLICE":         `system.slice`,
		"_STREAM_ID":             `f16d7d0400a44f228554869816ab1dfa`,
		"SYSLOG_IDENTIFIER":      `dockerd`,
		"_PID":                   `1061`,
		"_COMM":                  `dockerd`,
		"_EXE":                   `/usr/bin/dockerd`,
		"_CMDLINE":               `/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock`,
		"_SYSTEMD_CGROUP":        `/system.slice/docker.service`,
		"fluent.tag":             `host.docker.service`,
		"_SYSTEMD_INVOCATION_ID": `1b689467e52f4fc4aa95d3c36fa1d7fa`,
	}

	t.Run("works using existing resource attribute", func(t *testing.T) {
		config := NewFactory().CreateDefaultConfig().(*Config)
		config.SourceName = "will-it-work-%{_HOSTNAME}"
		config.SourceHostKey = "_HOSTNAME"

		pLogs := newLogsDataWithLogs(resourceAttrs, logAttrs)

		sp := newSourceProcessor(config)
		out, err := sp.ProcessLogs(context.Background(), pLogs)
		require.NoError(t, err)

		out.ResourceLogs().At(0).Resource().Attributes().Range(func(k string, v pdata.AttributeValue) bool {
			t.Logf("k %s : v %v\n", k, v.StringVal())
			return true
		})

		require.Equal(t, out.ResourceLogs().Len(), 1)
		resAttrs := out.ResourceLogs().At(0).Resource().Attributes()

		{
			v, ok := resAttrs.Get("_sourceName")
			require.True(t, ok)
			assert.Equal(t, "will-it-work-sumologic-kubernetes-collection-hostname", v.StringVal())
		}

		{
			v, ok := resAttrs.Get("_sourceHost")
			require.True(t, ok)
			assert.Equal(t, "sumologic-kubernetes-collection-hostname", v.StringVal())
		}
	})

	t.Run("does not work using record attribute", func(t *testing.T) {
		config := NewFactory().CreateDefaultConfig().(*Config)
		config.SourceName = "will-it-work-%{_CMDLINE}"
		config.SourceHostKey = "_CMDLINE"

		pLogs := newLogsDataWithLogs(resourceAttrs, logAttrs)

		sp := newSourceProcessor(config)
		out, err := sp.ProcessLogs(context.Background(), pLogs)
		require.NoError(t, err)

		out.ResourceLogs().At(0).Resource().Attributes().Range(func(k string, v pdata.AttributeValue) bool {
			t.Logf("k %s : v %v\n", k, v.StringVal())
			return true
		})

		require.Equal(t, out.ResourceLogs().Len(), 1)
		resAttrs := out.ResourceLogs().At(0).Resource().Attributes()

		{
			_, ok := resAttrs.Get("_sourceName")
			require.False(t, ok)
		}

		{
			_, ok := resAttrs.Get("_sourceHost")
			require.False(t, ok)
		}
	})
}

func TestTraceSourceProcessor(t *testing.T) {
	want := newTraceData(mergedK8sLabelsWithMeta)
	test := newTraceData(k8sLabels)

	rtp := newSourceProcessor(cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, td, want)
}

func TestTraceSourceProcessorNewTaxonomy(t *testing.T) {
	want := newTraceData(mergedK8sNewLabelsWithMeta)
	test := newTraceData(k8sNewLabels)

	config := createConfig()
	config.NamespaceKey = "k8s.namespace.name"
	config.PodIDKey = "k8s.pod.id"
	config.PodNameKey = "k8s.pod.pod_name"
	config.PodKey = "k8s.pod.name"
	config.PodTemplateHashKey = "k8s.pod.label.pod-template-hash"
	config.ContainerKey = "k8s.container.name"

	rtp := newSourceProcessor(config)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, td, want)
}

func TestTraceSourceProcessorEmpty(t *testing.T) {
	want := newTraceData(limitedLabelsWithMeta)
	test := newTraceData(limitedLabels)

	rtp := newSourceProcessor(cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)
	assertTracesEqual(t, td, want)
}

func TestTraceSourceFilteringOutByRegex(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *Config
		want pdata.Traces
	}{
		{
			name: "pod exclude regex",
			cfg: func() *Config {
				cfg := createConfig()
				cfg.Exclude = map[string]string{
					"pod": ".*",
				}
				return cfg
			}(),
			want: func() pdata.Traces {
				want := newTraceDataWithSpans(mergedK8sLabelsWithMeta, k8sLabels)
				want.ResourceSpans().At(0).InstrumentationLibrarySpans().
					RemoveIf(func(pdata.InstrumentationLibrarySpans) bool { return true })
				return want
			}(),
		},
		{
			name: "container exclude regex",
			cfg: func() *Config {
				cfg := createConfig()
				cfg.Exclude = map[string]string{
					"container": ".*",
				}
				return cfg
			}(),
			want: func() pdata.Traces {
				want := newTraceDataWithSpans(mergedK8sLabelsWithMeta, k8sLabels)
				want.ResourceSpans().At(0).InstrumentationLibrarySpans().
					RemoveIf(func(pdata.InstrumentationLibrarySpans) bool { return true })
				return want
			}(),
		},
		{
			name: "namespace exclude regex",
			cfg: func() *Config {
				cfg := createConfig()
				cfg.Exclude = map[string]string{
					"namespace": ".*",
				}
				return cfg
			}(),
			want: func() pdata.Traces {
				want := newTraceDataWithSpans(mergedK8sLabelsWithMeta, k8sLabels)
				want.ResourceSpans().At(0).InstrumentationLibrarySpans().
					RemoveIf(func(pdata.InstrumentationLibrarySpans) bool { return true })
				return want
			}(),
		},
		{
			name: "no exclude regex",
			cfg: func() *Config {
				return createConfig()
			}(),
			want: func() pdata.Traces {
				return newTraceDataWithSpans(mergedK8sLabelsWithMeta, k8sLabels)
			}(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test := newTraceDataWithSpans(mergedK8sLabels, k8sLabels)

			rtp := newSourceProcessor(tc.cfg)

			td, err := rtp.ProcessTraces(context.Background(), test)
			assert.NoError(t, err)

			assertTracesEqual(t, td, tc.want)
		})
	}
}

func TestTraceSourceFilteringOutByExclude(t *testing.T) {
	test := newTraceDataWithSpans(k8sLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().
		UpsertString("pod_annotation_sumologic.com/exclude", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, mergedK8sLabels)
	want.ResourceSpans().At(0).InstrumentationLibrarySpans().
		RemoveIf(func(pdata.InstrumentationLibrarySpans) bool { return true })

	rtp := newSourceProcessor(cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertSpansEqual(t, td, want)
}

func TestTraceSourceIncludePrecedence(t *testing.T) {
	test := newTraceDataWithSpans(limitedLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().UpsertString("pod_annotation_sumologic.com/include", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, k8sLabels)
	want.ResourceSpans().At(0).Resource().Attributes().UpsertString("pod_annotation_sumologic.com/include", "true")

	cfg1 := createConfig()
	cfg1.Exclude = map[string]string{
		"pod": ".*",
	}
	rtp := newSourceProcessor(cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, td, want)
}

func TestTraceSourceProcessorAnnotations(t *testing.T) {
	k8sLabels["pod_annotation_sumologic.com/sourceHost"] = "sh:%{pod_id}"
	k8sLabels["pod_annotation_sumologic.com/sourceCategory"] = "sc:%{pod_id}"
	test := newTraceData(k8sLabels)

	mergedK8sLabelsWithMeta["pod_annotation_sumologic.com/sourceHost"] = "sh:%{pod_id}"
	mergedK8sLabelsWithMeta["pod_annotation_sumologic.com/sourceCategory"] = "sc:%{pod_id}"
	mergedK8sLabelsWithMeta["_sourceHost"] = "sh:pod-1234"
	mergedK8sLabelsWithMeta["_sourceCategory"] = "prefix/sc:pod#1234"
	want := newTraceData(mergedK8sLabelsWithMeta)

	rtp := newSourceProcessor(cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, td, want)
}

func TestTemplateWithCustomAttribute(t *testing.T) {
	t.Run("attribute name is a single word", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["someattr"] = "somevalue"
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{someattr}/123"

		processedTraces, err := newSourceProcessor(config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/somevalue/123")
	})

	t.Run("attribute name has dot in it", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["some.attr"] = "somevalue"
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{some.attr}/123"

		processedTraces, err := newSourceProcessor(config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/%!{(MISSING)some.attr}/123")
	})
}

func assertAttribute(t *testing.T, attributes pdata.AttributeMap, attributeName string, expectedValue string) {
	value, exists := attributes.Get(attributeName)

	if expectedValue == "" {
		assert.False(t, exists, "Attribute '%s' should not exist", attributeName)
	} else {
		assert.True(t, exists)
		actualValue := value.StringVal()
		assert.Equal(t, expectedValue, actualValue, "Attribute '%s' should be '%s' but was '%s'", attributeName, expectedValue, actualValue)

	}
}

func TestLogProcessorJson(t *testing.T) {
	testcases := []struct {
		name               string
		body               string
		expectedBody       string
		expectedAttributes map[string]string
		testLogs           pdata.Logs
	}{
		{
			name:         "dockerFormat",
			body:         `{"log": "test\n", "stream": "stdout", "time": "2021"}`,
			expectedBody: "test",
			expectedAttributes: map[string]string{
				"stream": "stdout",
				"time":   "2021",
			},
		},
		{
			// additional fields are going to be removed
			name:         "additionalFields",
			body:         `{"log": "test", "stream": "stdout", "time": "2021", "additional_field": "random_value"}`,
			expectedBody: "test",
			expectedAttributes: map[string]string{
				"stream": "stdout",
				"time":   "2021",
			},
		},
		{
			// nested json log is treated as invalid data and not processed
			name:               "nestedLog",
			body:               `{"log": {"nested_key": "nested_value"}, "stream": "stdout", "time": "2021"}`,
			expectedBody:       `{"log": {"nested_key": "nested_value"}, "stream": "stdout", "time": "2021"}`,
			expectedAttributes: map[string]string{},
		},
		{
			// non docker json log is not parsed
			name:               "additionalFields",
			body:               `{"log": "some_log", "stream": "stdout"}`,
			expectedBody:       `{"log": "some_log", "stream": "stdout"}`,
			expectedAttributes: map[string]string{},
		},
		{
			// non json log is not parsed
			name:               "additionalFields",
			body:               `log": "some_log", "stream": "stdout"}`,
			expectedBody:       `log": "some_log", "stream": "stdout"}`,
			expectedAttributes: map[string]string{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			inputLog := pdata.NewLogs()
			inputLog.
				ResourceLogs().
				AppendEmpty().
				InstrumentationLibraryLogs().
				AppendEmpty().
				Logs().
				AppendEmpty().
				Body().
				SetStringVal(tc.body)

			rtp := newSourceProcessor(cfg)

			td, err := rtp.ProcessLogs(context.Background(), inputLog)
			assert.NoError(t, err)

			rss := td.ResourceLogs()
			require.Equal(t, 1, rss.Len())

			ills := rss.At(0).InstrumentationLibraryLogs()
			require.Equal(t, 1, ills.Len())

			logs := ills.At(0).Logs()
			require.Equal(t, 1, logs.Len())

			log := logs.At(0)
			assert.Equal(t, tc.expectedBody, log.Body().AsString())

			for key, value := range tc.expectedAttributes {
				attr, ok := log.Attributes().Get(key)
				require.True(t, ok)
				assert.Equal(t, value, attr.AsString())
			}
		})
	}
}
