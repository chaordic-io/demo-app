package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chaordic-io/demo-app/internal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func main() {
	registry := prometheus.NewRegistry()

	requestHits := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "demo",
		Subsystem: "app",
		Name:      "request",
		Help:      "The requests",
	}, []string{"userAgent"})

	registry.Register(requestHits)

	handler := promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
	ctx := context.Background()
	tp, err := tracerProvider(ctx, "10.0.0.5:4317")
	if err != nil {
		log.Fatal(err)
	}

	defer tp.Shutdown(ctx)

	go http.ListenAndServe(":8080", http.HandlerFunc(HelloServer(requestHits, tp)))
	http.ListenAndServe(":8090", handler)
}

func tracerProvider(ctx context.Context, host string) (*tracesdk.TracerProvider, error) {

	exporter, err := otlptrace.New(ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(host),
			otlptracegrpc.WithInsecure(),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	}

	provider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("demo-app"),
			attribute.String("environment", "prod"),
			attribute.Int64("ID", 1),
		)),
	)

	return provider, nil
}

func HelloServer(counter *prometheus.CounterVec, tp *tracesdk.TracerProvider) func(w http.ResponseWriter, r *http.Request) {
	logger, _ := zap.NewProduction()
	logger = logger.With(
		zap.String("service", "demo-app"),
		zap.String("version", internal.Version),
		zap.String("namespace", os.Getenv("NOMAD_NAMESPACE")),
		zap.String("task_name", os.Getenv("NOMAD_TASK_NAME")),
		zap.String("alloc_id", os.Getenv("NOMAD_ALLOC_ID")),
		zap.String("job_name", os.Getenv("NOMAD_JOB_NAME")),
		zap.String("namespace", os.Getenv("NOMAD_NAMESPACE")),
		zap.String("dc", os.Getenv("NOMAD_DC")),
		zap.String("host", os.Getenv("NOMAD_HOST_IP_prometheus")),
	)
	defer logger.Sync()

	logger.Info("starting app",
		// Structured context as strongly typed Field values.
		zap.String("version", internal.Version),
		zap.String("go", internal.GoVersion),
		zap.String("platform", internal.Platform),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		tr := tp.Tracer("component-main")
		ctx, span := tr.Start(r.Context(), "foo")
		defer span.End()
		counter.WithLabelValues(r.Header.Get("User-Agent")).Inc()
		logger.Info("requested URL",
			// Structured context as strongly typed Field values.
			zap.String("url", r.URL.Path[1:]),
			zap.String("trace_id", trace.SpanFromContext(ctx).SpanContext().TraceID().String()),
		)
		backendThing(ctx, tr)
		fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])

	}
}

func backendThing(ctx context.Context, tr trace.Tracer) {
	_, span := tr.Start(ctx, "bar")
	span.SetAttributes(attribute.Key("testset").String("value"))
	defer span.End()
	time.Sleep(50 * time.Millisecond)
}
