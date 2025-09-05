package commands

import (
	"fmt"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/spf13/cobra"
)

// InstallCommand implements the install command functionality
type InstallCommand struct {
	*BaseCommand
	sourceName string
}

// NewInstallCommand creates a new install command instance
func NewInstallCommand() *InstallCommand {
	cmd := &InstallCommand{}
	cmd.BaseCommand = NewBaseCommand(cmd)
	return cmd
}

// Name returns the command name
func (c *InstallCommand) Name() string {
	return "install"
}

// Description returns the command description
func (c *InstallCommand) Description() string {
	return "Install agents from configured sources"
}

// CreateCommand creates the cobra command for install functionality
func (c *InstallCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: c.Description(),
		Long:  `Install agents from all enabled sources defined in the configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Execute(sharedCtx)
		},
	}

	cmd.Flags().StringVarP(&c.sourceName, "source", "s", "", "install specific source only")

	return cmd
}

// Execute runs the install command logic
func (c *InstallCommand) Execute(sharedCtx *SharedContext) error {
	return c.ExecuteWithCommonPattern(sharedCtx, c.sourceName)
}

// ExecuteOperation implements CommandExecutor interface for install operations
func (c *InstallCommand) ExecuteOperation(ctx *SharedContext, sources []config.Source) error {
	// Create installer
	inst, err := ctx.CreateInstaller()
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	// Execute install operation on each source
	for _, source := range sources {
		if err := inst.InstallSource(source); err != nil {
			return err
		}
	}
	return nil
}

// GetOperationName implements CommandExecutor interface
func (c *InstallCommand) GetOperationName() string {
	return "Installing"
}

// GetCompletionMessage implements CommandExecutor interface
func (c *InstallCommand) GetCompletionMessage() string {
	return "Installation complete"
}

// ShouldContinueOnError implements CommandExecutor interface
func (c *InstallCommand) ShouldContinueOnError(ctx *SharedContext) bool {
	return ctx.Config.Settings.ContinueOnError
}
