package templates

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"strings"
	"text/template"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

//go:embed role_prompts/*.md.tmpl
var rolePromptFS embed.FS

type RolePromptTemplates struct{}

func NewRolePromptTemplates() *RolePromptTemplates {
	return &RolePromptTemplates{}
}

func (templates *RolePromptTemplates) TemplateForRole(projectRoot string, role domain.PromptRole) (string, error) {
	_ = projectRoot

	templatePath := path.Join("role_prompts", string(role)+".md.tmpl")
	contents, err := rolePromptFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read embedded role prompt template %s: %w", role, err)
	}
	return string(contents), nil
}

type PromptTemplateRenderer struct{}

func NewPromptTemplateRenderer() *PromptTemplateRenderer {
	return &PromptTemplateRenderer{}
}

func (renderer *PromptTemplateRenderer) Render(templateSource string, request domain.PromptRenderRequest) (string, error) {
	data, err := promptRenderData(request)
	if err != nil {
		return "", err
	}

	parsed, err := template.New("prompt").Funcs(templateFunctions(data)).Option("missingkey=error").Parse(templateSource)
	if err != nil {
		return "", err
	}

	var output bytes.Buffer
	if err := parsed.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

func promptRenderData(request domain.PromptRenderRequest) (map[string]string, error) {
	targetAgent := request.TargetAgent
	if targetAgent == "" {
		targetAgent = domain.DefaultPromptTargetAgent()
	}
	if _, err := domain.ParsePromptTargetAgent(string(targetAgent)); err != nil {
		return nil, err
	}

	targetSection, err := promptTargetSection(request.Role, targetAgent)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"change_id":                request.ChangeID,
		"task":                     request.Task,
		"project_context":          request.ProjectContext,
		"project_brief_read_first": request.ProjectBriefReadFirst,
		"target_agent_section":     targetSection,
	}, nil
}

func promptTargetSection(role domain.PromptRole, targetAgent domain.PromptTargetAgent) (string, error) {
	guidance, err := promptTargetGuidance(targetAgent)
	if err != nil {
		return "", err
	}

	displayName := domain.PromptTargetAgentDisplayName(targetAgent)
	var section strings.Builder
	fmt.Fprintln(&section, "## Prompt Target")
	fmt.Fprintln(&section)
	fmt.Fprintf(&section, "- Workflow role: `%s`\n", role)
	fmt.Fprintf(&section, "- Target agent: `%s` (%s)\n", targetAgent, displayName)
	fmt.Fprintln(&section, "- Paste or use this prompt in the selected external tool.")
	fmt.Fprintln(&section, "- SpecHarbor only generated this prompt text; it does not execute, configure, authenticate, start, supervise, or automate the selected external tool.")
	fmt.Fprintln(&section)
	fmt.Fprintln(&section, "Target-agent guidance:")
	fmt.Fprintf(&section, "- %s\n", guidance)
	fmt.Fprintln(&section)
	return section.String(), nil
}

func promptTargetGuidance(targetAgent domain.PromptTargetAgent) (string, error) {
	switch targetAgent {
	case domain.PromptTargetAgentGeneric:
		return "Use neutral instructions suitable for any coding assistant.", nil
	case domain.PromptTargetAgentCodex:
		return "Intended for Codex-style repository work; keep edits scoped and follow explicit verification steps.", nil
	case domain.PromptTargetAgentClaudeCode:
		return "Apply careful repository edits, reason from project context, and run requested validation commands.", nil
	case domain.PromptTargetAgentDevin:
		return "Respect autonomous-task boundaries, prepare PR-ready work, and report completed work and verification.", nil
	case domain.PromptTargetAgentCursor:
		return "Use editor-assisted implementation and keep changes file-scoped to the requested area.", nil
	case domain.PromptTargetAgentCopilot:
		return "Use inside the coding environment and keep suggestions aligned with the active repository context.", nil
	case domain.PromptTargetAgentGemini:
		return "Use repository context carefully and verify commands and results explicitly.", nil
	case domain.PromptTargetAgentRoo:
		return "Follow the role-based task within the selected workflow responsibility.", nil
	case domain.PromptTargetAgentWindsurf:
		return "Use through an IDE agent workflow while keeping repository edits scoped to the prompt.", nil
	case domain.PromptTargetAgentAider:
		return "Prefer patch-oriented changes with explicit file scope and clear verification output.", nil
	default:
		return "", fmt.Errorf("unsupported prompt target agent: %s", targetAgent)
	}
}

func templateFunctions(data map[string]string) template.FuncMap {
	functions := make(template.FuncMap, len(data))
	for key, value := range data {
		value := value
		functions[key] = func() string {
			return value
		}
	}
	return functions
}
