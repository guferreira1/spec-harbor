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

func TestGenerateChangeCreatesNewTemplateChange(t *testing.T) {
	tests := []struct {
		name         string
		templateName domain.TemplateName
	}{
		{name: "feature", templateName: domain.FeatureTemplate},
		{name: "bugfix", templateName: domain.BugfixTemplate},
		{name: "docs", templateName: domain.DocsTemplate},
		{name: "refactor", templateName: domain.RefactorTemplate},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changeID := test.name + "-change"
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			templateContent := newFakeTemplateChangeContent()
			useCase := NewGenerateChangeWithTemplateContent(fileSystem, newFakeBlankChangeContent(), templateContent)

			result, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot:  "/project",
				ChangeID:     changeID,
				Mode:         domain.TemplateMode,
				TemplateName: string(test.templateName),
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			changePath := openspecChangesDirectory + "/" + changeID
			if result.Mode != domain.TemplateMode {
				t.Fatalf("Mode = %q, want %q", result.Mode, domain.TemplateMode)
			}
			if result.TemplateName != test.templateName {
				t.Fatalf("TemplateName = %q, want %q", result.TemplateName, test.templateName)
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
				want := defaultTemplateContent(test.templateName, requiredFile)
				if fileSystem.files[path] != want {
					t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
				}
			}
			assertTemplateRequestsEqual(t, templateContent.requests, test.templateName, domain.RequiredOpenSpecChangeFiles())
		})
	}
}

func TestGenerateChangeCreatesNewGuidedChange(t *testing.T) {
	tests := []struct {
		name       string
		guidedType domain.GuidedType
	}{
		{name: "feature", guidedType: domain.FeatureGuidedType},
		{name: "bugfix", guidedType: domain.BugfixGuidedType},
		{name: "docs", guidedType: domain.DocsGuidedType},
		{name: "refactor", guidedType: domain.RefactorGuidedType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changeID := test.name + "-guided-change"
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			guidedContent := newFakeGuidedChangeContent()
			useCase := NewGenerateChangeWithContent(
				fileSystem,
				newFakeBlankChangeContent(),
				newFakeTemplateChangeContent(),
				guidedContent,
			)

			result, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    changeID,
				Mode:        domain.GuidedMode,
				GuidedType:  string(test.guidedType),
				Title:       " Add reports ",
				Summary:     " Create report generation support ",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			changePath := openspecChangesDirectory + "/" + changeID
			if result.Mode != domain.GuidedMode {
				t.Fatalf("Mode = %q, want %q", result.Mode, domain.GuidedMode)
			}
			if result.GuidedType != test.guidedType {
				t.Fatalf("GuidedType = %q, want %q", result.GuidedType, test.guidedType)
			}
			if result.GuidedTitle != "Add reports" {
				t.Fatalf("GuidedTitle = %q, want Add reports", result.GuidedTitle)
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
				want := defaultGuidedContent(test.guidedType, "Add reports", "Create report generation support", requiredFile)
				if fileSystem.files[path] != want {
					t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
				}
			}
			assertGuidedRequestsEqual(
				t,
				guidedContent.requests,
				test.guidedType,
				"Add reports",
				"Create report generation support",
				domain.RequiredOpenSpecChangeFiles(),
			)
		})
	}
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

func TestGenerateChangeTemplateFillsMissingFilesAndPreservesExisting(t *testing.T) {
	changeID := "partial-template-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "custom proposal"
	fileSystem.files[changePath+"/tasks.md"] = "custom tasks"

	result, err := NewGenerateChangeWithTemplateContent(
		fileSystem,
		newFakeBlankChangeContent(),
		newFakeTemplateChangeContent(),
	).Execute(GenerateChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     changeID,
		Mode:         domain.TemplateMode,
		TemplateName: string(domain.FeatureTemplate),
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
	for _, requiredFile := range []string{"design.md", "acceptance-criteria.md", "risks.md"} {
		path := changePath + "/" + requiredFile
		want := defaultTemplateContent(domain.FeatureTemplate, requiredFile)
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
		}
	}
}

func TestGenerateChangeGuidedFillsMissingFilesAndPreservesExisting(t *testing.T) {
	changeID := "partial-guided-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "custom proposal"
	fileSystem.files[changePath+"/tasks.md"] = "custom tasks"
	guidedContent := newFakeGuidedChangeContent()

	result, err := NewGenerateChangeWithContent(
		fileSystem,
		newFakeBlankChangeContent(),
		newFakeTemplateChangeContent(),
		guidedContent,
	).Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    changeID,
		Mode:        domain.GuidedMode,
		GuidedType:  string(domain.FeatureGuidedType),
		Title:       "Add reports",
		Summary:     "Create report generation support",
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
	for _, requiredFile := range []string{"design.md", "acceptance-criteria.md", "risks.md"} {
		path := changePath + "/" + requiredFile
		want := defaultGuidedContent(domain.FeatureGuidedType, "Add reports", "Create report generation support", requiredFile)
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
		}
	}
	assertGuidedRequestsEqual(
		t,
		guidedContent.requests,
		domain.FeatureGuidedType,
		"Add reports",
		"Create report generation support",
		domain.RequiredOpenSpecChangeFiles(),
	)
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

func TestGenerateChangeRejectsInvalidTemplateInputBeforeTargetWrites(t *testing.T) {
	tests := []struct {
		name  string
		input GenerateChangeInput
		want  string
	}{
		{
			name: "missing template name",
			input: GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.TemplateMode,
			},
			want: "template name is required",
		},
		{
			name: "unknown template name",
			input: GenerateChangeInput{
				ProjectRoot:  "/project",
				ChangeID:     "change",
				Mode:         domain.TemplateMode,
				TemplateName: "maintenance",
			},
			want: "unknown template name: maintenance",
		},
		{
			name: "unsafe template change id",
			input: GenerateChangeInput{
				ProjectRoot:  "/project",
				ChangeID:     "../unsafe",
				Mode:         domain.TemplateMode,
				TemplateName: string(domain.FeatureTemplate),
			},
			want: "change id must be a safe single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			templateContent := newFakeTemplateChangeContent()

			_, err := NewGenerateChangeWithTemplateContent(
				fileSystem,
				newFakeBlankChangeContent(),
				templateContent,
			).Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("filesystem operations = %d, want 0 before rejecting invalid template input", fileSystem.operationCount())
			}
			if len(templateContent.requests) != 0 {
				t.Fatalf("template content requests = %v, want none before rejecting invalid template input", templateContent.requests)
			}
		})
	}
}

func TestGenerateChangeRejectsMissingGuidedContentDependency(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()

	_, err := NewGenerateChangeWithTemplateContent(
		fileSystem,
		newFakeBlankChangeContent(),
		newFakeTemplateChangeContent(),
	).Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "change",
		Mode:        domain.GuidedMode,
		GuidedType:  string(domain.FeatureGuidedType),
		Title:       "Add reports",
		Summary:     "Create report generation support",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want guided content dependency error")
	}
	if !strings.Contains(err.Error(), "guided change content is required") {
		t.Fatalf("Execute() error = %q, want guided content dependency context", err.Error())
	}
	if fileSystem.operationCount() != 0 {
		t.Fatalf("filesystem operations = %d, want 0 before rejecting missing dependency", fileSystem.operationCount())
	}
}

func TestGenerateChangeRejectsInvalidGuidedInputBeforeTargetWrites(t *testing.T) {
	tests := []struct {
		name  string
		input GenerateChangeInput
		want  string
	}{
		{
			name: "missing guided type",
			input: GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.GuidedMode,
				Title:       "Add reports",
				Summary:     "Create report generation support",
			},
			want: "guided type is required",
		},
		{
			name: "unknown guided type",
			input: GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.GuidedMode,
				GuidedType:  "maintenance",
				Title:       "Add reports",
				Summary:     "Create report generation support",
			},
			want: "unknown guided type: maintenance",
		},
		{
			name: "missing guided title",
			input: GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.GuidedMode,
				GuidedType:  string(domain.FeatureGuidedType),
				Title:       " ",
				Summary:     "Create report generation support",
			},
			want: "guided title is required",
		},
		{
			name: "missing guided summary",
			input: GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.GuidedMode,
				GuidedType:  string(domain.FeatureGuidedType),
				Title:       "Add reports",
				Summary:     " ",
			},
			want: "guided summary is required",
		},
		{
			name: "unsafe guided change id",
			input: GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "../unsafe",
				Mode:        domain.GuidedMode,
				GuidedType:  string(domain.FeatureGuidedType),
				Title:       "Add reports",
				Summary:     "Create report generation support",
			},
			want: "change id must be a safe single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			guidedContent := newFakeGuidedChangeContent()

			_, err := NewGenerateChangeWithContent(
				fileSystem,
				newFakeBlankChangeContent(),
				newFakeTemplateChangeContent(),
				guidedContent,
			).Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if fileSystem.operationCount() != 0 {
				t.Fatalf("filesystem operations = %d, want 0 before rejecting invalid guided input", fileSystem.operationCount())
			}
			if len(guidedContent.requests) != 0 {
				t.Fatalf("guided content requests = %v, want none before rejecting invalid guided input", guidedContent.requests)
			}
		})
	}
}

func TestGenerateChangeTemplateAndGuidedUseSameUnsafeChangeIDValidationAsBlankGeneration(t *testing.T) {
	unsafeChangeIDs := []string{
		".",
		"..",
		"../outside",
		"/outside",
		"bad/id",
		`bad\id`,
		"C:bad",
		"-bad",
	}

	for _, changeID := range unsafeChangeIDs {
		t.Run(changeID, func(t *testing.T) {
			blankFileSystem := newFakeGenerationFileSystem()
			_, blankErr := NewGenerateChange(blankFileSystem, newFakeBlankChangeContent()).Execute(GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    changeID,
				Mode:        domain.BlankMode,
			})
			if blankErr == nil {
				t.Fatalf("blank Execute() error = nil, want unsafe change id error")
			}

			templateFileSystem := newFakeGenerationFileSystem()
			templateContent := newFakeTemplateChangeContent()
			_, templateErr := NewGenerateChangeWithTemplateContent(
				templateFileSystem,
				newFakeBlankChangeContent(),
				templateContent,
			).Execute(GenerateChangeInput{
				ProjectRoot:  "/project",
				ChangeID:     changeID,
				Mode:         domain.TemplateMode,
				TemplateName: string(domain.FeatureTemplate),
			})
			if templateErr == nil {
				t.Fatalf("template Execute() error = nil, want unsafe change id error")
			}
			if templateErr.Error() != blankErr.Error() {
				t.Fatalf("template Execute() error = %q, want same unsafe id error as blank %q", templateErr.Error(), blankErr.Error())
			}
			if blankFileSystem.operationCount() != 0 {
				t.Fatalf("blank filesystem operations = %d, want 0 before rejecting unsafe id", blankFileSystem.operationCount())
			}
			if templateFileSystem.operationCount() != 0 {
				t.Fatalf("template filesystem operations = %d, want 0 before rejecting unsafe id", templateFileSystem.operationCount())
			}
			if len(templateContent.requests) != 0 {
				t.Fatalf("template content requests = %v, want none before rejecting unsafe id", templateContent.requests)
			}

			guidedFileSystem := newFakeGenerationFileSystem()
			guidedContent := newFakeGuidedChangeContent()
			_, guidedErr := NewGenerateChangeWithContent(
				guidedFileSystem,
				newFakeBlankChangeContent(),
				newFakeTemplateChangeContent(),
				guidedContent,
			).Execute(GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    changeID,
				Mode:        domain.GuidedMode,
				GuidedType:  string(domain.FeatureGuidedType),
				Title:       "Add reports",
				Summary:     "Create report generation support",
			})
			if guidedErr == nil {
				t.Fatalf("guided Execute() error = nil, want unsafe change id error")
			}
			if guidedErr.Error() != blankErr.Error() {
				t.Fatalf("guided Execute() error = %q, want same unsafe id error as blank %q", guidedErr.Error(), blankErr.Error())
			}
			if guidedFileSystem.operationCount() != 0 {
				t.Fatalf("guided filesystem operations = %d, want 0 before rejecting unsafe id", guidedFileSystem.operationCount())
			}
			if len(guidedContent.requests) != 0 {
				t.Fatalf("guided content requests = %v, want none before rejecting unsafe id", guidedContent.requests)
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

func TestGenerateChangeTemplateRejectsMissingOpenSpecProjectBeforeTargetWrites(t *testing.T) {
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
			templateContent := newFakeTemplateChangeContent()
			test.setup(fileSystem)

			_, err := NewGenerateChangeWithTemplateContent(
				fileSystem,
				newFakeBlankChangeContent(),
				templateContent,
			).Execute(GenerateChangeInput{
				ProjectRoot:  "/project",
				ChangeID:     "change",
				Mode:         domain.TemplateMode,
				TemplateName: string(domain.FeatureTemplate),
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
			if len(templateContent.requests) != 0 {
				t.Fatalf("template content requests = %v, want none before project structure is available", templateContent.requests)
			}
		})
	}
}

func TestGenerateChangeGuidedRejectsMissingOpenSpecProjectBeforeTargetWrites(t *testing.T) {
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
			guidedContent := newFakeGuidedChangeContent()
			test.setup(fileSystem)

			_, err := NewGenerateChangeWithContent(
				fileSystem,
				newFakeBlankChangeContent(),
				newFakeTemplateChangeContent(),
				guidedContent,
			).Execute(GenerateChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				Mode:        domain.GuidedMode,
				GuidedType:  string(domain.FeatureGuidedType),
				Title:       "Add reports",
				Summary:     "Create report generation support",
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
			if len(guidedContent.requests) != 0 {
				t.Fatalf("guided content requests = %v, want none before project structure is available", guidedContent.requests)
			}
		})
	}
}

func TestGenerateChangeTemplateReturnsContentLoadingErrorsBeforeFileWrites(t *testing.T) {
	wantErr := errors.New("template content unavailable")
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	templateContent.errors["feature:proposal.md"] = wantErr

	_, err := NewGenerateChangeWithTemplateContent(
		fileSystem,
		newFakeBlankChangeContent(),
		templateContent,
	).Execute(GenerateChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     "change",
		Mode:         domain.TemplateMode,
		TemplateName: string(domain.FeatureTemplate),
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want content error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "load feature template content for proposal.md") {
		t.Fatalf("Execute() error = %q, want template content context", err.Error())
	}
	if len(fileSystem.writtenFiles) != 0 {
		t.Fatalf("written files = %v, want none when first template file content cannot load", fileSystem.writtenFiles)
	}
	if len(templateContent.requests) != 1 {
		t.Fatalf("template content requests = %v, want only first required file", templateContent.requests)
	}
	firstRequiredFile := domain.RequiredOpenSpecChangeFiles()[0]
	if templateContent.requests[0].templateName != domain.FeatureTemplate || templateContent.requests[0].relativePath != firstRequiredFile {
		t.Fatalf("template content requests = %v, want first %s request for %s", templateContent.requests, domain.FeatureTemplate, firstRequiredFile)
	}
}

func TestGenerateChangeGuidedReturnsContentLoadingErrorsBeforeFileWrites(t *testing.T) {
	wantErr := errors.New("guided content unavailable")
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	guidedContent := newFakeGuidedChangeContent()
	guidedContent.errors["feature:proposal.md"] = wantErr

	_, err := NewGenerateChangeWithContent(
		fileSystem,
		newFakeBlankChangeContent(),
		newFakeTemplateChangeContent(),
		guidedContent,
	).Execute(GenerateChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "change",
		Mode:        domain.GuidedMode,
		GuidedType:  string(domain.FeatureGuidedType),
		Title:       "Add reports",
		Summary:     "Create report generation support",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want content error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "load feature guided content for proposal.md") {
		t.Fatalf("Execute() error = %q, want guided content context", err.Error())
	}
	if len(fileSystem.writtenFiles) != 0 {
		t.Fatalf("written files = %v, want none when first guided file content cannot load", fileSystem.writtenFiles)
	}
	if len(guidedContent.requests) != 1 {
		t.Fatalf("guided content requests = %v, want only first required file", guidedContent.requests)
	}
	firstRequiredFile := domain.RequiredOpenSpecChangeFiles()[0]
	if guidedContent.requests[0].guidedType != domain.FeatureGuidedType ||
		guidedContent.requests[0].title != "Add reports" ||
		guidedContent.requests[0].summary != "Create report generation support" ||
		guidedContent.requests[0].relativePath != firstRequiredFile {
		t.Fatalf("guided content requests = %v, want first feature request for %s", guidedContent.requests, firstRequiredFile)
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

func TestGenerateChangeReturnsTemplateContentLoadingErrors(t *testing.T) {
	wantErr := errors.New("template content unavailable")
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	templateContent.errors["feature:proposal.md"] = wantErr

	_, err := NewGenerateChangeWithTemplateContent(
		fileSystem,
		newFakeBlankChangeContent(),
		templateContent,
	).Execute(GenerateChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     "change",
		Mode:         domain.TemplateMode,
		TemplateName: string(domain.FeatureTemplate),
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want content error")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "load feature template content for proposal.md") {
		t.Fatalf("Execute() error = %q, want template content context", err.Error())
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

type templateContentRequest struct {
	templateName domain.TemplateName
	relativePath string
}

type fakeTemplateChangeContent struct {
	errors   map[string]error
	requests []templateContentRequest
}

func newFakeTemplateChangeContent() *fakeTemplateChangeContent {
	return &fakeTemplateChangeContent{errors: make(map[string]error)}
}

func (content *fakeTemplateChangeContent) ContentFor(templateName domain.TemplateName, relativePath string) (string, error) {
	content.requests = append(content.requests, templateContentRequest{
		templateName: templateName,
		relativePath: relativePath,
	})
	if err := content.errors[string(templateName)+":"+relativePath]; err != nil {
		return "", err
	}
	return defaultTemplateContent(templateName, relativePath), nil
}

func defaultTemplateContent(templateName domain.TemplateName, relativePath string) string {
	return "template:" + string(templateName) + ":" + relativePath
}

type guidedContentRequest struct {
	guidedType   domain.GuidedType
	title        string
	summary      string
	relativePath string
}

type fakeGuidedChangeContent struct {
	errors   map[string]error
	requests []guidedContentRequest
}

func newFakeGuidedChangeContent() *fakeGuidedChangeContent {
	return &fakeGuidedChangeContent{errors: make(map[string]error)}
}

func (content *fakeGuidedChangeContent) ContentFor(
	guidedType domain.GuidedType,
	title string,
	summary string,
	relativePath string,
) (string, error) {
	content.requests = append(content.requests, guidedContentRequest{
		guidedType:   guidedType,
		title:        title,
		summary:      summary,
		relativePath: relativePath,
	})
	if err := content.errors[string(guidedType)+":"+relativePath]; err != nil {
		return "", err
	}
	return defaultGuidedContent(guidedType, title, summary, relativePath), nil
}

func defaultGuidedContent(guidedType domain.GuidedType, title string, summary string, relativePath string) string {
	return "guided:" + string(guidedType) + ":" + title + ":" + summary + ":" + relativePath
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

func assertTemplateRequestsEqual(
	t *testing.T,
	got []templateContentRequest,
	templateName domain.TemplateName,
	requiredFiles []string,
) {
	t.Helper()

	if len(got) != len(requiredFiles) {
		t.Fatalf("template requests = %v, want %d requests", got, len(requiredFiles))
	}
	for index, requiredFile := range requiredFiles {
		if got[index].templateName != templateName || got[index].relativePath != requiredFile {
			t.Fatalf("template requests = %v, want %s request for %s at index %d", got, templateName, requiredFile, index)
		}
	}
}

func assertGuidedRequestsEqual(
	t *testing.T,
	got []guidedContentRequest,
	guidedType domain.GuidedType,
	title string,
	summary string,
	requiredFiles []string,
) {
	t.Helper()

	if len(got) != len(requiredFiles) {
		t.Fatalf("guided requests = %v, want %d requests", got, len(requiredFiles))
	}
	for index, requiredFile := range requiredFiles {
		request := got[index]
		if request.guidedType != guidedType ||
			request.title != title ||
			request.summary != summary ||
			request.relativePath != requiredFile {
			t.Fatalf("guided requests = %v, want %s request for %s at index %d", got, guidedType, requiredFile, index)
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
