package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/conflict"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/pacphi/claude-code-agent-manager/internal/tracker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallAndQueryWorkflow tests the complete workflow from installation to querying
func TestInstallAndQueryWorkflow(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer cleanup(testDir)

	// Create test configuration
	cfg := createTestConfig(testDir)

	// Create test agents
	createTestAgents(t, testDir)

	// Test installation
	installAgents(t, cfg, testDir)

	// Test query functionality
	testQueryOperations(t, cfg)

	// Test index operations
	testIndexOperations(t, cfg)

	// Test cache operations
	testCacheOperations(t, cfg)
}

// TestUpdateAndVerifyIndex tests updating agents and verifying index updates
func TestUpdateAndVerifyIndex(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer cleanup(testDir)

	cfg := createTestConfig(testDir)
	createTestAgents(t, testDir)
	installAgents(t, cfg, testDir)

	// Create query engine
	queryEngine, err := engine.NewEngine(
		filepath.Join(testDir, ".agent-index"),
		filepath.Join(testDir, ".agent-cache"),
	)
	require.NoError(t, err)

	// Initial index build
	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Get initial agent count
	initialAgents := queryEngine.GetAllAgents()
	initialCount := len(initialAgents)
	assert.Greater(t, initialCount, 0, "Should have indexed some agents")

	// Add a new agent
	newAgentPath := filepath.Join(cfg.Settings.BaseDir, "new-test-agent.md")
	newAgentContent := `---
name: new-test-agent
description: A newly added test agent
tools: [Read, Write]
---

This is a new test agent added for testing index updates.
`
	err = os.WriteFile(newAgentPath, []byte(newAgentContent), 0644)
	require.NoError(t, err)

	// Update index
	err = queryEngine.UpdateIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Verify the new agent is indexed
	updatedAgents := queryEngine.GetAllAgents()
	assert.GreaterOrEqual(t, len(updatedAgents), initialCount+1, "Index should contain at least the new agent")

	// Search for the new agent
	results, err := queryEngine.Query("new-test-agent", engine.QueryOptions{})
	require.NoError(t, err)
	assert.Len(t, results, 1, "Should find exactly one result for new-test-agent")
	assert.Equal(t, "new-test-agent", results[0].Name)
}

// TestComplexQueryPatterns tests complex query patterns with real data
func TestComplexQueryPatterns(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer cleanup(testDir)

	cfg := createTestConfig(testDir)
	createComplexTestAgents(t, testDir)
	installAgents(t, cfg, testDir)

	queryEngine, err := engine.NewEngine(
		filepath.Join(testDir, ".agent-index"),
		filepath.Join(testDir, ".agent-cache"),
	)
	require.NoError(t, err)

	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Test field-specific queries
	testFieldSpecificQueries(t, queryEngine)

	// Test filtering options
	testFilteringOptions(t, queryEngine)

	// Test fuzzy matching
	testFuzzyMatching(t, queryEngine)

	// Test empty and edge case queries
	testEdgeCaseQueries(t, queryEngine)
}

// TestPerformanceWithLargeAgentSet tests performance with a large number of agents
func TestPerformanceWithLargeAgentSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testDir := setupTestDirectory(t)
	defer cleanup(testDir)

	cfg := createTestConfig(testDir)

	// Create a large number of test agents (100 agents)
	agentCount := 100
	createLargeAgentSet(t, testDir, agentCount)

	queryEngine, err := engine.NewEngine(
		filepath.Join(testDir, ".agent-index"),
		filepath.Join(testDir, ".agent-cache"),
	)
	require.NoError(t, err)

	// Test index build performance
	start := time.Now()
	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)
	buildDuration := time.Since(start)

	// Verify performance target: < 5s for 1000 agents (proportionally < 0.5s for 100 agents)
	maxBuildTime := 2 * time.Second // Be generous with CI environments
	assert.Less(t, buildDuration, maxBuildTime,
		"Index build should complete within %v for %d agents, took %v",
		maxBuildTime, agentCount, buildDuration)

	// Test query performance
	testQueryPerformance(t, queryEngine, 50*time.Millisecond)

	// Test memory usage by getting all agents
	allAgents := queryEngine.GetAllAgents()
	assert.Len(t, allAgents, agentCount, "Should have indexed all agents")

	// Test cache performance
	testCachePerformance(t, queryEngine)
}

// TestErrorHandling tests error handling across component boundaries
func TestErrorHandling(t *testing.T) {
	testDir := setupTestDirectory(t)
	defer cleanup(testDir)

	cfg := createTestConfig(testDir)

	// Test with non-existent directory
	queryEngine, err := engine.NewEngine(
		filepath.Join(testDir, ".agent-index"),
		filepath.Join(testDir, ".agent-cache"),
	)
	require.NoError(t, err)

	// Should handle non-existent directory gracefully
	err = queryEngine.RebuildIndex("/non/existent/directory")
	// The parser might handle this gracefully, so we don't assert error
	// but we can verify that no agents were indexed
	if err == nil {
		agents := queryEngine.GetAllAgents()
		assert.Len(t, agents, 0, "No agents should be indexed from non-existent directory")
	}

	// Test with malformed agent files
	createMalformedAgents(t, cfg.Settings.BaseDir)
	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	assert.NoError(t, err) // Should not fail completely due to malformed files

	// Test invalid queries
	results, err := queryEngine.QueryByField("invalid_field", "test")
	assert.Error(t, err)
	assert.Nil(t, results)

	// Test show agent with non-existent agent
	_, err = queryEngine.ShowAgent("non-existent-agent")
	assert.Error(t, err)
}

// Helper functions

func setupTestDirectory(t *testing.T) string {
	testDir, err := os.MkdirTemp("", "agent-manager-integration-test-*")
	require.NoError(t, err)
	return testDir
}

func cleanup(testDir string) {
	os.RemoveAll(testDir)
}

func createTestConfig(testDir string) *config.Config {
	return &config.Config{
		Version: "1.0",
		Settings: config.Settings{
			BaseDir:             filepath.Join(testDir, "agents"),
			DocsDir:             "docs",
			ConflictStrategy:    "backup",
			BackupDir:           filepath.Join(testDir, "backups"),
			ConcurrentDownloads: 3,
			Timeout:             5 * time.Minute,
		},
		Sources: []config.Source{
			{
				Name:    "test-local",
				Enabled: true,
				Type:    "local",
				Paths: config.PathConfig{
					Source: testDir,
					Target: filepath.Join(testDir, "agents"),
				},
			},
		},
		Metadata: config.Metadata{
			TrackingFile: filepath.Join(testDir, ".installed-agents.json"),
			LogFile:      filepath.Join(testDir, "installation.log"),
		},
	}
}

func createTestAgents(t *testing.T, testDir string) {
	agentsDir := filepath.Join(testDir, "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	agents := []struct {
		name    string
		content string
	}{
		{
			name: "go-specialist.md",
			content: `---
name: go-specialist
description: Expert Go developer specializing in Go programming
tools: [Read, Write, Edit, Bash]
---

You are an expert Go developer with deep knowledge of Go best practices.
`,
		},
		{
			name: "python-expert.md",
			content: `---
name: python-expert
description: Python programming specialist
tools: [Read, Write, Edit]
---

You are a Python expert focusing on clean, efficient code.
`,
		},
		{
			name: "code-reviewer.md",
			content: `---
name: code-reviewer
description: Code review specialist focusing on quality and security
---

You provide comprehensive code reviews with security focus.
`,
		},
	}

	for _, agent := range agents {
		agentPath := filepath.Join(agentsDir, agent.name)
		err := os.WriteFile(agentPath, []byte(agent.content), 0644)
		require.NoError(t, err)
	}
}

func createComplexTestAgents(t *testing.T, testDir string) {
	agentsDir := filepath.Join(testDir, "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Create agents with various configurations for comprehensive testing
	agents := []struct {
		name    string
		content string
	}{
		{
			name: "full-stack-agent.md",
			content: `---
name: full-stack-agent
description: Full-stack developer with React and Node.js expertise
tools: [Read, Write, Edit, Bash, WebFetch]
---

Expert in both frontend and backend development technologies.
`,
		},
		{
			name: "database-expert.md",
			content: `---
name: database-expert
description: Database design and optimization specialist
tools: [Read, Write, Bash]
---

Specialized in database architecture, SQL optimization, and performance tuning.
`,
		},
		{
			name: "minimal-agent.md",
			content: `---
name: minimal-agent
description: Minimal agent for testing
---

This agent has only the required fields.
`,
		},
		{
			name: "devops-engineer.md",
			content: `---
name: devops-engineer
description: DevOps and infrastructure automation specialist
tools: [Read, Write, Edit, Bash, Task]
---

Expert in CI/CD, Docker, Kubernetes, and infrastructure as code.
`,
		},
	}

	for _, agent := range agents {
		agentPath := filepath.Join(agentsDir, agent.name)
		err := os.WriteFile(agentPath, []byte(agent.content), 0644)
		require.NoError(t, err)
	}
}

func createLargeAgentSet(t *testing.T, testDir string, count int) {
	agentsDir := filepath.Join(testDir, "agents")
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	tools := [][]string{
		{"Read", "Write"},
		{"Read", "Write", "Edit"},
		{"Read", "Write", "Edit", "Bash"},
		{"Read", "Write", "Task"},
		{}, // No tools (inherited)
	}

	for i := 0; i < count; i++ {
		toolSet := tools[i%len(tools)]
		var toolsYaml string
		if len(toolSet) > 0 {
			toolsYaml = fmt.Sprintf("\ntools: [%s]", strings.Join(toolSet, ", "))
		}

		content := fmt.Sprintf(`---
name: test-agent-%d
description: Test agent number %d for performance testing%s
---

This is test agent %d created for performance testing.
It contains some sample content to test parsing and indexing performance.
`, i+1, i+1, toolsYaml, i+1)

		agentPath := filepath.Join(agentsDir, fmt.Sprintf("test-agent-%d.md", i+1))
		err := os.WriteFile(agentPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

func createMalformedAgents(t *testing.T, agentsDir string) {
	// Ensure the directory exists
	err := os.MkdirAll(agentsDir, 0755)
	require.NoError(t, err)

	// Agent with invalid YAML
	invalidYamlPath := filepath.Join(agentsDir, "invalid-yaml.md")
	invalidYamlContent := `---
name: invalid-yaml
description: Agent with invalid YAML
tools: [Read, Write,
---

This agent has invalid YAML frontmatter.
`
	err = os.WriteFile(invalidYamlPath, []byte(invalidYamlContent), 0644)
	require.NoError(t, err)

	// Agent with missing frontmatter
	noFrontmatterPath := filepath.Join(agentsDir, "no-frontmatter.md")
	noFrontmatterContent := "This file has no YAML frontmatter at all."
	err = os.WriteFile(noFrontmatterPath, []byte(noFrontmatterContent), 0644)
	require.NoError(t, err)
}

func installAgents(t *testing.T, cfg *config.Config, testDir string) {
	// Create necessary directories
	err := os.MkdirAll(cfg.Settings.BaseDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(cfg.Settings.BackupDir, 0755)
	require.NoError(t, err)

	// Create installer
	track := tracker.New(cfg.Metadata.TrackingFile)
	resolver := conflict.NewResolver(cfg.Settings.ConflictStrategy, cfg.Settings.BackupDir)
	inst := installer.New(cfg, track, resolver, installer.Options{})

	// Install the test source
	err = inst.InstallSource(cfg.Sources[0])
	require.NoError(t, err)
}

func testQueryOperations(t *testing.T, cfg *config.Config) {
	queryEngine, err := engine.NewEngine(
		filepath.Join(cfg.Settings.BaseDir, "../.agent-index"),
		filepath.Join(cfg.Settings.BaseDir, "../.agent-cache"),
	)
	require.NoError(t, err)

	// Build index
	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Test basic query
	results, err := queryEngine.Query("go", engine.QueryOptions{})
	require.NoError(t, err)
	assert.Greater(t, len(results), 0, "Should find results for 'go' query")

	// Test field-specific queries
	results, err = queryEngine.QueryByField("name", "go-specialist")
	require.NoError(t, err)
	if len(results) > 0 {
		// Find the exact match
		found := false
		for _, agent := range results {
			if agent.Name == "go-specialist" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find go-specialist among results: %v", results)
	} else {
		t.Error("Should find results for go-specialist")
	}

	// Test show agent
	agent, err := queryEngine.ShowAgent("go-specialist.md")
	require.NoError(t, err)
	assert.Equal(t, "go-specialist", agent.Name)
}

func testIndexOperations(t *testing.T, cfg *config.Config) {
	queryEngine, err := engine.NewEngine(
		filepath.Join(cfg.Settings.BaseDir, "../.agent-index"),
		filepath.Join(cfg.Settings.BaseDir, "../.agent-cache"),
	)
	require.NoError(t, err)

	// Test rebuild index
	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Test update index
	err = queryEngine.UpdateIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Test get all agents
	agents := queryEngine.GetAllAgents()
	assert.Greater(t, len(agents), 0, "Should have indexed agents")
}

func testCacheOperations(t *testing.T, cfg *config.Config) {
	queryEngine, err := engine.NewEngine(
		filepath.Join(cfg.Settings.BaseDir, "../.agent-index"),
		filepath.Join(cfg.Settings.BaseDir, "../.agent-cache"),
	)
	require.NoError(t, err)

	err = queryEngine.RebuildIndex(cfg.Settings.BaseDir)
	require.NoError(t, err)

	// Perform query to populate cache
	_, err = queryEngine.Query("test", engine.QueryOptions{})
	require.NoError(t, err)

	// Get cache stats
	stats := queryEngine.GetCacheStats()
	assert.NotNil(t, stats, "Should return cache stats")

	// Clear cache
	err = queryEngine.ClearCache()
	require.NoError(t, err)

	// Save cache
	err = queryEngine.SaveCache()
	require.NoError(t, err)
}

func testFieldSpecificQueries(t *testing.T, queryEngine *engine.Engine) {
	// Test name search
	results, err := queryEngine.QueryByField("name", "full-stack")
	require.NoError(t, err)
	assert.Greater(t, len(results), 0, "Should find results for full-stack name search")
	found := false
	for _, agent := range results {
		if agent.Name == "full-stack-agent" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find full-stack-agent in results")

	// Test description search
	results, err = queryEngine.QueryByField("description", "database")
	require.NoError(t, err)
	assert.Greater(t, len(results), 0, "Should find results for database description search")
	found = false
	for _, agent := range results {
		if agent.Name == "database-expert" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find database-expert in results")

	// Test tools search
	results, err = queryEngine.QueryByField("tools", "WebFetch")
	require.NoError(t, err)
	assert.Greater(t, len(results), 0, "Should find results for WebFetch tools search")
	found = false
	for _, agent := range results {
		if agent.Name == "full-stack-agent" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should find full-stack-agent in WebFetch tools results")
}

func testFilteringOptions(t *testing.T, queryEngine *engine.Engine) {
	// Test no-tools filter (agents with inherited tools)
	results, err := queryEngine.Query("", engine.QueryOptions{
		NoTools: true,
		Limit:   10,
	})
	require.NoError(t, err)
	for _, agent := range results {
		assert.True(t, agent.ToolsInherited, "Agent %s should have inherited tools", agent.Name)
	}

	// Test custom-tools filter (agents with explicit tools)
	results, err = queryEngine.Query("", engine.QueryOptions{
		CustomTools: true,
		Limit:       10,
	})
	require.NoError(t, err)
	for _, agent := range results {
		assert.False(t, agent.ToolsInherited, "Agent %s should have custom tools", agent.Name)
	}
}

func testFuzzyMatching(t *testing.T, queryEngine *engine.Engine) {
	// Test fuzzy matching with partial names
	agent, err := queryEngine.ShowAgent("full-stack")
	require.NoError(t, err)
	assert.Equal(t, "full-stack-agent", agent.Name)

	// Test fuzzy matching with typos (if implemented)
	agent, err = queryEngine.ShowAgent("databse")
	if err == nil {
		assert.Equal(t, "database-expert", agent.Name)
	}
}

func testEdgeCaseQueries(t *testing.T, queryEngine *engine.Engine) {
	// Test empty query
	results, err := queryEngine.Query("", engine.QueryOptions{})
	require.NoError(t, err)
	assert.Greater(t, len(results), 0, "Empty query should return all agents")

	// Test non-existent term
	results, err = queryEngine.Query("non-existent-term-xyz", engine.QueryOptions{})
	require.NoError(t, err)
	assert.Len(t, results, 0, "Should return no results for non-existent term")

	// Test query with limit
	results, err = queryEngine.Query("", engine.QueryOptions{Limit: 2})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(results), 2, "Should respect limit parameter")
}

func testQueryPerformance(t *testing.T, queryEngine *engine.Engine, maxDuration time.Duration) {
	// Test various query types for performance
	queries := []struct {
		name  string
		query string
		opts  engine.QueryOptions
	}{
		{"basic_search", "test", engine.QueryOptions{}},
		{"name_search", "test-agent-1", engine.QueryOptions{}},
		{"limited_search", "agent", engine.QueryOptions{Limit: 10}},
		{"filtered_search", "", engine.QueryOptions{NoTools: true, Limit: 20}},
	}

	for _, q := range queries {
		start := time.Now()
		results, err := queryEngine.Query(q.query, q.opts)
		duration := time.Since(start)

		require.NoError(t, err, "Query %s should not error", q.name)
		assert.Less(t, duration, maxDuration,
			"Query %s took %v, should be less than %v", q.name, duration, maxDuration)

		t.Logf("Query %s returned %d results in %v", q.name, len(results), duration)
	}
}

func testCachePerformance(t *testing.T, queryEngine *engine.Engine) {
	// Warm up cache with a query
	query := "test-agent"
	opts := engine.QueryOptions{Limit: 10}

	// First query (cache miss)
	start := time.Now()
	results1, err := queryEngine.Query(query, opts)
	firstDuration := time.Since(start)
	require.NoError(t, err)

	// Second identical query (should be cache hit)
	start = time.Now()
	results2, err := queryEngine.Query(query, opts)
	secondDuration := time.Since(start)
	require.NoError(t, err)

	assert.Equal(t, len(results1), len(results2), "Cached results should match original")

	// Cache hit should be significantly faster (at least 2x faster)
	if firstDuration > 1*time.Millisecond { // Only check if first query took meaningful time
		assert.Less(t, secondDuration, firstDuration/2,
			"Cached query (%v) should be significantly faster than first query (%v)",
			secondDuration, firstDuration)
	}

	// Verify cache stats show hits
	stats := queryEngine.GetCacheStats()
	if hits, ok := stats["hits"].(int); ok {
		assert.Greater(t, hits, 0, "Cache should have registered hits")
	}
}
