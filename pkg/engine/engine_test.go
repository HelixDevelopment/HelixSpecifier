package engine

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock Implementations ---

type mockSpecKitPillar struct {
	phases map[types.SpecKitPhase]*types.PhaseResult
	err    error
}

func (m *mockSpecKitPillar) ExecutePhase(
	ctx context.Context,
	phase types.SpecKitPhase,
	input *types.PhaseInput,
) (*types.PhaseResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if result, ok := m.phases[phase]; ok {
		return result, nil
	}
	return &types.PhaseResult{
		Phase:        phase,
		StartTime:    time.Now(),
		EndTime:      time.Now(),
		Duration:     100 * time.Millisecond,
		Success:      true,
		Output:       fmt.Sprintf("output for %s", phase),
		QualityScore: 0.8,
	}, nil
}

func (m *mockSpecKitPillar) GetPhaseOrder(
	ceremony types.CeremonyLevel,
) []types.SpecKitPhase {
	return types.PhasesForCeremony(ceremony)
}

func (m *mockSpecKitPillar) ValidateTransition(
	from, to types.SpecKitPhase,
) bool {
	return true
}

type mockCeremonyScaler struct {
	scaleResult  types.CeremonyLevel
	adjustResult types.CeremonyLevel
}

func (m *mockCeremonyScaler) Scale(
	classification *types.EffortClassification,
) types.CeremonyLevel {
	return m.scaleResult
}

func (m *mockCeremonyScaler) Adjust(
	current types.CeremonyLevel,
	phaseResults []types.PhaseResult,
) types.CeremonyLevel {
	return m.adjustResult
}

type mockSpecMemory struct {
	learnErr error
	learned  []*types.FlowResult
	mu       sync.Mutex
}

func (m *mockSpecMemory) Store(
	ctx context.Context,
	spec *types.Specification,
) error {
	return nil
}

func (m *mockSpecMemory) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]types.Specification, error) {
	return nil, nil
}

func (m *mockSpecMemory) LearnFromFlow(
	ctx context.Context,
	flow *types.FlowResult,
) error {
	if m.learnErr != nil {
		return m.learnErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.learned = append(m.learned, flow)
	return nil
}

type mockAdapter struct {
	agentType string
}

func (m *mockAdapter) FormatOutput(
	result *types.FlowResult,
) (string, error) {
	return "formatted", nil
}

func (m *mockAdapter) ParseInput(
	input string,
) (string, error) {
	return input, nil
}

func (m *mockAdapter) GetAgentType() string {
	return m.agentType
}

func (m *mockAdapter) SupportsStreaming() bool {
	return true
}

// --- Helper ---

func newTestEngine() *FusionEngine {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	cfg := config.DefaultConfig()
	return New(cfg, logger)
}

// --- Tests ---

func TestNew(t *testing.T) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	e := New(cfg, logger)

	require.NotNil(t, e)
	assert.NotNil(t, e.cfg)
	assert.NotNil(t, e.logger)
	assert.NotNil(t, e.adapters)
	assert.NotNil(t, e.flows)
	assert.Nil(t, e.speckit)
	assert.Nil(t, e.powers)
	assert.Nil(t, e.gsd)
	assert.Nil(t, e.ceremony)
	assert.Nil(t, e.memory)
}

func TestFusionEngine_Name(t *testing.T) {
	e := newTestEngine()
	assert.Equal(t, "HelixSpecifier", e.Name())
}

func TestFusionEngine_Version(t *testing.T) {
	e := newTestEngine()
	assert.Equal(t, "1.0.0", e.Version())
}

func TestFusionEngine_Health_NoPillar(t *testing.T) {
	e := newTestEngine()
	err := e.Health(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "speckit pillar not registered")
}

func TestFusionEngine_Health_WithPillar(t *testing.T) {
	e := newTestEngine()
	e.RegisterSpecKit(&mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	})
	err := e.Health(context.Background())
	assert.NoError(t, err)
}

func TestFusionEngine_RegisterSpecKit(t *testing.T) {
	e := newTestEngine()
	assert.Nil(t, e.speckit)

	mock := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mock)
	assert.NotNil(t, e.speckit)
}

func TestFusionEngine_RegisterSuperpowers(t *testing.T) {
	e := newTestEngine()
	assert.Nil(t, e.powers)
	// We just verify it sets the field without panic.
	// No mock needed since we only check assignment.
	e.RegisterSuperpowers(nil)
	assert.Nil(t, e.powers)
}

func TestFusionEngine_RegisterGSD(t *testing.T) {
	e := newTestEngine()
	assert.Nil(t, e.gsd)
	e.RegisterGSD(nil)
	assert.Nil(t, e.gsd)
}

func TestFusionEngine_RegisterCeremonyScaler(t *testing.T) {
	e := newTestEngine()
	assert.Nil(t, e.ceremony)

	scaler := &mockCeremonyScaler{
		scaleResult: types.CeremonyStandard,
	}
	e.RegisterCeremonyScaler(scaler)
	assert.NotNil(t, e.ceremony)
}

func TestFusionEngine_RegisterSpecMemory(t *testing.T) {
	e := newTestEngine()
	assert.Nil(t, e.memory)

	mem := &mockSpecMemory{}
	e.RegisterSpecMemory(mem)
	assert.NotNil(t, e.memory)
}

func TestFusionEngine_RegisterAdapter(t *testing.T) {
	e := newTestEngine()
	adapter := &mockAdapter{agentType: "opencode"}
	e.RegisterAdapter("opencode", adapter)

	e.mu.RLock()
	defer e.mu.RUnlock()
	assert.Len(t, e.adapters, 1)
	assert.Equal(t, adapter, e.adapters["opencode"])
}

func TestFusionEngine_GetAdapter_Found(t *testing.T) {
	e := newTestEngine()
	adapter := &mockAdapter{agentType: "crush"}
	e.RegisterAdapter("crush", adapter)

	got, err := e.GetAdapter("crush")
	require.NoError(t, err)
	assert.Equal(t, "crush", got.GetAgentType())
}

func TestFusionEngine_GetAdapter_NotFound(t *testing.T) {
	e := newTestEngine()
	got, err := e.GetAdapter("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "adapter nonexistent not found")
}

func TestFusionEngine_ClassifyEffort(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	classification, err := e.ClassifyEffort(
		ctx, "add a new feature",
	)
	require.NoError(t, err)
	require.NotNil(t, classification)
	assert.Equal(t, types.EffortMedium, classification.Level)
	assert.Equal(t, 0.5, classification.Confidence)
	assert.Equal(
		t, types.CeremonyLight, classification.CeremonyLevel,
	)
	assert.True(t, classification.RequiresDebate)
	assert.False(t, classification.RequiresSpecKit)
	assert.Contains(
		t, classification.Signals, "default_classification",
	)
}

func TestFusionEngine_ClassifyEffort_EmptyRequest(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	classification, err := e.ClassifyEffort(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, classification)
	assert.Contains(t, err.Error(), "empty request")
}

func TestFusionEngine_ExecuteFlow_NilClassification(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	result, err := e.ExecuteFlow(ctx, "test", nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "classification required")
}

func TestFusionEngine_ExecuteFlow_NoPillar(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(ctx, "test", classification)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "speckit pillar not registered")
}

func TestFusionEngine_ExecuteFlow_Success(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(
		ctx, "fix a typo", classification,
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, result.Success)
	assert.NotEmpty(t, result.FlowID)
	assert.True(t, result.FlowID[:4] == "hsf-")
	assert.Equal(t, types.EffortQuick, result.EffortLevel)
	assert.Equal(
		t, types.CeremonyMinimal, result.CeremonyLevel,
	)
	assert.False(t, result.StartTime.IsZero())
	assert.False(t, result.EndTime.IsZero())
	assert.True(t, result.Duration > 0)
	// Minimal ceremony: only implement phase
	assert.Len(t, result.PhaseResults, 1)
	assert.Equal(
		t,
		types.PhaseImplement,
		result.PhaseResults[0].Phase,
	)
	assert.Equal(t, 0.8, result.OverallQualityScore)
}

func TestFusionEngine_ExecuteFlow_PhaseFailure(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
		err: fmt.Errorf("phase execution failed"),
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(
		ctx, "broken request", classification,
	)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "phase implement failed")
	assert.Contains(t, err.Error(), "phase execution failed")
}

func TestFusionEngine_GetFlowStatus_Found(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(
		ctx, "test", classification,
	)
	require.NoError(t, err)

	status, err := e.GetFlowStatus(result.FlowID)
	require.NoError(t, err)
	assert.Equal(t, result.FlowID, status.FlowID)
	assert.True(t, status.Success)
}

func TestFusionEngine_GetFlowStatus_NotFound(t *testing.T) {
	e := newTestEngine()

	status, err := e.GetFlowStatus("nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(
		t, err.Error(), "flow nonexistent-id not found",
	)
}

func TestFusionEngine_ResumeFlow_NotFound(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	result, err := e.ResumeFlow(
		ctx, "nonexistent-id", "request",
	)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(
		t, err.Error(), "flow nonexistent-id not found",
	)
}

func TestFusionEngine_ResumeFlow_AlreadyComplete(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	original, err := e.ExecuteFlow(
		ctx, "test", classification,
	)
	require.NoError(t, err)
	require.True(t, original.Success)

	resumed, err := e.ResumeFlow(
		ctx, original.FlowID, "test",
	)
	require.NoError(t, err)
	assert.Equal(t, original.FlowID, resumed.FlowID)
	assert.True(t, resumed.Success)
	// Already complete flows should not be marked as resumed
	assert.False(t, resumed.ResumedFromCache)
}

func TestFusionEngine_ConcurrentFlows(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	const numFlows = 20
	var wg sync.WaitGroup
	wg.Add(numFlows)

	results := make([]*types.FlowResult, numFlows)
	errors := make([]error, numFlows)

	for i := 0; i < numFlows; i++ {
		go func(idx int) {
			defer wg.Done()
			classification := &types.EffortClassification{
				Level:         types.EffortQuick,
				CeremonyLevel: types.CeremonyMinimal,
			}
			results[idx], errors[idx] = e.ExecuteFlow(
				ctx,
				fmt.Sprintf("task-%d", idx),
				classification,
			)
		}(i)
	}

	wg.Wait()

	// All flows should succeed
	flowIDs := make(map[string]bool)
	for i := 0; i < numFlows; i++ {
		require.NoError(t, errors[i],
			"flow %d should not error", i,
		)
		require.NotNil(t, results[i],
			"flow %d result should not be nil", i,
		)
		assert.True(t, results[i].Success,
			"flow %d should succeed", i,
		)
		// Each flow should have a unique ID
		assert.False(t, flowIDs[results[i].FlowID],
			"flow ID %s should be unique", results[i].FlowID,
		)
		flowIDs[results[i].FlowID] = true
	}

	// All flows should be retrievable
	for id := range flowIDs {
		status, err := e.GetFlowStatus(id)
		require.NoError(t, err)
		assert.True(t, status.Success)
	}
}

func TestFusionEngine_AdaptiveCeremony(t *testing.T) {
	e := newTestEngine()
	e.cfg.AdaptiveCeremony = true
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	// Scaler that adjusts ceremony from light to standard
	scaler := &mockCeremonyScaler{
		scaleResult:  types.CeremonyLight,
		adjustResult: types.CeremonyStandard,
	}
	e.RegisterCeremonyScaler(scaler)

	classification := &types.EffortClassification{
		Level:         types.EffortMedium,
		CeremonyLevel: types.CeremonyLight,
	}

	result, err := e.ExecuteFlow(
		ctx, "medium task", classification,
	)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Light ceremony: specify + plan + implement = 3 phases
	// After first phase the scaler returns standard, but phases
	// were already determined at the start of the flow. The
	// ceremony level on classification gets updated though.
	assert.Equal(
		t,
		types.CeremonyStandard,
		classification.CeremonyLevel,
	)
}

func TestFusionEngine_ExecuteFlow_WithMemory(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	mem := &mockSpecMemory{}
	e.RegisterSpecMemory(mem)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(
		ctx, "test with memory", classification,
	)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Memory should have learned from the flow
	mem.mu.Lock()
	defer mem.mu.Unlock()
	require.Len(t, mem.learned, 1)
	assert.Equal(t, result.FlowID, mem.learned[0].FlowID)
}

func TestFusionEngine_ExecuteFlow_MemoryLearnError(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	mem := &mockSpecMemory{
		learnErr: fmt.Errorf("memory learn failed"),
	}
	e.RegisterSpecMemory(mem)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	// Memory failure should not cause flow failure
	result, err := e.ExecuteFlow(
		ctx, "test with broken memory", classification,
	)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestFusionEngine_ExecuteFlow_MultiplePhases(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: map[types.SpecKitPhase]*types.PhaseResult{
			types.PhaseSpecify: {
				Phase:        types.PhaseSpecify,
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				Duration:     50 * time.Millisecond,
				Success:      true,
				Output:       "specification output",
				QualityScore: 0.9,
			},
			types.PhasePlan: {
				Phase:        types.PhasePlan,
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				Duration:     75 * time.Millisecond,
				Success:      true,
				Output:       "plan output",
				QualityScore: 0.7,
			},
			types.PhaseImplement: {
				Phase:        types.PhaseImplement,
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				Duration:     200 * time.Millisecond,
				Success:      true,
				Output:       "implementation output",
				QualityScore: 0.8,
			},
		},
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortMedium,
		CeremonyLevel: types.CeremonyLight,
	}

	result, err := e.ExecuteFlow(
		ctx, "medium task", classification,
	)
	require.NoError(t, err)
	assert.True(t, result.Success)
	// Light ceremony: specify + plan + implement
	assert.Len(t, result.PhaseResults, 3)

	// Quality score should be average: (0.9 + 0.7 + 0.8) / 3
	expectedQuality := (0.9 + 0.7 + 0.8) / 3.0
	assert.InDelta(
		t, expectedQuality, result.OverallQualityScore, 0.001,
	)
}

func TestFusionEngine_ClassifyEffort_WithCeremonyScaler(
	t *testing.T,
) {
	e := newTestEngine()
	ctx := context.Background()

	scaler := &mockCeremonyScaler{
		scaleResult: types.CeremonyHeavy,
	}
	e.RegisterCeremonyScaler(scaler)

	classification, err := e.ClassifyEffort(
		ctx, "large refactoring task",
	)
	require.NoError(t, err)
	assert.Equal(
		t, types.CeremonyHeavy, classification.CeremonyLevel,
	)
}

func TestFusionEngine_ResumeFlow_InProgress(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	// Manually insert an incomplete flow
	flow := &types.FlowResult{
		FlowID:      "hsf-test1234",
		EffortLevel: types.EffortLarge,
		StartTime:   time.Now(),
		Success:     false,
		PhaseResults: []types.PhaseResult{
			{
				Phase:        types.PhaseConstitution,
				Success:      true,
				QualityScore: 0.8,
			},
		},
	}
	e.mu.Lock()
	e.flows["hsf-test1234"] = flow
	e.mu.Unlock()

	resumed, err := e.ResumeFlow(
		ctx, "hsf-test1234", "continue work",
	)
	require.NoError(t, err)
	assert.True(t, resumed.ResumedFromCache)
	assert.Equal(t, "hsf-test1234", resumed.FlowID)
	assert.Len(t, resumed.PhaseResults, 1)
}

func TestFusionEngine_ExecuteFlow_PhaseBelowQualityThreshold(
	t *testing.T,
) {
	e := newTestEngine()
	ctx := context.Background()

	// Mock that returns a low quality score
	mockSK := &mockSpecKitPillar{
		phases: map[types.SpecKitPhase]*types.PhaseResult{
			types.PhaseImplement: {
				Phase:        types.PhaseImplement,
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				Duration:     50 * time.Millisecond,
				Success:      true,
				Output:       "low quality output",
				QualityScore: 0.2, // Below default 0.5
			},
		},
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(
		ctx, "low quality task", classification,
	)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.InDelta(t, 0.2, result.OverallQualityScore, 0.001)
}

// mockDebatablePillar implements both SpecKitPillar and
// DebateFuncSetter.
type mockDebatablePillar struct {
	mockSpecKitPillar
	debateFunc types.DebateFunc
}

func (m *mockDebatablePillar) SetDebateFunc(
	fn types.DebateFunc,
) {
	m.debateFunc = fn
}

func TestFusionEngine_SetDebateFunc_Supported(t *testing.T) {
	e := newTestEngine()
	pillar := &mockDebatablePillar{
		mockSpecKitPillar: mockSpecKitPillar{
			phases: make(
				map[types.SpecKitPhase]*types.PhaseResult,
			),
		},
	}
	e.RegisterSpecKit(pillar)

	called := false
	fn := func(
		_ context.Context,
		_ string,
		_ int,
		_ map[string]interface{},
	) (string, float64, string, error) {
		called = true
		return "debated", 0.95, "d-1", nil
	}

	e.SetDebateFunc(fn)
	assert.NotNil(t, pillar.debateFunc)

	// Invoke through the pillar to verify injection
	out, score, id, err := pillar.debateFunc(
		context.Background(), "topic", 3, nil,
	)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "debated", out)
	assert.Equal(t, 0.95, score)
	assert.Equal(t, "d-1", id)
}

func TestFusionEngine_SetDebateFunc_NotSupported(t *testing.T) {
	e := newTestEngine()
	// mockSpecKitPillar does NOT implement DebateFuncSetter
	pillar := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(pillar)

	// Should not panic; just silently skip injection
	e.SetDebateFunc(func(
		_ context.Context,
		_ string,
		_ int,
		_ map[string]interface{},
	) (string, float64, string, error) {
		return "", 0, "", nil
	})
}

func TestFusionEngine_SetDebateFunc_NilSpecKit(t *testing.T) {
	e := newTestEngine()
	// No SpecKit registered -- should not panic
	e.SetDebateFunc(func(
		_ context.Context,
		_ string,
		_ int,
		_ map[string]interface{},
	) (string, float64, string, error) {
		return "", 0, "", nil
	})
}

func TestFusionEngine_ExecuteFlow_FlowMetadata(t *testing.T) {
	e := newTestEngine()
	ctx := context.Background()

	mockSK := &mockSpecKitPillar{
		phases: make(
			map[types.SpecKitPhase]*types.PhaseResult,
		),
	}
	e.RegisterSpecKit(mockSK)

	classification := &types.EffortClassification{
		Level:         types.EffortQuick,
		CeremonyLevel: types.CeremonyMinimal,
	}

	result, err := e.ExecuteFlow(
		ctx, "test metadata", classification,
	)
	require.NoError(t, err)
	require.NotNil(t, result.Metadata)
	assert.Equal(
		t, "test metadata", result.Metadata["request"],
	)
	assert.NotNil(t, result.Metadata["classification"])
}
