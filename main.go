package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// downloadToTemp downloads a file from the given URL to a temporary directory.
// It returns the path to the temporary file.
func downloadToTemp(url string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "download-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			panic(err)
		}
	}()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return tmpFile.Name(), nil
}

func main() {
	// Example usage
	url := "https://www.google.com"
	path, err := downloadToTemp(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("File downloaded to: %s\n", path)

	// Clean up for demo purposes
	// os.Remove(path)
}
