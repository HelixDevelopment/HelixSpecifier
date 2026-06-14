package i18n

import (
	"embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

// bundlesFS embeds the shipped translation bundles so a fresh
// checkout resolves keys with zero external provisioning. The
// embed is part of the binary — no runtime file lookup, no
// host-path assumptions (CONST-051(B) project-not-aware decoupling).
//
//go:embed bundles/*.yaml
var bundlesFS embed.FS

// BundleTranslator resolves keys from an embedded YAML message
// bundle. It is the production-default Translator: it loads the
// `key: "format string"` map once and, on T, looks the key up and
// applies fmt.Sprintf with the supplied args so `%s`/`%v`/`%.2f`
// placeholders in the bundle value are interpolated.
//
// Anti-bluff contract (Article XI §11.9, HXC-081 root cause): when
// a key is MISSING from the bundle, T returns the key verbatim
// (never feeds the raw key to Sprintf with args) so a missing
// translation surfaces as a visible key on screen — NOT as a
// `helixspecifier_..._key%!(EXTRA string=...)` Go format-error
// artifact, and NOT as silent emptiness. When the key IS present,
// its bundle value (a real format string with matching verbs) is
// formatted with the args, producing coherent prose.
type BundleTranslator struct {
	messages map[string]string
}

var (
	defaultBundleOnce sync.Once
	defaultBundle     *BundleTranslator
	defaultBundleErr  error
)

// parseBundleFile decodes one embedded YAML bundle into a flat
// key→message map. The bundle schema is a single-level mapping of
// string keys to string format-strings (see bundles/active.en.yaml).
func parseBundleFile(name string) (map[string]string, error) {
	raw, err := bundlesFS.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("i18n: read embedded bundle %s: %w", name, err)
	}
	out := make(map[string]string)
	if err := yaml.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("i18n: parse bundle %s: %w", name, err)
	}
	return out, nil
}

// NewBundleTranslator constructs a BundleTranslator from the
// embedded active English bundle. It returns an error only when the
// embedded resource cannot be decoded (a build/packaging defect,
// caught by tests) — never at runtime in a shipped binary, since
// the bundle is compiled in.
func NewBundleTranslator() (*BundleTranslator, error) {
	msgs, err := parseBundleFile("bundles/active.en.yaml")
	if err != nil {
		return nil, err
	}
	return &BundleTranslator{messages: msgs}, nil
}

// DefaultTranslator returns the process-wide bundle-backed
// translator, loading the embedded bundle exactly once. If the
// embedded bundle cannot be decoded (build defect) it falls back to
// NoopTranslator so call sites never panic — the failure is still
// observable via DefaultTranslatorErr for tests/diagnostics.
func DefaultTranslator() Translator {
	defaultBundleOnce.Do(func() {
		defaultBundle, defaultBundleErr = NewBundleTranslator()
	})
	if defaultBundleErr != nil || defaultBundle == nil {
		return NoopTranslator{}
	}
	return defaultBundle
}

// DefaultTranslatorErr reports the error (if any) encountered while
// loading the embedded default bundle. nil means the bundle loaded
// successfully.
func DefaultTranslatorErr() error {
	defaultBundleOnce.Do(func() {
		defaultBundle, defaultBundleErr = NewBundleTranslator()
	})
	return defaultBundleErr
}

// T implements Translator. A present key is formatted with args via
// fmt.Sprintf against its bundle value (a real format string). A
// missing key returns the key verbatim — args are NOT applied to
// the bare key, so no `%!(EXTRA ...)` marker can ever be produced
// for a missing translation (HXC-081 guarantee).
func (b *BundleTranslator) T(key string, args ...any) string {
	if b == nil || b.messages == nil {
		return key
	}
	format, ok := b.messages[key]
	if !ok {
		// Missing key: surface the key visibly, never Sprintf the
		// bare key with args (that is exactly the HXC-081 bug).
		return key
	}
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
