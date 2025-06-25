package main

import (
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
)

// Calculate the digital signature of the unzipped http.log.gz

func main() {
	const fileName = "http.log.gz"

	// sha1Sum("noexist")
	sig, err := sha1Sum(fileName)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	fmt.Println(sig)

	sig, err = sha1Sum("sha1.go")
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	fmt.Println(sig)
}

/*
* if file name ends with .gz
* 	cat http.log.gz | gunzip | sha1sum
* else
* 	cat http.log | sha1sum
* */
func sha1Sum(fileName string) (string, error) {
	// idiom: acquire a resource, check for error, defer release
	// Deferring file.Close() before checking err means you might be deferring Close() on a nil pointer.
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer closeOrLog(file) // defers are called in LIFO order

	var r io.ReadCloser = file

	if strings.HasSuffix(fileName, "gz") {
		// At this point file returns us the compressed content (file is a .gz)
		r, err = gzip.NewReader(file)
		if err != nil {
			return "", err
		}
		defer closeOrLog(r)
		// io.CopyN(os.Stdout, r, 100)
	}

	w := sha1.New()
	if _, err := io.Copy(w, r); err != nil {
		return "", err
	}

	sig := w.Sum(nil)

	// %x - convert bytes to strings, 2 hex chars per byte
	return fmt.Sprintf("%x", sig), nil
}

func closeOrLog(closer io.Closer) {
	if err := closer.Close(); err != nil {
		slog.Error("Failed to close resource", "error", err)
	}
}
