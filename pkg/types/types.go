// Package types defines the core types and interfaces for the
// HelixSpecifier Spec-Driven Development Fusion Engine.
//
// HelixSpecifier fuses three pillars:
//   - SpecKit: 7-phase SDD workflow
//   - Superpowers: TDD/subagent execution
//   - GSD: Get-Shit-Done milestone lifecycle
package types

import (
	"context"
	"time"
)

// EffortLevel represents the scale of work effort.
type EffortLevel string

const (
	// EffortQuick is for tasks under 5 minutes with minimal ceremony.
	EffortQuick EffortLevel = "quick"
	// EffortMedium is for 30-90 minute tasks with light ceremony.
	EffortMedium EffortLevel = "medium"
	// EffortLarge is for 4-16 hour tasks with full ceremony.
	EffortLarge EffortLevel = "large"
	// EffortEpic is for days-to-weeks tasks with maximum ceremony.
	EffortEpic EffortLevel = "epic"
)

// FusionStage represents a stage in the fusion pipeline.
type FusionStage string

const (
	// StageIdeation is the ideation and foundation stage.
	StageIdeation FusionStage = "ideation"
	// StageConstitution is the constitution and specification stage.
	StageConstitution FusionStage = "constitution"
	// StagePlanning is the planning and task breakdown stage.
	StagePlanning FusionStage = "planning"
	// StageExecution is the execution stage.
	StageExecution FusionStage = "execution"
	// StageVerification is the verification and completion stage.
	StageVerification FusionStage = "verification"
)

// SpecKitPhase represents a phase in the SpecKit 7-phase SDD
// workflow.
type SpecKitPhase string

const (
	// PhaseConstitution establishes project constitution.
	PhaseConstitution SpecKitPhase = "constitution"
	// PhaseSpecify creates the specification document.
	PhaseSpecify SpecKitPhase = "specify"
	// PhaseClarify resolves ambiguities in the specification.
	PhaseClarify SpecKitPhase = "clarify"
	// PhasePlan creates the implementation plan.
	PhasePlan SpecKitPhase = "plan"
	// PhaseTasks breaks the plan into executable tasks.
	PhaseTasks SpecKitPhase = "tasks"
	// PhaseAnalyze performs analysis before implementation.
	PhaseAnalyze SpecKitPhase = "analyze"
	// PhaseImplement executes the implementation.
	PhaseImplement SpecKitPhase = "implement"
)

// CeremonyLevel represents the amount of process rigor.
type CeremonyLevel string

const (
	// CeremonyMinimal skips most phases.
	CeremonyMinimal CeremonyLevel = "minimal"
	// CeremonyLight applies quick spec and plan.
	CeremonyLight CeremonyLevel = "light"
	// CeremonyStandard applies the full 7-phase SDD.
	CeremonyStandard CeremonyLevel = "standard"
	// CeremonyHeavy extends standard with additional reviews.
	CeremonyHeavy CeremonyLevel = "heavy"
)

// MilestoneStatus represents GSD milestone lifecycle states.
type MilestoneStatus string

const (
	// MilestoneQueued means the milestone is waiting to start.
	MilestoneQueued MilestoneStatus = "queued"
	// MilestoneActive means the milestone is in progress.
	MilestoneActive MilestoneStatus = "active"
	// MilestoneBlocked means the milestone is blocked.
	MilestoneBlocked MilestoneStatus = "blocked"
	// MilestoneCompleted means the milestone finished successfully.
	MilestoneCompleted MilestoneStatus = "completed"
	// MilestoneFailed means the milestone failed.
	MilestoneFailed MilestoneStatus = "failed"
)

// SpecSource identifies where a specification originated.
type SpecSource string

const (
	// SourceSpecKit means the spec came from SpecKit workflow.
	SourceSpecKit SpecSource = "speckit"
	// SourceSuperpowers means the spec came from Superpowers.
	SourceSuperpowers SpecSource = "superpowers"
	// SourceGSD means the spec came from GSD lifecycle.
	SourceGSD SpecSource = "gsd"
	// SourceFusion means the spec came from fusion engine.
	SourceFusion SpecSource = "fusion"
	// SourceUser means the spec was provided by the user.
	SourceUser SpecSource = "user"
)

// --- Core Data Structures ---

// EffortClassification holds the result of effort analysis.
type EffortClassification struct {
	Level           EffortLevel   `json:"level"`
	Confidence      float64       `json:"confidence"`
	CeremonyLevel   CeremonyLevel `json:"ceremony_level"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	RequiresDebate  bool          `json:"requires_debate"`
	RequiresSpecKit bool          `json:"requires_speckit"`
	Signals         []string      `json:"signals"`
	Reasoning       string        `json:"reasoning"`
}

// Specification represents a unified specification document.
type Specification struct {
	ID           string                `json:"id"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	Source       SpecSource            `json:"source"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Requirements []Requirement         `json:"requirements"`
	Constraints  []Constraint          `json:"constraints"`
	Architecture *ArchitectureDecision `json:"architecture,omitempty"`
	TestStrategy *TestStrategy         `json:"test_strategy,omitempty"`
	QualityScore float64               `json:"quality_score"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Requirement represents a single requirement in a spec.
type Requirement struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Acceptance  string `json:"acceptance"`
}

// Constraint represents a project constraint.
type Constraint struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ArchitectureDecision records an architectural choice.
type ArchitectureDecision struct {
	Pattern      string   `json:"pattern"`
	Reasoning    string   `json:"reasoning"`
	Alternatives []string `json:"alternatives"`
	TradeOffs    string   `json:"trade_offs"`
}

// TestStrategy defines the testing approach.
type TestStrategy struct {
	UnitTests        bool `json:"unit_tests"`
	IntegrationTests bool `json:"integration_tests"`
	E2ETests         bool `json:"e2e_tests"`
	SecurityTests    bool `json:"security_tests"`
	StressTests      bool `json:"stress_tests"`
	BenchmarkTests   bool `json:"benchmark_tests"`
	ChallengeScript  bool `json:"challenge_script"`
}

// Milestone represents a GSD milestone.
type Milestone struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Status       MilestoneStatus `json:"status"`
	Tasks        []Task          `json:"tasks"`
	Dependencies []string        `json:"dependencies"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	Progress     float64         `json:"progress"`
}

// Task represents an atomic unit of work.
type Task struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Priority        string        `json:"priority"`
	Status          string        `json:"status"`
	EstimatedEffort time.Duration `json:"estimated_effort"`
	ActualEffort    time.Duration `json:"actual_effort,omitempty"`
	Dependencies    []string      `json:"dependencies"`
	Acceptance      []string      `json:"acceptance"`
	FilesAffected   []string      `json:"files_affected"`
	TestFiles       []string      `json:"test_files"`
}

// PhaseResult contains the output of a single phase execution.
type PhaseResult struct {
	Phase        SpecKitPhase          `json:"phase"`
	Stage        FusionStage           `json:"stage"`
	StartTime    time.Time             `json:"start_time"`
	EndTime      time.Time             `json:"end_time"`
	Duration     time.Duration         `json:"duration"`
	Success      bool                  `json:"success"`
	Output       string                `json:"output"`
	Artifacts    map[string]interface{} `json:"artifacts,omitempty"`
	QualityScore float64               `json:"quality_score"`
	DebateID     string                `json:"debate_id,omitempty"`
	Error        string                `json:"error,omitempty"`
	Cached       bool                  `json:"cached,omitempty"`
}

// FlowResult contains the complete result of a HelixSpecifier
// flow.
type FlowResult struct {
	FlowID              string                `json:"flow_id"`
	EffortLevel         EffortLevel           `json:"effort_level"`
	CeremonyLevel       CeremonyLevel         `json:"ceremony_level"`
	StartTime           time.Time             `json:"start_time"`
	EndTime             time.Time             `json:"end_time"`
	Duration            time.Duration         `json:"duration"`
	Success             bool                  `json:"success"`
	Stages              []FusionStage         `json:"stages"`
	PhaseResults        []PhaseResult         `json:"phase_results"`
	Specification       *Specification        `json:"specification,omitempty"`
	Milestones          []Milestone           `json:"milestones,omitempty"`
	FinalArtifact       string                `json:"final_artifact"`
	OverallQualityScore float64               `json:"overall_quality_score"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	ResumedFromCache    bool                  `json:"resumed_from_cache,omitempty"`
}

// --- Core Interfaces ---

// SpecEngine is the main interface for the spec-driven
// development fusion engine.
type SpecEngine interface {
	// ExecuteFlow runs the complete spec-driven development flow.
	ExecuteFlow(
		ctx context.Context,
		request string,
		classification *EffortClassification,
	) (*FlowResult, error)

	// ResumeFlow resumes a previously cached flow.
	ResumeFlow(
		ctx context.Context,
		flowID string,
		request string,
	) (*FlowResult, error)

	// ClassifyEffort determines the effort level for a request.
	ClassifyEffort(
		ctx context.Context,
		request string,
	) (*EffortClassification, error)

	// GetFlowStatus returns the current status of a flow.
	GetFlowStatus(flowID string) (*FlowResult, error)

	// Health returns health status of the engine.
	Health(ctx context.Context) error

	// Name returns the engine name.
	Name() string

	// Version returns the engine version.
	Version() string
}

// SpecKitPillar is the interface for the SpecKit component.
type SpecKitPillar interface {
	// ExecutePhase runs a single SpecKit phase.
	ExecutePhase(
		ctx context.Context,
		phase SpecKitPhase,
		input *PhaseInput,
	) (*PhaseResult, error)

	// GetPhaseOrder returns the ordered list of phases for a
	// given ceremony level.
	GetPhaseOrder(ceremony CeremonyLevel) []SpecKitPhase

	// ValidateTransition checks if a phase transition is valid.
	ValidateTransition(from, to SpecKitPhase) bool
}

// SuperpowersPillar is the interface for the Superpowers
// component.
type SuperpowersPillar interface {
	// ExecuteWithTDD runs implementation with TDD discipline.
	ExecuteWithTDD(
		ctx context.Context,
		tasks []Task,
	) ([]TaskResult, error)

	// DispatchSubagents dispatches parallel subagents for tasks.
	DispatchSubagents(
		ctx context.Context,
		tasks []Task,
		maxParallel int,
	) ([]TaskResult, error)

	// ReviewCode performs code review on completed tasks.
	ReviewCode(
		ctx context.Context,
		taskResults []TaskResult,
	) (*ReviewResult, error)
}

// GSDPillar is the interface for the Get-Shit-Done component.
type GSDPillar interface {
	// CreateMilestones breaks work into milestones.
	CreateMilestones(
		ctx context.Context,
		spec *Specification,
		tasks []Task,
	) ([]Milestone, error)

	// TrackProgress updates milestone progress.
	TrackProgress(
		ctx context.Context,
		milestoneID string,
		taskResults []TaskResult,
	) (*Milestone, error)

	// GetProgress returns overall progress for a flow.
	GetProgress(flowID string) (float64, error)
}

// CeremonyScaler determines appropriate ceremony level.
type CeremonyScaler interface {
	// Scale determines ceremony level based on effort.
	Scale(
		classification *EffortClassification,
	) CeremonyLevel

	// Adjust dynamically adjusts ceremony during execution.
	Adjust(
		current CeremonyLevel,
		phaseResults []PhaseResult,
	) CeremonyLevel
}

// SpecMemory provides spec memory and cross-project learning.
type SpecMemory interface {
	// Store saves a specification for future reference.
	Store(ctx context.Context, spec *Specification) error

	// Search finds similar specifications.
	Search(
		ctx context.Context,
		query string,
		limit int,
	) ([]Specification, error)

	// LearnFromFlow extracts patterns from completed flows.
	LearnFromFlow(
		ctx context.Context,
		flow *FlowResult,
	) error
}

// CLIAgentAdapter adapts the engine for specific CLI agents.
type CLIAgentAdapter interface {
	// FormatOutput formats engine output for the agent.
	FormatOutput(result *FlowResult) (string, error)

	// ParseInput parses agent-specific input.
	ParseInput(input string) (string, error)

	// GetAgentType returns the agent type name.
	GetAgentType() string

	// SupportsStreaming returns if the agent supports streaming.
	SupportsStreaming() bool
}

// --- Supporting Types ---

// PhaseInput contains input for a phase execution.
type PhaseInput struct {
	UserRequest    string                `json:"user_request"`
	PreviousOutput string                `json:"previous_output"`
	Constitution   string                `json:"constitution"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TaskResult contains the result of executing a task.
type TaskResult struct {
	TaskID       string        `json:"task_id"`
	Success      bool          `json:"success"`
	Output       string        `json:"output"`
	FilesChanged []string      `json:"files_changed"`
	TestsPassed  int           `json:"tests_passed"`
	TestsFailed  int           `json:"tests_failed"`
	Duration     time.Duration `json:"duration"`
	Error        string        `json:"error,omitempty"`
}

// ReviewResult contains code review findings.
type ReviewResult struct {
	Approved    bool          `json:"approved"`
	Score       float64       `json:"score"`
	Issues      []ReviewIssue `json:"issues"`
	Suggestions []string      `json:"suggestions"`
}

// ReviewIssue represents a code review finding.
type ReviewIssue struct {
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line,omitempty"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// AllPhases returns all SpecKit phases in order.
func AllPhases() []SpecKitPhase {
	return []SpecKitPhase{
		PhaseConstitution,
		PhaseSpecify,
		PhaseClarify,
		PhasePlan,
		PhaseTasks,
		PhaseAnalyze,
		PhaseImplement,
	}
}

// AllStages returns all fusion stages in order.
func AllStages() []FusionStage {
	return []FusionStage{
		StageIdeation,
		StageConstitution,
		StagePlanning,
		StageExecution,
		StageVerification,
	}
}

// EffortLevels returns all effort levels.
func EffortLevels() []EffortLevel {
	return []EffortLevel{
		EffortQuick,
		EffortMedium,
		EffortLarge,
		EffortEpic,
	}
}

// CeremonyForEffort returns the default ceremony level for a
// given effort level.
func CeremonyForEffort(effort EffortLevel) CeremonyLevel {
	switch effort {
	case EffortQuick:
		return CeremonyMinimal
	case EffortMedium:
		return CeremonyLight
	case EffortLarge:
		return CeremonyStandard
	case EffortEpic:
		return CeremonyHeavy
	default:
		return CeremonyStandard
	}
}

// StagesForEffort returns the fusion stages appropriate for
// a given effort level.
func StagesForEffort(effort EffortLevel) []FusionStage {
	switch effort {
	case EffortQuick:
		return []FusionStage{StageExecution}
	case EffortMedium:
		return []FusionStage{
			StagePlanning,
			StageExecution,
			StageVerification,
		}
	case EffortLarge:
		return AllStages()
	case EffortEpic:
		return AllStages()
	default:
		return AllStages()
	}
}

// PhasesForCeremony returns the SpecKit phases appropriate
// for a given ceremony level.
func PhasesForCeremony(ceremony CeremonyLevel) []SpecKitPhase {
	switch ceremony {
	case CeremonyMinimal:
		return []SpecKitPhase{PhaseImplement}
	case CeremonyLight:
		return []SpecKitPhase{
			PhaseSpecify,
			PhasePlan,
			PhaseImplement,
		}
	case CeremonyStandard:
		return AllPhases()
	case CeremonyHeavy:
		return AllPhases()
	default:
		return AllPhases()
	}
}
