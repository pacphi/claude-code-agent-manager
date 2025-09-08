package commands

import (
	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/cli"
	"github.com/spf13/cobra"
)

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands   []Command
	sharedOpts *SharedOptions
	sharedCtx  *SharedContext
}

// NewCommandRegistry creates a new command registry with all available commands
func NewCommandRegistry() *CommandRegistry {
	sharedOpts := &SharedOptions{
		ConfigFile: "agents-config.yaml",
	}

	registry := &CommandRegistry{
		sharedOpts: sharedOpts,
		sharedCtx:  NewSharedContext(sharedOpts),
		commands: []Command{
			NewInstallCommand(),
			NewUninstallCommand(),
			NewUpdateCommand(),
			NewListCommand(),
			NewQueryCommand(),
			NewShowCommand(),
			NewStatsCommand(),
			NewValidateCommand(),
			NewIndexCommand(),
		},
	}

	return registry
}

// CreateRootCommand creates the root cobra command with all subcommands
func (r *CommandRegistry) CreateRootCommand(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "agent-manager",
		Short: "Manage Claude Code subagents via YAML configuration",
		Long: `Agent Manager is a tool for installing, updating, and managing
Claude Code subagents from various sources using YAML configuration.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			r.setupGlobalOptions()
		},
	}

	// Add persistent flags
	AddPersistentFlags(rootCmd, r.sharedOpts)

	// Add all registered commands
	for _, command := range r.commands {
		subCmd := command.CreateCommand(r.sharedCtx)
		rootCmd.AddCommand(subCmd)
	}

	// Add version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			color.Green("agent-manager version %s\n", version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Add marketplace command (external)
	rootCmd.AddCommand(cli.NewMarketplaceCmd())

	return rootCmd
}

// setupGlobalOptions configures global options before command execution
func (r *CommandRegistry) setupGlobalOptions() {
	// Setup colors
	SetupColors(r.sharedOpts.NoColor)

	// Setup progress manager
	SetupProgress(r.sharedOpts)
}
