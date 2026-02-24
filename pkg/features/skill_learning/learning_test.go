package skill_learning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLearner(t *testing.T) {
	l := NewLearner()
	require.NotNil(t, l)
	assert.Equal(t, 0, l.SkillCount())
}

func TestLearner_RegisterSkill(t *testing.T) {
	l := NewLearner()
	err := l.RegisterSkill(Skill{
		ID:     "s1",
		Name:   "Go Programming",
		Domain: "engineering",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, l.SkillCount())
}

func TestLearner_RegisterSkill_EmptyID(t *testing.T) {
	l := NewLearner()
	err := l.RegisterSkill(Skill{Name: "no id"})
	require.Error(t, err)
}

func TestLearner_RegisterSkill_Duplicate(t *testing.T) {
	l := NewLearner()
	_ = l.RegisterSkill(Skill{ID: "s1", Name: "A"})
	err := l.RegisterSkill(Skill{ID: "s1", Name: "B"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestLearner_RecordUsage(t *testing.T) {
	l := NewLearner()
	_ = l.RegisterSkill(Skill{
		ID: "s1", Name: "Code Review",
	})

	err := l.RecordUsage("s1", 0.8)
	require.NoError(t, err)

	skill, _ := l.GetSkill("s1")
	assert.InDelta(t, 0.8, skill.Proficiency, 0.001)
	assert.Equal(t, 1, skill.UseCount)
}

func TestLearner_RecordUsage_NotFound(t *testing.T) {
	l := NewLearner()
	err := l.RecordUsage("missing", 0.5)
	require.Error(t, err)
}

func TestLearner_GetSkill_NotFound(t *testing.T) {
	l := NewLearner()
	skill, err := l.GetSkill("nonexistent")
	require.Error(t, err)
	assert.Nil(t, skill)
	assert.Contains(t, err.Error(), "not found")
}

func TestLearner_TopSkills(t *testing.T) {
	l := NewLearner()
	_ = l.RegisterSkill(Skill{
		ID: "s1", Name: "A", Proficiency: 0.6,
	})
	_ = l.RegisterSkill(Skill{
		ID: "s2", Name: "B", Proficiency: 0.9,
	})
	_ = l.RegisterSkill(Skill{
		ID: "s3", Name: "C", Proficiency: 0.75,
	})

	top := l.TopSkills(2)
	assert.Len(t, top, 2)
	assert.Equal(t, "s2", top[0].ID)
}

func TestLearner_TopSkills_LimitExceedsCount(t *testing.T) {
	l := NewLearner()
	_ = l.RegisterSkill(Skill{
		ID: "s1", Name: "A", Proficiency: 0.6,
	})
	_ = l.RegisterSkill(Skill{
		ID: "s2", Name: "B", Proficiency: 0.9,
	})

	// Requesting more than available
	top := l.TopSkills(10)
	assert.Len(t, top, 2)
	assert.Equal(t, "s2", top[0].ID)
}

func TestLearner_TopSkills_Empty(t *testing.T) {
	l := NewLearner()
	top := l.TopSkills(5)
	assert.Empty(t, top)
}
