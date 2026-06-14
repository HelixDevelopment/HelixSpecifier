package speckit

import (
	"os"
	"strings"
	"testing"

	"digital.vasic.helixspecifier/pkg/types"
)

// HXC-081 — raw i18n key + fmt.Sprintf verb/arg mismatch in the
// default (no-injected-translator) /specify topic.
//
// Live evidence captured by the operator:
//
//	Topic: helixspecifier_speckit_topic_specify%!(EXTRA string=Build
//	a URL shortener service, string=)
//
// Two defects, one root cause: the default Pillar resolves topic
// keys through a translator that (a) does NOT resolve the key to
// prose (the i18n bundle is never loaded — same class as HXC-079),
// and (b) feeds the raw key to fmt.Sprintf with positional args
// while the key carries NO `%` verbs, producing the `%!(EXTRA ...)`
// Go format-error marker.
//
// §11.4.115 polarity switch: with RED_MODE=1 this test REPRODUCES
// the defect — it asserts the broken markers ARE present, proving
// the defect is real on a pre-fix artifact (a RED test that fails
// to reproduce on the broken artifact is a blind test). With
// RED_MODE unset/0 (the standing default) it is the GREEN
// regression-guard: it asserts the topic is resolved prose
// referencing the request, with NO raw key and NO `%!(EXTRA`.
// The captured RED reproduction (on the pre-fix artifact) is in
// the HXC-081 work-item evidence; the suite runs the GREEN guard.
//
// The default Pillar (NewPillar, no SetTranslator) is exercised
// deliberately — that is the production default a live /specify run
// hits, the exact path the operator's evidence came from.
func TestPillar_HXC081_SpecifyTopic_RawKeyAndSprintfMismatch(t *testing.T) {
	// §11.4.115 polarity switch. Default is the standing GREEN
	// regression-guard (RED_MODE=0/unset) so the suite stays green on
	// the fixed artifact; set RED_MODE=1 to opt into reproducing the
	// historical defect on a pre-fix artifact.
	redMode := os.Getenv("RED_MODE") == "1"

	p := newTestPillar() // production default: NewPillar → bundle-backed/Noop
	input := &types.PhaseInput{
		UserRequest:    "Build a URL shortener service",
		PreviousOutput: "",
		Constitution:   "Project constitution text",
	}

	topic := p.buildPhaseTopic(types.PhaseSpecify, input)
	t.Logf("RED_MODE=%v built /specify topic: %q", redMode, topic)

	const rawKey = "helixspecifier_speckit_topic_specify"
	const sprintfErr = "%!(EXTRA"

	if redMode {
		// RED: prove BOTH defects are present on the broken artifact.
		if !strings.Contains(topic, rawKey) {
			t.Fatalf(
				"RED expected raw i18n key %q in topic but it was absent; "+
					"defect not reproduced (blind test?). topic=%q",
				rawKey, topic,
			)
		}
		if !strings.Contains(topic, sprintfErr) {
			t.Fatalf(
				"RED expected Go format-error marker %q in topic but it "+
					"was absent; defect not reproduced. topic=%q",
				sprintfErr, topic,
			)
		}
		t.Logf(
			"RED reproduced HXC-081: raw key %q AND %q both present",
			rawKey, sprintfErr,
		)
		return
	}

	// GREEN (RED_MODE=0): the fix is in place.
	if strings.Contains(topic, rawKey) {
		t.Fatalf(
			"GREEN: raw i18n key %q leaked into resolved topic; "+
				"translator did not resolve the key. topic=%q",
			rawKey, topic,
		)
	}
	if strings.Contains(topic, sprintfErr) {
		t.Fatalf(
			"GREEN: Go format-error marker %q present; Sprintf "+
				"verb/arg mismatch not fixed. topic=%q",
			sprintfErr, topic,
		)
	}
	if !strings.Contains(topic, input.UserRequest) {
		t.Fatalf(
			"GREEN: resolved topic does not reference the user "+
				"request %q; interpolation broken. topic=%q",
			input.UserRequest, topic,
		)
	}
	// Resolved prose must contain real bundle words, not the key.
	if !strings.Contains(topic, "specification") {
		t.Fatalf(
			"GREEN: resolved topic is not the expected /specify "+
				"prose (missing \"specification\"). topic=%q",
			topic,
		)
	}
	t.Logf("GREEN: /specify topic resolved to prose referencing the request")
}
