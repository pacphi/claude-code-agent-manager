package commands

import (
	"fmt"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/spf13/cobra"
)

// UpdateCommand implements the update command functionality
type UpdateCommand struct {
	*BaseCommand
	sourceName string
	checkOnly  bool
}

// NewUpdateCommand creates a new update command instance
func NewUpdateCommand() *UpdateCommand {
	cmd := &UpdateCommand{}
	cmd.BaseCommand = NewBaseCommand(cmd)
	return cmd
}

// Name returns the command name
func (c *UpdateCommand) Name() string {
	return "update"
}

// Description returns the command description
func (c *UpdateCommand) Description() string {
	return "Update existing agent installations"
}

// CreateCommand creates the cobra command for update functionality
func (c *UpdateCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: c.Description(),
		Long:  `Update agents from their sources to get the latest versions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Execute(sharedCtx)
		},
	}

	cmd.Flags().StringVarP(&c.sourceName, "source", "s", "", "update specific source only")
	cmd.Flags().BoolVar(&c.checkOnly, "check-only", false, "check for updates without applying")

	return cmd
}

// Execute runs the update command logic
func (c *UpdateCommand) Execute(sharedCtx *SharedContext) error {
	return c.ExecuteWithCommonPattern(sharedCtx, c.sourceName)
}

// ExecuteOperation implements CommandExecutor interface for update operations
func (c *UpdateCommand) ExecuteOperation(ctx *SharedContext, sources []config.Source) error {
	// Create installer with check-only mode if requested
	var inst *installer.Installer
	var err error

	if c.checkOnly {
		// Create a modified installer with dry-run enabled
		inst, err = ctx.createInstallerWithOptions(installer.Options{
			Verbose:     ctx.Options.Verbose,
			DryRun:      true, // Force dry-run for check-only
			KeepBackups: false,
		})
	} else {
		inst, err = ctx.CreateInstaller()
	}

	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	// Execute update operation on each source
	for _, source := range sources {
		if err := inst.UpdateSource(source.Name); err != nil {
			return err
		}
	}
	return nil
}

// GetOperationName implements CommandExecutor interface
func (c *UpdateCommand) GetOperationName() string {
	if c.checkOnly {
		return "Checking updates for"
	}
	return "Updating"
}

// GetCompletionMessage implements CommandExecutor interface
func (c *UpdateCommand) GetCompletionMessage() string {
	if c.checkOnly {
		return "Check complete"
	}
	return "Update complete"
}

// ShouldContinueOnError implements CommandExecutor interface
func (c *UpdateCommand) ShouldContinueOnError(ctx *SharedContext) bool {
	return ctx.Config.Settings.ContinueOnError
}
