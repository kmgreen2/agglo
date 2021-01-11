package observability

import (
	"context"
	"go.opentelemetry.io/otel"
	_ "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ExporterType int
const (
	StdoutExporter ExporterType = iota
	ZipkinExporter
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

func NewStdoutExporter() (*stdoutExporter, error) {
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

func (exporter stdoutExporter) Start() error {
	exporter.pusher.Start()
	return nil
}

func (exporter stdoutExporter) Stop() error {
	_ = exporter.traceProvider.Shutdown(context.Background())
	exporter.pusher.Stop()
	return nil
}

type zipkinExporter struct {
	url string
	serviceName string
	underlying *zipkin.Exporter
}

func NewZipkinExporter(url, serviceName string) *zipkinExporter {
	return &zipkinExporter{
		url: url,
		serviceName: serviceName,
	}
}

func (exporter zipkinExporter) Start() error {
	var err error
	exporter.underlying, err = zipkin.NewRawExporter(exporter.url,
		exporter.serviceName,
		zipkin.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter.underlying))

	otel.SetTracerProvider(tp)

	return nil
}

func (exporter zipkinExporter) Stop() error {
	if exporter.underlying != nil {
		return exporter.underlying.Shutdown(context.Background())
	}
	return nil
}
