package intent

import (
	"strings"

	"digital.vasic.helixspecifier/pkg/types"
)

// Classifier provides effort-level classification for requests
type Classifier struct{}

// NewClassifier creates a new effort classifier
func NewClassifier() *Classifier {
	return &Classifier{}
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

// generateReasoning explains the classification
func (c *Classifier) generateReasoning(
	cl *types.EffortClassification,
) string {
	switch cl.Level {
	case types.EffortQuick:
		return "Simple task requiring minimal process"
	case types.EffortMedium:
		return "Moderate task with light planning"
	case types.EffortLarge:
		return "Substantial task requiring full SDD workflow"
	case types.EffortEpic:
		return "Major undertaking requiring maximum ceremony"
	default:
		return "Unknown effort level"
	}
}
