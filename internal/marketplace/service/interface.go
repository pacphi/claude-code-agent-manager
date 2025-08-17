package service

import (
	"context"
	"errors"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
)

var (
	ErrServiceUnavailable = errors.New("marketplace service unavailable")
	ErrInvalidCategory    = errors.New("invalid category")
	ErrAgentNotFound      = errors.New("agent not found")
	ErrRateLimited        = errors.New("rate limited")
)

// MarketplaceService defines the interface for marketplace operations
type MarketplaceService interface {
	// Categories
	GetCategories(ctx context.Context) ([]types.Category, error)

	// Agents
	GetAgents(ctx context.Context, category string) ([]types.Agent, error)
	GetAgent(ctx context.Context, agentID string) (*types.Agent, error)

	// Content
	GetAgentContent(ctx context.Context, agentID string) (string, error)

	// Cache management
	RefreshCache(ctx context.Context) error
	ClearCache() error

	// Health
	HealthCheck(ctx context.Context) error
}

// SearchQuery defines search parameters
type SearchQuery struct {
	Query     string
	Category  string
	Limit     int
	MinRating float32
	SortBy    string // name, rating, downloads, date
	SortOrder string // asc, desc
}

// ExtractorSet holds all data extractors
type ExtractorSet struct {
	Categories CategoryExtractor
	Agents     AgentExtractor
}

// CategoryExtractor extracts category data
type CategoryExtractor interface {
	Extract(ctx context.Context, browser browser.Controller) ([]types.Category, error)
}

// AgentExtractor extracts agent data
type AgentExtractor interface {
	Extract(ctx context.Context, browser browser.Controller, category string) ([]types.Agent, error)
}
