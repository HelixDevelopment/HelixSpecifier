// Package brownfield provides analysis capabilities for
// existing codebases, enabling HelixSpecifier to understand
// legacy code patterns and constraints before generating
// specifications.
package brownfield

import (
	"fmt"
	"strings"
	"sync"
)

// CodePattern represents a detected code pattern.
type CodePattern struct {
	Name       string  `json:"name"`
	Frequency  int     `json:"frequency"`
	Confidence float64 `json:"confidence"`
}

// AnalysisResult holds the result of a codebase analysis.
type AnalysisResult struct {
	Patterns     []CodePattern `json:"patterns"`
	Dependencies []string      `json:"dependencies"`
	Complexity   float64       `json:"complexity"`
	LinesOfCode  int           `json:"lines_of_code"`
}

// Analyzer examines existing codebases for patterns and
// constraints that inform specification generation.
type Analyzer struct {
	patterns map[string]*CodePattern
	deps     map[string]bool
	mu       sync.RWMutex
}

// NewAnalyzer creates a new brownfield analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		patterns: make(map[string]*CodePattern),
		deps:     make(map[string]bool),
	}
}

// RecordPattern records a detected pattern in the codebase.
func (a *Analyzer) RecordPattern(
	name string,
	confidence float64,
) error {
	if name == "" {
		return fmt.Errorf("pattern name cannot be empty")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	p, ok := a.patterns[name]
	if !ok {
		a.patterns[name] = &CodePattern{
			Name:       name,
			Frequency:  1,
			Confidence: confidence,
		}
		return nil
	}

	p.Frequency++
	// Running average for confidence
	p.Confidence += (confidence - p.Confidence) /
		float64(p.Frequency)
	return nil
}

// AddDependency records a detected dependency.
func (a *Analyzer) AddDependency(dep string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.deps[dep] = true
}

// Analyze returns the current analysis result.
func (a *Analyzer) Analyze() *AnalysisResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	patterns := make([]CodePattern, 0, len(a.patterns))
	for _, p := range a.patterns {
		patterns = append(patterns, *p)
	}

	deps := make([]string, 0, len(a.deps))
	for d := range a.deps {
		deps = append(deps, d)
	}

	totalFreq := 0
	for _, p := range patterns {
		totalFreq += p.Frequency
	}

	return &AnalysisResult{
		Patterns:     patterns,
		Dependencies: deps,
		Complexity:   float64(len(patterns)) * 0.1,
		LinesOfCode:  totalFreq * 100, // rough estimate
	}
}

// HasPattern checks if a specific pattern was detected.
func (a *Analyzer) HasPattern(name string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.patterns[strings.ToLower(name)]
	if ok {
		return true
	}
	// Also check case-insensitive
	for k := range a.patterns {
		if strings.EqualFold(k, name) {
			return true
		}
	}
	return false
}

// PatternCount returns the number of unique patterns.
func (a *Analyzer) PatternCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.patterns)
}
