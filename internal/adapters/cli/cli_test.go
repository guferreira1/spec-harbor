package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

	configContents, err := os.ReadFile(filepath.Join(root, ".specharbor", "config.yml"))
	if err != nil {
		t.Fatalf("ReadFile(config.yml) error = %v", err)
	}
	for _, want := range []string{"version: 1", "templates:", "  aliases: {}"} {
		if !strings.Contains(string(configContents), want) {
			t.Fatalf("config.yml = %q, want %q", string(configContents), want)
		}
	}

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

func TestExecuteGenerateTemplatePrintsCreatedReport(t *testing.T) {
	tests := []struct {
		templateName string
		changeID     string
		wantProposal string
	}{
		{
			templateName: "feature",
			changeID:     "add-feature",
			wantProposal: "## Proposed Solution",
		},
		{
			templateName: "bugfix",
			changeID:     "fix-bug",
			wantProposal: "## Current Behavior",
		},
		{
			templateName: "docs",
			changeID:     "update-docs",
			wantProposal: "## Documentation Goal",
		},
		{
			templateName: "refactor",
			changeID:     "cleanup-code",
			wantProposal: "## Refactor Goal",
		},
	}

	for _, test := range tests {
		t.Run(test.templateName, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)

			var output bytes.Buffer
			if err := execute([]string{"generate", test.changeID, "--template", test.templateName}, &output); err != nil {
				t.Fatalf("execute(generate) error = %v", err)
			}

			want := `SpecHarbor template change generated.
Change: ` + test.changeID + `
Template: ` + test.templateName + `
Path: openspec/changes/` + test.changeID + `
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

			changeDirectory := filepath.Join(root, "openspec", "changes", test.changeID)
			for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
				path := filepath.Join(changeDirectory, requiredFile)
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

			proposal, err := os.ReadFile(filepath.Join(changeDirectory, "proposal.md"))
			if err != nil {
				t.Fatalf("ReadFile(proposal.md) error = %v", err)
			}
			if !strings.Contains(string(proposal), test.wantProposal) {
				t.Fatalf("proposal.md = %q, want to contain %q", string(proposal), test.wantProposal)
			}
		})
	}
}

func TestExecuteGenerateGuidedPrintsCreatedReport(t *testing.T) {
	tests := []struct {
		guidedType   string
		changeID     string
		wantProposal string
		title        string
		summary      string
	}{
		{
			guidedType:   "feature",
			changeID:     "guided-feature",
			wantProposal: "## Proposed Solution",
			title:        "Add reports",
			summary:      "Create report generation support",
		},
		{
			guidedType:   "bugfix",
			changeID:     "guided-bugfix",
			wantProposal: "## Current Behavior",
			title:        "Fix validation",
			summary:      "Correct invalid validation status",
		},
		{
			guidedType:   "docs",
			changeID:     "guided-docs",
			wantProposal: "## Documentation Goal",
			title:        "Update docs",
			summary:      "Document guided generation",
		},
		{
			guidedType:   "refactor",
			changeID:     "guided-refactor",
			wantProposal: "## Refactor Goal",
			title:        "Simplify generation",
			summary:      "Refactor generation orchestration",
		},
	}

	for _, test := range tests {
		t.Run(test.guidedType, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)

			var output bytes.Buffer
			if err := execute([]string{
				"generate",
				test.changeID,
				"--guided",
				"--type",
				test.guidedType,
				"--title",
				test.title,
				"--summary",
				test.summary,
			}, &output); err != nil {
				t.Fatalf("execute(generate) error = %v", err)
			}

			want := `SpecHarbor guided change generated.
Change: ` + test.changeID + `
Guided type: ` + test.guidedType + `
Title: ` + test.title + `
Path: openspec/changes/` + test.changeID + `
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

			changeDirectory := filepath.Join(root, "openspec", "changes", test.changeID)
			for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
				path := filepath.Join(changeDirectory, requiredFile)
				contents, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("ReadFile(%q) error = %v", path, err)
				}
				generated := string(contents)
				if strings.TrimSpace(generated) == "" {
					t.Fatalf("generated file %q is empty", requiredFile)
				}
				if !strings.Contains(generated, test.title) {
					t.Fatalf("%s = %q, want title %q", requiredFile, generated, test.title)
				}
				if !strings.Contains(generated, test.summary) {
					t.Fatalf("%s = %q, want summary %q", requiredFile, generated, test.summary)
				}
				if requiredFile == "tasks.md" {
					assertUncheckedTasksOnly(t, generated)
				}
			}

			proposal, err := os.ReadFile(filepath.Join(changeDirectory, "proposal.md"))
			if err != nil {
				t.Fatalf("ReadFile(proposal.md) error = %v", err)
			}
			if !strings.Contains(string(proposal), test.wantProposal) {
				t.Fatalf("proposal.md = %q, want to contain %q", string(proposal), test.wantProposal)
			}
		})
	}
}

func TestExecuteGenerateAgentAssistedPrintsDryRunReport(t *testing.T) {
	tests := []struct {
		authoringType string
		changeID      string
		title         string
		summary       string
	}{
		{
			authoringType: "feature",
			changeID:      "agent-feature",
			title:         "Add payment retry policy",
			summary:       "Create a controlled retry policy",
		},
		{
			authoringType: "bugfix",
			changeID:      "agent-bugfix",
			title:         "Fix payment retry policy",
			summary:       "Correct retry classification",
		},
		{
			authoringType: "docs",
			changeID:      "agent-docs",
			title:         "Document payment retry policy",
			summary:       "Describe retry behavior",
		},
		{
			authoringType: "refactor",
			changeID:      "agent-refactor",
			title:         "Refactor payment retry policy",
			summary:       "Simplify retry orchestration",
		},
	}

	for _, test := range tests {
		t.Run(test.authoringType, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)

			var output bytes.Buffer
			err := execute([]string{
				"generate",
				test.changeID,
				"--agent-assisted",
				"--agent",
				"codex",
				"--type",
				test.authoringType,
				"--title",
				test.title,
				"--summary",
				test.summary,
			}, &output)
			if err != nil {
				t.Fatalf("execute(generate agent-assisted) error = %v", err)
			}

			generateOutput := output.String()
			for _, want := range []string{
				"SpecHarbor agent-assisted spec authoring dry run.",
				"Change: " + test.changeID,
				"Agent: codex",
				"Authoring type: " + test.authoringType,
				"Title: " + test.title,
				"Summary: " + test.summary,
				"Path: openspec/changes/" + test.changeID,
				"Dry run: yes",
				"Required files:",
				"- proposal.md",
				"- design.md",
				"- tasks.md",
				"- acceptance-criteria.md",
				"- risks.md",
				"Plan:",
				"- Validate the agent-assisted authoring request.",
				"- Build the OpenSpec authoring plan for openspec/changes/" + test.changeID + ".",
				"- Render a deterministic, copy-pasteable Markdown authoring prompt.",
				"- Stop before writing files, executing agents, or running external commands.",
				"Status:",
				"- No files written: yes",
				"- No prompt file written: yes",
				"- No agent executed: yes",
				"- No external command executed: yes",
				"- No agent output parsed or applied: yes",
				"Generated prompt:",
				"# Agent-Assisted OpenSpec Authoring Prompt",
				"Change id: `" + test.changeID + "`",
				"Authoring type: `" + test.authoringType + "`",
				"Title: " + test.title,
				"Summary: " + test.summary,
				"Create or refine only files under `openspec/changes/" + test.changeID + "/`",
				"Do not implement production code.",
				"specharbor validate " + test.changeID,
			} {
				if !strings.Contains(generateOutput, want) {
					t.Fatalf("generate output = %q, want to contain %q", generateOutput, want)
				}
			}

			if strings.Contains(generateOutput, root) {
				t.Fatalf("generate output = %q, want no absolute temp root path", generateOutput)
			}
			assertAgentAssistedDryRunOutputSafety(t, generateOutput, root)
			assertPathDoesNotExist(t, root, "openspec")
		})
	}
}

func TestExecuteGenerateAgentAssistedAcceptsFlagsBeforeChangeID(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	if err := execute([]string{
		"generate",
		"--agent-assisted",
		"--agent",
		"codex",
		"--type",
		"feature",
		"--title",
		"Add reports",
		"--summary",
		"Create report support",
		"order-independent-agent-assisted",
	}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	if !strings.Contains(output.String(), "Change: order-independent-agent-assisted") {
		t.Fatalf("generate output = %q, want change id", output.String())
	}
	if !strings.Contains(output.String(), "Agent: codex") {
		t.Fatalf("generate output = %q, want agent", output.String())
	}
	if !strings.Contains(output.String(), "Authoring type: feature") {
		t.Fatalf("generate output = %q, want authoring type", output.String())
	}
	assertPathDoesNotExist(t, root, "openspec")
}

func TestExecuteGenerateAgentAssistedDryRunSupportsRecognizedTargets(t *testing.T) {
	for _, target := range domain.RecognizedAgentTargets() {
		t.Run(string(target.ID), func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)

			var output bytes.Buffer
			err := execute([]string{
				"generate",
				"agent-" + string(target.ID),
				"--agent-assisted",
				"--agent",
				string(target.ID),
				"--type",
				"feature",
				"--title",
				"Title",
				"--summary",
				"Summary",
			}, &output)
			if err != nil {
				t.Fatalf("execute(generate agent-assisted) error = %v", err)
			}
			generateOutput := output.String()
			if !strings.Contains(generateOutput, "SpecHarbor agent-assisted spec authoring dry run.") {
				t.Fatalf("generate output = %q, want dry-run report", generateOutput)
			}
			if !strings.Contains(generateOutput, "Agent: "+string(target.ID)) {
				t.Fatalf("generate output = %q, want target id", generateOutput)
			}
			if strings.Contains(generateOutput, "Resolved command:") {
				t.Fatalf("generate output = %q, want no executable mapping in dry-run", generateOutput)
			}
			assertPathDoesNotExist(t, root, "openspec")
		})
	}
}

func TestExecuteGenerateAgentAssistedExecutePrintsRunReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	installFakeAgentCommand(t, "codex", `
input=$(cat)
case "$input" in
*"Title: Add reports"*) printf 'stdout prompt received\n' ;;
*) printf 'stdout prompt missing\n' ;;
esac
printf 'stderr captured\n' >&2
exit 0
`)

	var output bytes.Buffer
	err := execute([]string{
		"generate",
		"agent-execute",
		"--agent-assisted",
		"--agent",
		"codex",
		"--type",
		"feature",
		"--title",
		"Add reports",
		"--summary",
		"Create report support",
		"--execute",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate agent-assisted --execute) error = %v", err)
	}

	generateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor agent-assisted spec authoring execute run.",
		"Mode: execute",
		"Change: agent-execute",
		"Agent target id: codex",
		"Agent display name: Codex",
		"Resolved command: codex",
		"Resolved args: (none)",
		"Authoring type: feature",
		"Title: Add reports",
		"Summary: Create report support",
		"Path: openspec/changes/agent-execute",
		"Working directory: project root (" + root + ")",
		"Prompt sent through stdin: yes",
		"Execution status: success",
		"Exit code: 0",
		"Stdout:",
		"stdout prompt received",
		"Stderr:",
		"stderr captured",
		"- Output parsed or applied by SpecHarbor: no",
		"- OpenSpec files written from runner output: no",
		"- Production code modified by SpecHarbor: no",
		"- Auto-commit, auto-push, or auto-merge by SpecHarbor: no",
	} {
		if !strings.Contains(generateOutput, want) {
			t.Fatalf("generate output = %q, want to contain %q", generateOutput, want)
		}
	}
	assertPathDoesNotExist(t, root, "openspec")
}

func TestExecuteGenerateAgentAssistedExecuteAcceptsMappedTargets(t *testing.T) {
	for _, command := range domain.ExecutableAgentCommands() {
		t.Run(string(command.AgentID), func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			installFakeAgentCommand(t, command.CommandName, "printf 'ok\\n'\nexit 0\n")

			var output bytes.Buffer
			err := execute([]string{
				"generate",
				"agent-" + string(command.AgentID),
				"--agent-assisted",
				"--agent",
				string(command.AgentID),
				"--type",
				"feature",
				"--title",
				"Title",
				"--summary",
				"Summary",
				"--execute",
			}, &output)
			if err != nil {
				t.Fatalf("execute(generate agent-assisted --execute) error = %v", err)
			}
			generateOutput := output.String()
			for _, want := range []string{
				"Agent target id: " + string(command.AgentID),
				"Agent display name: " + command.AgentDisplayName,
				"Resolved command: " + command.CommandName,
				"Execution status: success",
				"ok",
			} {
				if !strings.Contains(generateOutput, want) {
					t.Fatalf("generate output = %q, want %q", generateOutput, want)
				}
			}
		})
	}
}

func TestExecuteGenerateAgentAssistedExecutePrintsFullReportBeforeNonZeroExit(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	installFakeAgentCommand(t, "codex", `
printf 'stdout nonzero\n'
printf 'stderr nonzero\n' >&2
exit 6
`)

	var output bytes.Buffer
	err := execute([]string{
		"generate",
		"agent-nonzero",
		"--agent-assisted",
		"--agent",
		"codex",
		"--type",
		"bugfix",
		"--title",
		"Fix reports",
		"--summary",
		"Correct report behavior",
		"--execute",
	}, &output)
	assertExitCode(t, err, 6)

	generateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor agent-assisted spec authoring execute run.",
		"Execution status: non_zero_exit",
		"Exit code: 6",
		"stdout nonzero",
		"stderr nonzero",
		"- Output parsed or applied by SpecHarbor: no",
		"- OpenSpec files written from runner output: no",
		"- Production code modified by SpecHarbor: no",
		"- Auto-commit, auto-push, or auto-merge by SpecHarbor: no",
	} {
		if !strings.Contains(generateOutput, want) {
			t.Fatalf("generate output = %q, want to contain %q", generateOutput, want)
		}
	}
}

func TestExecuteGenerateAgentAssistedExecuteReportsEmptyStdoutAndStderr(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	installFakeAgentCommand(t, "codex", "exit 0\n")

	var output bytes.Buffer
	err := execute([]string{
		"generate",
		"agent-empty-output",
		"--agent-assisted",
		"--agent",
		"codex",
		"--type",
		"docs",
		"--title",
		"Document reports",
		"--summary",
		"Describe report behavior",
		"--execute",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate agent-assisted --execute) error = %v", err)
	}
	if count := strings.Count(output.String(), "(empty)"); count != 2 {
		t.Fatalf("generate output = %q, want two empty output markers", output.String())
	}
}

func TestExecuteGenerateAgentAssistedExecuteRejectsGenericAndUnknownAgents(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "generic execute",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "generic", "--type", "feature", "--title", "Title", "--summary", "Summary", "--execute"},
			want: "agent target has no executable local runner mapping in this change: generic",
		},
		{
			name: "unknown dry run",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "unknown", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "unknown agent target: unknown",
		},
		{
			name: "unknown execute",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "unknown", "--type", "feature", "--title", "Title", "--summary", "Summary", "--execute"},
			want: "unknown agent target: unknown",
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
			if err.Error() != test.want {
				t.Fatalf("execute(%v) error = %q, want %q", test.args, err.Error(), test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(%v) output = %q, want empty output", test.args, output.String())
			}
		})
	}
}

func TestExecuteGenerateAgentAssistedExecuteStartupFailureDoesNotPrintNormalReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("PATH", t.TempDir())

	var output bytes.Buffer
	err := execute([]string{
		"generate",
		"agent-startup-failure",
		"--agent-assisted",
		"--agent",
		"codex",
		"--type",
		"refactor",
		"--title",
		"Refactor reports",
		"--summary",
		"Simplify report flow",
		"--execute",
	}, &output)
	if err == nil {
		t.Fatalf("execute(generate agent-assisted --execute) error = nil, want startup failure")
	}
	if !strings.Contains(err.Error(), `start agent runner command "codex"`) {
		t.Fatalf("execute() error = %q, want command context", err.Error())
	}
	if strings.Contains(err.Error(), "Exit code:") {
		t.Fatalf("execute() error = %q, want no exit code for startup failure", err.Error())
	}
	if output.String() != "" {
		t.Fatalf("execute() output = %q, want no normal runner report", output.String())
	}
}

func TestExecuteGenerateTemplateAcceptsFlagBeforeChangeID(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	if err := execute([]string{"generate", "--template", "feature", "order-independent-template"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	if !strings.Contains(output.String(), "Change: order-independent-template") {
		t.Fatalf("generate output = %q, want change id", output.String())
	}
	if !strings.Contains(output.String(), "Template: feature") {
		t.Fatalf("generate output = %q, want template name", output.String())
	}
}

func TestExecuteGenerateGuidedAcceptsFlagsBeforeChangeID(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	if err := execute([]string{
		"generate",
		"--guided",
		"--type",
		"feature",
		"--title",
		"Add reports",
		"--summary",
		"Create report generation support",
		"order-independent-guided",
	}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	if !strings.Contains(output.String(), "Change: order-independent-guided") {
		t.Fatalf("generate output = %q, want change id", output.String())
	}
	if !strings.Contains(output.String(), "Guided type: feature") {
		t.Fatalf("generate output = %q, want guided type", output.String())
	}
	if !strings.Contains(output.String(), "Title: Add reports") {
		t.Fatalf("generate output = %q, want title", output.String())
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

func TestExecuteGenerateTemplateCompletesPartialChangeAndPrintsSkippedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createOpenSpecChange(t, root, "partial-template-change", []string{"proposal.md", "tasks.md"})

	changeDirectory := filepath.Join(root, "openspec", "changes", "partial-template-change")
	proposalPath := filepath.Join(changeDirectory, "proposal.md")
	tasksPath := filepath.Join(changeDirectory, "tasks.md")
	if err := os.WriteFile(proposalPath, []byte("custom proposal"), 0o644); err != nil {
		t.Fatalf("WriteFile(proposal.md) error = %v", err)
	}
	if err := os.WriteFile(tasksPath, []byte("custom tasks"), 0o644); err != nil {
		t.Fatalf("WriteFile(tasks.md) error = %v", err)
	}

	var output bytes.Buffer
	if err := execute([]string{"generate", "partial-template-change", "--template", "feature"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor template change generated.
Change: partial-template-change
Template: feature
Path: openspec/changes/partial-template-change
Directory: existing
Created files: 3
Skipped existing files: 2

Created:
- design.md
- acceptance-criteria.md
- risks.md

Skipped existing:
- proposal.md
- tasks.md
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}

	proposal, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if string(proposal) != "custom proposal" {
		t.Fatalf("proposal.md contents = %q, want preserved custom proposal", string(proposal))
	}
	tasks, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatalf("ReadFile(tasks.md) error = %v", err)
	}
	if string(tasks) != "custom tasks" {
		t.Fatalf("tasks.md contents = %q, want preserved custom tasks", string(tasks))
	}
	design, err := os.ReadFile(filepath.Join(changeDirectory, "design.md"))
	if err != nil {
		t.Fatalf("ReadFile(design.md) error = %v", err)
	}
	if !strings.Contains(string(design), "## Architecture Notes") {
		t.Fatalf("design.md = %q, want feature template content", string(design))
	}
}

func TestExecuteGenerateAIAssistedPrintsSuccessReportAndWritesFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeAIOutputFile(t, root, "agent-output.txt", validAIAssistedCLIOutput(nil))

	var output bytes.Buffer
	if err := execute([]string{"generate", "ai-change", "--ai-assisted", "--from-file", "agent-output.txt"}, &output); err != nil {
		t.Fatalf("execute(generate --ai-assisted) error = %v", err)
	}

	want := `SpecHarbor AI-assisted change generated.
Change: ai-change
Source file: agent-output.txt
Path: openspec/changes/ai-change
Directory: created
Overwrite: no
Generated files: 5
Skipped existing files: 0
Overwritten files: 0

Generated:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md

Validation:
Status: valid
Required files: 5
Errors: 0
Warnings: 0

Safety:
- Provider APIs called: no
- Remote AI services called: no
- Agent commands executed: no
- Production code modified: no
- Source-control commands run: no
- Auto-commit, auto-push, PR, merge, or archive: no
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		path := filepath.Join(root, "openspec", "changes", "ai-change", requiredFile)
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if strings.TrimSpace(string(contents)) == "" {
			t.Fatalf("%s is empty", requiredFile)
		}
	}
	assertPathDoesNotExist(t, root, "internal")
}

func TestExecuteGenerateAIAssistedPrintsParseErrorsAndWritesNothing(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeAIOutputFile(t, root, "bad-output.txt", "---FILE: notes.md---\n# Notes\n\nNope.\n---END FILE---\n")

	var output bytes.Buffer
	err := execute([]string{"generate", "bad-ai-change", "--ai-assisted", "--from-file", "bad-output.txt"}, &output)
	assertExitCode(t, err, 1)

	generateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor AI-assisted import failed.",
		"Change: bad-ai-change",
		"Source file: bad-output.txt",
		"Parse status: invalid",
		"Files written: 0",
		"No files written: yes",
		"Parse errors:",
		"unknown_file_block",
		"missing_file_block",
		"file: notes.md",
		"line: 1",
		"Safety:",
		"- Provider APIs called: no",
		"- Production code modified: no",
	} {
		if !strings.Contains(generateOutput, want) {
			t.Fatalf("generate output = %q, want to contain %q", generateOutput, want)
		}
	}
	assertPathDoesNotExist(t, root, "openspec/changes/bad-ai-change")
}

func TestExecuteGenerateAIAssistedSkipsAndOverwritesExistingFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeAIOutputFile(t, root, "agent-output.txt", validAIAssistedCLIOutput(nil))
	createAuthoredOpenSpecChange(t, root, "ai-existing", nil)

	var output bytes.Buffer
	if err := execute([]string{"generate", "ai-existing", "--ai-assisted", "--from-file", "agent-output.txt"}, &output); err != nil {
		t.Fatalf("execute(generate --ai-assisted skip) error = %v", err)
	}
	skipOutput := output.String()
	for _, want := range []string{
		"Directory: existing",
		"Overwrite: no",
		"Generated files: 0",
		"Skipped existing files: 5",
		"Skipped existing:",
		"- proposal.md",
	} {
		if !strings.Contains(skipOutput, want) {
			t.Fatalf("skip output = %q, want %q", skipOutput, want)
		}
	}
	proposalPath := filepath.Join(root, "openspec", "changes", "ai-existing", "proposal.md")
	proposal, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if !strings.Contains(string(proposal), "Users cannot tell whether a change package is ready") {
		t.Fatalf("proposal after skip = %q, want original authored content", string(proposal))
	}

	output.Reset()
	if err := execute([]string{"generate", "ai-existing", "--ai-assisted", "--from-file", "agent-output.txt", "--overwrite"}, &output); err != nil {
		t.Fatalf("execute(generate --ai-assisted overwrite) error = %v", err)
	}
	overwriteOutput := output.String()
	for _, want := range []string{
		"Overwrite: yes",
		"Generated files: 0",
		"Skipped existing files: 0",
		"Overwritten files: 5",
		"Overwritten:",
		"- proposal.md",
	} {
		if !strings.Contains(overwriteOutput, want) {
			t.Fatalf("overwrite output = %q, want %q", overwriteOutput, want)
		}
	}
	proposal, err = os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if !strings.Contains(string(proposal), "AI-assisted imports need a safe bridge.") {
		t.Fatalf("proposal after overwrite = %q, want AI-assisted content", string(proposal))
	}
}

func TestExecuteGenerateAIAssistedOverwriteRejectsSymlinkOutputFile(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeAIOutputFile(t, root, "agent-output.txt", validAIAssistedCLIOutput(nil))

	changeDirectory := filepath.Join(root, "openspec", "changes", "symlink-ai")
	if err := os.MkdirAll(changeDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll(change) error = %v", err)
	}
	outsidePath := filepath.Join(t.TempDir(), "outside-proposal.md")
	if err := os.WriteFile(outsidePath, []byte("outside original"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside) error = %v", err)
	}
	if err := os.Symlink(outsidePath, filepath.Join(changeDirectory, "proposal.md")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	var output bytes.Buffer
	err := execute([]string{"generate", "symlink-ai", "--ai-assisted", "--from-file", "agent-output.txt", "--overwrite"}, &output)
	if err == nil || !strings.Contains(err.Error(), "symlink target paths are not allowed for generated OpenSpec files") {
		t.Fatalf("execute(generate --ai-assisted --overwrite) error = %v, want symlink safety failure", err)
	}
	if strings.Contains(output.String(), "SpecHarbor AI-assisted change generated.") {
		t.Fatalf("output = %q, want no success report", output.String())
	}
	contents, readErr := os.ReadFile(outsidePath)
	if readErr != nil {
		t.Fatalf("ReadFile(outside) error = %v", readErr)
	}
	if string(contents) != "outside original" {
		t.Fatalf("outside target = %q, want unchanged", string(contents))
	}
}

func TestExecuteGenerateAIAssistedValidationWarningsExitZero(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeAIOutputFile(t, root, "warning-output.txt", validAIAssistedCLIOutput(map[string]string{
		"risks.md": "# Risks\n\n## Risks\n\n- Strict import can reject useful drafts.\n",
	}))

	var output bytes.Buffer
	if err := execute([]string{"generate", "ai-warning", "--ai-assisted", "--from-file", "warning-output.txt"}, &output); err != nil {
		t.Fatalf("execute(generate warnings) error = %v, want exit 0", err)
	}

	generateOutput := output.String()
	for _, want := range []string{
		"Validation:",
		"Status: valid",
		"Errors: 0",
		"Warnings: 1",
		"Warnings:",
		"risks_mitigation_missing",
	} {
		if !strings.Contains(generateOutput, want) {
			t.Fatalf("generate output = %q, want %q", generateOutput, want)
		}
	}
}

func TestExecuteGenerateAIAssistedValidationErrorsExitNonZeroAfterReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeAIOutputFile(t, root, "invalid-output.txt", validAIAssistedCLIOutput(map[string]string{
		"tasks.md": "# Tasks\n\n## Phase 1\n\nNo checkbox tasks here.\n",
	}))

	var output bytes.Buffer
	err := execute([]string{"generate", "ai-invalid", "--ai-assisted", "--from-file", "invalid-output.txt"}, &output)
	assertExitCode(t, err, 1)

	generateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor AI-assisted change generated.",
		"Generated files: 5",
		"Validation:",
		"Status: invalid",
		"Errors: 1",
		"Errors:",
		"tasks_checkbox_missing",
		"Safety:",
	} {
		if !strings.Contains(generateOutput, want) {
			t.Fatalf("generate output = %q, want %q", generateOutput, want)
		}
	}
	assertPathExists(t, root, "openspec/changes/ai-invalid/tasks.md")
}

func TestExecuteGenerateAIAssistedRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing from-file", args: []string{"generate", "change", "--ai-assisted"}, want: "source file is required"},
		{name: "from-file without ai", args: []string{"generate", "change", "--from-file", "agent-output.txt"}, want: "from-file requires --ai-assisted"},
		{name: "overwrite without ai", args: []string{"generate", "change", "--blank", "--overwrite"}, want: "overwrite requires --ai-assisted"},
		{name: "duplicate ai-assisted", args: []string{"generate", "change", "--ai-assisted", "--ai-assisted", "--from-file", "agent-output.txt"}, want: "ai-assisted generation flag specified more than once"},
		{name: "duplicate from-file", args: []string{"generate", "change", "--ai-assisted", "--from-file", "one.txt", "--from-file", "two.txt"}, want: "from-file flag specified more than once"},
		{name: "duplicate overwrite", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--overwrite", "--overwrite"}, want: "overwrite flag specified more than once"},
		{name: "blank conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--blank"}, want: "ai-assisted and blank generation flags cannot be used together"},
		{name: "template conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--template", "feature"}, want: "ai-assisted and template generation flags cannot be used together"},
		{name: "guided conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--guided"}, want: "ai-assisted and guided generation flags cannot be used together"},
		{name: "agent-assisted conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--agent-assisted"}, want: "ai-assisted and agent-assisted generation flags cannot be used together"},
		{name: "execute conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--execute"}, want: "ai-assisted and execute flags cannot be used together"},
		{name: "agent flag conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--agent", "codex"}, want: "agent-assisted input flags cannot be used with --ai-assisted"},
		{name: "guided input conflict", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "--type", "feature"}, want: "guided input flags cannot be used with --ai-assisted"},
		{name: "extra argument", args: []string{"generate", "change", "--ai-assisted", "--from-file", "agent-output.txt", "extra"}, want: "unexpected argument: extra"},
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
			name: "missing generation mode flag",
			args: []string{"generate", "change"},
			want: "generation mode flag is required",
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
			name: "guided missing type",
			args: []string{"generate", "change", "--guided"},
			want: "guided type is required",
		},
		{
			name: "guided type without value",
			args: []string{"generate", "change", "--guided", "--type"},
			want: "guided type is required",
		},
		{
			name: "guided type followed by flag",
			args: []string{"generate", "change", "--guided", "--type", "--title", "Title", "--summary", "Summary"},
			want: "guided type is required",
		},
		{
			name: "guided missing title",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--summary", "Summary"},
			want: "guided title is required",
		},
		{
			name: "guided title without value",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title"},
			want: "guided title is required",
		},
		{
			name: "guided title followed by flag",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "--summary", "Summary"},
			want: "guided title is required",
		},
		{
			name: "guided missing summary",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title"},
			want: "guided summary is required",
		},
		{
			name: "guided summary without value",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--summary"},
			want: "guided summary is required",
		},
		{
			name: "guided summary followed by flag",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--summary", "--blank"},
			want: "guided summary is required",
		},
		{
			name: "guided empty type",
			args: []string{"generate", "change", "--guided", "--type", "", "--title", "Title", "--summary", "Summary"},
			want: "guided type is required",
		},
		{
			name: "guided empty title",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "", "--summary", "Summary"},
			want: "guided title is required",
		},
		{
			name: "guided empty summary",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--summary", ""},
			want: "guided summary is required",
		},
		{
			name: "unknown guided type",
			args: []string{"generate", "change", "--guided", "--type", "maintenance", "--title", "Title", "--summary", "Summary"},
			want: "unknown guided type: maintenance",
		},
		{
			name: "agent-assisted missing agent",
			args: []string{"generate", "change", "--agent-assisted", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent name is required",
		},
		{
			name: "agent-assisted agent without value",
			args: []string{"generate", "change", "--agent-assisted", "--agent"},
			want: "agent name is required",
		},
		{
			name: "agent-assisted agent followed by flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent name is required",
		},
		{
			name: "agent-assisted missing type",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted authoring type is required",
		},
		{
			name: "agent-assisted type without value",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type"},
			want: "agent-assisted authoring type is required",
		},
		{
			name: "agent-assisted type followed by flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted authoring type is required",
		},
		{
			name: "agent-assisted missing title",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--summary", "Summary"},
			want: "agent-assisted title is required",
		},
		{
			name: "agent-assisted title without value",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title"},
			want: "agent-assisted title is required",
		},
		{
			name: "agent-assisted title followed by flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "--summary", "Summary"},
			want: "agent-assisted title is required",
		},
		{
			name: "agent-assisted missing summary",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title"},
			want: "agent-assisted summary is required",
		},
		{
			name: "agent-assisted summary without value",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary"},
			want: "agent-assisted summary is required",
		},
		{
			name: "agent-assisted summary followed by flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "--blank"},
			want: "agent-assisted summary is required",
		},
		{
			name: "agent-assisted empty agent",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent name is required",
		},
		{
			name: "agent-assisted empty type",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted authoring type is required",
		},
		{
			name: "agent-assisted empty title",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "", "--summary", "Summary"},
			want: "agent-assisted title is required",
		},
		{
			name: "agent-assisted empty summary",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", ""},
			want: "agent-assisted summary is required",
		},
		{
			name: "unknown agent-assisted authoring type",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "maintenance", "--title", "Title", "--summary", "Summary"},
			want: "unknown agent-assisted authoring type: maintenance",
		},
		{
			name: "missing template name",
			args: []string{"generate", "change", "--template"},
			want: "template name is required",
		},
		{
			name: "template without change id",
			args: []string{"generate", "--template", "feature"},
			want: "change id is required",
		},
		{
			name: "empty template name",
			args: []string{"generate", "change", "--template", ""},
			want: "template name is required",
		},
		{
			name: "unknown template name",
			args: []string{"generate", "change", "--template", "maintenance"},
			want: "unknown template name: maintenance",
		},
		{
			name: "blank and template flags together",
			args: []string{"generate", "change", "--blank", "--template", "feature"},
			want: "blank and template generation flags cannot be used together",
		},
		{
			name: "template and blank flags together",
			args: []string{"generate", "change", "--template", "feature", "--blank"},
			want: "blank and template generation flags cannot be used together",
		},
		{
			name: "blank and template flags together without change id",
			args: []string{"generate", "--blank", "--template", "feature"},
			want: "blank and template generation flags cannot be used together",
		},
		{
			name: "guided and blank flags together",
			args: []string{"generate", "change", "--guided", "--blank", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "guided and blank generation flags cannot be used together",
		},
		{
			name: "blank and guided flags together",
			args: []string{"generate", "change", "--blank", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "guided and blank generation flags cannot be used together",
		},
		{
			name: "guided and template flags together",
			args: []string{"generate", "change", "--guided", "--template", "feature", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "guided and template generation flags cannot be used together",
		},
		{
			name: "template and guided flags together",
			args: []string{"generate", "change", "--template", "feature", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "guided and template generation flags cannot be used together",
		},
		{
			name: "agent-assisted and blank flags together",
			args: []string{"generate", "change", "--agent-assisted", "--blank", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted and blank generation flags cannot be used together",
		},
		{
			name: "agent-assisted and template flags together",
			args: []string{"generate", "change", "--agent-assisted", "--template", "feature", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted and template generation flags cannot be used together",
		},
		{
			name: "agent-assisted and guided flags together",
			args: []string{"generate", "change", "--agent-assisted", "--guided", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted and guided generation flags cannot be used together",
		},
		{
			name: "duplicate template flag",
			args: []string{"generate", "change", "--template", "feature", "--template", "bugfix"},
			want: "template generation flag specified more than once",
		},
		{
			name: "duplicate guided flag",
			args: []string{"generate", "change", "--guided", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "guided generation flag specified more than once",
		},
		{
			name: "duplicate guided type flag",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--type", "bugfix", "--title", "Title", "--summary", "Summary"},
			want: "guided type flag specified more than once",
		},
		{
			name: "duplicate guided title flag",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--title", "Other", "--summary", "Summary"},
			want: "guided title flag specified more than once",
		},
		{
			name: "duplicate guided summary flag",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary", "--summary", "Other"},
			want: "guided summary flag specified more than once",
		},
		{
			name: "duplicate agent-assisted flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted generation flag specified more than once",
		},
		{
			name: "duplicate agent flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--agent", "claude-code", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "agent flag specified more than once",
		},
		{
			name: "duplicate agent-assisted type flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--type", "bugfix", "--title", "Title", "--summary", "Summary"},
			want: "agent-assisted authoring type flag specified more than once",
		},
		{
			name: "duplicate agent-assisted title flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--title", "Other", "--summary", "Summary"},
			want: "agent-assisted title flag specified more than once",
		},
		{
			name: "duplicate agent-assisted summary flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary", "--summary", "Other"},
			want: "agent-assisted summary flag specified more than once",
		},
		{
			name: "duplicate execute flag",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary", "--execute", "--execute"},
			want: "execute flag specified more than once",
		},
		{
			name: "guided input without guided flag",
			args: []string{"generate", "change", "--type", "feature"},
			want: "guided input flags require --guided",
		},
		{
			name: "agent input without agent-assisted flag",
			args: []string{"generate", "change", "--agent", "codex"},
			want: "agent-assisted input flags require --agent-assisted",
		},
		{
			name: "unsupported ai flag",
			args: []string{"generate", "change", "--ai"},
			want: "unsupported flag: --ai",
		},
		{
			name: "unsupported execute flag without agent-assisted",
			args: []string{"generate", "change", "--execute"},
			want: "unsupported flag: --execute",
		},
		{
			name: "execute flag with blank",
			args: []string{"generate", "change", "--blank", "--execute"},
			want: "unsupported flag: --execute",
		},
		{
			name: "execute flag with template",
			args: []string{"generate", "change", "--template", "feature", "--execute"},
			want: "unsupported flag: --execute",
		},
		{
			name: "execute flag with guided",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary", "--execute"},
			want: "unsupported flag: --execute",
		},
		{
			name: "hybrid missing source selector",
			args: []string{"generate", "change", "--hybrid"},
			want: "hybrid source selector is required",
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
			name: "template extra argument",
			args: []string{"generate", "change", "--template", "feature", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "guided extra argument",
			args: []string{"generate", "change", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "agent-assisted extra argument",
			args: []string{"generate", "change", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "unsafe traversal change id",
			args: []string{"generate", "../unsafe", "--blank"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe template traversal change id",
			args: []string{"generate", "../unsafe", "--template", "feature"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe guided traversal change id",
			args: []string{"generate", "../unsafe", "--guided", "--type", "feature", "--title", "Title", "--summary", "Summary"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe agent-assisted traversal change id",
			args: []string{"generate", "../unsafe", "--agent-assisted", "--agent", "codex", "--type", "feature", "--title", "Title", "--summary", "Summary"},
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

func TestExecuteGenerateTemplateRejectsMissingOpenSpecProjectWithoutCreatingStructure(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	err := execute([]string{"generate", "missing-project", "--template", "feature"}, &output)
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

func TestExecuteGenerateGuidedRejectsMissingOpenSpecProjectWithoutCreatingStructure(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	err := execute([]string{
		"generate",
		"missing-project",
		"--guided",
		"--type",
		"feature",
		"--title",
		"Add reports",
		"--summary",
		"Create report generation support",
	}, &output)
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

func TestExecuteGenerateTemplateRejectsUnknownTemplateWithoutCreatingChange(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	err := execute([]string{"generate", "unknown-template-change", "--template", "maintenance"}, &output)
	if err == nil {
		t.Fatalf("execute(generate) error = nil, want unknown template error")
	}
	if !strings.Contains(err.Error(), "unknown template name: maintenance") {
		t.Fatalf("execute(generate) error = %q, want unknown template context", err.Error())
	}
	if output.String() != "" {
		t.Fatalf("execute(generate) output = %q, want empty output", output.String())
	}
	assertPathDoesNotExist(t, root, "openspec/changes/unknown-template-change")
}

func TestExecuteGenerateGuidedRejectsUnknownTypeWithoutCreatingChange(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	err := execute([]string{
		"generate",
		"unknown-guided-change",
		"--guided",
		"--type",
		"maintenance",
		"--title",
		"Title",
		"--summary",
		"Summary",
	}, &output)
	if err == nil {
		t.Fatalf("execute(generate) error = nil, want unknown guided type error")
	}
	if !strings.Contains(err.Error(), "unknown guided type: maintenance") {
		t.Fatalf("execute(generate) error = %q, want unknown guided type context", err.Error())
	}
	if output.String() != "" {
		t.Fatalf("execute(generate) output = %q, want empty output", output.String())
	}
	assertPathDoesNotExist(t, root, "openspec/changes/unknown-guided-change")
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
	if strings.Contains(promptOutput, "{{project_context}}") {
		t.Fatalf("prompt output = %q, want no raw project_context placeholder", promptOutput)
	}
	if strings.Contains(promptOutput, "not implemented") {
		t.Fatalf("prompt output = %q, want rendered prompt instead of placeholder", promptOutput)
	}
}

func TestExecutePromptIncludesContextAwareProjectContext(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join(root, ".specharbor", "rules"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor/rules) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".specharbor", "project-brief.md"), []byte(`# Project Brief

## Stack

Answer: Go

## Commands

### Test

Answer: go test ./...

## Agent behavior

Answer: Ask before assuming

## Unknown Section

Answer: Must not appear
`), 0o644); err != nil {
		t.Fatalf("WriteFile(project-brief.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".specharbor", "rules", "global.md"), []byte("# Global\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(global.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# Agents\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENTS.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/project\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("WriteFile(package.json) error = %v", err)
	}

	var output bytes.Buffer
	if err := execute([]string{"prompt", "context-aware-prompts", "--role", "spec-author"}, &output); err != nil {
		t.Fatalf("execute(prompt) error = %v", err)
	}

	promptOutput := output.String()
	for _, want := range []string{
		"# Spec Author Agent",
		"- `.specharbor/project-brief.md`",
		"## Project Context",
		"### User-confirmed context",
		"- Stack: Go",
		"- Test command: go test ./...",
		"- Agent behavior: Ask before assuming",
		"### Detected facts",
		"- Agent rules: AGENTS.md",
		"- Agent rules: .specharbor/rules/",
		"Source: go.mod",
		"Confidence: high",
		"### Suggested assumptions",
		"Build command may be `go build ./...`",
		"### Conflict notes",
		"detected Stack includes Node.js from package.json",
		"Do not invent stack, architecture, commands, persistence decisions, workflow decisions, or project direction.",
	} {
		if !strings.Contains(promptOutput, want) {
			t.Fatalf("prompt output = %q, want to contain %q", promptOutput, want)
		}
	}
	if strings.Contains(promptOutput, "Must not appear") {
		t.Fatalf("prompt output = %q, want unknown project brief section omitted", promptOutput)
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
	createAuthoredOpenSpecChange(t, root, "implement-validation-foundation", nil)

	var output bytes.Buffer
	if err := execute([]string{"validate", "implement-validation-foundation"}, &output); err != nil {
		t.Fatalf("execute(validate) error = %v", err)
	}

	want := `SpecHarbor change is valid.
Change: implement-validation-foundation
Checked path: openspec/changes/implement-validation-foundation
Required files: 5
Errors: 0
Warnings: 0
`
	if output.String() != want {
		t.Fatalf("validate output = %q, want %q", output.String(), want)
	}
}

func TestExecuteValidatePrintsInvalidReportForMissingRequiredFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createAuthoredOpenSpecChange(t, root, "implement-validation-foundation", map[string]string{
		"proposal.md": "",
		"risks.md":    "",
	})

	var output bytes.Buffer
	err := execute([]string{"validate", "implement-validation-foundation"}, &output)
	assertExitCode(t, err, 1)

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is invalid.",
		"Change: implement-validation-foundation",
		"Checked path: openspec/changes/implement-validation-foundation",
		"Errors:",
		"- [error] required_file_missing: Missing required file: proposal.md (openspec/changes/implement-validation-foundation/proposal.md)",
		"- [error] required_file_missing: Missing required file: risks.md (openspec/changes/implement-validation-foundation/risks.md)",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
}

func TestExecuteValidatePrintsWarningsAndExitsZeroForWarningsOnlyChange(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createAuthoredOpenSpecChange(t, root, "warning-change", map[string]string{
		"risks.md": "# Risks\n\n## Risks\n\n- Strict rules could reject existing changes.\n",
	})

	var output bytes.Buffer
	if err := execute([]string{"validate", "warning-change"}, &output); err != nil {
		t.Fatalf("execute(validate) error = %v, want exit 0 for warnings-only result", err)
	}

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is valid.",
		"Errors: 0",
		"Warnings: 1",
		"Warnings:",
		"- [warning] risks_mitigation_missing: Risks are listed without mitigation notes. (openspec/changes/warning-change/risks.md)",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
}

func TestExecuteValidateGroupsErrorsAndWarningsAndExitsNonZero(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createAuthoredOpenSpecChange(t, root, "broken-change", map[string]string{
		"tasks.md": "",
		"risks.md": "# Risks\n\n## Risks\n\n- Strict rules could reject existing changes.\n",
	})

	var output bytes.Buffer
	err := execute([]string{"validate", "broken-change"}, &output)
	assertExitCode(t, err, 1)

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is invalid.",
		"Errors:",
		"- [error] required_file_missing: Missing required file: tasks.md (openspec/changes/broken-change/tasks.md)",
		"Warnings:",
		"- [warning] risks_mitigation_missing: Risks are listed without mitigation notes. (openspec/changes/broken-change/risks.md)",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
}

func TestExecuteValidateReportsEmptyFileAsErrorWithSuppressedContentRules(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createAuthoredOpenSpecChange(t, root, "empty-file-change", nil)
	emptyPath := filepath.Join(root, "openspec", "changes", "empty-file-change", "design.md")
	if err := os.WriteFile(emptyPath, []byte("  \n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var output bytes.Buffer
	err := execute([]string{"validate", "empty-file-change"}, &output)
	assertExitCode(t, err, 1)

	validateOutput := output.String()
	if !strings.Contains(validateOutput, "- [error] file_empty: File is empty. (openspec/changes/empty-file-change/design.md)") {
		t.Fatalf("validate output = %q, want file_empty error with path", validateOutput)
	}
	if strings.Contains(validateOutput, "file_missing_heading") {
		t.Fatalf("validate output = %q, want same-file content rules suppressed for empty file", validateOutput)
	}
}

func TestExecuteValidateReportsFreshBlankChangeAsValidWithBoilerplateWarnings(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var generateOutput bytes.Buffer
	if err := execute([]string{"generate", "fresh-blank-change", "--blank"}, &generateOutput); err != nil {
		t.Fatalf("execute(generate --blank) error = %v", err)
	}

	var output bytes.Buffer
	if err := execute([]string{"validate", "fresh-blank-change"}, &output); err != nil {
		t.Fatalf("execute(validate) error = %v, want exit 0 for fresh blank change", err)
	}

	validateOutput := output.String()
	for _, want := range []string{
		"SpecHarbor change is valid.",
		"Errors: 0",
		"Warnings: 5",
		"- [warning] boilerplate_only_content: Only starter boilerplate content found. (openspec/changes/fresh-blank-change/proposal.md)",
		"- [warning] boilerplate_only_content: Only starter boilerplate content found. (openspec/changes/fresh-blank-change/tasks.md)",
	} {
		if !strings.Contains(validateOutput, want) {
			t.Fatalf("validate output = %q, want to contain %q", validateOutput, want)
		}
	}
	if strings.Contains(validateOutput, "[error]") {
		t.Fatalf("validate output = %q, want no error findings for fresh blank change", validateOutput)
	}
}

func TestExecuteValidateRejectsUnsafeChangeIDs(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "traversal", id: "..", want: "change id must not contain '.' or '..' path sequences"},
		{name: "separator", id: "a/b", want: "change id must be a single path segment"},
		{name: "leading dot", id: ".hidden", want: "change id must not start with '.'"},
		{name: "unsafe character", id: "change$id", want: "change id contains unsupported character"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute([]string{"validate", test.id}, &output)
			if err == nil {
				t.Fatalf("execute(validate %q) error = nil, want %q", test.id, test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(validate %q) error = %q, want %q", test.id, err.Error(), test.want)
			}
		})
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

func TestExecuteReviewPrintsApprovedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createReviewOpenSpecChange(t, root, "implement-review-foundation", reviewTasks([]string{
		"Add review domain types.",
		"Add review filesystem port.",
		"Add review use case.",
		"Add local filesystem compatibility.",
		"Add deterministic task parsing.",
		"Wire the review CLI command.",
		"Format the approved report.",
		"Format the needs-work report.",
		"Format the invalid report.",
		"Add CLI exit-code behavior.",
		"Run gofmt.",
		"Run go test ./...",
	}, nil), nil)

	var output bytes.Buffer
	if err := execute([]string{"review", "implement-review-foundation"}, &output); err != nil {
		t.Fatalf("execute(review) error = %v", err)
	}

	want := `SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: approved
Tasks: 12 total, 12 completed, 0 incomplete
Findings: 0
`
	if output.String() != want {
		t.Fatalf("review output = %q, want %q", output.String(), want)
	}
}

func TestExecuteReviewPrintsNeedsWorkReportAndReturnsNonZero(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createReviewOpenSpecChange(t, root, "implement-review-foundation", reviewTasks([]string{
		"Add review domain types.",
		"Add review filesystem port.",
		"Add review use case.",
		"Add local filesystem compatibility.",
		"Add deterministic task parsing.",
		"Wire the review CLI command.",
		"Format the approved report.",
		"Format the needs-work report.",
		"Format the invalid report.",
		"Add CLI exit-code behavior.",
	}, []string{
		"Run go test ./...",
		"Update this tasks.md by checking off only completed tasks.",
	}), nil)

	var output bytes.Buffer
	err := execute([]string{"review", "implement-review-foundation"}, &output)
	assertExitCode(t, err, 1)

	want := `SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: needs-work
Tasks: 12 total, 10 completed, 2 incomplete

Findings:
- [warning] incomplete_task: Task is not completed: Run go test ./...
- [warning] incomplete_task: Task is not completed: Update this tasks.md by checking off only completed tasks.
`
	if output.String() != want {
		t.Fatalf("review output = %q, want %q", output.String(), want)
	}
}

func TestExecuteReviewPrintsInvalidReportAndReturnsNonZero(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createReviewOpenSpecChange(t, root, "implement-review-foundation", "- [x] Done\n", map[string]bool{"risks.md": true})

	var output bytes.Buffer
	err := execute([]string{"review", "implement-review-foundation"}, &output)
	assertExitCode(t, err, 1)

	want := `SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: invalid
Tasks: 0 total, 0 completed, 0 incomplete

Findings:
- [error] required_file_missing: Missing required file: risks.md
`
	if output.String() != want {
		t.Fatalf("review output = %q, want %q", output.String(), want)
	}
}

func TestExecuteReviewRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing change id",
			args: []string{"review"},
			want: "change id is required",
		},
		{
			name: "unsupported json flag",
			args: []string{"review", "change", "--json"},
			want: "unsupported flag: --json",
		},
		{
			name: "unsupported format flag",
			args: []string{"review", "change", "--format", "json"},
			want: "unsupported flag: --format",
		},
		{
			name: "unsupported ai flag",
			args: []string{"review", "change", "--ai"},
			want: "unsupported flag: --ai",
		},
		{
			name: "unsupported github flag",
			args: []string{"review", "change", "--github"},
			want: "unsupported flag: --github",
		},
		{
			name: "unsupported gitlab flag",
			args: []string{"review", "change", "--gitlab"},
			want: "unsupported flag: --gitlab",
		},
		{
			name: "unsupported diff flag",
			args: []string{"review", "change", "--diff"},
			want: "unsupported flag: --diff",
		},
		{
			name: "unsupported fix flag",
			args: []string{"review", "change", "--fix"},
			want: "unsupported flag: --fix",
		},
		{
			name: "unsupported agent flag",
			args: []string{"review", "change", "--agent"},
			want: "unsupported flag: --agent",
		},
		{
			name: "extra argument",
			args: []string{"review", "change", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "unsafe traversal change id",
			args: []string{"review", "../unsafe"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe absolute change id",
			args: []string{"review", "/unsafe"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe slash change id",
			args: []string{"review", "bad/id"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe backslash change id",
			args: []string{"review", `bad\id`},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe colon change id",
			args: []string{"review", "bad:id"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe dot change id",
			args: []string{"review", "."},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe dotdot change id",
			args: []string{"review", ".."},
			want: "change id must be a safe single path segment",
		},
		{
			name: "leading dash change id",
			args: []string{"review", "-bad-id"},
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
		})
	}
}

func TestExecuteArchivePrintsSuccessReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createOpenSpecChange(t, root, "implement-archive-foundation", domain.RequiredOpenSpecChangeFiles())

	var output bytes.Buffer
	if err := execute([]string{"archive", "implement-archive-foundation"}, &output); err != nil {
		t.Fatalf("execute(archive) error = %v", err)
	}

	archiveDate := currentLocalArchiveDate()
	want := `SpecHarbor change archived.
Change: implement-archive-foundation
Source: openspec/changes/implement-archive-foundation
Archive: openspec/archive/` + archiveDate + `/implement-archive-foundation
Archive date: ` + archiveDate + `
Moved: yes

Moved directory:
- openspec/changes/implement-archive-foundation -> openspec/archive/` + archiveDate + `/implement-archive-foundation
`
	if output.String() != want {
		t.Fatalf("archive output = %q, want %q", output.String(), want)
	}

	assertPathDoesNotExist(t, root, "openspec/changes/implement-archive-foundation")
	assertPathExists(t, root, "openspec/archive/"+archiveDate+"/implement-archive-foundation")
}

func TestFormatLocalArchiveDate(t *testing.T) {
	location := time.FixedZone("UTC-3", -3*60*60)
	now := time.Date(2026, 6, 6, 23, 59, 0, 0, location)

	if got := formatLocalArchiveDate(now); got != "2026-06-06" {
		t.Fatalf("formatLocalArchiveDate() = %q, want 2026-06-06", got)
	}
}

func TestExecuteArchiveRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing change id",
			args: []string{"archive"},
			want: "change id is required",
		},
		{
			name: "unsupported force flag",
			args: []string{"archive", "change", "--force"},
			want: "unsupported flag: --force",
		},
		{
			name: "unsupported date flag",
			args: []string{"archive", "change", "--date", "2026-06-06"},
			want: "unsupported flag: --date",
		},
		{
			name: "unsupported dry run flag",
			args: []string{"archive", "change", "--dry-run"},
			want: "unsupported flag: --dry-run",
		},
		{
			name: "unsupported metadata flag",
			args: []string{"archive", "change", "--metadata"},
			want: "unsupported flag: --metadata",
		},
		{
			name: "unsupported summary flag",
			args: []string{"archive", "change", "--summary"},
			want: "unsupported flag: --summary",
		},
		{
			name: "unsupported github flag",
			args: []string{"archive", "change", "--github"},
			want: "unsupported flag: --github",
		},
		{
			name: "unsupported gitlab flag",
			args: []string{"archive", "change", "--gitlab"},
			want: "unsupported flag: --gitlab",
		},
		{
			name: "extra argument",
			args: []string{"archive", "change", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "unsafe traversal change id",
			args: []string{"archive", "../unsafe"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe absolute change id",
			args: []string{"archive", "/unsafe"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe slash change id",
			args: []string{"archive", "bad/id"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe backslash change id",
			args: []string{"archive", `bad\id`},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe colon change id",
			args: []string{"archive", "bad:id"},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe dot change id",
			args: []string{"archive", "."},
			want: "change id must be a safe single path segment",
		},
		{
			name: "unsafe dotdot change id",
			args: []string{"archive", ".."},
			want: "change id must be a safe single path segment",
		},
		{
			name: "leading dash change id",
			args: []string{"archive", "-bad-id"},
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
			assertPathDoesNotExist(t, root, "openspec/archive")
		})
	}
}

func TestExecuteScanPrintsPopulatedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeScanFile(t, root, "go.mod", "module example")
	writeScanFile(t, root, "package.json", "{}")
	writeScanFile(t, root, "package-lock.json", "{}")
	writeScanFile(t, root, "Dockerfile", "FROM scratch")
	if err := os.MkdirAll(filepath.Join(root, ".github", "workflows"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.github/workflows) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("MkdirAll(openspec/changes) error = %v", err)
	}
	writeScanFile(t, root, "openspec/project.md", "project")

	var output bytes.Buffer
	if err := execute([]string{"scan"}, &output); err != nil {
		t.Fatalf("execute(scan) error = %v", err)
	}

	want := "SpecHarbor project scan completed.\n" +
		"Project root: " + scanReportRoot(t) + "\n" +
		"\nDetected ecosystems:\n- Go: go.mod\n- Node: package.json\n" +
		"\nPackage managers:\n- npm: package-lock.json\n" +
		"\nTest command hints:\n- go test ./...\n- npm test\n" +
		"\nCI:\n- GitHub Actions: .github/workflows/\n" +
		"\nContainers/deployment:\n- Dockerfile\n" +
		"\nSpecHarbor/OpenSpec:\n- openspec/project.md\n- openspec/changes/\n" +
		"\nNotes:\n- No Kubernetes manifests detected.\n"
	if output.String() != want {
		t.Fatalf("scan output = %q, want %q", output.String(), want)
	}
}

func TestExecuteScanPrintsEmptyReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	if err := execute([]string{"scan"}, &output); err != nil {
		t.Fatalf("execute(scan) error = %v", err)
	}

	want := "SpecHarbor project scan completed.\n" +
		"Project root: " + scanReportRoot(t) + "\n" +
		"\nDetected ecosystems:\n- none detected\n" +
		"\nPackage managers:\n- none detected\n" +
		"\nTest command hints:\n- none detected\n" +
		"\nCI:\n- none detected\n" +
		"\nContainers/deployment:\n- none detected\n" +
		"\nSpecHarbor/OpenSpec:\n- none detected\n" +
		"\nNotes:\n- No known project signals detected.\n"
	if output.String() != want {
		t.Fatalf("scan output = %q, want %q", output.String(), want)
	}
}

func TestExecuteScanIsStackAgnostic(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeScanFile(t, root, "package.json", "{}")
	writeScanFile(t, root, "Dockerfile", "FROM scratch")
	writeScanFile(t, root, ".gitlab-ci.yml", "stages: []")

	var output bytes.Buffer
	if err := execute([]string{"scan"}, &output); err != nil {
		t.Fatalf("execute(scan) error = %v", err)
	}

	scanOutput := output.String()
	for _, want := range []string{
		"- Node: package.json",
		"- GitLab CI: .gitlab-ci.yml",
		"- Dockerfile",
	} {
		if !strings.Contains(scanOutput, want) {
			t.Fatalf("scan output = %q, want to contain %q", scanOutput, want)
		}
	}
	if strings.Contains(scanOutput, "- Go: go.mod") {
		t.Fatalf("scan output = %q, want no Go detection in a non-Go project", scanOutput)
	}
}

func TestExecuteScanRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "positional argument",
			args: []string{"scan", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "unsupported json flag",
			args: []string{"scan", "--json"},
			want: "unsupported flag: --json",
		},
		{
			name: "unsupported path flag",
			args: []string{"scan", "--path", "/tmp"},
			want: "unsupported flag: --path",
		},
		{
			name: "unsupported deep flag",
			args: []string{"scan", "--deep"},
			want: "unsupported flag: --deep",
		},
		{
			name: "unsupported ai flag",
			args: []string{"scan", "--ai"},
			want: "unsupported flag: --ai",
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
		})
	}
}

func TestExecuteScanReturnsZeroWhenReportPrinted(t *testing.T) {
	t.Chdir(t.TempDir())

	var output bytes.Buffer
	err := execute([]string{"scan"}, &output)
	if err != nil {
		t.Fatalf("execute(scan) error = %v, want nil", err)
	}
	if !strings.HasPrefix(output.String(), "SpecHarbor project scan completed.\n") {
		t.Fatalf("scan output = %q, want completion line", output.String())
	}
}

func TestExecuteContextDiscoverPrintsStructuredReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeScanFile(t, root, "go.mod", "module example.com/project\n")
	writeScanFile(t, root, "AGENTS.md", "# Agents\n\nUse Hexagonal Architecture.\n")
	if err := os.MkdirAll(filepath.Join(root, "openspec", "specs", "architecture"), 0o755); err != nil {
		t.Fatalf("MkdirAll(openspec/specs/architecture) error = %v", err)
	}
	writeScanFile(t, root, "openspec/specs/architecture/spec.md", "# Architecture\n\nClean Architecture applies.\n")

	var output bytes.Buffer
	if err := execute([]string{"context", "discover"}, &output); err != nil {
		t.Fatalf("execute(context discover) error = %v", err)
	}

	report := output.String()
	for _, want := range []string{
		"Detected project context:",
		"User-confirmed context:",
		"- none detected",
		"Detected facts:",
		"- Stack: Go",
		"  Source: go.mod",
		"  Classification: detected_fact",
		"  Confidence: high",
		"- Agent rules: AGENTS.md",
		"- Architecture: Hexagonal Architecture",
		"- Architecture: Clean Architecture",
		"Suggested assumptions:",
		"- Test command: go test ./...",
		"  Classification: suggested_assumption",
		"Notes:",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("context discover output = %q, want %q", report, want)
		}
	}
	if strings.Contains(report, "module example.com/project") {
		t.Fatalf("context discover output dumped raw go.mod contents: %q", report)
	}
}

func TestExecuteContextDiscoverPrintsConfirmedContextFirst(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	writeScanFile(t, root, ".specharbor/project-brief.md", `# Project Brief

## Stack

Answer: Go

## Commands

### Test

Answer: go test ./...
`)
	writeScanFile(t, root, "go.mod", "module example.com/project\n")

	var output bytes.Buffer
	if err := execute([]string{"context", "discover"}, &output); err != nil {
		t.Fatalf("execute(context discover) error = %v", err)
	}

	report := output.String()
	confirmedIndex := strings.Index(report, "User-confirmed context:")
	factsIndex := strings.Index(report, "Detected facts:")
	if confirmedIndex == -1 || factsIndex == -1 || confirmedIndex > factsIndex {
		t.Fatalf("context discover output = %q, want confirmed context before facts", report)
	}
	for _, want := range []string{
		"- Stack: Go\n  Source: .specharbor/project-brief.md (Stack)\n  Classification: user_confirmed_context\n  Confidence: high",
		"- Test command: go test ./...\n  Source: .specharbor/project-brief.md (Test)\n  Classification: user_confirmed_context\n  Confidence: high",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("context discover output = %q, want %q", report, want)
		}
	}
}

func TestExecuteContextDiscoverPrintsEmptyReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	if err := execute([]string{"context", "discover"}, &output); err != nil {
		t.Fatalf("execute(context discover) error = %v", err)
	}

	want := `Detected project context:

User-confirmed context:
- none detected

Detected facts:
- none detected

Suggested assumptions:
- none detected

Notes:
- No supported context sources detected.
`
	if output.String() != want {
		t.Fatalf("context discover output = %q, want %q", output.String(), want)
	}
}

func TestExecuteContextIndexPrintsMetadataOnlyReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeScanFile(t, root, "README.md", "# Project\nraw project prose must not appear\n")
	writeScanFile(t, root, "go.mod", "module example.com/project\n")

	var output bytes.Buffer
	if err := execute([]string{"context", "index"}, &output); err != nil {
		t.Fatalf("execute(context index) error = %v", err)
	}

	report := output.String()
	for _, want := range []string{
		"Repository context index:",
		"Mode: report",
		"Status: built",
		"Path: .specharbor/context-index.json",
		"Schema version: 1",
		"Indexed files: 2",
		"Truncated: no",
		"Limits:",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("context index output = %q, want %q", report, want)
		}
	}
	for _, unwanted := range []string{"raw project prose", "module example.com/project"} {
		if strings.Contains(report, unwanted) {
			t.Fatalf("context index output dumped raw content %q: %q", unwanted, report)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".specharbor", "context-index.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("context index without --write created index file, stat err = %v", err)
	}
}

func TestExecuteContextIndexWriteAndCheck(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeScanFile(t, root, "README.md", "# Project\n")

	var output bytes.Buffer
	if err := execute([]string{"context", "index", "--write"}, &output); err != nil {
		t.Fatalf("execute(context index --write) error = %v", err)
	}
	if !strings.Contains(output.String(), "Status: written") {
		t.Fatalf("write output = %q, want written status", output.String())
	}

	indexContents, err := os.ReadFile(filepath.Join(root, ".specharbor", "context-index.json"))
	if err != nil {
		t.Fatalf("ReadFile(context-index.json) error = %v", err)
	}
	if !strings.Contains(string(indexContents), `"path": "README.md"`) {
		t.Fatalf("index contents = %q, want README metadata", string(indexContents))
	}
	if strings.Contains(string(indexContents), "# Project") {
		t.Fatalf("index stored raw README contents: %s", string(indexContents))
	}

	output.Reset()
	if err := execute([]string{"context", "index", "--check"}, &output); err != nil {
		t.Fatalf("execute(context index --check) error = %v", err)
	}
	if !strings.Contains(output.String(), "Status: current") {
		t.Fatalf("check output = %q, want current status", output.String())
	}

	writeScanFile(t, root, "README.md", "# Project changed\n")
	output.Reset()
	err = execute([]string{"context", "index", "--check"}, &output)
	var exitErr ExitError
	if err == nil || !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("execute(stale check) error = %v, want ExitError{1}", err)
	}
	if !strings.Contains(output.String(), "Status: stale") || !strings.Contains(output.String(), "Stale reasons:") {
		t.Fatalf("stale check output = %q, want stale report", output.String())
	}
}

func TestExecuteContextIndexCheckMissingAndInvalidExitNonZero(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	err := execute([]string{"context", "index", "--check"}, &output)
	var exitErr ExitError
	if err == nil || !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("execute(missing check) error = %v, want ExitError{1}", err)
	}
	if !strings.Contains(output.String(), "Status: missing") {
		t.Fatalf("missing check output = %q, want missing", output.String())
	}

	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	writeScanFile(t, root, ".specharbor/context-index.json", "{not json")

	output.Reset()
	err = execute([]string{"context", "index", "--check"}, &output)
	if err == nil || !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("execute(invalid check) error = %v, want ExitError{1}", err)
	}
	if !strings.Contains(output.String(), "Status: invalid") {
		t.Fatalf("invalid check output = %q, want invalid", output.String())
	}
}

func TestExecuteContextRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing subcommand", args: []string{"context"}, want: "context subcommand is required: discover, index, retrieve, or github"},
		{name: "unsupported flag before subcommand", args: []string{"context", "--json"}, want: "unsupported flag: --json"},
		{name: "unsupported subcommand", args: []string{"context", "update"}, want: "unsupported context subcommand: update"},
		{name: "extra argument", args: []string{"context", "discover", "extra"}, want: "unexpected argument: extra"},
		{name: "json flag", args: []string{"context", "discover", "--json"}, want: "unsupported flag: --json"},
		{name: "path flag", args: []string{"context", "discover", "--path", "/tmp"}, want: "unsupported flag: --path"},
		{name: "deep flag", args: []string{"context", "discover", "--deep"}, want: "unsupported flag: --deep"},
		{name: "github flag", args: []string{"context", "discover", "--github"}, want: "unsupported flag: --github"},
		{name: "rag flag", args: []string{"context", "discover", "--rag"}, want: "unsupported flag: --rag"},
		{name: "index extra argument", args: []string{"context", "index", "extra"}, want: "unexpected argument: extra"},
		{name: "index unsupported flag", args: []string{"context", "index", "--json"}, want: "unsupported flag: --json"},
		{name: "index path flag", args: []string{"context", "index", "--path", "/tmp"}, want: "unsupported flag: --path"},
		{name: "index deep flag", args: []string{"context", "index", "--deep"}, want: "unsupported flag: --deep"},
		{name: "index github flag", args: []string{"context", "index", "--github"}, want: "unsupported flag: --github"},
		{name: "index rag flag", args: []string{"context", "index", "--rag"}, want: "unsupported flag: --rag"},
		{name: "index write check conflict", args: []string{"context", "index", "--write", "--check"}, want: "context index write and check flags cannot be used together"},
		{name: "index duplicate write", args: []string{"context", "index", "--write", "--write"}, want: "context index write flag specified more than once"},
		{name: "index duplicate check", args: []string{"context", "index", "--check", "--check"}, want: "context index check flag specified more than once"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)

			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(%v) error = %v, want %q", test.args, err, test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(%v) output = %q, want empty", test.args, output.String())
			}
		})
	}
}

func TestExecuteWorkflowPrintsRecommendedWorkflow(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	if err := execute([]string{"workflow"}, &output); err != nil {
		t.Fatalf("execute(workflow) error = %v", err)
	}

	if output.String() != expectedWorkflowReport() {
		t.Fatalf("workflow output = %q, want %q", output.String(), expectedWorkflowReport())
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", root, err)
	}
	if len(entries) != 0 {
		t.Fatalf("workflow created files in %q: %v", root, entries)
	}
}

func TestExecuteWorkflowRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "unsupported format flag",
			args: []string{"workflow", "--format", "json"},
			want: "unsupported flag: --format",
		},
		{
			name: "unsupported json flag",
			args: []string{"workflow", "--json"},
			want: "unsupported flag: --json",
		},
		{
			name: "extra argument",
			args: []string{"workflow", "extra"},
			want: "unexpected argument: extra",
		},
		{
			name: "show unsupported",
			args: []string{"workflow", "show"},
			want: "unexpected argument: show",
		},
		{
			name: "status unsupported",
			args: []string{"workflow", "status", "example-change"},
			want: "unexpected argument: status",
		},
		{
			name: "next unsupported",
			args: []string{"workflow", "next", "example-change"},
			want: "unexpected argument: next",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

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

func TestExecuteWorkflowOutputContainsRequiredFacts(t *testing.T) {
	t.Chdir(t.TempDir())

	var output bytes.Buffer
	if err := execute([]string{"workflow"}, &output); err != nil {
		t.Fatalf("execute(workflow) error = %v", err)
	}

	workflowOutput := output.String()
	for _, want := range []string{
		"Title: OpenSpec/SDD agent-driven workflow",
		"1. spec-author - Spec Author Agent",
		"2. architecture-reviewer - Architecture Reviewer Agent",
		"3. implementer - Implementer Agent",
		"4. test-engineer - Test Engineer Agent",
		"5. change-reviewer - Change Reviewer Agent",
		"6. commit - Commit",
		"7. pull-request - Pull Request",
		"8. merge - Merge",
		"9. archive - Archive",
		"Mode: agent-assisted",
		"Mode: manual",
		"Supported by SpecHarbor: yes",
		"Supported by SpecHarbor: no",
		"Advisory only: yes",
		"Requires: none",
		"Requires: change-reviewer",
		"Purpose: Create or refine the OpenSpec change package.",
		"Purpose: Commit the reviewed local changes manually.",
		`- specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"`,
		"- specharbor validate <change-id>",
		"- specharbor prompt <change-id> --role implementer",
		"- specharbor review <change-id>",
		"- specharbor archive <change-id>",
		"- none",
		"SpecHarbor does not commit, stage files, modify branches, push, or sign commits.",
		"SpecHarbor does not create PRs, call source-control APIs",
		"SpecHarbor does not merge, approve, inspect CI, trigger CI, or update remote repositories.",
		"specharbor workflow does not archive automatically",
		"Command suggestions are advisory and are not executed by this command.",
		"does not inspect Git, GitHub, GitLab, CI, provider APIs, agent CLIs, or remote workflow state.",
	} {
		if !strings.Contains(workflowOutput, want) {
			t.Fatalf("workflow output = %q, want to contain %q", workflowOutput, want)
		}
	}
}

func writeAIOutputFile(t *testing.T, root string, relativePath string, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(relativePath)), []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", relativePath, err)
	}
}

func validAIAssistedCLIOutput(overrides map[string]string) string {
	contents := map[string]string{
		"proposal.md": `# Proposal: AI-Assisted Import

## Problem

AI-assisted imports need a safe bridge.

## Goal

Import strict local file blocks into OpenSpec change files.
`,
		"design.md": `# Design: AI-Assisted Import

## Overview

The command parses all local source content before writing.

## Architecture

Domain parser rules stay pure and the use case writes approved paths only.
`,
		"tasks.md": `# Tasks

## Phase 1

- [ ] Implement the strict parser.
- [ ] Update documentation.
`,
		"acceptance-criteria.md": `# Acceptance Criteria

- Valid strict output writes only the required OpenSpec files.
- Validation runs after successful writes.
`,
		"risks.md": `# Risks

## Risks

- Untrusted AI output could include unsafe paths.

## Mitigations

- Reject unsafe filenames before writes.
`,
	}

	for fileName, content := range overrides {
		contents[fileName] = content
	}

	var blocks []string
	for _, fileName := range domain.RequiredOpenSpecChangeFiles() {
		blocks = append(blocks, "---FILE: "+fileName+"---\n"+contents[fileName]+"---END FILE---")
	}
	return strings.Join(blocks, "\n\n")
}

func writeScanFile(t *testing.T, root string, relativePath string, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(relativePath)), []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", relativePath, err)
	}
}

func scanReportRoot(t *testing.T) string {
	t.Helper()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	return root
}

func TestExecutePreservesHelpVersionAndUnknownCommandBehavior(t *testing.T) {
	var output bytes.Buffer
	if err := execute(nil, &output); err != nil {
		t.Fatalf("execute(no args) error = %v", err)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("no-arg output = %q, want Usage", output.String())
	}

	output.Reset()
	if err := execute([]string{"help"}, &output); err != nil {
		t.Fatalf("execute(help) error = %v", err)
	}
	if !strings.Contains(output.String(), "Usage:") {
		t.Fatalf("help output = %q, want Usage", output.String())
	}
	if !strings.Contains(output.String(), "workflow    Show the recommended SpecHarbor workflow") {
		t.Fatalf("help output = %q, want workflow command", output.String())
	}
	if !strings.Contains(output.String(), "brief       Collect confirmed project context") {
		t.Fatalf("help output = %q, want brief command", output.String())
	}

	output.Reset()
	if err := execute([]string{"version"}, &output); err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}
	wantVersion := version.Current().Format() + "\n"
	if output.String() != wantVersion {
		t.Fatalf("version output = %q, want %q", output.String(), wantVersion)
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

func TestExecuteVersionPrintsMultilineReport(t *testing.T) {
	var output bytes.Buffer
	if err := execute([]string{"version"}, &output); err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}

	want := `SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
`
	if output.String() != want {
		t.Fatalf("version output = %q, want %q", output.String(), want)
	}
}

func TestExecuteVersionRejectsUnsupportedArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "json flag", args: []string{"version", "--json"}, want: "unsupported flag: --json"},
		{name: "short flag", args: []string{"version", "--short"}, want: "unsupported flag: --short"},
		{name: "format flag", args: []string{"version", "--format", "json"}, want: "unsupported flag: --format"},
		{name: "extra argument", args: []string{"version", "extra"}, want: "unexpected argument: extra"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil {
				t.Fatalf("execute(%v) error = nil, want %q", test.args, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("execute(%v) error = %q, want %q", test.args, err.Error(), test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(%v) output = %q, want empty output", test.args, output.String())
			}
		})
	}
}

func TestExecuteVersionIsReadOnlyAndWorksOutsideProject(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)

	var output bytes.Buffer
	if err := execute([]string{"version"}, &output); err != nil {
		t.Fatalf("execute(version) error = %v", err)
	}

	if output.String() != version.Current().Format()+"\n" {
		t.Fatalf("version output = %q, want current metadata report", output.String())
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", root, err)
	}
	if len(entries) != 0 {
		t.Fatalf("version command created files in %q: %v", root, entries)
	}
}

func TestExecuteTopLevelVersionFlagRemainsUnsupported(t *testing.T) {
	var output bytes.Buffer
	err := execute([]string{"--version"}, &output)
	if err == nil {
		t.Fatalf("execute(--version) error = nil, want unknown command")
	}
	if err.Error() != "unknown command: --version" {
		t.Fatalf("execute(--version) error = %q, want unknown command", err.Error())
	}
	if output.String() != "" {
		t.Fatalf("execute(--version) output = %q, want empty output", output.String())
	}
}

func TestExecuteConfigShowPrintsConfigReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeLocalConfig(t, root, versionOneConfigYAML())

	var output bytes.Buffer
	if err := execute([]string{"config", "show"}, &output); err != nil {
		t.Fatalf("execute(config show) error = %v", err)
	}

	if output.String() != expectedConfigReport() {
		t.Fatalf("config show output = %q, want %q", output.String(), expectedConfigReport())
	}
}

func TestExecuteConfigAliasMatchesConfigShow(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	writeLocalConfig(t, root, versionOneConfigYAML())

	var showOutput bytes.Buffer
	if err := execute([]string{"config", "show"}, &showOutput); err != nil {
		t.Fatalf("execute(config show) error = %v", err)
	}

	var aliasOutput bytes.Buffer
	if err := execute([]string{"config"}, &aliasOutput); err != nil {
		t.Fatalf("execute(config) error = %v", err)
	}

	if aliasOutput.String() != showOutput.String() {
		t.Fatalf("config alias output = %q, want config show output %q", aliasOutput.String(), showOutput.String())
	}
}

func TestExecuteConfigRejectsUnsupportedFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "config json", args: []string{"config", "--json"}, want: "unsupported flag: --json"},
		{name: "show json", args: []string{"config", "show", "--json"}, want: "unsupported flag: --json"},
		{name: "format", args: []string{"config", "--format", "json"}, want: "unsupported flag: --format"},
		{name: "path", args: []string{"config", "--path", "/tmp"}, want: "unsupported flag: --path"},
		{name: "global", args: []string{"config", "--global"}, want: "unsupported flag: --global"},
		{name: "set flag", args: []string{"config", "--set"}, want: "unsupported flag: --set"},
		{name: "get flag", args: []string{"config", "--get"}, want: "unsupported flag: --get"},
		{name: "env", args: []string{"config", "--env"}, want: "unsupported flag: --env"},
		{name: "secrets", args: []string{"config", "--secrets"}, want: "unsupported flag: --secrets"},
		{name: "interactive", args: []string{"config", "--interactive"}, want: "unsupported flag: --interactive"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

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

func TestExecuteConfigRejectsUnsupportedSubcommands(t *testing.T) {
	tests := []string{"get", "set", "unset", "list", "edit", "other"}

	for _, subcommand := range tests {
		t.Run(subcommand, func(t *testing.T) {
			t.Chdir(t.TempDir())

			var output bytes.Buffer
			err := execute([]string{"config", subcommand}, &output)
			if err == nil {
				t.Fatalf("execute(config %s) error = nil, want unsupported subcommand", subcommand)
			}

			want := "unsupported config subcommand: " + subcommand
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("execute(config %s) error = %q, want %q", subcommand, err.Error(), want)
			}
			if output.String() != "" {
				t.Fatalf("execute(config %s) output = %q, want empty output", subcommand, output.String())
			}
		})
	}
}

func TestExecuteConfigRejectsExtraArgumentsAfterShow(t *testing.T) {
	t.Chdir(t.TempDir())

	var output bytes.Buffer
	err := execute([]string{"config", "show", "extra"}, &output)
	if err == nil {
		t.Fatalf("execute(config show extra) error = nil, want unexpected argument")
	}
	if !strings.Contains(err.Error(), "unexpected argument: extra") {
		t.Fatalf("execute(config show extra) error = %q, want unexpected argument: extra", err.Error())
	}
	if output.String() != "" {
		t.Fatalf("execute(config show extra) output = %q, want empty output", output.String())
	}
}

func TestExecuteConfigReturnsErrorsForConfigFailures(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, root string)
		want  string
	}{
		{
			name:  "missing config",
			setup: func(_ *testing.T, _ string) {},
			want:  "missing config file",
		},
		{
			name: "invalid yaml",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, "version: [\n")
			},
			want: "invalid config YAML",
		},
		{
			name: "unsupported version",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, "version: 2\n")
			},
			want: "unsupported config version",
		},
		{
			name: "unreadable config",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, versionOneConfigYAML())
				configPath := filepath.Join(root, ".specharbor", "config.yml")
				if err := os.Chmod(configPath, 0); err != nil {
					t.Fatalf("Chmod(config.yml) error = %v", err)
				}
				t.Cleanup(func() {
					_ = os.Chmod(configPath, 0o644)
				})
			},
			want: "unreadable config",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			test.setup(t, root)

			var output bytes.Buffer
			err := execute([]string{"config", "show"}, &output)
			if err == nil {
				t.Fatalf("execute(config show) error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(config show) error = %q, want %q", err.Error(), test.want)
			}
			if output.String() != "" {
				t.Fatalf("execute(config show) output = %q, want empty output", output.String())
			}
		})
	}
}

func writeLocalConfig(t *testing.T, root string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".specharbor", "config.yml"), []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(config.yml) error = %v", err)
	}
}

func versionOneConfigYAML() string {
	return `version: 1

defaults:
  agent_role: implementer
  generation_mode: blank

validation:
  require_all_change_files: true

review:
  require_completed_tasks: true

archive:
  date_layout: "2006-01-02"

scan:
  include_common_project_files: true

output:
  format: text
`
}

func expectedConfigReport() string {
	return `SpecHarbor configuration loaded.
Path: .specharbor/config.yml
Version: 1

Defaults:
- agent role: implementer
- generation mode: blank

Validation:
- require all change files: true

Review:
- require completed tasks: true

Archive:
- date layout: 2006-01-02

Scan:
- include common project files: true

Output:
- format: text
`
}

func expectedWorkflowReport() string {
	return `SpecHarbor recommended workflow.
Title: OpenSpec/SDD agent-driven workflow

Steps:
1. spec-author - Spec Author Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: none
   Purpose: Create or refine the OpenSpec change package.
   Commands:
   - specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>" (Create a guided OpenSpec change package.)
   - specharbor prompt <change-id> --role spec-author (Generate the spec author role prompt.)

2. architecture-reviewer - Architecture Reviewer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: spec-author
   Purpose: Review the proposed scope and design against architecture boundaries.
   Commands:
   - specharbor validate <change-id> (Check required OpenSpec change files.)
   - specharbor prompt <change-id> --role architecture-reviewer (Generate the architecture reviewer role prompt.)

3. implementer - Implementer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: architecture-reviewer
   Purpose: Apply the approved OpenSpec change.
   Commands:
   - specharbor prompt <change-id> --role implementer (Generate the implementer role prompt.)

4. test-engineer - Test Engineer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: implementer
   Purpose: Add or run focused verification for the implemented change.
   Commands:
   - specharbor prompt <change-id> --role test-engineer (Generate the test engineer role prompt.)

5. change-reviewer - Change Reviewer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: test-engineer
   Purpose: Review the final diff, task state, and verification evidence.
   Commands:
   - specharbor review <change-id> (Review local task completion and change package state.)
   - specharbor prompt <change-id> --role change-reviewer (Generate the change reviewer role prompt.)

6. commit - Commit
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Requires: change-reviewer
   Purpose: Commit the reviewed local changes manually.
   Commands:
   - none
   Safety:
   - SpecHarbor does not commit, stage files, modify branches, push, or sign commits.

7. pull-request - Pull Request
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Requires: commit
   Purpose: Open a pull request manually in your source-control workflow.
   Commands:
   - none
   Safety:
   - SpecHarbor does not create PRs, call source-control APIs, set reviewers, edit labels, or inspect remote branches.

8. merge - Merge
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Requires: pull-request
   Purpose: Merge manually after your review and CI process passes.
   Commands:
   - none
   Safety:
   - SpecHarbor does not merge, approve, inspect CI, trigger CI, or update remote repositories.

9. archive - Archive
   Mode: manual
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: merge
   Purpose: Archive the completed OpenSpec change after the work is merged or otherwise accepted.
   Commands:
   - specharbor archive <change-id> (Archive the completed OpenSpec change explicitly.)
   Safety:
   - specharbor workflow does not archive automatically; specharbor archive <change-id> remains an explicit user command.

Notes:
- Command suggestions are advisory and are not executed by this command.
- This command is read-only and does not inspect Git, GitHub, GitLab, CI, provider APIs, agent CLIs, or remote workflow state.
`
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

func installFakeAgentCommand(t *testing.T, name string, scriptBody string) string {
	t.Helper()

	directory := t.TempDir()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+scriptBody), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", directory+string(os.PathListSeparator)+oldPath)
	return path
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

// createAuthoredOpenSpecChange writes a fully authored change that validates
// with zero findings. Overrides replace a file's content; an empty override
// skips creating that file.
func createAuthoredOpenSpecChange(t *testing.T, root string, changeID string, overrides map[string]string) {
	t.Helper()

	authored := map[string]string{
		"proposal.md": `# Proposal: Example

## Problem

Users cannot tell whether a change package is ready for implementation.

## Goal

Validation reports content-quality findings with clear severities.
`,
		"design.md": `# Design: Example

## Overview

The validation pipeline reads files once and evaluates deterministic rules.

## Architecture

Rules live in the domain layer and the use case orchestrates them.
`,
		"tasks.md": `# Tasks

## Phase 1 - Implementation

- [x] Read the architecture documentation.
- [ ] Implement the validation rules.
`,
		"acceptance-criteria.md": `# Acceptance Criteria

- Validation reports findings with severity and code.
- Warnings alone keep the exit code at zero.
`,
		"risks.md": `# Risks

## Risks

- Strict rules could reject existing changes.

## Mitigations

- Quality findings stay warnings and are covered by tests.
`,
	}

	changeDirectory := filepath.Join(root, "openspec", "changes", changeID)
	if err := os.MkdirAll(changeDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	for _, file := range domain.RequiredOpenSpecChangeFiles() {
		contents := authored[file]
		if override, exists := overrides[file]; exists {
			if override == "" {
				continue
			}
			contents = override
		}
		if err := os.WriteFile(filepath.Join(changeDirectory, file), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}
}

func createReviewOpenSpecChange(t *testing.T, root string, changeID string, tasksContents string, skipFiles map[string]bool) {
	t.Helper()

	changeDirectory := filepath.Join(root, "openspec", "changes", changeID)
	if err := os.MkdirAll(changeDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	for _, file := range domain.RequiredOpenSpecChangeFiles() {
		if skipFiles[file] {
			continue
		}

		contents := file
		if file == "tasks.md" {
			contents = tasksContents
		}
		if err := os.WriteFile(filepath.Join(changeDirectory, file), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}
}

func reviewTasks(completed []string, incomplete []string) string {
	var builder strings.Builder
	for _, task := range completed {
		builder.WriteString("- [x] ")
		builder.WriteString(task)
		builder.WriteByte('\n')
	}
	for _, task := range incomplete {
		builder.WriteString("- [ ] ")
		builder.WriteString(task)
		builder.WriteByte('\n')
	}
	return builder.String()
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

func assertAgentAssistedDryRunOutputSafety(t *testing.T, output string, root string) {
	t.Helper()

	if strings.Contains(output, root) {
		t.Fatalf("agent-assisted output = %q, want no absolute project root", output)
	}

	lowerOutput := strings.ToLower(output)
	for _, forbidden := range []string{
		"http://",
		"https://",
		"localhost",
		"127.0.0.1",
		"api_key",
		"api key:",
		"secret:",
		"token:",
		"password",
		"sk-",
		"github.com/",
		"gitlab.com/",
		"registry.npmjs.org",
		"ghcr.io",
		"docker.io",
		"git commit",
		"git push",
		"git merge",
		"gh workflow",
		"workflow run",
		"go test",
		"go run",
		"curl ",
		"docker build",
		"kubectl",
		"terraform apply",
		"deploy to",
		"apply patch",
		"edit internal/",
		"modify internal/",
		"debug:",
		"trace:",
		"panic:",
		"specharbor change is valid",
		"validation status: valid",
		"validated: yes",
		"\nfiles written: yes",
		"prompt file:",
	} {
		if strings.Contains(lowerOutput, forbidden) {
			t.Fatalf("agent-assisted output contains unsafe dry-run artifact %q:\n%s", forbidden, output)
		}
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
