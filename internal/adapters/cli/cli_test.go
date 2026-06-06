package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestExecutePromptPrintsRenderedPromptOnly(t *testing.T) {
	t.Chdir(findProjectRoot(t))

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

func assertPathExists(t *testing.T, root string, relativePath string) {
	t.Helper()

	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(relativePath))); err != nil {
		t.Fatalf("expected %q to exist: %v", relativePath, err)
	}
}

func findProjectRoot(t *testing.T) string {
	t.Helper()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	for {
		templatePath := filepath.Join(root, "agent-prompts", "roles", "implementer.md.tmpl")
		if _, err := os.Stat(templatePath); err == nil {
			return root
		}

		parent := filepath.Dir(root)
		if parent == root {
			t.Fatalf("could not find project root from %s", root)
		}
		root = parent
	}
}
