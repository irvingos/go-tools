package tx

import (
	"context"

	"gorm.io/gorm"
)

func NewBaseRepo(db *gorm.DB) BaseRepo {
	return BaseRepo{db: db}
}

type BaseRepo struct {
	db *gorm.DB
}

func (b BaseRepo) DBFrom(ctx context.Context) *gorm.DB {
	if t, ok := GormTxFrom(ctx); ok {
		return t
	}
	return b.db
}
