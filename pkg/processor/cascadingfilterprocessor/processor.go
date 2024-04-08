// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cascadingfilterprocessor

import (
	"context"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"

	lru "github.com/hashicorp/golang-lru"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/config"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/idbatcher"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/sampling"
)

// TraceAcceptEvaluator combines a sampling policy evaluator with the destinations to be
// used for that policy.
type TraceAcceptEvaluator struct {
	// Name used to identify this policy instance.
	Name string
	// Evaluator that decides if a trace is sampled or not by this policy instance.
	Evaluator sampling.PolicyEvaluator
	// ctx used to carry metric tags of each policy.
	ctx context.Context
	// probabilisticFilter determines whether `sampling.probability` field must be calculated and added
	probabilisticFilter bool
}

// TraceRejectEvaluator holds checking if trace should be dropped completely before further processing
type TraceRejectEvaluator struct {
	// Name used to identify this policy instance.
	Name string
	// Evaluator that decides if a trace is sampled or not by this policy instance.
	Evaluator sampling.DropTraceEvaluator
	// ctx used to carry metric tags of each policy.
	ctx context.Context
}

// traceKey is defined since sync.Map requires a comparable type, isolating it on its own
// type to help track usage.
type traceKey pcommon.TraceID

// cascadingFilterSpanProcessor handles the incoming trace data and uses the given sampling
// policy to sample traces.
type cascadingFilterSpanProcessor struct {
	ctx              context.Context
	nextConsumer     consumer.Traces
	instanceName     string
	start            sync.Once
	maxNumTraces     uint64
	traceAcceptRules []*TraceAcceptEvaluator
	traceRejectRules []*TraceRejectEvaluator
	logger           *zap.Logger
	idToTrace        sync.Map
	policyTicker     tTicker
	decisionBatcher  idbatcher.Batcher
	decisionHistory  *lru.TwoQueueCache
	deleteChan       chan traceKey
	numTracesOnMap   uint64

	filteringEnabled bool

	decisionSpansLimitter *rateLimiter
	priorSpansLimitter    *rateLimiter
}

type decisionHistoryInfo struct {
	finalDecision       sampling.Decision
	filterName          string
	probabilisticFilter bool
}

const (
	probabilisticFilterPolicyName = "probabilistic_filter"
	probabilisticRuleVale         = "probabilistic"
	filteredRuleValue             = "filtered"
	AttributeSamplingRule         = "sampling.rule"
	AttributeSamplingFilter       = "sampling.filter"
	AttributeSamplingLateArrival  = "sampling.late_arrival"

	AttributeSamplingProbability = "sampling.probability"

	defaultCollectorInstancesNo = 1
)

// newTraceProcessor returns a processor.TraceProcessor that will perform Cascading Filter according to the given
// configuration.
func newTraceProcessor(logger *zap.Logger, nextConsumer consumer.Traces, cfg config.Config, id component.ID) (*cascadingFilterSpanProcessor, error) {
	numDecisionBatches := uint64(cfg.DecisionWait.Seconds())
	inBatcher, err := idbatcher.New(numDecisionBatches, cfg.ExpectedNewTracesPerSec, uint64(2*runtime.NumCPU()))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	var policies []*TraceAcceptEvaluator
	var dropTraceEvals []*TraceRejectEvaluator

	// In case of lack of collectorInstances set default.
	if cfg.CollectorInstances == 0 {
		cfg.CollectorInstances = defaultCollectorInstancesNo
		logger.Info("Using default collector instances", zap.Uint("value", defaultCollectorInstancesNo))
	}

	// Prepare Trace Reject config

	for _, dropCfg := range cfg.TraceRejectCfgs {
		dropCtx, err := tag.New(ctx, tag.Upsert(tagPolicyKey, dropCfg.Name), tag.Upsert(tagPolicyDecisionKey, statusDropped))
		if err != nil {
			return nil, err
		}
		evaluator, err := sampling.NewDropTraceEvaluator(logger, dropCfg)
		if err != nil {
			return nil, err
		}
		dropEval := &TraceRejectEvaluator{
			Name:      dropCfg.Name,
			Evaluator: evaluator,
			ctx:       dropCtx,
		}
		logger.Info("Adding trace reject rule", zap.String("name", dropCfg.Name))
		dropTraceEvals = append(dropTraceEvals, dropEval)
	}

	// Prepare Trace Accept config

	var policyCfgs []config.TraceAcceptCfg
	totalRate := int32(0)

	if len(cfg.TraceAcceptCfgs) > 0 {
		policyCfgs = append(policyCfgs, cfg.TraceAcceptCfgs...)
	}

	if len(cfg.PolicyCfgs) > 0 {
		logger.Warn("'traceAcceptRules' is deprecated and will be removed in future versions, please use 'trace_accept_filters' instead")
		policyCfgs = append(policyCfgs, cfg.PolicyCfgs...)
	}

	for i := range policyCfgs {
		policyCfg := policyCfgs[i]
		policyCtx, err := tag.New(ctx, tag.Upsert(tagPolicyKey, policyCfg.Name))
		if err != nil {
			return nil, err
		}

		if policyCfg.SpansPerSecond > 0 {
			policyCalculatedSpansPerSecond := calculateSpansPerSecond(policyCfg.SpansPerSecond, cfg.CollectorInstances)
			policyCfg.SpansPerSecond = policyCalculatedSpansPerSecond
			totalRate += policyCfg.SpansPerSecond
		}

		eval, err := buildPolicyEvaluator(logger, &policyCfg)
		if err != nil {
			return nil, err
		}
		policy := &TraceAcceptEvaluator{
			Name:                policyCfg.Name,
			Evaluator:           eval,
			ctx:                 policyCtx,
			probabilisticFilter: false,
		}

		logger.Info("Adding trace accept rule",
			zap.String("name", policyCfg.Name),
			zap.Int32("spans_per_second", policyCfg.SpansPerSecond),
			zap.Uint("collector_instances", cfg.CollectorInstances),
		)

		policies = append(policies, policy)
	}

	// Recalculate the total spans per second rate if needed
	calculatedGlobalSpansPerSecond := calculateSpansPerSecond(cfg.SpansPerSecond, cfg.CollectorInstances)
	cfg.SpansPerSecond = calculatedGlobalSpansPerSecond
	spansPerSecond := calculatedGlobalSpansPerSecond
	if spansPerSecond == 0 {
		spansPerSecond = totalRate
		if cfg.ProbabilisticFilteringRate != nil && *cfg.ProbabilisticFilteringRate > 0 {
			spansPerSecond += *cfg.ProbabilisticFilteringRate
		}
	}

	if spansPerSecond != 0 {
		logger.Info("Setting total spans per second limit, based on configured collector instances",
			zap.Int32("spans_per_second", spansPerSecond),
			zap.Uint("collector_instances", cfg.CollectorInstances),
		)
	} else {
		logger.Info("Not setting total spans per second limit (only selected traces will be filtered out)")
	}

	// Setup probabilistic filtering - using either ratio or rate.
	// This must be always evaluated first as it must select traces independently of other traceAcceptRules

	probabilisticFilteringRate := int32(-1)

	if cfg.ProbabilisticFilteringRatio != nil && *cfg.ProbabilisticFilteringRatio > 0.0 && spansPerSecond > 0 {
		probabilisticFilteringRate = int32(float32(spansPerSecond) * *cfg.ProbabilisticFilteringRatio)
	} else if cfg.ProbabilisticFilteringRate != nil && *cfg.ProbabilisticFilteringRate > 0 {
		probabilisticFilteringRate = *cfg.ProbabilisticFilteringRate
	}

	if probabilisticFilteringRate > 0 {
		logger.Info("Setting probabilistic filtering rate", zap.Int32("probabilistic_filtering_rate", probabilisticFilteringRate))

		policyCtx, err := tag.New(ctx, tag.Upsert(tagPolicyKey, probabilisticFilterPolicyName))
		if err != nil {
			return nil, err
		}
		eval, err := buildProbabilisticFilterEvaluator(logger, probabilisticFilteringRate)
		if err != nil {
			return nil, err
		}
		policy := &TraceAcceptEvaluator{
			Name:                probabilisticFilterPolicyName,
			Evaluator:           eval,
			ctx:                 policyCtx,
			probabilisticFilter: true,
		}
		policies = append([]*TraceAcceptEvaluator{policy}, policies...)
	} else {
		logger.Info("Not setting probabilistic filtering rate")
	}

	// This allows to not buffer data when no filters are defined (and the processor is still in the pipeline)
	if len(policies) == 0 && len(dropTraceEvals) == 0 {
		logger.Info("No rules set for cascading_filter processor. Processor wil output all incoming spans without filtering.")
	}

	historySize := cfg.HistorySize
	if historySize == nil {
		logger.Info("setting history size to the same value as num_traces", zap.Uint64("num_traces", cfg.NumTraces))
		historySize = &cfg.NumTraces
	}
	cache, err := lru.New2Q(int(*historySize))
	if err != nil {
		return nil, err
	}

	var priorSpansRate int32
	if cfg.PriorSpansRate != nil {
		priorSpansRate = *cfg.PriorSpansRate
	} else {
		priorSpansRate = cfg.SpansPerSecond / 2
		logger.Info("setting prior spans rate to half of spans per second", zap.Int32("prior_spans_rate", priorSpansRate))
	}

	// Build the span processor
	cfsp := &cascadingFilterSpanProcessor{
		ctx:                   ctx,
		nextConsumer:          nextConsumer,
		instanceName:          id.String(),
		maxNumTraces:          cfg.NumTraces,
		decisionSpansLimitter: newRateLimitter(spansPerSecond),
		priorSpansLimitter:    newRateLimitter(priorSpansRate),
		logger:                logger,
		decisionBatcher:       inBatcher,
		decisionHistory:       cache,
		traceAcceptRules:      policies,
		traceRejectRules:      dropTraceEvals,
		filteringEnabled:      len(policies) > 0 || len(dropTraceEvals) > 0,
	}

	cfsp.policyTicker = &policyTicker{onTick: cfsp.samplingPolicyOnTick}
	cfsp.deleteChan = make(chan traceKey, cfg.NumTraces)

	return cfsp, nil
}

func buildPolicyEvaluator(logger *zap.Logger, cfg *config.TraceAcceptCfg) (sampling.PolicyEvaluator, error) {
	return sampling.NewFilter(logger, cfg)
}

func buildProbabilisticFilterEvaluator(logger *zap.Logger, maxSpanRate int32) (sampling.PolicyEvaluator, error) {
	return sampling.NewProbabilisticFilter(logger, maxSpanRate)
}

type policyMetrics struct {
	idNotFoundOnMapCount, evaluateErrorCount, decisionSampled, decisionNotSampled int64
}

func (cfsp *cascadingFilterSpanProcessor) samplingPolicyOnTick() {
	batch, _ := cfsp.decisionBatcher.CloseCurrentAndTakeFirstBatch()
	t := newCascade(cfsp)
	t.decideOnBatch(&batch)
}

// ConsumeTraces is required by the SpanProcessor interface.
func (cfsp *cascadingFilterSpanProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	if !cfsp.filteringEnabled {
		return cfsp.nextConsumer.ConsumeTraces(ctx, td)
	}

	cfsp.start.Do(func() {
		cfsp.logger.Info("First trace data arrived, starting cascading_filter timers")
		cfsp.policyTicker.Start(1 * time.Second)
	})
	resourceSpans := td.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		resourceSpan := resourceSpans.At(i)
		cfsp.processTraces(ctx, resourceSpan)
	}
	return nil
}

// TODO: include InstrumentationScope
func (cfsp *cascadingFilterSpanProcessor) groupSpansByTraceKey(resourceSpans ptrace.ResourceSpans) map[traceKey][]*ptrace.Span {
	idToSpans := make(map[traceKey][]*ptrace.Span)
	ss := resourceSpans.ScopeSpans()
	for j := 0; j < ss.Len(); j++ {
		ils := ss.At(j)
		spansLen := ils.Spans().Len()
		for k := 0; k < spansLen; k++ {
			span := ils.Spans().At(k)
			tk := traceKey(span.TraceID())
			if len(tk) != 16 {
				cfsp.logger.Warn("Span without valid TraceId")
			}
			idToSpans[tk] = append(idToSpans[tk], &span)
		}
	}
	return idToSpans
}

func (cfsp *cascadingFilterSpanProcessor) bufferTraces(id traceKey, res pcommon.Resource, spans []*ptrace.Span) int64 {
	newTraceIDs := int64(0)
	lenSpans := int32(len(spans))
	lenPolicies := len(cfsp.traceAcceptRules)
	initialDecisions := make([]sampling.Decision, lenPolicies)

	for i := 0; i < lenPolicies; i++ {
		initialDecisions[i] = sampling.Pending
	}
	initialTraceData := &sampling.TraceData{
		Decisions:   initialDecisions,
		ArrivalTime: time.Now(),
		SpanCount:   lenSpans,
	}
	d, loaded := cfsp.idToTrace.LoadOrStore(id, initialTraceData)

	actualData := d.(*sampling.TraceData)
	if loaded {
		atomic.AddInt32(&actualData.SpanCount, lenSpans)
	} else {
		newTraceIDs++
		cfsp.decisionBatcher.AddToCurrentBatch(pcommon.TraceID(id))
		atomic.AddUint64(&cfsp.numTracesOnMap, 1)
		postDeletion := false

		for !postDeletion {
			select {
			case cfsp.deleteChan <- id:
				postDeletion = true
			default:
				// Note this is a buffered channel, so this will only delete excessive traces (if they exist)
				traceKeyToDrop := <-cfsp.deleteChan
				cfsp.dropTrace(traceKeyToDrop)
			}
		}
	}

	// Add the spans to the trace, but only once for all policy, otherwise same spans will
	// be duplicated in the final trace.
	actualData.Lock()
	finalDecision := actualData.FinalDecision

	// If decision is pending, we want to add the new spans still under the lock, so the decision doesn't happen
	// in between the transition from pending.
	if finalDecision == sampling.Pending || finalDecision == sampling.Unspecified {
		// Add the spans to the trace, but only once for all policy, otherwise same spans will
		// be duplicated in the final trace.

		traceTd := prepareTraceBatch(res, spans)
		actualData.ReceivedBatches = append(actualData.ReceivedBatches, traceTd)
	}

	actualData.Unlock()

	return newTraceIDs
}

func (cfsp *cascadingFilterSpanProcessor) processTraces(ctx context.Context, resourceSpans ptrace.ResourceSpans) {
	// Group spans per their traceId to minimize contention on idToTrace
	idToSpans := cfsp.groupSpansByTraceKey(resourceSpans)
	currTime := time.Now().Unix()

	var newTraceIDs int64
	for id, spans := range idToSpans {
		if decision, found := cfsp.decisionHistory.Get(id); found {
			info := decision.(decisionHistoryInfo)
			finalDecision := info.finalDecision
			if finalDecision == sampling.Sampled {
				// First check if it even fits within the overall prior limit
				finalDecision = cfsp.priorSpansLimitter.updateRate(currTime, int32(len(spans)))
			}

			switch finalDecision {
			case sampling.Sampled:
				// Forward the spans to the policy destinations
				traceTd := prepareTraceBatch(resourceSpans.Resource(), spans)
				updateLateArrival(traceTd, info.filterName, info.probabilisticFilter)
				if err := cfsp.nextConsumer.ConsumeTraces(ctx, traceTd); err != nil {
					cfsp.logger.Warn("Error sending late arrived spans to destination",
						zap.Error(err))
				}
				recordSpanLateDecision(cfsp.ctx, cfsp.instanceName, statusSampled, len(spans))
				continue
			case sampling.NotSampled:
				recordSpanLateDecision(cfsp.ctx, cfsp.instanceName, statusNotSampled, len(spans))
				continue
			case sampling.Dropped:
				recordSpanLateDecision(cfsp.ctx, cfsp.instanceName, statusDropped, len(spans))
				continue
			default:
				cfsp.logger.Warn("Encountered unexpected sampling decision",
					zap.Int("decision", int(info.finalDecision)))
			}

		}

		newTraceIDs += cfsp.bufferTraces(id, resourceSpans.Resource(), spans)
	}

	//nolint:errcheck
	_ = stats.RecordWithTags(
		cfsp.ctx,
		[]tag.Mutator{tag.Insert(tagProcessorKey, cfsp.instanceName)},
		statNewTraceIDReceivedCount.M(newTraceIDs),
	)
}

func (cfsp *cascadingFilterSpanProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// Start is invoked during service startup.
func (cfsp *cascadingFilterSpanProcessor) Start(context.Context, component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (cfsp *cascadingFilterSpanProcessor) Shutdown(context.Context) error {
	return nil
}

func (cfsp *cascadingFilterSpanProcessor) dropTrace(traceID traceKey) {
	var trace *sampling.TraceData
	if d, ok := cfsp.idToTrace.Load(traceID); ok {
		trace = d.(*sampling.TraceData)
		cfsp.idToTrace.Delete(traceID)
		// Subtract one from numTracesOnMap per https://godoc.org/sync/atomic#AddUint64
		atomic.AddUint64(&cfsp.numTracesOnMap, ^uint64(0))
	}
	if trace == nil {
		cfsp.logger.Debug("Attempt to delete traceID not on table")
		return
	}
}

func prepareTraceBatch(res pcommon.Resource, spans []*ptrace.Span) ptrace.Traces {
	traceTd := ptrace.NewTraces()
	rs := traceTd.ResourceSpans().AppendEmpty()
	res.CopyTo(rs.Resource())
	ss := rs.ScopeSpans().AppendEmpty()
	ilsSpans := ss.Spans()
	for _, span := range spans {
		span.CopyTo(ilsSpans.AppendEmpty())
	}
	return traceTd
}

// tTicker interface allows easier testing of cascade related functionality used by cascadingfilterprocessor
type tTicker interface {
	// Start sets the frequency of the ticker and starts the periodic calls to OnTick.
	Start(d time.Duration)
	// OnTick is called when the ticker fires.
	OnTick()
	// Stops firing the ticker.
	Stop()
}

type policyTicker struct {
	ticker *time.Ticker
	onTick func()
}

func (pt *policyTicker) Start(d time.Duration) {
	pt.ticker = time.NewTicker(d)
	go func() {
		for range pt.ticker.C {
			pt.OnTick()
		}
	}()
}
func (pt *policyTicker) OnTick() {
	pt.onTick()
}
func (pt *policyTicker) Stop() {
	pt.ticker.Stop()
}

var _ tTicker = (*policyTicker)(nil)

func calculateSpansPerSecond(spansPerSecond int32, collectorInstances uint) int32 {
	calculateSpansPerSecond := float64(spansPerSecond) / float64(collectorInstances)
	roundedSpansPerSecond := int32(math.Ceil(calculateSpansPerSecond))

	return roundedSpansPerSecond
}
