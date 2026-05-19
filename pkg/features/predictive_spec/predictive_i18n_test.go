package predictive_spec

import (
	"testing"

	"digital.vasic.helixspecifier/pkg/i18n"
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

// TestPredictor_Predict_DescriptionUsesI18nKey asserts the
// round-208 migration: Prediction.Description is produced via the
// i18n translator, not the old "Based on: %s" literal. A regression
// that reinstates the literal would fail this test — paired-
// mutation guard for CONST-046.
func TestPredictor_Predict_DescriptionUsesI18nKey(t *testing.T) {
	rec := &recordingTranslator{}
	p := NewPredictorWithTranslator(rec)
	require.NotNil(t, p)

	p.AddHistory("login-event")
	p.AddHistory("checkout-event")

	preds := p.Predict()
	require.Len(t, preds, 2)
	for _, pr := range preds {
		assert.Equal(
			t,
			"stub:helixspecifier_predictive_basis_label",
			pr.Description,
		)
		assert.NotContains(
			t, pr.Description, "Based on:",
			"regression: hardcoded literal reintroduced",
		)
	}
	// translator should have been called once per history entry.
	assert.GreaterOrEqual(t, len(rec.calls), 2)
}

// TestPredictor_Predict_NoopFallback verifies the legacy
// NewPredictor() constructor stays operational and the noop
// fallback produces a visible key + event in the output, so
// missing-bundle situations are visibly debuggable.
func TestPredictor_Predict_NoopFallback(t *testing.T) {
	p := NewPredictor()
	p.AddHistory("login-event")
	preds := p.Predict()
	require.Len(t, preds, 1)
	assert.Contains(
		t, preds[0].Description,
		"helixspecifier_predictive_basis_label",
	)
	assert.Contains(t, preds[0].Description, "login-event")
}

// TestPredictor_NewPredictorWithTranslator_NilFallback proves
// the constructor accepts nil and falls back to NoopTranslator.
func TestPredictor_NewPredictorWithTranslator_NilFallback(t *testing.T) {
	p := NewPredictorWithTranslator(nil)
	require.NotNil(t, p)
	require.NotNil(t, p.translator)
}

// TestPredictor_SetTranslator_Swap exercises runtime swap
// including nil-fallback. Confirms no panic and subsequent
// Predict() calls use the new backend.
func TestPredictor_SetTranslator_Swap(t *testing.T) {
	p := NewPredictor()
	p.AddHistory("e1")
	preds1 := p.Predict()
	require.Len(t, preds1, 1)
	assert.Contains(
		t, preds1[0].Description,
		"helixspecifier_predictive_basis_label",
	)

	rec := &recordingTranslator{}
	p.SetTranslator(rec)
	preds2 := p.Predict()
	require.Len(t, preds2, 1)
	assert.Equal(
		t,
		"stub:helixspecifier_predictive_basis_label",
		preds2[0].Description,
	)

	p.SetTranslator(nil)
	preds3 := p.Predict()
	require.Len(t, preds3, 1)
	assert.NotContains(t, preds3[0].Description, "stub:")
}

// Sentinel that the i18n package is referenced.
func TestPredictive_I18nPackagePresent(t *testing.T) {
	tr := i18n.NewNoopTranslator()
	assert.NotNil(t, tr)
}
