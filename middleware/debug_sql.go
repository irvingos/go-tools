package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/irvingos/go-tools/logx"
)

func DebugSQLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if debugSQL, err := strconv.ParseBool(c.GetHeader("X-Debug-SQL")); err == nil && debugSQL {
			ctx := logx.WithTraceSQL(c.Request.Context())
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
