package downloader_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ArditZubaku/parallel-downloader/downloader"
)

// setupTestServer creates an HTTP test server that responds with file content.
// It can simulate errors or specific ranges.
func setupTestServer(t *testing.T, content []byte, etag string, simulateError bool) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if simulateError {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Simulated error"))
			return
		}

		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
			if etag != "" {
				w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			rangeHeader := r.Header.Get("Range")
			if rangeHeader == "" {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
				if etag != "" {
					w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
				}
				_, _ = w.Write(content)
				return
			}

			// Parse range header: "bytes=start-end"
			parts := strings.Split(rangeHeader, "=")
			if len(parts) != 2 || parts[0] != "bytes" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			rangeParts := strings.Split(parts[1], "-")
			if len(rangeParts) != 2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			startStr, endStr := rangeParts[0], rangeParts[1]
			start, err := strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			end, err := strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if start >= int64(len(content)) || end >= int64(len(content)) || start > end {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}

			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(content)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write(content[start : end+1])
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	return server
}

// TestGetMetadata tests the getMetadata function.
func TestGetMetadata(t *testing.T) {
	testContent := []byte("This is a test file content for metadata.")
	testEtag := "testetag123"
	server := setupTestServer(t, testContent, testEtag, false)
	defer server.Close()

	d := downloader.NewDownloader(server.URL, "test_output.txt", 1, 1024, 0, 5*time.Second)
	// Initialize context for metadata fetch
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d.SetContext(ctx, cancel) // Helper to set private fields for test

	err := d.RunMetadataFetchOnly() // Separate method for testing metadata fetch
	if err != nil {
		t.Fatalf("getMetadata failed: %v", err)
	}

	if d.GetFileSize() != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), d.GetFileSize())
	}
	if d.GetEtag() != testEtag {
		t.Errorf("Expected ETag %q, got %q", testEtag, d.GetEtag())
	}
}

// Add a helper method to Downloader for testing metadata fetch in isolation
func (d *downloader.Downloader) RunMetadataFetchOnly() error {
	d.ctx, d.cancel = context.WithCancel(context.Background())
	defer d.cancel()
	return d.GetMetadata()
}

// Add getters for private fields for testing
func (d *downloader.Downloader) GetFileSize() int64 {
	return d.fileSize
}

func (d *downloader.Downloader) GetEtag() string {
	return d.etag
}

func (d *downloader.Downloader) SetContext(ctx context.Context, cancel context.CancelFunc) {
	d.ctx = ctx
	d.cancel = cancel
}

// TestCreateEmptyFile tests the createEmptyFile function.
func TestCreateEmptyFile(t *testing.T) {
	destFile := "test_empty_file.tmp"
	fileSize := int64(100)

	d := downloader.NewDownloader("", destFile, 1, 10, 0, 5*time.Second)
	d.SetFileSize(fileSize) // Helper to set private field for test

	err := d.RunCreateEmptyFileOnly()
	if err != nil {
		t.Fatalf("createEmptyFile failed: %v", err)
	}
	defer os.Remove(destFile) // Clean up file
	defer d.CloseFile()       // Close the file handle used by the downloader

	stat, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("Failed to stat created file: %v", err)
	}
	if stat.Size() != fileSize {
		t.Errorf("Expected file size %d, got %d", fileSize, stat.Size())
	}
}

// Add helper methods to Downloader for testing createEmptyFile in isolation
func (d *downloader.Downloader) SetFileSize(size int64) {
	d.fileSize = size
}

func (d *downloader.Downloader) RunCreateEmptyFileOnly() error {
	return d.CreateEmptyFile()
}

func (d *downloader.Downloader) CloseFile() {
	if d.file != nil {
		d.file.Close()
		d.file = nil
	}
}

// TestCalculateChunks tests the calculateChunks function.
func TestCalculateChunks(t *testing.T) {
	tests := []struct {
		fileSize  int64
		chunkSize int64
		expected  []downloader.Chunk
	}{
		{
			fileSize:  10,
			chunkSize: 3,
			expected: []downloader.Chunk{
				{ID: 0, Offset: 0, Size: 3},
				{ID: 1, Offset: 3, Size: 3},
				{ID: 2, Offset: 6, Size: 3},
				{ID: 3, Offset: 9, Size: 1},
			},
		},
		{
			fileSize:  10,
			chunkSize: 10,
			expected: []downloader.Chunk{
				{ID: 0, Offset: 0, Size: 10},
			},
		},
		{
			fileSize:  0,
			chunkSize: 100,
			expected:  []downloader.Chunk{},
		},
		{
			fileSize:  1,
			chunkSize: 100,
			expected: []downloader.Chunk{
				{ID: 0, Offset: 0, Size: 1},
			},
		},
	}

	for _, tt := range tests {
		d := downloader.NewDownloader("", "", 1, tt.chunkSize, 0, 0)
		d.SetFileSize(tt.fileSize) // Helper to set private field for test
		chunks := d.CalculateChunks()

		if len(chunks) != len(tt.expected) {
			t.Fatalf("For fileSize %d, chunkSize %d: Expected %d chunks, got %d",
				tt.fileSize, tt.chunkSize, len(tt.expected), len(chunks))
		}

		for i, chunk := range chunks {
			if chunk != tt.expected[i] {
				t.Errorf("For chunk %d, expected %+v, got %+v", i, tt.expected[i], chunk)
			}
		}
	}
}

// Helper methods for testing calculateChunks
func (d *downloader.Downloader) SetEtag(etag string) {
	d.etag = etag
}

func (d *downloader.Downloader) CalculateChunks() []downloader.Chunk {
	// Expose the private method for testing
	return d.calculateChunks()
}

// TestDownloadChunkSuccess tests a single chunk download.
func TestDownloadChunkSuccess(t *testing.T) {
	testContent := []byte("0123456789abcdefghijklmnopqrstuvwxyz")
	testChunk := downloader.Chunk{ID: 0, Offset: 5, Size: 10} // "56789abcde"
	destFile := "test_chunk_success.tmp"

	server := setupTestServer(t, testContent, "", false)
	defer server.Close()

	d := downloader.NewDownloader(server.URL, destFile, 1, 10, 0, 5*time.Second)
	d.SetFileSize(int64(len(testContent)))

	// Simulate setup for download, including creating the file
	err := d.RunCreateEmptyFileOnly()
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	defer os.Remove(destFile)
	defer d.CloseFile()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetContext(ctx, cancel)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.DownloadChunk(testChunk) // Expose for testing
	}()
	wg.Wait()

	// Check if any error was reported through errChan (should not be for success)
	select {
	case err := <-d.GetErrChan():
		t.Fatalf("Unexpected error received: %v", err)
	default:
		// No error
	}

	// Read content from the file and verify the chunk
	fileContent := make([]byte, testChunk.Size)
	n, err := d.GetFile().ReadAt(fileContent, testChunk.Offset)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read from file: %v", err)
	}

	expectedChunkContent := testContent[testChunk.Offset : testChunk.Offset+testChunk.Size]
	if !bytes.Equal(fileContent[:n], expectedChunkContent) {
		t.Errorf("Downloaded chunk mismatch:\nExpected: %s\nGot: %s",
			string(expectedChunkContent), string(fileContent[:n]))
	}
}

// TestDownloadChunkRetry tests the retry mechanism for chunk downloads.
func TestDownloadChunkRetry(t *testing.T) {
	testContent := []byte("0123456789")
	testChunk := downloader.Chunk{ID: 0, Offset: 0, Size: 10}
	destFile := "test_chunk_retry.tmp"
	failCount := 0
	maxFails := 1 // Server will fail once, then succeed

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "GET" {
			if failCount < maxFails {
				failCount++
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Simulated transient error"))
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(testContent)
			return
		}
	}))
	defer server.Close()

	d := downloader.NewDownloader(server.URL, destFile, 1, 10, 2, 5*time.Second) // 2 retries means 3 attempts total
	d.SetFileSize(int64(len(testContent)))

	err := d.RunCreateEmptyFileOnly()
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	defer os.Remove(destFile)
	defer d.CloseFile()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetContext(ctx, cancel)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.DownloadChunk(testChunk)
	}()
	wg.Wait()

	// Ensure no error was sent to errChan, as it should have succeeded on retry
	select {
	case err := <-d.GetErrChan():
		t.Fatalf("Unexpected error received: %v", err)
	default:
		// No error, good
	}

	// Verify content
	readContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if !bytes.Equal(readContent, testContent) {
		t.Errorf("Downloaded content mismatch:\nExpected: %s\nGot: %s",
			string(testContent), string(readContent))
	}
}

// TestDownloadChunkFailureAfterRetries tests if chunk download fails after all retries.
func TestDownloadChunkFailureAfterRetries(t *testing.T) {
	testContent := []byte("0123456789")
	testChunk := downloader.Chunk{ID: 0, Offset: 0, Size: 10}
	destFile := "test_chunk_fail_retries.tmp"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
			w.WriteHeader(http.StatusOK)
			return
		}
		// Always fail GET requests
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Always failing"))
	}))
	defer server.Close()

	d := downloader.NewDownloader(server.URL, destFile, 1, 10, 1, 5*time.Second) // 1 retry means 2 attempts total
	d.SetFileSize(int64(len(testContent)))

	err := d.RunCreateEmptyFileOnly()
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}
	defer os.Remove(destFile)
	defer d.CloseFile()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetContext(ctx, cancel)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.DownloadChunk(testChunk)
	}()
	wg.Wait()

	// Expect an error to be sent to errChan
	select {
	case err := <-d.GetErrChan():
		if !strings.Contains(err.Error(), "failed after 1 retries") {
			t.Errorf("Expected 'failed after 1 retries' error, got: %v", err)
		}
	case <-time.After(2 * time.Second): // Give some time for error to propagate
		t.Fatal("Expected an error but none was received within timeout.")
	}
}

// Add helper methods to Downloader for testing downloadChunk
func (d *downloader.Downloader) DownloadChunk(chunk downloader.Chunk) {
	// Expose the private method for testing
	d.downloadChunk(chunk)
}

func (d *downloader.Downloader) GetErrChan() <-chan error {
	return d.errChan
}

func (d *downloader.Downloader) GetFile() *os.File {
	return d.file
}

// TestParallelDownloadFull tests the full parallel download process.
func TestParallelDownloadFull(t *testing.T) {
	testContent := make([]byte, 1024*1024*2) // 2MB file
	for i := 0; i < len(testContent); i++ {
		testContent[i] = byte(i % 256)
	}
	testEtag := "d41d8cd98f00b204e9800998ecf8427e" // MD5 of an empty string for simplicity in ETag
	destFile := "test_full_download.tmp"

	server := setupTestServer(t, testContent, testEtag, false)
	defer server.Close()

	d := downloader.NewDownloader(server.URL, destFile, 4, 512*1024, 1, 10*time.Second) // 4 goroutines, 512KB chunks
	err := d.Run()
	if err != nil {
		t.Fatalf("Parallel download failed: %v", err)
	}
	defer os.Remove(destFile)

	downloadedContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if !bytes.Equal(downloadedContent, testContent) {
		t.Errorf("Downloaded file content mismatch. Expected %d bytes, got %d bytes.", len(testContent), len(downloadedContent))
		// Optionally, print a diff or just say sizes are different for large files
	}

	// Additional check for file size
	stat, err := os.Stat(destFile)
	if err != nil {
		t.Fatalf("Failed to stat downloaded file: %v", err)
	}
	if stat.Size() != int64(len(testContent)) {
		t.Errorf("Downloaded file size mismatch. Expected %d, got %d", len(testContent), stat.Size())
	}
}

// TestCleanupOnError tests if the partially downloaded file is removed on error.
func TestCleanupOnError(t *testing.T) {
	testContent := make([]byte, 1024)
	destFile := "test_cleanup_on_error.tmp"

	// Server that always returns an error for GET requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := downloader.NewDownloader(server.URL, destFile, 1, 512, 0, 5*time.Second) // No retries for quick failure

	// Create an empty file to simulate partial download
	initialFile, err := os.Create(destFile)
	if err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}
	_, err = initialFile.WriteString("partial content") // Write some content
	initialFile.Close()
	if err != nil {
		t.Fatalf("Failed to write to initial file: %v", err)
	}

	err = d.Run()
	if err == nil {
		t.Fatal("Expected download to fail, but it succeeded.")
	}

	// Check if the file was removed
	_, err = os.Stat(destFile)
	if !os.IsNotExist(err) {
		t.Errorf("Expected file %s to be removed, but it still exists or other error: %v", destFile, err)
	}
}

// TestGetFilenameFromURL
func TestGetFilenameFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://example.com/path/to/file.zip", "file.zip"},
		{"https://d37ci6vzurychx.cloudfront.net/trip-data/yellow_tripdata_2018-05.parquet", "yellow_tripdata_2018-05.parquet"},
		{"http://example.com/no-extension", "no-extension"},
		{"http://example.com/", "downloaded_file"}, // Default fallback
		{"http://example.com/path/", "path"},       // Should still get the last path segment
		{"http://example.com?file=mydata.txt", "mydata.txt"},
		{"http://example.com?name=document.pdf", "document.pdf"},
		{"http://example.com/query?id=123", "query"}, // Should prioritize path over generic query
		{"", ""}, // Invalid URL
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := downloader.GetFilenameFromURL(tt.url)
			if tt.expected == "downloaded_file" && !strings.HasPrefix(result, tt.expected) {
				t.Errorf("For URL %s, expected prefix %s, got %s", tt.url, tt.expected, result)
			} else if tt.expected != "downloaded_file" && result != tt.expected {
				t.Errorf("For URL %s, expected %s, got %s", tt.url, tt.expected, result)
			}
		})
	}
}
