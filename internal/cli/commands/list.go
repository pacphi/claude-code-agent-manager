package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/pacphi/claude-code-agent-manager/internal/tracker"
	"github.com/spf13/cobra"
)

// ListCommand implements the list command functionality
type ListCommand struct {
	sourceName  string
	search      string
	name        string
	description string
	tools       []string
	noTools     bool
	customTools bool
	limit       int
}

// NewListCommand creates a new list command instance
func NewListCommand() *ListCommand {
	return &ListCommand{
		limit: 50,
	}
}

// Name returns the command name
func (c *ListCommand) Name() string {
	return "list"
}

// Description returns the command description
func (c *ListCommand) Description() string {
	return "List installed agents"
}

// CreateCommand creates the cobra command for list functionality
func (c *ListCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: c.Description(),
		Long:  `List all installed agents or filter by source.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Execute(sharedCtx)
		},
	}

	cmd.Flags().StringVarP(&c.sourceName, "source", "s", "", "list agents from specific source")
	cmd.Flags().StringVar(&c.search, "search", "", "search in agent names, descriptions, or content")
	cmd.Flags().StringVar(&c.name, "name", "", "search by agent name")
	cmd.Flags().StringVar(&c.description, "description", "", "search in description")
	cmd.Flags().StringSliceVar(&c.tools, "tools", nil, "filter by tools")
	cmd.Flags().BoolVar(&c.noTools, "no-tools", false, "show agents with inherited tools only")
	cmd.Flags().BoolVar(&c.customTools, "custom-tools", false, "show agents with explicit tools only")
	cmd.Flags().IntVar(&c.limit, "limit", 50, "limit number of results")

	return cmd
}

// Execute runs the list command logic
func (c *ListCommand) Execute(sharedCtx *SharedContext) error {
	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Check if any search parameters are provided
	hasSearchParams := c.search != "" || c.name != "" || c.description != "" ||
		len(c.tools) > 0 || c.noTools || c.customTools

	if hasSearchParams {
		// Use enhanced search with query engine
		return c.executeSearchList(sharedCtx)
	}

	// Original list functionality for backward compatibility
	return c.executeBasicList(sharedCtx)
}

// executeBasicList runs the original list functionality
func (c *ListCommand) executeBasicList(sharedCtx *SharedContext) error {
	var installations map[string]*tracker.Installation

	// Load tracking data
	err := sharedCtx.PM.WithSpinner("Loading installation data", func() error {
		track := tracker.New(sharedCtx.Config.Metadata.TrackingFile)
		var loadErr error
		installations, loadErr = track.List()
		return loadErr
	})
	if err != nil {
		return fmt.Errorf("failed to load installation data: %w", err)
	}

	if c.sourceName != "" {
		if inst, exists := installations[c.sourceName]; exists {
			c.printInstallation(c.sourceName, *inst)
		} else {
			PrintWarning("No installation found for source: %s", c.sourceName)
		}
		return nil
	}

	// List all installations
	if len(installations) == 0 {
		PrintWarning("No agents installed")
		return nil
	}

	for name, inst := range installations {
		c.printInstallation(name, *inst)
		fmt.Println()
	}

	return nil
}

// executeSearchList runs the enhanced search-based list functionality
func (c *ListCommand) executeSearchList(sharedCtx *SharedContext) error {
	// Initialize query engine
	cachePath := filepath.Join(sharedCtx.Config.Settings.BaseDir, ".agent-cache")
	indexPath := filepath.Join(sharedCtx.Config.Settings.BaseDir, ".agent-index")

	var queryEngine *engine.Engine
	err := sharedCtx.PM.WithSpinner("Initializing search engine", func() error {
		var engineErr error
		queryEngine, engineErr = engine.NewEngine(indexPath, cachePath)
		if engineErr != nil {
			return fmt.Errorf("failed to initialize query engine: %w", engineErr)
		}

		// Build index from tracking data
		track := tracker.New(sharedCtx.Config.Metadata.TrackingFile)
		agentData, err := track.GetAllAgentMetadata()
		if err != nil {
			return fmt.Errorf("failed to load agent metadata: %w", err)
		}

		// Convert to agent specs for querying
		agents := make([]*parser.AgentSpec, 0, len(agentData))
		for _, agentInfo := range agentData {
			// Filter by source if specified
			if c.sourceName != "" && agentInfo.Source != c.sourceName {
				continue
			}

			agentSpec := &parser.AgentSpec{
				Name:           agentInfo.Name,
				Description:    agentInfo.Description,
				Tools:          agentInfo.Tools,
				ToolsInherited: agentInfo.ToolsInherited,
				FilePath:       agentInfo.FilePath,
				FileName:       agentInfo.FileName,
				FileSize:       agentInfo.FileSize,
				ModTime:        agentInfo.ModTime,
				Source:         agentInfo.Source,
				InstalledAt:    agentInfo.InstalledAt,
			}
			agents = append(agents, agentSpec)
		}

		// Rebuild index with current agents
		return queryEngine.RebuildWithAgents(agents)
	})
	if err != nil {
		return err
	}

	// Execute search
	var results []*parser.AgentSpec
	err = sharedCtx.PM.WithSpinner("Searching agents", func() error {
		opts := engine.QueryOptions{
			Limit:       c.limit,
			NoTools:     c.noTools,
			CustomTools: c.customTools,
			Source:      c.sourceName,
		}

		var searchErr error

		// Execute appropriate search based on flags
		if c.search != "" {
			results, searchErr = queryEngine.Query(c.search, opts)
		} else if c.name != "" {
			results, searchErr = queryEngine.QueryByField("name", c.name)
		} else if c.description != "" {
			results, searchErr = queryEngine.QueryByField("description", c.description)
		} else if len(c.tools) > 0 {
			results, searchErr = queryEngine.QueryByField("tools", strings.Join(c.tools, ","))
		} else {
			// Just filter by options (no-tools, custom-tools, source)
			results, searchErr = queryEngine.Query("", opts)
		}

		return searchErr
	})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(results) == 0 {
		PrintWarning("No agents found matching search criteria")
		return nil
	}

	PrintSuccess("Found %d agents:", len(results))
	fmt.Println()

	for _, agent := range results {
		c.printAgentSummary(agent)
		fmt.Println()
	}

	return nil
}

// printInstallation prints installation details in the original format
func (c *ListCommand) printInstallation(name string, inst tracker.Installation) {
	color.Green("Source: %s\n", name)
	fmt.Printf("  Installed: %s\n", inst.Timestamp.Format("2006-01-02 15:04:05"))
	if inst.SourceCommit != "" {
		fmt.Printf("  Commit: %s\n", inst.SourceCommit)
	}
	fmt.Printf("  Files: %d\n", len(inst.Files))

	if len(inst.Directories) > 0 {
		fmt.Println("  Directories:")
		for _, dir := range inst.Directories {
			fmt.Printf("    - %s\n", dir)
		}
	}
	if len(inst.DocsGenerated) > 0 {
		fmt.Println("  Documentation:")
		for _, doc := range inst.DocsGenerated {
			fmt.Printf("    - %s\n", doc)
		}
	}
}

// printAgentSummary prints agent details in search result format
func (c *ListCommand) printAgentSummary(agent *parser.AgentSpec) {
	color.Cyan("● %s", agent.Name)
	fmt.Printf("  %s\n", agent.Description)
	fmt.Printf("  Source: %s | File: %s\n", agent.Source, agent.FileName)

	if !agent.ToolsInherited && len(agent.GetToolsAsSlice()) > 0 {
		fmt.Printf("  Tools: %s\n", strings.Join(agent.GetToolsAsSlice(), ", "))
	} else if agent.ToolsInherited {
		fmt.Printf("  Tools: inherited\n")
	}

	fmt.Printf("  Updated: %s\n", agent.ModTime.Format("2006-01-02"))
}
