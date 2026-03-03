# HelixSpecifier Module Architecture

**Module:** `digital.vasic.helixspecifier`

## Overview

HelixSpecifier is a Spec-Driven Development (SDD) Fusion Engine for HelixAgent. It fuses three development pillars -- SpecKit, Superpowers, and GSD -- into a unified, adaptive workflow. The engine classifies incoming requests by effort level, scales ceremony accordingly, executes debate-backed specification phases, dispatches parallel subagents for implementation, and learns from completed flows via persistent spec memory.

The module contains 27 packages (21 core + 6 test suites), 835+ tests, and is active by default in HelixAgent (opt out with `-tags nohelixspecifier`).

## Core Components

### FusionEngine (`pkg/engine/engine.go`)

The central orchestrator that fuses the three pillars into a single SDD flow. It manages flow lifecycle, phase execution, adaptive ceremony adjustment, task dispatch, milestone tracking, and post-flow learning.

**Key responsibilities:**
- Register and coordinate SpecKit, Superpowers, and GSD pillars
- Execute flows as a sequence of SpecKit phases determined by ceremony level
- Apply adaptive ceremony adjustment during execution (increase on low quality, decrease on high quality)
- Delegate task execution to Superpowers pillar for parallel subagent dispatch
- Create and track milestones through the GSD pillar
- Learn from completed flows via SpecMemory
- Support flow caching and resumption
- Inject debate functions into the SpecKit pillar for real multi-LLM debate

**Implements:** `SpecEngine` interface (ExecuteFlow, ResumeFlow, ClassifyEffort, GetFlowStatus, Health, Name, Version)

### SpecKit Pillar (`pkg/speckit/speckit.go`)

Implements the 7-phase Spec-Driven Development workflow. Each phase generates a debate topic from user input and previous phase output, then either executes a real debate via the injected `DebateFunc` or produces placeholder output (0.75 quality score) when no debate function is available.

**7 Phases (in order):**
1. **Constitution** -- Analyze the request and create/update the project constitution
2. **Specify** -- Create a detailed specification from the request and constitution
3. **Clarify** -- Resolve ambiguities and validate the specification
4. **Plan** -- Create an implementation plan from the specification
5. **Tasks** -- Break the plan into executable tasks
6. **Analyze** -- Pre-implementation analysis of tasks
7. **Implement** -- Execute the implementation

**Phase selection by ceremony level:**

| Ceremony | Phases Executed |
|----------|----------------|
| Minimal | Implement only |
| Light | Specify, Plan, Implement |
| Standard | All 7 phases |
| Heavy | All 7 phases with extended reviews |

**Implements:** `SpecKitPillar` interface + `DebateFuncSetter` for debate injection

### Superpowers Pillar (`pkg/superpowers/superpowers.go`)

Provides TDD-driven execution and parallel subagent dispatch. Tasks are executed with bounded concurrency using a channel-based semaphore pattern.

**Key operations:**
- `ExecuteWithTDD` -- Sequential task execution with TDD discipline (context-aware cancellation)
- `DispatchSubagents` -- Parallel task execution with configurable concurrency limit (default from `MaxParallelAgents` config). Uses goroutines with `sync.WaitGroup` and channel-based semaphore.
- `ReviewCode` -- Aggregates task results into a review score, identifies failures as critical issues

**Implements:** `SuperpowersPillar` interface (ExecuteWithTDD, DispatchSubagents, ReviewCode)

### GSD Pillar (`pkg/gsd/gsd.go`)

Get-Shit-Done milestone lifecycle management. Groups tasks by priority into sequential milestones with dependency chains, and tracks progress as task results arrive.

**Milestone lifecycle:** `queued` -> `active` -> `completed` (or `blocked`/`failed`)

**Key operations:**
- `CreateMilestones` -- Groups tasks by priority (critical -> high -> medium -> low), creates milestones with dependency chains
- `TrackProgress` -- Updates milestone progress based on completed task ratios, transitions status automatically
- `GetProgress` -- Returns aggregate progress across all milestones

**Implements:** `GSDPillar` interface (CreateMilestones, TrackProgress, GetProgress)

### Effort Classifier (`pkg/intent/classifier.go`)

Signal-based request analysis that determines the effort level of incoming requests. Analyzes request text for keyword signals to classify effort.

**Classification logic:**
- **Quick** signals: "fix typo", "add log", "rename", "update comment", "bump version"
- **Large** signals: "implement" + "service", "build" + "api", etc. (2+ matches or 200+ chars)
- **Epic** signals: "entire system", "complete rewrite", "whole architecture"
- **Refactoring** boost: "refactor", "restructure", "rewrite" boost to at least Large
- **Default:** Medium effort (0.5 confidence)

**Derived fields:** CeremonyLevel, RequiresDebate (non-quick), RequiresSpecKit (large/epic)

### Ceremony Scaler (`pkg/ceremony/ceremony.go`)

Adaptive ceremony level determination. Determines the initial ceremony level from effort classification, then dynamically adjusts during flow execution based on phase quality scores.

**Scaling rules:**
- Initial: Maps effort level to default ceremony (Quick->Minimal, Medium->Light, Large->Standard, Epic->Heavy)
- Boost: Optional boost from config (`HELIX_SPECIFIER_CEREMONY_BOOST`) increases ceremony by one level
- Adaptive adjustment: Average quality > 0.85 with Heavy ceremony -> reduce to Standard. Average quality < 0.5 with Light ceremony -> increase to Standard.

**Implements:** `CeremonyScaler` interface (Scale, Adjust)

### Spec Memory (`pkg/memory/memory.go`)

Persistent specification storage and cross-project learning. Stores completed specifications, searches by title/description, and extracts patterns from successful flows.

**Key operations:**
- `Store` -- Saves specifications for future reference
- `Search` -- Finds similar specifications using case-insensitive title/description matching
- `LearnFromFlow` -- Extracts patterns from successful flows (effort, ceremony, phase count, quality score)

Tracks learned patterns with use count and success rate for continuous improvement.

**Implements:** `SpecMemory` interface (Store, Search, LearnFromFlow)

### CLI Agent Adapters (`pkg/adapters/adapters.go`)

Format engine output for specific CLI agents. Supports multiple output categories:

| Category | Format | Used By |
|----------|--------|---------|
| Markdown | Structured markdown with headers | OpenCode, Crush, Claude, Cursor, Windsurf, Cline, Roo |
| GitHub | GitHub-flavored formatting | Copilot |
| TOML | TOML key-value format | (custom agents) |
| Generic/JSON | Pretty-printed JSON | Gemini, Aider |

Ships with 10 pre-configured adapters via `WellKnownAdapters()`.

**Implements:** `CLIAgentAdapter` interface (FormatOutput, ParseInput, GetAgentType, SupportsStreaming)

### Metrics (`pkg/metrics/metrics.go`)

Operational metrics tracking for all HelixSpecifier components. Thread-safe with `sync.RWMutex`.

**Tracked metrics:**
- Flow counts (started, completed, failed) with success rate calculation
- Phase execution counts per phase (executions, successes, failures, average score, total duration)
- Effort level distribution
- Ceremony level distribution
- Snapshot support for read-only copies

### Configuration (`pkg/config/config.go`)

Loads all settings from `HELIX_SPECIFIER_*` environment variables with sensible defaults. Covers effort thresholds, debate rounds per phase, parallelism limits, caching, memory, quality thresholds, and circuit breaker settings.

### 10 Power Features (`pkg/features/`)

| # | Feature | Package | Description |
|---|---------|---------|-------------|
| 1 | Parallel Execution | `parallel_execution` | Bounded-concurrency task dispatch with configurable worker pools and per-task timeouts |
| 2 | Constitution as Code | `constitution_code` | Machine-readable project constitution with mandatory rule enforcement and compliance validation |
| 3 | Nyquist TDD | `nyquist_tdd` | Test-to-implementation ratio tracking enforcing >= 2x test coverage ratio (Nyquist sampling theorem) |
| 4 | Debate Architecture | `debate_architecture` | Multi-round, multi-agent debate orchestration for specification refinement with position scoring |
| 5 | Skill Learning | `skill_learning` | Adaptive skill proficiency tracking with running-average improvement based on quality feedback |
| 6 | Brownfield Analysis | `brownfield` | Legacy codebase pattern detection, dependency analysis, and constraint-aware spec generation |
| 7 | Predictive Specification | `predictive_spec` | Future requirement prediction from accumulated historical patterns with confidence scoring |
| 8 | Cross-Project Transfer | `cross_project` | Shared knowledge base enabling pattern and decision transfer between projects |
| 9 | Adaptive Ceremony | `adaptive_ceremony` | Dynamic ceremony level adjustment from real-time quality metrics during flow execution |
| 10 | Spec Memory | `spec_memory` | Persistent specification index with semantic search, access tracking, and pattern learning |

## Execution Flow

### Complete SDD Flow

```
                          User Request
                               |
                      +--------v---------+
                      | Effort Classifier |
                      | (pkg/intent)      |
                      | Analyze signals   |
                      +--------+---------+
                               |
                  +------------v-------------+
                  |  EffortClassification     |
                  |  level, confidence,       |
                  |  requires_debate,         |
                  |  requires_speckit         |
                  +------------+-------------+
                               |
                      +--------v---------+
                      | Ceremony Scaler   |
                      | (pkg/ceremony)    |
                      | Scale + boost     |
                      +--------+---------+
                               |
                  +------------v--------------+
                  |       FusionEngine         |
                  |       (pkg/engine)         |
                  |                            |
                  |  For each phase in order:  |
                  |  +----------------------+  |
                  |  |    SpecKit Pillar    |  |
                  |  | 1. Build topic       |  |
                  |  | 2. Execute debate    |  |
                  |  |    (via DebateFunc)  |  |
                  |  | 3. Score quality     |  |
                  |  | 4. Return result     |  |
                  |  +----------+-----------+  |
                  |             |               |
                  |  +----------v-----------+  |
                  |  |  Adaptive Ceremony   |  |
                  |  |  Adjust if needed    |  |
                  |  +----------+-----------+  |
                  |             |               |
                  +-------------+--------------+
                               |
              +----------------+----------------+
              |                                 |
     +--------v---------+            +----------v---------+
     | Superpowers Pillar|            |    GSD Pillar      |
     | (pkg/superpowers) |            |    (pkg/gsd)       |
     |                   |            |                    |
     | DispatchSubagents |            | CreateMilestones   |
     | (parallel tasks)  |            | (priority groups)  |
     |        |          |            |        |           |
     | ReviewCode        |            | TrackProgress      |
     | (quality check)   |            | (completion ratio) |
     +--------+----------+            +----------+---------+
              |                                  |
              +----------------+-----------------+
                               |
                      +--------v---------+
                      |   Spec Memory     |
                      |   (pkg/memory)    |
                      |  LearnFromFlow()  |
                      +--------+---------+
                               |
                      +--------v---------+
                      |   FlowResult      |
                      |  (quality score,  |
                      |   milestones,     |
                      |   specification)  |
                      +------------------+
```

### Effort-to-Ceremony Mapping

```
  Request Analysis              Ceremony Selection              Phase Execution
  ================              ==================              ===============

  Quick (<5 min)   -------->    Minimal          -------->    [Implement]
  fix typo, rename

  Medium (30-90m)  -------->    Light            -------->    [Specify]->[Plan]->[Implement]
  moderate feature

  Large (4-16h)    -------->    Standard         -------->    [Constitution]->[Specify]->[Clarify]->
  new system/API                                               [Plan]->[Tasks]->[Analyze]->[Implement]

  Epic (days-weeks) ------->    Heavy            -------->    All 7 phases + extended reviews
  full rewrite
```

### Injection Points

```
  HelixAgent                            HelixSpecifier
  ==========                            ==============

  debate_service  --SetDebateFunc()-->  SpecKit Pillar
  (real multi-LLM                       (receives DebateFunc,
   debate)                               calls during phases)

  enhanced_intent --RegisterClassifier-> FusionEngine
  _classifier()                          (delegates ClassifyEffort
                                          to injected function)

  speckit_        --ExecuteFlow()------>  FusionEngine
  orchestrator                            (full flow execution)
```

## Package Structure

```
HelixSpecifier/
├── pkg/
│   ├── types/
│   │   └── types.go                      # Core types: EffortLevel, FusionStage,
│   │                                     #   SpecKitPhase, CeremonyLevel,
│   │                                     #   MilestoneStatus, SpecSource,
│   │                                     #   DebateFunc, EffortClassifierFunc,
│   │                                     #   7 core interfaces, data structures
│   ├── config/
│   │   └── config.go                     # HELIX_SPECIFIER_* env var config
│   ├── engine/
│   │   └── engine.go                     # FusionEngine: 3-pillar orchestrator
│   ├── speckit/
│   │   └── speckit.go                    # SpecKit Pillar: 7-phase SDD workflow
│   ├── superpowers/
│   │   └── superpowers.go                # Superpowers Pillar: TDD + parallel
│   ├── gsd/
│   │   └── gsd.go                        # GSD Pillar: milestone lifecycle
│   ├── intent/
│   │   └── classifier.go                 # Signal-based effort classifier
│   ├── ceremony/
│   │   └── ceremony.go                   # Adaptive ceremony scaler
│   ├── memory/
│   │   └── memory.go                     # Spec memory store + learning
│   ├── adapters/
│   │   └── adapters.go                   # CLI agent output adapters
│   ├── metrics/
│   │   └── metrics.go                    # Operational metrics tracking
│   └── features/
│       ├── parallel_execution/parallel.go
│       ├── constitution_code/constitution.go
│       ├── nyquist_tdd/nyquist.go
│       ├── debate_architecture/debate.go
│       ├── skill_learning/learning.go
│       ├── brownfield/brownfield.go
│       ├── predictive_spec/predictive.go
│       ├── cross_project/transfer.go
│       ├── adaptive_ceremony/adaptive.go
│       └── spec_memory/specmem.go
├── tests/
│   ├── integration/                      # 26 cross-package workflow tests
│   ├── e2e/                              # 18 full workflow simulation tests
│   ├── security/                         # 26 input validation/thread safety tests
│   ├── stress/                           # 12 concurrent load/deadlock tests
│   ├── benchmark/                        # 22 performance measurement tests
│   └── automation/                       # 14 structure/API contract tests
├── go.mod                                # Module: digital.vasic.helixspecifier
├── Makefile                              # Build, test, lint targets
├── CLAUDE.md                             # Development instructions
└── README.md                             # Project documentation
```

## Key Interfaces

| Interface | Package | Methods | Purpose |
|-----------|---------|---------|---------|
| `SpecEngine` | `types` | ExecuteFlow, ResumeFlow, ClassifyEffort, GetFlowStatus, Health, Name, Version | Main fusion engine contract |
| `SpecKitPillar` | `types` | ExecutePhase, GetPhaseOrder, ValidateTransition | 7-phase SDD workflow |
| `SuperpowersPillar` | `types` | ExecuteWithTDD, DispatchSubagents, ReviewCode | TDD and parallel subagent execution |
| `GSDPillar` | `types` | CreateMilestones, TrackProgress, GetProgress | Milestone lifecycle management |
| `CeremonyScaler` | `types` | Scale, Adjust | Adaptive ceremony level determination |
| `SpecMemory` | `types` | Store, Search, LearnFromFlow | Specification storage and learning |
| `DebateFuncSetter` | `types` | SetDebateFunc | Optional debate function injection |
| `CLIAgentAdapter` | `types` | FormatOutput, ParseInput, GetAgentType, SupportsStreaming | CLI agent output formatting |

## Design Patterns

| Pattern | Usage |
|---------|-------|
| **Strategy** | Pillar interfaces (SpecKit, Superpowers, GSD) are interchangeable implementations |
| **Facade** | FusionEngine provides a unified interface over three pillars and supporting services |
| **Observer** | Metrics tracking records events from all components |
| **Factory** | `NewPillar()`, `NewStore()`, `NewClassifier()`, `NewGenericAdapter()` |
| **Template Method** | Phase execution follows a fixed sequence with pillar-specific behavior |
| **Adapter** | CLIAgentAdapter adapts engine output per agent type; MemoryStoreAdapter bridges module interfaces |
| **Semaphore** | Bounded concurrency in Superpowers (channel-based) and ParallelExecutor |
| **Dependency Injection** | DebateFunc and EffortClassifierFunc allow external behavior injection |

## Integration with HelixAgent

HelixSpecifier integrates with HelixAgent through several connection points:

1. **SpecKit Orchestrator** -- `internal/services/speckit_orchestrator.go` creates the FusionEngine, registers all three pillars and supporting services, injects the debate function from HelixAgent's debate service, and exposes the engine through HelixAgent's service layer.

2. **Enhanced Intent Classifier** -- `internal/services/enhanced_intent_classifier.go` registers HelixAgent's LLM-based intent classifier as the `EffortClassifierFunc`, replacing the built-in signal-based classifier with LLM-powered analysis.

3. **Debate Service Integration** -- The SpecKit pillar's `DebateFunc` is injected with a real multi-LLM debate execution function from HelixAgent's debate service (`internal/services/debate_service.go`). This enables specification phases to be refined through actual multi-agent debate with 5 positions across 5 LLMs.

4. **Auto-Activation** -- HelixAgent detects work granularity (single action, small creation, big creation, whole functionality, refactoring) and auto-activates HelixSpecifier for large changes that benefit from the full SDD workflow. Phase caching enables resumption of interrupted flows.

5. **Constitution Watcher** -- HelixAgent's `ConstitutionWatcher` monitors project changes and can trigger HelixSpecifier's Constitution phase when project structure changes are detected.

6. **CLI Agent Config** -- All 48 CLI agents in HelixAgent can receive HelixSpecifier-formatted output through the adapters package, with pre-configured adapters for OpenCode, Crush, Claude, Cursor, Windsurf, Copilot, Gemini, Aider, Cline, and Roo.
