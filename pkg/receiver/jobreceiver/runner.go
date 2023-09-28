package jobreceiver

import (
	"context"
	"sync"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/builder"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/command"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.uber.org/zap"
)

// Build returns the job runner, the process responsible for scheduling and
// running commands and piping their output to the output consumer.
func (c Config) Build(logger *zap.SugaredLogger, out consumer.Interface) (builder.JobRunner, error) {
	return &runner{
		exec:     c.Exec,
		schedule: c.Schedule,
		consumer: out,
	}, nil
}

// runner schedules and executes commands
type runner struct {
	exec     ExecutionConfig
	schedule ScheduleConfig
	consumer consumer.Interface

	logger *zap.SugaredLogger

	wg     sync.WaitGroup
	cancel func()
}

// Start stub impl. runs command once at startup then idles indefinitely.
func (r *runner) Start(operator.Persister) error {
	r.wg.Add(1)

	ctx := context.WithValue(context.Background(), consumer.ContextKeyCommandName, r.exec.Command)
	ctx, r.cancel = context.WithCancel(ctx)
	go func() {
		defer r.wg.Done()

		// TODO(ck) spec using persistence for interval timing.
		ticker := time.NewTicker(r.schedule.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			cmd := command.NewExecution(ctx, command.ExecutionRequest{
				Command:   r.exec.Command,
				Arguments: r.exec.Arguments,
				Timeout:   r.exec.Timeout,
			})

			stdout, err := cmd.Stdout()
			if err != nil {
				r.logger.Errorf("monitoringjob runner failed to create command Stdout pipe: %s", err)
				continue
			}
			stderr, err := cmd.Stderr()
			if err != nil {
				r.logger.Errorf("monitoringjob runner failed to create command Stderr pipe: %s", err)
				continue
			}
			cb := r.consumer.Consume(ctx, stdout, stderr)

			resp, err := cmd.Run()
			if err != nil {
				r.logger.Errorf("monitoringjob runner failed to run command: %s", err)
				continue
			}
			cb(consumer.ExecutionSummary{
				Command:     r.exec.Command,
				ExitCode:    resp.Status,
				RunDuration: resp.Duration,
			})
		}

	}()
	return nil
}

func (r *runner) Stop() error {
	r.cancel()
	r.wg.Wait()
	return nil
}
