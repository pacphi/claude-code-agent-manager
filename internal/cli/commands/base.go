package commands

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/config"
)

// CommandExecutor defines the interface for command-specific execution logic
type CommandExecutor interface {
	// ExecuteOperation performs the actual command operation
	ExecuteOperation(ctx *SharedContext, sources []config.Source) error

	// GetOperationName returns a human-readable name for the operation (e.g., "Installing", "Updating")
	GetOperationName() string

	// GetCompletionMessage returns the completion message for successful operations
	GetCompletionMessage() string

	// ShouldContinueOnError returns true if the operation should continue when individual sources fail
	ShouldContinueOnError(ctx *SharedContext) bool
}

// BaseCommand provides shared execution patterns for commands
type BaseCommand struct {
	executor CommandExecutor
}

// NewBaseCommand creates a new base command with the specified executor
func NewBaseCommand(executor CommandExecutor) *BaseCommand {
	return &BaseCommand{
		executor: executor,
	}
}

// ExecuteWithCommonPattern executes command logic using the common pattern
func (bc *BaseCommand) ExecuteWithCommonPattern(sharedCtx *SharedContext, sourceName string) error {
	// Load and validate configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Get sources to process
	sources, err := sharedCtx.FilterEnabledSources(sourceName)
	if err != nil {
		return err
	}

	// Validate sources
	if err := bc.validateSources(sources, sourceName); err != nil {
		return err
	}

	// Execute operation on sources
	return bc.executeOnSources(sharedCtx, sources)
}

// validateSources validates that we have sources to process
func (bc *BaseCommand) validateSources(sources []config.Source, sourceName string) error {
	if len(sources) == 0 {
		if sourceName != "" {
			return fmt.Errorf("source '%s' is not enabled or not found", sourceName)
		}
		PrintWarning("No enabled sources found in configuration")
		return nil
	}
	return nil
}

// executeOnSources executes the operation on all provided sources
func (bc *BaseCommand) executeOnSources(sharedCtx *SharedContext, sources []config.Source) error {
	if len(sources) == 0 {
		return nil // No sources to process
	}

	successCount := 0
	failCount := 0
	operationName := bc.executor.GetOperationName()

	for _, source := range sources {
		if bc.shouldUseSpinner(sharedCtx) {
			// Use spinner for non-verbose mode
			err := sharedCtx.PM.WithSpinner(fmt.Sprintf("%s %s", operationName, source.Name), func() error {
				return bc.executor.ExecuteOperation(sharedCtx, []config.Source{source})
			})

			if err != nil {
				PrintError("Failed to %s %s: %v", bc.getOperationVerb(), source.Name, err)
				failCount++
				if !bc.executor.ShouldContinueOnError(sharedCtx) {
					return err
				}
			} else {
				successCount++
			}
		} else {
			// Verbose mode with detailed output
			color.Blue("%s source: %s\n", operationName, source.Name)

			if err := bc.executor.ExecuteOperation(sharedCtx, []config.Source{source}); err != nil {
				PrintError("Failed to %s %s: %v", bc.getOperationVerb(), source.Name, err)
				failCount++
				if !bc.executor.ShouldContinueOnError(sharedCtx) {
					return err
				}
			} else {
				PrintSuccess("Successfully %s %s", bc.getOperationPastTense(), source.Name)
				successCount++
			}
		}
	}

	// Print summary
	bc.printSummary(successCount, failCount)
	return nil
}

// shouldUseSpinner determines if spinner should be used based on options
func (bc *BaseCommand) shouldUseSpinner(sharedCtx *SharedContext) bool {
	return !sharedCtx.Options.NoProgress && !sharedCtx.Options.Verbose
}

// getOperationVerb returns the verb form of the operation (e.g., "install", "update")
func (bc *BaseCommand) getOperationVerb() string {
	// Convert "Installing" -> "install", "Updating" -> "update"
	name := bc.executor.GetOperationName()

	// Handle specific known cases
	switch name {
	case "Installing":
		return "install"
	case "Updating":
		return "update"
	case "Uninstalling":
		return "uninstall"
	}

	// Generic case: remove "ing" suffix and convert to lowercase
	if len(name) > 3 && name[len(name)-3:] == "ing" {
		verb := name[:len(name)-3]
		return strings.ToLower(verb)
	}

	// Convert to lowercase for other cases
	return strings.ToLower(name)
}

// getOperationPastTense returns the past tense form of the operation
func (bc *BaseCommand) getOperationPastTense() string {
	// Convert "Installing" -> "installed", "Updating" -> "updated"
	verb := bc.getOperationVerb()
	switch verb {
	case "install":
		return "installed"
	case "update":
		return "updated"
	case "uninstall":
		return "uninstalled"
	default:
		return verb + "ed"
	}
}

// printSummary prints the operation summary
func (bc *BaseCommand) printSummary(successCount, failCount int) {
	fmt.Println()

	completionMsg := bc.executor.GetCompletionMessage()

	if successCount > 0 {
		PrintSuccess("%s: %d succeeded", completionMsg, successCount)
	}

	if failCount > 0 {
		PrintError("%d failed", failCount)
	}

	if successCount == 0 && failCount == 0 {
		PrintInfo("No sources processed")
	}
}
