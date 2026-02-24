// Package config provides configuration management for the
// HelixSpecifier Spec-Driven Development Fusion Engine.
//
// Configuration is loaded from environment variables with the
// HELIX_SPECIFIER_ prefix, falling back to sensible defaults.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all HelixSpecifier configuration.
type Config struct {
	// Effort classification
	QuickMaxMinutes  int `json:"quick_max_minutes"`
	MediumMaxMinutes int `json:"medium_max_minutes"`
	LargeMaxHours    int `json:"large_max_hours"`

	// Ceremony scaling
	AdaptiveCeremony bool    `json:"adaptive_ceremony"`
	CeremonyBoost    float64 `json:"ceremony_boost"`

	// Debate rounds per phase
	ConstitutionRounds int `json:"constitution_rounds"`
	SpecifyRounds      int `json:"specify_rounds"`
	ClarifyRounds      int `json:"clarify_rounds"`
	PlanRounds         int `json:"plan_rounds"`
	TasksRounds        int `json:"tasks_rounds"`
	AnalyzeRounds      int `json:"analyze_rounds"`
	ImplementRounds    int `json:"implement_rounds"`

	// Phase timeouts
	ConstitutionTimeout time.Duration `json:"constitution_timeout"`
	SpecifyTimeout      time.Duration `json:"specify_timeout"`
	ClarifyTimeout      time.Duration `json:"clarify_timeout"`
	PlanTimeout         time.Duration `json:"plan_timeout"`
	TasksTimeout        time.Duration `json:"tasks_timeout"`
	AnalyzeTimeout      time.Duration `json:"analyze_timeout"`
	ImplementTimeout    time.Duration `json:"implement_timeout"`

	// Caching
	CacheEnabled bool   `json:"cache_enabled"`
	CacheDir     string `json:"cache_dir"`

	// Parallel execution
	MaxParallelAgents int `json:"max_parallel_agents"`
	MaxParallelTasks  int `json:"max_parallel_tasks"`

	// Memory / learning
	SpecMemoryEnabled bool `json:"spec_memory_enabled"`
	CrossProjectLearn bool `json:"cross_project_learn"`

	// Quality thresholds
	MinQualityScore float64 `json:"min_quality_score"`
	MinPhaseScore   float64 `json:"min_phase_score"`

	// Circuit breaker
	CircuitBreakerThreshold int           `json:"circuit_breaker_threshold"`
	CircuitBreakerTimeout   time.Duration `json:"circuit_breaker_timeout"`
}

// DefaultConfig returns config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		QuickMaxMinutes:  5,
		MediumMaxMinutes: 90,
		LargeMaxHours:    16,

		AdaptiveCeremony: true,
		CeremonyBoost:    0.0,

		ConstitutionRounds: 5,
		SpecifyRounds:      3,
		ClarifyRounds:      3,
		PlanRounds:         4,
		TasksRounds:        2,
		AnalyzeRounds:      4,
		ImplementRounds:    6,

		ConstitutionTimeout: 15 * time.Minute,
		SpecifyTimeout:      10 * time.Minute,
		ClarifyTimeout:      10 * time.Minute,
		PlanTimeout:         12 * time.Minute,
		TasksTimeout:        8 * time.Minute,
		AnalyzeTimeout:      12 * time.Minute,
		ImplementTimeout:    20 * time.Minute,

		CacheEnabled: true,
		CacheDir:     ".helixspecifier/cache",

		MaxParallelAgents: 4,
		MaxParallelTasks:  8,

		SpecMemoryEnabled: true,
		CrossProjectLearn: true,

		MinQualityScore: 0.6,
		MinPhaseScore:   0.5,

		CircuitBreakerThreshold: 5,
		CircuitBreakerTimeout:   30 * time.Second,
	}
}

// FromEnv loads configuration from environment variables.
// Variables use the HELIX_SPECIFIER_ prefix. Any variable not
// set in the environment falls back to the default value.
func FromEnv() *Config {
	cfg := DefaultConfig()

	if v := os.Getenv(
		"HELIX_SPECIFIER_QUICK_MAX_MINUTES",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.QuickMaxMinutes = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_MEDIUM_MAX_MINUTES",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MediumMaxMinutes = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_LARGE_MAX_HOURS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.LargeMaxHours = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_ADAPTIVE_CEREMONY",
	); v != "" {
		cfg.AdaptiveCeremony = v == "true" || v == "1"
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CEREMONY_BOOST",
	); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.CeremonyBoost = f
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CONSTITUTION_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.ConstitutionRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_SPECIFY_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.SpecifyRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CLARIFY_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.ClarifyRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_PLAN_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.PlanRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_TASKS_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.TasksRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_ANALYZE_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.AnalyzeRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_IMPLEMENT_ROUNDS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.ImplementRounds = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CACHE_ENABLED",
	); v != "" {
		cfg.CacheEnabled = v == "true" || v == "1"
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CACHE_DIR",
	); v != "" {
		cfg.CacheDir = v
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_MAX_PARALLEL_AGENTS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxParallelAgents = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_MAX_PARALLEL_TASKS",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxParallelTasks = n
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_SPEC_MEMORY_ENABLED",
	); v != "" {
		cfg.SpecMemoryEnabled = v == "true" || v == "1"
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CROSS_PROJECT_LEARN",
	); v != "" {
		cfg.CrossProjectLearn = v == "true" || v == "1"
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_MIN_QUALITY_SCORE",
	); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.MinQualityScore = f
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_MIN_PHASE_SCORE",
	); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.MinPhaseScore = f
		}
	}
	if v := os.Getenv(
		"HELIX_SPECIFIER_CIRCUIT_BREAKER_THRESHOLD",
	); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.CircuitBreakerThreshold = n
		}
	}

	return cfg
}

// GetPhaseRounds returns the configured debate rounds for a
// given phase name.
func (c *Config) GetPhaseRounds(phase string) int {
	switch phase {
	case "constitution":
		return c.ConstitutionRounds
	case "specify":
		return c.SpecifyRounds
	case "clarify":
		return c.ClarifyRounds
	case "plan":
		return c.PlanRounds
	case "tasks":
		return c.TasksRounds
	case "analyze":
		return c.AnalyzeRounds
	case "implement":
		return c.ImplementRounds
	default:
		return 3
	}
}

// GetPhaseTimeout returns the configured timeout duration for
// a given phase name.
func (c *Config) GetPhaseTimeout(phase string) time.Duration {
	switch phase {
	case "constitution":
		return c.ConstitutionTimeout
	case "specify":
		return c.SpecifyTimeout
	case "clarify":
		return c.ClarifyTimeout
	case "plan":
		return c.PlanTimeout
	case "tasks":
		return c.TasksTimeout
	case "analyze":
		return c.AnalyzeTimeout
	case "implement":
		return c.ImplementTimeout
	default:
		return 10 * time.Minute
	}
}
