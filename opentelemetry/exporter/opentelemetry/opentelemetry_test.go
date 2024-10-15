package opentelemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfigFromInterface(t *testing.T) {
	cfg := map[string]interface{}{
		"port": 4317,
		"host": "localhost1",
	}
	collectorConfig, err := ParseConfig(cfg)
	require.NoError(t, err)
	expectedConfig := &CollectorConfig{
		Port: 4317,
		Host: "localhost1",
	}
	require.Equal(t, expectedConfig, collectorConfig)

	cfg = map[string]interface{}{}
	collectorConfig, err = ParseConfig(cfg)
	require.NoError(t, err)
	expectedConfig = &CollectorConfig{
		Port: 4317,
		Host: "localhost",
	}
	require.Equal(t, expectedConfig, collectorConfig)
}

func TestExporter(t *testing.T) {
	cfg := map[string]interface{}{
		"port": 4317,
		"host": "localhost1",
	}
	ctx, cancel := context.WithCancel(context.Background())
	exporterInstance, err := CreateExporter(ctx, cfg)
	require.NoError(t, err)
	_, ok := exporterInstance.(*Collector)
	require.True(t, ok)
	cancel()
	time.Sleep(time.Millisecond * 200)
}
