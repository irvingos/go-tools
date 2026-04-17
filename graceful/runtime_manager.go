package graceful

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrShuttingDown         = errors.New("runtime manager is shutting down")
	ErrRuntimeAlreadyExists = errors.New("runtime is already exists")
)

type RuntimeManager struct {
	wg       sync.WaitGroup
	shutting atomic.Bool

	mu       sync.Mutex
	runtimes map[string]*RuntimeHandle
}

type RuntimeHandle struct {
	key string

	ctx    context.Context
	cancel context.CancelFunc

	manager  *RuntimeManager
	doneOnce sync.Once
}

func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		runtimes: make(map[string]*RuntimeHandle),
	}
}

func (m *RuntimeManager) Start(parent context.Context, key string) (*RuntimeHandle, error) {
	if m.shutting.Load() {
		return nil, ErrShuttingDown
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shutting.Load() {
		return nil, ErrShuttingDown
	}
	if _, exists := m.runtimes[key]; exists {
		return nil, ErrRuntimeAlreadyExists
	}

	ctx, cancel := context.WithCancel(parent)
	handle := &RuntimeHandle{
		key:     key,
		ctx:     ctx,
		cancel:  cancel,
		manager: m,
	}

	m.runtimes[key] = handle
	m.wg.Add(1)

	return handle, nil
}

func (m *RuntimeManager) Cancel(key string) bool {
	m.mu.Lock()
	handle, exists := m.runtimes[key]
	m.mu.Unlock()
	if !exists {
		return false
	}

	handle.Cancel()
	return true
}

func (m *RuntimeManager) IsRunning(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.runtimes[key]
	return exists
}

func (m *RuntimeManager) Drain(ctx context.Context) error {
	m.shutting.Store(true)
	return m.wait(ctx)
}

func (m *RuntimeManager) Shutdown(ctx context.Context) error {
	m.shutting.Store(true)

	m.mu.Lock()
	handles := make([]*RuntimeHandle, 0, len(m.runtimes))
	for _, h := range m.runtimes {
		handles = append(handles, h)
	}
	m.mu.Unlock()

	for _, handle := range handles {
		handle.Cancel()
	}

	return m.wait(ctx)
}

func (m *RuntimeManager) wait(ctx context.Context) error {
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

func (h *RuntimeHandle) Context() context.Context {
	return h.ctx
}

func (h *RuntimeHandle) Cancel() {
	h.cancel()
}

func (h *RuntimeHandle) Done() {
	h.doneOnce.Do(func() {
		m := h.manager

		m.mu.Lock()
		if current, exists := m.runtimes[h.key]; exists && current == h {
			delete(m.runtimes, h.key)
		}
		m.mu.Unlock()

		h.cancel()
		m.wg.Done()
	})
}
