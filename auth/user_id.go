package auth

import (
	"context"
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserIDFrom(ctx context.Context) int64 {
	userID, ok := ctx.Value(userIDKey{}).(int64)
	if !ok {
		return 0
	}
	return userID
}
