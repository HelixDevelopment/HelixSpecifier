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

// TestPillar_BuildPhaseTopic_DefaultResolvesToProse verifies the
// production default path (NewPillar, no translator injected) now
// resolves the topic key to its embedded bundle prose with the
// input payload interpolated — NOT the raw i18n key.
//
// §11.4.120 reconciliation (HXC-081): this test previously asserted
// the default topic CONTAINED the raw key
// `helixspecifier_speckit_topic_constitution` — that assertion
// codified the HXC-081 bug (NoopTranslator default leaking the raw
// key + producing `%!(EXTRA ...)`). The default is now the
// bundle-backed translator, so the correct invariant is: resolved
// prose, request + constitution interpolated, NO raw key.
func TestPillar_BuildPhaseTopic_DefaultResolvesToProse(t *testing.T) {
	p := newTestPillar()
	input := newTestInput()

	topic := p.buildPhaseTopic(types.PhaseConstitution, input)
	assert.NotContains(
		t, topic, "helixspecifier_speckit_topic_constitution",
		"raw i18n key leaked into default-resolved topic",
	)
	assert.NotContains(
		t, topic, "%!(EXTRA",
		"Go format-error marker present in default-resolved topic",
	)
	assert.Contains(t, topic, input.UserRequest)
	assert.Contains(t, topic, input.Constitution)
	assert.Contains(
		t, topic, "Constitution",
		"default topic is not the expected resolved prose",
	)
	assert.NotEmpty(t, topic, "default translator returned empty topic")
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
//
// §11.4.120 reconciliation (HXC-081): the default-path assertions
// (t1, t3) previously asserted the raw key
// `helixspecifier_speckit_topic_plan` — codifying the bug. The
// default + nil-fallback now resolve through the bundle-backed
// DefaultTranslator, so they assert resolved prose with the input
// interpolated and no raw key. The injected-translator assertion
// (t2, recordingTranslator stub) is unchanged — injection still
// routes through whatever backend is set.
func TestPillar_SetTranslator_Swap(t *testing.T) {
	p := newTestPillar()
	input := newTestInput()

	t1 := p.buildPhaseTopic(types.PhasePlan, input)
	assert.NotContains(t, t1, "helixspecifier_speckit_topic_plan")
	assert.NotContains(t, t1, "%!(EXTRA")
	assert.Contains(t, t1, input.PreviousOutput)

	rec := &recordingTranslator{}
	p.SetTranslator(rec)
	t2 := p.buildPhaseTopic(types.PhasePlan, input)
	assert.Contains(t, t2, "stub:helixspecifier_speckit_topic_plan")

	p.SetTranslator(nil)
	t3 := p.buildPhaseTopic(types.PhasePlan, input)
	assert.NotContains(t, t3, "helixspecifier_speckit_topic_plan")
	assert.NotContains(t, t3, "%!(EXTRA")
	assert.Contains(t, t3, input.PreviousOutput)
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
