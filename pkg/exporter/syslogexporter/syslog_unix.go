package syslogexporter

import (
	"errors"
	"io"
	"net"
)

func unixSyslog() (conn serverConn, err error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog", "/var/run/log"}
	for _, network := range logTypes {
		for _, path := range logPaths {
			conn, err := net.Dial(network, path)
			if err != nil {
				continue
			} else {
				return &localConn{conn: conn}, nil
			}
		}
	}
	return nil, errors.New("Unix syslog delivery error")
}

// localConn adheres to the serverConn interface, allowing us to send syslog
// messages to the local syslog daemon over a Unix domain socket.
type localConn struct {
	conn io.WriteCloser
}

// writeString formats syslog messages using time.Stamp instead of time.RFC3339,
// and omits the hostname (because it is expected to be used locally).
func (n *localConn) writeString(framer Framer, formatter Formatter, p Priority, hostname, tag, msg string) error {
	if framer == nil {
		framer = DefaultFramer
	}
	if formatter == nil {
		formatter = UnixFormatter
	}
	_, err := n.conn.Write([]byte(framer(formatter(p, hostname, tag, msg))))
	return err
}

// close the (local) network connection
func (n *localConn) close() error {
	return n.conn.Close()
}
