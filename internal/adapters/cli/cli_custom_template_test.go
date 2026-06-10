package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestExecuteGenerateCustomTemplatePrintsCreatedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")

	var output bytes.Buffer
	if err := execute([]string{"generate", "add-payment-flow", "--custom-template", "api-feature"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor custom template change generated.
Change: add-payment-flow
Template: api-feature (custom)
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

	changeDirectory := filepath.Join(root, "openspec", "changes", "add-payment-flow")
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents, err := os.ReadFile(filepath.Join(changeDirectory, requiredFile))
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", requiredFile, err)
		}
		if !strings.Contains(string(contents), "Change id: add-payment-flow") {
			t.Fatalf("generated %q = %q, want substituted change id", requiredFile, string(contents))
		}
		if !strings.Contains(string(contents), "Title: {{title}}") {
			t.Fatalf("generated %q = %q, want verbatim title token when omitted", requiredFile, string(contents))
		}
		if !strings.Contains(string(contents), "Summary: {{summary}}") {
			t.Fatalf("generated %q = %q, want verbatim summary token when omitted", requiredFile, string(contents))
		}
	}
}

func TestExecuteGenerateCustomTemplateSubstitutesProvidedTitleAndSummary(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")

	var output bytes.Buffer
	args := []string{
		"generate", "add-payment-flow",
		"--custom-template", "api-feature",
		"--title", "Add payments",
		"--summary", "Adds a payment flow.",
	}
	if err := execute(args, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
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
	if strings.Contains(string(proposal), "{{title}}") || strings.Contains(string(proposal), "{{summary}}") {
		t.Fatalf("proposal.md = %q, want no unreplaced provided tokens", string(proposal))
	}
}

func TestExecuteGenerateCustomTemplateIgnoresExtraTemplateFiles(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")
	extraPath := filepath.Join(root, ".specharbor", "templates", "api-feature", "README.md")
	if err := os.WriteFile(extraPath, []byte("template notes"), 0o644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}

	var output bytes.Buffer
	if err := execute([]string{"generate", "add-payment-flow", "--custom-template", "api-feature"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	assertPathDoesNotExist(t, root, "openspec/changes/add-payment-flow/README.md")
}

func TestExecuteGenerateCustomTemplateSecondRunPrintsSkippedReport(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")

	var firstOutput bytes.Buffer
	if err := execute([]string{"generate", "add-payment-flow", "--custom-template", "api-feature"}, &firstOutput); err != nil {
		t.Fatalf("execute(generate) first run error = %v", err)
	}

	proposalPath := filepath.Join(root, "openspec", "changes", "add-payment-flow", "proposal.md")
	if err := os.WriteFile(proposalPath, []byte("manually edited"), 0o644); err != nil {
		t.Fatalf("WriteFile(proposal.md) error = %v", err)
	}

	var secondOutput bytes.Buffer
	if err := execute([]string{"generate", "add-payment-flow", "--custom-template", "api-feature"}, &secondOutput); err != nil {
		t.Fatalf("execute(generate) second run error = %v", err)
	}

	want := `SpecHarbor custom template change generated.
Change: add-payment-flow
Template: api-feature (custom)
Template source: .specharbor/templates/api-feature
Change path: openspec/changes/add-payment-flow
Change directory: existing
Skipped existing files:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
Only OpenSpec change files under openspec/changes/add-payment-flow/ were written.
`
	if secondOutput.String() != want {
		t.Fatalf("second generate output = %q, want %q", secondOutput.String(), want)
	}

	contents, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if string(contents) != "manually edited" {
		t.Fatalf("proposal.md = %q, want preserved manual edit", string(contents))
	}
}

func TestExecuteGenerateCustomTemplateErrorsWithoutCreatingChange(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		setup        func(t *testing.T, root string)
		want         string
	}{
		{
			name:         "invalid template name",
			templateName: "../escape",
			setup:        func(t *testing.T, root string) {},
			want:         "custom template name must be a single path segment",
		},
		{
			name:         "missing template directory",
			templateName: "missing-template",
			setup:        func(t *testing.T, root string) {},
			want:         "unknown custom template: missing-template. Expected directory: .specharbor/templates/missing-template",
		},
		{
			name:         "missing required template files",
			templateName: "api-feature",
			setup: func(t *testing.T, root string) {
				createCustomTemplateDirectory(t, root, "api-feature")
				for _, file := range []string{"design.md", "risks.md"} {
					if err := os.Remove(filepath.Join(root, ".specharbor", "templates", "api-feature", file)); err != nil {
						t.Fatalf("Remove(%q) error = %v", file, err)
					}
				}
			},
			want: "custom template api-feature is missing required files: design.md, risks.md",
		},
		{
			name:         "empty template file",
			templateName: "api-feature",
			setup: func(t *testing.T, root string) {
				createCustomTemplateDirectory(t, root, "api-feature")
				path := filepath.Join(root, ".specharbor", "templates", "api-feature", "design.md")
				if err := os.WriteFile(path, []byte("  \n\t\n"), 0o644); err != nil {
					t.Fatalf("WriteFile(design.md) error = %v", err)
				}
			},
			want: "custom template file api-feature/design.md is empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			createOpenSpecProject(t, root)
			test.setup(t, root)

			var output bytes.Buffer
			err := execute([]string{"generate", "new-change", "--custom-template", test.templateName}, &output)
			if err == nil {
				t.Fatalf("execute(generate) error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("execute(generate) error = %q, want %q", err.Error(), test.want)
			}
			assertPathDoesNotExist(t, root, "openspec/changes/new-change")
		})
	}
}

func TestExecuteGenerateCustomTemplateRequiresOpenSpecProject(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createCustomTemplateDirectory(t, root, "api-feature")

	var output bytes.Buffer
	err := execute([]string{"generate", "new-change", "--custom-template", "api-feature"}, &output)
	if err == nil {
		t.Fatalf("execute(generate) error = nil, want missing OpenSpec project error")
	}
	want := "OpenSpec project structure is missing. Run specharbor init first."
	if err.Error() != want {
		t.Fatalf("execute(generate) error = %q, want %q", err.Error(), want)
	}
	assertPathDoesNotExist(t, root, "openspec/changes/new-change")
}

func TestExecuteGenerateCustomTemplateRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "missing change id",
			args: []string{"generate", "--custom-template", "api-feature"},
			want: "change id is required",
		},
		{
			name: "missing template name value",
			args: []string{"generate", "change", "--custom-template"},
			want: "custom template name is required",
		},
		{
			name: "template name followed by flag",
			args: []string{"generate", "change", "--custom-template", "--blank"},
			want: "custom template name is required",
		},
		{
			name: "empty template name",
			args: []string{"generate", "change", "--custom-template", ""},
			want: "custom template name is required",
		},
		{
			name: "duplicate custom-template flag",
			args: []string{"generate", "change", "--custom-template", "one", "--custom-template", "two"},
			want: "custom-template generation flag specified more than once",
		},
		{
			name: "custom-template with blank",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--blank"},
			want: "custom-template and blank generation flags cannot be used together",
		},
		{
			name: "custom-template with template",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--template", "feature"},
			want: "custom-template and template generation flags cannot be used together",
		},
		{
			name: "custom-template with guided",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--guided"},
			want: "custom-template and guided generation flags cannot be used together",
		},
		{
			name: "custom-template with agent-assisted",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--agent-assisted"},
			want: "custom-template and agent-assisted generation flags cannot be used together",
		},
		{
			name: "custom-template with type",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--type", "feature"},
			want: "guided input flags require --guided",
		},
		{
			name: "custom-template with agent",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--agent", "codex"},
			want: "agent-assisted input flags require --agent-assisted",
		},
		{
			name: "custom-template with execute",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--execute"},
			want: "unsupported flag: --execute",
		},
		{
			name: "extra positional argument",
			args: []string{"generate", "change", "extra", "--custom-template", "api-feature"},
			want: "unexpected argument: extra",
		},
		{
			name: "unsupported flag",
			args: []string{"generate", "change", "--custom-template", "api-feature", "--force"},
			want: "unsupported flag: --force",
		},
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
			assertPathDoesNotExist(t, root, "openspec/changes/change")
		})
	}
}

func TestExecuteGenerateCustomTemplateAcceptsFlagBeforeChangeID(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")

	var output bytes.Buffer
	if err := execute([]string{"generate", "--custom-template", "api-feature", "order-independent"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	if !strings.Contains(output.String(), "Change: order-independent") {
		t.Fatalf("generate output = %q, want change id", output.String())
	}
	if !strings.Contains(output.String(), "Template: api-feature (custom)") {
		t.Fatalf("generate output = %q, want custom template label", output.String())
	}
}

func TestExecuteGenerateBuiltInTemplateUnchangedWhenCustomTemplateSharesName(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "feature")

	var output bytes.Buffer
	if err := execute([]string{"generate", "add-feature", "--template", "feature"}, &output); err != nil {
		t.Fatalf("execute(generate) error = %v", err)
	}

	want := `SpecHarbor template change generated.
Change: add-feature
Template: feature
Path: openspec/changes/add-feature
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
		t.Fatalf("generate output = %q, want byte-identical built-in report %q", output.String(), want)
	}

	proposal, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "add-feature", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if strings.Contains(string(proposal), "Change id:") {
		t.Fatalf("proposal.md = %q, want built-in content, not custom template content", string(proposal))
	}
}

func TestExecuteGenerateBuiltInTemplateKeepsUnknownNameErrorWhenCustomTemplateExists(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createCustomTemplateDirectory(t, root, "api-feature")

	var output bytes.Buffer
	err := execute([]string{"generate", "add-feature", "--template", "api-feature"}, &output)
	if err == nil {
		t.Fatalf("execute(generate) error = nil, want unknown template name error")
	}
	if err.Error() != "unknown template name: api-feature" {
		t.Fatalf("execute(generate) error = %q, want unknown template name: api-feature", err.Error())
	}
	assertPathDoesNotExist(t, root, "openspec/changes/add-feature")
}

func createCustomTemplateDirectory(t *testing.T, root string, templateName string) {
	t.Helper()

	templateDirectory := filepath.Join(root, ".specharbor", "templates", templateName)
	if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", templateDirectory, err)
	}
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents := "# Custom " + requiredFile + "\n\n" +
			"Change id: {{change_id}}\n" +
			"Title: {{title}}\n" +
			"Summary: {{summary}}\n"
		if err := os.WriteFile(filepath.Join(templateDirectory, requiredFile), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", requiredFile, err)
		}
	}
}
