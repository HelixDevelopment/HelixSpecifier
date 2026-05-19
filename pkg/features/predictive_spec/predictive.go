// Package predictive_spec provides predictive specification
// capabilities that anticipate future requirements based on
// historical data and project patterns.
//
// User-visible prediction descriptions are resolved through the
// i18n translator (CONST-046). A nil translator falls back to
// NoopTranslator so the zero-value Predictor stays panic-free.
package predictive_spec

import (
	"fmt"
	"sync"
	"time"

	"digital.vasic.helixspecifier/pkg/i18n"
)

// Prediction represents a predicted future requirement.
type Prediction struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	Basis       string    `json:"basis"`
	CreatedAt   time.Time `json:"created_at"`
}

// Predictor generates specification predictions based on
// historical patterns and project context.
type Predictor struct {
	predictions []Prediction
	history     []string
	translator  i18n.Translator
	mu          sync.RWMutex
}

// NewPredictor creates a new spec predictor with the default
// NoopTranslator. Production callers SHOULD prefer
// NewPredictorWithTranslator and inject a real backend.
func NewPredictor() *Predictor {
	return &Predictor{
		predictions: []Prediction{},
		history:     []string{},
		translator:  i18n.NewNoopTranslator(),
	}
}

// NewPredictorWithTranslator creates a Predictor with an injected
// translator. Passing nil falls back to NoopTranslator.
func NewPredictorWithTranslator(t i18n.Translator) *Predictor {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	return &Predictor{
		predictions: []Prediction{},
		history:     []string{},
		translator:  t,
	}
}

// SetTranslator swaps the Predictor's i18n translator at
// runtime. Passing nil falls back to NoopTranslator so the
// Predictor stays panic-free.
func (p *Predictor) SetTranslator(t i18n.Translator) {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.translator = t
}

// AddHistory records a historical event for analysis.
func (p *Predictor) AddHistory(event string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.history = append(p.history, event)
}

// Predict generates predictions based on accumulated history.
// Each history entry produces a prediction with decreasing
// confidence for older entries.
func (p *Predictor) Predict() []Prediction {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.predictions = []Prediction{}

	t := p.translator
	if t == nil {
		t = i18n.NewNoopTranslator()
	}

	for i, event := range p.history {
		confidence := 1.0 - float64(i)*0.1
		if confidence < 0.1 {
			confidence = 0.1
		}

		pred := Prediction{
			ID: fmt.Sprintf("pred-%d", i+1),
			Description: t.T(
				"helixspecifier_predictive_basis_label",
				event,
			),
			Confidence: confidence,
			Basis:      event,
			CreatedAt:  time.Now(),
		}
		p.predictions = append(p.predictions, pred)
	}

	return p.predictions
}

// PredictionCount returns the number of predictions.
func (p *Predictor) PredictionCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.predictions)
}

// HistoryCount returns the number of history entries.
func (p *Predictor) HistoryCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.history)
}

// HighConfidence returns predictions above the threshold.
func (p *Predictor) HighConfidence(
	threshold float64,
) []Prediction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := []Prediction{}
	for _, pred := range p.predictions {
		if pred.Confidence >= threshold {
			result = append(result, pred)
		}
	}
	return result
}
