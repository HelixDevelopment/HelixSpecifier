# HelixSpecifier - Getting Started

**Module:** `digital.vasic.helixspecifier`

## Overview

HelixSpecifier is a Spec-Driven Development (SDD) Fusion Engine that
classifies incoming requests by effort level, scales ceremony
accordingly, executes debate-backed specification phases, dispatches
parallel subagents, and learns from completed flows.

## Quick Start

### 1. Create the Fusion Engine

```go
package main

import (
    "context"
    "fmt"

    "digital.vasic.helixspecifier/pkg/config"
    "digital.vasic.helixspecifier/pkg/engine"
    "digital.vasic.helixspecifier/pkg/speckit"
    "digital.vasic.helixspecifier/pkg/superpowers"
    "digital.vasic.helixspecifier/pkg/gsd"
    "digital.vasic.helixspecifier/pkg/intent"
    "digital.vasic.helixspecifier/pkg/ceremony"
    "digital.vasic.helixspecifier/pkg/memory"
    "digital.vasic.helixspecifier/pkg/types"
)

func main() {
    cfg := config.FromEnv()

    eng := engine.NewFusionEngine(cfg)

    // Register the three pillars
    eng.RegisterSpecKit(speckit.NewPillar(cfg))
    eng.RegisterSuperpowers(superpowers.NewPillar(cfg))
    eng.RegisterGSD(gsd.NewPillar(cfg))

    // Register supporting services
    eng.RegisterCeremonyScaler(ceremony.NewScaler(cfg))
    eng.RegisterSpecMemory(memory.NewStore())
    eng.RegisterClassifier(intent.NewClassifier().Classify)
```

### 2. Classify Effort

```go
    classification := eng.ClassifyEffort("Build a REST API for user management with auth")
    fmt.Printf("Effort: %s (confidence: %.2f)\n", classification.Level, classification.Confidence)
    fmt.Printf("Ceremony: %s\n", classification.CeremonyLevel)
    fmt.Printf("Requires debate: %v\n", classification.RequiresDebate)
```

### 3. Execute a Flow

```go
    result, err := eng.ExecuteFlow(context.Background(), types.FlowRequest{
        Description: "Build a REST API for user management with auth",
        UserID:      "developer-1",
    })
    if err != nil {
        panic(err)
    }

    fmt.Printf("Flow ID:  %s\n", result.FlowID)
    fmt.Printf("Quality:  %.2f\n", result.QualityScore)
    fmt.Printf("Phases:   %d\n", len(result.PhaseResults))
    fmt.Printf("Status:   %s\n", result.Status)
}
```

## Using SpecKit Phases

The SpecKit pillar executes a 7-phase SDD workflow. The number of
phases depends on the ceremony level:

| Ceremony Level | Phases Executed |
|----------------|----------------|
| Minimal | Implement only |
| Light | Specify, Plan, Implement |
| Standard | All 7 phases |
| Heavy | All 7 phases + extended reviews |

### The 7 Phases

1. **Constitution** -- Analyze request, establish project rules
2. **Specify** -- Create detailed specification
3. **Clarify** -- Resolve ambiguities
4. **Plan** -- Create implementation plan
5. **Tasks** -- Break plan into executable tasks
6. **Analyze** -- Pre-implementation analysis
7. **Implement** -- Execute the implementation

## Effort Classification

The classifier analyzes request text for keyword signals:

| Effort Level | Typical Duration | Signal Examples |
|--------------|-----------------|-----------------|
| Quick | < 5 min | "fix typo", "add log", "rename" |
| Medium | 30-90 min | Moderate features, standard tasks |
| Large | 4-16 hours | "implement service", "build API" |
| Epic | Days-weeks | "entire system", "complete rewrite" |

## Injecting a Debate Function

For real multi-LLM debate execution, inject a `DebateFunc`:

```go
    eng.SetDebateFunc(func(ctx context.Context, topic string, rounds int,
        metadata map[string]interface{}) (string, float64, string, error) {
        // Execute real debate using your debate service
        return output, qualityScore, debateID, nil
    })
```

Without injection, phases produce placeholder output with a 0.75
quality score.

## Resuming Interrupted Flows

Flows support caching and resumption:

```go
    result, err := eng.ResumeFlow(ctx, flowID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Resumed from phase: %s\n", result.ResumedFromPhase)
```

Enable caching in configuration:

```bash
export HELIX_SPECIFIER_CACHE_ENABLED=true
export HELIX_SPECIFIER_CACHE_DIR=.helixspecifier/cache
```

## Adaptive Ceremony

Ceremony levels adjust automatically during execution:

- If phase quality scores average > 0.85 with Heavy ceremony, it
  reduces to Standard
- If quality drops below 0.5 with Light ceremony, it increases to
  Standard

Disable adaptive scaling:

```bash
export HELIX_SPECIFIER_ADAPTIVE_CEREMONY=false
```

## CLI Agent Output

Format engine output for specific CLI agents:

```go
    adapter := adapters.GetAdapter("opencode")
    formatted, err := adapter.FormatOutput(result)
    fmt.Println(formatted)
```

Supported adapters: OpenCode, Crush, Claude, Cursor, Windsurf,
Copilot, Gemini, Aider, Cline, Roo.

## Build and Test

```bash
make build       # Build all packages
make test        # Run all tests (resource-limited)
make test-unit   # Unit tests only
make lint        # Run linters
```

## Next Steps

- See [ARCHITECTURE.md](ARCHITECTURE.md) for the 3-pillar design
- See [API_REFERENCE.md](API_REFERENCE.md) for all types and interfaces
- See [CONFIGURATION.md](CONFIGURATION.md) for environment variables
- See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
