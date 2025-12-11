package config

// ProjectConfig represents the parsed docker-compose configuration
type ProjectConfig struct {
	Services map[string]Service `yaml:"services"`
	Volumes  map[string]Volume  `yaml:"volumes"`

	// Top-level extensions
	Template     []TemplateItem `yaml:"x-template,omitempty"`
	RequiredEnv  []string       `yaml:"x-required-env,omitempty"`
	GenerateCert []CertConfig   `yaml:"x-generate-cert,omitempty"`
	Fetch        []FetchItem    `yaml:"x-fetch,omitempty"`
	Chown        []ChownConfig  `yaml:"x-chown,omitempty"`
}

type Service struct {
	Image       string            `yaml:"image"`
	Volumes     []VolumeMount     `yaml:"volumes"`
	Environment map[string]string `yaml:"environment,omitempty"`

	// Service-level extensions
	Chown        []ChownConfig  `yaml:"x-chown,omitempty"`
	Template     []TemplateItem `yaml:"x-template,omitempty"`
	RequiredEnv  []string       `yaml:"x-required-env,omitempty"`
	GenerateCert []CertConfig   `yaml:"x-generate-cert,omitempty"`
	Fetch        []FetchItem    `yaml:"x-fetch,omitempty"`
}

// VolumeMount handles the complex volume syntax or simple string
type VolumeMount struct {
	Type   string `yaml:"type"`
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type Volume struct {
	Name   string `yaml:"name"` // Physical name
	Driver string `yaml:"driver"`

	// Volume-level extensions
	Chown *ChownConfig `yaml:"x-chown,omitempty"`
}

// Extension Types

type ChownConfig struct {
	Path      string `yaml:"path"`  // For service-level: target path. For volume-level: relative to volume root? Or just applied to volume?
	Owner     string `yaml:"owner"` // "host" or "uid:gid"
	Mode      string `yaml:"mode"`  // "0755"
	Recursive bool   `yaml:"recursive"`
}

type TemplateItem struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type CertConfig struct {
	Domain    string `yaml:"domain"`
	OutputDir string `yaml:"output_dir"`
	CertName  string `yaml:"cert_name"`
	KeyName   string `yaml:"key_name"`
	Force     bool   `yaml:"force"`
}

type FetchItem struct {
	URL     string `yaml:"url"`
	Dest    string `yaml:"dest"`
	SHA256  string `yaml:"sha256"`
	Force   bool   `yaml:"force"`
	Retries int    `yaml:"retries"`
	Extract bool   `yaml:"extract"`
}
