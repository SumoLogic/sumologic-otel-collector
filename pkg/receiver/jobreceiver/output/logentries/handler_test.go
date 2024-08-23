package logentries

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConsumeWithMaxLogSize(t *testing.T) {
	var w stubWriter
	cfg := LogEntriesConfig{
		IncludeCommandName: true,
		IncludeStreamName:  true,
		MaxLogSize:         128,
	}
	h, err := cfg.Build(zap.NewNop().Sugar(), &w)
	if err != nil {
		t.Fatal(err)
	}

	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()

	writeStdout := func(w io.WriteCloser) {
		fmt.Fprint(w, strings.Repeat("a", 64))
		fmt.Fprint(w, strings.Repeat("a", 64))
		for i := 0; i < 64; i = i + 8 {
			fmt.Fprint(w, strings.Repeat("b", 8))
		}
		fmt.Fprint(w, "\n")
		w.Close()
	}

	writeStderr := func(w io.WriteCloser) {
		fmt.Fprint(w, "hello world")
		w.Close()
	}
	h.Consume(context.Background(), stdoutR, stderrR)

	go writeStdout(stdoutW)
	go writeStderr(stderrW)

	done := make(chan struct{})
	go func() {
		close(done)
	}()

	require.Eventually(t,
		func() bool {
			w.MU.Lock()
			defer w.MU.Unlock()
			return len(w.Out) == 3
		},
		time.Second,
		time.Millisecond*100,
		"expected three log entries out",
	)
	for _, ent := range w.Out {
		body, ok := ent.Body.(string)
		require.True(t, ok)
		assert.LessOrEqual(t, len(body), 128)
	}
}

type stubWriter struct {
	MU  sync.Mutex
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
	s.MU.Lock()
	s.Out = append(s.Out, e)
	s.MU.Unlock()
	return nil
}
