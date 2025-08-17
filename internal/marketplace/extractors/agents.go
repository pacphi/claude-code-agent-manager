package extractors

import (
	"context"
	"fmt"

	"github.com/pacphi/claude-code-agent-manager/internal/marketplace/browser"
	"github.com/pacphi/claude-code-agent-manager/internal/types"
	"github.com/pacphi/claude-code-agent-manager/internal/util"
)

// agentExtractor implements AgentExtractor interface
type agentExtractor struct {
	scriptProvider ScriptProvider
}

// NewAgentExtractor creates a new agent extractor
func NewAgentExtractor(scriptProvider ScriptProvider) AgentExtractor {
	return &agentExtractor{
		scriptProvider: scriptProvider,
	}
}

// Extract extracts agents from a category page
func (e *agentExtractor) Extract(ctx context.Context, browser browser.Controller, category string) ([]types.Agent, error) {
	// Load the agents extraction script
	script, err := e.scriptProvider.LoadAgentsScript()
	if err != nil {
		return nil, fmt.Errorf("failed to load agents script: %w", err)
	}

	util.DebugPrintf("Executing agents extraction script for category: %s\n", category)

	// Execute the script
	result, err := browser.ExecuteScript(ctx, script)
	if err != nil {
		return nil, fmt.Errorf("failed to execute agents script: %w", err)
	}

	util.DebugPrintf("Script execution result type: %T\n", result)

	// If result is a map, let's see the debug info
	if resultMap, ok := result.(map[string]interface{}); ok {
		if debug, hasDebug := resultMap["debug"]; hasDebug {
			util.DebugPrintf("Debug info from script: %+v\n", debug)
		}
	}

	// Parse the results
	agents, err := e.parseAgents(result, category)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agents: %w", err)
	}

	util.DebugLogf("Successfully extracted %d agents for category %s", len(agents), category)
	return agents, nil
}

// parseAgents converts JavaScript result to Agent slice
func (e *agentExtractor) parseAgents(result interface{}, category string) ([]types.Agent, error) {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	// Extract agents array from the result
	agentsSlice, ok := resultMap["agents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no agents array found in result")
	}

	var agents []types.Agent
	for _, agentItem := range agentsSlice {
		agentMap, ok := agentItem.(map[string]interface{})
		if !ok {
			continue
		}

		agent := types.Agent{
			Name:        util.GetString(agentMap, "name"),
			Description: util.GetString(agentMap, "description"),
			Author:      util.GetString(agentMap, "author"),
			Rating:      util.GetFloat32(agentMap, "rating"),
			ContentURL:  util.GetString(agentMap, "url"),
			Category:    category,
		}

		if agent.Name != "" {
			agent.ID = util.GenerateSlug(agent.Name)
			agent.Slug = agent.ID
			agents = append(agents, agent)
		}
	}

	return agents, nil
}
