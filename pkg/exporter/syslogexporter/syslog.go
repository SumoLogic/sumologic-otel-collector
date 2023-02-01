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
	"log/syslog"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

const severityMask = 0x07
const facilityMask = 0xf8

const emptyTag = "-"
const defaultPriority = syslog.LOG_INFO

const formatRFC5424 = "RFC5424"
const formatRFC3164 = "RFC3164"
const formatAny = "any"

const regexpRFC5424 = "^\\<(?<priority>\\d+)\\>\\d+ (?<timestamp>\\S+) (?<hostname>\\S+) (?<app>\\S+) (?<pid>\\S+) (?<msgid>\\S+) (?<structured_data>\\-|\\[.*\\]) (?<message>.*)"
const regexpRFC3164 = "^\\<(?<priority>\\d+)\\>(?<timestamp>\\w+\\s+\\d+\\s+\\d+:\\d+:\\S+) (?<hostname>\\S+) (?<other_fields>.*)"

type Syslog struct {
	hostname       string
	network        string
	addr           string
	format         string
	app            string
	dropInvalidMsg bool
	pid            int
	tlsConfig      *tls.Config
	formatRegexp   *regexp.Regexp

	mu   sync.Mutex
	conn net.Conn
}

func Connect(cfg *Config, tlsConfig *tls.Config, hostname string, pid int, app string) (*Syslog, error) {
	s := &Syslog{
		hostname:       hostname,
		network:        cfg.Protocol,
		addr:           fmt.Sprintf("%s:%d", cfg.Endpoint, cfg.Port),
		format:         cfg.Format,
		tlsConfig:      tlsConfig,
		pid:            pid,
		app:            app,
		dropInvalidMsg: cfg.DropInvalidMsg,
	}

	s.setFormatRegexp()

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
		return fmt.Errorf("Invalid message format, expected format %s, message: %s", s.format, msg)
	} else if !isFormatCorrect {
		msg = s.formatMsg(msg)
	}

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

func (s *Syslog) setFormatRegexp() {
	r := ""
	switch s.format {
	case formatRFC3164:
		r = regexpRFC3164
	case formatRFC5424:
		r = regexpRFC3164
	}
	s.formatRegexp = regexp.MustCompile(r)
}

func (s Syslog) validateFormat(msg string) bool {
	switch s.format {
	case formatAny:
		return true
	case formatRFC3164:
		return s.isRFC3164(msg)
	case formatRFC5424:
		return s.isRFC5424(msg)
	default:
		return false
	}
}

func (s Syslog) formatMsg(msg string) string {
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

func (s Syslog) formatRFC3164(msg string) string {
	timestamp := time.Now().Format(time.Stamp)
	return fmt.Sprintf("<%d>%s %s %s", defaultPriority, timestamp, s.hostname, msg)
}

func (s Syslog) formatRFC5424(msg string) string {
	timestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf("<%d>%d %s %s %s %d %s - %s", defaultPriority, 1, timestamp, s.hostname, s.app, s.pid, emptyTag, msg)
}

// Example messages RFC5424 compliant
// <34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - BOM'su root' failed for lonvick on /dev/pts/8
// <165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.
// <165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] BOMAn application event log entry...
// <165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"][examplePriority@32473 class="high"]
// for more information see: https://www.rfc-editor.org/rfc/rfc5424
func (s Syslog) isRFC5424(msg string) bool {
	return s.formatRegexp.MatchString(msg)
}

// Example messages RFC3164 compliant
// <34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8
// <13>Feb  5 17:32:18 10.0.0.99 Use the BFG!
// for more information see: https://www.ietf.org/rfc/rfc3164.txt
func (s Syslog) isRFC3164(msg string) bool {
	return s.formatRegexp.MatchString(msg)
}
