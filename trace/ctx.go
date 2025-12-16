package trace

import (
	"context"

	"github.com/gin-gonic/gin"
)

type traceKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return WithTraceID(gCtx.Request.Context(), traceID)
	}
	return context.WithValue(ctx, traceKey{}, traceID)
}

func TraceIDFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if gCtx, ok := ctx.(*gin.Context); ok {
		return TraceIDFrom(gCtx.Request.Context())
	}
	if v, ok := ctx.Value(traceKey{}).(string); ok {
		return v
	}
	return ""
}
