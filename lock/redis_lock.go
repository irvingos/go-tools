package lock

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"

	"github.com/irvingos/go-tools/logx"
	"github.com/redis/go-redis/v9"
)

var (
	unlockScript = redis.NewScript(`
					if redis.call("GET", KEYS[1]) == ARGV[1] then
						return redis.call("DEL", KEYS[1])
					else
						return 0
					end
				`)

	renewScript = redis.NewScript(`
					if redis.call("GET", KEYS[1]) == ARGV[1] then
						return redis.call("PEXPIRE", KEYS[1], ARGV[2])
					else
						return 0
					end
				`)
)

func runBoolScript(
	ctx context.Context,
	script *redis.Script,
	client redis.Cmdable,
	keys []string,
	args ...any,
) (bool, error) {
	n, err := script.Run(ctx, client, keys, args...).Int64()
	if err != nil {
		// 注意：Lua 脚本一般不会通过 redis.Nil 表达“不存在”，而是返回 0
		// 所以这里的 err 基本只代表“执行失败”（网络/超时/脚本错误等）
		return false, err
	}
	return n != 0, nil
}

type redisLocker struct {
	client redis.Cmdable
}

func (r *redisLocker) Acquire(ctx context.Context, key string, opts ...Option) (Guard, error) {
	o, err := r.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	if o.Wait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.Wait)
		defer cancel()
	}

	startTime := time.Now()
	token := newToken()

	// 先 try 一次
	ok, err := r.trySetNX(ctx, key, token, o.TTL)
	if err != nil {
		return nil, err
	}
	if ok {
		return newRedisGuard(r.client, key, token, o.TTL, o.RenewInterval, time.Since(startTime)), nil
	}

	// 重试
	ticker := time.NewTicker(o.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return nil, ctx.Err()
		case <-ticker.C:
			ok, err := r.trySetNX(ctx, key, token, o.TTL)
			if err != nil {
				return nil, err
			}
			if ok {
				return newRedisGuard(r.client, key, token, o.TTL, o.RenewInterval, time.Since(startTime)), nil
			}
		}
	}
}

func (r *redisLocker) TryAcquire(ctx context.Context, key string, opts ...Option) (Guard, error) {
	o, err := r.applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	token := newToken()

	ok, err := r.trySetNX(ctx, key, token, o.TTL)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrBusy
	}
	return newRedisGuard(r.client, key, token, o.TTL, o.RenewInterval, time.Since(startTime)), nil
}

func (r *redisLocker) applyOptions(opts ...Option) (*AcquireOptions, error) {
	o := &AcquireOptions{
		TTL:           30 * time.Second,
		Wait:          0,
		RetryInterval: 100 * time.Millisecond,
		RenewInterval: 0,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}

	// 校验
	if o.TTL < 0 || o.Wait < 0 || o.RetryInterval < 0 || o.RenewInterval < 0 {
		return nil, ErrInvalidOption
	}
	if o.TTL == 0 {
		o.TTL = 30 * time.Second
	}
	if o.RetryInterval == 0 {
		o.RetryInterval = 100 * time.Millisecond
	}
	if o.RenewInterval > 0 && o.RenewInterval >= o.TTL {
		return nil, ErrInvalidOption
	}

	return o, nil
}

func (r *redisLocker) trySetNX(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return r.client.SetNX(ctx, key, token, ttl).Result()
}

func newToken() string {
	// 16 bytes random -> 32 hex chars
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

type redisGuard struct {
	client        redis.Cmdable
	key           string
	token         string
	ttl           time.Duration
	renewInterval time.Duration
	cost          time.Duration

	holdTime time.Time
	unlocked atomic.Bool

	stopRenew chan struct{}
	stopOnce  sync.Once
}

func newRedisGuard(client redis.Cmdable, key, token string, ttl, renewInterval, cost time.Duration) Guard {
	g := &redisGuard{
		client:        client,
		key:           key,
		token:         token,
		ttl:           ttl,
		renewInterval: renewInterval,
		cost:          cost,
		holdTime:      time.Now(),
		stopRenew:     make(chan struct{}),
	}
	if renewInterval > 0 {
		g.startAutoRenew()
	}
	return g
}

// Key implements Guard.
func (g *redisGuard) Key() string {
	return g.key
}

// Cost implements Guard.
func (g *redisGuard) Cost() time.Duration {
	return g.cost
}

// HeldFor implements Guard.
func (g *redisGuard) HeldFor() time.Duration {
	return time.Since(g.holdTime)
}

// Renew implements Guard.
func (g *redisGuard) Renew(ctx context.Context, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = g.ttl
	}
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	ok, err := runBoolScript(
		ctx,
		renewScript,
		g.client,
		[]string{g.key},
		g.token,
		int64(ttl/time.Millisecond),
	)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockLost // 或者更准确：ErrLockLost
	}
	return nil
}

// Unlock implements Guard.
func (g *redisGuard) Unlock(ctx context.Context) error {
	if g.unlocked.Load() {
		return nil
	}

	if g.unlocked.CompareAndSwap(false, true) {
		g.stopAutoRenew()
	}

	ok, err := runBoolScript(
		ctx,
		unlockScript,
		g.client,
		[]string{g.key},
		g.token,
	)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockLost
	}

	return nil
}

// startAutoRenew implements Guard.
func (g *redisGuard) startAutoRenew() {
	ticker := time.NewTicker(g.renewInterval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ok, err := runBoolScript(
					context.Background(),
					renewScript,
					g.client,
					[]string{g.key},
					g.token,
					int64(g.ttl/time.Millisecond),
				)
				if err != nil {
					// 临时错误：记录并继续（避免一次网络抖动就停止续租）
					logx.Error(err)
					continue
				}
				if !ok {
					// 锁已不属于当前持有者（过期/被抢/手动删）
					logx.Errorf("lock lost during auto-renew, key=%s", g.key)
					return
				}
			case <-g.stopRenew:
				return
			}
		}
	}()
}

func (g *redisGuard) stopAutoRenew() {
	g.stopOnce.Do(func() {
		close(g.stopRenew)
	})
}

func NewRedisLocker(client redis.Cmdable) Locker {
	return &redisLocker{
		client: client,
	}
}
