package consumer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"go.uber.org/zap"
)

// Interface consumes command output and emits telemetry data
type Interface interface {
	Consume(stdin, stderr io.Reader) CloseFunc
}

// CloseFunc
type CloseFunc func(ExecutionSummary)

// ExecutionSummary describes a command execution
type ExecutionSummary struct {
	Command     string
	ExitCode    int
	RunDuration time.Duration
}

// Builder builds a consumer
type Builder interface {
	Build(*zap.SugaredLogger, WriterOp) (Interface, error)
}

// WriterOp is the consumer's interface to the stanza pipeline
type WriterOp interface {
	NewEntry(value interface{}) (*entry.Entry, error)
	Write(ctx context.Context, e *entry.Entry)
}

// DemoConsumer stub consumer implementation.
// todo(ck) delete - this is a stub implementation for PoC purposes only.
type DemoConsumer struct {
	WriterOp
	Logger *zap.SugaredLogger
}

// Consume reads stdout line by line and produces entries
func (p *DemoConsumer) Consume(stdout, _ io.Reader) CloseFunc {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		scanner := bufio.NewScanner(stdout)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if !scanner.Scan() {
				return
			}
			ent, err := p.NewEntry(scanner.Text())
			if err != nil {
				ent = entry.New()
				ent.Body = fmt.Sprintf("error: %s", err)
			}
			p.Write(ctx, ent)

		}
	}()
	return func(_ ExecutionSummary) { cancel() }
}
