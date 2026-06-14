package i18n

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBundleTranslator_LoadsEmbeddedBundle proves the embedded
// active English bundle decodes — a packaging/build defect would
// surface here, never at runtime in a shipped binary.
func TestBundleTranslator_LoadsEmbeddedBundle(t *testing.T) {
	bt, err := NewBundleTranslator()
	require.NoError(t, err)
	require.NotNil(t, bt)
	assert.NotEmpty(t, bt.messages, "embedded bundle decoded to empty map")
}

// TestDefaultTranslator_NotNoop confirms the process default is the
// bundle-backed translator (not NoopTranslator) when the embedded
// bundle is intact — the HXC-081 fix's load-bearing precondition.
func TestDefaultTranslator_NotNoop(t *testing.T) {
	require.NoError(t, DefaultTranslatorErr())
	tr := DefaultTranslator()
	require.NotNil(t, tr)
	_, isNoop := tr.(NoopTranslator)
	assert.False(t, isNoop, "default translator must be bundle-backed, got Noop")
}

// TestBundleTranslator_PresentKeyResolvesAndInterpolates is the
// HXC-081 positive guard: the /specify key resolves to its bundle
// prose and the supplied args are interpolated into the format
// string's %s verbs — NOT echoed as a raw key.
func TestBundleTranslator_PresentKeyResolvesAndInterpolates(t *testing.T) {
	bt, err := NewBundleTranslator()
	require.NoError(t, err)

	got := bt.T(
		"helixspecifier_speckit_topic_specify",
		"Build a URL shortener service", "prev-output",
	)
	assert.NotContains(
		t, got, "helixspecifier_speckit_topic_specify",
		"raw key leaked instead of resolving",
	)
	assert.NotContains(t, got, "%!(EXTRA", "format verb/arg mismatch")
	assert.NotContains(t, got, "%!", "any Go format-error marker")
	assert.Contains(t, got, "Build a URL shortener service")
	assert.Contains(t, got, "prev-output")
	assert.Contains(t, got, "specification")
}

// TestBundleTranslator_MissingKeyReturnsBareKeyNoSprintf is the
// HXC-081 ROOT-CAUSE guard: the original bug was T doing
// fmt.Sprintf(rawKey, args...) on a key with no `%` verbs, yielding
// `rawKey%!(EXTRA string=...)`. A MISSING key MUST return the bare
// key verbatim and MUST NOT feed args to Sprintf — so no
// `%!(EXTRA)` marker can ever be produced for a missing translation.
func TestBundleTranslator_MissingKeyReturnsBareKeyNoSprintf(t *testing.T) {
	bt, err := NewBundleTranslator()
	require.NoError(t, err)

	const missing = "helixspecifier_definitely_absent_key_xyz"
	got := bt.T(missing, "arg-a", "arg-b")
	assert.Equal(
		t, missing, got,
		"missing key must return the bare key, never Sprintf(key, args)",
	)
	assert.False(
		t, strings.Contains(got, "%!(EXTRA"),
		"missing-key path must not produce a Go format-error marker",
	)
}

// TestBundleTranslator_NoArgsReturnsFormatVerbatim confirms the
// no-args path returns the bundle format string unchanged (no
// spurious Sprintf with zero args mangling literal `%` if present).
func TestBundleTranslator_NoArgsReturnsFormatVerbatim(t *testing.T) {
	bt, err := NewBundleTranslator()
	require.NoError(t, err)

	got := bt.T("helixspecifier_engine_err_empty_request")
	assert.Equal(t, "empty request", got)
}
