# HelixSpecifier - Configuration

**Module:** `digital.vasic.helixspecifier`

## Overview

All configuration is loaded from environment variables prefixed with
`HELIX_SPECIFIER_`. Missing variables fall back to sensible defaults.
Configuration is loaded via `config.FromEnv()` with fallback to
`config.DefaultConfig()`.

## Effort Classification

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_QUICK_MAX_MINUTES` | `5` | Maximum minutes for Quick effort level |
| `HELIX_SPECIFIER_MEDIUM_MAX_MINUTES` | `90` | Maximum minutes for Medium effort level |
| `HELIX_SPECIFIER_LARGE_MAX_HOURS` | `16` | Maximum hours for Large effort level |

### Effort-to-Ceremony Mapping

| Effort Level | Duration Range | Default Ceremony |
|-------------|---------------|-----------------|
| Quick | < 5 minutes | Minimal |
| Medium | 5 min - 90 min | Light |
| Large | 90 min - 16 hours | Standard |
| Epic | > 16 hours | Heavy |

## Ceremony Scaling

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_ADAPTIVE_CEREMONY` | `true` | Enable runtime ceremony adjustment |
| `HELIX_SPECIFIER_CEREMONY_BOOST` | `0.0` | Boost factor: increases ceremony by one level when > 0 |

### Adaptive Adjustment Rules

When `ADAPTIVE_CEREMONY` is enabled:

- Average quality > 0.85 with Heavy ceremony reduces to Standard
- Average quality < 0.50 with Light ceremony increases to Standard
- Intermediate quality scores keep current ceremony level

## Debate Rounds Per Phase

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_CONSTITUTION_ROUNDS` | `5` | Debate rounds for Constitution phase |
| `HELIX_SPECIFIER_SPECIFY_ROUNDS` | `3` | Debate rounds for Specify phase |
| `HELIX_SPECIFIER_CLARIFY_ROUNDS` | `3` | Debate rounds for Clarify phase |
| `HELIX_SPECIFIER_PLAN_ROUNDS` | `4` | Debate rounds for Plan phase |
| `HELIX_SPECIFIER_TASKS_ROUNDS` | `2` | Debate rounds for Tasks phase |
| `HELIX_SPECIFIER_ANALYZE_ROUNDS` | `4` | Debate rounds for Analyze phase |
| `HELIX_SPECIFIER_IMPLEMENT_ROUNDS` | `6` | Debate rounds for Implement phase |

Higher round counts produce higher quality but cost more time and tokens.

## Parallelism

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_MAX_PARALLEL_AGENTS` | `4` | Maximum concurrent subagents |
| `HELIX_SPECIFIER_MAX_PARALLEL_TASKS` | `8` | Maximum concurrent tasks |

The Superpowers pillar uses channel-based semaphore patterns bounded
by these limits.

## Caching and Resumption

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_CACHE_ENABLED` | `true` | Enable phase result caching |
| `HELIX_SPECIFIER_CACHE_DIR` | `.helixspecifier/cache` | Directory for cached phase results |

Caching enables interrupted flows to be resumed from the last completed
phase using `ResumeFlow()`.

## Memory and Learning

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_SPEC_MEMORY_ENABLED` | `true` | Enable specification persistence |
| `HELIX_SPECIFIER_CROSS_PROJECT_LEARN` | `true` | Enable cross-project pattern learning |

When enabled, completed flows are analyzed for patterns (effort level,
ceremony used, phase count, quality score) and stored for future reference.

## Quality Thresholds

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_MIN_QUALITY_SCORE` | `0.6` | Minimum acceptable overall quality |
| `HELIX_SPECIFIER_MIN_PHASE_SCORE` | `0.5` | Minimum acceptable per-phase quality |

Phases scoring below `MIN_PHASE_SCORE` may trigger adaptive ceremony
increases when `ADAPTIVE_CEREMONY` is enabled.

## Circuit Breaker

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_CIRCUIT_BREAKER_THRESHOLD` | `5` | Consecutive failures before circuit opens |

## Example .env File

```bash
# Effort classification
HELIX_SPECIFIER_QUICK_MAX_MINUTES=5
HELIX_SPECIFIER_MEDIUM_MAX_MINUTES=90
HELIX_SPECIFIER_LARGE_MAX_HOURS=16

# Ceremony
HELIX_SPECIFIER_ADAPTIVE_CEREMONY=true
HELIX_SPECIFIER_CEREMONY_BOOST=0.0

# Debate rounds
HELIX_SPECIFIER_CONSTITUTION_ROUNDS=5
HELIX_SPECIFIER_SPECIFY_ROUNDS=3
HELIX_SPECIFIER_PLAN_ROUNDS=4
HELIX_SPECIFIER_IMPLEMENT_ROUNDS=6

# Parallelism
HELIX_SPECIFIER_MAX_PARALLEL_AGENTS=4
HELIX_SPECIFIER_MAX_PARALLEL_TASKS=8

# Caching
HELIX_SPECIFIER_CACHE_ENABLED=true
HELIX_SPECIFIER_CACHE_DIR=.helixspecifier/cache

# Quality
HELIX_SPECIFIER_MIN_QUALITY_SCORE=0.6
HELIX_SPECIFIER_MIN_PHASE_SCORE=0.5

# Memory
HELIX_SPECIFIER_SPEC_MEMORY_ENABLED=true
HELIX_SPECIFIER_CROSS_PROJECT_LEARN=true
```

## Integration with HelixAgent

When used within HelixAgent, HelixSpecifier auto-activates based on
work granularity detection:

| Granularity | HelixSpecifier Active |
|-------------|----------------------|
| Single Action | No |
| Small Creation | No |
| Big Creation | Yes |
| Whole Functionality | Yes |
| Refactoring | Yes |

Auto-activation is managed by HelixAgent's `speckit_orchestrator.go`
and `enhanced_intent_classifier.go`.

Opt out of HelixSpecifier with the build tag: `-tags nohelixspecifier`.
