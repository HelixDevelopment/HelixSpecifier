package brownfield

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnalyzer(t *testing.T) {
	a := NewAnalyzer()
	require.NotNil(t, a)
	assert.Equal(t, 0, a.PatternCount())
}

func TestAnalyzer_RecordPattern(t *testing.T) {
	a := NewAnalyzer()
	err := a.RecordPattern("singleton", 0.9)
	require.NoError(t, err)
	assert.Equal(t, 1, a.PatternCount())
}

func TestAnalyzer_RecordPattern_EmptyName(t *testing.T) {
	a := NewAnalyzer()
	err := a.RecordPattern("", 0.5)
	require.Error(t, err)
}

func TestAnalyzer_RecordPattern_Duplicate(t *testing.T) {
	a := NewAnalyzer()
	_ = a.RecordPattern("factory", 0.8)
	_ = a.RecordPattern("factory", 0.9)
	assert.Equal(t, 1, a.PatternCount())

	result := a.Analyze()
	for _, p := range result.Patterns {
		if p.Name == "factory" {
			assert.Equal(t, 2, p.Frequency)
		}
	}
}

func TestAnalyzer_AddDependency(t *testing.T) {
	a := NewAnalyzer()
	a.AddDependency("github.com/gin-gonic/gin")
	a.AddDependency("github.com/stretchr/testify")

	result := a.Analyze()
	assert.Len(t, result.Dependencies, 2)
}

func TestAnalyzer_Analyze(t *testing.T) {
	a := NewAnalyzer()
	_ = a.RecordPattern("observer", 0.85)
	_ = a.RecordPattern("strategy", 0.75)
	a.AddDependency("dep1")

	result := a.Analyze()
	assert.Len(t, result.Patterns, 2)
	assert.Len(t, result.Dependencies, 1)
	assert.Greater(t, result.Complexity, 0.0)
}

func TestAnalyzer_HasPattern(t *testing.T) {
	a := NewAnalyzer()
	_ = a.RecordPattern("Facade", 0.8)
	assert.True(t, a.HasPattern("Facade"))
	assert.True(t, a.HasPattern("facade"))
	assert.False(t, a.HasPattern("missing"))
}

func TestAnalyzer_HasPattern_CaseInsensitiveLoop(
	t *testing.T,
) {
	a := NewAnalyzer()
	// Store with mixed case -- ToLower("FACADE") = "facade"
	// won't match "Facade" key directly, falls to loop
	_ = a.RecordPattern("Facade", 0.85)
	assert.True(t, a.HasPattern("FACADE"))
}

func TestAnalyzer_HasPattern_ExactLowerMatch(t *testing.T) {
	a := NewAnalyzer()
	// Store with lowercase key -- direct map lookup succeeds
	_ = a.RecordPattern("observer", 0.7)
	assert.True(t, a.HasPattern("observer"))
	assert.True(t, a.HasPattern("Observer"))
	assert.True(t, a.HasPattern("OBSERVER"))
}

func TestAnalyzer_HasPattern_Empty(t *testing.T) {
	a := NewAnalyzer()
	assert.False(t, a.HasPattern("anything"))
}
