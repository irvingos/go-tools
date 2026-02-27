package trace

import (
	"context"

	"github.com/gin-gonic/gin"
)

type traceKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(traceKey{}, traceID)
		return gCtx
	}
	return context.WithValue(ctx, traceKey{}, traceID)
}

func TraceIDFrom(ctx context.Context) string {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return gCtx.GetString(traceKey{})
	}
	if v, ok := ctx.Value(traceKey{}).(string); ok {
		return v
	}
	return ""
}
