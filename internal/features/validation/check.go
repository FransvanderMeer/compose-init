package validation

import (
	"fmt"
	"os"
	"strings"

	"compose-init/internal/config"
)

func Check(cfg *config.ProjectConfig) error {
	var missing []string

	for _, key := range cfg.RequiredEnv {
		if _, ok := os.LookupEnv(key); !ok {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if len(cfg.RequiredEnv) > 0 {
		fmt.Printf("Environment validation passed (%d variables checked)\n", len(cfg.RequiredEnv))
	}

	return nil
}
