package templates

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type AgentAssistedAuthoringPromptTemplate struct{}

func NewAgentAssistedAuthoringPromptTemplate() *AgentAssistedAuthoringPromptTemplate {
	return &AgentAssistedAuthoringPromptTemplate{}
}

func (template *AgentAssistedAuthoringPromptTemplate) Render(
	request domain.AgentAssistedAuthoringPromptRequest,
) (string, error) {
	if template == nil {
		return "", errors.New("agent-assisted authoring prompt template is required")
	}
	if strings.TrimSpace(request.ChangeID) == "" {
		return "", errors.New("change id is required")
	}
	if strings.TrimSpace(string(request.AgentName)) == "" {
		return "", errors.New("agent name is required")
	}
	if strings.TrimSpace(string(request.AuthoringType)) == "" {
		return "", errors.New("agent-assisted authoring type is required")
	}
	if strings.TrimSpace(request.Title) == "" {
		return "", errors.New("agent-assisted title is required")
	}
	if strings.TrimSpace(request.Summary) == "" {
		return "", errors.New("agent-assisted summary is required")
	}
	if strings.TrimSpace(request.ChangePath) == "" {
		return "", errors.New("target OpenSpec path is required")
	}
	if len(request.RequiredFiles) == 0 {
		return "", errors.New("required OpenSpec files are required")
	}

	var prompt strings.Builder
	fmt.Fprintln(&prompt, "# Agent-Assisted OpenSpec Authoring Prompt")
	fmt.Fprintln(&prompt)
	fmt.Fprintf(
		&prompt,
		"You are `%s`, helping author or refine an OpenSpec change for SpecHarbor. This prompt is self-contained and copy-pasteable.\n",
		request.AgentName,
	)
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "## Project Context")
	fmt.Fprintln(&prompt)
	fmt.Fprintln(
		&prompt,
		"SpecHarbor is an open source CLI for generating, validating, reviewing, and archiving OpenSpec-based development workflows for coding agents.",
	)
	fmt.Fprintln(
		&prompt,
		"SpecHarbor helps developers convert a loose idea into a structured OpenSpec change, technical design, implementation tasks, acceptance criteria, risk notes, and agent-specific prompts.",
	)
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "## Change Request")
	fmt.Fprintln(&prompt)
	fmt.Fprintf(&prompt, "- Change id: `%s`\n", request.ChangeID)
	fmt.Fprintf(&prompt, "- Authoring type: `%s`\n", request.AuthoringType)
	fmt.Fprintf(&prompt, "- Title: %s\n", request.Title)
	fmt.Fprintf(&prompt, "- Summary: %s\n", request.Summary)
	fmt.Fprintf(&prompt, "- Target OpenSpec path: `%s`\n", request.ChangePath)
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "## Required OpenSpec Files")
	fmt.Fprintln(&prompt)
	fmt.Fprintf(&prompt, "Create or refine Markdown content only under `%s/` for these files:\n", request.ChangePath)
	for _, requiredFile := range request.RequiredFiles {
		fmt.Fprintf(&prompt, "- `%s`\n", requiredFile)
	}
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "## Scope Rules")
	fmt.Fprintln(&prompt)
	fmt.Fprintf(&prompt, "- Create or refine only files under `%s/`.\n", request.ChangePath)
	fmt.Fprintln(&prompt, "- Do not implement production code.")
	fmt.Fprintln(&prompt, "- Do not modify unrelated files.")
	fmt.Fprintln(&prompt, readmeDocsBoundary(request.AuthoringType))
	fmt.Fprintln(&prompt, "- Leave implementation tasks unchecked in `tasks.md`; use unchecked Markdown boxes such as `- [ ]`.")
	fmt.Fprintln(&prompt, "- Preserve architecture boundaries.")
	fmt.Fprintln(&prompt, "- Do not run implementation, tests, source-control commands, workflow commands, provider setup, credential setup, commits, pushes, merges, deployments, or production code edits.")
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "## Architecture Boundaries")
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "- Domain code belongs in `internal/core/domain`.")
	fmt.Fprintln(&prompt, "- Ports belong in `internal/core/ports`.")
	fmt.Fprintln(&prompt, "- Use cases belong in `internal/core/usecase`.")
	fmt.Fprintln(&prompt, "- Concrete implementations belong in `internal/adapters`.")
	fmt.Fprintln(&prompt, "- Core must not import adapters.")
	fmt.Fprintln(&prompt, "- CLI must not contain business rules.")
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "## Output Expectations")
	fmt.Fprintln(&prompt)
	fmt.Fprintln(&prompt, "- Output Markdown-only OpenSpec content.")
	fmt.Fprintf(&prompt, "- Keep all output scoped to `%s/`.\n", request.ChangePath)
	fmt.Fprintln(&prompt, "- Do not depend on a prompt file or on files written by SpecHarbor.")
	fmt.Fprintln(&prompt, "- Do not include source-control details, workflow details, provider details, credentials, network steps, or production code patches.")
	fmt.Fprintf(&prompt, "- Run or recommend `specharbor validate %s` when the OpenSpec files exist.\n", request.ChangeID)

	return prompt.String(), nil
}

func readmeDocsBoundary(authoringType domain.AgentAssistedAuthoringType) string {
	if authoringType == domain.DocsAgentAssistedAuthoringType {
		return "- Documentation scope may be described only when it comes from the title and summary; do not edit README or docs files in this authoring step."
	}
	return "- Do not change README or documentation files; this authoring type is not documentation scope."
}
