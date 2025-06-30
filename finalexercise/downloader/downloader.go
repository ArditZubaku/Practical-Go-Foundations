package downloader

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Downloader manages the parallel download process.
type Downloader struct {
	URL           string
	DestFile      string
	NumGoroutines int
	ChunkSize     int64
	Retries       int
	Timeout       time.Duration

	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	file      *os.File // File handle for writing
	fileSize  int64
	etag      string
	errChan   chan error // Channel to propagate errors from goroutines
	mu        sync.Mutex // Mutex for file writing, though os.File.WriteAt is generally safe
	client    *http.Client
	startTime time.Time
}

// Chunk represents a segment of the file to be downloaded.
type Chunk struct {
	ID     int
	Offset int64
	Size   int64
}

// NewDownloader creates and initializes a new Downloader.
func NewDownloader(url, destFile string, numGoroutines int, chunkSize int64, retries int, timeout time.Duration) *Downloader {
	return &Downloader{
		URL:           url,
		DestFile:      destFile,
		NumGoroutines: numGoroutines,
		ChunkSize:     chunkSize,
		Retries:       retries,
		Timeout:       timeout,
		errChan:       make(chan error, numGoroutines), // Buffered to prevent blocking
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Run orchestrates the entire download process.
func (d *Downloader) Run() error {
	d.startTime = time.Now()
	d.ctx, d.cancel = context.WithCancel(context.Background())
	defer d.cancel() // Ensure cancel is called on exit

	log.Printf("Getting metadata for %s...", d.URL)
	if err := d.getMetadata(); err != nil {
		d.cleanup() // Clean up file if metadata fetch fails
		return fmt.Errorf("failed to get file metadata: %w", err)
	}

	log.Printf("File size: %d bytes, ETag: %s", d.fileSize, d.etag)

	log.Printf("Creating empty file %s with size %d bytes...", d.DestFile, d.fileSize)
	if err := d.createEmptyFile(); err != nil {
		d.cleanup()
		return fmt.Errorf("failed to create empty file: %w", err)
	}
	defer d.file.Close() // Close the file when Run exits

	chunks := d.calculateChunks()
	log.Printf("Dividing into %d chunks. Starting parallel download with %d goroutines...", len(chunks), d.NumGoroutines)

	// Start a goroutine to listen for errors and cancel if one occurs
	go func() {
		select {
		case err := <-d.errChan:
			log.Printf("Error received from a goroutine: %v. Cancelling all downloads...", err)
			d.cancel() // Cancel all other goroutines
		case <-d.ctx.Done():
			// Context was cancelled, potentially from an external signal or cleanup
		}
	}()

	// Semaphore to limit the number of concurrent goroutines
	sem := make(chan struct{}, d.NumGoroutines)

	for _, chunk := range chunks {
		d.wg.Add(1)
		sem <- struct{}{} // Acquire a token
		go func(c Chunk) {
			defer func() {
				<-sem // Release the token
				d.wg.Done()
			}()
			d.downloadChunk(c)
		}(chunk)
	}

	d.wg.Wait() // Wait for all download goroutines to finish

	// Check if the context was cancelled due to an error
	select {
	case <-d.ctx.Done():
		if err := d.ctx.Err(); err != nil && !errors.Is(err, context.Canceled) { // context.Canceled is expected if cancel() was called
			log.Printf("Download was cancelled due to an error: %v", err)
			d.cleanup()
			return fmt.Errorf("download interrupted: %w", err)
		} else if errors.Is(err, context.Canceled) {
			// Check if there was an actual error on errChan that caused the cancellation
			select {
			case finalErr := <-d.errChan:
				log.Printf("Download was cancelled due to an error: %v", finalErr)
				d.cleanup()
				return fmt.Errorf("download interrupted: %w", finalErr)
			default:
				// If no error is in errChan, it means it was cancelled externally or finished
				// This case should ideally not be hit if we rely on errChan for internal errors.
				// But including for robustness.
				log.Println("Download context cancelled, but no specific error reported. Assuming clean cancellation or already handled.")
			}
		}
	default:
		// No cancellation, proceed to verify
	}

	// Verify MD5 if ETag was provided (and assumed to be MD5)
	if d.etag != "" {
		log.Println("Verifying file MD5 signature...")
		if err := d.verifyMD5(); err != nil {
			d.cleanup()
			return fmt.Errorf("MD5 verification failed: %w", err)
		}
		log.Println("MD5 verification successful!")
	} else {
		log.Println("No ETag provided, skipping MD5 verification.")
	}

	elapsed := time.Since(d.startTime)
	log.Printf("Total download time: %s", elapsed)

	return nil
}

// getMetadata performs a HEAD request to get file size and ETag.
func (d *Downloader) getMetadata() error {
	req, err := http.NewRequestWithContext(d.ctx, "HEAD", d.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HEAD request returned non-OK status: %s", resp.Status)
	}

	contentLengthStr := resp.Header.Get("Content-Length")
	if contentLengthStr == "" {
		return errors.New("Content-Length header not found")
	}

	fileSize, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid Content-Length: %w", err)
	}
	d.fileSize = fileSize

	d.etag = strings.Trim(resp.Header.Get("ETag"), `"`) // Remove quotes from ETag
	return nil
}

// createEmptyFile creates the destination file and truncates it to the required size.
func (d *Downloader) createEmptyFile() error {
	file, err := os.Create(d.DestFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	d.file = file

	if err := d.file.Truncate(d.fileSize); err != nil {
		d.file.Close() // Close file before returning error
		return fmt.Errorf("failed to truncate file to size %d: %w", d.fileSize, err)
	}
	return nil
}

// calculateChunks determines the offset and size for each chunk.
func (d *Downloader) calculateChunks() []Chunk {
	var chunks []Chunk
	currentOffset := int64(0)
	chunkID := 0

	for currentOffset < d.fileSize {
		chunkSize := d.ChunkSize
		if currentOffset+chunkSize > d.fileSize {
			chunkSize = d.fileSize - currentOffset // Last chunk might be smaller
		}
		chunks = append(chunks, Chunk{
			ID:     chunkID,
			Offset: currentOffset,
			Size:   chunkSize,
		})
		currentOffset += chunkSize
		chunkID++
	}
	return chunks
}

// downloadChunk downloads a specific chunk and writes it to the file.
func (d *Downloader) downloadChunk(chunk Chunk) {
	attempt := 0
	for attempt <= d.Retries {
		select {
		case <-d.ctx.Done():
			log.Printf("Chunk %d: Download cancelled before starting or during retry.", chunk.ID)
			return // Context cancelled, stop this goroutine
		default:
			// Continue
		}

		req, err := http.NewRequestWithContext(d.ctx, "GET", d.URL, nil)
		if err != nil {
			d.reportError(fmt.Errorf("chunk %d: failed to create GET request: %w", chunk.ID, err))
			return
		}

		// Set the Range header
		endByte := chunk.Offset + chunk.Size - 1 // Inclusive end byte
		rangeHeader := fmt.Sprintf("bytes=%d-%d", chunk.Offset, endByte)
		req.Header.Set("Range", rangeHeader)

		log.Printf("Chunk %d: Attempt %d/%d. Downloading range %s...", chunk.ID, attempt+1, d.Retries+1, rangeHeader)

		resp, err := d.client.Do(req)
		if err != nil {
			log.Printf("Chunk %d: Download failed (attempt %d/%d): %v", chunk.ID, attempt+1, d.Retries+1, err)
			attempt++
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("chunk %d: unexpected status code: %s", chunk.ID, resp.Status)
			log.Printf("Chunk %d: Download failed (attempt %d/%d): %v", chunk.ID, attempt+1, d.Retries+1, err)
			attempt++
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		// Read the content and write it to the file

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		n, writeErr := d.file.WriteAt(data, chunk.Offset)
		if writeErr != nil {
			err = fmt.Errorf("chunk %d: failed to write to file: %w", chunk.ID, writeErr)
			log.Printf("Chunk %d: Write failed (attempt %d/%d): %v", chunk.ID, attempt+1, d.Retries+1, err)
			attempt++
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		if int64(n) != chunk.Size {
			err = fmt.Errorf("chunk %d: incomplete write. Expected %d bytes, got %d", chunk.ID, chunk.Size, n)
			log.Printf("Chunk %d: Incomplete write (attempt %d/%d): %v", chunk.ID, attempt+1, d.Retries+1, err)
			attempt++
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		log.Printf("Chunk %d: Downloaded and written %d bytes.", chunk.ID, n)
		return // Success
	}

	// If all retries fail
	d.reportError(fmt.Errorf("chunk %d: failed after %d retries", chunk.ID, d.Retries))
}

// reportError sends an error to the error channel and triggers cancellation.
func (d *Downloader) reportError(err error) {
	select {
	case d.errChan <- err:
		// Error sent successfully
	default:
		// Channel full, likely another error already reported or context cancelled.
		log.Printf("Error channel full, ignoring: %v", err)
	}
	d.cancel() // Ensure cancellation is triggered
}

// verifyMD5 calculates the MD5 hash of the downloaded file and compares it with the ETag.
func (d *Downloader) verifyMD5() error {
	// Ensure the file is flushed to disk before calculating MD5
	if err := d.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file before MD5 verification: %w", err)
	}

	file, err := os.Open(d.DestFile)
	if err != nil {
		return fmt.Errorf("failed to open file for MD5 verification: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate MD5 hash: %w", err)
	}

	calculatedMD5 := hex.EncodeToString(hash.Sum(nil))

	if !strings.EqualFold(calculatedMD5, d.etag) {
		return fmt.Errorf("MD5 mismatch: expected %s, got %s", d.etag, calculatedMD5)
	}
	return nil
}

// cleanup removes the partially downloaded file.
func (d *Downloader) cleanup() {
	if d.file != nil {
		d.file.Close() // Ensure the file handle is closed
	}
	if _, err := os.Stat(d.DestFile); err == nil { // Check if file exists
		log.Printf("Cleaning up partially downloaded file: %s", d.DestFile)
		if err := os.Remove(d.DestFile); err != nil {
			log.Printf("Failed to remove partially downloaded file %s: %v", d.DestFile, err)
		}
	}
}

// GetFilenameFromURL extracts a filename from a URL.
func GetFilenameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	// Extract the base filename
	filename := filepath.Base(u.Path)
	if filename == "." || filename == "/" {
		// If the path is just a slash or dot, try to find a name from query parameters or default
		if q := u.Query().Get("file"); q != "" {
			return q
		}
		if q := u.Query().Get("name"); q != "" {
			return q
		}
		// Fallback if no specific filename can be deduced
		return "downloaded_file" + time.Now().Format("20060102150405")
	}
	return filename
}

// WriteAt is a helper function to write bytes.Buffer content to a file at a specific offset.
// This is used for testing purposes where we want to simulate writing directly from a buffer.
// In the actual download, http.Response.Body is directly streamed to d.file.WriteAt.
func (d *Downloader) WriteAt(buffer *bytes.Buffer, offset int64) (int, error) {
	if d.file == nil {
		return 0, errors.New("file not open")
	}
	return d.file.WriteAt(buffer.Bytes(), offset)
}
