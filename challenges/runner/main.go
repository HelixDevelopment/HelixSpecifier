// Command runner is the HelixSpecifier round-273 Challenge runner.
//
// It exercises the real FusionEngine, the real intent.Classifier, and
// the real i18n.Translator surface across five locale fixtures. Every
// PASS line is backed by a runtime invariant — never a metadata-only
// check (CONST-035 / Article XI §11.9).
//
// Anti-bluff invariants enforced:
//
//  1. Engine.Name() and Engine.Version() return the documented
//     constants (not empty / not placeholder).
//  2. Engine.Health() reports the speckit-required error when no
//     pillar is registered, then PASSes after registration.
//  3. Engine.ClassifyEffort errors on empty input (defensive
//     contract preserved across rounds).
//  4. intent.Classifier returns the expected effort level + at
//     least one signal substring matching the fixture's
//     expect_signal_substr (proves the keyword lexicon really
//     fires; not a tautology).
//  5. i18n.NoopTranslator returns the key verbatim when no args
//     are passed (the anti-bluff contract documented in
//     translator.go — missing translations surface as the key,
//     not as silent empty strings).
//
// Mutation hook: when env HELIXSPECIFIER_MUTATE_RUNNER=1 is set,
// the runner inverts invariant (3) (treats a successful empty
// classification as PASS instead of FAIL). Paired Challenge wraps
// this to assert the runner exits 99 under mutation, guaranteeing
// the runner actually checks what it claims (CONST-050(A) paired
// mutation, §1.1).
//
// Verbatim 2026-05-19 operator mandate (preserved per
// CONST-049 §11.4.17):
//
//	"all existing tests and Challenges do work in anti-bluff
//	manner - they MUST confirm that all tested codebase really
//	works as expected! We had been in position that all tests
//	do execute with success and all Challenges as well, but
//	in reality the most of the features does not work and
//	can't be used! This MUST NOT be the case and execution
//	of tests and Challenges MUST guarantee the quality, the
//	completition and full usability by end users of the
//	product!"
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.helixspecifier/pkg/config"
	"digital.vasic.helixspecifier/pkg/engine"
	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/intent"
	"digital.vasic.helixspecifier/pkg/speckit"
	"digital.vasic.helixspecifier/pkg/types"

	"github.com/sirupsen/logrus"
)

// fixture is a 5-field projection of challenges/fixtures/<locale>.yaml.
// We parse minimal YAML in-process so the runner stays dependency-free
// beyond what the module already depends on (CONST-051(B): no new
// transitive deps creeping into a reusable submodule).
type fixture struct {
	locale                string
	prompt                string
	expectEffort          types.EffortLevel
	expectSignalSubstr    string
	expectReasoningKey    string
	expectReasoningSubstr string
}

func main() {
	if code := run(os.Stdout); code != 0 {
		os.Exit(code)
	}
}

func run(out io.Writer) int {
	fmt.Fprintln(out, "=== HelixSpecifier Challenge Runner (round-273) ===")

	fixDir := os.Getenv("HELIXSPECIFIER_FIXTURES_DIR")
	if fixDir == "" {
		// Default: ../fixtures relative to this binary's source.
		// When invoked via `go run ./challenges/runner` from the
		// module root this resolves to challenges/fixtures/.
		fixDir = filepath.Join("challenges", "fixtures")
	}

	fixtures, err := loadFixtures(fixDir)
	if err != nil {
		fmt.Fprintf(out, "FAIL: load fixtures from %s: %v\n",
			fixDir, err)
		return 1
	}
	if len(fixtures) < 5 {
		fmt.Fprintf(out, "FAIL: expected >=5 fixtures, got %d\n",
			len(fixtures))
		return 1
	}
	fmt.Fprintf(out, "[setup] loaded %d locale fixtures from %s\n",
		len(fixtures), fixDir)

	mutate := os.Getenv("HELIXSPECIFIER_MUTATE_RUNNER") == "1"
	if mutate {
		fmt.Fprintln(out, "[setup] MUTATION MODE: runner will treat"+
			" successful empty classification as PASS")
	}

	logger := logrus.New()
	logger.SetOutput(io.Discard)
	cfg := config.FromEnv()
	tr := i18n.NewNoopTranslator()

	pass, fail := 0, 0
	step := func(name string, ok bool, detail string) {
		if ok {
			pass++
			fmt.Fprintf(out, "  PASS  %-44s  %s\n", name, detail)
			return
		}
		fail++
		fmt.Fprintf(out, "  FAIL  %-44s  %s\n", name, detail)
	}

	// Invariant 1: Engine.Name() + Version() — documented constants.
	eng := engine.NewWithTranslator(cfg, logger, tr)
	step("engine.Name",
		eng.Name() == "HelixSpecifier",
		fmt.Sprintf("got=%q", eng.Name()))
	step("engine.Version",
		eng.Version() == "1.0.0",
		fmt.Sprintf("got=%q", eng.Version()))

	// Invariant 2: Health errors w/o speckit, passes w/ pillar.
	ctx, cancel := context.WithTimeout(
		context.Background(), 10*time.Second)
	defer cancel()
	if err := eng.Health(ctx); err == nil {
		step("engine.Health.no_pillar_returns_error", false,
			"expected error, got nil")
	} else {
		step("engine.Health.no_pillar_returns_error", true,
			fmt.Sprintf("got=%v", err))
	}
	eng.RegisterSpecKit(speckit.NewPillar(cfg, logger))
	step("engine.Health.with_pillar_ok",
		eng.Health(ctx) == nil,
		"speckit registered")

	// Invariant 3: ClassifyEffort errors on empty input.
	_, errEmpty := eng.ClassifyEffort(ctx, "")
	if mutate {
		// Mutation flips polarity: expect PASS when error is nil.
		step("engine.ClassifyEffort.empty_errors[MUTATED]",
			errEmpty == nil,
			"mutation-inverted check")
	} else {
		step("engine.ClassifyEffort.empty_errors",
			errEmpty != nil,
			fmt.Sprintf("got=%v", errEmpty))
	}

	// Invariant 4: per-fixture classifier exercise.
	classifier := intent.NewClassifierWithTranslator(tr)
	for _, f := range fixtures {
		cls := classifier.Classify(f.prompt)
		levelOK := cls.Level == f.expectEffort
		sigOK := false
		for _, s := range cls.Signals {
			if strings.Contains(s, f.expectSignalSubstr) {
				sigOK = true
				break
			}
		}
		detail := fmt.Sprintf(
			"locale=%s want=%s got=%s signals=%v",
			f.locale, f.expectEffort, cls.Level, cls.Signals)
		step("intent.Classify."+f.locale+".level",
			levelOK, detail)
		step("intent.Classify."+f.locale+".signal",
			sigOK,
			fmt.Sprintf("want_substr=%q signals=%v",
				f.expectSignalSubstr, cls.Signals))
	}

	// Invariant 5: NoopTranslator returns key verbatim — the
	// documented anti-bluff contract (translator.go).
	for _, f := range fixtures {
		got := tr.T(f.expectReasoningKey)
		ok := got == f.expectReasoningKey &&
			strings.Contains(got, f.expectReasoningSubstr)
		step("i18n.Noop.key_roundtrip."+f.locale,
			ok,
			fmt.Sprintf("key=%s got=%q",
				f.expectReasoningKey, got))
	}

	fmt.Fprintf(out, "\n=== Summary: PASS=%d FAIL=%d ===\n",
		pass, fail)
	if fail > 0 {
		return 1
	}
	return 0
}

// loadFixtures parses every *.yaml in dir using a tiny line-based
// parser. We only support the 6 keys our fixtures use; anything
// else is ignored. Keeping the parser in-runner avoids pulling
// yaml.v3 into the runtime path of a submodule that other projects
// reuse (CONST-051(B)).
func loadFixtures(dir string) ([]fixture, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []fixture
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		f := parseFixture(string(data))
		if f.locale == "" {
			return nil, fmt.Errorf(
				"%s: missing locale key", e.Name())
		}
		out = append(out, f)
	}
	return out, nil
}

func parseFixture(text string) fixture {
	f := fixture{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		colon := strings.Index(line, ":")
		if colon < 0 {
			continue
		}
		k := strings.TrimSpace(line[:colon])
		v := strings.TrimSpace(line[colon+1:])
		v = strings.Trim(v, "\"'")
		switch k {
		case "locale":
			f.locale = v
		case "prompt":
			f.prompt = v
		case "expect_effort":
			f.expectEffort = types.EffortLevel(v)
		case "expect_signal_substr":
			f.expectSignalSubstr = v
		case "expect_reasoning_key":
			f.expectReasoningKey = v
		case "expect_reasoning_substr":
			f.expectReasoningSubstr = v
		}
	}
	return f
}
