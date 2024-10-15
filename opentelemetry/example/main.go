package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	gotel "github.com/razorpay/golib/opentelemetry"
	"github.com/razorpay/golib/opentelemetry/config"
	"github.com/razorpay/golib/opentelemetry/exporter/opentelemetry"
	"github.com/razorpay/golib/opentelemetry/exporter/prometheus"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	setupObs(ctx)
	instrumentRandomData(ctx)
	// Wait for flushing telemetry data via exporters
	<-time.After(5 * time.Second)
	// After cancel traceprovider, meterprovider and metrics endpoint would shut down
	cancel()
}

func instrumentRandomData(ctx context.Context) {
	commonAttrs := []attribute.KeyValue{
		attribute.String("attrA", "chocolate"),
		attribute.String("attrB", "raspberry"),
		attribute.String("attrC", "vanilla"),
	}
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(
		ctx,
		"test-example",
		trace.WithAttributes(commonAttrs...))
	for i := 0; i < 10; i++ {
		_, iSpan := tracer.Start(ctx, fmt.Sprintf("Sample-%d", i))
		<-time.After(10 * time.Millisecond)
		iSpan.End()
	}
	span.End()

	meter := otel.Meter("test")
	opt := api.WithAttributes(
		attribute.Key("A").String("B"),
		attribute.Key("C").String("D"),
	)
	counter, err := meter.Float64Counter("foo12", api.WithDescription("a simple counter"))
	if err != nil {
		panic(err)
	}
	counter.Add(ctx, 5, opt)
	histogram, err := meter.Float64Histogram(
		"baz",
		api.WithDescription("a histogram with custom buckets and rename"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	if err != nil {
		panic(err)
	}
	histogram.Record(ctx, 136, opt)
	histogram.Record(ctx, 64, opt)
	histogram.Record(ctx, 701, opt)
	histogram.Record(ctx, 830, opt)
}

func setupObs(ctx context.Context) {
	cfg := &config.Config{
		ServiceName: "test-service",
		Exporters: []config.Exporter{
			{
				Name: "prom",
				Kind: prometheus.ExporterKey,
				Config: map[string]interface{}{
					"port":            9092,
					"process_metrics": true,
					"go_metrics":      true,
				},
			},
			{
				Name: "otel",
				Kind: opentelemetry.ExporterKey,
				Config: map[string]interface{}{
					"host":     "localhost",
					"port":     4317,
					"use_http": false,
				},
			},
		},
		Metrics: &config.MetricsConfig{
			Exporters: []string{"prom"},
		},
		Trace: &config.TraceConfig{
			Exporters:  []string{"otel"},
			SampleRate: 1.0,
		},
	}
	err := gotel.Register(ctx, cfg, nil)
	if err != nil {
		panic(err)
	}
}
