package domain

import (
	"reflect"
	"testing"
)

func TestSupportedPromptTargetAgentsAreStable(t *testing.T) {
	got := SupportedPromptTargetAgents()
	want := []PromptTargetAgentInfo{
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

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedPromptTargetAgents() = %#v, want %#v", got, want)
	}
	got[0].ID = "mutated"
	if SupportedPromptTargetAgents()[0].ID != PromptTargetAgentGeneric {
		t.Fatalf("SupportedPromptTargetAgents() returned mutable policy")
	}
}

func TestSupportedPromptTargetAgentIDsAreDeterministic(t *testing.T) {
	got := SupportedPromptTargetAgentIDs()
	want := []string{
		"generic",
		"codex",
		"claude-code",
		"devin",
		"cursor",
		"copilot",
		"gemini",
		"roo",
		"windsurf",
		"aider",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedPromptTargetAgentIDs() = %v, want %v", got, want)
	}
	if SupportedPromptTargetAgentIDsText() != "generic, codex, claude-code, devin, cursor, copilot, gemini, roo, windsurf, aider" {
		t.Fatalf("SupportedPromptTargetAgentIDsText() = %q", SupportedPromptTargetAgentIDsText())
	}
}

func TestDefaultPromptTargetAgentIsGeneric(t *testing.T) {
	if DefaultPromptTargetAgent() != PromptTargetAgentGeneric {
		t.Fatalf("DefaultPromptTargetAgent() = %q, want generic", DefaultPromptTargetAgent())
	}
}

func TestParsePromptTargetAgentAcceptsSupportedAgents(t *testing.T) {
	for _, agent := range SupportedPromptTargetAgents() {
		t.Run(string(agent.ID), func(t *testing.T) {
			got, err := ParsePromptTargetAgent(" " + string(agent.ID) + " ")
			if err != nil {
				t.Fatalf("ParsePromptTargetAgent(%q) error = %v", agent.ID, err)
			}
			if got != agent.ID {
				t.Fatalf("ParsePromptTargetAgent(%q) = %q, want %q", agent.ID, got, agent.ID)
			}
			if PromptTargetAgentDisplayName(got) != agent.DisplayName {
				t.Fatalf("PromptTargetAgentDisplayName(%q) = %q, want %q", got, PromptTargetAgentDisplayName(got), agent.DisplayName)
			}
		})
	}
}

func TestParsePromptTargetAgentRejectsEmptyAndUnknownValues(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "", want: "prompt target agent is required"},
		{value: " ", want: "prompt target agent is required"},
		{value: "unknown", want: "unsupported prompt target agent: unknown (supported: generic, codex, claude-code, devin, cursor, copilot, gemini, roo, windsurf, aider)"},
		{value: "Codex", want: "unsupported prompt target agent: Codex (supported: generic, codex, claude-code, devin, cursor, copilot, gemini, roo, windsurf, aider)"},
		{value: "claude", want: "unsupported prompt target agent: claude (supported: generic, codex, claude-code, devin, cursor, copilot, gemini, roo, windsurf, aider)"},
		{value: "claude_code", want: "unsupported prompt target agent: claude_code (supported: generic, codex, claude-code, devin, cursor, copilot, gemini, roo, windsurf, aider)"},
		{value: "claude-code extra", want: "unsupported prompt target agent: claude-code extra (supported: generic, codex, claude-code, devin, cursor, copilot, gemini, roo, windsurf, aider)"},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			_, err := ParsePromptTargetAgent(test.value)
			if err == nil {
				t.Fatalf("ParsePromptTargetAgent(%q) error = nil, want %q", test.value, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParsePromptTargetAgent(%q) error = %q, want %q", test.value, err.Error(), test.want)
			}
		})
	}
}
