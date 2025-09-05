package stats

import (
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/pacphi/claude-code-agent-manager/internal/query/validator"
)

// Calculator computes agent statistics
type Calculator struct {
	agents []*parser.AgentSpec
}

// NewCalculator creates a new stats calculator
func NewCalculator(agents []*parser.AgentSpec) *Calculator {
	return &Calculator{agents: agents}
}

// Statistics holds computed statistics
type Statistics struct {
	TotalAgents    int                 `json:"total_agents"`
	BySource       map[string]int      `json:"by_source"`
	Coverage       CoverageStats       `json:"coverage"`
	ToolUsage      ToolStats           `json:"tool_usage"`
	Duplicates     map[string][]string `json:"duplicates"`
	OrphanedAgents int                 `json:"orphaned_agents"`
}

// CoverageStats shows field coverage statistics
type CoverageStats struct {
	WithName        int     `json:"with_name"`
	WithDescription int     `json:"with_description"`
	WithTools       int     `json:"with_tools"`
	WithPrompt      int     `json:"with_prompt"`
	AverageCoverage float64 `json:"average_coverage"`
}

// ToolStats shows tool usage statistics
type ToolStats struct {
	InheritedTools   int            `json:"inherited_tools"`
	ExplicitTools    int            `json:"explicit_tools"`
	ToolDistribution map[string]int `json:"tool_distribution"`
}

// Calculate computes all statistics for the agent collection
func (c *Calculator) Calculate() *Statistics {
	stats := &Statistics{
		TotalAgents: len(c.agents),
		BySource:    make(map[string]int),
		Duplicates:  make(map[string][]string),
	}

	// Count by source
	for _, agent := range c.agents {
		if agent.Source != "" {
			stats.BySource[agent.Source]++
		}
	}

	// Calculate coverage metrics
	stats.Coverage = c.calculateCoverage()

	// Calculate tool usage statistics
	stats.ToolUsage = c.calculateToolUsage()

	// Find duplicate agent names
	nameMap := make(map[string][]string)
	for _, agent := range c.agents {
		nameMap[agent.Name] = append(nameMap[agent.Name], agent.FilePath)
	}
	for name, files := range nameMap {
		if len(files) > 1 {
			stats.Duplicates[name] = files
		}
	}

	// Count orphaned (invalid) agents
	validator := validator.NewValidator()
	for _, agent := range c.agents {
		if err := validator.Validate(agent); err != nil {
			stats.OrphanedAgents++
		}
	}

	return stats
}

// calculateCoverage computes field coverage metrics
func (c *Calculator) calculateCoverage() CoverageStats {
	coverage := CoverageStats{}
	totalScore := 0.0

	for _, agent := range c.agents {
		fieldsPresent := 0

		// Count required fields that are present
		if agent.Name != "" {
			coverage.WithName++
			fieldsPresent++
		}

		if agent.Description != "" {
			coverage.WithDescription++
			fieldsPresent++
		}

		if len(agent.GetToolsAsSlice()) > 0 {
			coverage.WithTools++
		}

		if agent.Prompt != "" {
			coverage.WithPrompt++
			fieldsPresent++
		}

		// Calculate coverage percentage for this agent (based on 3 required fields)
		// Tools are optional, so not counted in coverage calculation
		agentCoverage := float64(fieldsPresent) / 3.0 * 100
		totalScore += agentCoverage
	}

	// Calculate average coverage across all agents
	if len(c.agents) > 0 {
		coverage.AverageCoverage = totalScore / float64(len(c.agents))
	}

	return coverage
}

// calculateToolUsage computes tool usage statistics
func (c *Calculator) calculateToolUsage() ToolStats {
	stats := ToolStats{
		ToolDistribution: make(map[string]int),
	}

	for _, agent := range c.agents {
		if agent.ToolsInherited {
			stats.InheritedTools++
		} else {
			stats.ExplicitTools++
			// Count tool usage
			for _, tool := range agent.GetToolsAsSlice() {
				stats.ToolDistribution[tool]++
			}
		}
	}

	return stats
}

// CalculateSourceStats calculates statistics grouped by source
func (c *Calculator) CalculateSourceStats() map[string]*Statistics {
	sourceGroups := make(map[string][]*parser.AgentSpec)

	// Group agents by source
	for _, agent := range c.agents {
		source := agent.Source
		if source == "" {
			source = "unknown"
		}
		sourceGroups[source] = append(sourceGroups[source], agent)
	}

	// Calculate stats for each source
	result := make(map[string]*Statistics)
	for source, agents := range sourceGroups {
		calc := NewCalculator(agents)
		result[source] = calc.Calculate()
	}

	return result
}

// GetTopTools returns the most commonly used tools in descending order
func (c *Calculator) GetTopTools(limit int) []struct {
	Tool  string `json:"tool"`
	Count int    `json:"count"`
} {
	toolStats := c.calculateToolUsage()

	// Convert map to slice for sorting
	type toolCount struct {
		Tool  string
		Count int
	}

	tools := make([]toolCount, 0, len(toolStats.ToolDistribution))
	for tool, count := range toolStats.ToolDistribution {
		tools = append(tools, toolCount{Tool: tool, Count: count})
	}

	// Sort by count (descending)
	for i := 0; i < len(tools)-1; i++ {
		for j := i + 1; j < len(tools); j++ {
			if tools[j].Count > tools[i].Count {
				tools[i], tools[j] = tools[j], tools[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(tools) > limit {
		tools = tools[:limit]
	}

	// Convert to return format
	result := make([]struct {
		Tool  string `json:"tool"`
		Count int    `json:"count"`
	}, len(tools))

	for i, tc := range tools {
		result[i].Tool = tc.Tool
		result[i].Count = tc.Count
	}

	return result
}

// GetValidationReport provides detailed validation results
func (c *Calculator) GetValidationReport() map[string]interface{} {
	validator := validator.NewValidator()
	validCount := 0
	invalidCount := 0
	errors := make(map[string]int)
	warnings := make(map[string]int)

	for _, agent := range c.agents {
		if err := validator.Validate(agent); err != nil {
			invalidCount++
			errors[err.Error()]++
		} else {
			validCount++
		}

		// Get detailed report for warnings
		report := validator.ValidateWithReport(agent)
		for _, warning := range report.Warnings {
			warnings[warning]++
		}
	}

	return map[string]interface{}{
		"total_agents":   len(c.agents),
		"valid_agents":   validCount,
		"invalid_agents": invalidCount,
		"validation_rate": func() float64 {
			if len(c.agents) == 0 {
				return 0.0
			}
			return float64(validCount) / float64(len(c.agents)) * 100
		}(),
		"common_errors":   errors,
		"common_warnings": warnings,
	}
}
