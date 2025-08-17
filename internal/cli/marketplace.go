package cli

import (
	"github.com/pacphi/claude-code-agent-manager/internal/cli/marketplace"
	marketplaceService "github.com/pacphi/claude-code-agent-manager/internal/marketplace"
	"github.com/spf13/cobra"
)

// NewMarketplaceCmd creates the marketplace command using the new architecture
func NewMarketplaceCmd() *cobra.Command {
	// Create marketplace container with default configuration
	container, err := marketplaceService.WithDefaults()
	if err != nil {
		// Return a command that shows the error immediately
		return &cobra.Command{
			Use:   "marketplace",
			Short: "Browse and search the subagents.sh marketplace",
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.Printf("ERROR: Failed to initialize marketplace: %v\n", err)
				return err
			},
		}
	}

	// Create commands with dependency injection
	commands := marketplace.NewCommands(container.Service)

	// Return the marketplace command
	cmd := commands.NewMarketplaceCmd()

	// Add cleanup on command completion
	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if container != nil {
			_ = container.Close()
		}
	}

	return cmd
}
