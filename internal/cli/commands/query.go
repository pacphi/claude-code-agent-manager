package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/query/engine"
	"github.com/pacphi/claude-code-agent-manager/internal/query/parser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// QueryCommand implements enhanced query functionality with regex and multi-field search
type QueryCommand struct {
	query       string
	field       string
	limit       int
	noTools     bool
	customTools bool
	source      string
	output      string
	useRegex    bool
	fuzzyScore  float64
	timeout     time.Duration
}

// NewQueryCommand creates a new query command instance
func NewQueryCommand() *QueryCommand {
	return &QueryCommand{
		fuzzyScore: 0.7,
		timeout:    30 * time.Second,
	}
}

// Name returns the command name
func (c *QueryCommand) Name() string {
	return "query"
}

// Description returns the command description
func (c *QueryCommand) Description() string {
	return "Search agents with complex queries"
}

// CreateCommand creates the cobra command for query functionality
func (c *QueryCommand) CreateCommand(sharedCtx *SharedContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [QUERY]",
		Short: c.Description(),
		Long: `Search agents using field-specific queries with support for regex patterns and fuzzy matching.

Examples:
  # Basic field searches
  agent-manager query "name:go"                 # Find agents with 'go' in name
  agent-manager query "tools:bash,git"          # Find agents using bash and git tools
  agent-manager query "description:automation"  # Find agents with automation in description

  # Regex pattern matching
  agent-manager query "name:^data.*processor$" --regex  # Regex pattern in name field
  agent-manager query "description:.*API.*" --regex     # Regex in description

  # Multi-field fuzzy search
  agent-manager query "database management" --fuzzy-score 0.6  # Lower threshold for broader matches

  # Complex filtering
  agent-manager query --no-tools                # Find agents with inherited tools only
  agent-manager query --custom-tools            # Find agents with explicit tools only
  agent-manager query --source github           # Find agents from github source
  agent-manager query --limit 10                # Limit results to 10 agents

  # Output formats
  agent-manager query "go" --output json        # JSON output
  agent-manager query "go" --output yaml        # YAML output`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				c.query = args[0]
			}
			return c.Execute(sharedCtx)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&c.field, "field", "f", "", "search specific field (name, description, content, tools, source)")
	cmd.Flags().IntVarP(&c.limit, "limit", "l", 0, "limit number of results")
	cmd.Flags().BoolVar(&c.noTools, "no-tools", false, "find agents with inherited tools only")
	cmd.Flags().BoolVar(&c.customTools, "custom-tools", false, "find agents with explicit tools only")
	cmd.Flags().StringVarP(&c.source, "source", "s", "", "filter by source")
	cmd.Flags().StringVarP(&c.output, "output", "o", "table", "output format (table, json, yaml)")
	cmd.Flags().BoolVar(&c.useRegex, "regex", false, "use regex pattern matching")
	cmd.Flags().Float64Var(&c.fuzzyScore, "fuzzy-score", 0.7, "fuzzy matching threshold (0.0-1.0)")
	cmd.Flags().DurationVar(&c.timeout, "timeout", 30*time.Second, "query timeout")

	return cmd
}

// Execute runs the query command logic
func (c *QueryCommand) Execute(sharedCtx *SharedContext) error {
	// Load configuration
	if err := sharedCtx.LoadConfig(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create query engine with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	queryEngine, err := sharedCtx.CreateQueryEngine()
	if err != nil {
		return err
	}

	// Configure fuzzy matching threshold
	queryEngine.SetFuzzyThreshold(c.fuzzyScore)

	// Execute search with progress indication
	var results []*parser.AgentSpec
	var queryErr error

	queryAction := "Searching agents"
	if c.query != "" {
		if c.useRegex {
			queryAction = fmt.Sprintf("Searching with regex pattern '%s'", c.query)
		} else {
			queryAction = fmt.Sprintf("Searching for '%s'", c.query)
		}
	}

	err = sharedCtx.PM.WithSpinner(queryAction, func() error {
		results, queryErr = c.executeQuery(ctx, queryEngine)
		return queryErr
	})

	if err != nil {
		return err
	}

	// Handle timeout
	select {
	case <-ctx.Done():
		return fmt.Errorf("query timed out after %v", c.timeout)
	default:
	}

	// Output results
	return c.outputResults(results, sharedCtx)
}

// executeQuery executes the appropriate query based on command options
func (c *QueryCommand) executeQuery(ctx context.Context, queryEngine *engine.Engine) ([]*parser.AgentSpec, error) {
	// Set up query options
	opts := engine.QueryOptions{
		Limit:       c.limit,
		NoTools:     c.noTools,
		CustomTools: c.customTools,
		Source:      c.source,
		Context:     ctx,
	}

	// Enable regex matching if requested
	opts.Regex = c.useRegex

	// Execute appropriate query type
	if c.field != "" && c.query != "" {
		return c.executeFieldQuery(queryEngine, opts)
	} else if c.query != "" {
		return c.executeComplexQuery(queryEngine, opts)
	} else {
		// Get all agents with filters applied
		allAgents := queryEngine.GetAllAgents()
		return c.applyFilters(allAgents, opts), nil
	}
}

// executeFieldQuery executes a field-specific query with optional regex support
func (c *QueryCommand) executeFieldQuery(queryEngine *engine.Engine, opts engine.QueryOptions) ([]*parser.AgentSpec, error) {
	if c.useRegex {
		return c.executeRegexFieldQuery(queryEngine, opts)
	}
	return queryEngine.QueryByField(c.field, c.query)
}

// executeRegexFieldQuery executes a field query with regex pattern matching
func (c *QueryCommand) executeRegexFieldQuery(queryEngine *engine.Engine, opts engine.QueryOptions) ([]*parser.AgentSpec, error) {
	// Compile regex pattern
	pattern, err := regexp.Compile(c.query)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Get all agents and filter by regex
	allAgents := queryEngine.GetAllAgents()
	var matches []*parser.AgentSpec

	for _, agent := range allAgents {
		var fieldValue string

		switch strings.ToLower(c.field) {
		case "name":
			fieldValue = agent.Name
		case "description":
			fieldValue = agent.Description
		case "content", "prompt":
			fieldValue = agent.Prompt
		case "tools":
			fieldValue = strings.Join(agent.Tools, " ")
		case "source":
			fieldValue = agent.Source
		default:
			return nil, fmt.Errorf("unsupported field for regex search: %s", c.field)
		}

		if pattern.MatchString(fieldValue) {
			matches = append(matches, agent)
		}
	}

	// Apply additional filters
	return c.applyFilters(matches, opts), nil
}

// executeComplexQuery executes a complex multi-field query
func (c *QueryCommand) executeComplexQuery(queryEngine *engine.Engine, opts engine.QueryOptions) ([]*parser.AgentSpec, error) {
	if c.useRegex {
		return c.executeRegexComplexQuery(queryEngine, opts)
	}

	// Use enhanced fuzzy matching for better relevance when no specific field is targeted
	if c.fuzzyScore < 0.7 || len(strings.Fields(c.query)) > 1 {
		return queryEngine.QueryWithFuzzy(c.query, opts)
	}

	return queryEngine.Query(c.query, opts)
}

// executeRegexComplexQuery executes a complex query with regex across multiple fields
func (c *QueryCommand) executeRegexComplexQuery(queryEngine *engine.Engine, opts engine.QueryOptions) ([]*parser.AgentSpec, error) {
	// Parse query for field-specific patterns
	patterns := c.parseComplexRegexQuery()

	allAgents := queryEngine.GetAllAgents()
	var matches []*parser.AgentSpec

	for _, agent := range allAgents {
		matched := true

		// Apply each pattern
		for field, pattern := range patterns {
			if !c.matchAgentField(agent, field, pattern) {
				matched = false
				break
			}
		}

		if matched {
			matches = append(matches, agent)
		}
	}

	return c.applyFilters(matches, opts), nil
}

// parseComplexRegexQuery parses a complex query string for field:pattern pairs
func (c *QueryCommand) parseComplexRegexQuery() map[string]*regexp.Regexp {
	patterns := make(map[string]*regexp.Regexp)

	// Check if query contains field:pattern format
	if strings.Contains(c.query, ":") {
		parts := strings.Fields(c.query)
		for _, part := range parts {
			if colonIndex := strings.Index(part, ":"); colonIndex > 0 {
				field := part[:colonIndex]
				pattern := part[colonIndex+1:]

				if regex, err := regexp.Compile(pattern); err == nil {
					patterns[strings.ToLower(field)] = regex
				}
			}
		}
	} else {
		// Single pattern applies to all searchable fields
		if regex, err := regexp.Compile(c.query); err == nil {
			patterns["name"] = regex
			patterns["description"] = regex
			patterns["content"] = regex
		}
	}

	return patterns
}

// matchAgentField checks if an agent field matches a regex pattern
func (c *QueryCommand) matchAgentField(agent *parser.AgentSpec, field string, pattern *regexp.Regexp) bool {
	var fieldValue string

	switch field {
	case "name":
		fieldValue = agent.Name
	case "description":
		fieldValue = agent.Description
	case "content", "prompt":
		fieldValue = agent.Prompt
	case "tools":
		fieldValue = strings.Join(agent.Tools, " ")
	case "source":
		fieldValue = agent.Source
	default:
		return false
	}

	return pattern.MatchString(fieldValue)
}

// applyFilters applies additional filters to results
func (c *QueryCommand) applyFilters(agents []*parser.AgentSpec, opts engine.QueryOptions) []*parser.AgentSpec {
	// Pre-allocate slice with estimated capacity to avoid reallocations
	filtered := make([]*parser.AgentSpec, 0, len(agents))

	for _, agent := range agents {
		// Apply source filter
		if opts.Source != "" && agent.Source != opts.Source {
			continue
		}

		// Apply tools filters
		if opts.NoTools && !agent.ToolsInherited {
			continue
		}
		if opts.CustomTools && agent.ToolsInherited {
			continue
		}

		// Apply date filter
		if !opts.After.IsZero() && agent.InstalledAt.Before(opts.After) {
			continue
		}

		filtered = append(filtered, agent)
	}

	// Apply limit
	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered
}

// outputResults outputs the query results in the specified format
func (c *QueryCommand) outputResults(results []*parser.AgentSpec, sharedCtx *SharedContext) error {
	if !sharedCtx.Options.Verbose && !sharedCtx.Options.NoProgress {
		fmt.Println() // Add spacing after spinner
	}

	title := fmt.Sprintf("Found %d agents", len(results))
	color.Blue("%s\n", title)

	if len(results) == 0 {
		PrintWarning("No agents found matching search criteria")
		return nil
	}

	switch c.output {
	case "json":
		return c.outputJSON(results)
	case "yaml":
		return c.outputYAML(results)
	case "table":
		fallthrough
	default:
		return c.outputTable(results)
	}
}

// outputJSON outputs results as JSON
func (c *QueryCommand) outputJSON(results []*parser.AgentSpec) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// outputYAML outputs results as YAML
func (c *QueryCommand) outputYAML(results []*parser.AgentSpec) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer func() {
		if err := encoder.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close YAML encoder: %v\n", err)
		}
	}()
	encoder.SetIndent(2)
	return encoder.Encode(results)
}

// outputTable outputs results as a formatted table
func (c *QueryCommand) outputTable(results []*parser.AgentSpec) error {
	// Print table header
	fmt.Printf("%-25s %-15s %-40s %-15s\n", "NAME", "SOURCE", "DESCRIPTION", "TOOLS")
	fmt.Println(strings.Repeat("-", 95))

	// Print each agent
	for _, agent := range results {
		name := c.truncate(agent.Name, 24)
		source := c.truncate(agent.Source, 14)
		description := c.truncate(agent.Description, 39)

		toolsStr := ""
		if agent.ToolsInherited {
			toolsStr = "inherited"
		} else if len(agent.Tools) > 0 {
			if len(agent.Tools) == 1 {
				toolsStr = agent.Tools[0]
			} else {
				toolsStr = fmt.Sprintf("%s (+%d)", agent.Tools[0], len(agent.Tools)-1)
			}
		}
		toolsStr = c.truncate(toolsStr, 14)

		fmt.Printf("%-25s %-15s %-40s %-15s\n", name, source, description, toolsStr)
	}

	return nil
}

// truncate truncates a string to the specified length with ellipsis
func (c *QueryCommand) truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
