package main

import (
	"fmt"
	"time"
)

func main() {
	go fmt.Println("goroutine")
	fmt.Println("main")

	for i := range 3 {
		// BUG: All goroutines use the same "i" for the for loop
		// When the functions start executing, the i will be 3 for all of them because the closure captures variables by reference
		// go func() {
		// 	fmt.Println(i)
		// }()

		// FIX 1:
		go func(i int) {
			fmt.Println(i)
		}(i)

		// FIX 2:
		// Use a loop body variable
		i := i // new `i` shadows the `i` from the loop
		go func() {
			fmt.Println(i)
		}()
	}

	time.Sleep(10 * time.Millisecond)

	shadowExample()

	ch := make(chan string)
	// BUG:
	// ch <- "hi"  // send

	// THEFIX:
	go func() {
		ch <- "hi"
	}()

	msg := <-ch // receive
	fmt.Println(msg)

	go func() {
		for i := range 3 {
			msg := fmt.Sprintf("Message #%d", i+1)
			ch <- msg
		}
		// INFO: If we don't close the channel, go won't know when to stop iterating over it
		close(ch)
	}()

	for msg := range ch {
		fmt.Println("Got: ", msg)
	}

	// The for range is just a syntactic sugar for this:
	// for {
	// 	msg, ok := <-ch
	// 	if !ok {
	// 		break
	// 	}
	// 	fmt.Println("Got: ", msg)
	// }

	msg = <-ch // ch is closed above
	fmt.Printf("CLOSED: %#v\n", msg)

	msg, ok := <-ch // ch is closed above
	fmt.Printf("CLOSED: %#v, OK?: %#v\n", msg, ok)

	// BUG:
	// ch <- "test" // ch is closed -> panic

	// BUG:
	// close(ch) // ch is already closed -> panic

	values := []int{15, 8, 42, 16, 4, 23}
	fmt.Println(sleepSort(values))
}

// For every value "n" in values, spin a goroutine that will:
// - sleep "n" milliseconds
// - send "n" over a channel
//
// In the funcion body, collect values from the channel to a slice and return it
func sleepSort(values []int) []int {
	ch := make(chan int)
	for _, n := range values {
		n := n
		go func() {
			time.Sleep(time.Duration(n) * (time.Millisecond))
			ch <- n
		}()
	}

	var out []int
	for range values {
		n := <-ch
		out = append(out, n)
	}

	return out
}

/*
*	Channel semantics
*	- send and receive will block until opposite operation (*)
*	- receiving from a closed channel will return the zero value without blocking
*	- sending to a closed channel will panic
*	- closing a closed channel will panic
*	- sending/receiving to/from a nil channel will block forever
* */

func shadowExample() {
	n := 7
	{
		// n := n
		n := 2
		fmt.Println("inner: ", n)
	}
	fmt.Println("outer: ", n)
}
