package util

import (
	"fmt"
	"os"
	"syscall"
)

// DetectFileOwner returns the UID and GID of the given file.
func DetectFileOwner(path string) (int, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, fmt.Errorf("failed to get sys stats for %s", path)
	}

	return int(stat.Uid), int(stat.Gid), nil
}
