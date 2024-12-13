# Opentelemetry
Package opentelemetry provides telemetry instrumentation support for applications
using Opentelemetry specification. This package owns the responsibility of
creating and destroying telemetry providers which sends data to the external
components via exporters provided in the configuration.

## Features
1. Instrumentation support for Tracing
2. Instrumentation support for Metrics

## Usage Examples

### Fetch the package as a dependency
```go
    go get -u github.com/razorpay/goutils/opentelemetry
```
### Parse and generate the [opentelemetry configuration] (opentelemetry/config/config.go)
The client is responsible for generating opentelemetry configuration. Client can have an example opentelemetry configuration
block like below in its configuration file:
```
{
    "service_name": "local_service",
    "exporters": [
        {
            "name": "local_prometheus",
            "kind": "prometheus",
            "config": {
                "port": 9092,
                "process_metrics": true,
                "go_metrics": true
            }
        },
        {
            "name": "local_tempo",
            "kind": "opentelemetry",
            "config": {
                "host": "localhost",
                "port": 4317,
            }
        },
        {
            "name": "local_jaeger",
            "kind": "opentelemetry",
            "config": {
                "host": "localhost",
                "port": 14268,
            }
        }
    ],
    "metrics": {
      "exporters": ["local_prometheus"]
    },
    "traces": {
      "exporters": ["local_tempo", "local_jaeger"],
      "sample_rate": 1
    }
}
```

### Initialize the instrumentation providers
After generating the above configuration for opentelemetry, initialize the instrumentation providers like below:
```go
    err := opentelemetry.Register(context, opentelemetry)
```

### Instrument application for Tracing
```go
      tracer := otel.Tracer("test")
      commonAttrs := []attribute.KeyValue{
        attribute.String("attrA", "A"),
        attribute.String("attrB", "B"),
        attribute.String("attrC", "C"),
      }
      ctx, span := tracer.Start(
        ctx,
        "test-example",
        trace.WithAttributes(commonAttrs...))

	  // some operation

      span.End()
```
### Instrument application for Metrics
```go
        meter := otel.Meter("test-example")
	opt := api.WithAttributes(
		attribute.Key("attrA").String("A"),
		attribute.Key("attrB").String("B"),
	)
	counter, err := meter.Float64Counter("foo")
	counter.Add(context, 5, opt)
	histogram, err := meter.Float64Histogram(
		"baz",
		api.WithDescription("a histogram with custom buckets and rename"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	histogram.Record(context, 136, opt)
```

For more details, refer [opentelemetry/example/main.go] (example instrumentation file)

### Viewing example telemetry data
Run ``` make run-example  ``` to run a sample instrumentation 'test-service' application and to set up Prometheus, Otel-collector and Jaeger which will collect metrics and tracing data
of the application.
1. To view traces, use jaeger endpoint ``` localhost:16686  ```.
2. To view metrics, use prometheus endpoint ``` localhost:9090  ```.
3. Run ``` make obs-stack-down ``` to bring down the observability stack.

### Viewing telemetry data for any application
Run ``` make obs-stack  ``` to setup Prometheus, Otel-collector and Jaeger which will collect metrics and tracing data
of the application which are exporting data to them.
1. To view traces, use jaeger endpoint ``` localhost:16686  ```.
2. To view metrics, use prometheus endpoint ``` localhost:9090  ```.
3. Run ``` make obs-stack-down ``` to bring down the observability stack.

## TODOs
- Support for push based configuration for exporting metrics
- Allow different formats for the propagation of the trace (the `TextMapPropagator`)
- Support for exporting Metrics and Logging using Opentelemetry Exporter

