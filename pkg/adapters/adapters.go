package adapters

import (
	"encoding/json"
	"fmt"

	"digital.vasic.helixspecifier/pkg/types"
)

// AgentCategory represents a category of CLI agents
type AgentCategory string

const (
	// CategoryMarkdown outputs markdown formatted results
	CategoryMarkdown AgentCategory = "markdown"
	// CategoryTOML outputs TOML formatted results
	CategoryTOML AgentCategory = "toml"
	// CategoryGitHub outputs GitHub-flavored results
	CategoryGitHub AgentCategory = "github"
	// CategoryMinimal outputs minimal results
	CategoryMinimal AgentCategory = "minimal"
	// CategoryGeneric outputs JSON formatted results
	CategoryGeneric AgentCategory = "generic"
)

// GenericAdapter provides CLI agent output formatting
type GenericAdapter struct {
	agentType string
	category  AgentCategory
	streaming bool
}

// NewGenericAdapter creates a new generic adapter
func NewGenericAdapter(
	agentType string,
	category AgentCategory,
) *GenericAdapter {
	return &GenericAdapter{
		agentType: agentType,
		category:  category,
		streaming: true,
	}
}

// FormatOutput formats engine output for the agent
func (a *GenericAdapter) FormatOutput(
	result *types.FlowResult,
) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil result")
	}

	switch a.category {
	case CategoryMarkdown:
		return a.formatMarkdown(result)
	case CategoryTOML:
		return a.formatTOML(result)
	default:
		return a.formatJSON(result)
	}
}

// ParseInput parses agent-specific input
func (a *GenericAdapter) ParseInput(
	input string,
) (string, error) {
	return input, nil
}

// GetAgentType returns the agent type
func (a *GenericAdapter) GetAgentType() string {
	return a.agentType
}

// SupportsStreaming returns streaming support
func (a *GenericAdapter) SupportsStreaming() bool {
	return a.streaming
}

// formatMarkdown formats as markdown
func (a *GenericAdapter) formatMarkdown(
	result *types.FlowResult,
) (string, error) {
	md := fmt.Sprintf(
		"# HelixSpecifier Flow: %s\n\n",
		result.FlowID,
	)
	md += fmt.Sprintf(
		"**Effort:** %s | ", result.EffortLevel,
	)
	md += fmt.Sprintf(
		"**Ceremony:** %s | ", result.CeremonyLevel,
	)
	md += fmt.Sprintf(
		"**Quality:** %.2f\n\n",
		result.OverallQualityScore,
	)

	if result.Success {
		md += "**Status:** Completed\n\n"
	} else {
		md += "**Status:** Failed\n\n"
	}

	md += "## Phases\n\n"
	for _, pr := range result.PhaseResults {
		status := "PASS"
		if !pr.Success {
			status = "FAIL"
		}
		md += fmt.Sprintf(
			"- **%s** [%s] (%.2f) %s\n",
			pr.Phase, status,
			pr.QualityScore, pr.Duration,
		)
	}

	return md, nil
}

// formatTOML formats as TOML-like
func (a *GenericAdapter) formatTOML(
	result *types.FlowResult,
) (string, error) {
	toml := fmt.Sprintf(
		"[flow]\nid = %q\n", result.FlowID,
	)
	toml += fmt.Sprintf(
		"effort = %q\n", result.EffortLevel,
	)
	toml += fmt.Sprintf(
		"ceremony = %q\n", result.CeremonyLevel,
	)
	toml += fmt.Sprintf(
		"quality = %.2f\n", result.OverallQualityScore,
	)
	toml += fmt.Sprintf(
		"success = %v\n", result.Success,
	)
	return toml, nil
}

// formatJSON formats as JSON
func (a *GenericAdapter) formatJSON(
	result *types.FlowResult,
) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	return string(data), nil
}

// WellKnownAdapters returns adapters for known agent types
func WellKnownAdapters() map[string]*GenericAdapter {
	return map[string]*GenericAdapter{
		"opencode": NewGenericAdapter(
			"opencode", CategoryMarkdown,
		),
		"crush": NewGenericAdapter(
			"crush", CategoryMarkdown,
		),
		"claude": NewGenericAdapter(
			"claude", CategoryMarkdown,
		),
		"cursor": NewGenericAdapter(
			"cursor", CategoryMarkdown,
		),
		"windsurf": NewGenericAdapter(
			"windsurf", CategoryMarkdown,
		),
		"copilot": NewGenericAdapter(
			"copilot", CategoryGitHub,
		),
		"gemini": NewGenericAdapter(
			"gemini", CategoryGeneric,
		),
		"aider": NewGenericAdapter(
			"aider", CategoryGeneric,
		),
		"cline": NewGenericAdapter(
			"cline", CategoryMarkdown,
		),
		"roo": NewGenericAdapter(
			"roo", CategoryMarkdown,
		),
	}
}
