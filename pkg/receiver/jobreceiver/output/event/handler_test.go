package event

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
)

func TestConsume(t *testing.T) {
	var w stubWriter
	h := handler{
		logger: nil,
		writer: &w,
		config: EventConfig{
			IncludeCommandName:   true,
			IncludeCommandStatus: true,
			IncludeDuration:      true,
			MaxBodySize:          5,
		},
	}

	var stdout, stderr bytes.Buffer
	fmt.Fprint(&stdout, "hello world")
	fmt.Fprint(&stderr, "hello world")
	cb := h.Consume(context.Background(), io.NopCloser(&stdout), io.NopCloser(&stderr))
	cb(consumer.ExecutionSummary{
		Command:     "exit",
		RunDuration: time.Millisecond * 500,
		ExitCode:    2,
	})

	if len(w.Out) != 1 {
		t.Fatalf("expected handler to write single entry, got %v", w.Out)
	}
	actualEntry := w.Out[0]
	if actualEntry.Body.(string) != "hello" {
		t.Errorf("expected handler to write single entry with empty string output, got %v", actualEntry.Body)
	}
	if actualEntry.Attributes[commandNameLabel] != "exit" {
		t.Errorf("expected handler to write entry with command.name, got %v", actualEntry.Attributes)
	}
	if actualEntry.Attributes[commandStatusLabel] != 2 {
		t.Errorf("expected handler to write entry with command.status, got %v", actualEntry.Attributes)
	}
	if actualEntry.Attributes[commandDurationLabel] != 0.5 {
		t.Errorf("expected handler to write entry with command.duration, got %v", actualEntry.Attributes)
	}
}

type stubWriter struct {
	Out []*entry.Entry
}

func (s *stubWriter) NewEntry(value interface{}) (*entry.Entry, error) {
	e := entry.New()
	e.Attributes = make(map[string]interface{})
	e.Resource = make(map[string]interface{})
	e.Body = value
	return e, nil
}

func (s *stubWriter) Write(ctx context.Context, e *entry.Entry) error {
	s.Out = append(s.Out, e)
	return nil
}
