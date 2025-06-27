package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	p := Payment{From: "Wile. E. Coyote", To: "ACME", Amount: 123.34}
	// What if we wanted this to be executed only once? Solution => sync.Once
	p.Process()
	p.Process()
}

type Payment struct {
	From   string
	To     string
	Amount float64 // USD

	once sync.Once
}

func (p *Payment) Process() {
	t := time.Now()
	p.once.Do(func() { p.process(t) })
}

func (p *Payment) process(t time.Time) {
	ts := t.Format(time.RFC3339)
	fmt.Printf("[%s] %s -> $%.2f -> %s\n", ts, p.From, p.Amount, p.To)
}
