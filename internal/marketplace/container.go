package marketplace

import (
	"fmt"
	"time"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/cache"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/extractors"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/service"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// Container holds all marketplace dependencies
type Container struct {
	Browser browser.Controller
	Cache   cache.Manager
	Service service.MarketplaceService
	Config  ContainerConfig
}

// ContainerConfig holds configuration for the container
type ContainerConfig struct {
	BaseURL         string
	CacheEnabled    bool
	CacheTTLHours   int
	CacheMaxSizeMB  int64
	BrowserHeadless bool
	BrowserTimeout  int
	UserAgent       string
}

// DefaultContainerConfig returns sensible defaults
func DefaultContainerConfig() ContainerConfig {
	return ContainerConfig{
		BaseURL:         "https://subagents.sh",
		CacheEnabled:    true,
		CacheTTLHours:   1,
		CacheMaxSizeMB:  50,
		BrowserHeadless: true,
		BrowserTimeout:  30,
		UserAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// NewContainer creates a new dependency injection container
func NewContainer(config ContainerConfig) (*Container, error) {
	// Create browser controller
	browserOpts := browser.Options{
		Headless:     config.BrowserHeadless,
		Timeout:      config.BrowserTimeout,
		UserAgent:    config.UserAgent,
		WindowWidth:  1920,
		WindowHeight: 1080,
	}

	util.DebugPrintf("Creating browser controller with options: %+v\n", browserOpts)
	browserController, err := browser.NewController(browserOpts)
	if err != nil {
		util.DebugPrintf("Browser controller creation failed: %v\n", err)
		return nil, fmt.Errorf("failed to create browser controller: %w", err)
	}
	util.DebugPrintf("Browser controller created successfully\n")

	// Create cache manager
	cacheConfig := cache.Config{
		Enabled:   config.CacheEnabled,
		TTL:       time.Duration(config.CacheTTLHours) * time.Hour,
		MaxSizeMB: config.CacheMaxSizeMB,
	}

	cacheManager, err := cache.NewManager(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}

	// Create extractors
	extractorFactory := extractors.NewFactory()
	extractorSet := extractorFactory.CreateExtractorSet()

	// Create service
	serviceConfig := service.Config{
		BaseURL:        config.BaseURL,
		CacheEnabled:   config.CacheEnabled,
		CacheTTL:       time.Duration(config.CacheTTLHours) * time.Hour,
		RequestTimeout: time.Duration(config.BrowserTimeout) * time.Second,
		UserAgent:      config.UserAgent,
	}

	marketplaceService := service.NewMarketplaceService(
		browserController,
		cacheManager,
		extractorSet,
		serviceConfig,
	)

	return &Container{
		Browser: browserController,
		Cache:   cacheManager,
		Service: marketplaceService,
		Config:  config,
	}, nil
}

// Close cleans up container resources
func (c *Container) Close() error {
	if c.Browser != nil {
		return c.Browser.Close()
	}
	return nil
}

// WithConfig creates a new container with custom configuration
func WithConfig(config ContainerConfig) (*Container, error) {
	return NewContainer(config)
}

// WithDefaults creates a new container with default configuration
func WithDefaults() (*Container, error) {
	return NewContainer(DefaultContainerConfig())
}
