// Package parallel_execution provides parallel task execution
// capabilities for the HelixSpecifier engine.
package parallel_execution

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskFunc is a function that executes a task and returns
// an output string and optional error.
type TaskFunc func(ctx context.Context) (string, error)

// Executor runs tasks in parallel with bounded concurrency.
type Executor struct {
	maxWorkers int
	timeout    time.Duration
	mu         sync.Mutex
	completed  int
	failed     int
}

// Result holds the outcome of a single parallel task.
type Result struct {
	TaskID   string        `json:"task_id"`
	Output   string        `json:"output"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error,omitempty"`
}

// NewExecutor creates a new parallel executor with the given
// concurrency limit and per-task timeout.
func NewExecutor(
	maxWorkers int,
	timeout time.Duration,
) *Executor {
	if maxWorkers <= 0 {
		maxWorkers = 4
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &Executor{
		maxWorkers: maxWorkers,
		timeout:    timeout,
	}
}

// Execute runs all tasks in parallel with bounded concurrency
// and returns results for each task.
func (e *Executor) Execute(
	ctx context.Context,
	tasks map[string]TaskFunc,
) []Result {
	results := make([]Result, 0, len(tasks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, e.maxWorkers)

	for id, fn := range tasks {
		wg.Add(1)
		go func(taskID string, taskFn TaskFunc) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			start := time.Now()
			taskCtx, cancel := context.WithTimeout(
				ctx, e.timeout,
			)
			defer cancel()

			output, err := taskFn(taskCtx)
			dur := time.Since(start)

			r := Result{
				TaskID:   taskID,
				Output:   output,
				Duration: dur,
				Error:    err,
			}

			mu.Lock()
			results = append(results, r)
			e.mu.Lock()
			if err != nil {
				e.failed++
			} else {
				e.completed++
			}
			e.mu.Unlock()
			mu.Unlock()
		}(id, fn)
	}

	wg.Wait()
	return results
}

// Stats returns completed and failed counts.
func (e *Executor) Stats() (completed, failed int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.completed, e.failed
}

// MaxWorkers returns the max concurrency level.
func (e *Executor) MaxWorkers() int {
	return e.maxWorkers
}

// Timeout returns the per-task timeout.
func (e *Executor) Timeout() time.Duration {
	return e.timeout
}

// String returns a human-readable description.
func (e *Executor) String() string {
	return fmt.Sprintf(
		"ParallelExecutor(workers=%d, timeout=%s)",
		e.maxWorkers, e.timeout,
	)
}
