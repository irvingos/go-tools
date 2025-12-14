package safego

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

func Go(ctx context.Context, fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[Go] panic: %v\n%s", err, debug.Stack())
			}
		}()

		fn(ctx)
	}()
}

func GoWithWaitGroup(ctx context.Context, wg *sync.WaitGroup, fn func(ctx context.Context)) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("[Go] panic: %v\n%s", r, debug.Stack())
			}
		}()

		fn(ctx)
	}()
}
