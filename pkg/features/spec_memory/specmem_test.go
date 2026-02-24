package spec_memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIndex(t *testing.T) {
	idx := NewIndex()
	require.NotNil(t, idx)
	assert.Equal(t, 0, idx.Count())
}

func TestIndex_Add(t *testing.T) {
	idx := NewIndex()
	err := idx.Add(Entry{
		ID:      "e1",
		Title:   "Auth Spec",
		Content: "Authentication specification",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, idx.Count())
}

func TestIndex_Add_EmptyID(t *testing.T) {
	idx := NewIndex()
	err := idx.Add(Entry{Title: "no id"})
	require.Error(t, err)
}

func TestIndex_Get(t *testing.T) {
	idx := NewIndex()
	_ = idx.Add(Entry{ID: "e1", Title: "Test"})

	entry, err := idx.Get("e1")
	require.NoError(t, err)
	assert.Equal(t, "Test", entry.Title)
	assert.Equal(t, 1, entry.AccessCount)
}

func TestIndex_Get_NotFound(t *testing.T) {
	idx := NewIndex()
	_, err := idx.Get("missing")
	require.Error(t, err)
}

func TestIndex_Search(t *testing.T) {
	idx := NewIndex()
	_ = idx.Add(Entry{
		ID:      "e1",
		Title:   "Database Design",
		Content: "PostgreSQL schema",
	})
	_ = idx.Add(Entry{
		ID:      "e2",
		Title:   "API Design",
		Content: "REST endpoints",
	})

	results := idx.Search("database", 10)
	assert.Len(t, results, 1)
	assert.Equal(t, "e1", results[0].ID)
}

func TestIndex_Remove(t *testing.T) {
	idx := NewIndex()
	_ = idx.Add(Entry{ID: "e1", Title: "A"})
	assert.Equal(t, 1, idx.Count())

	err := idx.Remove("e1")
	require.NoError(t, err)
	assert.Equal(t, 0, idx.Count())
}

func TestIndex_Remove_NotFound(t *testing.T) {
	idx := NewIndex()
	err := idx.Remove("missing")
	require.Error(t, err)
}

func TestIndex_MostAccessed(t *testing.T) {
	idx := NewIndex()
	_ = idx.Add(Entry{ID: "e1", Title: "A"})
	_ = idx.Add(Entry{ID: "e2", Title: "B"})

	// Access e2 three times
	_, _ = idx.Get("e2")
	_, _ = idx.Get("e2")
	_, _ = idx.Get("e2")
	// Access e1 once
	_, _ = idx.Get("e1")

	best := idx.MostAccessed()
	require.NotNil(t, best)
	assert.Equal(t, "e2", best.ID)
	assert.Equal(t, 3, best.AccessCount)
}

func TestIndex_MostAccessed_Empty(t *testing.T) {
	idx := NewIndex()
	assert.Nil(t, idx.MostAccessed())
}
