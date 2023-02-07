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
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const emptyStructuredData = "-"
const defaultPriority = 165
const emptyMessageID = "-"
const versionRFC5424 = 1

const formatRFC5424 = "RFC5424"
const formatRFC3164 = "RFC3164"
const formatAny = "any"

const priority = "priority"
const version = "version"
const timestamp = "timestamp"
const hostname = "hostname"
const app = "app"
const pid = "pid"
const msgid = "msgid"
const structuredData = "structured_data"
const message = "message"

var regexpRFC5424 = regexp.MustCompile(`^\<(?P<priority>\d+)\>(?P<version>\d+) (?P<timestamp>\S+) (?P<hostname>\S+) (?P<app>\S+) (?P<pid>\S+) (?P<msgid>\S+) (?P<structured_data>\-|\[.*\]) (?P<message>.*)`)
var regexpRFC3164 = regexp.MustCompile(`^\<(?P<priority>\d+)\>(?P<timestamp>\w+\s+\d+\s+\d+:\d+:\S+) (?P<hostname>\S+) (?P<other_fields>.*)`)

type Syslog struct {
	hostname                 string
	network                  string
	addr                     string
	format                   string
	app                      string
	dropInvalidMsg           bool
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
		dropInvalidMsg:           cfg.DropInvalidMsg,
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

func (s *Syslog) Write(msg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	isFormatCorrect := s.validateFormat(msg)

	if !isFormatCorrect && s.dropInvalidMsg {
		s.logger.Debug("Invalid message format",
			zap.String("format", s.format),
			zap.String("msg", msg),
		)
		return nil
	} else if !isFormatCorrect {
		msg = s.formatMsg(msg)
	}

	msg = s.addStructuredData(msg)

	if s.conn != nil {
		if err := s.write(msg); err == nil {
			return nil
		}
	}
	if err := s.connect(); err != nil {
		return err
	}
	return s.write(msg)
}

func (s *Syslog) write(msg string) error {
	// check if logs contains new line character at the end, if not add it
	if !strings.HasSuffix(msg, "\n") {
		msg = fmt.Sprintf("%s%s", msg, "\n")
	}
	_, err := fmt.Fprint(s.conn, msg)
	return err
}

func (s *Syslog) addStructuredData(msg string) string {
	c := getMessageComponents(regexpRFC5424, msg)
	if len(s.additionalStructuredData) == 0 || len(c) == 0 {
		return msg
	}
	var sd []string
	if len(c[structuredData]) == 1 && c[structuredData] == emptyStructuredData {
		sd = s.additionalStructuredData
	} else {
		c[structuredData] = strings.ReplaceAll(c[structuredData], "[", "")
		c[structuredData] = strings.ReplaceAll(c[structuredData], "]", "")
		sd = append(s.additionalStructuredData, strings.Split(c[structuredData], " ")...)
	}
	structuredData := fmt.Sprintf("[%s]", strings.Join(sd, " "))

	return fmt.Sprintf("<%s>%s %s %s %s %s %s %s %s",
		c[priority], c[version], c[timestamp], c[hostname],
		c[app], c[pid], c[msgid], structuredData, c[message])
}

func (s *Syslog) validateFormat(msg string) bool {
	switch s.format {
	case formatAny:
		return true
	case formatRFC3164:
		return isRFC3164(msg)
	case formatRFC5424:
		return isRFC5424(msg)
	default:
		return false
	}
}

func (s *Syslog) formatMsg(msg string) string {
	switch s.format {
	case formatAny:
		return msg
	case formatRFC3164:
		return s.formatRFC3164(msg)
	case formatRFC5424:
		return s.formatRFC5424(msg)
	default:
		return ""
	}
}

func (s *Syslog) formatRFC3164(msg string) string {
	timestamp := time.Now().Format(time.Stamp)
	return fmt.Sprintf("<%d>%s %s %s", defaultPriority, timestamp, s.hostname, msg)
}

func (s *Syslog) formatRFC5424(msg string) string {
	timestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("<%d>%d %s %s %s %d %s %s %s", defaultPriority, versionRFC5424, timestamp, s.hostname, s.app, s.pid, emptyMessageID, emptyStructuredData, msg)
}

// Example messages RFC5424 compliant
// <34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - BOM'su root' failed for lonvick on /dev/pts/8
// <165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.
// <165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] BOMAn application event log entry...
// <165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"][examplePriority@32473 class="high"]
// for more information see: https://www.rfc-editor.org/rfc/rfc5424
func isRFC5424(msg string) bool {
	return regexpRFC5424.MatchString(msg)
}

// Example messages RFC3164 compliant
// <34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8
// <13>Feb  5 17:32:18 10.0.0.99 Use the BFG!
// for more information see: https://www.ietf.org/rfc/rfc3164.txt
func isRFC3164(msg string) bool {
	return regexpRFC3164.MatchString(msg)
}

func getMessageComponents(r *regexp.Regexp, msg string) map[string]string {
	match := r.FindStringSubmatch(msg)

	components := make(map[string]string)
	for i, name := range r.SubexpNames() {
		if i > 0 && i <= len(match) {
			components[name] = match[i]
		}
	}
	return components
}
