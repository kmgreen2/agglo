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
	Int64Recorder
	Float64Recorder
	Float64Gauge
)

type Numeric float64

type Emitter struct {
	tracer trace.Tracer
	meter metric.Meter
	int64Counters map[string]metric.Int64Counter
	float64UpDownCounters map[string]metric.Float64UpDownCounter
	int64Recorders map[string]metric.Int64ValueRecorder
	float64Recorders map[string]metric.Float64ValueRecorder
}

func NewEmitter(name string) *Emitter {
	return &Emitter{
		otel.Tracer(name),
		otel.Meter(name),
		make(map[string]metric.Int64Counter),
		make(map[string]metric.Float64UpDownCounter),
		make(map[string]metric.Int64ValueRecorder),
		make(map[string]metric.Float64ValueRecorder),
	}
}

func (e Emitter) CreateSpan(ctx context.Context, name string, opts ...trace.SpanOption) (context.Context, trace.Span) {
	return e.tracer.Start(ctx, name, opts...)
}

func (e Emitter) AddMetric(name string, metricType MetricType, opts ...metric.InstrumentOption) {
	switch metricType{
	case Int64Counter:
		e.int64Counters[name] = metric.Must(e.meter).NewInt64Counter(name, opts...)
	case Float64Recorder:
		e.float64Recorders[name] = metric.Must(e.meter).NewFloat64ValueRecorder(name, opts...)
	case Int64Recorder:
		e.int64Recorders[name] = metric.Must(e.meter).NewInt64ValueRecorder(name, opts...)
	case Float64Gauge:
		e.float64UpDownCounters[name] = metric.Must(e.meter).NewFloat64UpDownCounter(name, opts...)
	}
}

func (e Emitter) AddInt64(name string, value int64, labels ...label.KeyValue) {
	e.AddInt64WithContext(context.Background(), name, value, labels...)
}

func (e Emitter) AddInt64WithContext(ctx context.Context, name string, value int64, labels ...label.KeyValue) {
	if counter, ok := e.int64Counters[name]; ok {
		counter.Add(ctx, value, labels...)
	}
}

func (e Emitter) RecordInt64(name string, value int64, labels ...label.KeyValue) {
	e.RecordInt64WithContext(context.Background(), name, value, labels...)
}

func (e Emitter) RecordInt64WithContext(ctx context.Context, name string, value int64, labels ...label.KeyValue) {
	if recorder, ok := e.int64Recorders[name]; ok {
		recorder.Record(ctx, value, labels...)
	}
}

func (e Emitter) RecordFloat64(name string, value float64, labels ...label.KeyValue) {
	e.RecordFloat64WithContext(context.Background(), name, value, labels...)
}

func (e Emitter) RecordFloat64WithContext(ctx context.Context, name string, value float64, labels ...label.KeyValue) {
	if recorder, ok := e.float64Recorders[name]; ok {
		recorder.Record(ctx, value, labels...)
	}
}

func (e Emitter) GaugeFloat64(name string, value float64, labels ...label.KeyValue) {
	e.GaugeFloat64WithContext(context.Background(), name, value, labels...)
}

func (e Emitter) GaugeFloat64WithContext(ctx context.Context, name string, value float64, labels ...label.KeyValue) {
	if gauge, ok := e.float64UpDownCounters[name]; ok {
		gauge.Add(ctx, value, labels...)
	}
}




