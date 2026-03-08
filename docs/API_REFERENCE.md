# HelixSpecifier - API Reference

**Module:** `digital.vasic.helixspecifier`

## Core Interfaces (`pkg/types`)

### SpecEngine

Main fusion engine contract.

```go
type SpecEngine interface {
    ExecuteFlow(ctx context.Context, request FlowRequest) (*FlowResult, error)
    ResumeFlow(ctx context.Context, flowID string) (*FlowResult, error)
    ClassifyEffort(request string) *EffortClassification
    GetFlowStatus(flowID string) (*FlowStatus, error)
    Health() error
    Name() string
    Version() string
}
```

### SpecKitPillar

7-phase SDD workflow execution.

```go
type SpecKitPillar interface {
    ExecutePhase(ctx context.Context, phase SpecKitPhase, input PhaseInput) (*PhaseResult, error)
    GetPhaseOrder(ceremony CeremonyLevel) []SpecKitPhase
    ValidateTransition(from, to SpecKitPhase) bool
}
```

### SuperpowersPillar

TDD discipline and parallel subagent dispatch.

```go
type SuperpowersPillar interface {
    ExecuteWithTDD(ctx context.Context, tasks []*Task) ([]*TaskResult, error)
    DispatchSubagents(ctx context.Context, tasks []*Task) ([]*TaskResult, error)
    ReviewCode(ctx context.Context, results []*TaskResult) (*ReviewResult, error)
}
```

### GSDPillar

Milestone lifecycle management.

```go
type GSDPillar interface {
    CreateMilestones(ctx context.Context, tasks []*Task) ([]*Milestone, error)
    TrackProgress(ctx context.Context, milestones []*Milestone, results []*TaskResult) error
    GetProgress(ctx context.Context) (*Progress, error)
}
```

### CeremonyScaler

Adaptive ceremony level determination.

```go
type CeremonyScaler interface {
    Scale(effort *EffortClassification) CeremonyLevel
    Adjust(current CeremonyLevel, qualityScores []float64) CeremonyLevel
}
```

### SpecMemory

Specification storage and cross-project learning.

```go
type SpecMemory interface {
    Store(ctx context.Context, spec *Specification) error
    Search(ctx context.Context, query string) ([]*Specification, error)
    LearnFromFlow(ctx context.Context, result *FlowResult) error
}
```

### CLIAgentAdapter

Output formatting for CLI agents.

```go
type CLIAgentAdapter interface {
    FormatOutput(result *FlowResult) (string, error)
    ParseInput(raw string) (*FlowRequest, error)
    GetAgentType() string
    SupportsStreaming() bool
}
```

### DebateFuncSetter

Optional debate function injection.

```go
type DebateFuncSetter interface {
    SetDebateFunc(fn DebateFunc)
}
```

## Function Types

```go
type DebateFunc func(ctx context.Context, topic string, rounds int,
    metadata map[string]interface{}) (output string, score float64, debateID string, err error)

type EffortClassifierFunc func(request string) *EffortClassification
```

## Data Types

### FlowRequest / FlowResult

```go
type FlowRequest struct {
    Description string
    UserID      string
    ProjectID   string
    Context     map[string]interface{}
}

type FlowResult struct {
    FlowID       string
    Status       string
    QualityScore float64
    PhaseResults []*PhaseResult
    Milestones   []*Milestone
    Specification *Specification
    Duration     time.Duration
}
```

### PhaseInput / PhaseResult

```go
type PhaseInput struct {
    Request       string
    PreviousPhase *PhaseResult
    Context       map[string]interface{}
}

type PhaseResult struct {
    Phase        SpecKitPhase
    Output       string
    QualityScore float64
    DebateID     string
    Duration     time.Duration
    Metadata     map[string]interface{}
}
```

### EffortClassification

```go
type EffortClassification struct {
    Level          EffortLevel
    Confidence     float64
    CeremonyLevel  CeremonyLevel
    RequiresDebate bool
    RequiresSpecKit bool
}
```

### Task / TaskResult

```go
type Task struct {
    ID          string
    Name        string
    Description string
    Priority    string  // critical, high, medium, low
    Dependencies []string
}

type TaskResult struct {
    TaskID  string
    Status  string
    Output  string
    Score   float64
    Error   error
}
```

### Milestone

```go
type Milestone struct {
    ID           string
    Name         string
    Tasks        []*Task
    Dependencies []string
    Status       MilestoneStatus  // queued, active, completed, blocked, failed
    Progress     float64
}
```

## Enums

### EffortLevel

| Constant | Value | Duration |
|----------|-------|----------|
| `EffortQuick` | `"quick"` | < 5 minutes |
| `EffortMedium` | `"medium"` | 30-90 minutes |
| `EffortLarge` | `"large"` | 4-16 hours |
| `EffortEpic` | `"epic"` | Days to weeks |

### CeremonyLevel

| Constant | Value | Phases |
|----------|-------|--------|
| `CeremonyMinimal` | `"minimal"` | Implement only |
| `CeremonyLight` | `"light"` | Specify + Plan + Implement |
| `CeremonyStandard` | `"standard"` | All 7 phases |
| `CeremonyHeavy` | `"heavy"` | All 7 + extended reviews |

### SpecKitPhase

| Constant | Value | Order |
|----------|-------|-------|
| `PhaseConstitution` | `"constitution"` | 1 |
| `PhaseSpecify` | `"specify"` | 2 |
| `PhaseClarify` | `"clarify"` | 3 |
| `PhasePlan` | `"plan"` | 4 |
| `PhaseTasks` | `"tasks"` | 5 |
| `PhaseAnalyze` | `"analyze"` | 6 |
| `PhaseImplement` | `"implement"` | 7 |

### FusionStage

| Constant | Value |
|----------|-------|
| `StageIdeation` | `"ideation"` |
| `StageConstitution` | `"constitution"` |
| `StagePlanning` | `"planning"` |
| `StageExecution` | `"execution"` |
| `StageVerification` | `"verification"` |

### MilestoneStatus

| Constant | Value |
|----------|-------|
| `MilestoneQueued` | `"queued"` |
| `MilestoneActive` | `"active"` |
| `MilestoneCompleted` | `"completed"` |
| `MilestoneBlocked` | `"blocked"` |
| `MilestoneFailed` | `"failed"` |

## Metrics (`pkg/metrics`)

Thread-safe operational metrics:

| Metric | Description |
|--------|-------------|
| `FlowsStarted` | Total flows initiated |
| `FlowsCompleted` | Successfully completed flows |
| `FlowsFailed` | Failed flows |
| `SuccessRate` | `Completed / Started` |
| `PhaseMetrics[phase]` | Per-phase: executions, successes, failures, avg score, total duration |
| `EffortDistribution` | Count per effort level |
| `CeremonyDistribution` | Count per ceremony level |
