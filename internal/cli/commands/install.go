package commands

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// InstallCommand implements the install command functionality
type InstallCommand struct {
	sourceName string
}

// NewInstallCommand creates a new install command instance
func NewInstallCommand() *InstallCommand {
	return &InstallCommand{}
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
	// Load and validate configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create installer
	inst, err := sharedCtx.CreateInstaller()
	if err != nil {
		return fmt.Errorf("failed to create installer: %w", err)
	}

	// Get sources to install
	sources, err := sharedCtx.FilterEnabledSources(c.sourceName)
	if err != nil {
		return err
	}

	if len(sources) == 0 {
		if c.sourceName != "" {
			return fmt.Errorf("source '%s' is not enabled or not found", c.sourceName)
		}
		PrintWarning("No enabled sources found in configuration")
		return nil
	}

	// Install from each source
	successCount := 0
	failCount := 0

	for _, source := range sources {
		if c.shouldUseSpinner(sharedCtx) {
			// Use spinner for non-verbose mode
			err := sharedCtx.PM.WithSpinner(fmt.Sprintf("Installing %s", source.Name), func() error {
				return inst.InstallSource(source)
			})

			if err != nil {
				PrintError("Failed to install %s: %v", source.Name, err)
				failCount++
				if !sharedCtx.Config.Settings.ContinueOnError {
					return err
				}
			} else {
				successCount++
			}
		} else {
			// Verbose mode with detailed output
			color.Blue("Installing from source: %s\n", source.Name)

			if err := inst.InstallSource(source); err != nil {
				PrintError("Failed to install %s: %v", source.Name, err)
				failCount++
				if !sharedCtx.Config.Settings.ContinueOnError {
					return err
				}
			} else {
				PrintSuccess("Successfully installed %s", source.Name)
				successCount++
			}
		}
	}

	// Print summary
	c.printSummary(successCount, failCount)
	return nil
}

// shouldUseSpinner determines if spinner should be used based on options
func (c *InstallCommand) shouldUseSpinner(sharedCtx *SharedContext) bool {
	return !sharedCtx.Options.NoProgress && !sharedCtx.Options.Verbose
}

// printSummary prints the installation summary
func (c *InstallCommand) printSummary(successCount, failCount int) {
	fmt.Println()

	if successCount > 0 {
		PrintSuccess("Installation complete: %d succeeded", successCount)
	}

	if failCount > 0 {
		PrintError("%d failed", failCount)
	}

	if successCount == 0 && failCount == 0 {
		PrintInfo("No sources processed")
	}
}
