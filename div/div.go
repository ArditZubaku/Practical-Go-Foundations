package main

import (
	"fmt"
	"log"
)

func main() {
	// fmt.Println(div(1, 0))
	fmt.Println(safeDiv(1, 0))
}

// INFO: Functions should log, only the main should print things

// func safeDiv(a, b int) (int, error) {
func safeDiv(a, b int) (q int, err error) {
	// `q` and `err` are local variables to this function, just like `a` and `b`
	defer func() {
		// e's type is any **not** error
		if e := recover(); e != nil {
			log.Printf("ERROR: %#v", e)
			err = fmt.Errorf("%v", e)
		}
	}()

	return a / b, nil
	// OR:
	// q = a / b
	// return
}

func div(a, b int) int {
	return a / b
}
