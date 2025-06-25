package main

import (
	"fmt"
	"log"
	"unicode/utf8"
)

func main() {
	banner("Go", 6)

	s := "G☺"
	fmt.Println("len", len(s))

	for i, r := range s {
		fmt.Println(i, r)
		if i == 0 {
			fmt.Printf("%c of type %T\n", r, r)
		}
	}

	// byte (uint8)
	// rune (int32)
	//
	b := s[0]
	fmt.Printf("%c of type %T\n", b, b)

	x, y := 1, "1"
	fmt.Printf("x=%v, y=%v\n", x, y)
	fmt.Printf("x=%#v, y=%#v\n", x, y)
	log.Printf("x=%#v, y=%#v\n", x, y)

	fmt.Printf("%20s\n", "Test")

	fmt.Println("g", isPalindrome("g"))
	fmt.Println("gog", isPalindrome("g"))
	fmt.Println("gogo", isPalindrome("g"))

	fmt.Println("g", isPalindromeUnicodeAware("g"))
	fmt.Println("gog", isPalindromeUnicodeAware("g"))
	fmt.Println("gogo", isPalindromeUnicodeAware("g"))
}

// isPalindrome("g") -> true
// isPalindrome("go") -> false
// isPalindrome("gog") -> true
// isPalindrome("gogo") -> false
func isPalindrome(s string) bool {
	for i := range len(s) / 2 {
		if s[i] != s[len(s)-i-1] {
			return false
		}
	}
	return true
}

func isPalindromeUnicodeAware(s string) bool {
	// Convert to a slice of runes and let go do its job
	// A rune is basically a code-point
	// You could think of it as a character, but that may vary in size
	rs := []rune(s)
	for i := range len(rs) / 2 {
		if rs[i] != rs[len(rs)-i-1] {
			return false
		}
	}
	return true
}

func banner(text string, width int) {
	// padding := (width - len(text)) / 2 // BUG: Len is in bytes
	padding := (width - utf8.RuneCountInString(text)) / 2
	for range width {
		fmt.Print("-")
	}
	fmt.Println()
	for range padding {
		fmt.Print(" ")
	}
	fmt.Println(text)
	for range width {
		fmt.Print("-")
	}
	fmt.Println()
}
