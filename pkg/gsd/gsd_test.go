package gsd

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPillar() *Pillar {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return NewPillar(logger)
}

func newTestSpec() *types.Specification {
	return &types.Specification{
		ID:    "spec-1",
		Title: "Auth Module",
	}
}

func newTestTasks() []types.Task {
	return []types.Task{
		{
			ID:       "t1",
			Name:     "Create auth handler",
			Priority: "critical",
			Acceptance: []string{
				"handler responds 200",
			},
		},
		{
			ID:       "t2",
			Name:     "Add JWT middleware",
			Priority: "high",
			Acceptance: []string{
				"tokens validated",
			},
		},
		{
			ID:       "t3",
			Name:     "Write auth tests",
			Priority: "medium",
			Acceptance: []string{
				"coverage > 80%",
			},
		},
		{
			ID:       "t4",
			Name:     "Update docs",
			Priority: "low",
			Acceptance: []string{
				"docs updated",
			},
		},
	}
}

// --- TestNewPillar ---

func TestNewPillar(t *testing.T) {
	logger := logrus.New()
	p := NewPillar(logger)

	require.NotNil(t, p)
	assert.Equal(t, logger, p.logger)
	assert.NotNil(t, p.milestones)
	assert.Empty(t, p.milestones)
}

// --- TestPillar_CreateMilestones_Success ---

func TestPillar_CreateMilestones_Success(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	spec := newTestSpec()
	tasks := newTestTasks()

	milestones, err := p.CreateMilestones(ctx, spec, tasks)

	require.NoError(t, err)
	assert.Len(t, milestones, 4,
		"Should create 4 milestones (one per priority)")
	for _, ms := range milestones {
		assert.NotEmpty(t, ms.ID)
		assert.NotEmpty(t, ms.Name)
		assert.Equal(t, types.MilestoneQueued, ms.Status)
		assert.Equal(t, 0.0, ms.Progress)
	}
}

// --- TestPillar_CreateMilestones_NilSpec ---

func TestPillar_CreateMilestones_NilSpec(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	milestones, err := p.CreateMilestones(
		ctx, nil, newTestTasks(),
	)

	require.Error(t, err)
	assert.Nil(t, milestones)
	assert.Contains(t, err.Error(),
		"specification required")
}

// --- TestPillar_CreateMilestones_EmptyTasks ---

func TestPillar_CreateMilestones_EmptyTasks(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), []types.Task{},
	)

	require.Error(t, err)
	assert.Nil(t, milestones)
	assert.Contains(t, err.Error(),
		"at least one task required")
}

// --- TestPillar_CreateMilestones_GroupsByPriority ---

func TestPillar_CreateMilestones_GroupsByPriority(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	spec := newTestSpec()

	// All tasks same priority
	tasks := []types.Task{
		{ID: "t1", Name: "A", Priority: "high"},
		{ID: "t2", Name: "B", Priority: "high"},
		{ID: "t3", Name: "C", Priority: "high"},
	}

	milestones, err := p.CreateMilestones(ctx, spec, tasks)

	require.NoError(t, err)
	assert.Len(t, milestones, 1,
		"Same priority should produce 1 milestone")
	assert.Len(t, milestones[0].Tasks, 3)
}

// --- TestPillar_CreateMilestones_Dependencies ---

func TestPillar_CreateMilestones_Dependencies(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	spec := newTestSpec()
	tasks := newTestTasks()

	milestones, err := p.CreateMilestones(ctx, spec, tasks)

	require.NoError(t, err)
	require.Len(t, milestones, 4)

	// First milestone has no dependencies
	assert.Empty(t, milestones[0].Dependencies,
		"First milestone should have no dependencies")

	// Second depends on first
	assert.Len(t, milestones[1].Dependencies, 1)
	assert.Equal(t, milestones[0].ID,
		milestones[1].Dependencies[0])

	// Third depends on first and second
	assert.Len(t, milestones[2].Dependencies, 2)

	// Fourth depends on all three
	assert.Len(t, milestones[3].Dependencies, 3)
}

// --- TestPillar_TrackProgress_Success ---

func TestPillar_TrackProgress_Success(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), newTestTasks(),
	)
	require.NoError(t, err)
	require.NotEmpty(t, milestones)

	msID := milestones[0].ID
	taskResults := []types.TaskResult{
		{TaskID: "t1", Success: true},
	}

	ms, err := p.TrackProgress(ctx, msID, taskResults)

	require.NoError(t, err)
	require.NotNil(t, ms)
	assert.Equal(t, 1.0, ms.Progress,
		"1 of 1 critical task done = 100%")
	assert.Equal(t, types.MilestoneCompleted, ms.Status)
	assert.NotNil(t, ms.CompletedAt)
}

// --- TestPillar_TrackProgress_NotFound ---

func TestPillar_TrackProgress_NotFound(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	ms, err := p.TrackProgress(
		ctx, "nonexistent", []types.TaskResult{},
	)

	require.Error(t, err)
	assert.Nil(t, ms)
	assert.Contains(t, err.Error(), "not found")
}

// --- TestPillar_TrackProgress_EmptyResults ---

func TestPillar_TrackProgress_EmptyResults(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), newTestTasks(),
	)
	require.NoError(t, err)

	msID := milestones[0].ID
	ms, err := p.TrackProgress(
		ctx, msID, []types.TaskResult{},
	)

	require.NoError(t, err)
	require.NotNil(t, ms)
	assert.Equal(t, 0.0, ms.Progress,
		"Empty results should not change progress")
	assert.Equal(t, types.MilestoneQueued, ms.Status)
}

// --- TestPillar_TrackProgress_AllComplete ---

func TestPillar_TrackProgress_AllComplete(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	// Create with single-priority tasks to get 1 milestone
	tasks := []types.Task{
		{ID: "t1", Name: "A", Priority: "medium"},
		{ID: "t2", Name: "B", Priority: "medium"},
	}

	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), tasks,
	)
	require.NoError(t, err)
	require.Len(t, milestones, 1)

	msID := milestones[0].ID
	taskResults := []types.TaskResult{
		{TaskID: "t1", Success: true},
		{TaskID: "t2", Success: true},
	}

	ms, err := p.TrackProgress(ctx, msID, taskResults)

	require.NoError(t, err)
	assert.Equal(t, 1.0, ms.Progress)
	assert.Equal(t, types.MilestoneCompleted, ms.Status)
	assert.NotNil(t, ms.CompletedAt)
}

// --- TestPillar_TrackProgress_StatusTransition ---

func TestPillar_TrackProgress_StatusTransition(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	tasks := []types.Task{
		{ID: "t1", Name: "A", Priority: "high"},
		{ID: "t2", Name: "B", Priority: "high"},
		{ID: "t3", Name: "C", Priority: "high"},
	}

	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), tasks,
	)
	require.NoError(t, err)
	msID := milestones[0].ID

	// Partial completion: queued -> active
	partialResults := []types.TaskResult{
		{TaskID: "t1", Success: true},
	}
	ms, err := p.TrackProgress(ctx, msID, partialResults)
	require.NoError(t, err)
	assert.Equal(t, types.MilestoneActive, ms.Status)
	assert.NotNil(t, ms.StartedAt)
	assert.InDelta(t, 0.333, ms.Progress, 0.01)
}

// --- TestPillar_GetProgress_Empty ---

func TestPillar_GetProgress_Empty(t *testing.T) {
	p := newTestPillar()

	progress, err := p.GetProgress("any-flow")

	require.NoError(t, err)
	assert.Equal(t, 0.0, progress)
}

// --- TestPillar_GetProgress_WithMilestones ---

func TestPillar_GetProgress_WithMilestones(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), newTestTasks(),
	)
	require.NoError(t, err)
	require.NotEmpty(t, milestones)

	// Complete first milestone
	msID := milestones[0].ID
	taskResults := []types.TaskResult{
		{TaskID: "t1", Success: true},
	}
	_, err = p.TrackProgress(ctx, msID, taskResults)
	require.NoError(t, err)

	progress, err := p.GetProgress("flow-1")

	require.NoError(t, err)
	assert.Greater(t, progress, 0.0,
		"Progress should be > 0 after completing a milestone")
	assert.LessOrEqual(t, progress, 1.0,
		"Progress should be <= 1.0")
}

// --- TestPillar_ImplementsInterface ---

func TestPillar_ImplementsInterface(t *testing.T) {
	var _ types.GSDPillar = (*Pillar)(nil)
	p := newTestPillar()
	var iface types.GSDPillar = p
	assert.NotNil(t, iface,
		"Pillar must implement GSDPillar interface")
}

// --- TestPillar_ConcurrentAccess ---

func TestPillar_ConcurrentAccess(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()

	// Create milestones first
	milestones, err := p.CreateMilestones(
		ctx, newTestSpec(), newTestTasks(),
	)
	require.NoError(t, err)
	require.NotEmpty(t, milestones)

	// Concurrent reads and writes
	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.GetProgress("flow")
			if err != nil {
				errCh <- err
			}
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			spec := &types.Specification{
				ID:    fmt.Sprintf("spec-%d", idx),
				Title: fmt.Sprintf("Spec %d", idx),
			}
			tasks := []types.Task{
				{
					ID:       fmt.Sprintf("ct-%d", idx),
					Name:     "task",
					Priority: "medium",
				},
			}
			_, err := p.CreateMilestones(ctx, spec, tasks)
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		t.Errorf("Concurrent access error: %v", e)
	}
}

// --- TestPillar_CreateMilestones_DefaultPriority ---

func TestPillar_CreateMilestones_DefaultPriority(t *testing.T) {
	p := newTestPillar()
	ctx := context.Background()
	spec := newTestSpec()

	// Tasks without explicit priority default to medium
	tasks := []types.Task{
		{ID: "t1", Name: "No priority"},
	}

	milestones, err := p.CreateMilestones(ctx, spec, tasks)

	require.NoError(t, err)
	assert.Len(t, milestones, 1)
	assert.Contains(t, milestones[0].Description, "medium")
}
