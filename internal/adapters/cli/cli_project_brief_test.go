package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/platform/version"
)

func TestExecuteBriefRejectsArgumentsAndFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "positional", args: []string{"brief", "extra"}, want: "unexpected argument: extra"},
		{name: "force flag", args: []string{"brief", "--force"}, want: "unsupported flag: --force"},
		{name: "overwrite flag", args: []string{"brief", "--overwrite"}, want: "unsupported flag: --overwrite"},
		{name: "json flag", args: []string{"brief", "--json"}, want: "unsupported flag: --json"},
		{name: "duplicate update flag", args: []string{"brief", "--update", "--update"}, want: "duplicate flag: --update"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := executeWithTerminal(test.args, &output, &cliFakeInteractiveTerminal{isTTY: true, output: &output})
			if err == nil || err.Error() != test.want {
				t.Fatalf("execute(%v) error = %v, want %q", test.args, err, test.want)
			}
			if output.String() != "" {
				t.Fatalf("output = %q, want empty", output.String())
			}
		})
	}
}

func TestExecuteBriefNonTTYFailsWithoutPromptOrWrites(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{
		isTTY:  false,
		output: &output,
		inputs: append(defaultProjectBriefPromptInputs(), "y"),
	}
	err := executeWithTerminal([]string{"brief"}, &output, terminal)
	if err == nil || err.Error() != "brief requires an interactive TTY" {
		t.Fatalf("execute brief non-tty error = %v, want TTY error", err)
	}
	if output.String() != "" {
		t.Fatalf("output = %q, want empty", output.String())
	}
	if terminal.reads != 0 {
		t.Fatalf("terminal reads = %d, want 0", terminal.reads)
	}
	assertPathDoesNotExist(t, root, ".specharbor/project-brief.md")
}

func TestExecuteBriefCreatesProjectBriefAfterConfirmation(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	output, err := executeBrief(t, append(defaultProjectBriefPromptInputs(), "y"))
	if err != nil {
		t.Fatalf("execute brief error = %v\noutput:\n%s", err, output)
	}

	for _, want := range []string{
		"SpecHarbor could not find enough confirmed project context.",
		"What type of project is this?",
		"Other / custom",
		"What should agents do when context is missing?",
		"SpecHarbor will create:",
		".specharbor/project-brief.md",
		"Safety:",
		"- Stack, architecture, and commands come from confirmed answers only.",
		"- Detected context remains separate from user answers.",
		"- Assumptions are not confirmed facts.",
		"Confirm? [y/N]:",
		"SpecHarbor project brief created.",
		"Path: .specharbor/project-brief.md",
		"File written: yes",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("brief output = %q, want %q", output, want)
		}
	}

	contents := readProjectBrief(t, root)
	for _, want := range []string{
		"# Project Brief",
		"Answer: CLI/tooling project",
		"Answer: Developer productivity tool",
		"Answer: Developers or platform engineers",
		"Answer: Go",
		"Answer: Clean Architecture / Hexagonal",
		"Answer: go mod download",
		"Answer: go test ./...",
		"Answer: go build ./...",
		"Answer: go run ./cmd/<name>",
		"Answer: Ask before assuming",
		"### Detected context",
		"None recorded.",
		"## Assumptions",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("project brief = %q, want %q", contents, want)
		}
	}
	assertPathDoesNotExist(t, root, "openspec")
}

func TestExecuteBriefSupportsCustomAnswers(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	inputs := []string{
		"5", "Desktop application",
		"5", "Coordinate field inspections",
		"5", "Inspection managers",
		"5", "Go and SQLite",
		"5", "Ports and adapters",
		"5", "make deps",
		"5", "make test",
		"5", "make build",
		"5", "make run",
		"4", "Ask in the active OpenSpec change",
		"YES",
	}
	output, err := executeBrief(t, inputs)
	if err != nil {
		t.Fatalf("execute brief custom error = %v\noutput:\n%s", err, output)
	}
	if strings.Count(output, "Custom answer:") != 10 {
		t.Fatalf("custom output = %q, want ten custom prompts", output)
	}

	contents := readProjectBrief(t, root)
	for _, want := range []string{
		"Answer: Desktop application",
		"Answer: Coordinate field inspections",
		"Answer: Inspection managers",
		"Answer: Go and SQLite",
		"Answer: Ports and adapters",
		"Answer: make deps",
		"Answer: make test",
		"Answer: make build",
		"Answer: make run",
		"Answer: Ask in the active OpenSpec change",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("custom project brief = %q, want %q", contents, want)
		}
	}
}

func TestExecuteBriefRecordsDiscoveredContextSeparatelyFromConfirmedAnswers(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/project\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}

	inputs := []string{
		"3",
		"3",
		"3",
		"1",
		"2",
		"4",
		"1",
		"1",
		"1",
		"1",
		"y",
	}
	output, err := executeBrief(t, inputs)
	if err != nil {
		t.Fatalf("execute brief with discovery error = %v\noutput:\n%s", err, output)
	}

	contents := readProjectBrief(t, root)
	for _, want := range []string{
		"Answer: Go",
		"Answer: go test ./...",
		"Answer: go build ./...",
		"### Detected context",
		"- Stack from go.mod: Go (Source: detected context)",
		"- Language from go.mod: Go (Source: detected context)",
		"- Package manager from go.mod: Go modules (Source: detected context)",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("project brief with discovery = %q, want %q", contents, want)
		}
	}
	if strings.Contains(contents, "Assumption: Test command: go test ./...") {
		t.Fatalf("confirmed test suggestion was also recorded as an assumption:\n%s", contents)
	}
	if strings.Contains(contents, "Source: detected context\n\n## Stack") {
		t.Fatalf("detected context was rendered as confirmed stack:\n%s", contents)
	}
}

func TestExecuteBriefCancellationWritesNoFile(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
	}{
		{name: "n", inputs: append(defaultProjectBriefPromptInputs(), "n")},
		{name: "NO", inputs: append(defaultProjectBriefPromptInputs(), "NO")},
		{name: "empty confirmation", inputs: append(defaultProjectBriefPromptInputs(), "")},
		{name: "EOF confirmation", inputs: defaultProjectBriefPromptInputs()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)

			output, err := executeBrief(t, test.inputs)
			if err == nil || err.Error() != "operation cancelled" {
				t.Fatalf("execute brief cancel error = %v, want operation cancelled\noutput:\n%s", err, output)
			}
			if !strings.Contains(output, "Confirm? [y/N]:") {
				t.Fatalf("cancel output = %q, want confirmation prompt", output)
			}
			assertPathDoesNotExist(t, root, ".specharbor/project-brief.md")
		})
	}
}

func TestExecuteBriefConfirmationRetryExhaustionWritesNoFile(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	inputs := append(defaultProjectBriefPromptInputs(), "maybe", "later", "ok")
	output, err := executeBrief(t, inputs)
	if err == nil || err.Error() != "confirmation retry limit exceeded" {
		t.Fatalf("execute brief confirmation exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
	}
	if !strings.Contains(output, "Invalid answer: enter y/yes or n/no.") {
		t.Fatalf("output = %q, want invalid confirmation retry", output)
	}
	assertPathDoesNotExist(t, root, ".specharbor/project-brief.md")
}

func TestExecuteBriefInvalidChoiceAndCustomRetryExhaustionWriteNoFile(t *testing.T) {
	t.Run("menu retry exhausted", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)

		output, err := executeBrief(t, []string{"bad", "0", "6"})
		if err == nil || err.Error() != "project type retry limit exceeded" {
			t.Fatalf("execute brief menu exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		assertPathDoesNotExist(t, root, ".specharbor/project-brief.md")
	})

	t.Run("custom retry exhausted", func(t *testing.T) {
		root := t.TempDir()
		t.Chdir(root)

		output, err := executeBrief(t, []string{"5", "", " ", "\t"})
		if err == nil || err.Error() != "project type custom answer retry limit exceeded" {
			t.Fatalf("execute brief custom exhaustion error = %v, want retry limit\noutput:\n%s", err, output)
		}
		if !strings.Contains(output, "Invalid answer: custom answer is required") {
			t.Fatalf("output = %q, want custom answer validation", output)
		}
		assertPathDoesNotExist(t, root, ".specharbor/project-brief.md")
	})
}

func TestExecuteBriefDoesNotOverwriteExistingProjectBrief(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	path := filepath.Join(root, ".specharbor", "project-brief.md")
	if err := os.WriteFile(path, []byte("existing project brief"), 0o644); err != nil {
		t.Fatalf("WriteFile(project-brief.md) error = %v", err)
	}

	output, err := executeBrief(t, append(defaultProjectBriefPromptInputs(), "y"))
	if err == nil || !strings.Contains(err.Error(), "project brief already exists at .specharbor/project-brief.md") {
		t.Fatalf("execute brief existing error = %v, want existing refusal\noutput:\n%s", err, output)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(project-brief.md) error = %v", err)
	}
	if string(contents) != "existing project brief" {
		t.Fatalf("existing project brief = %q, want preserved", string(contents))
	}
	if strings.Contains(output, "SpecHarbor project brief created.") {
		t.Fatalf("output = %q, must not report success", output)
	}
}

func TestExecuteBriefUpdateRequiresExistingProjectBrief(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{isTTY: true, output: &output}
	err := executeWithTerminal([]string{"brief", "--update"}, &output, terminal)
	if err == nil || !strings.Contains(err.Error(), "project brief does not exist at .specharbor/project-brief.md") {
		t.Fatalf("execute brief --update error = %v, want missing brief error", err)
	}
	if terminal.reads != 0 {
		t.Fatalf("terminal reads = %d, want 0", terminal.reads)
	}
	assertPathDoesNotExist(t, root, "openspec")
}

func TestExecuteBriefUpdateCancellationWritesNoFile(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeCLIUpdateProjectBrief(t, root, "Go")
	original := readProjectBrief(t, root)

	inputs := append(defaultProjectBriefUpdateKeepInputs(), "n")
	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{isTTY: true, output: &output, inputs: inputs}
	err := executeWithTerminal([]string{"brief", "--update"}, &output, terminal)
	if err == nil || err.Error() != "operation cancelled" {
		t.Fatalf("execute brief --update cancel error = %v, want operation cancelled\noutput:\n%s", err, output.String())
	}
	if !strings.Contains(output.String(), "Project brief update preview:") {
		t.Fatalf("output = %q, want update preview", output.String())
	}
	if !strings.Contains(output.String(), "Write updated project brief? [y/N]:") {
		t.Fatalf("output = %q, want final confirmation", output.String())
	}
	if got := readProjectBrief(t, root); got != original {
		t.Fatalf("project brief changed after cancellation:\n%s", got)
	}
}

func TestExecuteBriefUpdateAcceptsDetectedFactAfterConfirmation(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeCLIUpdateProjectBrief(t, root, "Node.js")
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/project\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}

	inputs := []string{
		"1",
		"1",
		"1",
		"4",
		"1",
		"1",
		"1",
		"1",
		"1",
		"y",
	}
	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{isTTY: true, output: &output, inputs: inputs}
	err := executeWithTerminal([]string{"brief", "--update"}, &output, terminal)
	if err != nil {
		t.Fatalf("execute brief --update error = %v\noutput:\n%s", err, output.String())
	}

	contents := readProjectBrief(t, root)
	if !strings.Contains(contents, "## Stack\n\nAnswer: Go\n\nSource: user-provided answer") {
		t.Fatalf("updated brief = %q, want accepted detected Go stack", contents)
	}
	if !strings.Contains(output.String(), "SpecHarbor project brief updated.") {
		t.Fatalf("output = %q, want update success", output.String())
	}
}

func TestExecuteBriefKeepsVersionCommandWorking(t *testing.T) {
	var output bytes.Buffer
	if err := execute([]string{"version"}, &output); err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}
	if output.String() != version.Current().Format()+"\n" {
		t.Fatalf("version output = %q, want current metadata report", output.String())
	}
}

func executeBrief(t *testing.T, inputs []string) (string, error) {
	t.Helper()

	var output bytes.Buffer
	terminal := &cliFakeInteractiveTerminal{
		isTTY:  true,
		output: &output,
		inputs: inputs,
	}
	err := executeWithTerminal([]string{"brief"}, &output, terminal)
	return output.String(), err
}

func defaultProjectBriefPromptInputs() []string {
	return []string{
		"3",
		"3",
		"3",
		"1",
		"2",
		"4",
		"2",
		"2",
		"2",
		"1",
	}
}

func readProjectBrief(t *testing.T, root string) string {
	t.Helper()

	contents, err := os.ReadFile(filepath.Join(root, ".specharbor", "project-brief.md"))
	if err != nil {
		t.Fatalf("ReadFile(project-brief.md) error = %v", err)
	}
	return string(contents)
}

func writeCLIUpdateProjectBrief(t *testing.T, root string, stack string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	contents := `# Project Brief

## Project type

Answer: CLI/tooling project

Source: user-provided answer

## Purpose

Answer: Developer productivity tool

Source: user-provided answer

## Target users

Answer: Developers or platform engineers

Source: user-provided answer

## Stack

Answer: ` + stack + `

Source: user-provided answer

## Architecture

Answer: Clean Architecture / Hexagonal

Source: user-provided answer

## Commands

### Install

Answer: go mod download

Source: user-provided answer

### Test

Answer: go test ./...

Source: user-provided answer

### Build

Answer: go build ./...

Source: user-provided answer

### Run

Answer: go run ./cmd/specharbor

Source: user-provided answer

## Agent behavior

Answer: Ask before assuming

Source: user-provided answer

## Context sources

### User-provided answers

- Project type: CLI/tooling project (Source: user-provided answer)
- Purpose: Developer productivity tool (Source: user-provided answer)
- Target users: Developers or platform engineers (Source: user-provided answer)
- Stack: ` + stack + ` (Source: user-provided answer)
- Architecture: Clean Architecture / Hexagonal (Source: user-provided answer)
- Install command: go mod download (Source: user-provided answer)
- Test command: go test ./... (Source: user-provided answer)
- Build command: go build ./... (Source: user-provided answer)
- Run command: go run ./cmd/specharbor (Source: user-provided answer)
- Agent behavior: Ask before assuming (Source: user-provided answer)

### Detected context

None recorded.

## Assumptions

None recorded.
`
	if err := os.WriteFile(filepath.Join(root, ".specharbor", "project-brief.md"), []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(project-brief.md) error = %v", err)
	}
}

func defaultProjectBriefUpdateKeepInputs() []string {
	return []string{
		"1",
		"1",
		"1",
		"1",
		"1",
		"1",
		"1",
		"1",
		"1",
	}
}
