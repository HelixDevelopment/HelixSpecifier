package ceremony

import (
	"testing"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScaler() *Scaler {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return NewScaler(cfg, logger)
}

func newTestScalerWithBoost(boost float64) *Scaler {
	cfg := config.DefaultConfig()
	cfg.CeremonyBoost = boost
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return NewScaler(cfg, logger)
}

// --- TestNewScaler ---

func TestNewScaler(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	s := NewScaler(cfg, logger)

	require.NotNil(t, s)
	assert.Equal(t, cfg, s.cfg)
	assert.Equal(t, logger, s.logger)
}

// --- TestScaler_Scale_Quick ---

func TestScaler_Scale_Quick(t *testing.T) {
	s := newTestScaler()
	c := &types.EffortClassification{
		Level: types.EffortQuick,
	}

	result := s.Scale(c)
	assert.Equal(t, types.CeremonyMinimal, result,
		"Quick effort should map to minimal ceremony")
}

// --- TestScaler_Scale_Medium ---

func TestScaler_Scale_Medium(t *testing.T) {
	s := newTestScaler()
	c := &types.EffortClassification{
		Level: types.EffortMedium,
	}

	result := s.Scale(c)
	assert.Equal(t, types.CeremonyLight, result,
		"Medium effort should map to light ceremony")
}

// --- TestScaler_Scale_Large ---

func TestScaler_Scale_Large(t *testing.T) {
	s := newTestScaler()
	c := &types.EffortClassification{
		Level: types.EffortLarge,
	}

	result := s.Scale(c)
	assert.Equal(t, types.CeremonyStandard, result,
		"Large effort should map to standard ceremony")
}

// --- TestScaler_Scale_Epic ---

func TestScaler_Scale_Epic(t *testing.T) {
	s := newTestScaler()
	c := &types.EffortClassification{
		Level: types.EffortEpic,
	}

	result := s.Scale(c)
	assert.Equal(t, types.CeremonyHeavy, result,
		"Epic effort should map to heavy ceremony")
}

// --- TestScaler_Scale_WithBoost ---

func TestScaler_Scale_WithBoost(t *testing.T) {
	s := newTestScalerWithBoost(1.0)

	tests := []struct {
		name     string
		effort   types.EffortLevel
		expected types.CeremonyLevel
	}{
		{
			"quick boosted to light",
			types.EffortQuick,
			types.CeremonyLight,
		},
		{
			"medium boosted to standard",
			types.EffortMedium,
			types.CeremonyStandard,
		},
		{
			"large boosted to heavy",
			types.EffortLarge,
			types.CeremonyHeavy,
		},
		{
			"epic stays heavy (max)",
			types.EffortEpic,
			types.CeremonyHeavy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &types.EffortClassification{
				Level: tt.effort,
			}
			result := s.Scale(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- TestScaler_Adjust_NoResults ---

func TestScaler_Adjust_NoResults(t *testing.T) {
	s := newTestScaler()

	result := s.Adjust(
		types.CeremonyStandard,
		[]types.PhaseResult{},
	)
	assert.Equal(t, types.CeremonyStandard, result,
		"No results should not change ceremony")
}

// --- TestScaler_Adjust_HighQuality ---

func TestScaler_Adjust_HighQuality(t *testing.T) {
	s := newTestScaler()

	phaseResults := []types.PhaseResult{
		{QualityScore: 0.90},
		{QualityScore: 0.95},
		{QualityScore: 0.88},
	}

	result := s.Adjust(types.CeremonyHeavy, phaseResults)
	assert.Equal(t, types.CeremonyStandard, result,
		"High quality heavy should reduce to standard")

	// Standard should stay standard (not reduced further)
	result2 := s.Adjust(
		types.CeremonyStandard, phaseResults,
	)
	assert.Equal(t, types.CeremonyStandard, result2,
		"Standard should not be reduced by high quality")
}

// --- TestScaler_Adjust_LowQuality ---

func TestScaler_Adjust_LowQuality(t *testing.T) {
	s := newTestScaler()

	phaseResults := []types.PhaseResult{
		{QualityScore: 0.30},
		{QualityScore: 0.40},
		{QualityScore: 0.45},
	}

	result := s.Adjust(types.CeremonyLight, phaseResults)
	assert.Equal(t, types.CeremonyStandard, result,
		"Low quality light should increase to standard")

	// Standard should stay standard (not increased further)
	result2 := s.Adjust(
		types.CeremonyStandard, phaseResults,
	)
	assert.Equal(t, types.CeremonyStandard, result2,
		"Standard should not increase on low quality")
}

// --- TestScaler_Adjust_NormalQuality ---

func TestScaler_Adjust_NormalQuality(t *testing.T) {
	s := newTestScaler()

	phaseResults := []types.PhaseResult{
		{QualityScore: 0.65},
		{QualityScore: 0.70},
		{QualityScore: 0.75},
	}

	levels := []types.CeremonyLevel{
		types.CeremonyMinimal,
		types.CeremonyLight,
		types.CeremonyStandard,
		types.CeremonyHeavy,
	}

	for _, level := range levels {
		result := s.Adjust(level, phaseResults)
		assert.Equal(t, level, result,
			"Normal quality should not change %s", level)
	}
}

// --- TestScaler_ImplementsInterface ---

func TestScaler_ImplementsInterface(t *testing.T) {
	var _ types.CeremonyScaler = (*Scaler)(nil)
	s := newTestScaler()
	var iface types.CeremonyScaler = s
	assert.NotNil(t, iface,
		"Scaler must implement CeremonyScaler interface")
}

// --- TestScaler_BoostCeremony (table-driven) ---

func TestScaler_BoostCeremony(t *testing.T) {
	s := newTestScaler()

	tests := []struct {
		name     string
		input    types.CeremonyLevel
		expected types.CeremonyLevel
	}{
		{
			"minimal to light",
			types.CeremonyMinimal,
			types.CeremonyLight,
		},
		{
			"light to standard",
			types.CeremonyLight,
			types.CeremonyStandard,
		},
		{
			"standard to heavy",
			types.CeremonyStandard,
			types.CeremonyHeavy,
		},
		{
			"heavy stays heavy",
			types.CeremonyHeavy,
			types.CeremonyHeavy,
		},
		{
			"unknown stays unknown",
			types.CeremonyLevel("unknown"),
			types.CeremonyLevel("unknown"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.boostCeremony(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- TestScaler_Adjust_SingleResult ---

func TestScaler_Adjust_SingleResult(t *testing.T) {
	s := newTestScaler()

	// Single high-quality result on heavy -> standard
	high := []types.PhaseResult{{QualityScore: 0.95}}
	result := s.Adjust(types.CeremonyHeavy, high)
	assert.Equal(t, types.CeremonyStandard, result)

	// Single low-quality result on light -> standard
	low := []types.PhaseResult{{QualityScore: 0.20}}
	result = s.Adjust(types.CeremonyLight, low)
	assert.Equal(t, types.CeremonyStandard, result)
}
