package gormx

import (
	"github.com/irvingos/go-tools/consts"
	"github.com/irvingos/go-tools/errorx"
	"gorm.io/gen/field"
)

var (
	ErrUnknownSortByField = errorx.NewError(1004200, "unknown sort_by field")
)

type Model interface {
	GetFieldByName(fieldName string) (field.OrderExpr, bool)
}

type Do[T any] interface {
	Order(...field.Expr) T
}

func ApplySort[T Do[T]](model Model, do T, sortBy, order string) (new T, err error) {
	if sortBy == "" {
		return do, nil
	}

	sortField, ok := model.GetFieldByName(sortBy)
	if !ok {
		return do, ErrUnknownSortByField
	}

	if order == consts.GORM_ASC {
		new = do.Order(sortField.Asc())
	} else {
		new = do.Order(sortField.Desc())
	}

	return
}
