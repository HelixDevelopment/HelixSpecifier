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
