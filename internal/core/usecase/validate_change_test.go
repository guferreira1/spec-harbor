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
	fileSystem.files[checkedPath+"/design.md"] = true
	fileSystem.files[checkedPath+"/tasks.md"] = true
	fileSystem.files[checkedPath+"/acceptance-criteria.md"] = true

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
	fileSystem.files[checkedPath+"/proposal.md"] = true
	fileSystem.files[checkedPath+"/design.md"] = true
	fileSystem.files[checkedPath+"/tasks.md"] = true
	fileSystem.files[checkedPath+"/acceptance-criteria.md"] = true

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
			want:  "change id must be a single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := NewValidateChange(newFakeValidationFileSystem())

			_, err := useCase.Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
		})
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
		fileSystem.files[checkedPath+"/"+requiredFile] = true
	}
}

type fakeValidationFileSystem struct {
	directories     map[string]bool
	files           map[string]bool
	directoryErrors map[string]error
	fileErrors      map[string]error
}

func newFakeValidationFileSystem() *fakeValidationFileSystem {
	return &fakeValidationFileSystem{
		directories:     make(map[string]bool),
		files:           make(map[string]bool),
		directoryErrors: make(map[string]error),
		fileErrors:      make(map[string]error),
	}
}

func (fileSystem *fakeValidationFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeValidationFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.files[relativePath], nil
}
