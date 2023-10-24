package jobreceiver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/asset"
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
		logger:   logger,
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

// Start running the configured command.
//
// Returns after starting a goroutine that will:
// 1. Ensures all runtime assets are downloaded and available
// 2. Build the command environment the monitoring job will be executed with
// 3. Begin scheduling the monitoring job
func (r *runner) Start(operator.Persister) error {

	ctx := context.WithValue(context.Background(), consumer.ContextKeyCommandName, r.exec.Command)
	ctx, r.cancel = context.WithCancel(ctx)

	//TODO expose asset fetching http, retry and storage strategy as user
	//configuration?
	assetManager := asset.Manager{
		Fetcher: asset.NewFetcher(
			r.logger,
			&http.Client{
				Transport: http.DefaultTransport.(*http.Transport).Clone(),
			},
		),
		StoragePath: os.TempDir(),
		Logger:      r.logger,
	}

	runtimeAssets := r.exec.RuntimeAssets
	if err := assetManager.Validate(runtimeAssets); err != nil {
		return fmt.Errorf("invalid runtime asset(s): %w", err)
	}

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		refs, err := assetManager.InstallAll(ctx, runtimeAssets)
		if err != nil {
			r.logger.Errorf("failed to load runtime assets. monitoringjob will not be scheduled: %s", err)
			return
		}
		commandEnv := os.Environ()
		for _, ref := range refs {
			commandEnv = ref.MergeEnvironment(commandEnv)
		}
		r.run(ctx, commandEnv)
	}()

	return nil
}

func (r *runner) Stop() error {
	r.cancel()
	r.wg.Wait()
	return nil
}

func (r *runner) run(ctx context.Context, commandEnv []string) {
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
			Env:       commandEnv,
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
}
