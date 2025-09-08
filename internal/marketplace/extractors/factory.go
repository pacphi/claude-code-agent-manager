package extractors

import (
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/scripts"
	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/service"
)

// Factory creates extractors with their dependencies
type Factory struct {
	scriptLoader ScriptProvider
}

// NewFactory creates a new extractor factory
func NewFactory() *Factory {
	return &Factory{
		scriptLoader: scripts.NewScriptLoader(),
	}
}

// CreateExtractorSet creates a complete set of extractors
func (f *Factory) CreateExtractorSet() service.ExtractorSet {
	return service.ExtractorSet{
		Categories: NewCategoryExtractor(f.scriptLoader),
		Agents:     NewAgentExtractor(f.scriptLoader),
		Content:    NewContentExtractor(f.scriptLoader),
	}
}
