package speckit

import (
	"context"
	"errors"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPillar() *Pillar {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return NewPillar(cfg, logger)
}

func newTestInput() *types.PhaseInput {
	return &types.PhaseInput{
		UserRequest:    "Add authentication module",
		PreviousOutput: "Previous phase output",
		Constitution:   "Project constitution text",
		Metadata: map[string]interface{}{
			"project": "test",
		},
	}
}

// --- TestNewPillar ---

func TestNewPillar(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	p := NewPillar(cfg, logger)

	require.NotNil(t, p, "NewPillar should return non-nil")
	assert.Equal(t, cfg, p.cfg,
		"Config should be stored")
	assert.Equal(t, logger, p.logger,
		"Logger should be stored")
	assert.Nil(t, p.DebateFunc,
		"DebateFunc should be nil initially")
}

// --- TestPillar_ExecutePhase_Constitution ---

func TestPillar_ExecutePhase_Constitution(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	input := newTestInput()

	result, err := p.ExecutePhase(
		ctx, types.PhaseConstitution, input,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, types.PhaseConstitution, result.Phase)
	assert.Equal(t, types.StageConstitution, result.Stage)
	assert.Contains(t, result.Output, "constitution")
	assert.Equal(t, 0.75, result.QualityScore)
	assert.NotZero(t, result.Duration)
	assert.NotNil(t, result.Artifacts)
	assert.Equal(t, 5, result.Artifacts["rounds"])
}

// --- TestPillar_ExecutePhase_Specify ---

func TestPillar_ExecutePhase_Specify(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	input := newTestInput()

	result, err := p.ExecutePhase(
		ctx, types.PhaseSpecify, input,
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, types.PhaseSpecify, result.Phase)
	assert.Equal(t, types.StageConstitution, result.Stage)
	assert.Contains(t, result.Output, "specify")
	assert.Equal(t, 3, result.Artifacts["rounds"])
}

// --- TestPillar_ExecutePhase_AllPhases (table-driven) ---

func TestPillar_ExecutePhase_AllPhases(t *testing.T) {
	phases := types.AllPhases()
	expectedStages := map[types.SpecKitPhase]types.FusionStage{
		types.PhaseConstitution: types.StageConstitution,
		types.PhaseSpecify:      types.StageConstitution,
		types.PhaseClarify:      types.StageConstitution,
		types.PhasePlan:         types.StagePlanning,
		types.PhaseTasks:        types.StagePlanning,
		types.PhaseAnalyze:      types.StagePlanning,
		types.PhaseImplement:    types.StageExecution,
	}

	for _, phase := range phases {
		t.Run(string(phase), func(t *testing.T) {
			p := newTestPillar()
			ctx := context.Background()
			input := newTestInput()

			result, err := p.ExecutePhase(
				ctx, phase, input,
			)

			require.NoError(t, err,
				"Phase %s should not error", phase)
			require.NotNil(t, result)
			assert.True(t, result.Success)
			assert.Equal(t, phase, result.Phase)
			assert.Equal(t, expectedStages[phase],
				result.Stage,
				"Phase %s stage mismatch", phase)
			assert.NotEmpty(t, result.Output)
			assert.Greater(t, result.QualityScore, 0.0)
		})
	}
}

// --- TestPillar_ExecutePhase_WithDebateFunc ---

func TestPillar_ExecutePhase_WithDebateFunc(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	input := newTestInput()

	called := false
	p.SetDebateFunc(func(
		_ context.Context,
		topic string,
		rounds int,
		metadata map[string]interface{},
	) (string, float64, string, error) {
		called = true
		assert.NotEmpty(t, topic)
		assert.Greater(t, rounds, 0)
		assert.Contains(t, metadata, "phase")
		return "debate-output", 0.92, "debate-123", nil
	})

	result, err := p.ExecutePhase(
		ctx, types.PhaseConstitution, input,
	)

	require.NoError(t, err)
	assert.True(t, called, "DebateFunc should be called")
	assert.Equal(t, "debate-output", result.Output)
	assert.Equal(t, 0.92, result.QualityScore)
	assert.Equal(t, "debate-123", result.DebateID)
	assert.True(t, result.Success)
}

// --- TestPillar_ExecutePhase_DebateFuncError ---

func TestPillar_ExecutePhase_DebateFuncError(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	input := newTestInput()

	debateErr := errors.New("debate service unavailable")
	p.SetDebateFunc(func(
		_ context.Context,
		_ string,
		_ int,
		_ map[string]interface{},
	) (string, float64, string, error) {
		return "", 0, "", debateErr
	})

	result, err := p.ExecutePhase(
		ctx, types.PhasePlan, input,
	)

	require.Error(t, err)
	assert.Equal(t, debateErr, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "debate failed")
	assert.NotZero(t, result.Duration)
}

// --- TestPillar_GetPhaseOrder_Minimal ---

func TestPillar_GetPhaseOrder_Minimal(t *testing.T) {
	p := newTestPillar()
	phases := p.GetPhaseOrder(types.CeremonyMinimal)

	assert.Len(t, phases, 1)
	assert.Equal(t, types.PhaseImplement, phases[0])
}

// --- TestPillar_GetPhaseOrder_Light ---

func TestPillar_GetPhaseOrder_Light(t *testing.T) {
	p := newTestPillar()
	phases := p.GetPhaseOrder(types.CeremonyLight)

	assert.Len(t, phases, 3)
	expected := []types.SpecKitPhase{
		types.PhaseSpecify,
		types.PhasePlan,
		types.PhaseImplement,
	}
	assert.Equal(t, expected, phases)
}

// --- TestPillar_GetPhaseOrder_Standard ---

func TestPillar_GetPhaseOrder_Standard(t *testing.T) {
	p := newTestPillar()
	phases := p.GetPhaseOrder(types.CeremonyStandard)

	assert.Len(t, phases, 7)
	assert.Equal(t, types.AllPhases(), phases)
}

// --- TestPillar_GetPhaseOrder_Heavy ---

func TestPillar_GetPhaseOrder_Heavy(t *testing.T) {
	p := newTestPillar()
	phases := p.GetPhaseOrder(types.CeremonyHeavy)

	assert.Len(t, phases, 7)
	assert.Equal(t, types.AllPhases(), phases)
}

// --- TestPillar_ValidateTransition_Valid ---

func TestPillar_ValidateTransition_Valid(t *testing.T) {
	p := newTestPillar()
	allPhases := types.AllPhases()

	for i := 0; i < len(allPhases)-1; i++ {
		from := allPhases[i]
		to := allPhases[i+1]
		assert.True(t,
			p.ValidateTransition(from, to),
			"Transition %s -> %s should be valid",
			from, to,
		)
	}
}

// --- TestPillar_ValidateTransition_Invalid ---

func TestPillar_ValidateTransition_Invalid(t *testing.T) {
	p := newTestPillar()

	// Backward transition
	assert.False(t,
		p.ValidateTransition(
			types.PhasePlan, types.PhaseSpecify,
		),
		"Backward transition should be invalid",
	)

	// Same phase
	assert.False(t,
		p.ValidateTransition(
			types.PhasePlan, types.PhasePlan,
		),
		"Same-phase transition should be invalid",
	)
}

// --- TestPillar_ValidateTransition_Skip ---

func TestPillar_ValidateTransition_Skip(t *testing.T) {
	p := newTestPillar()

	// Skipping a phase
	assert.False(t,
		p.ValidateTransition(
			types.PhaseConstitution, types.PhaseClarify,
		),
		"Skipping phases should be invalid",
	)
	assert.False(t,
		p.ValidateTransition(
			types.PhasePlan, types.PhaseImplement,
		),
		"Skipping plan->implement should be invalid",
	)
}

// --- TestPillar_ValidateTransition_Unknown ---

func TestPillar_ValidateTransition_Unknown(t *testing.T) {
	p := newTestPillar()

	assert.False(t,
		p.ValidateTransition(
			types.SpecKitPhase("unknown"),
			types.PhaseSpecify,
		),
		"Unknown 'from' phase should be invalid",
	)
	assert.False(t,
		p.ValidateTransition(
			types.PhaseConstitution,
			types.SpecKitPhase("unknown"),
		),
		"Unknown 'to' phase should be invalid",
	)
	assert.False(t,
		p.ValidateTransition(
			types.SpecKitPhase("foo"),
			types.SpecKitPhase("bar"),
		),
		"Both unknown should be invalid",
	)
}

// --- TestPillar_ImplementsInterface ---

func TestPillar_ImplementsInterface(t *testing.T) {
	var _ types.SpecKitPillar = (*Pillar)(nil)
	p := newTestPillar()
	var iface types.SpecKitPillar = p
	assert.NotNil(t, iface,
		"Pillar must implement SpecKitPillar interface")
}

// --- TestPillar_ExecutePhase_MetadataMerge ---

func TestPillar_ExecutePhase_MetadataMerge(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	receivedMeta := map[string]interface{}{}
	p.SetDebateFunc(func(
		_ context.Context,
		_ string,
		_ int,
		metadata map[string]interface{},
	) (string, float64, string, error) {
		for k, v := range metadata {
			receivedMeta[k] = v
		}
		return "out", 0.8, "d1", nil
	})

	input := &types.PhaseInput{
		UserRequest: "test",
		Metadata: map[string]interface{}{
			"custom_key": "custom_value",
		},
	}

	_, err := p.ExecutePhase(
		ctx, types.PhaseConstitution, input,
	)
	require.NoError(t, err)

	assert.Contains(t, receivedMeta, "phase")
	assert.Contains(t, receivedMeta, "custom_key")
	assert.Equal(t, "custom_value",
		receivedMeta["custom_key"])
}

// --- TestPillar_ExecutePhase_NilMetadata ---

func TestPillar_ExecutePhase_NilMetadata(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	input := &types.PhaseInput{
		UserRequest: "test",
		Metadata:    nil,
	}

	result, err := p.ExecutePhase(
		ctx, types.PhaseImplement, input,
	)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, types.StageExecution, result.Stage)
}

// --- TestPillar_ExecutePhase_Timing ---

func TestPillar_ExecutePhase_Timing(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	input := newTestInput()

	before := time.Now()
	result, err := p.ExecutePhase(
		ctx, types.PhaseConstitution, input,
	)
	after := time.Now()

	require.NoError(t, err)
	assert.False(t, result.StartTime.Before(before),
		"StartTime should be >= test start")
	assert.False(t, result.EndTime.After(after),
		"EndTime should be <= test end")
	assert.True(t, result.Duration >= 0,
		"Duration should be non-negative")
}
