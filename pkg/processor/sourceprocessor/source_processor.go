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
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/sourceprocessor/observability"
)

type sourceKeys struct {
	annotationPrefix          string
	namespaceAnnotationPrefix string
	podKey                    string
	podNameKey                string
	podTemplateHashKey        string
}

// dockerLog represents log from k8s using docker log driver send by FluentBit
type dockerLog struct {
	Stream string
	Time   string
	Log    string
}

type sourceProcessor struct {
	logger               *zap.Logger
	collector            string
	sourceCategoryFiller sourceCategoryFiller
	sourceNameFiller     attributeFiller
	sourceHostFiller     attributeFiller

	exclude map[string]*regexp.Regexp
	keys    sourceKeys
}

const (
	alphanums = "bcdfghjklmnpqrstvwxz2456789"

	sourceHostSpecialAnnotation         = "sumologic.com/sourceHost"
	sourceNameSpecialAnnotation         = "sumologic.com/sourceName"
	sourceCategorySpecialAnnotation     = "sumologic.com/sourceCategory"
	sourceCategoryPrefixAnnotation      = "sumologic.com/sourceCategoryPrefix"
	sourceCategoryReplaceDashAnnotation = "sumologic.com/sourceCategoryReplaceDash"

	includeAnnotation = "sumologic.com/include"
	excludeAnnotation = "sumologic.com/exclude"

	collectorKey      = "_collector"
	sourceCategoryKey = "_sourceCategory"
	sourceHostKey     = "_sourceHost"
	sourceNameKey     = "_sourceName"
)

func compileRegex(regex string) *regexp.Regexp {
	if regex == "" {
		return nil
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		log.Fatalf("Cannot compile regular expression: %s Error: %v\n", regex, err)
	}

	return re
}

func newSourceProcessor(set processor.Settings, cfg *Config) *sourceProcessor {
	keys := sourceKeys{
		annotationPrefix:          cfg.AnnotationPrefix,
		namespaceAnnotationPrefix: cfg.NamespaceAnnotationPrefix,
		podKey:                    cfg.PodKey,
		podNameKey:                cfg.PodNameKey,
		podTemplateHashKey:        cfg.PodTemplateHashKey,
	}

	exclude := make(map[string]*regexp.Regexp)
	for field, regexStr := range cfg.Exclude {
		if r := compileRegex(regexStr); r != nil {
			exclude[field] = r
		}
	}

	return &sourceProcessor{
		logger:               set.Logger,
		collector:            cfg.Collector,
		keys:                 keys,
		sourceHostFiller:     createSourceHostFiller(cfg),
		sourceCategoryFiller: newSourceCategoryFiller(cfg, set.Logger),
		sourceNameFiller:     createSourceNameFiller(cfg),
		exclude:              exclude,
	}
}

func (sp *sourceProcessor) fillOtherMeta(atts pcommon.Map) {
	if sp.collector != "" {
		atts.PutStr(collectorKey, sp.collector)
	}
}

func (sp *sourceProcessor) isFilteredOut(atts pcommon.Map) bool {
	// TODO: This is quite inefficient when done for each package (ore even more so, span) separately.
	// It should be moved to K8S Meta Processor and done once per new pod/changed pod

	isFiltered, useAnnotation := sp.isFilteredOutUsingAnnotation(atts, sp.annotationAttribute)
	if useAnnotation {
		return isFiltered
	}

	isFilteredNamespace, useNamespaceAnnotation := sp.isFilteredOutUsingAnnotation(atts, sp.NamespaceAnnotationAttribute)
	if useNamespaceAnnotation {
		return isFilteredNamespace
	}

	// Check fields by matching them against field exclusion regexes
	for field, r := range sp.exclude {
		_, ok := matchFieldByRegex(atts, field, r)
		if ok {
			return true
		}
	}

	return false
}

func (sp *sourceProcessor) isFilteredOutUsingAnnotation(atts pcommon.Map, formatter func(string) string) (bool, bool) {
	useAnnotation := false
	isFiltered := false

	if value, found := atts.Get(formatter(excludeAnnotation)); found {
		if value.Type() == pcommon.ValueTypeStr && value.Str() == "true" {
			useAnnotation = true
			isFiltered = true
		} else if value.Type() == pcommon.ValueTypeBool && value.Bool() {
			useAnnotation = true
			isFiltered = true
		}
	}

	if value, found := atts.Get(formatter(includeAnnotation)); found {
		if value.Type() == pcommon.ValueTypeStr && value.Str() == "true" {
			useAnnotation = true
			isFiltered = false
		} else if value.Type() == pcommon.ValueTypeBool && value.Bool() {
			useAnnotation = true
			isFiltered = false
		}
	}
	return isFiltered, useAnnotation
}

func (sp *sourceProcessor) annotationAttribute(annotationKey string) string {
	return sp.keys.annotationPrefix + annotationKey
}

func (sp *sourceProcessor) NamespaceAnnotationAttribute(annotationKey string) string {
	return sp.keys.namespaceAnnotationPrefix + annotationKey
}

// ProcessTraces processes traces
func (sp *sourceProcessor) ProcessTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	rss := td.ResourceSpans()

	for i := 0; i < rss.Len(); i++ {
		observability.RecordResourceSpansProcessed()

		rs := rss.At(i)
		res := sp.processResource(rs.Resource())
		atts := res.Attributes()

		ss := rs.ScopeSpans()
		totalSpans := 0
		for j := 0; j < ss.Len(); j++ {
			ils := ss.At(j)
			totalSpans += ils.Spans().Len()
		}

		if sp.isFilteredOut(atts) {
			rs.ScopeSpans().RemoveIf(func(ptrace.ScopeSpans) bool { return true })
			observability.RecordFilteredOutN(totalSpans)
		} else {
			observability.RecordFilteredInN(totalSpans)
		}
	}

	return td, nil
}

// ProcessMetrics processes metrics
func (sp *sourceProcessor) ProcessMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	rss := md.ResourceMetrics()

	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		res := sp.processResource(rs.Resource())
		atts := res.Attributes()

		if sp.isFilteredOut(atts) {
			rs.ScopeMetrics().RemoveIf(func(pmetric.ScopeMetrics) bool { return true })
		}
	}

	return md, nil
}

// ProcessLogs processes logs
func (sp *sourceProcessor) ProcessLogs(ctx context.Context, md plog.Logs) (plog.Logs, error) {
	rss := md.ResourceLogs()

	var dockerLog dockerLog

	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		res := sp.processResource(rs.Resource())
		atts := res.Attributes()

		if sp.isFilteredOut(atts) {
			rs.ScopeLogs().RemoveIf(func(plog.ScopeLogs) bool { return true })
		}

		// Due to fluent-bit configuration for sumologic kubernetes collection,
		// logs from kubernetes with docker log driver are send as json with
		// `log`, `stream` and `time` keys.
		// We would like to extract `time` and `stream` to record level attributes
		// and treat `log` as body of the log
		//
		// Related issue: https://github.com/SumoLogic/sumologic-kubernetes-collection/issues/1758
		// ToDo: remove this functionality when the issue is resolved

		sls := rs.ScopeLogs()
		for j := 0; j < sls.Len(); j++ {
			sl := sls.At(j)
			logs := sl.LogRecords()
			for k := 0; k < logs.Len(); k++ {
				log := logs.At(k)
				if log.Body().Type() == pcommon.ValueTypeStr {
					err := json.Unmarshal([]byte(log.Body().Str()), &dockerLog)

					// If there was any parsing error or any of the expected key have no value
					// skip extraction and leave log unchanged
					if err != nil || dockerLog.Stream == "" || dockerLog.Time == "" || dockerLog.Log == "" {
						continue
					}

					// Extract `stream` and `time` as record level attributes
					log.Attributes().PutStr("stream", dockerLog.Stream)
					log.Attributes().PutStr("time", dockerLog.Time)

					// Set log body to `log` content
					log.Body().SetStr(strings.TrimSpace(dockerLog.Log))
				}
			}
		}
	}

	return md, nil
}

// processResource performs multiple actions on resource in the following order:
//   - enrich pod name, so it can be used in templates
//   - set metadata (collector name), so it can be used in templates as well
//   - fills source attributes based on config or annotations
func (sp *sourceProcessor) processResource(res pcommon.Resource) pcommon.Resource {
	atts := res.Attributes()

	sp.enrichPodName(&atts)
	sp.fillOtherMeta(atts)

	sp.sourceHostFiller.fillResourceOrUseAnnotation(&atts,
		sp.annotationAttribute(sourceHostSpecialAnnotation),
		sp.NamespaceAnnotationAttribute(sourceHostSpecialAnnotation),
	)
	sp.sourceCategoryFiller.fill(&atts)

	sp.sourceNameFiller.fillResourceOrUseAnnotation(&atts,
		sp.annotationAttribute(sourceNameSpecialAnnotation),
		sp.NamespaceAnnotationAttribute(sourceNameSpecialAnnotation),
	)

	return res
}

// Start is invoked during service startup.
func (*sourceProcessor) Start(_context context.Context, _host component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (*sourceProcessor) Shutdown(_context context.Context) error {
	return nil
}

// Convert the pod_template_hash to an alphanumeric string using the same logic Kubernetes
// uses at https://github.com/kubernetes/apimachinery/blob/18a5ff3097b4b189511742e39151a153ee16988b/pkg/util/rand/rand.go#L119
func SafeEncodeString(s string) string {
	r := make([]byte, len(s))
	for i, b := range []rune(s) {
		r[i] = alphanums[(int(b) % len(alphanums))]
	}
	return string(r)
}

func (sp *sourceProcessor) enrichPodName(atts *pcommon.Map) {
	// This replicates sanitize_pod_name function
	// Strip out dynamic bits from pod name.
	// NOTE: Kubernetes deployments append a template hash.
	// At the moment this can be in 3 different forms:
	//   1) pre-1.8: numeric in pod_template_hash and pod_parts[-2]
	//   2) 1.8-1.11: numeric in pod_template_hash, hash in pod_parts[-2]
	//   3) post-1.11: hash in pod_template_hash and pod_parts[-2]

	if atts == nil {
		return
	}
	pod, found := atts.Get(sp.keys.podKey)
	if !found {
		return
	}

	podParts := strings.Split(pod.Str(), "-")
	if len(podParts) < 2 {
		// This is unexpected, fallback
		return
	}

	podTemplateHashAttr, found := atts.Get(sp.keys.podTemplateHashKey)

	if found && len(podParts) > 2 {
		podTemplateHash := podTemplateHashAttr.Str()
		if podTemplateHash == podParts[len(podParts)-2] || SafeEncodeString(podTemplateHash) == podParts[len(podParts)-2] {
			atts.PutStr(sp.keys.podNameKey, strings.Join(podParts[:len(podParts)-2], "-"))
			return
		}
	}
	atts.PutStr(sp.keys.podNameKey, strings.Join(podParts[:len(podParts)-1], "-"))
}

// matchFieldByRegex searches the provided attribute map for a particular field
// and matches is with the provided regex.
// It returns the string value of found elements and a boolean flag whether the
// value matched the provided regex.
func matchFieldByRegex(atts pcommon.Map, field string, r *regexp.Regexp) (string, bool) {
	att, ok := atts.Get(field)
	if !ok {
		return "", false
	}

	if att.Type() != pcommon.ValueTypeStr {
		return "", false
	}

	v := att.Str()
	return v, r.MatchString(v)
}
