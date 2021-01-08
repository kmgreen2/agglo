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

type Observer interface {
	Start() error
	Stop() error
}

type stdoutObserver struct {
	exporter *stdout.Exporter
	spanProcessor *sdktrace.BatchSpanProcessor
	traceProvider *sdktrace.TracerProvider
	pusher *push.Controller
}

func NewObserver(exporterType ExporterType) (*stdoutObserver, error) {
	var err error
	observer := &stdoutObserver{}

	observer.exporter, err = stdout.NewExporter([]stdout.Option{
		stdout.WithQuantiles([]float64{0.5, 0.9, 0.99}),
		stdout.WithPrettyPrint(),
	}...)

	if err != nil {
		return nil, err
	}

	observer.spanProcessor = sdktrace.NewBatchSpanProcessor(observer.exporter)
	observer.traceProvider = sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(observer.spanProcessor))

	observer.pusher = push.New(
		basic.New(
			simple.NewWithExactDistribution(),
			observer.exporter,
		),
		observer.exporter,
	)

	otel.SetTracerProvider(observer.traceProvider)
	otel.SetMeterProvider(observer.pusher.MeterProvider())

	otel.SetTextMapPropagator(propagation.Baggage{})

	return observer, nil
}

func (observer stdoutObserver) Start() error {
	observer.pusher.Start()
	return nil
}

func (observer stdoutObserver) Stop() error {
	_ = observer.traceProvider.Shutdown(context.Background())
	observer.pusher.Stop()
	return nil
}

