package logx

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ctxKey string

const traceKey ctxKey = "trace_id"

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return context.WithValue(gCtx.Request.Context(), traceKey, traceID)
	}
	return context.WithValue(ctx, traceKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if gCtx, ok := ctx.(*gin.Context); ok {
		return TraceIDFromContext(gCtx.Request.Context())
	}
	if v, ok := ctx.Value(traceKey).(string); ok {
		return v
	}
	return ""
}
