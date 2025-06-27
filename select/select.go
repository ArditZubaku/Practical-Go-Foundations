package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ch1, ch2, ch3 := make(chan int), make(chan int), make(chan int)

	go func() {
		time.Sleep(10 * time.Millisecond)
		ch3 <- 3
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		ch1 <- 1
	}()

	go func() {
		time.Sleep(20 * time.Millisecond)
		ch2 <- 2
	}()

	go func() {
		time.Sleep(10 * time.Millisecond)
		ch3 <- 3
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
	defer cancel()

	// The idea came from a network syscall named `select`
	select {
	// INFO: The select is supposed to choose a ready case,
	// but when multiple are ready, it picks one at random — not in lexical order.

	case val := <-ch3:
		fmt.Println("ch3: ", val)
	case val := <-ch1:
		fmt.Println("ch1: ", val)
	case val := <-ch2:
		fmt.Println("ch2: ", val)
	case ch1 <- 1:
		fmt.Println("Can be a send too")
		// case val := <-ch3:
		// 	fmt.Println("ch3: ", val)
		// Cancellation:
	case <-time.After(5 * time.Millisecond):
		fmt.Println("TIMEDOUT")
	case <-ctx.Done():
		fmt.Println("DONE")
	}

	select {} // blocks forever without consuming CPU (unlike for{})
}
