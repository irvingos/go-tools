package resp

import (
	"context"

	"github.com/gin-gonic/gin"
)

type codeKey struct{}

func WithCode(ctx context.Context, code int) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(codeKey{}, code)
		return gCtx
	}
	return context.WithValue(ctx, codeKey{}, code)
}

func CodeFrom(ctx context.Context) int {
	if gCtx, ok := ctx.(*gin.Context); ok {
		return gCtx.GetInt(codeKey{})
	}
	if code, ok := ctx.Value(codeKey{}).(int); ok {
		return code
	}
	return 0
}
