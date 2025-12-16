package tx

import (
	"context"

	"gorm.io/gorm"
)

func NewGormUow(db *gorm.DB) Uow {
	return &gormUow{
		db: db,
	}
}

type gormUow struct {
	db *gorm.DB
}

// Do implements Uow.
func (u *gormUow) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithGormTx(ctx, tx))
	})
}
