package observability

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ShutdownFunc func(context.Context) error

func Configure(ctx context.Context, serviceName, exporterKind, endpoint string) (ShutdownFunc, error) {
	exporter, err := newExporter(ctx, exporterKind, endpoint)
	if err != nil {
		return nil, err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			attribute.String("service.name", serviceName),
			attribute.String("service.version", "0.1.0"),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(500*time.Millisecond),
			sdktrace.WithExportTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return provider.Shutdown, nil
}

func newExporter(ctx context.Context, kind string, endpoint string) (sdktrace.SpanExporter, error) {
	switch strings.ToLower(kind) {
	case "", "stdout", "console":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlp", "grpc":
		return otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(normalizeOTLPEndpoint(endpoint)),
			otlptracegrpc.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("unsupported OpenTelemetry exporter %q", kind)
	}
}

func normalizeOTLPEndpoint(endpoint string) string {
	value := strings.TrimSpace(endpoint)
	if value == "" {
		return "127.0.0.1:4317"
	}

	parsed, err := url.Parse(value)
	if err == nil && parsed.Host != "" {
		return parsed.Host
	}

	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")

	return value
}
