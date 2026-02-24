package speckit

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
)

// Pillar implements the SpecKitPillar interface providing the
// 7-phase Spec-Driven Development workflow.
type Pillar struct {
	cfg    *config.Config
	logger *logrus.Logger
	// DebateFunc is called to execute a debate for a phase.
	// This allows injection of the actual debate service.
	DebateFunc func(
		ctx context.Context,
		topic string,
		rounds int,
		metadata map[string]interface{},
	) (string, float64, string, error) // output, score, debateID, err
}

// NewPillar creates a new SpecKit pillar
func NewPillar(
	cfg *config.Config,
	logger *logrus.Logger,
) *Pillar {
	return &Pillar{
		cfg:    cfg,
		logger: logger,
	}
}

// SetDebateFunc sets the debate execution function
func (p *Pillar) SetDebateFunc(fn func(
	ctx context.Context,
	topic string,
	rounds int,
	metadata map[string]interface{},
) (string, float64, string, error)) {
	p.DebateFunc = fn
}

// ExecutePhase runs a single SpecKit phase
func (p *Pillar) ExecutePhase(
	ctx context.Context,
	phase types.SpecKitPhase,
	input *types.PhaseInput,
) (*types.PhaseResult, error) {
	startTime := time.Now()

	result := &types.PhaseResult{
		Phase:     phase,
		StartTime: startTime,
		Success:   false,
	}

	p.logger.WithField("phase", phase).Info(
		"[SpecKit] Starting phase",
	)

	// Build debate topic for the phase
	topic := p.buildPhaseTopic(phase, input)
	rounds := p.cfg.GetPhaseRounds(string(phase))

	metadata := map[string]interface{}{
		"phase": string(phase),
	}
	if input.Metadata != nil {
		for k, v := range input.Metadata {
			metadata[k] = v
		}
	}

	// Execute debate if function is available
	var output string
	var score float64
	var debateID string

	if p.DebateFunc != nil {
		var err error
		output, score, debateID, err = p.DebateFunc(
			ctx, topic, rounds, metadata,
		)
		if err != nil {
			result.Error = fmt.Sprintf(
				"debate failed: %v", err,
			)
			result.EndTime = time.Now()
			result.Duration = time.Since(startTime)
			return result, err
		}
	} else {
		// No debate function -- produce placeholder output
		output = fmt.Sprintf(
			"[%s] Phase output for: %s",
			phase, input.UserRequest,
		)
		score = 0.75
	}

	result.Success = true
	result.Output = output
	result.QualityScore = score
	result.DebateID = debateID
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)
	result.Artifacts = map[string]interface{}{
		"topic":  topic,
		"rounds": rounds,
	}

	// Determine fusion stage
	result.Stage = p.phaseToStage(phase)

	p.logger.WithFields(logrus.Fields{
		"phase":    phase,
		"duration": result.Duration,
		"score":    score,
	}).Info("[SpecKit] Completed phase")

	return result, nil
}

// GetPhaseOrder returns phases for a ceremony level
func (p *Pillar) GetPhaseOrder(
	ceremony types.CeremonyLevel,
) []types.SpecKitPhase {
	return types.PhasesForCeremony(ceremony)
}

// ValidateTransition checks if a phase transition is valid
func (p *Pillar) ValidateTransition(
	from, to types.SpecKitPhase,
) bool {
	allPhases := types.AllPhases()
	fromIdx := -1
	toIdx := -1
	for i, ph := range allPhases {
		if ph == from {
			fromIdx = i
		}
		if ph == to {
			toIdx = i
		}
	}
	if fromIdx == -1 || toIdx == -1 {
		return false
	}
	return toIdx == fromIdx+1
}

// buildPhaseTopic generates the debate topic for a phase
func (p *Pillar) buildPhaseTopic(
	phase types.SpecKitPhase,
	input *types.PhaseInput,
) string {
	switch phase {
	case types.PhaseConstitution:
		return fmt.Sprintf(
			"Analyze the request and create/update the "+
				"project Constitution:\n\nRequest: %s"+
				"\n\nConstitution: %s",
			input.UserRequest, input.Constitution,
		)
	case types.PhaseSpecify:
		return fmt.Sprintf(
			"Create a detailed specification based on "+
				"the request and Constitution:\n\n"+
				"Request: %s\n\nPrevious: %s",
			input.UserRequest, input.PreviousOutput,
		)
	case types.PhaseClarify:
		return fmt.Sprintf(
			"Clarify and validate the specification:"+
				"\n\nSpecification: %s",
			input.PreviousOutput,
		)
	case types.PhasePlan:
		return fmt.Sprintf(
			"Create implementation plan:\n\nSpec: %s",
			input.PreviousOutput,
		)
	case types.PhaseTasks:
		return fmt.Sprintf(
			"Break down plan into tasks:\n\nPlan: %s",
			input.PreviousOutput,
		)
	case types.PhaseAnalyze:
		return fmt.Sprintf(
			"Analyze tasks before implementation:"+
				"\n\nTasks: %s",
			input.PreviousOutput,
		)
	case types.PhaseImplement:
		return fmt.Sprintf(
			"Execute implementation:\n\nAnalysis: %s",
			input.PreviousOutput,
		)
	default:
		return fmt.Sprintf(
			"Execute phase %s: %s",
			phase, input.UserRequest,
		)
	}
}

// phaseToStage maps a SpecKit phase to a fusion stage
func (p *Pillar) phaseToStage(
	phase types.SpecKitPhase,
) types.FusionStage {
	switch phase {
	case types.PhaseConstitution:
		return types.StageConstitution
	case types.PhaseSpecify, types.PhaseClarify:
		return types.StageConstitution
	case types.PhasePlan, types.PhaseTasks:
		return types.StagePlanning
	case types.PhaseAnalyze:
		return types.StagePlanning
	case types.PhaseImplement:
		return types.StageExecution
	default:
		return types.StageExecution
	}
}
