package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/spf13/cobra"
)

// UninstallCommand implements the uninstall command functionality
type UninstallCommand struct {
	sourceName  string
	all         bool
	keepBackups bool
}

// NewUninstallCommand creates a new uninstall command instance
func NewUninstallCommand() *UninstallCommand {
	return &UninstallCommand{}
}

// Name returns the command name
func (c *UninstallCommand) Name() string {
	return "uninstall"
}

// Description returns the command description
func (c *UninstallCommand) Description() string {
	return "Remove installed agents"
}

// CreateCommand creates the cobra command for uninstall functionality
func (c *UninstallCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: c.Description(),
		Long:  `Uninstall agents that were previously installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Execute(sharedCtx)
		},
	}

	cmd.Flags().StringVarP(&c.sourceName, "source", "s", "", "uninstall specific source")
	cmd.Flags().BoolVarP(&c.all, "all", "a", false, "uninstall all sources")
	cmd.Flags().BoolVar(&c.keepBackups, "keep-backups", false, "keep backup files")

	return cmd
}

// Execute runs the uninstall command logic
func (c *UninstallCommand) Execute(sharedCtx *SharedContext) error {
	// Validate flags
	if c.all && c.sourceName != "" {
		return fmt.Errorf("cannot specify both --all and --source")
	}
	if !c.all && c.sourceName == "" {
		return fmt.Errorf("must specify either --all or --source")
	}

	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create installer with keep-backups option
	inst, err := sharedCtx.createInstallerWithOptions(installer.Options{
		Verbose:     sharedCtx.Options.Verbose,
		DryRun:      sharedCtx.Options.DryRun,
		KeepBackups: c.keepBackups,
	})
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	if c.all {
		return c.uninstallAll(sharedCtx, inst)
	}

	return c.uninstallSource(sharedCtx, inst)
}

// uninstallAll removes all installed sources
func (c *UninstallCommand) uninstallAll(sharedCtx *SharedContext, inst *installer.Installer) error {
	if c.shouldUseSpinner(sharedCtx) {
		return sharedCtx.PM.WithSpinner("Uninstalling all sources", func() error {
			return inst.UninstallAll()
		})
	}

	PrintWarning("Uninstalling all sources...")
	err := inst.UninstallAll()
	if err != nil {
		PrintError("Failed to uninstall all sources: %v", err)
		return err
	}

	PrintSuccess("All sources uninstalled successfully")
	return nil
}

// uninstallSource removes a specific source
func (c *UninstallCommand) uninstallSource(sharedCtx *SharedContext, inst *installer.Installer) error {
	if c.shouldUseSpinner(sharedCtx) {
		return sharedCtx.PM.WithSpinner(fmt.Sprintf("Uninstalling %s", c.sourceName), func() error {
			return inst.UninstallSource(c.sourceName)
		})
	}

	color.Yellow("Uninstalling source: %s\n", c.sourceName)
	err := inst.UninstallSource(c.sourceName)
	if err != nil {
		PrintError("Failed to uninstall %s: %v", c.sourceName, err)
		return err
	}

	PrintSuccess("Successfully uninstalled %s", c.sourceName)
	return nil
}

// shouldUseSpinner determines if spinner should be used based on options
func (c *UninstallCommand) shouldUseSpinner(sharedCtx *SharedContext) bool {
	return !sharedCtx.Options.NoProgress && !sharedCtx.Options.Verbose
}
