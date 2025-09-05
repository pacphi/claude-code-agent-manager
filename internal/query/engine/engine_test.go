package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.index)
	assert.NotNil(t, engine.cache)
	assert.NotNil(t, engine.parser)
	assert.NotNil(t, engine.fuzzy)
}

func TestEngine_Query(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	// Add test agents
	agents := []*parser.AgentSpec{
		{
			Name:           "data-processor",
			Description:    "Process data files efficiently",
			FileName:       "data-processor.md",
			Prompt:         "You are a data processing expert",
			ToolsInherited: true, // This agent uses inherited tools
		},
		{
			Name:           "code-reviewer",
			Description:    "Review code for quality and style",
			FileName:       "code-reviewer.md",
			Prompt:         "You are a code review specialist",
			Tools:          []string{"Read", "Write"},
			ToolsInherited: false, // This agent has explicit tools
		},
		{
			Name:           "web-scraper",
			Description:    "Scrape websites for data",
			FileName:       "web-scraper.md",
			Prompt:         "You are a web scraping expert",
			Tools:          []string{"WebFetch", "WebSearch"},
			ToolsInherited: false, // This agent has explicit tools
		},
	}

	for _, agent := range agents {
		engine.index.AddAgent(agent)
	}

	tests := []struct {
		name     string
		query    string
		opts     QueryOptions
		minCount int
	}{
		{
			name:     "simple query",
			query:    "data",
			opts:     QueryOptions{},
			minCount: 1,
		},
		{
			name:     "description search",
			query:    "review",
			opts:     QueryOptions{},
			minCount: 1,
		},
		{
			name:     "empty query returns all",
			query:    "",
			opts:     QueryOptions{},
			minCount: 3,
		},
		{
			name:     "limited results",
			query:    "",
			opts:     QueryOptions{Limit: 2},
			minCount: 2,
		},
		{
			name:     "custom tools filter",
			query:    "",
			opts:     QueryOptions{CustomTools: true},
			minCount: 2, // code-reviewer and web-scraper have explicit tools
		},
		{
			name:     "inherited tools filter",
			query:    "",
			opts:     QueryOptions{NoTools: true},
			minCount: 1, // data-processor has inherited tools
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.Query(tt.query, tt.opts)
			require.NoError(t, err)
			if tt.opts.Limit > 0 {
				assert.LessOrEqual(t, len(results), tt.opts.Limit)
			}
			assert.GreaterOrEqual(t, len(results), tt.minCount)
		})
	}
}

func TestEngine_QueryByField(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	// Add test agents
	agents := []*parser.AgentSpec{
		{
			Name:        "data-processor",
			Description: "Process data files efficiently",
			FileName:    "data-processor.md",
			Prompt:      "You are a data processing expert",
		},
		{
			Name:        "code-reviewer",
			Description: "Review code for quality and style",
			FileName:    "code-reviewer.md",
			Prompt:      "You are a code review specialist",
			Tools:       []string{"Read", "Write"},
		},
	}

	for _, agent := range agents {
		engine.index.AddAgent(agent)
	}

	tests := []struct {
		name      string
		field     string
		value     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "search by name",
			field:     "name",
			value:     "data",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "search by description",
			field:     "description",
			value:     "review",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "search by content",
			field:     "content",
			value:     "processing",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "search by prompt",
			field:     "prompt",
			value:     "specialist",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "search by tools",
			field:     "tools",
			value:     "Read,Write",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid field",
			field:     "invalid",
			value:     "test",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.QueryByField(tt.field, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tt.wantCount)
			}
		})
	}
}

func TestEngine_ShowAgent(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	// Lower fuzzy matching threshold for testing
	engine.SetFuzzyThreshold(0.3)

	// Add test agents
	agent := &parser.AgentSpec{
		Name:        "test-agent",
		Description: "Test agent",
		FileName:    "test-agent.md",
		Prompt:      "You are a test assistant",
	}
	engine.index.AddAgent(agent)

	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "exact match",
			filename: "test-agent.md",
			wantErr:  false,
		},
		{
			name:     "without extension",
			filename: "test-agent",
			wantErr:  false,
		},
		{
			name:     "fuzzy match",
			filename: "test",
			wantErr:  false,
		},
		{
			name:     "not found",
			filename: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ShowAgent(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, agent.Name, result.Name)
			}
		})
	}
}

func TestEngine_WithCache(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	// Add test agent
	agent := &parser.AgentSpec{
		Name:        "cached-agent",
		Description: "Test caching",
		FileName:    "cached-agent.md",
		Prompt:      "You are cached",
	}
	engine.index.AddAgent(agent)

	// First query should hit the index
	results1, err := engine.Query("cached", QueryOptions{})
	require.NoError(t, err)
	assert.Len(t, results1, 1)

	// Second query should hit the cache (we can't directly test this, but it should work)
	results2, err := engine.Query("cached", QueryOptions{})
	require.NoError(t, err)
	assert.Len(t, results2, 1)
	assert.Equal(t, results1[0].Name, results2[0].Name)
}

func TestEngine_QueryWithTimeFilter(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Add test agents with different install times
	agents := []*parser.AgentSpec{
		{
			Name:        "old-agent",
			Description: "Old agent",
			FileName:    "old-agent.md",
			InstalledAt: yesterday,
		},
		{
			Name:        "new-agent",
			Description: "New agent",
			FileName:    "new-agent.md",
			InstalledAt: now,
		},
	}

	for _, agent := range agents {
		engine.index.AddAgent(agent)
	}

	// Query for agents installed after yesterday
	results, err := engine.Query("", QueryOptions{After: yesterday.Add(time.Hour)})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "new-agent", results[0].Name)

	// Query for agents installed after tomorrow (should return none)
	results, err = engine.Query("", QueryOptions{After: tomorrow})
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestEngine_Integration(t *testing.T) {
	tempDir := t.TempDir()

	// Create test agent files
	agentsDir := filepath.Join(tempDir, "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	agentContent := `---
name: integration-test
description: Integration test agent
tools: ["Read", "Write"]
---

You are an integration test assistant.`

	agentFile := filepath.Join(agentsDir, "integration-test.md")
	err = os.WriteFile(agentFile, []byte(agentContent), 0644)
	require.NoError(t, err)

	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	// Rebuild index from directory
	err = engine.RebuildIndex(agentsDir)
	require.NoError(t, err)

	// Test that agent was indexed
	results, err := engine.Query("integration", QueryOptions{})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "integration-test", results[0].Name)

	// Test show functionality
	agent, err := engine.ShowAgent("integration-test")
	require.NoError(t, err)
	assert.Equal(t, "integration-test", agent.Name)
	assert.Equal(t, "Integration test agent", agent.Description)
	assert.Contains(t, agent.Prompt, "integration test assistant")
}

func TestQueryOptions_Validation(t *testing.T) {
	tempDir := t.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(t, err)

	// Add test agents
	agents := []*parser.AgentSpec{
		{
			Name:           "inherited-tools",
			Description:    "Agent with inherited tools",
			FileName:       "inherited-tools.md",
			ToolsInherited: true,
		},
		{
			Name:           "custom-tools",
			Description:    "Agent with custom tools",
			FileName:       "custom-tools.md",
			Tools:          []string{"Read"},
			ToolsInherited: false,
		},
	}

	for _, agent := range agents {
		engine.index.AddAgent(agent)
	}

	// Test NoTools filter
	results, err := engine.Query("", QueryOptions{NoTools: true})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "inherited-tools", results[0].Name)

	// Test CustomTools filter
	results, err = engine.Query("", QueryOptions{CustomTools: true})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "custom-tools", results[0].Name)

	// Test conflicting filters (should return empty)
	results, err = engine.Query("", QueryOptions{NoTools: true, CustomTools: true})
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

// Benchmarks
func BenchmarkEngine_Query(b *testing.B) {
	tempDir := b.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(b, err)

	// Add many test agents
	for i := 0; i < 100; i++ {
		agent := &parser.AgentSpec{
			Name:        "agent-" + string(rune(i)),
			Description: "Test agent " + string(rune(i)),
			FileName:    "agent-" + string(rune(i)) + ".md",
		}
		engine.index.AddAgent(agent)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Query("agent", QueryOptions{Limit: 10})
	}
}

func BenchmarkEngine_ShowAgent(b *testing.B) {
	tempDir := b.TempDir()
	indexPath := filepath.Join(tempDir, "index.json")
	cachePath := filepath.Join(tempDir, "cache")

	engine, err := NewEngine(indexPath, cachePath)
	require.NoError(b, err)

	// Add test agent
	agent := &parser.AgentSpec{
		Name:        "bench-agent",
		Description: "Benchmark agent",
		FileName:    "bench-agent.md",
	}
	engine.index.AddAgent(agent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ShowAgent("bench")
	}
}
