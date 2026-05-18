package speckit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"digital.vasic.helixspecifier/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubResponder is a unit-test-only LLMResponder per CONST-050(A)
// (mocks/stubs PERMITTED in *_test.go files invoked without the
// integration build tag). Production code MUST NOT import this.
type stubResponder struct {
	response    string
	err         error
	calledCtx   context.Context
	calledPrompt string
	delay       time.Duration
}

func (s *stubResponder) Generate(ctx context.Context, prompt string) (string, error) {
	s.calledCtx = ctx
	s.calledPrompt = prompt
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	if s.err != nil {
		return "", s.err
	}
	return s.response, nil
}

// cannedDebate is a realistic debate transcript that triggers all
// four structural markers — used by the happy-path tests so the
// heuristic scorer returns a non-trivial score.
const cannedDebate = `Round 1:
FOR: The authentication module should use JWT tokens because they are stateless and scale horizontally.
AGAINST: JWT tokens cannot be revoked before expiration without server-side state, which weakens incident response.
SYNTHESIS: Adopt JWT with a short TTL (15min) plus a refresh-token table; revocation by deleting the refresh token.

Round 2:
FOR: bcrypt provides industry-standard password hashing with adjustable work factor.
AGAINST: argon2id is the modern OWASP recommendation and resists GPU/ASIC attacks better.
SYNTHESIS: Use argon2id for new deployments, bcrypt fallback for compatibility, configurable per environment.

CONCLUSION: Implement authentication as a JWT (RS256) + refresh-token pair with argon2id password hashing,
revocation via refresh-token deletion, and per-environment configuration of legacy bcrypt fallback.`

// --- Round-65 §11.4 anti-bluff tests for LLMBackedDebateFunc ---

// TestLLMBackedDebateFunc_NilResponder_ReturnsSentinel asserts that
// invoking the DebateFunc with a nil responder yields the dedicated
// round-65 sentinel (distinguishable from round-28's
// ErrDebateFuncNotConfigured via errors.Is).
func TestLLMBackedDebateFunc_NilResponder_ReturnsSentinel(t *testing.T) {
	fn := LLMBackedDebateFunc(nil)
	out, score, debateID, err := fn(context.Background(), "topic", 3, map[string]interface{}{"phase": "specify"})

	require.Error(t, err)
	require.ErrorIs(t, err, ErrLLMDebateResponderNotConfigured)
	assert.Empty(t, out, "output MUST be empty when responder is nil — no fabrication")
	assert.Zero(t, score, "score MUST be zero when responder is nil — no fabricated 0.75")
	assert.Empty(t, debateID, "debateID MUST be empty when responder is nil")

	// Paired-mutation: round-28 and round-65 sentinels MUST be
	// distinguishable so callers can branch precisely.
	assert.False(t, errors.Is(err, ErrDebateFuncNotConfigured),
		"round-65 sentinel MUST NOT collapse into round-28 sentinel")
}

// TestLLMBackedDebateFunc_ResponderError_PropagatesAsError asserts
// that responder errors propagate verbatim (wrapped with %w).
func TestLLMBackedDebateFunc_ResponderError_PropagatesAsError(t *testing.T) {
	upstreamErr := errors.New("ollama: connection refused")
	resp := &stubResponder{err: upstreamErr}
	fn := LLMBackedDebateFunc(resp)

	out, score, debateID, err := fn(context.Background(), "topic", 2, map[string]interface{}{"phase": "plan"})

	require.Error(t, err)
	require.ErrorIs(t, err, upstreamErr, "responder error MUST be wrapped with %%w so callers can unwrap")
	assert.Empty(t, out)
	assert.Zero(t, score)
	assert.Empty(t, debateID)
}

// TestLLMBackedDebateFunc_EmptyResponse_ReturnsSentinel asserts that
// an empty model response yields ErrLLMDebateEmptyResponse rather
// than a fabricated transcript.
func TestLLMBackedDebateFunc_EmptyResponse_ReturnsSentinel(t *testing.T) {
	resp := &stubResponder{response: "   \n\t  "}
	fn := LLMBackedDebateFunc(resp)

	out, score, debateID, err := fn(context.Background(), "topic", 1, map[string]interface{}{"phase": "specify"})

	require.Error(t, err)
	require.ErrorIs(t, err, ErrLLMDebateEmptyResponse)
	assert.Empty(t, out)
	assert.Zero(t, score)
	assert.Empty(t, debateID)
}

// TestLLMBackedDebateFunc_HappyPath_PopulatesPhaseOutput asserts the
// success path — real responder output flows through to PhaseResult
// with a derived (non-fabricated) QualityScore.
func TestLLMBackedDebateFunc_HappyPath_PopulatesPhaseOutput(t *testing.T) {
	resp := &stubResponder{response: cannedDebate}
	fn := LLMBackedDebateFunc(resp)

	out, score, debateID, err := fn(context.Background(), "Design authentication module", 2, map[string]interface{}{"phase": "specify"})

	require.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(cannedDebate), out, "output MUST be the trimmed responder response, not fabricated")
	assert.Greater(t, score, 0.0, "QualityScore MUST be derived from response content")
	assert.LessOrEqual(t, score, 1.0, "QualityScore MUST be in [0, 1]")
	// canned debate has all 4 structural markers + decent length → score should be > 0.6
	assert.Greater(t, score, 0.6, "canned debate with all 4 markers MUST score above structural floor")
	assert.NotEmpty(t, debateID, "debateID MUST be populated (deterministic content hash)")
	assert.True(t, strings.HasPrefix(debateID, "speckit-llm-"), "debateID MUST carry the speckit-llm- namespace prefix")

	// Verify prompt actually contained substituted placeholders.
	assert.Contains(t, resp.calledPrompt, "specify", "prompt MUST substitute the {phase} placeholder")
	assert.Contains(t, resp.calledPrompt, "Design authentication module", "prompt MUST substitute the {topic} placeholder")
	assert.Contains(t, resp.calledPrompt, "2", "prompt MUST substitute the {rounds} placeholder")
}

// TestLLMBackedDebateFunc_DeterministicDebateID asserts the content
// hash is reproducible across invocations with identical input —
// no randomness, no clock dependency.
func TestLLMBackedDebateFunc_DeterministicDebateID(t *testing.T) {
	resp1 := &stubResponder{response: cannedDebate}
	resp2 := &stubResponder{response: cannedDebate}
	fn1 := LLMBackedDebateFunc(resp1)
	fn2 := LLMBackedDebateFunc(resp2)

	meta := map[string]interface{}{"phase": "specify"}
	_, _, id1, err1 := fn1(context.Background(), "T", 1, meta)
	_, _, id2, err2 := fn2(context.Background(), "T", 1, meta)
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, id1, id2, "debateID MUST be deterministic across identical invocations")
}

// TestLLMBackedDebateFunc_HonoursContextCancel asserts the
// implementation surfaces context cancellation rather than blocking
// or fabricating output.
func TestLLMBackedDebateFunc_HonoursContextCancel(t *testing.T) {
	resp := &stubResponder{response: cannedDebate, delay: 200 * time.Millisecond}
	fn := LLMBackedDebateFunc(resp)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	out, score, debateID, err := fn(ctx, "topic", 1, map[string]interface{}{"phase": "plan"})
	require.Error(t, err)
	// Either context.DeadlineExceeded surfaces through, or the
	// pre-check in the closure catches it. Both are valid.
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"context error MUST propagate (got %v)", err)
	assert.Empty(t, out)
	assert.Zero(t, score)
	assert.Empty(t, debateID)
}

// TestLLMBackedDebateFunc_PreCheckContextCancellation asserts the
// pre-flight ctx.Err() guard fires before any responder call when
// the context is already cancelled at entry.
func TestLLMBackedDebateFunc_PreCheckContextCancellation(t *testing.T) {
	resp := &stubResponder{response: cannedDebate}
	fn := LLMBackedDebateFunc(resp)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, _, err := fn(ctx, "topic", 1, map[string]interface{}{"phase": "plan"})
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, resp.calledCtx, "responder MUST NOT be called when ctx is pre-cancelled")
}

// TestLLMBackedDebateFunc_OptionsApplied asserts the functional
// options actually take effect.
func TestLLMBackedDebateFunc_OptionsApplied(t *testing.T) {
	resp := &stubResponder{response: cannedDebate}
	customTpl := "CUSTOM TPL phase={phase} topic={topic} rounds={rounds}"
	fn := LLMBackedDebateFunc(resp,
		WithPromptTemplate(customTpl),
		WithMaxResponseLength(1000),
		WithMinResponseLength(8),
	)

	_, _, _, err := fn(context.Background(), "T", 3, map[string]interface{}{"phase": "tasks"})
	require.NoError(t, err)
	assert.Equal(t, "CUSTOM TPL phase=tasks topic=T rounds=3", resp.calledPrompt,
		"custom template MUST be used and all placeholders substituted")
}

// TestLLMBackedDebateFunc_OptionsIgnoresNilAndBadInput asserts the
// option helpers gracefully ignore degenerate input (empty template,
// zero or negative lengths) per CONST-035 — silent degradation is
// preferable to producing a broken closure.
func TestLLMBackedDebateFunc_OptionsIgnoresNilAndBadInput(t *testing.T) {
	resp := &stubResponder{response: cannedDebate}
	fn := LLMBackedDebateFunc(resp,
		WithPromptTemplate(""), // ignored — default kept
		WithMaxResponseLength(-5),
		WithMinResponseLength(-1),
	)
	_, _, _, err := fn(context.Background(), "T", 1, map[string]interface{}{"phase": "specify"})
	require.NoError(t, err)
	assert.Contains(t, resp.calledPrompt, "SpecKit phase", "default template MUST be retained when WithPromptTemplate('') is passed")
}

// TestLLMBackedDebateFunc_MissingPhaseMetadata_StillSucceeds asserts
// graceful degradation when metadata lacks a phase entry.
func TestLLMBackedDebateFunc_MissingPhaseMetadata_StillSucceeds(t *testing.T) {
	resp := &stubResponder{response: cannedDebate}
	fn := LLMBackedDebateFunc(resp)

	_, _, _, err := fn(context.Background(), "topic", 1, map[string]interface{}{})
	require.NoError(t, err)
	assert.Contains(t, resp.calledPrompt, "unknown", "phase MUST default to 'unknown' when metadata lacks it")
}

// TestLLMBackedDebateFunc_ZeroOrNegativeRounds_NormalisedToOne
// asserts the rounds parameter is clamped to >=1.
func TestLLMBackedDebateFunc_ZeroOrNegativeRounds_NormalisedToOne(t *testing.T) {
	resp := &stubResponder{response: cannedDebate}
	fn := LLMBackedDebateFunc(resp)
	_, _, _, err := fn(context.Background(), "topic", 0, map[string]interface{}{"phase": "p"})
	require.NoError(t, err)
	assert.Contains(t, resp.calledPrompt, "1 round(s)", "rounds MUST be clamped to 1")
}

// --- End-to-end via ExecutePhase ---

// TestExecutePhase_WithLLMBackedDebateFunc_EndToEnd wires
// LLMBackedDebateFunc into a Pillar and exercises ExecutePhase
// against it. This is the round-65 anti-bluff guarantee: the
// previously-bluffed nil-DebateFunc branch is replaced by a real,
// composable concrete implementation that flows through to a real
// PhaseResult.Success/Output/QualityScore.
func TestExecutePhase_WithLLMBackedDebateFunc_EndToEnd(t *testing.T) {
	p := newTestPillar()
	resp := &stubResponder{response: cannedDebate}
	p.SetDebateFunc(LLMBackedDebateFunc(resp))

	input := newTestInput()
	result, err := p.ExecutePhase(context.Background(), types.PhaseSpecify, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success, "PhaseResult.Success MUST be true on happy path")
	assert.Equal(t, types.PhaseSpecify, result.Phase)
	assert.Equal(t, strings.TrimSpace(cannedDebate), result.Output,
		"PhaseResult.Output MUST be the real model transcript, not a [<phase>] Phase output for: ... fabrication")
	assert.Greater(t, result.QualityScore, 0.0,
		"PhaseResult.QualityScore MUST be derived from response content, not the hardcoded 0.75 bluff")
	assert.NotEqual(t, 0.75, result.QualityScore,
		"PhaseResult.QualityScore MUST NOT be the round-28 fabricated 0.75 value")
	assert.NotEmpty(t, result.DebateID)
	assert.NotZero(t, result.Duration)
}

// TestExecutePhase_Round28Canary asserts the round-28 fix is still
// in force: a Pillar created without SetDebateFunc MUST still
// return ErrDebateFuncNotConfigured. Round 65 ADDS a wireable
// concrete implementation; it does NOT remove the round-28 sentinel.
func TestExecutePhase_Round28Canary(t *testing.T) {
	p := newTestPillar()
	_, err := p.ExecutePhase(context.Background(), types.PhaseConstitution, newTestInput())

	require.Error(t, err)
	require.ErrorIs(t, err, ErrDebateFuncNotConfigured,
		"round-28 sentinel MUST remain in force when no DebateFunc is wired")
	assert.False(t, errors.Is(err, ErrLLMDebateResponderNotConfigured),
		"round-28 and round-65 sentinels MUST remain distinguishable")
}

// TestScoreDebate_StructuralFloor asserts that a response with zero
// structural markers and minimal length receives a low score.
func TestScoreDebate_StructuralFloor(t *testing.T) {
	score, err := scoreDebate("just a short reply with no markers", 1, 8192)
	require.NoError(t, err)
	assert.Less(t, score, 0.1, "marker-less short reply MUST score low")
}

// TestScoreDebate_AllMarkersPresent asserts that all four markers
// contribute toward the structural ceiling.
func TestScoreDebate_AllMarkersPresent(t *testing.T) {
	resp := "FOR: a AGAINST: b SYNTHESIS: c CONCLUSION: d"
	score, err := scoreDebate(resp, 1, 8192)
	require.NoError(t, err)
	// 4 markers * 0.15 = 0.6 structural; tiny length contribution; ~0.1 rounds
	assert.GreaterOrEqual(t, score, 0.6, "all 4 markers MUST yield >= 0.6")
}

// TestScoreDebate_EmptyResponse_ReturnsSentinel asserts the scorer
// refuses to score an empty string.
func TestScoreDebate_EmptyResponse_ReturnsSentinel(t *testing.T) {
	_, err := scoreDebate("", 1, 8192)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrLLMDebateEmptyResponse)
}

// TestContentDebateID_Deterministic asserts the content-hash ID is
// stable across calls with identical input.
func TestContentDebateID_Deterministic(t *testing.T) {
	a := contentDebateID("specify", "prompt", "response")
	b := contentDebateID("specify", "prompt", "response")
	assert.Equal(t, a, b, "identical input MUST yield identical debateID")
	c := contentDebateID("specify", "prompt", "DIFFERENT")
	assert.NotEqual(t, a, c, "different response MUST yield different debateID")
	assert.True(t, strings.HasPrefix(a, "speckit-llm-"))
}

// TestRenderPrompt_PlaceholderSubstitution asserts placeholder
// substitution is correct and idempotent.
func TestRenderPrompt_PlaceholderSubstitution(t *testing.T) {
	got := renderPrompt("p={phase} t={topic} r={rounds}", "T", 5, "specify")
	assert.Equal(t, "p=specify t=T r=5", got)
}
