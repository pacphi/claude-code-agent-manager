package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/spf13/cobra"
)

// IndexCommand implements the index command functionality
type IndexCommand struct {
	action string
}

// NewIndexCommand creates a new index command instance
func NewIndexCommand() *IndexCommand {
	return &IndexCommand{}
}

// Name returns the command name
func (c *IndexCommand) Name() string {
	return "index"
}

// Description returns the command description
func (c *IndexCommand) Description() string {
	return "Manage search index and cache"
}

// CreateCommand creates the cobra command for index functionality
func (c *IndexCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: c.Description(),
		Long: `Build or rebuild the search index for faster queries and manage the query cache.

Examples:
  agent-manager index build       # Build/update index
  agent-manager index rebuild     # Force rebuild index
  agent-manager index stats       # Show index statistics
  agent-manager index cache-clear # Clear query cache
  agent-manager index cache-stats # Show cache statistics`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"build", "rebuild", "stats", "cache-clear", "cache-stats"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c.action = args[0]
			return c.Execute(sharedCtx)
		},
	}

	return cmd
}

// Execute runs the index command logic
func (c *IndexCommand) Execute(sharedCtx *SharedContext) error {
	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create query engine
	queryEngine, err := sharedCtx.CreateQueryEngine()
	if err != nil {
		return err
	}

	agentsDir := sharedCtx.GetAgentsDirectory()

	switch c.action {
	case "build":
		return c.executeBuild(sharedCtx, queryEngine, agentsDir)
	case "rebuild":
		return c.executeRebuild(sharedCtx, queryEngine, agentsDir)
	case "stats":
		return c.executeStats(sharedCtx, queryEngine)
	case "cache-clear":
		return c.executeCacheClear(sharedCtx, queryEngine)
	case "cache-stats":
		return c.executeCacheStats(sharedCtx, queryEngine)
	default:
		return fmt.Errorf("unknown index action: %s", c.action)
	}
}

// executeBuild builds or updates the search index
func (c *IndexCommand) executeBuild(sharedCtx *SharedContext, queryEngine interface{}, agentsDir string) error {
	// Use type assertion to access methods (this is a safe pattern since we created the engine)
	engine := queryEngine.(*engine.Engine)

	err := sharedCtx.PM.WithSpinner("Building index", func() error {
		return engine.UpdateIndex(agentsDir)
	})
	if err != nil {
		return err
	}

	PrintSuccess("Index built successfully")
	return nil
}

// executeRebuild forces a complete rebuild of the search index
func (c *IndexCommand) executeRebuild(sharedCtx *SharedContext, queryEngine interface{}, agentsDir string) error {
	engine := queryEngine.(*engine.Engine)

	err := sharedCtx.PM.WithSpinner("Rebuilding index", func() error {
		return engine.RebuildIndex(agentsDir)
	})
	if err != nil {
		return err
	}

	PrintSuccess("Index rebuilt successfully")
	return nil
}

// executeStats displays index statistics
func (c *IndexCommand) executeStats(sharedCtx *SharedContext, queryEngine interface{}) error {
	engine := queryEngine.(*engine.Engine)

	var indexStats map[string]interface{}
	err := sharedCtx.PM.WithSpinner("Gathering index statistics", func() error {
		indexStats = engine.GetStats()
		return nil
	})
	if err != nil {
		return err
	}

	c.displayIndexStats(indexStats, sharedCtx)
	return nil
}

// executeCacheClear clears the query cache
func (c *IndexCommand) executeCacheClear(sharedCtx *SharedContext, queryEngine interface{}) error {
	engine := queryEngine.(*engine.Engine)

	err := sharedCtx.PM.WithSpinner("Clearing query cache", func() error {
		return engine.ClearCache()
	})
	if err != nil {
		return err
	}

	PrintSuccess("Cache cleared successfully")
	return nil
}

// executeCacheStats displays cache statistics
func (c *IndexCommand) executeCacheStats(sharedCtx *SharedContext, queryEngine interface{}) error {
	engine := queryEngine.(*engine.Engine)

	var cacheStats map[string]interface{}
	err := sharedCtx.PM.WithSpinner("Gathering cache statistics", func() error {
		cacheStats = engine.GetCacheStats()
		return nil
	})
	if err != nil {
		return err
	}

	c.displayCacheStats(cacheStats, sharedCtx)
	return nil
}

// displayIndexStats displays index statistics
func (c *IndexCommand) displayIndexStats(indexStats map[string]interface{}, sharedCtx *SharedContext) {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	color.Blue("Index Statistics\n")
	fmt.Println(strings.Repeat("=", 40))

	if totalAgents, ok := indexStats["total_agents"].(int); ok {
		fmt.Printf("Total Indexed Agents: %d\n", totalAgents)
	}

	if cacheStats, ok := indexStats["cache_stats"].(map[string]interface{}); ok {
		if hits, exists := cacheStats["hits"].(int); exists {
			fmt.Printf("Cache Hits: %d\n", hits)
		}
		if size, exists := cacheStats["size"].(int); exists {
			fmt.Printf("Cache Size: %d\n", size)
		}
	}

	if indexInfo, ok := indexStats["index_stats"].(map[string]interface{}); ok {
		if lastUpdate, exists := indexInfo["last_updated"].(time.Time); exists {
			fmt.Printf("Last Updated: %s\n", lastUpdate.Format("2006-01-02 15:04:05"))
		}
	}

	if sources, ok := indexStats["by_source"].(map[string]int); ok && len(sources) > 0 {
		fmt.Printf("\nBy Source:\n")
		for source, count := range sources {
			fmt.Printf("  %s: %d\n", source, count)
		}
	}

	// Performance metrics
	if toolsInherited, ok := indexStats["tools_inherited"].(int); ok {
		if customTools, ok2 := indexStats["custom_tools"].(int); ok2 {
			fmt.Printf("\nTools Distribution:\n")
			fmt.Printf("  Inherited Tools: %d\n", toolsInherited)
			fmt.Printf("  Custom Tools: %d\n", customTools)
		}
	}
}

// displayCacheStats displays cache performance statistics
func (c *IndexCommand) displayCacheStats(cacheStats map[string]interface{}, sharedCtx *SharedContext) {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	color.Blue("Cache Statistics\n")
	fmt.Println(strings.Repeat("=", 40))

	if hits, ok := cacheStats["hits"].(int); ok {
		fmt.Printf("Cache Hits: %d\n", hits)
	}

	if misses, ok := cacheStats["misses"].(int); ok {
		fmt.Printf("Cache Misses: %d\n", misses)
	}

	if size, ok := cacheStats["size"].(int); ok {
		fmt.Printf("Cache Size: %d entries\n", size)
	}

	// Calculate and display hit rate
	if hits, hok := cacheStats["hits"].(int); hok {
		if misses, mok := cacheStats["misses"].(int); mok {
			total := hits + misses
			if total > 0 {
				hitRate := float64(hits) / float64(total) * 100
				fmt.Printf("Hit Rate: %.1f%%\n", hitRate)

				// Performance assessment
				if hitRate >= 80.0 {
					PrintSuccess("Excellent cache performance")
				} else if hitRate >= 60.0 {
					PrintInfo("Good cache performance")
				} else if hitRate >= 40.0 {
					PrintWarning("Fair cache performance - consider clearing cache")
				} else {
					PrintWarning("Poor cache performance - recommend cache clear and rebuild")
				}
			}
		}
	}
}
