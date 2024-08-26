package event

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"go.uber.org/zap"
)

var errSizeLimitExceeded = errors.New("buffer size limit exceeded")

type handler struct {
	logger *zap.SugaredLogger
	writer consumer.WriterOp

	config EventConfig
}

var _ consumer.Interface = (*handler)(nil)

// Consume
func (h *handler) Consume(ctx context.Context, stdout, stderr io.Reader) consumer.CloseFunc {
	e := eventOutputBuffer{
		ctx:         ctx,
		logger:      h.logger,
		writer:      h.writer,
		EventConfig: h.config,
	}
	e.Start(stdout, stderr)
	return e.Close
}

type eventOutputBuffer struct {
	EventConfig

	ctx    context.Context
	logger *zap.SugaredLogger
	writer consumer.WriterOp

	wg  sync.WaitGroup
	mu  sync.Mutex
	buf bytes.Buffer
}

// Start consuming the intput streams into the buffer
func (b *eventOutputBuffer) Start(stdout, stderr io.Reader) {
	b.wg.Add(2)
	go b.consume(stdout)
	go b.consume(stderr)
}

// consume reads the input into the buffer until either EOF is reached or the
// buffer is full. Once the buffer is full, consume discards remaining input
// until EOF or read error
func (b *eventOutputBuffer) consume(in io.Reader) {
	defer b.wg.Done()

	_, err := io.Copy(b, in)
	if errors.Is(err, errSizeLimitExceeded) {
		_, err = io.Copy(io.Discard, in)
	}
	// os.ErrClosed likely when an OS Pipe is closed due to a problem executing
	// a command.
	if errors.Is(err, os.ErrClosed) {
		b.logger.Infof("input from closed file: %s", err)
	} else if err != nil {
		b.logger.Errorf("io error consuming event input: %s", err)
	}
}

// Close builds a new log entry based off the exeuction summary and contents
// of the buffer. Writes the entry to the pipeline.
func (b *eventOutputBuffer) Close(summary consumer.ExecutionSummary) {
	// Wait for all content to be written to the buffer
	b.wg.Wait()

	ent, err := b.writer.NewEntry(b.String())
	if err != nil {
		b.logger.Errorf("event output buffer could not create a new log entry: %s", err)
	}
	if ent.Attributes == nil {
		ent.Attributes = map[string]interface{}{}
	}
	if b.IncludeCommandName {
		ent.Attributes[commandNameLabel] = summary.Command
	}
	if b.IncludeCommandStatus {
		ent.Attributes[commandStatusLabel] = summary.ExitCode
	}
	if b.IncludeDuration {
		ent.Attributes[commandDurationLabel] = summary.RunDuration.Seconds()
	}
	if err := b.writer.Write(b.ctx, ent); err != nil {
		b.logger.Errorf("failed to write entry: %s", err)
	}
}

// Write to the buffer. Meant to be used by both output streams
// in a monitoring plugin spec compliant way.
// Will accept writes until MaxBodySize is reached.
func (b *eventOutputBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if rem := int(b.MaxBodySize) - b.buf.Len(); b.MaxBodySize > 0 && len(p) > rem {
		if w, wErr := b.buf.Write(p[:rem]); wErr != nil {
			return w, wErr
		}
		return rem, errSizeLimitExceeded
	}
	return b.buf.Write(p)
}

func (b *eventOutputBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
