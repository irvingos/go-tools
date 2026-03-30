package auth

import (
	"context"
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserIDFrom(ctx context.Context) int {
	userID, ok := ctx.Value(userIDKey{}).(int)
	if !ok {
		return 0
	}
	return userID
}
