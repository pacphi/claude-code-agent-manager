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
		util.DebugPrintf("Cached %d categories\n", len(categories))
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
				util.DebugPrintf("Cache hit: returning %d agents for category %s\n", len(agents), category)
				return agents, nil
			}
		}
		util.DebugPrintf("Cache miss: fetching agents for category %s\n", category)
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

	// Cache results and populate individual agent cache + category mappings
	if s.config.CacheEnabled {
		s.cache.SetAgents(category, agents)

		// Warm individual agent cache and category mappings
		for _, agent := range agents {
			s.cache.SetAgent(agent.ID, &agent)
			s.cache.SetAgentCategory(agent.ID, category)
			// Also cache by slug if different from ID
			if agent.Slug != agent.ID {
				s.cache.SetAgent(agent.Slug, &agent)
				s.cache.SetAgentCategory(agent.Slug, category)
			}
		}
		util.DebugPrintf("Warmed cache with %d individual agents from category %s\n", len(agents), category)
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
				util.DebugPrintf("Cache hit: found agent %s directly\n", agentID)
				return agent, nil
			}
		}

		// Check if we know which category this agent belongs to
		if category := s.cache.GetAgentCategory(agentID); category != "" {
			util.DebugPrintf("Cache optimization: found cached category %s for agent %s, searching directly\n", category, agentID)
			agents, err := s.GetAgents(ctx, category)
			if err == nil {
				for _, agent := range agents {
					if agent.ID == agentID || agent.Slug == agentID {
						// Cache the agent since GetAgents would have populated category agents but not individual agent
						s.cache.SetAgent(agentID, &agent)
						util.DebugPrintf("Cache optimization: found agent %s in predicted category %s\n", agentID, category)
						return &agent, nil
					}
				}
			}
		}
	}

	util.DebugPrintf("No cached data for agent %s, searching all categories\n", agentID)

	// Fall back to searching through all categories
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
				// Cache the agent and its category mapping
				if s.config.CacheEnabled {
					s.cache.SetAgent(agentID, &agent)
					s.cache.SetAgentCategory(agentID, category.Slug)
					// Also cache by slug if different from ID
					if agent.Slug != agent.ID {
						s.cache.SetAgentCategory(agent.Slug, category.Slug)
					}
				}
				return &agent, nil
			}
		}
	}

	return nil, ErrAgentNotFound
}

// GetAgentContent retrieves the full content/definition of an agent
func (s *marketplaceService) GetAgentContent(ctx context.Context, agentID string) (string, error) {
	agent, err := s.GetAgent(ctx, agentID)
	if err != nil {
		return "", err
	}

	// Determine the detail URL for content extraction
	var detailURL string
	if agent.ContentURL != "" &&
		!strings.Contains(agent.ContentURL, "search?q=") &&
		!strings.HasPrefix(agent.ContentURL, "NAVIGATE_TO:") {
		detailURL = agent.ContentURL
		util.DebugPrintf("Using stored ContentURL: %s\n", detailURL)
	} else {
		// Try to find the actual agent detail page URL by navigation
		detailURL, err = s.findAgentDetailURL(ctx, agent.Name)
		if err != nil {
			util.DebugPrintf("Failed to find agent detail URL for %s: %v\n", agentID, err)
			return agent.Description, nil
		}

		// Update the agent's ContentURL cache for future use
		if detailURL != "" {
			agent.ContentURL = detailURL
			// Cache the updated agent
			if s.config.CacheEnabled {
				s.cache.SetAgent(agentID, agent)
			}
		}
	}

	// Use the content extractor to get the agent definition
	content, err := s.extractors.Content.Extract(ctx, s.browser, detailURL)
	if err != nil {
		// Log the error but don't fail completely
		util.DebugPrintf("Failed to extract content for agent %s from URL %s: %v\n", agentID, detailURL, err)
		// Fall back to description if content extraction fails
		return agent.Description, nil
	}

	// Validate that we got actual content, not just an empty string
	if content == "" || len(content) < len(agent.Description) {
		util.DebugPrintf("Extracted content seems invalid (empty or too short), falling back to description\n")
		return agent.Description, nil
	}

	return content, nil
}

// findAgentDetailURL attempts to find the actual detail page URL for an agent
func (s *marketplaceService) findAgentDetailURL(ctx context.Context, agentName string) (string, error) {
	// Strategy 1: Navigate and click approach for agents page
	agentsURL := fmt.Sprintf("%s/agents", s.baseURL)
	if err := s.browser.Navigate(ctx, agentsURL); err != nil {
		return "", fmt.Errorf("failed to navigate to agents page: %w", err)
	}

	// Wait for page to load
	util.DebugPrintf("Waiting for agents page to load...\n")
	time.Sleep(3 * time.Second)

	// Execute a script to find and click the agent card
	clickScript := fmt.Sprintf(`
(function() {
	const targetAgentName = %q;
	console.log('Searching for agent to click:', targetAgentName);

	// Strategy 1: Look for heading elements containing the agent name
	const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
	for (const heading of headings) {
		if (heading.textContent && heading.textContent.trim().toLowerCase() === targetAgentName.toLowerCase()) {
			console.log('Found agent heading:', heading.textContent);

			// Look for clickable parent containers
			let parent = heading.parentElement;
			let depth = 0;
			while (parent && depth < 5) {
				if (parent.style.cursor === 'pointer' ||
					parent.onclick ||
					parent.classList.contains('cursor-pointer') ||
					parent.getAttribute('role') === 'button') {
					console.log('Found clickable parent container at depth', depth);
					parent.click();
					return 'CLICKED';
				}
				parent = parent.parentElement;
				depth++;
			}

			// If no clickable parent, try clicking the heading itself
			console.log('Trying to click heading directly');
			heading.click();
			return 'CLICKED';
		}
	}

	// Strategy 2: Look for View buttons and find the one in the same container as the agent name
	const viewButtons = Array.from(document.querySelectorAll('button')).filter(btn =>
		btn.textContent && btn.textContent.trim() === 'View'
	);

	for (const viewButton of viewButtons) {
		let container = viewButton;
		let depth = 0;

		// Walk up to find the agent card container
		while (container && depth < 10) {
			container = container.parentElement;
			depth++;

			if (container && container.textContent &&
				container.textContent.toLowerCase().includes(targetAgentName.toLowerCase())) {
				console.log('Found agent container via View button');
				viewButton.click();
				return 'CLICKED';
			}
		}
	}

	console.log('No clickable agent card found');
	return null;
})();`, agentName)

	result, err := s.browser.ExecuteScript(ctx, clickScript)
	if err != nil {
		return "", fmt.Errorf("failed to execute click script: %w", err)
	}

	if result == "CLICKED" {
		// Wait for navigation to complete
		util.DebugPrintf("Waiting for navigation after click...\n")
		time.Sleep(2 * time.Second)

		// Get the current URL
		getCurrentURLScript := `window.location.href`
		urlResult, err := s.browser.ExecuteScript(ctx, getCurrentURLScript)
		if err != nil {
			return "", fmt.Errorf("failed to get current URL: %w", err)
		}

		if currentURL, ok := urlResult.(string); ok {
			if currentURL != agentsURL && strings.Contains(currentURL, "/agents/") {
				util.DebugPrintf("Successfully navigated to agent detail page: %s\n", currentURL)
				return currentURL, nil
			}
		}
	}

	return "", fmt.Errorf("agent detail URL not found for %s", agentName)
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

// GetCacheStats returns cache performance statistics
func (s *marketplaceService) GetCacheStats() interface{} {
	return s.cache.GetStats()
}

// HealthCheck verifies the service is operational
func (s *marketplaceService) HealthCheck(ctx context.Context) error {
	// Try to navigate to the base URL
	if err := s.browser.Navigate(ctx, s.baseURL); err != nil {
		return fmt.Errorf("marketplace unreachable: %w", err)
	}

	return nil
}
