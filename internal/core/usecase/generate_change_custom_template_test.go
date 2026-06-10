package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateChangeCreatesNewCustomTemplateChange(t *testing.T) {
	changeID := "add-payment-flow"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           changeID,
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
		Title:              "Add payments",
		Summary:            "Adds a payment flow.",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	if result.Mode != domain.TemplateMode {
		t.Fatalf("Mode = %q, want %q", result.Mode, domain.TemplateMode)
	}
	if result.TemplateSource != domain.CustomTemplateSource {
		t.Fatalf("TemplateSource = %q, want %q", result.TemplateSource, domain.CustomTemplateSource)
	}
	if result.CustomTemplateName != "api-feature" {
		t.Fatalf("CustomTemplateName = %q, want api-feature", result.CustomTemplateName)
	}
	if result.TemplatePath != customTemplatesDirectory+"/api-feature" {
		t.Fatalf("TemplatePath = %q, want %q", result.TemplatePath, customTemplatesDirectory+"/api-feature")
	}
	if result.ChangePath != changePath {
		t.Fatalf("ChangePath = %q, want %q", result.ChangePath, changePath)
	}
	if !result.ChangeDirectoryCreated {
		t.Fatalf("ChangeDirectoryCreated = false, want true")
	}
	assertStringSlicesEqual(t, result.CreatedFiles(), domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), nil)

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		want := "custom:" + requiredFile + ":" + changeID + ":Add payments:Adds a payment flow."
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
		}
	}
}

func TestGenerateChangeCustomTemplateLeavesUnresolvedTokensVerbatim(t *testing.T) {
	changeID := "add-payment-flow"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           changeID,
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		want := "custom:" + requiredFile + ":" + changeID + ":{{title}}:{{summary}}"
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
		}
	}
}

func TestGenerateChangeRejectsInvalidCustomTemplateNamesBeforeFilesystemAccess(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		want         string
	}{
		{name: "empty", templateName: "", want: "custom template name is required"},
		{name: "separator", templateName: "nested/template", want: "custom template name must be a single path segment"},
		{name: "backslash", templateName: "nested\\template", want: "custom template name must be a single path segment"},
		{name: "absolute path", templateName: "/etc/passwd", want: "custom template name must be a single path segment"},
		{name: "traversal", templateName: "a..b", want: "custom template name must not contain '.' or '..' path sequences"},
		{name: "leading dot", templateName: ".hidden", want: "custom template name must not start with '.'"},
		{name: "leading dash", templateName: "-flag", want: "custom template name must not start with '-'"},
		{name: "unsupported character", templateName: "api template", want: "custom template name contains unsupported character ' '"},
		{name: "over-length", templateName: strings.Repeat("a", 129), want: "custom template name must be at most 128 characters"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			customTemplateFiles := newFakeCustomTemplateFileSystem()
			useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

			_, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot:        "/project",
				ChangeID:           "valid-change",
				Mode:               domain.TemplateMode,
				TemplateSource:     domain.CustomTemplateSource,
				CustomTemplateName: test.templateName,
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if customTemplateFiles.operationCount() != 0 {
				t.Fatalf("custom template filesystem operations = %d, want 0", customTemplateFiles.operationCount())
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("generation filesystem operations = %d, want 0", fileSystem.operationCount())
			}
		})
	}
}

func TestGenerateChangeCustomTemplateUnknownDirectoryError(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "missing-template",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want unknown custom template error")
	}
	want := "unknown custom template: missing-template. Expected directory: .specharbor/templates/missing-template"
	if err.Error() != want {
		t.Fatalf("Execute() error = %q, want %q", err.Error(), want)
	}
	assertNoGenerationWrites(t, fileSystem)
}

func TestGenerateChangeCustomTemplateMissingFilesReturnAggregatedError(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	delete(customTemplateFiles.files, customTemplatesDirectory+"/api-feature/design.md")
	delete(customTemplateFiles.files, customTemplatesDirectory+"/api-feature/risks.md")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want missing files error")
	}
	want := "custom template api-feature is missing required files: design.md, risks.md"
	if err.Error() != want {
		t.Fatalf("Execute() error = %q, want %q", err.Error(), want)
	}
	assertNoGenerationWrites(t, fileSystem)
}

func TestGenerateChangeCustomTemplateEmptyFileError(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	customTemplateFiles.files[customTemplatesDirectory+"/api-feature/design.md"] = "  \n\t "
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want empty file error")
	}
	want := "custom template file api-feature/design.md is empty"
	if err.Error() != want {
		t.Fatalf("Execute() error = %q, want %q", err.Error(), want)
	}
	assertNoGenerationWrites(t, fileSystem)
}

func TestGenerateChangeCustomTemplateReadsOnlyKnownFiles(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	customTemplateFiles.files[customTemplatesDirectory+"/api-feature/README.md"] = "notes"
	customTemplateFiles.files[customTemplatesDirectory+"/api-feature/setup.sh"] = "#!/bin/sh"
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var wantReads []string
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		wantReads = append(wantReads, customTemplatesDirectory+"/api-feature/"+requiredFile)
	}
	assertStringSlicesEqual(t, customTemplateFiles.readFiles, wantReads)
	assertStringSlicesEqual(t, customTemplateFiles.checkedFiles, wantReads)
	assertStringSlicesEqual(t, customTemplateFiles.checkedDirectories, []string{customTemplatesDirectory + "/api-feature"})

	changePath := openspecChangesDirectory + "/valid-change"
	if _, exists := fileSystem.files[changePath+"/README.md"]; exists {
		t.Fatalf("extra template file README.md was copied into the change")
	}
	if _, exists := fileSystem.files[changePath+"/setup.sh"]; exists {
		t.Fatalf("extra template file setup.sh was copied into the change")
	}
}

func TestGenerateChangeCustomTemplateSkipsExistingFilesAndDoesNotOverwrite(t *testing.T) {
	changeID := "add-payment-flow"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "existing proposal"
	fileSystem.files[changePath+"/tasks.md"] = "existing tasks"
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           changeID,
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ChangeDirectoryCreated {
		t.Fatalf("ChangeDirectoryCreated = true, want false")
	}
	assertStringSlicesEqual(t, result.CreatedFiles(), []string{"design.md", "acceptance-criteria.md", "risks.md"})
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), []string{"proposal.md", "tasks.md"})

	if fileSystem.files[changePath+"/proposal.md"] != "existing proposal" {
		t.Fatalf("proposal.md content = %q, want preserved existing content", fileSystem.files[changePath+"/proposal.md"])
	}
	if fileSystem.files[changePath+"/tasks.md"] != "existing tasks" {
		t.Fatalf("tasks.md content = %q, want preserved existing content", fileSystem.files[changePath+"/tasks.md"])
	}
}

func TestGenerateChangeCustomTemplateWritesOnlyUnderChangePath(t *testing.T) {
	changeID := "add-payment-flow"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           changeID,
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	for _, writtenFile := range fileSystem.writtenFiles {
		if !strings.HasPrefix(writtenFile, changePath+"/") {
			t.Fatalf("write target %q is outside %q", writtenFile, changePath)
		}
	}
	assertStringSlicesEqual(t, fileSystem.createdDirectories, []string{changePath})
}

func TestGenerateChangeCustomTemplateRejectsMissingOpenSpecProjectBeforeTemplateReads(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want missing OpenSpec project error")
	}
	want := "OpenSpec project structure is missing. Run specharbor init first."
	if err.Error() != want {
		t.Fatalf("Execute() error = %q, want %q", err.Error(), want)
	}
	if customTemplateFiles.operationCount() != 0 {
		t.Fatalf("custom template filesystem operations = %d, want 0", customTemplateFiles.operationCount())
	}
	assertNoGenerationWrites(t, fileSystem)
}

func TestGenerateChangeCustomTemplateReadErrorsAreReported(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	customTemplateFiles.readErrors[customTemplatesDirectory+"/api-feature/design.md"] = errors.New("disk failure")
	useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want read error")
	}
	want := "read file .specharbor/templates/api-feature/design.md: disk failure"
	if err.Error() != want {
		t.Fatalf("Execute() error = %q, want %q", err.Error(), want)
	}
	assertNoGenerationWrites(t, fileSystem)
}

func TestGenerateChangeCustomTemplateRequiresReadPortDependency(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	useCase := NewGenerateChangeWithContent(
		fileSystem,
		newFakeBlankChangeContent(),
		newFakeTemplateChangeContent(),
		newFakeGuidedChangeContent(),
	)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "valid-change",
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "api-feature",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want missing dependency error")
	}
	if err.Error() != "custom template filesystem is required" {
		t.Fatalf("Execute() error = %q, want custom template filesystem is required", err.Error())
	}
	if fileSystem.operationCount() != 0 {
		t.Fatalf("generation filesystem operations = %d, want 0", fileSystem.operationCount())
	}
}

func TestGenerateChangeRejectsUnsupportedTemplateSources(t *testing.T) {
	tests := []struct {
		name  string
		input GenerateChangeInput
		want  string
	}{
		{
			name: "unknown template source",
			input: GenerateChangeInput{
				ProjectRoot:        "/project",
				ChangeID:           "valid-change",
				Mode:               domain.TemplateMode,
				TemplateSource:     "remote",
				CustomTemplateName: "api-feature",
			},
			want: "unsupported template source: remote",
		},
		{
			name: "custom source outside template mode",
			input: GenerateChangeInput{
				ProjectRoot:        "/project",
				ChangeID:           "valid-change",
				Mode:               domain.BlankMode,
				TemplateSource:     domain.CustomTemplateSource,
				CustomTemplateName: "api-feature",
			},
			want: "custom templates require template generation mode, got: blank",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			customTemplateFiles := newFakeCustomTemplateFileSystem()
			seedCustomTemplate(customTemplateFiles, "api-feature")
			useCase := newCustomTemplateGenerateChange(fileSystem, customTemplateFiles)

			_, err := useCase.Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if customTemplateFiles.operationCount() != 0 {
				t.Fatalf("custom template filesystem operations = %d, want 0", customTemplateFiles.operationCount())
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("generation filesystem operations = %d, want 0", fileSystem.operationCount())
			}
		})
	}
}

func TestGenerateChangeBuiltInTemplateBehaviorUnchangedWhenCustomTemplateSharesName(t *testing.T) {
	changeID := "feature-change"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "feature")
	useCase := NewGenerateChangeWithCustomTemplates(
		fileSystem,
		newFakeBlankChangeContent(),
		templateContent,
		newFakeGuidedChangeContent(),
		customTemplateFiles,
	)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     changeID,
		Mode:         domain.TemplateMode,
		TemplateName: "feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.TemplateSource != domain.BuiltInTemplateSource {
		t.Fatalf("TemplateSource = %q, want %q", result.TemplateSource, domain.BuiltInTemplateSource)
	}
	if result.TemplateName != domain.FeatureTemplate {
		t.Fatalf("TemplateName = %q, want %q", result.TemplateName, domain.FeatureTemplate)
	}
	if result.CustomTemplateName != "" {
		t.Fatalf("CustomTemplateName = %q, want empty", result.CustomTemplateName)
	}
	if customTemplateFiles.operationCount() != 0 {
		t.Fatalf("custom template filesystem operations = %d, want 0", customTemplateFiles.operationCount())
	}
	assertTemplateRequestsEqual(t, templateContent.requests, domain.FeatureTemplate, domain.RequiredOpenSpecChangeFiles())

	changePath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		want := defaultTemplateContent(domain.FeatureTemplate, requiredFile)
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want built-in template content", path, fileSystem.files[path])
		}
	}
}

func TestGenerateChangeCustomTemplateSharingBuiltInNameResolvesFromCustomDirectory(t *testing.T) {
	changeID := "feature-change"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "feature")
	useCase := NewGenerateChangeWithCustomTemplates(
		fileSystem,
		newFakeBlankChangeContent(),
		templateContent,
		newFakeGuidedChangeContent(),
		customTemplateFiles,
	)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           changeID,
		Mode:               domain.TemplateMode,
		TemplateSource:     domain.CustomTemplateSource,
		CustomTemplateName: "feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.TemplateSource != domain.CustomTemplateSource {
		t.Fatalf("TemplateSource = %q, want %q", result.TemplateSource, domain.CustomTemplateSource)
	}
	if result.CustomTemplateName != "feature" {
		t.Fatalf("CustomTemplateName = %q, want feature", result.CustomTemplateName)
	}
	if len(templateContent.requests) != 0 {
		t.Fatalf("built-in template requests = %v, want none", templateContent.requests)
	}

	changePath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		want := "custom:" + requiredFile + ":" + changeID + ":{{title}}:{{summary}}"
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want custom template content", path, fileSystem.files[path])
		}
	}
}

func newCustomTemplateGenerateChange(
	fileSystem *fakeGenerationFileSystem,
	customTemplateFiles *fakeCustomTemplateFileSystem,
) *GenerateChange {
	return NewGenerateChangeWithCustomTemplates(
		fileSystem,
		newFakeBlankChangeContent(),
		newFakeTemplateChangeContent(),
		newFakeGuidedChangeContent(),
		customTemplateFiles,
	)
}

func seedCustomTemplate(fileSystem *fakeCustomTemplateFileSystem, templateName string) {
	templatePath := customTemplatesDirectory + "/" + templateName
	fileSystem.directories[templatePath] = true
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[templatePath+"/"+requiredFile] = "custom:" + requiredFile + ":{{change_id}}:{{title}}:{{summary}}"
	}
}

func assertNoGenerationWrites(t *testing.T, fileSystem *fakeGenerationFileSystem) {
	t.Helper()

	if len(fileSystem.createdDirectories) != 0 {
		t.Fatalf("created directories = %v, want none", fileSystem.createdDirectories)
	}
	if len(fileSystem.writtenFiles) != 0 {
		t.Fatalf("written files = %v, want none", fileSystem.writtenFiles)
	}
}

type fakeCustomTemplateFileSystem struct {
	directories        map[string]bool
	files              map[string]string
	directoryErrors    map[string]error
	fileErrors         map[string]error
	readErrors         map[string]error
	checkedDirectories []string
	checkedFiles       []string
	readFiles          []string
}

func newFakeCustomTemplateFileSystem() *fakeCustomTemplateFileSystem {
	return &fakeCustomTemplateFileSystem{
		directories:     make(map[string]bool),
		files:           make(map[string]string),
		directoryErrors: make(map[string]error),
		fileErrors:      make(map[string]error),
		readErrors:      make(map[string]error),
	}
}

func (fileSystem *fakeCustomTemplateFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedDirectories = append(fileSystem.checkedDirectories, relativePath)
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeCustomTemplateFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedFiles = append(fileSystem.checkedFiles, relativePath)
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeCustomTemplateFileSystem) ReadFile(_ string, relativePath string) (string, error) {
	fileSystem.readFiles = append(fileSystem.readFiles, relativePath)
	if err := fileSystem.readErrors[relativePath]; err != nil {
		return "", err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeCustomTemplateFileSystem) operationCount() int {
	return len(fileSystem.checkedDirectories) +
		len(fileSystem.checkedFiles) +
		len(fileSystem.readFiles)
}
