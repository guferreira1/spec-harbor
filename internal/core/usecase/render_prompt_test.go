package usecase

import (
	"fmt"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestRenderPromptRendersSupportedRoles(t *testing.T) {
	for _, role := range domain.SupportedPromptRoles() {
		t.Run(string(role), func(t *testing.T) {
			templates := &fakePromptTemplates{
				template: "# Prompt\nchange={{change_id}}\ntask={{task}}\n",
			}
			renderer := fakePromptRenderer{}
			useCase := NewRenderPrompt(templates, renderer)

			result, err := useCase.Execute(RenderPromptInput{
				ProjectRoot: "/project",
				ChangeID:    "implement-prompt-command",
				Role:        string(role),
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if templates.projectRoot != "/project" {
				t.Fatalf("project root = %q, want /project", templates.projectRoot)
			}
			if templates.role != role {
				t.Fatalf("role = %q, want %q", templates.role, role)
			}
			if !strings.Contains(result.Prompt, "change=implement-prompt-command") {
				t.Fatalf("prompt = %q, want rendered change id", result.Prompt)
			}
			if !strings.Contains(result.Prompt, "task="+DefaultPromptTask) {
				t.Fatalf("prompt = %q, want rendered default task", result.Prompt)
			}
			if strings.Contains(result.Prompt, "{{change_id}}") {
				t.Fatalf("prompt = %q, want no raw change_id placeholder", result.Prompt)
			}
			if strings.Contains(result.Prompt, "{{task}}") {
				t.Fatalf("prompt = %q, want no raw task placeholder", result.Prompt)
			}
		})
	}
}

func TestRenderPromptRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input RenderPromptInput
		want  string
	}{
		{
			name:  "empty project root",
			input: RenderPromptInput{ProjectRoot: " ", ChangeID: "change", Role: "implementer"},
			want:  "project root is required",
		},
		{
			name:  "empty change id",
			input: RenderPromptInput{ProjectRoot: "/project", ChangeID: " ", Role: "implementer"},
			want:  "change id is required",
		},
		{
			name:  "empty role",
			input: RenderPromptInput{ProjectRoot: "/project", ChangeID: "change", Role: " "},
			want:  "prompt role is required",
		},
		{
			name:  "unsupported role",
			input: RenderPromptInput{ProjectRoot: "/project", ChangeID: "change", Role: "unknown"},
			want:  "unsupported prompt role: unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := NewRenderPrompt(&fakePromptTemplates{}, fakePromptRenderer{})

			_, err := useCase.Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

type fakePromptTemplates struct {
	template    string
	projectRoot string
	role        domain.PromptRole
}

func (templates *fakePromptTemplates) TemplateForRole(projectRoot string, role domain.PromptRole) (string, error) {
	templates.projectRoot = projectRoot
	templates.role = role
	if templates.template == "" {
		return "# Prompt\n{{change_id}}\n{{task}}\n", nil
	}
	return templates.template, nil
}

type fakePromptRenderer struct{}

func (renderer fakePromptRenderer) Render(templateSource string, data map[string]string) (string, error) {
	output := templateSource
	for key, value := range data {
		output = strings.ReplaceAll(output, fmt.Sprintf("{{%s}}", key), value)
	}
	return output, nil
}
