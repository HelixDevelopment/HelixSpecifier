# CLAUDE.md - HelixSpecifier Module


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

```bash
# Effort classification → ceremony scaling → TDD gating for quick / medium / large changes
cd HelixSpecifier && GOMAXPROCS=2 nice -n 19 go test -count=1 -race -p 1 -v \
  -run 'TestE2E_QuickFixWorkflow|TestE2E_MediumFeatureWorkflow|TestE2E_LargeImplementationWorkflow' \
  ./tests/e2e/...
```
Expect: three E2E PASS covering all three effort tiers with appropriate ceremony and TDD gates.


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
| `DebateFuncSetter` | SetDebateFunc | Optional: debate function injection |
| `CLIAgentAdapter` | FormatOutput, ParseInput, GetAgentType, SupportsStreaming | CLI agent output formatting |

### Key Types

| Type | Purpose |
|------|---------|
| `DebateFunc` | Debate execution injection for SpecKit phases |
| `EffortClassifierFunc` | Custom effort classification logic injection |

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

## Injection Points

The FusionEngine supports external dependency injection through
three extension mechanisms:

### DebateFunc

The `DebateFunc` type enables injection of real debate execution
into the SpecKit pillar. Without injection, phases produce
placeholder output with a 0.75 quality score.

```go
type DebateFunc func(
    ctx context.Context,
    topic string,
    rounds int,
    metadata map[string]interface{},
) (output string, score float64, debateID string, err error)
```

Injection via `FusionEngine.SetDebateFunc(fn)` delegates to the
registered SpecKit pillar if it implements `DebateFuncSetter`.

### EffortClassifierFunc

The `EffortClassifierFunc` type enables custom effort
classification logic. Without injection, `ClassifyEffort`
defaults to medium effort.

```go
type EffortClassifierFunc func(
    request string,
) *EffortClassification
```

Register via `FusionEngine.RegisterClassifier(fn)`. The built-in
`intent.Classifier.Classify` provides signal-based classification
detecting quick/medium/large/epic effort from request keywords.

### DebateFuncSetter Interface

Optional interface for pillars accepting debate functions:

```go
type DebateFuncSetter interface {
    SetDebateFunc(fn DebateFunc)
}
```

The `speckit.Pillar` implements this interface. Custom pillar
implementations can opt in by implementing this interface.

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

## Integration Seams

| Direction | Sibling modules |
|-----------|-----------------|
| Upstream (this module imports) | none |
| Downstream (these import this module) | root only |

*Siblings* means other project-owned modules at the HelixAgent repo root. The root HelixAgent app and external systems are not listed here — the list above is intentionally scoped to module-to-module seams, because drift *between* sibling modules is where the "tests pass, product broken" class of bug most often lives. See root `CLAUDE.md` for the rules that keep these seams contract-tested.

## Universal Mandatory Constraints

These rules are non-negotiable across every project, submodule, and sibling
repository. They are derived from the HelixAgent root `CLAUDE.md`. Each
project MUST surface them in its own `CLAUDE.md`, `AGENTS.md`, and
`CONSTITUTION.md`. Project-specific addenda are welcome but cannot weaken
or override these.

### Hard Stops (permanent, non-negotiable)

1. **NO CI/CD pipelines.** No `.github/workflows/`, `.gitlab-ci.yml`,
   `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated pipeline.
   No Git hooks either. All builds and tests run manually or via Makefile/
   script targets.
2. **NO HTTPS for Git.** SSH URLs only (`git@github.com:…`,
   `git@gitlab.com:…`, etc.) for clones, fetches, pushes, and submodule
   updates. Including for public repos. SSH keys are configured on every
   service.
3. **NO manual container commands.** Container orchestration is owned by
   the project's binary/orchestrator (e.g. `make build` → `./bin/<app>`).
   Direct `docker`/`podman start|stop|rm` and `docker-compose up|down`
   are prohibited as workflows. The orchestrator reads its configured
   `.env` and brings up everything.

### Mandatory Development Standards

1. **100% Test Coverage.** Every component MUST have unit, integration,
   E2E, automation, security/penetration, and benchmark tests. No false
   positives. Mocks/stubs ONLY in unit tests; all other test types use
   real data and live services.
2. **Challenge Coverage.** Every component MUST have Challenge scripts
   (`./challenges/scripts/`) validating real-life use cases. No false
   success — validate actual behavior, not return codes.
3. **Real Data.** Beyond unit tests, all components MUST use actual API
   calls, real databases, live services. No simulated success. Fallback
   chains tested with actual failures.
4. **Health & Observability.** Every service MUST expose health
   endpoints. Circuit breakers for all external dependencies. Prometheus
   / OpenTelemetry integration where applicable.
5. **Documentation & Quality.** Update `CLAUDE.md`, `AGENTS.md`, and
   relevant docs alongside code changes. Pass language-appropriate
   format/lint/security gates. Conventional Commits:
   `<type>(<scope>): <description>`.
6. **Validation Before Release.** Pass the project's full validation
   suite (`make ci-validate-all`-equivalent) plus all challenges
   (`./challenges/scripts/run_all_challenges.sh`).
7. **No Mocks or Stubs in Production.** Mocks, stubs, fakes, placeholder
   classes, TODO implementations are STRICTLY FORBIDDEN in production
   code. All production code is fully functional with real integrations.
   Only unit tests may use mocks/stubs.
8. **Comprehensive Verification.** Every fix MUST be verified from all
   angles: runtime testing (actual HTTP requests / real CLI invocations),
   compile verification, code structure checks, dependency existence
   checks, backward compatibility, and no false positives in tests or
   challenges. Grep-only validation is NEVER sufficient.
9. **Resource Limits for Tests & Challenges (CRITICAL).** ALL test and
   challenge execution MUST be strictly limited to 30-40% of host system
   resources. Use `GOMAXPROCS=2`, `nice -n 19`, `ionice -c 3`, `-p 1`
   for `go test`. Container limits required. The host runs
   mission-critical processes — exceeding limits causes system crashes.
10. **Bugfix Documentation.** All bug fixes MUST be documented in
    `docs/issues/fixed/BUGFIXES.md` (or the project's equivalent) with
    root cause analysis, affected files, fix description, and a link to
    the verification test/challenge.
11. **Real Infrastructure for All Non-Unit Tests.** Mocks/fakes/stubs/
    placeholders MAY be used ONLY in unit tests (files ending `_test.go`
    run under `go test -short`, equivalent for other languages). ALL
    other test types — integration, E2E, functional, security, stress,
    chaos, challenge, benchmark, runtime verification — MUST execute
    against the REAL running system with REAL containers, REAL
    databases, REAL services, and REAL HTTP calls. Non-unit tests that
    cannot connect to real services MUST skip (not fail).
12. **Reproduction-Before-Fix (CONST-032 — MANDATORY).** Every reported
    error, defect, or unexpected behavior MUST be reproduced by a
    Challenge script BEFORE any fix is attempted. Sequence:
    (1) Write the Challenge first. (2) Run it; confirm fail (it
    reproduces the bug). (3) Then write the fix. (4) Re-run; confirm
    pass. (5) Commit Challenge + fix together. The Challenge becomes
    the regression guard for that bug forever.
13. **Concurrent-Safe Containers (Go-specific, where applicable).** Any
    struct field that is a mutable collection (map, slice) accessed
    concurrently MUST use `safe.Store[K,V]` / `safe.Slice[T]` from
    `digital.vasic.concurrency/pkg/safe` (or the project's equivalent
    primitives). Bare `sync.Mutex + map/slice` combinations are
    prohibited for new code.

### Definition of Done (universal)

A change is NOT done because code compiles and tests pass. "Done"
requires pasted terminal output from a real run, produced in the same
session as the change.

- **No self-certification.** Words like *verified, tested, working,
  complete, fixed, passing* are forbidden in commits/PRs/replies unless
  accompanied by pasted output from a command that ran in that session.
- **Demo before code.** Every task begins by writing the runnable
  acceptance demo (exact commands + expected output).
- **Real system, every time.** Demos run against real artifacts.
- **Skips are loud.** `t.Skip` / `@Ignore` / `xit` / `describe.skip`
  without a trailing `SKIP-OK: #<ticket>` comment break validation.
- **Evidence in the PR.** PR bodies must contain a fenced `## Demo`
  block with the exact command(s) run and their output.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->



<!-- CONST-035 anti-bluff addendum (cascaded) -->

## CONST-035 — Anti-Bluff Tests & Challenges (mandatory; inherits from root)

Tests and Challenges in this submodule MUST verify the product, not
the LLM's mental model of the product. A test that passes when the
feature is broken is worse than a missing test — it gives false
confidence and lets defects ship to users. Functional probes at the
protocol layer are mandatory:

- TCP-open is the FLOOR, not the ceiling. Postgres → execute
  `SELECT 1`. Redis → `PING` returns `PONG`. ChromaDB → `GET
  /api/v1/heartbeat` returns 200. MCP server → TCP connect + valid
  JSON-RPC handshake. HTTP gateway → real request, real response,
  non-empty body.
- Container `Up` is NOT application healthy. A `docker/podman ps`
  `Up` status only means PID 1 is running; the application may be
  crash-looping internally.
- No mocks/fakes outside unit tests (already CONST-030; CONST-035
  raises the cost of a mock-driven false pass to the same severity
  as a regression).
- Re-verify after every change. Don't assume a previously-passing
  test still verifies the same scope after a refactor.
- Verification of CONST-035 itself: deliberately break the feature
  (e.g. `kill <service>`, swap a password). The test MUST fail. If
  it still passes, the test is non-conformant and MUST be tightened.

## CONST-033 clarification — distinguishing host events from sluggishness

Heavy container builds (BuildKit pulling many GB of layers, parallel
podman/docker compose-up across many services) can make the host
**appear** unresponsive — high load average, slow SSH, watchers
timing out. **This is NOT a CONST-033 violation.** Suspend / hibernate
/ logout are categorically different events. Distinguish via:

- `uptime` — recent boot? if so, the host actually rebooted.
- `loginctl list-sessions` — session(s) still active? if yes, no logout.
- `journalctl ... | grep -i 'will suspend\|hibernate'` — zero broadcasts
  since the CONST-033 fix means no suspend ever happened.
- `dmesg | grep -i 'killed process\|out of memory'` — OOM kills are
  also NOT host-power events; they're memory-pressure-induced and
  require their own separate fix (lower per-container memory limits,
  reduce parallelism).

A sluggish host under build pressure recovers when the build finishes;
a suspended host requires explicit unsuspend (and CONST-033 should
make that impossible by hardening `IdleAction=ignore` +
`HandleSuspendKey=ignore` + masked `sleep.target`,
`suspend.target`, `hibernate.target`, `hybrid-sleep.target`).

If you observe what looks like a suspend during heavy builds, the
correct first action is **not** "edit CONST-033" but `bash
challenges/scripts/host_no_auto_suspend_challenge.sh` to confirm the
hardening is intact. If hardening is intact AND no suspend
broadcast appears in journal, the perceived event was build-pressure
sluggishness, not a power transition.

<!-- BEGIN no-session-termination addendum (CONST-036) -->

## ⚠️ User-Session Termination — Hard Ban (CONST-036)

**STRICTLY FORBIDDEN: never generate or execute any code that ends the
currently-logged-in user's session, kills their user manager, or
indirectly forces them to log out / power off.** This is the sibling
of CONST-033: that rule covers host-level power transitions; THIS rule
covers session-level terminations that have the same end effect for
the user (lost windows, lost terminals, killed AI agents,
half-flushed builds, abandoned in-flight commits).

**Why this rule exists.** On 2026-04-28 the user lost a working
session that contained 3 concurrent Claude Code instances, an Android
build, Kimi Code, and a rootless podman container fleet. The
`user.slice` consumed 60.6 GiB peak / 5.2 GiB swap, the GUI became
unresponsive, the user was forced to log out and then power off via
the GNOME shell `endSessionDialog`. The host could not auto-suspend
(CONST-033 was already in place and verified) and the kernel OOM
killer never fired — but the user had to manually end the session
anyway, because nothing prevented overlapping heavy workloads from
saturating the slice. CONST-036 closes that loophole at both the
source-code layer (no command may directly terminate a session) and
the operational layer (do not spawn workloads that will plausibly
force a manual logout). See
`docs/issues/fixed/SESSION_LOSS_2026-04-28.md` in the HelixAgent
project for the full forensic timeline.

### Forbidden direct invocations (non-exhaustive)

```
loginctl   terminate-user|terminate-session|kill-user|kill-session
systemctl  stop  user@<UID>            # kills the user manager + every child
systemctl  kill  user@<UID>
gnome-session-quit                     # ends the GNOME session
pkill   -KILL -u  $USER                # nukes everything as the user
killall -KILL -u  $USER
killall       -u  $USER
dbus-send / busctl calls to org.gnome.SessionManager.{Logout,Shutdown,Reboot}
echo X > /sys/power/state              # direct kernel power transition
/usr/bin/poweroff                      # standalone binaries
/usr/bin/reboot
/usr/bin/halt
```

### Indirect-pressure clauses

1. Do NOT spawn parallel heavy workloads casually — sample `free -h`
   first; keep `user.slice` under 70% of physical RAM.
2. Long-lived background subagents go in `system.slice`, not
   `user.slice` (rootless podman containers die with the user manager).
3. Document AI-agent concurrency caps in CLAUDE.md per submodule.
4. Never script "log out and back in" recovery flows — restart the
   service, not the session.

### Verification

```bash
bash challenges/scripts/no_session_termination_calls_challenge.sh  # source clean
bash challenges/scripts/no_suspend_calls_challenge.sh              # CONST-033 still clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh          # host hardened
```

All three must PASS.

<!-- END no-session-termination addendum (CONST-036) -->
