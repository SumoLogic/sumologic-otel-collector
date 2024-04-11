// Copyright 2020, OpenTelemetry Authors
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
	"errors"
	"fmt"
	"net/url"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configcompression"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension"
)

// Config defines configuration for Sumo Logic exporter.
type Config struct {
	confighttp.ClientConfig      `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
	exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	configretry.BackOffConfig    `mapstructure:"retry_on_failure"`

	// Compression encoding format, either empty string, gzip or deflate (default gzip)
	// Empty string means no compression
	// NOTE: CompressEncoding is deprecated and will be removed in an upcoming release
	CompressEncoding configcompression.Type `mapstructure:"compress_encoding"`
	// Max HTTP request body size in bytes before compression (if applied).
	// By default 1MB is recommended.
	MaxRequestBodySize int `mapstructure:"max_request_body_size"`

	// Logs related configuration
	// Format to post logs into Sumo. (default json)
	//   * text - Logs will appear in Sumo Logic in text format.
	//   * json - Logs will appear in Sumo Logic in json format.
	//   * otlp - Logs will be send in otlp format and will appear in Sumo Logic in text format.
	LogFormat LogFormatType `mapstructure:"log_format"`

	// Metrics related configuration
	// The format of metrics you will be sending, either otlp or prometheus (Default is otlp)
	MetricFormat MetricFormatType `mapstructure:"metric_format"`

	// Decompose OTLP Histograms into individual metrics, similar to how they're represented in Prometheus format
	DecomposeOtlpHistograms bool `mapstructure:"decompose_otlp_histograms"`

	// Traces related configuration
	// The format of traces you will be sending, currently only otlp format is supported
	TraceFormat TraceFormatType `mapstructure:"trace_format"`

	// Sumo specific options
	// Name of the client
	Client string `mapstructure:"client"`

	// StickySessionEnabled defines if sticky session support is enable.
	// By default this is false.
	StickySessionEnabled bool `mapstructure:"sticky_session_enabled"`
}

// CreateDefaultClientConfig returns default http client settings
func CreateDefaultClientConfig() confighttp.ClientConfig {
	return confighttp.ClientConfig{
		Timeout:     defaultTimeout,
		Compression: DefaultCompressEncoding,
		Auth: &configauth.Authentication{
			AuthenticatorID: component.NewID(sumologicextension.Type),
		},
	}
}

func (cfg *Config) Validate() error {

	switch cfg.CompressEncoding {
	case configcompression.TypeGzip:
	case configcompression.TypeDeflate:
	case NoCompression:

	default:
		return fmt.Errorf("invalid compression encoding type: %v", cfg.ClientConfig.Compression)
	}

	switch cfg.ClientConfig.Compression {
	case configcompression.TypeGzip:
	case configcompression.TypeDeflate:
	case configcompression.TypeZstd:
	case NoCompression:

	default:
		return fmt.Errorf("invalid compression encoding type: %v", cfg.ClientConfig.Compression)
	}

	if cfg.CompressEncoding != NoCompression && cfg.ClientConfig.Compression != DefaultCompressEncoding {
		return fmt.Errorf("compress_encoding is deprecated and should not be used when compression is set to a non-default value")
	}

	switch cfg.LogFormat {
	case OTLPLogFormat:
	case JSONFormat:
	case TextFormat:
	default:
		return fmt.Errorf("unexpected log format: %s", cfg.LogFormat)
	}

	switch cfg.MetricFormat {
	case OTLPMetricFormat:
	case PrometheusFormat:
	case RemovedGraphiteFormat:
		return fmt.Errorf("support for the graphite metric format was removed, please use prometheus or otlp instead")
	case RemovedCarbon2Format:
		return fmt.Errorf("support for the carbon2 metric format was removed, please use prometheus or otlp instead")
	default:
		return fmt.Errorf("unexpected metric format: %s", cfg.MetricFormat)
	}

	switch cfg.TraceFormat {
	case OTLPTraceFormat:
	default:
		return fmt.Errorf("unexpected trace format: %s", cfg.TraceFormat)
	}

	if len(cfg.ClientConfig.Endpoint) == 0 && cfg.ClientConfig.Auth == nil {
		return errors.New("no endpoint and no auth extension specified")
	}

	if _, err := url.Parse(cfg.ClientConfig.Endpoint); err != nil {
		return fmt.Errorf("failed parsing endpoint URL: %s; err: %w",
			cfg.ClientConfig.Endpoint, err,
		)
	}

	if err := cfg.QueueSettings.Validate(); err != nil {
		return fmt.Errorf("queue settings has invalid configuration: %w", err)
	}

	return nil
}

// LogFormatType represents log_format
type LogFormatType string

// MetricFormatType represents metric_format
type MetricFormatType string

// TraceFormatType represents trace_format
type TraceFormatType string

// PipelineType represents type of the pipeline
type PipelineType string

const (
	// TextFormat represents log_format: text
	TextFormat LogFormatType = "text"
	// JSONFormat represents log_format: json
	JSONFormat LogFormatType = "json"
	// OTLPLogFormat represents log_format: otlp
	OTLPLogFormat LogFormatType = "otlp"
	// RemovedGraphiteFormat represents the no longer supported graphite metric format
	RemovedGraphiteFormat MetricFormatType = "graphite"
	// RemovedCarbon2Format represents the no longer supported carbon2 metric format
	RemovedCarbon2Format MetricFormatType = "carbon2"
	// PrometheusFormat represents metric_format: prometheus
	PrometheusFormat MetricFormatType = "prometheus"
	// OTLPMetricFormat represents metric_format: otlp
	OTLPMetricFormat MetricFormatType = "otlp"
	// OTLPTraceFormat represents trace_format: otlp
	OTLPTraceFormat TraceFormatType = "otlp"
	// NoCompression represents disabled compression
	NoCompression configcompression.Type = ""
	// MetricsPipeline represents metrics pipeline
	MetricsPipeline PipelineType = "metrics"
	// LogsPipeline represents metrics pipeline
	LogsPipeline PipelineType = "logs"
	// TracesPipeline represents traces pipeline
	TracesPipeline PipelineType = "traces"
	// defaultTimeout
	defaultTimeout time.Duration = 30 * time.Second
	// DefaultCompress defines default Compress
	DefaultCompress bool = true
	// DefaultCompressEncoding defines default CompressEncoding
	DefaultCompressEncoding configcompression.Type = "gzip"
	// DefaultMaxRequestBodySize defines default MaxRequestBodySize in bytes
	DefaultMaxRequestBodySize int = 1 * 1024 * 1024
	// DefaultLogFormat defines default LogFormat
	DefaultLogFormat LogFormatType = OTLPLogFormat
	// DefaultMetricFormat defines default MetricFormat
	DefaultMetricFormat MetricFormatType = OTLPMetricFormat
	// DefaultClient defines default Client
	DefaultClient string = "otelcol"
	// DefaultLogKey defines default LogKey value
	DefaultLogKey string = "log"
	// DefaultDropRoutingAttribute defines default DropRoutingAttribute
	DefaultDropRoutingAttribute string = ""
	// DefaultStickySessionEnabled defines default StickySessionEnabled value
	DefaultStickySessionEnabled bool = false
)
