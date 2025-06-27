package nlp_test

import (
	"fmt"

	"github.com/ArditZubaku/nlp"
)

func ExampleTokenize() {
	text := "Who's on first?"
	tokens := nlp.Tokenize(text)
	fmt.Println(tokens)

	// Go will basically capture the stdout and compare it to the output below the //Output:
	// You should not touch that part, it should be AS IT IS, all the formatting and spaces and stuff

	// Output:
	// [who s on first]
}

/*
*	Test discovery in Go:
*	For every file ending with `_test.go`, run every function that matches either:
*	- Example[A-Z_].*, body MUST include `// Output:` comment
*	- Test[A-Z_].*
* */
