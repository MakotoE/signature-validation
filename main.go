package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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

// SignatureInfo represents the output of Get-AuthenticodeSignature.
type SignatureInfo struct {
	SignerCertificate struct {
		NotAfter  string
		NotBefore string
		Subject   string
	}
	Status        int
	StatusMessage string
	Path          string
}

// getSignatureInfo calls Get-AuthenticodeSignature and returns the parsed info.
func getSignatureInfo(filePath string) (SignatureInfo, error) {
	cmdStr := fmt.Sprintf("Get-AuthenticodeSignature '%s' | Select-Object SignerCertificate, Status, StatusMessage | ConvertTo-Json", filePath)
	out, err := exec.Command("powershell", "-Command", cmdStr).Output()
	if err != nil {
		return SignatureInfo{}, fmt.Errorf("powershell command failed: %w", err)
	}

	info := SignatureInfo{}
	if err := json.Unmarshal(out, &info); err != nil {
		return SignatureInfo{}, fmt.Errorf("failed to parse json: %w", err)
	}

	return info, nil
}

func main() {
	// Example usage
	url := "https://download.emeditor.info/emed64_25.4.3.msi"
	path, err := downloadToTemp(url)
	if err != nil {
		panic(err)
	}
	fmt.Printf("File downloaded to: %s\n", path)
	info, err := getSignatureInfo(path)
	if err != nil {
		fmt.Printf("Error checking signature: %v\n", err)
		return
	}

	fmt.Printf("Signature status: %+v\n", info)

	// Clean up
	if err := os.Remove(path); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
