package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestWorkflowStepIDValues(t *testing.T) {
	want := []WorkflowStepID{
		"spec-author",
		"architecture-reviewer",
		"implementer",
		"test-engineer",
		"change-reviewer",
		"commit",
		"pull-request",
		"merge",
		"archive",
	}

	if !reflect.DeepEqual(recommendedWorkflowStepIDs(t), want) {
		t.Fatalf("recommended step ids = %v, want %v", recommendedWorkflowStepIDs(t), want)
	}
}

func TestWorkflowStepOrderingAndDisplayNames(t *testing.T) {
	workflow := recommendedWorkflow(t)
	if workflow.Title != RecommendedWorkflowTitle {
		t.Fatalf("Title = %q, want %q", workflow.Title, RecommendedWorkflowTitle)
	}

	wantDisplayNames := []string{
		"Spec Author Agent",
		"Architecture Reviewer Agent",
		"Implementer Agent",
		"Test Engineer Agent",
		"Change Reviewer Agent",
		"Commit",
		"Pull Request",
		"Merge",
		"Archive",
	}

	steps := workflow.Steps()
	if len(steps) != len(wantDisplayNames) {
		t.Fatalf("step count = %d, want %d", len(steps), len(wantDisplayNames))
	}
	for index, step := range steps {
		wantOrder := index + 1
		if step.Order != wantOrder {
			t.Fatalf("step %q order = %d, want %d", step.ID, step.Order, wantOrder)
		}
		if step.DisplayName != wantDisplayNames[index] {
			t.Fatalf("step %q display name = %q, want %q", step.ID, step.DisplayName, wantDisplayNames[index])
		}
		if step.Description == "" {
			t.Fatalf("step %q description is empty", step.ID)
		}
	}
}

func TestWorkflowStepModesSupportedAndAdvisoryFlags(t *testing.T) {
	stepsByID := recommendedWorkflowStepsByID(t)

	for _, id := range []WorkflowStepID{
		WorkflowStepIDSpecAuthor,
		WorkflowStepIDArchitectureReviewer,
		WorkflowStepIDImplementer,
		WorkflowStepIDTestEngineer,
		WorkflowStepIDChangeReviewer,
	} {
		step := stepsByID[id]
		if step.Mode != WorkflowStepModeAgentAssisted {
			t.Fatalf("%s mode = %q, want %q", id, step.Mode, WorkflowStepModeAgentAssisted)
		}
		if !step.Supported {
			t.Fatalf("%s Supported = false, want true", id)
		}
		if step.AdvisoryOnly {
			t.Fatalf("%s AdvisoryOnly = true, want false", id)
		}
	}

	for _, id := range []WorkflowStepID{WorkflowStepIDCommit, WorkflowStepIDPullRequest, WorkflowStepIDMerge} {
		step := stepsByID[id]
		if step.Mode != WorkflowStepModeManual {
			t.Fatalf("%s mode = %q, want %q", id, step.Mode, WorkflowStepModeManual)
		}
		if step.Supported {
			t.Fatalf("%s Supported = true, want false", id)
		}
		if !step.AdvisoryOnly {
			t.Fatalf("%s AdvisoryOnly = false, want true", id)
		}
	}

	archive := stepsByID[WorkflowStepIDArchive]
	if archive.Mode != WorkflowStepModeManual {
		t.Fatalf("archive mode = %q, want %q", archive.Mode, WorkflowStepModeManual)
	}
	if !archive.Supported {
		t.Fatalf("archive Supported = false, want true")
	}
	if archive.AdvisoryOnly {
		t.Fatalf("archive AdvisoryOnly = true, want false")
	}
}

func TestWorkflowStepDependencies(t *testing.T) {
	stepsByID := recommendedWorkflowStepsByID(t)
	want := map[WorkflowStepID][]WorkflowStepID{
		WorkflowStepIDSpecAuthor:           nil,
		WorkflowStepIDArchitectureReviewer: {WorkflowStepIDSpecAuthor},
		WorkflowStepIDImplementer:          {WorkflowStepIDArchitectureReviewer},
		WorkflowStepIDTestEngineer:         {WorkflowStepIDImplementer},
		WorkflowStepIDChangeReviewer:       {WorkflowStepIDTestEngineer},
		WorkflowStepIDCommit:               {WorkflowStepIDChangeReviewer},
		WorkflowStepIDPullRequest:          {WorkflowStepIDCommit},
		WorkflowStepIDMerge:                {WorkflowStepIDPullRequest},
		WorkflowStepIDArchive:              {WorkflowStepIDMerge},
	}

	for id, dependencies := range want {
		if !reflect.DeepEqual(stepsByID[id].Requires(), dependencies) {
			t.Fatalf("%s dependencies = %v, want %v", id, stepsByID[id].Requires(), dependencies)
		}
	}
}

func TestWorkflowDependenciesReferenceKnownStepIDs(t *testing.T) {
	steps := recommendedWorkflow(t).Steps()
	known := map[WorkflowStepID]bool{}
	for _, step := range steps {
		known[step.ID] = true
	}

	for _, step := range steps {
		for _, requiredStepID := range step.Requires() {
			if !known[requiredStepID] {
				t.Fatalf("%s requires unknown step id %s", step.ID, requiredStepID)
			}
		}
	}

	_, err := NewWorkflow("invalid", []WorkflowStep{
		NewWorkflowStep("only-step", "Only", "Only step.", 1, WorkflowStepModeManual, true, false, []WorkflowStepID{"missing"}, nil, nil),
	})
	if err == nil || !strings.Contains(err.Error(), "workflow dependencies must reference known step ids") {
		t.Fatalf("NewWorkflow() error = %v, want unknown dependency error", err)
	}
}

func TestWorkflowCommandSuggestionMetadata(t *testing.T) {
	stepsByID := recommendedWorkflowStepsByID(t)

	specAuthorSuggestions := stepsByID[WorkflowStepIDSpecAuthor].CommandSuggestions()
	if len(specAuthorSuggestions) != 2 {
		t.Fatalf("spec author suggestions = %v, want 2 suggestions", specAuthorSuggestions)
	}
	if specAuthorSuggestions[0].Command != `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"` {
		t.Fatalf("spec author first command = %q, want guided generation command", specAuthorSuggestions[0].Command)
	}
	if specAuthorSuggestions[0].Description == "" {
		t.Fatalf("spec author first command description is empty")
	}

	architectureReviewerSuggestions := stepsByID[WorkflowStepIDArchitectureReviewer].CommandSuggestions()
	wantArchitectureCommands := []string{
		"specharbor validate <change-id>",
		"specharbor prompt <change-id> --role architecture-reviewer",
	}
	for index, want := range wantArchitectureCommands {
		if architectureReviewerSuggestions[index].Command != want {
			t.Fatalf("architecture suggestion %d command = %q, want %q", index, architectureReviewerSuggestions[index].Command, want)
		}
		if architectureReviewerSuggestions[index].Description == "" {
			t.Fatalf("architecture suggestion %d description is empty", index)
		}
	}

	if len(stepsByID[WorkflowStepIDCommit].CommandSuggestions()) != 0 {
		t.Fatalf("commit suggestions = %v, want none", stepsByID[WorkflowStepIDCommit].CommandSuggestions())
	}
	if got := stepsByID[WorkflowStepIDArchive].CommandSuggestions()[0].Command; got != "specharbor archive <change-id>" {
		t.Fatalf("archive suggestion = %q, want specharbor archive <change-id>", got)
	}
}

func TestWorkflowRoleStepIDsAlignWithPromptRoles(t *testing.T) {
	want := []WorkflowStepID{
		WorkflowStepID(PromptRoleSpecAuthor),
		WorkflowStepID(PromptRoleArchitectureReviewer),
		WorkflowStepID(PromptRoleImplementer),
		WorkflowStepID(PromptRoleTestEngineer),
		WorkflowStepID(PromptRoleChangeReviewer),
	}

	got := recommendedWorkflowStepIDs(t)[:5]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("role workflow ids = %v, want prompt role ids %v", got, want)
	}
}

func TestWorkflowSafetyNotes(t *testing.T) {
	stepsByID := recommendedWorkflowStepsByID(t)
	wantSnippets := map[WorkflowStepID]string{
		WorkflowStepIDCommit:      "does not commit, stage files, modify branches, push, or sign commits",
		WorkflowStepIDPullRequest: "does not create PRs, call source-control APIs",
		WorkflowStepIDMerge:       "does not merge, approve, inspect CI, trigger CI",
		WorkflowStepIDArchive:     "does not archive automatically",
	}

	for id, snippet := range wantSnippets {
		notes := strings.Join(stepsByID[id].SafetyNotes(), "\n")
		if !strings.Contains(notes, snippet) {
			t.Fatalf("%s safety notes = %q, want snippet %q", id, notes, snippet)
		}
	}
}

func TestWorkflowSlicesAreDefensivelyCopied(t *testing.T) {
	requires := []WorkflowStepID{"first"}
	suggestions := []WorkflowCommandSuggestion{{Command: "specharbor example", Description: "Example command."}}
	safetyNotes := []string{"Original note."}

	first := NewWorkflowStep("first", "First", "First step.", 1, WorkflowStepModeManual, true, false, nil, nil, nil)
	second := NewWorkflowStep("second", "Second", "Second step.", 2, WorkflowStepModeManual, true, false, requires, suggestions, safetyNotes)
	workflow, err := NewWorkflow("Workflow", []WorkflowStep{second, first})
	if err != nil {
		t.Fatalf("NewWorkflow() error = %v", err)
	}

	requires[0] = "mutated"
	suggestions[0].Command = "mutated"
	safetyNotes[0] = "mutated"

	secondRequires := second.Requires()
	secondSuggestions := second.CommandSuggestions()
	secondSafetyNotes := second.SafetyNotes()
	if secondRequires[0] != "first" {
		t.Fatalf("step requirement mutated through input slice: %v", secondRequires)
	}
	if secondSuggestions[0].Command != "specharbor example" {
		t.Fatalf("step command suggestion mutated through input slice: %v", secondSuggestions)
	}
	if secondSafetyNotes[0] != "Original note." {
		t.Fatalf("step safety note mutated through input slice: %v", secondSafetyNotes)
	}

	secondRequires[0] = "mutated"
	secondSuggestions[0].Command = "mutated"
	secondSafetyNotes[0] = "mutated"
	if second.Requires()[0] != "first" {
		t.Fatalf("step requirement mutated through accessor slice: %v", second.Requires())
	}
	if second.CommandSuggestions()[0].Command != "specharbor example" {
		t.Fatalf("step command suggestion mutated through accessor slice: %v", second.CommandSuggestions())
	}
	if second.SafetyNotes()[0] != "Original note." {
		t.Fatalf("step safety note mutated through accessor slice: %v", second.SafetyNotes())
	}

	steps := workflow.Steps()
	if steps[0].ID != "first" {
		t.Fatalf("workflow steps were not sorted by order: %v", steps)
	}
	steps[0].ID = "mutated"
	if workflow.Steps()[0].ID != "first" {
		t.Fatalf("workflow step mutated through accessor slice: %v", workflow.Steps())
	}

	nestedRequires := steps[1].Requires()
	nestedRequires[0] = "mutated"
	if workflow.Steps()[1].Requires()[0] != "first" {
		t.Fatalf("workflow nested step requirement mutated through accessor slice: %v", workflow.Steps()[1].Requires())
	}
}

func recommendedWorkflow(t *testing.T) Workflow {
	t.Helper()

	workflow, err := RecommendedWorkflow()
	if err != nil {
		t.Fatalf("RecommendedWorkflow() error = %v", err)
	}
	return workflow
}

func recommendedWorkflowStepIDs(t *testing.T) []WorkflowStepID {
	t.Helper()

	steps := recommendedWorkflow(t).Steps()
	stepIDs := make([]WorkflowStepID, 0, len(steps))
	for _, step := range steps {
		stepIDs = append(stepIDs, step.ID)
	}
	return stepIDs
}

func recommendedWorkflowStepsByID(t *testing.T) map[WorkflowStepID]WorkflowStep {
	t.Helper()

	stepsByID := map[WorkflowStepID]WorkflowStep{}
	for _, step := range recommendedWorkflow(t).Steps() {
		stepsByID[step.ID] = step
	}
	return stepsByID
}
