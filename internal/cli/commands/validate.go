package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/spf13/cobra"
)

// ValidateCommand implements the validate command functionality
type ValidateCommand struct {
	agents bool
	query  bool
}

// NewValidateCommand creates a new validate command instance
func NewValidateCommand() *ValidateCommand {
	return &ValidateCommand{}
}

// Name returns the command name
func (c *ValidateCommand) Name() string {
	return "validate"
}

// Description returns the command description
func (c *ValidateCommand) Description() string {
	return "Validate configuration file and agents"
}

// CreateCommand creates the cobra command for validate functionality
func (c *ValidateCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: c.Description(),
		Long: `Validate the YAML configuration file for correctness and optionally validate installed agents.

Examples:
  agent-manager validate             # Validate configuration only
  agent-manager validate --agents    # Also validate installed agents
  agent-manager validate --query     # Test query functionality`,
		SilenceUsage:  true,  // Don't show usage on error
		SilenceErrors: true,  // Don't print errors (we handle them ourselves)
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Execute(sharedCtx)
		},
	}

	cmd.Flags().BoolVar(&c.agents, "agents", false, "also validate installed agents")
	cmd.Flags().BoolVar(&c.query, "query", false, "test query functionality")

	return cmd
}

// Execute runs the validate command logic
func (c *ValidateCommand) Execute(sharedCtx *SharedContext) error {
	var cfg *config.Config
	var validationErr error

	// Load and validate configuration with progress
	err := sharedCtx.PM.WithSpinner("Validating configuration", func() error {
		var loadErr error
		cfg, loadErr = config.Load(sharedCtx.Options.ConfigFile)
		if loadErr != nil {
			return loadErr
		}

		validationErr = config.Validate(cfg)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Store config in shared context for later use
	sharedCtx.Config = cfg

	if validationErr != nil {
		PrintError("Configuration is invalid:")
		return validationErr
	}

	PrintSuccess("Configuration is valid")

	// Show configuration summary in verbose mode
	if sharedCtx.Options.Verbose {
		c.showConfigSummary(cfg)
	}

	// Check for potential issues
	c.checkForWarnings(cfg)

	// Enhanced validation: check agents if requested
	if c.agents {
		fmt.Println()
		if err := c.validateInstalledAgents(sharedCtx); err != nil {
			// Error already printed with details in validateInstalledAgents
			return err
		}
		PrintSuccess("All installed agents are valid")
	}

	// Test query functionality if requested
	if c.query {
		fmt.Println()
		if err := c.testQueryFunctionality(sharedCtx); err != nil {
			PrintError("Query functionality test failed: %v", err)
			return err
		}
		PrintSuccess("Query functionality is working")
	}

	return nil
}

// showConfigSummary displays a summary of the configuration
func (c *ValidateCommand) showConfigSummary(cfg *config.Config) {
	fmt.Printf("\nConfiguration summary:\n")
	fmt.Printf("  Version: %s\n", cfg.Version)
	fmt.Printf("  Base directory: %s\n", cfg.Settings.BaseDir)
	fmt.Printf("  Sources: %d\n", len(cfg.Sources))

	for _, source := range cfg.Sources {
		status := "enabled"
		if !source.Enabled {
			status = "disabled"
		}
		fmt.Printf("    - %s (%s, %s)\n", source.Name, source.Type, status)
	}
}

// checkForWarnings checks for potential configuration issues
func (c *ValidateCommand) checkForWarnings(cfg *config.Config) {
	warnings := []string{}

	// Check if directories exist
	if _, err := os.Stat(cfg.Settings.BaseDir); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("Base directory does not exist: %s", cfg.Settings.BaseDir))
	}

	// Check for duplicate source names
	seen := make(map[string]bool)
	for _, source := range cfg.Sources {
		if seen[source.Name] {
			warnings = append(warnings, fmt.Sprintf("Duplicate source name: %s", source.Name))
		}
		seen[source.Name] = true
	}

	// Check for sources with no enabled state
	enabledCount := 0
	for _, source := range cfg.Sources {
		if source.Enabled {
			enabledCount++
		}
	}
	if enabledCount == 0 {
		warnings = append(warnings, "No sources are enabled")
	}

	if len(warnings) > 0 {
		fmt.Println()
		PrintWarning("Configuration warnings:")
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}
}

// validateInstalledAgents validates all installed agent files
func (c *ValidateCommand) validateInstalledAgents(sharedCtx *SharedContext) error {
	agentsDir := sharedCtx.GetAgentsDirectory()
	
	// Count all .md files first to get total
	totalFiles := 0
	err := filepath.Walk(agentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		if strings.HasSuffix(path, ".md") && !info.IsDir() {
			totalFiles++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk agents directory: %w", err)
	}

	if totalFiles == 0 {
		PrintWarning("No agent files found to validate")
		return nil
	}

	// Parse agents with warnings enabled to detect parsing errors
	parserWithWarnings := parser.NewParserWithOptions(false) // Show warnings
	parsedAgents, _ := parserWithWarnings.ParseDirectory(agentsDir)
	
	// Track statistics
	validCount := 0
	invalidCount := 0
	parseFailureCount := totalFiles - len(parsedAgents) // Files that failed to parse
	warningCount := 0
	
	// Validate successfully parsed agents
	for _, agent := range parsedAgents {
		isValid := true
		
		// Check for missing required fields
		if agent.Name == "" {
			PrintError("Agent at %s is missing name", agent.FilePath)
			isValid = false
		}

		// Check if file exists (shouldn't happen for parsed agents, but double-check)
		if _, err := os.Stat(agent.FilePath); os.IsNotExist(err) {
			PrintError("Agent file does not exist: %s", agent.FilePath)
			isValid = false
		}

		// Check if prompt is reasonable length
		if len(agent.Prompt) < 10 {
			PrintWarning("Agent %s has very short prompt", agent.Name)
			warningCount++
		}

		// Check if description is present
		if agent.Description == "" {
			PrintWarning("Agent %s has no description", agent.Name)
			warningCount++
		}
		
		if isValid {
			validCount++
		} else {
			invalidCount++
		}
	}
	
	// Add parse failures to invalid count
	invalidCount += parseFailureCount
	
	// Display summary
	fmt.Println()
	color.Blue("Agent Validation Summary")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("Total agent files: %d\n", totalFiles)
	color.Green("✓ Valid agents: %d\n", validCount)
	if invalidCount > 0 {
		color.Red("✗ Invalid agents: %d\n", invalidCount)
		if parseFailureCount > 0 {
			color.Red("  - Failed to parse: %d\n", parseFailureCount)
		}
	}
	if warningCount > 0 {
		color.Yellow("⚠ Warnings: %d\n", warningCount)
	}

	if invalidCount > 0 {
		// Exit with error code but don't return an error (to avoid duplicate message)
		os.Exit(1)
	}

	return nil
}

// testQueryFunctionality tests basic query operations
func (c *ValidateCommand) testQueryFunctionality(sharedCtx *SharedContext) error {
	queryEngine, err := sharedCtx.CreateQueryEngine()
	if err != nil {
		return fmt.Errorf("failed to create query engine: %w", err)
	}

	// Test 1: Get all agents
	agents := queryEngine.GetAllAgents()
	if sharedCtx.Options.Verbose {
		color.Blue("Test 1: Retrieved %d agents\n", len(agents))
	}

	if len(agents) == 0 {
		PrintWarning("No agents found for query testing")
		return nil
	}

	// Test 2: Simple query with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := queryEngine.Query("", engine.QueryOptions{
		Limit:   5,
		Context: ctx,
	})
	if err != nil {
		return fmt.Errorf("basic query failed: %w", err)
	}
	if sharedCtx.Options.Verbose {
		color.Blue("Test 2: Basic query returned %d results\n", len(results))
	}

	// Test 3: Field-based query with first agent's name
	if len(agents) > 0 {
		firstAgentName := agents[0].Name
		nameResults, err := queryEngine.QueryByField("name", firstAgentName)
		if err != nil {
			return fmt.Errorf("field query failed: %w", err)
		}
		if sharedCtx.Options.Verbose {
			color.Blue("Test 3: Name query for '%s' returned %d results\n", firstAgentName, len(nameResults))
		}
	}

	// Test 4: Show agent functionality
	if len(agents) > 0 {
		firstAgent, err := queryEngine.ShowAgent(agents[0].Name)
		if err != nil {
			return fmt.Errorf("show agent failed: %w", err)
		}
		if firstAgent == nil {
			return fmt.Errorf("show agent returned nil")
		}
		if sharedCtx.Options.Verbose {
			color.Blue("Test 4: ShowAgent successfully found '%s'\n", firstAgent.Name)
		}
	}

	// Test 5: Cache functionality
	cacheStats := queryEngine.GetCacheStats()
	if sharedCtx.Options.Verbose {
		color.Blue("Test 5: Cache stats retrieved: %v\n", cacheStats)
	}

	return nil
}
