package usecase

import (
	"reflect"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestShowWorkflowReturnsOrderedRecommendedWorkflow(t *testing.T) {
	result, err := NewShowWorkflow().Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Workflow.Title != domain.RecommendedWorkflowTitle {
		t.Fatalf("Title = %q, want %q", result.Workflow.Title, domain.RecommendedWorkflowTitle)
	}

	want := []domain.WorkflowStepID{
		domain.WorkflowStepIDSpecAuthor,
		domain.WorkflowStepIDArchitectureReviewer,
		domain.WorkflowStepIDImplementer,
		domain.WorkflowStepIDTestEngineer,
		domain.WorkflowStepIDChangeReviewer,
		domain.WorkflowStepIDCommit,
		domain.WorkflowStepIDPullRequest,
		domain.WorkflowStepIDMerge,
		domain.WorkflowStepIDArchive,
	}
	if got := workflowStepIDs(result.Workflow); !reflect.DeepEqual(got, want) {
		t.Fatalf("workflow step ids = %v, want %v", got, want)
	}
}

func TestShowWorkflowIncludesRequiredStepData(t *testing.T) {
	result, err := NewShowWorkflow().Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	steps := result.Workflow.Steps()
	if len(steps) != 9 {
		t.Fatalf("step count = %d, want 9", len(steps))
	}
	for _, step := range steps {
		if step.ID == "" {
			t.Fatalf("step id is empty: %+v", step)
		}
		if step.DisplayName == "" {
			t.Fatalf("step display name is empty: %+v", step)
		}
		if step.Description == "" {
			t.Fatalf("step description is empty: %+v", step)
		}
		if step.Mode != domain.WorkflowStepModeAgentAssisted && step.Mode != domain.WorkflowStepModeManual {
			t.Fatalf("step mode = %q, want supported workflow mode", step.Mode)
		}
	}
}

func TestShowWorkflowIncludesAdvisoryCommandSuggestions(t *testing.T) {
	result, err := NewShowWorkflow().Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	stepsByID := workflowStepsByID(result.Workflow)
	wantCommands := map[domain.WorkflowStepID]string{
		domain.WorkflowStepIDSpecAuthor:           "specharbor generate <change-id> --guided",
		domain.WorkflowStepIDArchitectureReviewer: "specharbor validate <change-id>",
		domain.WorkflowStepIDImplementer:          "specharbor prompt <change-id> --role implementer",
		domain.WorkflowStepIDTestEngineer:         "specharbor prompt <change-id> --role test-engineer",
		domain.WorkflowStepIDChangeReviewer:       "specharbor review <change-id>",
		domain.WorkflowStepIDArchive:              "specharbor archive <change-id>",
	}

	for id, want := range wantCommands {
		commands := commandSuggestionText(stepsByID[id].CommandSuggestions())
		if !strings.Contains(commands, want) {
			t.Fatalf("%s command suggestions = %q, want %q", id, commands, want)
		}
	}

	for _, id := range []domain.WorkflowStepID{
		domain.WorkflowStepIDCommit,
		domain.WorkflowStepIDPullRequest,
		domain.WorkflowStepIDMerge,
	} {
		if len(stepsByID[id].CommandSuggestions()) != 0 {
			t.Fatalf("%s command suggestions = %v, want none", id, stepsByID[id].CommandSuggestions())
		}
	}
}

func TestShowWorkflowIncludesSafetyNotesForManualSourceControlSteps(t *testing.T) {
	result, err := NewShowWorkflow().Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	stepsByID := workflowStepsByID(result.Workflow)
	wantSnippets := map[domain.WorkflowStepID]string{
		domain.WorkflowStepIDCommit:      "does not commit",
		domain.WorkflowStepIDPullRequest: "does not create PRs",
		domain.WorkflowStepIDMerge:       "does not merge",
		domain.WorkflowStepIDArchive:     "does not archive automatically",
	}

	for id, want := range wantSnippets {
		notes := strings.Join(stepsByID[id].SafetyNotes(), "\n")
		if !strings.Contains(notes, want) {
			t.Fatalf("%s safety notes = %q, want %q", id, notes, want)
		}
	}
}

func TestShowWorkflowDoesNotRequireExternalCollaborators(t *testing.T) {
	useCase := NewShowWorkflow()

	first, err := useCase.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	second, err := useCase.Execute()
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}

	if !reflect.DeepEqual(workflowStepIDs(first.Workflow), workflowStepIDs(second.Workflow)) {
		t.Fatalf("workflow step ids changed between executions: %v then %v", workflowStepIDs(first.Workflow), workflowStepIDs(second.Workflow))
	}
}

func TestShowWorkflowRejectsNilUseCase(t *testing.T) {
	_, err := (*ShowWorkflow)(nil).Execute()
	if err == nil || !strings.Contains(err.Error(), "show workflow use case is required") {
		t.Fatalf("nil use case error = %v, want show workflow use case is required", err)
	}
}

func workflowStepIDs(workflow domain.Workflow) []domain.WorkflowStepID {
	steps := workflow.Steps()
	stepIDs := make([]domain.WorkflowStepID, 0, len(steps))
	for _, step := range steps {
		stepIDs = append(stepIDs, step.ID)
	}
	return stepIDs
}

func workflowStepsByID(workflow domain.Workflow) map[domain.WorkflowStepID]domain.WorkflowStep {
	stepsByID := map[domain.WorkflowStepID]domain.WorkflowStep{}
	for _, step := range workflow.Steps() {
		stepsByID[step.ID] = step
	}
	return stepsByID
}

func commandSuggestionText(suggestions []domain.WorkflowCommandSuggestion) string {
	var builder strings.Builder
	for _, suggestion := range suggestions {
		builder.WriteString(suggestion.Command)
		builder.WriteByte('\n')
		builder.WriteString(suggestion.Description)
		builder.WriteByte('\n')
	}
	return builder.String()
}
