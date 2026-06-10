package domain

import (
	"strings"
	"testing"
)

func TestAllowedCustomTemplateFilesMatchRequiredOpenSpecChangeFiles(t *testing.T) {
	allowed := AllowedCustomTemplateFiles()
	required := RequiredOpenSpecChangeFiles()

	if len(allowed) != len(required) {
		t.Fatalf("AllowedCustomTemplateFiles() = %v, want %v", allowed, required)
	}
	for index := range allowed {
		if allowed[index] != required[index] {
			t.Fatalf("AllowedCustomTemplateFiles() = %v, want %v", allowed, required)
		}
	}

	want := []string{"proposal.md", "design.md", "tasks.md", "acceptance-criteria.md", "risks.md"}
	for index := range want {
		if allowed[index] != want[index] {
			t.Fatalf("AllowedCustomTemplateFiles() = %v, want %v", allowed, want)
		}
	}
}

func TestTemplateSourceValues(t *testing.T) {
	if BuiltInTemplateSource != "built-in" {
		t.Fatalf("BuiltInTemplateSource = %q, want built-in", BuiltInTemplateSource)
	}
	if CustomTemplateSource != "custom" {
		t.Fatalf("CustomTemplateSource = %q, want custom", CustomTemplateSource)
	}
}

func validCustomTemplateFiles() map[string]string {
	files := make(map[string]string)
	for _, requiredFile := range RequiredOpenSpecChangeFiles() {
		files[requiredFile] = "# Content for " + requiredFile + "\n"
	}
	return files
}

func mustNewCustomTemplateName(t *testing.T, raw string) CustomTemplateName {
	t.Helper()

	name, err := NewCustomTemplateName(raw)
	if err != nil {
		t.Fatalf("NewCustomTemplateName(%q) error = %v", raw, err)
	}
	return name
}

func TestNewCustomTemplateAcceptsCompleteTemplate(t *testing.T) {
	name := mustNewCustomTemplateName(t, "api-feature")

	template, err := NewCustomTemplate(name, validCustomTemplateFiles())
	if err != nil {
		t.Fatalf("NewCustomTemplate() error = %v", err)
	}

	if template.Name().String() != "api-feature" {
		t.Fatalf("Name() = %q, want api-feature", template.Name().String())
	}
	files := template.Files()
	for _, requiredFile := range RequiredOpenSpecChangeFiles() {
		if files[requiredFile] != "# Content for "+requiredFile+"\n" {
			t.Fatalf("Files()[%q] = %q, want template content", requiredFile, files[requiredFile])
		}
	}
}

func TestNewCustomTemplateRejectsMissingFilesWithAggregatedError(t *testing.T) {
	name := mustNewCustomTemplateName(t, "api-feature")
	files := validCustomTemplateFiles()
	delete(files, "design.md")
	delete(files, "risks.md")

	_, err := NewCustomTemplate(name, files)
	if err == nil {
		t.Fatalf("NewCustomTemplate() error = nil, want missing files error")
	}
	want := "custom template api-feature is missing required files: design.md, risks.md"
	if err.Error() != want {
		t.Fatalf("NewCustomTemplate() error = %q, want %q", err.Error(), want)
	}
}

func TestNewCustomTemplateRejectsEmptyFiles(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{name: "empty", contents: ""},
		{name: "whitespace only", contents: " \n\t\n "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			name := mustNewCustomTemplateName(t, "api-feature")
			files := validCustomTemplateFiles()
			files["design.md"] = test.contents

			_, err := NewCustomTemplate(name, files)
			if err == nil {
				t.Fatalf("NewCustomTemplate() error = nil, want empty file error")
			}
			want := "custom template file api-feature/design.md is empty"
			if err.Error() != want {
				t.Fatalf("NewCustomTemplate() error = %q, want %q", err.Error(), want)
			}
		})
	}
}

func TestNewCustomTemplateIgnoresUnknownExtraFiles(t *testing.T) {
	name := mustNewCustomTemplateName(t, "api-feature")
	files := validCustomTemplateFiles()
	files["README.md"] = "notes"
	files["extra.sh"] = "#!/bin/sh"

	template, err := NewCustomTemplate(name, files)
	if err != nil {
		t.Fatalf("NewCustomTemplate() error = %v", err)
	}

	copied := template.Files()
	if len(copied) != len(RequiredOpenSpecChangeFiles()) {
		t.Fatalf("Files() = %v, want only the five required files", copied)
	}
	if _, exists := copied["README.md"]; exists {
		t.Fatalf("Files() contains unknown extra file README.md")
	}
}

func TestCustomTemplateCopiesFilesDefensively(t *testing.T) {
	name := mustNewCustomTemplateName(t, "api-feature")
	files := validCustomTemplateFiles()

	template, err := NewCustomTemplate(name, files)
	if err != nil {
		t.Fatalf("NewCustomTemplate() error = %v", err)
	}

	files["proposal.md"] = "mutated input"
	if template.Files()["proposal.md"] != "# Content for proposal.md\n" {
		t.Fatalf("mutating the input map changed the template model")
	}

	returned := template.Files()
	returned["proposal.md"] = "mutated output"
	if template.Files()["proposal.md"] != "# Content for proposal.md\n" {
		t.Fatalf("mutating a returned map changed the template model")
	}

	rendered := template.Render("change", "", "")
	rendered["proposal.md"] = "mutated rendered"
	if template.Files()["proposal.md"] != "# Content for proposal.md\n" {
		t.Fatalf("mutating a rendered map changed the template model")
	}
}

func TestRenderCustomTemplateContentSubstitutesVariables(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		title   string
		summary string
		want    string
	}{
		{
			name:   "change id always replaced",
			source: "# Change {{change_id}}",
			want:   "# Change add-payment-flow",
		},
		{
			name:   "title replaced when provided",
			source: "# {{title}} ({{change_id}})",
			title:  "Add payments",
			want:   "# Add payments (add-payment-flow)",
		},
		{
			name:    "summary replaced when provided",
			source:  "Summary: {{summary}}",
			summary: "Adds a payment flow.",
			want:    "Summary: Adds a payment flow.",
		},
		{
			name:   "title left verbatim when omitted",
			source: "# {{title}}",
			want:   "# {{title}}",
		},
		{
			name:    "title left verbatim when whitespace only",
			source:  "# {{title}}",
			title:   "   ",
			want:    "# {{title}}",
			summary: "",
		},
		{
			name:   "summary left verbatim when omitted",
			source: "Summary: {{summary}}",
			want:   "Summary: {{summary}}",
		},
		{
			name:   "unknown tokens left verbatim",
			source: "{{type}} {{template_name}} {{unknown}}",
			want:   "{{type}} {{template_name}} {{unknown}}",
		},
		{
			name:    "provided values are trimmed",
			source:  "# {{title}}: {{summary}}",
			title:   "  Add payments  ",
			summary: "  Adds a payment flow.  ",
			want:    "# Add payments: Adds a payment flow.",
		},
		{
			name:   "repeated tokens all replaced",
			source: "{{change_id}} and {{change_id}}",
			want:   "add-payment-flow and add-payment-flow",
		},
		{
			name:   "non-token content unchanged",
			source: "plain content without variables\n",
			want:   "plain content without variables\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RenderCustomTemplateContent(test.source, "add-payment-flow", test.title, test.summary)
			if got != test.want {
				t.Fatalf("RenderCustomTemplateContent() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRenderCustomTemplateContentIsDeterministic(t *testing.T) {
	source := "# {{title}}\n{{summary}}\n{{change_id}}\n"
	first := RenderCustomTemplateContent(source, "change", "Title", "Summary")
	second := RenderCustomTemplateContent(source, "change", "Title", "Summary")
	if first != second {
		t.Fatalf("rendering is not deterministic: %q != %q", first, second)
	}
}

func TestCustomTemplateRenderRendersEveryFile(t *testing.T) {
	name := mustNewCustomTemplateName(t, "api-feature")
	files := validCustomTemplateFiles()
	for file := range files {
		files[file] = "# {{change_id}} " + file + "\n"
	}

	template, err := NewCustomTemplate(name, files)
	if err != nil {
		t.Fatalf("NewCustomTemplate() error = %v", err)
	}

	rendered := template.Render("add-payment-flow", "", "")
	for _, requiredFile := range RequiredOpenSpecChangeFiles() {
		want := "# add-payment-flow " + requiredFile + "\n"
		if rendered[requiredFile] != want {
			t.Fatalf("Render()[%q] = %q, want %q", requiredFile, rendered[requiredFile], want)
		}
		if strings.Contains(rendered[requiredFile], "{{change_id}}") {
			t.Fatalf("Render()[%q] still contains the change id token", requiredFile)
		}
	}
}
