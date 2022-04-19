// Copyright 2021, OpenTelemetry Authors
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

package sumologicexporter

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// carbon2TagString returns all attributes as space spearated key=value pairs.
// In addition, metric name and unit are also included.
// In case `metric` or `unit` attributes has been set too, they are prefixed
// with underscore `_` to avoid overwriting the metric name and unit.
func carbon2TagString(metric pmetric.Metric, attributes pcommon.Map) string {
	length := attributes.Len()

	if _, ok := attributes.Get("metric"); ok {
		length++
	}

	if _, ok := attributes.Get("unit"); ok && len(metric.Unit()) > 0 {
		length++
	}

	returnValue := make([]string, 0, length)
	attributes.Range(func(k string, v pcommon.Value) bool {
		if k == "name" || k == "unit" {
			k = fmt.Sprintf("_%s", k)
		}
		returnValue = append(returnValue, fmt.Sprintf(
			"%s=%s",
			sanitizeCarbonString(k),
			sanitizeCarbonString(v.AsString()),
		))
		return true
	})

	returnValue = append(returnValue, fmt.Sprintf("metric=%s", sanitizeCarbonString(metric.Name())))

	if len(metric.Unit()) > 0 {
		returnValue = append(returnValue, fmt.Sprintf("unit=%s", sanitizeCarbonString(metric.Unit())))
	}

	return strings.Join(returnValue, " ")
}

// sanitizeCarbonString replaces problematic characters with underscore
func sanitizeCarbonString(text string) string {
	return strings.NewReplacer(" ", "_", "=", ":", "\n", "_", "\r", "_").Replace(text)
}

// carbon2NumberRecord converts NumberDataPoint to carbon2 metric string
// with additional information from metricPair.
func carbon2NumberRecord(metric pmetric.Metric, attributes pcommon.Map, dataPoint pmetric.NumberDataPoint) string {
	switch dataPoint.ValueType() {
	case pmetric.MetricValueTypeDouble:
		return fmt.Sprintf("%s  %g %d",
			carbon2TagString(metric, attributes),
			dataPoint.DoubleVal(),
			dataPoint.Timestamp()/1e9,
		)
	case pmetric.MetricValueTypeInt:
		return fmt.Sprintf("%s  %d %d",
			carbon2TagString(metric, attributes),
			dataPoint.IntVal(),
			dataPoint.Timestamp()/1e9,
		)
	}
	return ""
}

// carbon2metric2String converts metric to Carbon2 formatted string.
func carbon2Metric2String(metric pmetric.Metric, attributes pcommon.Map) string {
	var nextLines []string

	switch metric.DataType() {
	case pmetric.MetricDataTypeGauge:
		dps := metric.Gauge().DataPoints()
		nextLines = make([]string, 0, dps.Len())
		for i := 0; i < dps.Len(); i++ {
			nextLines = append(nextLines, carbon2NumberRecord(metric, attributes, dps.At(i)))
		}
	case pmetric.MetricDataTypeSum:
		dps := metric.Sum().DataPoints()
		nextLines = make([]string, 0, dps.Len())
		for i := 0; i < dps.Len(); i++ {
			nextLines = append(nextLines, carbon2NumberRecord(metric, attributes, dps.At(i)))
		}
	// Skip complex metrics
	case pmetric.MetricDataTypeHistogram:
	case pmetric.MetricDataTypeSummary:
	}

	return strings.Join(nextLines, "\n")
}
