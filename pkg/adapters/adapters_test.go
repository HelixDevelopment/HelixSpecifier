package adapters

import (
	"strings"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGenericAdapter(t *testing.T) {
	a := NewGenericAdapter("test", CategoryMarkdown)
	require.NotNil(t, a)
	assert.Equal(t, "test", a.agentType)
	assert.Equal(t, CategoryMarkdown, a.category)
	assert.True(t, a.streaming)
}

func TestGenericAdapter_GetAgentType(t *testing.T) {
	a := NewGenericAdapter("opencode", CategoryMarkdown)
	assert.Equal(t, "opencode", a.GetAgentType())
}

func TestGenericAdapter_SupportsStreaming(t *testing.T) {
	a := NewGenericAdapter("crush", CategoryMarkdown)
	assert.True(t, a.SupportsStreaming())
}

func TestGenericAdapter_ParseInput(t *testing.T) {
	a := NewGenericAdapter("test", CategoryGeneric)
	result, err := a.ParseInput("hello world")
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestGenericAdapter_FormatOutput_Nil(t *testing.T) {
	a := NewGenericAdapter("test", CategoryMarkdown)
	_, err := a.FormatOutput(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGenericAdapter_FormatOutput_Markdown(t *testing.T) {
	a := NewGenericAdapter("opencode", CategoryMarkdown)
	result := &types.FlowResult{
		FlowID:              "hsf-abc123",
		EffortLevel:         types.EffortLarge,
		CeremonyLevel:       types.CeremonyStandard,
		Success:             true,
		OverallQualityScore: 0.85,
		PhaseResults: []types.PhaseResult{
			{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.9,
				Duration:     2 * time.Second,
			},
		},
	}

	output, err := a.FormatOutput(result)
	require.NoError(t, err)
	assert.Contains(t, output, "# HelixSpecifier Flow")
	assert.Contains(t, output, "hsf-abc123")
	assert.Contains(t, output, "Completed")
	assert.Contains(t, output, "PASS")
}

func TestGenericAdapter_FormatOutput_TOML(t *testing.T) {
	a := NewGenericAdapter("toml-agent", CategoryTOML)
	result := &types.FlowResult{
		FlowID:              "hsf-toml1",
		EffortLevel:         types.EffortMedium,
		CeremonyLevel:       types.CeremonyLight,
		Success:             true,
		OverallQualityScore: 0.75,
	}

	output, err := a.FormatOutput(result)
	require.NoError(t, err)
	assert.Contains(t, output, "[flow]")
	assert.Contains(t, output, "hsf-toml1")
	assert.Contains(t, output, "success = true")
}

func TestGenericAdapter_FormatOutput_JSON(t *testing.T) {
	a := NewGenericAdapter("gemini", CategoryGeneric)
	result := &types.FlowResult{
		FlowID:        "hsf-json1",
		EffortLevel:   types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
		Success:       true,
	}

	output, err := a.FormatOutput(result)
	require.NoError(t, err)
	assert.Contains(t, output, "hsf-json1")
	assert.Contains(t, output, "\"success\": true")
}

func TestGenericAdapter_ImplementsInterface(t *testing.T) {
	a := NewGenericAdapter("test", CategoryGeneric)
	// Compile-time check
	var _ types.CLIAgentAdapter = a
}

func TestWellKnownAdapters_Count(t *testing.T) {
	adapters := WellKnownAdapters()
	assert.Len(t, adapters, 10)
}

func TestWellKnownAdapters_HasOpenCode(t *testing.T) {
	adapters := WellKnownAdapters()
	oc, ok := adapters["opencode"]
	require.True(t, ok)
	assert.Equal(t, "opencode", oc.GetAgentType())
}

func TestWellKnownAdapters_HasCrush(t *testing.T) {
	adapters := WellKnownAdapters()
	cr, ok := adapters["crush"]
	require.True(t, ok)
	assert.Equal(t, "crush", cr.GetAgentType())
}

func TestWellKnownAdapters_Categories(t *testing.T) {
	adapters := WellKnownAdapters()

	// Markdown agents
	mdAgents := []string{
		"opencode", "crush", "claude",
		"cursor", "windsurf", "cline", "roo",
	}
	for _, name := range mdAgents {
		a, ok := adapters[name]
		require.True(t, ok, "missing %s", name)
		assert.Equal(t, CategoryMarkdown, a.category,
			"%s should be markdown", name)
	}

	// GitHub agent
	assert.Equal(t, CategoryGitHub,
		adapters["copilot"].category)

	// Generic agents
	assert.Equal(t, CategoryGeneric,
		adapters["gemini"].category)
	assert.Equal(t, CategoryGeneric,
		adapters["aider"].category)
}

func TestGenericAdapter_FormatMarkdown_Success(
	t *testing.T,
) {
	a := NewGenericAdapter("test", CategoryMarkdown)
	result := &types.FlowResult{
		FlowID:              "flow-ok",
		Success:             true,
		EffortLevel:         types.EffortLarge,
		CeremonyLevel:       types.CeremonyStandard,
		OverallQualityScore: 0.92,
		PhaseResults: []types.PhaseResult{
			{
				Phase:        types.PhaseConstitution,
				Success:      true,
				QualityScore: 0.95,
			},
			{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.89,
			},
		},
	}

	output, err := a.FormatOutput(result)
	require.NoError(t, err)
	assert.Contains(t, output, "Completed")
	// Both phases should appear
	assert.Equal(t, 2,
		strings.Count(output, "PASS"))
}

func TestGenericAdapter_FormatMarkdown_Failed(
	t *testing.T,
) {
	a := NewGenericAdapter("test", CategoryMarkdown)
	result := &types.FlowResult{
		FlowID:        "flow-fail",
		Success:       false,
		EffortLevel:   types.EffortMedium,
		CeremonyLevel: types.CeremonyLight,
		PhaseResults: []types.PhaseResult{
			{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.8,
			},
			{
				Phase:        types.PhasePlan,
				Success:      false,
				QualityScore: 0.3,
			},
		},
	}

	output, err := a.FormatOutput(result)
	require.NoError(t, err)
	assert.Contains(t, output, "Failed")
	assert.Contains(t, output, "PASS")
	assert.Contains(t, output, "FAIL")
}
