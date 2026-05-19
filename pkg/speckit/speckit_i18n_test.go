package speckit

import (
	"strings"
	"testing"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingTranslator captures every key passed to T. Lives in
// _test.go only (CONST-050(A)) — never imported by production.
type recordingTranslator struct {
	calls []string
}

func (r *recordingTranslator) T(key string, args ...any) string {
	r.calls = append(r.calls, key)
	_ = args
	return "stub:" + key
}

// TestPillar_BuildPhaseTopic_AllPhasesUseI18nKeys exercises every
// migrated SpecKit phase topic and proves the translator is the
// resolution path — NOT the old hardcoded English literals. A
// regression that reinstates "Analyze the request and create/
// update the project Constitution", "Break down plan into tasks",
// etc. would fail this test (paired-mutation guard).
func TestPillar_BuildPhaseTopic_AllPhasesUseI18nKeys(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	rec := &recordingTranslator{}
	p := NewPillarWithTranslator(cfg, logger, rec)
	require.NotNil(t, p)

	input := newTestInput()

	cases := []struct {
		phase   types.SpecKitPhase
		key     string
		oldLit  string
	}{
		{
			types.PhaseConstitution,
			"helixspecifier_speckit_topic_constitution",
			"Analyze the request and create/update the",
		},
		{
			types.PhaseSpecify,
			"helixspecifier_speckit_topic_specify",
			"Create a detailed specification based on",
		},
		{
			types.PhaseClarify,
			"helixspecifier_speckit_topic_clarify",
			"Clarify and validate the specification",
		},
		{
			types.PhasePlan,
			"helixspecifier_speckit_topic_plan",
			"Create implementation plan",
		},
		{
			types.PhaseTasks,
			"helixspecifier_speckit_topic_tasks",
			"Break down plan into tasks",
		},
		{
			types.PhaseAnalyze,
			"helixspecifier_speckit_topic_analyze",
			"Analyze tasks before implementation",
		},
		{
			types.PhaseImplement,
			"helixspecifier_speckit_topic_implement",
			"Execute implementation",
		},
	}

	for _, c := range cases {
		topic := p.buildPhaseTopic(c.phase, input)
		assert.Contains(
			t, topic, "stub:"+c.key,
			"phase %s did not route through translator",
			c.phase,
		)
		assert.NotContains(
			t, topic, c.oldLit,
			"phase %s regressed to hardcoded literal %q",
			c.phase, c.oldLit,
		)
		// Confirm the translator was actually called with
		// the canonical key.
		assert.Contains(
			t, rec.calls, c.key,
			"phase %s did not invoke key %s", c.phase, c.key,
		)
	}
}

// TestPillar_BuildPhaseTopic_NoopFallbackVisible verifies the
// NoopTranslator path keeps the migration safe: with the default
// constructor (no translator injected) the topic still contains
// the canonical key plus the input payload, so a missing-bundle
// situation is visibly debuggable rather than silently empty.
func TestPillar_BuildPhaseTopic_NoopFallbackVisible(t *testing.T) {
	p := newTestPillar()
	input := newTestInput()

	topic := p.buildPhaseTopic(types.PhaseConstitution, input)
	assert.Contains(
		t, topic, "helixspecifier_speckit_topic_constitution",
	)
	assert.Contains(t, topic, input.UserRequest)
	assert.Contains(t, topic, input.Constitution)
	assert.NotEmpty(t, topic, "noop translator returned empty topic")
}

// TestPillar_NewPillarWithTranslator_NilFallback proves the
// constructor accepts nil and falls back to NoopTranslator.
func TestPillar_NewPillarWithTranslator_NilFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	p := NewPillarWithTranslator(cfg, logger, nil)
	require.NotNil(t, p)
	require.NotNil(t, p.translator)
}

// TestPillar_SetTranslator_Swap exercises runtime swap including
// nil-fallback. Confirms no panic and subsequent topic builds use
// the new backend.
func TestPillar_SetTranslator_Swap(t *testing.T) {
	p := newTestPillar()
	input := newTestInput()

	t1 := p.buildPhaseTopic(types.PhasePlan, input)
	assert.Contains(t, t1, "helixspecifier_speckit_topic_plan")

	rec := &recordingTranslator{}
	p.SetTranslator(rec)
	t2 := p.buildPhaseTopic(types.PhasePlan, input)
	assert.Contains(t, t2, "stub:helixspecifier_speckit_topic_plan")

	p.SetTranslator(nil)
	t3 := p.buildPhaseTopic(types.PhasePlan, input)
	assert.Contains(t, t3, "helixspecifier_speckit_topic_plan")
	assert.NotContains(t, t3, "stub:")
}

// TestPillar_BuildPhaseTopic_DefaultUnknownPhase confirms the
// default branch — which intentionally keeps a Sprintf format
// (no user-facing English) — still embeds both the phase name
// and request for diagnostics.
func TestPillar_BuildPhaseTopic_DefaultUnknownPhase(t *testing.T) {
	p := newTestPillar()
	input := newTestInput()
	topic := p.buildPhaseTopic(
		types.SpecKitPhase("unknown-phase-x"), input,
	)
	assert.True(
		t, strings.Contains(topic, "unknown-phase-x"),
		"unknown phase name missing from fallback topic",
	)
	assert.Contains(t, topic, input.UserRequest)
}

// Sentinel that the i18n package is referenced.
func TestSpecKit_I18nPackagePresent(t *testing.T) {
	tr := i18n.NewNoopTranslator()
	assert.NotNil(t, tr)
}
