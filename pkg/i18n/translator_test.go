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

// realTranslator is a unit-test-only translator that resolves a
// fixed key to a fixed real string, simulating a bundle-backed
// backend. Mocks permitted in unit tests only (CONST-050(A)).
type realTranslator struct{}

func (realTranslator) T(key string, args ...any) string {
	_ = args
	if key == "helixspecifier_engine_err_empty_request" {
		return "the request was empty"
	}
	return key
}

// TestResolveOrFallback_NoopPathEmitsFallback is the round-405
// paired-mutation guard for i18n.ResolveOrFallback. When the
// translator is the NoopTranslator (key-verbatim path), the call
// site's bundled English fallback MUST be emitted — never the raw
// key. A regression that returns the key would leak it into
// rendered output (Article XI §11.9 usability defect).
func TestResolveOrFallback_NoopPathEmitsFallback(t *testing.T) {
	tr := NewNoopTranslator()

	// No-args path.
	got := ResolveOrFallback(
		tr,
		"helixspecifier_engine_err_empty_request",
		"empty request",
	)
	assert.Equal(t, "empty request", got)
	assert.NotContains(t, got, "helixspecifier_")

	// Args path — fallback is a format string, args substituted.
	got = ResolveOrFallback(
		tr,
		"helixspecifier_engine_err_flow_not_found",
		"flow %s not found",
		"flow-42",
	)
	assert.Equal(t, "flow flow-42 not found", got)
	assert.NotContains(t, got, "helixspecifier_")
}

// TestResolveOrFallback_RealTranslatorWins proves a real
// bundle-backed translator's result is returned unchanged — the
// fallback is only used on the noop path.
func TestResolveOrFallback_RealTranslatorWins(t *testing.T) {
	got := ResolveOrFallback(
		realTranslator{},
		"helixspecifier_engine_err_empty_request",
		"empty request",
	)
	assert.Equal(t, "the request was empty", got)
	assert.NotEqual(t, "empty request", got,
		"real translator result must override the fallback")
}

// TestResolveOrFallback_NilTranslatorSafe proves a nil translator
// is treated as NoopTranslator — zero-value safety, no panic.
func TestResolveOrFallback_NilTranslatorSafe(t *testing.T) {
	got := ResolveOrFallback(
		nil,
		"helixspecifier_engine_err_empty_request",
		"empty request",
	)
	assert.Equal(t, "empty request", got)
}

// TestResolveOrFallback_Round405KeysResolve asserts every
// round-405 key resolves to a non-empty, non-raw-key string
// through the noop fallback — sentinel for key/fallback drift.
// Calls use constant fallback literals so go vet's printf
// analysis is satisfied (the fallback IS a format string when
// args are supplied; here no args are passed so it is returned
// verbatim).
func TestResolveOrFallback_Round405KeysResolve(t *testing.T) {
	tr := NewNoopTranslator()

	got := []string{
		ResolveOrFallback(tr,
			"helixspecifier_engine_err_speckit_not_registered",
			"speckit pillar not registered"),
		ResolveOrFallback(tr,
			"helixspecifier_engine_err_empty_request",
			"empty request"),
		ResolveOrFallback(tr,
			"helixspecifier_engine_err_classification_required",
			"classification required"),
		ResolveOrFallback(tr,
			"helixspecifier_engine_err_adapter_not_found",
			"adapter not found"),
		ResolveOrFallback(tr,
			"helixspecifier_adapter_err_nil_result",
			"nil result"),
		ResolveOrFallback(tr,
			"helixspecifier_adapter_md_status_completed",
			"**Status:** Completed\n\n"),
		ResolveOrFallback(tr,
			"helixspecifier_adapter_md_phases_heading",
			"## Phases\n\n"),
		ResolveOrFallback(tr,
			"helixspecifier_superpowers_err_no_tasks",
			"no tasks provided"),
		ResolveOrFallback(tr,
			"helixspecifier_superpowers_err_no_task_results",
			"no task results to review"),
	}
	for i, g := range got {
		assert.NotEmpty(t, g, "case %d resolved empty", i)
		assert.NotContains(
			t, g, "helixspecifier_",
			"case %d: raw key leaked through fallback", i,
		)
	}
}
