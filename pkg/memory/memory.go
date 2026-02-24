package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"digital.vasic.helixspecifier/pkg/types"
)

// Store implements SpecMemory for specification storage and
// cross-project learning.
type Store struct {
	specs    []types.Specification
	patterns []LearnedPattern
	mu       sync.RWMutex
}

// LearnedPattern represents a pattern learned from flows
type LearnedPattern struct {
	ID          string    `json:"id"`
	Pattern     string    `json:"pattern"`
	Context     string    `json:"context"`
	LearnedAt   time.Time `json:"learned_at"`
	UseCount    int       `json:"use_count"`
	SuccessRate float64   `json:"success_rate"`
}

// NewStore creates a new spec memory store
func NewStore() *Store {
	return &Store{
		specs:    []types.Specification{},
		patterns: []LearnedPattern{},
	}
}

// Store saves a specification
func (s *Store) Store(
	ctx context.Context,
	spec *types.Specification,
) error {
	if spec == nil {
		return fmt.Errorf("spec cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	spec.UpdatedAt = time.Now()
	s.specs = append(s.specs, *spec)
	return nil
}

// Search finds similar specifications
func (s *Store) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]types.Specification, error) {
	if limit <= 0 {
		limit = 10
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	lower := strings.ToLower(query)
	results := []types.Specification{}

	for _, spec := range s.specs {
		if strings.Contains(
			strings.ToLower(spec.Title), lower,
		) || strings.Contains(
			strings.ToLower(spec.Description), lower,
		) {
			results = append(results, spec)
			if len(results) >= limit {
				break
			}
		}
	}

	return results, nil
}

// LearnFromFlow extracts patterns from completed flows
func (s *Store) LearnFromFlow(
	ctx context.Context,
	flow *types.FlowResult,
) error {
	if flow == nil {
		return fmt.Errorf("flow cannot be nil")
	}
	if !flow.Success {
		return nil // Only learn from successful flows
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pattern := LearnedPattern{
		ID: flow.FlowID,
		Pattern: fmt.Sprintf(
			"effort=%s ceremony=%s phases=%d "+
				"quality=%.2f",
			flow.EffortLevel,
			flow.CeremonyLevel,
			len(flow.PhaseResults),
			flow.OverallQualityScore,
		),
		Context:     string(flow.EffortLevel),
		LearnedAt:   time.Now(),
		UseCount:    0,
		SuccessRate: 1.0,
	}

	s.patterns = append(s.patterns, pattern)
	return nil
}

// GetPatterns returns all learned patterns
func (s *Store) GetPatterns() []LearnedPattern {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]LearnedPattern, len(s.patterns))
	copy(result, s.patterns)
	return result
}

// SpecCount returns the number of stored specs
func (s *Store) SpecCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.specs)
}

// PatternCount returns the number of learned patterns
func (s *Store) PatternCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.patterns)
}
