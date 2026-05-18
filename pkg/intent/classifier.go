package intent

import (
	"strings"

	"digital.vasic.helixspecifier/pkg/i18n"
	"digital.vasic.helixspecifier/pkg/types"
)

// Classifier provides effort-level classification for requests.
//
// The translator resolves user-facing reasoning strings per
// CONST-046. A nil translator is treated as a NoopTranslator at
// resolution time, so the zero value remains usable and existing
// constructors stay backward compatible (CONST-051(B) decoupling
// — no project-specific defaults baked in).
type Classifier struct {
	translator i18n.Translator
}

// NewClassifier creates a new effort classifier with the default
// NoopTranslator. Production callers SHOULD prefer
// NewClassifierWithTranslator and inject a locale-aware backend.
func NewClassifier() *Classifier {
	return &Classifier{translator: i18n.NewNoopTranslator()}
}

// NewClassifierWithTranslator creates a classifier with an
// injected translator. Passing nil falls back to NoopTranslator.
func NewClassifierWithTranslator(t i18n.Translator) *Classifier {
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	return &Classifier{translator: t}
}

// Classify determines the effort level of a request
func (c *Classifier) Classify(
	request string,
) *types.EffortClassification {
	lower := strings.ToLower(request)

	result := &types.EffortClassification{
		Level:      types.EffortMedium,
		Confidence: 0.5,
		Signals:    []string{},
	}

	// Epic indicators
	epicSignals := []string{
		"entire system", "from scratch", "complete rewrite",
		"full platform", "whole architecture",
		"migrate everything", "rebuild all",
	}
	for _, sig := range epicSignals {
		if strings.Contains(lower, sig) {
			result.Level = types.EffortEpic
			result.Confidence = 0.8
			result.Signals = append(
				result.Signals, "epic:"+sig,
			)
		}
	}

	// Large indicators
	if result.Level == types.EffortMedium {
		largeSignals := []string{
			"implement", "build", "create module",
			"new system", "add feature", "develop",
			"authentication", "authorization",
			"database", "api", "service",
		}
		matches := 0
		for _, sig := range largeSignals {
			if strings.Contains(lower, sig) {
				matches++
				result.Signals = append(
					result.Signals, "large:"+sig,
				)
			}
		}
		if matches >= 2 || len(request) > 200 {
			result.Level = types.EffortLarge
			result.Confidence = 0.7
		}
	}

	// Quick indicators
	quickSignals := []string{
		"fix typo", "add log", "rename",
		"update comment", "change color",
		"bump version", "add import",
	}
	for _, sig := range quickSignals {
		if strings.Contains(lower, sig) {
			result.Level = types.EffortQuick
			result.Confidence = 0.85
			result.Signals = append(
				result.Signals, "quick:"+sig,
			)
		}
	}

	// Refactoring indicators boost to at least Large
	refactorSignals := []string{
		"refactor", "restructure", "rewrite", "migrate",
	}
	for _, sig := range refactorSignals {
		if strings.Contains(lower, sig) {
			if result.Level == types.EffortMedium ||
				result.Level == types.EffortQuick {
				result.Level = types.EffortLarge
				result.Confidence = 0.75
			}
			result.Signals = append(
				result.Signals, "refactor:"+sig,
			)
		}
	}

	// Set derived fields
	result.CeremonyLevel = types.CeremonyForEffort(
		result.Level,
	)
	result.RequiresDebate =
		result.Level != types.EffortQuick
	result.RequiresSpecKit =
		result.Level == types.EffortLarge ||
			result.Level == types.EffortEpic
	result.Reasoning = c.generateReasoning(result)

	return result
}

// generateReasoning explains the classification. Reasoning
// strings are resolved through the i18n translator (CONST-046).
// A nil translator (defensive — should not occur via either
// constructor) is treated as NoopTranslator so the function
// stays panic-free, matching the anti-bluff contract that
// missing translations surface visibly rather than crash.
func (c *Classifier) generateReasoning(
	cl *types.EffortClassification,
) string {
	t := c.translator
	if t == nil {
		t = i18n.NewNoopTranslator()
	}
	switch cl.Level {
	case types.EffortQuick:
		return t.T("helixspecifier_intent_quick_reasoning")
	case types.EffortMedium:
		return t.T("helixspecifier_intent_medium_reasoning")
	case types.EffortLarge:
		return t.T("helixspecifier_intent_large_reasoning")
	case types.EffortEpic:
		return t.T("helixspecifier_intent_epic_reasoning")
	default:
		return t.T("helixspecifier_intent_unknown_reasoning")
	}
}
