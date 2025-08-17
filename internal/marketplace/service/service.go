package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/cache"
	"github.com/pacphi/claude-code-agent-manager/internal/progress"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// marketplaceService implements MarketplaceService interface
type marketplaceService struct {
	browser    browser.Controller
	cache      cache.Manager
	extractors ExtractorSet
	baseURL    string
	config     Config
}

// Config holds service configuration
type Config struct {
	BaseURL        string
	CacheEnabled   bool
	CacheTTL       time.Duration
	RequestTimeout time.Duration
	UserAgent      string
}

// NewMarketplaceService creates a new marketplace service
func NewMarketplaceService(
	browser browser.Controller,
	cache cache.Manager,
	extractors ExtractorSet,
	config Config,
) MarketplaceService {
	return &marketplaceService{
		browser:    browser,
		cache:      cache,
		extractors: extractors,
		baseURL:    config.BaseURL,
		config:     config,
	}
}

// GetCategories retrieves all marketplace categories
func (s *marketplaceService) GetCategories(ctx context.Context) ([]types.Category, error) {
	util.DebugPrintf("GetCategories called\n")

	// Check cache first
	if s.config.CacheEnabled {
		if cached := s.cache.GetCategories(); cached != nil {
			if categories, ok := cached.([]types.Category); ok {
				util.DebugPrintf("Returning cached categories: %d\n", len(categories))
				return categories, nil
			}
		}
	}
	util.DebugPrintf("No cached categories, proceeding with extraction\n")

	// Navigate to categories page where all categories are displayed with agent counts
	pm := progress.Default()
	var categories []types.Category

	err := pm.WithSpinner("Fetching marketplace categories", func() error {
		categoriesURL := fmt.Sprintf("%s/categories", s.baseURL)
		util.DebugPrintf("Navigating to: %s\n", categoriesURL)
		if err := s.browser.Navigate(ctx, categoriesURL); err != nil {
			util.DebugPrintf("Navigation failed: %v\n", err)
			return fmt.Errorf("failed to navigate to categories page: %w", err)
		}
		util.DebugPrintf("Navigation successful\n")

		// Extract categories from categories page (gets names and descriptions)
		util.DebugPrintf("Starting category extraction\n")
		var extractErr error
		categories, extractErr = s.extractors.Categories.Extract(ctx, s.browser)
		if extractErr != nil {
			util.DebugPrintf("Category extraction failed: %v\n", extractErr)
			return fmt.Errorf("failed to extract categories: %w", extractErr)
		}
		util.DebugPrintf("Category extraction completed: %d categories\n", len(categories))
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Categories extracted from /categories page include real agent counts

	// Cache results
	if s.config.CacheEnabled {
		s.cache.SetCategories(categories)
	}

	return categories, nil
}

// GetAgents retrieves agents from a specific category
func (s *marketplaceService) GetAgents(ctx context.Context, category string) ([]types.Agent, error) {
	if category == "" {
		return nil, ErrInvalidCategory
	}

	// Check cache first
	if s.config.CacheEnabled {
		if cached := s.cache.GetAgents(category); cached != nil {
			if agents, ok := cached.([]types.Agent); ok {
				return agents, nil
			}
		}
	}

	// Fetch agents with progress
	pm := progress.Default()
	var agents []types.Agent

	err := pm.WithSpinner(fmt.Sprintf("Fetching agents from %s", category), func() error {
		// Navigate to category page
		categoryURL := fmt.Sprintf("%s/categories/%s", s.baseURL, category)
		if err := s.browser.Navigate(ctx, categoryURL); err != nil {
			return fmt.Errorf("failed to navigate to category %s: %w", category, err)
		}

		// Extract agents
		var extractErr error
		agents, extractErr = s.extractors.Agents.Extract(ctx, s.browser, category)
		if extractErr != nil {
			return fmt.Errorf("failed to extract agents from category %s: %w", category, extractErr)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Cache results
	if s.config.CacheEnabled {
		s.cache.SetAgents(category, agents)
	}

	return agents, nil
}

// GetAgent retrieves a specific agent by ID
func (s *marketplaceService) GetAgent(ctx context.Context, agentID string) (*types.Agent, error) {
	if agentID == "" {
		return nil, ErrAgentNotFound
	}

	// Check cache first
	if s.config.CacheEnabled {
		if cached := s.cache.GetAgent(agentID); cached != nil {
			if agent, ok := cached.(*types.Agent); ok {
				return agent, nil
			}
		}
	}

	// For now, search through all categories to find the agent
	// TODO: Implement direct agent lookup when API supports it
	categories, err := s.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	for _, category := range categories {
		agents, err := s.GetAgents(ctx, category.Slug)
		if err != nil {
			continue // Skip categories that fail
		}

		for _, agent := range agents {
			if agent.ID == agentID || agent.Slug == agentID {
				// Cache the agent
				if s.config.CacheEnabled {
					s.cache.SetAgent(agentID, &agent)
				}
				return &agent, nil
			}
		}
	}

	return nil, ErrAgentNotFound
}

// SearchAgents searches for agents matching a query
func (s *marketplaceService) SearchAgents(ctx context.Context, query string) ([]types.Agent, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	query = strings.ToLower(strings.TrimSpace(query))

	// Get all categories and search through them
	categories, err := s.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	var allAgents []types.Agent
	for _, category := range categories {
		agents, err := s.GetAgents(ctx, category.Slug)
		if err != nil {
			continue // Skip categories that fail
		}
		allAgents = append(allAgents, agents...)
	}

	// Filter agents based on query
	var results []types.Agent
	for _, agent := range allAgents {
		if s.matchesQuery(agent, query) {
			results = append(results, agent)
		}
	}

	return results, nil
}

// GetAgentContent retrieves the full content/definition of an agent
func (s *marketplaceService) GetAgentContent(ctx context.Context, agentID string) (string, error) {
	agent, err := s.GetAgent(ctx, agentID)
	if err != nil {
		return "", err
	}

	if agent.ContentURL == "" {
		return agent.Description, nil
	}

	// Navigate to agent content page and extract content
	if err := s.browser.Navigate(ctx, agent.ContentURL); err != nil {
		return "", fmt.Errorf("failed to navigate to agent content: %w", err)
	}

	// Extract content using JavaScript
	script := `
		(function() {
			// Look for common content containers
			const selectors = [
				'pre', 'code', '.content', '.agent-content',
				'[data-content]', '.markdown-body'
			];

			for (const selector of selectors) {
				const element = document.querySelector(selector);
				if (element && element.textContent.trim().length > 50) {
					return element.textContent.trim();
				}
			}

			// Fallback to body content
			return document.body.textContent.trim();
		})();
	`

	result, err := s.browser.ExecuteScript(ctx, script)
	if err != nil {
		return "", fmt.Errorf("failed to extract agent content: %w", err)
	}

	if content, ok := result.(string); ok {
		return content, nil
	}

	return agent.Description, nil
}

// RefreshCache clears and rebuilds the cache
func (s *marketplaceService) RefreshCache(ctx context.Context) error {
	s.cache.Clear()

	// Warm up cache with categories
	_, err := s.GetCategories(ctx)
	return err
}

// ClearCache clears all cached data
func (s *marketplaceService) ClearCache() error {
	s.cache.Clear()
	return nil
}

// HealthCheck verifies the service is operational
func (s *marketplaceService) HealthCheck(ctx context.Context) error {
	// Try to navigate to the base URL
	if err := s.browser.Navigate(ctx, s.baseURL); err != nil {
		return fmt.Errorf("marketplace unreachable: %w", err)
	}

	return nil
}

// matchesQuery checks if an agent matches the search query
func (s *marketplaceService) matchesQuery(agent types.Agent, query string) bool {
	searchFields := []string{
		strings.ToLower(agent.Name),
		strings.ToLower(agent.Description),
		strings.ToLower(agent.Author),
		strings.ToLower(agent.Category),
	}

	for _, field := range searchFields {
		if strings.Contains(field, query) {
			return true
		}
	}

	// Check tags
	for _, tag := range agent.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}
