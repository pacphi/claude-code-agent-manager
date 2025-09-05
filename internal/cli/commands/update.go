package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/spf13/cobra"
)

// UpdateCommand implements the update command functionality
type UpdateCommand struct {
	sourceName string
	checkOnly  bool
}

// NewUpdateCommand creates a new update command instance
func NewUpdateCommand() *UpdateCommand {
	return &UpdateCommand{}
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
	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create installer with check-only mode if requested
	inst, err := sharedCtx.CreateInstaller()
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	// Enable dry-run if check-only mode is requested
	if c.checkOnly {
		// Create a modified installer with dry-run enabled
		inst, err = sharedCtx.createInstallerWithOptions(installer.Options{
			Verbose:     sharedCtx.Options.Verbose,
			DryRun:      true, // Force dry-run for check-only
			KeepBackups: false,
		})
		if err != nil {
			return fmt.Errorf("failed to create installer: %w", err)
		}
	}

	if c.sourceName != "" {
		return c.updateSingleSource(sharedCtx, inst)
	}

	return c.updateAllSources(sharedCtx, inst)
}

// updateSingleSource updates a specific source
func (c *UpdateCommand) updateSingleSource(sharedCtx *SharedContext, inst *installer.Installer) error {
	action := "Updating"
	if c.checkOnly {
		action = "Checking updates for"
	}

	if c.shouldUseSpinner(sharedCtx) {
		return sharedCtx.PM.WithSpinner(fmt.Sprintf("%s %s", action, c.sourceName), func() error {
			return inst.UpdateSource(c.sourceName)
		})
	}

	color.Blue("%s source: %s\n", action, c.sourceName)
	err := inst.UpdateSource(c.sourceName)
	if err != nil {
		PrintError("Failed to update %s: %v", c.sourceName, err)
		return err
	}

	PrintSuccess("Successfully updated %s", c.sourceName)
	return nil
}

// updateAllSources updates all enabled sources
func (c *UpdateCommand) updateAllSources(sharedCtx *SharedContext, inst *installer.Installer) error {
	// Get enabled sources
	sources, err := sharedCtx.FilterEnabledSources("")
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		PrintWarning("No enabled sources found in configuration")
		return nil
	}

	successCount := 0
	failCount := 0
	action := "Updating"
	if c.checkOnly {
		action = "Checking updates for"
	}

	for _, source := range sources {
		if c.shouldUseSpinner(sharedCtx) {
			err := sharedCtx.PM.WithSpinner(fmt.Sprintf("%s %s", action, source.Name), func() error {
				return inst.UpdateSource(source.Name)
			})
			if err != nil {
				PrintError("Failed to update %s: %v", source.Name, err)
				failCount++
				if !sharedCtx.Config.Settings.ContinueOnError {
					return err
				}
			} else {
				successCount++
			}
		} else {
			color.Blue("%s: %s\n", action, source.Name)
			if err := inst.UpdateSource(source.Name); err != nil {
				PrintError("Failed to update %s: %v", source.Name, err)
				failCount++
				if !sharedCtx.Config.Settings.ContinueOnError {
					return err
				}
			} else {
				PrintSuccess("Successfully updated %s", source.Name)
				successCount++
			}
		}
	}

	// Print summary
	c.printSummary(successCount, failCount)
	return nil
}

// shouldUseSpinner determines if spinner should be used based on options
func (c *UpdateCommand) shouldUseSpinner(sharedCtx *SharedContext) bool {
	return !sharedCtx.Options.NoProgress && !sharedCtx.Options.Verbose
}

// printSummary prints the update summary
func (c *UpdateCommand) printSummary(successCount, failCount int) {
	fmt.Println()

	action := "Update"
	if c.checkOnly {
		action = "Check"
	}

	if successCount > 0 {
		PrintSuccess("%s complete: %d succeeded", action, successCount)
	}

	if failCount > 0 {
		PrintError("%d failed", failCount)
	}

	if successCount == 0 && failCount == 0 {
		PrintInfo("No sources processed")
	}
}
