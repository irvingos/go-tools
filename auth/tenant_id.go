package auth

import (
	"context"
)

type tenantIDKey struct{}

func WithTenantID(ctx context.Context, tenantID int) context.Context {
	return context.WithValue(ctx, tenantIDKey{}, tenantID)
}

func TenantIDFrom(ctx context.Context) int {
	tenantID, ok := ctx.Value(tenantIDKey{}).(int)
	if !ok {
		return 0
	}
	return tenantID
}
