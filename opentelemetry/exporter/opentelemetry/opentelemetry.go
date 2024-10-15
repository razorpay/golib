// Package opentelemetry implements the Open Telemetry exporter.
package opentelemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/razorpay/golib/opentelemetry/config"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	// ExporterKey is the name for the opentelemetry collector
	ExporterKey      = config.ExporterKind("opentelemetry")
	RemoteServerPort = 4317
	RemoteServerHost = "localhost"
)

// CollectorConfig has the variables to configure
// the otel collector
type CollectorConfig struct {
	// Host is the host of remote server receiving telemetry data (exp: otel-collector endpoint host)
	Host string `json:"host"`
	// Port is the port of remote server receiving telemetry data (exp: otel-collector endpoint port)
	Port int `json:"port"`
}

// Collector implements the traces exporter.
type Collector struct {
	exporter sdktrace.SpanExporter
}

// SpanExporter implements the interface to export traces.
func (c *Collector) SpanExporter() sdktrace.SpanExporter {
	return c.exporter
}

// ParseConfig creates an Open Telemetry configuration.
func ParseConfig(in map[string]interface{}) (*CollectorConfig, error) {
	defaultConfig := CollectorConfig{
		Host: RemoteServerHost,
		Port: RemoteServerPort,
	}
	err := config.Parse(in, &defaultConfig)
	if err != nil {
		return nil, err
	}
	return &defaultConfig, nil
}

// CreateExporter creates an Open Telemetry exporter instance.
func CreateExporter(ctx context.Context, cfg map[string]interface{}) (interface{}, error) {
	otelCfg, err := ParseConfig(cfg)
	if err != nil {
		return nil, err
	}

	var exporter sdktrace.SpanExporter
	exporter, err = otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", otelCfg.Host, otelCfg.Port)))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = exporter.Shutdown(ctx)
	}()
	return &Collector{
		exporter: exporter,
	}, nil
}
