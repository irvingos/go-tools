package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/irvingos/go-tools/auth"
	"github.com/irvingos/go-tools/consts"
	"github.com/irvingos/go-tools/logx"
	"github.com/irvingos/go-tools/resp"
	"github.com/irvingos/go-tools/trace"
)

func Log() gin.HandlerFunc {
	return func(g *gin.Context) {
		start := time.Now()
		traceID := g.GetHeader(consts.HTTP_HEADER_TRACE_ID)

		// fill traceID if need
		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx := g.Request.Context()
		ctx = trace.WithTraceID(ctx, traceID)
		ctx = trace.WithClientIP(ctx, g.ClientIP())
		ctx = trace.WithUserAgent(ctx, g.Request.UserAgent())
		g.Request = g.Request.WithContext(ctx)
		g.Header(consts.HTTP_HEADER_TRACE_ID, traceID)

		g.Next()

		latency := time.Since(start)

		// 注意：这里一定要再取一次，因为 g.Request 可能在 g.Next() 过程中被重新赋值
		ctx = g.Request.Context()
		logx.WithContext(ctx).
			WithField(logx.FieldTenantID, auth.TenantIDFrom(ctx)).
			WithField(logx.FieldUserID, auth.UserIDFrom(ctx)).
			WithField(logx.FieldUsername, auth.UsernameFrom(ctx)).
			WithField(logx.FieldOpenAPIClient, auth.OpenAPIClientFrom(ctx)).
			WithField(logx.FieldLatency, fmt.Sprintf("%.3fms", float64(latency.Nanoseconds())/1e6)).
			WithField(logx.FieldStatus, g.Writer.Status()).
			WithField(logx.FieldCode, resp.CodeFrom(ctx)).
			WithField(logx.FieldRemoteIP, g.ClientIP()).
			WithField(logx.FieldMethod, g.Request.Method).
			WithField(logx.FieldPath, g.Request.URL.Path).
			WithField(logx.FieldQuery, g.Request.URL.RawQuery).
			WithField(logx.FieldUA, g.Request.UserAgent()).
			Info()
	}
}
