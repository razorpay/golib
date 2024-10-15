package config

import (
	"errors"
)

var ErrNoExporter = errors.New("no exporters declared")
var ErrTelemetryConfigMissing = errors.New("config for telemetry instance is not provided")

// Config is the root configuration for the OTEL observability stack
type Config struct {
	ServiceName string `mapstructure:"service_name" json:"service_name"`
	Exporters   []Exporter
	Metrics     *MetricsConfig
	Trace       *TraceConfig
}

type ExporterKind string

// Exporter has the information to configure an exporter
// instance.
//
// The Kind is the name of the kind of exporter we want:
// OTEL, Prometheus, ...
//
// The Config is the configuration for this provider
type Exporter struct {
	Name   string
	Kind   ExporterKind
	Config map[string]interface{}
}

type MetricsConfig struct {
	Exporters []string
}

type TraceConfig struct {
	Exporters  []string
	SampleRate float64 `mapstructure:"sample_rate" json:"sample_rate"`
}

func Validate(cfg *Config) error {
	if cfg.Exporters == nil || len(cfg.Exporters) == 0 {
		return ErrNoExporter
	}
	if cfg.Metrics == nil && cfg.Trace == nil {
		return ErrTelemetryConfigMissing
	}
	return nil
}
