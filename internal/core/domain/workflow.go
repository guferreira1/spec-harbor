package domain

import (
	"errors"
	"sort"
)

type WorkflowStepID string

const (
	WorkflowStepIDSpecAuthor           WorkflowStepID = WorkflowStepID(PromptRoleSpecAuthor)
	WorkflowStepIDArchitectureReviewer WorkflowStepID = WorkflowStepID(PromptRoleArchitectureReviewer)
	WorkflowStepIDImplementer          WorkflowStepID = WorkflowStepID(PromptRoleImplementer)
	WorkflowStepIDTestEngineer         WorkflowStepID = WorkflowStepID(PromptRoleTestEngineer)
	WorkflowStepIDChangeReviewer       WorkflowStepID = WorkflowStepID(PromptRoleChangeReviewer)
	WorkflowStepIDCommit               WorkflowStepID = "commit"
	WorkflowStepIDPullRequest          WorkflowStepID = "pull-request"
	WorkflowStepIDMerge                WorkflowStepID = "merge"
	WorkflowStepIDArchive              WorkflowStepID = "archive"
)

type WorkflowStepMode string

const (
	WorkflowStepModeAgentAssisted WorkflowStepMode = "agent-assisted"
	WorkflowStepModeManual        WorkflowStepMode = "manual"
)

const RecommendedWorkflowTitle = "OpenSpec/SDD agent-driven workflow"

type WorkflowCommandSuggestion struct {
	Command     string
	Description string
}

type WorkflowStep struct {
	ID           WorkflowStepID
	DisplayName  string
	Description  string
	Order        int
	Mode         WorkflowStepMode
	Supported    bool
	AdvisoryOnly bool

	requires           []WorkflowStepID
	commandSuggestions []WorkflowCommandSuggestion
	safetyNotes        []string
}

func NewWorkflowStep(
	id WorkflowStepID,
	displayName string,
	description string,
	order int,
	mode WorkflowStepMode,
	supported bool,
	advisoryOnly bool,
	requires []WorkflowStepID,
	commandSuggestions []WorkflowCommandSuggestion,
	safetyNotes []string,
) WorkflowStep {
	return WorkflowStep{
		ID:                 id,
		DisplayName:        displayName,
		Description:        description,
		Order:              order,
		Mode:               mode,
		Supported:          supported,
		AdvisoryOnly:       advisoryOnly,
		requires:           append([]WorkflowStepID(nil), requires...),
		commandSuggestions: append([]WorkflowCommandSuggestion(nil), commandSuggestions...),
		safetyNotes:        append([]string(nil), safetyNotes...),
	}
}

func (step WorkflowStep) Requires() []WorkflowStepID {
	return append([]WorkflowStepID(nil), step.requires...)
}

func (step WorkflowStep) CommandSuggestions() []WorkflowCommandSuggestion {
	return append([]WorkflowCommandSuggestion(nil), step.commandSuggestions...)
}

func (step WorkflowStep) SafetyNotes() []string {
	return append([]string(nil), step.safetyNotes...)
}

type Workflow struct {
	Title string

	steps []WorkflowStep
}

func NewWorkflow(title string, steps []WorkflowStep) (Workflow, error) {
	copiedSteps := copyWorkflowSteps(steps)
	sort.SliceStable(copiedSteps, func(left int, right int) bool {
		return copiedSteps[left].Order < copiedSteps[right].Order
	})

	if err := validateWorkflowSteps(copiedSteps); err != nil {
		return Workflow{}, err
	}

	return Workflow{
		Title: title,
		steps: copiedSteps,
	}, nil
}

func (workflow Workflow) Steps() []WorkflowStep {
	return copyWorkflowSteps(workflow.steps)
}

func RecommendedWorkflow() (Workflow, error) {
	return NewWorkflow(RecommendedWorkflowTitle, []WorkflowStep{
		NewWorkflowStep(
			WorkflowStepIDSpecAuthor,
			"Spec Author Agent",
			"Create or refine the OpenSpec change package.",
			1,
			WorkflowStepModeAgentAssisted,
			true,
			false,
			nil,
			[]WorkflowCommandSuggestion{
				{
					Command:     `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"`,
					Description: "Create a guided OpenSpec change package.",
				},
				{
					Command:     "specharbor prompt <change-id> --role spec-author",
					Description: "Generate the spec author role prompt.",
				},
			},
			nil,
		),
		NewWorkflowStep(
			WorkflowStepIDArchitectureReviewer,
			"Architecture Reviewer Agent",
			"Review the proposed scope and design against architecture boundaries.",
			2,
			WorkflowStepModeAgentAssisted,
			true,
			false,
			[]WorkflowStepID{WorkflowStepIDSpecAuthor},
			[]WorkflowCommandSuggestion{
				{
					Command:     "specharbor validate <change-id>",
					Description: "Check required OpenSpec change files.",
				},
				{
					Command:     "specharbor prompt <change-id> --role architecture-reviewer",
					Description: "Generate the architecture reviewer role prompt.",
				},
			},
			nil,
		),
		NewWorkflowStep(
			WorkflowStepIDImplementer,
			"Implementer Agent",
			"Apply the approved OpenSpec change.",
			3,
			WorkflowStepModeAgentAssisted,
			true,
			false,
			[]WorkflowStepID{WorkflowStepIDArchitectureReviewer},
			[]WorkflowCommandSuggestion{
				{
					Command:     "specharbor prompt <change-id> --role implementer",
					Description: "Generate the implementer role prompt.",
				},
			},
			nil,
		),
		NewWorkflowStep(
			WorkflowStepIDTestEngineer,
			"Test Engineer Agent",
			"Add or run focused verification for the implemented change.",
			4,
			WorkflowStepModeAgentAssisted,
			true,
			false,
			[]WorkflowStepID{WorkflowStepIDImplementer},
			[]WorkflowCommandSuggestion{
				{
					Command:     "specharbor prompt <change-id> --role test-engineer",
					Description: "Generate the test engineer role prompt.",
				},
			},
			nil,
		),
		NewWorkflowStep(
			WorkflowStepIDChangeReviewer,
			"Change Reviewer Agent",
			"Review the final diff, task state, and verification evidence.",
			5,
			WorkflowStepModeAgentAssisted,
			true,
			false,
			[]WorkflowStepID{WorkflowStepIDTestEngineer},
			[]WorkflowCommandSuggestion{
				{
					Command:     "specharbor review <change-id>",
					Description: "Review local task completion and change package state.",
				},
				{
					Command:     "specharbor prompt <change-id> --role change-reviewer",
					Description: "Generate the change reviewer role prompt.",
				},
			},
			nil,
		),
		NewWorkflowStep(
			WorkflowStepIDCommit,
			"Commit",
			"Commit the reviewed local changes manually.",
			6,
			WorkflowStepModeManual,
			false,
			true,
			[]WorkflowStepID{WorkflowStepIDChangeReviewer},
			nil,
			[]string{
				"SpecHarbor does not commit, stage files, modify branches, push, or sign commits.",
			},
		),
		NewWorkflowStep(
			WorkflowStepIDPullRequest,
			"Pull Request",
			"Open a pull request manually in your source-control workflow.",
			7,
			WorkflowStepModeManual,
			false,
			true,
			[]WorkflowStepID{WorkflowStepIDCommit},
			nil,
			[]string{
				"SpecHarbor does not create PRs, call source-control APIs, set reviewers, edit labels, or inspect remote branches.",
			},
		),
		NewWorkflowStep(
			WorkflowStepIDMerge,
			"Merge",
			"Merge manually after your review and CI process passes.",
			8,
			WorkflowStepModeManual,
			false,
			true,
			[]WorkflowStepID{WorkflowStepIDPullRequest},
			nil,
			[]string{
				"SpecHarbor does not merge, approve, inspect CI, trigger CI, or update remote repositories.",
			},
		),
		NewWorkflowStep(
			WorkflowStepIDArchive,
			"Archive",
			"Archive the completed OpenSpec change after the work is merged or otherwise accepted.",
			9,
			WorkflowStepModeManual,
			true,
			false,
			[]WorkflowStepID{WorkflowStepIDMerge},
			[]WorkflowCommandSuggestion{
				{
					Command:     "specharbor archive <change-id>",
					Description: "Archive the completed OpenSpec change explicitly.",
				},
			},
			[]string{
				"specharbor workflow does not archive automatically; specharbor archive <change-id> remains an explicit user command.",
			},
		),
	})
}

func copyWorkflowSteps(steps []WorkflowStep) []WorkflowStep {
	copiedSteps := make([]WorkflowStep, len(steps))
	for index, step := range steps {
		copiedSteps[index] = NewWorkflowStep(
			step.ID,
			step.DisplayName,
			step.Description,
			step.Order,
			step.Mode,
			step.Supported,
			step.AdvisoryOnly,
			step.requires,
			step.commandSuggestions,
			step.safetyNotes,
		)
	}
	return copiedSteps
}

func validateWorkflowSteps(steps []WorkflowStep) error {
	if len(steps) == 0 {
		return errors.New("workflow steps are required")
	}

	knownStepIDs := make(map[WorkflowStepID]bool, len(steps))
	usedOrders := make(map[int]bool, len(steps))
	for _, step := range steps {
		if step.ID == "" {
			return errors.New("workflow step id is required")
		}
		if knownStepIDs[step.ID] {
			return errors.New("workflow step ids must be unique")
		}
		if step.Order <= 0 {
			return errors.New("workflow step order must be positive")
		}
		if usedOrders[step.Order] {
			return errors.New("workflow step orders must be unique")
		}
		knownStepIDs[step.ID] = true
		usedOrders[step.Order] = true
	}

	for _, step := range steps {
		for _, requiredStepID := range step.requires {
			if !knownStepIDs[requiredStepID] {
				return errors.New("workflow dependencies must reference known step ids")
			}
		}
	}

	return nil
}
