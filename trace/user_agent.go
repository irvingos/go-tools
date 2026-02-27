package trace

import (
	"context"

	"github.com/gin-gonic/gin"
)

type userAgentKey struct{}

func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(userAgentKey{}, userAgent)
		return gCtx
	}
	return context.WithValue(ctx, userAgentKey{}, userAgent)
}

func UserAgentFrom(ctx context.Context) string {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return gCtx.GetString(userAgentKey{})
	}
	if v, ok := ctx.Value(userAgentKey{}).(string); ok {
		return v
	}
	return ""
}
