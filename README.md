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

831+ tests (including subtests) across 7 test suites, all
race-detector clean.

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

## Anti-Bluff Guarantees (round-273)

HelixSpecifier ships with a layered defence against the failure
mode the constitution captures verbatim:

> Verbatim 2026-05-19 operator mandate (preserved per
> CONST-049 §11.4.17):
> *"all existing tests and Challenges do work in anti-bluff
> manner - they MUST confirm that all tested codebase really
> works as expected! We had been in position that all tests
> do execute with success and all Challenges as well, but in
> reality the most of the features does not work and can't
> be used! This MUST NOT be the case and execution of tests
> and Challenges MUST guarantee the quality, the completition
> and full usability by end users of the product!"*

The layers, top to bottom:

1. **Unit + race**: `GOMAXPROCS=2 go test -count=1 -race -p 1
   ./...` covers every package in `pkg/` plus the new
   `challenges/runner/` package. Mocks are permitted ONLY here
   (CONST-050(A)).
2. **Per-locale Challenge runner** (`challenges/runner/main.go`):
   real `FusionEngine` + real `intent.Classifier` + real
   `i18n.NoopTranslator` exercised across 5 locale fixtures
   (`en`, `sr`, `de`, `ja`, `es`). 20 invariants per run; each
   carries the actual runtime value in the PASS line.
3. **Paired-mutation Challenge**
   (`challenges/helixspecifier_describe_challenge.sh`): wraps the
   runner with `HELIXSPECIFIER_MUTATE_RUNNER=1`, asserts the
   runner DETECTS the planted invariant flip, exits 99 to signal
   "mutation correctly observed" (CONST-050(A) §1.1).
4. **Domain-level Challenges** (`challenges/scripts/*.sh`): chaos,
   DDoS, scaling, stress, UI, UX — env-gated to avoid bluffing
   when no target endpoint is configured (`SKIP-OK:` markers per
   §11.4.6).
5. **Test-coverage ledger** (`docs/test-coverage.md`): every
   public symbol of `pkg/engine`, `pkg/intent`, `pkg/i18n`,
   `pkg/speckit`, `pkg/superpowers`, `pkg/gsd` traced to the
   test + Challenge that proves it works.

### Quick verification

```bash
# Layer 1: race-detector clean across every package
GOMAXPROCS=2 go test -count=1 -race -p 1 ./...

# Layer 2: per-locale runner (exit 0 on success)
go run ./challenges/runner/

# Layer 3: paired-mutation Challenge (exit 0 normal, 99 mutate)
./challenges/helixspecifier_describe_challenge.sh normal
./challenges/helixspecifier_describe_challenge.sh mutate

# Layer 4: domain Challenges (env-gated; SKIP-OK without target)
for c in challenges/scripts/*.sh; do bash "$c" || true; done
```

### What is NOT bluffed

- The runner uses `engine.NewWithTranslator` + real
  `speckit.NewPillar` — no in-process stub pillar.
- Every PASS line includes the runtime value (`got=...`,
  `signals=[...]`), so a regression that returned the wrong
  level/signal would surface in the diff, not just in the
  exit code.
- The mutation hook is checked by `TestRun_MutationDetected` —
  if a future edit silently removes the mutation logic, the
  test fails. This is the §1.1 "paired mutation" requirement.
- Fixtures are checked into the repo as plain YAML; they MUST
  match the live keys in `pkg/i18n/bundles/active.en.yaml`.
  Mismatch → runner FAIL → Challenge exit 1.

### Cascade (CONST-047 / CONST-051(A))

This anti-bluff stack mirrors the pattern in sibling submodules
LeakHub (round 266), MCP_Module (round 267), Models (round 268),
Ouroborous (round 269), and is bumped as part of every meta-repo
close-out round.

## License

Same license as HelixAgent. See the root LICENSE file.
