package gormx

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

func IsUniqueViolationConstraint(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == constraint
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// Deprecated: Use IsUniqueViolation instead
func IsDuplicateKeyError(err error) bool {
	return IsUniqueViolation(err)
}

// Deprecated: Use IsRecordNotFound instead
func IsRecordNotFoundError(err error) bool {
	return IsRecordNotFound(err)
}
