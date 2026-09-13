package tracing

import (
	"context"
	"fmt"

	"github.com/mahdipeydai/taskmanager-go/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func Tracer(cfg *config.Config) trace.Tracer {
	return otel.Tracer(cfg.Jaeger.App)
}

func Init(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	jaegerEndpoint := fmt.Sprintf("%s:%d", cfg.Jaeger.Host, cfg.Jaeger.Port)

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(jaegerEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
