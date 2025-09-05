package index

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
)

// createTestAgent creates a test agent with given parameters
func createTestAgent(name, description string, tools []string, prompt string) *parser.AgentSpec {
	return &parser.AgentSpec{
		Name:           name,
		Description:    description,
		Tools:          tools,
		ToolsInherited: len(tools) == 0,
		Prompt:         prompt,
		FilePath:       "/test/path/" + name + ".md",
		FileName:       name + ".md",
		FileSize:       int64(len(prompt) + 100),
		ModTime:        time.Now(),
		Source:         "test-source",
		InstalledAt:    time.Now(),
	}
}

// TestNewIndexManager tests index manager creation
func TestNewIndexManager(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	if im == nil {
		t.Fatal("NewIndexManager returned nil")
	}

	// Should start empty
	agents := im.GetAll()
	if len(agents) != 0 {
		t.Errorf("Expected empty index, got %d agents", len(agents))
	}
}

// TestAddAgent tests adding a single agent to the index
func TestAddAgent(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	agent := createTestAgent("test-agent", "Test description", []string{"Read"}, "Test prompt")
	im.AddAgent(agent)

	// Verify agent was added
	agents := im.GetAll()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	if agents[0].Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agents[0].Name)
	}

	// Verify lookup by name
	foundAgent := im.byName["test-agent"]
	if foundAgent == nil {
		t.Error("Agent not found in name lookup map")
	}

	// Verify lookup by filename
	foundAgent = im.GetByFilename("test-agent.md")
	if foundAgent == nil {
		t.Error("Agent not found in filename lookup")
	}
}

// TestSearch tests basic text search functionality
func TestSearch(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	// Add test agents
	agents := []*parser.AgentSpec{
		createTestAgent("go-expert", "Golang programming expert", []string{"Read", "Write"}, "Expert Go programming assistant"),
		createTestAgent("python-helper", "Python coding assistant", []string{"Read", "Edit"}, "Help with Python development"),
		createTestAgent("web-developer", "Web development specialist", []string{"Read", "Write", "Bash"}, "Full-stack web development"),
		createTestAgent("data-analyst", "Data analysis expert", []string{"Read"}, "Analyze data and create reports"),
	}

	for _, agent := range agents {
		im.AddAgent(agent)
	}

	testCases := []struct {
		query           string
		expectedMatches []string
		description     string
	}{
		{
			"golang",
			[]string{"go-expert"},
			"search in description",
		},
		{
			"python",
			[]string{"python-helper"},
			"search in name and description",
		},
		{
			"development",
			[]string{"python-helper", "web-developer"},
			"search matches multiple agents",
		},
		{
			"expert",
			[]string{"go-expert", "data-analyst"},
			"search in name and description across agents",
		},
		{
			"nonexistent",
			[]string{},
			"search with no matches",
		},
		{
			"",
			[]string{"go-expert", "python-helper", "web-developer", "data-analyst"},
			"empty search returns all",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			opts := QueryOptions{}
			results, err := im.Search(tc.query, opts)
			if err != nil {
				t.Errorf("Search failed: %v", err)
			}

			if len(results) != len(tc.expectedMatches) {
				t.Errorf("Expected %d results, got %d", len(tc.expectedMatches), len(results))
			}

			// Check that all expected matches are found
			foundNames := make(map[string]bool)
			for _, result := range results {
				foundNames[result.Name] = true
			}

			for _, expected := range tc.expectedMatches {
				if !foundNames[expected] {
					t.Errorf("Expected to find agent '%s' in results", expected)
				}
			}
		})
	}
}

// TestSearch_WithFilters tests search with various filter options
func TestSearch_WithFilters(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	// Add test agents with different sources and installation times
	baseTime := time.Now()
	agents := []*parser.AgentSpec{
		{
			Name:           "agent1",
			Description:    "First agent",
			Tools:          []string{"Read"},
			ToolsInherited: false,
			Prompt:         "First prompt",
			Source:         "source-a",
			InstalledAt:    baseTime.Add(-2 * time.Hour),
		},
		{
			Name:           "agent2",
			Description:    "Second agent",
			Tools:          []string{},
			ToolsInherited: true,
			Prompt:         "Second prompt",
			Source:         "source-b",
			InstalledAt:    baseTime.Add(-1 * time.Hour),
		},
		{
			Name:           "agent3",
			Description:    "Third agent",
			Tools:          []string{"Read", "Write"},
			ToolsInherited: false,
			Prompt:         "Third prompt",
			Source:         "source-a",
			InstalledAt:    baseTime,
		},
	}

	for _, agent := range agents {
		im.AddAgent(agent)
	}

	// Test source filter
	opts := QueryOptions{Source: "source-a"}
	results, err := im.Search("agent", opts)
	if err != nil {
		t.Errorf("Search with source filter failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results with source filter, got %d", len(results))
	}

	// Test time filter
	opts = QueryOptions{After: baseTime.Add(-90 * time.Minute)}
	results, err = im.Search("agent", opts)
	if err != nil {
		t.Errorf("Search with time filter failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results with time filter, got %d", len(results))
	}

	// Test NoTools filter (inherited tools)
	opts = QueryOptions{NoTools: true}
	results, err = im.Search("agent", opts)
	if err != nil {
		t.Errorf("Search with NoTools filter failed: %v", err)
	}
	if len(results) != 1 || results[0].Name != "agent2" {
		t.Errorf("Expected 1 result (agent2) with NoTools filter, got %d", len(results))
	}

	// Test CustomTools filter (explicit tools)
	opts = QueryOptions{CustomTools: true}
	results, err = im.Search("agent", opts)
	if err != nil {
		t.Errorf("Search with CustomTools filter failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results with CustomTools filter, got %d", len(results))
	}

	// Test limit
	opts = QueryOptions{Limit: 1}
	results, err = im.Search("agent", opts)
	if err != nil {
		t.Errorf("Search with limit failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result with limit=1, got %d", len(results))
	}
}

// TestSearchByName tests name-specific search
func TestSearchByName(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	agents := []*parser.AgentSpec{
		createTestAgent("go-expert", "Go programming", []string{"Read"}, "Go help"),
		createTestAgent("python-go", "Python with Go", []string{"Write"}, "Python help"),
		createTestAgent("web-dev", "Web development", []string{"Edit"}, "Web help"),
	}

	for _, agent := range agents {
		im.AddAgent(agent)
	}

	results, err := im.SearchByName("go")
	if err != nil {
		t.Errorf("SearchByName failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for name search 'go', got %d", len(results))
	}

	// Check that both go-related agents are found
	foundNames := make(map[string]bool)
	for _, result := range results {
		foundNames[result.Name] = true
	}

	if !foundNames["go-expert"] || !foundNames["python-go"] {
		t.Error("Expected to find both go-expert and python-go")
	}
}

// TestSearchByTools tests tool-specific search
func TestSearchByTools(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	agents := []*parser.AgentSpec{
		createTestAgent("agent1", "First", []string{"Read", "Write"}, "Prompt 1"),
		createTestAgent("agent2", "Second", []string{"Read", "Edit", "Bash"}, "Prompt 2"),
		createTestAgent("agent3", "Third", []string{"Write"}, "Prompt 3"),
		createTestAgent("agent4", "Fourth", []string{}, "Prompt 4"), // Inherited tools
	}

	for _, agent := range agents {
		im.AddAgent(agent)
	}

	// Search for agents with Read tool
	results, err := im.SearchByTools([]string{"Read"})
	if err != nil {
		t.Errorf("SearchByTools failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 agents with Read tool, got %d", len(results))
	}

	// Search for agents with Read AND Write tools
	results, err = im.SearchByTools([]string{"Read", "Write"})
	if err != nil {
		t.Errorf("SearchByTools failed: %v", err)
	}

	if len(results) != 1 || results[0].Name != "agent1" {
		t.Errorf("Expected 1 agent (agent1) with Read AND Write, got %d", len(results))
	}

	// Search for agents with nonexistent tool
	results, err = im.SearchByTools([]string{"NonexistentTool"})
	if err != nil {
		t.Errorf("SearchByTools failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 agents with nonexistent tool, got %d", len(results))
	}
}

// TestGetByFilename tests filename lookup
func TestGetByFilename(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	agent := createTestAgent("test-agent", "Test description", []string{"Read"}, "Test prompt")
	im.AddAgent(agent)

	// Test exact filename match
	found := im.GetByFilename("test-agent.md")
	if found == nil {
		t.Error("Expected to find agent by filename")
	} else if found.Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", found.Name)
	}

	// Test nonexistent filename
	notFound := im.GetByFilename("nonexistent.md")
	if notFound != nil {
		t.Error("Expected nil for nonexistent filename")
	}
}

// TestGetAll tests retrieving all agents
func TestGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	// Start empty
	agents := im.GetAll()
	if len(agents) != 0 {
		t.Errorf("Expected 0 agents initially, got %d", len(agents))
	}

	// Add some agents
	testAgents := []*parser.AgentSpec{
		createTestAgent("agent1", "First", []string{"Read"}, "First prompt"),
		createTestAgent("agent2", "Second", []string{"Write"}, "Second prompt"),
		createTestAgent("agent3", "Third", []string{"Edit"}, "Third prompt"),
	}

	for _, agent := range testAgents {
		im.AddAgent(agent)
	}

	agents = im.GetAll()
	if len(agents) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(agents))
	}

	// Verify it returns copies (not original slices that could be modified)
	agents[0] = nil // This shouldn't affect the internal state

	agentsAgain := im.GetAll()
	if agentsAgain[0] == nil {
		t.Error("GetAll should return copies, not references to internal slice")
	}
}

// TestRebuild tests rebuilding the index from a directory
func TestRebuild(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	// Create test agent files
	agents := []struct {
		filename string
		content  string
	}{
		{
			"agent1.md",
			`---
name: agent1
description: First test agent
tools: [Read]
---
First agent prompt.`,
		},
		{
			"agent2.md",
			`---
name: agent2
description: Second test agent
---
Second agent prompt.`,
		},
		{
			"invalid.md",
			`This is not a valid agent file.`,
		},
		{
			"not-agent.txt",
			`This file should be ignored.`,
		},
	}

	for _, agent := range agents {
		filePath := filepath.Join(tmpDir, agent.filename)
		if err := os.WriteFile(filePath, []byte(agent.content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", agent.filename, err)
		}
	}

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	// Index should be empty initially
	agents_before := im.GetAll()
	if len(agents_before) != 0 {
		t.Errorf("Expected 0 agents before rebuild, got %d", len(agents_before))
	}

	// Rebuild from directory
	err = im.Rebuild(tmpDir)
	if err != nil {
		t.Errorf("Rebuild failed: %v", err)
	}

	// Should now have 2 valid agents (invalid.md should be skipped)
	agents_after := im.GetAll()
	if len(agents_after) != 2 {
		t.Errorf("Expected 2 agents after rebuild, got %d", len(agents_after))
	}

	// Verify specific agents are present
	agentNames := make(map[string]bool)
	for _, agent := range agents_after {
		agentNames[agent.Name] = true
	}

	if !agentNames["agent1"] || !agentNames["agent2"] {
		t.Error("Expected to find agent1 and agent2 after rebuild")
	}

	// Verify name and filename lookups work
	agent1 := im.byName["agent1"]
	if agent1 == nil {
		t.Error("agent1 not found in name lookup after rebuild")
	}

	agent1_file := im.GetByFilename("agent1.md")
	if agent1_file == nil {
		t.Error("agent1.md not found in filename lookup after rebuild")
	}
}

// TestConcurrentAccess tests thread-safety of the index manager
func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	// Add some initial agents
	for i := 0; i < 10; i++ {
		agent := createTestAgent(
			fmt.Sprintf("agent%d", i),
			fmt.Sprintf("Agent %d description", i),
			[]string{"Read"},
			fmt.Sprintf("Agent %d prompt", i),
		)
		im.AddAgent(agent)
	}

	// Run concurrent operations
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				agents := im.GetAll()
				if len(agents) < 10 {
					errors <- fmt.Errorf("expected at least 10 agents, got %d", len(agents))
					return
				}

				// Search operations
				_, err := im.Search("agent", QueryOptions{})
				if err != nil {
					errors <- err
					return
				}

				// Name search
				_, err = im.SearchByName("agent")
				if err != nil {
					errors <- err
					return
				}

				// Filename lookup
				agent := im.GetByFilename("agent0.md")
				if agent == nil {
					errors <- fmt.Errorf("agent0.md not found")
					return
				}
			}
		}()
	}

	// Concurrent writers (adding more agents)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				agent := createTestAgent(
					fmt.Sprintf("concurrent-agent%d-%d", id, j),
					fmt.Sprintf("Concurrent agent %d-%d", id, j),
					[]string{"Write"},
					fmt.Sprintf("Concurrent prompt %d-%d", id, j),
				)
				im.AddAgent(agent)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify final state
	finalAgents := im.GetAll()
	expectedMin := 10 + (5 * 50) // Original + concurrent additions
	if len(finalAgents) < expectedMin {
		t.Errorf("Expected at least %d agents after concurrent operations, got %d", expectedMin, len(finalAgents))
	}
}

// TestIndexConsistency tests that internal lookups stay consistent
func TestIndexConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "test-index.json")

	im, err := NewIndexManager(indexPath)
	if err != nil {
		t.Fatalf("NewIndexManager failed: %v", err)
	}

	agents := []*parser.AgentSpec{
		createTestAgent("agent1", "First agent", []string{"Read"}, "First prompt"),
		createTestAgent("agent2", "Second agent", []string{"Write"}, "Second prompt"),
		createTestAgent("agent3", "Third agent", []string{"Edit"}, "Third prompt"),
	}

	// Add agents
	for _, agent := range agents {
		im.AddAgent(agent)
	}

	// Verify consistency between different access methods
	allAgents := im.GetAll()
	if len(allAgents) != 3 {
		t.Errorf("Expected 3 agents in GetAll(), got %d", len(allAgents))
	}

	// Check each agent is accessible via both name and filename lookups
	for _, agent := range allAgents {
		// Name lookup
		byName := im.byName[agent.Name]
		if byName == nil {
			t.Errorf("Agent %s not found in name lookup", agent.Name)
		}
		if !reflect.DeepEqual(byName, agent) {
			t.Errorf("Agent %s from name lookup doesn't match", agent.Name)
		}

		// Filename lookup
		byFile := im.GetByFilename(agent.FileName)
		if byFile == nil {
			t.Errorf("Agent %s not found in filename lookup", agent.FileName)
		}
		if !reflect.DeepEqual(byFile, agent) {
			t.Errorf("Agent %s from filename lookup doesn't match", agent.FileName)
		}
	}
}
