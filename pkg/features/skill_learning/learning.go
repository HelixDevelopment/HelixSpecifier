// Package skill_learning provides adaptive skill acquisition
// capabilities allowing the engine to improve its specification
// quality over time based on feedback.
package skill_learning

import (
	"fmt"
	"sync"
	"time"
)

// Skill represents a learned skill with proficiency tracking.
type Skill struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	Proficiency float64   `json:"proficiency"`
	UseCount    int       `json:"use_count"`
	LastUsed    time.Time `json:"last_used"`
}

// Learner tracks and improves skills based on feedback.
type Learner struct {
	skills map[string]*Skill
	mu     sync.RWMutex
}

// NewLearner creates a new skill learner.
func NewLearner() *Learner {
	return &Learner{
		skills: make(map[string]*Skill),
	}
}

// RegisterSkill registers a new skill.
func (l *Learner) RegisterSkill(skill Skill) error {
	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.skills[skill.ID]; exists {
		return fmt.Errorf(
			"skill %s already registered", skill.ID,
		)
	}

	l.skills[skill.ID] = &skill
	return nil
}

// RecordUsage records a skill usage with a quality score.
// The proficiency is updated as a running average.
func (l *Learner) RecordUsage(
	skillID string,
	qualityScore float64,
) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	skill, ok := l.skills[skillID]
	if !ok {
		return fmt.Errorf("skill %s not found", skillID)
	}

	skill.UseCount++
	skill.LastUsed = time.Now()

	// Running average: new = old + (score - old) / count
	skill.Proficiency += (qualityScore -
		skill.Proficiency) / float64(skill.UseCount)

	return nil
}

// GetSkill returns a skill by ID.
func (l *Learner) GetSkill(id string) (*Skill, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	skill, ok := l.skills[id]
	if !ok {
		return nil, fmt.Errorf("skill %s not found", id)
	}
	cp := *skill
	return &cp, nil
}

// SkillCount returns the number of registered skills.
func (l *Learner) SkillCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.skills)
}

// TopSkills returns skills sorted by proficiency descending,
// up to the given limit.
func (l *Learner) TopSkills(limit int) []Skill {
	l.mu.RLock()
	defer l.mu.RUnlock()

	all := make([]Skill, 0, len(l.skills))
	for _, s := range l.skills {
		all = append(all, *s)
	}

	// Simple insertion sort for small N
	for i := 1; i < len(all); i++ {
		j := i
		for j > 0 &&
			all[j].Proficiency > all[j-1].Proficiency {
			all[j], all[j-1] = all[j-1], all[j]
			j--
		}
	}

	if limit > len(all) {
		limit = len(all)
	}
	return all[:limit]
}
