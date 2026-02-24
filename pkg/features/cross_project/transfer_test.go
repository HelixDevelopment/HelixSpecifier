package cross_project

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTransferEngine(t *testing.T) {
	te := NewTransferEngine()
	require.NotNil(t, te)
	assert.Equal(t, 0, te.Count())
}

func TestTransferEngine_Add(t *testing.T) {
	te := NewTransferEngine()
	err := te.Add(Knowledge{
		ID:       "k1",
		Project:  "ProjectA",
		Category: "architecture",
		Content:  "Use hexagonal architecture",
		Tags:     []string{"arch", "patterns"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, te.Count())
}

func TestTransferEngine_Add_EmptyID(t *testing.T) {
	te := NewTransferEngine()
	err := te.Add(Knowledge{Project: "P"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID")
}

func TestTransferEngine_Add_EmptyProject(t *testing.T) {
	te := NewTransferEngine()
	err := te.Add(Knowledge{ID: "k1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
}

func TestTransferEngine_Search_ByContent(t *testing.T) {
	te := NewTransferEngine()
	_ = te.Add(Knowledge{
		ID:      "k1",
		Project: "P1",
		Content: "Use PostgreSQL for persistence",
	})
	_ = te.Add(Knowledge{
		ID:      "k2",
		Project: "P2",
		Content: "Redis caching layer",
	})

	results := te.Search("postgresql", 10)
	assert.Len(t, results, 1)
	assert.Equal(t, "k1", results[0].ID)
}

func TestTransferEngine_Search_ByTag(t *testing.T) {
	te := NewTransferEngine()
	_ = te.Add(Knowledge{
		ID:      "k1",
		Project: "P1",
		Content: "some content",
		Tags:    []string{"database", "sql"},
	})

	results := te.Search("database", 10)
	assert.Len(t, results, 1)
}

func TestTransferEngine_ForProject(t *testing.T) {
	te := NewTransferEngine()
	_ = te.Add(Knowledge{
		ID: "k1", Project: "Alpha", Content: "a",
	})
	_ = te.Add(Knowledge{
		ID: "k2", Project: "Beta", Content: "b",
	})
	_ = te.Add(Knowledge{
		ID: "k3", Project: "Alpha", Content: "c",
	})

	alpha := te.ForProject("Alpha")
	assert.Len(t, alpha, 2)
}
