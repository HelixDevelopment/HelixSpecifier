// Package spec_memory provides persistent specification
// memory with semantic indexing for rapid retrieval of
// historical specifications and design decisions.
package spec_memory

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Entry represents a single specification memory entry.
type Entry struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Tags        []string          `json:"tags"`
	Score       float64           `json:"score"`
	CreatedAt   time.Time         `json:"created_at"`
	AccessCount int               `json:"access_count"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Index provides in-memory specification indexing and
// retrieval with access tracking.
type Index struct {
	entries map[string]*Entry
	mu      sync.RWMutex
}

// NewIndex creates a new spec memory index.
func NewIndex() *Index {
	return &Index{
		entries: make(map[string]*Entry),
	}
}

// Add stores an entry in the index.
func (idx *Index) Add(entry Entry) error {
	if entry.ID == "" {
		return fmt.Errorf("entry ID cannot be empty")
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	idx.entries[entry.ID] = &entry
	return nil
}

// Get retrieves an entry by ID and increments access count.
func (idx *Index) Get(id string) (*Entry, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	e, ok := idx.entries[id]
	if !ok {
		return nil, fmt.Errorf("entry %s not found", id)
	}
	e.AccessCount++
	cp := *e
	return &cp, nil
}

// Search finds entries whose title or content matches the
// query string, returning up to limit results.
func (idx *Index) Search(
	query string,
	limit int,
) []Entry {
	if limit <= 0 {
		limit = 10
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	lower := strings.ToLower(query)
	results := []Entry{}

	for _, e := range idx.entries {
		if strings.Contains(
			strings.ToLower(e.Title), lower,
		) || strings.Contains(
			strings.ToLower(e.Content), lower,
		) {
			results = append(results, *e)
			if len(results) >= limit {
				break
			}
		}
	}
	return results
}

// Count returns the total number of entries.
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.entries)
}

// Remove deletes an entry by ID.
func (idx *Index) Remove(id string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, ok := idx.entries[id]; !ok {
		return fmt.Errorf("entry %s not found", id)
	}
	delete(idx.entries, id)
	return nil
}

// MostAccessed returns the entry with the highest access
// count, or nil if the index is empty.
func (idx *Index) MostAccessed() *Entry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var best *Entry
	for _, e := range idx.entries {
		if best == nil ||
			e.AccessCount > best.AccessCount {
			best = e
		}
	}
	if best == nil {
		return nil
	}
	cp := *best
	return &cp
}
