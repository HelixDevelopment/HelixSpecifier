// Package nyquist_tdd implements Nyquist-frequency inspired
// TDD discipline ensuring tests are written at twice the rate
// of implementation code (the "sampling theorem" of quality).
package nyquist_tdd

import (
	"fmt"
	"sync"
)

// Tracker monitors the ratio of test code to implementation
// code, enforcing the Nyquist TDD principle: test coverage
// must be at least 2x the implementation surface.
type Tracker struct {
	implLines  int
	testLines  int
	violations []string
	mu         sync.RWMutex
}

// NewTracker creates a new Nyquist TDD tracker.
func NewTracker() *Tracker {
	return &Tracker{
		violations: []string{},
	}
}

// RecordImplementation records implementation lines added.
func (tr *Tracker) RecordImplementation(lines int) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.implLines += lines
}

// RecordTests records test lines added.
func (tr *Tracker) RecordTests(lines int) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.testLines += lines
}

// Ratio returns the current test-to-implementation ratio.
// Returns 0 if no implementation lines recorded.
func (tr *Tracker) Ratio() float64 {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	if tr.implLines == 0 {
		return 0.0
	}
	return float64(tr.testLines) / float64(tr.implLines)
}

// IsCompliant checks if the ratio meets the Nyquist
// threshold (>= 2.0).
func (tr *Tracker) IsCompliant() bool {
	return tr.Ratio() >= 2.0
}

// Check validates the current state and records a violation
// if the ratio is below threshold. Returns nil if compliant.
func (tr *Tracker) Check() error {
	ratio := tr.Ratio()
	if ratio >= 2.0 {
		return nil
	}
	msg := fmt.Sprintf(
		"Nyquist TDD violation: ratio=%.2f (need >=2.0)",
		ratio,
	)
	tr.mu.Lock()
	tr.violations = append(tr.violations, msg)
	tr.mu.Unlock()
	return fmt.Errorf("%s", msg)
}

// Violations returns all recorded violations.
func (tr *Tracker) Violations() []string {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	result := make([]string, len(tr.violations))
	copy(result, tr.violations)
	return result
}

// Stats returns implementation lines and test lines.
func (tr *Tracker) Stats() (impl, tests int) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	return tr.implLines, tr.testLines
}
