package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"compose-init/internal/config"
	"compose-init/internal/util"

	"compose-init/internal/features/fetch"
	"compose-init/internal/features/permissions"
	"compose-init/internal/features/ssl"
	"compose-init/internal/features/templates"
	"compose-init/internal/features/validation"
)

var projectDir string

var rootCmd = &cobra.Command{
	Use:   "compose-init",
	Short: "Initialize Docker Compose environment",
	Long:  `A robust initialization tool that prepares the environment for a Docker Compose project by handling permissions, templating, validation, SSL, and resource fetching.`,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	rootCmd.Flags().StringVar(&projectDir, "project-dir", ".", "Directory containing compose.yaml")
}

func run() {
	fmt.Printf("Starting compose-init in %s\n", projectDir)

	// 1. Detect Host Owner
	hostUID, hostGID := detectOwner(projectDir)
	fmt.Printf("Detected Host Owner: %d:%d\n", hostUID, hostGID)

	// 2. Load Configuration
	cfg, err := config.Load(projectDir)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 3. Validation
	if err := validation.Check(cfg); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	// 4. Resource Fetching
	if err := fetch.Apply(cfg); err != nil {
		fmt.Printf("Fetch failed: %v\n", err)
		os.Exit(1)
	}

	// 5. Templating
	if err := templates.Apply(cfg); err != nil {
		fmt.Printf("Templating failed: %v\n", err)
		os.Exit(1)
	}

	// 6. SSL
	if err := ssl.Apply(cfg); err != nil {
		fmt.Printf("SSL generation failed: %v\n", err)
		os.Exit(1)
	}

	// 7. Permissions
	if err := permissions.Apply(cfg, hostUID, hostGID); err != nil {
		fmt.Printf("Permission fix failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Initialization complete.")
}

func detectOwner(dir string) (int, int) {
	candidates := []string{"compose.yaml", "docker-compose.yml", "docker-compose.yaml"}
	for _, f := range candidates {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err == nil {
			u, g, err := util.DetectFileOwner(path)
			if err == nil {
				return u, g
			}
		}
	}
	// Fallback to 0:0 if no file found (unlikely if loading succeeds) or stat fails
	return 0, 0
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
