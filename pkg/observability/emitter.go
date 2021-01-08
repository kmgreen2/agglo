package observability

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type MetricType int
const (
	Int64Counter MetricType = iota
	Float64Counter
	Int64Recorder
	Float64Recorder
)

type Numeric float64

type Emitter struct {
	tracer trace.Tracer
	meter metric.Meter
	int64Counters map[string]metric.Int64Counter
	float64Counters map[string]metric.Float64Counter
	int64Recorders map[string]metric.Int64ValueRecorder
	float64Recorders map[string]metric.Float64ValueRecorder
}

func NewEmitter(name string) *Emitter {
	return &Emitter{
		otel.Tracer(name),
		otel.Meter(name),
		make(map[string]metric.Int64Counter),
		make(map[string]metric.Float64Counter),
		make(map[string]metric.Int64ValueRecorder),
		make(map[string]metric.Float64ValueRecorder),
	}
}

func (e Emitter) CreateSpan(ctx context.Context, name string, opts ...trace.SpanOption) (context.Context, trace.Span) {
	return e.tracer.Start(ctx, name, opts...)
}

func (e Emitter) AddMetric(name string, metricType MetricType, opts ...metric.InstrumentOption) {
	switch metricType{
	case Float64Counter:
		e.float64Counters[name] = metric.Must(e.meter).NewFloat64Counter(name, opts...)
	case Int64Counter:
		e.int64Counters[name] = metric.Must(e.meter).NewInt64Counter(name, opts...)
	case Float64Recorder:
		e.float64Recorders[name] = metric.Must(e.meter).NewFloat64ValueRecorder(name, opts...)
	case Int64Recorder:
		e.int64Recorders[name] = metric.Must(e.meter).NewInt64ValueRecorder(name, opts...)
	}
}

func (e Emitter) Emit(ctx context.Context, name string, metricType MetricType, value Numeric, labels ...label.KeyValue) {
	switch metricType {
	case Float64Counter:
		e.float64Counters[name].Add(ctx, float64(value), labels...)
	case Int64Counter:
		e.int64Counters[name].Add(ctx, int64(value), labels...)
	case Float64Recorder:
		e.float64Recorders[name].Record(ctx, float64(value), labels...)
	case Int64Recorder:
		e.int64Recorders[name].Record(ctx, int64(value), labels...)
	}
}




