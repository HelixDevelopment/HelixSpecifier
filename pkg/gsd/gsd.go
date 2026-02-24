package gsd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Pillar implements the GSDPillar interface providing
// milestone lifecycle management (Get-Shit-Done).
type Pillar struct {
	logger     *logrus.Logger
	milestones map[string]*types.Milestone
	mu         sync.RWMutex
}

// NewPillar creates a new GSD pillar
func NewPillar(logger *logrus.Logger) *Pillar {
	return &Pillar{
		logger:     logger,
		milestones: make(map[string]*types.Milestone),
	}
}

// CreateMilestones breaks work into milestones from spec
// and tasks.
func (p *Pillar) CreateMilestones(
	ctx context.Context,
	spec *types.Specification,
	tasks []types.Task,
) ([]types.Milestone, error) {
	if spec == nil {
		return nil, fmt.Errorf("specification required")
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf(
			"at least one task required",
		)
	}

	// Group tasks by priority for milestone creation
	groups := map[string][]types.Task{
		"critical": {},
		"high":     {},
		"medium":   {},
		"low":      {},
	}

	for _, task := range tasks {
		priority := task.Priority
		if priority == "" {
			priority = "medium"
		}
		groups[priority] = append(groups[priority], task)
	}

	milestones := []types.Milestone{}
	order := []string{"critical", "high", "medium", "low"}
	depChain := []string{}

	for _, priority := range order {
		groupTasks := groups[priority]
		if len(groupTasks) == 0 {
			continue
		}

		id := fmt.Sprintf(
			"ms-%s", uuid.New().String()[:8],
		)
		ms := types.Milestone{
			ID:   id,
			Name: fmt.Sprintf(
				"%s: %s tasks", spec.Title, priority,
			),
			Description: fmt.Sprintf(
				"%d %s-priority tasks",
				len(groupTasks), priority,
			),
			Status:       types.MilestoneQueued,
			Tasks:        groupTasks,
			Dependencies: append([]string{}, depChain...),
			Progress:     0.0,
		}

		milestones = append(milestones, ms)
		depChain = append(depChain, id)

		p.mu.Lock()
		p.milestones[id] = &ms
		p.mu.Unlock()
	}

	p.logger.WithField("count", len(milestones)).Info(
		"[GSD] Created milestones",
	)

	return milestones, nil
}

// TrackProgress updates milestone progress
func (p *Pillar) TrackProgress(
	ctx context.Context,
	milestoneID string,
	taskResults []types.TaskResult,
) (*types.Milestone, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ms, ok := p.milestones[milestoneID]
	if !ok {
		return nil, fmt.Errorf(
			"milestone %s not found", milestoneID,
		)
	}

	if len(taskResults) == 0 {
		return ms, nil
	}

	completed := 0
	for _, tr := range taskResults {
		if tr.Success {
			completed++
		}
	}

	total := len(ms.Tasks)
	if total > 0 {
		ms.Progress = float64(completed) / float64(total)
	}

	now := time.Now()
	if ms.Status == types.MilestoneQueued && completed > 0 {
		ms.Status = types.MilestoneActive
		ms.StartedAt = &now
	}

	if ms.Progress >= 1.0 {
		ms.Status = types.MilestoneCompleted
		ms.CompletedAt = &now
	}

	return ms, nil
}

// GetProgress returns overall progress across all milestones
func (p *Pillar) GetProgress(flowID string) (float64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.milestones) == 0 {
		return 0.0, nil
	}

	total := 0.0
	for _, ms := range p.milestones {
		total += ms.Progress
	}

	return total / float64(len(p.milestones)), nil
}
