package jobreceiver

import (
	"bytes"
	"fmt"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/builder"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.uber.org/zap"
)

// Build returns the job runner, the process responsible for scheduling and
// running commands and piping their output to the output consumer.
func (c Config) Build(logger *zap.SugaredLogger, out consumer.Interface) (builder.JobRunner, error) {
	return &stubRunner{Consumer: out}, nil
}

// stubRunner is a stub implementation.
type stubRunner struct {
	Consumer consumer.Interface
}

func (r *stubRunner) Start(operator.Persister) error {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "hello world. This is a placeholder message.")
	r.Consumer.Consume(&buf, &bytes.Buffer{})
	return nil
}

func (r *stubRunner) Stop() error {
	return nil
}
