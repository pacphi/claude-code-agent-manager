package display

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace"
)

// Formatter handles output formatting for marketplace data
type Formatter struct {
	tabWriter *tabwriter.Writer
}

// NewFormatter creates a new display formatter
func NewFormatter() *Formatter {
	return &Formatter{
		tabWriter: tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0),
	}
}

// PrintCategories displays categories in a formatted table
func (f *Formatter) PrintCategories(categories []marketplace.Category) {
	if len(categories) == 0 {
		f.PrintWarning("No categories found")
		return
	}

	f.PrintHeader(fmt.Sprintf("Found %d categories:", len(categories)))

	// Create a new tabwriter with proper spacing for left alignment
	// Parameters: output, minwidth, tabwidth, padding, padchar, flags
	// Use 0 for flags (default is left alignment)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()

	// Print headers with consistent spacing
	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		color.HiCyanString("NAME"),
		color.HiCyanString("SLUG"),
		color.HiCyanString("AGENTS"),
		color.HiCyanString("DESCRIPTION"))

	for _, category := range categories {
		agentCount := color.YellowString("%d", category.AgentCount)
		description := f.truncateString(category.Description, 50)

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			category.Name,
			category.Slug,
			agentCount,
			description)
	}
}

// PrintAgents displays agents in a formatted table
func (f *Formatter) PrintAgents(agents []marketplace.Agent) {
	if len(agents) == 0 {
		f.PrintWarning("No agents found")
		return
	}

	f.PrintHeader(fmt.Sprintf("Found %d agents:", len(agents)))

	defer func() { _ = f.tabWriter.Flush() }()
	_, _ = fmt.Fprintf(f.tabWriter, "%s\t%s\t%s\n",
		color.HiCyanString("NAME"),
		color.HiCyanString("CATEGORY"),
		color.HiCyanString("DESCRIPTION"))

	for _, agent := range agents {
		description := f.truncateString(agent.Description, 50)
		category := f.truncateString(agent.Category, 15)

		_, _ = fmt.Fprintf(f.tabWriter, "%s\t%s\t%s\n",
			agent.Name,
			category,
			description)
	}
}

// PrintSearchResults displays search results with highlighted query
func (f *Formatter) PrintSearchResults(agents []marketplace.Agent, query string) {
	if len(agents) == 0 {
		f.PrintWarning(fmt.Sprintf("No agents found matching '%s'", query))
		return
	}

	f.PrintHeader(fmt.Sprintf("Found %d agents matching '%s':", len(agents), color.YellowString(query)))
	f.PrintAgents(agents)
}

// PrintAgentDetails displays detailed information about a single agent
func (f *Formatter) PrintAgentDetails(agent marketplace.Agent, content string) {
	fmt.Printf("\n%s\n", color.HiCyanString("Agent Details"))
	fmt.Printf("%s\n\n", strings.Repeat("=", 50))

	fmt.Printf("%s: %s\n", color.HiWhiteString("Name"), agent.Name)
	fmt.Printf("%s: %s\n", color.HiWhiteString("Slug"), agent.Slug)
	if agent.Author != "" {
		fmt.Printf("%s: %s\n", color.HiWhiteString("Author"), agent.Author)
	}
	if agent.Category != "" {
		fmt.Printf("%s: %s\n", color.HiWhiteString("Category"), agent.Category)
	}
	if agent.Rating > 0 {
		fmt.Printf("%s: %s\n", color.HiWhiteString("Rating"), f.formatRating(agent.Rating))
	}
	if agent.ContentURL != "" {
		fmt.Printf("%s: %s\n", color.HiWhiteString("URL"), agent.ContentURL)
	}

	fmt.Printf("\n%s:\n%s\n", color.HiWhiteString("Description"), agent.Description)

	if len(agent.Tags) > 0 {
		fmt.Printf("\n%s: %s\n", color.HiWhiteString("Tags"), strings.Join(agent.Tags, ", "))
	}

	if content != "" {
		fmt.Printf("\n%s:\n", color.HiWhiteString("Content"))
		fmt.Printf("%s\n", strings.Repeat("-", 50))
		fmt.Printf("%s\n", content)
	}

	fmt.Println()
}

// PrintSuccess displays a success message
func (f *Formatter) PrintSuccess(message string) {
	fmt.Printf("%s %s\n", color.GreenString("✓"), message)
}

// PrintError displays an error message
func (f *Formatter) PrintError(message string) {
	fmt.Printf("%s %s\n", color.RedString("✗"), message)
}

// PrintWarning displays a warning message
func (f *Formatter) PrintWarning(message string) {
	fmt.Printf("%s %s\n", color.YellowString("⚠"), message)
}

// PrintHeader displays a section header
func (f *Formatter) PrintHeader(message string) {
	fmt.Printf("\n%s\n", color.HiCyanString(message))
}

// formatRating formats a rating with stars
func (f *Formatter) formatRating(rating float32) string {
	if rating <= 0 {
		return color.HiBlackString("N/A")
	}

	stars := int(rating)
	starStr := strings.Repeat("⭐", stars)
	if rating > float32(stars) {
		starStr += "½"
	}

	return fmt.Sprintf("%s (%.1f)", starStr, rating)
}

// truncateString truncates a string to the specified length
func (f *Formatter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
