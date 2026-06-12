package usecase

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestCreateProjectBriefWritesNewBrief(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	useCase := NewCreateProjectBrief(fileSystem)

	result, err := useCase.Execute(CreateProjectBriefInput{
		ProjectRoot: "/project",
		Answers:     validProjectBriefAnswers(t),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.TargetPath != projectBriefRelativePath {
		t.Fatalf("TargetPath = %q, want %q", result.TargetPath, projectBriefRelativePath)
	}
	if !result.DirectoryCreated {
		t.Fatalf("DirectoryCreated = false, want true")
	}
	if !result.FileWritten {
		t.Fatalf("FileWritten = false, want true")
	}
	if !fileSystem.directories[projectBriefDirectory] {
		t.Fatalf("%s directory was not created", projectBriefDirectory)
	}
	contents := fileSystem.files[projectBriefRelativePath]
	for _, want := range []string{
		"# Project Brief",
		"## Project type",
		"Answer: CLI/tooling project",
		"### Test",
		"Answer: go test ./...",
		"### Detected context",
		"None recorded.",
		"## Assumptions",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("project brief = %q, want %q", contents, want)
		}
	}
}

func TestCreateProjectBriefUsesExistingDirectory(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	fileSystem.directories[projectBriefDirectory] = true

	result, err := NewCreateProjectBrief(fileSystem).Execute(CreateProjectBriefInput{
		ProjectRoot: "/project",
		Answers:     validProjectBriefAnswers(t),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.DirectoryCreated {
		t.Fatalf("DirectoryCreated = true, want false")
	}
	if len(fileSystem.createdDirectories) != 0 {
		t.Fatalf("created directories = %v, want none", fileSystem.createdDirectories)
	}
}

func TestCreateProjectBriefRefusesExistingBrief(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	fileSystem.directories[projectBriefDirectory] = true
	fileSystem.files[projectBriefRelativePath] = "existing brief"

	_, err := NewCreateProjectBrief(fileSystem).Execute(CreateProjectBriefInput{
		ProjectRoot: "/project",
		Answers:     validProjectBriefAnswers(t),
	})
	if err == nil || !strings.Contains(err.Error(), "project brief already exists at .specharbor/project-brief.md") {
		t.Fatalf("Execute() error = %v, want existing brief refusal", err)
	}
	if fileSystem.files[projectBriefRelativePath] != "existing brief" {
		t.Fatalf("existing brief was overwritten: %q", fileSystem.files[projectBriefRelativePath])
	}
	if len(fileSystem.writtenFiles) != 0 {
		t.Fatalf("written files = %v, want none", fileSystem.writtenFiles)
	}
}

func TestCreateProjectBriefValidationFailureWritesNothing(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	answers := validProjectBriefAnswers(t)
	answers.Stack = domain.ProjectBriefAnswer{Value: "Go", Source: domain.ProjectBriefAnswerSourceDetectedContext}

	_, err := NewCreateProjectBrief(fileSystem).Execute(CreateProjectBriefInput{
		ProjectRoot: "/project",
		Answers:     answers,
	})
	if err == nil || err.Error() != "stack must be a user-provided answer" {
		t.Fatalf("Execute() error = %v, want source validation error", err)
	}
	if len(fileSystem.createdDirectories) != 0 {
		t.Fatalf("created directories = %v, want none", fileSystem.createdDirectories)
	}
	if len(fileSystem.writtenFiles) != 0 {
		t.Fatalf("written files = %v, want none", fileSystem.writtenFiles)
	}
}

func TestCreateProjectBriefRejectsInvalidDependenciesAndInput(t *testing.T) {
	_, err := (*CreateProjectBrief)(nil).Execute(CreateProjectBriefInput{
		ProjectRoot: "/project",
		Answers:     validProjectBriefAnswers(t),
	})
	if err == nil || err.Error() != "create project brief use case is required" {
		t.Fatalf("nil use case error = %v, want required error", err)
	}

	_, err = NewCreateProjectBrief(nil).Execute(CreateProjectBriefInput{
		ProjectRoot: "/project",
		Answers:     validProjectBriefAnswers(t),
	})
	if err == nil || err.Error() != "project brief filesystem is required" {
		t.Fatalf("nil filesystem error = %v, want required error", err)
	}

	_, err = NewCreateProjectBrief(newFakeGenerationFileSystem()).Execute(CreateProjectBriefInput{
		ProjectRoot: " ",
		Answers:     validProjectBriefAnswers(t),
	})
	if err == nil || err.Error() != "project root is required" {
		t.Fatalf("empty root error = %v, want root required", err)
	}
}

func validProjectBriefAnswers(t *testing.T) domain.ProjectBriefAnswers {
	t.Helper()

	answer := func(value string) domain.ProjectBriefAnswer {
		t.Helper()
		briefAnswer, err := domain.NewUserProvidedProjectBriefAnswer(value)
		if err != nil {
			t.Fatalf("NewUserProvidedProjectBriefAnswer(%q) error = %v", value, err)
		}
		return briefAnswer
	}

	return domain.ProjectBriefAnswers{
		ProjectType:  answer("CLI/tooling project"),
		Purpose:      answer("Developer productivity tool"),
		TargetUsers:  answer("Developers or platform engineers"),
		Stack:        answer("Go"),
		Architecture: answer("Clean Architecture / Hexagonal"),
		Commands: domain.ProjectBriefCommandSet{
			Install: answer("go mod download"),
			Test:    answer("go test ./..."),
			Build:   answer("go build ./..."),
			Run:     answer("go run ./cmd/specharbor"),
		},
		AgentBehavior: answer("Ask before assuming"),
	}
}
