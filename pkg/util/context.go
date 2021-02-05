package util

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

type ContextKey string
const (
	ParentContext      ContextKey = "agglo.io/parentContext"
	SpanContext        ContextKey = "agglo.io/spanContext"
	DistributedLockKey            = "agglo.io/distributedLockKey"
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

func InjectDistributedLockIndex(ctx context.Context, index int) context.Context {
	return context.WithValue(ctx, DistributedLockKey, index)
}

func ExtractDistributedLockIndex(ctx context.Context) int {
	if value := ctx.Value(DistributedLockKey); value != nil {
		if idx, ok := value.(int); ok {
			return idx
		}
	}
	return -1
}

