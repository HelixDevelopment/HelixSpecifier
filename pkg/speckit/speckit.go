package speckit

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
)

// ErrDebateFuncNotConfigured is returned by ExecutePhase when the
// Pillar's DebateFunc has not been wired before the call.
//
// Round-28 §11.4 audit (2026-05-17): the previous nil-DebateFunc
// branch fabricated a `[<phase>] Phase output for: <request>`
// string, hardcoded `0.75` as QualityScore, and set
// `result.Success = true` — a PASS-bluff at the baseline-runner-
// class default layer. Any consumer that gated on
// `result.Success && result.QualityScore > 0.5` passed against
// fabricated data with no error surfaced. Per the operator's
// anti-bluff mandate, ExecutePhase now returns this sentinel and
// writes no fake output / score when DebateFunc is nil.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests), Article XI §11.9 (forensic
// anchor).
var ErrDebateFuncNotConfigured = fmt.Errorf("helixspecifier speckit: DebateFunc has not been wired into the phase — set Phase.DebateFunc to a real debate orchestrator before invoking ExecutePhase (the previous nil-DebateFunc branch fabricated [phase] Phase output strings with a hardcoded 0.75 quality score and marked Success=true; §11.4 PASS-bluff removed)")

// Pillar implements the SpecKitPillar interface providing the
// 7-phase Spec-Driven Development workflow.
type Pillar struct {
	cfg    *config.Config
	logger *logrus.Logger
	// DebateFunc is called to execute a debate for a phase.
	// This allows injection of the actual debate service.
	// MUST be set via SetDebateFunc before ExecutePhase is
	// invoked. A nil DebateFunc causes ExecutePhase to return
	// ErrDebateFuncNotConfigured without writing any phase
	// output (round-28 §11.4 audit).
	DebateFunc types.DebateFunc
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

// SetDebateFunc sets the debate execution function.
func (p *Pillar) SetDebateFunc(fn types.DebateFunc) {
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

	// Execute debate. DebateFunc MUST be wired via SetDebateFunc
	// before reaching this point — round-28 §11.4 audit removed
	// the prior nil-branch that fabricated phase output with a
	// hardcoded 0.75 score and Success=true.
	if p.DebateFunc == nil {
		result.Error = ErrDebateFuncNotConfigured.Error()
		result.EndTime = time.Now()
		result.Duration = time.Since(startTime)
		return result, ErrDebateFuncNotConfigured
	}

	var output string
	var score float64
	var debateID string

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
