package graceful

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewRuntimeManager(t *testing.T) {
	m := NewRuntimeManager()
	if m == nil {
		t.Fatal("NewRuntimeManager returned nil")
	}
	if m.IsRunning("any") {
		t.Fatal("new manager should have no runtimes")
	}
}

func TestStartAndDone(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	h, err := m.Start(ctx, "a")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !m.IsRunning("a") {
		t.Fatal("IsRunning(a) should be true after Start")
	}

	h.Done()
	if m.IsRunning("a") {
		t.Fatal("IsRunning(a) should be false after Done")
	}

	if err := m.Drain(context.Background()); err != nil {
		t.Fatalf("Drain: %v", err)
	}
}

func TestStartDuplicateKey(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	if _, err := m.Start(ctx, "x"); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	_, err := m.Start(ctx, "x")
	if !errors.Is(err, ErrRuntimeAlreadyExists) {
		t.Fatalf("second Start: got %v, want ErrRuntimeAlreadyExists", err)
	}
}

func TestStartAfterDrain(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	h, err := m.Start(ctx, "r")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	h.Done()

	if err := m.Drain(context.Background()); err != nil {
		t.Fatalf("Drain: %v", err)
	}

	if _, err := m.Start(ctx, "r2"); !errors.Is(err, ErrShuttingDown) {
		t.Fatalf("Start after Drain: got %v, want ErrShuttingDown", err)
	}
}

func TestCancel(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	h, err := m.Start(ctx, "c")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if !m.Cancel("c") {
		t.Fatal("Cancel should return true for existing key")
	}
	select {
	case <-h.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("context should be canceled after Cancel")
	}
	if !m.IsRunning("c") {
		t.Fatal("Cancel does not remove runtime; IsRunning should still be true until Done")
	}

	h.Done()
	if m.IsRunning("c") {
		t.Fatal("after Done, runtime should be gone")
	}
}

func TestCancelMissing(t *testing.T) {
	m := NewRuntimeManager()
	if m.Cancel("nope") {
		t.Fatal("Cancel missing key should return false")
	}
}

func TestShutdownWaitsForDone(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	var wg sync.WaitGroup
	for _, key := range []string{"u", "v"} {
		h, err := m.Start(ctx, key)
		if err != nil {
			t.Fatalf("Start %s: %v", key, err)
		}
		wg.Add(1)
		go func(h *RuntimeHandle) {
			defer wg.Done()
			<-h.Context().Done()
			h.Done()
		}(h)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	wg.Wait()
}

func TestDrainWaitsForDone(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	h, err := m.Start(ctx, "d")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.Done()
	}()

	drainCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := m.Drain(drainCtx); err != nil {
		t.Fatalf("Drain: %v", err)
	}
	<-done
}

func TestWaitContextCanceled(t *testing.T) {
	m := NewRuntimeManager()
	ctx := context.Background()

	if _, err := m.Start(ctx, "block"); err != nil {
		t.Fatalf("Start: %v", err)
	}

	waitCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := m.wait(waitCtx); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait: got %v, want context.Canceled", err)
	}
}

func TestDoneIdempotent(t *testing.T) {
	m := NewRuntimeManager()
	h, err := m.Start(context.Background(), "id")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	h.Done()
	h.Done()
	if m.IsRunning("id") {
		t.Fatal("runtime should be removed after first Done")
	}
	if err := m.Drain(context.Background()); err != nil {
		t.Fatalf("Drain: %v", err)
	}
}

func TestDoneReplacesOnlyCurrentHandle(t *testing.T) {
	m := NewRuntimeManager()
	h1, err := m.Start(context.Background(), "same")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	h1.Done()

	h2, err := m.Start(context.Background(), "same")
	if err != nil {
		t.Fatalf("second Start: %v", err)
	}
	h1.Done()
	if !m.IsRunning("same") {
		t.Fatal("Done on stale handle must not delete current runtime")
	}
	h2.Done()
	if err := m.Drain(context.Background()); err != nil {
		t.Fatalf("Drain: %v", err)
	}
}
