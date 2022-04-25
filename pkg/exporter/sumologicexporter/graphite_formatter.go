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
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

type graphiteFormatter struct {
	template sourceFormat
	replacer *strings.Replacer
}

const (
	graphiteMetricNamePlaceholder = "_metric_"
)

// newGraphiteFormatter creates new formatter for given SourceFormat template
func newGraphiteFormatter(template string) (graphiteFormatter, error) {
	r, err := regexp.Compile(sourceRegex)
	if err != nil {
		return graphiteFormatter{}, err
	}

	sf := newSourceFormat(r, template)

	replacementChar := `_`
	// replace characters with special meaning in the Graphite transport format
	// Note: \r is technically ok, but it's safer to replace anyway
	replacer := strings.NewReplacer(`.`, replacementChar, ` `, replacementChar, "\n", replacementChar, "\r", replacementChar)
	return graphiteFormatter{
		template: sf,
		replacer: replacer,
	}, nil
}

// escapeGraphiteString replaces dot and space using replacer,
// as dot is special character for graphite format
func (gf *graphiteFormatter) escapeGraphiteString(value string) string {
	return gf.replacer.Replace(value)
}

// format returns metric name basing on template for given fields nas metric name
func (gf *graphiteFormatter) format(f fields, metricName string) string {
	s := gf.template
	labels := make([]interface{}, 0, len(s.matches))

	for _, matchset := range s.matches {
		if matchset == graphiteMetricNamePlaceholder {
			labels = append(labels, gf.escapeGraphiteString(metricName))
		} else {
			attr, ok := f.orig.Get(matchset)
			var value string
			if ok {
				value = attr.AsString()
			} else {
				value = ""
			}
			labels = append(labels, gf.escapeGraphiteString(value))
		}
	}

	return fmt.Sprintf(s.template, labels...)
}

// numberRecord converts NumberDataPoint to graphite metric string
// with additional information from fields
func (gf *graphiteFormatter) numberRecord(fs fields, name string, dataPoint pdata.NumberDataPoint) string {
	switch dataPoint.ValueType() {
	case pdata.MetricValueTypeDouble:
		return fmt.Sprintf("%s %g %d",
			gf.format(fs, name),
			dataPoint.DoubleVal(),
			dataPoint.Timestamp()/pdata.Timestamp(time.Second),
		)
	case pdata.MetricValueTypeInt:
		return fmt.Sprintf("%s %d %d",
			gf.format(fs, name),
			dataPoint.IntVal(),
			dataPoint.Timestamp()/pdata.Timestamp(time.Second),
		)
	}
	return ""
}

// metric2String returns stringified metricPair
func (gf *graphiteFormatter) metric2String(metric pdata.Metric, attributes pdata.Map) string {
	var nextLines []string
	fs := newFields(attributes)
	name := metric.Name()

	switch metric.DataType() {
	case pdata.MetricDataTypeGauge:
		dps := metric.Gauge().DataPoints()
		nextLines = make([]string, 0, dps.Len())
		for i := 0; i < dps.Len(); i++ {
			nextLines = append(nextLines, gf.numberRecord(fs, name, dps.At(i)))
		}
	case pdata.MetricDataTypeSum:
		dps := metric.Sum().DataPoints()
		nextLines = make([]string, 0, dps.Len())
		for i := 0; i < dps.Len(); i++ {
			nextLines = append(nextLines, gf.numberRecord(fs, name, dps.At(i)))
		}
	// Skip complex metrics
	case pdata.MetricDataTypeHistogram:
	case pdata.MetricDataTypeSummary:
	}

	return strings.Join(nextLines, "\n")
}
