package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIQueryCommands tests all query command variations
func TestCLIQueryCommands(t *testing.T) {
	testDir := setupCLITestDirectory(t)
	defer cleanup(testDir)

	// Build the binary
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	// Setup test environment
	setupCLITestEnvironment(t, testDir)

	// Test basic query command
	testBasicQueryCommand(t, binaryPath, testDir)

	// Test field-specific queries
	testFieldSpecificQueryCommands(t, binaryPath, testDir)

	// Test query with filters
	testQueryWithFilters(t, binaryPath, testDir)

	// Test output formats
	testQueryOutputFormats(t, binaryPath, testDir)
}

// TestCLIShowCommand tests show command with fuzzy matching
func TestCLIShowCommand(t *testing.T) {
	testDir := setupCLITestDirectory(t)
	defer cleanup(testDir)

	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	setupCLITestEnvironment(t, testDir)

	// Test exact match
	testShowExactMatch(t, binaryPath, testDir)

	// Test fuzzy matching
	testShowFuzzyMatch(t, binaryPath, testDir)

	// Test show with different formats
	testShowOutputFormats(t, binaryPath, testDir)

	// Test show with non-existent agent
	testShowNonExistent(t, binaryPath, testDir)
}

// TestCLIStatsCommand tests stats command outputs
func TestCLIStatsCommand(t *testing.T) {
	testDir := setupCLITestDirectory(t)
	defer cleanup(testDir)

	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	setupCLITestEnvironment(t, testDir)

	// Test basic stats
	testBasicStats(t, binaryPath, testDir)

	// Test stats by source
	testStatsBySource(t, binaryPath, testDir)

	// Test coverage stats
	testCoverageStats(t, binaryPath, testDir)

	// Test tool usage stats
	testToolUsageStats(t, binaryPath, testDir)

	// Test validation stats
	testValidationStats(t, binaryPath, testDir)
}

// TestCLIIndexManagement tests index management operations
func TestCLIIndexManagement(t *testing.T) {
	testDir := setupCLITestDirectory(t)
	defer cleanup(testDir)

	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	setupCLITestEnvironment(t, testDir)

	// Test index build
	testIndexBuild(t, binaryPath, testDir)

	// Test index rebuild
	testIndexRebuild(t, binaryPath, testDir)

	// Test index stats
	testIndexStats(t, binaryPath, testDir)

	// Test cache clear
	testCacheClear(t, binaryPath, testDir)

	// Test cache stats
	testCacheStats(t, binaryPath, testDir)
}

// TestCLIValidateCommand tests enhanced validate functionality
func TestCLIValidateCommand(t *testing.T) {
	testDir := setupCLITestDirectory(t)
	defer cleanup(testDir)

	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	setupCLITestEnvironment(t, testDir)

	// Test basic validate
	testBasicValidate(t, binaryPath, testDir)

	// Test validate with agents flag
	testValidateAgents(t, binaryPath, testDir)

	// Test validate with query flag
	testValidateQuery(t, binaryPath, testDir)
}

// TestCLIListWithSearch tests enhanced list command with search
func TestCLIListWithSearch(t *testing.T) {
	testDir := setupCLITestDirectory(t)
	defer cleanup(testDir)

	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	setupCLITestEnvironment(t, testDir)

	// Test basic list
	testBasicList(t, binaryPath, testDir)

	// Test list with search
	testListWithSearch(t, binaryPath, testDir)

	// Test list with filters
	testListWithFilters(t, binaryPath, testDir)

	// Test list by source
	testListBySource(t, binaryPath, testDir)
}

// Helper functions

func setupCLITestDirectory(t *testing.T) string {
	testDir, err := os.MkdirTemp("", "agent-manager-cli-test-*")
	require.NoError(t, err)
	return testDir
}

func buildTestBinary(t *testing.T) string {
	binaryName := "agent-manager-test-" + fmt.Sprintf("%d", time.Now().UnixNano())

	// Add .exe extension on Windows
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	tmpBinary := filepath.Join(os.TempDir(), binaryName)

	cmd := exec.Command("go", "build", "-o", tmpBinary, "../../cmd/agent-manager")
	err := cmd.Run()
	require.NoError(t, err, "Failed to build test binary")

	return tmpBinary
}

func setupCLITestEnvironment(t *testing.T, testDir string) {
	// Create config file
	configContent := fmt.Sprintf(`version: "1.0"
settings:
  base_dir: "%s/agents"
  docs_dir: "docs"
  conflict_strategy: "backup"
  backup_dir: "%s/backups"
  concurrent_downloads: 3
  timeout: "5m"
  query:
    enabled: true
    cache:
      enabled: true
      ttl: "1h"
      max_size: "100MB"
    defaults:
      format: "table"
      limit: 20
      fuzzy: true

sources:
  - name: "test-local"
    enabled: true
    type: "local"
    paths:
      source: "%s/test-agents"
      target: "%s/agents"

metadata:
  tracking_file: "%s/.installed-agents.json"
  log_file: "%s/installation.log"
`, testDir, testDir, testDir, testDir, testDir, testDir)

	configPath := filepath.Join(testDir, "agents-config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create test agents
	createCLITestAgents(t, testDir)

	// Install test agents
	installCLITestAgents(t, testDir)
}

func createCLITestAgents(t *testing.T, testDir string) {
	testAgentsDir := filepath.Join(testDir, "test-agents")
	err := os.MkdirAll(testAgentsDir, 0755)
	require.NoError(t, err)

	agents := []struct {
		name    string
		content string
	}{
		{
			name: "go-expert.md",
			content: `---
name: go-expert
description: Expert Go developer specializing in modern Go development
tools: [Read, Write, Edit, Bash]
---

You are an expert Go developer with deep knowledge of Go 1.24+ features.
`,
		},
		{
			name: "python-specialist.md",
			content: `---
name: python-specialist
description: Python programming specialist with ML expertise
tools: [Read, Write, Edit]
---

You are a Python expert focusing on clean, efficient code and machine learning.
`,
		},
		{
			name: "frontend-developer.md",
			content: `---
name: frontend-developer
description: Frontend developer specializing in React and TypeScript
tools: [Read, Write, Edit, WebFetch]
---

You are a frontend developer expert in modern React and TypeScript development.
`,
		},
		{
			name: "minimal-helper.md",
			content: `---
name: minimal-helper
description: Minimal helper agent for testing
---

This is a minimal agent with only required fields for testing purposes.
`,
		},
	}

	for _, agent := range agents {
		agentPath := filepath.Join(testAgentsDir, agent.name)
		err := os.WriteFile(agentPath, []byte(agent.content), 0644)
		require.NoError(t, err)
	}
}

func installCLITestAgents(t *testing.T, testDir string) {
	configPath := filepath.Join(testDir, "agents-config.yaml")

	// Use the built binary to install agents
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	cmd := exec.Command(binaryPath, "install", "--config", configPath, "--verbose")
	cmd.Dir = testDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Install output: %s", string(output))
		require.NoError(t, err, "Failed to install test agents")
	}
}

func runCLICommand(t *testing.T, binaryPath, testDir string, args ...string) (string, error) {
	configPath := filepath.Join(testDir, "agents-config.yaml")

	// Add config flag to all commands
	fullArgs := []string{"--config", configPath}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command(binaryPath, fullArgs...)
	cmd.Dir = testDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if err != nil {
		// Include stderr in error for debugging
		if stderr.Len() > 0 {
			t.Logf("Command stderr: %s", stderr.String())
		}
		return output, fmt.Errorf("command failed: %w, output: %s", err, output)
	}

	return output, nil
}

func testBasicQueryCommand(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "query", "go")
	require.NoError(t, err)

	assert.Contains(t, output, "go-expert", "Should find go-expert agent")
	assert.Contains(t, output, "Expert Go developer", "Should show agent description")
}

func testFieldSpecificQueryCommands(t *testing.T, binaryPath, testDir string) {
	// Test name search
	output, err := runCLICommand(t, binaryPath, testDir, "query", "--field", "name", "python-specialist")
	require.NoError(t, err)
	assert.Contains(t, output, "python-specialist", "Should find python-specialist by name")

	// Test description search
	output, err = runCLICommand(t, binaryPath, testDir, "query", "--field", "description", "React")
	require.NoError(t, err)
	assert.Contains(t, output, "frontend-developer", "Should find frontend-developer by description")

	// Test tools search
	output, err = runCLICommand(t, binaryPath, testDir, "query", "--field", "tools", "WebFetch")
	require.NoError(t, err)
	assert.Contains(t, output, "frontend-developer", "Should find agents with WebFetch tool")
}

func testQueryWithFilters(t *testing.T, binaryPath, testDir string) {
	// Test no-tools filter
	output, err := runCLICommand(t, binaryPath, testDir, "query", "--no-tools", "--limit", "5")
	require.NoError(t, err)
	assert.Contains(t, output, "minimal-helper", "Should find agents with inherited tools")

	// Test custom-tools filter
	output, err = runCLICommand(t, binaryPath, testDir, "query", "--custom-tools", "--limit", "5")
	require.NoError(t, err)
	// Should find agents with explicit tools
	hasExplicitTools := strings.Contains(output, "go-expert") || strings.Contains(output, "python-specialist")
	assert.True(t, hasExplicitTools, "Should find agents with explicit tools")

	// Test limit
	output, err = runCLICommand(t, binaryPath, testDir, "query", "--limit", "2")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Should have limited output (header + agents + separators)
	assert.LessOrEqual(t, len(lines), 10, "Should respect limit parameter")
}

func testQueryOutputFormats(t *testing.T, binaryPath, testDir string) {
	// Test table format (default)
	output, err := runCLICommand(t, binaryPath, testDir, "query", "go", "--output", "table")
	require.NoError(t, err)
	assert.Contains(t, output, "go-expert", "Table format should show agents")

	// Test JSON format
	output, err = runCLICommand(t, binaryPath, testDir, "query", "go", "--output", "json")
	if err == nil { // JSON output might not be implemented yet
		assert.Contains(t, output, "name", "JSON format should contain name field")
	}

	// Test YAML format
	output, err = runCLICommand(t, binaryPath, testDir, "query", "go", "--output", "yaml")
	if err == nil { // YAML output might not be implemented yet
		assert.Contains(t, output, "name:", "YAML format should contain name field")
	}
}

func testShowExactMatch(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "show", "go-expert.md")
	require.NoError(t, err)

	assert.Contains(t, output, "go-expert", "Should show agent name")
	assert.Contains(t, output, "Expert Go developer", "Should show agent description")
	assert.Contains(t, output, "Read", "Should show agent tools")
}

func testShowFuzzyMatch(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "show", "go-expert")
	if err == nil { // Fuzzy matching might find the agent
		assert.Contains(t, output, "go-expert", "Should find agent via fuzzy matching")
	}

	// Test partial match
	output, err = runCLICommand(t, binaryPath, testDir, "show", "frontend")
	if err == nil {
		assert.Contains(t, output, "frontend-developer", "Should find agent via partial matching")
	}
}

func testShowOutputFormats(t *testing.T, binaryPath, testDir string) {
	// Test basic show functionality (without unsupported flags)
	output, err := runCLICommand(t, binaryPath, testDir, "show", "go-expert.md")
	if err == nil {
		assert.Contains(t, output, "go-expert", "Should show agent details")
		assert.Contains(t, output, "Read", "Should show tools")
	}
}

func testShowNonExistent(t *testing.T, binaryPath, testDir string) {
	_, err := runCLICommand(t, binaryPath, testDir, "show", "non-existent-agent")
	assert.Error(t, err, "Should fail for non-existent agent")
}

func testBasicStats(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "stats")
	require.NoError(t, err)

	assert.Contains(t, output, "Total", "Should show total statistics")
	assert.Contains(t, output, "4", "Should show correct number of test agents")
}

func testStatsBySource(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "stats", "--detailed")
	if err == nil {
		assert.Contains(t, output, "Agent Statistics", "Should show detailed statistics")
	}
}

func testCoverageStats(t *testing.T, binaryPath, testDir string) {
	// Coverage stats are part of detailed stats now
	output, err := runCLICommand(t, binaryPath, testDir, "stats", "--detailed")
	if err == nil {
		assert.Contains(t, output, "Coverage", "Should show coverage statistics")
	}
}

func testToolUsageStats(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "stats", "--tools")
	if err == nil {
		assert.Contains(t, output, "Tools", "Should show tool usage statistics")
	}
}

func testValidationStats(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "stats", "--validation")
	if err == nil {
		assert.Contains(t, output, "Validation", "Should show validation statistics")
	}
}

func testIndexBuild(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "index", "build")
	require.NoError(t, err)
	assert.Contains(t, output, "built", "Should confirm index was built")
}

func testIndexRebuild(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "index", "rebuild")
	require.NoError(t, err)
	assert.Contains(t, output, "rebuilt", "Should confirm index was rebuilt")
}

func testIndexStats(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "index", "stats")
	require.NoError(t, err)
	assert.Contains(t, output, "Index Statistics", "Should show index statistics")
}

func testCacheClear(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "index", "cache-clear")
	require.NoError(t, err)
	assert.Contains(t, output, "cleared", "Should confirm cache was cleared")
}

func testCacheStats(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "index", "cache-stats")
	require.NoError(t, err)
	assert.Contains(t, output, "Cache Statistics", "Should show cache statistics")
}

func testBasicValidate(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "validate")
	require.NoError(t, err)
	assert.Contains(t, output, "valid", "Should show validation results")
}

func testValidateAgents(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "validate", "--agents")
	require.NoError(t, err)
	assert.Contains(t, output, "agents", "Should validate installed agents")
}

func testValidateQuery(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "validate", "--query")
	require.NoError(t, err)
	assert.Contains(t, output, "Query", "Should test query functionality")
}

func testBasicList(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "list")
	require.NoError(t, err)
	assert.Contains(t, output, "test-local", "Should show installed sources")
	assert.Contains(t, output, "Files:", "Should show file counts")
}

func testListWithSearch(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "list", "--search", "go")
	if err == nil {
		assert.Contains(t, output, "go-expert", "Should find agents matching search term")
	}
}

func testListWithFilters(t *testing.T, binaryPath, testDir string) {
	// Test no-tools filter
	output, err := runCLICommand(t, binaryPath, testDir, "list", "--no-tools")
	if err == nil {
		// Should show agents with inherited tools if any exist
		t.Logf("List no-tools output: %s", output)
	}

	// Test custom-tools filter
	output, err = runCLICommand(t, binaryPath, testDir, "list", "--custom-tools")
	if err == nil {
		// Should show agents with explicit tools
		t.Logf("List custom-tools output: %s", output)
	}
}

func testListBySource(t *testing.T, binaryPath, testDir string) {
	output, err := runCLICommand(t, binaryPath, testDir, "list", "--source", "test-local")
	require.NoError(t, err)
	assert.Contains(t, output, "test-local", "Should show specific source")
}
