package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateHybridChangeBuiltInTemplateSource(t *testing.T) {
	changeID := "add-login"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	templateContent := newFakeHybridTemplateContent()
	validator := newFakeHybridValidator()
	useCase := newHybridGenerateChange(
		fileSystem,
		templateContent,
		newFakeCustomTemplateFileSystem(),
		newFakeConfigTemplateConfigFileSystem(),
		&fakeConfigTemplateParser{},
		&fakeRemoteTemplateFetcher{},
		&fakeRemoteTemplateBundleReader{},
		validator,
	)

	result, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     changeID,
		TemplateName: "feature",
		Title:        " Add login ",
		Summary:      " Add authentication support. ",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertHybridResultBasics(t, result, changeID, domain.HybridSelectedSourceBuiltIn, "feature", domain.HybridResolvedSourceBuiltin, "feature")
	if !result.TypeAvailable || result.EffectiveType != domain.HybridTypeFeature {
		t.Fatalf("EffectiveType = %q available=%t, want feature true", result.EffectiveType, result.TypeAvailable)
	}
	assertStringSlicesEqual(t, result.CreatedFiles(), domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), nil)
	assertHybridFileContent(t, fileSystem, changeID, "proposal.md", "hybrid:feature:proposal.md:add-login:Add login:Add authentication support.:feature:{{unknown}}")
	if validator.calls != 1 || validator.inputs[0].ChangeID != changeID {
		t.Fatalf("validator calls = %d inputs = %+v, want one validation", validator.calls, validator.inputs)
	}
}

func TestGenerateHybridChangeCustomTemplateSource(t *testing.T) {
	changeID := "add-payment-flow"
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedHybridCustomTemplate(customTemplateFiles, "api-feature")
	validator := newFakeHybridValidator()
	useCase := newHybridGenerateChange(
		fileSystem,
		newFakeHybridTemplateContent(),
		customTemplateFiles,
		newFakeConfigTemplateConfigFileSystem(),
		&fakeConfigTemplateParser{},
		&fakeRemoteTemplateFetcher{},
		&fakeRemoteTemplateBundleReader{},
		validator,
	)

	result, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           changeID,
		CustomTemplateName: "api-feature",
		Title:              "Add payments",
		Summary:            "Adds a payment flow.",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertHybridResultBasics(t, result, changeID, domain.HybridSelectedSourceCustom, "api-feature", domain.HybridResolvedSourceCustom, "api-feature")
	if result.TypeAvailable {
		t.Fatalf("TypeAvailable = true, want false for custom without type")
	}
	assertHybridFileContent(t, fileSystem, changeID, "proposal.md", "custom:proposal.md:add-payment-flow:Add payments:Adds a payment flow.:{{type}}:{{unknown}}")

	result, err = useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:        "/project",
		ChangeID:           "typed-custom",
		CustomTemplateName: "api-feature",
		Title:              "Add payments",
		Summary:            "Adds a payment flow.",
		Type:               "docs",
	})
	if err != nil {
		t.Fatalf("Execute(typed custom) error = %v", err)
	}
	if !result.TypeAvailable || result.EffectiveType != domain.HybridTypeDocs {
		t.Fatalf("typed custom EffectiveType = %q available=%t, want docs true", result.EffectiveType, result.TypeAvailable)
	}
	assertHybridFileContent(t, fileSystem, "typed-custom", "proposal.md", "custom:proposal.md:typed-custom:Add payments:Adds a payment flow.:docs:{{unknown}}")
}

func TestGenerateHybridChangeConfigTemplateSources(t *testing.T) {
	tests := []struct {
		name               string
		alias              string
		config             domain.LocalConfig
		setupCustom        bool
		remoteBundle       domain.RemoteTemplateBundle
		wantSelected       string
		wantResolvedKind   domain.HybridResolvedSourceKind
		wantResolvedName   string
		wantTypeAvailable  bool
		wantType           domain.HybridType
		wantProposal       string
		wantFetcherCalls   int
		wantReaderCalls    int
		optionalType       string
		downloadedBytes    []byte
		remoteChecksumFrom []byte
	}{
		{
			name:              "config builtin alias",
			alias:             "default-feature",
			config:            configWithTemplateAlias(t, "default-feature", "builtin", "feature"),
			wantSelected:      "default-feature",
			wantResolvedKind:  domain.HybridResolvedSourceBuiltin,
			wantResolvedName:  "feature",
			wantTypeAvailable: true,
			wantType:          domain.HybridTypeFeature,
			wantProposal:      "hybrid:feature:proposal.md:config-builtin-alias:Add login:Add auth.:feature:{{unknown}}",
		},
		{
			name:             "config custom alias",
			alias:            "api-feature",
			config:           configWithTemplateAlias(t, "api-feature", "custom", "api-feature"),
			setupCustom:      true,
			wantSelected:     "api-feature",
			wantResolvedKind: domain.HybridResolvedSourceCustom,
			wantResolvedName: "api-feature",
			wantProposal:     "custom:proposal.md:config-custom-alias:Add login:Add auth.:{{type}}:{{unknown}}",
		},
		{
			name:              "config custom provided type",
			alias:             "typed-api-feature",
			config:            configWithTemplateAlias(t, "typed-api-feature", "custom", "api-feature"),
			setupCustom:       true,
			wantSelected:      "typed-api-feature",
			wantResolvedKind:  domain.HybridResolvedSourceCustom,
			wantResolvedName:  "api-feature",
			wantTypeAvailable: true,
			wantType:          domain.HybridTypeDocs,
			wantProposal:      "custom:proposal.md:config-custom-provided-type:Add login:Add auth.:docs:{{unknown}}",
			optionalType:      "docs",
		},
		{
			name:               "config remote alias",
			alias:              "service-feature",
			wantSelected:       "service-feature",
			wantResolvedKind:   domain.HybridResolvedSourceRemote,
			wantResolvedName:   "example.com",
			wantProposal:       "remote:proposal.md:config-remote-alias:Add login:Add auth.:{{type}}:{{unknown}}",
			wantFetcherCalls:   1,
			wantReaderCalls:    1,
			downloadedBytes:    []byte("zip bytes"),
			remoteChecksumFrom: []byte("zip bytes"),
		},
		{
			name:               "config remote provided type",
			alias:              "typed-service",
			wantSelected:       "typed-service",
			wantResolvedKind:   domain.HybridResolvedSourceRemote,
			wantResolvedName:   "example.com",
			wantTypeAvailable:  true,
			wantType:           domain.HybridTypeRefactor,
			wantProposal:       "remote:proposal.md:config-remote-provided-type:Add login:Add auth.:refactor:{{unknown}}",
			wantFetcherCalls:   1,
			wantReaderCalls:    1,
			optionalType:       "refactor",
			downloadedBytes:    []byte("zip bytes"),
			remoteChecksumFrom: []byte("zip bytes"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changeID := strings.ReplaceAll(test.name, " ", "-")
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			templateContent := newFakeHybridTemplateContent()
			customTemplateFiles := newFakeCustomTemplateFileSystem()
			if test.setupCustom {
				seedHybridCustomTemplate(customTemplateFiles, "api-feature")
			}
			configFileSystem := newFakeConfigTemplateConfigFileSystem()
			seedConfigTemplateConfig(configFileSystem)
			parser := &fakeConfigTemplateParser{config: test.config}
			fetcher := &fakeRemoteTemplateFetcher{}
			reader := &fakeRemoteTemplateBundleReader{}
			if test.wantResolvedKind == domain.HybridResolvedSourceRemote {
				checksumBytes := test.remoteChecksumFrom
				checksum := domain.NewRemoteTemplateChecksumFromBytes(checksumBytes).String()
				parser.config = configWithRemoteTemplateAlias(t, test.alias, "https://example.com/templates/service-feature.zip", checksum, "zip")
				fetcher.result = domain.NewRemoteTemplateFetchResult(200, test.downloadedBytes)
				reader.bundle = mustRemoteTemplateBundle(t, hybridRemoteFiles("remote"))
			}
			validator := newFakeHybridValidator()
			useCase := newHybridGenerateChange(fileSystem, templateContent, customTemplateFiles, configFileSystem, parser, fetcher, reader, validator)

			result, err := useCase.Execute(GenerateHybridChangeInput{
				ProjectRoot:         "/project",
				ChangeID:            changeID,
				ConfigTemplateAlias: test.alias,
				Title:               "Add login",
				Summary:             "Add auth.",
				Type:                test.optionalType,
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			assertHybridResultBasics(t, result, changeID, domain.HybridSelectedSourceConfig, test.wantSelected, test.wantResolvedKind, test.wantResolvedName)
			if result.TypeAvailable != test.wantTypeAvailable || result.EffectiveType != test.wantType {
				t.Fatalf("type = %q available=%t, want %q %t", result.EffectiveType, result.TypeAvailable, test.wantType, test.wantTypeAvailable)
			}
			assertHybridFileContent(t, fileSystem, changeID, "proposal.md", test.wantProposal)
			if fetcher.calls != test.wantFetcherCalls {
				t.Fatalf("fetcher calls = %d, want %d", fetcher.calls, test.wantFetcherCalls)
			}
			if reader.calls != test.wantReaderCalls {
				t.Fatalf("reader calls = %d, want %d", reader.calls, test.wantReaderCalls)
			}
		})
	}
}

func TestGenerateHybridChangeRejectsInvalidInputBeforeWrites(t *testing.T) {
	tests := []struct {
		name  string
		input GenerateHybridChangeInput
		want  string
	}{
		{name: "missing source", input: GenerateHybridChangeInput{ProjectRoot: "/project", ChangeID: "change", Title: "Title", Summary: "Summary"}, want: "hybrid source selection error: hybrid source selector is required"},
		{name: "multiple sources", input: GenerateHybridChangeInput{ProjectRoot: "/project", ChangeID: "change", TemplateName: "feature", CustomTemplateName: "api-feature", Title: "Title", Summary: "Summary"}, want: "hybrid source selection error: hybrid requires exactly one source selector"},
		{name: "invalid source", input: GenerateHybridChangeInput{ProjectRoot: "/project", ChangeID: "change", TemplateName: "maintenance", Title: "Title", Summary: "Summary"}, want: "hybrid source selection error: unknown template name: maintenance"},
		{name: "missing title", input: GenerateHybridChangeInput{ProjectRoot: "/project", ChangeID: "change", TemplateName: "feature", Summary: "Summary"}, want: "hybrid metadata error: hybrid title is required"},
		{name: "missing summary", input: GenerateHybridChangeInput{ProjectRoot: "/project", ChangeID: "change", TemplateName: "feature", Title: "Title"}, want: "hybrid metadata error: hybrid summary is required"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			useCase := newHybridGenerateChange(
				fileSystem,
				newFakeHybridTemplateContent(),
				newFakeCustomTemplateFileSystem(),
				newFakeConfigTemplateConfigFileSystem(),
				&fakeConfigTemplateParser{},
				&fakeRemoteTemplateFetcher{},
				&fakeRemoteTemplateBundleReader{},
				newFakeHybridValidator(),
			)

			_, err := useCase.Execute(test.input)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Execute() error = %v, want %q", err, test.want)
			}
			assertNoGenerationWrites(t, fileSystem)
			if fileSystem.directories[openspecChangesDirectory+"/change"] {
				t.Fatalf("change directory was created")
			}
		})
	}
}

func TestGenerateHybridChangeBuiltInTypeMismatchWritesNothing(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	useCase := newHybridGenerateChange(
		fileSystem,
		newFakeHybridTemplateContent(),
		newFakeCustomTemplateFileSystem(),
		newFakeConfigTemplateConfigFileSystem(),
		&fakeConfigTemplateParser{},
		&fakeRemoteTemplateFetcher{},
		&fakeRemoteTemplateBundleReader{},
		newFakeHybridValidator(),
	)

	_, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     "mismatch",
		TemplateName: "feature",
		Title:        "Title",
		Summary:      "Summary",
		Type:         "bugfix",
	})
	if err == nil || !strings.Contains(err.Error(), "hybrid type mismatch") {
		t.Fatalf("Execute() error = %v, want type mismatch", err)
	}
	assertNoGenerationWrites(t, fileSystem)
	if fileSystem.directories[openspecChangesDirectory+"/mismatch"] {
		t.Fatalf("change directory was created")
	}
}

func TestGenerateHybridChangeSkipsExistingFilesAndWritesOnlyOpenSpecFiles(t *testing.T) {
	changeID := "existing-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	fileSystem.files[changePath+"/proposal.md"] = "existing proposal"
	useCase := newHybridGenerateChange(
		fileSystem,
		newFakeHybridTemplateContent(),
		newFakeCustomTemplateFileSystem(),
		newFakeConfigTemplateConfigFileSystem(),
		&fakeConfigTemplateParser{},
		&fakeRemoteTemplateFetcher{},
		&fakeRemoteTemplateBundleReader{},
		newFakeHybridValidator(),
	)

	result, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     changeID,
		TemplateName: "feature",
		Title:        "Title",
		Summary:      "Summary",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertStringSlicesEqual(t, result.SkippedExistingFiles(), []string{"proposal.md"})
	if fileSystem.files[changePath+"/proposal.md"] != "existing proposal" {
		t.Fatalf("proposal.md was overwritten")
	}
	for _, writtenFile := range fileSystem.writtenFiles {
		if !strings.HasPrefix(writtenFile, changePath+"/") {
			t.Fatalf("write target %q is outside %q", writtenFile, changePath)
		}
	}
	assertPathNotWritten(t, fileSystem, "internal/app.go")
	assertPathNotWritten(t, fileSystem, "docs/usage.md")
	assertPathNotWritten(t, fileSystem, ".github/workflows/ci.yml")
	assertPathNotWritten(t, fileSystem, ".git/config")
	assertPathNotWritten(t, fileSystem, "agent-prompts/codex.md")
	assertPathNotWritten(t, fileSystem, "archives/existing-change")
}

func TestGenerateHybridChangeRunsValidationAfterSkipOnlyCompletion(t *testing.T) {
	changeID := "skip-only-change"
	changePath := openspecChangesDirectory + "/" + changeID
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	fileSystem.directories[changePath] = true
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[changePath+"/"+requiredFile] = "preserved " + requiredFile
	}
	validator := newFakeHybridValidator()
	useCase := newHybridGenerateChange(
		fileSystem,
		newFakeHybridTemplateContent(),
		newFakeCustomTemplateFileSystem(),
		newFakeConfigTemplateConfigFileSystem(),
		&fakeConfigTemplateParser{},
		&fakeRemoteTemplateFetcher{},
		&fakeRemoteTemplateBundleReader{},
		validator,
	)

	result, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:  "/project",
		ChangeID:     changeID,
		TemplateName: "feature",
		Title:        "Title",
		Summary:      "Summary",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertStringSlicesEqual(t, result.CreatedFiles(), nil)
	assertStringSlicesEqual(t, result.SkippedExistingFiles(), domain.RequiredOpenSpecChangeFiles())
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + requiredFile
		if fileSystem.files[path] != "preserved "+requiredFile {
			t.Fatalf("%s = %q, want preserved content", path, fileSystem.files[path])
		}
	}
	if validator.calls != 1 || validator.inputs[0].ChangeID != changeID {
		t.Fatalf("validator calls = %d inputs = %+v, want one skip-only validation", validator.calls, validator.inputs)
	}
}

func TestGenerateHybridChangeValidationResultsAreReturned(t *testing.T) {
	tests := []struct {
		name   string
		result domain.ValidationResult
		status domain.ValidationStatus
	}{
		{
			name: "warnings",
			result: domain.NewValidationResult("validation-change", openspecChangesDirectory+"/validation-change", domain.RequiredOpenSpecChangeFiles(), []domain.ValidationFinding{{
				Severity: domain.ValidationFindingSeverityWarning,
				Code:     domain.ValidationFindingCodeRisksMitigationMissing,
				Message:  "warning",
			}}),
			status: domain.ValidationStatusValid,
		},
		{
			name: "errors",
			result: domain.NewValidationResult("validation-change", openspecChangesDirectory+"/validation-change", domain.RequiredOpenSpecChangeFiles(), []domain.ValidationFinding{{
				Severity: domain.ValidationFindingSeverityError,
				Code:     domain.ValidationFindingCodeTasksCheckboxMissing,
				Message:  "error",
			}}),
			status: domain.ValidationStatusInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			validator := &fakeHybridValidator{result: test.result}
			useCase := newHybridGenerateChange(
				fileSystem,
				newFakeHybridTemplateContent(),
				newFakeCustomTemplateFileSystem(),
				newFakeConfigTemplateConfigFileSystem(),
				&fakeConfigTemplateParser{},
				&fakeRemoteTemplateFetcher{},
				&fakeRemoteTemplateBundleReader{},
				validator,
			)

			result, err := useCase.Execute(GenerateHybridChangeInput{
				ProjectRoot:  "/project",
				ChangeID:     "validation-change",
				TemplateName: "feature",
				Title:        "Title",
				Summary:      "Summary",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			validationResult, ok := result.ValidationResult()
			if !ok || validationResult.Status != test.status {
				t.Fatalf("ValidationResult() = %+v, %t; want status %s", validationResult, ok, test.status)
			}
		})
	}
}

func TestGenerateHybridChangeNetworkOnlyForSelectedRemoteAlias(t *testing.T) {
	fileSystem := newFakeGenerationFileSystem()
	seedGenerationOpenSpecProject(fileSystem)
	customTemplateFiles := newFakeCustomTemplateFileSystem()
	seedHybridCustomTemplate(customTemplateFiles, "api-feature")
	configFileSystem := newFakeConfigTemplateConfigFileSystem()
	seedConfigTemplateConfig(configFileSystem)
	parser := &fakeConfigTemplateParser{config: configWithTemplateAlias(t, "default-feature", "builtin", "feature")}
	fetcher := &fakeRemoteTemplateFetcher{}
	reader := &fakeRemoteTemplateBundleReader{bundle: mustRemoteTemplateBundle(t, hybridRemoteFiles("remote"))}
	useCase := newHybridGenerateChange(fileSystem, newFakeHybridTemplateContent(), customTemplateFiles, configFileSystem, parser, fetcher, reader, newFakeHybridValidator())

	requests := []GenerateHybridChangeInput{
		{ProjectRoot: "/project", ChangeID: "hybrid-builtin", TemplateName: "feature", Title: "Title", Summary: "Summary"},
		{ProjectRoot: "/project", ChangeID: "hybrid-custom", CustomTemplateName: "api-feature", Title: "Title", Summary: "Summary"},
		{ProjectRoot: "/project", ChangeID: "hybrid-config-builtin", ConfigTemplateAlias: "default-feature", Title: "Title", Summary: "Summary"},
	}
	for _, request := range requests {
		if _, err := useCase.Execute(request); err != nil {
			t.Fatalf("Execute(%s) error = %v", request.ChangeID, err)
		}
	}

	parser.config = configWithTemplateAlias(t, "api-feature", "custom", "api-feature")
	if _, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "hybrid-config-custom",
		ConfigTemplateAlias: "api-feature",
		Title:               "Title",
		Summary:             "Summary",
	}); err != nil {
		t.Fatalf("Execute(config custom) error = %v", err)
	}

	if fetcher.calls != 0 || reader.calls != 0 {
		t.Fatalf("remote calls for non-remote sources: fetcher=%d reader=%d, want 0", fetcher.calls, reader.calls)
	}

	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	parser.config = configWithRemoteTemplateAlias(t, "service-feature", "https://example.com/templates/service-feature.zip", checksum, "zip")
	fetcher.result = domain.NewRemoteTemplateFetchResult(200, downloadedBytes)
	if _, err := useCase.Execute(GenerateHybridChangeInput{
		ProjectRoot:         "/project",
		ChangeID:            "hybrid-config-remote",
		ConfigTemplateAlias: "service-feature",
		Title:               "Title",
		Summary:             "Summary",
	}); err != nil {
		t.Fatalf("Execute(config remote) error = %v", err)
	}
	if fetcher.calls != 1 || reader.calls != 1 {
		t.Fatalf("remote calls = fetcher %d reader %d, want 1 each", fetcher.calls, reader.calls)
	}
}

func TestGenerateHybridChangeRemoteFailuresWriteNothing(t *testing.T) {
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	tests := []struct {
		name       string
		fetchBytes []byte
		readerErr  error
		want       string
	}{
		{name: "checksum mismatch", fetchBytes: []byte("different"), want: "remote template checksum mismatch"},
		{name: "archive safety", fetchBytes: downloadedBytes, readerErr: errors.New("remote template archive contains unsupported file: README.md"), want: "remote template archive contains unsupported file"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeGenerationFileSystem()
			seedGenerationOpenSpecProject(fileSystem)
			configFileSystem := newFakeConfigTemplateConfigFileSystem()
			seedConfigTemplateConfig(configFileSystem)
			parser := &fakeConfigTemplateParser{
				config: configWithRemoteTemplateAlias(t, "service-feature", "https://example.com/templates/service-feature.zip", checksum, "zip"),
			}
			fetcher := &fakeRemoteTemplateFetcher{result: domain.NewRemoteTemplateFetchResult(200, test.fetchBytes)}
			reader := &fakeRemoteTemplateBundleReader{bundle: mustRemoteTemplateBundle(t, hybridRemoteFiles("remote")), err: test.readerErr}
			useCase := newHybridGenerateChange(fileSystem, newFakeHybridTemplateContent(), newFakeCustomTemplateFileSystem(), configFileSystem, parser, fetcher, reader, newFakeHybridValidator())

			_, err := useCase.Execute(GenerateHybridChangeInput{
				ProjectRoot:         "/project",
				ChangeID:            "remote-failure",
				ConfigTemplateAlias: "service-feature",
				Title:               "Title",
				Summary:             "Summary",
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %v, want %q", err, test.want)
			}
			assertNoGenerationWrites(t, fileSystem)
			if fileSystem.directories[openspecChangesDirectory+"/remote-failure"] {
				t.Fatalf("change directory was created")
			}
		})
	}
}

func newHybridGenerateChange(
	fileSystem *fakeGenerationFileSystem,
	templateContent *fakeHybridTemplateContent,
	customTemplateFiles *fakeCustomTemplateFileSystem,
	configFileSystem *fakeConfigTemplateConfigFileSystem,
	parser *fakeConfigTemplateParser,
	fetcher *fakeRemoteTemplateFetcher,
	reader *fakeRemoteTemplateBundleReader,
	validator hybridChangeValidator,
) *GenerateHybridChange {
	return NewGenerateHybridChange(
		fileSystem,
		templateContent,
		customTemplateFiles,
		configFileSystem,
		parser,
		fetcher,
		reader,
		validator,
	)
}

func assertHybridResultBasics(
	t *testing.T,
	result domain.HybridGenerationResult,
	changeID string,
	selectedKind domain.HybridSelectedSourceKind,
	selectedName string,
	resolvedKind domain.HybridResolvedSourceKind,
	resolvedName string,
) {
	t.Helper()

	if result.ChangeID != changeID || result.Mode != domain.HybridMode || result.ChangePath != openspecChangesDirectory+"/"+changeID {
		t.Fatalf("result basics = %+v, want change %s hybrid path", result, changeID)
	}
	if result.SelectedSourceKind != selectedKind || result.SelectedSourceName != selectedName {
		t.Fatalf("selected source = %s/%s, want %s/%s", result.SelectedSourceKind, result.SelectedSourceName, selectedKind, selectedName)
	}
	if result.ResolvedSourceKind != resolvedKind || result.ResolvedSourceName != resolvedName {
		t.Fatalf("resolved source = %s/%s, want %s/%s", result.ResolvedSourceKind, result.ResolvedSourceName, resolvedKind, resolvedName)
	}
}

func assertHybridFileContent(t *testing.T, fileSystem *fakeGenerationFileSystem, changeID string, file string, want string) {
	t.Helper()

	path := openspecChangesDirectory + "/" + changeID + "/" + file
	if fileSystem.files[path] != want {
		t.Fatalf("file %q content = %q, want %q", path, fileSystem.files[path], want)
	}
}

func seedHybridCustomTemplate(fileSystem *fakeCustomTemplateFileSystem, templateName string) {
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[customTemplatesDirectory+"/"+templateName+"/"+requiredFile] = "custom:" + requiredFile + ":{{change_id}}:{{title}}:{{summary}}:{{type}}:{{unknown}}"
	}
	fileSystem.directories[customTemplatesDirectory+"/"+templateName] = true
}

func hybridRemoteFiles(prefix string) map[string]string {
	files := make(map[string]string)
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		files[requiredFile] = prefix + ":" + requiredFile + ":{{change_id}}:{{title}}:{{summary}}:{{type}}:{{unknown}}"
	}
	return files
}

type fakeHybridTemplateContent struct {
	requests []templateContentRequest
}

func newFakeHybridTemplateContent() *fakeHybridTemplateContent {
	return &fakeHybridTemplateContent{}
}

func (content *fakeHybridTemplateContent) ContentFor(templateName domain.TemplateName, relativePath string) (string, error) {
	content.requests = append(content.requests, templateContentRequest{
		templateName: templateName,
		relativePath: relativePath,
	})
	return "hybrid:" + string(templateName) + ":" + relativePath + ":{{change_id}}:{{title}}:{{summary}}:{{type}}:{{unknown}}", nil
}

type fakeHybridValidator struct {
	result domain.ValidationResult
	err    error
	calls  int
	inputs []ValidateChangeInput
}

func newFakeHybridValidator() *fakeHybridValidator {
	return &fakeHybridValidator{
		result: domain.NewValidationResult("change", openspecChangesDirectory+"/change", domain.RequiredOpenSpecChangeFiles(), nil),
	}
}

func (validator *fakeHybridValidator) Execute(input ValidateChangeInput) (domain.ValidationResult, error) {
	validator.calls++
	validator.inputs = append(validator.inputs, input)
	if validator.err != nil {
		return domain.ValidationResult{}, validator.err
	}
	if validator.result.ChangeID == "change" {
		return domain.NewValidationResult(input.ChangeID, openspecChangesDirectory+"/"+input.ChangeID, domain.RequiredOpenSpecChangeFiles(), nil), nil
	}
	return validator.result, nil
}
