package config

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()

	// Effort classification
	assert.Equal(t, 5, cfg.QuickMaxMinutes)
	assert.Equal(t, 90, cfg.MediumMaxMinutes)
	assert.Equal(t, 16, cfg.LargeMaxHours)

	// Ceremony scaling
	assert.True(t, cfg.AdaptiveCeremony)
	assert.Equal(t, 0.0, cfg.CeremonyBoost)

	// Debate rounds
	assert.Equal(t, 5, cfg.ConstitutionRounds)
	assert.Equal(t, 3, cfg.SpecifyRounds)
	assert.Equal(t, 3, cfg.ClarifyRounds)
	assert.Equal(t, 4, cfg.PlanRounds)
	assert.Equal(t, 2, cfg.TasksRounds)
	assert.Equal(t, 4, cfg.AnalyzeRounds)
	assert.Equal(t, 6, cfg.ImplementRounds)

	// Phase timeouts
	assert.Equal(t, 15*time.Minute, cfg.ConstitutionTimeout)
	assert.Equal(t, 10*time.Minute, cfg.SpecifyTimeout)
	assert.Equal(t, 10*time.Minute, cfg.ClarifyTimeout)
	assert.Equal(t, 12*time.Minute, cfg.PlanTimeout)
	assert.Equal(t, 8*time.Minute, cfg.TasksTimeout)
	assert.Equal(t, 12*time.Minute, cfg.AnalyzeTimeout)
	assert.Equal(t, 20*time.Minute, cfg.ImplementTimeout)

	// Caching
	assert.True(t, cfg.CacheEnabled)
	assert.Equal(t, ".helixspecifier/cache", cfg.CacheDir)

	// Parallel execution
	assert.Equal(t, 4, cfg.MaxParallelAgents)
	assert.Equal(t, 8, cfg.MaxParallelTasks)

	// Memory / learning
	assert.True(t, cfg.SpecMemoryEnabled)
	assert.True(t, cfg.CrossProjectLearn)

	// Quality thresholds
	assert.Equal(t, 0.6, cfg.MinQualityScore)
	assert.Equal(t, 0.5, cfg.MinPhaseScore)

	// Circuit breaker
	assert.Equal(t, 5, cfg.CircuitBreakerThreshold)
	assert.Equal(t, 30*time.Second, cfg.CircuitBreakerTimeout)
}

func TestFromEnv_Defaults(t *testing.T) {
	cfg := FromEnv()
	defaults := DefaultConfig()

	assert.Equal(t, defaults.QuickMaxMinutes, cfg.QuickMaxMinutes)
	assert.Equal(t, defaults.MediumMaxMinutes, cfg.MediumMaxMinutes)
	assert.Equal(t, defaults.LargeMaxHours, cfg.LargeMaxHours)
	assert.Equal(t, defaults.AdaptiveCeremony, cfg.AdaptiveCeremony)
	assert.Equal(t, defaults.CeremonyBoost, cfg.CeremonyBoost)
	assert.Equal(t, defaults.ConstitutionRounds, cfg.ConstitutionRounds)
	assert.Equal(t, defaults.CacheEnabled, cfg.CacheEnabled)
	assert.Equal(t, defaults.CacheDir, cfg.CacheDir)
	assert.Equal(t, defaults.MaxParallelAgents, cfg.MaxParallelAgents)
	assert.Equal(t, defaults.MaxParallelTasks, cfg.MaxParallelTasks)
	assert.Equal(t, defaults.SpecMemoryEnabled, cfg.SpecMemoryEnabled)
	assert.Equal(t, defaults.CrossProjectLearn, cfg.CrossProjectLearn)
	assert.Equal(t, defaults.MinQualityScore, cfg.MinQualityScore)
	assert.Equal(t, defaults.MinPhaseScore, cfg.MinPhaseScore)
	assert.Equal(
		t,
		defaults.CircuitBreakerThreshold,
		cfg.CircuitBreakerThreshold,
	)
}

func TestFromEnv_QuickMaxMinutes(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_QUICK_MAX_MINUTES", "10")
	cfg := FromEnv()
	assert.Equal(t, 10, cfg.QuickMaxMinutes)
}

func TestFromEnv_MediumMaxMinutes(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_MEDIUM_MAX_MINUTES", "120")
	cfg := FromEnv()
	assert.Equal(t, 120, cfg.MediumMaxMinutes)
}

func TestFromEnv_LargeMaxHours(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_LARGE_MAX_HOURS", "24")
	cfg := FromEnv()
	assert.Equal(t, 24, cfg.LargeMaxHours)
}

func TestFromEnv_AdaptiveCeremony(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true_string", "true", true},
		{"one_string", "1", true},
		{"false_string", "false", false},
		{"zero_string", "0", false},
		{"random_string", "yes", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(
				"HELIX_SPECIFIER_ADAPTIVE_CEREMONY",
				tc.envValue,
			)
			cfg := FromEnv()
			assert.Equal(t, tc.expected, cfg.AdaptiveCeremony)
		})
	}
}

func TestFromEnv_CeremonyBoost(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_CEREMONY_BOOST", "1.5")
	cfg := FromEnv()
	assert.Equal(t, 1.5, cfg.CeremonyBoost)
}

func TestFromEnv_AllRounds(t *testing.T) {
	tests := []struct {
		envVar   string
		value    string
		getField func(*Config) int
	}{
		{
			"HELIX_SPECIFIER_CONSTITUTION_ROUNDS", "7",
			func(c *Config) int { return c.ConstitutionRounds },
		},
		{
			"HELIX_SPECIFIER_SPECIFY_ROUNDS", "5",
			func(c *Config) int { return c.SpecifyRounds },
		},
		{
			"HELIX_SPECIFIER_CLARIFY_ROUNDS", "6",
			func(c *Config) int { return c.ClarifyRounds },
		},
		{
			"HELIX_SPECIFIER_PLAN_ROUNDS", "8",
			func(c *Config) int { return c.PlanRounds },
		},
		{
			"HELIX_SPECIFIER_TASKS_ROUNDS", "4",
			func(c *Config) int { return c.TasksRounds },
		},
		{
			"HELIX_SPECIFIER_ANALYZE_ROUNDS", "9",
			func(c *Config) int { return c.AnalyzeRounds },
		},
		{
			"HELIX_SPECIFIER_IMPLEMENT_ROUNDS", "10",
			func(c *Config) int { return c.ImplementRounds },
		},
	}

	for _, tc := range tests {
		t.Run(tc.envVar, func(t *testing.T) {
			t.Setenv(tc.envVar, tc.value)
			cfg := FromEnv()
			expected, _ := strconv.Atoi(tc.value)
			assert.Equal(t, expected, tc.getField(cfg))
		})
	}
}

func TestFromEnv_CacheEnabled(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_CACHE_ENABLED", "false")
	cfg := FromEnv()
	assert.False(t, cfg.CacheEnabled)

	t.Setenv("HELIX_SPECIFIER_CACHE_ENABLED", "true")
	cfg = FromEnv()
	assert.True(t, cfg.CacheEnabled)
}

func TestFromEnv_CacheDir(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_CACHE_DIR", "/tmp/custom-cache")
	cfg := FromEnv()
	assert.Equal(t, "/tmp/custom-cache", cfg.CacheDir)
}

func TestFromEnv_MaxParallelAgents(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_MAX_PARALLEL_AGENTS", "8")
	cfg := FromEnv()
	assert.Equal(t, 8, cfg.MaxParallelAgents)
}

func TestFromEnv_MaxParallelTasks(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_MAX_PARALLEL_TASKS", "16")
	cfg := FromEnv()
	assert.Equal(t, 16, cfg.MaxParallelTasks)
}

func TestFromEnv_SpecMemoryEnabled(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_SPEC_MEMORY_ENABLED", "false")
	cfg := FromEnv()
	assert.False(t, cfg.SpecMemoryEnabled)

	t.Setenv("HELIX_SPECIFIER_SPEC_MEMORY_ENABLED", "1")
	cfg = FromEnv()
	assert.True(t, cfg.SpecMemoryEnabled)
}

func TestFromEnv_CrossProjectLearn(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_CROSS_PROJECT_LEARN", "false")
	cfg := FromEnv()
	assert.False(t, cfg.CrossProjectLearn)

	t.Setenv("HELIX_SPECIFIER_CROSS_PROJECT_LEARN", "true")
	cfg = FromEnv()
	assert.True(t, cfg.CrossProjectLearn)
}

func TestFromEnv_MinQualityScore(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_MIN_QUALITY_SCORE", "0.85")
	cfg := FromEnv()
	assert.Equal(t, 0.85, cfg.MinQualityScore)
}

func TestFromEnv_MinPhaseScore(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_MIN_PHASE_SCORE", "0.7")
	cfg := FromEnv()
	assert.Equal(t, 0.7, cfg.MinPhaseScore)
}

func TestFromEnv_CircuitBreakerThreshold(t *testing.T) {
	t.Setenv("HELIX_SPECIFIER_CIRCUIT_BREAKER_THRESHOLD", "10")
	cfg := FromEnv()
	assert.Equal(t, 10, cfg.CircuitBreakerThreshold)
}

func TestGetPhaseRounds_AllPhases(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		phase    string
		expected int
	}{
		{"constitution", 5},
		{"specify", 3},
		{"clarify", 3},
		{"plan", 4},
		{"tasks", 2},
		{"analyze", 4},
		{"implement", 6},
	}

	for _, tc := range tests {
		t.Run(tc.phase, func(t *testing.T) {
			assert.Equal(
				t,
				tc.expected,
				cfg.GetPhaseRounds(tc.phase),
			)
		})
	}
}

func TestGetPhaseRounds_Unknown(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 3, cfg.GetPhaseRounds("nonexistent"))
	assert.Equal(t, 3, cfg.GetPhaseRounds(""))
}

func TestGetPhaseTimeout_AllPhases(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		phase    string
		expected time.Duration
	}{
		{"constitution", 15 * time.Minute},
		{"specify", 10 * time.Minute},
		{"clarify", 10 * time.Minute},
		{"plan", 12 * time.Minute},
		{"tasks", 8 * time.Minute},
		{"analyze", 12 * time.Minute},
		{"implement", 20 * time.Minute},
	}

	for _, tc := range tests {
		t.Run(tc.phase, func(t *testing.T) {
			assert.Equal(
				t,
				tc.expected,
				cfg.GetPhaseTimeout(tc.phase),
			)
		})
	}
}

func TestGetPhaseTimeout_Unknown(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(
		t,
		10*time.Minute,
		cfg.GetPhaseTimeout("nonexistent"),
	)
	assert.Equal(t, 10*time.Minute, cfg.GetPhaseTimeout(""))
}

func TestFromEnv_InvalidValues(t *testing.T) {
	defaults := DefaultConfig()

	// Non-numeric strings should be ignored, keeping defaults
	t.Setenv("HELIX_SPECIFIER_QUICK_MAX_MINUTES", "abc")
	t.Setenv("HELIX_SPECIFIER_MEDIUM_MAX_MINUTES", "xyz")
	t.Setenv("HELIX_SPECIFIER_LARGE_MAX_HOURS", "not_a_number")
	t.Setenv("HELIX_SPECIFIER_CEREMONY_BOOST", "invalid")
	t.Setenv("HELIX_SPECIFIER_CONSTITUTION_ROUNDS", "!@#")
	t.Setenv("HELIX_SPECIFIER_MAX_PARALLEL_AGENTS", "many")
	t.Setenv("HELIX_SPECIFIER_MAX_PARALLEL_TASKS", "lots")
	t.Setenv("HELIX_SPECIFIER_MIN_QUALITY_SCORE", "high")
	t.Setenv("HELIX_SPECIFIER_MIN_PHASE_SCORE", "low")
	t.Setenv(
		"HELIX_SPECIFIER_CIRCUIT_BREAKER_THRESHOLD",
		"broken",
	)

	cfg := FromEnv()

	assert.Equal(t, defaults.QuickMaxMinutes, cfg.QuickMaxMinutes)
	assert.Equal(
		t, defaults.MediumMaxMinutes, cfg.MediumMaxMinutes,
	)
	assert.Equal(t, defaults.LargeMaxHours, cfg.LargeMaxHours)
	assert.Equal(t, defaults.CeremonyBoost, cfg.CeremonyBoost)
	assert.Equal(
		t, defaults.ConstitutionRounds, cfg.ConstitutionRounds,
	)
	assert.Equal(
		t, defaults.MaxParallelAgents, cfg.MaxParallelAgents,
	)
	assert.Equal(
		t, defaults.MaxParallelTasks, cfg.MaxParallelTasks,
	)
	assert.Equal(
		t, defaults.MinQualityScore, cfg.MinQualityScore,
	)
	assert.Equal(t, defaults.MinPhaseScore, cfg.MinPhaseScore)
	assert.Equal(
		t,
		defaults.CircuitBreakerThreshold,
		cfg.CircuitBreakerThreshold,
	)
}
