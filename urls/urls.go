package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

func main() {
	urls := []string{
		"https://go.dev",
		"https://ardanlabs.com",
		"https://ibm.com/no/such/page",
	}

	start := time.Now()
	// for _, url := range urls {
	// 	s, err := urlCheck(url)
	// 	fmt.Printf("%q: %d (%v)\n", url, s, err)
	// }
	fanOutResult(urls)
	fmt.Println(strings.Repeat("-", 20))
	fanOutWait(urls)
	fmt.Println(strings.Repeat("-", 20))
	fanOutPool(urls)
	duration := time.Since(start)
	fmt.Printf("%d urls in %v\n", len(urls), duration)
}

func fanOutWait(urls []string) {
	var wg sync.WaitGroup

	wg.Add(len(urls))
	// fan-out
	for _, url := range urls {
		// wg.Add(1)
		go func() {
			defer wg.Done()
			urlLog(url)
		}()
	}

	// wait for goroutines to finish
	// if we need errors we could use errgroup pkg
	wg.Wait()
}

func fanOutPool(urls []string) {
	var wg sync.WaitGroup

	ch := make(chan string)
	// producer
	go func() {
		for _, url := range urls {
			ch <- url
		}
		close(ch)
	}()

	const size = 2

	wg.Add(size)

	for range size {
		// consumers
		go func() {
			defer wg.Done()
			for url := range ch {
				urlLog(url)
			}
		}()
	}

	// wait for goroutines to finish
	wg.Wait()
}

func urlLog(url string) {
	res, err := http.Get(url)
	if err != nil {
		slog.Error("urlLog", "url", url, "error", err)
		return
	}

	slog.Info("urlLog", "url", url, "status", res.StatusCode)
}

func fanOutResult(urls []string) {
	type result struct {
		url    string
		status int
		err    error
	}

	ch := make(chan result)

	// fan-out
	for _, url := range urls {
		go func() {
			r := result{url: url}
			defer func() { ch <- r }()

			r.status, r.err = urlCheck(url)
		}()
	}

	// collect results
	for range urls {
		r := <-ch
		fmt.Printf("%q: %d (%v)\n", r.url, r.status, r.err)
	}
}

func urlCheck(url string) (int, error) {
	res, err := http.Get(url)
	if err != nil {
		return 0, nil
	}

	return res.StatusCode, nil
}
