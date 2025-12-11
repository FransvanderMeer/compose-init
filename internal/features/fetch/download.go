package fetch

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"compose-init/internal/config"
)

func Apply(cfg *config.ProjectConfig) error {
	// Top-level
	for _, item := range cfg.Fetch {
		if err := fetchOne(item); err != nil {
			return fmt.Errorf("failed to fetch %s: %w", item.URL, err)
		}
	}

	// Service-level
	for _, svc := range cfg.Services {
		for _, item := range svc.Fetch {
			if err := fetchOne(item); err != nil {
				return fmt.Errorf("failed to fetch %s: %w", item.URL, err)
			}
		}
	}
	return nil
}

func fetchOne(f config.FetchItem) error {
	destPath, err := filepath.Abs(f.Dest)
	if err != nil {
		return err
	}

	// Force check
	if !f.Force {
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("File %s already exists, skipping download.\n", destPath)
			return nil
		}
	} else {
		fmt.Printf("Forcing download of %s\n", destPath)
	}

	fmt.Printf("Downloading %s -> %s\n", f.URL, destPath)

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Retry Loop
	var lastErr error
	maxAttempts := f.Retries + 1
	for i := 0; i < maxAttempts; i++ {
		if i > 0 {
			fmt.Printf("Retry %d/%d for %s...\n", i, f.Retries, f.URL)
			time.Sleep(time.Duration(i) * time.Second) // Simple backoff
		}

		lastErr = download(f, destPath)
		if lastErr == nil {
			break
		}
		fmt.Printf("Download failed: %v\n", lastErr)
	}

	if lastErr != nil {
		return fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
	}

	// Extract Logic
	if f.Extract {
		if err := extract(destPath); err != nil {
			return fmt.Errorf("failed to extract %s: %w", destPath, err)
		}
	}

	return nil
}

func download(f config.FetchItem, destPath string) error {
	resp, err := http.Get(f.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	var writer io.Writer = out
	var hasher = sha256.New()

	if f.SHA256 != "" {
		writer = io.MultiWriter(out, hasher)
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return err
	}

	if f.SHA256 != "" {
		sum := fmt.Sprintf("%x", hasher.Sum(nil))
		if sum != f.SHA256 {
			os.Remove(destPath)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", f.SHA256, sum)
		}
		fmt.Println("Checksum verified.")
	}
	return nil
}

func extract(archivePath string) error {
	destDir := filepath.Dir(archivePath)
	fmt.Printf("Extracting %s to %s\n", archivePath, destDir)

	cmd := exec.Command("unzip", "-o", archivePath, "-d", destDir)
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		cmd = exec.Command("tar", "-xzf", archivePath, "-C", destDir)
	} else if strings.HasSuffix(archivePath, ".tar") {
		cmd = exec.Command("tar", "-xf", archivePath, "-C", destDir)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("extract failed: %s: %w", string(out), err)
	}
	return nil
}
