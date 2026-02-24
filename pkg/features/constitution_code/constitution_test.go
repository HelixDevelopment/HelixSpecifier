package constitution_code

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConstitution(t *testing.T) {
	c := NewConstitution()
	require.NotNil(t, c)
	assert.Equal(t, 0, c.RuleCount())
}

func TestConstitution_AddRule_Success(t *testing.T) {
	c := NewConstitution()
	err := c.AddRule(Rule{
		ID:        "r1",
		Title:     "Test Coverage",
		Category:  "testing",
		Priority:  1,
		Mandatory: true,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, c.RuleCount())
}

func TestConstitution_AddRule_EmptyID(t *testing.T) {
	c := NewConstitution()
	err := c.AddRule(Rule{Title: "No ID"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID")
}

func TestConstitution_AddRule_Duplicate(t *testing.T) {
	c := NewConstitution()
	_ = c.AddRule(Rule{ID: "r1", Title: "First"})
	err := c.AddRule(Rule{ID: "r1", Title: "Second"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestConstitution_GetRule(t *testing.T) {
	c := NewConstitution()
	_ = c.AddRule(Rule{
		ID:    "r1",
		Title: "Security",
	})
	r, err := c.GetRule("r1")
	require.NoError(t, err)
	assert.Equal(t, "Security", r.Title)
}

func TestConstitution_GetRule_NotFound(t *testing.T) {
	c := NewConstitution()
	_, err := c.GetRule("missing")
	require.Error(t, err)
}

func TestConstitution_MandatoryRules(t *testing.T) {
	c := NewConstitution()
	_ = c.AddRule(Rule{
		ID: "r1", Title: "A", Mandatory: true,
	})
	_ = c.AddRule(Rule{
		ID: "r2", Title: "B", Mandatory: false,
	})
	_ = c.AddRule(Rule{
		ID: "r3", Title: "C", Mandatory: true,
	})

	mandatory := c.MandatoryRules()
	assert.Len(t, mandatory, 2)
}

func TestConstitution_Validate(t *testing.T) {
	c := NewConstitution()
	_ = c.AddRule(Rule{
		ID: "r1", Title: "Testing", Mandatory: true,
	})
	_ = c.AddRule(Rule{
		ID: "r2", Title: "Security", Mandatory: true,
	})

	violations := c.Validate("we cover testing fully")
	assert.Len(t, violations, 1)
	assert.Contains(t, violations, "r2")
}
