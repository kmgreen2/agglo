package common

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type Key string
const (
	ParentContext Key = "agglo.io/parentContext"
	SpanContext Key = "agglo.io/spanContext"
	ProcessSpan Key = "agglo.io/processSpan"
	ProcessStartTime = "agglo.io/processStartTime"
	IntraProcessCheckpoint = "agglo.io/intraProcessCheckpoint"
)

func ExtractPubSubContext(payload []byte) context.Context {
	return context.Background()
}

func InjectPubSubContext(ctx context.Context, payload []byte) context.Context {
	return context.Background()
}

func InjectParentSpanContext(currCtx context.Context, parentSpanCtx trace.SpanContext) context.Context {
	return context.WithValue(currCtx, ParentContext, parentSpanCtx)
}

var EmptySpanContext = trace.SpanContext{}
func ExtractParentSpanContext(ctx context.Context) trace.SpanContext {
	if value := ctx.Value(ParentContext); value != nil {
		if parentCtx, ok := value.(trace.SpanContext); ok {
			return parentCtx
		}
	}
	return EmptySpanContext
}

func ExtractSpanContext(ctx context.Context) trace.SpanContext {
	if value := ctx.Value(SpanContext); value != nil {
		if spanCtx, ok := value.(trace.SpanContext); ok {
			return spanCtx
		}
	}
	return EmptySpanContext
}

func InjectSpanContext(ctx context.Context, spanCtx trace.SpanContext) context.Context {
	return context.WithValue(ctx, SpanContext, spanCtx)
}

func InjectProcessStartTime(processKey string, time time.Time, ctx context.Context) context.Context {
	contextKey := fmt.Sprintf("%s.%s", ProcessStartTime, processKey)
	return context.WithValue(ctx, contextKey, time)
}

var InvalidTime = time.Time{}
func ExtractProcessStartTime(processKey string, ctx context.Context) time.Time {
	contextKey := fmt.Sprintf("%s.%s", ProcessStartTime, processKey)
	if value := ctx.Value(contextKey); value != nil {
		if t, ok := value.(time.Time); ok {
			return t
		}
	}
	return InvalidTime
}

