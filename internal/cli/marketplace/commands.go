package marketplace

import (
	"fmt"

	"github.com/pacphi/claude-code-agent-manager/internal/cli/marketplace/display"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/service"
	"github.com/spf13/cobra"
)

// Commands handles marketplace CLI commands
type Commands struct {
	service service.MarketplaceService
	display *display.Formatter
}

// NewCommands creates a new marketplace commands handler
func NewCommands(svc service.MarketplaceService) *Commands {
	return &Commands{
		service: svc,
		display: display.NewFormatter(),
	}
}

// NewMarketplaceCmd creates the main marketplace command
func (c *Commands) NewMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Browse the subagents.sh marketplace",
		Long: `Discover and view details of Claude Code subagents from the subagents.sh marketplace.

Examples:
  agent-manager marketplace list                    # List all categories
  agent-manager marketplace list --category dev     # List agents in development category
  agent-manager marketplace show code-reviewer      # Show details for a specific agent
  agent-manager marketplace refresh                 # Refresh cached marketplace data`,
	}

	cmd.AddCommand(c.newListCmd())
	cmd.AddCommand(c.newShowCmd())
	cmd.AddCommand(c.newRefreshCmd())

	return cmd
}

// newListCmd creates the list command
func (c *Commands) newListCmd() *cobra.Command {
	var category string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List marketplace categories or agents",
		Long:  "List all available categories or agents within a specific category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if category == "" {
				return c.listCategories(cmd)
			}
			return c.listAgents(cmd, category, limit)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "filter by category")
	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "limit number of results")

	return cmd
}

// newShowCmd creates the show command
func (c *Commands) newShowCmd() *cobra.Command {
	var showContent bool

	cmd := &cobra.Command{
		Use:   "show <agent-id>",
		Short: "Show details for a specific agent",
		Long:  "Display detailed information about a marketplace agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]
			return c.showAgent(cmd, agentID, showContent)
		},
	}

	cmd.Flags().BoolVar(&showContent, "content", false, "show agent content/definition")

	return cmd
}

// newRefreshCmd creates the refresh command
func (c *Commands) newRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh cached marketplace data",
		Long:  "Clear the cache and fetch fresh data from the marketplace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.refreshCache(cmd)
		},
	}

	return cmd
}

// listCategories lists all marketplace categories
func (c *Commands) listCategories(cmd *cobra.Command) error {
	categories, err := c.service.GetCategories(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to fetch categories: %w", err)
	}

	c.display.PrintCategories(categories)
	return nil
}

// listAgents lists agents in a specific category
func (c *Commands) listAgents(cmd *cobra.Command, category string, limit int) error {
	agents, err := c.service.GetAgents(cmd.Context(), category)
	if err != nil {
		return fmt.Errorf("failed to fetch agents for category %s: %w", category, err)
	}

	if limit > 0 && limit < len(agents) {
		agents = agents[:limit]
	}

	c.display.PrintAgents(agents)
	return nil
}

// showAgent displays detailed information about an agent
func (c *Commands) showAgent(cmd *cobra.Command, agentID string, showContent bool) error {
	agent, err := c.service.GetAgent(cmd.Context(), agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent %s: %w", agentID, err)
	}

	var content string
	if showContent {
		content, err = c.service.GetAgentContent(cmd.Context(), agentID)
		if err != nil {
			return fmt.Errorf("failed to get agent content: %w", err)
		}
	}

	c.display.PrintAgentDetails(*agent, content)
	return nil
}

// refreshCache refreshes the marketplace cache
func (c *Commands) refreshCache(cmd *cobra.Command) error {
	err := c.service.RefreshCache(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}

	c.display.PrintSuccess("Cache refreshed successfully")
	return nil
}
