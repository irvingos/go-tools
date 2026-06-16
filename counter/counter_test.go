package counter

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCounter(t *testing.T) (Counter, func()) {
	t.Helper()
	s, err := miniredis.Run()
	require.NoError(t, err)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	cleanup := func() {
		_ = client.Close()
		s.Close()
	}
	return NewRedisCounter(client), cleanup
}

func TestRedisCounter_Incr_IncrBy_Get(t *testing.T) {
	c, cleanup := newTestCounter(t)
	defer cleanup()
	ctx := context.Background()
	key := "test:cnt:1"

	n, err := c.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)

	n, err = c.Incr(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	n, err = c.IncrBy(ctx, key, 4)
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)

	n, err = c.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
}

func TestRedisCounter_IncrBy_negative(t *testing.T) {
	c, cleanup := newTestCounter(t)
	defer cleanup()
	ctx := context.Background()
	key := "test:cnt:2"

	_, err := c.IncrBy(ctx, key, 10)
	require.NoError(t, err)
	n, err := c.IncrBy(ctx, key, -3)
	require.NoError(t, err)
	assert.Equal(t, int64(7), n)
}
