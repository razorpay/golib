package opentelemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/razorpay/golib/opentelemetry/config"
	"github.com/razorpay/golib/opentelemetry/exporter"

	"go.opentelemetry.io/otel/metric"
	noopmetric "go.opentelemetry.io/otel/metric/noop"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Register all the known exporter factories (Opentelemetry, prometheus, etc.)
// and uses the provided Config to instantiate the configured exporters.
func Register(ctx context.Context, cfg *config.Config, views []sdkmetric.View) error {
	err := config.Validate(cfg)
	if err != nil {
		return err
	}
	exporter.RegisterKnownFactories()

	metricExporters, spanExporters, errs := exporter.CreateInstances(ctx, cfg.Exporters)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	// if we do not have any metrics exporter config but exporters to use, we default
	// to report to all configured exporters.
	if cfg.Metrics != nil && cfg.Metrics.Exporters == nil {
		cfg.Metrics = &config.MetricsConfig{
			Exporters: make([]string, 0, len(metricExporters)),
		}
		exporters := cfg.Metrics.Exporters
		for metricExporter := range metricExporters {
			exporters = append(exporters, metricExporter)
		}
	}

	// if we do not have any trace exporter config but exporters to use, we default
	// to report to all configured exporters.
	if cfg.Trace != nil && cfg.Trace.Exporters == nil {
		cfg.Trace = &config.TraceConfig{
			Exporters:  make([]string, 0, len(spanExporters)),
			SampleRate: 1,
		}
		exporters := cfg.Trace.Exporters
		for spanExporter := range spanExporters {
			exporters = append(exporters, spanExporter)
		}
	}

	res := sdkresource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName))
	if cfg.Trace != nil {
		err := initTraceProvider(ctx, res, cfg.Trace, spanExporters)
		if err != nil {
			return err
		}
	}
	if cfg.Metrics != nil {
		err := initMeterProvider(ctx, res, cfg.Metrics, metricExporters, views)
		if err != nil {
			return err
		}
	}
	otel.SetTextMapPropagator(prop)
	return nil
}

func initTraceProvider(ctx context.Context, resource *sdkresource.Resource, traceCfg *config.TraceConfig, spanExporters map[string]exporter.SpanExporter) error {
	traceOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(resource)}
	for _, exporterName := range traceCfg.Exporters {
		spanExporter, ok := spanExporters[exporterName]
		if !ok {
			return fmt.Errorf("span exporter: %s provided in trace config does not exist. (spanExporters: %#v)", exporterName, spanExporters)
		}
		//nolint:godox
		// Todo: Add support for exposing BatchSpanProcessorOption config.
		// Currently default exporter timeout, retry, max queue size configs are being used.
		traceOpts = append(traceOpts, sdktrace.WithBatcher(spanExporter.SpanExporter()))
	}

	var tracerProvider trace.TracerProvider = nooptrace.NewTracerProvider()
	var sdkTracerProvider *sdktrace.TracerProvider
	if len(traceOpts) > 0 {
		samplerOpt := sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(traceCfg.SampleRate)))
		traceOpts = append(traceOpts, samplerOpt)
		sdkTracerProvider = sdktrace.NewTracerProvider(traceOpts...)
		go func() {
			<-ctx.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = sdkTracerProvider.Shutdown(ctx)
			cancel()
		}()
		tracerProvider = sdkTracerProvider
	}
	otel.SetTracerProvider(tracerProvider)
	return nil
}

func initMeterProvider(ctx context.Context, resource *sdkresource.Resource, cfg *config.MetricsConfig, metricExporters map[string]exporter.MetricReader, views []sdkmetric.View) error {
	metricOpts := []sdkmetric.Option{sdkmetric.WithResource(resource)}
	if len(views) > 0 {
		metricOpts = append(metricOpts, sdkmetric.WithView(views...))
	}
	for _, exporterName := range cfg.Exporters {
		metricExporter, ok := metricExporters[exporterName]
		if !ok {
			return fmt.Errorf("metric exporter %s provided in metrics config does not exist. (metricExporters: %#v)", metricExporter, metricExporters)
		}
		//nolint:godox
		// TODO: pass reporting period for use-cases where metrics is pushed instead of scraped?
		metricOpts = append(metricOpts, sdkmetric.WithReader(metricExporter.MetricReader()))
	}

	var meterProvider metric.MeterProvider = noopmetric.NewMeterProvider()
	var sdkMetricProvider *sdkmetric.MeterProvider
	if len(metricOpts) > 0 {
		sdkMetricProvider = sdkmetric.NewMeterProvider(metricOpts...)
		go func() {
			<-ctx.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = sdkMetricProvider.Shutdown(ctx)
			cancel()
		}()
		meterProvider = sdkMetricProvider
	}
	otel.SetMeterProvider(meterProvider)
	return nil
}
