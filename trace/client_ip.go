package trace

import (
	"context"

	"github.com/gin-gonic/gin"
)

type clientIPKey struct{}

func WithClientIP(ctx context.Context, clientIP string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(clientIPKey{}, clientIP)
		return gCtx
	}
	return context.WithValue(ctx, clientIPKey{}, clientIP)
}

func ClientIPFrom(ctx context.Context) string {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return gCtx.GetString(clientIPKey{})
	}
	if v, ok := ctx.Value(clientIPKey{}).(string); ok {
		return v
	}
	return ""
}
