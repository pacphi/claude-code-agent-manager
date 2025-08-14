package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration
type Config struct {
	Version  string   `yaml:"version"`
	Settings Settings `yaml:"settings"`
	Sources  []Source `yaml:"sources"`
	Metadata Metadata `yaml:"metadata"`
}

// Settings contains global settings
type Settings struct {
	BaseDir             string        `yaml:"base_dir"`
	DocsDir             string        `yaml:"docs_dir"`
	ConflictStrategy    string        `yaml:"conflict_strategy"`
	BackupDir           string        `yaml:"backup_dir"`
	LogLevel            string        `yaml:"log_level"`
	ConcurrentDownloads int           `yaml:"concurrent_downloads"`
	Timeout             time.Duration `yaml:"timeout"`
	ContinueOnError     bool          `yaml:"continue_on_error"`
}

// Source represents an agent source
type Source struct {
	Name             string           `yaml:"name"`
	Enabled          bool             `yaml:"enabled"`
	Type             string           `yaml:"type"`
	Repository       string           `yaml:"repository,omitempty"`
	URL              string           `yaml:"url,omitempty"`
	Branch           string           `yaml:"branch,omitempty"`
	Auth             AuthConfig       `yaml:"auth,omitempty"`
	Paths            PathConfig       `yaml:"paths"`
	Filters          FilterConfig     `yaml:"filters,omitempty"`
	Transformations  []Transformation `yaml:"transformations,omitempty"`
	PostInstall      []PostInstall    `yaml:"post_install,omitempty"`
	ConflictStrategy string           `yaml:"conflict_strategy,omitempty"`
	Watch            bool             `yaml:"watch,omitempty"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Method   string `yaml:"method,omitempty"`
	TokenEnv string `yaml:"token_env,omitempty"`
	SSHKey   string `yaml:"ssh_key,omitempty"`
}

// PathConfig contains source and target paths
type PathConfig struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// FilterConfig contains include/exclude filters
type FilterConfig struct {
	Include IncludeFilter `yaml:"include,omitempty"`
	Exclude ExcludeFilter `yaml:"exclude,omitempty"`
}

// IncludeFilter contains inclusion rules
type IncludeFilter struct {
	Extensions []string `yaml:"extensions,omitempty"`
	Patterns   []string `yaml:"patterns,omitempty"`
	Regex      []string `yaml:"regex,omitempty"`
}

// ExcludeFilter contains exclusion rules
type ExcludeFilter struct {
	Patterns []string `yaml:"patterns,omitempty"`
}

// Transformation represents a file transformation
type Transformation struct {
	Type          string   `yaml:"type"`
	Pattern       string   `yaml:"pattern,omitempty"`
	SourcePattern string   `yaml:"source_pattern,omitempty"`
	TargetDir     string   `yaml:"target_dir,omitempty"`
	Naming        string   `yaml:"naming,omitempty"`
	Script        string   `yaml:"script,omitempty"`
	Args          []string `yaml:"args,omitempty"`
}

// PostInstall represents a post-installation action
type PostInstall struct {
	Type string   `yaml:"type"`
	Path string   `yaml:"path,omitempty"`
	Args []string `yaml:"args,omitempty"`
}

// Metadata contains tracking and logging configuration
type Metadata struct {
	TrackingFile string `yaml:"tracking_file"`
	LogFile      string `yaml:"log_file"`
	LockFile     string `yaml:"lock_file,omitempty"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML with variable substitution
	data = substituteVariables(data)

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Apply defaults
	applyDefaults(&cfg)

	return &cfg, nil
}

// substituteVariables replaces ${variable} patterns in the configuration
func substituteVariables(data []byte) []byte {
	content := string(data)

	// First pass: collect all settings values
	var tempCfg struct {
		Settings map[string]interface{} `yaml:"settings"`
	}
	yaml.Unmarshal(data, &tempCfg)

	// Replace ${settings.*} variables
	if tempCfg.Settings != nil {
		for key, value := range tempCfg.Settings {
			placeholder := fmt.Sprintf("${settings.%s}", key)
			if strVal, ok := value.(string); ok {
				content = strings.ReplaceAll(content, placeholder, strVal)
			}
		}
	}

	// Replace ${env.*} variables
	re := regexp.MustCompile(`\$\{env\.([^}]+)\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		envVar := re.FindStringSubmatch(match)[1]
		if val := os.Getenv(envVar); val != "" {
			return val
		}
		return match
	})

	return []byte(content)
}

// applyDefaults sets default values for missing configuration
func applyDefaults(cfg *Config) {
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}

	if cfg.Settings.BaseDir == "" {
		cfg.Settings.BaseDir = ".claude/agents"
	}

	if cfg.Settings.DocsDir == "" {
		cfg.Settings.DocsDir = "docs"
	}

	if cfg.Settings.ConflictStrategy == "" {
		cfg.Settings.ConflictStrategy = "backup"
	}

	if cfg.Settings.BackupDir == "" {
		cfg.Settings.BackupDir = ".claude/backups"
	}

	if cfg.Settings.LogLevel == "" {
		cfg.Settings.LogLevel = "info"
	}

	if cfg.Settings.ConcurrentDownloads == 0 {
		cfg.Settings.ConcurrentDownloads = 3
	}

	if cfg.Settings.Timeout == 0 {
		cfg.Settings.Timeout = 5 * time.Minute
	}

	if cfg.Metadata.TrackingFile == "" {
		cfg.Metadata.TrackingFile = ".claude/.installed-agents.json"
	}

	if cfg.Metadata.LogFile == "" {
		cfg.Metadata.LogFile = ".claude/installation.log"
	}

	if cfg.Metadata.LockFile == "" {
		cfg.Metadata.LockFile = ".claude/.lock"
	}

	// Apply defaults to sources
	for i := range cfg.Sources {
		if cfg.Sources[i].Branch == "" && cfg.Sources[i].Type == "github" {
			cfg.Sources[i].Branch = "main"
		}
	}
}

// Save writes the configuration to a file
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
