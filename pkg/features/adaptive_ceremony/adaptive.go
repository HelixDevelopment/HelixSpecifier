// Package adaptive_ceremony provides dynamic ceremony level
// adjustment based on real-time quality metrics and phase
// results during flow execution.
package adaptive_ceremony

import (
	"fmt"
	"sync"

	"digital.vasic.helixspecifier/pkg/types"
)

// Adjuster dynamically adjusts ceremony levels based on
// phase quality scores and success rates.
type Adjuster struct {
	upThreshold   float64
	downThreshold float64
	adjustments   int
	mu            sync.Mutex
}

// NewAdjuster creates a new ceremony adjuster.
// upThreshold: quality above this triggers ceremony reduction.
// downThreshold: quality below this triggers ceremony increase.
func NewAdjuster(
	upThreshold float64,
	downThreshold float64,
) *Adjuster {
	if upThreshold <= 0 {
		upThreshold = 0.9
	}
	if downThreshold <= 0 {
		downThreshold = 0.5
	}
	return &Adjuster{
		upThreshold:   upThreshold,
		downThreshold: downThreshold,
	}
}

// Adjust evaluates phase results and returns the adjusted
// ceremony level. If average quality is very high, ceremony
// can be reduced. If quality is low, ceremony is increased.
func (a *Adjuster) Adjust(
	current types.CeremonyLevel,
	results []types.PhaseResult,
) types.CeremonyLevel {
	if len(results) == 0 {
		return current
	}

	avgScore := a.averageScore(results)

	a.mu.Lock()
	defer a.mu.Unlock()

	if avgScore >= a.upThreshold {
		reduced := a.reduce(current)
		if reduced != current {
			a.adjustments++
		}
		return reduced
	}

	if avgScore < a.downThreshold {
		increased := a.increase(current)
		if increased != current {
			a.adjustments++
		}
		return increased
	}

	return current
}

// AdjustmentCount returns the total adjustments made.
func (a *Adjuster) AdjustmentCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.adjustments
}

// Thresholds returns the up and down thresholds.
func (a *Adjuster) Thresholds() (up, down float64) {
	return a.upThreshold, a.downThreshold
}

// String returns a human-readable description.
func (a *Adjuster) String() string {
	return fmt.Sprintf(
		"AdaptiveCeremony(up=%.2f, down=%.2f)",
		a.upThreshold, a.downThreshold,
	)
}

// averageScore calculates the average quality score.
func (a *Adjuster) averageScore(
	results []types.PhaseResult,
) float64 {
	total := 0.0
	for _, r := range results {
		total += r.QualityScore
	}
	return total / float64(len(results))
}

// reduce lowers the ceremony level by one step.
func (a *Adjuster) reduce(
	level types.CeremonyLevel,
) types.CeremonyLevel {
	switch level {
	case types.CeremonyHeavy:
		return types.CeremonyStandard
	case types.CeremonyStandard:
		return types.CeremonyLight
	case types.CeremonyLight:
		return types.CeremonyMinimal
	default:
		return level
	}
}

// increase raises the ceremony level by one step.
func (a *Adjuster) increase(
	level types.CeremonyLevel,
) types.CeremonyLevel {
	switch level {
	case types.CeremonyMinimal:
		return types.CeremonyLight
	case types.CeremonyLight:
		return types.CeremonyStandard
	case types.CeremonyStandard:
		return types.CeremonyHeavy
	default:
		return level
	}
}
