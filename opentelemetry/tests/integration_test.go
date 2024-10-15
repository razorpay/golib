//go:build integration

package tests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	gotel "github.com/razorpay/golib/opentelemetry"
	"github.com/razorpay/golib/opentelemetry/config"
	"github.com/razorpay/golib/opentelemetry/exporter/opentelemetry"
	"github.com/razorpay/golib/opentelemetry/exporter/prometheus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	JaegerQueryPort     = "16686"
	PrometheusQueryPort = "9090"
	ServiceName         = "test-service"
	sleepTime           = 7 * time.Second
)

func setupObservabilityStack(t *testing.T) {
	compose, err := tc.NewDockerCompose("testdata/docker-compose.yaml")
	assert.NoError(t, err, "NewDockerComposeAPI()")

	t.Cleanup(func() {
		assert.NoError(t, compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal), "compose.Down()")
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	assert.NoError(t, compose.Up(ctx, tc.Wait(true)), "compose.Up()")
}

// TestInstrumentation creates the observability stack using docker which receives the telemetry data
// pushed as part of this test. Data is queried from the external receivers for verifying the data
// collection.
func TestInstrumentation(t *testing.T) {
	ctx := context.Background()
	setupObservabilityStack(t)
	generateTelemetryData(ctx, t)
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/traces?service=%s", JaegerQueryPort, ServiceName))
	require.NoError(t, err)

	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyString := string(bodyBytes)
	require.NotEmpty(t, bodyString)
	require.Contains(t, bodyString, "raspberry")
	require.Contains(t, bodyString, "vanilla")
	require.Contains(t, bodyString, "Sample-9")
	require.Contains(t, bodyString, "test-example")

	resp, err = http.Get(fmt.Sprintf("http://localhost:%s/api/v1/label/__name__/values", PrometheusQueryPort))
	require.NoError(t, err)

	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bodyBytes, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyString = string(bodyBytes)
	require.NotEmpty(t, bodyString)
	require.Contains(t, bodyString, "baz_bucket")
	require.Contains(t, bodyString, "go_threads")
	require.Contains(t, bodyString, "foo12")
}

func generateTelemetryData(ctx context.Context, t *testing.T) {
	cfg := &config.Config{
		ServiceName: ServiceName,
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
	require.NoError(t, err)
	tracer := otel.Tracer("test")
	commonAttrs := []attribute.KeyValue{
		attribute.String("attrA", "chocolate"),
		attribute.String("attrB", "raspberry"),
		attribute.String("attrC", "vanilla"),
	}
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
	require.NoError(t, err)
	counter.Add(ctx, 5, opt)
	histogram, err := meter.Float64Histogram(
		"baz",
		api.WithDescription("a histogram with custom buckets and rename"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	require.NoError(t, err)
	histogram.Record(ctx, 136, opt)
	histogram.Record(ctx, 64, opt)
	histogram.Record(ctx, 701, opt)
	histogram.Record(ctx, 830, opt)
	// Wait for flushing telemetry data via exporters
	<-time.After(sleepTime)
}
