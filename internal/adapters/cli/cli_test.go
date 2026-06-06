package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/platform/version"
)

func TestExecuteInitFirstAndSecondRun(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	if err := execute([]string{"init"}, &output); err != nil {
		t.Fatalf("execute(init) error = %v", err)
	}

	firstRunOutput := output.String()
	if !strings.Contains(firstRunOutput, "SpecHarbor initialized") {
		t.Fatalf("init output = %q, want initialized message", firstRunOutput)
	}
	if !strings.Contains(firstRunOutput, "Created: 13") {
		t.Fatalf("init output = %q, want created count", firstRunOutput)
	}
	if !strings.Contains(firstRunOutput, "Skipped existing: 0") {
		t.Fatalf("init output = %q, want skipped count", firstRunOutput)
	}

	assertPathExists(t, root, "openspec/project.md")
	assertPathExists(t, root, "openspec/specs")
	assertPathExists(t, root, "openspec/changes")
	assertPathExists(t, root, ".specharbor/config.yml")
	assertPathExists(t, root, ".specharbor/rules/global.md")
	assertPathExists(t, root, ".specharbor/rules/spec-author.md")
	assertPathExists(t, root, ".specharbor/rules/implementer.md")
	assertPathExists(t, root, ".specharbor/rules/architecture-reviewer.md")
	assertPathExists(t, root, ".specharbor/rules/test-engineer.md")
	assertPathExists(t, root, ".specharbor/rules/change-reviewer.md")

	projectPath := filepath.Join(root, "openspec", "project.md")
	if err := os.WriteFile(projectPath, []byte("custom project"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	output.Reset()
	if err := execute([]string{"init"}, &output); err != nil {
		t.Fatalf("execute(init) second run error = %v", err)
	}

	secondRunOutput := output.String()
	if !strings.Contains(secondRunOutput, "SpecHarbor already initialized") {
		t.Fatalf("second init output = %q, want already initialized message", secondRunOutput)
	}

	contents, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "custom project" {
		t.Fatalf("project.md contents = %q, want preserved custom content", string(contents))
	}
}

func TestExecuteInitPartialProject(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	if err := os.MkdirAll(filepath.Join(root, "openspec"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	projectPath := filepath.Join(root, "openspec", "project.md")
	if err := os.WriteFile(projectPath, []byte("existing project"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var output bytes.Buffer
	if err := execute([]string{"init"}, &output); err != nil {
		t.Fatalf("execute(init) error = %v", err)
	}

	initOutput := output.String()
	if !strings.Contains(initOutput, "SpecHarbor initialized") {
		t.Fatalf("init output = %q, want initialized message", initOutput)
	}
	if !strings.Contains(initOutput, "Skipped existing: 2") {
		t.Fatalf("init output = %q, want skipped count for partial project", initOutput)
	}

	contents, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "existing project" {
		t.Fatalf("project.md contents = %q, want preserved existing content", string(contents))
	}
	assertPathExists(t, root, ".specharbor/config.yml")
	assertPathExists(t, root, ".specharbor/rules/change-reviewer.md")
}

func TestExecuteGenerateBlankPrintsCreatedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	if err := execute([]string{"generate", "implement-generation-foundation", "--blank"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor blank change generated.
Change: implement-generation-foundation
Path: openspec/changes/implement-generation-foundation
Directory: created
Created files: 5
Skipped existing files: 0

Created:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := filepath.Join(root, "openspec", "changes", "implement-generation-foundation", requiredFile)
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if strings.TrimSpace(string(contents)) == "" {
			t.Fatalf("generated file %q is empty", requiredFile)
		}
		if requiredFile == "tasks.md" {
			assertUncheckedTasksOnly(t, string(contents))
		}
	}
}

func TestExecuteGenerateBlankAcceptsFlagBeforeChangeID(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	if err := execute([]string{"generate", "--blank", "order-independent"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	if !strings.Contains(output.String(), "Change: order-independent") {
		t.Fatalf("generate output = %q, want change id", output.String())
	}
}

func TestExecuteGenerateBlankPrintsSkippedExistingReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createOpenSpecChange(t, root, "implement-generation-foundation", domain.RequiredOpenSpecChangeFiles())

	var output bytes.Buffer
	if err := execute([]string{"generate", "implement-generation-foundation", "--blank"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor blank change generated.
Change: implement-generation-foundation
Path: openspec/changes/implement-generation-foundation
Directory: existing
Created files: 0
Skipped existing files: 5

Skipped existing:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := filepath.Join(root, "openspec", "changes", "implement-generation-foundation", requiredFile)
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if string(contents) != requiredFile {
			t.Fatalf("%s contents = %q, want preserved %q", requiredFile, string(contents), requiredFile)
		}
	}
}

func TestExecuteGenerateRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing change id",
			args: []string{"generate"},
			want: "change id is required",
		},
		{
			name: "missing blank flag",
			args: []string{"generate", "change"},
			want: "blank generation flag is required",
		},
		{
			name: "blank without change id",
			args: []string{"generate", "--blank"},
			want: "change id is required",
		},
		{
			name: "duplicate blank flag",
			args: []string{"generate", "change", "--blank", "--blank"},
			want: "blank generation flag specified more than once",
		},
		{
			name: "unsupported flag",
			args: []string{"generate", "change", "--guided"},
			want: "unsupported flag: --guided",
		},
		{
			name: "unsupported template flag",
			args: []string{"generate", "change", "--template"},
			want: "unsupported flag: --template",
		},
		{
			name: "unsupported ai flag",
			args: []string{"generate", "change", "--ai"},
			want: "unsupported flag: --ai",
		},
		{
			name: "unsupported agent flag",
			args: []string{"generate", "change", "--agent"},
			want: "unsupported flag: --agent",
		},
		{
			name: "unsupported hybrid flag",
			args: []string{"generate", "change", "--hybrid"},
			want: "unsupported flag: --hybrid",
		},
		{
			name: "unsupported format flag",
			args: []string{"generate", "change", "--format", "json"},
			want: "unsupported flag: --format",
		},
		{
			name: "unsupported force flag",
			args: []string{"generate", "change", "--force"},
			want: "unsupported flag: --force",
		},
		{
			name: "extra argument",
			args: []string{"generate", "change", "--blank", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "unsafe traversal change id",
			args: []string{"generate", "../unsafe", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe absolute change id",
			args: []string{"generate", "/unsafe", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe slash change id",
			args: []string{"generate", "bad/id", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe backslash change id",
			args: []string{"generate", `bad\id`, "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe colon change id",
			args: []string{"generate", "bad:id", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe dot change id",
			args: []string{"generate", ".", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe dotdot change id",
			args: []string{"generate", "..", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "leading dash change id",
			args: []string{"generate", "-bad-id", "--blank"},
			want: "unsupported flag: -bad-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)

			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil {
				t.Fatalf("execute(%v) error = nil, want %q", test.args, test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(%v) error = %q, want %q", test.args, err.Error(), test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(%v) output = %q, want empty output", test.args, output.String())
			}
			assertPathDoesNotExist(t, root, "openspec")
		})
	}
}

func TestExecuteGenerateRejectsMissingOpenSpecProjectWithoutCreatingStructure(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	err := execute([]string{"generate", "missing-project", "--blank"}, &output)
	if err == nil {
		t.Fatalf("execute(generate) error = nil, want missing project structure error")
	}
	for _, want := range []string{"OpenSpec project structure is missing", "specharbor init"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("execute(generate) error = %q, want to contain %q", err.Error(), want)
		}
	}
	if output.String() != "" {
		t.Fatalf("execute(generate) output = %q, want empty output", output.String())
	}
	assertPathDoesNotExist(t, root, "openspec")
}

func TestExecutePromptPrintsRenderedPromptOnly(t *testing.T) {
	t.Chdir(t.TempDir())

	var output bytes.Buffer
	if err := execute([]string{"prompt", "implement-prompt-command", "--role", "implementer"}, &output); err != nil {
		t.Fatalf("execute(prompt) error = %v", err)
	}

	promptOutput := output.String()
	if !strings.HasPrefix(promptOutput, "# Implementer Agent\n") {
		t.Fatalf("prompt output = %q, want implementer prompt only", promptOutput)
	}
	if !strings.Contains(promptOutput, "openspec/changes/implement-prompt-command/") {
		t.Fatalf("prompt output = %q, want rendered change id", promptOutput)
	}
	if strings.Contains(promptOutput, "{{change_id}}") {
		t.Fatalf("prompt output = %q, want no raw change_id placeholder", promptOutput)
	}
	if strings.Contains(promptOutput, "{{task}}") {
		t.Fatalf("prompt output = %q, want no raw task placeholder", promptOutput)
	}
	if strings.Contains(promptOutput, "not implemented") {
		t.Fatalf("prompt output = %q, want rendered prompt instead of placeholder", promptOutput)
	}
}

func TestExecutePromptRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing change id",
			args: []string{"prompt"},
			want: "change id is required",
		},
		{
			name: "missing role",
			args: []string{"prompt", "implement-prompt-command"},
			want: "prompt role is required",
		},
		{
			name: "missing role value",
			args: []string{"prompt", "implement-prompt-command", "--role"},
			want: "prompt role value is required",
		},
		{
			name: "unsupported role",
			args: []string{"prompt", "implement-prompt-command", "--role", "unknown"},
			want: "unsupported prompt role: unknown",
		},
		{
			name: "unsupported flag",
			args: []string{"prompt", "implement-prompt-command", "--target", "codex"},
			want: "unsupported flag: --target",
		},
		{
			name: "extra argument",
			args: []string{"prompt", "implement-prompt-command", "--role", "implementer", "extra"},
			want: "unexpected argument: extra",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil {
				t.Fatalf("execute(%v) error = nil, want %q", test.args, test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(%v) error = %q, want %q", test.args, err.Error(), test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(%v) output = %q, want empty output", test.args, output.String())
			}
		})
	}
}

func TestExecuteValidatePrintsValidReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createOpenSpecChange(t, root, "implement-validation-foundation", domain.RequiredOpenSpecChangeFiles())

	var output bytes.Buffer
	if err := execute([]string{"validate", "implement-validation-foundation"}, &output); err != nil {
		t.Fatalf("execute(validate) error = %v", err)
	}

	want := `SpecHarbor change is valid.
Change: implement-validation-foundation
Checked path: openspec/changes/implement-validation-foundation
Required files: 5
Findings: 0
`
	if output.String() != want {
		t.Fatalf("validate output = %q, want %q", output.String(), want)
	}
}

func TestExecuteValidatePrintsInvalidReportForMissingRequiredFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createOpenSpecChange(t, root, "implement-validation-foundation", []string{
		"design.md",
		"tasks.md",
		"acceptance-criteria.md",
	})

	var output bytes.Buffer
	err := execute([]string{"validate", "implement-validation-foundation"}, &output)
	assertExitCode(t, err, 1)

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is invalid.",
		"Change: implement-validation-foundation",
		"Checked path: openspec/changes/implement-validation-foundation",
		"Findings:",
		"- [error] required_file_missing: Missing required file: proposal.md",
		"- [error] required_file_missing: Missing required file: risks.md",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
}

func TestExecuteValidatePrintsInvalidReportForMissingChangeDirectory(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	err := execute([]string{"validate", "missing-change"}, &output)
	assertExitCode(t, err, 1)

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is invalid.",
		"Change: missing-change",
		"Checked path: openspec/changes/missing-change",
		"- [error] change_directory_missing: Missing change directory: openspec/changes/missing-change",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
}

func TestExecuteValidatePrintsInvalidReportForMissingOpenSpecProjectStructure(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	err := execute([]string{"validate", "change"}, &output)
	assertExitCode(t, err, 1)

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is invalid.",
		"Change: change",
		"Checked path: openspec/changes/change",
		"- [error] project_root_unavailable: OpenSpec project structure is unavailable.",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
	if strings.Contains(validateOutput, "change_directory_missing") {
		t.Fatalf("validate output = %q, want no change directory finding when project structure is unavailable", validateOutput)
	}
}

func TestExecuteValidateRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing change id",
			args: []string{"validate"},
			want: "change id is required",
		},
		{
			name: "unsupported flag",
			args: []string{"validate", "--format", "json"},
			want: "unsupported flag: --format",
		},
		{
			name: "extra argument",
			args: []string{"validate", "change", "extra"},
			want: "unexpected argument: extra",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil {
				t.Fatalf("execute(%v) error = nil, want %q", test.args, test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(%v) error = %q, want %q", test.args, err.Error(), test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(%v) output = %q, want empty output", test.args, output.String())
			}
		})
	}
}

func TestExecutePreservesHelpVersionAndUnknownCommandBehavior(t *testing.T) {
	var output bytes.Buffer
	if err := execute([]string{"help"}, &output); err != nil {
		t.Fatalf("execute(help) error = %v", err)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("help output = %q, want Usage", output.String())
	}

	output.Reset()
	if err := execute([]string{"version"}, &output); err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}
	if output.String() != version.Version+"\n" {
		t.Fatalf("version output = %q, want %q", output.String(), version.Version+"\n")
	}

	output.Reset()
	err := execute([]string{"missing"}, &output)
	if err == nil {
		t.Fatalf("execute(missing) error = nil, want error")
	}
	if err.Error() != "unknown command: missing" {
		t.Fatalf("unknown command error = %q, want %q", err.Error(), "unknown command: missing")
	}
}

func assertExitCode(t *testing.T, err error, code int) {
	t.Helper()

	if err == nil {
		t.Fatalf("execute() error = nil, want exit code %d", code)
	}

	var exitError ExitError
	if !errors.As(err, &exitError) {
		t.Fatalf("execute() error = %T %v, want ExitError", err, err)
	}
	if exitError.Code != code {
		t.Fatalf("exit code = %d, want %d", exitError.Code, code)
	}
}

func createOpenSpecProject(t *testing.T, root string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "openspec", "project.md"), []byte("project"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func createOpenSpecChange(t *testing.T, root string, changeID string, files []string) {
	t.Helper()

	changeDirectory := filepath.Join(root, "openspec", "changes", changeID)
	if err := os.MkdirAll(changeDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(changeDirectory, file), []byte(file), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}
}

func assertPathExists(t *testing.T, root string, relativePath string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(relativePath))); err != nil {
		t.Fatalf("expected %q to exist: %v", relativePath, err)
	}
}

func assertPathDoesNotExist(t *testing.T, root string, relativePath string) {
	t.Helper()

	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err == nil {
		t.Fatalf("expected %q not to exist", relativePath)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("checking %q existence returned unexpected error: %v", relativePath, err)
	}
}

func assertUncheckedTasksOnly(t *testing.T, contents string) {
	t.Helper()

	checkboxes := 0
	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- [") {
			continue
		}

		checkboxes++
		if !strings.HasPrefix(line, "- [ ]") {
			t.Fatalf("tasks.md checkbox line = %q, want unchecked task", line)
		}
	}
	if checkboxes == 0 {
		t.Fatalf("tasks.md content = %q, want unchecked tasks", contents)
	}
}
