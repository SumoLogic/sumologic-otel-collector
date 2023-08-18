package builder

import (
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

type JobRunnerBuilder interface {
	Build(logger *zap.SugaredLogger, consumer consumer.Interface) (JobRunner, error)
}

type JobRunner interface {
	Start(operator.Persister) error
	Stop() error
}

// NewOperatorBuilder builds a stanza operator.Builder from monitoring job
// configuration objects.
func NewOperatorBuilder(outputCfg output.Config, executorCfg JobRunnerBuilder) operator.Builder {
	return &pipelineInputConfig{
		InputConfig:      outputCfg.InputConfig,
		ConsumerBuilder:  outputCfg.Builder,
		JobRunnerBuilder: executorCfg,
	}
}

type pipelineInputConfig struct {
	helper.InputConfig `mapstructure:",squash"`
	ConsumerBuilder    consumer.Builder

	JobRunnerBuilder JobRunnerBuilder
}

// Build the stanza input operator.
func (cfg *pipelineInputConfig) Build(logger *zap.SugaredLogger) (operator.Operator, error) {
	inputBase, err := cfg.InputConfig.Build(logger)
	if err != nil {
		return nil, err
	}
	inputOp := &inputOperator{
		InputOperator: inputBase,
	}

	// point the consumer at this input operator
	consumer, err := cfg.ConsumerBuilder.Build(logger, inputOp)
	if err != nil {
		return nil, err
	}
	// point the job runner at the consumer
	runner, err := cfg.JobRunnerBuilder.Build(logger, consumer)
	if err != nil {
		return nil, err
	}
	inputOp.JobRunner = runner
	return inputOp, nil
}

// inputOperator is the actual stanza input operator implementation.
type inputOperator struct {
	helper.InputOperator
	JobRunner JobRunner
}

func (op *inputOperator) Start(p operator.Persister) error {
	return op.JobRunner.Start(p)
}
