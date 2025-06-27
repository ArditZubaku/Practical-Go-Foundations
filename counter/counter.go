package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {
	// How to avoid race conditions
	// Solution 1: Use mutexes
	// INFO: Always put the mutex above the variable that it is supposed to guard
	// var mu sync.Mutex
	// count := 0

	// Solution 2: Use atomics (lower level than mutexes)
	// Always reach last for this, it is hard to get it right
	count := int64(0)

	const n = 10

	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			for range 10_000 {
				// mu.Lock()
				// count++
				// mu.Unlock()
				atomic.AddInt64(&count, 1)
			}
		}()
	}

	wg.Wait()

	fmt.Println(count)
}
