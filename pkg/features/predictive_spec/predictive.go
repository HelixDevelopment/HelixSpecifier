// Package predictive_spec provides predictive specification
// capabilities that anticipate future requirements based on
// historical data and project patterns.
package predictive_spec

import (
	"fmt"
	"sync"
	"time"
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
	mu          sync.RWMutex
}

// NewPredictor creates a new spec predictor.
func NewPredictor() *Predictor {
	return &Predictor{
		predictions: []Prediction{},
		history:     []string{},
	}
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

	for i, event := range p.history {
		confidence := 1.0 - float64(i)*0.1
		if confidence < 0.1 {
			confidence = 0.1
		}

		pred := Prediction{
			ID: fmt.Sprintf("pred-%d", i+1),
			Description: fmt.Sprintf(
				"Based on: %s", event,
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
