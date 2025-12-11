package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"compose-init/internal/config"
)

func Apply(cfg *config.ProjectConfig) error {
	// Top-level
	for _, tpl := range cfg.Template {
		if err := renderOne(tpl); err != nil {
			return fmt.Errorf("failed to render template %s: %w", tpl.Source, err)
		}
	}

	// Service-level
	for _, svc := range cfg.Services {
		for _, tpl := range svc.Template {
			if err := renderOne(tpl); err != nil {
				return fmt.Errorf("failed to render template %s: %w", tpl.Source, err)
			}
		}
	}
	return nil
}

func renderOne(t config.TemplateItem) error {
	srcPath, err := filepath.Abs(t.Source)
	if err != nil {
		return err
	}
	targetPath, err := filepath.Abs(t.Target)
	if err != nil {
		return err
	}

	fmt.Printf("Rendering template %s -> %s\n", srcPath, targetPath)

	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(content))
	if err != nil {
		return err
	}

	// Prepare environment map
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}

	// Create target directory if needed
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, env); err != nil {
		return err
	}

	return nil
}
