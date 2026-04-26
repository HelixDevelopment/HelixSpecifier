# AGENTS.md - HelixSpecifier Agent Integration Guide

## Overview

HelixSpecifier integrates with HelixAgent's 48+ CLI agents through
the `CLIAgentAdapter` interface. Each agent receives spec-driven
development output formatted for its particular consumption model.
The adapter layer bridges the FusionEngine's internal data
structures to agent-specific formats.

## Agent Categories

HelixSpecifier classifies CLI agents into five categories, each
with a dedicated output formatter:

| Category | Format | Agents |
|----------|--------|--------|
| `markdown` | Markdown with headers, lists, status badges | OpenCode, Crush, Claude, Cursor, Windsurf, Cline, Roo |
| `toml` | TOML key-value pairs | Agents using TOML config |
| `github` | GitHub-flavored markdown with task lists | Copilot |
| `minimal` | Plain text, minimal decoration | Resource-constrained agents |
| `generic` | JSON-serialized FlowResult | Aider, Gemini, and all others |

## Well-Known Adapters

Pre-registered adapters are available via
`adapters.WellKnownAdapters()`:

- `opencode` -- Markdown, streaming enabled
- `crush` -- Markdown, streaming enabled
- `claude` -- Markdown, streaming enabled
- `cursor` -- Markdown, streaming enabled
- `windsurf` -- Markdown, streaming enabled
- `copilot` -- GitHub-flavored, streaming enabled
- `gemini` -- Generic JSON, streaming enabled
- `aider` -- Generic JSON, streaming enabled
- `cline` -- Markdown, streaming enabled
- `roo` -- Markdown, streaming enabled

Additional agents use the generic JSON adapter by default.

## Integration Points

### 1. Effort Classification

When an agent sends a user request, HelixSpecifier first classifies
the effort level using `pkg/intent.Classifier`:

```
Agent Request --> Classifier.Classify(request)
  --> EffortClassification{Level, Confidence, CeremonyLevel, ...}
```

Signals detected include epic indicators ("entire system",
"complete rewrite"), large indicators ("implement", "build",
"new system"), quick indicators ("fix typo", "rename"), and
refactoring indicators ("refactor", "migrate").

### 2. Debate Service Integration

The SpecKit pillar accepts a `DebateFunc` injection point for
connecting to HelixAgent's multi-LLM debate service:

```go
speckit.SetDebateFunc(func(
    ctx context.Context,
    topic string,
    rounds int,
    metadata map[string]interface{},
) (output string, score float64, debateID string, err error) {
    // Delegate to HelixAgent debate service
})
```

Each SpecKit phase constructs a debate topic appropriate for that
phase (constitution analysis, specification creation, clarification,
planning, task breakdown, analysis, implementation). The number of
debate rounds is configurable per phase via `HELIX_SPECIFIER_*_ROUNDS`
environment variables.

### 3. Flow Execution

The full integration flow:

1. Agent submits request via HelixAgent API
2. HelixSpecifier classifies effort level
3. Ceremony scaler determines appropriate ceremony
4. FusionEngine executes phases in order:
   - Each phase invokes the debate service (if registered)
   - Adaptive ceremony adjusts mid-flow based on quality
5. Superpowers dispatches parallel subagents for tasks
6. GSD creates and tracks milestones
7. Spec memory learns from the completed flow
8. CLIAgentAdapter formats the FlowResult for the agent

### 4. Metrics Collection

The `pkg/metrics.Metrics` tracker records operational data:

- Flow starts, completions, and failures by effort level
- Phase executions and quality scores by phase type
- Ceremony level distribution
- Total and average durations

## Build Tags

HelixSpecifier uses build tags for conditional compilation in
the parent HelixAgent project:

- **Default (no tag)**: HelixSpecifier is included. Build tag
  `!nohelixspecifier` is the default condition.
- **`nohelixspecifier`**: Opt-out tag. When specified, the
  HelixSpecifier integration is excluded and the engine is
  replaced with a no-op stub.

Usage in HelixAgent:

```go
//go:build !nohelixspecifier

package specifier

import "digital.vasic.helixspecifier/pkg/engine"
```

To build HelixAgent without HelixSpecifier:

```bash
go build -tags nohelixspecifier ./cmd/helixagent
```

## Adding a New Adapter

To add support for a new CLI agent:

### Step 1: Choose or Create a Category

If the agent's output format matches an existing category
(markdown, TOML, GitHub, minimal, generic), use that category.
Otherwise, extend the `FormatOutput` method in
`pkg/adapters/adapters.go`.

### Step 2: Register the Adapter

Add the agent to the `WellKnownAdapters()` map in
`pkg/adapters/adapters.go`:

```go
func WellKnownAdapters() map[string]*GenericAdapter {
    return map[string]*GenericAdapter{
        // ... existing adapters ...
        "newagent": NewGenericAdapter(
            "newagent", CategoryMarkdown,
        ),
    }
}
```

### Step 3: Register at Engine Initialization

In the HelixAgent integration layer, register the adapter
with the FusionEngine:

```go
engine.RegisterAdapter(
    "newagent",
    adapters.NewGenericAdapter("newagent", adapters.CategoryMarkdown),
)
```

### Step 4: Add Tests

Add tests in `pkg/adapters/adapters_test.go` covering:

- `FormatOutput` with a sample `FlowResult`
- `ParseInput` for any agent-specific input parsing
- `GetAgentType` returns the correct name
- `SupportsStreaming` returns the expected value

### Step 5: Update Documentation

Update this file and the adapter table above.

## Configuration for Agents

Agents influence HelixSpecifier behavior through these mechanisms:

1. **Effort override**: An agent can pre-classify effort level
   and pass it directly to `ExecuteFlow`, bypassing the
   automatic classifier.

2. **Phase caching**: Agents can resume interrupted flows via
   `ResumeFlow(flowID)` using the cached flow state.

3. **Parallel limits**: `HELIX_SPECIFIER_MAX_PARALLEL_AGENTS`
   controls how many subagents run simultaneously during the
   Superpowers execution phase.

4. **Quality thresholds**: `HELIX_SPECIFIER_MIN_QUALITY_SCORE`
   and `HELIX_SPECIFIER_MIN_PHASE_SCORE` determine minimum
   acceptable quality. Phases below threshold trigger warnings
   in logs but do not block execution.

## Resource Limits

All test and execution operations respect the HelixAgent resource
management policy:

- `GOMAXPROCS=2` for test execution
- `nice -n 19` and `ionice -c 3` for system-level throttling
- `-p 1` for sequential test package execution
- Container resource limits when running in Docker/Podman

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
