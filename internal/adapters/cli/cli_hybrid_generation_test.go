package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestExecuteGenerateHybridBuiltInPrintsReportAndDerivedType(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "add-login",
		"--hybrid",
		"--template", "feature",
		"--title", "Add login",
		"--summary", "Add an OpenSpec change for login",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate hybrid built-in) error = %v", err)
	}

	hybridOutput := output.String()
	for _, want := range []string{
		"SpecHarbor hybrid change generated.",
		"Change: add-login",
		"Mode: hybrid",
		"Source kind: built-in",
		"Source: feature",
		"Resolved source: builtin",
		"Resolved template: feature",
		"Title: Add login",
		"Summary: Add an OpenSpec change for login",
		"Type: feature",
		"Change path: openspec/changes/add-login",
		"Change directory: created",
		"Created files:",
		"- proposal.md",
		"- design.md",
		"- tasks.md",
		"- acceptance-criteria.md",
		"- risks.md",
		"Validation:",
		"Status: valid",
		"Required files: 5",
		"Errors: 0",
		"Warnings:",
		"Safety:",
		"- Provider APIs called: no",
		"- LLM APIs called: no",
		"- Agent commands executed: no",
		"- AI output file imported: no",
		"- Production code modified: no",
		"- Source-control commands run: no",
		"- Auto-commit, auto-push, PR, merge, or archive: no",
	} {
		if !strings.Contains(hybridOutput, want) {
			t.Fatalf("hybrid output = %q, want to contain %q", hybridOutput, want)
		}
	}
}

func TestExecuteGenerateHybridCustomOmittedTypeOmitsTypeLineAndPreservesToken(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createHybridCustomTemplateDirectory(t, root, "api-feature", true)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "add-payment-flow",
		"--hybrid",
		"--custom-template", "api-feature",
		"--title", "Add payments",
		"--summary", "Adds a payment flow.",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate hybrid custom) error = %v", err)
	}

	hybridOutput := output.String()
	if strings.Contains(hybridOutput, "Type:") {
		t.Fatalf("hybrid output = %q, want omitted Type line", hybridOutput)
	}
	for _, want := range []string{
		"Source kind: custom",
		"Source: api-feature",
		"Resolved source: custom",
		"Resolved template: api-feature",
		"Validation:",
		"Status: valid",
	} {
		if !strings.Contains(hybridOutput, want) {
			t.Fatalf("hybrid output = %q, want %q", hybridOutput, want)
		}
	}

	proposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "add-payment-flow", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if !strings.Contains(string(proposal), "Type: {{type}}") {
		t.Fatalf("proposal.md = %q, want unresolved type token", string(proposal))
	}
}

func TestExecuteGenerateHybridConfigBuiltInPrintsDerivedType(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeConfigTemplateConfig(t, root, `
    default-feature:
      source: builtin
      template: feature
`)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "add-login",
		"--hybrid",
		"--config-template", "default-feature",
		"--title", "Add login",
		"--summary", "Add login support",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate hybrid config built-in) error = %v", err)
	}

	hybridOutput := output.String()
	for _, want := range []string{
		"Source kind: config",
		"Source: default-feature",
		"Resolved source: builtin",
		"Resolved template: feature",
		"Type: feature",
	} {
		if !strings.Contains(hybridOutput, want) {
			t.Fatalf("hybrid output = %q, want %q", hybridOutput, want)
		}
	}
}

func TestExecuteGenerateHybridConfigCustomProvidedTypeRendersType(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createHybridCustomTemplateDirectory(t, root, "api-feature", true)
	writeConfigTemplateConfig(t, root, `
    api-feature:
      source: custom
      template: api-feature
`)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "add-payment-flow",
		"--hybrid",
		"--config-template", "api-feature",
		"--title", "Add payments",
		"--summary", "Adds a payment flow.",
		"--type", "docs",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate hybrid config custom) error = %v", err)
	}
	if !strings.Contains(output.String(), "Type: docs") {
		t.Fatalf("hybrid output = %q, want Type: docs", output.String())
	}

	proposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "add-payment-flow", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if !strings.Contains(string(proposal), "Type: docs") {
		t.Fatalf("proposal.md = %q, want rendered type", string(proposal))
	}
}

func TestExecuteGenerateHybridConfigRemotePrintsSanitizedFacts(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	downloadedBytes := []byte("zip bytes")
	checksum := domain.NewRemoteTemplateChecksumFromBytes(downloadedBytes).String()
	writeConfigTemplateConfig(t, root, `
    service-feature:
      source: remote
      url: https://example.com/templates/service-feature.zip
      checksum: `+checksum+`
      format: zip
`)
	fetcher := &cliFakeRemoteTemplateFetcher{result: domain.NewRemoteTemplateFetchResult(200, downloadedBytes)}
	reader := &cliFakeRemoteTemplateBundleReader{bundle: mustCLIHybridRemoteTemplateBundle(t)}
	withCLIRemoteTemplateFactories(t, fetcher, reader)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "add-service",
		"--hybrid",
		"--config-template", "service-feature",
		"--title", "Add service",
		"--summary", "Adds a service workflow.",
	}, &output)
	if err != nil {
		t.Fatalf("execute(generate hybrid remote) error = %v", err)
	}

	hybridOutput := output.String()
	for _, want := range []string{
		"Resolved source: remote",
		"Resolved source name: example.com",
		"Remote host: example.com",
		"Remote format: zip",
		"Checksum: sha256",
	} {
		if !strings.Contains(hybridOutput, want) {
			t.Fatalf("hybrid output = %q, want %q", hybridOutput, want)
		}
	}
	for _, forbidden := range []string{"?", "#", "token=", "user:", "supersecret"} {
		if strings.Contains(hybridOutput, forbidden) {
			t.Fatalf("hybrid output = %q, must not contain %q", hybridOutput, forbidden)
		}
	}
}

func TestExecuteGenerateHybridRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "duplicate hybrid", args: []string{"generate", "change", "--hybrid", "--hybrid"}, want: "hybrid generation flag specified more than once"},
		{name: "duplicate template", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--template", "bugfix", "--title", "Title", "--summary", "Summary"}, want: "template generation flag specified more than once"},
		{name: "duplicate type", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--type", "feature", "--type", "feature", "--title", "Title", "--summary", "Summary"}, want: "hybrid type flag specified more than once"},
		{name: "duplicate title", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--title", "Title", "--title", "Other", "--summary", "Summary"}, want: "hybrid title flag specified more than once"},
		{name: "duplicate summary", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--title", "Title", "--summary", "Summary", "--summary", "Other"}, want: "hybrid summary flag specified more than once"},
		{name: "missing source", args: []string{"generate", "change", "--hybrid", "--title", "Title", "--summary", "Summary"}, want: "hybrid source selector is required"},
		{name: "multiple source flags", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--custom-template", "api-feature", "--title", "Title", "--summary", "Summary"}, want: "hybrid requires exactly one source selector"},
		{name: "missing title", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--summary", "Summary"}, want: "hybrid title is required"},
		{name: "missing summary", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--title", "Title"}, want: "hybrid summary is required"},
		{name: "optional type invalid", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--title", "Title", "--summary", "Summary", "--type", "maintenance"}, want: "unsupported hybrid type: maintenance"},
		{name: "blank conflict", args: []string{"generate", "change", "--hybrid", "--blank", "--template", "feature", "--title", "Title", "--summary", "Summary"}, want: "hybrid and blank generation flags cannot be used together"},
		{name: "guided conflict", args: []string{"generate", "change", "--hybrid", "--guided", "--template", "feature", "--title", "Title", "--summary", "Summary"}, want: "hybrid and guided generation flags cannot be used together"},
		{name: "ai conflict", args: []string{"generate", "change", "--hybrid", "--ai-assisted", "--template", "feature", "--title", "Title", "--summary", "Summary"}, want: "hybrid and ai-assisted generation flags cannot be used together"},
		{name: "agent-assisted conflict", args: []string{"generate", "change", "--hybrid", "--agent-assisted", "--template", "feature", "--title", "Title", "--summary", "Summary"}, want: "hybrid and agent-assisted generation flags cannot be used together"},
		{name: "from-file conflict", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--from-file", "output.txt", "--title", "Title", "--summary", "Summary"}, want: "hybrid and from-file flags cannot be used together"},
		{name: "overwrite conflict", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--overwrite", "--title", "Title", "--summary", "Summary"}, want: "hybrid and overwrite flags cannot be used together"},
		{name: "agent conflict", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--agent", "codex", "--title", "Title", "--summary", "Summary"}, want: "hybrid and agent flags cannot be used together"},
		{name: "execute conflict", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--execute", "--title", "Title", "--summary", "Summary"}, want: "hybrid and execute flags cannot be used together"},
		{name: "extra argument", args: []string{"generate", "change", "--hybrid", "--template", "feature", "--title", "Title", "--summary", "Summary", "extra"}, want: "unexpected argument: extra"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)

			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("execute(%v) error = %v, want %q", test.args, err, test.want)
			}
			assertPathDoesNotExist(t, root, "openspec/changes/change")
		})
	}
}

func TestExecuteGenerateHybridBuiltInTypeMismatchWritesNothing(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "add-login",
		"--hybrid",
		"--template", "feature",
		"--type", "bugfix",
		"--title", "Add login",
		"--summary", "Add login support",
	}, &output)
	if err == nil || !strings.Contains(err.Error(), "hybrid type mismatch") {
		t.Fatalf("execute(generate hybrid mismatch) error = %v, want type mismatch", err)
	}
	assertPathDoesNotExist(t, root, "openspec/changes/add-login")
}

func TestExecuteGenerateHybridValidationErrorsExitNonZeroAfterReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createHybridCustomTemplateDirectory(t, root, "invalid-template", false)

	var output bytes.Buffer
	err := execute([]string{
		"generate", "invalid-change",
		"--hybrid",
		"--custom-template", "invalid-template",
		"--title", "Invalid change",
		"--summary", "Uses invalid task content.",
	}, &output)
	assertExitCode(t, err, 1)

	hybridOutput := output.String()
	for _, want := range []string{
		"SpecHarbor hybrid change generated.",
		"Validation:",
		"Status: invalid",
		"Errors:",
	} {
		if !strings.Contains(hybridOutput, want) {
			t.Fatalf("hybrid output = %q, want %q", hybridOutput, want)
		}
	}
}

func createHybridCustomTemplateDirectory(t *testing.T, root string, templateName string, valid bool) {
	t.Helper()

	templateDirectory := filepath.Join(root, ".specharbor", "templates", templateName)
	if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", templateDirectory, err)
	}

	files := map[string]string{
		"proposal.md": `# Proposal: {{title}}

## Summary

{{summary}}

## Problem

Type: {{type}}
Unknown: {{unknown}}
`,
		"design.md": `# Design: {{title}}

## Architecture

The implementation stays inside the existing architecture.
`,
		"tasks.md": `# Tasks: {{title}}

## Phase 1 - Implementation

- [ ] Implement the requested change.
`,
		"acceptance-criteria.md": `# Acceptance Criteria: {{title}}

- The generated change includes all required files.
`,
		"risks.md": `# Risks: {{title}}

## Risks

- Template output may need editing.

## Mitigations

- Validation reports remaining issues.
`,
	}
	if !valid {
		files["tasks.md"] = "# Tasks\n\nNo checkbox tasks here.\n"
	}

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		if err := os.WriteFile(filepath.Join(templateDirectory, requiredFile), []byte(files[requiredFile]), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", requiredFile, err)
		}
	}
}

func mustCLIHybridRemoteTemplateBundle(t *testing.T) domain.RemoteTemplateBundle {
	t.Helper()

	files := map[string]string{
		"proposal.md": `# Proposal: {{title}}

## Summary

{{summary}}

## Problem

Remote template type: {{type}}
`,
		"design.md": `# Design: {{title}}

## Architecture

Remote content remains static Markdown.
`,
		"tasks.md": `# Tasks: {{title}}

## Phase 1 - Implementation

- [ ] Implement the remote-templated change.
`,
		"acceptance-criteria.md": `# Acceptance Criteria: {{title}}

- Remote template output includes the required OpenSpec files.
`,
		"risks.md": `# Risks: {{title}}

## Risks

- Remote content is external.

## Mitigations

- Checksum verification runs before archive parsing.
`,
	}
	bundle, err := domain.NewRemoteTemplateBundle(files)
	if err != nil {
		t.Fatalf("NewRemoteTemplateBundle() error = %v", err)
	}
	return bundle
}
