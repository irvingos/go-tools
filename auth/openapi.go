package auth

import "context"

type openapiClientKey struct{}

func WithOpenAPIClient(ctx context.Context, clientName string) context.Context {
	return context.WithValue(ctx, openapiClientKey{}, clientName)
}

func OpenAPIClientFrom(ctx context.Context) string {
	v, ok := ctx.Value(openapiClientKey{}).(string)
	if !ok {
		return ""
	}
	return v
}
