package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/cache"
	"github.com/pacphi/claude-code-agent-manager/internal/query/fuzzy"
	"github.com/pacphi/claude-code-agent-manager/internal/query/index"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

// Engine handles agent queries with caching and advanced search capabilities
type Engine struct {
	index  *index.IndexManager
	cache  *cache.CacheManager
	parser *parser.Parser
	fuzzy  *fuzzy.FuzzyMatcher
}

// NewEngine creates a new query engine with the specified index and cache paths
func NewEngine(indexPath, cachePath string) (*Engine, error) {
	indexManager, err := index.NewIndexManager(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create index manager: %w", err)
	}

	cacheManager, err := cache.NewCacheManager(cachePath, cache.Config{
		MaxSize: 100,
		TTL:     time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	return &Engine{
		index:  indexManager,
		cache:  cacheManager,
		parser: parser.NewParserWithOptions(true), // Suppress warnings by default
		fuzzy:  fuzzy.NewFuzzyMatcher(0.7),
	}, nil
}

// QueryOptions provides filtering and configuration options for queries
type QueryOptions struct {
	Limit       int             // Maximum number of results to return
	NoTools     bool            // Find agents with inherited tools only
	CustomTools bool            // Find agents with explicit tools only
	Regex       bool            // Use regex pattern matching
	Source      string          // Filter by installation source
	After       time.Time       // Filter agents installed after this time
	Context     context.Context // For cancellation and timeouts
}

// QueryWithFuzzy searches for agents using enhanced multi-field fuzzy matching
func (e *Engine) QueryWithFuzzy(query string, opts QueryOptions) ([]*parser.AgentSpec, error) {
	// Set default context if not provided
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// Check for cancellation
	select {
	case <-opts.Context.Done():
		return nil, opts.Context.Err()
	default:
	}

	// Check cache first
	cacheKey := e.buildCacheKey("fuzzy:"+query, opts)
	if cached := e.cache.Get(cacheKey); cached != nil {
		if agents, ok := cached.([]*parser.AgentSpec); ok {
			return agents, nil
		}
	}

	// Use fuzzy multi-field search for enhanced matching
	allAgents := e.index.GetAll()
	results := e.fuzzy.MultiFieldSearch(query, allAgents, nil, opts.Limit)

	// Apply additional filters
	results = e.applyQueryFilters(results, opts)

	// Cache results
	e.cache.Set(cacheKey, results)

	return results, nil
}

// Query searches for agents using the provided query string and options
func (e *Engine) Query(query string, opts QueryOptions) ([]*parser.AgentSpec, error) {
	// Set default context if not provided
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// Check for cancellation
	select {
	case <-opts.Context.Done():
		return nil, opts.Context.Err()
	default:
	}

	// Check cache first
	cacheKey := e.buildCacheKey(query, opts)
	if cached := e.cache.Get(cacheKey); cached != nil {
		if agents, ok := cached.([]*parser.AgentSpec); ok {
			return agents, nil
		}
	}

	// Execute search - maintain original behavior unless explicitly using regex
	results, err := e.index.Search(query, index.QueryOptions{
		Limit:       opts.Limit,
		NoTools:     opts.NoTools,
		CustomTools: opts.CustomTools,
		Source:      opts.Source,
		After:       opts.After,
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Cache results
	e.cache.Set(cacheKey, results)

	return results, nil
}

// QueryByField searches specific fields with the provided value
func (e *Engine) QueryByField(field, value string) ([]*parser.AgentSpec, error) {
	field = strings.ToLower(strings.TrimSpace(field))
	value = strings.TrimSpace(value)

	switch field {
	case "name":
		return e.index.SearchByName(value)
	case "description":
		return e.index.SearchByDescription(value)
	case "content", "prompt":
		return e.index.SearchByContent(value)
	case "tools":
		tools := strings.Split(value, ",")
		for i := range tools {
			tools[i] = strings.TrimSpace(tools[i])
		}
		return e.index.SearchByTools(tools)
	case "source":
		return e.index.SearchBySource(value)
	default:
		return nil, fmt.Errorf("invalid field: %s", field)
	}
}

// ShowAgent retrieves an agent by filename with fuzzy matching fallback
func (e *Engine) ShowAgent(filename string) (*parser.AgentSpec, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	// Try exact match first
	if agent := e.index.GetByFilename(filename); agent != nil {
		return agent, nil
	}

	// Try with .md extension if not present
	if !strings.HasSuffix(filename, ".md") {
		if agent := e.index.GetByFilename(filename + ".md"); agent != nil {
			return agent, nil
		}
	}

	// Fallback to fuzzy matching
	agents := e.index.GetAll()
	if match := e.fuzzy.FindBest(filename, agents); match != nil {
		return match, nil
	}

	return nil, fmt.Errorf("agent not found: %s", filename)
}

// RebuildIndex rebuilds the search index from the specified directory
func (e *Engine) RebuildIndex(dir string) error {
	// Clear cache when rebuilding index
	e.cache.Clear()

	if err := e.index.Rebuild(dir); err != nil {
		return err
	}

	// Save the rebuilt index to disk
	return e.index.Save()
}

// RebuildWithAgents rebuilds the index with a provided list of agents
func (e *Engine) RebuildWithAgents(agents []*parser.AgentSpec) error {
	// Clear cache when rebuilding index
	e.cache.Clear()

	return e.index.RebuildWithAgents(agents)
}

// UpdateIndex updates the index with new or modified agents
func (e *Engine) UpdateIndex(dir string) error {
	// Parse agents from directory
	agents, err := e.parser.ParseDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to parse agents: %w", err)
	}

	// Rebuild index with all agents
	if err := e.index.RebuildWithAgents(agents); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	// Save the index to disk
	if err := e.index.Save(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	// Clear cache to ensure fresh results
	e.cache.Clear()

	return nil
}

// GetAllAgents returns all agents in the index
func (e *Engine) GetAllAgents() []*parser.AgentSpec {
	return e.index.GetAll()
}

// GetStats returns statistics about the indexed agents
func (e *Engine) GetStats() map[string]interface{} {
	agents := e.index.GetAll()

	stats := map[string]interface{}{
		"total_agents": len(agents),
		"cache_stats":  e.cache.Stats(),
		"index_stats":  e.index.Stats(),
	}

	// Count by source
	sources := make(map[string]int)
	toolsInherited := 0
	customTools := 0

	for _, agent := range agents {
		if agent.Source != "" {
			sources[agent.Source]++
		}

		if agent.ToolsInherited {
			toolsInherited++
		} else {
			customTools++
		}
	}

	stats["by_source"] = sources
	stats["tools_inherited"] = toolsInherited
	stats["custom_tools"] = customTools

	return stats
}

// ClearCache clears the query cache
func (e *Engine) ClearCache() error {
	e.cache.Clear()
	return nil
}

// SaveCache saves the cache to disk
func (e *Engine) SaveCache() error {
	return e.cache.Save()
}

// GetCacheStats returns cache performance statistics
func (e *Engine) GetCacheStats() map[string]interface{} {
	return e.cache.Stats()
}

// ShowCacheStats displays cache statistics
func (e *Engine) ShowCacheStats() error {
	stats := e.cache.Stats()
	fmt.Printf("Cache Statistics:\n")
	fmt.Printf("  Size: %v entries\n", stats["size"])
	fmt.Printf("  Hits: %v\n", stats["hits"])
	fmt.Printf("  Misses: %v\n", stats["misses"])

	if hits, ok := stats["hits"].(int); ok {
		if misses, ok := stats["misses"].(int); ok {
			total := hits + misses
			if total > 0 {
				hitRate := float64(hits) / float64(total) * 100
				fmt.Printf("  Hit Rate: %.1f%%\n", hitRate)
			}
		}
	}

	return nil
}

// SetFuzzyThreshold adjusts the fuzzy matching threshold
func (e *Engine) SetFuzzyThreshold(threshold float64) {
	e.fuzzy.SetThreshold(threshold)
}

// applyQueryFilters applies additional filters to query results
func (e *Engine) applyQueryFilters(agents []*parser.AgentSpec, opts QueryOptions) []*parser.AgentSpec {
	// Pre-allocate slice with estimated capacity to avoid reallocations
	filtered := make([]*parser.AgentSpec, 0, len(agents))

	for _, agent := range agents {
		// Apply source filter
		if opts.Source != "" && agent.Source != opts.Source {
			continue
		}

		// Apply tools filters
		if opts.NoTools && !agent.ToolsInherited {
			continue
		}
		if opts.CustomTools && agent.ToolsInherited {
			continue
		}

		// Apply date filter
		if !opts.After.IsZero() && agent.InstalledAt.Before(opts.After) {
			continue
		}

		filtered = append(filtered, agent)
	}

	// Apply limit if not already handled by search
	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered
}

// buildCacheKey creates a cache key from query and options
func (e *Engine) buildCacheKey(query string, opts QueryOptions) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("q:%s", query))

	if opts.Limit > 0 {
		parts = append(parts, fmt.Sprintf("l:%d", opts.Limit))
	}

	if opts.NoTools {
		parts = append(parts, "nt:true")
	}

	if opts.CustomTools {
		parts = append(parts, "ct:true")
	}

	if opts.Regex {
		parts = append(parts, "r:true")
	}

	if opts.Source != "" {
		parts = append(parts, fmt.Sprintf("s:%s", opts.Source))
	}

	if !opts.After.IsZero() {
		parts = append(parts, fmt.Sprintf("a:%d", opts.After.Unix()))
	}

	return strings.Join(parts, "|")
}
