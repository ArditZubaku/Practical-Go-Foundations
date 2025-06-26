package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

// Q: What is the most common word (case insensitive) in sherlock.txt
// => Word frequency
func main() {
	file, err := os.Open("sherlock.txt")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
	defer file.Close()

	path := "C:\to\new\file.csv"
	fmt.Println(path)

	// "Raw" string
	path = `C:\to\new\file.csv`
	fmt.Println(path)

	// Multiline string using raw strings
	req := `GET /ip / HTTP/1.1
	Host: httpbin.org
	Connection: Close

	`
	fmt.Println(req)

	mapDemo()

	w, err := mostCommon(file)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	fmt.Println("Most common word: ", w)
}

func mostCommon(r io.Reader) (string, error) {
	freqs, err := wordFrequency(r)
	if err != nil {
		return "", err
	}

	return maxWord(freqs)
}

func mapDemo() {
	var stocks map[string]float64 // symbol -> price
	symbol := "TTWO"
	price := stocks[symbol]
	fmt.Printf("%s -> $%.2f\n", symbol, price)

	if price, ok := stocks[symbol]; ok {
		fmt.Printf("%s -> $%.2f\n", symbol, price)
	} else {
		fmt.Printf("%s not found\n", symbol)
	}

	// INFO: Reading from maps is no problem even if they are not initialized
	// Iterating and stuff
	// Setting values is the problem

	// The Fix:
	// stocks = make(map[string]float64)
	// stocks[symbol] = 136.74
	// OR:
	stocks = map[string]float64{
		symbol: 137.74,
		"AAPL": 172.35,
	}
	if price, ok := stocks[symbol]; ok {
		fmt.Printf("%s -> $%.2f\n", symbol, price)
	} else {
		fmt.Printf("%s not found\n", symbol)
	}

	for k := range stocks { // Keys
		fmt.Println(k)
	}

	for k, v := range stocks { // Keys and Values
		fmt.Println(k, "->", v)
	}

	for _, v := range stocks { // Values
		fmt.Println(v)
	}

	delete(stocks, "AAPL")
	fmt.Println(stocks)
	// What if we try to delete again - NO panic, nil safe
	delete(stocks, "AAPL")
}

// INFO: This will run before main
// Global variables run before main
// And also the special func init()
//
// "Whos's on first?" -> [Who s on first]
var wordRegEx = regexp.MustCompile(`[a-zA-Z]+`)

func maxWord(freqs map[string]int) (string, error) {
	if len(freqs) == 0 {
		return "", fmt.Errorf("ERROR: empty map")
	}

	maxN, maxW := 0, ""
	for w, c := range freqs {
		if c > maxN {
			maxN, maxW = c, w
		}
	}

	return maxW, nil
}

func wordFrequency(r io.Reader) (map[string]int, error) {
	s := bufio.NewScanner(r)
	lnum := 0
	freqs := make(map[string]int) // word -> count

	for s.Scan() {
		lnum++
		// s.Text()// Current line
		words := wordRegEx.FindAllString(s.Text(), -1) // -1 basically means return all the matches
		// if len(words) != 0 {
		// 	fmt.Println(words)
		// 	break
		// }
		for _, w := range words {
			freqs[strings.ToLower(w)]++
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	// fmt.Println("NUM LINES: ", lnum)

	return freqs, nil
}
