package graceful

import (
	"context"
	"sync"
	"sync/atomic"
)

type RuntimeManager struct {
	wg       sync.WaitGroup
	shutting atomic.Bool
}

func (m *RuntimeManager) Begin() bool {
	if m.shutting.Load() {
		return false
	}
	m.wg.Add(1)

	// double check, 避免 Add 之后立即进入 shutdown 竞态
	if m.shutting.Load() {
		m.wg.Done()
		return false
	}

	return true
}

func (m *RuntimeManager) End() {
	m.wg.Done()
}

func (m *RuntimeManager) Shutdown(ctx context.Context) error {
	m.shutting.Store(true)

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *RuntimeManager) IsShuttingDown() bool {
	return m.shutting.Load()
}
