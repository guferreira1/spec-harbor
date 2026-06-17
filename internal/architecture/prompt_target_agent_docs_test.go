package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPromptTargetAgentDocumentationDescribesRoleAgentBoundary(t *testing.T) {
	englishDocuments := map[string]string{
		"README.md":           filepath.Join("..", "..", "README.md"),
		"docs/en/usage.md":    filepath.Join("..", "..", "docs", "en", "usage.md"),
		"docs/en/agent-roles": filepath.Join("..", "..", "docs", "en", "agent-roles.md"),
		"docs/en/workflow.md": filepath.Join("..", "..", "docs", "en", "workflow.md"),
	}
	portugueseDocuments := map[string]string{
		"docs/pt-br/usage.md":    filepath.Join("..", "..", "docs", "pt-br", "usage.md"),
		"docs/pt-br/agent-roles": filepath.Join("..", "..", "docs", "pt-br", "agent-roles.md"),
		"docs/pt-br/workflow.md": filepath.Join("..", "..", "docs", "pt-br", "workflow.md"),
	}

	agentSnippets := []string{
		"`generic`",
		"`codex`",
		"`claude-code`",
		"`devin`",
		"`cursor`",
		"`copilot`",
		"`gemini`",
		"`roo`",
		"`windsurf`",
		"`aider`",
	}

	assertPromptTargetAgentDocs(t, englishDocuments, append([]string{
		"--role",
		"--agent",
		"default",
		"does not execute",
		"only adapts",
	}, agentSnippets...))
	assertPromptTargetAgentDocs(t, portugueseDocuments, append([]string{
		"--role",
		"--agent",
		"padrão",
		"não executa",
		"adapta",
	}, agentSnippets...))
}

func assertPromptTargetAgentDocs(t *testing.T, documents map[string]string, snippets []string) {
	t.Helper()

	for name, path := range documents {
		t.Run(name, func(t *testing.T) {
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", path, err)
			}
			source := string(contents)
			for _, snippet := range snippets {
				if !strings.Contains(source, snippet) {
					t.Fatalf("%s missing prompt target agent documentation snippet %q", name, snippet)
				}
			}
		})
	}
}
