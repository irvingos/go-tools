package auth

import (
	"context"
)

type usernameKey struct{}

func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

func UsernameFrom(ctx context.Context) string {
	username, ok := ctx.Value(usernameKey{}).(string)
	if !ok {
		return ""
	}
	return username
}
