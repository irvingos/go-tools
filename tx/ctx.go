package tx

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type key struct{}

func WithGormTx(ctx context.Context, tx *gorm.DB) context.Context {
	if gCtx, ok := ctx.(*gin.Context); ok {
		gCtx.Set(key{}, tx)
		return gCtx
	}
	return context.WithValue(ctx, key{}, tx)
}

func GormTxFrom(ctx context.Context) (*gorm.DB, bool) {
	if gCtx, ok := ctx.(*gin.Context); ok {
		v, exists := gCtx.Get(key{})
		if !exists {
			return nil, false
		}
		tx, ok := v.(*gorm.DB)
		return tx, ok
	}
	v := ctx.Value(key{})
	tx, ok := v.(*gorm.DB)
	return tx, ok
}
