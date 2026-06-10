package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateAIAssistedChangeWritesRequiredFilesAndRunsValidation(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	seedAIAssistedOpenSpecProject(fileSystem)
	fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
	validator := newFakeAIAssistedValidator()

	result, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "ai-change",
		SourcePath:  "agent-output.txt",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	changePath := openspecChangesDirectory + "/ai-change"
	if result.ChangeID != "ai-change" || result.SourcePath != "agent-output.txt" || result.TargetPath != changePath {
		t.Fatalf("result = %+v, want change/source/path fields", result)
	}
	if !result.ChangeDirectoryCreated {
		t.Fatalf("ChangeDirectoryCreated = false, want true")
	}
	assertStringSlicesEqual(t, result.GeneratedFiles(), domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.SkippedFiles(), nil)
	assertStringSlicesEqual(t, result.OverwrittenFiles(), nil)
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + fileName
		if !strings.Contains(fileSystem.files[path], "#") {
			t.Fatalf("file %s content = %q, want generated markdown", path, fileSystem.files[path])
		}
	}
	if validator.calls != 1 || validator.inputs[0].ChangeID != "ai-change" {
		t.Fatalf("validator calls = %d inputs = %+v, want one ai-change validation", validator.calls, validator.inputs)
	}
	validationResult, ok := result.ValidationResult()
	if !ok || validationResult.Status != domain.ValidationStatusValid {
		t.Fatalf("ValidationResult() = %+v, %v; want valid", validationResult, ok)
	}
}

func TestGenerateAIAssistedChangeRejectsMalformedSourcesWithoutWrites(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   domain.AIOutputParseFindingCode
	}{
		{name: "malformed", source: "---FILE:proposal.md---\n# Proposal\n---END FILE---", code: domain.AIOutputParseFindingCodeMalformedFileBlock},
		{name: "unknown", source: aiBlock("notes.md", "# Notes\n\nContent."), code: domain.AIOutputParseFindingCodeUnknownFileBlock},
		{name: "duplicate", source: validAIAssistedSource() + "\n" + aiBlock("proposal.md", "# Proposal\n\nDuplicate."), code: domain.AIOutputParseFindingCodeDuplicateFileBlock},
		{name: "missing", source: strings.Replace(validAIAssistedSource(), aiBlock("risks.md", validAIAssistedContent("risks.md")), "", 1), code: domain.AIOutputParseFindingCodeMissingFileBlock},
		{name: "empty", source: strings.Replace(validAIAssistedSource(), aiBlock("tasks.md", validAIAssistedContent("tasks.md")), aiBlock("tasks.md", " "), 1), code: domain.AIOutputParseFindingCodeEmptyFileContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeAIAssistedFileSystem()
			seedAIAssistedOpenSpecProject(fileSystem)
			fileSystem.sources["agent-output.txt"] = test.source
			validator := newFakeAIAssistedValidator()

			_, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "bad-ai-change",
				SourcePath:  "agent-output.txt",
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want parse failure")
			}
			var parseFailure *AIAssistedParseFailure
			if !errors.As(err, &parseFailure) {
				t.Fatalf("Execute() error = %T %v, want AIAssistedParseFailure", err, err)
			}
			if !hasAIParseCode(parseFailure.ParseResult, test.code) {
				t.Fatalf("parse findings = %v, want %s", parseFailure.ParseResult.Findings(), test.code)
			}
			assertStringSlicesEqual(t, fileSystem.writes, nil)
			if len(fileSystem.createdDirectories) != 0 {
				t.Fatalf("created directories = %v, want none", fileSystem.createdDirectories)
			}
			if validator.calls != 0 {
				t.Fatalf("validator calls = %d, want 0", validator.calls)
			}
		})
	}
}

func TestGenerateAIAssistedChangeSkipsExistingFilesByDefault(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	seedAIAssistedOpenSpecProject(fileSystem)
	fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
	changePath := openspecChangesDirectory + "/existing-ai-change"
	fileSystem.directories[changePath] = true
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[changePath+"/"+fileName] = "existing:" + fileName
	}

	result, err := NewGenerateAIAssistedChange(fileSystem, newFakeAIAssistedValidator()).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "existing-ai-change",
		SourcePath:  "agent-output.txt",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	assertStringSlicesEqual(t, result.GeneratedFiles(), nil)
	assertStringSlicesEqual(t, result.SkippedFiles(), domain.RequiredOpenSpecChangeFiles())
	assertStringSlicesEqual(t, result.OverwrittenFiles(), nil)
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + fileName
		if fileSystem.files[path] != "existing:"+fileName {
			t.Fatalf("%s = %q, want preserved existing content", path, fileSystem.files[path])
		}
	}
}

func TestGenerateAIAssistedChangeOverwritesOnlyWithExplicitFlag(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	seedAIAssistedOpenSpecProject(fileSystem)
	fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
	changePath := openspecChangesDirectory + "/overwrite-ai-change"
	fileSystem.directories[changePath] = true
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		fileSystem.files[changePath+"/"+fileName] = "existing:" + fileName
	}

	result, err := NewGenerateAIAssistedChange(fileSystem, newFakeAIAssistedValidator()).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "overwrite-ai-change",
		SourcePath:  "agent-output.txt",
		Overwrite:   true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !result.Overwrite {
		t.Fatalf("Overwrite = false, want true")
	}
	assertStringSlicesEqual(t, result.GeneratedFiles(), nil)
	assertStringSlicesEqual(t, result.SkippedFiles(), nil)
	assertStringSlicesEqual(t, result.OverwrittenFiles(), domain.RequiredOpenSpecChangeFiles())
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		path := changePath + "/" + fileName
		if strings.HasPrefix(fileSystem.files[path], "existing:") {
			t.Fatalf("%s was not overwritten", path)
		}
	}
}

func TestGenerateAIAssistedChangeRejectsUnsafeTargetsBeforeWrites(t *testing.T) {
	tests := []struct {
		name      string
		overwrite bool
	}{
		{name: "default skip mode", overwrite: false},
		{name: "overwrite mode", overwrite: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeAIAssistedFileSystem()
			seedAIAssistedOpenSpecProject(fileSystem)
			fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
			changePath := openspecChangesDirectory + "/unsafe-ai-change"
			fileSystem.directories[changePath] = true
			fileSystem.files[changePath+"/proposal.md"] = "symlink placeholder"
			fileSystem.files["outside/proposal.md"] = "outside original"
			fileSystem.safeTargetErrors[changePath+"/proposal.md"] = errors.New("symlink target paths are not allowed for generated OpenSpec files: " + changePath + "/proposal.md")
			validator := newFakeAIAssistedValidator()

			_, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "unsafe-ai-change",
				SourcePath:  "agent-output.txt",
				Overwrite:   test.overwrite,
			})
			if err == nil || !strings.Contains(err.Error(), "symlink target paths are not allowed for generated OpenSpec files") {
				t.Fatalf("Execute() error = %v, want symlink target rejection", err)
			}
			assertStringSlicesEqual(t, fileSystem.writes, nil)
			assertStringSlicesEqual(t, fileSystem.overwrites, nil)
			if fileSystem.files["outside/proposal.md"] != "outside original" {
				t.Fatalf("external target = %q, want unchanged", fileSystem.files["outside/proposal.md"])
			}
			if validator.calls != 0 {
				t.Fatalf("validator calls = %d, want no validation after symlink safety failure", validator.calls)
			}
		})
	}
}

func TestGenerateAIAssistedChangeRechecksTargetsImmediatelyBeforeWriting(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	seedAIAssistedOpenSpecProject(fileSystem)
	fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
	changePath := openspecChangesDirectory + "/prewrite-ai-change"
	relativePath := changePath + "/proposal.md"
	fileSystem.safeTargetErrorsAfterChecks[relativePath] = 2
	fileSystem.safeTargetErrors[relativePath] = errors.New("symlink target paths are not allowed for generated OpenSpec files: " + relativePath)
	validator := newFakeAIAssistedValidator()

	_, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "prewrite-ai-change",
		SourcePath:  "agent-output.txt",
	})
	if err == nil || !strings.Contains(err.Error(), "symlink target paths are not allowed for generated OpenSpec files") {
		t.Fatalf("Execute() error = %v, want immediate pre-write symlink rejection", err)
	}
	assertStringSlicesEqual(t, fileSystem.writes, nil)
	if validator.calls != 0 {
		t.Fatalf("validator calls = %d, want no validation after immediate safety failure", validator.calls)
	}
}

func TestGenerateAIAssistedChangePreflightFailuresWriteNothing(t *testing.T) {
	tests := []struct {
		name  string
		setup func(fileSystem *fakeAIAssistedFileSystem, changePath string)
		want  string
	}{
		{
			name: "target path is a file",
			setup: func(fileSystem *fakeAIAssistedFileSystem, changePath string) {
				fileSystem.files[changePath] = "not a directory"
			},
			want: "target path exists and is not a directory",
		},
		{
			name: "target required path is a directory",
			setup: func(fileSystem *fakeAIAssistedFileSystem, changePath string) {
				fileSystem.directories[changePath] = true
				fileSystem.directories[changePath+"/proposal.md"] = true
			},
			want: "target file path exists and is not a file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeAIAssistedFileSystem()
			seedAIAssistedOpenSpecProject(fileSystem)
			fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
			changePath := openspecChangesDirectory + "/preflight-ai-change"
			test.setup(fileSystem, changePath)
			validator := newFakeAIAssistedValidator()

			_, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "preflight-ai-change",
				SourcePath:  "agent-output.txt",
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %v, want %q", err, test.want)
			}
			assertStringSlicesEqual(t, fileSystem.writes, nil)
			if validator.calls != 0 {
				t.Fatalf("validator calls = %d, want 0", validator.calls)
			}
		})
	}
}

func TestGenerateAIAssistedChangeRejectsUnsafeChangeIDBeforeFilesystemAccess(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	fileSystem.sources["agent-output.txt"] = validAIAssistedSource()

	_, err := NewGenerateAIAssistedChange(fileSystem, newFakeAIAssistedValidator()).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "../unsafe",
		SourcePath:  "agent-output.txt",
	})
	if err == nil || !strings.Contains(err.Error(), "change id must be a single path segment") {
		t.Fatalf("Execute() error = %v, want unsafe change id", err)
	}
	if fileSystem.operationCount() != 0 || fileSystem.sourceReads != 0 {
		t.Fatalf("filesystem operations = %d sourceReads = %d, want none", fileSystem.operationCount(), fileSystem.sourceReads)
	}
}

func TestGenerateAIAssistedChangeSourceReadErrorWritesNothing(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	seedAIAssistedOpenSpecProject(fileSystem)
	wantErr := errors.New("source unavailable")
	fileSystem.sourceErrors["missing-output.txt"] = wantErr

	_, err := NewGenerateAIAssistedChange(fileSystem, newFakeAIAssistedValidator()).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "source-error",
		SourcePath:  "missing-output.txt",
	})
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want source error %v", err, wantErr)
	}
	assertStringSlicesEqual(t, fileSystem.writes, nil)
}

func TestGenerateAIAssistedChangeReportsRuntimeWriteFailureWithoutRollbackOrValidation(t *testing.T) {
	fileSystem := newFakeAIAssistedFileSystem()
	seedAIAssistedOpenSpecProject(fileSystem)
	fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
	changePath := openspecChangesDirectory + "/write-failure"
	wantErr := errors.New("disk full")
	fileSystem.writeAbsentErrors[changePath+"/design.md"] = wantErr
	validator := newFakeAIAssistedValidator()

	_, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
		ProjectRoot: "/project",
		ChangeID:    "write-failure",
		SourcePath:  "agent-output.txt",
	})
	if err == nil || !errors.Is(err, wantErr) || !strings.Contains(err.Error(), "write file "+changePath+"/design.md") {
		t.Fatalf("Execute() error = %v, want clear design.md write error wrapping %v", err, wantErr)
	}
	if fileSystem.files[changePath+"/proposal.md"] == "" {
		t.Fatalf("proposal.md was rolled back or not written before runtime failure")
	}
	if _, exists := fileSystem.files[changePath+"/design.md"]; exists {
		t.Fatalf("design.md exists after failed write, want absent")
	}
	if validator.calls != 0 {
		t.Fatalf("validator calls = %d, want no validation after runtime write failure", validator.calls)
	}
}

func TestGenerateAIAssistedChangeReturnsValidationWarningsAndErrors(t *testing.T) {
	tests := []struct {
		name   string
		result domain.ValidationResult
		status domain.ValidationStatus
	}{
		{
			name: "warnings",
			result: domain.NewValidationResult("validation-ai", openspecChangesDirectory+"/validation-ai", domain.RequiredOpenSpecChangeFiles(), []domain.ValidationFinding{{
				Severity: domain.ValidationFindingSeverityWarning,
				Code:     domain.ValidationFindingCodeRisksMitigationMissing,
			}}),
			status: domain.ValidationStatusValid,
		},
		{
			name: "errors",
			result: domain.NewValidationResult("validation-ai", openspecChangesDirectory+"/validation-ai", domain.RequiredOpenSpecChangeFiles(), []domain.ValidationFinding{{
				Severity: domain.ValidationFindingSeverityError,
				Code:     domain.ValidationFindingCodeTasksCheckboxMissing,
			}}),
			status: domain.ValidationStatusInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileSystem := newFakeAIAssistedFileSystem()
			seedAIAssistedOpenSpecProject(fileSystem)
			fileSystem.sources["agent-output.txt"] = validAIAssistedSource()
			validator := newFakeAIAssistedValidator()
			validator.result = test.result

			result, err := NewGenerateAIAssistedChange(fileSystem, validator).Execute(GenerateAIAssistedChangeInput{
				ProjectRoot: "/project",
				ChangeID:    "validation-ai",
				SourcePath:  "agent-output.txt",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			validationResult, ok := result.ValidationResult()
			if !ok || validationResult.Status != test.status {
				t.Fatalf("ValidationResult() = %+v, %v; want %s", validationResult, ok, test.status)
			}
		})
	}
}

func seedAIAssistedOpenSpecProject(fileSystem *fakeAIAssistedFileSystem) {
	fileSystem.files[openspecProjectFile] = "project"
	fileSystem.directories[openspecChangesDirectory] = true
}

func validAIAssistedSource() string {
	var blocks []string
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		blocks = append(blocks, aiBlock(fileName, validAIAssistedContent(fileName)))
	}
	return strings.Join(blocks, "\n")
}

func aiBlock(fileName string, content string) string {
	return "---FILE: " + fileName + "---\n" + content + "\n---END FILE---"
}

func validAIAssistedContent(fileName string) string {
	switch fileName {
	case "proposal.md":
		return "# Proposal\n\n## Problem\n\nManual copying is error-prone.\n\n## Goal\n\nImport local AI-authored blocks safely."
	case "design.md":
		return "# Design\n\n## Overview\n\nThe parser validates all blocks first.\n\n## Architecture\n\nThe use case writes approved paths only."
	case "tasks.md":
		return "# Tasks\n\n## Phase 1\n\n- [ ] Parse strict blocks.\n- [ ] Validate generated files."
	case "acceptance-criteria.md":
		return "# Acceptance Criteria\n\n- Valid strict output writes the required files."
	case "risks.md":
		return "# Risks\n\n## Risks\n\n- Untrusted output could be unsafe.\n\n## Mitigations\n\n- Reject unsafe names."
	default:
		return "# " + fileName + "\n\nContent."
	}
}

func hasAIParseCode(result domain.AIOutputParseResult, code domain.AIOutputParseFindingCode) bool {
	for _, finding := range result.Findings() {
		if finding.Code == code {
			return true
		}
	}
	return false
}

type fakeAIAssistedFileSystem struct {
	directories                 map[string]bool
	files                       map[string]string
	sources                     map[string]string
	sourceErrors                map[string]error
	directoryErrors             map[string]error
	fileErrors                  map[string]error
	pathErrors                  map[string]error
	createErrors                map[string]error
	writeAbsentErrors           map[string]error
	writeErrors                 map[string]error
	safeTargetErrors            map[string]error
	safeTargetErrorsAfterChecks map[string]int
	safeTargetChecks            map[string]int
	createdDirectories          []string
	writes                      []string
	overwrites                  []string
	checkedDirectories          []string
	checkedFiles                []string
	checkedPaths                []string
	sourceReads                 int
}

func newFakeAIAssistedFileSystem() *fakeAIAssistedFileSystem {
	return &fakeAIAssistedFileSystem{
		directories:                 make(map[string]bool),
		files:                       make(map[string]string),
		sources:                     make(map[string]string),
		sourceErrors:                make(map[string]error),
		directoryErrors:             make(map[string]error),
		fileErrors:                  make(map[string]error),
		pathErrors:                  make(map[string]error),
		createErrors:                make(map[string]error),
		writeAbsentErrors:           make(map[string]error),
		writeErrors:                 make(map[string]error),
		safeTargetErrors:            make(map[string]error),
		safeTargetErrorsAfterChecks: make(map[string]int),
		safeTargetChecks:            make(map[string]int),
	}
}

func (fileSystem *fakeAIAssistedFileSystem) ReadSourceFile(path string) (string, error) {
	fileSystem.sourceReads++
	if err := fileSystem.sourceErrors[path]; err != nil {
		return "", err
	}
	return fileSystem.sources[path], nil
}

func (fileSystem *fakeAIAssistedFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedDirectories = append(fileSystem.checkedDirectories, relativePath)
	if err := fileSystem.directoryErrors[relativePath]; err != nil {
		return false, err
	}
	return fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeAIAssistedFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedFiles = append(fileSystem.checkedFiles, relativePath)
	if err := fileSystem.fileErrors[relativePath]; err != nil {
		return false, err
	}
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeAIAssistedFileSystem) PathExists(_ string, relativePath string) (bool, error) {
	fileSystem.checkedPaths = append(fileSystem.checkedPaths, relativePath)
	if err := fileSystem.pathErrors[relativePath]; err != nil {
		return false, err
	}
	_, fileExists := fileSystem.files[relativePath]
	return fileExists || fileSystem.directories[relativePath], nil
}

func (fileSystem *fakeAIAssistedFileSystem) CreateDirectory(_ string, relativePath string) error {
	fileSystem.createdDirectories = append(fileSystem.createdDirectories, relativePath)
	if err := fileSystem.createErrors[relativePath]; err != nil {
		return err
	}
	fileSystem.directories[relativePath] = true
	return nil
}

func (fileSystem *fakeAIAssistedFileSystem) EnsureSafeWriteTarget(_ string, relativePath string) error {
	fileSystem.safeTargetChecks[relativePath]++
	if threshold := fileSystem.safeTargetErrorsAfterChecks[relativePath]; threshold > 0 && fileSystem.safeTargetChecks[relativePath] < threshold {
		return nil
	}
	if err := fileSystem.safeTargetErrors[relativePath]; err != nil {
		return err
	}
	return nil
}

func (fileSystem *fakeAIAssistedFileSystem) WriteFileIfAbsent(_ string, relativePath string, contents string) (bool, error) {
	fileSystem.writes = append(fileSystem.writes, relativePath)
	if err := fileSystem.writeAbsentErrors[relativePath]; err != nil {
		return false, err
	}
	if _, exists := fileSystem.files[relativePath]; exists {
		return false, nil
	}
	fileSystem.files[relativePath] = contents
	return true, nil
}

func (fileSystem *fakeAIAssistedFileSystem) WriteFile(_ string, relativePath string, contents string) error {
	fileSystem.overwrites = append(fileSystem.overwrites, relativePath)
	if err := fileSystem.writeErrors[relativePath]; err != nil {
		return err
	}
	fileSystem.files[relativePath] = contents
	return nil
}

func (fileSystem *fakeAIAssistedFileSystem) operationCount() int {
	return len(fileSystem.checkedDirectories) +
		len(fileSystem.checkedFiles) +
		len(fileSystem.checkedPaths) +
		len(fileSystem.createdDirectories) +
		len(fileSystem.writes) +
		len(fileSystem.overwrites) +
		len(fileSystem.safeTargetChecks)
}

type fakeAIAssistedValidator struct {
	result domain.ValidationResult
	err    error
	calls  int
	inputs []ValidateChangeInput
}

func newFakeAIAssistedValidator() *fakeAIAssistedValidator {
	return &fakeAIAssistedValidator{}
}

func (validator *fakeAIAssistedValidator) Execute(input ValidateChangeInput) (domain.ValidationResult, error) {
	validator.calls++
	validator.inputs = append(validator.inputs, input)
	if validator.err != nil {
		return domain.ValidationResult{}, validator.err
	}
	if validator.result.ChangeID != "" {
		return validator.result, nil
	}
	return domain.NewValidationResult(
		input.ChangeID,
		openspecChangesDirectory+"/"+input.ChangeID,
		domain.RequiredOpenSpecChangeFiles(),
		nil,
	), nil
}
