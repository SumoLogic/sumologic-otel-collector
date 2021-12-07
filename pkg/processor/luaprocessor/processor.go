package luaprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/model/pdata"
)

type luaProcessor struct{}

func newLuaProcessor(cfg *Config) *luaProcessor {
	return &luaProcessor{}
}

// ProcessMetrics processes metrics
func (lp *luaProcessor) ProcessMetrics(ctx context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua metrics processor***")

	return md, nil
}

// ProcessTraces processes traces
func (lp *luaProcessor) ProcessTraces(ctx context.Context, md pdata.Traces) (pdata.Traces, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua traces processor***")

	return md, nil
}

// ProcessLogs processes logs
func (lp *luaProcessor) ProcessLogs(ctx context.Context, md pdata.Logs) (pdata.Logs, error) {
	// TODO: add processor logic here
	fmt.Println("***Hello from Lua logs processor***")

	return md, nil
}
