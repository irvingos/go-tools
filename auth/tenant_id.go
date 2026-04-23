package auth

import (
	"context"
)

type tenantIDKey struct{}

func WithTenantID(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

func TenantIDFrom(ctx context.Context) int64 {
	tenantID, ok := ctx.Value(tenantIDKey{}).(int64)
	if !ok {
		return 0
	}
	return tenantID
}
