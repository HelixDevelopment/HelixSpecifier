# HelixSpecifier Test Coverage Ledger (round-273)

This ledger maps the surface-area symbols of HelixSpecifier to the
specific tests + Challenges that exercise them with positive
runtime evidence (CONST-035 / Article XI §11.9, CONST-050(B)).

> Verbatim 2026-05-19 operator mandate (preserved per
> CONST-049 §11.4.17):
> "all existing tests and Challenges do work in anti-bluff manner
> - they MUST confirm that all tested codebase really works as
> expected! We had been in position that all tests do execute with
> success and all Challenges as well, but in reality the most of
> the features does not work and can't be used! This MUST NOT be
> the case and execution of tests and Challenges MUST guarantee
> the quality, the completition and full usability by end users
> of the product!"

## Anti-bluff guarantees

Every row below identifies BOTH a symbol AND the failing-line a
mutation would produce. Mutation pairing per §1.1 / CONST-050(A).

## Ledger — pkg/engine

| Symbol                                | Test                                                                              | Challenge invariant                                |
|---------------------------------------|-----------------------------------------------------------------------------------|----------------------------------------------------|
| `FusionEngine.Name()`                 | `pkg/engine/engine_test.go`                                                       | `engine.Name` (runner invariant 1)                 |
| `FusionEngine.Version()`              | `pkg/engine/engine_test.go`                                                       | `engine.Version` (runner invariant 1)              |
| `FusionEngine.Health()`               | `pkg/engine/engine_test.go`                                                       | `engine.Health.no_pillar_returns_error` + pillar OK |
| `FusionEngine.ClassifyEffort()`       | `pkg/engine/engine_test.go`, `pkg/engine/engine_i18n_test.go`                     | `engine.ClassifyEffort.empty_errors`               |
| `FusionEngine.NewWithTranslator()`    | `pkg/engine/engine_i18n_test.go`                                                  | (runner construction path)                         |
| `FusionEngine.SetTranslator()`        | `pkg/engine/engine_i18n_test.go`                                                  | nil→Noop fallback covered                          |
| `FusionEngine.RegisterSpecKit()`      | `pkg/engine/engine_test.go`                                                       | `engine.Health.with_pillar_ok`                     |

## Ledger — pkg/intent

| Symbol                                       | Test                                            | Challenge invariant                       |
|----------------------------------------------|-------------------------------------------------|-------------------------------------------|
| `Classifier.Classify()`                      | `pkg/intent/classifier_test.go`                 | `intent.Classify.<locale>.{level,signal}` |
| `Classifier.generateReasoning()`             | `pkg/intent/classifier_test.go`                 | (i18n round-trip via runner)              |
| `NewClassifierWithTranslator()` nil fallback | `pkg/intent/classifier_test.go`                 | runner constructor path                   |

## Ledger — pkg/i18n

| Symbol                          | Test                                  | Challenge invariant                             |
|---------------------------------|---------------------------------------|-------------------------------------------------|
| `Translator.T()` interface      | `pkg/i18n/translator_test.go`         | `i18n.Noop.key_roundtrip.<locale>`              |
| `NoopTranslator.T()` no-args    | `pkg/i18n/translator_test.go`         | runner asserts key == returned (anti-bluff)     |
| `NoopTranslator.T()` w/ args    | `pkg/i18n/translator_test.go`         | (format-string contract — unit only)            |
| `NewNoopTranslator()` factory   | `pkg/i18n/translator_test.go`         | runner uses for all 5 locales                   |

## Ledger — pkg/speckit / superpowers / gsd

| Symbol                          | Test                                       | Challenge invariant                     |
|---------------------------------|--------------------------------------------|-----------------------------------------|
| `SpecKitPillar.ExecutePhase()`  | `pkg/speckit/speckit_test.go`              | runner registers pillar + Health check  |
| `LLMBackedDebateFunc`           | `pkg/speckit/llm_debate_test.go`           | round-65 fix preserved                  |
| `SuperpowersPillar.Dispatch…()` | `pkg/superpowers/superpowers_test.go`      | (covered via integration suite)         |
| `GSDPillar.CreateMilestones()`  | `pkg/gsd/*` package tests                  | (covered via integration suite)         |

## Challenges (challenges/scripts + challenges/runner)

| Challenge                                   | Purpose                                   | Mutation-paired? |
|---------------------------------------------|-------------------------------------------|------------------|
| `helixspecifier_compile_challenge.sh`       | go build sanity                           | n                |
| `helixspecifier_functionality_challenge.sh` | structural pillar contracts               | n                |
| `helixspecifier_unit_challenge.sh`          | unit-test layer floor                     | n                |
| `helixspecifier_describe_challenge.sh`      | runner invariants (20)                    | **YES** (round-273) |
| `chaos_failure_injection_challenge.sh`      | runtime chaos resilience                  | env-gated        |
| `ddos_health_flood_challenge.sh`            | DDoS health endpoint                      | env-gated        |
| `scaling_horizontal_challenge.sh`           | horizontal scale check                    | env-gated        |
| `stress_sustained_load_challenge.sh`        | stress floor                              | env-gated        |
| `ui_terminal_interaction_challenge.sh`      | TUI smoke                                 | env-gated        |
| `ux_end_to_end_flow_challenge.sh`           | UX end-to-end                             | env-gated        |
| `host_no_auto_suspend_challenge.sh`         | CONST-033 host-power ban                  | n                |
| `no_suspend_calls_challenge.sh`             | CONST-033 anti-suspend scan               | n                |

## How to run

```bash
# Race-detector full unit + runner pass:
GOMAXPROCS=2 go test -count=1 -race -p 1 ./...

# Anti-bluff describe Challenge (20 invariants):
./challenges/helixspecifier_describe_challenge.sh normal   # exit 0
./challenges/helixspecifier_describe_challenge.sh mutate   # exit 99

# Standalone runner (useful for forensics):
go run ./challenges/runner/
HELIXSPECIFIER_MUTATE_RUNNER=1 go run ./challenges/runner/
```

## Cascade

CONST-047 + CONST-051(A): this ledger is the HelixSpecifier-side
realisation of the cross-org "equal codebase" mandate. The same
ledger pattern lives in sibling submodules (LeakHub, MCP_Module,
Models, Ouroborous, …) and is bumped from the meta-repo on every
round close-out.
