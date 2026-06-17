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
				template: "# Prompt\nchange={{change_id}}\ntask={{task}}\nrole={{role}}\nagent={{target_agent}}\n",
			}
			renderer := &fakePromptRenderer{}
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
			if len(renderer.requests) != 1 {
				t.Fatalf("renderer requests = %d, want 1", len(renderer.requests))
			}
			if renderer.requests[0].Role != role {
				t.Fatalf("render request role = %q, want %q", renderer.requests[0].Role, role)
			}
			if renderer.requests[0].TargetAgent != domain.PromptTargetAgentGeneric {
				t.Fatalf("render request target agent = %q, want generic", renderer.requests[0].TargetAgent)
			}
			if !strings.Contains(result.Prompt, "change=implement-prompt-command") {
				t.Fatalf("prompt = %q, want rendered change id", result.Prompt)
			}
			if !strings.Contains(result.Prompt, "task="+DefaultPromptTask) {
				t.Fatalf("prompt = %q, want rendered default task", result.Prompt)
			}
			if !strings.Contains(result.Prompt, "role="+string(role)) {
				t.Fatalf("prompt = %q, want rendered role", result.Prompt)
			}
			if !strings.Contains(result.Prompt, "agent=generic") {
				t.Fatalf("prompt = %q, want rendered default agent", result.Prompt)
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

func TestRenderPromptPassesExplicitSupportedAgentsToRenderer(t *testing.T) {
	for _, agent := range domain.SupportedPromptTargetAgents() {
		t.Run(string(agent.ID), func(t *testing.T) {
			templates := &fakePromptTemplates{
				template: "# Prompt\nrole={{role}}\nagent={{target_agent}}\n",
			}
			renderer := &fakePromptRenderer{}
			useCase := NewRenderPrompt(templates, renderer)

			result, err := useCase.Execute(RenderPromptInput{
				ProjectRoot: "/project",
				ChangeID:    "implement-agent-aware-prompts",
				Role:        "implementer",
				Agent:       string(agent.ID),
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if len(renderer.requests) != 1 {
				t.Fatalf("renderer requests = %d, want 1", len(renderer.requests))
			}
			request := renderer.requests[0]
			if request.Role != domain.PromptRoleImplementer {
				t.Fatalf("render request role = %q, want implementer", request.Role)
			}
			if request.TargetAgent != agent.ID {
				t.Fatalf("render request target agent = %q, want %q", request.TargetAgent, agent.ID)
			}
			if !strings.Contains(result.Prompt, "role=implementer") {
				t.Fatalf("prompt = %q, want rendered role", result.Prompt)
			}
			if !strings.Contains(result.Prompt, "agent="+string(agent.ID)) {
				t.Fatalf("prompt = %q, want rendered target agent %q", result.Prompt, agent.ID)
			}
		})
	}
}

func TestRenderPromptIncludesProjectContextForSupportedRoles(t *testing.T) {
	for _, role := range domain.SupportedPromptRoles() {
		t.Run(string(role), func(t *testing.T) {
			templates := &fakePromptTemplates{
				template: "# Prompt\n{{project_brief_read_first}}{{project_context}}\nTask: {{task}}\n",
			}
			contextProvider := &fakePromptContextProvider{
				result: domain.NewContextDiscoveryResult([]domain.ContextSignal{
					renderPromptContextSignal(t, domain.ContextSignalKindLanguage, "Go", domain.ContextSignalClassificationDetectedFact, domain.ContextConfidenceHigh, domain.ContextSource{
						Path:     "go.mod",
						Category: domain.ContextSourceCategoryPackageManifest,
					}),
				}, nil),
			}
			useCase := NewRenderPromptWithContext(templates, &fakePromptRenderer{}, contextProvider)

			result, err := useCase.Execute(RenderPromptInput{
				ProjectRoot: "/project",
				ChangeID:    "context-aware-prompts",
				Role:        string(role),
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			for _, want := range []string{
				"## Project Context",
				"### Detected facts",
				"- Language: Go",
				"Source: go.mod",
				"Confidence: high",
				"Task: " + DefaultPromptTask,
			} {
				if !strings.Contains(result.Prompt, want) {
					t.Fatalf("prompt = %q, want to contain %q", result.Prompt, want)
				}
			}
			if contextProvider.projectRoot != "/project" {
				t.Fatalf("context provider projectRoot = %q, want /project", contextProvider.projectRoot)
			}
		})
	}
}

func TestRenderPromptIncludesProjectBriefReadFirstWhenBriefExists(t *testing.T) {
	templates := &fakePromptTemplates{
		template: "Read first:\n{{project_brief_read_first}}\n{{project_context}}\n",
	}
	contextProvider := &fakePromptContextProvider{
		briefExists: true,
	}
	useCase := NewRenderPromptWithContext(templates, &fakePromptRenderer{}, contextProvider)

	result, err := useCase.Execute(RenderPromptInput{
		ProjectRoot: "/project",
		ChangeID:    "context-aware-prompts",
		Role:        "implementer",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Prompt, "- `.specharbor/project-brief.md`") {
		t.Fatalf("prompt = %q, want project brief read-first entry", result.Prompt)
	}
	if !strings.Contains(result.Prompt, "No confirmed project context") {
		t.Fatalf("prompt = %q, want missing context instructions", result.Prompt)
	}
}

func TestRenderPromptIncludesConfirmedContextAndConflictNotes(t *testing.T) {
	templates := &fakePromptTemplates{
		template: "{{project_brief_read_first}}{{project_context}}\n",
	}
	contextProvider := &fakePromptContextProvider{
		result: domain.NewContextDiscoveryResult([]domain.ContextSignal{
			renderPromptContextSignal(t, domain.ContextSignalKindStack, "Go", domain.ContextSignalClassificationUserConfirmedContext, domain.ContextConfidenceHigh, domain.ContextSource{
				Path:     ".specharbor/project-brief.md",
				Category: domain.ContextSourceCategoryProjectBrief,
				Evidence: "Stack",
			}),
			renderPromptContextSignal(t, domain.ContextSignalKindStack, "Node.js", domain.ContextSignalClassificationDetectedFact, domain.ContextConfidenceHigh, domain.ContextSource{
				Path:     "package.json",
				Category: domain.ContextSourceCategoryPackageManifest,
			}),
		}, nil),
	}
	useCase := NewRenderPromptWithContext(templates, &fakePromptRenderer{}, contextProvider)

	result, err := useCase.Execute(RenderPromptInput{
		ProjectRoot: "/project",
		ChangeID:    "context-aware-prompts",
		Role:        "architecture-reviewer",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, want := range []string{
		"- `.specharbor/project-brief.md`",
		"### User-confirmed context",
		"- Stack: Go",
		"### Conflict notes",
		"detected Stack includes Node.js from package.json",
	} {
		if !strings.Contains(result.Prompt, want) {
			t.Fatalf("prompt = %q, want to contain %q", result.Prompt, want)
		}
	}
}

func TestRenderPromptWithoutContextProviderStillRendersMissingContextInstructions(t *testing.T) {
	templates := &fakePromptTemplates{
		template: "{{project_context}}\n",
	}
	useCase := NewRenderPrompt(templates, &fakePromptRenderer{})

	result, err := useCase.Execute(RenderPromptInput{
		ProjectRoot: "/project",
		ChangeID:    "context-aware-prompts",
		Role:        "implementer",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(result.Prompt, "No confirmed project context") {
		t.Fatalf("prompt = %q, want missing context instructions", result.Prompt)
	}
}

func TestRenderPromptPreservesTemplateFinalDecisionLabels(t *testing.T) {
	templates := &fakePromptTemplates{
		template: "{{project_context}}\n\nFinal decision must be exactly one of:\nIMPLEMENTATION_COMPLETE\nBLOCKED\n",
	}
	useCase := NewRenderPrompt(templates, &fakePromptRenderer{})

	result, err := useCase.Execute(RenderPromptInput{
		ProjectRoot: "/project",
		ChangeID:    "context-aware-prompts",
		Role:        "implementer",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, want := range []string{
		"Final decision must be exactly one of:",
		"IMPLEMENTATION_COMPLETE",
		"BLOCKED",
	} {
		if !strings.Contains(result.Prompt, want) {
			t.Fatalf("prompt = %q, want to preserve %q", result.Prompt, want)
		}
	}
}

func TestRenderPromptReturnsContextProviderErrors(t *testing.T) {
	useCase := NewRenderPromptWithContext(&fakePromptTemplates{}, &fakePromptRenderer{}, &fakePromptContextProvider{
		discoverErr: fmt.Errorf("discovery failed"),
	})

	_, err := useCase.Execute(RenderPromptInput{
		ProjectRoot: "/project",
		ChangeID:    "context-aware-prompts",
		Role:        "implementer",
	})
	if err == nil || !strings.Contains(err.Error(), "discover prompt context: discovery failed") {
		t.Fatalf("Execute() error = %v, want discovery context error", err)
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
		{
			name:  "unsupported agent",
			input: RenderPromptInput{ProjectRoot: "/project", ChangeID: "change", Role: "implementer", Agent: "unknown"},
			want:  "unsupported prompt target agent: unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCase := NewRenderPrompt(&fakePromptTemplates{}, &fakePromptRenderer{})

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

type fakePromptRenderer struct {
	requests []domain.PromptRenderRequest
}

func (renderer *fakePromptRenderer) Render(templateSource string, request domain.PromptRenderRequest) (string, error) {
	renderer.requests = append(renderer.requests, request)
	output := templateSource
	for key, value := range map[string]string{
		"change_id":                request.ChangeID,
		"task":                     request.Task,
		"project_context":          request.ProjectContext,
		"project_brief_read_first": request.ProjectBriefReadFirst,
		"role":                     string(request.Role),
		"target_agent":             string(request.TargetAgent),
		"target_agent_section":     "target agent section",
	} {
		output = strings.ReplaceAll(output, fmt.Sprintf("{{%s}}", key), value)
	}
	return output, nil
}

type fakePromptContextProvider struct {
	result      domain.ContextDiscoveryResult
	briefExists bool
	discoverErr error
	briefErr    error
	projectRoot string
}

func (provider *fakePromptContextProvider) DiscoverPromptContext(projectRoot string) (domain.ContextDiscoveryResult, error) {
	provider.projectRoot = projectRoot
	if provider.discoverErr != nil {
		return domain.ContextDiscoveryResult{}, provider.discoverErr
	}
	return provider.result, nil
}

func (provider *fakePromptContextProvider) ProjectBriefExists(projectRoot string) (bool, error) {
	provider.projectRoot = projectRoot
	if provider.briefErr != nil {
		return false, provider.briefErr
	}
	return provider.briefExists, nil
}

func renderPromptContextSignal(
	t *testing.T,
	kind domain.ContextSignalKind,
	value string,
	classification domain.ContextSignalClassification,
	confidence domain.ContextConfidence,
	source domain.ContextSource,
) domain.ContextSignal {
	t.Helper()

	signal, err := domain.NewContextSignal(domain.ContextSignalInput{
		Kind:           kind,
		Value:          value,
		Classification: classification,
		Confidence:     confidence,
		Source:         source,
	})
	if err != nil {
		t.Fatalf("NewContextSignal() error = %v", err)
	}
	return signal
}
