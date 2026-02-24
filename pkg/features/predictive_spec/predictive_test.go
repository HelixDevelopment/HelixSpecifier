package predictive_spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPredictor(t *testing.T) {
	p := NewPredictor()
	require.NotNil(t, p)
	assert.Equal(t, 0, p.PredictionCount())
	assert.Equal(t, 0, p.HistoryCount())
}

func TestPredictor_AddHistory(t *testing.T) {
	p := NewPredictor()
	p.AddHistory("added auth module")
	p.AddHistory("added database layer")
	assert.Equal(t, 2, p.HistoryCount())
}

func TestPredictor_Predict(t *testing.T) {
	p := NewPredictor()
	p.AddHistory("added auth module")
	p.AddHistory("added caching layer")

	preds := p.Predict()
	assert.Len(t, preds, 2)
	assert.Equal(t, "pred-1", preds[0].ID)
	assert.Greater(t, preds[0].Confidence, 0.0)
}

func TestPredictor_Predict_DecreasingConfidence(
	t *testing.T,
) {
	p := NewPredictor()
	for i := 0; i < 5; i++ {
		p.AddHistory("event")
	}
	preds := p.Predict()
	// First should have highest confidence
	assert.Greater(t,
		preds[0].Confidence, preds[4].Confidence)
}

func TestPredictor_HighConfidence(t *testing.T) {
	p := NewPredictor()
	for i := 0; i < 15; i++ {
		p.AddHistory("event")
	}
	p.Predict()

	high := p.HighConfidence(0.8)
	for _, pred := range high {
		assert.GreaterOrEqual(t, pred.Confidence, 0.8)
	}
}

func TestPredictor_PredictionCount(t *testing.T) {
	p := NewPredictor()
	assert.Equal(t, 0, p.PredictionCount())
	p.AddHistory("x")
	p.Predict()
	assert.Equal(t, 1, p.PredictionCount())
}
