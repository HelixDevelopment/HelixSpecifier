package integration

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/adapters"
	"digital.vasic.helixspecifier/pkg/ceremony"
	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/engine"
	"digital.vasic.helixspecifier/pkg/features/adaptive_ceremony"
	"digital.vasic.helixspecifier/pkg/features/brownfield"
	"digital.vasic.helixspecifier/pkg/features/constitution_code"
	"digital.vasic.helixspecifier/pkg/features/cross_project"
	"digital.vasic.helixspecifier/pkg/features/debate_architecture"
	"digital.vasic.helixspecifier/pkg/features/nyquist_tdd"
	"digital.vasic.helixspecifier/pkg/features/parallel_execution"
	"digital.vasic.helixspecifier/pkg/features/predictive_spec"
	"digital.vasic.helixspecifier/pkg/features/skill_learning"
	"digital.vasic.helixspecifier/pkg/features/spec_memory"
	"digital.vasic.helixspecifier/pkg/gsd"
	"digital.vasic.helixspecifier/pkg/intent"
	"digital.vasic.helixspecifier/pkg/memory"
	"digital.vasic.helixspecifier/pkg/metrics"
	"digital.vasic.helixspecifier/pkg/speckit"
	"digital.vasic.helixspecifier/pkg/superpowers"
	"digital.vasic.helixspecifier/pkg/types"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// newTestLogger returns a silent logger for tests.
func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	return logger
}

// newFullEngine creates a fully wired FusionEngine with all
// three real pillars, ceremony scaler, and spec memory.
func newFullEngine() *engine.FusionEngine {
	cfg := config.DefaultConfig()
	logger := newTestLogger()

	eng := engine.New(cfg, logger)

	sk := speckit.NewPillar(cfg, logger)
	sp := superpowers.NewPillar(cfg, logger)
	g := gsd.NewPillar(logger)
	cs := ceremony.NewScaler(cfg, logger)
	mem := memory.NewStore()

	eng.RegisterSpecKit(sk)
	eng.RegisterSuperpowers(sp)
	eng.RegisterGSD(g)
	eng.RegisterCeremonyScaler(cs)
	eng.RegisterSpecMemory(mem)

	return eng
}

// sampleTasks creates a slice of tasks for testing.
func sampleTasks(count int) []types.Task {
	tasks := make([]types.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = types.Task{
			ID:          fmt.Sprintf("task-%d", i+1),
			Name:        fmt.Sprintf("Task %d", i+1),
			Description: fmt.Sprintf("Description for task %d", i+1),
			Priority:    "medium",
			Status:      "pending",
			Acceptance:  []string{"criterion-1", "criterion-2"},
			FilesAffected: []string{
				fmt.Sprintf("pkg/module%d/file.go", i+1),
			},
		}
	}
	return tasks
}

// sampleSpec creates a specification for testing.
func sampleSpec(title string) *types.Specification {
	return &types.Specification{
		ID:          "spec-001",
		Title:       title,
		Description: "Test specification for integration tests",
		Source:      types.SourceFusion,
		CreatedAt:   time.Now(),
		Requirements: []types.Requirement{
			{
				ID:          "req-1",
				Type:        "functional",
				Description: "Must handle concurrent requests",
				Priority:    "high",
				Acceptance:  "All concurrent requests succeed",
			},
		},
		Constraints: []types.Constraint{
			{
				Type:        "performance",
				Description: "Response time under 100ms",
			},
		},
		QualityScore: 0.8,
	}
}

// --- 1. Full Flow Integration ---

func TestFullFlowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	eng := newFullEngine()
	ctx := context.Background()

	t.Run("QuickEffortFlow", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortQuick,
			Confidence:    0.9,
			CeremonyLevel: types.CeremonyMinimal,
			Signals:       []string{"quick_fix"},
		}

		result, err := eng.ExecuteFlow(
			ctx, "fix typo in README", cl,
		)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Success)
		assert.Equal(t, types.EffortQuick, result.EffortLevel)
		assert.Equal(t, types.CeremonyMinimal, result.CeremonyLevel)
		assert.NotEmpty(t, result.FlowID)
		assert.True(t, strings.HasPrefix(result.FlowID, "hsf-"))
		assert.Greater(t, len(result.PhaseResults), 0)
		assert.True(t, result.Duration > 0)
		assert.Greater(t, result.OverallQualityScore, 0.0)
	})

	t.Run("MediumEffortFlow", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortMedium,
			Confidence:    0.7,
			CeremonyLevel: types.CeremonyLight,
			Signals:       []string{"moderate_task"},
		}

		result, err := eng.ExecuteFlow(
			ctx, "add validation to user input handler", cl,
		)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Success)
		assert.Equal(t, types.EffortMedium, result.EffortLevel)
		// Light ceremony: specify, plan, implement
		assert.Equal(t, 3, len(result.PhaseResults))

		for _, pr := range result.PhaseResults {
			assert.True(t, pr.Success)
			assert.NotEmpty(t, pr.Output)
			assert.Greater(t, pr.QualityScore, 0.0)
		}
	})

	t.Run("LargeEffortFlow", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortLarge,
			Confidence:    0.8,
			CeremonyLevel: types.CeremonyStandard,
			Signals:       []string{"new_module"},
		}

		result, err := eng.ExecuteFlow(
			ctx,
			"implement new authentication service with OAuth2",
			cl,
		)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Success)
		// Standard ceremony: all 7 phases
		assert.Equal(t, 7, len(result.PhaseResults))

		phases := make([]types.SpecKitPhase, len(result.PhaseResults))
		for i, pr := range result.PhaseResults {
			phases[i] = pr.Phase
		}
		assert.Equal(t, types.AllPhases(), phases)
	})

	t.Run("EpicEffortFlow", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortEpic,
			Confidence:    0.9,
			CeremonyLevel: types.CeremonyHeavy,
			Signals:       []string{"full_platform"},
		}

		result, err := eng.ExecuteFlow(
			ctx,
			"rebuild entire system from scratch with microservices",
			cl,
		)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Success)
		// Heavy ceremony also uses all 7 phases
		assert.Equal(t, 7, len(result.PhaseResults))
		assert.Equal(t, types.EffortEpic, result.EffortLevel)
	})

	t.Run("FlowContainsAllStages", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortLarge,
			Confidence:    0.8,
			CeremonyLevel: types.CeremonyStandard,
			Signals:       []string{"test"},
		}

		result, err := eng.ExecuteFlow(
			ctx, "build new feature end-to-end", cl,
		)
		require.NoError(t, err)

		assert.Equal(
			t,
			types.StagesForEffort(types.EffortLarge),
			result.Stages,
		)
	})

	t.Run("FlowWithNilClassificationFails", func(t *testing.T) {
		result, err := eng.ExecuteFlow(ctx, "anything", nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "classification required")
	})
}

// --- 2. Effort Classification Pipeline ---

func TestEffortClassificationPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	classifier := intent.NewClassifier()
	cfg := config.DefaultConfig()
	logger := newTestLogger()
	scaler := ceremony.NewScaler(cfg, logger)

	cases := []struct {
		name             string
		request          string
		expectedEffort   types.EffortLevel
		expectedCeremony types.CeremonyLevel
		minPhases        int
		maxPhases        int
	}{
		{
			name:             "QuickFix",
			request:          "fix typo in the README file",
			expectedEffort:   types.EffortQuick,
			expectedCeremony: types.CeremonyMinimal,
			minPhases:        1,
			maxPhases:        1,
		},
		{
			name:             "MediumTask",
			request:          "add a new logging middleware",
			expectedEffort:   types.EffortMedium,
			expectedCeremony: types.CeremonyLight,
			minPhases:        3,
			maxPhases:        3,
		},
		{
			name:             "LargeWithSignals",
			request:          "implement new authentication service for the API",
			expectedEffort:   types.EffortLarge,
			expectedCeremony: types.CeremonyStandard,
			minPhases:        7,
			maxPhases:        7,
		},
		{
			name:             "LargeFeature",
			request:          "implement new authentication service with database and API endpoints",
			expectedEffort:   types.EffortLarge,
			expectedCeremony: types.CeremonyStandard,
			minPhases:        7,
			maxPhases:        7,
		},
		{
			name:             "EpicRewrite",
			request:          "rebuild entire system from scratch with new architecture",
			expectedEffort:   types.EffortEpic,
			expectedCeremony: types.CeremonyHeavy,
			minPhases:        7,
			maxPhases:        7,
		},
		{
			name:             "RefactorBoostsToLarge",
			request:          "refactor the handler code",
			expectedEffort:   types.EffortLarge,
			expectedCeremony: types.CeremonyStandard,
			minPhases:        7,
			maxPhases:        7,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cl := classifier.Classify(tc.request)
			require.NotNil(t, cl)

			assert.Equal(t, tc.expectedEffort, cl.Level)
			assert.NotNil(t, cl.Signals)
			assert.NotEmpty(t, cl.Reasoning)
			assert.Greater(t, cl.Confidence, 0.0)

			ceremonyLevel := scaler.Scale(cl)
			assert.Equal(t, tc.expectedCeremony, ceremonyLevel)

			phases := types.PhasesForCeremony(ceremonyLevel)
			assert.GreaterOrEqual(t, len(phases), tc.minPhases)
			assert.LessOrEqual(t, len(phases), tc.maxPhases)
		})
	}

	t.Run("ClassifierIntegrationWithEngine", func(t *testing.T) {
		eng := newFullEngine()
		ctx := context.Background()

		cl, err := eng.ClassifyEffort(
			ctx, "add a new module for the user service",
		)
		require.NoError(t, err)
		require.NotNil(t, cl)

		assert.NotEmpty(t, string(cl.Level))
		assert.NotEmpty(t, string(cl.CeremonyLevel))
		assert.Greater(t, cl.Confidence, 0.0)
	})

	t.Run("EmptyRequestFails", func(t *testing.T) {
		eng := newFullEngine()
		ctx := context.Background()

		cl, err := eng.ClassifyEffort(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, cl)
		assert.Contains(t, err.Error(), "empty request")
	})
}

// --- 3. Ceremony Adaptation Integration ---

func TestCeremonyAdaptationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()
	cfg.AdaptiveCeremony = true
	logger := newTestLogger()

	t.Run("CeremonyRemainsWhenQualityNormal", func(t *testing.T) {
		eng := engine.New(cfg, logger)

		sk := speckit.NewPillar(cfg, logger)
		cs := ceremony.NewScaler(cfg, logger)
		eng.RegisterSpecKit(sk)
		eng.RegisterCeremonyScaler(cs)

		ctx := context.Background()
		cl := &types.EffortClassification{
			Level:         types.EffortLarge,
			Confidence:    0.8,
			CeremonyLevel: types.CeremonyStandard,
		}

		result, err := eng.ExecuteFlow(
			ctx, "build new feature module", cl,
		)
		require.NoError(t, err)
		assert.True(t, result.Success)
		// Default phase score is 0.75, within normal range
		// so ceremony should remain standard
		assert.Equal(t, types.CeremonyStandard, result.CeremonyLevel)
	})

	t.Run("CeremonyAdjusterFeatureIntegration", func(t *testing.T) {
		adjuster := adaptive_ceremony.NewAdjuster(0.85, 0.5)

		// Simulate high-quality phases
		highResults := []types.PhaseResult{
			{Phase: types.PhaseConstitution, QualityScore: 0.95, Success: true},
			{Phase: types.PhaseSpecify, QualityScore: 0.92, Success: true},
		}

		newLevel := adjuster.Adjust(
			types.CeremonyHeavy, highResults,
		)
		assert.Equal(t, types.CeremonyStandard, newLevel)
		assert.Equal(t, 1, adjuster.AdjustmentCount())

		// Simulate low-quality phases
		lowResults := []types.PhaseResult{
			{Phase: types.PhaseSpecify, QualityScore: 0.3, Success: true},
			{Phase: types.PhasePlan, QualityScore: 0.4, Success: true},
		}

		newLevel = adjuster.Adjust(
			types.CeremonyLight, lowResults,
		)
		assert.Equal(t, types.CeremonyStandard, newLevel)
		assert.Equal(t, 2, adjuster.AdjustmentCount())
	})

	t.Run("NoAdjustmentWhenQualityMidRange", func(t *testing.T) {
		adjuster := adaptive_ceremony.NewAdjuster(0.9, 0.5)

		midResults := []types.PhaseResult{
			{Phase: types.PhaseSpecify, QualityScore: 0.75, Success: true},
			{Phase: types.PhasePlan, QualityScore: 0.70, Success: true},
		}

		level := adjuster.Adjust(
			types.CeremonyStandard, midResults,
		)
		assert.Equal(t, types.CeremonyStandard, level)
		assert.Equal(t, 0, adjuster.AdjustmentCount())
	})
}

// --- 4. Multi-Adapter Output ---

func TestMultiAdapterOutputIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	eng := newFullEngine()
	ctx := context.Background()

	cl := &types.EffortClassification{
		Level:         types.EffortMedium,
		Confidence:    0.7,
		CeremonyLevel: types.CeremonyLight,
	}

	result, err := eng.ExecuteFlow(
		ctx, "add input validation layer", cl,
	)
	require.NoError(t, err)
	require.True(t, result.Success)

	categories := []struct {
		name     string
		category adapters.AgentCategory
	}{
		{"Markdown", adapters.CategoryMarkdown},
		{"TOML", adapters.CategoryTOML},
		{"GitHub", adapters.CategoryGitHub},
		{"Minimal", adapters.CategoryMinimal},
		{"Generic", adapters.CategoryGeneric},
	}

	for _, cat := range categories {
		t.Run(cat.name, func(t *testing.T) {
			adapter := adapters.NewGenericAdapter(
				strings.ToLower(cat.name), cat.category,
			)
			require.NotNil(t, adapter)

			output, err := adapter.FormatOutput(result)
			require.NoError(t, err)
			assert.NotEmpty(t, output)

			assert.Equal(
				t,
				strings.ToLower(cat.name),
				adapter.GetAgentType(),
			)
			assert.True(t, adapter.SupportsStreaming())

			parsed, err := adapter.ParseInput("test input")
			require.NoError(t, err)
			assert.Equal(t, "test input", parsed)
		})
	}

	t.Run("MarkdownContainsFlowID", func(t *testing.T) {
		adapter := adapters.NewGenericAdapter(
			"test", adapters.CategoryMarkdown,
		)
		output, err := adapter.FormatOutput(result)
		require.NoError(t, err)
		assert.Contains(t, output, result.FlowID)
		assert.Contains(t, output, "# HelixSpecifier Flow")
	})

	t.Run("TOMLContainsFlowFields", func(t *testing.T) {
		adapter := adapters.NewGenericAdapter(
			"test", adapters.CategoryTOML,
		)
		output, err := adapter.FormatOutput(result)
		require.NoError(t, err)
		assert.Contains(t, output, "[flow]")
		assert.Contains(t, output, "success = true")
	})

	t.Run("GenericProducesValidJSON", func(t *testing.T) {
		adapter := adapters.NewGenericAdapter(
			"test", adapters.CategoryGeneric,
		)
		output, err := adapter.FormatOutput(result)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(output, "{"))
		assert.True(t, strings.HasSuffix(
			strings.TrimSpace(output), "}",
		))
	})

	t.Run("NilResultFails", func(t *testing.T) {
		adapter := adapters.NewGenericAdapter(
			"test", adapters.CategoryMarkdown,
		)
		_, err := adapter.FormatOutput(nil)
		assert.Error(t, err)
	})

	t.Run("WellKnownAdaptersAreRegistered", func(t *testing.T) {
		known := adapters.WellKnownAdapters()
		assert.Greater(t, len(known), 0)

		expectedAgents := []string{
			"opencode", "crush", "claude", "cursor",
		}
		for _, agent := range expectedAgents {
			a, ok := known[agent]
			assert.True(t, ok, "missing adapter for %s", agent)
			assert.Equal(t, agent, a.GetAgentType())
		}
	})

	t.Run("AdapterRegisteredOnEngine", func(t *testing.T) {
		adapter := adapters.NewGenericAdapter(
			"integration-test", adapters.CategoryGeneric,
		)
		eng.RegisterAdapter("integration-test", adapter)

		retrieved, err := eng.GetAdapter("integration-test")
		require.NoError(t, err)
		assert.Equal(t, "integration-test", retrieved.GetAgentType())
	})
}

// --- 5. Spec Memory Persistence ---

func TestSpecMemoryPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("StoreAndSearchSpecs", func(t *testing.T) {
		store := memory.NewStore()

		specs := []*types.Specification{
			sampleSpec("Authentication Module"),
			{
				ID:          "spec-002",
				Title:       "Database Migration System",
				Description: "Automated migration framework",
				Source:      types.SourceSpecKit,
				CreatedAt:   time.Now(),
			},
			{
				ID:          "spec-003",
				Title:       "REST API Gateway",
				Description: "API gateway with rate limiting",
				Source:      types.SourceUser,
				CreatedAt:   time.Now(),
			},
		}

		for _, spec := range specs {
			err := store.Store(ctx, spec)
			require.NoError(t, err)
		}

		assert.Equal(t, 3, store.SpecCount())

		results, err := store.Search(ctx, "auth", 10)
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "Authentication Module", results[0].Title)

		results, err = store.Search(ctx, "api", 10)
		require.NoError(t, err)
		assert.Equal(t, 1, len(results))

		results, err = store.Search(ctx, "nonexistent", 10)
		require.NoError(t, err)
		assert.Equal(t, 0, len(results))
	})

	t.Run("LearnFromSuccessfulFlow", func(t *testing.T) {
		store := memory.NewStore()

		flow := &types.FlowResult{
			FlowID:              "hsf-test-001",
			EffortLevel:         types.EffortLarge,
			CeremonyLevel:       types.CeremonyStandard,
			Success:             true,
			OverallQualityScore: 0.85,
			PhaseResults: []types.PhaseResult{
				{Phase: types.PhaseSpecify, QualityScore: 0.9},
				{Phase: types.PhasePlan, QualityScore: 0.8},
			},
		}

		err := store.LearnFromFlow(ctx, flow)
		require.NoError(t, err)
		assert.Equal(t, 1, store.PatternCount())

		patterns := store.GetPatterns()
		assert.Equal(t, 1, len(patterns))
		assert.Equal(t, "hsf-test-001", patterns[0].ID)
		assert.Contains(t, patterns[0].Pattern, "effort=large")
		assert.Contains(t, patterns[0].Pattern, "ceremony=standard")
	})

	t.Run("SkipLearningFromFailedFlow", func(t *testing.T) {
		store := memory.NewStore()

		failedFlow := &types.FlowResult{
			FlowID:  "hsf-failed",
			Success: false,
		}

		err := store.LearnFromFlow(ctx, failedFlow)
		require.NoError(t, err)
		assert.Equal(t, 0, store.PatternCount())
	})

	t.Run("EngineLearnsThroughFlow", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.SpecMemoryEnabled = true
		logger := newTestLogger()

		eng := engine.New(cfg, logger)
		store := memory.NewStore()

		eng.RegisterSpecKit(speckit.NewPillar(cfg, logger))
		eng.RegisterSpecMemory(store)

		cl := &types.EffortClassification{
			Level:         types.EffortMedium,
			Confidence:    0.7,
			CeremonyLevel: types.CeremonyLight,
		}

		result, err := eng.ExecuteFlow(
			ctx, "add new endpoint", cl,
		)
		require.NoError(t, err)
		require.True(t, result.Success)

		// Engine should have learned from the successful flow
		assert.Equal(t, 1, store.PatternCount())
	})

	t.Run("NilSpecRejected", func(t *testing.T) {
		store := memory.NewStore()
		err := store.Store(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("NilFlowRejected", func(t *testing.T) {
		store := memory.NewStore()
		err := store.LearnFromFlow(ctx, nil)
		assert.Error(t, err)
	})
}

// --- 6. GSD Milestone Lifecycle ---

func TestGSDMilestoneLifecycleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	logger := newTestLogger()
	pillar := gsd.NewPillar(logger)

	t.Run("CreateAndTrackMilestones", func(t *testing.T) {
		spec := sampleSpec("Feature Module")
		tasks := []types.Task{
			{
				ID:       "t-1",
				Name:     "Setup",
				Priority: "critical",
			},
			{
				ID:       "t-2",
				Name:     "Implement",
				Priority: "high",
			},
			{
				ID:       "t-3",
				Name:     "Test",
				Priority: "medium",
			},
		}

		milestones, err := pillar.CreateMilestones(
			ctx, spec, tasks,
		)
		require.NoError(t, err)
		assert.Equal(t, 3, len(milestones))

		for _, ms := range milestones {
			assert.NotEmpty(t, ms.ID)
			assert.True(t, strings.HasPrefix(ms.ID, "ms-"))
			assert.Equal(t, types.MilestoneQueued, ms.Status)
			assert.Equal(t, 0.0, ms.Progress)
		}
	})

	t.Run("ProgressTracking", func(t *testing.T) {
		spec := sampleSpec("Progress Test")
		tasks := []types.Task{
			{ID: "t-1", Name: "Task A", Priority: "high"},
			{ID: "t-2", Name: "Task B", Priority: "high"},
		}

		milestones, err := pillar.CreateMilestones(
			ctx, spec, tasks,
		)
		require.NoError(t, err)
		require.Equal(t, 1, len(milestones))

		msID := milestones[0].ID

		// Complete one of two tasks
		results := []types.TaskResult{
			{TaskID: "t-1", Success: true},
		}

		ms, err := pillar.TrackProgress(ctx, msID, results)
		require.NoError(t, err)
		assert.Equal(t, types.MilestoneActive, ms.Status)
		assert.Equal(t, 0.5, ms.Progress)
		assert.NotNil(t, ms.StartedAt)

		// Complete both tasks
		allResults := []types.TaskResult{
			{TaskID: "t-1", Success: true},
			{TaskID: "t-2", Success: true},
		}

		ms, err = pillar.TrackProgress(ctx, msID, allResults)
		require.NoError(t, err)
		assert.Equal(t, types.MilestoneCompleted, ms.Status)
		assert.Equal(t, 1.0, ms.Progress)
		assert.NotNil(t, ms.CompletedAt)
	})

	t.Run("MilestoneDependencyChaining", func(t *testing.T) {
		spec := sampleSpec("Dependency Chain")
		tasks := []types.Task{
			{ID: "t-crit", Name: "Critical", Priority: "critical"},
			{ID: "t-high", Name: "High", Priority: "high"},
			{ID: "t-med", Name: "Medium", Priority: "medium"},
		}

		milestones, err := pillar.CreateMilestones(
			ctx, spec, tasks,
		)
		require.NoError(t, err)
		assert.Equal(t, 3, len(milestones))

		// First milestone has no dependencies
		assert.Empty(t, milestones[0].Dependencies)
		// Second depends on first
		assert.Contains(t, milestones[1].Dependencies, milestones[0].ID)
		// Third depends on first and second
		assert.Contains(t, milestones[2].Dependencies, milestones[0].ID)
		assert.Contains(t, milestones[2].Dependencies, milestones[1].ID)
	})

	t.Run("OverallProgress", func(t *testing.T) {
		freshPillar := gsd.NewPillar(logger)

		progress, err := freshPillar.GetProgress("any-flow")
		require.NoError(t, err)
		assert.Equal(t, 0.0, progress)
	})

	t.Run("NilSpecFails", func(t *testing.T) {
		_, err := pillar.CreateMilestones(
			ctx, nil, []types.Task{{ID: "t-1"}},
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "specification required")
	})

	t.Run("EmptyTasksFails", func(t *testing.T) {
		_, err := pillar.CreateMilestones(
			ctx, sampleSpec("Empty"), []types.Task{},
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one task")
	})

	t.Run("UnknownMilestoneFails", func(t *testing.T) {
		_, err := pillar.TrackProgress(
			ctx, "nonexistent", []types.TaskResult{},
		)
		assert.Error(t, err)
	})
}

// --- 7. SpecKit Phase Chain ---

func TestSpecKitPhaseChainIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()
	logger := newTestLogger()
	pillar := speckit.NewPillar(cfg, logger)
	ctx := context.Background()

	t.Run("FullPhaseSequence", func(t *testing.T) {
		phases := types.AllPhases()
		previousOutput := ""

		for i, phase := range phases {
			input := &types.PhaseInput{
				UserRequest:    "build authentication module",
				PreviousOutput: previousOutput,
			}

			result, err := pillar.ExecutePhase(
				ctx, phase, input,
			)
			require.NoError(t, err,
				"phase %s should succeed", phase,
			)
			assert.True(t, result.Success)
			assert.Equal(t, phase, result.Phase)
			assert.NotEmpty(t, result.Output)
			assert.Greater(t, result.QualityScore, 0.0)
			assert.NotNil(t, result.Artifacts)

			// Each phase output should feed to next
			if i > 0 {
				assert.NotEqual(t, "", previousOutput)
			}
			previousOutput = result.Output
		}
	})

	t.Run("PhaseOutputFeedsToNext", func(t *testing.T) {
		specifyInput := &types.PhaseInput{
			UserRequest: "create user registration",
		}
		specifyResult, err := pillar.ExecutePhase(
			ctx, types.PhaseSpecify, specifyInput,
		)
		require.NoError(t, err)

		planInput := &types.PhaseInput{
			UserRequest:    "create user registration",
			PreviousOutput: specifyResult.Output,
		}
		planResult, err := pillar.ExecutePhase(
			ctx, types.PhasePlan, planInput,
		)
		require.NoError(t, err)

		assert.NotEqual(t, specifyResult.Output, planResult.Output)
		assert.Equal(t, types.PhaseSpecify, specifyResult.Phase)
		assert.Equal(t, types.PhasePlan, planResult.Phase)
	})

	t.Run("TransitionValidation", func(t *testing.T) {
		// Valid sequential transitions
		assert.True(t, pillar.ValidateTransition(
			types.PhaseConstitution, types.PhaseSpecify,
		))
		assert.True(t, pillar.ValidateTransition(
			types.PhaseSpecify, types.PhaseClarify,
		))
		assert.True(t, pillar.ValidateTransition(
			types.PhaseAnalyze, types.PhaseImplement,
		))

		// Invalid skipping transitions
		assert.False(t, pillar.ValidateTransition(
			types.PhaseConstitution, types.PhasePlan,
		))
		assert.False(t, pillar.ValidateTransition(
			types.PhaseSpecify, types.PhaseImplement,
		))

		// Backward transitions
		assert.False(t, pillar.ValidateTransition(
			types.PhaseImplement, types.PhaseConstitution,
		))
	})

	t.Run("PhaseOrderByCeremony", func(t *testing.T) {
		minPhases := pillar.GetPhaseOrder(types.CeremonyMinimal)
		assert.Equal(t, 1, len(minPhases))
		assert.Equal(t, types.PhaseImplement, minPhases[0])

		lightPhases := pillar.GetPhaseOrder(types.CeremonyLight)
		assert.Equal(t, 3, len(lightPhases))

		stdPhases := pillar.GetPhaseOrder(types.CeremonyStandard)
		assert.Equal(t, 7, len(stdPhases))

		heavyPhases := pillar.GetPhaseOrder(types.CeremonyHeavy)
		assert.Equal(t, 7, len(heavyPhases))
	})

	t.Run("PhaseStageMapping", func(t *testing.T) {
		constitutionResult, err := pillar.ExecutePhase(
			ctx, types.PhaseConstitution,
			&types.PhaseInput{UserRequest: "test"},
		)
		require.NoError(t, err)
		assert.Equal(t,
			types.StageConstitution,
			constitutionResult.Stage,
		)

		implResult, err := pillar.ExecutePhase(
			ctx, types.PhaseImplement,
			&types.PhaseInput{UserRequest: "test"},
		)
		require.NoError(t, err)
		assert.Equal(t, types.StageExecution, implResult.Stage)
	})
}

// --- 8. Superpowers Parallel Dispatch ---

func TestSuperpowersParallelDispatchIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()
	logger := newTestLogger()
	pillar := superpowers.NewPillar(cfg, logger)
	ctx := context.Background()

	t.Run("TDDExecution", func(t *testing.T) {
		tasks := sampleTasks(5)

		results, err := pillar.ExecuteWithTDD(ctx, tasks)
		require.NoError(t, err)
		assert.Equal(t, 5, len(results))

		for i, r := range results {
			assert.True(t, r.Success)
			assert.Equal(t, tasks[i].ID, r.TaskID)
			assert.NotEmpty(t, r.Output)
			assert.Equal(t, len(tasks[i].Acceptance), r.TestsPassed)
			assert.Equal(t, 0, r.TestsFailed)
		}
	})

	t.Run("ParallelSubagentDispatch", func(t *testing.T) {
		tasks := sampleTasks(8)

		results, err := pillar.DispatchSubagents(
			ctx, tasks, 4,
		)
		require.NoError(t, err)
		assert.Equal(t, 8, len(results))

		for _, r := range results {
			assert.True(t, r.Success)
			assert.NotEmpty(t, r.TaskID)
		}
	})

	t.Run("ParallelDispatchWithDefaultConcurrency", func(t *testing.T) {
		tasks := sampleTasks(3)

		results, err := pillar.DispatchSubagents(
			ctx, tasks, 0,
		)
		require.NoError(t, err)
		assert.Equal(t, 3, len(results))
	})

	t.Run("CodeReviewOnResults", func(t *testing.T) {
		tasks := sampleTasks(3)

		taskResults, err := pillar.ExecuteWithTDD(ctx, tasks)
		require.NoError(t, err)

		review, err := pillar.ReviewCode(ctx, taskResults)
		require.NoError(t, err)
		assert.True(t, review.Approved)
		assert.Equal(t, 1.0, review.Score)
		assert.Empty(t, review.Issues)
	})

	t.Run("ReviewDetectsFailures", func(t *testing.T) {
		results := []types.TaskResult{
			{TaskID: "t-1", Success: true},
			{TaskID: "t-2", Success: false, Error: "compilation failed"},
		}

		review, err := pillar.ReviewCode(ctx, results)
		require.NoError(t, err)
		assert.False(t, review.Approved)
		assert.Equal(t, 0.5, review.Score)
		assert.Equal(t, 1, len(review.Issues))
		assert.Equal(t, "critical", review.Issues[0].Severity)
	})

	t.Run("EmptyTasksFails", func(t *testing.T) {
		_, err := pillar.ExecuteWithTDD(ctx, []types.Task{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tasks")

		_, err = pillar.DispatchSubagents(
			ctx, []types.Task{}, 4,
		)
		assert.Error(t, err)

		_, err = pillar.ReviewCode(ctx, []types.TaskResult{})
		assert.Error(t, err)
	})

	t.Run("ContextCancellationStopsTDD", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		tasks := sampleTasks(10)
		results, err := pillar.ExecuteWithTDD(cancelCtx, tasks)

		// Should stop early due to cancellation
		assert.True(
			t,
			err != nil || len(results) < len(tasks),
			"expected early termination on cancelled context",
		)
	})
}

// --- 9. All Power Features Combined ---

func TestAllPowerFeaturesCombined(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	t.Run("AllTenFeaturesWorkTogether", func(t *testing.T) {
		// 1. Constitution as Code
		constitution := constitution_code.NewConstitution()
		err := constitution.AddRule(constitution_code.Rule{
			ID:        "rule-001",
			Category:  "quality",
			Priority:  1,
			Title:     "test coverage",
			Mandatory: true,
		})
		require.NoError(t, err)
		err = constitution.AddRule(constitution_code.Rule{
			ID:        "rule-002",
			Category:  "security",
			Priority:  1,
			Title:     "input validation",
			Mandatory: true,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, constitution.RuleCount())
		assert.Equal(t, 2, len(constitution.MandatoryRules()))

		// 2. Nyquist TDD Tracking
		tdd := nyquist_tdd.NewTracker()
		tdd.RecordImplementation(100)
		tdd.RecordTests(250)
		assert.True(t, tdd.IsCompliant())
		assert.GreaterOrEqual(t, tdd.Ratio(), 2.0)
		assert.NoError(t, tdd.Check())

		// 3. Parallel Execution
		executor := parallel_execution.NewExecutor(
			4, 5*time.Second,
		)
		tasks := map[string]parallel_execution.TaskFunc{
			"task-a": func(ctx context.Context) (string, error) {
				return "result-a", nil
			},
			"task-b": func(ctx context.Context) (string, error) {
				return "result-b", nil
			},
			"task-c": func(ctx context.Context) (string, error) {
				return "result-c", nil
			},
		}
		parallelResults := executor.Execute(ctx, tasks)
		assert.Equal(t, 3, len(parallelResults))
		completed, failed := executor.Stats()
		assert.Equal(t, 3, completed)
		assert.Equal(t, 0, failed)

		// 4. Debate Architecture
		debate := debate_architecture.NewDebate(
			"debate-001", "API Design", 3,
		)
		err = debate.AddRound(debate_architecture.Round{
			Positions: []debate_architecture.Position{
				{
					AgentID:  "agent-1",
					Argument: "Use REST",
					Score:    0.8,
				},
				{
					AgentID:  "agent-2",
					Argument: "Use gRPC",
					Score:    0.9,
				},
			},
			Winner: "agent-2",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, debate.RoundCount())
		best := debate.BestPosition()
		require.NotNil(t, best)
		assert.Equal(t, "agent-2", best.AgentID)

		// 5. Skill Learning
		learner := skill_learning.NewLearner()
		err = learner.RegisterSkill(skill_learning.Skill{
			ID:     "go-api",
			Name:   "Go API Development",
			Domain: "backend",
		})
		require.NoError(t, err)
		err = learner.RecordUsage("go-api", 0.9)
		require.NoError(t, err)
		err = learner.RecordUsage("go-api", 0.85)
		require.NoError(t, err)
		skill, err := learner.GetSkill("go-api")
		require.NoError(t, err)
		assert.Greater(t, skill.Proficiency, 0.0)
		assert.Equal(t, 2, skill.UseCount)

		// 6. Brownfield Analysis
		analyzer := brownfield.NewAnalyzer()
		err = analyzer.RecordPattern("singleton", 0.9)
		require.NoError(t, err)
		err = analyzer.RecordPattern("factory", 0.85)
		require.NoError(t, err)
		analyzer.AddDependency("github.com/gin-gonic/gin")
		result := analyzer.Analyze()
		assert.Equal(t, 2, len(result.Patterns))
		assert.Equal(t, 1, len(result.Dependencies))
		assert.True(t, analyzer.HasPattern("singleton"))

		// 7. Predictive Specification
		predictor := predictive_spec.NewPredictor()
		predictor.AddHistory("added auth module")
		predictor.AddHistory("added rate limiting")
		predictions := predictor.Predict()
		assert.Equal(t, 2, len(predictions))
		assert.Greater(t, predictions[0].Confidence, 0.0)

		// 8. Cross-Project Transfer
		transfer := cross_project.NewTransferEngine()
		err = transfer.Add(cross_project.Knowledge{
			ID:       "k-001",
			Project:  "ProjectAlpha",
			Category: "architecture",
			Content:  "Use hexagonal architecture",
			Tags:     []string{"architecture", "patterns"},
		})
		require.NoError(t, err)
		results := transfer.Search("architecture", 5)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, "ProjectAlpha", results[0].Project)

		// 9. Adaptive Ceremony
		adjuster := adaptive_ceremony.NewAdjuster(0.9, 0.4)
		highPhases := []types.PhaseResult{
			{QualityScore: 0.95, Success: true},
			{QualityScore: 0.92, Success: true},
		}
		adjusted := adjuster.Adjust(
			types.CeremonyHeavy, highPhases,
		)
		assert.Equal(t, types.CeremonyStandard, adjusted)

		// 10. Spec Memory Index
		index := spec_memory.NewIndex()
		err = index.Add(spec_memory.Entry{
			ID:      "entry-001",
			Title:   "Auth Spec",
			Content: "Authentication specification document",
			Tags:    []string{"auth", "security"},
			Score:   0.9,
		})
		require.NoError(t, err)
		entries := index.Search("auth", 5)
		assert.Equal(t, 1, len(entries))
		assert.Equal(t, 1, index.Count())
	})

	t.Run("FeaturesDoNotInterfere", func(t *testing.T) {
		// Run features in parallel to confirm no interference
		var wg sync.WaitGroup

		constitution := constitution_code.NewConstitution()
		tdd := nyquist_tdd.NewTracker()
		learner := skill_learning.NewLearner()
		analyzer := brownfield.NewAnalyzer()
		predictor := predictive_spec.NewPredictor()

		wg.Add(5)

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = constitution.AddRule(constitution_code.Rule{
					ID:        fmt.Sprintf("r-%d", i),
					Title:     fmt.Sprintf("Rule %d", i),
					Mandatory: i%2 == 0,
				})
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				tdd.RecordImplementation(10)
				tdd.RecordTests(25)
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = learner.RegisterSkill(skill_learning.Skill{
					ID:   fmt.Sprintf("skill-%d", i),
					Name: fmt.Sprintf("Skill %d", i),
				})
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = analyzer.RecordPattern(
					fmt.Sprintf("pattern-%d", i), 0.8,
				)
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				predictor.AddHistory(
					fmt.Sprintf("event-%d", i),
				)
			}
		}()

		wg.Wait()

		assert.Equal(t, 50, constitution.RuleCount())
		assert.Equal(t, 50, learner.SkillCount())
		assert.Equal(t, 50, analyzer.PatternCount())
		assert.Equal(t, 50, predictor.HistoryCount())
	})
}

// --- 10. Flow Resumption ---

func TestFlowResumptionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	eng := newFullEngine()
	ctx := context.Background()

	t.Run("ResumeCompletedFlow", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortMedium,
			Confidence:    0.7,
			CeremonyLevel: types.CeremonyLight,
		}

		result, err := eng.ExecuteFlow(
			ctx, "add validation logic", cl,
		)
		require.NoError(t, err)
		require.True(t, result.Success)

		flowID := result.FlowID

		// Resume the completed flow
		resumed, err := eng.ResumeFlow(
			ctx, flowID, "add validation logic",
		)
		require.NoError(t, err)
		assert.Equal(t, flowID, resumed.FlowID)
		assert.True(t, resumed.Success)
	})

	t.Run("GetFlowStatus", func(t *testing.T) {
		cl := &types.EffortClassification{
			Level:         types.EffortQuick,
			Confidence:    0.9,
			CeremonyLevel: types.CeremonyMinimal,
		}

		result, err := eng.ExecuteFlow(
			ctx, "bump version number", cl,
		)
		require.NoError(t, err)

		status, err := eng.GetFlowStatus(result.FlowID)
		require.NoError(t, err)
		assert.Equal(t, result.FlowID, status.FlowID)
		assert.True(t, status.Success)
	})

	t.Run("ResumeNonexistentFlowFails", func(t *testing.T) {
		_, err := eng.ResumeFlow(
			ctx, "nonexistent-id", "anything",
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetStatusNonexistentFlowFails", func(t *testing.T) {
		_, err := eng.GetFlowStatus("nonexistent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("MultipleFlowsTrackedIndependently", func(t *testing.T) {
		flowIDs := make([]string, 3)

		for i := 0; i < 3; i++ {
			cl := &types.EffortClassification{
				Level:         types.EffortQuick,
				Confidence:    0.9,
				CeremonyLevel: types.CeremonyMinimal,
			}
			result, err := eng.ExecuteFlow(
				ctx,
				fmt.Sprintf("fix issue #%d", i+1),
				cl,
			)
			require.NoError(t, err)
			flowIDs[i] = result.FlowID
		}

		// All flows should be independently retrievable
		for _, id := range flowIDs {
			status, err := eng.GetFlowStatus(id)
			require.NoError(t, err)
			assert.Equal(t, id, status.FlowID)
			assert.True(t, status.Success)
		}

		// All flow IDs should be unique
		assert.NotEqual(t, flowIDs[0], flowIDs[1])
		assert.NotEqual(t, flowIDs[1], flowIDs[2])
		assert.NotEqual(t, flowIDs[0], flowIDs[2])
	})
}

// --- 11. Metrics Collection ---

func TestMetricsCollectionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("FlowMetricsTracking", func(t *testing.T) {
		tracker := metrics.NewMetrics()

		// Record flow starts
		tracker.RecordFlowStart(types.EffortQuick)
		tracker.RecordFlowStart(types.EffortMedium)
		tracker.RecordFlowStart(types.EffortLarge)

		snap := tracker.Snapshot()
		assert.Equal(t, int64(3), snap.FlowsStarted)
		assert.Equal(t, int64(1), snap.EffortCounts[types.EffortQuick])
		assert.Equal(t, int64(1), snap.EffortCounts[types.EffortMedium])
		assert.Equal(t, int64(1), snap.EffortCounts[types.EffortLarge])
	})

	t.Run("FlowCompletionTracking", func(t *testing.T) {
		tracker := metrics.NewMetrics()

		tracker.RecordFlowStart(types.EffortMedium)
		tracker.RecordFlowStart(types.EffortMedium)

		successFlow := &types.FlowResult{
			Success:       true,
			CeremonyLevel: types.CeremonyLight,
			Duration:      500 * time.Millisecond,
		}
		failedFlow := &types.FlowResult{
			Success:       false,
			CeremonyLevel: types.CeremonyLight,
			Duration:      200 * time.Millisecond,
		}

		tracker.RecordFlowComplete(successFlow)
		tracker.RecordFlowComplete(failedFlow)

		snap := tracker.Snapshot()
		assert.Equal(t, int64(1), snap.FlowsCompleted)
		assert.Equal(t, int64(1), snap.FlowsFailed)
		assert.Equal(t, int64(2), snap.CeremonyCounts[types.CeremonyLight])
		assert.Equal(t, 700*time.Millisecond, snap.TotalDuration)
	})

	t.Run("PhaseMetricsTracking", func(t *testing.T) {
		tracker := metrics.NewMetrics()

		phases := []types.PhaseResult{
			{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.9,
				Duration:     100 * time.Millisecond,
			},
			{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.8,
				Duration:     150 * time.Millisecond,
			},
			{
				Phase:        types.PhasePlan,
				Success:      false,
				QualityScore: 0.3,
				Duration:     50 * time.Millisecond,
			},
		}

		for i := range phases {
			tracker.RecordPhase(&phases[i])
		}

		snap := tracker.Snapshot()
		assert.Equal(t, int64(3), snap.PhasesExecuted)
		assert.Equal(t, int64(1), snap.PhasesFailed)

		specifyMetric := tracker.GetPhaseMetric(types.PhaseSpecify)
		assert.Equal(t, int64(2), specifyMetric.Executions)
		assert.Equal(t, int64(2), specifyMetric.Successes)
		assert.Equal(t, int64(0), specifyMetric.Failures)
		assert.InDelta(t, 0.85, specifyMetric.AvgScore, 0.01)

		planMetric := tracker.GetPhaseMetric(types.PhasePlan)
		assert.Equal(t, int64(1), planMetric.Executions)
		assert.Equal(t, int64(1), planMetric.Failures)
	})

	t.Run("SuccessRateCalculation", func(t *testing.T) {
		tracker := metrics.NewMetrics()

		assert.Equal(t, 0.0, tracker.GetFlowSuccessRate())

		tracker.RecordFlowStart(types.EffortMedium)
		tracker.RecordFlowStart(types.EffortMedium)
		tracker.RecordFlowStart(types.EffortMedium)
		tracker.RecordFlowStart(types.EffortMedium)

		for i := 0; i < 3; i++ {
			tracker.RecordFlowComplete(&types.FlowResult{
				Success:       true,
				CeremonyLevel: types.CeremonyLight,
			})
		}
		tracker.RecordFlowComplete(&types.FlowResult{
			Success:       false,
			CeremonyLevel: types.CeremonyLight,
		})

		assert.InDelta(t, 0.75, tracker.GetFlowSuccessRate(), 0.01)
	})

	t.Run("EffortCountTracking", func(t *testing.T) {
		tracker := metrics.NewMetrics()

		for i := 0; i < 5; i++ {
			tracker.RecordFlowStart(types.EffortQuick)
		}
		for i := 0; i < 3; i++ {
			tracker.RecordFlowStart(types.EffortEpic)
		}

		assert.Equal(t, int64(5), tracker.GetEffortCount(types.EffortQuick))
		assert.Equal(t, int64(3), tracker.GetEffortCount(types.EffortEpic))
		assert.Equal(t, int64(0), tracker.GetEffortCount(types.EffortLarge))
	})

	t.Run("MetricsIntegrationWithEngine", func(t *testing.T) {
		eng := newFullEngine()
		ctx := context.Background()
		tracker := metrics.NewMetrics()

		efforts := []types.EffortLevel{
			types.EffortQuick,
			types.EffortMedium,
			types.EffortLarge,
		}

		for _, effort := range efforts {
			tracker.RecordFlowStart(effort)

			cl := &types.EffortClassification{
				Level:         effort,
				Confidence:    0.8,
				CeremonyLevel: types.CeremonyForEffort(effort),
			}

			result, err := eng.ExecuteFlow(
				ctx,
				fmt.Sprintf("task with %s effort", effort),
				cl,
			)
			require.NoError(t, err)

			tracker.RecordFlowComplete(result)
			for i := range result.PhaseResults {
				tracker.RecordPhase(&result.PhaseResults[i])
			}
		}

		snap := tracker.Snapshot()
		assert.Equal(t, int64(3), snap.FlowsStarted)
		assert.Equal(t, int64(3), snap.FlowsCompleted)
		assert.Greater(t, snap.PhasesExecuted, int64(0))
	})

	t.Run("UnrecordedPhaseReturnsEmpty", func(t *testing.T) {
		tracker := metrics.NewMetrics()
		pm := tracker.GetPhaseMetric(types.PhaseTasks)
		assert.Equal(t, int64(0), pm.Executions)
	})
}

// --- 12. Cross-Component Thread Safety ---

func TestCrossComponentThreadSafety(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	eng := newFullEngine()
	ctx := context.Background()
	tracker := metrics.NewMetrics()

	const goroutines = 20
	var wg sync.WaitGroup
	errors := make(chan error, goroutines*3)

	// Launch goroutines doing different operations
	for i := 0; i < goroutines; i++ {
		idx := i

		// Classify effort concurrently
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := fmt.Sprintf(
				"concurrent task %d for classification", idx,
			)
			cl, err := eng.ClassifyEffort(ctx, req)
			if err != nil {
				errors <- fmt.Errorf(
					"classify %d: %w", idx, err,
				)
				return
			}
			if cl == nil {
				errors <- fmt.Errorf(
					"classify %d: nil classification", idx,
				)
				return
			}
			tracker.RecordFlowStart(cl.Level)
		}()

		// Execute flows concurrently
		wg.Add(1)
		go func() {
			defer wg.Done()
			cl := &types.EffortClassification{
				Level:         types.EffortQuick,
				Confidence:    0.9,
				CeremonyLevel: types.CeremonyMinimal,
			}
			result, err := eng.ExecuteFlow(
				ctx,
				fmt.Sprintf("concurrent flow %d", idx),
				cl,
			)
			if err != nil {
				errors <- fmt.Errorf(
					"flow %d: %w", idx, err,
				)
				return
			}
			if result == nil || !result.Success {
				errors <- fmt.Errorf(
					"flow %d: unsuccessful", idx,
				)
				return
			}
			tracker.RecordFlowComplete(result)
			for j := range result.PhaseResults {
				tracker.RecordPhase(&result.PhaseResults[j])
			}
		}()

		// Query metrics concurrently
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tracker.GetFlowSuccessRate()
			_ = tracker.GetEffortCount(types.EffortQuick)
			_ = tracker.GetPhaseMetric(types.PhaseImplement)
			_ = tracker.Snapshot()
		}()
	}

	wg.Wait()
	close(errors)

	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	assert.Empty(t, errs,
		"expected no errors from concurrent operations, got: %v",
		errs,
	)

	t.Run("MetricsConsistentAfterConcurrency", func(t *testing.T) {
		snap := tracker.Snapshot()
		// 20 classify goroutines each called RecordFlowStart
		assert.Equal(t, int64(goroutines), snap.FlowsStarted)
		// 20 flow goroutines each called RecordFlowComplete
		assert.Equal(t, int64(goroutines), snap.FlowsCompleted)
	})

	t.Run("EngineHealthUnderConcurrency", func(t *testing.T) {
		err := eng.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("EngineNameAndVersion", func(t *testing.T) {
		assert.Equal(t, "HelixSpecifier", eng.Name())
		assert.NotEmpty(t, eng.Version())
	})
}

// --- Additional Integration Tests ---

func TestEngineHealthChecks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("HealthyWithSpecKit", func(t *testing.T) {
		eng := newFullEngine()
		err := eng.Health(context.Background())
		assert.NoError(t, err)
	})

	t.Run("UnhealthyWithoutSpecKit", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger := newTestLogger()
		eng := engine.New(cfg, logger)

		err := eng.Health(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "speckit")
	})
}

func TestConfigurationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("DefaultConfigValues", func(t *testing.T) {
		cfg := config.DefaultConfig()
		assert.Equal(t, 5, cfg.QuickMaxMinutes)
		assert.Equal(t, 90, cfg.MediumMaxMinutes)
		assert.Equal(t, 16, cfg.LargeMaxHours)
		assert.True(t, cfg.AdaptiveCeremony)
		assert.True(t, cfg.SpecMemoryEnabled)
		assert.True(t, cfg.CrossProjectLearn)
		assert.True(t, cfg.CacheEnabled)
		assert.Equal(t, 4, cfg.MaxParallelAgents)
		assert.Equal(t, 8, cfg.MaxParallelTasks)
		assert.Equal(t, 0.6, cfg.MinQualityScore)
		assert.Equal(t, 0.5, cfg.MinPhaseScore)
	})

	t.Run("PhaseRoundsConfig", func(t *testing.T) {
		cfg := config.DefaultConfig()
		assert.Equal(t, 5, cfg.GetPhaseRounds("constitution"))
		assert.Equal(t, 3, cfg.GetPhaseRounds("specify"))
		assert.Equal(t, 3, cfg.GetPhaseRounds("clarify"))
		assert.Equal(t, 4, cfg.GetPhaseRounds("plan"))
		assert.Equal(t, 2, cfg.GetPhaseRounds("tasks"))
		assert.Equal(t, 4, cfg.GetPhaseRounds("analyze"))
		assert.Equal(t, 6, cfg.GetPhaseRounds("implement"))
		assert.Equal(t, 3, cfg.GetPhaseRounds("unknown"))
	})

	t.Run("PhaseTimeoutsConfig", func(t *testing.T) {
		cfg := config.DefaultConfig()
		assert.Equal(t, 15*time.Minute, cfg.GetPhaseTimeout("constitution"))
		assert.Equal(t, 20*time.Minute, cfg.GetPhaseTimeout("implement"))
		assert.Equal(t, 10*time.Minute, cfg.GetPhaseTimeout("unknown"))
	})

	t.Run("ConfigDrivesEngineExecution", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.AdaptiveCeremony = false
		logger := newTestLogger()

		eng := engine.New(cfg, logger)
		eng.RegisterSpecKit(speckit.NewPillar(cfg, logger))
		eng.RegisterCeremonyScaler(
			ceremony.NewScaler(cfg, logger),
		)

		ctx := context.Background()
		cl := &types.EffortClassification{
			Level:         types.EffortLarge,
			Confidence:    0.8,
			CeremonyLevel: types.CeremonyStandard,
		}

		result, err := eng.ExecuteFlow(
			ctx, "build feature", cl,
		)
		require.NoError(t, err)
		assert.True(t, result.Success)
		// Ceremony should not change when adaptive is disabled
		assert.Equal(t, types.CeremonyStandard, result.CeremonyLevel)
	})
}

func TestTypeHelperFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("AllPhases", func(t *testing.T) {
		phases := types.AllPhases()
		assert.Equal(t, 7, len(phases))
		assert.Equal(t, types.PhaseConstitution, phases[0])
		assert.Equal(t, types.PhaseImplement, phases[6])
	})

	t.Run("AllStages", func(t *testing.T) {
		stages := types.AllStages()
		assert.Equal(t, 5, len(stages))
		assert.Equal(t, types.StageIdeation, stages[0])
		assert.Equal(t, types.StageVerification, stages[4])
	})

	t.Run("EffortLevels", func(t *testing.T) {
		levels := types.EffortLevels()
		assert.Equal(t, 4, len(levels))
	})

	t.Run("CeremonyForEffort", func(t *testing.T) {
		assert.Equal(t,
			types.CeremonyMinimal,
			types.CeremonyForEffort(types.EffortQuick),
		)
		assert.Equal(t,
			types.CeremonyLight,
			types.CeremonyForEffort(types.EffortMedium),
		)
		assert.Equal(t,
			types.CeremonyStandard,
			types.CeremonyForEffort(types.EffortLarge),
		)
		assert.Equal(t,
			types.CeremonyHeavy,
			types.CeremonyForEffort(types.EffortEpic),
		)
		assert.Equal(t,
			types.CeremonyStandard,
			types.CeremonyForEffort("unknown"),
		)
	})

	t.Run("StagesForEffort", func(t *testing.T) {
		quickStages := types.StagesForEffort(types.EffortQuick)
		assert.Equal(t, 1, len(quickStages))
		assert.Equal(t, types.StageExecution, quickStages[0])

		mediumStages := types.StagesForEffort(types.EffortMedium)
		assert.Equal(t, 3, len(mediumStages))

		largeStages := types.StagesForEffort(types.EffortLarge)
		assert.Equal(t, 5, len(largeStages))
	})

	t.Run("PhasesForCeremony", func(t *testing.T) {
		minPhases := types.PhasesForCeremony(types.CeremonyMinimal)
		assert.Equal(t, 1, len(minPhases))
		assert.Equal(t, types.PhaseImplement, minPhases[0])

		lightPhases := types.PhasesForCeremony(types.CeremonyLight)
		assert.Equal(t, 3, len(lightPhases))
		assert.Equal(t, types.PhaseSpecify, lightPhases[0])
		assert.Equal(t, types.PhasePlan, lightPhases[1])
		assert.Equal(t, types.PhaseImplement, lightPhases[2])

		stdPhases := types.PhasesForCeremony(types.CeremonyStandard)
		assert.Equal(t, 7, len(stdPhases))
	})
}

func TestParallelExecutorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ConcurrentTaskExecution", func(t *testing.T) {
		executor := parallel_execution.NewExecutor(
			2, 5*time.Second,
		)

		var mu sync.Mutex
		var order []string

		tasks := map[string]parallel_execution.TaskFunc{
			"alpha": func(ctx context.Context) (string, error) {
				time.Sleep(10 * time.Millisecond)
				mu.Lock()
				order = append(order, "alpha")
				mu.Unlock()
				return "alpha-done", nil
			},
			"beta": func(ctx context.Context) (string, error) {
				time.Sleep(10 * time.Millisecond)
				mu.Lock()
				order = append(order, "beta")
				mu.Unlock()
				return "beta-done", nil
			},
		}

		results := executor.Execute(context.Background(), tasks)
		assert.Equal(t, 2, len(results))
		assert.Equal(t, 2, len(order))
	})

	t.Run("TaskFailureTracked", func(t *testing.T) {
		executor := parallel_execution.NewExecutor(
			4, 5*time.Second,
		)

		tasks := map[string]parallel_execution.TaskFunc{
			"good": func(ctx context.Context) (string, error) {
				return "ok", nil
			},
			"bad": func(ctx context.Context) (string, error) {
				return "", fmt.Errorf("task failed")
			},
		}

		results := executor.Execute(context.Background(), tasks)
		assert.Equal(t, 2, len(results))

		completed, failed := executor.Stats()
		assert.Equal(t, 1, completed)
		assert.Equal(t, 1, failed)
	})

	t.Run("ExecutorProperties", func(t *testing.T) {
		executor := parallel_execution.NewExecutor(
			8, 30*time.Second,
		)
		assert.Equal(t, 8, executor.MaxWorkers())
		assert.Equal(t, 30*time.Second, executor.Timeout())
		assert.Contains(t, executor.String(), "workers=8")
	})

	t.Run("DefaultExecutorValues", func(t *testing.T) {
		executor := parallel_execution.NewExecutor(0, 0)
		assert.Equal(t, 4, executor.MaxWorkers())
		assert.Equal(t, 30*time.Second, executor.Timeout())
	})
}

func TestDebateArchitectureIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("MultiRoundDebate", func(t *testing.T) {
		debate := debate_architecture.NewDebate(
			"d-001", "Service Design", 3,
		)
		assert.Equal(t, "d-001", debate.ID())
		assert.Equal(t, "Service Design", debate.Topic())
		assert.False(t, debate.IsComplete())

		for i := 0; i < 3; i++ {
			round := debate_architecture.Round{
				Positions: []debate_architecture.Position{
					{
						AgentID:  fmt.Sprintf("agent-%d-a", i),
						Argument: fmt.Sprintf("Position A round %d", i),
						Score:    0.7 + float64(i)*0.05,
					},
					{
						AgentID:  fmt.Sprintf("agent-%d-b", i),
						Argument: fmt.Sprintf("Position B round %d", i),
						Score:    0.6 + float64(i)*0.1,
					},
				},
			}
			err := debate.AddRound(round)
			require.NoError(t, err)
		}

		assert.True(t, debate.IsComplete())
		assert.Equal(t, 3, debate.RoundCount())

		// Cannot add more rounds
		err := debate.AddRound(debate_architecture.Round{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max rounds")
	})

	t.Run("BestPositionTracking", func(t *testing.T) {
		debate := debate_architecture.NewDebate(
			"d-002", "Architecture", 5,
		)

		assert.Nil(t, debate.BestPosition())

		err := debate.AddRound(debate_architecture.Round{
			Positions: []debate_architecture.Position{
				{AgentID: "a1", Score: 0.7},
				{AgentID: "a2", Score: 0.95},
				{AgentID: "a3", Score: 0.8},
			},
		})
		require.NoError(t, err)

		best := debate.BestPosition()
		require.NotNil(t, best)
		assert.Equal(t, "a2", best.AgentID)
		assert.Equal(t, 0.95, best.Score)
	})
}

func TestSpecMemoryIndexIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("IndexLifecycle", func(t *testing.T) {
		index := spec_memory.NewIndex()

		entries := []spec_memory.Entry{
			{
				ID:      "e-1",
				Title:   "Auth Module Spec",
				Content: "Specification for authentication",
				Tags:    []string{"auth", "security"},
				Score:   0.9,
			},
			{
				ID:      "e-2",
				Title:   "Cache Layer Spec",
				Content: "Redis caching specification",
				Tags:    []string{"cache", "performance"},
				Score:   0.85,
			},
			{
				ID:      "e-3",
				Title:   "API Gateway Spec",
				Content: "Gateway routing specification",
				Tags:    []string{"api", "routing"},
				Score:   0.88,
			},
		}

		for _, e := range entries {
			err := index.Add(e)
			require.NoError(t, err)
		}
		assert.Equal(t, 3, index.Count())

		// Retrieve and access tracking
		e, err := index.Get("e-1")
		require.NoError(t, err)
		assert.Equal(t, "Auth Module Spec", e.Title)
		assert.Equal(t, 1, e.AccessCount)

		e, err = index.Get("e-1")
		require.NoError(t, err)
		assert.Equal(t, 2, e.AccessCount)

		// Search
		results := index.Search("auth", 10)
		assert.GreaterOrEqual(t, len(results), 1)

		// Remove
		err = index.Remove("e-2")
		require.NoError(t, err)
		assert.Equal(t, 2, index.Count())

		err = index.Remove("nonexistent")
		assert.Error(t, err)
	})

	t.Run("MostAccessedTracking", func(t *testing.T) {
		index := spec_memory.NewIndex()

		assert.Nil(t, index.MostAccessed())

		err := index.Add(spec_memory.Entry{
			ID:    "a",
			Title: "Entry A",
		})
		require.NoError(t, err)
		err = index.Add(spec_memory.Entry{
			ID:    "b",
			Title: "Entry B",
		})
		require.NoError(t, err)

		// Access "b" more times
		for i := 0; i < 5; i++ {
			_, _ = index.Get("b")
		}
		_, _ = index.Get("a")

		most := index.MostAccessed()
		require.NotNil(t, most)
		assert.Equal(t, "b", most.ID)
		assert.Equal(t, 5, most.AccessCount)
	})
}

func TestCrossProjectTransferIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("KnowledgeTransferWorkflow", func(t *testing.T) {
		transfer := cross_project.NewTransferEngine()

		// Add knowledge from multiple projects
		projects := []struct {
			project  string
			category string
			content  string
			tags     []string
		}{
			{
				"Alpha", "architecture",
				"Use hexagonal architecture for clean separation",
				[]string{"architecture", "clean-code"},
			},
			{
				"Alpha", "testing",
				"Table-driven tests for all handlers",
				[]string{"testing", "go"},
			},
			{
				"Beta", "deployment",
				"Blue-green deployment strategy",
				[]string{"deployment", "ci-cd"},
			},
			{
				"Gamma", "architecture",
				"Event-driven microservices pattern",
				[]string{"architecture", "events"},
			},
		}

		for i, p := range projects {
			err := transfer.Add(cross_project.Knowledge{
				ID:       fmt.Sprintf("k-%d", i+1),
				Project:  p.project,
				Category: p.category,
				Content:  p.content,
				Tags:     p.tags,
			})
			require.NoError(t, err)
		}

		assert.Equal(t, 4, transfer.Count())

		// Search by content
		archResults := transfer.Search("architecture", 10)
		assert.GreaterOrEqual(t, len(archResults), 2)

		// Filter by project
		alphaKnowledge := transfer.ForProject("Alpha")
		assert.Equal(t, 2, len(alphaKnowledge))

		betaKnowledge := transfer.ForProject("Beta")
		assert.Equal(t, 1, len(betaKnowledge))

		emptyKnowledge := transfer.ForProject("NonExistent")
		assert.Equal(t, 0, len(emptyKnowledge))
	})

	t.Run("ValidationRules", func(t *testing.T) {
		transfer := cross_project.NewTransferEngine()

		err := transfer.Add(cross_project.Knowledge{
			ID:      "",
			Project: "test",
		})
		assert.Error(t, err)

		err = transfer.Add(cross_project.Knowledge{
			ID:      "k-1",
			Project: "",
		})
		assert.Error(t, err)
	})
}

func TestSkillLearningIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ProficiencyImprovement", func(t *testing.T) {
		learner := skill_learning.NewLearner()

		err := learner.RegisterSkill(skill_learning.Skill{
			ID:     "spec-writing",
			Name:   "Specification Writing",
			Domain: "sdd",
		})
		require.NoError(t, err)

		scores := []float64{0.6, 0.7, 0.75, 0.82, 0.9}
		for _, score := range scores {
			err = learner.RecordUsage("spec-writing", score)
			require.NoError(t, err)
		}

		skill, err := learner.GetSkill("spec-writing")
		require.NoError(t, err)
		assert.Equal(t, 5, skill.UseCount)
		assert.Greater(t, skill.Proficiency, 0.0)
	})

	t.Run("TopSkillsRanking", func(t *testing.T) {
		learner := skill_learning.NewLearner()

		skills := []struct {
			id    string
			score float64
		}{
			{"api-design", 0.9},
			{"testing", 0.95},
			{"docs", 0.7},
		}

		for _, s := range skills {
			err := learner.RegisterSkill(skill_learning.Skill{
				ID:   s.id,
				Name: s.id,
			})
			require.NoError(t, err)
			err = learner.RecordUsage(s.id, s.score)
			require.NoError(t, err)
		}

		top := learner.TopSkills(2)
		assert.Equal(t, 2, len(top))
		// Highest proficiency should be first
		assert.GreaterOrEqual(t,
			top[0].Proficiency, top[1].Proficiency,
		)
	})

	t.Run("DuplicateRegistrationFails", func(t *testing.T) {
		learner := skill_learning.NewLearner()

		err := learner.RegisterSkill(skill_learning.Skill{
			ID: "skill-1", Name: "Skill 1",
		})
		require.NoError(t, err)

		err = learner.RegisterSkill(skill_learning.Skill{
			ID: "skill-1", Name: "Duplicate",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestConstitutionCodeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ConstitutionValidation", func(t *testing.T) {
		constitution := constitution_code.NewConstitution()

		rules := []constitution_code.Rule{
			{
				ID:        "r-tests",
				Category:  "quality",
				Title:     "test coverage",
				Mandatory: true,
				Priority:  1,
			},
			{
				ID:        "r-security",
				Category:  "security",
				Title:     "security scanning",
				Mandatory: true,
				Priority:  1,
			},
			{
				ID:        "r-docs",
				Category:  "documentation",
				Title:     "complete documentation",
				Mandatory: false,
				Priority:  2,
			},
		}

		for _, r := range rules {
			err := constitution.AddRule(r)
			require.NoError(t, err)
		}

		assert.Equal(t, 3, constitution.RuleCount())
		assert.Equal(t, 2, len(constitution.MandatoryRules()))

		// Text that mentions both mandatory rules
		compliant := "We ensured test coverage and security scanning throughout"
		violations := constitution.Validate(compliant)
		assert.Empty(t, violations)

		// Text missing security scanning
		partial := "We added test coverage for all modules"
		violations = constitution.Validate(partial)
		assert.Contains(t, violations, "r-security")
		assert.NotContains(t, violations, "r-tests")
	})

	t.Run("DuplicateRuleFails", func(t *testing.T) {
		constitution := constitution_code.NewConstitution()

		err := constitution.AddRule(constitution_code.Rule{
			ID: "r-1", Title: "Rule 1",
		})
		require.NoError(t, err)

		err = constitution.AddRule(constitution_code.Rule{
			ID: "r-1", Title: "Duplicate",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("EmptyIDFails", func(t *testing.T) {
		constitution := constitution_code.NewConstitution()
		err := constitution.AddRule(constitution_code.Rule{
			ID: "", Title: "No ID",
		})
		assert.Error(t, err)
	})

	t.Run("GetRuleByID", func(t *testing.T) {
		constitution := constitution_code.NewConstitution()
		err := constitution.AddRule(constitution_code.Rule{
			ID:       "r-1",
			Title:    "Test Rule",
			Category: "testing",
		})
		require.NoError(t, err)

		rule, err := constitution.GetRule("r-1")
		require.NoError(t, err)
		assert.Equal(t, "Test Rule", rule.Title)

		_, err = constitution.GetRule("nonexistent")
		assert.Error(t, err)
	})
}

func TestNyquistTDDIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ComplianceLifecycle", func(t *testing.T) {
		tracker := nyquist_tdd.NewTracker()

		// Initially at zero
		impl, tests := tracker.Stats()
		assert.Equal(t, 0, impl)
		assert.Equal(t, 0, tests)
		assert.Equal(t, 0.0, tracker.Ratio())

		// Add implementation without enough tests
		tracker.RecordImplementation(100)
		tracker.RecordTests(100)
		assert.False(t, tracker.IsCompliant())
		assert.InDelta(t, 1.0, tracker.Ratio(), 0.01)

		err := tracker.Check()
		assert.Error(t, err)
		assert.Equal(t, 1, len(tracker.Violations()))

		// Add more tests to become compliant
		tracker.RecordTests(150)
		assert.True(t, tracker.IsCompliant())
		assert.InDelta(t, 2.5, tracker.Ratio(), 0.01)

		err = tracker.Check()
		assert.NoError(t, err)
		// Violations should still contain the previous one
		assert.Equal(t, 1, len(tracker.Violations()))
	})

	t.Run("ZeroImplRatio", func(t *testing.T) {
		tracker := nyquist_tdd.NewTracker()
		tracker.RecordTests(50)
		assert.Equal(t, 0.0, tracker.Ratio())
	})
}

func TestBrownfieldAnalysisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("PatternAnalysis", func(t *testing.T) {
		analyzer := brownfield.NewAnalyzer()

		patterns := []struct {
			name       string
			confidence float64
			count      int
		}{
			{"singleton", 0.9, 3},
			{"factory", 0.85, 2},
			{"observer", 0.7, 1},
		}

		for _, p := range patterns {
			for i := 0; i < p.count; i++ {
				err := analyzer.RecordPattern(p.name, p.confidence)
				require.NoError(t, err)
			}
		}

		analyzer.AddDependency("gin")
		analyzer.AddDependency("pgx")

		result := analyzer.Analyze()
		assert.Equal(t, 3, len(result.Patterns))
		assert.Equal(t, 2, len(result.Dependencies))
		assert.Greater(t, result.Complexity, 0.0)
		assert.Greater(t, result.LinesOfCode, 0)
	})

	t.Run("CaseInsensitivePatternLookup", func(t *testing.T) {
		analyzer := brownfield.NewAnalyzer()
		err := analyzer.RecordPattern("Singleton", 0.9)
		require.NoError(t, err)

		assert.True(t, analyzer.HasPattern("singleton"))
		assert.True(t, analyzer.HasPattern("Singleton"))
		assert.True(t, analyzer.HasPattern("SINGLETON"))
	})

	t.Run("EmptyPatternNameFails", func(t *testing.T) {
		analyzer := brownfield.NewAnalyzer()
		err := analyzer.RecordPattern("", 0.9)
		assert.Error(t, err)
	})
}

func TestPredictiveSpecIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("PredictionFromHistory", func(t *testing.T) {
		predictor := predictive_spec.NewPredictor()

		events := []string{
			"added user authentication",
			"added role-based access control",
			"added audit logging",
			"added session management",
		}

		for _, e := range events {
			predictor.AddHistory(e)
		}
		assert.Equal(t, 4, predictor.HistoryCount())

		predictions := predictor.Predict()
		assert.Equal(t, 4, len(predictions))
		assert.Equal(t, 4, predictor.PredictionCount())

		// First prediction should have highest confidence
		assert.Greater(t,
			predictions[0].Confidence,
			predictions[len(predictions)-1].Confidence,
		)

		// High confidence filtering
		highConf := predictor.HighConfidence(0.8)
		assert.Greater(t, len(highConf), 0)
		for _, p := range highConf {
			assert.GreaterOrEqual(t, p.Confidence, 0.8)
		}
	})

	t.Run("EmptyHistory", func(t *testing.T) {
		predictor := predictive_spec.NewPredictor()
		predictions := predictor.Predict()
		assert.Equal(t, 0, len(predictions))
	})
}

func TestCeremonyBoostIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("BoostIncreasesLevel", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.CeremonyBoost = 1.0
		logger := newTestLogger()

		scaler := ceremony.NewScaler(cfg, logger)

		cl := &types.EffortClassification{
			Level: types.EffortQuick,
		}
		result := scaler.Scale(cl)
		// Quick + boost = Light
		assert.Equal(t, types.CeremonyLight, result)

		cl = &types.EffortClassification{
			Level: types.EffortMedium,
		}
		result = scaler.Scale(cl)
		// Light + boost = Standard
		assert.Equal(t, types.CeremonyStandard, result)
	})

	t.Run("NoBoostPreservesLevel", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.CeremonyBoost = 0.0
		logger := newTestLogger()

		scaler := ceremony.NewScaler(cfg, logger)

		cl := &types.EffortClassification{
			Level: types.EffortQuick,
		}
		result := scaler.Scale(cl)
		assert.Equal(t, types.CeremonyMinimal, result)
	})
}

func TestEndToEndWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("ClassifyThenExecuteWithMetrics", func(t *testing.T) {
		eng := newFullEngine()
		ctx := context.Background()
		tracker := metrics.NewMetrics()
		store := memory.NewStore()

		// Step 1: Classify the request
		cl, err := eng.ClassifyEffort(
			ctx,
			"implement new database migration service with API",
		)
		require.NoError(t, err)
		require.NotNil(t, cl)

		tracker.RecordFlowStart(cl.Level)

		// Step 2: Execute the flow
		result, err := eng.ExecuteFlow(
			ctx,
			"implement new database migration service with API",
			cl,
		)
		require.NoError(t, err)
		require.True(t, result.Success)

		tracker.RecordFlowComplete(result)
		for i := range result.PhaseResults {
			tracker.RecordPhase(&result.PhaseResults[i])
		}

		// Step 3: Store the generated spec
		spec := sampleSpec("Database Migration Service")
		err = store.Store(ctx, spec)
		require.NoError(t, err)

		// Step 4: Learn from the flow
		err = store.LearnFromFlow(ctx, result)
		require.NoError(t, err)

		// Step 5: Format output through adapter
		adapter := adapters.NewGenericAdapter(
			"opencode", adapters.CategoryMarkdown,
		)
		output, err := adapter.FormatOutput(result)
		require.NoError(t, err)
		assert.NotEmpty(t, output)

		// Step 6: Verify flow is retrievable
		status, err := eng.GetFlowStatus(result.FlowID)
		require.NoError(t, err)
		assert.Equal(t, result.FlowID, status.FlowID)

		// Verify metrics
		snap := tracker.Snapshot()
		assert.Equal(t, int64(1), snap.FlowsStarted)
		assert.Equal(t, int64(1), snap.FlowsCompleted)
		assert.Greater(t, snap.PhasesExecuted, int64(0))

		// Verify memory learned
		assert.Equal(t, 1, store.PatternCount())
		assert.Equal(t, 1, store.SpecCount())
	})
}
