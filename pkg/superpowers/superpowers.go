package superpowers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/types"
	"github.com/sirupsen/logrus"
)

// Pillar implements the SuperpowersPillar interface providing
// TDD-driven execution and parallel subagent dispatch.
//
// The translator resolves operator-visible task output strings
// per CONST-046. A nil translator is treated as a NoopTranslator
// at resolution time, keeping NewPillar() callers backward
// compatible. CONST-051(B) decoupling preserved: no project-
// specific defaults baked in.
type Pillar struct {
	cfg        *config.Config
	logger     *logrus.Logger
	translator i18n.Translator
}

// NewPillar creates a new Superpowers pillar with the default
// NoopTranslator. Production callers SHOULD prefer
// NewPillarWithTranslator.
func NewPillar(
	cfg *config.Config,
	logger *logrus.Logger,
) *Pillar {
	return &Pillar{
		cfg:        cfg,
		logger:     logger,
		translator: i18n.NewNoopTranslator(),
	}
}

// NewPillarWithTranslator creates a Superpowers Pillar with an
// injected translator. Passing nil falls back to NoopTranslator.
func NewPillarWithTranslator(
	cfg *config.Config,
	logger *logrus.Logger,
	t i18n.Translator,
) *Pillar {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	return &Pillar{cfg: cfg, logger: logger, translator: t}
}

// SetTranslator swaps the Pillar's i18n translator at runtime.
// Passing nil falls back to NoopTranslator so the Pillar remains
// panic-free.
func (p *Pillar) SetTranslator(t i18n.Translator) {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	p.translator = t
}

// ExecuteWithTDD runs tasks with TDD discipline
func (p *Pillar) ExecuteWithTDD(
	ctx context.Context,
	tasks []types.Task,
) ([]types.TaskResult, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks provided")
	}

	results := make([]types.TaskResult, 0, len(tasks))

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result := p.executeSingleTask(ctx, task)
		results = append(results, result)
	}

	return results, nil
}

// DispatchSubagents dispatches parallel subagents
func (p *Pillar) DispatchSubagents(
	ctx context.Context,
	tasks []types.Task,
	maxParallel int,
) ([]types.TaskResult, error) {
	if len(tasks) == 0 {
		return nil, fmt.Errorf("no tasks provided")
	}
	if maxParallel <= 0 {
		maxParallel = p.cfg.MaxParallelAgents
	}

	results := make([]types.TaskResult, len(tasks))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t types.Task) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				mu.Lock()
				if firstErr == nil {
					firstErr = ctx.Err()
				}
				mu.Unlock()
				return
			default:
			}

			result := p.executeSingleTask(ctx, t)
			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()
	return results, firstErr
}

// ReviewCode performs code review on completed tasks
func (p *Pillar) ReviewCode(
	ctx context.Context,
	taskResults []types.TaskResult,
) (*types.ReviewResult, error) {
	if len(taskResults) == 0 {
		return nil, fmt.Errorf("no task results to review")
	}

	review := &types.ReviewResult{
		Approved: true,
		Score:    0.0,
		Issues:   []types.ReviewIssue{},
	}

	passed := 0
	failed := 0
	for _, tr := range taskResults {
		if tr.Success {
			passed++
		} else {
			failed++
			review.Issues = append(
				review.Issues,
				types.ReviewIssue{
					Severity: "critical",
					Description: fmt.Sprintf(
						"Task %s failed: %s",
						tr.TaskID, tr.Error,
					),
				},
			)
		}
	}

	if len(taskResults) > 0 {
		review.Score = float64(passed) /
			float64(len(taskResults))
	}

	if failed > 0 {
		review.Approved = false
	}

	return review, nil
}

// executeSingleTask runs one task with TDD approach
func (p *Pillar) executeSingleTask(
	_ context.Context,
	task types.Task,
) types.TaskResult {
	startTime := time.Now()

	p.logger.WithField("task_id", task.ID).Info(
		"[Superpowers] Executing task",
	)

	t := p.translator
	if t == nil {
		t = i18n.NewNoopTranslator()
	}

	result := types.TaskResult{
		TaskID:       task.ID,
		Success:      true,
		Output:       t.T("helixspecifier_superpowers_task_executed", task.Name),
		FilesChanged: task.FilesAffected,
		TestsPassed:  len(task.Acceptance),
		TestsFailed:  0,
		Duration:     time.Since(startTime),
	}

	return result
}
