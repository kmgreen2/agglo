package common

import (
	"context"
	"fmt"
	"time"
)

type ContextKey string
const (
	ProcessSpan      ContextKey = "agglo.io/processSpan"
	ProcessStartTime            = "agglo.io/processStartTime"
)

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

