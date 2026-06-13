package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

const DefaultPromptTask = "Follow the active OpenSpec change."

type RenderPromptInput struct {
	ProjectRoot string
	ChangeID    string
	Role        string
}

type RenderPromptResult struct {
	Prompt string
}

type RenderPrompt struct {
	templates       ports.PromptTemplateRepository
	renderer        ports.TemplateRenderer
	contextProvider ports.PromptContextProvider
	contextPolicy   domain.PromptContextRenderPolicy
}

func NewRenderPrompt(templates ports.PromptTemplateRepository, renderer ports.TemplateRenderer) *RenderPrompt {
	return NewRenderPromptWithContext(templates, renderer, nil)
}

func NewRenderPromptWithContext(
	templates ports.PromptTemplateRepository,
	renderer ports.TemplateRenderer,
	contextProvider ports.PromptContextProvider,
) *RenderPrompt {
	return &RenderPrompt{
		templates:       templates,
		renderer:        renderer,
		contextProvider: contextProvider,
		contextPolicy:   domain.DefaultPromptContextRenderPolicy(),
	}
}

func (useCase *RenderPrompt) Execute(input RenderPromptInput) (RenderPromptResult, error) {
	if useCase == nil {
		return RenderPromptResult{}, errors.New("render prompt use case is required")
	}
	if useCase.templates == nil {
		return RenderPromptResult{}, errors.New("prompt template repository is required")
	}
	if useCase.renderer == nil {
		return RenderPromptResult{}, errors.New("template renderer is required")
	}

	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return RenderPromptResult{}, errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return RenderPromptResult{}, errors.New("change id is required")
	}

	roleValue := strings.TrimSpace(input.Role)
	if roleValue == "" {
		return RenderPromptResult{}, errors.New("prompt role is required")
	}

	role, supported := domain.ParsePromptRole(roleValue)
	if !supported {
		return RenderPromptResult{}, fmt.Errorf("unsupported prompt role: %s", roleValue)
	}

	templateSource, err := useCase.templates.TemplateForRole(projectRoot, role)
	if err != nil {
		return RenderPromptResult{}, fmt.Errorf("load prompt template for role %s: %w", role, err)
	}

	projectContext, projectBriefReadFirst, err := useCase.promptContextTemplateData(projectRoot)
	if err != nil {
		return RenderPromptResult{}, err
	}

	prompt, err := useCase.renderer.Render(templateSource, map[string]string{
		"change_id":                changeID,
		"task":                     DefaultPromptTask,
		"project_context":          projectContext,
		"project_brief_read_first": projectBriefReadFirst,
	})
	if err != nil {
		return RenderPromptResult{}, fmt.Errorf("render prompt template for role %s: %w", role, err)
	}

	return RenderPromptResult{Prompt: prompt}, nil
}

func (useCase *RenderPrompt) promptContextTemplateData(projectRoot string) (string, string, error) {
	discoveryResult := domain.NewContextDiscoveryResult(nil, nil)
	projectBriefExists := false

	if useCase.contextProvider != nil {
		var err error
		discoveryResult, err = useCase.contextProvider.DiscoverPromptContext(projectRoot)
		if err != nil {
			return "", "", fmt.Errorf("discover prompt context: %w", err)
		}
		projectBriefExists, err = useCase.contextProvider.ProjectBriefExists(projectRoot)
		if err != nil {
			return "", "", fmt.Errorf("check project brief for prompt context: %w", err)
		}
	}

	promptContext := domain.NewPromptProjectContext(discoveryResult, useCase.contextPolicy)
	projectBriefReadFirst := ""
	if projectBriefExists || promptContext.HasUserConfirmedContext() {
		projectBriefReadFirst = "- `.specharbor/project-brief.md`\n"
	}
	return promptContext.RenderMarkdown(), projectBriefReadFirst, nil
}
