package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestRolePromptTemplatesLoadAndRenderRequiredRoles(t *testing.T) {
	root := findProjectRoot(t)
	templates := NewRolePromptTemplates()
	renderer := NewPromptTemplateRenderer()

	for _, role := range domain.SupportedPromptRoles() {
		t.Run(string(role), func(t *testing.T) {
			templateSource, err := templates.TemplateForRole(root, role)
			if err != nil {
				t.Fatalf("TemplateForRole() error = %v", err)
			}
			if strings.TrimSpace(templateSource) == "" {
				t.Fatalf("TemplateForRole() returned empty template")
			}

			prompt, err := renderer.Render(templateSource, map[string]string{
				"change_id": "implement-prompt-command",
				"task":      "Follow the active OpenSpec change.",
			})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if !strings.Contains(prompt, "openspec/changes/implement-prompt-command/") {
				t.Fatalf("prompt = %q, want rendered change id", prompt)
			}
			if strings.Contains(prompt, "{{change_id}}") {
				t.Fatalf("prompt = %q, want no raw change_id placeholder", prompt)
			}
			if strings.Contains(prompt, "{{task}}") {
				t.Fatalf("prompt = %q, want no raw task placeholder", prompt)
			}
		})
	}
}

func TestRolePromptTemplatesReturnMissingTemplateError(t *testing.T) {
	templates := NewRolePromptTemplates()

	_, err := templates.TemplateForRole(t.TempDir(), domain.PromptRoleImplementer)
	if err == nil {
		t.Fatalf("TemplateForRole() error = nil, want missing template error")
	}
	if !strings.Contains(err.Error(), "read role prompt template implementer") {
		t.Fatalf("TemplateForRole() error = %q, want role template context", err.Error())
	}
}

func TestPromptTemplateRendererReturnsRenderErrors(t *testing.T) {
	renderer := NewPromptTemplateRenderer()

	if _, err := renderer.Render("{{unknown}}", map[string]string{
		"change_id": "change",
		"task":      "Follow the active OpenSpec change.",
	}); err == nil {
		t.Fatalf("Render() error = nil, want unknown placeholder error")
	}

	if _, err := renderer.Render("{{change_id", map[string]string{
		"change_id": "change",
		"task":      "Follow the active OpenSpec change.",
	}); err == nil {
		t.Fatalf("Render() error = nil, want parse error")
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
