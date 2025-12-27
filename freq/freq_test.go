package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkTop10(b *testing.B) {
	freq := make(map[string]int, 100_000)
	for i := range 100_000 {
		freq[fmt.Sprintf("word-%d", i)] = rand.Intn(10_000)
	}

	for b.Loop() {
		topN(freq, 10)
	}
}

func BenchmarkTop100(b *testing.B) {
	freq := make(map[string]int, 100_000)
	for i := range 100_000 {
		freq[fmt.Sprintf("word-%d", i)] = rand.Intn(10_000)
	}

	for b.Loop() {
		topN(freq, 100)
	}
}

func BenchmarkTop10v2(b *testing.B) {
	freq := make(map[string]int, 100_000)
	for i := range 100_000 {
		freq[fmt.Sprintf("word-%d", i)] = rand.Intn(10_000)
	}

	for b.Loop() {
		topNv2(freq, 10)
	}
}

func BenchmarkTop100v2(b *testing.B) {
	freq := make(map[string]int, 100_000)
	for i := range 100_000 {
		freq[fmt.Sprintf("word-%d", i)] = rand.Intn(10_000)
	}

	for b.Loop() {
		topNv2(freq, 100)
	}
}
