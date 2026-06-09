package domain

import (
	"fmt"
	"strings"
)

type AgentName string

func ParseAgentName(value string) (AgentName, error) {
	agentName := AgentName(strings.TrimSpace(value))
	if agentName == "" {
		return "", fmt.Errorf("agent name is required")
	}
	if strings.ContainsAny(string(agentName), "\r\n") {
		return "", fmt.Errorf("agent name must be safe display text")
	}
	return agentName, nil
}

type AgentAssistedAuthoringType string

const (
	FeatureAgentAssistedAuthoringType  AgentAssistedAuthoringType = "feature"
	BugfixAgentAssistedAuthoringType   AgentAssistedAuthoringType = "bugfix"
	DocsAgentAssistedAuthoringType     AgentAssistedAuthoringType = "docs"
	RefactorAgentAssistedAuthoringType AgentAssistedAuthoringType = "refactor"
)

func ParseAgentAssistedAuthoringType(value string) (AgentAssistedAuthoringType, error) {
	authoringType := AgentAssistedAuthoringType(strings.TrimSpace(value))
	if authoringType == "" {
		return "", fmt.Errorf("agent-assisted authoring type is required")
	}
	if !authoringType.IsSupported() {
		return "", fmt.Errorf("unknown agent-assisted authoring type: %s", authoringType)
	}
	return authoringType, nil
}

func SupportedAgentAssistedAuthoringTypes() []AgentAssistedAuthoringType {
	return []AgentAssistedAuthoringType{
		FeatureAgentAssistedAuthoringType,
		BugfixAgentAssistedAuthoringType,
		DocsAgentAssistedAuthoringType,
		RefactorAgentAssistedAuthoringType,
	}
}

func (authoringType AgentAssistedAuthoringType) IsSupported() bool {
	switch authoringType {
	case FeatureAgentAssistedAuthoringType,
		BugfixAgentAssistedAuthoringType,
		DocsAgentAssistedAuthoringType,
		RefactorAgentAssistedAuthoringType:
		return true
	default:
		return false
	}
}

type AgentAssistedAuthoringPromptRequest struct {
	ChangeID      string
	AgentName     AgentName
	AuthoringType AgentAssistedAuthoringType
	Title         string
	Summary       string
	ChangePath    string
	RequiredFiles []string
}

func NewAgentAssistedAuthoringPromptRequest(
	changeID string,
	agentName AgentName,
	authoringType AgentAssistedAuthoringType,
	title string,
	summary string,
	changePath string,
	requiredFiles []string,
) AgentAssistedAuthoringPromptRequest {
	return AgentAssistedAuthoringPromptRequest{
		ChangeID:      changeID,
		AgentName:     agentName,
		AuthoringType: authoringType,
		Title:         title,
		Summary:       summary,
		ChangePath:    changePath,
		RequiredFiles: append([]string(nil), requiredFiles...),
	}
}

type AgentAssistedAuthoringResult struct {
	ChangeID                     string
	AgentName                    AgentName
	AuthoringType                AgentAssistedAuthoringType
	Title                        string
	Summary                      string
	ChangePath                   string
	DryRun                       bool
	NoFilesWritten               bool
	NoPromptFileWritten          bool
	NoAgentExecuted              bool
	NoExternalCommandExecuted    bool
	NoAgentOutputParsedOrApplied bool
	Prompt                       string
	requiredFiles                []string
	plan                         []string
}

func NewAgentAssistedAuthoringResult(
	changeID string,
	agentName AgentName,
	authoringType AgentAssistedAuthoringType,
	title string,
	summary string,
	changePath string,
	requiredFiles []string,
	plan []string,
	prompt string,
) AgentAssistedAuthoringResult {
	return AgentAssistedAuthoringResult{
		ChangeID:                     changeID,
		AgentName:                    agentName,
		AuthoringType:                authoringType,
		Title:                        title,
		Summary:                      summary,
		ChangePath:                   changePath,
		DryRun:                       true,
		NoFilesWritten:               true,
		NoPromptFileWritten:          true,
		NoAgentExecuted:              true,
		NoExternalCommandExecuted:    true,
		NoAgentOutputParsedOrApplied: true,
		Prompt:                       prompt,
		requiredFiles:                append([]string(nil), requiredFiles...),
		plan:                         append([]string(nil), plan...),
	}
}

func (result AgentAssistedAuthoringResult) RequiredFiles() []string {
	return append([]string(nil), result.requiredFiles...)
}

func (result AgentAssistedAuthoringResult) Plan() []string {
	return append([]string(nil), result.plan...)
}
