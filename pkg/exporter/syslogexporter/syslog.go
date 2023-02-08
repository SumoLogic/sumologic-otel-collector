// Copyright 2023 OpenTelemetry Authors
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

package syslogexporter

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"

	"go.uber.org/zap"
)

const defaultPriority = 165
const defaultFacility = 1
const versionRFC5424 = 1

const formatRFC5424Str = "RFC5424"
const formatRFC3164Str = "RFC3164"

const priority = "priority"
const facility = "facility"
const version = "version"
const timestamp = "timestamp"
const hostname = "hostname"
const app = "appname"
const pid = "proc_id"
const msgId = "msg_id"
const structuredData = "structured_data"
const message = "message"

type Syslog struct {
	hostname                 string
	network                  string
	addr                     string
	format                   string
	app                      string
	pid                      int
	additionalStructuredData []string
	tlsConfig                *tls.Config
	logger                   *zap.Logger
	mu                       sync.Mutex
	conn                     net.Conn
}

func Connect(logger *zap.Logger, cfg *Config, tlsConfig *tls.Config, hostname string, pid int, app string) (*Syslog, error) {
	s := &Syslog{
		logger:                   logger,
		hostname:                 hostname,
		network:                  cfg.Protocol,
		addr:                     fmt.Sprintf("%s:%d", cfg.Endpoint, cfg.Port),
		format:                   cfg.Format,
		tlsConfig:                tlsConfig,
		pid:                      pid,
		app:                      app,
		additionalStructuredData: cfg.AdditionalStructuredData,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.connect()
	if err != nil {
		return nil, err
	}
	return s, err
}

func (s *Syslog) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		return err
	}
	return nil
}

func (s *Syslog) connect() error {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	var err error
	if s.tlsConfig != nil {
		s.conn, err = tls.Dial("tcp", s.addr, s.tlsConfig)
	} else {
		s.conn, err = net.Dial(s.network, s.addr)
	}
	return err
}

func (s *Syslog) Write(msg map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.addStructuredData(msg)
	msgStr := s.formatMsg(msg)

	if s.conn != nil {
		if err := s.write(msgStr); err == nil {
			return nil
		}
	}
	if err := s.connect(); err != nil {
		return err
	}

	return s.write(msgStr)
}

func (s *Syslog) write(msg string) error {
	// check if logs contains new line character at the end, if not add it
	if !strings.HasSuffix(msg, "\n") {
		msg = fmt.Sprintf("%s%s", msg, "\n")
	}
	_, err := fmt.Fprint(s.conn, msg)
	return err
}

func (s *Syslog) formatMsg(msg map[string]any) string {
	switch s.format {
	case formatRFC3164Str:
		return formatRFC3164(msg)
	case formatRFC5424Str:
		return formatRFC5424(msg)
	default:
		panic(fmt.Sprintf("unsupported syslog format, format: %s", s.format))
	}
}

func (s *Syslog) addStructuredData(msg map[string]any) {
	if len(s.additionalStructuredData) == 0 {
		return
	}
	_, ok := msg[structuredData]
	if !ok {
		msg[structuredData] = s.additionalStructuredData
	}
}

func populateDefaults(msg map[string]any, msgProperty string) {
	const emptyValue = "-"
	msgValue, ok := msg[msgProperty]
	if !ok && msgProperty == priority {
		msg[msgProperty] = defaultPriority
		return
	}
	if !ok && msgProperty == version {
		msg[msgProperty] = versionRFC5424
		return
	}
	if !ok && msgProperty == facility {
		msg[msgProperty] = defaultFacility
		return
	}
	if !ok {
		msg[msgProperty] = emptyValue
		return
	}
	msg[msgProperty] = msgValue
}

func formatRFC3164(msg map[string]any) string {
	msgProperties := []string{priority, hostname, message}
	for _, msgProperty := range msgProperties {
		populateDefaults(msg, msgProperty)
	}
	return fmt.Sprintf("<%d>%s %s %s", msg[priority], msg[timestamp], msg[hostname], msg[message])
}

func formatRFC5424(msg map[string]any) string {
	msgProperties := []string{priority, version, hostname, app, pid, msgId, message, structuredData}
	for _, msgProperty := range msgProperties {
		populateDefaults(msg, msgProperty)
	}
	return fmt.Sprintf("<%d>%d %s %s %s %s %s %s %s", msg[priority], msg[version], msg[timestamp], msg[hostname], msg[app], msg[pid], msg[msgId], msg[structuredData], msg[message])
}
