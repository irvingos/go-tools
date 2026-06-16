package counter

import "context"

// Counter 对 Redis 中的 key 做原子累加计数（无业务语义，调用方负责 key 命名空间）。
type Counter interface {
	// Incr 将 key 对应值加 1；key 不存在时从 0 开始，返回累加后的值。
	Incr(ctx context.Context, key string) (int64, error)
	// IncrBy 将 key 对应值加 delta；delta 可为负表示递减。key 不存在时视为从 0 开始。
	IncrBy(ctx context.Context, key string, delta int64) (int64, error)
	// Get 读取当前计数值；key 不存在时返回 (0, nil)。
	Get(ctx context.Context, key string) (int64, error)
}
