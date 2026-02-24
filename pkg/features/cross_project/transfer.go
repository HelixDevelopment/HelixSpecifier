// Package cross_project enables knowledge transfer between
// projects by maintaining a shared knowledge base of patterns,
// decisions, and lessons learned.
package cross_project

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Knowledge represents a transferable piece of knowledge.
type Knowledge struct {
	ID          string    `json:"id"`
	Project     string    `json:"project"`
	Category    string    `json:"category"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Relevance   float64   `json:"relevance"`
	TransferAt  time.Time `json:"transfer_at"`
}

// TransferEngine manages cross-project knowledge transfer.
type TransferEngine struct {
	knowledge []Knowledge
	mu        sync.RWMutex
}

// NewTransferEngine creates a new transfer engine.
func NewTransferEngine() *TransferEngine {
	return &TransferEngine{
		knowledge: []Knowledge{},
	}
}

// Add records a piece of knowledge from a project.
func (te *TransferEngine) Add(k Knowledge) error {
	if k.ID == "" {
		return fmt.Errorf("knowledge ID cannot be empty")
	}
	if k.Project == "" {
		return fmt.Errorf("project cannot be empty")
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	k.TransferAt = time.Now()
	te.knowledge = append(te.knowledge, k)
	return nil
}

// Search finds knowledge entries matching a query string
// against content and tags.
func (te *TransferEngine) Search(
	query string,
	limit int,
) []Knowledge {
	if limit <= 0 {
		limit = 10
	}

	te.mu.RLock()
	defer te.mu.RUnlock()

	lower := strings.ToLower(query)
	results := []Knowledge{}

	for _, k := range te.knowledge {
		if strings.Contains(
			strings.ToLower(k.Content), lower,
		) {
			results = append(results, k)
			if len(results) >= limit {
				break
			}
			continue
		}
		for _, tag := range k.Tags {
			if strings.Contains(
				strings.ToLower(tag), lower,
			) {
				results = append(results, k)
				if len(results) >= limit {
					break
				}
				break
			}
		}
		if len(results) >= limit {
			break
		}
	}

	return results
}

// ForProject returns all knowledge from a specific project.
func (te *TransferEngine) ForProject(
	project string,
) []Knowledge {
	te.mu.RLock()
	defer te.mu.RUnlock()

	results := []Knowledge{}
	for _, k := range te.knowledge {
		if k.Project == project {
			results = append(results, k)
		}
	}
	return results
}

// Count returns the total number of knowledge entries.
func (te *TransferEngine) Count() int {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return len(te.knowledge)
}
