package main

import (
	"fmt"
	"sort"
)

// Me: I would definitely look at slices as sliding windows

func main() {
	var s []int
	fmt.Println("len", len(s)) // len is "nil" safe

	if s == nil { // You can compare only a slice to nil
		fmt.Println("nil slice")
	}

	s2 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Printf("s2 = %#v\n", s2)

	s3 := s2[1:4] // slicing operation, half-open range
	fmt.Printf("s3 = %#v, from idx 1 to 4 (not including it)\n", s3)

	// fmt.Println(s2[:100]) // This will cause a panic - out of range error
	s3 = append(s3, 100)
	fmt.Printf("s3 = %#v after append\n", s3)
	fmt.Printf("s2 = %#v after append\n", s2) // s2 is changed as well
	fmt.Printf("s2: len=%d, cap=%d", len(s2), cap(s2))
	fmt.Printf("s3: len=%d, cap=%d", len(s3), cap(s3))

	// s4 := make([]int, 1_000) // Single allocation, way more efficient when you know what you will be doing
	var s4 []int
	for i := range 1_000 {
		s4 = appendInt(s4, i)
	}

	fmt.Printf("s4 = %#v after append\n", s4)
	fmt.Printf("s4: len=%d, cap=%d\n", len(s4), cap(s4))

	fmt.Println(concat([]string{"A", "B"}, []string{"C", "D", "E"}))          // Should print [A B C D E]
	fmt.Println(concatTrickyWay([]string{"A", "B"}, []string{"C", "D", "E"})) // Should print [A B C D E]

	vs := []float64{2, 1, 3}
	fmt.Println(median(vs))
	vs = []float64{2, 1, 3, 4, 5.2}
	fmt.Println(median(vs))
	// Be careful, the sort function will sort the slice, which will reflect in the underlying array
	fmt.Println(vs)

	fmt.Println(median(nil))
}

func median(values []float64) (float64, error) {
	// BUG: Copy in order not to change underlying values
	// sort.Float64s(values)

	if len(values) == 0 {
		return 0, fmt.Errorf("median of empty slice")
	}

	nums := make([]float64, len(values))
	copy(nums, values)
	sort.Float64s(nums)

	i := len(nums) / 2
	// if len(values)&1 == 1 {
	if len(nums)%2 == 1 {
		return nums[i], nil
	}

	const n = 2 // const's type is infered based on where it is going to be used, use it in a floating context it becomes a float type
	return (nums[i-1] + nums[i]) / n, nil
}

// #EXERCISE:
func concat(s1, s2 []string) []string {
	// #RESTRICTIONS: No "for" loops
	s := make([]string, len(s1)+len(s2))
	copy(s[:len(s1)], s1)
	copy(s[len(s1):], s2)
	return s
}

func concatTrickyWay(s1, s2 []string) []string {
	return append(s1, s2...)
}

func appendInt(s []int, v int) []int {
	i := len(s)
	if len(s) < cap(s) { // Enough space in the underlying array
		// Extend the slice by one element to make room for the new value
		s = s[:len(s)+1]
	} else { // Need to reallocate and copy
		fmt.Printf("reallocate: %d -> %d\n", len(s), 2*len(s)+1)
		// Allocate a new slice with double the size + 1 (minimum 1 element)
		s2 := make([]int, 2*len(s)+1)
		copy(s2, s)
		s = s2[:len(s)+1]
	}

	s[i] = v
	return s
}
