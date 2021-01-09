package observability

import (
	"context"
	_ "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ExporterType int
const (
	StdoutExporter ExporterType = iota
)

type Exporter interface {
	Start() error
	Stop() error
}

type stdoutExporter struct {
	exporter *stdout.Exporter
	spanProcessor *sdktrace.BatchSpanProcessor
	traceProvider *sdktrace.TracerProvider
	pusher *push.Controller
}

func NewExporter(exporterType ExporterType) (*stdoutExporter, error) {
	var err error
	exporter := &stdoutExporter{}

	exporter.exporter, err = stdout.NewExporter([]stdout.Option{
		stdout.WithQuantiles([]float64{0.5, 0.9, 0.99}),
		stdout.WithPrettyPrint(),
	}...)

	if err != nil {
		return nil, err
	}

	exporter.spanProcessor = sdktrace.NewBatchSpanProcessor(exporter.exporter)
	exporter.traceProvider = sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(exporter.spanProcessor))

	exporter.pusher = push.New(
		basic.New(
			simple.NewWithExactDistribution(),
			exporter.exporter,
		),
		exporter.exporter,
	)

	otel.SetTracerProvider(exporter.traceProvider)
	otel.SetMeterProvider(exporter.pusher.MeterProvider())

	otel.SetTextMapPropagator(propagation.Baggage{})

	return exporter, nil
}

func (observer stdoutExporter) Start() error {
	observer.pusher.Start()
	return nil
}

func (observer stdoutExporter) Stop() error {
	_ = observer.traceProvider.Shutdown(context.Background())
	observer.pusher.Stop()
	return nil
}

