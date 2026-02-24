package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- AllPhases tests ---

func TestAllPhases_Count(t *testing.T) {
	phases := AllPhases()
	assert.Len(t, phases, 7, "AllPhases should return 7 phases")
}

func TestAllPhases_Order(t *testing.T) {
	phases := AllPhases()
	expected := []SpecKitPhase{
		PhaseConstitution,
		PhaseSpecify,
		PhaseClarify,
		PhasePlan,
		PhaseTasks,
		PhaseAnalyze,
		PhaseImplement,
	}
	assert.Equal(t, expected, phases,
		"AllPhases must return phases in correct order")
}

// --- AllStages tests ---

func TestAllStages_Count(t *testing.T) {
	stages := AllStages()
	assert.Len(t, stages, 5, "AllStages should return 5 stages")
}

func TestAllStages_Order(t *testing.T) {
	stages := AllStages()
	expected := []FusionStage{
		StageIdeation,
		StageConstitution,
		StagePlanning,
		StageExecution,
		StageVerification,
	}
	assert.Equal(t, expected, stages,
		"AllStages must return stages in correct order")
}

// --- EffortLevels tests ---

func TestEffortLevels_Count(t *testing.T) {
	levels := EffortLevels()
	assert.Len(t, levels, 4,
		"EffortLevels should return 4 levels")
}

// --- CeremonyForEffort tests ---

func TestCeremonyForEffort_Quick(t *testing.T) {
	result := CeremonyForEffort(EffortQuick)
	assert.Equal(t, CeremonyMinimal, result,
		"Quick effort should map to minimal ceremony")
}

func TestCeremonyForEffort_Medium(t *testing.T) {
	result := CeremonyForEffort(EffortMedium)
	assert.Equal(t, CeremonyLight, result,
		"Medium effort should map to light ceremony")
}

func TestCeremonyForEffort_Large(t *testing.T) {
	result := CeremonyForEffort(EffortLarge)
	assert.Equal(t, CeremonyStandard, result,
		"Large effort should map to standard ceremony")
}

func TestCeremonyForEffort_Epic(t *testing.T) {
	result := CeremonyForEffort(EffortEpic)
	assert.Equal(t, CeremonyHeavy, result,
		"Epic effort should map to heavy ceremony")
}

func TestCeremonyForEffort_Default(t *testing.T) {
	result := CeremonyForEffort(EffortLevel("unknown"))
	assert.Equal(t, CeremonyStandard, result,
		"Unknown effort should default to standard ceremony")
}

// --- StagesForEffort tests ---

func TestStagesForEffort_Quick(t *testing.T) {
	stages := StagesForEffort(EffortQuick)
	assert.Len(t, stages, 1,
		"Quick effort should have 1 stage")
	assert.Equal(t, StageExecution, stages[0],
		"Quick effort stage should be execution")
}

func TestStagesForEffort_Medium(t *testing.T) {
	stages := StagesForEffort(EffortMedium)
	assert.Len(t, stages, 3,
		"Medium effort should have 3 stages")
	expected := []FusionStage{
		StagePlanning,
		StageExecution,
		StageVerification,
	}
	assert.Equal(t, expected, stages,
		"Medium effort stages must be in correct order")
}

func TestStagesForEffort_Large(t *testing.T) {
	stages := StagesForEffort(EffortLarge)
	assert.Len(t, stages, 5,
		"Large effort should have 5 stages (all)")
	assert.Equal(t, AllStages(), stages,
		"Large effort should return all stages")
}

func TestStagesForEffort_Epic(t *testing.T) {
	stages := StagesForEffort(EffortEpic)
	assert.Len(t, stages, 5,
		"Epic effort should have 5 stages (all)")
	assert.Equal(t, AllStages(), stages,
		"Epic effort should return all stages")
}

func TestStagesForEffort_Default(t *testing.T) {
	stages := StagesForEffort(EffortLevel("unknown"))
	assert.Equal(t, AllStages(), stages,
		"Unknown effort should default to all stages")
}

// --- PhasesForCeremony tests ---

func TestPhasesForCeremony_Minimal(t *testing.T) {
	phases := PhasesForCeremony(CeremonyMinimal)
	assert.Len(t, phases, 1,
		"Minimal ceremony should have 1 phase")
	assert.Equal(t, PhaseImplement, phases[0],
		"Minimal ceremony phase should be implement")
}

func TestPhasesForCeremony_Light(t *testing.T) {
	phases := PhasesForCeremony(CeremonyLight)
	assert.Len(t, phases, 3,
		"Light ceremony should have 3 phases")
	expected := []SpecKitPhase{
		PhaseSpecify,
		PhasePlan,
		PhaseImplement,
	}
	assert.Equal(t, expected, phases,
		"Light ceremony phases must be in correct order")
}

func TestPhasesForCeremony_Standard(t *testing.T) {
	phases := PhasesForCeremony(CeremonyStandard)
	assert.Len(t, phases, 7,
		"Standard ceremony should have 7 phases (all)")
	assert.Equal(t, AllPhases(), phases,
		"Standard ceremony should return all phases")
}

func TestPhasesForCeremony_Heavy(t *testing.T) {
	phases := PhasesForCeremony(CeremonyHeavy)
	assert.Len(t, phases, 7,
		"Heavy ceremony should have 7 phases (all)")
	assert.Equal(t, AllPhases(), phases,
		"Heavy ceremony should return all phases")
}

func TestPhasesForCeremony_Default(t *testing.T) {
	phases := PhasesForCeremony(CeremonyLevel("unknown"))
	assert.Equal(t, AllPhases(), phases,
		"Unknown ceremony should default to all phases")
}

// --- Enum value tests ---

func TestEffortLevel_Values(t *testing.T) {
	tests := []struct {
		name     string
		level    EffortLevel
		expected string
	}{
		{"quick", EffortQuick, "quick"},
		{"medium", EffortMedium, "medium"},
		{"large", EffortLarge, "large"},
		{"epic", EffortEpic, "epic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.level))
		})
	}
}

func TestFusionStage_Values(t *testing.T) {
	tests := []struct {
		name     string
		stage    FusionStage
		expected string
	}{
		{"ideation", StageIdeation, "ideation"},
		{"constitution", StageConstitution, "constitution"},
		{"planning", StagePlanning, "planning"},
		{"execution", StageExecution, "execution"},
		{"verification", StageVerification, "verification"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.stage))
		})
	}
}

func TestSpecKitPhase_Values(t *testing.T) {
	tests := []struct {
		name     string
		phase    SpecKitPhase
		expected string
	}{
		{"constitution", PhaseConstitution, "constitution"},
		{"specify", PhaseSpecify, "specify"},
		{"clarify", PhaseClarify, "clarify"},
		{"plan", PhasePlan, "plan"},
		{"tasks", PhaseTasks, "tasks"},
		{"analyze", PhaseAnalyze, "analyze"},
		{"implement", PhaseImplement, "implement"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.phase))
		})
	}
}

func TestCeremonyLevel_Values(t *testing.T) {
	tests := []struct {
		name     string
		level    CeremonyLevel
		expected string
	}{
		{"minimal", CeremonyMinimal, "minimal"},
		{"light", CeremonyLight, "light"},
		{"standard", CeremonyStandard, "standard"},
		{"heavy", CeremonyHeavy, "heavy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.level))
		})
	}
}

func TestMilestoneStatus_Values(t *testing.T) {
	tests := []struct {
		name     string
		status   MilestoneStatus
		expected string
	}{
		{"queued", MilestoneQueued, "queued"},
		{"active", MilestoneActive, "active"},
		{"blocked", MilestoneBlocked, "blocked"},
		{"completed", MilestoneCompleted, "completed"},
		{"failed", MilestoneFailed, "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestSpecSource_Values(t *testing.T) {
	tests := []struct {
		name     string
		source   SpecSource
		expected string
	}{
		{"speckit", SourceSpecKit, "speckit"},
		{"superpowers", SourceSuperpowers, "superpowers"},
		{"gsd", SourceGSD, "gsd"},
		{"fusion", SourceFusion, "fusion"},
		{"user", SourceUser, "user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.source))
		})
	}
}

// --- Cross-validation tests ---

func TestCeremonyForEffort_AllLevels(t *testing.T) {
	for _, level := range EffortLevels() {
		result := CeremonyForEffort(level)
		assert.NotEmpty(t, string(result),
			"CeremonyForEffort(%s) should not be empty",
			level)
	}
}

func TestStagesForEffort_AllLevels(t *testing.T) {
	for _, level := range EffortLevels() {
		stages := StagesForEffort(level)
		assert.NotEmpty(t, stages,
			"StagesForEffort(%s) should not be empty",
			level)
	}
}

func TestPhasesForCeremony_IncreasingPhases(t *testing.T) {
	minimal := PhasesForCeremony(CeremonyMinimal)
	light := PhasesForCeremony(CeremonyLight)
	standard := PhasesForCeremony(CeremonyStandard)
	heavy := PhasesForCeremony(CeremonyHeavy)

	assert.True(t, len(minimal) <= len(light),
		"Minimal phases should be <= light phases")
	assert.True(t, len(light) <= len(standard),
		"Light phases should be <= standard phases")
	assert.True(t, len(standard) <= len(heavy),
		"Standard phases should be <= heavy phases")
}

func TestStagesForEffort_IncreasingStages(t *testing.T) {
	quick := StagesForEffort(EffortQuick)
	medium := StagesForEffort(EffortMedium)
	large := StagesForEffort(EffortLarge)
	epic := StagesForEffort(EffortEpic)

	assert.True(t, len(quick) <= len(medium),
		"Quick stages should be <= medium stages")
	assert.True(t, len(medium) <= len(large),
		"Medium stages should be <= large stages")
	assert.True(t, len(large) <= len(epic),
		"Large stages should be <= epic stages")
}

func TestAllPhases_ContainsImplement(t *testing.T) {
	phases := AllPhases()
	assert.Contains(t, phases, PhaseImplement,
		"AllPhases must contain implement phase")
}

func TestAllStages_ContainsExecution(t *testing.T) {
	stages := AllStages()
	assert.Contains(t, stages, StageExecution,
		"AllStages must contain execution stage")
}

func TestAllPhases_FirstIsConstitution(t *testing.T) {
	phases := AllPhases()
	assert.Equal(t, PhaseConstitution, phases[0],
		"First phase must be constitution")
}

func TestAllPhases_LastIsImplement(t *testing.T) {
	phases := AllPhases()
	assert.Equal(t, PhaseImplement, phases[len(phases)-1],
		"Last phase must be implement")
}

func TestAllStages_FirstIsIdeation(t *testing.T) {
	stages := AllStages()
	assert.Equal(t, StageIdeation, stages[0],
		"First stage must be ideation")
}

func TestAllStages_LastIsVerification(t *testing.T) {
	stages := AllStages()
	last := stages[len(stages)-1]
	assert.Equal(t, StageVerification, last,
		"Last stage must be verification")
}
