package adaptive_ceremony

import (
	"testing"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdjuster(t *testing.T) {
	a := NewAdjuster(0.9, 0.5)
	require.NotNil(t, a)
	up, down := a.Thresholds()
	assert.InDelta(t, 0.9, up, 0.001)
	assert.InDelta(t, 0.5, down, 0.001)
}

func TestNewAdjuster_Defaults(t *testing.T) {
	a := NewAdjuster(0, 0)
	up, down := a.Thresholds()
	assert.InDelta(t, 0.9, up, 0.001)
	assert.InDelta(t, 0.5, down, 0.001)
}

func TestAdjuster_Adjust_NoChange(t *testing.T) {
	a := NewAdjuster(0.9, 0.5)
	results := []types.PhaseResult{
		{QualityScore: 0.7},
		{QualityScore: 0.75},
	}
	level := a.Adjust(types.CeremonyStandard, results)
	assert.Equal(t, types.CeremonyStandard, level)
	assert.Equal(t, 0, a.AdjustmentCount())
}

func TestAdjuster_Adjust_ReduceOnHighQuality(t *testing.T) {
	a := NewAdjuster(0.85, 0.4)
	results := []types.PhaseResult{
		{QualityScore: 0.95},
		{QualityScore: 0.92},
	}
	level := a.Adjust(types.CeremonyHeavy, results)
	assert.Equal(t, types.CeremonyStandard, level)
	assert.Equal(t, 1, a.AdjustmentCount())
}

func TestAdjuster_Adjust_IncreaseOnLowQuality(
	t *testing.T,
) {
	a := NewAdjuster(0.9, 0.5)
	results := []types.PhaseResult{
		{QualityScore: 0.3},
		{QualityScore: 0.4},
	}
	level := a.Adjust(types.CeremonyLight, results)
	assert.Equal(t, types.CeremonyStandard, level)
	assert.Equal(t, 1, a.AdjustmentCount())
}

func TestAdjuster_Adjust_EmptyResults(t *testing.T) {
	a := NewAdjuster(0.9, 0.5)
	level := a.Adjust(
		types.CeremonyStandard,
		[]types.PhaseResult{},
	)
	assert.Equal(t, types.CeremonyStandard, level)
}

func TestAdjuster_String(t *testing.T) {
	a := NewAdjuster(0.9, 0.5)
	s := a.String()
	assert.Contains(t, s, "AdaptiveCeremony")
	assert.Contains(t, s, "0.90")
}
