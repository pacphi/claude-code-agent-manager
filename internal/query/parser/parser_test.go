package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestParseFile_Valid tests parsing of a valid agent file with all fields
func TestParseFile_Valid(t *testing.T) {
	content := `---
name: test-agent
description: A comprehensive test agent for validation
tools: [Read, Write, Edit]
---

This is the agent prompt content.
It can be multiple lines.

## Instructions
Follow these guidelines...`

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-agent.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	agent, err := parser.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify YAML frontmatter fields
	if agent.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", agent.Name)
	}

	if agent.Description != "A comprehensive test agent for validation" {
		t.Errorf("Expected description 'A comprehensive test agent for validation', got '%s'", agent.Description)
	}

	expectedTools := []string{"Read", "Write", "Edit"}
	if len(agent.GetToolsAsSlice()) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(agent.GetToolsAsSlice()))
	}
	for i, tool := range expectedTools {
		if i >= len(agent.GetToolsAsSlice()) || agent.GetToolsAsSlice()[i] != tool {
			t.Errorf("Expected tool[%d] to be '%s', got '%s'", i, tool, agent.GetToolsAsSlice()[i])
		}
	}

	// Verify derived fields
	if agent.ToolsInherited {
		t.Error("Expected ToolsInherited to be false when tools are explicitly specified")
	}

	expectedPrompt := `This is the agent prompt content.
It can be multiple lines.

## Instructions
Follow these guidelines...`
	if agent.Prompt != expectedPrompt {
		t.Errorf("Expected prompt content mismatch.\nExpected: %q\nGot: %q", expectedPrompt, agent.Prompt)
	}

	// Verify file metadata
	if agent.FilePath != testFile {
		t.Errorf("Expected FilePath '%s', got '%s'", testFile, agent.FilePath)
	}

	if agent.FileName != "test-agent.md" {
		t.Errorf("Expected FileName 'test-agent.md', got '%s'", agent.FileName)
	}

	if agent.FileSize <= 0 {
		t.Errorf("Expected FileSize > 0, got %d", agent.FileSize)
	}

	if agent.ModTime.IsZero() {
		t.Error("Expected ModTime to be set")
	}
}

// TestParseFile_InheritedTools tests parsing of agent with no tools specified
func TestParseFile_InheritedTools(t *testing.T) {
	content := `---
name: simple-agent
description: Simple agent without explicit tools
---

Simple agent prompt.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "simple-agent.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	agent, err := parser.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if !agent.ToolsInherited {
		t.Error("Expected ToolsInherited to be true when no tools specified")
	}

	if len(agent.GetToolsAsSlice()) != 0 {
		t.Errorf("Expected empty tools slice, got %v", agent.GetToolsAsSlice())
	}

	if agent.Prompt != "Simple agent prompt." {
		t.Errorf("Expected prompt 'Simple agent prompt.', got '%s'", agent.Prompt)
	}
}

// TestParseFile_EmptyTools tests parsing of agent with empty tools array
func TestParseFile_EmptyTools(t *testing.T) {
	content := `---
name: empty-tools-agent
description: Agent with empty tools array
tools: []
---

Agent with empty tools.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty-tools-agent.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	agent, err := parser.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if !agent.ToolsInherited {
		t.Error("Expected ToolsInherited to be true when tools array is empty")
	}

	if len(agent.GetToolsAsSlice()) != 0 {
		t.Errorf("Expected empty tools slice, got %v", agent.GetToolsAsSlice())
	}
}

// TestParseFile_InvalidYAML tests parsing of file with invalid YAML frontmatter
func TestParseFile_InvalidYAML(t *testing.T) {
	content := `---
name: invalid-yaml-agent
description: Agent with invalid YAML
tools: [Read, Write,  # Missing closing bracket
---

This should fail to parse.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid-yaml.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	_, err := parser.ParseFile(testFile)

	if err == nil {
		t.Error("Expected ParseFile to fail with invalid YAML")
	}

	if !strings.Contains(err.Error(), "failed to parse frontmatter") {
		t.Errorf("Expected error to mention frontmatter parsing, got: %v", err)
	}
}

// TestParseFile_MissingFrontmatter tests parsing of file without frontmatter delimiters
func TestParseFile_MissingFrontmatter(t *testing.T) {
	content := `This is just a regular markdown file without frontmatter.

## Heading
Some content here.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-frontmatter.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	_, err := parser.ParseFile(testFile)

	if err == nil {
		t.Error("Expected ParseFile to fail without frontmatter")
	}

	if !strings.Contains(err.Error(), "missing frontmatter") {
		t.Errorf("Expected error to mention missing frontmatter, got: %v", err)
	}
}

// TestParseFile_IncompleteFrontmatter tests parsing with only one frontmatter delimiter
func TestParseFile_IncompleteFrontmatter(t *testing.T) {
	content := `---
name: incomplete-agent
description: Agent with incomplete frontmatter

This content doesn't have closing frontmatter delimiter.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "incomplete-frontmatter.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	_, err := parser.ParseFile(testFile)

	if err == nil {
		t.Error("Expected ParseFile to fail with incomplete frontmatter")
	}

	if !strings.Contains(err.Error(), "missing frontmatter") {
		t.Errorf("Expected error to mention missing frontmatter, got: %v", err)
	}
}

// TestParseFile_EmptyFile tests parsing of empty file
func TestParseFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	_, err := parser.ParseFile(testFile)

	if err == nil {
		t.Error("Expected ParseFile to fail with empty file")
	}
}

// TestParseFile_UnicodeContent tests parsing with Unicode characters
func TestParseFile_UnicodeContent(t *testing.T) {
	content := `---
name: unicode-agent
description: "Agent with Unicode characters: 中文, Español, 🤖"
tools: [Read]
---

This is a prompt with Unicode content:
- Chinese: 你好世界
- Spanish: ¡Hola mundo!
- Emoji: 🚀 🎯 💡

Instructions with accented characters: naïve café.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unicode-agent.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	agent, err := parser.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile failed with Unicode content: %v", err)
	}

	expectedDescription := "Agent with Unicode characters: 中文, Español, 🤖"
	if agent.Description != expectedDescription {
		t.Errorf("Unicode description not preserved correctly.\nExpected: %q\nGot: %q", expectedDescription, agent.Description)
	}

	if !strings.Contains(agent.Prompt, "你好世界") {
		t.Error("Unicode characters in prompt not preserved")
	}

	if !strings.Contains(agent.Prompt, "🚀") {
		t.Error("Emoji characters in prompt not preserved")
	}
}

// TestParseFile_NonexistentFile tests parsing of file that doesn't exist
func TestParseFile_NonexistentFile(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFile("/nonexistent/path/file.md")

	if err == nil {
		t.Error("Expected ParseFile to fail with nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected error to mention file reading failure, got: %v", err)
	}
}

// TestParseDirectory_ValidDirectory tests parsing all agents in a directory
func TestParseDirectory_ValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test agent files
	agents := []struct {
		filename string
		content  string
	}{
		{
			"agent1.md",
			`---
name: agent-one
description: First test agent
tools: [Read]
---
First agent prompt.`,
		},
		{
			"agent2.md",
			`---
name: agent-two
description: Second test agent
---
Second agent prompt.`,
		},
		{
			"not-an-agent.txt", // Should be ignored
			"This is not an agent file.",
		},
		{
			"invalid-agent.md", // Should be skipped gracefully
			`This file has no frontmatter and should be skipped.`,
		},
	}

	for _, agent := range agents {
		filePath := filepath.Join(tmpDir, agent.filename)
		if err := os.WriteFile(filePath, []byte(agent.content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", agent.filename, err)
		}
	}

	parser := NewParser()
	parsedAgents, err := parser.ParseDirectory(tmpDir)

	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	// Should successfully parse 2 valid agents, skip invalid ones
	if len(parsedAgents) != 2 {
		t.Errorf("Expected 2 parsed agents, got %d", len(parsedAgents))
	}

	// Verify the parsed agents
	agentNames := make(map[string]bool)
	for _, agent := range parsedAgents {
		agentNames[agent.Name] = true
	}

	if !agentNames["agent-one"] {
		t.Error("Expected to find agent-one")
	}

	if !agentNames["agent-two"] {
		t.Error("Expected to find agent-two")
	}
}

// TestParseDirectory_EmptyDirectory tests parsing empty directory
func TestParseDirectory_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewParser()
	agents, err := parser.ParseDirectory(tmpDir)

	if err != nil {
		t.Fatalf("ParseDirectory failed on empty directory: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected 0 agents in empty directory, got %d", len(agents))
	}
}

// TestParseDirectory_NonexistentDirectory tests parsing nonexistent directory
func TestParseDirectory_NonexistentDirectory(t *testing.T) {
	parser := NewParser()
	agents, err := parser.ParseDirectory("/nonexistent/directory")

	if err != nil {
		t.Fatalf("ParseDirectory should handle nonexistent directory gracefully: %v", err)
	}

	if agents != nil {
		t.Errorf("Expected nil agents for nonexistent directory, got %v", agents)
	}
}

// TestParseDirectory_NestedDirectories tests parsing with nested subdirectories
func TestParseDirectory_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create agents in both root and subdirectory
	rootAgent := filepath.Join(tmpDir, "root-agent.md")
	subAgent := filepath.Join(subDir, "sub-agent.md")

	rootContent := `---
name: root-agent
description: Agent in root directory
---
Root agent prompt.`

	subContent := `---
name: sub-agent
description: Agent in subdirectory
---
Sub agent prompt.`

	if err := os.WriteFile(rootAgent, []byte(rootContent), 0644); err != nil {
		t.Fatalf("Failed to create root agent: %v", err)
	}

	if err := os.WriteFile(subAgent, []byte(subContent), 0644); err != nil {
		t.Fatalf("Failed to create sub agent: %v", err)
	}

	parser := NewParser()
	agents, err := parser.ParseDirectory(tmpDir)

	if err != nil {
		t.Fatalf("ParseDirectory failed: %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("Expected 2 agents (including nested), got %d", len(agents))
	}

	// Verify both agents were found
	agentNames := make(map[string]bool)
	for _, agent := range agents {
		agentNames[agent.Name] = true
	}

	if !agentNames["root-agent"] {
		t.Error("Expected to find root-agent")
	}

	if !agentNames["sub-agent"] {
		t.Error("Expected to find sub-agent")
	}
}

// TestParseFile_PermissionDenied tests handling of files without read permissions
func TestParseFile_PermissionDenied(t *testing.T) {
	// Skip this test on Windows where chmod doesn't work the same way
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "no-read.md")

	content := `---
name: no-read-agent
description: Agent file without read permission
---
This should fail to read.`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Remove read permission
	if err := os.Chmod(testFile, 0000); err != nil {
		t.Fatalf("Failed to change file permissions: %v", err)
	}

	// Restore permissions for cleanup
	defer os.Chmod(testFile, 0644)

	parser := NewParser()
	_, err := parser.ParseFile(testFile)

	if err == nil {
		t.Error("Expected ParseFile to fail with permission denied")
	}

	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("Expected error to mention file reading failure, got: %v", err)
	}
}

// TestNewParser tests parser creation
func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Error("NewParser should not return nil")
	}
}

// TestParseFile_LargeFile tests parsing of a large agent file
func TestParseFile_LargeFile(t *testing.T) {
	// Create a large prompt content (around 10KB)
	largePrompt := strings.Repeat("This is a line of the agent prompt.\n", 400)

	content := `---
name: large-agent
description: Agent with very large prompt content
tools: [Read, Write, Edit, Bash, Grep]
---

` + largePrompt

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large-agent.md")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	agent, err := parser.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile failed on large file: %v", err)
	}

	if agent.Name != "large-agent" {
		t.Errorf("Expected name 'large-agent', got '%s'", agent.Name)
	}

	if len(agent.Prompt) < 5000 {
		t.Errorf("Expected large prompt content, got length %d", len(agent.Prompt))
	}

	if agent.FileSize < 10000 {
		t.Errorf("Expected file size > 10KB, got %d bytes", agent.FileSize)
	}
}

// TestParseFile_TimestampAccuracy tests that file modification time is captured accurately
func TestParseFile_TimestampAccuracy(t *testing.T) {
	content := `---
name: timestamp-agent
description: Agent for testing timestamp accuracy
---
Timestamp test prompt.`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "timestamp-agent.md")

	// Record time before writing
	beforeWrite := time.Now()

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Record time after writing
	afterWrite := time.Now()

	parser := NewParser()
	agent, err := parser.ParseFile(testFile)

	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify timestamp is within expected range
	if agent.ModTime.Before(beforeWrite) || agent.ModTime.After(afterWrite) {
		t.Errorf("ModTime %v should be between %v and %v", agent.ModTime, beforeWrite, afterWrite)
	}
}

// TestFlexibleTools_StringFormat tests that tools can be parsed from comma-separated string
func TestFlexibleTools_StringFormat(t *testing.T) {
	yamlContent := `---
name: test-agent
description: A test agent
tools: Read, Write, Edit, MultiEdit, Bash
---

This is the test prompt content.
`

	tmpFile, err := os.CreateTemp("", "test-agent-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := NewParser()
	agent, err := parser.ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	expected := []string{"Read", "Write", "Edit", "MultiEdit", "Bash"}
	tools := agent.GetToolsAsSlice()

	if len(tools) != len(expected) {
		t.Errorf("Expected %d tools, got %d", len(expected), len(tools))
	}

	for i, expectedTool := range expected {
		if i >= len(tools) || tools[i] != expectedTool {
			t.Errorf("Expected tool[%d] to be '%s', got '%s'", i, expectedTool, tools[i])
		}
	}

	if agent.ToolsInherited {
		t.Error("Expected ToolsInherited to be false when tools are explicitly specified")
	}
}

// TestFlexibleTools_ArrayFormat tests that tools can be parsed from YAML array
func TestFlexibleTools_ArrayFormat(t *testing.T) {
	yamlContent := `---
name: test-agent
description: A test agent
tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash
---

This is the test prompt content.
`

	tmpFile, err := os.CreateTemp("", "test-agent-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := NewParser()
	agent, err := parser.ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	expected := []string{"Read", "Write", "Edit", "MultiEdit", "Bash"}
	tools := agent.GetToolsAsSlice()

	if len(tools) != len(expected) {
		t.Errorf("Expected %d tools, got %d", len(expected), len(tools))
	}

	for i, expectedTool := range expected {
		if i >= len(tools) || tools[i] != expectedTool {
			t.Errorf("Expected tool[%d] to be '%s', got '%s'", i, expectedTool, tools[i])
		}
	}

	if agent.ToolsInherited {
		t.Error("Expected ToolsInherited to be false when tools are explicitly specified")
	}
}

// TestFlexibleTools_EmptyString tests handling of empty string tools
func TestFlexibleTools_EmptyString(t *testing.T) {
	yamlContent := `---
name: test-agent
description: A test agent
tools: ""
---

This is the test prompt content.
`

	tmpFile, err := os.CreateTemp("", "test-agent-*.md")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	parser := NewParser()
	agent, err := parser.ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	tools := agent.GetToolsAsSlice()
	if len(tools) != 0 {
		t.Errorf("Expected empty tools slice for empty string, got %v", tools)
	}

	if !agent.ToolsInherited {
		t.Error("Expected ToolsInherited to be true when tools is empty string")
	}
}

// TestFlexibleTools_DirectUnmarshal tests the UnmarshalYAML method directly
func TestFlexibleTools_DirectUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected []string
		wantErr  bool
	}{
		{
			name:     "string format",
			yaml:     `"Read, Write, Edit"`,
			expected: []string{"Read", "Write", "Edit"},
			wantErr:  false,
		},
		{
			name:     "array format",
			yaml:     `["Read", "Write", "Edit"]`,
			expected: []string{"Read", "Write", "Edit"},
			wantErr:  false,
		},
		{
			name:     "empty string",
			yaml:     `""`,
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "single tool string",
			yaml:     `"Read"`,
			expected: []string{"Read"},
			wantErr:  false,
		},
		{
			name:     "string with spaces",
			yaml:     `"Read , Write , Edit "`,
			expected: []string{"Read", "Write", "Edit"},
			wantErr:  false,
		},
		{
			name:     "number format",
			yaml:     `123`,
			expected: []string{"123"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexibleTools
			err := yaml.Unmarshal([]byte(tt.yaml), &ft)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			tools := ft.GetTools()
			if len(tools) != len(tt.expected) {
				t.Errorf("Expected %d tools, got %d", len(tt.expected), len(tools))
			}

			for i, expected := range tt.expected {
				if i >= len(tools) || tools[i] != expected {
					t.Errorf("Expected tool[%d] to be '%s', got '%s'", i, expected, tools[i])
				}
			}
		})
	}
}
