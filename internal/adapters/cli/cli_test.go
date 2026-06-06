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
