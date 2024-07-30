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
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/ptracetest"
)

func createConfig() *Config {
	factory := NewFactory()
	config := factory.CreateDefaultConfig().(*Config)
	config.Collector = "foocollector"
	config.SourceCategoryPrefix = "prefix/"
	config.SourceCategoryReplaceDash = "#"
	config.PodNameKey = "k8s.pod.pod_name"
	config.PodKey = "k8s.pod.name"
	config.PodTemplateHashKey = "k8s.pod.label.pod-template-hash"
	config.AnnotationPrefix = "pod_annotation_"
	config.NamespaceAnnotationPrefix = "namespace_annotation_"
	return config
}

func createK8sLabels() map[string]string {
	return map[string]string{
		"k8s.namespace.name":              "namespace-1",
		"k8s.pod.uid":                     "pod-1234",
		"k8s.pod.name":                    "pod-5db86d8867-sdqlj",
		"k8s.pod.label.pod-template-hash": "5db86d8867",
		"k8s.container.name":              "container-1",
	}
}

var (
	cfg = createConfig()

	k8sLabels = createK8sLabels()

	mergedK8sLabels = map[string]string{
		"k8s.namespace.name":              "namespace-1",
		"k8s.pod.uid":                     "pod-1234",
		"k8s.pod.name":                    "pod-5db86d8867-sdqlj",
		"k8s.pod.label.pod-template-hash": "5db86d8867",
		"k8s.container.name":              "container-1",
		"k8s.pod.pod_name":                "pod",
		"_sourceHost":                     "undefined",
		"_sourceName":                     "namespace-1.pod-5db86d8867-sdqlj.container-1",
		"_sourceCategory":                 "prefix/namespace#1/pod",
		"_collector":                      "foocollector",
	}

	limitedLabels = map[string]string{
		"k8s.pod.uid": "pod-1234",
	}

	limitedLabelsWithMeta = map[string]string{
		"k8s.pod.uid":     "pod-1234",
		"_collector":      "foocollector",
		"_sourceCategory": "prefix/undefined/undefined",
		"_sourceHost":     "undefined",
		"_sourceName":     "undefined.undefined.undefined",
	}
)

func newLogsDataWithLogs(resourceAttrs map[string]string, logAttrs map[string]string) plog.Logs {
	ld := plog.NewLogs()
	rs := ld.ResourceLogs().AppendEmpty()
	attrs := rs.Resource().Attributes()
	for k, v := range resourceAttrs {
		attrs.PutStr(k, v)
	}

	sls := rs.ScopeLogs().AppendEmpty()
	log := sls.LogRecords().AppendEmpty()
	log.Body().SetStr("dummy log")
	for k, v := range logAttrs {
		log.Attributes().PutStr(k, v)
	}

	return ld
}

func newTraceData(labels map[string]string) ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	attrs := rs.Resource().Attributes()
	for k, v := range labels {
		attrs.PutStr(k, v)
	}
	return td
}

func newTraceDataWithSpans(_resourceLabels map[string]string, _spanLabels map[string]string) ptrace.Traces {
	// This will be very small attribute set, the actual data will be at span level
	td := newTraceData(_resourceLabels)
	sls := td.ResourceSpans().At(0).ScopeSpans().AppendEmpty()
	span := sls.Spans().AppendEmpty()
	span.SetName("foo")
	spanAttrs := span.Attributes()
	for k, v := range _spanLabels {
		spanAttrs.PutStr(k, v)
	}
	return td
}

func assertTracesEqual(t *testing.T, t1 ptrace.Traces, t2 ptrace.Traces) {
	err := ptracetest.CompareTraces(t1, t2)
	assert.NoError(t, err)
}

func assertSpansEqual(t *testing.T, t1 ptrace.Traces, t2 ptrace.Traces) {
	assert.Equal(t, t1.ResourceSpans().Len(), t2.ResourceSpans().Len())
	for i := 0; i < t1.ResourceSpans().Len(); i++ {
		rss1 := t1.ResourceSpans().At(i)
		rss2 := t2.ResourceSpans().At(i)
		assert.Equal(t, rss1.ScopeSpans().Len(), rss2.ScopeSpans().Len())

		for j := 0; j < rss1.ScopeSpans().Len(); j++ {
			err := ptracetest.CompareScopeSpans(rss1.ScopeSpans().At(j), rss2.ScopeSpans().At(j))
			assert.NoError(t, err)
		}
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
		config.SourceHost = "%{_HOSTNAME}"

		pLogs := newLogsDataWithLogs(resourceAttrs, logAttrs)

		sp := newSourceProcessor(newProcessorCreateSettings(), config)
		out, err := sp.ProcessLogs(context.Background(), pLogs)
		require.NoError(t, err)

		out.ResourceLogs().At(0).Resource().Attributes().Range(func(k string, v pcommon.Value) bool {
			t.Logf("k %s : v %v\n", k, v.Str())
			return true
		})

		require.Equal(t, out.ResourceLogs().Len(), 1)
		resAttrs := out.ResourceLogs().At(0).Resource().Attributes()

		{
			v, ok := resAttrs.Get("_sourceName")
			require.True(t, ok)
			assert.Equal(t, "will-it-work-sumologic-kubernetes-collection-hostname", v.Str())
		}

		{
			v, ok := resAttrs.Get("_sourceHost")
			require.True(t, ok)
			assert.Equal(t, "sumologic-kubernetes-collection-hostname", v.Str())
		}
	})

	t.Run("does not work using record attribute", func(t *testing.T) {
		config := NewFactory().CreateDefaultConfig().(*Config)
		config.SourceName = "will-it-work-%{_CMDLINE}"
		config.SourceHost = "%{_CMDLINE}"

		pLogs := newLogsDataWithLogs(resourceAttrs, logAttrs)

		sp := newSourceProcessor(newProcessorCreateSettings(), config)
		out, err := sp.ProcessLogs(context.Background(), pLogs)
		require.NoError(t, err)

		out.ResourceLogs().At(0).Resource().Attributes().Range(func(k string, v pcommon.Value) bool {
			t.Logf("k %s : v %v\n", k, v.Str())
			return true
		})

		require.Equal(t, out.ResourceLogs().Len(), 1)
		resAttrs := out.ResourceLogs().At(0).Resource().Attributes()

		assertAttribute(t, resAttrs, "_sourceName", "will-it-work-undefined")
		assertAttribute(t, resAttrs, "_sourceHost", "undefined")
	})
}

func TestTraceSourceProcessor(t *testing.T) {
	want := newTraceData(mergedK8sLabels)
	test := newTraceData(k8sLabels)

	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, want, td)
}

func TestTraceSourceProcessorEmpty(t *testing.T) {
	want := newTraceData(limitedLabelsWithMeta)
	test := newTraceData(limitedLabels)

	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)
	assertTracesEqual(t, want, td)
}

func TestTraceSourceFilteringOutByRegex(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *Config
		want ptrace.Traces
	}{
		{
			name: "pod exclude regex",
			cfg: func() *Config {
				cfg := createConfig()
				cfg.Exclude = map[string]string{
					"k8s.pod.name": ".*",
				}
				return cfg
			}(),
			want: func() ptrace.Traces {
				want := newTraceDataWithSpans(mergedK8sLabels, k8sLabels)
				want.ResourceSpans().At(0).ScopeSpans().
					RemoveIf(func(ptrace.ScopeSpans) bool { return true })
				return want
			}(),
		},
		{
			name: "container exclude regex",
			cfg: func() *Config {
				cfg := createConfig()
				cfg.Exclude = map[string]string{
					"k8s.container.name": ".*",
				}
				return cfg
			}(),
			want: func() ptrace.Traces {
				want := newTraceDataWithSpans(mergedK8sLabels, k8sLabels)
				want.ResourceSpans().At(0).ScopeSpans().
					RemoveIf(func(ptrace.ScopeSpans) bool { return true })
				return want
			}(),
		},
		{
			name: "namespace exclude regex",
			cfg: func() *Config {
				cfg := createConfig()
				cfg.Exclude = map[string]string{
					"k8s.namespace.name": ".*",
				}
				return cfg
			}(),
			want: func() ptrace.Traces {
				want := newTraceDataWithSpans(mergedK8sLabels, k8sLabels)
				want.ResourceSpans().At(0).ScopeSpans().
					RemoveIf(func(ptrace.ScopeSpans) bool { return true })
				return want
			}(),
		},
		{
			name: "no exclude regex",
			cfg: func() *Config {
				return createConfig()
			}(),
			want: func() ptrace.Traces {
				return newTraceDataWithSpans(mergedK8sLabels, k8sLabels)
			}(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test := newTraceDataWithSpans(mergedK8sLabels, k8sLabels)

			rtp := newSourceProcessor(newProcessorCreateSettings(), tc.cfg)

			td, err := rtp.ProcessTraces(context.Background(), test)
			assert.NoError(t, err)

			assertTracesEqual(t, td, tc.want)
		})
	}
}

func TestTraceSourceFilteringOutByExclude(t *testing.T) {
	test := newTraceDataWithSpans(k8sLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().
		PutStr("pod_annotation_sumologic.com/exclude", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, mergedK8sLabels)
	want.ResourceSpans().At(0).ScopeSpans().
		RemoveIf(func(ptrace.ScopeSpans) bool { return true })

	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertSpansEqual(t, want, td)
}

func TestTraceSourceFilteringOutByExcludeNamespaceAnnotation(t *testing.T) {
	test := newTraceDataWithSpans(k8sLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().
		PutStr("namespace_annotation_sumologic.com/exclude", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, mergedK8sLabels)
	want.ResourceSpans().At(0).ScopeSpans().
		RemoveIf(func(ptrace.ScopeSpans) bool { return true })

	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertSpansEqual(t, want, td)
}

func TestTraceSourceIncludePrecedenceByNamespaceAnnotation(t *testing.T) {
	test := newTraceDataWithSpans(limitedLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().PutStr("namespace_annotation_sumologic.com/include", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, k8sLabels)
	want.ResourceSpans().At(0).Resource().Attributes().PutStr("namespace_annotation_sumologic.com/include", "true")

	cfg1 := createConfig()
	cfg1.Exclude = map[string]string{
		"pod": ".*",
	}
	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg1)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, want, td)
}

func TestTraceSourceIncludePrecedence(t *testing.T) {
	test := newTraceDataWithSpans(limitedLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().PutStr("pod_annotation_sumologic.com/include", "true")
	test.ResourceSpans().At(0).Resource().Attributes().PutStr("namespace_annotation_sumologic.com/exclude", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, k8sLabels)
	want.ResourceSpans().At(0).Resource().Attributes().PutStr("pod_annotation_sumologic.com/include", "true")
	want.ResourceSpans().At(0).Resource().Attributes().PutStr("namespace_annotation_sumologic.com/exclude", "true")

	cfg1 := createConfig()
	cfg1.Exclude = map[string]string{
		"pod": ".*",
	}
	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg1)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, want, td)
}

func TestTraceSourceExcludePrecedence(t *testing.T) {
	test := newTraceDataWithSpans(limitedLabels, k8sLabels)
	test.ResourceSpans().At(0).Resource().Attributes().PutStr("pod_annotation_sumologic.com/exclude", "true")
	test.ResourceSpans().At(0).Resource().Attributes().PutStr("namespace_annotation_sumologic.com/include", "true")

	want := newTraceDataWithSpans(limitedLabelsWithMeta, k8sLabels)
	want.ResourceSpans().At(0).Resource().Attributes().PutStr("pod_annotation_sumologic.com/exclude", "true")
	want.ResourceSpans().At(0).Resource().Attributes().PutStr("namespace_annotation_sumologic.com/include", "true")
	want.ResourceSpans().At(0).ScopeSpans().
		RemoveIf(func(ptrace.ScopeSpans) bool { return true })

	rtp := newSourceProcessor(newProcessorCreateSettings(), cfg)

	td, err := rtp.ProcessTraces(context.Background(), test)
	assert.NoError(t, err)

	assertTracesEqual(t, want, td)
}

func TestSourceHostAnnotation(t *testing.T) {
	inputAttributes := createK8sLabels()
	inputAttributes["pod_annotation_sumologic.com/sourceHost"] = "sh:%{k8s.pod.uid}"
	inputAttributes["namespace_annotation_sumologic.com/sourceHost"] = "namespaceTest:%{k8s.pod.uid}"
	inputTraces := newTraceData(inputAttributes)

	processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
	assert.NoError(t, err)

	processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
	assertAttribute(t, processedAttributes, "_sourceHost", "sh:pod-1234")
}

func TestSourceHostNamespaceAnnotation(t *testing.T) {
	inputAttributes := createK8sLabels()
	inputAttributes["namespace_annotation_sumologic.com/sourceHost"] = "namespaceTest:%{k8s.pod.uid}"
	inputTraces := newTraceData(inputAttributes)

	processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
	assert.NoError(t, err)

	processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
	assertAttribute(t, processedAttributes, "_sourceHost", "namespaceTest:pod-1234")
}

func TestSourceNameAnnotation(t *testing.T) {
	inputAttributes := createK8sLabels()
	inputAttributes["pod_annotation_sumologic.com/sourceName"] = "sn:%{k8s.pod.name}"
	inputAttributes["namespace_annotation_sumologic.com/sourceName"] = "namespaceTest:%{k8s.pod.name}"
	inputTraces := newTraceData(inputAttributes)

	processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
	assert.NoError(t, err)

	processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
	assertAttribute(t, processedAttributes, "_sourceName", "sn:pod-5db86d8867-sdqlj")
}

func TestSourceNameNamespaceAnnotation(t *testing.T) {
	inputAttributes := createK8sLabels()
	inputAttributes["namespace_annotation_sumologic.com/sourceName"] = "namespaceTest:%{k8s.pod.name}"
	inputTraces := newTraceData(inputAttributes)

	processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
	assert.NoError(t, err)

	processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
	assertAttribute(t, processedAttributes, "_sourceName", "namespaceTest:pod-5db86d8867-sdqlj")
}

func TestSourceCategoryAnnotations(t *testing.T) {
	t.Run("source category annotation", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_annotation_sumologic.com/sourceCategory"] = "sc-%{k8s.namespace.name}"
		inputTraces := newTraceData(inputAttributes)

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "prefix/sc#namespace#1")
	})

	t.Run("source category prefix annotation", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_annotation_sumologic.com/sourceCategoryPrefix"] = "annot>"
		inputTraces := newTraceData(inputAttributes)

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "annot>namespace#1/pod")
	})

	t.Run("source category empty prefix annotation", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_annotation_sumologic.com/sourceCategoryPrefix"] = ""
		inputTraces := newTraceData(inputAttributes)

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "namespace#1/pod")
	})

	t.Run("source category dash replacement annotation", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_annotation_sumologic.com/sourceCategoryReplaceDash"] = "^"
		inputTraces := newTraceData(inputAttributes)

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "prefix/namespace^1/pod")
	})

	t.Run("all source category annotations together", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_annotation_sumologic.com/sourceCategory"] = "sc-%{k8s.namespace.name}"
		inputAttributes["pod_annotation_sumologic.com/sourceCategoryPrefix"] = "annot>"
		inputAttributes["pod_annotation_sumologic.com/sourceCategoryReplaceDash"] = "^"
		inputTraces := newTraceData(inputAttributes)

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "annot>sc^namespace^1")
	})

	t.Run("container-level annotations with default settings", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_annotation_sumologic.com/container-1.sourceCategory"] = "container-sc"
		inputTraces := newTraceData(inputAttributes)

		cfg.ContainerAnnotations.Enabled = true
		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "container-sc")
	})

	t.Run("container-level annotations with custom settings", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		delete(inputAttributes, "k8s.container.name")
		inputAttributes["containername"] = "container-2"
		inputAttributes["pod_annotation_custom.prefix/container-2.sourceCategory"] = "container-sc"
		inputTraces := newTraceData(inputAttributes)

		cfg.ContainerAnnotations.Enabled = true
		cfg.ContainerAnnotations.ContainerNameKey = "containername"
		cfg.ContainerAnnotations.Prefixes = []string{"custom.prefix/"}
		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), cfg).ProcessTraces(context.Background(), inputTraces)
		assert.NoError(t, err)

		processedAttributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, processedAttributes, "_sourceCategory", "container-sc")
	})
}

func TestSourceCategoryTemplateWithCustomAttribute(t *testing.T) {
	t.Run("attribute name is a single word", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["someattr"] = "somevalue"
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{someattr}/123"

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), config).ProcessTraces(context.Background(), traces)
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

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/somevalue/123")
	})

	t.Run("attribute does not exist", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{nonexistent.attr}/123"

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/undefined/123")
	})

	t.Run("attribute does not exist and contains a slash", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{pod_labels_app.kubernetes.io/name}/123"

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/undefined/123")
	})

	t.Run("attribute contains a slash", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		inputAttributes["pod_labels_app.kubernetes.io/name"] = "foobar"
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{pod_labels_app.kubernetes.io/name}/123"

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/foobar/123")
	})

	t.Run("attribute is a collector name", func(t *testing.T) {
		inputAttributes := createK8sLabels()
		traces := newTraceData(inputAttributes)

		config := createDefaultConfig().(*Config)
		config.SourceCategory = "abc/%{_collector}/123"
		config.Collector = "my-collector"

		processedTraces, err := newSourceProcessor(newProcessorCreateSettings(), config).ProcessTraces(context.Background(), traces)
		assert.NoError(t, err)

		attributes := processedTraces.ResourceSpans().At(0).Resource().Attributes()
		assertAttribute(t, attributes, "_sourceCategory", "kubernetes/abc/my/collector/123")
	})
}

func assertAttribute(t *testing.T, attributes pcommon.Map, attributeName string, expectedValue string) {
	value, exists := attributes.Get(attributeName)

	if !exists {
		assert.False(t, exists, "Attribute '%s' should not exist.", attributeName)
	} else {
		assert.True(t, exists, "Attribute '%s' should exist, but it does not.", attributeName)
		actualValue := value.Str()
		assert.Equal(t, expectedValue, actualValue, "Attribute '%s' should be '%s', but was '%s'.", attributeName, expectedValue, actualValue)
	}
}

func TestLogProcessorJson(t *testing.T) {
	testcases := []struct {
		name               string
		body               string
		expectedBody       string
		expectedAttributes map[string]string
		testLogs           plog.Logs
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
			inputLog := plog.NewLogs()
			inputLog.
				ResourceLogs().
				AppendEmpty().
				ScopeLogs().
				AppendEmpty().
				LogRecords().
				AppendEmpty().
				Body().
				SetStr(tc.body)

			rtp := newSourceProcessor(newProcessorCreateSettings(), cfg)

			td, err := rtp.ProcessLogs(context.Background(), inputLog)
			assert.NoError(t, err)

			rss := td.ResourceLogs()
			require.Equal(t, 1, rss.Len())

			sls := rss.At(0).ScopeLogs()
			require.Equal(t, 1, sls.Len())

			logs := sls.At(0).LogRecords()
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

func newProcessorCreateSettings() processor.Settings {
	return processor.Settings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zap.NewNop(),
		},
	}
}
