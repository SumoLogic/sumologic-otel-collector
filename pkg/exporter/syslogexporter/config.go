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

import "go.opentelemetry.io/collector/exporter/exporterhelper"

// Config defines configuration for Syslog exporter.
type Config struct {
	// Syslog server address
	Endpoint string `mapstructure:"endpoint"`
	// Syslog server port
	Port int `mapstructure:"port"`
	// Protocol for syslog communication
	// options: tcp, udp, tcp+tls, tcp+mtls
	Protocol                     string `mapstructure:"protocol"`
	exporterhelper.QueueSettings `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings `mapstructure:"retry_on_failure"`
}
