// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ArditZubaku/parallel-downloader/downloader"
	"github.com/urfave/cli/v2"
)

/*
Final Exercise - Parallel Downloader

Write a command line program that will download a file over HTTP in parallel.

First, issue a HEAD request to the URL and get the file size from the Content-Length HTTP header and the file MD5 signature from the ETag HTTP header.

Then create an empty file with the required size and spin n goroutines that will download the file in parallel.
Each goroutine will get the URL, destination file name, offset and size to download. It’ll use an HTTP range request to download a chunk and write it to the correct section of the file.

You can use https://d37ci6vzurychx.cloudfront.net/trip-data/yellow_tripdata_2018-05.parquet to test your program.
Possible Extensions

    Test everything
    Add a command line parameter to limit the number of downloading goroutintes
    Add a command line parameter to set the chunk size
    Support retrying of a failed download
        Add command line parameter to control number of retries
    Add connection timeout
    Cancel all goroutines on error & delete the file
*/

func main() {
	app := &cli.App{
		Name:  "parallel-downloader",
		Usage: "Download files over HTTP in parallel",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "url",
				Aliases:  []string{"u"},
				Usage:    "URL of the file to download",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file name",
				Value:   "", // Default will be derived from URL
			},
			&cli.IntFlag{
				Name:    "goroutines",
				Aliases: []string{"g"},
				Usage:   "Number of parallel downloading goroutines",
				Value:   4, // Default to 4 goroutines
			},
			&cli.Int64Flag{
				Name:    "chunk-size",
				Aliases: []string{"c"},
				Usage:   "Size of each download chunk in bytes",
				Value:   1024 * 1024 * 5, // Default to 5MB chunks
			},
			&cli.IntFlag{
				Name:    "retries",
				Aliases: []string{"r"},
				Usage:   "Number of retries for a failed chunk download",
				Value:   3, // Default to 3 retries
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Usage:   "Connection timeout for each HTTP request (e.g., 10s)",
				Value:   30 * time.Second, // Default to 30 seconds timeout
			},
		},
		Action: func(c *cli.Context) error {
			url := c.String("url")
			output := c.String("output")
			numGoroutines := c.Int("goroutines")
			chunkSize := c.Int64("chunk-size")
			retries := c.Int("retries")
			timeout := c.Duration("timeout")

			// If output filename is not provided, derive it from the URL
			if output == "" {
				output = downloader.GetFilenameFromURL(url)
				if output == "" {
					log.Fatalf("Could not determine output filename from URL. Please specify with -o flag.")
				}
				fmt.Printf("Output filename not specified, using: %s\n", output)
			}

			fmt.Printf("Starting download for %s to %s...\n", url, output)
			fmt.Printf("Goroutines: %d, Chunk Size: %d bytes, Retries: %d, Timeout: %s\n",
				numGoroutines, chunkSize, retries, timeout)

			dl := downloader.NewDownloader(url, output, numGoroutines, chunkSize, retries, timeout)
			if err := dl.Run(); err != nil {
				log.Fatalf("Download failed: %v", err)
			}

			fmt.Println("Download completed successfully!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
