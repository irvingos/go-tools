package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID int) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(userIDKey{}, userID)
		return gCtx
	}
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserIDFrom(ctx context.Context) int {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return gCtx.GetInt(userIDKey{})
	}
	userID, ok := ctx.Value(userIDKey{}).(int)
	if !ok {
		return 0
	}
	return userID
}
