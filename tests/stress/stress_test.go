package stress

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/ceremony"
	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/engine"
	"digital.vasic.helixspecifier/pkg/features/cross_project"
	"digital.vasic.helixspecifier/pkg/features/parallel_execution"
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

const deadlockTimeout = 120 * time.Second

// newTestLogger returns a logrus logger with output suppressed
// for stress tests to avoid excessive log noise.
func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	return logger
}

// newTestEngine creates a fully wired FusionEngine with all
// three pillars, ceremony scaler, and spec memory registered.
func newTestEngine() *engine.FusionEngine {
	cfg := config.DefaultConfig()
	logger := newTestLogger()

	e := engine.New(cfg, logger)
	e.RegisterSpecKit(speckit.NewPillar(cfg, logger))
	e.RegisterSuperpowers(superpowers.NewPillar(cfg, logger))
	e.RegisterGSD(gsd.NewPillar(logger))
	e.RegisterCeremonyScaler(ceremony.NewScaler(cfg, logger))
	e.RegisterSpecMemory(memory.NewStore())

	return e
}

// defaultClassification returns a medium-effort classification
// suitable for flow execution in stress tests.
func defaultClassification() *types.EffortClassification {
	return &types.EffortClassification{
		Level:           types.EffortMedium,
		Confidence:      0.7,
		CeremonyLevel:   types.CeremonyLight,
		RequiresDebate:  true,
		RequiresSpecKit: false,
		Signals:         []string{"stress_test"},
		Reasoning:       "Stress test classification",
	}
}

// --- 1. Concurrent Flow Execution ---

func TestStress_ConcurrentFlowExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 50

	e := newTestEngine()
	ctx := context.Background()

	var successCount atomic.Int64
	var failureCount atomic.Int64
	var wg sync.WaitGroup

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				request := fmt.Sprintf(
					"stress flow request %d", idx,
				)
				cl := defaultClassification()

				result, err := e.ExecuteFlow(
					ctx, request, cl,
				)
				if err != nil || !result.Success {
					failureCount.Add(1)
					return
				}
				successCount.Add(1)
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// All goroutines completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected: goroutines did not " +
			"complete within timeout")
	}

	total := successCount.Load() + failureCount.Load()
	assert.Equal(t, int64(goroutines), total,
		"all goroutines must complete")
	assert.Greater(t, successCount.Load(), int64(0),
		"at least some flows must succeed")
	t.Logf("completed: success=%d failure=%d",
		successCount.Load(), failureCount.Load())
}

// --- 2. High Volume Classification ---

func TestStress_HighVolumeClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const iterations = 1000

	classifier := intent.NewClassifier()
	validLevels := map[types.EffortLevel]bool{
		types.EffortQuick:  true,
		types.EffortMedium: true,
		types.EffortLarge:  true,
		types.EffortEpic:   true,
	}

	requests := []string{
		"fix typo in readme",
		"add feature for user authentication",
		"implement new database service with api",
		"complete rewrite of entire system from scratch",
		"update comment in helper",
		"refactor the caching layer",
		"build a full platform for analytics",
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < iterations; i++ {
			req := requests[i%len(requests)]
			result := classifier.Classify(req)

			assert.True(t, validLevels[result.Level],
				"classification %d returned invalid "+
					"effort level: %s", i, result.Level)
			assert.NotEmpty(t, result.Reasoning,
				"classification %d missing reasoning", i)
		}
	}()

	select {
	case <-done:
		t.Logf("classified %d requests successfully",
			iterations)
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected during classification")
	}
}

// --- 3. Concurrent Classification ---

func TestStress_ConcurrentClassification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 100
	const perGoroutine = 10

	classifier := intent.NewClassifier()
	var completedCount atomic.Int64
	var wg sync.WaitGroup

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(gID int) {
				defer wg.Done()

				for j := 0; j < perGoroutine; j++ {
					req := fmt.Sprintf(
						"implement service %d-%d", gID, j,
					)
					result := classifier.Classify(req)
					assert.NotNil(t, result,
						"goroutine %d iteration %d: "+
							"nil result", gID, j)
					completedCount.Add(1)
				}
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected during concurrent " +
			"classification")
	}

	expected := int64(goroutines * perGoroutine)
	assert.Equal(t, expected, completedCount.Load(),
		"all classifications must complete")
}

// --- 4. Spec Memory Under Load ---

func TestStress_SpecMemoryUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const writers = 50
	const readers = 50

	store := memory.NewStore()
	ctx := context.Background()
	var wg sync.WaitGroup
	var writeCount atomic.Int64
	var readCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < writers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				spec := &types.Specification{
					ID:    fmt.Sprintf("spec-%d", idx),
					Title: fmt.Sprintf("Spec Title %d", idx),
					Description: fmt.Sprintf(
						"Description for spec %d", idx,
					),
					Source:       types.SourceSpecKit,
					QualityScore: 0.8,
				}
				err := store.Store(ctx, spec)
				if err == nil {
					writeCount.Add(1)
				}
			}(i)
		}

		for i := 0; i < readers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				query := fmt.Sprintf("Spec Title %d", idx%10)
				results, err := store.Search(ctx, query, 5)
				if err == nil {
					_ = results
					readCount.Add(1)
				}
			}(i)
		}

		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in spec memory test")
	}

	assert.Equal(t, int64(writers), writeCount.Load(),
		"all writes must succeed")
	assert.Equal(t, int64(readers), readCount.Load(),
		"all reads must succeed")
	assert.Equal(t, writers, store.SpecCount(),
		"store must contain all written specs")
}

// --- 5. Metrics Tracker Concurrency ---

func TestStress_MetricsTrackerConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 100

	tracker := metrics.NewMetrics()
	var wg sync.WaitGroup

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				effort := types.EffortLevels()[idx%4]
				tracker.RecordFlowStart(effort)

				phaseResult := &types.PhaseResult{
					Phase:        types.AllPhases()[idx%7],
					Success:      idx%3 != 0,
					QualityScore: 0.5 + float64(idx%5)*0.1,
					Duration:     time.Millisecond * time.Duration(idx),
				}
				tracker.RecordPhase(phaseResult)

				flowResult := &types.FlowResult{
					Success:       idx%2 == 0,
					CeremonyLevel: types.CeremonyLight,
					Duration:      time.Millisecond * time.Duration(idx),
				}
				tracker.RecordFlowComplete(flowResult)
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in metrics test")
	}

	snap := tracker.Snapshot()

	assert.Equal(t, int64(goroutines), snap.FlowsStarted,
		"all flow starts must be recorded")

	completedPlusFailed := snap.FlowsCompleted + snap.FlowsFailed
	assert.Equal(t, int64(goroutines), completedPlusFailed,
		"completed + failed must equal total goroutines")

	assert.Equal(t, int64(goroutines), snap.PhasesExecuted,
		"all phase executions must be recorded")

	totalEffort := int64(0)
	for _, count := range snap.EffortCounts {
		totalEffort += count
	}
	assert.Equal(t, int64(goroutines), totalEffort,
		"effort counts must sum to total goroutines")
}

// --- 6. Parallel Executor Saturation ---

func TestStress_ParallelExecutorSaturation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const taskCount = 50
	const maxWorkers = 2

	executor := parallel_execution.NewExecutor(
		maxWorkers, 5*time.Second,
	)
	ctx := context.Background()

	var concurrentCount atomic.Int64
	var maxConcurrent atomic.Int64

	tasks := make(map[string]parallel_execution.TaskFunc)
	for i := 0; i < taskCount; i++ {
		taskID := fmt.Sprintf("task-%d", i)
		tasks[taskID] = func(
			ctx context.Context,
		) (string, error) {
			current := concurrentCount.Add(1)

			// Track peak concurrency
			for {
				old := maxConcurrent.Load()
				if current <= old {
					break
				}
				if maxConcurrent.CompareAndSwap(old, current) {
					break
				}
			}

			time.Sleep(5 * time.Millisecond)
			concurrentCount.Add(-1)
			return "done", nil
		}
	}

	done := make(chan struct{})
	var results []parallel_execution.Result

	go func() {
		defer close(done)
		results = executor.Execute(ctx, tasks)
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in parallel executor test")
	}

	assert.Len(t, results, taskCount,
		"all tasks must produce results")

	for _, r := range results {
		assert.NoError(t, r.Error,
			"task %s should not error", r.TaskID)
		assert.Equal(t, "done", r.Output,
			"task %s output mismatch", r.TaskID)
	}

	assert.LessOrEqual(t, maxConcurrent.Load(),
		int64(maxWorkers),
		"concurrency must not exceed max workers")

	completed, failed := executor.Stats()
	assert.Equal(t, taskCount, completed,
		"all tasks must complete")
	assert.Equal(t, 0, failed,
		"no tasks should fail")
}

// --- 7. GSD Milestone Rapid Updates ---

func TestStress_GSDMilestoneRapidUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const updaters = 50

	logger := newTestLogger()
	pillar := gsd.NewPillar(logger)
	ctx := context.Background()

	spec := &types.Specification{
		ID:    "stress-spec",
		Title: "Stress Test Spec",
	}

	tasks := []types.Task{
		{ID: "t1", Name: "Task 1", Priority: "high"},
		{ID: "t2", Name: "Task 2", Priority: "high"},
		{ID: "t3", Name: "Task 3", Priority: "medium"},
		{ID: "t4", Name: "Task 4", Priority: "medium"},
	}

	milestones, err := pillar.CreateMilestones(
		ctx, spec, tasks,
	)
	require.NoError(t, err, "milestone creation must succeed")
	require.NotEmpty(t, milestones,
		"must create at least one milestone")

	milestoneID := milestones[0].ID

	var wg sync.WaitGroup
	var panicCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < updaters; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						panicCount.Add(1)
					}
				}()

				taskResults := []types.TaskResult{
					{
						TaskID:  "t1",
						Success: idx%2 == 0,
					},
					{
						TaskID:  "t2",
						Success: idx%3 == 0,
					},
				}

				_, err := pillar.TrackProgress(
					ctx, milestoneID, taskResults,
				)
				assert.NoError(t, err,
					"TrackProgress must not error")
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in GSD milestone test")
	}

	assert.Equal(t, int64(0), panicCount.Load(),
		"no panics should occur during rapid updates")

	progress, err := pillar.GetProgress("")
	assert.NoError(t, err, "GetProgress must not error")
	assert.GreaterOrEqual(t, progress, 0.0,
		"final progress must be non-negative")
	assert.LessOrEqual(t, progress, 1.0,
		"final progress must not exceed 1.0")
}

// --- 8. Spec Memory Feature Flood ---

func TestStress_SpecMemoryFeatureFlood(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const entryCount = 500
	const searchGoroutines = 20

	idx := spec_memory.NewIndex()

	for i := 0; i < entryCount; i++ {
		entry := spec_memory.Entry{
			ID:    fmt.Sprintf("entry-%d", i),
			Title: fmt.Sprintf("Feature %d specification", i),
			Content: fmt.Sprintf(
				"Detailed content for feature %d "+
					"with authentication and database", i,
			),
			Tags:  []string{"feature", fmt.Sprintf("tag-%d", i%10)},
			Score: float64(i%10) * 0.1,
		}
		err := idx.Add(entry)
		require.NoError(t, err,
			"adding entry %d must succeed", i)
	}

	assert.Equal(t, entryCount, idx.Count(),
		"index must contain all entries")

	var wg sync.WaitGroup
	var searchCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		queries := []string{
			"Feature 1", "authentication", "database",
			"specification", "content",
		}

		for i := 0; i < searchGoroutines; i++ {
			wg.Add(1)
			go func(gID int) {
				defer wg.Done()

				query := queries[gID%len(queries)]
				results := idx.Search(query, 10)
				assert.NotNil(t, results,
					"search from goroutine %d must "+
						"return non-nil", gID)
				searchCount.Add(1)
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in spec memory flood test")
	}

	assert.Equal(t, int64(searchGoroutines),
		searchCount.Load(),
		"all search goroutines must complete")
	assert.Equal(t, entryCount, idx.Count(),
		"entry count must remain stable after searches")
}

// --- 9. Cross-Project Knowledge Flood ---

func TestStress_CrossProjectKnowledgeFlood(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const entryCount = 200
	const searchGoroutines = 30

	te := cross_project.NewTransferEngine()

	for i := 0; i < entryCount; i++ {
		k := cross_project.Knowledge{
			ID:        fmt.Sprintf("know-%d", i),
			Project:   fmt.Sprintf("project-%d", i%5),
			Category:  "architecture",
			Content:   fmt.Sprintf("Pattern %d for caching layer", i),
			Tags:      []string{"caching", fmt.Sprintf("v%d", i%3)},
			Relevance: float64(i%10) * 0.1,
		}
		err := te.Add(k)
		require.NoError(t, err,
			"adding knowledge %d must succeed", i)
	}

	assert.Equal(t, entryCount, te.Count(),
		"engine must contain all knowledge entries")

	var wg sync.WaitGroup
	var panicCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < searchGoroutines; i++ {
			wg.Add(1)
			go func(gID int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						panicCount.Add(1)
					}
				}()

				results := te.Search("caching", 20)
				assert.NotNil(t, results,
					"search from goroutine %d must "+
						"return non-nil", gID)
				assert.Greater(t, len(results), 0,
					"search from goroutine %d must "+
						"find results", gID)
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in cross-project " +
			"knowledge test")
	}

	assert.Equal(t, int64(0), panicCount.Load(),
		"no panics should occur during concurrent search")
}

// --- 10. Skill Learning Rapid Updates ---

func TestStress_SkillLearningRapidUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const skillCount = 20
	const updaters = 50

	learner := skill_learning.NewLearner()

	for i := 0; i < skillCount; i++ {
		skill := skill_learning.Skill{
			ID:          fmt.Sprintf("skill-%d", i),
			Name:        fmt.Sprintf("Skill %d", i),
			Domain:      "engineering",
			Proficiency: 0.5,
		}
		err := learner.RegisterSkill(skill)
		require.NoError(t, err,
			"registering skill %d must succeed", i)
	}

	assert.Equal(t, skillCount, learner.SkillCount(),
		"all skills must be registered")

	var wg sync.WaitGroup
	var panicCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < updaters; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						panicCount.Add(1)
					}
				}()

				skillID := fmt.Sprintf(
					"skill-%d", idx%skillCount,
				)
				score := float64(idx%10) * 0.1
				_ = learner.RecordUsage(skillID, score)
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in skill learning test")
	}

	assert.Equal(t, int64(0), panicCount.Load(),
		"no panics should occur during rapid skill updates")

	for i := 0; i < skillCount; i++ {
		skillID := fmt.Sprintf("skill-%d", i)
		skill, err := learner.GetSkill(skillID)
		require.NoError(t, err,
			"getting skill %s must succeed", skillID)
		assert.GreaterOrEqual(t, skill.Proficiency, 0.0,
			"skill %s proficiency must be >= 0.0", skillID)
		assert.LessOrEqual(t, skill.Proficiency, 1.0,
			"skill %s proficiency must be <= 1.0", skillID)
	}
}

// --- 11. Engine Resume Under Load ---

func TestStress_EngineResumeUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const flowCount = 20
	const resumersPerFlow = 3

	e := newTestEngine()
	ctx := context.Background()

	flowIDs := make([]string, 0, flowCount)

	for i := 0; i < flowCount; i++ {
		request := fmt.Sprintf("resume test flow %d", i)
		cl := defaultClassification()

		result, err := e.ExecuteFlow(ctx, request, cl)
		require.NoError(t, err,
			"flow %d execution must succeed", i)
		require.True(t, result.Success,
			"flow %d must succeed", i)
		flowIDs = append(flowIDs, result.FlowID)
	}

	var wg sync.WaitGroup
	var panicCount atomic.Int64
	var resumeCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		for _, flowID := range flowIDs {
			for r := 0; r < resumersPerFlow; r++ {
				wg.Add(1)
				go func(fID string, rIdx int) {
					defer wg.Done()
					defer func() {
						if r := recover(); r != nil {
							panicCount.Add(1)
						}
					}()

					result, err := e.ResumeFlow(
						ctx, fID, "resume request",
					)
					if err == nil && result != nil {
						resumeCount.Add(1)
					}
				}(flowID, r)
			}
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in engine resume test")
	}

	assert.Equal(t, int64(0), panicCount.Load(),
		"no panics should occur during concurrent resume")

	expectedResumes := int64(flowCount * resumersPerFlow)
	assert.Equal(t, expectedResumes, resumeCount.Load(),
		"all resume attempts must succeed for "+
			"completed flows")
}

// --- 12. Adapter Formatting Under Load ---

func TestStress_AdapterFormattingUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	const goroutines = 50

	e := newTestEngine()
	ctx := context.Background()

	result, err := e.ExecuteFlow(
		ctx, "adapter stress test", defaultClassification(),
	)
	require.NoError(t, err, "flow execution must succeed")
	require.True(t, result.Success, "flow must succeed")

	categories := []struct {
		name    string
		adapter types.CLIAgentAdapter
	}{
		{
			name: "markdown",
			adapter: newTestAdapter(
				"opencode", "markdown",
			),
		},
		{
			name: "toml",
			adapter: newTestAdapter(
				"custom-toml", "toml",
			),
		},
		{
			name: "github",
			adapter: newTestAdapter(
				"copilot", "github",
			),
		},
		{
			name: "generic",
			adapter: newTestAdapter(
				"aider", "generic",
			),
		},
		{
			name: "minimal",
			adapter: newTestAdapter(
				"minimal-agent", "minimal",
			),
		},
	}

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var panicCount atomic.Int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						panicCount.Add(1)
					}
				}()

				cat := categories[idx%len(categories)]
				output, fmtErr := cat.adapter.FormatOutput(
					result,
				)
				if fmtErr == nil && output != "" {
					successCount.Add(1)
				}
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		// Completed
	case <-time.After(deadlockTimeout):
		t.Fatal("deadlock detected in adapter " +
			"formatting test")
	}

	assert.Equal(t, int64(0), panicCount.Load(),
		"no panics should occur during formatting")
	assert.Equal(t, int64(goroutines), successCount.Load(),
		"all format operations must produce output")
}

// testAdapter wraps adapter fields for stress testing by
// implementing the CLIAgentAdapter interface with simple
// formatting based on category.
type testAdapter struct {
	agentType string
	category  string
}

func newTestAdapter(
	agentType string,
	category string,
) *testAdapter {
	return &testAdapter{
		agentType: agentType,
		category:  category,
	}
}

func (a *testAdapter) FormatOutput(
	result *types.FlowResult,
) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil result")
	}
	switch a.category {
	case "markdown":
		return fmt.Sprintf(
			"# Flow %s\nEffort: %s\nQuality: %.2f\n",
			result.FlowID, result.EffortLevel,
			result.OverallQualityScore,
		), nil
	case "toml":
		return fmt.Sprintf(
			"[flow]\nid = %q\neffort = %q\n",
			result.FlowID, result.EffortLevel,
		), nil
	default:
		return fmt.Sprintf(
			"{\"flow_id\":%q,\"effort\":%q}",
			result.FlowID, result.EffortLevel,
		), nil
	}
}

func (a *testAdapter) ParseInput(
	input string,
) (string, error) {
	return input, nil
}

func (a *testAdapter) GetAgentType() string {
	return a.agentType
}

func (a *testAdapter) SupportsStreaming() bool {
	return true
}
