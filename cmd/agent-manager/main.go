package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/conflict"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/pacphi/claude-code-agent-manager/internal/tracker"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	configFile string
	verbose    bool
	dryRun     bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:   "agent-manager",
	Short: "Manage Claude Code agents via YAML configuration",
	Long: `Agent Manager is a tool for installing, updating, and managing
Claude Code agents from various sources using YAML configuration.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install agents from configured sources",
	Long:  `Install agents from all enabled sources defined in the configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName, _ := cmd.Flags().GetString("source")
		return runInstall(sourceName)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove installed agents",
	Long:  `Uninstall agents that were previously installed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName, _ := cmd.Flags().GetString("source")
		all, _ := cmd.Flags().GetBool("all")
		keepBackups, _ := cmd.Flags().GetBool("keep-backups")

		if all && sourceName != "" {
			return fmt.Errorf("cannot specify both --all and --source")
		}
		if !all && sourceName == "" {
			return fmt.Errorf("must specify either --all or --source")
		}

		return runUninstall(sourceName, all, keepBackups)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update existing agent installations",
	Long:  `Update agents from their sources to get the latest versions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName, _ := cmd.Flags().GetString("source")
		checkOnly, _ := cmd.Flags().GetBool("check-only")
		return runUpdate(sourceName, checkOnly)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed agents",
	Long:  `List all installed agents or filter by source.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceName, _ := cmd.Flags().GetString("source")
		return runList(sourceName)
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Validate the YAML configuration file for correctness.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runValidate()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("agent-manager version %s\n", version)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "agents-config.yaml", "configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "simulate actions without making changes")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	installCmd.Flags().StringP("source", "s", "", "install specific source only")
	rootCmd.AddCommand(installCmd)

	uninstallCmd.Flags().StringP("source", "s", "", "uninstall specific source")
	uninstallCmd.Flags().BoolP("all", "a", false, "uninstall all sources")
	uninstallCmd.Flags().Bool("keep-backups", false, "keep backup files")
	rootCmd.AddCommand(uninstallCmd)

	updateCmd.Flags().StringP("source", "s", "", "update specific source only")
	updateCmd.Flags().Bool("check-only", false, "check for updates without applying")
	rootCmd.AddCommand(updateCmd)

	listCmd.Flags().StringP("source", "s", "", "list agents from specific source")
	rootCmd.AddCommand(listCmd)

	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runInstall(sourceName string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize components
	track := tracker.New(cfg.Metadata.TrackingFile)
	resolver := conflict.NewResolver(cfg.Settings.ConflictStrategy, cfg.Settings.BackupDir)
	inst := installer.New(cfg, track, resolver, installer.Options{
		Verbose: verbose,
		DryRun:  dryRun,
	})

	// Filter sources if specific one requested
	sources := cfg.Sources
	if sourceName != "" {
		found := false
		for _, s := range cfg.Sources {
			if s.Name == sourceName {
				sources = []config.Source{s}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("source '%s' not found in configuration", sourceName)
		}
	}

	// Install from each enabled source
	successCount := 0
	failCount := 0

	for _, source := range sources {
		if !source.Enabled {
			if verbose {
				color.Yellow("Skipping disabled source: %s\n", source.Name)
			}
			continue
		}

		color.Blue("Installing from source: %s\n", source.Name)

		if err := inst.InstallSource(source); err != nil {
			color.Red("Failed to install %s: %v\n", source.Name, err)
			failCount++
			if !cfg.Settings.ContinueOnError {
				return err
			}
		} else {
			color.Green("✓ Successfully installed %s\n", source.Name)
			successCount++
		}
	}

	// Summary
	fmt.Println()
	color.Green("Installation complete: %d succeeded", successCount)
	if failCount > 0 {
		color.Red("%d failed", failCount)
	}

	return nil
}

func runUninstall(sourceName string, all bool, keepBackups bool) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	track := tracker.New(cfg.Metadata.TrackingFile)
	resolver := conflict.NewResolver(cfg.Settings.ConflictStrategy, cfg.Settings.BackupDir)
	inst := installer.New(cfg, track, resolver, installer.Options{
		Verbose:     verbose,
		DryRun:      dryRun,
		KeepBackups: keepBackups,
	})

	if all {
		color.Yellow("Uninstalling all sources...\n")
		return inst.UninstallAll()
	}

	color.Yellow("Uninstalling source: %s\n", sourceName)
	return inst.UninstallSource(sourceName)
}

func runUpdate(sourceName string, checkOnly bool) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	track := tracker.New(cfg.Metadata.TrackingFile)
	resolver := conflict.NewResolver(cfg.Settings.ConflictStrategy, cfg.Settings.BackupDir)
	inst := installer.New(cfg, track, resolver, installer.Options{
		Verbose: verbose,
		DryRun:  dryRun || checkOnly,
	})

	if sourceName != "" {
		return inst.UpdateSource(sourceName)
	}

	// Update all sources
	for _, source := range cfg.Sources {
		if !source.Enabled {
			continue
		}

		color.Blue("Checking updates for: %s\n", source.Name)
		if err := inst.UpdateSource(source.Name); err != nil {
			color.Red("Failed to update %s: %v\n", source.Name, err)
			if !cfg.Settings.ContinueOnError {
				return err
			}
		}
	}

	return nil
}

func runList(sourceName string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	track := tracker.New(cfg.Metadata.TrackingFile)
	installations, err := track.List()
	if err != nil {
		return fmt.Errorf("failed to list installations: %w", err)
	}

	if sourceName != "" {
		if inst, exists := installations[sourceName]; exists {
			printInstallation(sourceName, *inst)
		} else {
			color.Yellow("No installation found for source: %s\n", sourceName)
		}
		return nil
	}

	// List all installations
	if len(installations) == 0 {
		color.Yellow("No agents installed\n")
		return nil
	}

	for name, inst := range installations {
		printInstallation(name, *inst)
		fmt.Println()
	}

	return nil
}

func printInstallation(name string, inst tracker.Installation) {
	color.Green("Source: %s\n", name)
	fmt.Printf("  Installed: %s\n", inst.Timestamp.Format("2006-01-02 15:04:05"))
	if inst.SourceCommit != "" {
		fmt.Printf("  Commit: %s\n", inst.SourceCommit)
	}
	fmt.Printf("  Files: %d\n", len(inst.Files))

	if verbose {
		fmt.Println("  Directories:")
		for _, dir := range inst.Directories {
			fmt.Printf("    - %s\n", dir)
		}
		if len(inst.DocsGenerated) > 0 {
			fmt.Println("  Documentation:")
			for _, doc := range inst.DocsGenerated {
				fmt.Printf("    - %s\n", doc)
			}
		}
	}
}

func runValidate() error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.Validate(cfg); err != nil {
		color.Red("✗ Configuration is invalid:\n")
		return err
	}

	color.Green("✓ Configuration is valid\n")

	if verbose {
		fmt.Printf("\nConfiguration summary:\n")
		fmt.Printf("  Version: %s\n", cfg.Version)
		fmt.Printf("  Base directory: %s\n", cfg.Settings.BaseDir)
		fmt.Printf("  Sources: %d\n", len(cfg.Sources))

		for _, source := range cfg.Sources {
			status := "enabled"
			if !source.Enabled {
				status = "disabled"
			}
			fmt.Printf("    - %s (%s, %s)\n", source.Name, source.Type, status)
		}
	}

	// Check for potential issues
	warnings := []string{}

	// Check if directories exist
	if _, err := os.Stat(cfg.Settings.BaseDir); os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("Base directory does not exist: %s", cfg.Settings.BaseDir))
	}

	// Check for duplicate source names
	seen := make(map[string]bool)
	for _, source := range cfg.Sources {
		if seen[source.Name] {
			warnings = append(warnings, fmt.Sprintf("Duplicate source name: %s", source.Name))
		}
		seen[source.Name] = true
	}

	if len(warnings) > 0 {
		color.Yellow("\nWarnings:\n")
		for _, w := range warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	return nil
}

// Helper function to resolve path
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// Expand ~ to home directory
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}

	// Make relative to current directory
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, path)
}
