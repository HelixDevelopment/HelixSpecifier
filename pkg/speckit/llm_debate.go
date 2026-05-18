// Package speckit — round-65 §11.4 anti-bluff fix.
//
// This file provides a concrete, wireable DebateFunc implementation
// (LLMBackedDebateFunc) that orchestrates real LLM-backed debates
// for the 7-phase Spec-Driven Development workflow.
//
// Round-28 §11.4 audit (2026-05-17, commit 4e66ff5) removed the
// nil-DebateFunc fabrication branch from ExecutePhase that returned
// `[<phase>] Phase output for: <request>` with a hardcoded 0.75
// score and Success=true — a PASS-bluff at the baseline-runner-
// class default layer. The fix introduced ErrDebateFuncNotConfigured
// (returned when no DebateFunc is wired) but left consumers without
// a ready-made concrete implementation.
//
// Round-65 (this file, 2026-05-18) closes the round-28 sentinel by
// providing LLMBackedDebateFunc — a dependency-light constructor
// that adapts any LLMResponder (minimal Generate-style interface)
// into a real types.DebateFunc. Consumers wire it via:
//
//     responder := http_responder.New(httpClient, endpoint)  // any impl
//     pillar.SetDebateFunc(speckit.LLMBackedDebateFunc(responder))
//
// The implementation:
//   - builds a structured debate prompt (FOR / AGAINST / SYNTHESIS)
//     that scales to the rounds count requested by ExecutePhase;
//   - invokes responder.Generate(ctx, prompt) — real HTTP, real
//     model, no simulation;
//   - parses the response to derive a heuristic QualityScore in
//     [0.0, 1.0] driven by presence of structural debate markers
//     and response length relative to the rounds budget;
//   - returns the raw debate transcript as the phase output so the
//     downstream PhaseResult.Output carries real model content.
//
// Constitutional anchors: CONST-035 (anti-bluff), CONST-050(A)
// (no-fakes-beyond-unit-tests — this file is production code, never
// imports test mocks), Article XI §11.9 (forensic anchor — every
// returned QualityScore is derived from captured model output, not
// fabricated).
package speckit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"digital.vasic.helixspecifier/pkg/types"
)

// LLMResponder is the minimal interface HelixSpecifier requires
// from any LLM client to back a real DebateFunc. Consumers inject
// their own implementation (e.g. an HTTP-backed responder from the
// LLMOps round-62 work, an Ollama client, an OpenAI client, etc.).
//
// Keeping this interface minimal (single Generate method) is a
// deliberate CONST-051(B) decoupling choice: HelixSpecifier MUST
// remain project-not-aware and fully reusable; coupling to a
// concrete LLM SDK would violate that mandate.
type LLMResponder interface {
	// Generate sends prompt to the underlying model and returns
	// the raw response text. Implementations MUST honour ctx
	// cancellation and MUST surface upstream errors verbatim
	// (wrapping with %w is acceptable, swallowing is not).
	Generate(ctx context.Context, prompt string) (string, error)
}

// Round-65 sentinels — distinguishable from the round-28
// ErrDebateFuncNotConfigured so callers can branch precisely.
//
//   - ErrLLMDebateResponderNotConfigured: a DebateFunc produced by
//     LLMBackedDebateFunc was invoked but the underlying responder
//     was nil at construction time (defensive — the constructor
//     itself rejects nil responders, so this only fires if a caller
//     bypasses the constructor).
//   - ErrLLMDebateEmptyResponse: the responder returned a non-error
//     empty string. Per CONST-035 we refuse to fabricate output
//     from an empty model response.
//   - ErrLLMDebateScoreCalculationFailed: the heuristic scorer
//     produced a NaN / out-of-range value (defensive — current
//     scorer is total-function but the sentinel is reserved for
//     future pluggable scorers).
var (
	ErrLLMDebateResponderNotConfigured = fmt.Errorf("helixspecifier speckit: LLMBackedDebateFunc invoked with nil LLMResponder — pass a non-nil responder to LLMBackedDebateFunc (round-65 §11.4 anti-bluff: refusing to fabricate debate output without a real model)")
	ErrLLMDebateEmptyResponse          = fmt.Errorf("helixspecifier speckit: LLMResponder.Generate returned empty string — refusing to fabricate a phase output / quality score from an empty model response (round-65 §11.4 anti-bluff: CONST-035)")
	ErrLLMDebateScoreCalculationFailed = fmt.Errorf("helixspecifier speckit: heuristic QualityScore calculation produced an out-of-range value — refusing to emit invalid score (round-65 §11.4 anti-bluff)")
)

// llmDebateConfig holds the tunable knobs for LLMBackedDebateFunc.
// Defaults are chosen so the function works out-of-the-box with
// any LLMResponder; callers tune via the functional options below.
type llmDebateConfig struct {
	promptTemplate string
	maxLength      int
	minLength      int
}

// LLMDebateOption is a functional option for LLMBackedDebateFunc.
type LLMDebateOption func(*llmDebateConfig)

// WithPromptTemplate replaces the default debate prompt template.
// The template MUST contain the placeholders {topic} (the
// phase-derived debate topic), {rounds} (the requested round
// count), and {phase} (the SpecKit phase name from metadata).
func WithPromptTemplate(template string) LLMDebateOption {
	return func(c *llmDebateConfig) {
		if strings.TrimSpace(template) != "" {
			c.promptTemplate = template
		}
	}
}

// WithMaxResponseLength sets the maximum response length (in
// runes) above which the heuristic scorer caps response-length
// contribution. Defaults to 8192.
func WithMaxResponseLength(n int) LLMDebateOption {
	return func(c *llmDebateConfig) {
		if n > 0 {
			c.maxLength = n
		}
	}
}

// WithMinResponseLength sets the minimum response length (in
// runes) below which the response is treated as an empty-debate
// failure (ErrLLMDebateEmptyResponse). Defaults to 16.
func WithMinResponseLength(n int) LLMDebateOption {
	return func(c *llmDebateConfig) {
		if n >= 0 {
			c.minLength = n
		}
	}
}

// defaultPromptTemplate produces a structured FOR/AGAINST/SYNTHESIS
// debate prompt. Placeholders are substituted at invocation time.
const defaultPromptTemplate = `You are conducting a structured debate for the SpecKit phase "{phase}".

Topic: {topic}

Conduct {rounds} round(s) of debate. For each round, produce:
  FOR: <argument supporting the proposal>
  AGAINST: <counter-argument or risk>
  SYNTHESIS: <reconciliation that advances the phase>

After the final round, emit:
  CONCLUSION: <one paragraph capturing the phase output>

Respond now with the full debate transcript.`

// LLMBackedDebateFunc returns a types.DebateFunc that orchestrates
// a real LLM-backed debate by invoking the supplied responder.
//
// The returned DebateFunc is safe for concurrent use provided the
// underlying responder is. Passing a nil responder at construction
// time returns a DebateFunc that immediately yields
// ErrLLMDebateResponderNotConfigured on invocation (defensive — the
// constructor does NOT panic so consumer test harnesses can
// exercise the misconfiguration path).
func LLMBackedDebateFunc(responder LLMResponder, opts ...LLMDebateOption) types.DebateFunc {
	cfg := &llmDebateConfig{
		promptTemplate: defaultPromptTemplate,
		maxLength:      8192,
		minLength:      16,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	return func(
		ctx context.Context,
		topic string,
		rounds int,
		metadata map[string]interface{},
	) (string, float64, string, error) {
		if responder == nil {
			return "", 0, "", ErrLLMDebateResponderNotConfigured
		}

		// Honour ctx cancellation before we issue any network I/O.
		if err := ctx.Err(); err != nil {
			return "", 0, "", fmt.Errorf("helixspecifier speckit: context cancelled before LLM debate invocation: %w", err)
		}

		phase, _ := metadata["phase"].(string)
		if phase == "" {
			phase = "unknown"
		}
		if rounds < 1 {
			rounds = 1
		}

		prompt := renderPrompt(cfg.promptTemplate, topic, rounds, phase)

		raw, err := responder.Generate(ctx, prompt)
		if err != nil {
			return "", 0, "", fmt.Errorf("helixspecifier speckit: LLMResponder.Generate failed for phase %q: %w", phase, err)
		}

		trimmed := strings.TrimSpace(raw)
		if len([]rune(trimmed)) < cfg.minLength {
			return "", 0, "", fmt.Errorf("helixspecifier speckit: phase %q: %w (got %d runes, minimum %d)", phase, ErrLLMDebateEmptyResponse, len([]rune(trimmed)), cfg.minLength)
		}

		score, err := scoreDebate(trimmed, rounds, cfg.maxLength)
		if err != nil {
			return "", 0, "", err
		}

		// Debate ID: deterministic content hash so repeated debates
		// over identical prompt+response are deduplicable. NOT a
		// random UUID — keeps test assertions reproducible.
		debateID := contentDebateID(phase, prompt, trimmed)

		return trimmed, score, debateID, nil
	}
}

// renderPrompt substitutes the three documented placeholders. No
// templating engine is used — kept dependency-light per CONST-051(B).
func renderPrompt(tpl, topic string, rounds int, phase string) string {
	out := strings.ReplaceAll(tpl, "{topic}", topic)
	out = strings.ReplaceAll(out, "{rounds}", fmt.Sprintf("%d", rounds))
	out = strings.ReplaceAll(out, "{phase}", phase)
	return out
}

// scoreDebate computes a heuristic QualityScore in [0.0, 1.0] from
// the trimmed model response. The heuristic considers:
//
//   - structural-marker coverage (FOR / AGAINST / SYNTHESIS /
//     CONCLUSION) — 0.0..0.6 contribution;
//   - response length relative to maxLength — 0.0..0.3 contribution;
//   - per-round marker density (FOR markers found vs rounds
//     requested) — 0.0..0.1 contribution.
//
// The heuristic is INTENTIONALLY simple and total-function — it does
// not pretend to evaluate semantic quality. It IS a real, deterministic
// derivation from captured model output (CONST-035), not a fabricated
// 0.75-style hardcoded number (round-28 §11.4 bluff).
func scoreDebate(response string, rounds, maxLength int) (float64, error) {
	if response == "" {
		return 0, ErrLLMDebateEmptyResponse
	}

	upper := strings.ToUpper(response)

	var structuralScore float64
	markers := []string{"FOR:", "AGAINST:", "SYNTHESIS:", "CONCLUSION:"}
	for _, m := range markers {
		if strings.Contains(upper, m) {
			structuralScore += 0.15
		}
	}
	// structuralScore now in [0.0, 0.6]

	// Length contribution — linear up to maxLength, capped.
	runeLen := len([]rune(response))
	lengthScore := float64(runeLen) / float64(maxLength) * 0.3
	if lengthScore > 0.3 {
		lengthScore = 0.3
	}

	// Per-round FOR marker density — caller asked for `rounds`
	// rounds of debate; reward responses that produced them.
	forCount := strings.Count(upper, "FOR:")
	var roundsScore float64
	if rounds > 0 {
		ratio := float64(forCount) / float64(rounds)
		if ratio > 1.0 {
			ratio = 1.0
		}
		roundsScore = ratio * 0.1
	}

	total := structuralScore + lengthScore + roundsScore
	if total < 0 || total > 1 {
		return 0, ErrLLMDebateScoreCalculationFailed
	}
	return total, nil
}

// contentDebateID returns a deterministic 16-hex-char ID derived
// from phase + prompt + response. Stable across runs (no clock /
// no randomness) so tests can assert on it without flakes.
func contentDebateID(phase, prompt, response string) string {
	h := sha256.New()
	h.Write([]byte(phase))
	h.Write([]byte{0})
	h.Write([]byte(prompt))
	h.Write([]byte{0})
	h.Write([]byte(response))
	sum := h.Sum(nil)
	return "speckit-llm-" + hex.EncodeToString(sum[:8])
}
