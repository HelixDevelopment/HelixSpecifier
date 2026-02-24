package parallel_execution

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	e := NewExecutor(8, 10*time.Second)
	require.NotNil(t, e)
	assert.Equal(t, 8, e.MaxWorkers())
	assert.Equal(t, 10*time.Second, e.Timeout())
}

func TestNewExecutor_Defaults(t *testing.T) {
	e := NewExecutor(0, 0)
	assert.Equal(t, 4, e.MaxWorkers())
	assert.Equal(t, 30*time.Second, e.Timeout())
}

func TestExecutor_Execute_Success(t *testing.T) {
	e := NewExecutor(2, 5*time.Second)
	ctx := context.Background()

	tasks := map[string]TaskFunc{
		"t1": func(_ context.Context) (string, error) {
			return "output1", nil
		},
		"t2": func(_ context.Context) (string, error) {
			return "output2", nil
		},
	}

	results := e.Execute(ctx, tasks)
	assert.Len(t, results, 2)

	completed, failed := e.Stats()
	assert.Equal(t, 2, completed)
	assert.Equal(t, 0, failed)
}

func TestExecutor_Execute_WithFailure(t *testing.T) {
	e := NewExecutor(2, 5*time.Second)
	ctx := context.Background()

	tasks := map[string]TaskFunc{
		"ok": func(_ context.Context) (string, error) {
			return "ok", nil
		},
		"fail": func(_ context.Context) (string, error) {
			return "", fmt.Errorf("task error")
		},
	}

	results := e.Execute(ctx, tasks)
	assert.Len(t, results, 2)

	completed, failed := e.Stats()
	assert.Equal(t, 1, completed)
	assert.Equal(t, 1, failed)
}

func TestExecutor_String(t *testing.T) {
	e := NewExecutor(4, 10*time.Second)
	s := e.String()
	assert.Contains(t, s, "ParallelExecutor")
	assert.Contains(t, s, "workers=4")
}
