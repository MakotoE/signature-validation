package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-errors/errors"
)

// downloadToTemp downloads a file from the given URL to a temporary directory.
// It returns the path to the temporary file.
func downloadToTemp(url string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "download-")
	if err != nil {
		return "", errors.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			panic(err)
		}
	}()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Errorf("failed to download file: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return "", errors.Errorf("failed to save file: %w", err)
	}

	return tmpFile.Name(), nil
}

// SignatureInfo represents the output of Get-AuthenticodeSignature.
type SignatureInfo struct {
	SignerCertificate struct {
		NotAfter  PowershellDate
		NotBefore PowershellDate
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
		return SignatureInfo{}, errors.Errorf("powershell command failed: %w", err)
	}

	info := SignatureInfo{}
	if err := json.Unmarshal(out, &info); err != nil {
		return SignatureInfo{}, errors.Errorf("failed to parse json: %w", err)
	}

	return info, nil
}

type PowershellDate time.Time

func (p *PowershellDate) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		return nil
	}

	// PowerShell date format is "/Date(1775779199000)/"
	// We need to extract the number
	var ms int64
	_, err := fmt.Sscanf(s, "\"\\/Date(%d)\\/\"", &ms)
	if err != nil {
		return errors.Errorf("failed to parse PowershellDate: %w", err)
	}

	*p = PowershellDate(time.Unix(0, ms*int64(time.Millisecond)))
	return nil
}

func mainWithError() error {
	// Example usage
	url := "https://download.emeditor.info/emed64_25.4.3.msi"
	path, err := downloadToTemp(url)
	if err != nil {
		return err
	}
	fmt.Printf("File downloaded to: %s\n", path)
	info, err := getSignatureInfo(path)
	if err != nil {
		return err
	}

	fmt.Printf("Signature status: %+v\n", info)

	// Clean up
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := mainWithError(); err != nil {
		var goErr *errors.Error
		if errors.As(err, &goErr) {
			fmt.Printf("Error:\n%s\n", goErr.ErrorStack())
		}
		os.Exit(1)
	}
}
