package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/adapters"
	"digital.vasic.helixspecifier/pkg/ceremony"
	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/engine"
	"digital.vasic.helixspecifier/pkg/features/brownfield"
	"digital.vasic.helixspecifier/pkg/features/constitution_code"
	"digital.vasic.helixspecifier/pkg/features/cross_project"
	"digital.vasic.helixspecifier/pkg/features/debate_architecture"
	"digital.vasic.helixspecifier/pkg/features/nyquist_tdd"
	"digital.vasic.helixspecifier/pkg/features/parallel_execution"
	"digital.vasic.helixspecifier/pkg/features/predictive_spec"
	"digital.vasic.helixspecifier/pkg/features/skill_learning"
	"digital.vasic.helixspecifier/pkg/features/spec_memory"
	"digital.vasic.helixspecifier/pkg/features/adaptive_ceremony"
	"digital.vasic.helixspecifier/pkg/gsd"
	"digital.vasic.helixspecifier/pkg/intent"
	"digital.vasic.helixspecifier/pkg/memory"
	"digital.vasic.helixspecifier/pkg/metrics"
	"digital.vasic.helixspecifier/pkg/speckit"
	"digital.vasic.helixspecifier/pkg/superpowers"
	"digital.vasic.helixspecifier/pkg/types"

	"github.com/sirupsen/logrus"
)

func init() {
	runtime.GOMAXPROCS(2)
}

// --- Helpers ---

func newBenchLogger() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.PanicLevel)
	return l
}

func newBenchConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.AdaptiveCeremony = false
	cfg.SpecMemoryEnabled = false
	return cfg
}

func newBenchEngine() *engine.FusionEngine {
	cfg := newBenchConfig()
	logger := newBenchLogger()

	eng := engine.New(cfg, logger)
	eng.RegisterSpecKit(speckit.NewPillar(cfg, logger))
	eng.RegisterSuperpowers(superpowers.NewPillar(cfg, logger))
	eng.RegisterGSD(gsd.NewPillar(logger))
	eng.RegisterCeremonyScaler(ceremony.NewScaler(cfg, logger))
	eng.RegisterSpecMemory(memory.NewStore())
	return eng
}

func makeTasks(n int) []types.Task {
	tasks := make([]types.Task, n)
	for i := range n {
		tasks[i] = types.Task{
			ID:          fmt.Sprintf("task-%d", i),
			Name:        fmt.Sprintf("Task %d", i),
			Description: fmt.Sprintf("Benchmark task number %d", i),
			Priority:    "medium",
			Status:      "pending",
			Acceptance:  []string{"criterion-a", "criterion-b"},
			FilesAffected: []string{
				fmt.Sprintf("pkg/module%d/file.go", i),
			},
		}
	}
	return tasks
}

func makeSpec() *types.Specification {
	return &types.Specification{
		ID:          "bench-spec-001",
		Title:       "Benchmark specification",
		Description: "Specification used for benchmark testing",
		Source:      types.SourceFusion,
		CreatedAt:   time.Now(),
		Requirements: []types.Requirement{
			{
				ID:          "req-1",
				Type:        "functional",
				Description: "Must pass benchmarks",
				Priority:    "high",
			},
		},
		QualityScore: 0.85,
	}
}

func makeClassification(
	level types.EffortLevel,
) *types.EffortClassification {
	return &types.EffortClassification{
		Level:         level,
		Confidence:    0.8,
		CeremonyLevel: types.CeremonyForEffort(level),
		RequiresDebate: level != types.EffortQuick,
		RequiresSpecKit: level == types.EffortLarge ||
			level == types.EffortEpic,
		Signals:   []string{"benchmark"},
		Reasoning: "Benchmark classification",
	}
}

func makeFlowResult() *types.FlowResult {
	return &types.FlowResult{
		FlowID:        "bench-flow-001",
		EffortLevel:   types.EffortMedium,
		CeremonyLevel: types.CeremonyLight,
		StartTime:     time.Now().Add(-5 * time.Second),
		EndTime:       time.Now(),
		Duration:      5 * time.Second,
		Success:       true,
		Stages:        types.StagesForEffort(types.EffortMedium),
		PhaseResults: []types.PhaseResult{
			{
				Phase:        types.PhaseSpecify,
				Success:      true,
				QualityScore: 0.8,
				Duration:     2 * time.Second,
			},
			{
				Phase:        types.PhasePlan,
				Success:      true,
				QualityScore: 0.75,
				Duration:     1 * time.Second,
			},
			{
				Phase:        types.PhaseImplement,
				Success:      true,
				QualityScore: 0.9,
				Duration:     2 * time.Second,
			},
		},
		OverallQualityScore: 0.82,
	}
}

// --- 1. Classifier ---

func BenchmarkClassifier_Classify(b *testing.B) {
	cl := intent.NewClassifier()
	cases := map[string]string{
		"short":  "fix typo in readme",
		"medium": "implement a new authentication service with JWT tokens and rate limiting",
		"long": "build an entire system from scratch that handles user management, " +
			"authentication, authorization, database migrations, API gateway, " +
			"service mesh, observability, and distributed tracing across all microservices " +
			"with full test coverage and challenge scripts for every component",
	}

	for name, input := range cases {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				cl.Classify(input)
			}
		})
	}
}

// --- 2. Engine ExecuteFlow ---

func BenchmarkEngine_ExecuteFlow(b *testing.B) {
	eng := newBenchEngine()
	ctx := context.Background()
	classification := makeClassification(types.EffortQuick)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = eng.ExecuteFlow(
			ctx,
			"add logging to handler",
			classification,
		)
	}
}

// --- 3. Engine ClassifyEffort ---

func BenchmarkEngine_ClassifyEffort(b *testing.B) {
	eng := newBenchEngine()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = eng.ClassifyEffort(
			ctx,
			"implement new database migration system",
		)
	}
}

// --- 4. CeremonyScaler Scale ---

func BenchmarkCeremonyScaler_Scale(b *testing.B) {
	cfg := newBenchConfig()
	scaler := ceremony.NewScaler(cfg, newBenchLogger())

	for _, level := range types.EffortLevels() {
		cl := makeClassification(level)
		b.Run(string(level), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				scaler.Scale(cl)
			}
		})
	}
}

// --- 5. SpecKit ExecutePhase ---

func BenchmarkSpecKit_ExecutePhase(b *testing.B) {
	cfg := newBenchConfig()
	pillar := speckit.NewPillar(cfg, newBenchLogger())
	ctx := context.Background()

	input := &types.PhaseInput{
		UserRequest:    "implement user authentication",
		PreviousOutput: "specification output from prior phase",
		Metadata:       map[string]interface{}{"effort_level": "large"},
	}

	for _, phase := range types.AllPhases() {
		b.Run(string(phase), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, _ = pillar.ExecutePhase(ctx, phase, input)
			}
		})
	}
}

// --- 6. Superpowers DispatchSubagents ---

func BenchmarkSuperpowers_DispatchSubagents(b *testing.B) {
	cfg := newBenchConfig()
	pillar := superpowers.NewPillar(cfg, newBenchLogger())
	ctx := context.Background()

	for _, count := range []int{2, 4, 8} {
		tasks := makeTasks(count)
		b.Run(fmt.Sprintf("agents_%d", count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, _ = pillar.DispatchSubagents(ctx, tasks, count)
			}
		})
	}
}

// --- 7. GSD CreateMilestones ---

func BenchmarkGSD_CreateMilestones(b *testing.B) {
	ctx := context.Background()
	spec := makeSpec()

	for _, count := range []int{5, 10, 25, 50} {
		tasks := makeTasks(count)
		b.Run(fmt.Sprintf("tasks_%d", count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				p := gsd.NewPillar(newBenchLogger())
				_, _ = p.CreateMilestones(ctx, spec, tasks)
			}
		})
	}
}

// --- 8. Memory Store ---

func BenchmarkMemory_Store(b *testing.B) {
	ctx := context.Background()

	for _, count := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("entries_%d", count), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				store := memory.NewStore()
				for i := range count {
					spec := &types.Specification{
						ID:          fmt.Sprintf("spec-%d", i),
						Title:       fmt.Sprintf("Spec number %d", i),
						Description: "Benchmark specification entry",
						Source:      types.SourceFusion,
					}
					_ = store.Store(ctx, spec)
				}
			}
		})
	}
}

// --- 9. Memory Search ---

func BenchmarkMemory_Search(b *testing.B) {
	ctx := context.Background()
	store := memory.NewStore()

	for i := range 500 {
		spec := &types.Specification{
			ID:          fmt.Sprintf("spec-%d", i),
			Title:       fmt.Sprintf("Module %d authentication service", i),
			Description: fmt.Sprintf("Description for spec %d with database integration", i),
			Source:      types.SourceFusion,
		}
		_ = store.Store(ctx, spec)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, _ = store.Search(ctx, "authentication", 10)
	}
}

// --- 10. Adapter FormatOutput ---

func BenchmarkAdapter_FormatOutput(b *testing.B) {
	result := makeFlowResult()
	categories := map[string]adapters.AgentCategory{
		"markdown": adapters.CategoryMarkdown,
		"toml":     adapters.CategoryTOML,
		"generic":  adapters.CategoryGeneric,
	}

	for name, cat := range categories {
		adapter := adapters.NewGenericAdapter(name, cat)
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				_, _ = adapter.FormatOutput(result)
			}
		})
	}
}

// --- 11. Metrics Record ---

func BenchmarkMetrics_Record(b *testing.B) {
	m := metrics.NewMetrics()
	result := makeFlowResult()
	phaseResult := &types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		QualityScore: 0.82,
		Duration:     100 * time.Millisecond,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		m.RecordFlowStart(types.EffortMedium)
		m.RecordPhase(phaseResult)
		m.RecordFlowComplete(result)
	}
}

// --- 12. Parallel Execution ---

func BenchmarkParallelExecution(b *testing.B) {
	ctx := context.Background()

	workerCounts := []int{1, 2, 4, 8}
	taskCounts := []int{10, 50}

	for _, workers := range workerCounts {
		for _, taskCount := range taskCounts {
			name := fmt.Sprintf(
				"workers_%d/tasks_%d",
				workers, taskCount,
			)
			b.Run(name, func(b *testing.B) {
				exec := parallel_execution.NewExecutor(
					workers, 5*time.Second,
				)
				tasks := make(
					map[string]parallel_execution.TaskFunc,
					taskCount,
				)
				for i := range taskCount {
					id := fmt.Sprintf("t-%d", i)
					tasks[id] = func(
						_ context.Context,
					) (string, error) {
						return "done", nil
					}
				}

				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					exec.Execute(ctx, tasks)
				}
			})
		}
	}
}

// --- 13. Constitution Validate ---

func BenchmarkConstitution_Validate(b *testing.B) {
	text := "this document covers testing, security, documentation, " +
		"performance, containerization, monitoring, configuration, " +
		"stability, networking, resource limits, observability, " +
		"git operations, architecture, quality, safety, " +
		"design patterns, software principles, ci/cd, decoupling, " +
		"health checks, stress testing, challenge scripts, " +
		"infrastructure, lazy loading, memory safety, scanning, " +
		"container builds, http3, resource management, " +
		"dead code removal, no broken components, " +
		"comprehensive tests, documentation sync, " +
		"non-interactive execution, ssh only, manual ci/cd, " +
		"mandatory orchestration, full containerization, " +
		"unified configuration, gitspec compliance, " +
		"rock-solid changes, certificate management, " +
		"complete documentation, hundred percent coverage, " +
		"rule-50, rule-60, rule-70, rule-80, rule-90, rule-100"

	for _, count := range []int{5, 10, 50} {
		b.Run(fmt.Sprintf("rules_%d", count), func(b *testing.B) {
			c := constitution_code.NewConstitution()
			for i := range count {
				_ = c.AddRule(constitution_code.Rule{
					ID:        fmt.Sprintf("rule-%d", i),
					Category:  "testing",
					Priority:  1,
					Title:     fmt.Sprintf("rule-%d", i),
					Mandatory: true,
				})
			}

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				c.Validate(text)
			}
		})
	}
}

// --- 14. Nyquist TDD Check ---

func BenchmarkNyquistTDD_Check(b *testing.B) {
	b.Run("compliant", func(b *testing.B) {
		tr := nyquist_tdd.NewTracker()
		tr.RecordImplementation(100)
		tr.RecordTests(250)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = tr.Check()
		}
	})

	b.Run("non_compliant", func(b *testing.B) {
		tr := nyquist_tdd.NewTracker()
		tr.RecordImplementation(100)
		tr.RecordTests(50)

		b.ReportAllocs()
		b.ResetTimer()
		for range b.N {
			_ = tr.Check()
		}
	})
}

// --- 15. Debate Architecture BestPosition ---

func BenchmarkDebateArchitecture_BestPosition(b *testing.B) {
	for _, roundCount := range []int{1, 3, 5, 10} {
		b.Run(
			fmt.Sprintf("rounds_%d", roundCount),
			func(b *testing.B) {
				debate := debate_architecture.NewDebate(
					"bench-debate", "spec topic", roundCount,
				)
				for r := range roundCount {
					round := debate_architecture.Round{
						Positions: []debate_architecture.Position{
							{
								AgentID:  "agent-a",
								Argument: "Position A argument",
								Score:    0.7 + float64(r)*0.02,
							},
							{
								AgentID:  "agent-b",
								Argument: "Position B argument",
								Score:    0.65 + float64(r)*0.03,
							},
							{
								AgentID:  "agent-c",
								Argument: "Position C argument",
								Score:    0.6 + float64(r)*0.01,
							},
						},
						Duration: 500 * time.Millisecond,
					}
					_ = debate.AddRound(round)
				}

				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					debate.BestPosition()
				}
			},
		)
	}
}

// --- 16. Skill Learning RecordUsage ---

func BenchmarkSkillLearning_RecordUsage(b *testing.B) {
	learner := skill_learning.NewLearner()
	skills := []string{
		"code-review", "spec-writing", "tdd",
		"architecture", "debugging",
	}
	for _, s := range skills {
		_ = learner.RegisterSkill(skill_learning.Skill{
			ID:     s,
			Name:   s,
			Domain: "engineering",
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		skillID := skills[i%len(skills)]
		score := 0.5 + float64(i%50)*0.01
		_ = learner.RecordUsage(skillID, score)
	}
}

// --- 17. Brownfield Analyze ---

func BenchmarkBrownfield_Analyze(b *testing.B) {
	for _, count := range []int{10, 50, 100} {
		b.Run(
			fmt.Sprintf("patterns_%d", count),
			func(b *testing.B) {
				analyzer := brownfield.NewAnalyzer()
				for i := range count {
					_ = analyzer.RecordPattern(
						fmt.Sprintf("pattern-%d", i),
						0.7+float64(i%30)*0.01,
					)
				}
				analyzer.AddDependency("github.com/gin-gonic/gin")
				analyzer.AddDependency("github.com/sirupsen/logrus")

				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					analyzer.Analyze()
				}
			},
		)
	}
}

// --- 18. Predictive Spec Predict ---

func BenchmarkPredictiveSpec_Predict(b *testing.B) {
	for _, histSize := range []int{5, 20, 50} {
		b.Run(
			fmt.Sprintf("history_%d", histSize),
			func(b *testing.B) {
				pred := predictive_spec.NewPredictor()
				for i := range histSize {
					pred.AddHistory(
						fmt.Sprintf("event-%d: module added", i),
					)
				}

				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					pred.Predict()
				}
			},
		)
	}
}

// --- 19. Cross Project Search ---

func BenchmarkCrossProject_Search(b *testing.B) {
	for _, count := range []int{10, 50, 200} {
		b.Run(
			fmt.Sprintf("entries_%d", count),
			func(b *testing.B) {
				te := cross_project.NewTransferEngine()
				for i := range count {
					_ = te.Add(cross_project.Knowledge{
						ID:       fmt.Sprintf("k-%d", i),
						Project:  "helix-agent",
						Category: "architecture",
						Content: fmt.Sprintf(
							"Pattern %d: circuit breaker for provider %d",
							i, i,
						),
						Tags:      []string{"resilience", "pattern"},
						Relevance: 0.8,
					})
				}

				b.ReportAllocs()
				b.ResetTimer()
				for range b.N {
					te.Search("circuit breaker", 10)
				}
			},
		)
	}
}

// --- 20. Spec Memory Search ---

func BenchmarkSpecMemory_Search(b *testing.B) {
	idx := spec_memory.NewIndex()
	for i := range 200 {
		_ = idx.Add(spec_memory.Entry{
			ID:      fmt.Sprintf("entry-%d", i),
			Title:   fmt.Sprintf("Specification for module %d", i),
			Content: fmt.Sprintf("Authentication and authorization patterns entry %d", i),
			Tags:    []string{"auth", "module"},
			Score:   0.75,
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		idx.Search("authentication", 10)
	}
}

// --- 21. Engine ExecuteFlow Parallel ---

func BenchmarkEngine_ExecuteFlow_Parallel(b *testing.B) {
	eng := newBenchEngine()
	classification := makeClassification(types.EffortQuick)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			_, _ = eng.ExecuteFlow(
				ctx,
				"add logging to handler",
				classification,
			)
		}
	})
}

// --- Supplementary: Adaptive Ceremony ---

func BenchmarkAdaptiveCeremony_Adjust(b *testing.B) {
	adjuster := adaptive_ceremony.NewAdjuster(0.9, 0.5)
	results := []types.PhaseResult{
		{Phase: types.PhaseSpecify, QualityScore: 0.85, Success: true},
		{Phase: types.PhasePlan, QualityScore: 0.78, Success: true},
		{Phase: types.PhaseImplement, QualityScore: 0.82, Success: true},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		adjuster.Adjust(types.CeremonyStandard, results)
	}
}
