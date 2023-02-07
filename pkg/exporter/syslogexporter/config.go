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
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

// Config defines configuration for Syslog exporter.
type Config struct {
	// Syslog server address
	Endpoint string `mapstructure:"endpoint" validate:"required,fqdn"`
	// Syslog server port
	Port int `mapstructure:"port" validate:"required,port"`
	// Protocol for syslog communication
	// options: tcp, udp
	Protocol string `mapstructure:"protocol" validate:"required,protocol type"`
	// CA certificate of syslog server
	CACertificate string `mapstructure:"ca_certificate"`
	// Certificate for mTLS communication
	Certificate string `mapstructure:"certificate"`
	// Key for mTLS communication
	Key string `mapstructure:"key"`
	// Format of syslog messages
	Format string `mapstructure:"format" validate:"required,format"`
	// Flag to control dropping messages in wrong format
	DropInvalidMsg bool `mapstructure:"drop_invalid_messages"`
	//Additional structured data added to structured data in RFC5424
	AdditionalStructuredData []string `mapstructure:"additional_structured_data"`

	exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings `mapstructure:"retry_on_failure"`
}

const (
	// Syslog Protocol
	DefaultProtocol = "tcp"
	// Syslog Port
	DefaultPort = 514
	// Syslog Endpoint
	DefaultEndpoint = "syslog-server.sumologic.net"
	// Syslog format
	DefaultFormat = "any"
	// Drop message if not in the above format
	DropInvalidMessagesDefault = false
)
