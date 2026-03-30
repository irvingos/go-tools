package trace

import (
	"context"
)

type userAgentKey struct{}

func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey{}, userAgent)
}

func UserAgentFrom(ctx context.Context) string {
	if v, ok := ctx.Value(userAgentKey{}).(string); ok {
		return v
	}
	return ""
}
