package engine

import (
	"context"
	"testing"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/i18n"
	"github.com/sirupsen/logrus"
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

// TestFusionEngine_ClassifyEffort_DefaultReasoningUsesI18nKey
// asserts the round-208 migration: the default-fallback Reasoning
// is resolved through the i18n translator, not a hardcoded English
// literal. A regression that reinstates the old "Default medium
// effort classification" literal would fail this test —
// paired-mutation guard for CONST-046 compliance.
func TestFusionEngine_ClassifyEffort_DefaultReasoningUsesI18nKey(
	t *testing.T,
) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	rec := &recordingTranslator{}
	e := NewWithTranslator(cfg, logger, rec)
	require.NotNil(t, e)

	ctx := context.Background()
	cl, err := e.ClassifyEffort(ctx, "do something")
	require.NoError(t, err)
	require.NotNil(t, cl)

	// Reasoning should be the stub-prefixed key, NOT the old
	// hardcoded English string.
	assert.Equal(
		t,
		"stub:helixspecifier_engine_default_medium_reasoning",
		cl.Reasoning,
	)
	assert.Contains(
		t, rec.calls,
		"helixspecifier_engine_default_medium_reasoning",
	)
	assert.NotEqual(
		t,
		"Default medium effort classification",
		cl.Reasoning,
		"regression: hardcoded literal reintroduced",
	)
}

// TestFusionEngine_NewWithTranslator_NilFallback proves the
// constructor accepts nil and falls back to NoopTranslator
// without panicking — anti-bluff zero-value safety contract.
func TestFusionEngine_NewWithTranslator_NilFallback(t *testing.T) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	e := NewWithTranslator(cfg, logger, nil)
	require.NotNil(t, e)
	require.NotNil(t, e.translator)

	// Reasoning should be the verbatim key (NoopTranslator).
	ctx := context.Background()
	cl, err := e.ClassifyEffort(ctx, "do something")
	require.NoError(t, err)
	assert.Equal(
		t,
		"helixspecifier_engine_default_medium_reasoning",
		cl.Reasoning,
	)
}

// TestFusionEngine_SetTranslator_Swap exercises the runtime
// translator-swap path and confirms subsequent classifications
// pick up the new backend.
func TestFusionEngine_SetTranslator_Swap(t *testing.T) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	e := New(cfg, logger)

	// Initial: NoopTranslator returns the key verbatim.
	ctx := context.Background()
	cl, err := e.ClassifyEffort(ctx, "do something")
	require.NoError(t, err)
	assert.Equal(
		t,
		"helixspecifier_engine_default_medium_reasoning",
		cl.Reasoning,
	)

	// Swap to recording stub.
	rec := &recordingTranslator{}
	e.SetTranslator(rec)
	cl2, err := e.ClassifyEffort(ctx, "do something else")
	require.NoError(t, err)
	assert.Equal(
		t,
		"stub:helixspecifier_engine_default_medium_reasoning",
		cl2.Reasoning,
	)

	// Nil should fall back to noop.
	e.SetTranslator(nil)
	cl3, err := e.ClassifyEffort(ctx, "third request")
	require.NoError(t, err)
	assert.Equal(
		t,
		"helixspecifier_engine_default_medium_reasoning",
		cl3.Reasoning,
	)
}

// Sentinel assertion that the i18n package is referenced by this
// test file — guards against accidental import drop while keeping
// the test fixture truthful about its dependencies.
func TestEngine_I18nPackagePresent(t *testing.T) {
	tr := i18n.NewNoopTranslator()
	assert.NotNil(t, tr)
}

// TestFusionEngine_Round405_ValidationErrorsUseI18nKeys is the
// paired-mutation guard for the round-405 CONST-046 migration of
// FusionEngine validation errors. Each error path is exercised
// against a recording translator; the assertions confirm (a) the
// i18n key was invoked and (b) the old hardcoded English literal
// is NOT present. A regression that reinstates any literal fails
// this test.
func TestFusionEngine_Round405_ValidationErrorsUseI18nKeys(
	t *testing.T,
) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	ctx := context.Background()

	// Health: speckit pillar not registered (speckit is nil).
	rec := &recordingTranslator{}
	e := NewWithTranslator(cfg, logger, rec)
	err := e.Health(ctx)
	require.Error(t, err)
	assert.Contains(
		t, rec.calls,
		"helixspecifier_engine_err_speckit_not_registered",
	)
	assert.Equal(
		t,
		"stub:helixspecifier_engine_err_speckit_not_registered",
		err.Error(),
	)
	assert.NotContains(
		t, err.Error(), "speckit pillar not registered",
		"regression: hardcoded literal reintroduced",
	)

	// ClassifyEffort: empty request.
	rec2 := &recordingTranslator{}
	e2 := NewWithTranslator(cfg, logger, rec2)
	_, err = e2.ClassifyEffort(ctx, "")
	require.Error(t, err)
	assert.Contains(
		t, rec2.calls,
		"helixspecifier_engine_err_empty_request",
	)
	assert.NotContains(t, err.Error(), "empty request")

	// ExecuteFlow: nil classification.
	rec3 := &recordingTranslator{}
	e3 := NewWithTranslator(cfg, logger, rec3)
	_, err = e3.ExecuteFlow(ctx, "req", nil)
	require.Error(t, err)
	assert.Contains(
		t, rec3.calls,
		"helixspecifier_engine_err_classification_required",
	)
	assert.NotContains(t, err.Error(), "classification required")

	// GetFlowStatus: flow not found.
	rec4 := &recordingTranslator{}
	e4 := NewWithTranslator(cfg, logger, rec4)
	_, err = e4.GetFlowStatus("missing-flow")
	require.Error(t, err)
	assert.Contains(
		t, rec4.calls,
		"helixspecifier_engine_err_flow_not_found",
	)
	assert.NotContains(t, err.Error(), "flow missing-flow not found")

	// ResumeFlow: flow not found.
	rec5 := &recordingTranslator{}
	e5 := NewWithTranslator(cfg, logger, rec5)
	_, err = e5.ResumeFlow(ctx, "missing-flow", "req")
	require.Error(t, err)
	assert.Contains(
		t, rec5.calls,
		"helixspecifier_engine_err_flow_not_found",
	)

	// GetAdapter: adapter not found.
	rec6 := &recordingTranslator{}
	e6 := NewWithTranslator(cfg, logger, rec6)
	_, err = e6.GetAdapter("missing-adapter")
	require.Error(t, err)
	assert.Contains(
		t, rec6.calls,
		"helixspecifier_engine_err_adapter_not_found",
	)
	assert.NotContains(t, err.Error(), "adapter missing-adapter")
}

// TestFusionEngine_Round405_NoopFallbackEmitsEnglish proves the
// NoopTranslator path (via i18n.ResolveOrFallback) emits the
// bundled English fallback — NOT a raw i18n key — so the real
// engine still produces a coherent message when no translator is
// wired. A raw key leaking into an error is itself a usability
// defect (Article XI §11.9 — "users can use the feature").
func TestFusionEngine_Round405_NoopFallbackEmitsEnglish(
	t *testing.T,
) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	e := New(cfg, logger)
	ctx := context.Background()

	_, err := e.ClassifyEffort(ctx, "")
	require.Error(t, err)
	assert.Equal(t, "empty request", err.Error())
	assert.NotContains(
		t, err.Error(), "helixspecifier_engine_err_",
		"raw i18n key leaked into noop-path output",
	)

	_, err = e.GetFlowStatus("nope")
	require.Error(t, err)
	assert.Equal(t, "flow nope not found", err.Error(),
		"flow id arg must be substituted into the fallback")
}
