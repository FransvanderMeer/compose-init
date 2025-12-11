package validation

import (
	"fmt"
	"os"
	"strings"

	"compose-init/internal/config"
)

func Check(cfg *config.ProjectConfig) error {
	var missing []string
	count := 0

	// Helper to check keys
	checkKeys := func(keys []string) {
		for _, key := range keys {
			count++
			if _, ok := os.LookupEnv(key); !ok {
				missing = append(missing, key)
			}
		}
	}

	// Top-level
	checkKeys(cfg.RequiredEnv)

	// Service-level
	for _, svc := range cfg.Services {
		checkKeys(svc.RequiredEnv)
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if count > 0 {
		fmt.Printf("Environment validation passed (%d variables checked)\n", count)
	}

	return nil
}
