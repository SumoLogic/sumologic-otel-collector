package jobreceiver

import (
	"context"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/builder"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/command"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.uber.org/zap"
)

// Build returns the job runner, the process responsible for scheduling and
// running commands and piping their output to the output consumer.
func (c Config) Build(logger *zap.SugaredLogger, out consumer.Interface) (builder.JobRunner, error) {
	return &stubRunner{
		Exec:     c.Exec,
		Consumer: out,
	}, nil
}

// stubRunner is a stub implementation.
type stubRunner struct {
	Exec     ExecutionConfig
	Consumer consumer.Interface
}

// Start stub impl. runs command once at startup then idles indefinitely.
func (r *stubRunner) Start(operator.Persister) error {
	go func() {
		ctx := context.Background()
		cmd, err := command.NewInvocation(command.ExecutionRequest{
			Command:   r.Exec.Command,
			Arguments: r.Exec.Arguments,
			Timeout:   r.Exec.Timeout,
		})

		if err != nil {
			panic(err)
		}
		cb := r.Consumer.Consume(ctx, cmd.Stdout(), cmd.Stderr())

		resp, err := cmd.Run(ctx)
		if err != nil {
			panic(err)
		}
		cb(consumer.ExecutionSummary{
			Command:     r.Exec.Command,
			ExitCode:    resp.Status,
			RunDuration: resp.Duration,
		})

	}()
	return nil
}

func (r *stubRunner) Stop() error {
	return nil
}
