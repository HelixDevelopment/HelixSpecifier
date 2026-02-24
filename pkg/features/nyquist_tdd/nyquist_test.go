package nyquist_tdd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTracker(t *testing.T) {
	tr := NewTracker()
	require.NotNil(t, tr)
	assert.Equal(t, 0.0, tr.Ratio())
}

func TestTracker_RecordAndRatio(t *testing.T) {
	tr := NewTracker()
	tr.RecordImplementation(100)
	tr.RecordTests(200)
	assert.InDelta(t, 2.0, tr.Ratio(), 0.001)
}

func TestTracker_IsCompliant_True(t *testing.T) {
	tr := NewTracker()
	tr.RecordImplementation(50)
	tr.RecordTests(150)
	assert.True(t, tr.IsCompliant())
}

func TestTracker_IsCompliant_False(t *testing.T) {
	tr := NewTracker()
	tr.RecordImplementation(100)
	tr.RecordTests(50)
	assert.False(t, tr.IsCompliant())
}

func TestTracker_Check_Pass(t *testing.T) {
	tr := NewTracker()
	tr.RecordImplementation(10)
	tr.RecordTests(25)
	err := tr.Check()
	assert.NoError(t, err)
	assert.Empty(t, tr.Violations())
}

func TestTracker_Check_Fail(t *testing.T) {
	tr := NewTracker()
	tr.RecordImplementation(100)
	tr.RecordTests(50)
	err := tr.Check()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Nyquist")
	assert.Len(t, tr.Violations(), 1)
}

func TestTracker_Stats(t *testing.T) {
	tr := NewTracker()
	tr.RecordImplementation(42)
	tr.RecordTests(84)
	impl, tests := tr.Stats()
	assert.Equal(t, 42, impl)
	assert.Equal(t, 84, tests)
}
