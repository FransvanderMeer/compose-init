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

	// Resolve Mode
	var fileMode os.FileMode
	if c.Mode != "" {
		val, err := strconv.ParseUint(c.Mode, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid mode format %s: %w", c.Mode, err)
		}
		fileMode = os.FileMode(val)
	}

	fmt.Printf("Fixing %s to %d:%d (mode: %s, recursive: %v)\n", absPath, uid, gid, c.Mode, c.Recursive)

	// Action
	ensureDir(absPath)

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := os.Lchown(path, uid, gid); err != nil {
			return fmt.Errorf("chown failed on %s: %w", path, err)
		}
		if c.Mode != "" {
			// Only apply mode to directories if it looks like a directory mode (executable bit set?)
			// Or just apply blindly? The user asked for it.
			// Usually we want different modes for files vs dirs (e.g. 644 vs 755).
			// If the user specifies 755, applying to files makes them executable.
			// For now, let's apply blindly as requested, user beware.
			if err := os.Chmod(path, fileMode); err != nil {
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
