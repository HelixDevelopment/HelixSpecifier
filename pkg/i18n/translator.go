// Package i18n provides locale-aware translation lookup for
// user-facing HelixSpecifier strings per CONST-046.
//
// CONST-046 forbids hardcoded user-facing literals in production
// code: every string visible to an operator MUST be either
// LLM-generated, composed from metadata, or loaded through this
// resource layer. The translator type below is decoupled from
// HelixSpecifier domain types so it remains reusable by any other
// project that incorporates this submodule per CONST-051(B).
//
// Verbatim 2026-05-19 operator mandate (preserved per
// CONST-049 §11.4.17):
//
//	"all existing tests and Challenges do work in anti-bluff
//	manner - they MUST confirm that all tested codebase really
//	works as expected! We had been in position that all tests
//	do execute with success and all Challenges as well, but in
//	reality the most of the features does not work and can't
//	be used! This MUST NOT be the case and execution of tests
//	and Challenges MUST guarantee the quality, the completition
//	and full usability by end users of the product!"
package i18n

import (
	"fmt"
	"strings"
)

// Translator resolves locale-aware messages for HelixSpecifier
// user-facing surfaces. Implementations are free to load from
// YAML/JSON bundles, environment-driven resolvers, or test-only
// noop adapters. The interface intentionally stays narrow so any
// consumer (HelixCode meta-repo or a downstream project) can plug
// its own backend without modifying this submodule, satisfying
// CONST-051(B) project-not-aware decoupling.
type Translator interface {
	// T returns the translated message for the given key and
	// arguments. Implementations MUST NOT panic on missing
	// keys; they MUST return a stable fallback so production
	// flows do not crash when a bundle is incomplete.
	T(key string, args ...any) string
}

// NoopTranslator returns the key (or formatted key with args) as
// its translation. It is the safe default for environments that
// have not provisioned a real bundle (early bootstrap, unit tests,
// CLI bring-up). Returning the key — not an empty string — is a
// deliberate anti-bluff choice: a missing translation is visibly
// "helixspecifier_intent_quick_reasoning" on screen, not silent
// emptiness that could be mistaken for "successful render".
type NoopTranslator struct{}

// T implements Translator. When args are supplied, the key is
// treated as a format string; otherwise it is returned verbatim.
// This mirrors the contract a real bundle backend would offer.
func (NoopTranslator) T(key string, args ...any) string {
	if len(args) == 0 {
		return key
	}
	return fmt.Sprintf(key, args...)
}

// NewNoopTranslator returns a NoopTranslator. Production code
// SHOULD wire a real bundle-backed translator instead; this
// constructor exists so call sites can declare an explicit
// fallback at module bootstrap without exposing the zero-value
// quirks of an empty struct.
func NewNoopTranslator() Translator {
	return NoopTranslator{}
}

// ResolveOrFallback routes a user-facing string through the
// translator and, when the translator is the NoopTranslator (the
// key-verbatim path), substitutes a bundled English fallback so
// the call site still emits a coherent, correctly-formatted string
// instead of a raw i18n key.
//
// This is the seam every CONST-046 migration of *display-formatted*
// content (Markdown report bodies, multi-part labels, validation
// errors) MUST pass through: a raw key leaking into rendered output
// is itself a usability defect (Article XI §11.9 — "users can use
// the feature" is the bar, not "tests pass").
//
// Detection contract: NoopTranslator.T returns the msgID verbatim
// when no args are supplied, or fmt.Sprintf(msgID, args...) when
// args are supplied — in the latter case the result still begins
// with the msgID because msgID is not a format string. So a result
// equal to OR prefixed by msgID signals the noop path; the call
// site's bundled English fallback is then emitted (formatted with
// the same args). A real bundle-backed Translator resolves the key
// to locale text and its result is returned unchanged.
func ResolveOrFallback(
	tr Translator,
	msgID, fallback string,
	args ...any,
) string {
	if tr == nil {
		tr = NoopTranslator{}
	}
	got := tr.T(msgID, args...)
	if got == msgID || strings.HasPrefix(got, msgID) {
		if len(args) == 0 {
			return fallback
		}
		return fmt.Sprintf(fallback, args...)
	}
	return got
}
