package tx

import (
	"context"
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGormUow_Do_StandardContext(t *testing.T) {
	db := setupTestDBWithTable(t)
	uow := NewGormUow(db)
	ctx := context.Background()

	// 測試正常執行
	err := uow.Do(ctx, func(ctx context.Context) error {
		// 驗證可以在回調中獲取 tx
		tx, ok := GormTxFrom(ctx)
		assert.True(t, ok)
		assert.NotNil(t, tx)

		// 驗證 tx 可以執行操作
		err := tx.Exec("INSERT INTO test_users (name) VALUES (?)", "test").Error
		assert.NoError(t, err)

		// 驗證在 transaction 中的操作
		var count int64
		tx.Raw("SELECT COUNT(*) FROM test_users WHERE name = ?", "test").Scan(&count)
		assert.Equal(t, int64(1), count)

		return nil
	})
	assert.NoError(t, err)

	// 驗證事務提交後數據存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM test_users WHERE name = ?", "test").Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestGormUow_Do_GinContext(t *testing.T) {
	db := setupTestDBWithTable(t)
	uow := NewGormUow(db)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// 測試使用 gin.Context
	err := uow.Do(c, func(ctx context.Context) error {
		// 驗證可以在回調中獲取 tx
		tx, ok := GormTxFrom(ctx)
		assert.True(t, ok)
		assert.NotNil(t, tx)

		// 驗證 tx 可以執行操作
		err := tx.Exec("INSERT INTO test_users (name) VALUES (?)", "test_gin").Error
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)

	// 驗證事務提交後數據存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM test_users WHERE name = ?", "test_gin").Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestGormUow_Do_RollbackOnError(t *testing.T) {
	db := setupTestDBWithTable(t)
	uow := NewGormUow(db)
	ctx := context.Background()

	testErr := errors.New("test error")

	// 測試錯誤時回滾
	err := uow.Do(ctx, func(ctx context.Context) error {
		tx, ok := GormTxFrom(ctx)
		assert.True(t, ok)

		// 插入數據
		err := tx.Exec("INSERT INTO test_users (name) VALUES (?)", "rollback_test").Error
		assert.NoError(t, err)

		// 返回錯誤，應該觸發回滾
		return testErr
	})
	assert.Error(t, err)
	assert.Equal(t, testErr, err)

	// 驗證數據被回滾，不存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM test_users WHERE name = ?", "rollback_test").Scan(&count)
	assert.Equal(t, int64(0), count)
}

func TestGormUow_Do_NestedTransaction(t *testing.T) {
	db := setupTestDBWithTable(t)
	uow := NewGormUow(db)
	ctx := context.Background()

	// 測試嵌套事務
	// 注意：GORM 的嵌套事務會創建保存點，內層事務失敗時只回滾到保存點
	err := uow.Do(ctx, func(ctx1 context.Context) error {
		tx1, ok := GormTxFrom(ctx1)
		assert.True(t, ok)
		assert.NotNil(t, tx1)

		// 在外層事務中插入數據
		err := tx1.Exec("INSERT INTO test_users (name) VALUES (?)", "outer").Error
		assert.NoError(t, err)

		// 嵌套內層事務 - 驗證可以正常創建嵌套事務
		return uow.Do(ctx1, func(ctx2 context.Context) error {
			// 驗證內層可以獲取到 tx
			tx2, ok := GormTxFrom(ctx2)
			assert.True(t, ok)
			assert.NotNil(t, tx2)

			// 驗證內層和外層的 tx 是同一個（GORM 嵌套事務使用保存點）
			// 注意：實際實現中，GORM 可能會創建新的 tx 實例，但底層連接相同
			// 這裡只驗證可以獲取到 tx，不驗證具體的數據操作

			return nil
		})
	})
	assert.NoError(t, err)

	// 驗證外層數據已提交
	var count int64
	db.Raw("SELECT COUNT(*) FROM test_users WHERE name = ?", "outer").Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestGormUow_Do_ContextPropagation(t *testing.T) {
	db := setupTestDB(t)
	uow := NewGormUow(db)
	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	// 測試 context 值傳播
	err := uow.Do(ctx, func(newCtx context.Context) error {
		// 驗證原始 context 的值可以傳播
		value := newCtx.Value("test_key")
		assert.Equal(t, "test_value", value)

		// 驗證可以獲取 tx
		tx, ok := GormTxFrom(newCtx)
		assert.True(t, ok)
		assert.NotNil(t, tx)

		return nil
	})
	assert.NoError(t, err)
}
