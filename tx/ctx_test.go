package tx

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWithGormTx_StandardContext(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	tx := db.Begin()

	// 測試 WithGormTx
	newCtx := WithGormTx(ctx, tx)

	// 驗證返回的 context 不是原來的 context
	assert.NotEqual(t, ctx, newCtx)

	// 驗證可以從 context 中讀取 tx
	retrievedTx, ok := GormTxFrom(newCtx)
	assert.True(t, ok)
	assert.Equal(t, tx, retrievedTx)
}

func TestWithGormTx_GinContext(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	tx := db.Begin()

	// 測試 WithGormTx 使用 gin.Context
	newCtx := WithGormTx(c, tx)

	// 驗證返回的是同一個 gin.Context
	assert.Equal(t, c, newCtx)

	// 驗證可以從 gin.Context 中讀取 tx
	retrievedTx, ok := GormTxFrom(newCtx)
	assert.True(t, ok)
	assert.Equal(t, tx, retrievedTx)
}

func TestGormTxFrom_StandardContext_WithTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	tx := db.Begin()

	ctxWithTx := WithGormTx(ctx, tx)

	// 測試從 context 中讀取 tx
	retrievedTx, ok := GormTxFrom(ctxWithTx)
	assert.True(t, ok)
	assert.Equal(t, tx, retrievedTx)
}

func TestGormTxFrom_StandardContext_WithoutTx(t *testing.T) {
	ctx := context.Background()

	// 測試從沒有 tx 的 context 中讀取
	retrievedTx, ok := GormTxFrom(ctx)
	assert.False(t, ok)
	assert.Nil(t, retrievedTx)
}

func TestGormTxFrom_GinContext_WithTx(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	tx := db.Begin()

	WithGormTx(c, tx)

	// 測試從 gin.Context 中讀取 tx
	retrievedTx, ok := GormTxFrom(c)
	assert.True(t, ok)
	assert.Equal(t, tx, retrievedTx)
}

func TestGormTxFrom_GinContext_WithoutTx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	// 測試從沒有 tx 的 gin.Context 中讀取
	retrievedTx, ok := GormTxFrom(c)
	assert.False(t, ok)
	assert.Nil(t, retrievedTx)
}

func TestGormTxFrom_ContextWithOtherValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "other_key", "other_value")

	// 測試從包含其他值的 context 中讀取 tx
	retrievedTx, ok := GormTxFrom(ctx)
	assert.False(t, ok)
	assert.Nil(t, retrievedTx)
}

func TestWithGormTx_ReplaceTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	tx1 := db.Begin()
	tx2 := db.Begin()

	// 先設置 tx1
	ctxWithTx1 := WithGormTx(ctx, tx1)
	retrievedTx1, ok := GormTxFrom(ctxWithTx1)
	assert.True(t, ok)
	assert.Equal(t, tx1, retrievedTx1)

	// 替換為 tx2
	ctxWithTx2 := WithGormTx(ctxWithTx1, tx2)
	retrievedTx2, ok := GormTxFrom(ctxWithTx2)
	assert.True(t, ok)
	assert.Equal(t, tx2, retrievedTx2)
}

func TestWithGormTx_ReplaceTx_GinContext(t *testing.T) {
	db := setupTestDB(t)
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	tx1 := db.Begin()
	tx2 := db.Begin()

	// 先設置 tx1
	WithGormTx(c, tx1)
	retrievedTx1, ok := GormTxFrom(c)
	assert.True(t, ok)
	assert.Equal(t, tx1, retrievedTx1)

	// 替換為 tx2
	WithGormTx(c, tx2)
	retrievedTx2, ok := GormTxFrom(c)
	assert.True(t, ok)
	assert.Equal(t, tx2, retrievedTx2)
}
