package syslogexporter

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRFC3164Formatter(t *testing.T) {
	out := RFC3164Formatter(LOG_ERR, "hostname", "tag", "content")
	expected := fmt.Sprintf("<%d>%s %s %s[%d]: %s",
		LOG_ERR, time.Now().Format(time.Stamp), "hostname", "tag", os.Getpid(), "content")
	if out != expected {
		t.Errorf("expected %v got %v", expected, out)
	}
}

func TestRFC5424Formatter(t *testing.T) {
	out := RFC5424Formatter(LOG_ERR, "hostname", "tag", "content")
	expected := fmt.Sprintf("<%d>%d %s %s %s %d %s - %s",
		LOG_ERR, 1, time.Now().Format(time.RFC3339), "hostname", truncateStartStr(os.Args[0], appNameMaxLength),
		os.Getpid(), "tag", "content")
	if out != expected {
		t.Errorf("expected %v got %v", expected, out)
	}
}
