// Package constitution_code manages project constitution
// documents that define non-negotiable rules and standards.
package constitution_code

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Rule represents a single constitution rule.
type Rule struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"`
	Priority    int       `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Mandatory   bool      `json:"mandatory"`
	CreatedAt   time.Time `json:"created_at"`
}

// Constitution manages project rules and standards.
type Constitution struct {
	rules []Rule
	mu    sync.RWMutex
}

// NewConstitution creates a new empty constitution.
func NewConstitution() *Constitution {
	return &Constitution{
		rules: []Rule{},
	}
}

// AddRule adds a rule to the constitution.
func (c *Constitution) AddRule(rule Rule) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	if rule.Title == "" {
		return fmt.Errorf("rule title cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range c.rules {
		if r.ID == rule.ID {
			return fmt.Errorf(
				"rule %s already exists", rule.ID,
			)
		}
	}

	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	c.rules = append(c.rules, rule)
	return nil
}

// GetRule returns a rule by ID.
func (c *Constitution) GetRule(id string) (*Rule, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, r := range c.rules {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("rule %s not found", id)
}

// RuleCount returns the number of rules.
func (c *Constitution) RuleCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.rules)
}

// MandatoryRules returns all mandatory rules.
func (c *Constitution) MandatoryRules() []Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := []Rule{}
	for _, r := range c.rules {
		if r.Mandatory {
			result = append(result, r)
		}
	}
	return result
}

// Validate checks if text complies with mandatory rules.
// Returns list of violated rule IDs.
func (c *Constitution) Validate(text string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	violations := []string{}
	lower := strings.ToLower(text)

	for _, r := range c.rules {
		if !r.Mandatory {
			continue
		}
		// Check that each mandatory rule keyword
		// is mentioned in the text
		keyword := strings.ToLower(r.Title)
		if !strings.Contains(lower, keyword) {
			violations = append(violations, r.ID)
		}
	}
	return violations
}
