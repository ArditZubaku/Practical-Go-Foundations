package main

import (
	"fmt"
	"slices"
	"strings"
)

func main() {
	cart := []string{"apple", "orange", "banana"}
	fmt.Println("len:", len(cart))
	fmt.Println("cart[1]:", cart[1])

	for i := range cart {
		fmt.Println(i)
	}

	for i, v := range cart {
		fmt.Println(i, v)
	}

	for _, v := range cart {
		fmt.Println(v)
	}

	cart = append(cart, "milk")
	fmt.Println(cart)

	fruit := cart[:3]
	fmt.Println("fruit:", fruit)
	fruit = append(fruit, "lemon")
	fmt.Println("fruit:", fruit)
	fmt.Println("cart:", cart)

	var s []int
	for i := range 100 {
		s = appendInt(s, i)
	}
	fmt.Println("s[:3]", s[:3])
	fmt.Println(strings.Repeat("-", 10))

	n := 100
	s2 := make([]int, 0, n)
	for i := range n {
		s2 = appendInt(s2, i)
	}
	fmt.Println("s2[:3]", s2[:3])
	fmt.Println(strings.Repeat("-", 10))

	// Exercise concat - without using a "for" loop
	out := concat([]string{"A", "B"}, []string{"C"})
	fmt.Println(out) // expect [A B C]

	values := []float64{3, 1, 2}
	fmt.Println(median(values)) // expect 2
	values = []float64{3, 1, 2, 4}
	fmt.Println(median(values)) // expect 2.5
	fmt.Println("values:", values)

	players := []Player{
		{"Rick", 10},
		{"Morty", 9},
	}

	// Value semantics - here we get a copy of a player
	for _, p := range players {
		p.score += 1
	}
	fmt.Println("players:", players)

	// The solution - use indices to access elements (mutate)
	for i := range players {
		players[i].score += 1
	}
	fmt.Println("players:", players)
}

type Player struct {
	name  string
	score int
}

// median
// - sort values
// - if odd number of values => return middle
// - else return average of middles
func median(values []float64) float64 {
	// slices.Sort(values)
	values = slices.Sorted(slices.Values(values))
	idx := len(values) / 2
	if len(values)%2 == 1 {
		return values[idx]
	}

	return (values[idx-1] + values[idx]) / 2
}

func concat(s1, s2 []string) []string {
	return append(s1, s2...)
}

func appendInt(s []int, v int) []int {
	idx := len(s)
	if len(s) == cap(s) {
		// no more space in the underlying array - need to reallocate
		size := 2 * (len(s) + 1) // +1 because what if the slice is empty?
		fmt.Println(cap(s), "->", size)
		newSlice := make([]int, size)
		copy(newSlice, s)
		s = newSlice[:len(s)] // now 's' points to the new array/region of memory
	}

	s = s[:len(s)+1] // increase size so we can put the value into the slice
	s[idx] = v

	return s
}
