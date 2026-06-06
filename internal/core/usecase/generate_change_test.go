package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateChangeCreatesNewBlankChange(t *testing.T) {
	changeID := "implement-generation-foundation"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	content := newFakeBlankChangeContent()
	useCase := NewGenerateChange(fileSystem, content)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
		Mode:        domain.BlankMode,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	if result.ChangeID != changeID {
		t.Fatalf("ChangeID = %q, want %q", result.ChangeID, changeID)
	}
	if result.Mode != domain.BlankMode {
		t.Fatalf("Mode = %q, want %q", result.Mode, domain.BlankMode)
	}
	if result.ChangePath != changePath {
		t.Fatalf("ChangePath = %q, want %q", result.ChangePath, changePath)
	}
	if !result.ChangeDirectoryCreated {
		t.Fatalf("ChangeDirectoryCreated = false, want true")
	}
	assertStringSlicesEqual(t, result.CreatedFiles(), domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), nil)

	if !fileSystem.directories[changePath] {
		t.Fatalf("change directory %q was not created", changePath)
	}
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		if fileSystem.files[path] != defaultBlankContent(requiredFile) {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], defaultBlankContent(requiredFile))
		}
	}
	assertStringSlicesEqual(t, content.requests, domain.RequiredOpenSpecChangeFiles())
}

func TestGenerateChangeCreatesTargetDirectoryWhenMissing(t *testing.T) {
	changeID := "missing-directory"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	useCase := NewGenerateChange(fileSystem, newFakeBlankChangeContent())

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
		Mode:        domain.BlankMode,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	if !result.ChangeDirectoryCreated {
		t.Fatalf("ChangeDirectoryCreated = false, want true")
	}
	if !containsString(fileSystem.createdDirectories, changePath) {
		t.Fatalf("created directories = %v, want %q", fileSystem.createdDirectories, changePath)
	}
}

func TestGenerateChangeFillsMissingFilesWhenTargetDirectoryExists(t *testing.T) {
	changeID := "partial-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "custom proposal"
	fileSystem.files[changePath+"/tasks.md"] = "custom tasks"

	result, err := NewGenerateChange(fileSystem, newFakeBlankChangeContent()).Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
		Mode:        domain.BlankMode,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ChangeDirectoryCreated {
		t.Fatalf("ChangeDirectoryCreated = true, want false")
	}
	assertStringSlicesEqual(t, result.CreatedFiles(), []string{"design.md", "acceptance-criteria.md", "risks.md"})
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), []string{"proposal.md", "tasks.md"})
	if fileSystem.files[changePath+"/proposal.md"] != "custom proposal" {
		t.Fatalf("existing proposal.md was overwritten")
	}
	if fileSystem.files[changePath+"/tasks.md"] != "custom tasks" {
		t.Fatalf("existing tasks.md was overwritten")
	}
}

func TestGenerateChangeSkipsExistingFilesAndDoesNotOverwrite(t *testing.T) {
	changeID := "existing-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[changePath+"/"+requiredFile] = "custom:" + requiredFile
	}

	result, err := NewGenerateChange(fileSystem, newFakeBlankChangeContent()).Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
		Mode:        domain.BlankMode,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertStringSlicesEqual(t, result.CreatedFiles(), nil)
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), domain.RequiredOpenSpecChangeFiles())
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		if fileSystem.files[path] != "custom:"+requiredFile {
			t.Fatalf("file %q content = %q, want preserved custom content", path, fileSystem.files[path])
		}
	}
}

func TestGenerateChangeRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input GenerateChangeInput
		want  string
	}{
		{
			name:  "empty project root",
			input: GenerateChangeInput{ProjectRoot: " ", ChangeID: "change", Mode: domain.BlankMode},
			want:  "project root is required",
		},
		{
			name:  "empty change id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: " ", Mode: domain.BlankMode},
			want:  "change id is required",
		},
		{
			name:  "unsupported guided mode",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "change", Mode: domain.GuidedMode},
			want:  "unsupported generation mode: guided",
		},
		{
			name:  "unsupported template mode",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "change", Mode: domain.TemplateMode},
			want:  "unsupported generation mode: template",
		},
		{
			name:  "unsupported ai-assisted mode",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "change", Mode: domain.AIAssistedMode},
			want:  "unsupported generation mode: ai-assisted",
		},
		{
			name:  "unsupported agent-assisted mode",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "change", Mode: domain.GenerationMode("agent-assisted")},
			want:  "unsupported generation mode: agent-assisted",
		},
		{
			name:  "unsupported hybrid mode",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "change", Mode: domain.HybridMode},
			want:  "unsupported generation mode: hybrid",
		},
		{
			name:  "dot id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: ".", Mode: domain.BlankMode},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "traversal id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "../outside", Mode: domain.BlankMode},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "absolute id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "/outside", Mode: domain.BlankMode},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "backslash id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: `bad\id`, Mode: domain.BlankMode},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "colon id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "C:bad", Mode: domain.BlankMode},
			want:  "change id must be a safe single path segment",
		},
		{
			name:  "leading dash id",
			input: GenerateChangeInput{ProjectRoot: "/project", ChangeID: "-bad", Mode: domain.BlankMode},
			want:  "change id must be a safe single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			_, err := NewGenerateChange(fileSystem, newFakeBlankChangeContent()).Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("filesystem operations = %d, want 0 before rejecting invalid input", fileSystem.operationCount())
			}
		})
	}
}

func TestGenerateChangeRejectsMissingOpenSpecProjectBeforeTargetWrites(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeGenerationFileSystem)
	}{
		{
			name:  "missing project file",
			setup: func(fileSystem *fakeGenerationFileSystem) { fileSystem.directories[openspecChangesDirectory] = true },
		},
		{
			name:  "missing changes directory",
			setup: func(fileSystem *fakeGenerationFileSystem) { fileSystem.files[openspecProjectFile] = "project" },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			content := newFakeBlankChangeContent()
			test.setup(fileSystem)

			_, err := NewGenerateChange(fileSystem, content).Execute(GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.BlankMode,
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want missing project structure error")
			}
			for _, want := range []string{"OpenSpec project structure is missing", "specharbor init"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("Execute() error = %q, want to contain %q", err.Error(), want)
				}
			}
			if len(fileSystem.createdDirectories) != 0 {
				t.Fatalf("created directories = %v, want none", fileSystem.createdDirectories)
			}
			if len(fileSystem.writtenFiles) != 0 {
				t.Fatalf("written files = %v, want none", fileSystem.writtenFiles)
			}
			if len(content.requests) != 0 {
				t.Fatalf("content requests = %v, want none before project structure is available", content.requests)
			}
			if fileSystem.directories["openspec"] {
				t.Fatalf("generation created openspec directory")
			}
			if fileSystem.files[openspecProjectFile] != "" && test.name == "missing project file" {
				t.Fatalf("generation created openspec/project.md")
			}
			if fileSystem.directories[openspecChangesDirectory] && test.name == "missing changes directory" {
				t.Fatalf("generation created openspec/changes")
			}
		})
	}
}

func TestGenerateChangeReturnsFilesystemErrors(t *testing.T) {
	wantErr := errors.New("filesystem unavailable")
	tests := []struct {
		name  string
		setup func(fileSystem *fakeGenerationFileSystem)
	}{
		{
			name: "project file check",
			setup: func(fileSystem *fakeGenerationFileSystem) {
				fileSystem.fileErrors[openspecProjectFile] = wantErr
			},
		},
		{
			name: "changes directory check",
			setup: func(fileSystem *fakeGenerationFileSystem) {
				fileSystem.files[openspecProjectFile] = "project"
				fileSystem.directoryErrors[openspecChangesDirectory] = wantErr
			},
		},
		{
			name: "target directory check",
			setup: func(fileSystem *fakeGenerationFileSystem) {
				seedGenerationOpenSpecProject(fileSystem)
				fileSystem.directoryErrors[openspecChangesDirectory+"/change"] = wantErr
			},
		},
		{
			name: "target directory creation",
			setup: func(fileSystem *fakeGenerationFileSystem) {
				seedGenerationOpenSpecProject(fileSystem)
				fileSystem.createDirectoryErrors[openspecChangesDirectory+"/change"] = wantErr
			},
		},
		{
			name: "file write",
			setup: func(fileSystem *fakeGenerationFileSystem) {
				seedGenerationOpenSpecProject(fileSystem)
				fileSystem.writeErrors[openspecChangesDirectory+"/change/proposal.md"] = wantErr
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			test.setup(fileSystem)

			_, err := NewGenerateChange(fileSystem, newFakeBlankChangeContent()).Execute(GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.BlankMode,
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

func TestGenerateChangeReturnsContentLoadingErrors(t *testing.T) {
	wantErr := errors.New("content unavailable")
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	content := newFakeBlankChangeContent()
	content.errors["proposal.md"] = wantErr

	_, err := NewGenerateChange(fileSystem, content).Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "change",
		Mode:        domain.BlankMode,
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want content error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
}

func seedGenerationOpenSpecProject(fileSystem *fakeGenerationFileSystem) {
	fileSystem.files[openspecProjectFile] = "project"
	fileSystem.directories[openspecChangesDirectory] = true
}

type fakeGenerationFileSystem struct {
	directories           map[string]bool
	files                 map[string]string
	directoryErrors       map[string]error
	fileErrors            map[string]error
	createDirectoryErrors map[string]error
	writeErrors           map[string]error
	createdDirectories    []string
	writtenFiles          []string
	checkedDirectories    []string
	checkedFiles          []string
}

func newFakeGenerationFileSystem() *fakeGenerationFileSystem {
	return &fakeGenerationFileSystem{
		directories:           make(map[string]bool),
		files:                 make(map[string]string),
		directoryErrors:       make(map[string]error),
		fileErrors:            make(map[string]error),
		createDirectoryErrors: make(map[string]error),
		writeErrors:           make(map[string]error),
	}
}

func (fileSystem *fakeGenerationFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedDirectories = append(fileSystem.checkedDirectories, relativePath)
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeGenerationFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedFiles = append(fileSystem.checkedFiles, relativePath)
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeGenerationFileSystem) CreateDirectory(_ string, relativePath string) error {
	fileSystem.createdDirectories = append(fileSystem.createdDirectories, relativePath)
	if err := fileSystem.createDirectoryErrors[relativePath]; err != nil {
		return err
	}
	fileSystem.directories[relativePath] = true
	return nil
}

func (fileSystem *fakeGenerationFileSystem) WriteFileIfAbsent(_ string, relativePath string, contents string) (bool, error) {
	fileSystem.writtenFiles = append(fileSystem.writtenFiles, relativePath)
	if err := fileSystem.writeErrors[relativePath]; err != nil {
		return false, err
	}
	if _, exists := fileSystem.files[relativePath]; exists {
		return false, nil
	}
	fileSystem.files[relativePath] = contents
	return true, nil
}

func (fileSystem *fakeGenerationFileSystem) operationCount() int {
	return len(fileSystem.checkedDirectories) +
		len(fileSystem.checkedFiles) +
		len(fileSystem.createdDirectories) +
		len(fileSystem.writtenFiles)
}

type fakeBlankChangeContent struct {
	errors   map[string]error
	requests []string
}

func newFakeBlankChangeContent() *fakeBlankChangeContent {
	return &fakeBlankChangeContent{errors: make(map[string]error)}
}

func (content *fakeBlankChangeContent) ContentFor(relativePath string) (string, error) {
	content.requests = append(content.requests, relativePath)
	if err := content.errors[relativePath]; err != nil {
		return "", err
	}
	return defaultBlankContent(relativePath), nil
}

func defaultBlankContent(relativePath string) string {
	return "blank:" + relativePath
}

func assertStringSlicesEqual(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice = %v, want %v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("slice = %v, want %v", got, want)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
