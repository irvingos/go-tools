package graceful

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRuntimeManager_BeginEnd(t *testing.T) {
	m := &RuntimeManager{}

	// 正常開始和結束
	if !m.Begin() {
		t.Error("Begin() should return true when not shutting down")
	}
	m.End()
}

func TestRuntimeManager_BeginAfterShutdown(t *testing.T) {
	m := &RuntimeManager{}

	// 開始關閉
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		_ = m.Shutdown(ctx)
	}()

	// 等待 shutdown 開始
	time.Sleep(10 * time.Millisecond)

	// 關閉後 Begin 應該返回 false
	if m.Begin() {
		t.Error("Begin() should return false after shutdown started")
		m.End() // 避免洩漏
	}
}

func TestRuntimeManager_IsShuttingDown(t *testing.T) {
	m := &RuntimeManager{}

	if m.IsShuttingDown() {
		t.Error("IsShuttingDown() should return false initially")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		_ = m.Shutdown(ctx)
	}()

	// 等待 shutdown 開始
	time.Sleep(10 * time.Millisecond)

	if !m.IsShuttingDown() {
		t.Error("IsShuttingDown() should return true after shutdown started")
	}
}

func TestRuntimeManager_ShutdownWaitsForTasks(t *testing.T) {
	m := &RuntimeManager{}

	taskDone := make(chan struct{})

	// 開始一個任務
	if !m.Begin() {
		t.Fatal("Begin() should return true")
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		m.End()
		close(taskDone)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	start := time.Now()
	err := m.Shutdown(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Shutdown() should not return error, got: %v", err)
	}

	// 應該等待任務完成
	if elapsed < 50*time.Millisecond {
		t.Errorf("Shutdown() should wait for tasks, elapsed: %v", elapsed)
	}

	<-taskDone
}

func TestRuntimeManager_ShutdownTimeout(t *testing.T) {
	m := &RuntimeManager{}

	// 開始一個永遠不會結束的任務
	if !m.Begin() {
		t.Fatal("Begin() should return true")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := m.Shutdown(ctx)

	if err != context.DeadlineExceeded {
		t.Errorf("Shutdown() should return DeadlineExceeded, got: %v", err)
	}

	// 清理：結束任務以避免洩漏
	m.End()
}

func TestRuntimeManager_ConcurrentBeginEnd(t *testing.T) {
	m := &RuntimeManager{}

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for range numGoroutines {
		go func() {
			defer wg.Done()
			if m.Begin() {
				time.Sleep(time.Millisecond)
				m.End()
			}
		}()
	}

	wg.Wait()

	// 所有任務完成後應該能正常關閉
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := m.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() should not return error after all tasks done, got: %v", err)
	}
}

func TestRuntimeManager_BeginDuringShutdown(t *testing.T) {
	m := &RuntimeManager{}

	// 開始一個任務
	if !m.Begin() {
		t.Fatal("Begin() should return true")
	}

	shutdownDone := make(chan error)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		shutdownDone <- m.Shutdown(ctx)
	}()

	// 等待 shutdown 開始
	time.Sleep(10 * time.Millisecond)

	// 在 shutdown 期間嘗試 Begin
	if m.Begin() {
		t.Error("Begin() should return false during shutdown")
		m.End()
	}

	// 結束原始任務
	m.End()

	// 等待 shutdown 完成
	err := <-shutdownDone
	if err != nil {
		t.Errorf("Shutdown() should complete without error, got: %v", err)
	}
}

func TestRuntimeManager_MultipleShutdown(t *testing.T) {
	m := &RuntimeManager{}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 第一次 shutdown
	err := m.Shutdown(ctx)
	if err != nil {
		t.Errorf("First Shutdown() should not return error, got: %v", err)
	}

	// 第二次 shutdown 也應該正常完成
	err = m.Shutdown(ctx)
	if err != nil {
		t.Errorf("Second Shutdown() should not return error, got: %v", err)
	}
}
