// Package debate_architecture provides multi-LLM debate
// orchestration for specification refinement. Debates allow
// multiple AI agents to argue positions, critique, and
// synthesize specifications.
package debate_architecture

import (
	"fmt"
	"sync"
	"time"
)

// Position represents a debate position.
type Position struct {
	AgentID   string    `json:"agent_id"`
	Argument  string    `json:"argument"`
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// Round represents a single debate round.
type Round struct {
	Number    int        `json:"number"`
	Positions []Position `json:"positions"`
	Winner    string     `json:"winner"`
	Duration  time.Duration `json:"duration"`
}

// Debate orchestrates multi-round specification debates.
type Debate struct {
	id        string
	topic     string
	rounds    []Round
	maxRounds int
	mu        sync.RWMutex
}

// NewDebate creates a new debate instance.
func NewDebate(
	id string,
	topic string,
	maxRounds int,
) *Debate {
	if maxRounds <= 0 {
		maxRounds = 3
	}
	return &Debate{
		id:        id,
		topic:     topic,
		rounds:    []Round{},
		maxRounds: maxRounds,
	}
}

// AddRound records a completed debate round.
func (d *Debate) AddRound(round Round) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.rounds) >= d.maxRounds {
		return fmt.Errorf(
			"max rounds (%d) reached", d.maxRounds,
		)
	}

	round.Number = len(d.rounds) + 1
	d.rounds = append(d.rounds, round)
	return nil
}

// RoundCount returns the number of completed rounds.
func (d *Debate) RoundCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.rounds)
}

// IsComplete returns true if all rounds are done.
func (d *Debate) IsComplete() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.rounds) >= d.maxRounds
}

// ID returns the debate identifier.
func (d *Debate) ID() string {
	return d.id
}

// Topic returns the debate topic.
func (d *Debate) Topic() string {
	return d.topic
}

// BestPosition returns the highest-scored position across
// all rounds. Returns nil if no rounds recorded.
func (d *Debate) BestPosition() *Position {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var best *Position
	for _, r := range d.rounds {
		for i := range r.Positions {
			p := &r.Positions[i]
			if best == nil || p.Score > best.Score {
				best = p
			}
		}
	}
	return best
}
