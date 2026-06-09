package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var ErrAgentAssistedExecuteUnsupported = errors.New("agent-assisted execute mode is unsupported in this version")

type AgentAssistedAuthoringInput struct {
	ProjectRoot   string
	ChangeID      string
	AgentName     string
	AuthoringType string
	Title         string
	Summary       string
	Execute       bool
}

type AgentAssistedAuthoring struct {
	promptRenderer ports.AgentAssistedAuthoringPromptRenderer
}

func NewAgentAssistedAuthoring(promptRenderer ports.AgentAssistedAuthoringPromptRenderer) *AgentAssistedAuthoring {
	return &AgentAssistedAuthoring{promptRenderer: promptRenderer}
}

func (useCase *AgentAssistedAuthoring) Execute(
	input AgentAssistedAuthoringInput,
) (domain.AgentAssistedAuthoringResult, error) {
	if useCase == nil {
		return domain.AgentAssistedAuthoringResult{}, errors.New("agent-assisted authoring use case is required")
	}
	if input.Execute {
		return domain.AgentAssistedAuthoringResult{}, ErrAgentAssistedExecuteUnsupported
	}
	if useCase.promptRenderer == nil {
		return domain.AgentAssistedAuthoringResult{}, errors.New("agent-assisted authoring prompt renderer is required")
	}

	request, err := normalizeAgentAssistedAuthoringInput(input)
	if err != nil {
		return domain.AgentAssistedAuthoringResult{}, err
	}

	prompt, err := useCase.promptRenderer.Render(request)
	if err != nil {
		return domain.AgentAssistedAuthoringResult{}, fmt.Errorf("render agent-assisted authoring prompt: %w", err)
	}

	return domain.NewAgentAssistedAuthoringResult(
		request.ChangeID,
		request.AgentName,
		request.AuthoringType,
		request.Title,
		request.Summary,
		request.ChangePath,
		request.RequiredFiles,
		agentAssistedAuthoringPlan(request.ChangePath),
		prompt,
	), nil
}

func normalizeAgentAssistedAuthoringInput(
	input AgentAssistedAuthoringInput,
) (domain.AgentAssistedAuthoringPromptRequest, error) {
	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.AgentAssistedAuthoringPromptRequest{}, errors.New("project root is required")
	}

	changeID := strings.TrimSpace(input.ChangeID)
	if changeID == "" {
		return domain.AgentAssistedAuthoringPromptRequest{}, errors.New("change id is required")
	}
	if err := validateGenerationChangeID(changeID); err != nil {
		return domain.AgentAssistedAuthoringPromptRequest{}, err
	}

	agentName, err := domain.ParseAgentName(input.AgentName)
	if err != nil {
		return domain.AgentAssistedAuthoringPromptRequest{}, err
	}

	authoringType, err := domain.ParseAgentAssistedAuthoringType(input.AuthoringType)
	if err != nil {
		return domain.AgentAssistedAuthoringPromptRequest{}, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return domain.AgentAssistedAuthoringPromptRequest{}, errors.New("agent-assisted title is required")
	}

	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		return domain.AgentAssistedAuthoringPromptRequest{}, errors.New("agent-assisted summary is required")
	}

	changePath := openspecChangesDirectory + "/" + changeID
	return domain.NewAgentAssistedAuthoringPromptRequest(
		changeID,
		agentName,
		authoringType,
		title,
		summary,
		changePath,
		domain.RequiredOpenSpecChangeFiles(),
	), nil
}

func agentAssistedAuthoringPlan(changePath string) []string {
	return []string{
		"Validate the agent-assisted authoring request.",
		"Build the OpenSpec authoring plan for " + changePath + ".",
		"Render a deterministic, copy-pasteable Markdown authoring prompt.",
		"Stop before writing files, executing agents, or running external commands.",
	}
}
