package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateChangeConfigTemplateBuiltInAlias(t *testing.T) {
	changeID := "add-feature"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithTemplateAlias(t, "default-feature", "builtin", "feature"),
	}
	useCase := newConfigTemplateGenerateChange(fileSystem, templateContent, newFakeCustomTemplateFileSystem(), configFileSystem, parser)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            changeID,
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "default-feature",
		Title:               "Ignored by built-in",
		Summary:             "Also ignored by built-in",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ConfigTemplateAlias != "default-feature" {
		t.Fatalf("ConfigTemplateAlias = %q, want default-feature", result.ConfigTemplateAlias)
	}
	if result.ConfigTemplateSource != domain.ConfigTemplateSourceBuiltin {
		t.Fatalf("ConfigTemplateSource = %q, want builtin", result.ConfigTemplateSource)
	}
	if result.ConfigTemplateName != "feature" {
		t.Fatalf("ConfigTemplateName = %q, want feature", result.ConfigTemplateName)
	}
	if result.TemplateSource != domain.BuiltInTemplateSource {
		t.Fatalf("TemplateSource = %q, want built-in", result.TemplateSource)
	}
	assertTemplateRequestsEqual(t, templateContent.requests, domain.FeatureTemplate, domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.CreatedFiles(), domain.RequiredOpenSpecChangeFiles())

	changePath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		want := defaultTemplateContent(domain.FeatureTemplate, requiredFile)
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
		}
	}
}

func TestGenerateChangeConfigTemplateCustomAliasPassesTitleAndSummary(t *testing.T) {
	changeID := "add-payment-flow"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithTemplateAlias(t, "api-feature", "custom", "api-feature"),
	}
	useCase := newConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), customTemplateFiles, configFileSystem, parser)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            changeID,
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "api-feature",
		Title:               "Add payments",
		Summary:             "Adds a payment flow.",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ConfigTemplateAlias != "api-feature" {
		t.Fatalf("ConfigTemplateAlias = %q, want api-feature", result.ConfigTemplateAlias)
	}
	if result.ConfigTemplateSource != domain.ConfigTemplateSourceCustom {
		t.Fatalf("ConfigTemplateSource = %q, want custom", result.ConfigTemplateSource)
	}
	if result.ConfigTemplateName != "api-feature" {
		t.Fatalf("ConfigTemplateName = %q, want api-feature", result.ConfigTemplateName)
	}
	if result.TemplatePath != customTemplatesDirectory+"/api-feature" {
		t.Fatalf("TemplatePath = %q, want %q", result.TemplatePath, customTemplatesDirectory+"/api-feature")
	}

	changePath := openspecChangesDirectory + "/" + changeID
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		want := "custom:" + requiredFile + ":" + changeID + ":Add payments:Adds a payment flow."
		if fileSystem.files[path] != want {
			t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
		}
	}
}

func TestGenerateChangeConfigTemplateReturnsConfigErrorsWithoutWrites(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*fakeConfigTemplateConfigFileSystem, *fakeConfigTemplateParser)
		want       string
		wantParser bool
	}{
		{
			name:       "missing config",
			setup:      func(_ *fakeConfigTemplateConfigFileSystem, _ *fakeConfigTemplateParser) {},
			want:       "missing config file: .specharbor/config.yml",
			wantParser: false,
		},
		{
			name: "unreadable config",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, _ *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				fileSystem.readErrors[localConfigPath] = errors.New("permission denied")
			},
			want:       "unreadable config .specharbor/config.yml: permission denied",
			wantParser: false,
		},
		{
			name: "invalid config",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, parser *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				parser.err = errors.New("parse local config YAML: bad yaml")
			},
			want:       "invalid config YAML in .specharbor/config.yml: parse local config YAML: bad yaml",
			wantParser: true,
		},
		{
			name: "missing version",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, parser *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				parser.config = domain.LocalConfig{}
			},
			want:       "missing config version in .specharbor/config.yml: supported version is 1",
			wantParser: true,
		},
		{
			name: "unsupported version",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, parser *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				parser.config = domain.LocalConfig{Version: 2}
			},
			want:       "unsupported config version 2 in .specharbor/config.yml: supported version is 1",
			wantParser: true,
		},
		{
			name: "missing alias",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, parser *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				parser.config = domain.LocalConfig{
					Version:   domain.SupportedLocalConfigVersion,
					Templates: domain.NewConfigTemplates(domain.EmptyConfigTemplateAliases()),
				}
			},
			want:       "config template alias not found: api-feature",
			wantParser: true,
		},
		{
			name: "unsupported source from config validation",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, parser *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				parser.err = errors.New(`invalid config template alias "api-feature": unsupported config template source: remote`)
			},
			want:       `invalid config YAML in .specharbor/config.yml: invalid config template alias "api-feature": unsupported config template source: remote`,
			wantParser: true,
		},
		{
			name: "unknown builtin template from config validation",
			setup: func(fileSystem *fakeConfigTemplateConfigFileSystem, parser *fakeConfigTemplateParser) {
				seedConfigTemplateConfig(fileSystem)
				parser.err = errors.New(`invalid config template alias "api-feature": unknown template name: maintenance`)
			},
			want:       `invalid config YAML in .specharbor/config.yml: invalid config template alias "api-feature": unknown template name: maintenance`,
			wantParser: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			configFileSystem := newFakeConfigTemplateConfigFileSystem()
			parser := &fakeConfigTemplateParser{}
			test.setup(configFileSystem, parser)
			useCase := newConfigTemplateGenerateChange(
				fileSystem,
				newFakeTemplateChangeContent(),
				newFakeCustomTemplateFileSystem(),
				configFileSystem,
				parser,
			)

			_, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot:         "/project",
				ChangeID:            "new-change",
				Mode:                domain.TemplateMode,
				ConfigTemplate:      true,
				ConfigTemplateAlias: "api-feature",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if parser.calls > 0 != test.wantParser {
				t.Fatalf("parser calls = %d, want parser called: %t", parser.calls, test.wantParser)
			}
			assertNoGenerationWrites(t, fileSystem)
		})
	}
}

func TestGenerateChangeConfigTemplateRejectsInvalidCLIAliasBeforeFilesystemAccess(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithTemplateAlias(t, "api-feature", "custom", "api-feature"),
	}
	useCase := newConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), customTemplateFiles, configFileSystem, parser)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "new-change",
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "../escape",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want invalid alias error")
	}
	if err.Error() != "config template alias must be a single path segment" {
		t.Fatalf("Execute() error = %q, want invalid alias error", err.Error())
	}
	if configFileSystem.operationCount() != 0 {
		t.Fatalf("config filesystem operations = %d, want 0", configFileSystem.operationCount())
	}
	if customTemplateFiles.operationCount() != 0 {
		t.Fatalf("custom template filesystem operations = %d, want 0", customTemplateFiles.operationCount())
	}
	if fileSystem.operationCount() != 0 {
		t.Fatalf("generation filesystem operations = %d, want 0", fileSystem.operationCount())
	}
}

func TestGenerateChangeConfigTemplateCustomSourceErrorsWriteNothing(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*fakeCustomTemplateFileSystem)
		want  string
	}{
		{
			name:  "missing custom template",
			setup: func(_ *fakeCustomTemplateFileSystem) {},
			want:  "unknown custom template: api-feature. Expected directory: .specharbor/templates/api-feature",
		},
		{
			name: "missing required custom template files",
			setup: func(customTemplateFiles *fakeCustomTemplateFileSystem) {
				seedCustomTemplate(customTemplateFiles, "api-feature")
				delete(customTemplateFiles.files, customTemplatesDirectory+"/api-feature/proposal.md")
			},
			want: "custom template api-feature is missing required files: proposal.md",
		},
		{
			name: "empty custom template file",
			setup: func(customTemplateFiles *fakeCustomTemplateFileSystem) {
				seedCustomTemplate(customTemplateFiles, "api-feature")
				customTemplateFiles.files[customTemplatesDirectory+"/api-feature/proposal.md"] = "  \n"
			},
			want: "custom template file api-feature/proposal.md is empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			customTemplateFiles := newFakeCustomTemplateFileSystem()
			test.setup(customTemplateFiles)
			configFileSystem := newFakeConfigTemplateConfigFileSystem()
			seedConfigTemplateConfig(configFileSystem)
			parser := &fakeConfigTemplateParser{
				config: configWithTemplateAlias(t, "api-feature", "custom", "api-feature"),
			}
			useCase := newConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), customTemplateFiles, configFileSystem, parser)

			_, err := useCase.Execute(GenerateChangeInput{
				ProjectRoot:         "/project",
				ChangeID:            "new-change",
				Mode:                domain.TemplateMode,
				ConfigTemplate:      true,
				ConfigTemplateAlias: "api-feature",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			assertNoGenerationWrites(t, fileSystem)
		})
	}
}

func TestGenerateChangeConfigTemplateWritesOnlyUnderChangePathAndSkipsExistingFiles(t *testing.T) {
	changeID := "existing-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "existing proposal"
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "api-feature")
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithTemplateAlias(t, "api-feature", "custom", "api-feature"),
	}
	useCase := newConfigTemplateGenerateChange(fileSystem, newFakeTemplateChangeContent(), customTemplateFiles, configFileSystem, parser)

	result, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            changeID,
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "api-feature",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertStringSlicesEqual(t, result.SkippedExistingFiles(), []string{"proposal.md"})
	for _, writtenFile := range fileSystem.writtenFiles {
		if !strings.HasPrefix(writtenFile, changePath+"/") {
			t.Fatalf("write target %q is outside %q", writtenFile, changePath)
		}
	}
	if fileSystem.files[changePath+"/proposal.md"] != "existing proposal" {
		t.Fatalf("proposal.md = %q, want preserved existing proposal", fileSystem.files[changePath+"/proposal.md"])
	}
	assertPathNotWritten(t, fileSystem, "internal/app.go")
	assertPathNotWritten(t, fileSystem, ".github/workflows/ci.yml")
}

func TestGenerateChangeConfigTemplateKeepsDirectNamespacesDisjoint(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeTemplateChangeContent()
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedCustomTemplate(customTemplateFiles, "feature")
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{
		config: configWithTemplateAlias(t, "feature", "custom", "feature"),
	}
	useCase := newConfigTemplateGenerateChange(fileSystem, templateContent, customTemplateFiles, configFileSystem, parser)

	_, err := useCase.Execute(GenerateChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     "direct-builtin",
		Mode:         domain.TemplateMode,
		TemplateName: "feature",
	})
	if err != nil {
		t.Fatalf("direct built-in Execute() error = %v", err)
	}
	if parser.calls != 0 {
		t.Fatalf("parser calls after direct built-in = %d, want 0", parser.calls)
	}
	if customTemplateFiles.operationCount() != 0 {
		t.Fatalf("custom template operations after direct built-in = %d, want 0", customTemplateFiles.operationCount())
	}

	_, err = useCase.Execute(GenerateChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "config-template",
		Mode:                domain.TemplateMode,
		ConfigTemplate:      true,
		ConfigTemplateAlias: "feature",
	})
	if err != nil {
		t.Fatalf("config template Execute() error = %v", err)
	}
	if parser.calls != 1 {
		t.Fatalf("parser calls after config template = %d, want 1", parser.calls)
	}
	changePath := openspecChangesDirectory + "/config-template"
	if fileSystem.files[changePath+"/proposal.md"] != "custom:proposal.md:config-template:{{title}}:{{summary}}" {
		t.Fatalf("config template proposal = %q, want custom content", fileSystem.files[changePath+"/proposal.md"])
	}
}

func newConfigTemplateGenerateChange(
	fileSystem *fakeGenerationFileSystem,
	templateContent *fakeTemplateChangeContent,
	customTemplateFiles *fakeCustomTemplateFileSystem,
	configFileSystem *fakeConfigTemplateConfigFileSystem,
	parser *fakeConfigTemplateParser,
) *GenerateChange {
	return NewGenerateChangeWithConfigTemplates(
		fileSystem,
		newFakeBlankChangeContent(),
		templateContent,
		newFakeGuidedChangeContent(),
		customTemplateFiles,
		configFileSystem,
		parser,
	)
}

func configWithTemplateAlias(t *testing.T, aliasValue string, source string, template string) domain.LocalConfig {
	t.Helper()

	alias, err := domain.NewConfigTemplateAlias(aliasValue)
	if err != nil {
		t.Fatalf("NewConfigTemplateAlias(%q) error = %v", aliasValue, err)
	}
	reference, err := domain.NewConfigTemplateReference(domain.ConfigTemplateReferenceInput{
		Alias:    alias,
		Source:   source,
		Template: template,
	})
	if err != nil {
		t.Fatalf("NewConfigTemplateReference(%q, %q, %q) error = %v", aliasValue, source, template, err)
	}
	return domain.LocalConfig{
		Version:   domain.SupportedLocalConfigVersion,
		Templates: domain.NewConfigTemplates(domain.NewConfigTemplateAliases([]domain.ConfigTemplateReference{reference})),
	}
}

func seedConfigTemplateConfig(fileSystem *fakeConfigTemplateConfigFileSystem) {
	fileSystem.files[localConfigPath] = "version: 1\n"
}

func assertPathNotWritten(t *testing.T, fileSystem *fakeGenerationFileSystem, relativePath string) {
	t.Helper()

	for _, writtenFile := range fileSystem.writtenFiles {
		if writtenFile == relativePath {
			t.Fatalf("unexpected write to %q", relativePath)
		}
	}
	if _, exists := fileSystem.files[relativePath]; exists {
		t.Fatalf("unexpected file at %q", relativePath)
	}
}

type fakeConfigTemplateConfigFileSystem struct {
	files        map[string]string
	fileErrors   map[string]error
	readErrors   map[string]error
	checkedFiles []string
	readFiles    []string
}

func newFakeConfigTemplateConfigFileSystem() *fakeConfigTemplateConfigFileSystem {
	return &fakeConfigTemplateConfigFileSystem{
		files:      make(map[string]string),
		fileErrors: make(map[string]error),
		readErrors: make(map[string]error),
	}
}

func (fileSystem *fakeConfigTemplateConfigFileSystem) DirectoryExists(_ string, _ string) (bool, error) {
	return true, nil
}

func (fileSystem *fakeConfigTemplateConfigFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedFiles = append(fileSystem.checkedFiles, relativePath)
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeConfigTemplateConfigFileSystem) ReadFile(_ string, relativePath string) (string, error) {
	fileSystem.readFiles = append(fileSystem.readFiles, relativePath)
	if err := fileSystem.readErrors[relativePath]; err != nil {
		return "", err
	}
	return fileSystem.files[relativePath], nil
}

func (fileSystem *fakeConfigTemplateConfigFileSystem) operationCount() int {
	return len(fileSystem.checkedFiles) + len(fileSystem.readFiles)
}

type fakeConfigTemplateParser struct {
	config   domain.LocalConfig
	err      error
	calls    int
	contents string
}

func (parser *fakeConfigTemplateParser) ParseLocalConfig(contents string) (domain.LocalConfig, error) {
	parser.calls++
	parser.contents = contents
	if parser.err != nil {
		return domain.LocalConfig{}, parser.err
	}
	return parser.config, nil
}
