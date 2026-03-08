# HelixSpecifier - Troubleshooting

**Module:** `digital.vasic.helixspecifier`

## Common Issues

### 1. Phases Produce Placeholder Output (Quality 0.75)

**Symptom:** All phase outputs say "placeholder" and quality scores
are exactly 0.75.

**Cause:** No `DebateFunc` has been injected into the SpecKit pillar.

**Solution:**

- When used standalone, inject a debate function via
  `engine.SetDebateFunc(fn)`
- When used within HelixAgent, the debate service automatically injects
  the function through `speckit_orchestrator.go`
- Verify the debate service is initialized before HelixSpecifier

### 2. Effort Classification Always Returns Medium

**Symptom:** All requests are classified as Medium effort with 0.5
confidence.

**Possible causes:**

| Cause | Solution |
|-------|----------|
| No classifier registered | Call `engine.RegisterClassifier(fn)` |
| Default classifier active | The built-in `intent.Classifier` returns Medium as default when no keyword signals are detected |
| HelixAgent classifier not injected | Verify `enhanced_intent_classifier.go` registers the LLM-based classifier |

**Tip:** Use explicit keywords in requests. The signal-based classifier
looks for: "fix typo", "rename" (Quick), "implement service", "build API"
(Large), "entire system", "complete rewrite" (Epic).

### 3. Flow Execution Hangs

**Symptom:** `ExecuteFlow` does not return.

**Possible causes:**

| Cause | Solution |
|-------|----------|
| Debate function blocks | Ensure the injected `DebateFunc` respects `context.Context` cancellation |
| Subagent deadlock | Check `HELIX_SPECIFIER_MAX_PARALLEL_AGENTS` (default: 4). Reduce if resource-constrained |
| Phase quality loop | If quality never meets threshold with adaptive ceremony, the engine may keep increasing ceremony. Set `HELIX_SPECIFIER_ADAPTIVE_CEREMONY=false` to disable |

### 4. Cache Directory Not Created

**Symptom:** Errors about missing cache directory when
`CACHE_ENABLED=true`.

**Solution:**

```bash
mkdir -p .helixspecifier/cache
```

Or set a custom directory:

```bash
export HELIX_SPECIFIER_CACHE_DIR=/tmp/helixspecifier-cache
```

### 5. Flow Resumption Fails

**Symptom:** `ResumeFlow` returns "flow not found" error.

**Possible causes:**

| Cause | Solution |
|-------|----------|
| Caching disabled | Set `HELIX_SPECIFIER_CACHE_ENABLED=true` |
| Cache cleared | Cached phase results are stored in `CACHE_DIR`; do not delete during execution |
| Flow ID mismatch | Flow IDs are UUID-based; verify the exact ID from the original `FlowResult` |

### 6. Adaptive Ceremony Not Adjusting

**Symptom:** Ceremony level stays constant despite varying quality scores.

**Checklist:**

1. Verify `HELIX_SPECIFIER_ADAPTIVE_CEREMONY=true`
2. Adjustment only occurs between phases, not after the flow completes
3. Thresholds: quality > 0.85 triggers reduction, quality < 0.50
   triggers increase
4. Ceremony only adjusts by one level at a time (Heavy to Standard,
   Light to Standard)

### 7. Milestone Tracking Shows No Progress

**Symptom:** `GSDPillar.GetProgress()` returns 0% after tasks complete.

**Solutions:**

- Ensure `TrackProgress` is called with the actual `TaskResult` objects
- Milestones transition: `queued` -> `active` -> `completed`
- A milestone requires all its tasks to complete before it transitions

### 8. CLI Agent Adapter Produces Wrong Format

**Symptom:** Output is not formatted correctly for the target CLI agent.

**Solutions:**

| Agent | Expected Format | Adapter Name |
|-------|----------------|--------------|
| OpenCode | Markdown | `opencode` |
| Crush | Markdown | `crush` |
| Copilot | GitHub-flavored | `copilot` |
| Gemini | JSON | `gemini` |
| Aider | JSON | `aider` |

Use `adapters.WellKnownAdapters()` to see all 10 pre-configured adapters.

### 9. Spec Memory Search Returns No Results

**Symptom:** `SpecMemory.Search()` returns empty results.

**Possible causes:**

| Cause | Solution |
|-------|----------|
| Memory disabled | Set `HELIX_SPECIFIER_SPEC_MEMORY_ENABLED=true` |
| No flows completed | Only completed flows are stored in memory |
| Query mismatch | Search uses case-insensitive title/description matching; try broader terms |

### 10. High Memory Usage Under Load

**Symptom:** Memory consumption grows unbounded during stress testing.

**Solutions:**

- Reduce `HELIX_SPECIFIER_MAX_PARALLEL_TASKS` (default: 8)
- Reduce `HELIX_SPECIFIER_MAX_PARALLEL_AGENTS` (default: 4)
- The in-memory spec store grows with completed flows; restart the
  process to reclaim memory, or implement a bounded store

## Logging

HelixSpecifier components log structured messages with these fields:

| Field | Description |
|-------|-------------|
| `component` | `engine`, `speckit`, `superpowers`, `gsd`, `ceremony`, `intent` |
| `flow_id` | Flow identifier |
| `phase` | Current SpecKit phase |
| `effort` | Classified effort level |
| `ceremony` | Current ceremony level |
| `quality` | Phase quality score |

## Resource Limits

Per project standards, all tests and challenges use:

- `GOMAXPROCS=2`
- `nice -n 19`
- `ionice -c 3`
- `-p 1` for `go test`

These limits apply to HelixSpecifier's test suites (835+ tests).

## Getting Help

1. Check [CONFIGURATION.md](CONFIGURATION.md) for all env variables
2. Review [ARCHITECTURE.md](ARCHITECTURE.md) for the execution flow
3. Inspect metrics via `pkg/metrics` snapshot support
4. Run the challenge: `./challenges/scripts/helixspecifier_challenge.sh`
