package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/pacphi/claude-code-agent-manager/internal/query/stats"
	"github.com/spf13/cobra"
)

// StatsCommand implements the stats command functionality
type StatsCommand struct {
	detailed   bool
	validation bool
	tools      bool
	toolsLimit int
}

// NewStatsCommand creates a new stats command instance
func NewStatsCommand() *StatsCommand {
	return &StatsCommand{
		toolsLimit: 10,
	}
}

// Name returns the command name
func (c *StatsCommand) Name() string {
	return "stats"
}

// Description returns the command description
func (c *StatsCommand) Description() string {
	return "Display aggregate statistics about agents"
}

// CreateCommand creates the cobra command for stats functionality
func (c *StatsCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: c.Description(),
		Long: `Display statistics about installed agents including coverage, tool usage, and validation metrics.

Examples:
  agent-manager stats                # Show basic statistics
  agent-manager stats --detailed     # Show detailed statistics by source
  agent-manager stats --validation   # Show validation report
  agent-manager stats --tools        # Show top tools usage`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Execute(sharedCtx)
		},
	}

	cmd.Flags().BoolVar(&c.detailed, "detailed", false, "show detailed statistics by source")
	cmd.Flags().BoolVar(&c.validation, "validation", false, "show validation report")
	cmd.Flags().BoolVar(&c.tools, "tools", false, "show top tools usage")
	cmd.Flags().IntVar(&c.toolsLimit, "tools-limit", 10, "limit number of tools shown")

	return cmd
}

// Execute runs the stats command logic
func (c *StatsCommand) Execute(sharedCtx *SharedContext) error {
	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create query engine and get agents
	queryEngine, err := sharedCtx.CreateQueryEngine()
	if err != nil {
		return err
	}

	// Get all agents for statistics
	var agents []*parser.AgentSpec
	var totalFiles int
	err = sharedCtx.PM.WithSpinner("Calculating statistics", func() error {
		agents = queryEngine.GetAllAgents()
		
		// Count all .md files to get true total
		agentsDir := sharedCtx.GetAgentsDirectory()
		filepath.Walk(agentsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if strings.HasSuffix(path, ".md") && !info.IsDir() {
				totalFiles++
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}

	if totalFiles == 0 && len(agents) == 0 {
		PrintWarning("No agents found for statistics")
		return nil
	}

	// Create stats calculator with total file count
	calculator := stats.NewCalculatorWithTotal(agents, totalFiles)

	// Display appropriate statistics based on flags
	if c.validation {
		c.displayValidationStats(calculator, sharedCtx)
	} else if c.tools {
		c.displayToolsStats(calculator, sharedCtx)
	} else if c.detailed {
		c.displayDetailedStats(calculator, sharedCtx)
	} else {
		c.displayBasicStats(calculator, sharedCtx)
	}

	return nil
}

// displayBasicStats shows basic agent statistics
func (c *StatsCommand) displayBasicStats(calculator *stats.Calculator, sharedCtx *SharedContext) {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	statistics := calculator.Calculate()
	validationReport := calculator.GetValidationReport()

	color.Blue("Agent Statistics\n")
	fmt.Println(strings.Repeat("=", 40))

	// Show total from validation report which includes unparseable files
	if totalAgents, ok := validationReport["total_agents"].(int); ok {
		fmt.Printf("Total Agents: %d\n", totalAgents)
	} else {
		fmt.Printf("Total Agents: %d\n", statistics.TotalAgents)
	}
	fmt.Printf("Average Coverage: %.1f%%\n", statistics.Coverage.AverageCoverage)
	fmt.Printf("Agents with Tools: %d\n", statistics.ToolUsage.ExplicitTools)
	fmt.Printf("Agents with Inherited Tools: %d\n", statistics.ToolUsage.InheritedTools)

	if len(statistics.BySource) > 0 {
		fmt.Printf("\nBy Source:\n")
		for source, count := range statistics.BySource {
			fmt.Printf("  %s: %d\n", source, count)
		}
	}

	if statistics.OrphanedAgents > 0 {
		PrintWarning("\nWarning: %d agents have validation issues", statistics.OrphanedAgents)
	}

	if len(statistics.Duplicates) > 0 {
		PrintWarning("\nWarning: %d duplicate agent names found", len(statistics.Duplicates))
	}
}

// displayDetailedStats shows detailed statistics including per-source breakdowns
func (c *StatsCommand) displayDetailedStats(calculator *stats.Calculator, sharedCtx *SharedContext) {
	c.displayBasicStats(calculator, sharedCtx)

	statistics := calculator.Calculate()

	fmt.Printf("\nDetailed Coverage:\n")
	fmt.Printf("  With Name: %d\n", statistics.Coverage.WithName)
	fmt.Printf("  With Description: %d\n", statistics.Coverage.WithDescription)
	fmt.Printf("  With Prompt: %d\n", statistics.Coverage.WithPrompt)
	fmt.Printf("  With Tools: %d\n", statistics.Coverage.WithTools)

	// Show source-specific stats
	sourceStats := calculator.CalculateSourceStats()
	if len(sourceStats) > 1 {
		fmt.Printf("\nPer-Source Statistics:\n")
		for source, stats := range sourceStats {
			fmt.Printf("  %s:\n", source)
			fmt.Printf("    Agents: %d\n", stats.TotalAgents)
			fmt.Printf("    Coverage: %.1f%%\n", stats.Coverage.AverageCoverage)
			fmt.Printf("    With Tools: %d\n", stats.ToolUsage.ExplicitTools)
		}
	}
}

// displayValidationStats shows validation report
func (c *StatsCommand) displayValidationStats(calculator *stats.Calculator, sharedCtx *SharedContext) {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	report := calculator.GetValidationReport()

	color.Blue("Validation Report\n")
	fmt.Println(strings.Repeat("=", 40))

	fmt.Printf("Total Agents: %d\n", report["total_agents"])
	fmt.Printf("Valid Agents: %d\n", report["valid_agents"])
	fmt.Printf("Invalid Agents: %d\n", report["invalid_agents"])
	fmt.Printf("Validation Rate: %.1f%%\n", report["validation_rate"])

	if errors, ok := report["common_errors"].(map[string]int); ok && len(errors) > 0 {
		fmt.Printf("\nCommon Errors:\n")
		for err, count := range errors {
			fmt.Printf("  %s: %d\n", err, count)
		}
	}

	if warnings, ok := report["common_warnings"].(map[string]int); ok && len(warnings) > 0 {
		fmt.Printf("\nCommon Warnings:\n")
		for warning, count := range warnings {
			fmt.Printf("  %s: %d\n", warning, count)
		}
	}
}

// displayToolsStats shows tools usage statistics
func (c *StatsCommand) displayToolsStats(calculator *stats.Calculator, sharedCtx *SharedContext) {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	topTools := calculator.GetTopTools(c.toolsLimit)
	statistics := calculator.Calculate()

	color.Blue("Tools Usage Statistics\n")
	fmt.Println(strings.Repeat("=", 40))

	fmt.Printf("Agents with Explicit Tools: %d\n", statistics.ToolUsage.ExplicitTools)
	fmt.Printf("Agents with Inherited Tools: %d\n", statistics.ToolUsage.InheritedTools)
	fmt.Printf("Total Unique Tools: %d\n", len(statistics.ToolUsage.ToolDistribution))

	if len(topTools) > 0 {
		fmt.Printf("\nTop Tools (max %d):\n", c.toolsLimit)
		for i, tool := range topTools {
			fmt.Printf("  %d. %s: %d agents\n", i+1, tool.Tool, tool.Count)
		}
	}
}
