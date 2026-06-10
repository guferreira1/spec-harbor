package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestValidateChangeReturnsValidForCompleteChange(t *testing.T) {
	changeID := "implement-validation-foundation"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	seedCompleteChange(fileSystem, changeID)

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ValidationStatusValid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ValidationStatusValid)
	}
	if result.ChangeID != changeID {
		t.Fatalf("ChangeID = %q, want %q", result.ChangeID, changeID)
	}
	if result.CheckedPath != "openspec/changes/"+changeID {
		t.Fatalf("CheckedPath = %q, want openspec/changes/%s", result.CheckedPath, changeID)
	}
	if len(result.RequiredFiles) != len(domain.RequiredOpenSpecChangeFiles()) {
		t.Fatalf("RequiredFiles count = %d, want %d", len(result.RequiredFiles), len(domain.RequiredOpenSpecChangeFiles()))
	}
	if len(result.Findings) != 0 {
		t.Fatalf("Findings count = %d, want 0", len(result.Findings))
	}
}

func TestValidateChangeReturnsInvalidForMissingOpenSpecProjectStructure(t *testing.T) {
	changeID := "implement-validation-foundation"
	fileSystem := newFakeValidationFileSystem()
	useCase := NewValidateChange(fileSystem)

	result, err := useCase.Execute(ValidateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertSingleFindingCode(t, result, domain.ValidationFindingCodeProjectRootUnavailable)
	if result.Status != domain.ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ValidationStatusInvalid)
	}
}

func TestValidateChangeReturnsInvalidForMissingChangeDirectory(t *testing.T) {
	changeID := "missing-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	useCase := NewValidateChange(fileSystem)

	result, err := useCase.Execute(ValidateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertSingleFindingCode(t, result, domain.ValidationFindingCodeChangeDirectoryMissing)
	if result.Findings[0].RelativePath != "openspec/changes/"+changeID {
		t.Fatalf("RelativePath = %q, want openspec/changes/%s", result.Findings[0].RelativePath, changeID)
	}
}

func TestValidateChangeReturnsMissingRequiredFileFindings(t *testing.T) {
	changeID := "missing-files"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	checkedPath := "openspec/changes/" + changeID
	fileSystem.directories[checkedPath] = true
	fileSystem.setFile(checkedPath+"/design.md", authoredChangeFileContent("design.md"))
	fileSystem.setFile(checkedPath+"/tasks.md", authoredChangeFileContent("tasks.md"))
	fileSystem.setFile(checkedPath+"/acceptance-criteria.md", authoredChangeFileContent("acceptance-criteria.md"))

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ValidationStatusInvalid)
	}
	if got := findingSubjectsByCode(result, domain.ValidationFindingCodeRequiredFileMissing); strings.Join(got, ",") != "proposal.md,risks.md" {
		t.Fatalf("missing file findings = %v, want [proposal.md risks.md]", got)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("Findings count = %d, want 2", len(result.Findings))
	}
}

func TestValidateChangeDoesNotReportExistingRequiredFilesAsMissing(t *testing.T) {
	changeID := "partial-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	checkedPath := "openspec/changes/" + changeID
	fileSystem.directories[checkedPath] = true
	fileSystem.setFile(checkedPath+"/proposal.md", authoredChangeFileContent("proposal.md"))
	fileSystem.setFile(checkedPath+"/design.md", authoredChangeFileContent("design.md"))
	fileSystem.setFile(checkedPath+"/tasks.md", authoredChangeFileContent("tasks.md"))
	fileSystem.setFile(checkedPath+"/acceptance-criteria.md", authoredChangeFileContent("acceptance-criteria.md"))

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	missingSubjects := findingSubjectsByCode(result, domain.ValidationFindingCodeRequiredFileMissing)
	if len(missingSubjects) != 1 || missingSubjects[0] != "risks.md" {
		t.Fatalf("missing file findings = %v, want [risks.md]", missingSubjects)
	}
}

func TestValidateChangeRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input ValidateChangeInput
		want  string
	}{
		{
			name:  "empty project root",
			input: ValidateChangeInput{ProjectRoot: " ", ChangeID: "change"},
			want:  "project root is required",
		},
		{
			name:  "empty change id",
			input: ValidateChangeInput{ProjectRoot: "/project", ChangeID: " "},
			want:  "change id is required",
		},
		{
			name:  "absolute change path",
			input: ValidateChangeInput{ProjectRoot: "/project", ChangeID: "/outside"},
			want:  "change id must be a single path segment",
		},
		{
			name:  "traversal change path",
			input: ValidateChangeInput{ProjectRoot: "/project", ChangeID: "../outside"},
			want:  "change id must be a single path segment",
		},
		{
			name:  "traversal segment",
			input: ValidateChangeInput{ProjectRoot: "/project", ChangeID: ".."},
			want:  "change id must not contain '.' or '..' path sequences",
		},
		{
			name:  "leading dash",
			input: ValidateChangeInput{ProjectRoot: "/project", ChangeID: "-change"},
			want:  "change id must not start with '-'",
		},
		{
			name:  "unsafe character",
			input: ValidateChangeInput{ProjectRoot: "/project", ChangeID: "change id"},
			want:  "change id contains unsupported character",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeValidationFileSystem()
			useCase := NewValidateChange(fileSystem)

			_, err := useCase.Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.calls != 0 {
				t.Fatalf("filesystem calls = %d, want 0 before change id validation passes", fileSystem.calls)
			}
		})
	}
}

func TestValidateChangeReturnsFileEmptyAndSuppressesOtherContentFindings(t *testing.T) {
	changeID := "empty-file-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	seedCompleteChange(fileSystem, changeID)
	checkedPath := "openspec/changes/" + changeID
	fileSystem.setFile(checkedPath+"/design.md", "  \n\n")

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{ProjectRoot: "/project", ChangeID: changeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ValidationStatusInvalid)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("Findings = %v, want only the file_empty finding", result.Findings)
	}
	finding := result.Findings[0]
	if finding.Code != domain.ValidationFindingCodeFileEmpty {
		t.Fatalf("finding code = %q, want %q", finding.Code, domain.ValidationFindingCodeFileEmpty)
	}
	if finding.RelativePath != checkedPath+"/design.md" {
		t.Fatalf("RelativePath = %q, want %s/design.md", finding.RelativePath, checkedPath)
	}
}

func TestValidateChangeReportsBlankStarterChangeAsValidWithBoilerplateWarnings(t *testing.T) {
	changeID := "fresh-blank-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	checkedPath := "openspec/changes/" + changeID
	fileSystem.directories[checkedPath] = true
	for fileName, content := range blankStarterChangeContents() {
		fileSystem.setFile(checkedPath+"/"+fileName, content)
	}

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{ProjectRoot: "/project", ChangeID: changeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ValidationStatusValid {
		t.Fatalf("Status = %q, want %q (findings: %v)", result.Status, domain.ValidationStatusValid, result.Findings)
	}
	if result.ErrorCount() != 0 {
		t.Fatalf("ErrorCount() = %d, want 0 (findings: %v)", result.ErrorCount(), result.Findings)
	}
	boilerplateFindings := 0
	for _, finding := range result.Findings {
		if finding.Code == domain.ValidationFindingCodeBoilerplateOnlyContent {
			boilerplateFindings++
			continue
		}
		if finding.Code != domain.ValidationFindingCodePlaceholderContent {
			t.Fatalf("unexpected finding %v for blank starter change", finding)
		}
	}
	if boilerplateFindings != len(domain.RequiredOpenSpecChangeFiles()) {
		t.Fatalf("boilerplate findings = %d, want %d", boilerplateFindings, len(domain.RequiredOpenSpecChangeFiles()))
	}
}

func TestValidateChangeDoesNotFlagEditedContentAsBoilerplate(t *testing.T) {
	changeID := "edited-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	seedCompleteChange(fileSystem, changeID)
	checkedPath := "openspec/changes/" + changeID
	starter := blankStarterChangeContents()["proposal.md"]
	fileSystem.setFile(checkedPath+"/proposal.md", starter+"\nThe validation pipeline must report severities per finding.\n")

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{ProjectRoot: "/project", ChangeID: changeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, finding := range result.Findings {
		if finding.Code == domain.ValidationFindingCodeBoilerplateOnlyContent && finding.Subject == "proposal.md" {
			t.Fatalf("edited proposal.md flagged as boilerplate-only: %v", result.Findings)
		}
	}
}

func TestValidateChangeReportsMalformedAndMissingCheckboxes(t *testing.T) {
	changeID := "broken-tasks-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	seedCompleteChange(fileSystem, changeID)
	checkedPath := "openspec/changes/" + changeID
	fileSystem.setFile(checkedPath+"/tasks.md", "# Tasks\n\n## Phase 1\n\n- [] broken checkbox\n- [y] wrong mark\n")

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{ProjectRoot: "/project", ChangeID: changeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ValidationStatusInvalid {
		t.Fatalf("Status = %q, want %q", result.Status, domain.ValidationStatusInvalid)
	}
	var malformedMessages []string
	missingFound := false
	for _, finding := range result.Findings {
		switch finding.Code {
		case domain.ValidationFindingCodeTasksCheckboxMalformed:
			malformedMessages = append(malformedMessages, finding.Message)
		case domain.ValidationFindingCodeTasksCheckboxMissing:
			missingFound = true
		}
	}
	if len(malformedMessages) != 2 {
		t.Fatalf("malformed findings = %v, want 2", malformedMessages)
	}
	if !strings.Contains(malformedMessages[0], "(line 5)") || !strings.Contains(malformedMessages[1], "(line 6)") {
		t.Fatalf("malformed messages = %v, want line numbers 5 and 6", malformedMessages)
	}
	if !missingFound {
		t.Fatalf("findings = %v, want tasks_checkbox_missing", result.Findings)
	}
}

func TestValidateChangeKeepsWarningsOnlyResultValid(t *testing.T) {
	changeID := "warning-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	seedCompleteChange(fileSystem, changeID)
	checkedPath := "openspec/changes/" + changeID
	fileSystem.setFile(checkedPath+"/risks.md", "# Risks\n\n## Risks\n\n- Strict rules could reject existing changes.\n")

	useCase := NewValidateChange(fileSystem)
	result, err := useCase.Execute(ValidateChangeInput{ProjectRoot: "/project", ChangeID: changeID})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Status != domain.ValidationStatusValid {
		t.Fatalf("Status = %q, want %q (findings: %v)", result.Status, domain.ValidationStatusValid, result.Findings)
	}
	if result.ErrorCount() != 0 || result.WarningCount() == 0 {
		t.Fatalf("ErrorCount() = %d, WarningCount() = %d, want 0 errors and at least one warning", result.ErrorCount(), result.WarningCount())
	}
	if result.Findings[0].Code != domain.ValidationFindingCodeRisksMitigationMissing {
		t.Fatalf("finding code = %q, want %q", result.Findings[0].Code, domain.ValidationFindingCodeRisksMitigationMissing)
	}
}

func TestValidateChangeReturnsReadErrorsAsGoErrors(t *testing.T) {
	changeID := "read-error-change"
	fileSystem := newFakeValidationFileSystem()
	seedOpenSpecProject(fileSystem)
	seedCompleteChange(fileSystem, changeID)
	wantErr := errors.New("read failure")
	fileSystem.readErrors["openspec/changes/"+changeID+"/design.md"] = wantErr

	useCase := NewValidateChange(fileSystem)
	_, err := useCase.Execute(ValidateChangeInput{ProjectRoot: "/project", ChangeID: changeID})
	if err == nil {
		t.Fatalf("Execute() error = nil, want read error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
}

// blankStarterChangeContents replicates blank starter content as in-test
// fixtures; it must not be imported from adapter template packages.
func blankStarterChangeContents() map[string]string {
	return map[string]string{
		"proposal.md": `# Proposal

## Problem

Describe the problem this change should solve and who is affected.

## Goal

Describe the outcome this change should deliver.

## Scope

- List the behavior, files, or interfaces included in this change.

## Out of Scope

- List related work that should not be implemented in this change.

## Success Criteria

- Describe how reviewers can tell the change is complete.
`,
		"design.md": `# Design

## Overview

Describe the proposed approach at a high level.

## Architecture

Describe the affected layers, boundaries, and dependencies.

## Technical Decisions

- Record important implementation choices and tradeoffs.

## Testing Strategy

- Describe the tests needed to cover the change.

## Validation

- List the commands or checks that should pass before completion.
`,
		"tasks.md": `# Tasks

## Implementation

- [ ] Read the project context, architecture spec, and active OpenSpec change.
- [ ] Keep implementation limited to the approved scope.
- [ ] Add or update domain, use case, port, adapter, CLI, and test code as needed.
- [ ] Run required formatting and verification commands.
- [ ] Update this task list only after implementation work is complete.
`,
		"acceptance-criteria.md": `# Acceptance Criteria

- The requested behavior is implemented within the approved scope.
- Existing behavior outside the scope remains unchanged.
- Errors are clear and actionable for users.
- Automated tests cover the important success and failure paths.
- Required verification commands pass.
`,
		"risks.md": `# Risks

## Risks

- Identify technical, product, security, or delivery risks introduced by this change.

## Mitigations

- Describe how each risk will be reduced, tested, or monitored.
`,
	}
}

func TestValidateChangeReturnsFilesystemErrors(t *testing.T) {
	wantErr := errors.New("filesystem unavailable")
	tests := []struct {
		name  string
		setup func(fileSystem *fakeValidationFileSystem)
	}{
		{
			name: "project file check",
			setup: func(fileSystem *fakeValidationFileSystem) {
				fileSystem.fileErrors[openspecProjectFile] = wantErr
			},
		},
		{
			name: "changes directory check",
			setup: func(fileSystem *fakeValidationFileSystem) {
				fileSystem.files[openspecProjectFile] = true
				fileSystem.directoryErrors[openspecChangesDirectory] = wantErr
			},
		},
		{
			name: "change directory check",
			setup: func(fileSystem *fakeValidationFileSystem) {
				seedOpenSpecProject(fileSystem)
				fileSystem.directoryErrors[openspecChangesDirectory+"/change"] = wantErr
			},
		},
		{
			name: "required file check",
			setup: func(fileSystem *fakeValidationFileSystem) {
				seedOpenSpecProject(fileSystem)
				fileSystem.directories[openspecChangesDirectory+"/change"] = true
				fileSystem.fileErrors[openspecChangesDirectory+"/change/proposal.md"] = wantErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeValidationFileSystem()
			test.setup(fileSystem)
			useCase := NewValidateChange(fileSystem)

			_, err := useCase.Execute(ValidateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want filesystem error")
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
			}
		})
	}
}

func assertSingleFindingCode(t *testing.T, result domain.ValidationResult, code domain.ValidationFindingCode) {
	t.Helper()

	if len(result.Findings) != 1 {
		t.Fatalf("Findings count = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].Code != code {
		t.Fatalf("finding code = %q, want %q", result.Findings[0].Code, code)
	}
}

func findingSubjectsByCode(result domain.ValidationResult, code domain.ValidationFindingCode) []string {
	var subjects []string
	for _, finding := range result.Findings {
		if finding.Code == code {
			subjects = append(subjects, finding.Subject)
		}
	}
	return subjects
}

func seedOpenSpecProject(fileSystem *fakeValidationFileSystem) {
	fileSystem.files[openspecProjectFile] = true
	fileSystem.directories[openspecChangesDirectory] = true
}

func seedCompleteChange(fileSystem *fakeValidationFileSystem, changeID string) {
	checkedPath := openspecChangesDirectory + "/" + changeID
	fileSystem.directories[checkedPath] = true
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.setFile(checkedPath+"/"+requiredFile, authoredChangeFileContent(requiredFile))
	}
}

func authoredChangeFileContent(fileName string) string {
	switch fileName {
	case "proposal.md":
		return `# Proposal: Example

## Problem

Users cannot tell whether a change package is ready for implementation.

## Goal

Validation reports content-quality findings with clear severities.
`
	case "design.md":
		return `# Design: Example

## Overview

The validation pipeline reads files once and evaluates deterministic rules.

## Architecture

Rules live in the domain layer and the use case orchestrates them.
`
	case "tasks.md":
		return `# Tasks

## Phase 1 - Implementation

- [x] Read the architecture documentation.
- [ ] Implement the validation rules.
`
	case "acceptance-criteria.md":
		return `# Acceptance Criteria

- Validation reports findings with severity and code.
- Warnings alone keep the exit code at zero.
`
	case "risks.md":
		return `# Risks

## Risks

- Strict rules could reject existing changes.

## Mitigations

- Quality findings stay warnings and are covered by tests.
`
	default:
		return "# " + fileName + "\n\nAuthored content.\n"
	}
}

type fakeValidationFileSystem struct {
	directories     map[string]bool
	files           map[string]bool
	contents        map[string]string
	directoryErrors map[string]error
	fileErrors      map[string]error
	readErrors      map[string]error
	calls           int
}

func newFakeValidationFileSystem() *fakeValidationFileSystem {
	return &fakeValidationFileSystem{
		directories:     make(map[string]bool),
		files:           make(map[string]bool),
		contents:        make(map[string]string),
		directoryErrors: make(map[string]error),
		fileErrors:      make(map[string]error),
		readErrors:      make(map[string]error),
	}
}

func (fileSystem *fakeValidationFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.calls++
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeValidationFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.calls++
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeValidationFileSystem) ReadFile(_ string, relativePath string) (string, error) {
	fileSystem.calls++
	if err := fileSystem.readErrors[relativePath]; err != nil {
		return "", err
	}
	return fileSystem.contents[relativePath], nil
}

func (fileSystem *fakeValidationFileSystem) setFile(relativePath string, content string) {
	fileSystem.files[relativePath] = true
	fileSystem.contents[relativePath] = content
}
