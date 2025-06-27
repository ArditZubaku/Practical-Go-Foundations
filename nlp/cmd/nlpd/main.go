package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ArditZubaku/nlp"
)

func main() {
	// Routing
	// `/health` is an exact match
	// `/health/` is a prefix match
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/tokenize", tokenizeHandler)

	// Run the server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("ERROR: %#v", err)
	}
}

// EXERCISE:
// Write a tokenizeHandler that will read the text from the request body
// and return JSON in the format `{ "tokens": ["who", "on", "first"] }`

func tokenizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// ALWAYS return
		return
	}

	defer r.Body.Close()

	// Step 1: Get, convert & validate data
	// Reads only 1MB of memory
	rdr := io.LimitReader(r.Body, 1_000_000)
	data, err := io.ReadAll(rdr)
	if err != nil {
		http.Error(w, "Can't read", http.StatusBadRequest)
		// ALWAYS return
		return
	}

	if len(data) == 0 {
		http.Error(w, "Missing data", http.StatusBadRequest)
		return
	}

	// data is a byte slice
	text := string(data)

	// Step 2: Work
	tokens := nlp.Tokenize(text)

	// Step 3: Encode & Emit output
	resp := map[string]any{"tokens": tokens}

	// We could also do:
	// err = json.NewEncoder(w).Encode(resp)
	data, err = json.Marshal(resp)
	if err != nil {
		http.Error(w, "Can't encode", http.StatusInternalServerError)
		// ALWAYS return
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Run a health check
	fmt.Fprintln(w, "OK")
}
