package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ArditZubaku/nlp"
	"github.com/ArditZubaku/nlp/stemmer"
	"github.com/gorilla/mux"
)

var (
	numTok = expvar.NewInt("tokenize.calls")
)

func main() {
	// Create server (dependency injection)
	logger := log.New(log.Writer(), "nlp ", log.LstdFlags|log.Lshortfile)
	s := Server{logger} // dependency injection - build all the components and then build the server
	// Routing
	// `/health` is an exact match
	// `/health/` is a prefix match
	// http.HandleFunc("/health", healthHandler)
	// http.HandleFunc("/tokenize", tokenizeHandler)
	r := mux.NewRouter()
	r.HandleFunc("/health", s.healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/tokenize", s.tokenizeHandler).Methods(http.MethodPost)
	r.HandleFunc("/stem/{word}", s.stemHandler).Methods(http.MethodGet)

	http.Handle("/", r)

	// Run the server
	addr := os.Getenv("NLPD_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	s.logger.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("ERROR: %#v", err)
	}
}

type Server struct {
	logger *log.Logger
}

func (s *Server) stemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	word := vars["word"]
	stem := stemmer.Stem(word)
	fmt.Fprintln(w, stem)
}

// EXERCISE:
// Write a tokenizeHandler that will read the text from the request body
// and return JSON in the format `{ "tokens": ["who", "on", "first"] }`
func (s *Server) tokenizeHandler(w http.ResponseWriter, r *http.Request) {
	// Before gorilla/mux
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	// ALWAYS return
	// 	return
	// }

	numTok.Add(1)

	// Step 1: Get, convert & validate data
	// Reads only 1MB of memory
	rdr := io.LimitReader(r.Body, 1_000_000)
	data, err := io.ReadAll(rdr)
	if err != nil {
		s.logger.Printf("ERROR: Can't read - %s", err)
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

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Run a health check
	fmt.Fprintln(w, "OK")
}
