package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/irvingos/go-tools/logx"
)

func DebugSQLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Debug-SQL") == "true" {
			ctx := logx.WithTraceSQL(c.Request.Context())
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
