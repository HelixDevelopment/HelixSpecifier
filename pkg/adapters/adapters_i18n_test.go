package adapters

import (
	"strings"
	"testing"

	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingTranslator captures every key passed to T. It is a
// unit-test-only fake (CONST-050(A)) that lives in this _test.go
// file and is never imported by production code.
type recordingTranslator struct {
	calls []string
}

func (r *recordingTranslator) T(key string, args ...any) string {
	r.calls = append(r.calls, key)
	_ = args
	return "stub:" + key
}

// TestGenericAdapter_Round405_MarkdownUsesI18nKeys is the
// paired-mutation guard for the round-405 CONST-046 migration of
// the GenericAdapter Markdown report-render content. It exercises
// formatMarkdown against a recording translator and asserts every
// migrated key is invoked and no old hardcoded literal survives.
func TestGenericAdapter_Round405_MarkdownUsesI18nKeys(
	t *testing.T,
) {
	rec := &recordingTranslator{}
	a := NewGenericAdapterWithTranslator(
		"claude", CategoryMarkdown, rec,
	)
	require.NotNil(t, a)

	result := &types.FlowResult{
		FlowID:              "hsf-abc123",
		EffortLevel:         types.EffortMedium,
		CeremonyLevel:       types.CeremonyStandard,
		OverallQualityScore: 0.91,
		Success:             true,
		PhaseResults: []types.PhaseResult{
			{
				Phase:        types.PhasePlan,
				Success:      true,
				QualityScore: 0.88,
			},
		},
	}

	out, err := a.FormatOutput(result)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	wantKeys := []string{
		"helixspecifier_adapter_md_flow_heading",
		"helixspecifier_adapter_md_effort_label",
		"helixspecifier_adapter_md_ceremony_label",
		"helixspecifier_adapter_md_quality_label",
		"helixspecifier_adapter_md_status_completed",
		"helixspecifier_adapter_md_phases_heading",
		"helixspecifier_adapter_md_phase_row",
	}
	for _, k := range wantKeys {
		assert.Contains(t, rec.calls, k,
			"i18n key %s not invoked by formatMarkdown", k)
	}

	// Old hardcoded literals must NOT appear in the output.
	for _, literal := range []string{
		"# HelixSpecifier Flow:",
		"**Effort:**",
		"## Phases",
	} {
		assert.NotContains(t, out, literal,
			"regression: hardcoded literal %q reintroduced",
			literal)
	}
}

// TestGenericAdapter_Round405_FailedStatusKey covers the
// status-failed branch of formatMarkdown.
func TestGenericAdapter_Round405_FailedStatusKey(t *testing.T) {
	rec := &recordingTranslator{}
	a := NewGenericAdapterWithTranslator(
		"crush", CategoryMarkdown, rec,
	)
	result := &types.FlowResult{
		FlowID:  "hsf-fail",
		Success: false,
	}
	_, err := a.FormatOutput(result)
	require.NoError(t, err)
	assert.Contains(
		t, rec.calls,
		"helixspecifier_adapter_md_status_failed",
	)
}

// TestGenericAdapter_Round405_NilResultUsesI18nKey covers the
// nil-result validation error.
func TestGenericAdapter_Round405_NilResultUsesI18nKey(
	t *testing.T,
) {
	rec := &recordingTranslator{}
	a := NewGenericAdapterWithTranslator(
		"claude", CategoryMarkdown, rec,
	)
	_, err := a.FormatOutput(nil)
	require.Error(t, err)
	assert.Contains(
		t, rec.calls,
		"helixspecifier_adapter_err_nil_result",
	)
	assert.NotContains(t, err.Error(), "nil result")
}

// TestGenericAdapter_Round405_NoopFallbackEmitsEnglish proves the
// NoopTranslator path (via i18n.ResolveOrFallback) emits the
// bundled English fallback Markdown — NOT raw i18n keys — so the
// real adapter still renders coherent Markdown when no translator
// is wired. A raw key leaking into rendered Markdown is itself a
// usability defect (Article XI §11.9 — "users can use the
// feature" is the bar, not "tests pass").
func TestGenericAdapter_Round405_NoopFallbackEmitsEnglish(
	t *testing.T,
) {
	a := NewGenericAdapter("claude", CategoryMarkdown)
	result := &types.FlowResult{
		FlowID:  "hsf-noop",
		Success: true,
		PhaseResults: []types.PhaseResult{
			{Phase: types.PhasePlan, Success: true},
		},
	}
	out, err := a.FormatOutput(result)
	require.NoError(t, err)

	// Coherent rendered Markdown, not raw keys.
	assert.Contains(t, out, "# HelixSpecifier Flow: hsf-noop")
	assert.Contains(t, out, "**Status:** Completed")
	assert.Contains(t, out, "## Phases")
	assert.True(t, strings.Contains(out, "- **"),
		"phase rows must render as Markdown list items")
	assert.NotContains(
		t, out, "helixspecifier_adapter_md_",
		"raw i18n key leaked into rendered Markdown",
	)
}

// TestGenericAdapter_Round405_NilTranslatorFallback proves the
// constructor + setter accept nil and fall back to NoopTranslator
// without panicking — zero-value safety contract.
func TestGenericAdapter_Round405_NilTranslatorFallback(
	t *testing.T,
) {
	a := NewGenericAdapterWithTranslator(
		"claude", CategoryMarkdown, nil,
	)
	require.NotNil(t, a)
	require.NotNil(t, a.translator)

	a.SetTranslator(nil)
	require.NotNil(t, a.translator)

	result := &types.FlowResult{FlowID: "hsf-x", Success: true}
	_, err := a.FormatOutput(result)
	require.NoError(t, err)
}

// TestAdapters_I18nPackagePresent is a sentinel that the i18n
// package is referenced by this test file.
func TestAdapters_I18nPackagePresent(t *testing.T) {
	tr := i18n.NewNoopTranslator()
	assert.NotNil(t, tr)
}
