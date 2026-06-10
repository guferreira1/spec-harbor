package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestExecuteGenerateConfigTemplateBuiltInPrintsCreatedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	writeConfigTemplateConfig(t, root, `
    default-feature:
      source: builtin
      template: feature
`)

	var output bytes.Buffer
	if err := execute([]string{"generate", "add-feature", "--config-template", "default-feature"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor config template change generated.
Change: add-feature
Config template: default-feature
Resolved source: builtin
Resolved template: feature
Change path: openspec/changes/add-feature
Change directory: created
Created files:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
Only OpenSpec change files under openspec/changes/add-feature/ were written.
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}

	proposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "add-feature", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if !strings.Contains(string(proposal), "## Proposed Solution") {
		t.Fatalf("proposal.md = %q, want built-in feature content", string(proposal))
	}
}

func TestExecuteGenerateConfigTemplateCustomPrintsCreatedReportAndPassesTitleSummary(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")
	writeConfigTemplateConfig(t, root, `
    api-feature:
      source: custom
      template: api-feature
`)

	var output bytes.Buffer
	args := []string{
		"generate", "add-payment-flow",
		"--config-template", "api-feature",
		"--title", "Add payments",
		"--summary", "Adds a payment flow.",
	}
	if err := execute(args, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor config template change generated.
Change: add-payment-flow
Config template: api-feature
Resolved source: custom
Resolved template: api-feature
Template source: .specharbor/templates/api-feature
Change path: openspec/changes/add-payment-flow
Change directory: created
Created files:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
Only OpenSpec change files under openspec/changes/add-payment-flow/ were written.
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}

	proposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "add-payment-flow", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if !strings.Contains(string(proposal), "Title: Add payments") {
		t.Fatalf("proposal.md = %q, want substituted title", string(proposal))
	}
	if !strings.Contains(string(proposal), "Summary: Adds a payment flow.") {
		t.Fatalf("proposal.md = %q, want substituted summary", string(proposal))
	}
}

func TestExecuteGenerateConfigTemplatePrintsSkippedExistingReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createOpenSpecChange(t, root, "add-feature", domain.RequiredOpenSpecChangeFiles())
	writeConfigTemplateConfig(t, root, `
    default-feature:
      source: builtin
      template: feature
`)

	var output bytes.Buffer
	if err := execute([]string{"generate", "add-feature", "--config-template", "default-feature"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor config template change generated.
Change: add-feature
Config template: default-feature
Resolved source: builtin
Resolved template: feature
Change path: openspec/changes/add-feature
Change directory: existing
Skipped existing files:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
Only OpenSpec change files under openspec/changes/add-feature/ were written.
`
	if output.String() != want {
		t.Fatalf("generate output = %q, want %q", output.String(), want)
	}
}

func TestExecuteGenerateConfigTemplateErrorsClearly(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, root string)
		args  []string
		want  string
	}{
		{
			name:  "missing config",
			setup: func(_ *testing.T, _ string) {},
			args:  []string{"generate", "new-change", "--config-template", "api-feature"},
			want:  "missing config file: .specharbor/config.yml",
		},
		{
			name: "invalid yaml",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, "version: [\n")
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: "invalid config YAML in .specharbor/config.yml: parse local config YAML",
		},
		{
			name: "missing version",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, "templates:\n  aliases: {}\n")
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: "missing config version in .specharbor/config.yml",
		},
		{
			name: "unsupported version",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, "version: 2\ntemplates:\n  aliases: {}\n")
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: "unsupported config version 2 in .specharbor/config.yml",
		},
		{
			name: "missing alias",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, "version: 1\ntemplates:\n  aliases: {}\n")
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: "config template alias not found: api-feature",
		},
		{
			name: "invalid cli alias",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    api-feature:
      source: builtin
      template: feature
`)
			},
			args: []string{"generate", "new-change", "--config-template", "../escape"},
			want: "config template alias must be a single path segment",
		},
		{
			name: "invalid config alias",
			setup: func(t *testing.T, root string) {
				writeLocalConfig(t, root, `version: 1
templates:
  aliases:
    nested/template:
      source: builtin
      template: feature
`)
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: `invalid config YAML in .specharbor/config.yml: invalid config template alias "nested/template"`,
		},
		{
			name: "unsupported source",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    api-feature:
      source: remote
      template: feature
`)
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: `invalid config YAML in .specharbor/config.yml: invalid config template alias "api-feature": unsupported config template source: remote`,
		},
		{
			name: "unsupported path field",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    api-feature:
      source: builtin
      template: feature
      path: ../templates/feature
`)
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: `invalid config YAML in .specharbor/config.yml: invalid config template alias "api-feature": unsupported config template field "path"`,
		},
		{
			name: "unknown builtin template",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    api-feature:
      source: builtin
      template: maintenance
`)
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: `invalid config YAML in .specharbor/config.yml: invalid config template alias "api-feature": unknown template name: maintenance`,
		},
		{
			name: "missing custom template",
			setup: func(t *testing.T, root string) {
				writeConfigTemplateConfig(t, root, `
    api-feature:
      source: custom
      template: api-feature
`)
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: "unknown custom template: api-feature. Expected directory: .specharbor/templates/api-feature",
		},
		{
			name: "missing custom template files",
			setup: func(t *testing.T, root string) {
				createCustomTemplateDirectory(t, root, "api-feature")
				if err := os.Remove(filepath.Join(root, ".specharbor", "templates", "api-feature", "risks.md")); err != nil {
					t.Fatalf("Remove(risks.md) error = %v", err)
				}
				writeConfigTemplateConfig(t, root, `
    api-feature:
      source: custom
      template: api-feature
`)
			},
			args: []string{"generate", "new-change", "--config-template", "api-feature"},
			want: "custom template api-feature is missing required files: risks.md",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)
			test.setup(t, root)

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
			assertPathDoesNotExist(t, root, "openspec/changes/new-change")
		})
	}
}

func TestExecuteGenerateConfigTemplateRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing alias value", args: []string{"generate", "change", "--config-template"}, want: "config template alias is required"},
		{name: "alias followed by flag", args: []string{"generate", "change", "--config-template", "--blank"}, want: "config template alias is required"},
		{name: "empty alias", args: []string{"generate", "change", "--config-template", ""}, want: "config template alias is required"},
		{name: "duplicate config-template", args: []string{"generate", "change", "--config-template", "one", "--config-template", "two"}, want: "config-template generation flag specified more than once"},
		{name: "with blank", args: []string{"generate", "change", "--config-template", "api-feature", "--blank"}, want: "config-template and blank generation flags cannot be used together"},
		{name: "with template", args: []string{"generate", "change", "--config-template", "api-feature", "--template", "feature"}, want: "config-template and template generation flags cannot be used together"},
		{name: "with custom-template", args: []string{"generate", "change", "--config-template", "api-feature", "--custom-template", "api-feature"}, want: "config-template and custom-template generation flags cannot be used together"},
		{name: "with guided", args: []string{"generate", "change", "--config-template", "api-feature", "--guided"}, want: "config-template and guided generation flags cannot be used together"},
		{name: "with agent-assisted", args: []string{"generate", "change", "--config-template", "api-feature", "--agent-assisted"}, want: "config-template and agent-assisted generation flags cannot be used together"},
		{name: "with ai-assisted", args: []string{"generate", "change", "--config-template", "api-feature", "--ai-assisted"}, want: "config-template and ai-assisted generation flags cannot be used together"},
		{name: "with execute", args: []string{"generate", "change", "--config-template", "api-feature", "--execute"}, want: "config-template and execute flags cannot be used together"},
		{name: "with type", args: []string{"generate", "change", "--config-template", "api-feature", "--type", "feature"}, want: "config-template and type flags cannot be used together"},
		{name: "with agent", args: []string{"generate", "change", "--config-template", "api-feature", "--agent", "codex"}, want: "config-template and agent flags cannot be used together"},
		{name: "with from-file", args: []string{"generate", "change", "--config-template", "api-feature", "--from-file", "agent-output.txt"}, want: "config-template and from-file flags cannot be used together"},
		{name: "with overwrite", args: []string{"generate", "change", "--config-template", "api-feature", "--overwrite"}, want: "config-template and overwrite flags cannot be used together"},
		{name: "duplicate title", args: []string{"generate", "change", "--config-template", "api-feature", "--title", "One", "--title", "Two"}, want: "config-template title flag specified more than once"},
		{name: "duplicate summary", args: []string{"generate", "change", "--config-template", "api-feature", "--summary", "One", "--summary", "Two"}, want: "config-template summary flag specified more than once"},
		{name: "unsupported flag", args: []string{"generate", "change", "--config-template", "api-feature", "--force"}, want: "unsupported flag: --force"},
		{name: "extra argument", args: []string{"generate", "change", "--config-template", "api-feature", "extra"}, want: "unexpected argument: extra"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)

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

func TestExecuteGenerateConfigTemplateKeepsNamespacesDisjoint(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "feature")
	writeConfigTemplateConfig(t, root, `
    feature:
      source: custom
      template: feature
`)

	var builtInOutput bytes.Buffer
	if err := execute([]string{"generate", "direct-built-in", "--template", "feature"}, &builtInOutput); err != nil {
		t.Fatalf("execute(direct built-in) error = %v", err)
	}
	builtInProposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "direct-built-in", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(direct built-in proposal) error = %v", err)
	}
	if strings.Contains(string(builtInProposal), "Change id:") {
		t.Fatalf("direct built-in proposal = %q, want built-in content", string(builtInProposal))
	}

	var customOutput bytes.Buffer
	if err := execute([]string{"generate", "direct-custom", "--custom-template", "feature"}, &customOutput); err != nil {
		t.Fatalf("execute(direct custom) error = %v", err)
	}
	customProposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "direct-custom", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(direct custom proposal) error = %v", err)
	}
	if !strings.Contains(string(customProposal), "Change id: direct-custom") {
		t.Fatalf("direct custom proposal = %q, want custom content", string(customProposal))
	}

	var configOutput bytes.Buffer
	if err := execute([]string{"generate", "config-alias", "--config-template", "feature"}, &configOutput); err != nil {
		t.Fatalf("execute(config alias) error = %v", err)
	}
	if !strings.Contains(configOutput.String(), "Config template: feature") {
		t.Fatalf("config output = %q, want config alias label", configOutput.String())
	}
	configProposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "config-alias", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(config proposal) error = %v", err)
	}
	if !strings.Contains(string(configProposal), "Change id: config-alias") {
		t.Fatalf("config proposal = %q, want custom content from config alias", string(configProposal))
	}
}

func writeConfigTemplateConfig(t *testing.T, root string, aliasesYAML string) {
	t.Helper()

	writeLocalConfig(t, root, `version: 1
templates:
  aliases:
`+aliasesYAML)
}
