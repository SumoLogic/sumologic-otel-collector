// Copyright 2023, OpenTelemetry Authors
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

package syslogexporter

import (
	"errors"
	"strings"

	"github.com/THREATINT/go-net"

	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

var (
	unsupportedPort     = errors.New("Unsupported port: port is required, must be in the range 1-65535")
	invalidEndpoint     = errors.New("Invalid endpoint: endpoint is required, must be a valid FQDN or IP address")
	unsupportedProtocol = errors.New("Unsupported protocol: protocol is required, only tcp/udp supported")
	unsupportedFormat   = errors.New("Unsupported format: Only rfc5424 and rfc3164 supported")
)

// Config defines configuration for Syslog exporter.
type Config struct {
	// Syslog server address
	Endpoint string `mapstructure:"endpoint"`
	// Syslog server port
	Port int `mapstructure:"port"`
	// Protocol for syslog communication
	// options: tcp, udp
	Protocol string `mapstructure:"protocol"`
	// CA certificate of syslog server
	CACertificate string `mapstructure:"ca_certificate"`
	// Format of syslog messages
	Format string `mapstructure:"format"`
	// Certificate for mTLS communication (client certificate)
	Certificate string `mapstructure:"certificate"`
	// Key for mTLS communication (client key)
	Key string `mapstructure:"key"`

	exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings `mapstructure:"retry_on_failure"`
}

// Validate the configuration for errors. This is required by component.Config.
func (cfg *Config) Validate() error {
	if cfg.Port < 1 || cfg.Port > 65525 {
		return unsupportedPort
	}

	if !net.IsFQDN(cfg.Endpoint) && !net.IsIPAddr(cfg.Endpoint) && cfg.Endpoint != "localhost" {
		return invalidEndpoint
	}

	if strings.ToLower(cfg.Protocol) != "tcp" && strings.ToLower(cfg.Protocol) != "udp" {
		return unsupportedProtocol
	}

	switch cfg.Format {
	case formatRFC3164Str:
	case formatRFC5424Str:
	default:
		return unsupportedFormat
	}

	return nil
}

const (
	// Syslog Protocol
	DefaultProtocol = "tcp"
	// Syslog Port
	DefaultPort = 514
	// Syslog Endpoint
	DefaultEndpoint = "host.domain.com"
	// Syslog format
	DefaultFormat = "rfc5424"
)
