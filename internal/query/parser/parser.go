package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FlexibleTools is a custom type that can unmarshal both string and array formats
type FlexibleTools []string

// UnmarshalYAML implements custom YAML unmarshaling for tools field
func (ft *FlexibleTools) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as array first
	var toolsArray []string
	if err := node.Decode(&toolsArray); err == nil {
		*ft = FlexibleTools(toolsArray)
		return nil
	}

	// If array unmarshaling fails, try as string
	var toolsString string
	if err := node.Decode(&toolsString); err != nil {
		return fmt.Errorf("tools field must be either a string or array of strings: %w", err)
	}

	// Parse comma-separated string into array
	if strings.TrimSpace(toolsString) == "" {
		*ft = FlexibleTools{}
		return nil
	}

	tools := strings.Split(toolsString, ",")
	for i, tool := range tools {
		tools[i] = strings.TrimSpace(tool)
	}

	*ft = FlexibleTools(tools)
	return nil
}

// GetTools returns tools as []string for compatibility
func (ft FlexibleTools) GetTools() []string {
	return []string(ft)
}

// AgentSpec represents a Claude Code subagent
type AgentSpec struct {
	// YAML frontmatter fields
	Name        string        `yaml:"name" json:"name"`
	Description string        `yaml:"description" json:"description"`
	Tools       FlexibleTools `yaml:"tools,omitempty" json:"tools,omitempty"`

	// Derived fields
	ToolsInherited bool   `json:"tools_inherited"`
	Prompt         string `json:"prompt"`

	// File metadata
	FilePath string    `json:"file_path"`
	FileName string    `json:"file_name"`
	FileSize int64     `json:"file_size"`
	ModTime  time.Time `json:"mod_time"`

	// Installation metadata
	Source      string    `json:"source,omitempty"`
	InstalledAt time.Time `json:"installed_at,omitempty"`
}

// GetToolsAsSlice returns tools as []string for compatibility with existing code
func (a *AgentSpec) GetToolsAsSlice() []string {
	return a.Tools.GetTools()
}

// Parser extracts agent specifications
type Parser struct {
	SuppressWarnings bool
}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// NewParserWithOptions creates a new parser with options
func NewParserWithOptions(suppressWarnings bool) *Parser {
	return &Parser{
		SuppressWarnings: suppressWarnings,
	}
}

// ParseFile extracts agent spec from a file
func (p *Parser) ParseFile(path string) (*AgentSpec, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split frontmatter and content
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid agent format: missing frontmatter")
	}

	// Parse YAML frontmatter
	var spec AgentSpec
	if err := yaml.Unmarshal([]byte(parts[1]), &spec); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set prompt content
	spec.Prompt = strings.TrimSpace(parts[2])

	// Handle tools field - if empty or nil, mark as inherited
	if len(spec.Tools.GetTools()) == 0 {
		spec.ToolsInherited = true
	}

	// Add file metadata
	spec.FilePath = path
	spec.FileName = filepath.Base(path)

	if info, err := os.Stat(path); err == nil {
		spec.FileSize = info.Size()
		spec.ModTime = info.ModTime()
	}

	return &spec, nil
}

// ParseDirectory parses all agents in a directory
func (p *Parser) ParseDirectory(dir string) ([]*AgentSpec, error) {
	var agents []*AgentSpec

	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, walkFuncErr error) error {
		if walkFuncErr != nil {
			// Log error but continue processing other files
			if !p.SuppressWarnings {
				fmt.Fprintf(os.Stderr, "Warning: error accessing %s: %v\n", path, walkFuncErr)
			}
			return nil
		}

		if strings.HasSuffix(path, ".md") {
			agent, parseErr := p.ParseFile(path)
			if parseErr != nil {
				// Log error but continue parsing other files
				if !p.SuppressWarnings {
					fmt.Fprintf(os.Stderr, "Warning: error parsing %s: %v\n", path, parseErr)
				}
				return nil
			}
			agents = append(agents, agent)
		}

		return nil
	})

	return agents, walkErr
}
