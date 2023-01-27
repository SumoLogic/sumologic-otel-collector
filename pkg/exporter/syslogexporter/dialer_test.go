package syslogexporter

import (
	"testing"
)

func TestTCPDialer(t *testing.T) {
	done := make(chan string)
	addr, sock, _ := startServer("tcp", "", done)
	defer sock.Close()

	w := Writer{
		priority: LOG_ERR,
		tag:      "tag",
		hostname: "",
		network:  "tcp",
		raddr:    addr,
	}

	_, hostname, err := w.basicDialer()

	if err != nil {
		t.Errorf("failed to dial: %v", err)
	}

	if hostname == "" {
		t.Errorf("should set default hostname")
	}

	w.hostname = "my other hostname"

	_, hostname, err = w.basicDialer()

	if err != nil {
		t.Errorf("failed to dial: %v", err)
	}

	if hostname != "my other hostname" {
		t.Errorf("should not interfere with hostname")
	}
}

func TestUDPDialer(t *testing.T) {
	done := make(chan string)
	addr, sock, _ := startServer("udp", "", done)
	defer sock.Close()

	w := Writer{
		priority: LOG_ERR,
		tag:      "tag",
		hostname: "",
		network:  "udp",
		raddr:    addr,
	}

	_, hostname, err := w.basicDialer()

	if err != nil {
		t.Errorf("failed to dial: %v", err)
	}

	if hostname == "" {
		t.Errorf("should set default hostname")
	}

	w.hostname = "my other hostname"

	_, hostname, err = w.basicDialer()

	if err != nil {
		t.Errorf("failed to dial: %v", err)
	}

	if hostname != "my other hostname" {
		t.Errorf("should not interfere with hostname")
	}
}
