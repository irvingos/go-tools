package lock

import (
	"context"
	"errors"
	"time"
)

var (
	// TryAcquire 专用
	ErrBusy          = errors.New("lock busy")
	ErrInvalidOption = errors.New("invalid lock option")
	ErrNotSupported  = errors.New("not supported")
	ErrLockLost      = errors.New("lock lost")
)

type Locker interface {
	// TryAcquire: 立即尝试一次；锁被占用返回 ErrBusy。
	// 注意：TryAcquire 忽略 Wait/RetryInterval（保持语义明确）
	TryAcquire(ctx context.Context, key string, opts ...Option) (Guard, error)

	// Acquire: 按 Wait 最大等待时长重试获取。
	// - Wait > 0: 最多等待 Wait
	// - Wait == 0: 不额外限制，直到 ctx.Done()
	Acquire(ctx context.Context, key string, opts ...Option) (Guard, error)
}

type Guard interface {
	Key() string
	Cost() time.Duration
	HeldFor() time.Duration
	Unlock(ctx context.Context) error
	Renew(ctx context.Context, ttl time.Duration) error
}

type AcquireOptions struct {
	TTL           time.Duration
	Wait          time.Duration
	RetryInterval time.Duration
	RenewInterval time.Duration
}

type Option func(*AcquireOptions)

func WithTTL(ttl time.Duration) Option {
	return func(o *AcquireOptions) {
		o.TTL = ttl
	}
}

func WithWait(wait time.Duration) Option {
	return func(o *AcquireOptions) {
		o.Wait = wait
	}
}

func WithRetryInterval(interval time.Duration) Option {
	return func(o *AcquireOptions) {
		o.RetryInterval = interval
	}
}

func WithRenewInterval(interval time.Duration) Option {
	return func(o *AcquireOptions) {
		o.RenewInterval = interval
	}
}
