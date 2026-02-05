package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

type tenantIDKey struct{}

func WithTenantID(ctx context.Context, tenantID int) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(tenantIDKey{}, tenantID)
		return gCtx
	}
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

func TenantIDFrom(ctx context.Context) int {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return gCtx.GetInt(tenantIDKey{})
	}
	tenantID, ok := ctx.Value(tenantIDKey{}).(int)
	if !ok {
		return 0
	}
	return tenantID
}
