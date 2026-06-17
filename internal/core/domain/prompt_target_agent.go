package domain

import (
	"fmt"
	"strings"
)

type PromptTargetAgent string

const (
	PromptTargetAgentGeneric    PromptTargetAgent = "generic"
	PromptTargetAgentCodex      PromptTargetAgent = "codex"
	PromptTargetAgentClaudeCode PromptTargetAgent = "claude-code"
	PromptTargetAgentDevin      PromptTargetAgent = "devin"
	PromptTargetAgentCursor     PromptTargetAgent = "cursor"
	PromptTargetAgentCopilot    PromptTargetAgent = "copilot"
	PromptTargetAgentGemini     PromptTargetAgent = "gemini"
	PromptTargetAgentRoo        PromptTargetAgent = "roo"
	PromptTargetAgentWindsurf   PromptTargetAgent = "windsurf"
	PromptTargetAgentAider      PromptTargetAgent = "aider"
)

type PromptTargetAgentInfo struct {
	ID          PromptTargetAgent
	DisplayName string
}

func DefaultPromptTargetAgent() PromptTargetAgent {
	return PromptTargetAgentGeneric
}

func SupportedPromptTargetAgents() []PromptTargetAgentInfo {
	return []PromptTargetAgentInfo{
		{ID: PromptTargetAgentGeneric, DisplayName: "Generic coding assistant"},
		{ID: PromptTargetAgentCodex, DisplayName: "Codex"},
		{ID: PromptTargetAgentClaudeCode, DisplayName: "Claude Code"},
		{ID: PromptTargetAgentDevin, DisplayName: "Devin"},
		{ID: PromptTargetAgentCursor, DisplayName: "Cursor"},
		{ID: PromptTargetAgentCopilot, DisplayName: "GitHub Copilot"},
		{ID: PromptTargetAgentGemini, DisplayName: "Gemini"},
		{ID: PromptTargetAgentRoo, DisplayName: "Roo"},
		{ID: PromptTargetAgentWindsurf, DisplayName: "Windsurf"},
		{ID: PromptTargetAgentAider, DisplayName: "Aider"},
	}
}

func SupportedPromptTargetAgentIDs() []string {
	agents := SupportedPromptTargetAgents()
	ids := make([]string, 0, len(agents))
	for _, agent := range agents {
		ids = append(ids, string(agent.ID))
	}
	return ids
}

func SupportedPromptTargetAgentIDsText() string {
	return strings.Join(SupportedPromptTargetAgentIDs(), ", ")
}

func ParsePromptTargetAgent(value string) (PromptTargetAgent, error) {
	agent := PromptTargetAgent(strings.TrimSpace(value))
	if agent == "" {
		return "", fmt.Errorf("prompt target agent is required")
	}
	for _, supported := range SupportedPromptTargetAgents() {
		if agent == supported.ID {
			return agent, nil
		}
	}
	return "", fmt.Errorf("unsupported prompt target agent: %s (supported: %s)", agent, SupportedPromptTargetAgentIDsText())
}

func PromptTargetAgentDisplayName(agent PromptTargetAgent) string {
	for _, supported := range SupportedPromptTargetAgents() {
		if agent == supported.ID {
			return supported.DisplayName
		}
	}
	return string(agent)
}

type PromptRenderRequest struct {
	ProjectRoot           string
	ChangeID              string
	Role                  PromptRole
	TargetAgent           PromptTargetAgent
	Task                  string
	ProjectContext        string
	ProjectBriefReadFirst string
}
