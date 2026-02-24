package superpowers

import (
	"context"
	"fmt"
	"sync/atomic"
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

func makeTasks(count int) []types.Task {
	tasks := make([]types.Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = types.Task{
			ID:       fmt.Sprintf("task-%d", i),
			Name:     fmt.Sprintf("Task %d", i),
			Priority: "medium",
			Acceptance: []string{
				"acceptance-1",
				"acceptance-2",
			},
			FilesAffected: []string{
				fmt.Sprintf("file_%d.go", i),
			},
		}
	}
	return tasks
}

// --- TestNewPillar ---

func TestNewPillar(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := logrus.New()
	p := NewPillar(cfg, logger)

	require.NotNil(t, p, "NewPillar should return non-nil")
	assert.Equal(t, cfg, p.cfg)
	assert.Equal(t, logger, p.logger)
}

// --- TestPillar_ExecuteWithTDD_Success ---

func TestPillar_ExecuteWithTDD_Success(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(3)

	results, err := p.ExecuteWithTDD(ctx, tasks)

	require.NoError(t, err)
	assert.Len(t, results, 3)
	for i, r := range results {
		assert.True(t, r.Success,
			"Task %d should succeed", i)
		assert.Equal(t, tasks[i].ID, r.TaskID)
		assert.Equal(t, 2, r.TestsPassed)
		assert.Equal(t, 0, r.TestsFailed)
	}
}

// --- TestPillar_ExecuteWithTDD_EmptyTasks ---

func TestPillar_ExecuteWithTDD_EmptyTasks(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	results, err := p.ExecuteWithTDD(ctx, []types.Task{})

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "no tasks provided")
}

// --- TestPillar_ExecuteWithTDD_MultipleTasks ---

func TestPillar_ExecuteWithTDD_MultipleTasks(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(10)

	results, err := p.ExecuteWithTDD(ctx, tasks)

	require.NoError(t, err)
	assert.Len(t, results, 10)
	for _, r := range results {
		assert.True(t, r.Success)
		assert.NotEmpty(t, r.Output)
	}
}

// --- TestPillar_ExecuteWithTDD_ContextCancel ---

func TestPillar_ExecuteWithTDD_ContextCancel(t *testing.T) {
	p := newTestPillar()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	tasks := makeTasks(5)
	results, err := p.ExecuteWithTDD(ctx, tasks)

	// Should return context error after processing at most
	// a partial set
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.True(t, len(results) < len(tasks),
		"Should not complete all tasks after cancel")
}

// --- TestPillar_DispatchSubagents_Success ---

func TestPillar_DispatchSubagents_Success(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(4)

	results, err := p.DispatchSubagents(ctx, tasks, 2)

	require.NoError(t, err)
	assert.Len(t, results, 4)
	for i, r := range results {
		assert.True(t, r.Success,
			"Task %d should succeed", i)
		assert.Equal(t, tasks[i].ID, r.TaskID)
	}
}

// --- TestPillar_DispatchSubagents_EmptyTasks ---

func TestPillar_DispatchSubagents_EmptyTasks(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	results, err := p.DispatchSubagents(
		ctx, []types.Task{}, 2,
	)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "no tasks provided")
}

// --- TestPillar_DispatchSubagents_Parallel ---

func TestPillar_DispatchSubagents_Parallel(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(8)

	results, err := p.DispatchSubagents(ctx, tasks, 4)

	require.NoError(t, err)
	assert.Len(t, results, 8)

	// Verify all task IDs present
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.TaskID] = true
	}
	for _, task := range tasks {
		assert.True(t, ids[task.ID],
			"Task %s should be in results", task.ID)
	}
}

// --- TestPillar_DispatchSubagents_MaxParallel ---

func TestPillar_DispatchSubagents_MaxParallel(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(6)

	// With maxParallel=1, execution is essentially serial
	results, err := p.DispatchSubagents(ctx, tasks, 1)

	require.NoError(t, err)
	assert.Len(t, results, 6)
	for _, r := range results {
		assert.True(t, r.Success)
	}
}

// --- TestPillar_DispatchSubagents_DefaultParallel ---

func TestPillar_DispatchSubagents_DefaultParallel(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(3)

	// maxParallel <= 0 triggers default from config
	results, err := p.DispatchSubagents(ctx, tasks, 0)

	require.NoError(t, err)
	assert.Len(t, results, 3)
}

// --- TestPillar_ReviewCode_AllPassed ---

func TestPillar_ReviewCode_AllPassed(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	taskResults := []types.TaskResult{
		{TaskID: "t1", Success: true},
		{TaskID: "t2", Success: true},
		{TaskID: "t3", Success: true},
	}

	review, err := p.ReviewCode(ctx, taskResults)

	require.NoError(t, err)
	require.NotNil(t, review)
	assert.True(t, review.Approved)
	assert.Equal(t, 1.0, review.Score)
	assert.Empty(t, review.Issues)
}

// --- TestPillar_ReviewCode_SomeFailed ---

func TestPillar_ReviewCode_SomeFailed(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	taskResults := []types.TaskResult{
		{TaskID: "t1", Success: true},
		{TaskID: "t2", Success: false, Error: "compile error"},
		{TaskID: "t3", Success: true},
		{TaskID: "t4", Success: false, Error: "test failure"},
	}

	review, err := p.ReviewCode(ctx, taskResults)

	require.NoError(t, err)
	require.NotNil(t, review)
	assert.False(t, review.Approved)
	assert.Equal(t, 0.5, review.Score)
	assert.Len(t, review.Issues, 2)
	assert.Equal(t, "critical", review.Issues[0].Severity)
	assert.Contains(t, review.Issues[0].Description,
		"compile error")
}

// --- TestPillar_ReviewCode_Empty ---

func TestPillar_ReviewCode_Empty(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	review, err := p.ReviewCode(ctx, []types.TaskResult{})

	require.Error(t, err)
	assert.Nil(t, review)
	assert.Contains(t, err.Error(),
		"no task results to review")
}

// --- TestPillar_ImplementsInterface ---

func TestPillar_ImplementsInterface(t *testing.T) {
	var _ types.SuperpowersPillar = (*Pillar)(nil)
	p := newTestPillar()
	var iface types.SuperpowersPillar = p
	assert.NotNil(t, iface,
		"Pillar must implement SuperpowersPillar")
}

// --- TestPillar_ConcurrentDispatch ---

func TestPillar_ConcurrentDispatch(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	tasks := makeTasks(20)

	var completed int64

	// Run dispatch concurrently from multiple goroutines
	done := make(chan struct{}, 3)
	for g := 0; g < 3; g++ {
		go func() {
			defer func() { done <- struct{}{} }()
			results, err := p.DispatchSubagents(
				ctx, tasks, 4,
			)
			if err == nil {
				atomic.AddInt64(
					&completed, int64(len(results)),
				)
			}
		}()
	}

	for g := 0; g < 3; g++ {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent dispatch timed out")
		}
	}

	assert.Equal(t, int64(60), atomic.LoadInt64(&completed),
		"All goroutines should complete all tasks")
}

// --- TestPillar_DispatchSubagents_ContextCancel ---

func TestPillar_DispatchSubagents_ContextCancel(t *testing.T) {
	p := newTestPillar()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately before dispatch
	cancel()

	tasks := makeTasks(10)
	results, err := p.DispatchSubagents(ctx, tasks, 2)

	// At least some goroutines should hit the cancelled context
	// The function returns firstErr which captures ctx.Err()
	if err != nil {
		assert.ErrorIs(t, err, context.Canceled)
	}
	// Results slice is always allocated to len(tasks)
	assert.Len(t, results, len(tasks))
}

// --- TestPillar_ExecuteWithTDD_FilesChanged ---

func TestPillar_ExecuteWithTDD_FilesChanged(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	tasks := []types.Task{
		{
			ID:   "t1",
			Name: "Auth task",
			FilesAffected: []string{
				"auth.go",
				"auth_test.go",
			},
			Acceptance: []string{"acc1"},
		},
	}

	results, err := p.ExecuteWithTDD(ctx, tasks)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, []string{"auth.go", "auth_test.go"},
		results[0].FilesChanged)
	assert.Equal(t, 1, results[0].TestsPassed)
}
