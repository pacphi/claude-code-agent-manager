package scripts

import (
	"embed"
	"fmt"
)

//go:embed *.js
var scriptFS embed.FS

// ScriptLoader handles loading JavaScript files
type ScriptLoader struct {
	fs embed.FS
}

// NewScriptLoader creates a new script loader
func NewScriptLoader() *ScriptLoader {
	return &ScriptLoader{
		fs: scriptFS,
	}
}

// LoadScript loads a JavaScript file by name
func (l *ScriptLoader) LoadScript(name string) (string, error) {
	content, err := l.fs.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed to load script %s: %w", name, err)
	}
	return string(content), nil
}

// LoadCategoriesScript loads the categories extraction script
func (l *ScriptLoader) LoadCategoriesScript() (string, error) {
	return l.LoadScript("extract_categories.js")
}

// LoadAgentsScript loads the agents extraction script
func (l *ScriptLoader) LoadAgentsScript() (string, error) {
	return l.LoadScript("extract_agents.js")
}

// LoadContentScript loads the content extraction script
func (l *ScriptLoader) LoadContentScript() (string, error) {
	return l.LoadScript("extract_content.js")
}

// AvailableScripts returns a list of available script names
func (l *ScriptLoader) AvailableScripts() ([]string, error) {
	entries, err := l.fs.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read script directory: %w", err)
	}

	var scripts []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name()[len(entry.Name())-3:] == ".js" {
			scripts = append(scripts, entry.Name())
		}
	}

	return scripts, nil
}
