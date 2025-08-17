package extractors

import (
	"context"
	"fmt"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
)

// contentExtractor implements ContentExtractor interface
type contentExtractor struct {
	scriptProvider ScriptProvider
}

// NewContentExtractor creates a new content extractor
func NewContentExtractor(scriptProvider ScriptProvider) ContentExtractor {
	return &contentExtractor{
		scriptProvider: scriptProvider,
	}
}

// Extract extracts content from an agent page
func (e *contentExtractor) Extract(ctx context.Context, browser browser.Controller, url string) (string, error) {
	// Navigate to the content URL
	if err := browser.Navigate(ctx, url); err != nil {
		return "", fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	// Load the content extraction script
	script, err := e.scriptProvider.LoadContentScript()
	if err != nil {
		return "", fmt.Errorf("failed to load content script: %w", err)
	}

	// Execute the script
	result, err := browser.ExecuteScript(ctx, script)
	if err != nil {
		return "", fmt.Errorf("failed to execute content script: %w", err)
	}

	// Convert result to string
	if content, ok := result.(string); ok {
		return content, nil
	}

	return "", fmt.Errorf("unexpected content type: %T", result)
}
