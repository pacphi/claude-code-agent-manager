package commands

import (
	"errors"
	"testing"

	"github.com/pacphi/claude-code-agent-manager/internal/config"
	"github.com/pacphi/claude-code-agent-manager/internal/progress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExecutor implements CommandExecutor for testing
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) ExecuteOperation(ctx *SharedContext, sources []config.Source) error {
	args := m.Called(ctx, sources)
	return args.Error(0)
}

func (m *MockExecutor) GetOperationName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockExecutor) GetCompletionMessage() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockExecutor) ShouldContinueOnError(ctx *SharedContext) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func TestBaseCommand(t *testing.T) {
	t.Run("NewBaseCommand", func(t *testing.T) {
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		assert.NotNil(t, baseCmd)
		assert.Equal(t, executor, baseCmd.executor)
	})

	t.Run("ExecuteWithCommonPattern - success", func(t *testing.T) {
		// Setup
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		// Create test configuration
		testConfig := &config.Config{
			Settings: config.Settings{
				BaseDir:         "/test/dir",
				ContinueOnError: false,
			},
			Sources: []config.Source{
				{Name: "test-source", Enabled: true, Type: "github"},
			},
		}

		sharedCtx := &SharedContext{
			Options: &SharedOptions{Verbose: false, NoProgress: false},
			Config:  testConfig,
			PM:      progress.New(progress.Options{Enabled: false}), // Disabled for testing
		}

		// Mock expectations
		executor.On("ExecuteOperation", sharedCtx, mock.MatchedBy(func(sources []config.Source) bool {
			return len(sources) == 1 && sources[0].Name == "test-source"
		})).Return(nil)
		executor.On("GetOperationName").Return("Testing")
		executor.On("GetCompletionMessage").Return("Test complete")
		// This call happens for verbose mode output

		// Execute
		err := baseCmd.executeOnSources(sharedCtx, []config.Source{{Name: "test-source", Enabled: true}})

		// Verify
		assert.NoError(t, err)
		executor.AssertExpectations(t)
	})

	t.Run("ExecuteWithCommonPattern - with error and continue", func(t *testing.T) {
		// Setup
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		testConfig := &config.Config{
			Settings: config.Settings{
				ContinueOnError: true,
			},
		}

		sharedCtx := &SharedContext{
			Options: &SharedOptions{Verbose: false, NoProgress: false},
			Config:  testConfig,
			PM:      progress.New(progress.Options{Enabled: false}), // Disabled for testing
		}

		sources := []config.Source{
			{Name: "source1", Enabled: true},
			{Name: "source2", Enabled: true},
		}

		// Mock expectations - first source fails, second succeeds
		executor.On("ExecuteOperation", sharedCtx, mock.MatchedBy(func(s []config.Source) bool {
			return len(s) == 1 && s[0].Name == "source1"
		})).Return(errors.New("test error"))
		executor.On("ExecuteOperation", sharedCtx, mock.MatchedBy(func(s []config.Source) bool {
			return len(s) == 1 && s[0].Name == "source2"
		})).Return(nil)
		executor.On("GetOperationName").Return("Testing")
		executor.On("GetCompletionMessage").Return("Test complete")
		executor.On("ShouldContinueOnError", sharedCtx).Return(true)

		// Execute
		err := baseCmd.executeOnSources(sharedCtx, sources)

		// Verify - should not return error because we continue on error
		assert.NoError(t, err)
		executor.AssertExpectations(t)
	})

	t.Run("ExecuteWithCommonPattern - with error and no continue", func(t *testing.T) {
		// Setup
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		testConfig := &config.Config{
			Settings: config.Settings{
				ContinueOnError: false,
			},
		}

		sharedCtx := &SharedContext{
			Options: &SharedOptions{Verbose: false, NoProgress: false},
			Config:  testConfig,
			PM:      progress.New(progress.Options{Enabled: false}), // Disabled for testing
		}

		sources := []config.Source{
			{Name: "source1", Enabled: true},
		}

		// Mock expectations
		executor.On("ExecuteOperation", sharedCtx, mock.MatchedBy(func(s []config.Source) bool {
			return len(s) == 1 && s[0].Name == "source1"
		})).Return(errors.New("test error"))
		executor.On("GetOperationName").Return("Testing")
		executor.On("ShouldContinueOnError", sharedCtx).Return(false)

		// Execute
		err := baseCmd.executeOnSources(sharedCtx, sources)

		// Verify - should return error because we don't continue on error
		assert.Error(t, err)
		executor.AssertExpectations(t)
	})

	t.Run("validateSources", func(t *testing.T) {
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		// Test with empty sources and no source name
		err := baseCmd.validateSources([]config.Source{}, "")
		assert.NoError(t, err) // Should not error, just print warning

		// Test with empty sources but specific source name
		err = baseCmd.validateSources([]config.Source{}, "test-source")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test-source")

		// Test with sources
		sources := []config.Source{{Name: "test"}}
		err = baseCmd.validateSources(sources, "")
		assert.NoError(t, err)
	})

	t.Run("shouldUseSpinner", func(t *testing.T) {
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		// Test spinner enabled (not verbose, progress enabled)
		sharedCtx := &SharedContext{
			Options: &SharedOptions{Verbose: false, NoProgress: false},
		}
		assert.True(t, baseCmd.shouldUseSpinner(sharedCtx))

		// Test spinner disabled by verbose mode
		sharedCtx.Options.Verbose = true
		assert.False(t, baseCmd.shouldUseSpinner(sharedCtx))

		// Test spinner disabled by no progress
		sharedCtx.Options.Verbose = false
		sharedCtx.Options.NoProgress = true
		assert.False(t, baseCmd.shouldUseSpinner(sharedCtx))
	})

	t.Run("operation name helpers", func(t *testing.T) {
		executor := &MockExecutor{}
		baseCmd := NewBaseCommand(executor)

		// Test verb extraction
		executor.On("GetOperationName").Return("Installing")
		assert.Equal(t, "install", baseCmd.getOperationVerb())

		executor.ExpectedCalls = nil // Reset mock
		executor.On("GetOperationName").Return("Updating")
		assert.Equal(t, "update", baseCmd.getOperationVerb())

		// Test past tense conversion
		executor.ExpectedCalls = nil
		executor.On("GetOperationName").Return("Installing")
		assert.Equal(t, "installed", baseCmd.getOperationPastTense())

		executor.ExpectedCalls = nil
		executor.On("GetOperationName").Return("Updating")
		assert.Equal(t, "updated", baseCmd.getOperationPastTense())
	})
}

func TestInstallCommandWithBase(t *testing.T) {
	t.Run("implements CommandExecutor correctly", func(t *testing.T) {
		installCmd := NewInstallCommand()

		// Verify it implements the interface
		assert.Equal(t, "Installing", installCmd.GetOperationName())
		assert.Equal(t, "Installation complete", installCmd.GetCompletionMessage())

		// Test ShouldContinueOnError
		testConfig := &config.Config{
			Settings: config.Settings{ContinueOnError: true},
		}
		sharedCtx := &SharedContext{Config: testConfig}
		assert.True(t, installCmd.ShouldContinueOnError(sharedCtx))
	})
}

func TestUpdateCommandWithBase(t *testing.T) {
	t.Run("implements CommandExecutor correctly", func(t *testing.T) {
		updateCmd := NewUpdateCommand()

		// Test normal mode
		assert.Equal(t, "Updating", updateCmd.GetOperationName())
		assert.Equal(t, "Update complete", updateCmd.GetCompletionMessage())

		// Test check-only mode
		updateCmd.checkOnly = true
		assert.Equal(t, "Checking updates for", updateCmd.GetOperationName())
		assert.Equal(t, "Check complete", updateCmd.GetCompletionMessage())

		// Test ShouldContinueOnError
		testConfig := &config.Config{
			Settings: config.Settings{ContinueOnError: false},
		}
		sharedCtx := &SharedContext{Config: testConfig}
		assert.False(t, updateCmd.ShouldContinueOnError(sharedCtx))
	})
}
