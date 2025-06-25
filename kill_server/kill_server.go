package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

func main() {
	// fmt.Println(killServer("server.pid"))
	err := killServer("server.pid")
	if err != nil {
		fmt.Println("ERROR:", err)
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("File not found")
		}
		for e := err; e != nil; e = errors.Unwrap(e) {
			fmt.Println(">", e)
		}
	}
}

func killServer(pidFile string) error {
	file, err := os.Open(pidFile)
	if err != nil {
		return err
	}
	// defer happens when function exits, no matter what (even panics)
	// defer works at the function level
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("Failed to close file", "file", file.Name())
		}
	}()

	pid := new(int)
	if _, err := fmt.Fscanf(file, "%d", pid); err != nil {
		return fmt.Errorf("%q - bad pid: %w", pidFile, err)
	}

	slog.Info("killing", "pid", pid)
	if err := os.Remove(pidFile); err != nil {
		slog.Warn("deleting", "file", pidFile, "error", err)
	}

	return nil
}
