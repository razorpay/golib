// Package exporter defines the interfaces required to implement in
// order to add additional exporters.
package exporter

import (
	"context"
	"fmt"
	"sync"

	"github.com/razorpay/golib/opentelemetry/config"
	"github.com/razorpay/golib/opentelemetry/exporter/opentelemetry"
	"github.com/razorpay/golib/opentelemetry/exporter/prometheus"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// MetricReader is the interface required in order to export metrics.
type MetricReader interface {
	MetricReader() sdkmetric.Reader
}

// SpanExporter is the interface required in order to export traces.
type SpanExporter interface {
	SpanExporter() sdktrace.SpanExporter
}

// Factory is the function type to obtain exporters, that can
// implement [MetricReader], [SpanExporter] or both.
type Factory func(context.Context, map[string]interface{}) (interface{}, error)

var (
	exporterFactories map[config.ExporterKind]Factory

	mu           = new(sync.RWMutex)
	registerOnce = new(sync.Once)
)

// RegisterKnownFactories registers all known exporter factories
func RegisterKnownFactories() {
	registerOnce.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		exporterFactories = map[config.ExporterKind]Factory{
			prometheus.ExporterKey:    prometheus.CreateExporter,
			opentelemetry.ExporterKey: opentelemetry.CreateExporter,
		}
	})
}

// CreateInstances create instances for a given configuration.
func CreateInstances(ctx context.Context, cfg []config.Exporter) (map[string]MetricReader, map[string]SpanExporter, []error) {
	metricReaderMap := make(map[string]MetricReader)
	spanExporterMap := make(map[string]SpanExporter)
	var errList []error

	mu.RLock()
	defer mu.RUnlock()
	uniqueNames := map[string]bool{}

	for idx, exporterCfg := range cfg {
		if uniqueNames[exporterCfg.Name] {
			err := fmt.Errorf("exporter with duplicate name: %s (at idx %d) ignored", exporterCfg.Name, idx)
			errList = append(errList, err)
			continue
		}
		f, ok := exporterFactories[exporterCfg.Kind]
		if !ok {
			err := fmt.Errorf("exporter %s of kind: %s (at idx %d) not found", exporterCfg.Name, exporterCfg.Kind, idx)
			errList = append(errList, err)
			continue
		}
		exporterInstance, err := f(ctx, exporterCfg.Config)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		if exporterInstance == nil {
			err := fmt.Errorf("implementation of kind: %s (at idx %d) creates nil instance", exporterCfg.Kind, idx)
			errList = append(errList, err)
		}

		uniqueNames[exporterCfg.Name] = true
		if spanExporter, ok := exporterInstance.(SpanExporter); ok && spanExporter != nil {
			spanExporterMap[exporterCfg.Name] = spanExporter
		} else if metricReader, ok := exporterInstance.(MetricReader); ok && metricReader != nil {
			metricReaderMap[exporterCfg.Name] = metricReader
		} else {
			errList = append(errList, fmt.Errorf("kind %s (at idx %d) is not a exporter", exporterCfg.Kind, idx))
		}
	}
	return metricReaderMap, spanExporterMap, errList
}
