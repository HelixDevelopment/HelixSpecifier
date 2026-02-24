# CLAUDE.md - HelixSpecifier Module

## Overview

`digital.vasic.helixspecifier` is the Spec-Driven Development Fusion
Engine for HelixAgent. It fuses three pillars into a unified workflow:

- **SpecKit**: 7-phase SDD workflow (Constitution, Specify, Clarify,
  Plan, Tasks, Analyze, Implement)
- **Superpowers**: TDD discipline and parallel subagent execution
- **GSD**: Get-Shit-Done milestone lifecycle management

The engine classifies incoming requests by effort level, scales
ceremony accordingly, executes debate-backed specification phases,
dispatches parallel subagents for implementation, and learns from
completed flows via persistent spec memory.

**Module**: `digital.vasic.helixspecifier` (Go 1.24+, toolchain
go1.24.11)

## Architecture

```
                        User Request
                             |
                    +--------v--------+
                    | Effort Classifier|
                    | (pkg/intent)     |
                    +--------+--------+
                             |
               +-------------v--------------+
               |      Ceremony Scaler       |
               |      (pkg/ceremony)        |
               +-------------+--------------+
                             |
          +------------------v------------------+
          |          FusionEngine                |
          |          (pkg/engine)                |
          |                                     |
          |  +----------+ +------+ +----------+ |
          |  | SpecKit  | |Super-| |   GSD    | |
          |  | Pillar   | |powers| |  Pillar  | |
          |  | 7-phase  | | TDD  | |milestone | |
          |  | SDD      | |agents| |lifecycle | |
          |  +----+-----+ +--+---+ +----+-----+ |
          |       |           |          |       |
          +-------+-----------+----------+-------+
                  |           |          |
          +-------v-----------v----------v-------+
          |            Spec Memory               |
          |            (pkg/memory)              |
          +--------------------------------------+
                             |
               +-------------v--------------+
               |      CLI Agent Adapters    |
               |      (pkg/adapters)        |
               +----------------------------+
                             |
     +----------+----------+---------+---------+
     | markdown |   TOML   | GitHub  | generic |
     +----------+----------+---------+---------+

  Power Features (pkg/features/):
  +------------------+  +------------------+
  | parallel_exec    |  | constitution_code|
  | nyquist_tdd      |  | debate_arch      |
  | skill_learning   |  | brownfield       |
  | predictive_spec  |  | cross_project    |
  | adaptive_ceremony|  | spec_memory      |
  +------------------+  +------------------+
```

## Package Structure

| # | Package | Import Path | Purpose |
|---|---------|-------------|---------|
| 1 | `types` | `pkg/types` | Core types, interfaces, enums (EffortLevel, FusionStage, SpecKitPhase, CeremonyLevel) |
| 2 | `config` | `pkg/config` | Configuration from HELIX_SPECIFIER_* env vars with defaults |
| 3 | `engine` | `pkg/engine` | FusionEngine orchestrating the 3-pillar SDD flow |
| 4 | `speckit` | `pkg/speckit` | SpecKit pillar: 7-phase debate-backed SDD workflow |
| 5 | `superpowers` | `pkg/superpowers` | Superpowers pillar: TDD execution and parallel subagent dispatch |
| 6 | `gsd` | `pkg/gsd` | GSD pillar: milestone lifecycle creation and progress tracking |
| 7 | `intent` | `pkg/intent` | Effort classifier: signal-based request analysis |
| 8 | `ceremony` | `pkg/ceremony` | Adaptive ceremony scaler with boost and quality-based adjustment |
| 9 | `memory` | `pkg/memory` | Spec memory store: specification storage and pattern learning |
| 10 | `adapters` | `pkg/adapters` | CLI agent output adapters (markdown, TOML, GitHub, generic) |
| 11 | `metrics` | `pkg/metrics` | Operational metrics: flow, phase, effort, and ceremony tracking |
| 12 | `parallel_execution` | `pkg/features/parallel_execution` | Bounded-concurrency parallel task executor |
| 13 | `constitution_code` | `pkg/features/constitution_code` | Constitution rule management and compliance validation |
| 14 | `nyquist_tdd` | `pkg/features/nyquist_tdd` | Nyquist-frequency TDD ratio tracker (test >= 2x impl) |
| 15 | `debate_architecture` | `pkg/features/debate_architecture` | Multi-round debate orchestration for spec refinement |
| 16 | `skill_learning` | `pkg/features/skill_learning` | Adaptive skill acquisition with proficiency tracking |
| 17 | `brownfield` | `pkg/features/brownfield` | Existing codebase pattern analysis for spec generation |
| 18 | `predictive_spec` | `pkg/features/predictive_spec` | Predictive specification from historical patterns |
| 19 | `cross_project` | `pkg/features/cross_project` | Cross-project knowledge transfer engine |
| 20 | `adaptive_ceremony` | `pkg/features/adaptive_ceremony` | Dynamic ceremony adjustment from real-time quality metrics |
| 21 | `spec_memory` | `pkg/features/spec_memory` | Persistent spec memory index with access tracking |
| 22 | `integration` | `tests/integration` | Cross-package workflow and lifecycle tests |
| 23 | `e2e` | `tests/e2e` | Full user workflow simulation tests |
| 24 | `security` | `tests/security` | Security validation and penetration tests |
| 25 | `stress` | `tests/stress` | Concurrent load and saturation tests |
| 26 | `benchmark` | `tests/benchmark` | Performance measurement benchmarks |
| 27 | `automation` | `tests/automation` | Project structure and API contract verification |

## Build & Test

```bash
go build ./...                                   # Build all packages
GOMAXPROCS=2 go test -count=1 -p 1 ./...         # Run all tests
GOMAXPROCS=2 go test -count=1 -race -p 1 ./...   # Race detection
GOMAXPROCS=2 go test -count=1 -bench=. ./...     # Benchmarks
go test -coverprofile=coverage.out ./...          # Coverage report
go tool cover -html=coverage.out -o coverage.html # HTML report
```

Use the Makefile for convenience:

```bash
make build          # Build all packages
make test           # Run tests (resource-limited)
make test-unit      # Unit tests only (pkg/*)
make test-integration # Integration tests
make test-e2e       # End-to-end tests
make test-security  # Security tests
make test-stress    # Stress tests
make test-automation # Automation tests
make test-all       # All tests (explicit alias for test)
make test-race      # Tests with race detector
make test-cover     # Coverage with HTML report
make test-bench     # Benchmarks
make test-bench-full # Benchmarks with 1s benchtime
make fmt            # gofmt + goimports
make vet            # go vet
make lint           # golangci-lint
```

## Test Suites

| Suite | Location | Functions | Purpose |
|-------|----------|-----------|---------|
| Unit | `pkg/*/` | ~60 | Per-package unit tests with table-driven design |
| Integration | `tests/integration/` | 26 | Cross-package workflow validation |
| E2E | `tests/e2e/` | 18 | Full user workflow simulation |
| Security | `tests/security/` | 26 | Input validation, resource exhaustion, thread safety |
| Stress | `tests/stress/` | 12 | Concurrent load, deadlock detection, saturation |
| Benchmark | `tests/benchmark/` | 22 | Performance measurement across all components |
| Automation | `tests/automation/` | 14 | Project structure, API contracts, build verification |

All test suites enforce resource limits per HelixAgent policy:
`runtime.GOMAXPROCS(2)` in `init()`, `-p 1` for sequential packages.

## Key Interfaces

Seven core interfaces defined in `pkg/types/types.go`:

| Interface | Methods | Purpose |
|-----------|---------|---------|
| `SpecEngine` | ExecuteFlow, ResumeFlow, ClassifyEffort, GetFlowStatus, Health, Name, Version | Main fusion engine contract |
| `SpecKitPillar` | ExecutePhase, GetPhaseOrder, ValidateTransition | 7-phase SDD workflow |
| `SuperpowersPillar` | ExecuteWithTDD, DispatchSubagents, ReviewCode | TDD and parallel subagent execution |
| `GSDPillar` | CreateMilestones, TrackProgress, GetProgress | Milestone lifecycle management |
| `CeremonyScaler` | Scale, Adjust | Adaptive ceremony level determination |
| `SpecMemory` | Store, Search, LearnFromFlow | Specification storage and learning |
| `CLIAgentAdapter` | FormatOutput, ParseInput, GetAgentType, SupportsStreaming | CLI agent output formatting |

## Effort Levels

| Level | Duration | Ceremony | Debate | SpecKit | Stages |
|-------|----------|----------|--------|---------|--------|
| `quick` | < 5 min | Minimal | No | No | Execution only |
| `medium` | 30-90 min | Light | Yes | No | Planning, Execution, Verification |
| `large` | 4-16 hours | Standard | Yes | Yes | All 5 stages |
| `epic` | Days-weeks | Heavy | Yes | Yes | All 5 stages + extended reviews |

## Fusion Stages

1. **Ideation** -- Foundation and vision establishment
2. **Constitution** -- Constitution validation and spec creation
3. **Planning** -- Implementation plan and task breakdown
4. **Execution** -- TDD-driven implementation with subagents
5. **Verification** -- Quality validation and completion

## SpecKit Phases (7-Phase SDD)

1. **Constitution** -- Establish/update project constitution
2. **Specify** -- Create detailed specification from request
3. **Clarify** -- Resolve ambiguities in the specification
4. **Plan** -- Create implementation plan
5. **Tasks** -- Break plan into executable tasks
6. **Analyze** -- Pre-implementation analysis
7. **Implement** -- Execute implementation

Each phase is backed by configurable debate rounds and quality
thresholds. Phases are selected based on ceremony level:

- Minimal: Implement only
- Light: Specify, Plan, Implement
- Standard: All 7 phases
- Heavy: All 7 phases with extended review

## Ceremony Levels

| Level | Phases | Use Case |
|-------|--------|----------|
| `minimal` | Implement only | Quick fixes, typos |
| `light` | Specify + Plan + Implement | Medium tasks |
| `standard` | All 7 phases | Large features |
| `heavy` | All 7 phases + extended reviews | Epic undertakings |

Ceremony is adaptive: if phase quality scores are consistently high
(> 0.85), ceremony is reduced. If quality drops below 0.5, ceremony
is increased.

## Power Features (10)

1. **Parallel Execution** -- Bounded-concurrency task dispatch
   with per-task timeouts
2. **Constitution as Code** -- Machine-readable project
   constitution with mandatory rule enforcement
3. **Nyquist TDD** -- Test-to-implementation ratio tracking
   enforcing >= 2x test coverage
4. **Debate Architecture** -- Multi-round, multi-agent debate
   for specification refinement
5. **Skill Learning** -- Adaptive skill proficiency tracking
   with running-average improvement
6. **Brownfield Analysis** -- Legacy codebase pattern detection
   and constraint analysis
7. **Predictive Specification** -- Future requirement prediction
   from historical patterns
8. **Cross-Project Transfer** -- Knowledge base for transferring
   patterns between projects
9. **Adaptive Ceremony** -- Dynamic ceremony adjustment based on
   real-time quality metrics
10. **Spec Memory** -- Persistent specification index with
    semantic search and access tracking

## Configuration

All configuration uses the `HELIX_SPECIFIER_` prefix. See
`.env.example` for the complete list. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIX_SPECIFIER_QUICK_MAX_MINUTES` | 5 | Max minutes for quick effort |
| `HELIX_SPECIFIER_MEDIUM_MAX_MINUTES` | 90 | Max minutes for medium effort |
| `HELIX_SPECIFIER_LARGE_MAX_HOURS` | 16 | Max hours for large effort |
| `HELIX_SPECIFIER_ADAPTIVE_CEREMONY` | true | Enable adaptive ceremony scaling |
| `HELIX_SPECIFIER_CEREMONY_BOOST` | 0.0 | Ceremony boost factor |
| `HELIX_SPECIFIER_CONSTITUTION_ROUNDS` | 5 | Debate rounds for constitution |
| `HELIX_SPECIFIER_SPECIFY_ROUNDS` | 3 | Debate rounds for specify |
| `HELIX_SPECIFIER_CLARIFY_ROUNDS` | 3 | Debate rounds for clarify |
| `HELIX_SPECIFIER_PLAN_ROUNDS` | 4 | Debate rounds for plan |
| `HELIX_SPECIFIER_TASKS_ROUNDS` | 2 | Debate rounds for tasks |
| `HELIX_SPECIFIER_ANALYZE_ROUNDS` | 4 | Debate rounds for analyze |
| `HELIX_SPECIFIER_IMPLEMENT_ROUNDS` | 6 | Debate rounds for implement |
| `HELIX_SPECIFIER_CACHE_ENABLED` | true | Enable phase caching |
| `HELIX_SPECIFIER_CACHE_DIR` | .helixspecifier/cache | Cache directory |
| `HELIX_SPECIFIER_MAX_PARALLEL_AGENTS` | 4 | Max parallel subagents |
| `HELIX_SPECIFIER_MAX_PARALLEL_TASKS` | 8 | Max parallel tasks |
| `HELIX_SPECIFIER_SPEC_MEMORY_ENABLED` | true | Enable spec memory |
| `HELIX_SPECIFIER_CROSS_PROJECT_LEARN` | true | Enable cross-project learning |
| `HELIX_SPECIFIER_MIN_QUALITY_SCORE` | 0.6 | Minimum overall quality |
| `HELIX_SPECIFIER_MIN_PHASE_SCORE` | 0.5 | Minimum per-phase quality |
| `HELIX_SPECIFIER_CIRCUIT_BREAKER_THRESHOLD` | 5 | Circuit breaker failure count |

Configuration is loaded via `config.FromEnv()` with fallback to
`config.DefaultConfig()`.

## Code Style

- Standard Go conventions, `gofmt` formatting
- Imports grouped: stdlib, third-party, internal (blank line
  separated). Use `goimports`.
- Line length <= 100 chars (readability first)
- Naming: `camelCase` private, `PascalCase` exported, acronyms
  all-caps (`HTTP`, `URL`, `ID`)
- Receivers: 1-2 letters (`e` for engine, `p` for pillar,
  `s` for store)
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`,
  `defer` for cleanup
- Interfaces: small/focused, accept interfaces return structs
- Concurrency: `context.Context` always, `sync.Mutex`/`sync.RWMutex`
  for shared data
- Tests: table-driven, `testify`, naming
  `Test<Struct>_<Method>_<Scenario>`

## Design Patterns

- **Strategy**: Pillar interfaces (SpecKit, Superpowers, GSD) are
  interchangeable implementations
- **Facade**: FusionEngine provides a unified interface over three
  pillars
- **Observer**: Metrics tracking records events from all components
- **Factory**: `NewPillar()`, `NewStore()`, `NewClassifier()`,
  `NewGenericAdapter()`
- **Template Method**: Phase execution follows a fixed sequence
  with pillar-specific behavior
- **Adapter**: CLIAgentAdapter adapts engine output per agent type
- **Semaphore**: Bounded concurrency in Superpowers and
  ParallelExecutor

## Commit Style

Conventional Commits: `feat(engine): add adaptive ceremony scaling`

Scopes: `types`, `config`, `engine`, `speckit`, `superpowers`, `gsd`,
`intent`, `ceremony`, `memory`, `adapters`, `metrics`, `features`
