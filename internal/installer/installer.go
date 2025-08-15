package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/conflict"
	"github.com/pacphi/claude-code-agent-manager/internal/tracker"
	"github.com/pacphi/claude-code-agent-manager/internal/transformer"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// Options contains installer options
type Options struct {
	Verbose     bool
	DryRun      bool
	KeepBackups bool
}

// Installer manages agent installation
type Installer struct {
	config   *config.Config
	tracker  *tracker.Tracker
	resolver *conflict.Resolver
	options  Options
}

// New creates a new installer instance
func New(cfg *config.Config, track *tracker.Tracker, resolver *conflict.Resolver, opts Options) *Installer {
	return &Installer{
		config:   cfg,
		tracker:  track,
		resolver: resolver,
		options:  opts,
	}
}

// InstallSource installs agents from a specific source
func (i *Installer) InstallSource(source config.Source) error {
	if i.options.DryRun {
		color.Yellow("[DRY RUN] Would install from source: %s\n", source.Name)
	}

	// Create temporary directory for cloning/copying
	tempDir, err := os.MkdirTemp("", "agent-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			// Log error but don't fail the entire operation
			if i.options.Verbose {
				fmt.Printf("Warning: failed to remove temp directory %s: %v\n", tempDir, err)
			}
		}
	}()

	// Get source handler based on type
	handler, err := i.getSourceHandler(source.Type)
	if err != nil {
		return err
	}

	// Fetch source to temp directory
	if i.options.Verbose {
		fmt.Printf("Fetching source %s...\n", source.Name)
	}

	fetchedPath, commit, err := handler.Fetch(source, tempDir)
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}

	// Apply filters
	files, err := i.applyFilters(fetchedPath, source.Filters)
	if err != nil {
		return fmt.Errorf("failed to apply filters: %w", err)
	}

	if len(files) == 0 {
		color.Yellow("No files matched the filters for source: %s\n", source.Name)
		return nil
	}

	// Prepare installation
	installation := tracker.Installation{
		SourceCommit:  commit,
		Files:         make(map[string]tracker.FileInfo),
		Directories:   []string{},
		DocsGenerated: []string{},
	}

	// Resolve target directory with variable substitution
	targetDir := i.resolveTargetPath(source.Paths.Target)

	// Create target directory if it doesn't exist
	if !i.options.DryRun {
		if err := os.MkdirAll(targetDir, 0750); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	// Apply transformations
	trans := transformer.New(i.config.Settings)
	transformedFiles := files

	for _, transform := range source.Transformations {
		if i.options.Verbose {
			fmt.Printf("Applying transformation: %s\n", transform.Type)
		}

		var err error
		transformedFiles, err = trans.Apply(transformedFiles, transform, fetchedPath, targetDir)
		if err != nil {
			return fmt.Errorf("transformation failed: %w", err)
		}

		// Track generated docs
		if transform.Type == "extract_docs" {
			for _, file := range transformedFiles {
				if filepath.Dir(file) == i.config.Settings.DocsDir {
					installation.DocsGenerated = append(installation.DocsGenerated, file)
				}
			}
		}
	}

	// Copy files to target with conflict resolution
	conflictStrategy := source.ConflictStrategy
	if conflictStrategy == "" {
		conflictStrategy = i.config.Settings.ConflictStrategy
	}

	for _, relPath := range transformedFiles {
		srcPath := filepath.Join(fetchedPath, relPath)
		dstPath := filepath.Join(targetDir, relPath)

		if !i.options.DryRun {
			// Check if file already exists (pre-existing)
			var wasPreExisting bool
			if _, err := os.Stat(dstPath); err == nil {
				wasPreExisting = true
				// File exists, resolve conflict
				resolved, err := i.resolver.Resolve(dstPath, srcPath, conflictStrategy)
				if err != nil {
					return fmt.Errorf("conflict resolution failed for %s: %w", dstPath, err)
				}
				if !resolved {
					if i.options.Verbose {
						fmt.Printf("Skipped: %s\n", dstPath)
					}
					continue
				}
			}

			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(dstPath), 0750); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			// Copy file
			if err := i.copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", relPath, err)
			}

			// Track installed file
			info, err := os.Stat(dstPath)
			if err != nil {
				return fmt.Errorf("failed to stat installed file %s: %w", dstPath, err)
			}
			installation.Files[dstPath] = tracker.FileInfo{
				Path:           dstPath,
				Size:           info.Size(),
				Modified:       info.ModTime(),
				WasPreExisting: wasPreExisting,
			}

			// Track directory
			dir := filepath.Dir(dstPath)
			if !contains(installation.Directories, dir) {
				installation.Directories = append(installation.Directories, dir)
			}
		}

		if i.options.Verbose {
			fmt.Printf("Installed: %s\n", dstPath)
		}
	}

	// Run post-install actions
	for _, action := range source.PostInstall {
		if i.options.Verbose {
			fmt.Printf("Running post-install: %s\n", action.Path)
		}

		if !i.options.DryRun {
			if err := i.runPostInstall(action); err != nil {
				color.Red("Post-install action failed: %v\n", err)
				if !i.config.Settings.ContinueOnError {
					return err
				}
			}
		}
	}

	// Save installation tracking
	if !i.options.DryRun {
		if err := i.tracker.RecordInstallation(source.Name, installation); err != nil {
			return fmt.Errorf("failed to record installation: %w", err)
		}
	}

	return nil
}

// UninstallSource removes agents from a specific source
func (i *Installer) UninstallSource(sourceName string) error {
	if i.options.DryRun {
		color.Yellow("[DRY RUN] Would uninstall source: %s\n", sourceName)
	}

	installation, err := i.tracker.GetInstallation(sourceName)
	if err != nil {
		return fmt.Errorf("source not found: %s", sourceName)
	}

	// Restore backups first (if resolver is available and not keeping backups)
	var restoredFiles map[string]bool
	if i.resolver != nil && !i.options.KeepBackups && !i.options.DryRun {
		if i.options.Verbose {
			fmt.Printf("Restoring original files from backup...\n")
		}
		var err error
		restoredFiles, err = i.resolver.RestoreBackupFilesWithTracking()
		if err != nil {
			color.Yellow("Warning: Failed to restore backups: %v\n", err)
			// Continue with uninstall even if restore fails
		}
	}

	// Remove files that were installed (skip pre-existing files and files restored from backup)
	for path, fileInfo := range installation.Files {
		if !i.options.DryRun {
			// Skip removing files that were restored from backup
			if restoredFiles != nil && restoredFiles[path] {
				if i.options.Verbose {
					fmt.Printf("Kept restored file: %s\n", path)
				}
				continue
			}

			// Skip removing pre-existing files - they should remain after uninstall
			if fileInfo.WasPreExisting {
				if i.options.Verbose {
					fmt.Printf("Kept pre-existing file: %s\n", path)
				}
				continue
			}

			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				color.Red("Failed to remove %s: %v\n", path, err)
			} else if i.options.Verbose {
				fmt.Printf("Removed: %s\n", path)
			}
		}
	}

	// Remove empty directories
	for _, dir := range installation.Directories {
		if !i.options.DryRun {
			// Only remove if empty
			isEmpty, err := isDirEmpty(dir)
			if err != nil {
				if i.options.Verbose {
					color.Yellow("Warning: failed to check if directory is empty %s: %v", dir, err)
				}
				continue
			}
			if isEmpty {
				if err := os.Remove(dir); err != nil {
					if i.options.Verbose {
						color.Yellow("Warning: failed to remove empty directory %s: %v", dir, err)
					}
				} else if i.options.Verbose {
					fmt.Printf("Removed directory: %s\n", dir)
				}
			}
		}
	}

	// Remove documentation
	for _, doc := range installation.DocsGenerated {
		if !i.options.DryRun {
			if err := os.Remove(doc); err != nil && !os.IsNotExist(err) {
				color.Red("Failed to remove doc %s: %v\n", doc, err)
			} else if i.options.Verbose {
				fmt.Printf("Removed doc: %s\n", doc)
			}
		}
	}

	// Remove from tracking
	if !i.options.DryRun {
		if err := i.tracker.RemoveInstallation(sourceName); err != nil {
			return fmt.Errorf("failed to update tracking: %w", err)
		}
	}

	// Clean up backups unless keeping them
	if !i.options.KeepBackups && !i.options.DryRun && i.resolver != nil {
		if err := i.resolver.CleanupBackups(sourceName); err != nil {
			color.Yellow("Warning: failed to cleanup backups: %v", err)
		}
	}

	color.Green("✓ Uninstalled source: %s\n", sourceName)
	return nil
}

// UninstallAll removes all installed agents
func (i *Installer) UninstallAll() error {
	installations, err := i.tracker.List()
	if err != nil {
		return fmt.Errorf("failed to list installations: %w", err)
	}

	for name := range installations {
		if err := i.UninstallSource(name); err != nil {
			color.Red("Failed to uninstall %s: %v\n", name, err)
			if !i.config.Settings.ContinueOnError {
				return err
			}
		}
	}

	return nil
}

// UpdateSource updates agents from a specific source
func (i *Installer) UpdateSource(sourceName string) error {
	// Find source in config
	var source *config.Source
	for _, s := range i.config.Sources {
		if s.Name == sourceName {
			source = &s
			break
		}
	}

	if source == nil {
		return fmt.Errorf("source not found in configuration: %s", sourceName)
	}

	// Check if already installed
	installation, err := i.tracker.GetInstallation(sourceName)
	if err != nil {
		// Not installed, do fresh install
		return i.InstallSource(*source)
	}

	// Get handler to check for updates
	handler, err := i.getSourceHandler(source.Type)
	if err != nil {
		return err
	}

	// Check if update is available
	hasUpdate, newCommit, err := handler.CheckUpdate(*source, installation.SourceCommit)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !hasUpdate {
		color.Green("✓ %s is up to date\n", sourceName)
		return nil
	}

	if i.options.DryRun {
		color.Yellow("[DRY RUN] Would update %s from %s to %s\n",
			sourceName, installation.SourceCommit[:7], newCommit[:7])
		return nil
	}

	// Perform update by reinstalling
	color.Blue("Updating %s...\n", sourceName)

	// Backup current installation
	if err := i.resolver.CreateBackup(sourceName); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Uninstall current version
	if err := i.UninstallSource(sourceName); err != nil {
		return fmt.Errorf("failed to uninstall old version: %w", err)
	}

	// Install new version
	if err := i.InstallSource(*source); err != nil {
		// Restore backup on failure
		if restoreErr := i.resolver.RestoreBackup(sourceName); restoreErr != nil {
			color.Yellow("Warning: failed to restore backup after installation failure: %v", restoreErr)
		}
		return fmt.Errorf("failed to install update: %w", err)
	}

	color.Green("✓ Updated %s to %s\n", sourceName, newCommit[:7])
	return nil
}

// Helper methods

func (i *Installer) getSourceHandler(sourceType string) (SourceHandler, error) {
	switch sourceType {
	case "github":
		return &GitHubHandler{}, nil
	case "git":
		return &GitHandler{}, nil
	case "local":
		return &LocalHandler{}, nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", sourceType)
	}
}

func (i *Installer) resolveTargetPath(path string) string {
	// Expand variables
	path = os.ExpandEnv(path)

	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		expandedPath, err := util.ExpandPath(path)
		if err != nil {
			// Log error but continue with original path as fallback
			if i.options.Verbose {
				color.Yellow("Warning: failed to expand path %s: %v", path, err)
			}
		} else {
			path = expandedPath
		}
	}

	// Make absolute if relative
	if !filepath.IsAbs(path) {
		pwd, err := os.Getwd()
		if err != nil {
			// Log error but continue with relative path
			if i.options.Verbose {
				color.Yellow("Warning: failed to get current directory: %v", err)
			}
		} else {
			path = filepath.Join(pwd, path)
		}
	}

	return path
}

func (i *Installer) copyFile(src, dst string) error {
	fm := util.NewFileManager()
	return fm.Copy(src, dst)
}

// validateScriptArg validates script arguments for security
func validateScriptArg(arg string) error {
	// Check for null bytes
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("null byte detected in argument: %s", arg)
	}

	// Check for command injection patterns
	injectionPatterns := []string{
		";", "|", "&", "$(", "`", "&&", "||", ">>", "<<",
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("potential command injection in argument: %s", arg)
		}
	}

	return nil
}

func (i *Installer) runPostInstall(action config.PostInstall) error {
	if action.Path == "" {
		return fmt.Errorf("post-install action path is required")
	}

	// Validate script path for security
	if err := util.ValidatePath(action.Path); err != nil {
		return fmt.Errorf("invalid script path: %w", err)
	}

	// Validate all arguments for security
	for i, arg := range action.Args {
		if err := validateScriptArg(arg); err != nil {
			return fmt.Errorf("invalid argument %d: %w", i, err)
		}
	}

	if i.options.Verbose {
		fmt.Printf("Executing post-install script: %s %v\n", action.Path, action.Args)
	}

	// Prepare the command with validated inputs using SecureCommand
	args := append([]string{action.Path}, action.Args...)
	cmd, err := util.SecureCommand("bash", args...)
	if err != nil {
		return fmt.Errorf("failed to create secure command for post-install script: %w", err)
	}

	// Set working directory to project root
	cmd.Dir, _ = os.Getwd()

	// Set up secure environment - SecureCommand already sets this, but we can add project-specific vars
	secureEnv := cmd.Env
	// Add any project-specific environment variables if needed
	cmd.Env = secureEnv

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("post-install script failed: %s\nOutput: %s", err, string(output))
	}

	if i.options.Verbose && len(output) > 0 {
		fmt.Printf("Post-install output:\n%s", string(output))
	}

	return nil
}

func isDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
