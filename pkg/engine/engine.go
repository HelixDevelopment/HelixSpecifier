// Package engine provides the core FusionEngine that orchestrates
// the HelixSpecifier spec-driven development flow.
//
// FusionEngine fuses three pillars:
//   - SpecKit: 7-phase SDD workflow (constitution, specify, clarify,
//     plan, tasks, analyze, implement)
//   - Superpowers: TDD discipline and parallel subagent execution
//   - GSD: Get-Shit-Done milestone lifecycle tracking
//
// The engine adapts ceremony level to effort, executes phases in
// order, and learns from completed flows via SpecMemory.
package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FusionEngine orchestrates the 3-pillar spec-driven development
// flow. It fuses SpecKit (7-phase SDD), Superpowers (TDD and
// subagents), and GSD (milestone lifecycle) into a unified engine.
//
// The translator resolves user-facing reasoning text per CONST-046.
// A nil translator is treated as a NoopTranslator at resolution
// time, keeping zero-value and legacy New() callers panic-free.
// Project-not-aware decoupling (CONST-051(B)) is preserved — the
// engine does not bake in any HelixCode-specific defaults.
type FusionEngine struct {
	cfg        *config.Config
	logger     *logrus.Logger
	speckit    types.SpecKitPillar
	powers     types.SuperpowersPillar
	gsd        types.GSDPillar
	ceremony   types.CeremonyScaler
	memory     types.SpecMemory
	classifier types.EffortClassifierFunc
	translator i18n.Translator
	adapters   map[string]types.CLIAgentAdapter
	flows      map[string]*types.FlowResult
	mu         sync.RWMutex
}

// New creates a new FusionEngine with the given configuration
// and logger. The engine defaults to a NoopTranslator; production
// callers SHOULD prefer NewWithTranslator and inject a real
// locale-aware backend.
func New(cfg *config.Config, logger *logrus.Logger) *FusionEngine {
	return &FusionEngine{
		cfg:        cfg,
		logger:     logger,
		translator: i18n.NewNoopTranslator(),
		adapters:   make(map[string]types.CLIAgentAdapter),
		flows:      make(map[string]*types.FlowResult),
	}
}

// NewWithTranslator creates a FusionEngine with an injected
// translator for user-facing strings (CONST-046). Passing nil
// falls back to NoopTranslator, matching the zero-value safety
// contract of New().
func NewWithTranslator(
	cfg *config.Config,
	logger *logrus.Logger,
	t i18n.Translator,
) *FusionEngine {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	return &FusionEngine{
		cfg:        cfg,
		logger:     logger,
		translator: t,
		adapters:   make(map[string]types.CLIAgentAdapter),
		flows:      make(map[string]*types.FlowResult),
	}
}

// SetTranslator swaps the engine's i18n translator at runtime.
// Passing nil falls back to NoopTranslator so the engine remains
// panic-free. Useful when locale changes mid-session.
func (e *FusionEngine) SetTranslator(t i18n.Translator) {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.translator = t
}

// RegisterSpecKit registers the SpecKit pillar implementation.
func (e *FusionEngine) RegisterSpecKit(sk types.SpecKitPillar) {
	e.speckit = sk
}

// RegisterSuperpowers registers the Superpowers pillar
// implementation.
func (e *FusionEngine) RegisterSuperpowers(
	sp types.SuperpowersPillar,
) {
	e.powers = sp
}

// RegisterGSD registers the GSD pillar implementation.
func (e *FusionEngine) RegisterGSD(gsd types.GSDPillar) {
	e.gsd = gsd
}

// RegisterCeremonyScaler registers the ceremony scaler
// implementation.
func (e *FusionEngine) RegisterCeremonyScaler(
	cs types.CeremonyScaler,
) {
	e.ceremony = cs
}

// RegisterSpecMemory registers the spec memory provider.
func (e *FusionEngine) RegisterSpecMemory(sm types.SpecMemory) {
	e.memory = sm
}

// RegisterAdapter registers a CLI agent adapter by name.
func (e *FusionEngine) RegisterAdapter(
	name string,
	adapter types.CLIAgentAdapter,
) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.adapters[name] = adapter
}

// RegisterClassifier registers the effort classifier function.
// When registered, ClassifyEffort delegates to this function
// instead of using the default medium-effort fallback.
func (e *FusionEngine) RegisterClassifier(
	fn types.EffortClassifierFunc,
) {
	e.classifier = fn
}

// SetDebateFunc injects a debate execution function into the
// SpecKit pillar. This allows the host system (e.g. HelixAgent's
// debate service) to provide real multi-LLM debate capability.
func (e *FusionEngine) SetDebateFunc(fn types.DebateFunc) {
	if setter, ok := e.speckit.(types.DebateFuncSetter); ok {
		setter.SetDebateFunc(fn)
		e.logger.Info(
			"[HelixSpecifier] Debate function injected " +
				"into SpecKit pillar",
		)
	}
}

// Name returns the engine name.
func (e *FusionEngine) Name() string {
	return "HelixSpecifier"
}

// Version returns the engine version.
func (e *FusionEngine) Version() string {
	return "1.0.0"
}

// tr returns the engine's translator, defaulting to a
// NoopTranslator when none is wired. Every user-facing string in
// this package resolves through this seam per CONST-046.
func (e *FusionEngine) tr() i18n.Translator {
	if e.translator == nil {
		return i18n.NewNoopTranslator()
	}
	return e.translator
}

// Health returns health status. It returns an error if the
// SpecKit pillar is not registered.
func (e *FusionEngine) Health(ctx context.Context) error {
	if e.speckit == nil {
		return fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_speckit_not_registered",
				"speckit pillar not registered",
			),
		)
	}
	return nil
}

// ClassifyEffort determines the effort level for a request.
// It delegates to the registered classifier if available,
// otherwise falls back to a default medium-effort classification.
// It returns an error if the request is empty.
func (e *FusionEngine) ClassifyEffort(
	ctx context.Context,
	request string,
) (*types.EffortClassification, error) {
	if request == "" {
		return nil, fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_empty_request",
				"empty request",
			),
		)
	}

	var classification *types.EffortClassification

	// Use registered classifier if available
	if e.classifier != nil {
		classification = e.classifier(request)
	} else {
		// Fallback to default — reasoning resolved through
		// the i18n translator (CONST-046). Nil translator
		// defensive fallback mirrors generateReasoning() in
		// pkg/intent/classifier.go.
		t := e.translator
		if t == nil {
			t = i18n.NewNoopTranslator()
		}
		classification = &types.EffortClassification{
			Level:      types.EffortMedium,
			Confidence: 0.5,
			Signals:    []string{"default_classification"},
			Reasoning: t.T(
				"helixspecifier_engine_default_medium_reasoning",
			),
		}
		classification.CeremonyLevel =
			types.CeremonyForEffort(
				classification.Level,
			)
		classification.RequiresDebate =
			classification.Level != types.EffortQuick
		classification.RequiresSpecKit =
			classification.Level == types.EffortLarge ||
				classification.Level == types.EffortEpic
	}

	// Apply ceremony scaler if registered
	if e.ceremony != nil {
		classification.CeremonyLevel = e.ceremony.Scale(
			classification,
		)
	}

	return classification, nil
}

// ExecuteFlow runs the complete spec-driven development flow for
// the given request and classification. It executes phases in
// order, applying adaptive ceremony adjustment when enabled.
func (e *FusionEngine) ExecuteFlow(
	ctx context.Context,
	request string,
	classification *types.EffortClassification,
) (*types.FlowResult, error) {
	if classification == nil {
		return nil, fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_classification_required",
				"classification required",
			),
		)
	}

	flowID := fmt.Sprintf("hsf-%s", uuid.New().String()[:8])
	startTime := time.Now()

	e.logger.WithFields(logrus.Fields{
		"flow_id":  flowID,
		"effort":   classification.Level,
		"ceremony": classification.CeremonyLevel,
	}).Info("[HelixSpecifier] Starting flow")

	result := &types.FlowResult{
		FlowID:        flowID,
		EffortLevel:   classification.Level,
		CeremonyLevel: classification.CeremonyLevel,
		StartTime:     startTime,
		Stages: types.StagesForEffort(
			classification.Level,
		),
		PhaseResults: []types.PhaseResult{},
		Metadata: map[string]interface{}{
			"request":        request,
			"classification": classification,
		},
	}

	// Store flow reference
	e.mu.Lock()
	e.flows[flowID] = result
	e.mu.Unlock()

	// Get phases for this ceremony level
	phases := types.PhasesForCeremony(
		classification.CeremonyLevel,
	)

	// Execute each phase in order
	previousOutput := ""
	for _, phase := range phases {
		phaseInput := &types.PhaseInput{
			UserRequest:    request,
			PreviousOutput: previousOutput,
			Metadata: map[string]interface{}{
				"effort_level":   classification.Level,
				"ceremony_level": classification.CeremonyLevel,
			},
		}

		phaseResult, err := e.executePhase(
			ctx, phase, phaseInput,
		)
		if err != nil {
			result.Success = false
			result.EndTime = time.Now()
			result.Duration = time.Since(startTime)
			return result, fmt.Errorf(
				"phase %s failed: %w", phase, err,
			)
		}

		result.PhaseResults = append(
			result.PhaseResults, *phaseResult,
		)
		previousOutput = phaseResult.Output

		// Adaptive ceremony: check if we should adjust
		if e.cfg.AdaptiveCeremony && e.ceremony != nil {
			newCeremony := e.ceremony.Adjust(
				classification.CeremonyLevel,
				result.PhaseResults,
			)
			if newCeremony != classification.CeremonyLevel {
				e.logger.WithFields(logrus.Fields{
					"from": classification.CeremonyLevel,
					"to":   newCeremony,
				}).Info(
					"[HelixSpecifier] Ceremony adjusted",
				)
				classification.CeremonyLevel = newCeremony
			}
		}
	}

	// Calculate overall quality score
	totalScore := 0.0
	for _, pr := range result.PhaseResults {
		totalScore += pr.QualityScore
	}
	if len(result.PhaseResults) > 0 {
		result.OverallQualityScore = totalScore /
			float64(len(result.PhaseResults))
	}

	// Build specification from phase outputs
	spec := e.buildSpecification(request, result)
	result.Specification = spec

	// Execute tasks via Superpowers if registered
	if e.powers != nil && len(phases) > 1 {
		tasks := e.buildTasksFromPhases(
			result.PhaseResults,
		)
		if len(tasks) > 0 {
			taskResults, dispErr := e.powers.DispatchSubagents(
				ctx, tasks, e.cfg.MaxParallelAgents,
			)
			if dispErr != nil {
				e.logger.WithError(dispErr).Warn(
					"[HelixSpecifier] Superpowers " +
						"dispatch failed",
				)
			} else {
				// Review completed tasks
				review, revErr := e.powers.ReviewCode(
					ctx, taskResults,
				)
				if revErr == nil && review != nil {
					result.Metadata["review_score"] =
						review.Score
					result.Metadata["review_approved"] =
						review.Approved
				}
			}

			// Create milestones via GSD if registered
			if e.gsd != nil && spec != nil {
				milestones, msErr := e.gsd.CreateMilestones(
					ctx, spec, tasks,
				)
				if msErr == nil {
					result.Milestones = milestones
					// Track progress
					if len(milestones) > 0 &&
						len(taskResults) > 0 {
						e.gsd.TrackProgress(
							ctx,
							milestones[0].ID,
							taskResults,
						)
					}
				}
			}
		}
	}

	// Set final artifact from last phase output
	if len(result.PhaseResults) > 0 {
		lastIdx := len(result.PhaseResults) - 1
		result.FinalArtifact =
			result.PhaseResults[lastIdx].Output
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = time.Since(startTime)

	// Learn from completed flow
	if e.memory != nil && e.cfg.SpecMemoryEnabled {
		if err := e.memory.LearnFromFlow(ctx, result); err != nil {
			e.logger.WithError(err).Warn(
				"[HelixSpecifier] Failed to learn from flow",
			)
		}
	}

	e.logger.WithFields(logrus.Fields{
		"flow_id":  flowID,
		"duration": result.Duration,
		"quality":  result.OverallQualityScore,
	}).Info("[HelixSpecifier] Flow completed")

	return result, nil
}

// ResumeFlow resumes a previously cached flow by its ID. It
// returns the flow if found. If the flow already succeeded it
// is returned as-is.
func (e *FusionEngine) ResumeFlow(
	ctx context.Context,
	flowID string,
	request string,
) (*types.FlowResult, error) {
	e.mu.RLock()
	existing, ok := e.flows[flowID]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_flow_not_found",
				"flow %s not found",
				flowID,
			),
		)
	}

	if existing.Success {
		return existing, nil
	}

	existing.ResumedFromCache = true

	e.logger.WithFields(logrus.Fields{
		"flow_id":     flowID,
		"phases_done": len(existing.PhaseResults),
	}).Info("[HelixSpecifier] Resuming flow")

	return existing, nil
}

// GetFlowStatus returns the current status of a flow by its ID.
func (e *FusionEngine) GetFlowStatus(
	flowID string,
) (*types.FlowResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result, ok := e.flows[flowID]
	if !ok {
		return nil, fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_flow_not_found",
				"flow %s not found",
				flowID,
			),
		)
	}
	return result, nil
}

// GetAdapter returns a registered CLI agent adapter by name.
func (e *FusionEngine) GetAdapter(
	name string,
) (types.CLIAgentAdapter, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	adapter, ok := e.adapters[name]
	if !ok {
		return nil, fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_adapter_not_found",
				"adapter %s not found",
				name,
			),
		)
	}
	return adapter, nil
}

// buildSpecification creates a Specification from flow results.
func (e *FusionEngine) buildSpecification(
	request string,
	result *types.FlowResult,
) *types.Specification {
	now := time.Now()
	spec := &types.Specification{
		ID: fmt.Sprintf("spec-%s", result.FlowID),
		Title:       request,
		Description: request,
		Source:      types.SourceFusion,
		CreatedAt:   now,
		UpdatedAt:   now,
		QualityScore: result.OverallQualityScore,
	}

	// Extract phase outputs into requirements
	for _, pr := range result.PhaseResults {
		if pr.Output != "" {
			spec.Requirements = append(
				spec.Requirements,
				types.Requirement{
					ID: fmt.Sprintf(
						"req-%s", pr.Phase,
					),
					Type:        string(pr.Phase),
					Description: pr.Output,
					Priority:    "medium",
				},
			)
		}
	}

	return spec
}

// buildTasksFromPhases creates tasks from phase results.
func (e *FusionEngine) buildTasksFromPhases(
	phases []types.PhaseResult,
) []types.Task {
	tasks := make([]types.Task, 0, len(phases))
	for _, pr := range phases {
		if pr.Output != "" {
			tasks = append(tasks, types.Task{
				ID: fmt.Sprintf(
					"task-%s", pr.Phase,
				),
				Name: fmt.Sprintf(
					"Execute %s", pr.Phase,
				),
				Description: pr.Output,
				Priority:    "medium",
				Status:      "pending",
			})
		}
	}
	return tasks
}

// executePhase runs a single SpecKit phase with the configured
// timeout and quality threshold checking.
func (e *FusionEngine) executePhase(
	ctx context.Context,
	phase types.SpecKitPhase,
	input *types.PhaseInput,
) (*types.PhaseResult, error) {
	if e.speckit == nil {
		return nil, fmt.Errorf(
			"%s",
			i18n.ResolveOrFallback(
				e.tr(),
				"helixspecifier_engine_err_speckit_not_registered",
				"speckit pillar not registered",
			),
		)
	}

	timeout := e.cfg.GetPhaseTimeout(string(phase))
	phaseCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := e.speckit.ExecutePhase(
		phaseCtx, phase, input,
	)
	if err != nil {
		return nil, err
	}

	// Check quality threshold
	if result.QualityScore < e.cfg.MinPhaseScore {
		e.logger.WithFields(logrus.Fields{
			"phase":     phase,
			"score":     result.QualityScore,
			"threshold": e.cfg.MinPhaseScore,
		}).Warn(
			"[HelixSpecifier] Phase below quality threshold",
		)
	}

	return result, nil
}
