package stats

import (
	"testing"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/stretchr/testify/assert"
)

func TestNewCalculator(t *testing.T) {
	agents := []*parser.AgentSpec{
		{Name: "test1", Description: "Test agent 1"},
		{Name: "test2", Description: "Test agent 2"},
	}

	calc := NewCalculator(agents)
	assert.NotNil(t, calc)
	assert.Equal(t, 2, len(calc.agents))
}

func TestCalculator_Calculate(t *testing.T) {
	agents := []*parser.AgentSpec{
		{
			Name:           "complete-agent",
			Description:    "Complete test agent with all fields",
			Tools:          []string{"Read", "Write"},
			ToolsInherited: false,
			Prompt:         "You are a complete agent",
			Source:         "github",
			FilePath:       "/path/to/complete-agent.md",
		},
		{
			Name:           "inherited-tools-agent",
			Description:    "Agent with inherited tools",
			ToolsInherited: true,
			Prompt:         "You are an inherited tools agent",
			Source:         "local",
			FilePath:       "/path/to/inherited-agent.md",
		},
		{
			Name:           "minimal-agent",
			Description:    "Minimal agent",
			ToolsInherited: true,
			Prompt:         "You are minimal",
			Source:         "github",
			FilePath:       "/path/to/minimal-agent.md",
		},
		{
			Name:           "duplicate-agent", // Duplicate name for testing
			Description:    "First duplicate",
			ToolsInherited: true,
			Prompt:         "You are first",
			Source:         "local",
			FilePath:       "/path/to/duplicate1.md",
		},
		{
			Name:           "duplicate-agent", // Duplicate name for testing
			Description:    "Second duplicate",
			Tools:          []string{"Bash"},
			ToolsInherited: false,
			Prompt:         "You are second",
			Source:         "git",
			FilePath:       "/path/to/duplicate2.md",
		},
	}

	calc := NewCalculator(agents)
	stats := calc.Calculate()

	// Test basic counts
	assert.Equal(t, 5, stats.TotalAgents)

	// Test source distribution
	expected := map[string]int{
		"github": 2,
		"local":  2,
		"git":    1,
	}
	assert.Equal(t, expected, stats.BySource)

	// Test coverage stats
	assert.Equal(t, 5, stats.Coverage.WithName)        // All have names
	assert.Equal(t, 5, stats.Coverage.WithDescription) // All have descriptions
	assert.Equal(t, 2, stats.Coverage.WithTools)       // 2 have explicit tools
	assert.Equal(t, 5, stats.Coverage.WithPrompt)      // All have prompts

	// Average coverage should be 100% since all have name, desc, and prompt
	// Tools are optional so not counted against coverage
	assert.Equal(t, 100.0, stats.Coverage.AverageCoverage)

	// Test tool usage
	assert.Equal(t, 3, stats.ToolUsage.InheritedTools) // 3 agents with inherited tools
	assert.Equal(t, 2, stats.ToolUsage.ExplicitTools)  // 2 agents with explicit tools

	expectedToolDist := map[string]int{
		"Read":  1,
		"Write": 1,
		"Bash":  1,
	}
	assert.Equal(t, expectedToolDist, stats.ToolUsage.ToolDistribution)

	// Test duplicates
	assert.Len(t, stats.Duplicates, 1)
	assert.Contains(t, stats.Duplicates, "duplicate-agent")
	assert.Len(t, stats.Duplicates["duplicate-agent"], 2)

	// Test orphaned agents (should be 0 for valid agents)
	assert.Equal(t, 0, stats.OrphanedAgents)
}

func TestCalculator_CalculateCoverage(t *testing.T) {
	tests := []struct {
		name             string
		agents           []*parser.AgentSpec
		expectedCoverage CoverageStats
	}{
		{
			name: "all fields complete",
			agents: []*parser.AgentSpec{
				{
					Name:        "complete",
					Description: "Complete agent",
					Tools:       []string{"Read"},
					Prompt:      "Complete prompt",
				},
			},
			expectedCoverage: CoverageStats{
				WithName:        1,
				WithDescription: 1,
				WithTools:       1,
				WithPrompt:      1,
				AverageCoverage: 100.0,
			},
		},
		{
			name: "partial fields",
			agents: []*parser.AgentSpec{
				{
					Name:        "partial",
					Description: "Has description",
					// No tools
					Prompt: "Has prompt",
				},
				{
					Name: "minimal",
					// No description, tools, or prompt
				},
			},
			expectedCoverage: CoverageStats{
				WithName:        2,
				WithDescription: 1,
				WithTools:       0,
				WithPrompt:      1,
				AverageCoverage: 66.67, // Average of 100% (3/3) and 33.33% (1/3)
			},
		},
		{
			name:   "empty agents",
			agents: []*parser.AgentSpec{},
			expectedCoverage: CoverageStats{
				WithName:        0,
				WithDescription: 0,
				WithTools:       0,
				WithPrompt:      0,
				AverageCoverage: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator(tt.agents)
			coverage := calc.calculateCoverage()

			assert.Equal(t, tt.expectedCoverage.WithName, coverage.WithName)
			assert.Equal(t, tt.expectedCoverage.WithDescription, coverage.WithDescription)
			assert.Equal(t, tt.expectedCoverage.WithTools, coverage.WithTools)
			assert.Equal(t, tt.expectedCoverage.WithPrompt, coverage.WithPrompt)
			assert.InDelta(t, tt.expectedCoverage.AverageCoverage, coverage.AverageCoverage, 0.1)
		})
	}
}

func TestCalculator_CalculateToolUsage(t *testing.T) {
	agents := []*parser.AgentSpec{
		{
			Name:           "explicit-single",
			Tools:          []string{"Read"},
			ToolsInherited: false,
		},
		{
			Name:           "explicit-multiple",
			Tools:          []string{"Read", "Write", "Bash"},
			ToolsInherited: false,
		},
		{
			Name:           "inherited1",
			ToolsInherited: true,
		},
		{
			Name:           "inherited2",
			ToolsInherited: true,
		},
		{
			Name:           "explicit-duplicate-tools",
			Tools:          []string{"Read", "Bash"}, // Read appears again
			ToolsInherited: false,
		},
	}

	calc := NewCalculator(agents)
	toolStats := calc.calculateToolUsage()

	assert.Equal(t, 2, toolStats.InheritedTools) // 2 agents with inherited tools
	assert.Equal(t, 3, toolStats.ExplicitTools)  // 3 agents with explicit tools

	expectedDist := map[string]int{
		"Read":  3, // appears in 3 agents (explicit-single, explicit-multiple, explicit-duplicate-tools)
		"Write": 1, // appears in 1 agent (explicit-multiple)
		"Bash":  2, // appears in 2 agents (explicit-multiple, explicit-duplicate-tools)
	}
	assert.Equal(t, expectedDist, toolStats.ToolDistribution)
}

func TestCalculator_WithOrphanedAgents(t *testing.T) {
	agents := []*parser.AgentSpec{
		{
			Name:        "valid-agent",
			Description: "Valid agent with all required fields",
			Prompt:      "You are valid",
		},
		{
			// Missing name - should be invalid
			Description: "Invalid agent missing name",
			Prompt:      "You are invalid",
		},
		{
			Name:        "invalid-name-format_with_underscores",
			Description: "Invalid name format",
			Prompt:      "You have invalid name",
		},
		{
			Name: "missing-description",
			// Missing description - should be invalid
			Prompt: "You are missing description",
		},
		{
			Name:        "missing-prompt",
			Description: "Missing prompt",
			// Missing prompt - should be invalid
		},
	}

	calc := NewCalculator(agents)
	stats := calc.Calculate()

	// Should have 4 orphaned agents (all except the first one)
	assert.Equal(t, 4, stats.OrphanedAgents)
}

func TestCalculator_EmptyAgents(t *testing.T) {
	calc := NewCalculator([]*parser.AgentSpec{})
	stats := calc.Calculate()

	assert.Equal(t, 0, stats.TotalAgents)
	assert.Equal(t, 0, len(stats.BySource))
	assert.Equal(t, 0, len(stats.Duplicates))
	assert.Equal(t, 0, stats.OrphanedAgents)
	assert.Equal(t, 0.0, stats.Coverage.AverageCoverage)
}

func TestCalculator_RealWorldScenario(t *testing.T) {
	// Test with realistic agent data
	agents := []*parser.AgentSpec{
		{
			Name:           "go-specialist",
			Description:    "Expert Go developer and architect",
			Tools:          []string{"Read", "Write", "Edit", "Bash"},
			ToolsInherited: false,
			Prompt:         "You are a master Go developer...",
			Source:         "github",
			FilePath:       "/Users/test/.claude/agents/go-specialist.md",
			ModTime:        time.Now().Add(-24 * time.Hour),
		},
		{
			Name:           "python-expert",
			Description:    "Python development specialist",
			Tools:          []string{"Read", "Write", "Bash"},
			ToolsInherited: false,
			Prompt:         "You are a Python expert...",
			Source:         "github",
			FilePath:       "/Users/test/.claude/agents/python-expert.md",
			ModTime:        time.Now().Add(-12 * time.Hour),
		},
		{
			Name:           "general-assistant",
			Description:    "General purpose assistant",
			ToolsInherited: true, // Uses inherited tools
			Prompt:         "You are a helpful assistant...",
			Source:         "local",
			FilePath:       "/Users/test/.claude/agents/general-assistant.md",
			ModTime:        time.Now(),
		},
		{
			Name:           "data-analyst",
			Description:    "Data analysis and visualization expert",
			Tools:          []string{"Read", "Write"}, // Fewer tools
			ToolsInherited: false,
			Prompt:         "You are a data analyst...",
			Source:         "git",
			FilePath:       "/Users/test/.claude/agents/data-analyst.md",
			ModTime:        time.Now().Add(-6 * time.Hour),
		},
	}

	calc := NewCalculator(agents)
	stats := calc.Calculate()

	// Verify realistic stats
	assert.Equal(t, 4, stats.TotalAgents)

	// All agents should have complete fields
	assert.Equal(t, 4, stats.Coverage.WithName)
	assert.Equal(t, 4, stats.Coverage.WithDescription)
	assert.Equal(t, 3, stats.Coverage.WithTools) // 3 have explicit tools
	assert.Equal(t, 4, stats.Coverage.WithPrompt)
	assert.Equal(t, 100.0, stats.Coverage.AverageCoverage) // All required fields present

	// Source distribution
	expectedSources := map[string]int{
		"github": 2,
		"local":  1,
		"git":    1,
	}
	assert.Equal(t, expectedSources, stats.BySource)

	// Tool usage
	assert.Equal(t, 1, stats.ToolUsage.InheritedTools) // general-assistant
	assert.Equal(t, 3, stats.ToolUsage.ExplicitTools)  // others

	// No duplicates or orphans in this clean dataset
	assert.Equal(t, 0, len(stats.Duplicates))
	assert.Equal(t, 0, stats.OrphanedAgents)

	// Verify tool distribution
	expectedTools := map[string]int{
		"Read":  3, // go-specialist, python-expert, data-analyst
		"Write": 3, // go-specialist, python-expert, data-analyst
		"Edit":  1, // go-specialist
		"Bash":  2, // go-specialist, python-expert
	}
	assert.Equal(t, expectedTools, stats.ToolUsage.ToolDistribution)
}
