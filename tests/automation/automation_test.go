package automation

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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

// --- Compile-time interface assertions ---

var _ types.SpecEngine = (*engine.FusionEngine)(nil)
var _ types.SpecKitPillar = (*speckit.Pillar)(nil)
var _ types.SuperpowersPillar = (*superpowers.Pillar)(nil)
var _ types.GSDPillar = (*gsd.Pillar)(nil)
var _ types.CeremonyScaler = (*ceremony.Scaler)(nil)
var _ types.SpecMemory = (*memory.Store)(nil)
var _ types.CLIAgentAdapter = (*adapters.GenericAdapter)(nil)

// projectRoot walks up from the current test directory to find
// the module root (the directory containing go.mod).
func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(
			filepath.Join(dir, "go.mod"),
		); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// newTestLogger returns a silent logger for automation tests.
func newTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	return logger
}

// --- 1. Project Structure ---

func TestAutomation_ProjectStructure(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	dirs := []string{
		"pkg/types",
		"pkg/config",
		"pkg/engine",
		"pkg/speckit",
		"pkg/superpowers",
		"pkg/gsd",
		"pkg/intent",
		"pkg/ceremony",
		"pkg/memory",
		"pkg/adapters",
		"pkg/metrics",
		"pkg/features/parallel_execution",
		"pkg/features/constitution_code",
		"pkg/features/nyquist_tdd",
		"pkg/features/debate_architecture",
		"pkg/features/skill_learning",
		"pkg/features/brownfield",
		"pkg/features/predictive_spec",
		"pkg/features/cross_project",
		"pkg/features/adaptive_ceremony",
		"pkg/features/spec_memory",
		"tests/security",
		"tests/integration",
		"tests/stress",
		"tests/benchmark",
		"tests/e2e",
		"tests/automation",
		"docs/guides",
	}

	for _, d := range dirs {
		t.Run(d, func(t *testing.T) {
			path := filepath.Join(root, d)
			info, err := os.Stat(path)
			require.NoError(t, err,
				"directory %s should exist", d,
			)
			assert.True(t, info.IsDir(),
				"%s should be a directory", d,
			)
		})
	}
}

// --- 2. Required Files ---

func TestAutomation_RequiredFiles(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	rootFiles := []string{
		"CLAUDE.md",
		"AGENTS.md",
		"README.md",
		"Makefile",
		".env.example",
		".gitignore",
		"go.mod",
		"go.sum",
	}

	for _, f := range rootFiles {
		t.Run("root/"+f, func(t *testing.T) {
			path := filepath.Join(root, f)
			info, err := os.Stat(path)
			require.NoError(t, err,
				"file %s should exist at project root", f,
			)
			assert.False(t, info.IsDir(),
				"%s should be a file", f,
			)
		})
	}

	guideFiles := []string{
		"docs/guides/architecture.md",
		"docs/guides/getting-started.md",
		"docs/guides/power-features.md",
	}

	for _, f := range guideFiles {
		t.Run(f, func(t *testing.T) {
			path := filepath.Join(root, f)
			info, err := os.Stat(path)
			require.NoError(t, err,
				"file %s should exist", f,
			)
			assert.False(t, info.IsDir(),
				"%s should be a file", f,
			)
		})
	}
}

// --- 3. Go Module Validity ---

func TestAutomation_GoModValidity(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	require.NoError(t, err, "go.mod should be readable")
	content := string(data)

	t.Run("module_name", func(t *testing.T) {
		assert.Contains(t, content,
			"module digital.vasic.helixspecifier",
			"go.mod must declare the correct module name",
		)
	})

	t.Run("go_version", func(t *testing.T) {
		assert.Contains(t, content, "go 1.24",
			"go.mod must require Go 1.24+",
		)
	})

	requiredDeps := []string{
		"github.com/stretchr/testify",
		"github.com/sirupsen/logrus",
		"github.com/google/uuid",
	}

	for _, dep := range requiredDeps {
		t.Run("dep/"+dep, func(t *testing.T) {
			assert.Contains(t, content, dep,
				"go.mod must include dependency %s", dep,
			)
		})
	}
}

// --- 4. Interface Contracts ---

func TestAutomation_InterfaceContracts(t *testing.T) {
	logger := newTestLogger()
	cfg := config.DefaultConfig()

	t.Run("SpecEngine", func(t *testing.T) {
		e := engine.New(cfg, logger)
		require.NotNil(t, e)
		assert.Equal(t, "HelixSpecifier", e.Name())
		assert.NotEmpty(t, e.Version())
	})

	t.Run("SpecKitPillar", func(t *testing.T) {
		p := speckit.NewPillar(cfg, logger)
		require.NotNil(t, p)
		phases := p.GetPhaseOrder(types.CeremonyStandard)
		assert.Equal(t, 7, len(phases))
		assert.True(t, p.ValidateTransition(
			types.PhaseConstitution, types.PhaseSpecify,
		))
	})

	t.Run("SuperpowersPillar", func(t *testing.T) {
		p := superpowers.NewPillar(cfg, logger)
		require.NotNil(t, p)
	})

	t.Run("GSDPillar", func(t *testing.T) {
		p := gsd.NewPillar(logger)
		require.NotNil(t, p)
		progress, err := p.GetProgress("nonexistent")
		assert.NoError(t, err)
		assert.Equal(t, 0.0, progress)
	})

	t.Run("CeremonyScaler", func(t *testing.T) {
		s := ceremony.NewScaler(cfg, logger)
		require.NotNil(t, s)
		level := s.Scale(&types.EffortClassification{
			Level: types.EffortLarge,
		})
		assert.Equal(t, types.CeremonyStandard, level)
	})

	t.Run("SpecMemory", func(t *testing.T) {
		store := memory.NewStore()
		require.NotNil(t, store)
		assert.Equal(t, 0, store.SpecCount())
		assert.Equal(t, 0, store.PatternCount())
	})

	t.Run("CLIAgentAdapter", func(t *testing.T) {
		a := adapters.NewGenericAdapter(
			"test-agent", adapters.CategoryGeneric,
		)
		require.NotNil(t, a)
		assert.Equal(t, "test-agent", a.GetAgentType())
		assert.True(t, a.SupportsStreaming())
	})

	t.Run("DebateFuncSetter_SpecKitPillar", func(t *testing.T) {
		p := speckit.NewPillar(cfg, logger)
		require.NotNil(t, p)

		// Verify *speckit.Pillar implements DebateFuncSetter
		var setter types.DebateFuncSetter = p
		require.NotNil(t, setter)

		// Verify SetDebateFunc is callable
		called := false
		setter.SetDebateFunc(func(
			ctx context.Context,
			topic string,
			rounds int,
			metadata map[string]interface{},
		) (string, float64, string, error) {
			called = true
			return "output", 0.9, "dbg-1", nil
		})

		// Execute a phase to confirm the func was set
		result, err := p.ExecutePhase(
			context.Background(),
			types.PhaseSpecify,
			&types.PhaseInput{
				UserRequest: "test request",
			},
		)
		require.NoError(t, err)
		assert.True(t, called,
			"debate func must be called after injection",
		)
		assert.Equal(t, 0.9, result.QualityScore)
	})

	t.Run("EffortClassifierFunc_Signature", func(t *testing.T) {
		c := intent.NewClassifier()
		require.NotNil(t, c)

		// Verify Classifier.Classify is assignable to
		// types.EffortClassifierFunc
		var fn types.EffortClassifierFunc = c.Classify
		require.NotNil(t, fn)

		// Invoke the func and verify the result
		result := fn("implement new auth service with database")
		require.NotNil(t, result)
		assert.NotEmpty(t, string(result.Level))
		assert.Greater(t, result.Confidence, 0.0)
		assert.NotEmpty(t, result.Reasoning)
		assert.Equal(t,
			types.CeremonyForEffort(result.Level),
			result.CeremonyLevel,
		)
	})
}

// --- 5. All Packages Compile ---

func TestAutomation_AllPackagesCompile(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	pkgRoot := filepath.Join(root, "pkg")
	fset := token.NewFileSet()
	parseErrors := []string{}

	err := filepath.Walk(pkgRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			_, parseErr := parser.ParseFile(
				fset, path, nil, parser.AllErrors,
			)
			if parseErr != nil {
				rel, _ := filepath.Rel(root, path)
				parseErrors = append(parseErrors,
					rel+": "+parseErr.Error(),
				)
			}
			return nil
		},
	)
	require.NoError(t, err, "walking pkg/ should not fail")
	assert.Empty(t, parseErrors,
		"all .go files under pkg/ must parse without errors: %v",
		parseErrors,
	)
}

// --- 6. Test File Existence ---

func TestAutomation_TestFileExistence(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	// Verify every package directory under pkg/ has a *_test.go
	pkgRoot := filepath.Join(root, "pkg")
	pkgDirs := map[string]bool{}

	err := filepath.Walk(pkgRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && path != pkgRoot {
				pkgDirs[path] = false
			}
			if !info.IsDir() &&
				strings.HasSuffix(info.Name(), "_test.go") {
				pkgDirs[filepath.Dir(path)] = true
			}
			return nil
		},
	)
	require.NoError(t, err, "walking pkg/ should not fail")

	for dir, hasTest := range pkgDirs {
		// Skip intermediate directories that only contain
		// subdirectories (e.g. pkg/features).
		hasGoFile := false
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if !e.IsDir() &&
				strings.HasSuffix(e.Name(), ".go") &&
				!strings.HasSuffix(e.Name(), "_test.go") {
				hasGoFile = true
				break
			}
		}
		if !hasGoFile {
			continue
		}

		rel, _ := filepath.Rel(root, dir)
		t.Run("pkg/"+rel, func(t *testing.T) {
			assert.True(t, hasTest,
				"package %s must have at least one "+
					"*_test.go file", rel,
			)
		})
	}

	// Verify test directories exist and have test files
	testDirs := []string{
		"tests/security",
		"tests/integration",
		"tests/stress",
		"tests/benchmark",
		"tests/e2e",
	}

	for _, d := range testDirs {
		t.Run(d, func(t *testing.T) {
			fullPath := filepath.Join(root, d)
			entries, err := os.ReadDir(fullPath)
			require.NoError(t, err,
				"test directory %s must exist", d,
			)

			hasTestFile := false
			for _, e := range entries {
				if strings.HasSuffix(
					e.Name(), "_test.go",
				) {
					hasTestFile = true
					break
				}
			}
			assert.True(t, hasTestFile,
				"%s must contain at least one *_test.go",
				d,
			)
		})
	}
}

// --- 7. Enum Completeness ---

func TestAutomation_EnumCompleteness(t *testing.T) {
	t.Run("AllPhases_count", func(t *testing.T) {
		phases := types.AllPhases()
		assert.Equal(t, 7, len(phases),
			"AllPhases must return exactly 7 phases",
		)
	})

	t.Run("AllPhases_order", func(t *testing.T) {
		phases := types.AllPhases()
		expected := []types.SpecKitPhase{
			types.PhaseConstitution,
			types.PhaseSpecify,
			types.PhaseClarify,
			types.PhasePlan,
			types.PhaseTasks,
			types.PhaseAnalyze,
			types.PhaseImplement,
		}
		assert.Equal(t, expected, phases,
			"AllPhases must return phases in correct order",
		)
	})

	t.Run("AllStages_count", func(t *testing.T) {
		stages := types.AllStages()
		assert.Equal(t, 5, len(stages),
			"AllStages must return exactly 5 stages",
		)
	})

	t.Run("EffortLevels_count", func(t *testing.T) {
		levels := types.EffortLevels()
		assert.Equal(t, 4, len(levels),
			"EffortLevels must return exactly 4 levels",
		)
	})

	t.Run("CeremonyForEffort_maps_all_levels",
		func(t *testing.T) {
			mapping := map[types.EffortLevel]types.CeremonyLevel{
				types.EffortQuick:  types.CeremonyMinimal,
				types.EffortMedium: types.CeremonyLight,
				types.EffortLarge:  types.CeremonyStandard,
				types.EffortEpic:   types.CeremonyHeavy,
			}
			for effort, expectedCeremony := range mapping {
				actual := types.CeremonyForEffort(effort)
				assert.Equal(t, expectedCeremony, actual,
					"CeremonyForEffort(%s) should be %s",
					effort, expectedCeremony,
				)
			}
		},
	)

	t.Run("PhasesForCeremony_counts", func(t *testing.T) {
		cases := []struct {
			ceremony types.CeremonyLevel
			count    int
		}{
			{types.CeremonyMinimal, 1},
			{types.CeremonyLight, 3},
			{types.CeremonyStandard, 7},
			{types.CeremonyHeavy, 7},
		}
		for _, tc := range cases {
			phases := types.PhasesForCeremony(tc.ceremony)
			assert.Equal(t, tc.count, len(phases),
				"PhasesForCeremony(%s) should return "+
					"%d phases", tc.ceremony, tc.count,
			)
		}
	})
}

// --- 8. Config Defaults ---

func TestAutomation_ConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	require.NotNil(t, cfg, "DefaultConfig must not return nil")

	t.Run("QuickMaxMinutes_positive", func(t *testing.T) {
		assert.Greater(t, cfg.QuickMaxMinutes, 0,
			"QuickMaxMinutes must be positive",
		)
	})

	t.Run("MediumMaxMinutes_gt_quick", func(t *testing.T) {
		assert.Greater(t, cfg.MediumMaxMinutes,
			cfg.QuickMaxMinutes,
			"MediumMaxMinutes must exceed QuickMaxMinutes",
		)
	})

	t.Run("LargeMaxHours_positive", func(t *testing.T) {
		assert.Greater(t, cfg.LargeMaxHours, 0,
			"LargeMaxHours must be positive",
		)
	})

	t.Run("MinQualityScore_range", func(t *testing.T) {
		assert.GreaterOrEqual(t, cfg.MinQualityScore, 0.0,
			"MinQualityScore must be >= 0.0",
		)
		assert.LessOrEqual(t, cfg.MinQualityScore, 1.0,
			"MinQualityScore must be <= 1.0",
		)
	})

	t.Run("AdaptiveCeremony_default_true", func(t *testing.T) {
		assert.True(t, cfg.AdaptiveCeremony,
			"AdaptiveCeremony should default to true",
		)
	})

	t.Run("MaxParallelAgents_positive", func(t *testing.T) {
		assert.Greater(t, cfg.MaxParallelAgents, 0,
			"MaxParallelAgents must be positive",
		)
	})

	t.Run("MaxParallelTasks_positive", func(t *testing.T) {
		assert.Greater(t, cfg.MaxParallelTasks, 0,
			"MaxParallelTasks must be positive",
		)
	})

	t.Run("CacheEnabled_default_true", func(t *testing.T) {
		assert.True(t, cfg.CacheEnabled,
			"CacheEnabled should default to true",
		)
	})

	t.Run("CacheDir_nonempty", func(t *testing.T) {
		assert.NotEmpty(t, cfg.CacheDir,
			"CacheDir must have a default value",
		)
	})

	t.Run("DebateRounds_positive", func(t *testing.T) {
		phases := []string{
			"constitution", "specify", "clarify",
			"plan", "tasks", "analyze", "implement",
		}
		for _, phase := range phases {
			rounds := cfg.GetPhaseRounds(phase)
			assert.Greater(t, rounds, 0,
				"GetPhaseRounds(%s) must be positive",
				phase,
			)
		}
	})

	t.Run("PhaseTimeouts_positive", func(t *testing.T) {
		phases := []string{
			"constitution", "specify", "clarify",
			"plan", "tasks", "analyze", "implement",
		}
		for _, phase := range phases {
			timeout := cfg.GetPhaseTimeout(phase)
			assert.Greater(t, timeout.Seconds(), 0.0,
				"GetPhaseTimeout(%s) must be positive",
				phase,
			)
		}
	})

	t.Run("CircuitBreakerThreshold_positive",
		func(t *testing.T) {
			assert.Greater(t, cfg.CircuitBreakerThreshold, 0,
				"CircuitBreakerThreshold must be positive",
			)
		},
	)
}

// --- 9. Well-Known Adapters ---

func TestAutomation_WellKnownAdapters(t *testing.T) {
	wka := adapters.WellKnownAdapters()

	t.Run("count", func(t *testing.T) {
		assert.Equal(t, 10, len(wka),
			"WellKnownAdapters must return exactly 10 "+
				"adapters",
		)
	})

	expectedNames := []string{
		"opencode", "crush", "claude", "cursor",
		"windsurf", "copilot", "gemini", "aider",
		"cline", "roo",
	}

	for _, name := range expectedNames {
		t.Run("adapter/"+name, func(t *testing.T) {
			adapter, ok := wka[name]
			require.True(t, ok,
				"adapter %s must be present", name,
			)
			assert.Equal(t, name, adapter.GetAgentType(),
				"adapter type must match key",
			)
			assert.True(t, adapter.SupportsStreaming(),
				"adapter %s must support streaming", name,
			)
		})
	}
}

// --- 10. Power Feature Independence ---

func TestAutomation_PowerFeatureIndependence(t *testing.T) {
	t.Run("parallel_execution", func(t *testing.T) {
		exec := parallel_execution.NewExecutor(4, 0)
		require.NotNil(t, exec)
		assert.Equal(t, 4, exec.MaxWorkers())
		assert.Greater(t, exec.Timeout().Seconds(), 0.0)
		assert.NotEmpty(t, exec.String())
	})

	t.Run("constitution_code", func(t *testing.T) {
		c := constitution_code.NewConstitution()
		require.NotNil(t, c)
		assert.Equal(t, 0, c.RuleCount())
		err := c.AddRule(constitution_code.Rule{
			ID:        "test-rule",
			Title:     "Test Rule",
			Mandatory: true,
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, c.RuleCount())
	})

	t.Run("nyquist_tdd", func(t *testing.T) {
		tr := nyquist_tdd.NewTracker()
		require.NotNil(t, tr)
		assert.Equal(t, 0.0, tr.Ratio())
		assert.False(t, tr.IsCompliant())
	})

	t.Run("debate_architecture", func(t *testing.T) {
		d := debate_architecture.NewDebate(
			"test-debate", "test topic", 3,
		)
		require.NotNil(t, d)
		assert.Equal(t, "test-debate", d.ID())
		assert.Equal(t, "test topic", d.Topic())
		assert.Equal(t, 0, d.RoundCount())
		assert.False(t, d.IsComplete())
	})

	t.Run("skill_learning", func(t *testing.T) {
		l := skill_learning.NewLearner()
		require.NotNil(t, l)
		assert.Equal(t, 0, l.SkillCount())
		err := l.RegisterSkill(skill_learning.Skill{
			ID:   "go",
			Name: "Go Programming",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, l.SkillCount())
	})

	t.Run("brownfield", func(t *testing.T) {
		a := brownfield.NewAnalyzer()
		require.NotNil(t, a)
		assert.Equal(t, 0, a.PatternCount())
		err := a.RecordPattern("singleton", 0.9)
		assert.NoError(t, err)
		assert.Equal(t, 1, a.PatternCount())
	})

	t.Run("predictive_spec", func(t *testing.T) {
		p := predictive_spec.NewPredictor()
		require.NotNil(t, p)
		assert.Equal(t, 0, p.HistoryCount())
		p.AddHistory("added auth module")
		assert.Equal(t, 1, p.HistoryCount())
		preds := p.Predict()
		assert.Equal(t, 1, len(preds))
	})

	t.Run("cross_project", func(t *testing.T) {
		te := cross_project.NewTransferEngine()
		require.NotNil(t, te)
		assert.Equal(t, 0, te.Count())
		err := te.Add(cross_project.Knowledge{
			ID:      "k1",
			Project: "proj-a",
			Content: "use factory pattern",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, te.Count())
	})

	t.Run("adaptive_ceremony", func(t *testing.T) {
		adj := adaptive_ceremony.NewAdjuster(0.9, 0.4)
		require.NotNil(t, adj)
		up, down := adj.Thresholds()
		assert.Equal(t, 0.9, up)
		assert.Equal(t, 0.4, down)
		assert.Equal(t, 0, adj.AdjustmentCount())
		assert.NotEmpty(t, adj.String())
	})

	t.Run("spec_memory", func(t *testing.T) {
		idx := spec_memory.NewIndex()
		require.NotNil(t, idx)
		assert.Equal(t, 0, idx.Count())
		err := idx.Add(spec_memory.Entry{
			ID:    "e1",
			Title: "Test Entry",
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, idx.Count())
	})
}

// --- 11. Metrics Snapshot ---

func TestAutomation_MetricsSnapshot(t *testing.T) {
	m := metrics.NewMetrics()
	require.NotNil(t, m)

	// Record a flow start
	m.RecordFlowStart(types.EffortLarge)

	// Record a phase execution
	m.RecordPhase(&types.PhaseResult{
		Phase:        types.PhaseSpecify,
		Success:      true,
		QualityScore: 0.8,
	})

	// Record flow completion
	m.RecordFlowComplete(&types.FlowResult{
		Success:       true,
		CeremonyLevel: types.CeremonyStandard,
	})

	snap := m.Snapshot()

	t.Run("FlowsStarted", func(t *testing.T) {
		assert.Equal(t, int64(1), snap.FlowsStarted)
	})

	t.Run("FlowsCompleted", func(t *testing.T) {
		assert.Equal(t, int64(1), snap.FlowsCompleted)
	})

	t.Run("FlowsFailed", func(t *testing.T) {
		assert.Equal(t, int64(0), snap.FlowsFailed)
	})

	t.Run("PhasesExecuted", func(t *testing.T) {
		assert.Equal(t, int64(1), snap.PhasesExecuted)
	})

	t.Run("PhasesFailed", func(t *testing.T) {
		assert.Equal(t, int64(0), snap.PhasesFailed)
	})

	t.Run("EffortCounts_populated", func(t *testing.T) {
		assert.NotNil(t, snap.EffortCounts)
		assert.Equal(t, int64(1),
			snap.EffortCounts[types.EffortLarge],
		)
	})

	t.Run("CeremonyCounts_populated", func(t *testing.T) {
		assert.NotNil(t, snap.CeremonyCounts)
		assert.Equal(t, int64(1),
			snap.CeremonyCounts[types.CeremonyStandard],
		)
	})

	t.Run("PhaseMetrics_populated", func(t *testing.T) {
		assert.NotNil(t, snap.PhaseMetrics)
		pm, ok := snap.PhaseMetrics[types.PhaseSpecify]
		require.True(t, ok,
			"PhaseMetrics must have an entry for Specify",
		)
		assert.Equal(t, int64(1), pm.Executions)
		assert.Equal(t, int64(1), pm.Successes)
		assert.Equal(t, 0.8, pm.AvgScore)
	})

	t.Run("SuccessRate", func(t *testing.T) {
		rate := m.GetFlowSuccessRate()
		assert.Equal(t, 1.0, rate)
	})

	t.Run("Snapshot_is_copy", func(t *testing.T) {
		// Mutating the snapshot must not affect the
		// original tracker.
		snap.FlowsStarted = 999
		assert.NotEqual(t, int64(999), m.Snapshot().FlowsStarted,
			"snapshot must be an independent copy",
		)
	})
}

// --- 12. Classifier Signals ---

func TestAutomation_ClassifierSignals(t *testing.T) {
	c := intent.NewClassifier()
	require.NotNil(t, c)

	cases := []struct {
		name     string
		input    string
		expected types.EffortLevel
	}{
		{
			name:     "quick_fix_typo",
			input:    "fix typo in README",
			expected: types.EffortQuick,
		},
		{
			name:     "medium_add_validation",
			input:    "add new validation",
			expected: types.EffortMedium,
		},
		{
			name:     "large_rest_api_with_auth",
			input:    "implement complete REST API with authentication",
			expected: types.EffortLarge,
		},
		{
			name:     "epic_rewrite_architecture",
			input:    "rewrite the entire system architecture",
			expected: types.EffortEpic,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := c.Classify(tc.input)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, result.Level,
				"input %q should classify as %s",
				tc.input, tc.expected,
			)
			assert.Greater(t, result.Confidence, 0.0,
				"confidence must be positive",
			)
			assert.NotEmpty(t, result.Reasoning,
				"reasoning must not be empty",
			)
			assert.Equal(t,
				types.CeremonyForEffort(tc.expected),
				result.CeremonyLevel,
				"CeremonyLevel must match effort mapping",
			)
		})
	}

	t.Run("debate_required_for_non_quick", func(t *testing.T) {
		for _, tc := range cases {
			result := c.Classify(tc.input)
			if tc.expected == types.EffortQuick {
				assert.False(t, result.RequiresDebate,
					"quick should not require debate",
				)
			} else {
				assert.True(t, result.RequiresDebate,
					"%s should require debate",
					tc.expected,
				)
			}
		}
	})

	t.Run("speckit_required_for_large_and_epic",
		func(t *testing.T) {
			for _, tc := range cases {
				result := c.Classify(tc.input)
				if tc.expected == types.EffortLarge ||
					tc.expected == types.EffortEpic {
					assert.True(t, result.RequiresSpecKit,
						"%s should require SpecKit",
						tc.expected,
					)
				} else {
					assert.False(t, result.RequiresSpecKit,
						"%s should not require SpecKit",
						tc.expected,
					)
				}
			}
		},
	)
}

// --- 13. No Exported Globals ---

func TestAutomation_NoExportedGlobals(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	pkgRoot := filepath.Join(root, "pkg")
	fset := token.NewFileSet()
	violations := []string{}

	err := filepath.Walk(pkgRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() ||
				!strings.HasSuffix(path, ".go") ||
				strings.HasSuffix(path, "_test.go") {
				return nil
			}

			file, parseErr := parser.ParseFile(
				fset, path, nil, parser.AllErrors,
			)
			if parseErr != nil {
				return nil
			}

			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.VAR {
					continue
				}
				for _, spec := range genDecl.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range vs.Names {
						if !ast.IsExported(name.Name) {
							continue
						}
						// Idiomatic Go permits exported sentinel
						// error values (Err*). They are a
						// CONST-035 anti-bluff requirement for
						// surfacing missing-dependency states
						// (round-28 §11.4 audit: e.g.
						// ErrDebateFuncNotConfigured replaces
						// fabricated placeholder output).
						if strings.HasPrefix(name.Name, "Err") {
							continue
						}
						rel, _ := filepath.Rel(
							root, path,
						)
						violations = append(
							violations,
							rel+": var "+name.Name,
						)
					}
				}
			}
			return nil
		},
	)
	require.NoError(t, err, "walking pkg/ should not fail")
	assert.Empty(t, violations,
		"exported global variables found in pkg/ "+
			"(use constructors instead): %v",
		violations,
	)
}

// --- 14. Makefile Targets ---

func TestAutomation_MakefileTargets(t *testing.T) {
	root := projectRoot()
	require.NotEmpty(t, root, "could not locate project root")

	data, err := os.ReadFile(
		filepath.Join(root, "Makefile"),
	)
	require.NoError(t, err, "Makefile should be readable")
	content := string(data)

	targets := []string{
		"build",
		"test",
		"test-race",
		"test-cover",
		"test-bench",
		"fmt",
		"vet",
		"lint",
		"clean",
	}

	for _, target := range targets {
		t.Run("target/"+target, func(t *testing.T) {
			// Match "target:" at the start of a line
			pattern := "\n" + target + ":"
			found := strings.Contains(content, pattern)
			if !found {
				// Also check if target is on the first line
				found = strings.HasPrefix(
					content, target+":",
				)
			}
			assert.True(t, found,
				"Makefile must contain target %q", target,
			)
		})
	}
}
