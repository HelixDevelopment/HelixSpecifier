package memory

import (
	"context"
	"sync"
	"testing"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	require.NotNil(t, s)
	assert.Equal(t, 0, s.SpecCount())
	assert.Equal(t, 0, s.PatternCount())
}

func TestStore_Store_Success(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	spec := &types.Specification{
		ID:          "spec-1",
		Title:       "Test Spec",
		Description: "A test specification",
	}

	err := s.Store(ctx, spec)
	require.NoError(t, err)
	assert.Equal(t, 1, s.SpecCount())
}

func TestStore_Store_Nil(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	err := s.Store(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestStore_Search_Found(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	_ = s.Store(ctx, &types.Specification{
		ID:          "spec-1",
		Title:       "Authentication Module",
		Description: "Handles user authentication",
	})
	_ = s.Store(ctx, &types.Specification{
		ID:          "spec-2",
		Title:       "Database Layer",
		Description: "PostgreSQL storage",
	})

	results, err := s.Search(ctx, "auth", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "spec-1", results[0].ID)
}

func TestStore_Search_NotFound(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	_ = s.Store(ctx, &types.Specification{
		ID:    "spec-1",
		Title: "Authentication Module",
	})

	results, err := s.Search(ctx, "kubernetes", 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_Search_Limit(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	for i := 0; i < 20; i++ {
		_ = s.Store(ctx, &types.Specification{
			ID:    "spec",
			Title: "Test spec item",
		})
	}

	results, err := s.Search(ctx, "test", 5)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestStore_LearnFromFlow_Success(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	flow := &types.FlowResult{
		FlowID:              "flow-1",
		Success:             true,
		EffortLevel:         types.EffortLarge,
		CeremonyLevel:       types.CeremonyStandard,
		OverallQualityScore: 0.85,
		PhaseResults: []types.PhaseResult{
			{Phase: types.PhaseSpecify, Success: true},
		},
	}

	err := s.LearnFromFlow(ctx, flow)
	require.NoError(t, err)
	assert.Equal(t, 1, s.PatternCount())

	patterns := s.GetPatterns()
	assert.Equal(t, "flow-1", patterns[0].ID)
	assert.Contains(t, patterns[0].Pattern, "large")
}

func TestStore_LearnFromFlow_Failed(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	flow := &types.FlowResult{
		FlowID:  "flow-fail",
		Success: false,
	}

	err := s.LearnFromFlow(ctx, flow)
	require.NoError(t, err)
	// Should not learn from failed flows
	assert.Equal(t, 0, s.PatternCount())
}

func TestStore_LearnFromFlow_Nil(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	err := s.LearnFromFlow(ctx, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestStore_GetPatterns(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	_ = s.LearnFromFlow(ctx, &types.FlowResult{
		FlowID:      "f1",
		Success:     true,
		EffortLevel: types.EffortQuick,
	})
	_ = s.LearnFromFlow(ctx, &types.FlowResult{
		FlowID:      "f2",
		Success:     true,
		EffortLevel: types.EffortLarge,
	})

	patterns := s.GetPatterns()
	assert.Len(t, patterns, 2)

	// Ensure it is a copy
	patterns[0].ID = "mutated"
	original := s.GetPatterns()
	assert.Equal(t, "f1", original[0].ID)
}

func TestStore_SpecCount(t *testing.T) {
	s := NewStore()
	ctx := context.Background()

	assert.Equal(t, 0, s.SpecCount())

	_ = s.Store(ctx, &types.Specification{
		ID: "a", Title: "A",
	})
	_ = s.Store(ctx, &types.Specification{
		ID: "b", Title: "B",
	})

	assert.Equal(t, 2, s.SpecCount())
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := NewStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Store(ctx, &types.Specification{
				ID:    "id",
				Title: "concurrent spec",
			})
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = s.Search(ctx, "concurrent", 10)
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.LearnFromFlow(ctx, &types.FlowResult{
				FlowID:  "f",
				Success: true,
			})
		}()
	}

	wg.Wait()
	assert.Equal(t, 50, s.SpecCount())
	assert.Equal(t, 50, s.PatternCount())
}
