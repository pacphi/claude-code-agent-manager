package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/spf13/cobra"
)

// ShowCommand implements the show command functionality
type ShowCommand struct {
	agentName string
}

// NewShowCommand creates a new show command instance
func NewShowCommand() *ShowCommand {
	return &ShowCommand{}
}

// Name returns the command name
func (c *ShowCommand) Name() string {
	return "show"
}

// Description returns the command description
func (c *ShowCommand) Description() string {
	return "Show detailed information for a single agent"
}

// CreateCommand creates the cobra command for show functionality
func (c *ShowCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [AGENT_NAME]",
		Short: c.Description(),
		Long: `Show detailed information for a single agent with fuzzy matching support.

Examples:
  agent-manager show go-specialist        # Show agent by exact name
  agent-manager show go                   # Show agent by fuzzy name matching
  agent-manager show go-specialist.md     # Show agent by filename`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c.agentName = args[0]
			return c.Execute(sharedCtx)
		},
	}

	return cmd
}

// Execute runs the show command logic
func (c *ShowCommand) Execute(sharedCtx *SharedContext) error {
	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create query engine
	queryEngine, err := sharedCtx.CreateQueryEngine()
	if err != nil {
		return err
	}

	// Find agent with progress indication
	var agent *parser.AgentSpec
	err = sharedCtx.PM.WithSpinner(fmt.Sprintf("Finding agent '%s'", c.agentName), func() error {
		var showErr error
		agent, showErr = queryEngine.ShowAgent(c.agentName)
		return showErr
	})
	if err != nil {
		return fmt.Errorf("failed to find agent: %w", err)
	}

	// Display agent details
	c.displayAgentDetails(agent, sharedCtx)
	return nil
}

// displayAgentDetails displays comprehensive agent information
func (c *ShowCommand) displayAgentDetails(agent *parser.AgentSpec, sharedCtx *SharedContext) {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	color.Green("Agent Details\n")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("Name: %s\n", color.CyanString(agent.Name))
	fmt.Printf("File: %s\n", agent.FileName)
	fmt.Printf("Path: %s\n", agent.FilePath)

	if agent.Description != "" {
		fmt.Printf("Description: %s\n", agent.Description)
	}

	if agent.Source != "" {
		fmt.Printf("Source: %s\n", agent.Source)
	}

	if !agent.InstalledAt.IsZero() {
		fmt.Printf("Installed: %s\n", agent.InstalledAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("File Size: %d bytes\n", agent.FileSize)
	fmt.Printf("Modified: %s\n", agent.ModTime.Format("2006-01-02 15:04:05"))

	// Tools section
	fmt.Printf("\nTools: ")
	if agent.ToolsInherited {
		color.Yellow("inherited from parent\n")
	} else if len(agent.Tools) == 0 {
		color.Red("none specified\n")
	} else {
		fmt.Println()
		for _, tool := range agent.Tools {
			fmt.Printf("  - %s\n", tool)
		}
	}

	// Prompt section
	fmt.Printf("\nPrompt Preview:\n")
	fmt.Println(strings.Repeat("-", 50))
	c.displayPromptPreview(agent.Prompt)
}

// displayPromptPreview displays a preview of the agent prompt
func (c *ShowCommand) displayPromptPreview(prompt string) {
	promptLines := strings.Split(prompt, "\n")
	maxLines := 10

	if len(promptLines) > maxLines {
		for i := 0; i < maxLines; i++ {
			fmt.Println(promptLines[i])
		}
		fmt.Printf("... (%d more lines)\n", len(promptLines)-maxLines)
	} else {
		fmt.Println(prompt)
	}
}
