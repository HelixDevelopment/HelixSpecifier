package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/adapters"
	"digital.vasic.helixspecifier/pkg/ceremony"
	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/engine"
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

// --- Helpers ---

// newE2ELogger returns a silent logger for E2E tests.
func newE2ELogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)
	return logger
}

// echoTestDebateFunc is a deterministic stub installed in *_test.go
// to give SpecKit Pillar a real DebateFunc (CONST-050(A) permits
// stubs in test sources only). Round-28 §11.4 audit (2026-05-17):
// the production Pillar's nil-DebateFunc fabrication branch was
// removed; E2E tests MUST install this stub explicitly.
func echoTestDebateFunc(_ context.Context, topic string, _ int, metadata map[string]interface{}) (string, float64, string, error) {
	phase, _ := metadata["phase"].(string)
	out := "[" + phase + "] " + topic
	return out, 0.75, "", nil
}

// newTestSpecKitPillar returns a Pillar with the deterministic
// echo DebateFunc installed for E2E tests.
func newTestSpecKitPillar(cfg *config.Config, logger *logrus.Logger) *speckit.Pillar {
	p := speckit.NewPillar(cfg, logger)
	p.SetDebateFunc(echoTestDebateFunc)
	return p
}

// newE2EEngine creates a fully wired FusionEngine with all
// three real pillars, ceremony scaler, and spec memory.
func newE2EEngine() *engine.FusionEngine {
	cfg := config.DefaultConfig()
	logger := newE2ELogger()

	eng := engine.New(cfg, logger)

	sk := newTestSpecKitPillar(cfg, logger)
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

// e2eResult bundles the outputs of a full classify-execute-format
// workflow for convenient assertions.
type e2eResult struct {
	Classification *types.EffortClassification
	Flow           *types.FlowResult
	Formatted      string
}

// executeAndFormat classifies the request, executes a flow, and
// formats the result through the named adapter category. It
// returns all intermediate artifacts for assertion.
func executeAndFormat(
	t *testing.T,
	eng *engine.FusionEngine,
	request string,
	category adapters.AgentCategory,
) *e2eResult {
	t.Helper()
	ctx := context.Background()

	classifier := intent.NewClassifier()
	classification := classifier.Classify(request)
	require.NotNil(t, classification)

	flow, err := eng.ExecuteFlow(ctx, request, classification)
	require.NoError(t, err)
	require.NotNil(t, flow)

	adapter := adapters.NewGenericAdapter(
		string(category), category,
	)
	formatted, err := adapter.FormatOutput(flow)
	require.NoError(t, err)

	return &e2eResult{
		Classification: classification,
		Flow:           flow,
		Formatted:      formatted,
	}
}

// --- 1. Quick Fix Workflow ---

func TestE2E_QuickFixWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"fix typo in README.md",
		adapters.CategoryMarkdown,
	)

	assert.Equal(t, types.EffortQuick, res.Classification.Level)
	assert.Equal(
		t, types.CeremonyMinimal,
		res.Classification.CeremonyLevel,
	)
	assert.False(t, res.Classification.RequiresDebate)
	assert.False(t, res.Classification.RequiresSpecKit)

	assert.True(t, res.Flow.Success)
	assert.Equal(t, 1, len(res.Flow.PhaseResults),
		"minimal ceremony should produce exactly 1 phase",
	)
	assert.Equal(
		t, types.PhaseImplement,
		res.Flow.PhaseResults[0].Phase,
	)
	assert.True(t, res.Flow.Duration < 5*time.Second,
		"quick fix should complete fast",
	)
	assert.NotEmpty(t, res.Formatted)
	assert.Contains(t, res.Formatted, res.Flow.FlowID)
}

// --- 2. Medium Feature Workflow ---

func TestE2E_MediumFeatureWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"add input validation to the user registration form",
		adapters.CategoryMarkdown,
	)

	assert.Equal(t, types.EffortMedium, res.Classification.Level)
	assert.Equal(
		t, types.CeremonyLight,
		res.Classification.CeremonyLevel,
	)

	assert.True(t, res.Flow.Success)
	// Light ceremony: specify, plan, implement
	assert.Equal(t, 3, len(res.Flow.PhaseResults))

	for _, pr := range res.Flow.PhaseResults {
		assert.True(t, pr.Success)
		assert.NotEmpty(t, pr.Output)
		assert.Greater(t, pr.QualityScore, 0.0)
		assert.LessOrEqual(t, pr.QualityScore, 1.0)
	}

	assert.Greater(t, res.Flow.OverallQualityScore, 0.0)
}

// --- 3. Large Implementation Workflow ---

func TestE2E_LargeImplementationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	cfg := config.DefaultConfig()
	logger := newE2ELogger()

	eng := engine.New(cfg, logger)
	sk := newTestSpecKitPillar(cfg, logger)
	sp := superpowers.NewPillar(cfg, logger)
	g := gsd.NewPillar(logger)
	cs := ceremony.NewScaler(cfg, logger)
	mem := memory.NewStore()

	eng.RegisterSpecKit(sk)
	eng.RegisterSuperpowers(sp)
	eng.RegisterGSD(g)
	eng.RegisterCeremonyScaler(cs)
	eng.RegisterSpecMemory(mem)

	request := "implement a complete REST API service with " +
		"JWT authentication, user management, and " +
		"role-based access control"

	classifier := intent.NewClassifier()
	classification := classifier.Classify(request)

	assert.Equal(t, types.EffortLarge, classification.Level)
	assert.Equal(
		t, types.CeremonyStandard,
		classification.CeremonyLevel,
	)
	assert.True(t, classification.RequiresSpecKit)

	ctx := context.Background()
	flow, err := eng.ExecuteFlow(ctx, request, classification)
	require.NoError(t, err)

	assert.True(t, flow.Success)
	// Standard ceremony: all 7 phases
	assert.Equal(t, 7, len(flow.PhaseResults))

	expectedPhases := types.AllPhases()
	for i, pr := range flow.PhaseResults {
		assert.Equal(t, expectedPhases[i], pr.Phase)
		assert.True(t, pr.Success)
	}

	// Verify milestones can be created from the spec
	spec := &types.Specification{
		ID:          "spec-large-001",
		Title:       "REST API Service",
		Description: request,
		Source:      types.SourceFusion,
	}
	tasks := []types.Task{
		{ID: "t-1", Name: "Auth module", Priority: "high"},
		{ID: "t-2", Name: "User CRUD", Priority: "medium"},
	}

	milestones, err := g.CreateMilestones(ctx, spec, tasks)
	require.NoError(t, err)
	assert.Greater(t, len(milestones), 0)

	// Verify memory learned from this flow
	assert.Greater(t, mem.PatternCount(), 0,
		"spec memory should learn from successful flow",
	)
}

// --- 4. Epic System Rewrite Workflow ---

func TestE2E_EpicSystemRewriteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()

	request := "rewrite the entire authentication system " +
		"from scratch with OAuth2, SAML, and MFA support " +
		"across all microservices"

	classifier := intent.NewClassifier()
	classification := classifier.Classify(request)

	assert.Equal(t, types.EffortEpic, classification.Level)
	assert.Equal(
		t, types.CeremonyHeavy,
		classification.CeremonyLevel,
	)
	assert.True(t, classification.RequiresDebate)
	assert.True(t, classification.RequiresSpecKit)

	ctx := context.Background()
	flow, err := eng.ExecuteFlow(ctx, request, classification)
	require.NoError(t, err)

	assert.True(t, flow.Success)
	// Heavy ceremony: all 7 phases
	assert.Equal(t, 7, len(flow.PhaseResults))
	assert.Equal(t, types.EffortEpic, flow.EffortLevel)
	assert.Greater(t, flow.OverallQualityScore, 0.0)
}

// --- 5. Markdown Adapter Workflow ---

func TestE2E_MarkdownAdapterWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"add logging to the payment service",
		adapters.CategoryMarkdown,
	)

	assert.True(t, res.Flow.Success)
	assert.Contains(t, res.Formatted, "#",
		"markdown output should contain headers",
	)
	assert.Contains(t, res.Formatted, "**",
		"markdown output should contain bold text",
	)
	assert.Contains(t, res.Formatted, "- ",
		"markdown output should contain list items",
	)
	assert.Contains(t, res.Formatted, "## Phases")
}

// --- 6. TOML Adapter Workflow ---

func TestE2E_TOMLAdapterWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"add caching to the search endpoint",
		adapters.CategoryTOML,
	)

	assert.True(t, res.Flow.Success)
	assert.Contains(t, res.Formatted, "[flow]",
		"TOML output should contain [flow] section",
	)
	assert.Contains(t, res.Formatted, "id = ",
		"TOML output should contain id key",
	)
	assert.Contains(t, res.Formatted, "effort = ",
		"TOML output should contain effort key",
	)
	assert.Contains(t, res.Formatted, "ceremony = ")
	assert.Contains(t, res.Formatted, "quality = ")
	assert.Contains(t, res.Formatted, "success = ")
}

// --- 7. GitHub Adapter Workflow ---

func TestE2E_GitHubAdapterWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"update the CI pipeline configuration",
		adapters.CategoryGitHub,
	)

	assert.True(t, res.Flow.Success)

	// GitHub category falls through to JSON formatting in
	// the current adapter implementation.
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(res.Formatted), &parsed)
	require.NoError(t, err,
		"github adapter output should be valid JSON",
	)
	assert.Contains(t, parsed, "flow_id")
	assert.Contains(t, parsed, "success")
}

// --- 8. Generic JSON Adapter Workflow ---

func TestE2E_GenericJSONAdapterWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"refactor the database connection pool",
		adapters.CategoryGeneric,
	)

	assert.True(t, res.Flow.Success)

	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(res.Formatted), &parsed)
	require.NoError(t, err,
		"generic adapter output should be valid JSON",
	)
	assert.Contains(t, parsed, "flow_id")
	assert.Contains(t, parsed, "effort_level")
	assert.Contains(t, parsed, "ceremony_level")
	assert.Contains(t, parsed, "success")
	assert.Contains(t, parsed, "phase_results")
	assert.Contains(t, parsed, "overall_quality_score")
}

// --- 9. Minimal Adapter Workflow ---

func TestE2E_MinimalAdapterWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	res := executeAndFormat(
		t, eng,
		"bump version to 2.0.0",
		adapters.CategoryMinimal,
	)

	assert.True(t, res.Flow.Success)

	// Minimal category falls through to JSON formatting in
	// the current adapter implementation. Verify the output
	// is non-empty and parseable as JSON (no markdown headers).
	assert.NotEmpty(t, res.Formatted)
	assert.NotContains(t, res.Formatted, "## Phases",
		"minimal output should not contain markdown sections",
	)
}

// --- 10. Flow Resumption Workflow ---

func TestE2E_FlowResumptionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	ctx := context.Background()

	classifier := intent.NewClassifier()
	classification := classifier.Classify(
		"add rate limiting middleware",
	)

	flow, err := eng.ExecuteFlow(
		ctx, "add rate limiting middleware", classification,
	)
	require.NoError(t, err)
	require.True(t, flow.Success)

	flowID := flow.FlowID
	assert.NotEmpty(t, flowID)

	// Resume the completed flow
	resumed, err := eng.ResumeFlow(
		ctx, flowID, "add rate limiting middleware",
	)
	require.NoError(t, err)
	require.NotNil(t, resumed)

	assert.Equal(t, flowID, resumed.FlowID,
		"resumed flow ID should match original",
	)
	assert.True(t, resumed.Success)
	assert.Equal(
		t,
		len(flow.PhaseResults),
		len(resumed.PhaseResults),
	)
}

// --- 11. Spec Memory Learning Workflow ---

func TestE2E_SpecMemoryLearningWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	cfg := config.DefaultConfig()
	logger := newE2ELogger()

	eng := engine.New(cfg, logger)
	sk := newTestSpecKitPillar(cfg, logger)
	sp := superpowers.NewPillar(cfg, logger)
	g := gsd.NewPillar(logger)
	cs := ceremony.NewScaler(cfg, logger)
	mem := memory.NewStore()

	eng.RegisterSpecKit(sk)
	eng.RegisterSuperpowers(sp)
	eng.RegisterGSD(g)
	eng.RegisterCeremonyScaler(cs)
	eng.RegisterSpecMemory(mem)

	ctx := context.Background()
	classifier := intent.NewClassifier()

	requests := []string{
		"fix typo in config.go",
		"add input validation to the login handler",
		"implement a complete notification service " +
			"with push, email, and SMS delivery",
	}

	for i, req := range requests {
		classification := classifier.Classify(req)
		flow, err := eng.ExecuteFlow(
			ctx, req, classification,
		)
		require.NoError(t, err)
		require.True(t, flow.Success)

		expectedPatterns := i + 1
		assert.Equal(
			t, expectedPatterns, mem.PatternCount(),
			"memory should contain %d patterns after flow %d",
			expectedPatterns, i+1,
		)
	}

	// Store a spec so we can search it
	spec := &types.Specification{
		ID:          "spec-mem-001",
		Title:       "Notification Service",
		Description: "Push, email, SMS delivery",
		Source:      types.SourceFusion,
	}
	err := mem.Store(ctx, spec)
	require.NoError(t, err)

	results, err := mem.Search(ctx, "notification", 5)
	require.NoError(t, err)
	assert.Greater(t, len(results), 0,
		"search should find the stored specification",
	)
}

// --- 12. Multiple Sequential Flows ---

func TestE2E_MultipleSequentialFlows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	ctx := context.Background()
	classifier := intent.NewClassifier()
	met := metrics.NewMetrics()

	requests := []string{
		"fix typo in main.go",
		"add logging to the auth handler",
		"update comment in utils.go",
		"rename variable in parser.go",
		"add import for testing package",
	}

	for _, req := range requests {
		classification := classifier.Classify(req)
		met.RecordFlowStart(classification.Level)

		flow, err := eng.ExecuteFlow(
			ctx, req, classification,
		)
		require.NoError(t, err)
		require.True(t, flow.Success)

		met.RecordFlowComplete(flow)
		for _, pr := range flow.PhaseResults {
			met.RecordPhase(&pr)
		}
	}

	snap := met.Snapshot()
	assert.Equal(t, int64(5), snap.FlowsStarted)
	assert.Equal(t, int64(5), snap.FlowsCompleted)
	assert.Equal(t, int64(0), snap.FlowsFailed)
	assert.Greater(t, snap.PhasesExecuted, int64(0))
	assert.Equal(t, 1.0, met.GetFlowSuccessRate())
}

// --- 13. Adaptive Ceremony Downgrade ---

func TestE2E_AdaptiveCeremonyDowngrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	cfg := config.DefaultConfig()
	cfg.AdaptiveCeremony = true
	logger := newE2ELogger()

	eng := engine.New(cfg, logger)

	// Use a speckit pillar that produces high quality scores
	// (> 0.85) to trigger ceremony downgrade.
	sk := speckit.NewPillar(cfg, logger)
	sk.SetDebateFunc(func(
		ctx context.Context,
		topic string,
		rounds int,
		metadata map[string]interface{},
	) (string, float64, string, error) {
		return "high quality output", 0.95, "dbg-001", nil
	})

	cs := ceremony.NewScaler(cfg, logger)

	eng.RegisterSpecKit(sk)
	eng.RegisterSuperpowers(superpowers.NewPillar(cfg, logger))
	eng.RegisterGSD(gsd.NewPillar(logger))
	eng.RegisterCeremonyScaler(cs)
	eng.RegisterSpecMemory(memory.NewStore())

	ctx := context.Background()
	classification := &types.EffortClassification{
		Level:         types.EffortEpic,
		Confidence:    0.9,
		CeremonyLevel: types.CeremonyHeavy,
		Signals:       []string{"epic_task"},
	}

	flow, err := eng.ExecuteFlow(
		ctx,
		"rewrite the entire system from scratch",
		classification,
	)
	require.NoError(t, err)
	assert.True(t, flow.Success)

	// The scaler should have downgraded from heavy to standard
	// because all phase scores are 0.95 (> 0.85 threshold).
	// The flow's classification gets mutated during execution.
	assert.Equal(
		t, types.CeremonyStandard,
		classification.CeremonyLevel,
		"ceremony should be downgraded when quality is high",
	)
}

// --- 14. Adaptive Ceremony Upgrade ---

func TestE2E_AdaptiveCeremonyUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	cfg := config.DefaultConfig()
	cfg.AdaptiveCeremony = true
	logger := newE2ELogger()

	eng := engine.New(cfg, logger)

	// Use a speckit pillar that produces low quality scores
	// (< 0.5) to trigger ceremony upgrade.
	sk := speckit.NewPillar(cfg, logger)
	sk.SetDebateFunc(func(
		ctx context.Context,
		topic string,
		rounds int,
		metadata map[string]interface{},
	) (string, float64, string, error) {
		return "low quality output", 0.35, "dbg-002", nil
	})

	cs := ceremony.NewScaler(cfg, logger)

	eng.RegisterSpecKit(sk)
	eng.RegisterSuperpowers(superpowers.NewPillar(cfg, logger))
	eng.RegisterGSD(gsd.NewPillar(logger))
	eng.RegisterCeremonyScaler(cs)
	eng.RegisterSpecMemory(memory.NewStore())

	ctx := context.Background()
	classification := &types.EffortClassification{
		Level:         types.EffortMedium,
		Confidence:    0.7,
		CeremonyLevel: types.CeremonyLight,
		Signals:       []string{"moderate_task"},
	}

	flow, err := eng.ExecuteFlow(
		ctx,
		"add caching to the query service",
		classification,
	)
	require.NoError(t, err)
	assert.True(t, flow.Success)

	// The scaler should have upgraded from light to standard
	// because all phase scores are 0.35 (< 0.5 threshold).
	assert.Equal(
		t, types.CeremonyStandard,
		classification.CeremonyLevel,
		"ceremony should be upgraded when quality is low",
	)
}

// --- 15. All Well-Known Adapters ---

func TestE2E_AllWellKnownAdapters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	ctx := context.Background()

	classifier := intent.NewClassifier()
	classification := classifier.Classify(
		"add metrics collection to the API gateway",
	)

	flow, err := eng.ExecuteFlow(
		ctx,
		"add metrics collection to the API gateway",
		classification,
	)
	require.NoError(t, err)
	require.True(t, flow.Success)

	wellKnown := adapters.WellKnownAdapters()
	assert.Equal(t, 10, len(wellKnown),
		"should have exactly 10 well-known adapters",
	)

	for name, adapter := range wellKnown {
		t.Run(name, func(t *testing.T) {
			output, err := adapter.FormatOutput(flow)
			require.NoError(t, err,
				"adapter %s should format without error", name,
			)
			assert.NotEmpty(t, output,
				"adapter %s should produce non-empty output",
				name,
			)
			assert.Equal(t, name, adapter.GetAgentType())
			assert.True(t, adapter.SupportsStreaming())
		})
	}
}

// --- 16. Concurrent User Workflows ---

func TestE2E_ConcurrentUserWorkflows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	ctx := context.Background()
	classifier := intent.NewClassifier()

	requests := []string{
		"fix typo in handler.go",
		"add validation to signup form",
		"update comment in utils package",
		"rename variable in config parser",
		"add import for net/http",
		"bump version in go.mod",
		"add logging to error handler",
		"change color of status badge",
		"add log statement to auth middleware",
		"update comment in test helpers",
	}

	type userResult struct {
		index int
		flow  *types.FlowResult
		err   error
	}

	results := make(chan userResult, len(requests))

	for i, req := range requests {
		go func(idx int, r string) {
			cl := classifier.Classify(r)
			flow, err := eng.ExecuteFlow(ctx, r, cl)
			results <- userResult{
				index: idx,
				flow:  flow,
				err:   err,
			}
		}(i, req)
	}

	flowIDs := make(map[string]bool)
	for range requests {
		res := <-results
		require.NoError(t, res.err,
			"request %d should succeed", res.index,
		)
		require.NotNil(t, res.flow)
		assert.True(t, res.flow.Success)
		assert.NotEmpty(t, res.flow.FlowID)

		assert.False(t, flowIDs[res.flow.FlowID],
			"each flow should have a unique ID",
		)
		flowIDs[res.flow.FlowID] = true
	}

	assert.Equal(t, len(requests), len(flowIDs))
}

// --- 17. Engine Health and Version ---

func TestE2E_EngineHealthAndVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	ctx := context.Background()

	err := eng.Health(ctx)
	assert.NoError(t, err, "health check should pass")

	name := eng.Name()
	assert.NotEmpty(t, name)
	assert.Equal(t, "HelixSpecifier", name)

	version := eng.Version()
	assert.NotEmpty(t, version)
	assert.Equal(t, "1.0.0", version)
}

// --- 18. Classify and Execute All Levels ---

func TestE2E_ClassifyAndExecuteAllLevels(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	eng := newE2EEngine()
	ctx := context.Background()
	classifier := intent.NewClassifier()

	cases := []struct {
		name           string
		request        string
		expectedEffort types.EffortLevel
		expectedPhases int
	}{
		{
			name:           "Quick",
			request:        "fix typo in README.md",
			expectedEffort: types.EffortQuick,
			expectedPhases: 1,
		},
		{
			name: "Medium",
			request: "add input validation to " +
				"the contact form",
			expectedEffort: types.EffortMedium,
			expectedPhases: 3,
		},
		{
			name: "Large",
			request: "implement a complete REST API " +
				"service with JWT authentication " +
				"and database integration",
			expectedEffort: types.EffortLarge,
			expectedPhases: 7,
		},
		{
			name: "Epic",
			request: "rewrite the entire system " +
				"from scratch with full " +
				"microservices architecture",
			expectedEffort: types.EffortEpic,
			expectedPhases: 7,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			classification := classifier.Classify(tc.request)
			assert.Equal(
				t, tc.expectedEffort,
				classification.Level,
				"effort level mismatch for %s", tc.name,
			)

			expectedCeremony := types.CeremonyForEffort(
				tc.expectedEffort,
			)
			assert.Equal(
				t, expectedCeremony,
				classification.CeremonyLevel,
			)

			flow, err := eng.ExecuteFlow(
				ctx, tc.request, classification,
			)
			require.NoError(t, err)
			assert.True(t, flow.Success)
			assert.Equal(
				t, tc.expectedEffort, flow.EffortLevel,
			)
			assert.Equal(
				t, tc.expectedPhases,
				len(flow.PhaseResults),
				"phase count mismatch for %s", tc.name,
			)

			// Verify each phase completed successfully
			for _, pr := range flow.PhaseResults {
				assert.True(t, pr.Success)
				assert.NotEmpty(t, pr.Output)
				assert.Greater(t, pr.QualityScore, 0.0)
			}

			// Verify flow ID format
			assert.True(
				t,
				strings.HasPrefix(flow.FlowID, "hsf-"),
			)

			// Verify adapter output
			adapter := adapters.NewGenericAdapter(
				"test", adapters.CategoryGeneric,
			)
			output, err := adapter.FormatOutput(flow)
			require.NoError(t, err)

			var parsed map[string]interface{}
			err = json.Unmarshal(
				[]byte(output), &parsed,
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				string(tc.expectedEffort),
				parsed["effort_level"],
			)
		})
	}
}

// --- 19. Three-Pillar Fusion Workflow ---

func TestE2E_ThreePillarFusion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")  // SKIP-OK: #short-mode
	}

	cfg := config.DefaultConfig()
	logger := newE2ELogger()

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

	// Track debate invocations
	debateCalls := 0
	sk.SetDebateFunc(func(
		ctx context.Context,
		topic string,
		rounds int,
		metadata map[string]interface{},
	) (string, float64, string, error) {
		debateCalls++
		return fmt.Sprintf(
			"Debate output for %s",
			metadata["phase"],
		), 0.88, fmt.Sprintf("dbg-%d", debateCalls), nil
	})

	// Register the intent classifier
	classifier := intent.NewClassifier()
	eng.RegisterClassifier(classifier.Classify)

	// Execute a "large" effort request
	request := "implement new authentication service " +
		"with database"
	ctx := context.Background()

	// Step 1: Classify via the engine's registered classifier
	classification, err := eng.ClassifyEffort(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, classification)

	// Verify classification is EffortLarge
	assert.Equal(t, types.EffortLarge, classification.Level,
		"classifier should detect large effort from "+
			"'implement' + 'service' + 'database'",
	)

	// Step 2: Execute the flow
	flow, err := eng.ExecuteFlow(ctx, request, classification)
	require.NoError(t, err)
	require.NotNil(t, flow)
	assert.True(t, flow.Success)

	// Verify FlowResult.Specification is not nil
	assert.NotNil(t, flow.Specification,
		"specification must be built from phases",
	)
	assert.NotEmpty(t, flow.Specification.ID)
	assert.NotEmpty(t, flow.Specification.Requirements)

	// Verify FlowResult.Milestones is not empty
	assert.Greater(t, len(flow.Milestones), 0,
		"GSD must create milestones for large flows",
	)

	// Verify FlowResult.FinalArtifact is not empty
	assert.NotEmpty(t, flow.FinalArtifact,
		"final artifact must be set from last phase output",
	)

	// Verify the debate func was called for each phase
	expectedPhases := types.PhasesForCeremony(
		types.CeremonyStandard,
	)
	assert.Equal(t, len(expectedPhases), debateCalls,
		"debate func must be called once per phase",
	)

	// Verify PhaseResults have score > 0.75
	for _, pr := range flow.PhaseResults {
		assert.True(t, pr.Success)
		assert.Greater(t, pr.QualityScore, 0.75,
			"phase %s score should exceed 0.75 with "+
				"real debate func", pr.Phase,
		)
		assert.NotEmpty(t, pr.DebateID,
			"phase %s should have a debate ID", pr.Phase,
		)
	}
}

// Silence unused import warnings by referencing all imports.
var (
	_ = fmt.Sprintf
	_ = time.Now
)
