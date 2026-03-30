package resp

import (
	"context"
)

type codeKey struct{}

func WithCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, codeKey{}, code)
}

func CodeFrom(ctx context.Context) int {
	if code, ok := ctx.Value(codeKey{}).(int); ok {
		return code
	}
	return 0
}
