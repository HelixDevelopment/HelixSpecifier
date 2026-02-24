package ceremony

import (
	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
)

// Scaler implements adaptive ceremony scaling based on
// effort classification and phase performance.
type Scaler struct {
	cfg    *config.Config
	logger *logrus.Logger
}

// NewScaler creates a new ceremony scaler
func NewScaler(
	cfg *config.Config,
	logger *logrus.Logger,
) *Scaler {
	return &Scaler{cfg: cfg, logger: logger}
}

// Scale determines initial ceremony level
func (s *Scaler) Scale(
	classification *types.EffortClassification,
) types.CeremonyLevel {
	base := types.CeremonyForEffort(classification.Level)

	// Apply boost from config
	if s.cfg.CeremonyBoost > 0 {
		base = s.boostCeremony(base)
	}

	return base
}

// Adjust dynamically adjusts ceremony during execution
func (s *Scaler) Adjust(
	current types.CeremonyLevel,
	phaseResults []types.PhaseResult,
) types.CeremonyLevel {
	if len(phaseResults) == 0 {
		return current
	}

	// Calculate average quality from completed phases
	avgQuality := 0.0
	for _, pr := range phaseResults {
		avgQuality += pr.QualityScore
	}
	avgQuality /= float64(len(phaseResults))

	// If quality is consistently high, can reduce ceremony
	if avgQuality > 0.85 &&
		current == types.CeremonyHeavy {
		s.logger.Info(
			"[Ceremony] High quality detected, " +
				"reducing ceremony",
		)
		return types.CeremonyStandard
	}

	// If quality is low, increase ceremony
	if avgQuality < 0.5 &&
		current == types.CeremonyLight {
		s.logger.Info(
			"[Ceremony] Low quality detected, " +
				"increasing ceremony",
		)
		return types.CeremonyStandard
	}

	return current
}

// boostCeremony increases ceremony by one level
func (s *Scaler) boostCeremony(
	current types.CeremonyLevel,
) types.CeremonyLevel {
	switch current {
	case types.CeremonyMinimal:
		return types.CeremonyLight
	case types.CeremonyLight:
		return types.CeremonyStandard
	case types.CeremonyStandard:
		return types.CeremonyHeavy
	default:
		return current
	}
}
