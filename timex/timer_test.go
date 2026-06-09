package timex

import (
	"context"
	"testing"
	"time"
)

func TestSleepWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := SleepWithContext(ctx, 1*time.Second)
	if err != nil {
		t.Fatalf("SleepWithContext: %v", err)
	}
}

func TestSleepWithContext_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := SleepWithContext(ctx, 2*time.Second)
	if err != context.DeadlineExceeded {
		t.Fatalf("SleepWithContext: %v", err)
	}
}
