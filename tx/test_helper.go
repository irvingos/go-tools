package tx

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func setupTestDBWithTable(t *testing.T) *gorm.DB {
	db := setupTestDB(t)

	// 創建測試表
	err := db.Exec("CREATE TABLE test_users (id INTEGER PRIMARY KEY, name TEXT)").Error
	require.NoError(t, err)

	return db
}
