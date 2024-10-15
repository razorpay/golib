// Package prometheus implements a Prometheus metrics exporter.
package prometheus

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"net/http"
	"time"

	"github.com/razorpay/golib/opentelemetry/config"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rs/zerolog/log"

	prom "github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	// ExporterKey is the name for the prometheus exporter
	ExporterKey    = config.ExporterKind("prometheus")
	ReadTimeoutMs  = 3000
	WriteTimeoutMs = 3000
	LocalPort      = 9091
)

// CollectorConfig has the variables to configure the
// prometheus exporter.
type CollectorConfig struct {
	// Port is the port on which /metrics endpoint will be exposed
	Port                    int  `json:"port"`
	ProcessMetrics          bool `json:"process_metrics"`
	GoMetrics               bool `json:"go_metrics"`
	ReadTimeoutInMillis     *int `json:"read_timeout_in_millis"`
	WriteTimeoutInMillis    *int `json:"write_timeout_in_millis"`
	OpenCensusBridgeEnabled bool `json:"opencensus_bridge_enabled"`
	DisableUnitSuffix       bool `json:"disable_unit_suffix"`
}

// Collector implements the metrics exporter
type Collector struct {
	registry *prom.Registry
	exporter *prometheus.Exporter
}

// MetricReader implements the interface to exporte metrics.
func (c *Collector) MetricReader() sdkmetric.Reader {
	return c.exporter
}

// ParseConfig creates a Prometheus configuration.
func ParseConfig(in map[string]interface{}) (*CollectorConfig, error) {
	defaultReadTimeout := ReadTimeoutMs
	defaultWriteTimeout := WriteTimeoutMs
	defaultCfg := CollectorConfig{
		Port:                 LocalPort,
		ProcessMetrics:       true,
		GoMetrics:            true,
		ReadTimeoutInMillis:  &defaultReadTimeout,
		WriteTimeoutInMillis: &defaultWriteTimeout,
	}
	err := config.Parse(in, &defaultCfg)
	if err != nil {
		return nil, err
	}
	return &defaultCfg, nil
}

// CreateExporter creates a Prometheus exporter instance.
func CreateExporter(ctx context.Context, cfg map[string]interface{}) (interface{}, error) {
	promCfg, err := ParseConfig(cfg)
	if err != nil {
		return nil, err
	}

	prometheusRegistry := prom.NewRegistry()

	if promCfg.ProcessMetrics {
		err := prometheusRegistry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		if err != nil {
			return nil, err
		}
	}

	if promCfg.GoMetrics {
		err = prometheusRegistry.Register(collectors.NewGoCollector())
		if err != nil {
			return nil, err
		}
	}

	opts := []prometheus.Option{prometheus.WithRegisterer(prometheusRegistry)}
	if promCfg.OpenCensusBridgeEnabled {
		opts = append(opts, prometheus.WithProducer(opencensus.NewMetricProducer()))
	}
	if promCfg.DisableUnitSuffix {
		opts = append(opts, prometheus.WithoutUnits())
		opts = append(opts, prometheus.WithoutCounterSuffixes())
	}
	exporter, err := prometheus.New(opts...)
	if err != nil {
		return nil, err
	}

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(prometheusRegistry,
		promhttp.HandlerOpts{}))
	server := http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", promCfg.Port),
		ReadTimeout:  time.Duration(*promCfg.ReadTimeoutInMillis) * time.Millisecond,
		WriteTimeout: time.Duration(*promCfg.WriteTimeoutInMillis) * time.Millisecond,
	}

	go func() {
		if serverErr := server.ListenAndServe(); !errors.Is(serverErr, http.ErrServerClosed) {
			log.Fatal().Str("SERVICE", "prometheus").Msgf("The Prometheus exporter failed to listen and serve: %v", serverErr)
		}
	}()

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	return &Collector{
		registry: prometheusRegistry,
		exporter: exporter,
	}, nil
}
