# Getting Started with HelixSpecifier

## Prerequisites

- Go 1.24 or later
- `goimports` and `golangci-lint` (optional, for development)

## Installation

### As a Submodule (Recommended for HelixAgent)

HelixSpecifier is designed to be used as a Git submodule within the
HelixAgent project:

```bash
cd /path/to/HelixAgent
git submodule add <repo-url> HelixSpecifier
git submodule update --init --recursive
```

Add a replace directive in the root `go.mod`:

```
require digital.vasic.helixspecifier v0.0.0

replace digital.vasic.helixspecifier => ./HelixSpecifier
```

### As a Standalone Module

For use outside HelixAgent:

```bash
go get digital.vasic.helixspecifier@latest
```

## Building

Build all packages to verify compilation:

```bash
cd HelixSpecifier
make build
```

Or directly:

```bash
go build ./...
```

## Running Tests

Run the full test suite with resource limits:

```bash
make test
```

With race detection:

```bash
make test-race
```

Generate a coverage report:

```bash
make test-cover
# Open coverage.html in a browser
```

## Basic Usage

### Step 1: Create Configuration

Configuration is loaded from environment variables. For defaults:

```go
cfg := config.DefaultConfig()
```

Or from environment:

```go
cfg := config.FromEnv()
```

### Step 2: Initialize the Engine

```go
logger := logrus.New()
eng := engine.New(cfg, logger)
```

### Step 3: Register Pillars

All three pillars must be registered before executing flows:

```go
eng.RegisterSpecKit(speckit.NewPillar(cfg, logger))
eng.RegisterSuperpowers(superpowers.NewPillar(cfg, logger))
eng.RegisterGSD(gsd.NewPillar(logger))
```

Optional components:

```go
eng.RegisterCeremonyScaler(ceremony.NewScaler(cfg, logger))
eng.RegisterSpecMemory(memory.NewStore())
```

### Step 4: Connect Debate Service (Optional)

For debate-backed specification, inject the debate function:

```go
sk := speckit.NewPillar(cfg, logger)
sk.SetDebateFunc(func(
    ctx context.Context,
    topic string,
    rounds int,
    metadata map[string]interface{},
) (string, float64, string, error) {
    // Your debate service implementation
    return output, score, debateID, nil
})
eng.RegisterSpecKit(sk)
```

Without a debate function, the SpecKit pillar produces placeholder
output with a default quality score of 0.75.

### Step 5: Classify and Execute

```go
ctx := context.Background()

// Classify the effort level
classification, err := eng.ClassifyEffort(
    ctx,
    "Build a REST API with JWT authentication",
)
if err != nil {
    log.Fatal(err)
}

// Execute the flow
result, err := eng.ExecuteFlow(ctx, request, classification)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Flow: %s\n", result.FlowID)
fmt.Printf("Quality: %.2f\n", result.OverallQualityScore)
fmt.Printf("Phases: %d\n", len(result.PhaseResults))
```

### Step 6: Format Output for an Agent

```go
adapter := adapters.NewGenericAdapter(
    "myagent", adapters.CategoryMarkdown,
)
output, err := adapter.FormatOutput(result)
if err != nil {
    log.Fatal(err)
}
fmt.Println(output)
```

## Using the Intent Classifier Standalone

The effort classifier can be used independently:

```go
classifier := intent.NewClassifier()
result := classifier.Classify(
    "Refactor the authentication module",
)
fmt.Printf("Level: %s\n", result.Level)
fmt.Printf("Ceremony: %s\n", result.CeremonyLevel)
fmt.Printf("Requires SpecKit: %v\n", result.RequiresSpecKit)
```

## Using Power Features Standalone

Each power feature is an independent package:

### Parallel Execution

```go
exec := parallel_execution.NewExecutor(4, 30*time.Second)
tasks := map[string]parallel_execution.TaskFunc{
    "task1": func(ctx context.Context) (string, error) {
        return "done", nil
    },
    "task2": func(ctx context.Context) (string, error) {
        return "done", nil
    },
}
results := exec.Execute(context.Background(), tasks)
```

### Nyquist TDD

```go
tracker := nyquist_tdd.NewTracker()
tracker.RecordImplementation(100)
tracker.RecordTests(250)
fmt.Printf("Ratio: %.2f\n", tracker.Ratio())
fmt.Printf("Compliant: %v\n", tracker.IsCompliant())
```

### Constitution as Code

```go
constitution := constitution_code.NewConstitution()
constitution.AddRule(constitution_code.Rule{
    ID:        "test-coverage",
    Category:  "quality",
    Priority:  1,
    Title:     "100% Test Coverage",
    Mandatory: true,
})
violations := constitution.Validate(
    "We have test coverage for all components",
)
```

## Flow Resumption

Interrupted flows can be resumed:

```go
// First execution (may be interrupted)
result, err := eng.ExecuteFlow(ctx, request, classification)
flowID := result.FlowID

// Later, resume from where it stopped
resumed, err := eng.ResumeFlow(ctx, flowID, request)
```

## Environment Variables

Copy the example environment file and adjust values:

```bash
cp .env.example .env
```

See `.env.example` for the complete list of configuration variables
with defaults and descriptions.

## Next Steps

- Read the [Architecture Guide](architecture.md) for a deep dive
  into the 3-pillar fusion design
- Read the [Power Features Guide](power-features.md) for details
  on all 10 power features
- See `CLAUDE.md` for developer conventions and interface reference
- See `AGENTS.md` for CLI agent integration details
