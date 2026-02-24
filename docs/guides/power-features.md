# Power Features Guide

HelixSpecifier includes 10 power features, each implemented as an
independent package under `pkg/features/`. These features can be
used standalone or as part of the full FusionEngine workflow.

---

## 1. Parallel Execution

**Package**: `pkg/features/parallel_execution`

Provides bounded-concurrency parallel task execution with per-task
timeouts.

### Key Types

- `Executor` -- Manages a pool of workers with semaphore-based
  concurrency control
- `TaskFunc` -- Function signature `func(ctx context.Context)
  (string, error)` for executable tasks
- `Result` -- Contains task ID, output, duration, and error

### How It Works

The executor accepts a map of named tasks and runs them concurrently,
limited by `maxWorkers`. Each task receives its own context with
the configured timeout. Results are collected in a thread-safe
slice.

### Usage

```go
exec := parallel_execution.NewExecutor(4, 30*time.Second)
tasks := map[string]parallel_execution.TaskFunc{
    "compile": func(ctx context.Context) (string, error) {
        return "compiled", nil
    },
    "lint": func(ctx context.Context) (string, error) {
        return "linted", nil
    },
}
results := exec.Execute(context.Background(), tasks)
completed, failed := exec.Stats()
```

---

## 2. Constitution as Code

**Package**: `pkg/features/constitution_code`

Manages machine-readable project constitutions with mandatory rule
enforcement and compliance validation.

### Key Types

- `Constitution` -- Thread-safe collection of rules
- `Rule` -- Individual rule with ID, category, priority, title,
  description, and mandatory flag

### How It Works

Rules are added to the constitution with unique IDs. Mandatory rules
are enforced during validation: the `Validate()` method checks that
the text mentions each mandatory rule's title keyword. Violations
are returned as a list of rule IDs.

### Usage

```go
c := constitution_code.NewConstitution()
c.AddRule(constitution_code.Rule{
    ID:        "testing",
    Category:  "quality",
    Priority:  1,
    Title:     "Test Coverage",
    Mandatory: true,
})
violations := c.Validate("Our plan includes test coverage")
// violations is empty because "test coverage" is mentioned
```

---

## 3. Nyquist TDD

**Package**: `pkg/features/nyquist_tdd`

Enforces a test-to-implementation ratio of at least 2.0, inspired
by the Nyquist-Shannon sampling theorem: to fully "sample" the
quality of implementation code, tests must cover at least twice the
implementation surface.

### Key Types

- `Tracker` -- Records implementation and test lines, checks
  compliance

### How It Works

As implementation and test lines are added, the tracker maintains
running totals. The `Ratio()` method returns `testLines / implLines`.
`IsCompliant()` returns true when the ratio is >= 2.0. `Check()`
records a violation if the ratio is below threshold.

### Usage

```go
tracker := nyquist_tdd.NewTracker()
tracker.RecordImplementation(100)
tracker.RecordTests(250)
ratio := tracker.Ratio()       // 2.50
compliant := tracker.IsCompliant() // true
err := tracker.Check()          // nil (compliant)
```

---

## 4. Debate Architecture

**Package**: `pkg/features/debate_architecture`

Provides multi-round debate orchestration for specification
refinement. Multiple AI agents argue positions, score arguments,
and the best position wins each round.

### Key Types

- `Debate` -- Manages rounds with a configurable maximum
- `Round` -- Contains positions from multiple agents with a winner
- `Position` -- An agent's argument with a quality score

### How It Works

A debate is created with a topic and maximum round count. Rounds
are added sequentially, each containing positions from participating
agents. `BestPosition()` returns the highest-scored position across
all rounds. `IsComplete()` returns true when all rounds are done.

### Usage

```go
debate := debate_architecture.NewDebate("d1", "API design", 3)
debate.AddRound(debate_architecture.Round{
    Positions: []debate_architecture.Position{
        {AgentID: "agent1", Argument: "REST", Score: 0.8},
        {AgentID: "agent2", Argument: "gRPC", Score: 0.9},
    },
    Winner: "agent2",
})
best := debate.BestPosition() // agent2's gRPC position
```

---

## 5. Skill Learning

**Package**: `pkg/features/skill_learning`

Provides adaptive skill proficiency tracking. Skills improve over
time based on quality feedback using a running-average algorithm.

### Key Types

- `Learner` -- Registry of skills with proficiency tracking
- `Skill` -- Named skill with domain, proficiency score, usage
  count, and last-used timestamp

### How It Works

Skills are registered with initial proficiency. Each usage records
a quality score, and proficiency is updated as a running average:
`new = old + (score - old) / count`. `TopSkills()` returns the
highest-proficiency skills.

### Usage

```go
learner := skill_learning.NewLearner()
learner.RegisterSkill(skill_learning.Skill{
    ID: "go-api", Name: "Go API Design", Domain: "backend",
})
learner.RecordUsage("go-api", 0.9)
learner.RecordUsage("go-api", 0.85)
skill, _ := learner.GetSkill("go-api")
// skill.Proficiency is ~0.875
```

---

## 6. Brownfield Analysis

**Package**: `pkg/features/brownfield`

Analyzes existing codebases to detect patterns and dependencies,
informing specification generation with real-world constraints.

### Key Types

- `Analyzer` -- Records patterns and dependencies from code
- `CodePattern` -- Pattern with name, frequency, and confidence
- `AnalysisResult` -- Aggregated patterns, dependencies,
  complexity, and lines of code

### How It Works

As the codebase is scanned, patterns are recorded with confidence
scores. Repeated patterns increase in frequency with running-average
confidence. Dependencies are tracked as a set. `Analyze()` returns
the complete analysis result.

### Usage

```go
analyzer := brownfield.NewAnalyzer()
analyzer.RecordPattern("repository-pattern", 0.9)
analyzer.RecordPattern("dependency-injection", 0.85)
analyzer.AddDependency("github.com/gin-gonic/gin")
result := analyzer.Analyze()
// result.Patterns, result.Dependencies, result.Complexity
```

---

## 7. Predictive Specification

**Package**: `pkg/features/predictive_spec`

Predicts future requirements based on accumulated historical events.
Predictions have decreasing confidence for older events.

### Key Types

- `Predictor` -- Accumulates history and generates predictions
- `Prediction` -- Predicted requirement with confidence and basis

### How It Works

Historical events are added to the predictor. When `Predict()` is
called, each event generates a prediction with confidence starting
at 1.0 for the most recent event and decreasing by 0.1 per older
event (minimum 0.1). `HighConfidence()` filters predictions above
a threshold.

### Usage

```go
predictor := predictive_spec.NewPredictor()
predictor.AddHistory("Added authentication module")
predictor.AddHistory("Added user profiles")
predictions := predictor.Predict()
highConf := predictor.HighConfidence(0.8)
```

---

## 8. Cross-Project Transfer

**Package**: `pkg/features/cross_project`

Maintains a shared knowledge base for transferring patterns,
decisions, and lessons learned between projects.

### Key Types

- `TransferEngine` -- Stores and retrieves cross-project knowledge
- `Knowledge` -- A transferable piece of knowledge with project,
  category, content, tags, and relevance

### How It Works

Knowledge entries from various projects are stored with tags and
relevance scores. `Search()` finds entries matching a query against
content and tags. `ForProject()` returns all knowledge from a
specific project.

### Usage

```go
te := cross_project.NewTransferEngine()
te.Add(cross_project.Knowledge{
    ID:       "k1",
    Project:  "project-alpha",
    Category: "architecture",
    Content:  "Event-driven architecture works well for...",
    Tags:     []string{"events", "microservices"},
})
results := te.Search("event-driven", 5)
```

---

## 9. Adaptive Ceremony

**Package**: `pkg/features/adaptive_ceremony`

Dynamically adjusts ceremony levels during flow execution based on
real-time quality metrics.

### Key Types

- `Adjuster` -- Evaluates phase results and adjusts ceremony

### How It Works

The adjuster is configured with two thresholds:

- `upThreshold` (default 0.9): If average quality exceeds this,
  ceremony is reduced by one level
- `downThreshold` (default 0.5): If average quality falls below
  this, ceremony is increased by one level

The `Adjust()` method evaluates all phase results and returns the
new ceremony level. `AdjustmentCount()` tracks how many adjustments
have been made.

### Usage

```go
adjuster := adaptive_ceremony.NewAdjuster(0.9, 0.5)
newLevel := adjuster.Adjust(
    types.CeremonyStandard,
    []types.PhaseResult{
        {QualityScore: 0.95},
        {QualityScore: 0.92},
    },
)
// newLevel is CeremonyLight (reduced due to high quality)
```

---

## 10. Spec Memory

**Package**: `pkg/features/spec_memory`

Provides persistent specification memory with semantic indexing,
access tracking, and rapid retrieval of historical specifications.

### Key Types

- `Index` -- In-memory specification index with CRUD operations
- `Entry` -- Specification entry with title, content, tags, score,
  access count, and metadata

### How It Works

Entries are stored in an in-memory index keyed by ID. `Get()`
retrieves an entry and increments its access count. `Search()` finds
entries whose title or content contains the query string.
`MostAccessed()` returns the entry with the highest access count.
`Remove()` deletes an entry.

### Usage

```go
idx := spec_memory.NewIndex()
idx.Add(spec_memory.Entry{
    ID:      "spec-auth",
    Title:   "Authentication Specification",
    Content: "JWT-based auth with refresh tokens...",
    Tags:    []string{"auth", "jwt", "security"},
    Score:   0.9,
})
results := idx.Search("authentication", 5)
entry, _ := idx.Get("spec-auth")
// entry.AccessCount is now 1
```

---

## Feature Independence

All 10 power features are designed as independent packages with
no cross-dependencies. Each can be imported and used standalone
without the FusionEngine or any other HelixSpecifier package.
This supports the HelixAgent principle of comprehensive decoupling.
