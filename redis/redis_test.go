package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	_, err := NewClient(context.Background(), Config{
		Mode: "single",
		Addr: ":6379",
	})
	assert.Nil(t, err)

	_, err = NewClient(context.Background(), Config{
		Mode: "single",
		Addr: ":63791",
	})
	assert.NotNil(t, err)
}
