package consumer

import (
	"context"
	"io"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/entry"
	"go.uber.org/zap"
)

// Interface consumes command output and emits telemetry data
type Interface interface {
	Consume(ctx context.Context, stdin, stderr io.Reader) CloseFunc
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
	Write(ctx context.Context, e *entry.Entry) error
}

type contextKey string

const ContextKeyCommandName = contextKey("commandName")
