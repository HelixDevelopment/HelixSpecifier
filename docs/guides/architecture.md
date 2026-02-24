# Architecture Guide

## Three-Pillar Fusion

HelixSpecifier's architecture is built around three complementary
pillars that are fused together by the central FusionEngine:

### Pillar 1: SpecKit (7-Phase SDD)

SpecKit provides the structured specification workflow. It takes a
user request through up to seven phases, each backed by configurable
multi-LLM debate rounds:

```
Constitution --> Specify --> Clarify --> Plan
     --> Tasks --> Analyze --> Implement
```

The number of phases executed depends on the ceremony level. Minimal
ceremony runs only the Implement phase. Light ceremony adds Specify
and Plan. Standard and Heavy ceremony run all seven phases.

Each phase produces a `PhaseResult` containing the debate output,
quality score, duration, and artifacts. The phase output feeds into
the next phase as `PreviousOutput`, creating a continuous
specification refinement pipeline.

### Pillar 2: Superpowers (TDD + Subagents)

Superpowers handles the actual execution of tasks with two modes:

- **Sequential TDD**: Tasks are executed one at a time with
  test-driven discipline. Each task writes tests first, then
  implementation, then verification.

- **Parallel Subagents**: Tasks are dispatched to concurrent
  goroutines bounded by a semaphore (`MaxParallelAgents`). Each
  subagent operates independently within a context with
  cancellation support.

After execution, Superpowers provides code review by analyzing
task results and producing a `ReviewResult` with issues and an
approval decision.

### Pillar 3: GSD (Milestone Lifecycle)

GSD (Get-Shit-Done) manages the project management layer:

- **Milestone Creation**: Tasks are grouped by priority (critical,
  high, medium, low) into sequential milestones with dependency
  chains.

- **Progress Tracking**: As tasks complete, milestone progress is
  updated. Milestones transition through states: Queued, Active,
  Blocked, Completed, Failed.

- **Overall Progress**: The engine reports aggregate progress across
  all milestones for a flow.

## Effort Classification

The effort classifier (`pkg/intent`) analyzes user requests using
signal-based pattern matching:

```
User Request
     |
     v
Signal Detection:
  - Epic: "entire system", "complete rewrite", "whole architecture"
  - Large: "implement" + "service", "build" + "api", len > 200
  - Quick: "fix typo", "rename", "add import"
  - Refactor: "refactor", "migrate", "rewrite" (boosts to Large)
     |
     v
EffortClassification:
  - Level: quick | medium | large | epic
  - Confidence: 0.0 - 1.0
  - CeremonyLevel: derived from effort
  - RequiresDebate: true for medium+
  - RequiresSpecKit: true for large+
```

## Adaptive Ceremony

Ceremony determines how many phases run and how much process
rigor is applied. The ceremony scaler works in two stages:

### Initial Scaling

Effort level maps to a default ceremony:

| Effort | Default Ceremony |
|--------|-----------------|
| Quick | Minimal |
| Medium | Light |
| Large | Standard |
| Epic | Heavy |

If `CeremonyBoost > 0`, the ceremony is raised by one level.

### Dynamic Adjustment

When `AdaptiveCeremony` is enabled (default), the engine evaluates
phase quality scores after each phase:

- **Average quality > 0.85**: Ceremony reduced by one level (heavy
  to standard, standard to light, etc.)
- **Average quality < 0.50**: Ceremony increased by one level

This prevents over-engineering simple tasks that produce high-quality
results quickly, and adds rigor when quality is struggling.

## Phase Flow

The complete execution flow through the FusionEngine:

```
1. Request received
2. ClassifyEffort(request)
   --> EffortClassification
3. CeremonyScaler.Scale(classification)
   --> CeremonyLevel
4. PhasesForCeremony(ceremonyLevel)
   --> [Phase1, Phase2, ...]
5. For each phase:
   a. Build debate topic for phase
   b. Execute debate (if DebateFunc registered)
   c. Collect PhaseResult
   d. Check quality threshold
   e. Adaptive ceremony adjustment (if enabled)
   f. Feed output to next phase
6. Calculate overall quality score
7. SpecMemory.LearnFromFlow (if enabled)
8. Return FlowResult
```

## Milestone Lifecycle

GSD milestones follow a state machine:

```
  Queued --> Active --> Completed
              |
              v
           Blocked
              |
              v
            Failed
```

Milestones are created in priority order with dependency chains:
critical milestones must complete before high-priority ones begin.
Each milestone contains a subset of tasks grouped by priority.

## Integration with HelixAgent

HelixSpecifier integrates with the parent HelixAgent project as a
Go submodule:

```
HelixAgent
  |-- internal/services/speckit_orchestrator.go
  |     (bridges HelixAgent debate service to SpecKit DebateFunc)
  |
  |-- HelixSpecifier/
  |     |-- pkg/engine/engine.go       (FusionEngine)
  |     |-- pkg/speckit/speckit.go     (SpecKit pillar)
  |     |-- pkg/superpowers/           (Superpowers pillar)
  |     |-- pkg/gsd/                   (GSD pillar)
  |     |-- pkg/adapters/              (CLI agent adapters)
  |     +-- pkg/features/             (10 power features)
  |
  |-- internal/adapters/specifier/
        (adapter connecting HelixSpecifier to internal types)
```

The debate service integration uses function injection: the
HelixAgent `speckit_orchestrator` sets the `DebateFunc` on the
SpecKit pillar, routing debate requests through the full ensemble
LLM debate infrastructure.

## Diagram: Complete Data Flow

```
  +--------+    +----------+    +---------+    +--------+
  | CLI    |--->| Effort   |--->|Ceremony |--->| Fusion |
  | Agent  |    |Classifier|    | Scaler  |    | Engine |
  +--------+    +----------+    +---------+    +---+----+
                                                   |
       +-------------------------------------------+
       |               |               |
  +----v----+    +-----v-----+    +----v----+
  | SpecKit |    |Superpowers|    |   GSD   |
  | 7-phase |    | TDD +     |    |milestone|
  | SDD     |    | subagents |    |lifecycle|
  +----+----+    +-----+-----+    +----+----+
       |               |               |
       +-------+-------+-------+-------+
               |               |
          +----v----+    +-----v-----+
          |  Spec   |    |  Metrics  |
          | Memory  |    |  Tracker  |
          +---------+    +-----------+
               |
          +----v----+
          |  CLI    |
          | Adapter |
          +----+----+
               |
               v
          Formatted Output
```
