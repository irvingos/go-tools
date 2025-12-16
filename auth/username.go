package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

type usernameKey struct{}

func WithUsername(ctx context.Context, username string) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return WithUsername(gCtx.Request.Context(), username)
	}
	return context.WithValue(ctx, usernameKey{}, username)
}

func UsernameFrom(ctx context.Context) string {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return UsernameFrom(gCtx.Request.Context())
	}
	username, ok := ctx.Value(usernameKey{}).(string)
	if !ok {
		return ""
	}
	return username
}
