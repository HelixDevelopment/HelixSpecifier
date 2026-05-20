package adapters

import (
	"encoding/json"
	"fmt"

	"digital.vasic.helixspecifier/pkg/i18n"
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

// GenericAdapter provides CLI agent output formatting.
//
// The translator resolves operator-visible report content (the
// Markdown flow summary headings, status labels, phase rows) per
// CONST-046. A nil translator is treated as a NoopTranslator at
// resolution time, keeping NewGenericAdapter() callers backward
// compatible. CONST-051(B) decoupling preserved: no project-
// specific defaults baked in.
type GenericAdapter struct {
	agentType  string
	category   AgentCategory
	streaming  bool
	translator i18n.Translator
}

// NewGenericAdapter creates a new generic adapter with the default
// NoopTranslator. Production callers SHOULD prefer
// NewGenericAdapterWithTranslator to wire a real i18n backend.
func NewGenericAdapter(
	agentType string,
	category AgentCategory,
) *GenericAdapter {
	return &GenericAdapter{
		agentType:  agentType,
		category:   category,
		streaming:  true,
		translator: i18n.NewNoopTranslator(),
	}
}

// NewGenericAdapterWithTranslator creates a generic adapter with an
// injected translator for user-facing strings (CONST-046). Passing
// nil falls back to NoopTranslator, matching zero-value safety.
func NewGenericAdapterWithTranslator(
	agentType string,
	category AgentCategory,
	t i18n.Translator,
) *GenericAdapter {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	return &GenericAdapter{
		agentType:  agentType,
		category:   category,
		streaming:  true,
		translator: t,
	}
}

// SetTranslator swaps the adapter's i18n translator at runtime.
// Passing nil falls back to NoopTranslator so the adapter remains
// panic-free.
func (a *GenericAdapter) SetTranslator(t i18n.Translator) {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	a.translator = t
}

// tr returns the adapter's translator, defaulting to a
// NoopTranslator when none is wired.
func (a *GenericAdapter) tr() i18n.Translator {
	if a.translator == nil {
		return i18n.NewNoopTranslator()
	}
	return a.translator
}

// FormatOutput formats engine output for the agent
func (a *GenericAdapter) FormatOutput(
	result *types.FlowResult,
) (string, error) {
	if result == nil {
		return "", fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				a.tr(),
				"helixspecifier_adapter_err_nil_result",
				"nil result",
			),
		)
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
	t := a.tr()
	md := i18n.ResolveOrFallback(
		t,
		"helixspecifier_adapter_md_flow_heading",
		"# HelixSpecifier Flow: %s\n\n",
		result.FlowID,
	)
	md += i18n.ResolveOrFallback(
		t,
		"helixspecifier_adapter_md_effort_label",
		"**Effort:** %s | ",
		result.EffortLevel,
	)
	md += i18n.ResolveOrFallback(
		t,
		"helixspecifier_adapter_md_ceremony_label",
		"**Ceremony:** %s | ",
		result.CeremonyLevel,
	)
	md += i18n.ResolveOrFallback(
		t,
		"helixspecifier_adapter_md_quality_label",
		"**Quality:** %.2f\n\n",
		result.OverallQualityScore,
	)

	if result.Success {
		md += i18n.ResolveOrFallback(
			t,
			"helixspecifier_adapter_md_status_completed",
			"**Status:** Completed\n\n",
		)
	} else {
		md += i18n.ResolveOrFallback(
			t,
			"helixspecifier_adapter_md_status_failed",
			"**Status:** Failed\n\n",
		)
	}

	md += i18n.ResolveOrFallback(
		t,
		"helixspecifier_adapter_md_phases_heading",
		"## Phases\n\n",
	)
	for _, pr := range result.PhaseResults {
		status := "PASS"
		if !pr.Success {
			status = "FAIL"
		}
		md += i18n.ResolveOrFallback(
			t,
			"helixspecifier_adapter_md_phase_row",
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
