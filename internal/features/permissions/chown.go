package permissions

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"compose-init/internal/config"
)

// Apply fixes permissions based on the provided config
func Apply(cfg *config.ProjectConfig, hostUID, hostGID int) error {
	// 1. Top-level extensions
	for _, chown := range cfg.Chown {
		if err := applyOne(chown, hostUID, hostGID); err != nil {
			return err
		}
	}

	// 2. Service-level extensions
	for _, svc := range cfg.Services {
		for _, chown := range svc.Chown {
			if err := applyOne(chown, hostUID, hostGID); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyOne(c config.ChownConfig, hostUID, hostGID int) error {
	// Resolve Path
	absPath, err := filepath.Abs(c.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve path %s: %w", c.Path, err)
	}

	// Resolve Owner
	uid, gid := hostUID, hostGID
	if c.Owner != "host" && c.Owner != "" {
		parsedUID, parsedGID, err := parseOwner(c.Owner)
		if err != nil {
			return fmt.Errorf("invalid owner format %s: %w", c.Owner, err)
		}
		uid, gid = parsedUID, parsedGID
	}

	// Resolve Modes
	var defaultMode, fileMode, dirMode os.FileMode

	parseMode := func(s string) (os.FileMode, error) {
		if s == "" {
			return 0, nil
		}
		val, err := strconv.ParseUint(s, 8, 32)
		if err != nil {
			return 0, err
		}
		return os.FileMode(val), nil
	}

	var errMode error
	if defaultMode, errMode = parseMode(c.Mode); errMode != nil {
		return fmt.Errorf("invalid mode %s: %w", c.Mode, errMode)
	}
	if fileMode, errMode = parseMode(c.FileMode); errMode != nil {
		return fmt.Errorf("invalid file_mode %s: %w", c.FileMode, errMode)
	}
	if dirMode, errMode = parseMode(c.DirMode); errMode != nil {
		return fmt.Errorf("invalid dir_mode %s: %w", c.DirMode, errMode)
	}

	fmt.Printf("Fixing %s to %d:%d (mode: %s, file: %s, dir: %s, rec: %v)\n",
		absPath, uid, gid, c.Mode, c.FileMode, c.DirMode, c.Recursive)

	// Action
	ensureDir(absPath)

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := os.Lchown(path, uid, gid); err != nil {
			return fmt.Errorf("chown failed on %s: %w", path, err)
		}

		// Determine mode to apply
		targetMode := defaultMode
		if info.IsDir() {
			if dirMode != 0 {
				targetMode = dirMode
			}
		} else {
			if fileMode != 0 {
				targetMode = fileMode
			}
		}

		if targetMode != 0 {
			if err := os.Chmod(path, targetMode); err != nil {
				return fmt.Errorf("chmod failed on %s: %w", path, err)
			}
		}
		return nil
	}

	if c.Recursive {
		return filepath.Walk(absPath, walker)
	}

	// Non-recursive: apply to the target only
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	return walker(absPath, info, nil)
}

func ensureDir(path string) {
	// If path doesn't exist, create it as a directory?
	// The user might be chmoding a file.
	// But usually this is for volumes/mounts which are dirs.
	// Let's check existence first.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create it (owned by root initially, fixed later)
		os.MkdirAll(path, 0755)
	}
}

func parseOwner(owner string) (int, int, error) {
	// Supported: "uid:gid", "uid" (gid=uid or 0?)
	// For simplicity, require "uid:gid" or "uid" (gid=uid)
	// TODO: Support parsing, simple Scanf
	var u, g int
	n, err := fmt.Sscanf(owner, "%d:%d", &u, &g)
	if err == nil && n == 2 {
		return u, g, nil
	}
	n, err = fmt.Sscanf(owner, "%d", &u)
	if err == nil && n == 1 {
		return u, u, nil // Default gid = uid
	}
	return 0, 0, fmt.Errorf("could not parse owner %s", owner)
}
