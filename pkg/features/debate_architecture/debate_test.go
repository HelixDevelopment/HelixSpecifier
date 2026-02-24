package debate_architecture

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDebate(t *testing.T) {
	d := NewDebate("d1", "architecture review", 5)
	require.NotNil(t, d)
	assert.Equal(t, "d1", d.ID())
	assert.Equal(t, "architecture review", d.Topic())
	assert.Equal(t, 0, d.RoundCount())
	assert.False(t, d.IsComplete())
}

func TestNewDebate_DefaultRounds(t *testing.T) {
	d := NewDebate("d2", "topic", 0)
	assert.False(t, d.IsComplete())
	// Default maxRounds should be 3
	for i := 0; i < 3; i++ {
		_ = d.AddRound(Round{})
	}
	assert.True(t, d.IsComplete())
}

func TestDebate_AddRound(t *testing.T) {
	d := NewDebate("d3", "topic", 2)

	err := d.AddRound(Round{
		Positions: []Position{
			{AgentID: "a1", Score: 0.8},
		},
		Duration: 1 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, d.RoundCount())
}

func TestDebate_AddRound_MaxReached(t *testing.T) {
	d := NewDebate("d4", "topic", 1)
	_ = d.AddRound(Round{})
	err := d.AddRound(Round{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max rounds")
}

func TestDebate_BestPosition(t *testing.T) {
	d := NewDebate("d5", "topic", 3)
	_ = d.AddRound(Round{
		Positions: []Position{
			{AgentID: "a1", Score: 0.7},
			{AgentID: "a2", Score: 0.9},
		},
	})
	_ = d.AddRound(Round{
		Positions: []Position{
			{AgentID: "a1", Score: 0.85},
		},
	})

	best := d.BestPosition()
	require.NotNil(t, best)
	assert.Equal(t, "a2", best.AgentID)
	assert.InDelta(t, 0.9, best.Score, 0.001)
}

func TestDebate_BestPosition_Empty(t *testing.T) {
	d := NewDebate("d6", "topic", 3)
	assert.Nil(t, d.BestPosition())
}
