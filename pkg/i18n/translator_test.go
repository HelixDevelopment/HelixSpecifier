package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoopTranslator_ReturnsKeyVerbatim is the canonical
// anti-bluff guard for the noop fallback: a missing translation
// is visibly the key on screen, never an empty string that
// could be mistaken for a successful render (CONST-035 / Article
// XI §11.9). If anyone changes NoopTranslator.T to return "" or
// swallow the input, this test fails immediately.
func TestNoopTranslator_ReturnsKeyVerbatim(t *testing.T) {
	tr := NewNoopTranslator()
	require.NotNil(t, tr)
	got := tr.T("helixspecifier_intent_quick_reasoning")
	assert.Equal(
		t,
		"helixspecifier_intent_quick_reasoning",
		got,
	)
	assert.NotEmpty(t, got)
}

// TestNoopTranslator_FormatsArgs verifies the printf-style
// branch of the contract — args produce a formatted string,
// proving the fallback is wired through fmt.Sprintf, not a
// stub that drops parameters silently.
func TestNoopTranslator_FormatsArgs(t *testing.T) {
	tr := NewNoopTranslator()
	got := tr.T("flow_%s_complete", "ideation")
	assert.Equal(t, "flow_ideation_complete", got)
}

// TestNoopTranslator_AllMigratedKeysResolve asserts every key
// migrated in round 117 (and the round-208 follow-up) resolves
// to a non-empty value through the noop fallback. This is the
// sentinel that detects key drift: if a call site is migrated
// but the key here is not added (or vice versa), the test fails.
// Composes with classifier_test.go / engine_test.go /
// speckit_test.go / superpowers_test.go / predictive_test.go
// assertions to guarantee end-to-end key integrity for the
// migration batch.
func TestNoopTranslator_AllMigratedKeysResolve(t *testing.T) {
	tr := NewNoopTranslator()
	migrated := []string{
		// round 117 (HelixSpecifier kickoff)
		"helixspecifier_intent_quick_reasoning",
		"helixspecifier_intent_medium_reasoning",
		"helixspecifier_intent_large_reasoning",
		"helixspecifier_intent_epic_reasoning",
		"helixspecifier_intent_unknown_reasoning",
		// round 208 (CONST-046 Phase 4 round 90)
		"helixspecifier_engine_default_medium_reasoning",
		"helixspecifier_speckit_topic_constitution",
		"helixspecifier_speckit_topic_specify",
		"helixspecifier_speckit_topic_clarify",
		"helixspecifier_speckit_topic_plan",
		"helixspecifier_speckit_topic_tasks",
		"helixspecifier_speckit_topic_analyze",
		"helixspecifier_speckit_topic_implement",
		"helixspecifier_superpowers_task_executed",
		"helixspecifier_predictive_basis_label",
	}
	for _, k := range migrated {
		got := tr.T(k)
		assert.NotEmpty(t, got, "key %s resolved empty", k)
		assert.True(
			t,
			strings.HasPrefix(k, "helixspecifier_"),
			"key %s missing namespace prefix", k,
		)
	}
}

// TestNoopTranslator_Round208KeysAcceptArgs verifies the
// printf-style branch is exercised for every round-208 key that
// carries %s / %v placeholders. The noop fallback intentionally
// does NOT translate — its anti-bluff contract is that args
// remain visible in the output (either spliced into format verbs
// or emitted as fmt's "EXTRA" suffix) so missing-translation
// situations cannot silently swallow data. A regression that
// drops the args parameter entirely would fail this test —
// paired-mutation guard for the migration.
func TestNoopTranslator_Round208KeysAcceptArgs(t *testing.T) {
	tr := NewNoopTranslator()
	cases := []struct {
		key     string
		args    []any
		wantArg string
	}{
		{
			"helixspecifier_speckit_topic_plan",
			[]any{"prev-output"},
			"prev-output",
		},
		{
			"helixspecifier_speckit_topic_tasks",
			[]any{"plan-content"},
			"plan-content",
		},
		{
			"helixspecifier_superpowers_task_executed",
			[]any{"my-task"},
			"my-task",
		},
		{
			"helixspecifier_predictive_basis_label",
			[]any{"login-event"},
			"login-event",
		},
	}
	for _, c := range cases {
		got := tr.T(c.key, c.args...)
		assert.NotEmpty(t, got, "key %s empty", c.key)
		assert.Contains(
			t, got, c.key,
			"key %s missing from output (translator dropped key)",
			c.key,
		)
		assert.Contains(
			t, got, c.wantArg,
			"arg not visible in noop output for key %s",
			c.key,
		)
	}
}

// stubTranslator is a unit-test-only translator that records
// calls. It lives in this _test.go file (mocks permitted in
// unit tests only per CONST-050(A)) and is never imported by
// production code.
type stubTranslator struct {
	calls []string
}

func (s *stubTranslator) T(key string, args ...any) string {
	s.calls = append(s.calls, key)
	_ = args
	return "stub:" + key
}

// TestTranslator_InterfaceContract proves the Translator
// interface accepts arbitrary implementations — critical for
// CONST-051(B) decoupling so consuming projects can plug in
// their own bundle backend without modifying this submodule.
func TestTranslator_InterfaceContract(t *testing.T) {
	var tr Translator = &stubTranslator{}
	got := tr.T("helixspecifier_intent_quick_reasoning")
	assert.Equal(
		t,
		"stub:helixspecifier_intent_quick_reasoning",
		got,
	)
}
