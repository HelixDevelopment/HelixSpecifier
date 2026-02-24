package metrics

import (
	"sync"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	require.NotNil(t, m)
	assert.NotNil(t, m.PhaseMetrics)
	assert.NotNil(t, m.EffortCounts)
	assert.NotNil(t, m.CeremonyCounts)
	assert.Equal(t, int64(0), m.FlowsStarted)
}

func TestMetrics_RecordFlowStart(t *testing.T) {
	m := NewMetrics()
	m.RecordFlowStart(types.EffortLarge)
	assert.Equal(t, int64(1), m.FlowsStarted)
	assert.Equal(t, int64(1),
		m.EffortCounts[types.EffortLarge])
}

func TestMetrics_RecordFlowComplete_Success(t *testing.T) {
	m := NewMetrics()
	m.RecordFlowStart(types.EffortMedium)

	result := &types.FlowResult{
		Success:       true,
		Duration:      2 * time.Second,
		CeremonyLevel: types.CeremonyLight,
	}
	m.RecordFlowComplete(result)

	assert.Equal(t, int64(1), m.FlowsCompleted)
	assert.Equal(t, int64(0), m.FlowsFailed)
	assert.Equal(t, 2*time.Second, m.TotalDuration)
}

func TestMetrics_RecordFlowComplete_Failure(t *testing.T) {
	m := NewMetrics()
	m.RecordFlowStart(types.EffortLarge)

	result := &types.FlowResult{
		Success:       false,
		Duration:      5 * time.Second,
		CeremonyLevel: types.CeremonyStandard,
	}
	m.RecordFlowComplete(result)

	assert.Equal(t, int64(0), m.FlowsCompleted)
	assert.Equal(t, int64(1), m.FlowsFailed)
}

func TestMetrics_RecordPhase_Success(t *testing.T) {
	m := NewMetrics()
	result := &types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		Duration:     1 * time.Second,
		QualityScore: 0.9,
	}
	m.RecordPhase(result)

	assert.Equal(t, int64(1), m.PhasesExecuted)
	assert.Equal(t, int64(0), m.PhasesFailed)

	pm := m.GetPhaseMetric(types.PhaseSpecify)
	assert.Equal(t, int64(1), pm.Executions)
	assert.Equal(t, int64(1), pm.Successes)
	assert.InDelta(t, 0.9, pm.AvgScore, 0.001)
}

func TestMetrics_RecordPhase_Failure(t *testing.T) {
	m := NewMetrics()
	result := &types.PhaseResult{
		Phase:        types.PhasePlan,
		Success:      false,
		Duration:     500 * time.Millisecond,
		QualityScore: 0.3,
	}
	m.RecordPhase(result)

	assert.Equal(t, int64(1), m.PhasesExecuted)
	assert.Equal(t, int64(1), m.PhasesFailed)

	pm := m.GetPhaseMetric(types.PhasePlan)
	assert.Equal(t, int64(1), pm.Failures)
	assert.Equal(t, int64(0), pm.Successes)
}

func TestMetrics_GetFlowSuccessRate_Zero(t *testing.T) {
	m := NewMetrics()
	assert.Equal(t, 0.0, m.GetFlowSuccessRate())
}

func TestMetrics_GetFlowSuccessRate_Some(t *testing.T) {
	m := NewMetrics()

	m.RecordFlowStart(types.EffortMedium)
	m.RecordFlowComplete(&types.FlowResult{
		Success:       true,
		CeremonyLevel: types.CeremonyLight,
	})

	m.RecordFlowStart(types.EffortLarge)
	m.RecordFlowComplete(&types.FlowResult{
		Success:       false,
		CeremonyLevel: types.CeremonyStandard,
	})

	assert.InDelta(t, 0.5, m.GetFlowSuccessRate(), 0.001)
}

func TestMetrics_GetPhaseMetric_Exists(t *testing.T) {
	m := NewMetrics()
	m.RecordPhase(&types.PhaseResult{
		Phase:        types.PhaseConstitution,
		Success:      true,
		QualityScore: 0.85,
	})

	pm := m.GetPhaseMetric(types.PhaseConstitution)
	assert.Equal(t, int64(1), pm.Executions)
	assert.InDelta(t, 0.85, pm.AvgScore, 0.001)
}

func TestMetrics_GetPhaseMetric_NotExists(t *testing.T) {
	m := NewMetrics()
	pm := m.GetPhaseMetric(types.PhaseAnalyze)
	assert.Equal(t, int64(0), pm.Executions)
}

func TestMetrics_GetEffortCount(t *testing.T) {
	m := NewMetrics()
	m.RecordFlowStart(types.EffortEpic)
	m.RecordFlowStart(types.EffortEpic)
	m.RecordFlowStart(types.EffortQuick)

	assert.Equal(t, int64(2),
		m.GetEffortCount(types.EffortEpic))
	assert.Equal(t, int64(1),
		m.GetEffortCount(types.EffortQuick))
	assert.Equal(t, int64(0),
		m.GetEffortCount(types.EffortMedium))
}

func TestMetrics_Snapshot(t *testing.T) {
	m := NewMetrics()
	m.RecordFlowStart(types.EffortLarge)
	m.RecordPhase(&types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		QualityScore: 0.8,
	})

	snap := m.Snapshot()
	assert.Equal(t, int64(1), snap.FlowsStarted)
	assert.Equal(t, int64(1), snap.PhasesExecuted)

	// Mutating original should not affect snapshot
	m.RecordFlowStart(types.EffortMedium)
	assert.Equal(t, int64(1), snap.FlowsStarted)
	assert.Equal(t, int64(2), m.FlowsStarted)
}

func TestMetrics_ConcurrentAccess(t *testing.T) {
	m := NewMetrics()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RecordFlowStart(types.EffortMedium)
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.RecordPhase(&types.PhaseResult{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.7,
			})
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.GetFlowSuccessRate()
			_ = m.Snapshot()
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(100), m.FlowsStarted)
	assert.Equal(t, int64(100), m.PhasesExecuted)
}

func TestMetrics_MultiplePhases(t *testing.T) {
	m := NewMetrics()

	phases := types.AllPhases()
	for _, ph := range phases {
		m.RecordPhase(&types.PhaseResult{
			Phase:        ph,
			Success:      true,
			QualityScore: 0.75,
		})
	}

	assert.Equal(t, int64(len(phases)), m.PhasesExecuted)
	for _, ph := range phases {
		pm := m.GetPhaseMetric(ph)
		assert.Equal(t, int64(1), pm.Executions)
	}
}

func TestMetrics_AvgScore(t *testing.T) {
	m := NewMetrics()

	m.RecordPhase(&types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		QualityScore: 0.6,
	})
	m.RecordPhase(&types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		QualityScore: 0.8,
	})
	m.RecordPhase(&types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		QualityScore: 1.0,
	})

	pm := m.GetPhaseMetric(types.PhaseSpecify)
	assert.Equal(t, int64(3), pm.Executions)
	assert.InDelta(t, 0.8, pm.AvgScore, 0.001)
}
