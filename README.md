# OpenTelemetry Metrics Demo

This demo showcases OpenTelemetry metrics implementation in Go with three types of metrics: Counter, Gauge, and Histogram. It includes an OpenTelemetry Collector setup to receive and export metrics to Prometheus.

## Metrics Types Demonstrated

### 1. Counter (`requests_total`)
- **Type**: Monotonic counter
- **Description**: Tracks total number of requests
- **Labels**: `method` (GET, POST, PUT, DELETE), `status` (200, 404, 500)

### 2. UpDownCounter (`cpu_usage`)
- **Type**: UpDownCounter (used as gauge alternative)
- **Description**: Current CPU usage percentage
- **Labels**: `host` (demo-host)

### 3. Histogram (`request_duration`)
- **Type**: Distribution of values
- **Description**: Request duration in milliseconds
- **Labels**: `endpoint` (/api/users, /api/orders, /api/products)
- **Buckets**: 10, 50, 100, 200, 500, 1000, 2000ms

## Architecture

```
Go Application ---> OpenTelemetry Collector ---> Prometheus
```

## Prerequisites

- Go 1.21+
- Docker and Docker Compose

## Quick Start

### 1. Start the OpenTelemetry Collector and Prometheus

```bash
docker-compose up -d
```

### 2. Install Go dependencies

```bash
go mod tidy
```

### 3. Run the demo application

**Default (console exporter):**
```bash
go run main.go
```

**With OTLP gRPC exporter:**
```bash
go run main.go -otlp-grpc
```

**With OTLP HTTP exporter:**
```bash
go run main.go -otlp-http
```

**With Prometheus exporter:**
```bash
go run main.go -prometheus
```

## What You'll See

The application will:
1. Initialize OpenTelemetry with console exporter (default), OTLP gRPC exporter (with `-otlp-grpc` flag), OTLP HTTP exporter (with `-otlp-http` flag), or Prometheus exporter (with `-prometheus` flag)
2. Create three metric instruments (counter, gauge, histogram)
3. Generate sample metrics every 2 seconds for 100 iterations
4. Export metrics to console, OpenTelemetry Collector, or Prometheus endpoint

## Viewing Metrics

### Console Output
Metrics are printed to stdout when using the default console exporter.

### Prometheus Exporter Endpoint
When using `-prometheus` flag, metrics are available at:
- http://localhost:2112/metrics

### OpenTelemetry Collector Logs
```bash
docker-compose logs otel-collector
```

### Prometheus UI (via Collector)
Open http://localhost:9090 and query:
- `requests_total` - Counter metrics
- `cpu_usage` - Gauge metrics
- `request_duration` - Histogram metrics

## Configuration Files

- `otel-collector.yaml` - OpenTelemetry Collector configuration
- `prometheus.yml` - Prometheus scraping configuration
- `docker-compose.yaml` - Container orchestration

## Key Implementation Details

### OTLP Exporter Configuration
```go
exporter, err := otlpmetricgrpc.New(ctx,
    otlpmetricgrpc.WithEndpoint("localhost:4317"),
    otlpmetricgrpc.WithInsecure(),
)
```

### Metric Instruments Creation
```go
counter, _ := meter.Int64Counter("requests_total")
gauge, _ := meter.Float64UpDownCounter("cpu_usage") 
histogram, _ := meter.Float64Histogram("request_duration",
    metric.WithExplicitBucketBoundaries(10, 50, 100, 200, 500, 1000, 2000),
)
```

### Recording Metrics with Attributes
```go
counter.Add(ctx, 1, metric.WithAttributes(
    attribute.String("method", "GET"),
    attribute.String("status", "200"),
))
```

## Cleanup

```bash
docker-compose down
```

## Troubleshooting

- Ensure ports 2112 (Prometheus exporter), 4317, 4318, 8889, and 9090 are available
- Check collector logs if metrics aren't appearing
- Verify Go module dependencies with `go mod tidy`