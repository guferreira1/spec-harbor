package templates

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestRolePromptTemplatesLoadAndRenderRequiredRoles(t *testing.T) {
	templates := NewRolePromptTemplates()
	renderer := NewPromptTemplateRenderer()

	for _, role := range domain.SupportedPromptRoles() {
		t.Run(string(role), func(t *testing.T) {
			templateSource, err := templates.TemplateForRole(t.TempDir(), role)
			if err != nil {
				t.Fatalf("TemplateForRole() error = %v", err)
			}
			if strings.TrimSpace(templateSource) == "" {
				t.Fatalf("TemplateForRole() returned empty template")
			}

			prompt, err := renderer.Render(templateSource, map[string]string{
				"change_id":                "implement-prompt-command",
				"task":                     "Follow the active OpenSpec change.",
				"project_context":          "## Project Context\n\nContext block.",
				"project_brief_read_first": "- `.specharbor/project-brief.md`\n",
			})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if !strings.Contains(prompt, "openspec/changes/implement-prompt-command/") {
				t.Fatalf("prompt = %q, want rendered change id", prompt)
			}
			for _, want := range []string{
				"`AGENTS.md`",
				"`.specharbor/rules/global.md`",
				"`README.md`",
				"`docs/`",
				"`.specharbor/project-brief.md`",
				"`openspec/project.md`",
				"`openspec/specs/architecture/spec.md`",
				"## Project Context",
			} {
				if !strings.Contains(prompt, want) {
					t.Fatalf("prompt = %q, want to contain %q", prompt, want)
				}
			}
			if strings.Contains(prompt, "{{change_id}}") {
				t.Fatalf("prompt = %q, want no raw change_id placeholder", prompt)
			}
			if strings.Contains(prompt, "{{task}}") {
				t.Fatalf("prompt = %q, want no raw task placeholder", prompt)
			}
			if strings.Contains(prompt, "{{project_context}}") {
				t.Fatalf("prompt = %q, want no raw project_context placeholder", prompt)
			}
			if strings.Contains(prompt, "{{project_brief_read_first}}") {
				t.Fatalf("prompt = %q, want no raw project_brief_read_first placeholder", prompt)
			}
		})
	}
}

func TestRolePromptTemplatesRenderFinalDecisionLabelsForSupportedRoles(t *testing.T) {
	templates := NewRolePromptTemplates()
	renderer := NewPromptTemplateRenderer()

	for role, labels := range expectedRoleFinalDecisionLabels() {
		t.Run(string(role), func(t *testing.T) {
			templateSource, err := templates.TemplateForRole(t.TempDir(), role)
			if err != nil {
				t.Fatalf("TemplateForRole() error = %v", err)
			}

			prompt, err := renderer.Render(templateSource, map[string]string{
				"change_id":                "implement-context-aware-agent-prompts",
				"task":                     "Follow the active OpenSpec change.",
				"project_context":          "## Project Context\n\nContext block.",
				"project_brief_read_first": "",
			})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			if strings.Count(prompt, "## Final Decision") != 1 {
				t.Fatalf("prompt = %q, want exactly one Final Decision section", prompt)
			}
			if strings.Count(prompt, "Final decision must be exactly one of:") != 1 {
				t.Fatalf("prompt = %q, want exactly one final decision instruction", prompt)
			}

			expected := make(map[string]bool, len(labels))
			for _, label := range labels {
				expected[label] = true
				bullet := "- " + label
				if strings.Count(prompt, bullet) != 1 {
					t.Fatalf("prompt = %q, want exactly one %q final decision label", prompt, label)
				}
			}

			for _, label := range allRoleFinalDecisionLabels() {
				if expected[label] {
					continue
				}
				if strings.Contains(prompt, "- "+label) {
					t.Fatalf("prompt = %q, role %q must not include label %q", prompt, role, label)
				}
			}
		})
	}
}

func TestRolePromptTemplatesDoNotDependOnProjectRoot(t *testing.T) {
	templates := NewRolePromptTemplates()

	templateSource, err := templates.TemplateForRole(t.TempDir(), domain.PromptRoleImplementer)
	if err != nil {
		t.Fatalf("TemplateForRole() error = %v", err)
	}
	if !strings.Contains(templateSource, "# Implementer Agent") {
		t.Fatalf("TemplateForRole() = %q, want embedded implementer template", templateSource)
	}
}

func expectedRoleFinalDecisionLabels() map[domain.PromptRole][]string {
	return map[domain.PromptRole][]string{
		domain.PromptRoleSpecAuthor: {
			"SPEC_READY_FOR_ARCHITECTURE_REVIEW",
			"BLOCKED",
		},
		domain.PromptRoleArchitectureReviewer: {
			"APPROVED_FOR_IMPLEMENTATION",
			"NEEDS_SPEC_CHANGES",
			"BLOCKED",
		},
		domain.PromptRoleImplementer: {
			"IMPLEMENTATION_COMPLETE",
			"BLOCKED",
		},
		domain.PromptRoleTestEngineer: {
			"TESTS_PASSED",
			"TESTS_FAILED",
			"BLOCKED",
		},
		domain.PromptRoleChangeReviewer: {
			"APPROVED_FOR_PR",
			"NEEDS_CHANGES",
			"BLOCKED",
		},
	}
}

func allRoleFinalDecisionLabels() []string {
	return []string{
		"SPEC_READY_FOR_ARCHITECTURE_REVIEW",
		"APPROVED_FOR_IMPLEMENTATION",
		"NEEDS_SPEC_CHANGES",
		"IMPLEMENTATION_COMPLETE",
		"TESTS_PASSED",
		"TESTS_FAILED",
		"APPROVED_FOR_PR",
		"NEEDS_CHANGES",
		"BLOCKED",
	}
}

func TestRolePromptTemplatesReturnMissingTemplateError(t *testing.T) {
	templates := NewRolePromptTemplates()

	_, err := templates.TemplateForRole(t.TempDir(), domain.PromptRole("unknown"))
	if err == nil {
		t.Fatalf("TemplateForRole() error = nil, want missing template error")
	}
	if !strings.Contains(err.Error(), "read embedded role prompt template unknown") {
		t.Fatalf("TemplateForRole() error = %q, want role template context", err.Error())
	}
}

func TestPromptTemplateRendererReturnsRenderErrors(t *testing.T) {
	renderer := NewPromptTemplateRenderer()

	if _, err := renderer.Render("{{unknown}}", map[string]string{
		"change_id":                "change",
		"task":                     "Follow the active OpenSpec change.",
		"project_context":          "## Project Context",
		"project_brief_read_first": "",
	}); err == nil {
		t.Fatalf("Render() error = nil, want unknown placeholder error")
	}

	if _, err := renderer.Render("{{change_id", map[string]string{
		"change_id":                "change",
		"task":                     "Follow the active OpenSpec change.",
		"project_context":          "## Project Context",
		"project_brief_read_first": "",
	}); err == nil {
		t.Fatalf("Render() error = nil, want parse error")
	}
}
