// runner unit tests — round-273.
//
// Exercises the parseFixture / loadFixtures helpers + the full run()
// path against a controlled fixtures directory. Race-detector clean.
//
// Unit tests are the ONLY layer where in-test mocks are permitted
// (CONST-050(A)). Here the only "fake" is an alternate
// FIXTURES_DIR pointing at t.TempDir() — the runner itself still
// exercises real engine / classifier / translator types.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"digital.vasic.helixspecifier/pkg/types"
)

func writeFixture(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(
		filepath.Join(dir, name),
		[]byte(body), 0o644,
	); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestParseFixture_MinimalKeys(t *testing.T) {
	body := `# comment line
locale: en
prompt: "Refactor module"
expect_effort: large
expect_signal_substr: "refactor:"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`
	f := parseFixture(body)
	if f.locale != "en" {
		t.Errorf("locale: want en got %q", f.locale)
	}
	if f.prompt != "Refactor module" {
		t.Errorf("prompt: got %q", f.prompt)
	}
	if f.expectEffort != types.EffortLarge {
		t.Errorf("effort: want large got %q", f.expectEffort)
	}
	if f.expectSignalSubstr != "refactor:" {
		t.Errorf("signal: %q", f.expectSignalSubstr)
	}
}

func TestLoadFixtures_FullRoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "en.yaml",
		`locale: en
prompt: "Add OAuth2 authentication to API gateway service"
expect_effort: large
expect_signal_substr: "large:"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`)
	writeFixture(t, dir, "de.yaml",
		`locale: de
prompt: "Fix typo in error message"
expect_effort: quick
expect_signal_substr: "quick:fix typo"
expect_reasoning_key: helixspecifier_intent_quick_reasoning
expect_reasoning_substr: helixspecifier_intent_quick_reasoning
`)
	writeFixture(t, dir, "sr.yaml",
		`locale: sr
prompt: "Implement user authentication database backend"
expect_effort: large
expect_signal_substr: "large:"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`)
	writeFixture(t, dir, "ja.yaml",
		`locale: ja
prompt: "Rebuild all of the entire system from scratch"
expect_effort: epic
expect_signal_substr: "epic:"
expect_reasoning_key: helixspecifier_intent_epic_reasoning
expect_reasoning_substr: helixspecifier_intent_epic_reasoning
`)
	writeFixture(t, dir, "es.yaml",
		`locale: es
prompt: "Refactor module structure for readability"
expect_effort: large
expect_signal_substr: "refactor:refactor"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`)

	fixtures, err := loadFixtures(dir)
	if err != nil {
		t.Fatalf("loadFixtures: %v", err)
	}
	if len(fixtures) != 5 {
		t.Fatalf("expected 5 fixtures, got %d", len(fixtures))
	}

	t.Setenv("HELIXSPECIFIER_FIXTURES_DIR", dir)
	t.Setenv("HELIXSPECIFIER_MUTATE_RUNNER", "")

	var out bytes.Buffer
	code := run(&out)
	if code != 0 {
		t.Fatalf("run() exit=%d, want 0\n%s",
			code, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "FAIL=0") {
		t.Errorf("expected FAIL=0 summary, got:\n%s", got)
	}
	for _, loc := range []string{"en", "de", "sr", "ja", "es"} {
		if !strings.Contains(got, "intent.Classify."+loc) {
			t.Errorf("missing per-locale invariant %q", loc)
		}
	}
}

func TestRun_MutationDetected(t *testing.T) {
	// Reuse the full-round-trip fixtures pattern but with the
	// mutation env var asserted by the runner. The runner MUST
	// then exit non-zero — proving paired-mutation actually
	// detects the planted invariant flip (CONST-050(A) §1.1).
	dir := t.TempDir()
	writeFixture(t, dir, "en.yaml",
		`locale: en
prompt: "Add OAuth2 authentication to API gateway service"
expect_effort: large
expect_signal_substr: "large:"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`)
	writeFixture(t, dir, "de.yaml",
		`locale: de
prompt: "Fix typo in error message"
expect_effort: quick
expect_signal_substr: "quick:fix typo"
expect_reasoning_key: helixspecifier_intent_quick_reasoning
expect_reasoning_substr: helixspecifier_intent_quick_reasoning
`)
	writeFixture(t, dir, "sr.yaml",
		`locale: sr
prompt: "Implement user authentication database backend"
expect_effort: large
expect_signal_substr: "large:"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`)
	writeFixture(t, dir, "ja.yaml",
		`locale: ja
prompt: "Rebuild all of the entire system from scratch"
expect_effort: epic
expect_signal_substr: "epic:"
expect_reasoning_key: helixspecifier_intent_epic_reasoning
expect_reasoning_substr: helixspecifier_intent_epic_reasoning
`)
	writeFixture(t, dir, "es.yaml",
		`locale: es
prompt: "Refactor module structure for readability"
expect_effort: large
expect_signal_substr: "refactor:refactor"
expect_reasoning_key: helixspecifier_intent_large_reasoning
expect_reasoning_substr: helixspecifier_intent_large_reasoning
`)

	t.Setenv("HELIXSPECIFIER_FIXTURES_DIR", dir)
	t.Setenv("HELIXSPECIFIER_MUTATE_RUNNER", "1")

	var out bytes.Buffer
	code := run(&out)
	if code == 0 {
		t.Fatalf("mutated runner exited 0 — mutation undetected:\n%s",
			out.String())
	}
}
