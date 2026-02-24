# HelixSpecifier

Spec-Driven Development Fusion Engine for HelixAgent.

HelixSpecifier fuses three development pillars into a unified,
adaptive workflow:

- **SpecKit** -- 7-phase Spec-Driven Development (Constitution,
  Specify, Clarify, Plan, Tasks, Analyze, Implement)
- **Superpowers** -- TDD discipline with parallel subagent execution
- **GSD** -- Get-Shit-Done milestone lifecycle management

The engine classifies work by effort level, scales ceremony
accordingly, executes debate-backed specification phases, and learns
from completed flows.

## Features

HelixSpecifier ships with 10 power features:

1. **Parallel Execution** -- Bounded-concurrency task dispatch with
   configurable worker pools and per-task timeouts
2. **Constitution as Code** -- Machine-readable project constitution
   with mandatory rule enforcement and compliance validation
3. **Nyquist TDD** -- Test-to-implementation ratio tracking that
   enforces a minimum 2x test coverage ratio (inspired by the
   Nyquist sampling theorem)
4. **Debate Architecture** -- Multi-round, multi-agent debate
   orchestration for specification refinement with position scoring
5. **Skill Learning** -- Adaptive skill proficiency tracking with
   running-average improvement based on quality feedback
6. **Brownfield Analysis** -- Legacy codebase pattern detection and
   dependency analysis for constraint-aware spec generation
7. **Predictive Specification** -- Future requirement prediction
   from accumulated historical patterns with confidence scoring
8. **Cross-Project Transfer** -- Shared knowledge base enabling
   pattern and decision transfer between projects
9. **Adaptive Ceremony** -- Dynamic ceremony level adjustment based
   on real-time quality metrics during flow execution
10. **Spec Memory** -- Persistent specification index with semantic
    search, access tracking, and pattern learning

## Quick Start

### As a Go Module Dependency

Add HelixSpecifier to your project:

```bash
go get digital.vasic.helixspecifier@latest
```

Or use a local replace directive in your `go.mod`:

```
replace digital.vasic.helixspecifier => ./HelixSpecifier
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "digital.vasic.helixspecifier/pkg/config"
    "digital.vasic.helixspecifier/pkg/engine"
    "digital.vasic.helixspecifier/pkg/ceremony"
    "digital.vasic.helixspecifier/pkg/gsd"
    "digital.vasic.helixspecifier/pkg/memory"
    "digital.vasic.helixspecifier/pkg/speckit"
    "digital.vasic.helixspecifier/pkg/superpowers"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    cfg := config.FromEnv()

    // Create the fusion engine
    eng := engine.New(cfg, logger)

    // Register all three pillars
    eng.RegisterSpecKit(speckit.NewPillar(cfg, logger))
    eng.RegisterSuperpowers(superpowers.NewPillar(cfg, logger))
    eng.RegisterGSD(gsd.NewPillar(logger))
    eng.RegisterCeremonyScaler(ceremony.NewScaler(cfg, logger))
    eng.RegisterSpecMemory(memory.NewStore())

    // Classify effort
    ctx := context.Background()
    classification, err := eng.ClassifyEffort(
        ctx, "Implement user authentication with OAuth2",
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf(
        "Effort: %s, Ceremony: %s\n",
        classification.Level,
        classification.CeremonyLevel,
    )

    // Execute the full flow
    result, err := eng.ExecuteFlow(
        ctx,
        "Implement user authentication with OAuth2",
        classification,
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf(
        "Flow %s completed: quality=%.2f\n",
        result.FlowID,
        result.OverallQualityScore,
    )
}
```

## Architecture

HelixSpecifier is organized as a 3-pillar fusion engine with
supporting infrastructure:

```
User Request
     |
     v
Effort Classifier --> Ceremony Scaler
     |
     v
FusionEngine
  |-- SpecKit Pillar (7-phase SDD with debate)
  |-- Superpowers Pillar (TDD + parallel subagents)
  |-- GSD Pillar (milestone lifecycle)
  |-- Spec Memory (learning + retrieval)
  |-- CLI Agent Adapters (output formatting)
     |
     v
Formatted Output --> CLI Agent
```

### Effort Levels

| Level | Duration | Ceremony | Description |
|-------|----------|----------|-------------|
| Quick | < 5 min | Minimal | Typos, renames, log additions |
| Medium | 30-90 min | Light | Moderate features with planning |
| Large | 4-16 hours | Standard | Full SDD with all 7 phases |
| Epic | Days-weeks | Heavy | System-wide changes, full ceremony |

### Package Overview

The module contains 27 packages (21 core + 6 test suites):

- **Core**: `types`, `config`, `engine`
- **Pillars**: `speckit`, `superpowers`, `gsd`
- **Infrastructure**: `intent`, `ceremony`, `memory`, `adapters`,
  `metrics`
- **Features**: `parallel_execution`, `constitution_code`,
  `nyquist_tdd`, `debate_architecture`, `skill_learning`,
  `brownfield`, `predictive_spec`, `cross_project`,
  `adaptive_ceremony`, `spec_memory`
- **Tests**: `integration`, `e2e`, `security`, `stress`,
  `benchmark`, `automation`

## Installation

HelixSpecifier is designed to be used as a Git submodule within the
HelixAgent project:

```bash
cd /path/to/HelixAgent
git submodule add <repo-url> HelixSpecifier
```

Then add a replace directive in the root `go.mod`:

```
replace digital.vasic.helixspecifier => ./HelixSpecifier
```

For standalone use, clone the repository and import the packages
directly.

## Configuration

All configuration uses environment variables with the
`HELIX_SPECIFIER_` prefix. Copy `.env.example` and adjust values:

```bash
cp .env.example .env
```

Key configuration groups:

- **Effort thresholds**: `QUICK_MAX_MINUTES`, `MEDIUM_MAX_MINUTES`,
  `LARGE_MAX_HOURS`
- **Ceremony**: `ADAPTIVE_CEREMONY`, `CEREMONY_BOOST`
- **Debate rounds per phase**: `CONSTITUTION_ROUNDS` through
  `IMPLEMENT_ROUNDS`
- **Caching**: `CACHE_ENABLED`, `CACHE_DIR`
- **Parallelism**: `MAX_PARALLEL_AGENTS`, `MAX_PARALLEL_TASKS`
- **Memory**: `SPEC_MEMORY_ENABLED`, `CROSS_PROJECT_LEARN`
- **Quality**: `MIN_QUALITY_SCORE`, `MIN_PHASE_SCORE`
- **Circuit breaker**: `CIRCUIT_BREAKER_THRESHOLD`

See `.env.example` for the complete list with defaults and
descriptions.

## Testing

773 tests across 7 test suites, all race-detector clean.

| Suite | Command | Tests | Purpose |
|-------|---------|-------|---------|
| Unit | `make test-unit` | ~60 | Per-package unit tests |
| Integration | `make test-integration` | 26 | Cross-package workflows |
| E2E | `make test-e2e` | 18 | Full user workflow simulation |
| Security | `make test-security` | 26 | Input validation, thread safety |
| Stress | `make test-stress` | 12 | Concurrent load, deadlocks |
| Benchmark | `make test-bench` | 22 | Performance measurement |
| Automation | `make test-automation` | 14 | Structure, API contracts |

Run all tests with resource limits (required by HelixAgent policy):

```bash
make test           # All tests, resource-limited
make test-race      # With race detector
make test-cover     # Coverage report
make test-all       # Explicit all-tests alias
```

Or directly:

```bash
GOMAXPROCS=2 go test -count=1 -race -p 1 ./...
```

Run a single test:

```bash
go test -v -run TestFusionEngine_ExecuteFlow ./pkg/engine/
```

## Contributing

1. Follow Go conventions and the code style in `CLAUDE.md`
2. Every change must include tests (table-driven, using testify)
3. Run `make fmt vet lint` before committing
4. Use Conventional Commits:
   `feat(engine): add adaptive ceremony scaling`
5. Maintain the Nyquist TDD ratio: test code >= 2x implementation

## License

Same license as HelixAgent. See the root LICENSE file.
