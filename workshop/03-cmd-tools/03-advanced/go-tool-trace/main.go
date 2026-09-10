package main

import (
	"context"
	"fmt"
	"runtime/trace"
	"sync"
	"time"
)

func processWork(ctx context.Context) {
	var lock sync.Mutex
	var wg sync.WaitGroup

	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			trace.WithRegion(ctx, "step", func() {
				lock.Lock()
				defer lock.Unlock()
				time.Sleep(25 * time.Millisecond)
			})
		}()
	}
	wg.Wait()
}

func main() {
	processWork(context.Background())
	fmt.Println("work completed")
}
