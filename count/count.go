package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	/*Solution 1:
	var mu sync.Mutex
	count := 0
	*/

	count := int32(0)

	nGr, nIter := 10, 10_000

	var wg sync.WaitGroup
	wg.Add(nGr)

	for range nGr {
		go func() {
			defer wg.Done()
			for range nIter {
				/*Solution 1:
				mu.Lock()
				count++
				mu.Unlock()
				*/
				//Solution 2:
				atomic.AddInt32(&count, 1)
				time.Sleep(time.Microsecond)
			}
		}()
	}
	wg.Wait()
	fmt.Println("count:", count)
}

/*
go run -race ./count
"-race" is supported by:
- run
- build
- test

- Rule of thumb: use it with "go test -race"
*/
