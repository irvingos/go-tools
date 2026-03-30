package tx

import (
	"context"

	"gorm.io/gorm"
)

type key struct{}

func WithGormTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, key{}, tx)
}

func GormTxFrom(ctx context.Context) (*gorm.DB, bool) {
	v := ctx.Value(key{})
	tx, ok := v.(*gorm.DB)
	return tx, ok
}
