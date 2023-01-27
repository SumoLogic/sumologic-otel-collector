package syslogexporter

import (
	"net"
)

type dialerFunctionWrapper struct {
	Name   string
	Dialer func() (serverConn, string, error)
}

func (df dialerFunctionWrapper) Call() (serverConn, string, error) {
	return df.Dialer()
}

func (w *Writer) getDialer() dialerFunctionWrapper {
	dialers := map[string]dialerFunctionWrapper{
		"": dialerFunctionWrapper{"unixDialer", w.unixDialer},
	}
	dialer, ok := dialers[w.network]
	if !ok {
		dialer = dialerFunctionWrapper{"basicDialer", w.basicDialer}
	}
	return dialer
}

func (w *Writer) unixDialer() (serverConn, string, error) {
	sc, err := unixSyslog()
	hostname := w.hostname
	if hostname == "" {
		hostname = "localhost"
	}
	return sc, hostname, err
}

func (w *Writer) basicDialer() (serverConn, string, error) {
	c, err := net.Dial(w.network, w.raddr)
	var sc serverConn
	hostname := w.hostname
	if err == nil {
		sc = &netConn{conn: c}
		if hostname == "" {
			hostname = c.LocalAddr().String()
		}
	}
	return sc, hostname, err
}
