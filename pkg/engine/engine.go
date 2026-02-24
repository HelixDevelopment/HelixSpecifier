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
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// FusionEngine orchestrates the 3-pillar spec-driven development
// flow. It fuses SpecKit (7-phase SDD), Superpowers (TDD and
// subagents), and GSD (milestone lifecycle) into a unified engine.
type FusionEngine struct {
	cfg      *config.Config
	logger   *logrus.Logger
	speckit  types.SpecKitPillar
	powers   types.SuperpowersPillar
	gsd      types.GSDPillar
	ceremony types.CeremonyScaler
	memory   types.SpecMemory
	adapters map[string]types.CLIAgentAdapter
	flows    map[string]*types.FlowResult
	mu       sync.RWMutex
}

// New creates a new FusionEngine with the given configuration
// and logger.
func New(cfg *config.Config, logger *logrus.Logger) *FusionEngine {
	return &FusionEngine{
		cfg:      cfg,
		logger:   logger,
		adapters: make(map[string]types.CLIAgentAdapter),
		flows:    make(map[string]*types.FlowResult),
	}
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

// Health returns health status. It returns an error if the
// SpecKit pillar is not registered.
func (e *FusionEngine) Health(ctx context.Context) error {
	if e.speckit == nil {
		return fmt.Errorf("speckit pillar not registered")
	}
	return nil
}

// ClassifyEffort determines the effort level for a request.
// It returns an error if the request is empty.
func (e *FusionEngine) ClassifyEffort(
	ctx context.Context,
	request string,
) (*types.EffortClassification, error) {
	if request == "" {
		return nil, fmt.Errorf("empty request")
	}

	classification := &types.EffortClassification{
		Level:      types.EffortMedium,
		Confidence: 0.5,
		Signals:    []string{"default_classification"},
		Reasoning:  "Default medium effort classification",
	}

	// Determine ceremony level from effort
	classification.CeremonyLevel = types.CeremonyForEffort(
		classification.Level,
	)
	classification.RequiresDebate =
		classification.Level != types.EffortQuick
	classification.RequiresSpecKit =
		classification.Level == types.EffortLarge ||
			classification.Level == types.EffortEpic

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
		return nil, fmt.Errorf("classification required")
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
		return nil, fmt.Errorf("flow %s not found", flowID)
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
		return nil, fmt.Errorf("flow %s not found", flowID)
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
		return nil, fmt.Errorf("adapter %s not found", name)
	}
	return adapter, nil
}

// executePhase runs a single SpecKit phase with the configured
// timeout and quality threshold checking.
func (e *FusionEngine) executePhase(
	ctx context.Context,
	phase types.SpecKitPhase,
	input *types.PhaseInput,
) (*types.PhaseResult, error) {
	if e.speckit == nil {
		return nil, fmt.Errorf("speckit pillar not registered")
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
