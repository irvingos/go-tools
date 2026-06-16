package counter

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisCounter struct {
	client redis.Cmdable
}

// NewRedisCounter 基于 Redis 的计数器实现（单节点 / Cluster 均可，与 lock 包一致使用 Cmdable）。
func NewRedisCounter(client redis.Cmdable) Counter {
	return &redisCounter{client: client}
}

func (r *redisCounter) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisCounter) IncrBy(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}

func (r *redisCounter) Get(ctx context.Context, key string) (int64, error) {
	v, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return v, err
}
