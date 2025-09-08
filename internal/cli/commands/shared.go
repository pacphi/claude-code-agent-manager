package commands

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/conflict"
	"github.com/pacphi/claude-code-agent-manager/internal/installer"
	"github.com/pacphi/claude-code-agent-manager/internal/progress"
	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/pacphi/claude-code-agent-manager/internal/tracker"
	"github.com/spf13/cobra"
)

// SharedOptions holds common configuration options used across commands
type SharedOptions struct {
	ConfigFile string
	Verbose    bool
	DryRun     bool
	NoColor    bool
	NoProgress bool
}

// SharedContext provides shared dependencies and helpers for commands
type SharedContext struct {
	Options *SharedOptions
	Config  *config.Config
	PM      *progress.Manager
}

// NewSharedContext creates a new shared context for commands
func NewSharedContext(opts *SharedOptions) *SharedContext {
	return &SharedContext{
		Options: opts,
		PM:      progress.Default(),
	}
}

// LoadConfig loads and validates the configuration file with progress indication
func (sc *SharedContext) LoadConfig() error {
	return sc.PM.WithSpinner("Loading configuration", func() error {
		var err error
		sc.Config, err = config.Load(sc.Options.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return config.Validate(sc.Config)
	})
}

// CreateInstaller creates a new installer with the current configuration and options
func (sc *SharedContext) CreateInstaller() (*installer.Installer, error) {
	return sc.createInstallerWithOptions(installer.Options{
		Verbose:     sc.Options.Verbose,
		DryRun:      sc.Options.DryRun,
		KeepBackups: false, // Can be overridden per command
	})
}

// createInstallerWithOptions creates an installer with specific options
func (sc *SharedContext) createInstallerWithOptions(opts installer.Options) (*installer.Installer, error) {
	if sc.Config == nil {
		return nil, fmt.Errorf("configuration not loaded - call LoadConfig() first")
	}

	track := tracker.New(sc.Config.Metadata.TrackingFile)
	resolver := conflict.NewResolver(sc.Config.Settings.ConflictStrategy, sc.Config.Settings.BackupDir)

	return installer.New(sc.Config, track, resolver, opts), nil
}

// CreateQueryEngine creates and initializes a query engine
func (sc *SharedContext) CreateQueryEngine() (*engine.Engine, error) {
	if sc.Config == nil {
		return nil, fmt.Errorf("configuration not loaded - call LoadConfig() first")
	}

	baseDir := sc.Config.Settings.BaseDir
	indexPath := filepath.Join(baseDir, ".agent-index")
	cachePath := filepath.Join(baseDir, ".agent-cache")

	var queryEngine *engine.Engine
	err := sc.PM.WithSpinner("Initializing query engine", func() error {
		var engineErr error
		queryEngine, engineErr = engine.NewEngine(indexPath, cachePath)
		if engineErr != nil {
			return fmt.Errorf("failed to create query engine: %w", engineErr)
		}

		// Update index if needed
		agentsDir := sc.Config.Settings.BaseDir
		if updateErr := queryEngine.UpdateIndex(agentsDir); updateErr != nil {
			// If update fails, try rebuilding
			if rebuildErr := queryEngine.RebuildIndex(agentsDir); rebuildErr != nil {
				return fmt.Errorf("failed to initialize index: %w", rebuildErr)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return queryEngine, nil
}

// GetSourceByName finds a source configuration by name
func (sc *SharedContext) GetSourceByName(sourceName string) (*config.Source, error) {
	if sc.Config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	for _, source := range sc.Config.Sources {
		if source.Name == sourceName {
			return &source, nil
		}
	}
	return nil, fmt.Errorf("source '%s' not found in configuration", sourceName)
}

// FilterEnabledSources filters sources to only enabled ones, optionally filtered by name
func (sc *SharedContext) FilterEnabledSources(sourceName string) ([]config.Source, error) {
	if sc.Config == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	var sources []config.Source

	if sourceName != "" {
		source, err := sc.GetSourceByName(sourceName)
		if err != nil {
			return nil, err
		}
		if source.Enabled {
			sources = []config.Source{*source}
		}
	} else {
		for _, source := range sc.Config.Sources {
			if source.Enabled {
				sources = append(sources, source)
			}
		}
	}

	return sources, nil
}

// GetAgentsDirectory returns the base directory where agents are installed
func (sc *SharedContext) GetAgentsDirectory() string {
	if sc.Config == nil {
		return ""
	}
	return sc.Config.Settings.BaseDir
}

// AddPersistentFlags adds common flags to a command
func AddPersistentFlags(cmd *cobra.Command, opts *SharedOptions) {
	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "c", "agents-config.yaml", "configuration file")
	cmd.PersistentFlags().BoolVarP(&opts.Verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVar(&opts.DryRun, "dry-run", false, "simulate actions without making changes")
	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "disable colored output")
	cmd.PersistentFlags().BoolVar(&opts.NoProgress, "no-progress", false, "disable progress indicators")
}

// SetupColors configures color output based on options
func SetupColors(noColor bool) {
	if noColor {
		color.NoColor = true
	}
}

// SetupProgress initializes the progress manager with options
func SetupProgress(opts *SharedOptions) {
	progress.Initialize(progress.Options{
		Enabled: !opts.NoProgress,
		Verbose: opts.Verbose,
		DryRun:  opts.DryRun,
		NoColor: opts.NoColor,
	})
}

// PrintSuccess prints a success message with consistent formatting
func PrintSuccess(format string, args ...interface{}) {
	color.Green("✓ "+format+"\n", args...)
}

// PrintWarning prints a warning message with consistent formatting
func PrintWarning(format string, args ...interface{}) {
	color.Yellow("⚠ "+format+"\n", args...)
}

// PrintError prints an error message with consistent formatting
func PrintError(format string, args ...interface{}) {
	color.Red("✗ "+format+"\n", args...)
}

// PrintInfo prints an info message with consistent formatting
func PrintInfo(format string, args ...interface{}) {
	color.Cyan("ℹ "+format+"\n", args...)
}

// Command interface for structured command implementations
type Command interface {
	// Name returns the command name
	Name() string

	// Description returns the command description
	Description() string

	// CreateCommand creates the cobra command
	CreateCommand(sharedCtx *SharedContext) *cobra.Command
}
