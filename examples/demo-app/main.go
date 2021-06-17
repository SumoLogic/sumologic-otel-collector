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

// Sample contains a program that exports to the OpenTelemetry service.
package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc"
)

func initProvider() func() {
	ctx := context.Background()

	otelAgentAddr, ok := os.LookupEnv("SUMO_OTEL_AGENT_ENDPOINT")
	if !ok {
		otelAgentAddr = "0.0.0.0:4317"
	}

	exp, err := otlp.NewExporter(ctx, otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(otelAgentAddr),
		otlpgrpc.WithDialOption(grpc.WithBlock()),
	))
	handleErr(err, "failed to create otlp exporter")

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("sumologic-otel-demo"),
		),
	)
	handleErr(err, "failed to create service name resource")

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithCollectPeriod(7*time.Second),
		controller.WithExporter(exp),
	)

	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetTracerProvider(tracerProvider)
	global.SetMeterProvider(cont.MeterProvider())
	handleErr(cont.Start(context.Background()), "failed to start metric controller")

	return func() {
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown provider")
		handleErr(cont.Stop(context.Background()), "failed to stop metrics controller") // pushes any last exports to the receiver
		handleErr(exp.Shutdown(ctx), "failed to stop exporter")
	}
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func main() {
	shutdown := initProvider()
	defer shutdown()

	tracer := otel.Tracer("sumologic-otel-demo-tracer")
	meter := global.Meter("sumologic-otel-demo-meter")

	commonLabels := []attribute.KeyValue{
		attribute.String("custom_label", "custom_value"),
	}

	rngVal := metric.Must(meter).
		NewFloat64ValueRecorder(
			"sumologic_otel_demo/rng_value",
			metric.WithDescription("Random number generator value"),
		).Bind(commonLabels...)
	defer rngVal.Unbind()

	constVal := metric.Must(meter).
		NewInt64ValueRecorder(
			"sumologic_otel_demo/const_value",
			metric.WithDescription("Constant value"),
		).Bind(commonLabels...)
	defer constVal.Unbind()

	counterVal := metric.Must(meter).
		NewInt64Counter(
			"sumologic_otel_demo/counter_value",
			metric.WithDescription("Incremental counter value"),
		).Bind(commonLabels...)
	defer counterVal.Unbind()

	defaultCtx := baggage.ContextWithValues(context.Background(), commonLabels...)
	for {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		ctx, span := tracer.Start(defaultCtx, "Generate metrics")
		rngVal.Record(ctx, rng.Float64())
		span.AddEvent("Random metric recorded")
		constVal.Record(ctx, 1024)
		span.AddEvent("Constant metric recorded")
		counterVal.Add(ctx, 1)
		span.AddEvent("Counter metric incremented")
		_, sleepSpan := tracer.Start(ctx, "Sleep for 10s")

		time.Sleep(time.Duration(10) * time.Second)
		sleepSpan.End()

		span.End()
	}
}
