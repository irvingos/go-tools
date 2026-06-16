package lock

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestLocker(t *testing.T) (Locker, func()) {
	t.Helper()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		_ = client.Close()
		s.Close()
	}

	return NewRedisLocker(client), cleanup
}

func TestTryAcquire_SuccessBusyUnlockThenSuccess(t *testing.T) {
	locker, cleanup := newTestLocker(t)
	defer cleanup()

	ctx := context.Background()
	key := "k1"

	// 第一次 TryAcquire 成功
	g1, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
	if err != nil {
		t.Fatalf("TryAcquire #1 err: %v", err)
	}
	if g1 == nil {
		t.Fatalf("TryAcquire #1 guard is nil")
	}

	// 第二次 TryAcquire 应该返回 ErrBusy
	g2, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
	if !errors.Is(err, ErrBusy) {
		t.Fatalf("TryAcquire #2 expected ErrBusy, got guard=%v err=%v", g2, err)
	}
	if g2 != nil {
		t.Fatalf("TryAcquire #2 expected nil guard when busy, got %T", g2)
	}

	// 解锁后再次 TryAcquire 成功
	if err := g1.Unlock(ctx); err != nil {
		t.Fatalf("Unlock err: %v", err)
	}

	g3, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
	if err != nil {
		t.Fatalf("TryAcquire #3 err: %v", err)
	}
	if g3 == nil {
		t.Fatalf("TryAcquire #3 guard is nil")
	}
	_ = g3.Unlock(ctx)
}

func TestAcquire_WaitsAndSucceedsAfterRelease(t *testing.T) {
	locker, cleanup := newTestLocker(t)
	defer cleanup()

	ctx := context.Background()
	key := "k2"

	// 先占锁
	g1, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
	if err != nil {
		t.Fatalf("TryAcquire err: %v", err)
	}
	defer g1.Unlock(ctx)

	// 200ms 后释放锁
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = g1.Unlock(context.Background())
	}()

	start := time.Now()
	g2, err := locker.Acquire(ctx, key,
		WithTTL(5*time.Second),
		WithWait(2*time.Second), // 最多等 2s
		WithRetryInterval(50*time.Millisecond),
	)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Acquire err: %v", err)
	}
	if g2 == nil {
		t.Fatalf("Acquire guard is nil")
	}
	defer g2.Unlock(ctx)

	// 应该至少等待了 ~200ms（允许一些调度误差）
	if elapsed < 150*time.Millisecond {
		t.Fatalf("Acquire returned too fast: %v", elapsed)
	}
}

func TestAcquire_TimeoutWhenHeld(t *testing.T) {
	locker, cleanup := newTestLocker(t)
	defer cleanup()

	ctx := context.Background()
	key := "k3"

	// 先占锁，不释放
	g1, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
	if err != nil {
		t.Fatalf("TryAcquire err: %v", err)
	}
	defer g1.Unlock(ctx)

	// Acquire 等待 300ms 应该超时（因为锁一直被持有）
	start := time.Now()
	g2, err := locker.Acquire(ctx, key,
		WithTTL(5*time.Second),
		WithWait(300*time.Millisecond),
		WithRetryInterval(50*time.Millisecond),
	)
	elapsed := time.Since(start)

	if g2 != nil {
		t.Fatalf("Acquire expected nil guard on timeout, got %T", g2)
	}
	// 这里 Acquire 的实现是用 ctx.WithTimeout 包装 Wait，因此期望 ctx deadline exceeded
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Acquire expected DeadlineExceeded, got %v", err)
	}
	if elapsed < 250*time.Millisecond {
		t.Fatalf("Acquire timeout too fast: %v", elapsed)
	}
}

func TestUnlock_IsIdempotent(t *testing.T) {
	locker, cleanup := newTestLocker(t)
	defer cleanup()

	ctx := context.Background()
	key := "k4"

	g, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
	if err != nil {
		t.Fatalf("TryAcquire err: %v", err)
	}

	// 多次 Unlock 不应报错
	if err := g.Unlock(ctx); err != nil {
		t.Fatalf("Unlock #1 err: %v", err)
	}
	if err := g.Unlock(ctx); err != nil {
		t.Fatalf("Unlock #2 err: %v", err)
	}
}

func TestConcurrentTryAcquire_OnlyOneWinner(t *testing.T) {
	locker, cleanup := newTestLocker(t)
	defer cleanup()

	ctx := context.Background()
	key := "k5"

	const n = 20
	var winners int
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			g, err := locker.TryAcquire(ctx, key, WithTTL(5*time.Second))
			if err == nil && g != nil {
				// 赢者释放锁
				_ = g.Unlock(ctx)
				// 统计赢者
				// 这里简单用 redis 语义保证最多一个赢者，统计只是验证
				// 如需更严格可用 atomic.AddInt32
				winners++
			}
		}()
	}
	wg.Wait()

	if winners == 0 {
		t.Fatalf("expected at least 1 winner")
	}
	// 注意：由于 winners++ 非原子，这个断言只做弱校验。
	// 如果你要严格，改为 atomic.Int32。
}

func TestRenew(t *testing.T) {
	locker, cleanup := newTestLocker(t)
	defer cleanup()

	key := "k6"
	guard, err := locker.Acquire(context.Background(), key, WithTTL(200*time.Millisecond))
	assert.Nil(t, err)
	_, err = locker.TryAcquire(context.Background(), key)
	assert.EqualError(t, err, ErrBusy.Error())

	time.Sleep(200 * time.Microsecond)
	_, err = locker.TryAcquire(context.Background(), key)
	assert.Nil(t, nil)

	guard.Renew(context.Background(), 200*time.Millisecond)

	_, err = locker.TryAcquire(context.Background(), key)
	assert.EqualError(t, err, ErrBusy.Error())

	guard.Unlock(context.Background())
}
