package logentries

import (
	"bufio"
	"fmt"
	"io"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/decode"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"go.uber.org/zap"
)

const (
	maxDefaultBufferSize = 32 * 1024
	defaultMaxLogSize    = 1024 * 1024
)

// LogEntriesConfig handles output as if it is a stream of distinct log events
type LogEntriesConfig struct {
	// IncludeCommandName indicates to include the attribute `command.name`
	IncludeCommandName bool `mapstructure:"include_command_name,omitempty"`
	// IncludeStreamName indicates to include the attribute `command.stream.name`
	// indicating the stream that log entry was consumed from. stdout | stdin.
	IncludeStreamName bool `mapstructure:"include_stream_name,omitempty"`
	// MaxLogSize restricts the length of log output to a specified size.
	// Excess output will overflow to subsequent log entries.
	MaxLogSize helper.ByteSize `mapstructure:"max_log_size,omitempty"`
	// Encoding to expect from output
	Encoding string `mapstructure:"encoding"`
	// Multiline configures alternate log line deliniation
	Multiline split.Config `mapstructure:"multiline"`
}

func (c *LogEntriesConfig) Build(logger *zap.SugaredLogger, op consumer.WriterOp) (consumer.Interface, error) {

	encoding, err := decode.LookupEncoding(c.Encoding)
	if err != nil {
		return nil, fmt.Errorf("log_entries configuration unable to use encoding %s: %w", c.Encoding, err)
	}
	splitFunc, err := c.Multiline.Func(encoding, true, int(c.MaxLogSize))
	if err != nil {
		return nil, fmt.Errorf("log_entries configuration could not build split function: %w", err)
	}
	return &handler{
		logger: logger,
		writer: op,
		config: *c,
		scanFactory: scannerFactory{
			splitFunc:  splitFunc,
			maxLogSize: int(c.MaxLogSize),
		},
	}, nil
}

type LogEntriesConfigFactory struct{}

func (LogEntriesConfigFactory) CreateDefaultConfig() consumer.Builder {
	return &LogEntriesConfig{
		IncludeCommandName: true,
		IncludeStreamName:  true,
		MaxLogSize:         defaultMaxLogSize,
	}
}

type scannerFactory struct {
	maxLogSize int
	splitFunc  bufio.SplitFunc
}

func (f scannerFactory) Build(in io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(in)

	if f.maxLogSize <= 0 {
		f.maxLogSize = defaultMaxLogSize
	}
	bufferSize := f.maxLogSize / 2
	if bufferSize > maxDefaultBufferSize {
		bufferSize = maxDefaultBufferSize
	}
	scanner.Buffer(make([]byte, 0, bufferSize), f.maxLogSize)
	scanner.Split(f.splitWithTruncate())
	return scanner
}

func (f scannerFactory) splitWithTruncate() bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = f.splitFunc(data, atEOF)
		if advance == 0 && token == nil && len(data) >= f.maxLogSize {
			advance, token = f.maxLogSize, data[:f.maxLogSize]
		} else if len(token) > f.maxLogSize {
			advance, token = f.maxLogSize, data[:f.maxLogSize]
		}
		return
	}
}
