package prometheus

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigFromInterface(t *testing.T) {
	cfg := map[string]interface{}{
		"port":            9092,
		"process_metrics": true,
		"go_metrics":      true,
	}
	collectorConfig, err := ParseConfig(cfg)
	require.NoError(t, err)
	duration := 3000
	expectedConfig := &CollectorConfig{
		Port:                    9092,
		ProcessMetrics:          true,
		GoMetrics:               true,
		ReadTimeoutInMillis:     &duration,
		WriteTimeoutInMillis:    &duration,
		DisableUnitSuffix:       false,
		OpenCensusBridgeEnabled: false,
	}
	require.Equal(t, expectedConfig, collectorConfig)
	readTimeout := 1000
	writeTimeout := 2000
	cfg = map[string]interface{}{
		"port":                      9091,
		"process_metrics":           true,
		"go_metrics":                true,
		"read_timeout_in_millis":    1000,
		"write_timeout_in_millis":   2000,
		"opencensus_bridge_enabled": true,
		"disable_unit_suffix":       true,
	}
	collectorConfig, err = ParseConfig(cfg)
	require.NoError(t, err)
	expectedConfig = &CollectorConfig{
		Port:                    9091,
		ProcessMetrics:          true,
		GoMetrics:               true,
		ReadTimeoutInMillis:     &readTimeout,
		WriteTimeoutInMillis:    &writeTimeout,
		OpenCensusBridgeEnabled: true,
		DisableUnitSuffix:       true,
	}
	require.Equal(t, expectedConfig, collectorConfig)
}

func TestExporter(t *testing.T) {
	cfg := map[string]interface{}{
		"port":                    9090,
		"process_metrics":         true,
		"go_metrics":              true,
		"read_timeout_in_millis":  1000,
		"write_timeout_in_millis": 2000,
	}
	ctx, cancel := context.WithCancel(context.Background())
	exporterInstance, err := CreateExporter(ctx, cfg)
	require.NoError(t, err)
	_, ok := exporterInstance.(*Collector)
	require.True(t, ok)
	cancel()
	time.Sleep(time.Millisecond * 200)
}
