package fetch

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"compose-init/internal/config"
)

func Apply(cfg *config.ProjectConfig) error {
	for _, item := range cfg.Fetch {
		if err := fetchOne(item); err != nil {
			return fmt.Errorf("failed to fetch %s: %w", item.URL, err)
		}
	}
	return nil
}

func fetchOne(f config.FetchItem) error {
	destPath, err := filepath.Abs(f.Dest)
	if err != nil {
		return err
	}

	// Skip if exists? Or overwrite?
	// Usually fetch is for initial setup. If exists, maybe check hash?
	if _, err := os.Stat(destPath); err == nil {
		fmt.Printf("File %s already exists, skipping download.\n", destPath)
		return nil
	}

	fmt.Printf("Downloading %s -> %s\n", f.URL, destPath)

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

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

	// If SHA256 provided, we need to hash while writing
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
			// Cleanup
			os.Remove(destPath)
			return fmt.Errorf("checksum mismatch: expected %s, got %s", f.SHA256, sum)
		}
		fmt.Println("Checksum verified.")
	}

	return nil
}
