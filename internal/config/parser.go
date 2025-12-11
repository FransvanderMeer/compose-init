package config

import (
	"bytes"
	"fmt"
	"os/exec"

	"gopkg.in/yaml.v3"
)

// Load runs `docker compose config` and parses the output
func Load(projectDir string) (*ProjectConfig, error) {
	// We assume docker is available in PATH
	cmd := exec.Command("docker", "compose", "config")
	cmd.Dir = projectDir

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run docker compose config: %w\nStderr: %s", err, stderr.String())
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(out.Bytes(), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse compose config: %w", err)
	}

	return &cfg, nil
}
