package main

import (
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	// siteTime("https://google.com")

	urls := []string{
		"https://google.com",
		"https://apple.com",
		"https://no-such-site.biz",
	}

	// INFO: Use this "pattern" when you don't care about the results but just want things to get done
	var wg sync.WaitGroup

	// INFO: Use this package instead if you care about the errors
	// golang.org/x/sync/errgroup

	// You are basically telling it for how many counters you are going to wait (a counter in this case means a goroutine)
	wg.Add(len(urls))

	for _, url := range urls {
		// OR:
		// wg.Add(1)
		go func(url string) {
			// Decrements the wg counter by 1
			defer wg.Done()
			siteTime(url)
		}(url)
	}

	// This basically says wait til every goroutine in this group is done a.k.a counter is zero
	wg.Wait()
}

func siteTime(url string) {
	start := time.Now()

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("ERROR: %s -> %s", url, err)
		return
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		log.Printf("ERROR: %s -> %s", url, err)
	}

	duration := time.Since(start)
	log.Printf("INFO: %s -> %v", url, duration)
}
