package intent

import (
	"strings"
	"testing"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClassifier(t *testing.T) {
	c := NewClassifier()
	require.NotNil(t, c)
}

func TestClassifier_Classify_Quick(t *testing.T) {
	c := NewClassifier()
	result := c.Classify("fix typo in readme")
	assert.Equal(t, types.EffortQuick, result.Level)
	assert.GreaterOrEqual(t, result.Confidence, 0.8)
}

func TestClassifier_Classify_Medium(t *testing.T) {
	c := NewClassifier()
	result := c.Classify("update the config")
	assert.Equal(t, types.EffortMedium, result.Level)
}

func TestClassifier_Classify_Large(t *testing.T) {
	c := NewClassifier()
	result := c.Classify(
		"implement new authentication service for the api",
	)
	assert.Equal(t, types.EffortLarge, result.Level)
	assert.GreaterOrEqual(t, result.Confidence, 0.6)
}

func TestClassifier_Classify_Epic(t *testing.T) {
	c := NewClassifier()
	result := c.Classify(
		"complete rewrite of the entire system",
	)
	assert.Equal(t, types.EffortEpic, result.Level)
	assert.GreaterOrEqual(t, result.Confidence, 0.7)
}

func TestClassifier_Classify_Refactor(t *testing.T) {
	c := NewClassifier()
	result := c.Classify("refactor the database layer")
	assert.Equal(t, types.EffortLarge, result.Level)
	found := false
	for _, s := range result.Signals {
		if strings.HasPrefix(s, "refactor:") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected refactor signal")
}

func TestClassifier_Classify_LongRequest(t *testing.T) {
	c := NewClassifier()
	// Generate a request longer than 200 characters
	long := strings.Repeat("do something useful ", 15)
	require.Greater(t, len(long), 200)
	result := c.Classify(long)
	assert.Equal(t, types.EffortLarge, result.Level)
}

func TestClassifier_Classify_CeremonyLevel(t *testing.T) {
	c := NewClassifier()

	quick := c.Classify("fix typo")
	assert.Equal(t, types.CeremonyMinimal,
		quick.CeremonyLevel)

	medium := c.Classify("update the config")
	assert.Equal(t, types.CeremonyLight,
		medium.CeremonyLevel)

	large := c.Classify(
		"implement new build system and deploy service",
	)
	assert.Equal(t, types.CeremonyStandard,
		large.CeremonyLevel)

	epic := c.Classify("rebuild all from scratch")
	assert.Equal(t, types.CeremonyHeavy,
		epic.CeremonyLevel)
}

func TestClassifier_Classify_RequiresDebate(t *testing.T) {
	c := NewClassifier()

	quick := c.Classify("fix typo")
	assert.False(t, quick.RequiresDebate)

	medium := c.Classify("update the config")
	assert.True(t, medium.RequiresDebate)

	large := c.Classify(
		"implement new authentication service for the api",
	)
	assert.True(t, large.RequiresDebate)
}

func TestClassifier_Classify_RequiresSpecKit(t *testing.T) {
	c := NewClassifier()

	quick := c.Classify("fix typo")
	assert.False(t, quick.RequiresSpecKit)

	medium := c.Classify("update the config")
	assert.False(t, medium.RequiresSpecKit)

	large := c.Classify(
		"implement new authentication service for api",
	)
	assert.True(t, large.RequiresSpecKit)

	epic := c.Classify("complete rewrite of entire system")
	assert.True(t, epic.RequiresSpecKit)
}

func TestClassifier_Classify_Signals(t *testing.T) {
	c := NewClassifier()

	result := c.Classify("complete rewrite of entire system")
	require.NotEmpty(t, result.Signals)
	found := false
	for _, s := range result.Signals {
		if strings.HasPrefix(s, "epic:") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected epic signal")
}

// TestClassifier_Classify_Reasoning asserts the i18n keys
// emitted by generateReasoning post-CONST-046 migration. Under
// the default NoopTranslator the keys round-trip verbatim, so
// asserting them here both validates the migration (every level
// resolves to its dedicated key) and pins the namespace prefix
// per CONST-046 — a regression that hardcodes English would
// fail this test because NoopTranslator never returns the
// underlying bundle value.
func TestClassifier_Classify_Reasoning(t *testing.T) {
	c := NewClassifier()

	quick := c.Classify("fix typo")
	assert.Equal(
		t,
		"helixspecifier_intent_quick_reasoning",
		quick.Reasoning,
	)

	medium := c.Classify("update the config")
	assert.Equal(
		t,
		"helixspecifier_intent_medium_reasoning",
		medium.Reasoning,
	)

	large := c.Classify(
		"implement new authentication service for api",
	)
	assert.Equal(
		t,
		"helixspecifier_intent_large_reasoning",
		large.Reasoning,
	)

	epic := c.Classify("rebuild all from scratch")
	assert.Equal(
		t,
		"helixspecifier_intent_epic_reasoning",
		epic.Reasoning,
	)
}

func TestClassifier_Classify_MultipleSignals(t *testing.T) {
	c := NewClassifier()
	result := c.Classify(
		"implement new authentication and authorization",
	)
	assert.Greater(t, len(result.Signals), 1)
}

func TestClassifier_Classify_EmptyRequest(t *testing.T) {
	c := NewClassifier()
	result := c.Classify("")
	assert.Equal(t, types.EffortMedium, result.Level)
	assert.Empty(t, result.Signals)
}

// TestClassifier_GenerateReasoning_UnknownLevel asserts the
// unknown-effort branch resolves to its i18n key (CONST-046).
// Pre-migration this asserted the literal English string; the
// key now wins via NoopTranslator so a regression that
// reintroduces a hardcoded literal is caught.
func TestClassifier_GenerateReasoning_UnknownLevel(
	t *testing.T,
) {
	c := NewClassifier()
	// Directly invoke generateReasoning with an unknown level
	cl := &types.EffortClassification{
		Level: types.EffortLevel("unknown"),
	}
	reasoning := c.generateReasoning(cl)
	assert.Equal(
		t,
		"helixspecifier_intent_unknown_reasoning",
		reasoning,
	)
}

// TestClassifier_WithTranslator_InjectsAlternate proves the
// translator-injection constructor wires through end-to-end:
// reasoning strings come from the injected backend, not from
// the noop fallback. This is the anti-bluff invariant for the
// i18n call site — a regression that hardcodes English literals
// would render this test inert because the injected stub would
// never be consulted.
func TestClassifier_WithTranslator_InjectsAlternate(
	t *testing.T,
) {
	stub := &recordingTranslator{
		responses: map[string]string{
			"helixspecifier_intent_quick_reasoning": "QK",
			"helixspecifier_intent_medium_reasoning": "MD",
			"helixspecifier_intent_large_reasoning": "LG",
			"helixspecifier_intent_epic_reasoning": "EP",
			"helixspecifier_intent_unknown_reasoning": "UN",
		},
	}
	c := NewClassifierWithTranslator(stub)
	require.NotNil(t, c)

	got := c.Classify("fix typo")
	assert.Equal(t, "QK", got.Reasoning)
	assert.Contains(
		t, stub.seen,
		"helixspecifier_intent_quick_reasoning",
	)
}

// TestClassifier_NewWithNilTranslator_FallsBackToNoop ensures
// constructor robustness — a nil injection lands on the noop
// fallback rather than crashing later inside generateReasoning.
func TestClassifier_NewWithNilTranslator_FallsBackToNoop(
	t *testing.T,
) {
	c := NewClassifierWithTranslator(nil)
	require.NotNil(t, c)
	got := c.Classify("fix typo")
	assert.Equal(
		t,
		"helixspecifier_intent_quick_reasoning",
		got.Reasoning,
	)
}

// recordingTranslator is a unit-test-only translator that maps
// keys to canned responses and records every lookup. Per
// CONST-050(A) mocks/stubs are permitted only in unit-test
// sources; this struct lives inside *_test.go and is never
// imported by production code.
type recordingTranslator struct {
	responses map[string]string
	seen      []string
}

func (r *recordingTranslator) T(
	key string, args ...any,
) string {
	r.seen = append(r.seen, key)
	_ = args
	if v, ok := r.responses[key]; ok {
		return v
	}
	return key
}

func TestClassifier_Classify_TableDriven(t *testing.T) {
	c := NewClassifier()

	tests := []struct {
		name     string
		request  string
		expected types.EffortLevel
	}{
		{
			name:     "empty_string",
			request:  "",
			expected: types.EffortMedium,
		},
		{
			name:     "simple_rename",
			request:  "rename the variable",
			expected: types.EffortQuick,
		},
		{
			name:     "bump_version",
			request:  "bump version to 2.0",
			expected: types.EffortQuick,
		},
		{
			name:     "add_log",
			request:  "add log to handler",
			expected: types.EffortQuick,
		},
		{
			name:     "generic_change",
			request:  "update configuration",
			expected: types.EffortMedium,
		},
		{
			name:     "implement_and_build",
			request:  "implement and build the new api",
			expected: types.EffortLarge,
		},
		{
			name:     "create_module_service",
			request:  "create module for service",
			expected: types.EffortLarge,
		},
		{
			name:     "migrate_refactor",
			request:  "migrate old data layer",
			expected: types.EffortLarge,
		},
		{
			name:     "entire_system_from_scratch",
			request:  "build entire system from scratch",
			expected: types.EffortEpic,
		},
		{
			name:     "full_platform_rewrite",
			request:  "full platform rebuild all services",
			expected: types.EffortEpic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Classify(tt.request)
			assert.Equal(t, tt.expected, result.Level,
				"request: %q", tt.request)
		})
	}
}
