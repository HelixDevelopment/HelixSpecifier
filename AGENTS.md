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
