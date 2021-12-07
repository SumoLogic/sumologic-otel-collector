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
	fmt.Println("***Hello from Lua processor***")

	return md, nil
}
