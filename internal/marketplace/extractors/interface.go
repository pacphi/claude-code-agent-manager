package extractors

import (
	"context"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
)

// CategoryExtractor extracts category data from web pages
type CategoryExtractor interface {
	Extract(ctx context.Context, browser browser.Controller) ([]types.Category, error)
}

// AgentExtractor extracts agent data from web pages
type AgentExtractor interface {
	Extract(ctx context.Context, browser browser.Controller, category string) ([]types.Agent, error)
}

// ContentExtractor extracts content from agent pages
type ContentExtractor interface {
	Extract(ctx context.Context, browser browser.Controller, url string) (string, error)
}

// ScriptProvider provides JavaScript scripts for extraction
type ScriptProvider interface {
	LoadCategoriesScript() (string, error)
	LoadAgentsScript() (string, error)
	LoadContentScript() (string, error)
}
