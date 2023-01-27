package syslogexporter

import (
	"log"
	"os"
)

type serverConn interface {
	writeString(framer Framer, formatter Formatter, p Priority, hostname, tag, s string) error
	close() error
}

func New(priority Priority, tag string) (w *Writer, err error) {
	return Dial("", "", priority, tag)
}

func Dial(network, raddr string, priority Priority, tag string) (*Writer, error) {
	return DialWithTLSConfig(network, raddr, priority, tag)
}

func DialWithTLSConfig(network, raddr string, priority Priority, tag string) (*Writer, error) {
	return dialAllParameters(network, raddr, priority, tag)
}

// implementation of the various functions above
func dialAllParameters(network, raddr string, priority Priority, tag string) (*Writer, error) {
	if err := validatePriority(priority); err != nil {
		return nil, err
	}

	if tag == "" {
		tag = os.Args[0]
	}
	hostname, _ := os.Hostname()

	w := &Writer{
		priority: priority,
		tag:      tag,
		hostname: hostname,
		network:  network,
		raddr:    raddr,
	}

	_, err := w.connect()
	if err != nil {
		return nil, err
	}
	return w, err
}

func NewLogger(p Priority, logFlag int) (*log.Logger, error) {
	s, err := New(p, "")
	if err != nil {
		return nil, err
	}
	return log.New(s, "", logFlag), nil
}
