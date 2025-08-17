package types

import "time"

// Category represents a marketplace category
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AgentCount  int    `json:"agent_count"`
	URL         string `json:"url"`
	Slug        string `json:"slug"`
}

// Agent represents a marketplace agent
type Agent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	CategoryID  string    `json:"category_id"`
	Rating      float32   `json:"rating"`
	Downloads   int       `json:"downloads"`
	Tags        []string  `json:"tags"`
	Author      string    `json:"author"`
	ContentURL  string    `json:"content_url"`
	Content     string    `json:"content,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MarketplaceData represents the complete marketplace data
type MarketplaceData struct {
	Categories []Category `json:"categories"`
	Agents     []Agent    `json:"agents"`
	Timestamp  time.Time  `json:"timestamp"`
	Version    string     `json:"version"`
}

// SearchOptions for filtering agents
type SearchOptions struct {
	Query     string
	Category  string
	Tags      []string
	MinRating float32
	Limit     int
	Offset    int
}

// CacheConfig for marketplace caching
type CacheConfig struct {
	Enabled   bool `yaml:"enabled"`
	TTLHours  int  `yaml:"ttl_hours"`
	MaxSizeMB int  `yaml:"max_size_mb"`
	RefreshBg bool `yaml:"refresh_background"`
}

// Metrics for monitoring marketplace operations
type Metrics struct {
	CacheHits     int64         `json:"cache_hits"`
	CacheMisses   int64         `json:"cache_misses"`
	ScrapeTime    time.Duration `json:"scrape_time"`
	ErrorCount    int64         `json:"error_count"`
	LastScrape    time.Time     `json:"last_scrape"`
	AgentsFetched int           `json:"agents_fetched"`
}
