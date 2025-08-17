package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace"
)




func TestSubagentsHandler_FormatAgentContent(t *testing.T) {
	// Create handler struct directly without initializing container
	handler := &SubagentsHandler{
		config: &config.Config{},
	}

	agent := marketplace.Agent{
		Name:        "Test Agent",
		Description: "A test agent for testing",
		Category:    "Testing",
		Author:      "Test Author",
		Rating:      4.5,
		Downloads:   1000,
		Tags:        []string{"test", "example"},
		CreatedAt:   time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2023, 12, 26, 0, 0, 0, 0, time.UTC),
		ContentURL:  "https://example.com/agent",
	}

	content := "This is the agent content."
	result := handler.formatAgentContent(agent, content)

	// Check for expected frontmatter fields
	expectedElements := []string{
		"name: Test Agent",
		"description: A test agent for testing",
		"category: Testing",
		"author: Test Author",
		"rating: 4.5",
		"downloads: 1000",
		"tags: test, example",
		"created_at: 2023-12-25",
		"updated_at: 2023-12-26",
		"source: subagents.sh",
		"source_url: https://example.com/agent",
		"This is the agent content.",
	}

	for _, element := range expectedElements {
		if !containsString(result, element) {
			t.Errorf("Formatted content missing expected element: %q", element)
		}
	}

	// Check that it starts and ends with proper frontmatter delimiters
	if !startsWithFrontmatter(result) {
		t.Error("Formatted content should start with frontmatter delimiter")
	}
}

func TestSubagentsHandler_GenerateVersionHash(t *testing.T) {
	// Create handler struct directly without initializing container
	handler := &SubagentsHandler{
		config: &config.Config{},
	}

	agents1 := make([]marketplace.Agent, 5)
	agents2 := make([]marketplace.Agent, 3)

	// Populate with test data
	for i := range agents1 {
		agents1[i] = marketplace.Agent{
			Name:      fmt.Sprintf("Agent %d", i),
			UpdatedAt: time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
		}
	}
	for i := range agents2 {
		agents2[i] = marketplace.Agent{
			Name:      fmt.Sprintf("Agent %d", i),
			UpdatedAt: time.Date(2023, 12, 25, 10, 0, 0, 0, time.UTC),
		}
	}

	hash1 := handler.generateVersionHash(agents1)
	hash2 := handler.generateVersionHash(agents2)

	// Hash should be non-empty
	if hash1 == "" {
		t.Error("Generated hash should not be empty")
	}

	// Hash should start with "subagents-"
	if !startsWith(hash1, "subagents-") {
		t.Error("Generated hash should start with 'subagents-'")
	}

	// The generateVersionHash function uses current time, so hashes will generally be different
	// even for the same input when called at different times
	if hash1 == hash2 {
		// This could happen if both calls occur within the same second
		// but with different agent counts they should still differ
		t.Log("Hashes generated at same timestamp with different agent counts")
	}

	// Sleep for at least a second to ensure different timestamp
	// (generateVersionHash uses second-precision timestamps)
	time.Sleep(time.Second + time.Millisecond*100)
	hash3 := handler.generateVersionHash(agents1)

	// Even with same agents, different timestamps should generate different hashes
	if hash1 == hash3 {
		t.Error("Same agents at different times should generate different hashes due to timestamp")
	}

	// Verify hash format is consistent
	if len(hash1) != len(hash3) {
		t.Error("Hash length should be consistent")
	}
}

func TestApplyFilters(t *testing.T) {
	// Create a mock installer to test the applyFilters method
	cfg := &config.Config{}
	installer := &Installer{config: cfg}

	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "filter-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"agent1.md",
		"agent2.md",
		"script.py",
		"readme.txt",
		"subdir/agent3.md",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	tests := []struct {
		name     string
		filters  config.FilterConfig
		expected int
	}{
		{
			name: "include .md files",
			filters: config.FilterConfig{
				Include: config.IncludeFilter{
					Extensions: []string{".md"},
				},
			},
			expected: 3,
		},
		{
			name: "include pattern",
			filters: config.FilterConfig{
				Include: config.IncludeFilter{
					Patterns: []string{"agent*"},
				},
			},
			expected: 3,
		},
		{
			name: "exclude pattern",
			filters: config.FilterConfig{
				Exclude: config.ExcludeFilter{
					Patterns: []string{"readme*"},
				},
			},
			expected: 4,
		},
		{
			name:     "no filters (include all)",
			filters:  config.FilterConfig{},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := installer.applyFilters(tempDir, tt.filters)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(files) != tt.expected {
				t.Errorf("Expected %d files, got %d", tt.expected, len(files))
			}
		})
	}
}

// Helper functions for testing
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findInString(s, substr)
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func startsWithFrontmatter(s string) bool {
	return startsWith(s, "---")
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
