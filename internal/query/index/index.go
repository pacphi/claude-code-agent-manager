package index

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

// IndexManager manages agent indices
type IndexManager struct {
	mu     sync.RWMutex
	agents []*parser.AgentSpec
	byName map[string]*parser.AgentSpec
	byFile map[string]*parser.AgentSpec
	path   string
}

// QueryOptions for searches
type QueryOptions struct {
	Limit       int
	NoTools     bool // Find agents with inherited tools
	CustomTools bool // Find agents with explicit tools
	Regex       bool
	Source      string
	After       time.Time
}

// NewIndexManager creates a new index manager
func NewIndexManager(path string) (*IndexManager, error) {
	im := &IndexManager{
		agents: make([]*parser.AgentSpec, 0),
		byName: make(map[string]*parser.AgentSpec),
		byFile: make(map[string]*parser.AgentSpec),
		path:   path,
	}

	// Load existing index if available
	if err := im.load(); err != nil {
		// Start with empty index if load fails
		fmt.Fprintf(os.Stderr, "Warning: failed to load index from %s: %v\n", path, err)
	}

	return im, nil
}

// AddAgent adds an agent to the index
func (im *IndexManager) AddAgent(agent *parser.AgentSpec) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.agents = append(im.agents, agent)
	im.byName[agent.Name] = agent
	im.byFile[agent.FileName] = agent
}

// Search performs a simple text search
func (im *IndexManager) Search(query string, opts QueryOptions) ([]*parser.AgentSpec, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var results []*parser.AgentSpec
	query = strings.ToLower(query)

	for _, agent := range im.agents {
		// Apply filters
		if opts.Source != "" && agent.Source != opts.Source {
			continue
		}

		if !opts.After.IsZero() && agent.InstalledAt.Before(opts.After) {
			continue
		}

		if opts.NoTools && !agent.ToolsInherited {
			continue
		}

		if opts.CustomTools && agent.ToolsInherited {
			continue
		}

		// Search in fields
		if query == "" || // Empty query matches all
			strings.Contains(strings.ToLower(agent.Name), query) ||
			strings.Contains(strings.ToLower(agent.Description), query) ||
			strings.Contains(strings.ToLower(agent.Prompt), query) {
			results = append(results, agent)

			if opts.Limit > 0 && len(results) >= opts.Limit {
				break
			}
		}
	}

	return results, nil
}

// SearchByName searches by agent name
func (im *IndexManager) SearchByName(name string) ([]*parser.AgentSpec, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var results []*parser.AgentSpec
	name = strings.ToLower(name)

	for _, agent := range im.agents {
		if strings.Contains(strings.ToLower(agent.Name), name) {
			results = append(results, agent)
		}
	}

	return results, nil
}

// SearchByDescription searches in descriptions
func (im *IndexManager) SearchByDescription(desc string) ([]*parser.AgentSpec, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var results []*parser.AgentSpec
	desc = strings.ToLower(desc)

	for _, agent := range im.agents {
		if strings.Contains(strings.ToLower(agent.Description), desc) {
			results = append(results, agent)
		}
	}

	return results, nil
}

// SearchByContent searches in prompt content
func (im *IndexManager) SearchByContent(content string) ([]*parser.AgentSpec, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var results []*parser.AgentSpec
	content = strings.ToLower(content)

	for _, agent := range im.agents {
		if strings.Contains(strings.ToLower(agent.Prompt), content) {
			results = append(results, agent)
		}
	}

	return results, nil
}

// SearchByTools searches by tool usage
func (im *IndexManager) SearchByTools(tools []string) ([]*parser.AgentSpec, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var results []*parser.AgentSpec

	for _, agent := range im.agents {
		if agent.ToolsInherited {
			continue
		}

		hasAll := true
		for _, tool := range tools {
			found := false
			for _, agentTool := range agent.Tools {
				if agentTool == tool {
					found = true
					break
				}
			}
			if !found {
				hasAll = false
				break
			}
		}

		if hasAll {
			results = append(results, agent)
		}
	}

	return results, nil
}

// SearchBySource searches by source
func (im *IndexManager) SearchBySource(source string) ([]*parser.AgentSpec, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var results []*parser.AgentSpec

	for _, agent := range im.agents {
		if agent.Source == source {
			results = append(results, agent)
		}
	}

	return results, nil
}

// GetByFilename retrieves agent by filename
func (im *IndexManager) GetByFilename(filename string) *parser.AgentSpec {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.byFile[filename]
}

// GetAll returns all agents
func (im *IndexManager) GetAll() []*parser.AgentSpec {
	im.mu.RLock()
	defer im.mu.RUnlock()

	result := make([]*parser.AgentSpec, len(im.agents))
	copy(result, im.agents)
	return result
}

// Rebuild rebuilds the index from a directory
func (im *IndexManager) Rebuild(dir string) error {
	p := parser.NewParser()
	agents, err := p.ParseDirectory(dir)
	if err != nil {
		return err
	}

	im.mu.Lock()
	defer im.mu.Unlock()

	im.agents = agents
	im.byName = make(map[string]*parser.AgentSpec)
	im.byFile = make(map[string]*parser.AgentSpec)

	for _, agent := range agents {
		im.byName[agent.Name] = agent
		im.byFile[agent.FileName] = agent
	}

	return nil
}

// RebuildWithAgents rebuilds the index with a provided list of agents
func (im *IndexManager) RebuildWithAgents(agents []*parser.AgentSpec) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.agents = agents
	im.byName = make(map[string]*parser.AgentSpec)
	im.byFile = make(map[string]*parser.AgentSpec)

	for _, agent := range agents {
		im.byName[agent.Name] = agent
		im.byFile[agent.FileName] = agent
	}

	return nil
}

// Save saves the index to disk
func (im *IndexManager) Save() error {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.save()
}

// load loads the index from disk (private helper)
func (im *IndexManager) load() error {
	if im.path == "" {
		return nil // No path specified
	}

	data, err := os.ReadFile(im.path)
	if err != nil {
		return err // File doesn't exist or can't be read
	}

	var agents []*parser.AgentSpec
	if err := json.Unmarshal(data, &agents); err != nil {
		return err
	}

	// Rebuild internal maps
	im.agents = agents
	im.byName = make(map[string]*parser.AgentSpec)
	im.byFile = make(map[string]*parser.AgentSpec)

	for _, agent := range agents {
		im.byName[agent.Name] = agent
		im.byFile[agent.FileName] = agent
	}

	return nil
}

// Stats returns index statistics
func (im *IndexManager) Stats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return map[string]interface{}{
		"total_agents":  len(im.agents),
		"indexed_names": len(im.byName),
		"indexed_files": len(im.byFile),
	}
}

// save saves the index to disk (private helper)
func (im *IndexManager) save() error {
	if im.path == "" {
		return nil // No path specified
	}

	data, err := json.MarshalIndent(im.agents, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(im.path, data, 0644)
}
