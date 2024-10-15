package opentelemetry

import (
	"context"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"testing"

	"github.com/razorpay/golib/opentelemetry/config"
	"github.com/razorpay/golib/opentelemetry/exporter/opentelemetry"
	"github.com/razorpay/golib/opentelemetry/exporter/prometheus"

	"github.com/stretchr/testify/require"
)

func TestObsWithValidConfig(t *testing.T) {
	ctx := context.Background()
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
	view := metric.NewView(
		metric.Instrument{
			Name:  "latency",
			Scope: instrumentation.Scope{Name: "opentelemetry"},
		},
		metric.Stream{
			Aggregation: metric.AggregationBase2ExponentialHistogram{
				MaxSize:  160,
				MaxScale: 20,
			},
		},
	)
	views := []metric.View{view}
	err := Register(ctx, cfg, views)
	require.NoError(t, err)
}

func TestObsWithNoExporterConfig(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		ServiceName: "test-service",
		Exporters:   []config.Exporter{},
		Metrics: &config.MetricsConfig{
			Exporters: []string{"prom"},
		},
		Trace: &config.TraceConfig{
			Exporters:  []string{"otel"},
			SampleRate: 1.0,
		},
	}
	err := Register(ctx, cfg, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "no exporters declared")
}
