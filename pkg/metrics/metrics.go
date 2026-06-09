package metrics

import (
	"sync"
	"time"

	"digital.vasic.helixspecifier/pkg/types"
)

// Metrics tracks HelixSpecifier operational metrics
type Metrics struct {
	FlowsStarted  int64
	FlowsCompleted int64
	FlowsFailed    int64
	PhasesExecuted int64
	PhasesFailed   int64
	TotalDuration  time.Duration
	PhaseMetrics   map[types.SpecKitPhase]*PhaseMetric
	EffortCounts   map[types.EffortLevel]int64
	CeremonyCounts map[types.CeremonyLevel]int64
	mu             sync.RWMutex
}

// MetricsSnapshot is a mutex-free, read-only copy of Metrics returned by
// Snapshot. It carries the same data fields as Metrics but omits the embedded
// sync.RWMutex, so handing it to callers by value never copies a lock (the
// copied-lock concurrency hazard go vet's copylocks analyzer flags).
type MetricsSnapshot struct {
	FlowsStarted   int64
	FlowsCompleted int64
	FlowsFailed    int64
	PhasesExecuted int64
	PhasesFailed   int64
	TotalDuration  time.Duration
	PhaseMetrics   map[types.SpecKitPhase]*PhaseMetric
	EffortCounts   map[types.EffortLevel]int64
	CeremonyCounts map[types.CeremonyLevel]int64
}

// PhaseMetric tracks per-phase metrics
type PhaseMetric struct {
	Executions    int64
	Successes     int64
	Failures      int64
	TotalDuration time.Duration
	AvgScore      float64
	totalScore    float64
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		PhaseMetrics: make(
			map[types.SpecKitPhase]*PhaseMetric,
		),
		EffortCounts: make(
			map[types.EffortLevel]int64,
		),
		CeremonyCounts: make(
			map[types.CeremonyLevel]int64,
		),
	}
}

// RecordFlowStart records a flow start
func (m *Metrics) RecordFlowStart(
	effort types.EffortLevel,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FlowsStarted++
	m.EffortCounts[effort]++
}

// RecordFlowComplete records a flow completion
func (m *Metrics) RecordFlowComplete(
	result *types.FlowResult,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if result.Success {
		m.FlowsCompleted++
	} else {
		m.FlowsFailed++
	}
	m.TotalDuration += result.Duration
	m.CeremonyCounts[result.CeremonyLevel]++
}

// RecordPhase records a phase execution
func (m *Metrics) RecordPhase(result *types.PhaseResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pm, ok := m.PhaseMetrics[result.Phase]
	if !ok {
		pm = &PhaseMetric{}
		m.PhaseMetrics[result.Phase] = pm
	}

	pm.Executions++
	pm.TotalDuration += result.Duration
	pm.totalScore += result.QualityScore

	if result.Success {
		pm.Successes++
	} else {
		pm.Failures++
	}

	if pm.Executions > 0 {
		pm.AvgScore = pm.totalScore /
			float64(pm.Executions)
	}

	m.PhasesExecuted++
	if !result.Success {
		m.PhasesFailed++
	}
}

// GetFlowSuccessRate returns the flow success rate
func (m *Metrics) GetFlowSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.FlowsStarted == 0 {
		return 0.0
	}
	return float64(m.FlowsCompleted) /
		float64(m.FlowsStarted)
}

// GetPhaseMetric returns metrics for a specific phase
func (m *Metrics) GetPhaseMetric(
	phase types.SpecKitPhase,
) *PhaseMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pm, ok := m.PhaseMetrics[phase]
	if !ok {
		return &PhaseMetric{}
	}
	return pm
}

// GetEffortCount returns count for an effort level
func (m *Metrics) GetEffortCount(
	level types.EffortLevel,
) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.EffortCounts[level]
}

// Snapshot returns a read-only copy of current metrics
func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := MetricsSnapshot{
		FlowsStarted:   m.FlowsStarted,
		FlowsCompleted: m.FlowsCompleted,
		FlowsFailed:    m.FlowsFailed,
		PhasesExecuted: m.PhasesExecuted,
		PhasesFailed:   m.PhasesFailed,
		TotalDuration:  m.TotalDuration,
	}
	cp.PhaseMetrics = make(
		map[types.SpecKitPhase]*PhaseMetric,
	)
	for k, v := range m.PhaseMetrics {
		vCopy := *v
		cp.PhaseMetrics[k] = &vCopy
	}
	cp.EffortCounts = make(
		map[types.EffortLevel]int64,
	)
	for k, v := range m.EffortCounts {
		cp.EffortCounts[k] = v
	}
	cp.CeremonyCounts = make(
		map[types.CeremonyLevel]int64,
	)
	for k, v := range m.CeremonyCounts {
		cp.CeremonyCounts[k] = v
	}
	return cp
}
