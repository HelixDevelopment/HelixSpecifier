package superpowers

import (
	"context"
	"testing"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingTranslator captures every key passed to T. Lives in
// _test.go only (CONST-050(A)) — never imported by production.
type recordingTranslator struct {
	calls []string
}

func (r *recordingTranslator) T(key string, args ...any) string {
	r.calls = append(r.calls, key)
	_ = args
	return "stub:" + key
}

// TestPillar_ExecuteSingleTask_OutputUsesI18nKey asserts the
// round-208 migration: TaskResult.Output is produced via the i18n
// translator, not the hardcoded "Executed task: %s" format. A
// regression that reinstates the literal would fail this test —
// paired-mutation guard for CONST-046.
func TestPillar_ExecuteSingleTask_OutputUsesI18nKey(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	rec := &recordingTranslator{}
	p := NewPillarWithTranslator(cfg, logger, rec)
	require.NotNil(t, p)

	task := types.Task{
		ID:   "task-i18n-1",
		Name: "demo-task-name",
	}
	res := p.executeSingleTask(context.Background(), task)

	assert.Equal(
		t,
		"stub:helixspecifier_superpowers_task_executed",
		res.Output,
	)
	assert.Contains(
		t, rec.calls,
		"helixspecifier_superpowers_task_executed",
	)
	assert.NotContains(
		t, res.Output, "Executed task:",
		"regression: hardcoded literal reintroduced",
	)
}

// TestPillar_NewPillar_DefaultNoopVisibleKey verifies the legacy
// NewPillar constructor stays operational and the noop fallback
// produces a visible key + task-name in the output, so missing-
// bundle situations are visibly debuggable.
func TestPillar_NewPillar_DefaultNoopVisibleKey(t *testing.T) {
	p := newTestPillar()
	task := types.Task{
		ID:   "task-noop",
		Name: "noop-demo",
	}
	res := p.executeSingleTask(context.Background(), task)
	assert.Contains(
		t, res.Output,
		"helixspecifier_superpowers_task_executed",
	)
	// NoopTranslator passes args via Sprintf — task name should
	// appear somewhere in the output.
	assert.Contains(t, res.Output, "noop-demo")
}

// TestPillar_NewPillarWithTranslator_NilFallback proves the
// constructor accepts nil and falls back to NoopTranslator.
func TestPillar_NewPillarWithTranslator_NilFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	p := NewPillarWithTranslator(cfg, logger, nil)
	require.NotNil(t, p)
	require.NotNil(t, p.translator)
}

// TestPillar_SetTranslator_Swap exercises runtime swap including
// nil-fallback. Confirms no panic and subsequent task executions
// use the new backend.
func TestPillar_SetTranslator_Swap(t *testing.T) {
	p := newTestPillar()
	task := types.Task{ID: "x", Name: "y"}

	r1 := p.executeSingleTask(context.Background(), task)
	assert.Contains(
		t, r1.Output,
		"helixspecifier_superpowers_task_executed",
	)

	rec := &recordingTranslator{}
	p.SetTranslator(rec)
	r2 := p.executeSingleTask(context.Background(), task)
	assert.Equal(
		t,
		"stub:helixspecifier_superpowers_task_executed",
		r2.Output,
	)

	p.SetTranslator(nil)
	r3 := p.executeSingleTask(context.Background(), task)
	assert.NotContains(t, r3.Output, "stub:")
}

// Sentinel that the i18n package is referenced.
func TestSuperpowers_I18nPackagePresent(t *testing.T) {
	tr := i18n.NewNoopTranslator()
	assert.NotNil(t, tr)
}

// TestPillar_Round405_ValidationErrorsUseI18nKeys is the
// paired-mutation guard for the round-405 CONST-046 migration of
// the Superpowers Pillar validation errors. Each error path is
// exercised against a recording translator confirming the i18n
// key is invoked and the old hardcoded literal is gone.
func TestPillar_Round405_ValidationErrorsUseI18nKeys(
	t *testing.T,
) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	ctx := context.Background()

	// ExecuteWithTDD: no tasks provided.
	rec := &recordingTranslator{}
	p := NewPillarWithTranslator(cfg, logger, rec)
	_, err := p.ExecuteWithTDD(ctx, nil)
	require.Error(t, err)
	assert.Contains(
		t, rec.calls,
		"helixspecifier_superpowers_err_no_tasks",
	)
	assert.Equal(
		t,
		"stub:helixspecifier_superpowers_err_no_tasks",
		err.Error(),
	)
	assert.NotContains(t, err.Error(), "no tasks provided",
		"regression: hardcoded literal reintroduced")

	// DispatchSubagents: no tasks provided.
	rec2 := &recordingTranslator{}
	p2 := NewPillarWithTranslator(cfg, logger, rec2)
	_, err = p2.DispatchSubagents(ctx, nil, 4)
	require.Error(t, err)
	assert.Contains(
		t, rec2.calls,
		"helixspecifier_superpowers_err_no_tasks",
	)

	// ReviewCode: no task results to review.
	rec3 := &recordingTranslator{}
	p3 := NewPillarWithTranslator(cfg, logger, rec3)
	_, err = p3.ReviewCode(ctx, nil)
	require.Error(t, err)
	assert.Contains(
		t, rec3.calls,
		"helixspecifier_superpowers_err_no_task_results",
	)
	assert.NotContains(t, err.Error(), "no task results to review")
}

// TestPillar_Round405_NoopFallbackEmitsEnglish proves the
// NoopTranslator path (via i18n.ResolveOrFallback) emits the
// bundled English fallback — NOT a raw i18n key — so the real
// Pillar still produces a coherent error when no translator is
// wired.
func TestPillar_Round405_NoopFallbackEmitsEnglish(t *testing.T) {
	logger := logrus.New()
	cfg := config.DefaultConfig()
	p := NewPillar(cfg, logger)
	ctx := context.Background()

	_, err := p.ExecuteWithTDD(ctx, nil)
	require.Error(t, err)
	assert.Equal(t, "no tasks provided", err.Error())
	assert.NotContains(
		t, err.Error(), "helixspecifier_superpowers_err_",
		"raw i18n key leaked into noop-path output",
	)
}
