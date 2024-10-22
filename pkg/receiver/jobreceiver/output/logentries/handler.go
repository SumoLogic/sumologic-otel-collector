package logentries

import (
	"context"
	"io"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"go.uber.org/zap"
)

const (
	streamNameStdout       = "stdout"
	streamNameStderr       = "stderr"
	commandNameLabel       = "command.name"
	commandStreamNameLabel = "command.stream.name"
)

type handler struct {
	logger *zap.SugaredLogger
	writer consumer.WriterOp

	config      LogEntriesConfig
	scanFactory scannerFactory
}

func (h *handler) Consume(ctx context.Context, stdout, stderr io.Reader) consumer.CloseFunc {
	go h.consume(ctx, stdout, streamNameStdout)
	go h.consume(ctx, stderr, streamNameStderr)
	return nopCloser
}

// TODO(ck) add buffer between reading and processing #1309
func (h *handler) consume(ctx context.Context, in io.Reader, stream string) {
	scanner := h.scanFactory.Build(in)
	for scanner.Scan() {
		ent, err := h.writer.NewEntry(scanner.Text())
		if err != nil {
			h.logger.Errorf("log entry handler could not create a new log entry: %s", err)
		}
		if ent.Attributes == nil {
			ent.Attributes = map[string]interface{}{}
		}
		if h.config.IncludeCommandName {
			ent.Attributes[commandNameLabel] = ctx.Value(consumer.ContextKeyCommandName)
		}
		if h.config.IncludeStreamName {
			ent.Attributes[commandStreamNameLabel] = stream
		}
		if err := h.writer.Write(ctx, ent); err != nil {
			h.logger.Errorf("failed to write log entry: %s", err)
		}
	}
	if err := scanner.Err(); err != nil {
		h.logger.Errorf("error reading input stream", zap.String("stream", stream), zap.Error(err))
	}
}

func nopCloser(_ consumer.ExecutionSummary) {}
