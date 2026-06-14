package speckit

import (
	"context"
	"fmt"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/i18n"
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
//
// The translator resolves debate-topic templates (the operator-
// visible prompts the orchestrator forwards to LLMs) per
// CONST-046. A nil translator is treated as a NoopTranslator at
// resolution time, keeping NewPillar() callers backward compatible.
type Pillar struct {
	cfg        *config.Config
	logger     *logrus.Logger
	translator i18n.Translator
	// DebateFunc is called to execute a debate for a phase.
	// This allows injection of the actual debate service.
	// MUST be set via SetDebateFunc before ExecutePhase is
	// invoked. A nil DebateFunc causes ExecutePhase to return
	// ErrDebateFuncNotConfigured without writing any phase
	// output (round-28 §11.4 audit).
	DebateFunc types.DebateFunc
}

// NewPillar creates a new SpecKit pillar with the bundle-backed
// default translator (i18n.DefaultTranslator). Production callers
// MAY still inject a custom backend via NewPillarWithTranslator.
//
// HXC-081 fix: the default was previously i18n.NewNoopTranslator(),
// whose T fed the raw i18n key to fmt.Sprintf with positional args
// (the key carries no `%` verb), producing a live /specify topic of
// `helixspecifier_speckit_topic_specify%!(EXTRA string=...)`. The
// bundle-backed default resolves the key to its embedded format
// string (whose `%s` verbs match the args), so the topic renders as
// coherent prose referencing the request.
func NewPillar(
	cfg *config.Config,
	logger *logrus.Logger,
) *Pillar {
	return &Pillar{
		cfg:        cfg,
		logger:     logger,
		translator: i18n.DefaultTranslator(),
	}
}

// NewPillarWithTranslator creates a Pillar with an injected
// translator. Passing nil falls back to NoopTranslator, matching
// the zero-value safety contract of NewPillar.
func NewPillarWithTranslator(
	cfg *config.Config,
	logger *logrus.Logger,
	t i18n.Translator,
) *Pillar {
	if t == nil {
		t = i18n.DefaultTranslator()
	}
	return &Pillar{
		cfg:        cfg,
		logger:     logger,
		translator: t,
	}
}

// SetTranslator swaps the Pillar's i18n translator at runtime.
// Passing nil falls back to NoopTranslator so the Pillar remains
// panic-free.
func (p *Pillar) SetTranslator(t i18n.Translator) {
	if t == nil {
		t = i18n.DefaultTranslator()
	}
	p.translator = t
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

// buildPhaseTopic generates the debate topic for a phase. Topic
// templates are resolved through the i18n translator (CONST-046)
// so non-English operators see locale-appropriate prompts. A nil
// translator is defensively treated as NoopTranslator: a missing
// translation surfaces as the raw key on screen rather than
// crashing the orchestrator (anti-bluff per Article XI §11.9).
func (p *Pillar) buildPhaseTopic(
	phase types.SpecKitPhase,
	input *types.PhaseInput,
) string {
	t := p.translator
	if t == nil {
		t = i18n.DefaultTranslator()
	}
	switch phase {
	case types.PhaseConstitution:
		return t.T(
			"helixspecifier_speckit_topic_constitution",
			input.UserRequest, input.Constitution,
		)
	case types.PhaseSpecify:
		return t.T(
			"helixspecifier_speckit_topic_specify",
			input.UserRequest, input.PreviousOutput,
		)
	case types.PhaseClarify:
		return t.T(
			"helixspecifier_speckit_topic_clarify",
			input.PreviousOutput,
		)
	case types.PhasePlan:
		return t.T(
			"helixspecifier_speckit_topic_plan",
			input.PreviousOutput,
		)
	case types.PhaseTasks:
		return t.T(
			"helixspecifier_speckit_topic_tasks",
			input.PreviousOutput,
		)
	case types.PhaseAnalyze:
		return t.T(
			"helixspecifier_speckit_topic_analyze",
			input.PreviousOutput,
		)
	case types.PhaseImplement:
		return t.T(
			"helixspecifier_speckit_topic_implement",
			input.PreviousOutput,
		)
	default:
		return fmt.Sprintf(
			"phase=%s request=%s",
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
