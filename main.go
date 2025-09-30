package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
	useGrpcExporter := flag.Bool("otlp-grpc", false, "Use OTLP gRPC exporter instead of console exporter")
	flag.Parse()

	ctx := context.Background()

	// Initialize OpenTelemetry
	shutdown, err := initOTel(ctx, *useGrpcExporter)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer shutdown()

	// Get meter
	meter := otel.Meter("otel-demo")

	// Create metrics instruments
	counter, err := meter.Int64Counter("requests_total", metric.WithDescription("Total number of requests"))
	if err != nil {
		log.Fatalf("Failed to create counter: %v", err)
	}

	gauge, err := meter.Float64UpDownCounter("cpu_usage", metric.WithDescription("Current CPU usage percentage"))
	if err != nil {
		log.Fatalf("Failed to create gauge: %v", err)
	}

	histogram, err := meter.Float64Histogram("request_duration", 
		metric.WithDescription("Request duration in milliseconds"),
		metric.WithExplicitBucketBoundaries(10, 50, 100, 200, 500, 1000, 2000),
	)
	if err != nil {
		log.Fatalf("Failed to create histogram: %v", err)
	}

	fmt.Println("OpenTelemetry Metrics Demo Started")
	fmt.Println("Generating metrics... Press Ctrl+C to stop")

	// Generate metrics continuously
	for i := 0; i < 100; i++ {
		// Counter: Increment request count
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", randomMethod()),
			attribute.String("status", randomStatus()),
		))

		// Gauge: Set current CPU usage (using UpDownCounter as gauge alternative)
		cpuUsage := rand.Float64() * 100
		gauge.Add(ctx, cpuUsage, metric.WithAttributes(
			attribute.String("host", "demo-host"),
		))

		// Histogram: Record request duration
		duration := rand.Float64() * 1000 // 0-1000ms
		histogram.Record(ctx, duration, metric.WithAttributes(
			attribute.String("endpoint", randomEndpoint()),
		))

		fmt.Printf("Iteration %d: Counter +1, Gauge %.2f%%, Histogram %.2fms\n", i+1, cpuUsage, duration)
		time.Sleep(2 * time.Second)
	}

	fmt.Println("Demo completed")
}

func initOTel(ctx context.Context, useGrpcExporter bool) (func(), error) {
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("otel-demo"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on flag
	var exporter sdkmetric.Exporter

	if useGrpcExporter {
		exporter, err = otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint("127.0.0.1:4317"),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		fmt.Println("Using OTLP gRPC exporter")
	} else {
		exporter, err = stdoutmetric.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create console exporter: %w", err)
		}
		fmt.Println("Using console exporter")
	}

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(3*time.Second))),
	)

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	return func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}, nil
}

func randomMethod() string {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return methods[rand.Intn(len(methods))]
}

func randomStatus() string {
	statuses := []string{"200", "404", "500"}
	return statuses[rand.Intn(len(statuses))]
}

func randomEndpoint() string {
	endpoints := []string{"/api/users", "/api/orders", "/api/products"}
	return endpoints[rand.Intn(len(endpoints))]
}
