package syslogexporter

import (
	"crypto/tls"
	"fmt"
	"log/syslog"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const severityMask = 0x07
const facilityMask = 0xf8

type Syslog struct {
	priority  syslog.Priority
	tag       string
	hostname  string
	network   string
	addr      string
	tlsConfig *tls.Config

	mu   sync.Mutex
	conn net.Conn
}

func Connect(network, addr string, priority syslog.Priority, tag string, tlsConfig *tls.Config) (*Syslog, error) {
	if priority < 0 || priority > syslog.LOG_LOCAL7|syslog.LOG_DEBUG {
		return nil, fmt.Errorf("invalid priority")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	s := &Syslog{
		priority:  priority,
		tag:       tag,
		hostname:  hostname,
		network:   network,
		addr:      addr,
		tlsConfig: tlsConfig,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	err = s.connect()
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

func (s *Syslog) WriteAndRetry(p syslog.Priority, msg string) error {
	priority := (s.priority & facilityMask) | (p & severityMask)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		if err := s.write(priority, msg); err == nil {
			return nil
		}
	}
	if err := s.connect(); err != nil {
		return err
	}
	return s.write(priority, msg)
}

func (w *Syslog) write(p syslog.Priority, msg string) error {
	// check if logs contains new line character at the end, if not add it
	if !strings.HasSuffix(msg, "\n") {
		msg = fmt.Sprintf("%s%s", msg, "\n")
	}

	timestamp := time.Now().Format(time.RFC3339)
	_, err := fmt.Fprintf(w.conn, "<%d>%s %s %s[%d]: %s%s", p, timestamp, w.hostname, w.tag, os.Getpid(), msg)
	return err
}
