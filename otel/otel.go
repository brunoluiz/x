package otel

import (
	"context"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	shutdown := func(ctx context.Context) error {
		var shutdownErr error
		for _, fn := range shutdownFuncs {
			shutdownErr = errors.Join(shutdownErr, fn(ctx))
		}
		shutdownFuncs = nil
		return shutdownErr
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		return shutdown, errors.Join(err, shutdown(ctx))
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		return shutdown, errors.Join(err, shutdown(ctx))
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	switch exporter := os.Getenv("OTEL_TRACES_EXPORTER"); exporter {
	case "otlp":
		exp, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, err
		}

		return sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(resource.Default()),
		), nil
	default:
		return sdktrace.NewTracerProvider(), nil
	}
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	switch os.Getenv("OTEL_METRICS_EXPORTER") {
	case "console":
		exp, err := stdoutmetric.New()
		if err != nil {
			return nil, err
		}

		return metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(3*time.Second))),
		), nil
	case "otlp":
		exp, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, err
		}

		return metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(3*time.Second))),
		), nil
	default:
		prometheus, err := prometheus.New()
		if err != nil {
			return nil, err
		}

		return metric.NewMeterProvider(metric.WithReader(prometheus)), nil
	}
}
