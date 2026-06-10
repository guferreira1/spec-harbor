package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentRunnerDocumentationDescribesMappingsAndSafetyBoundaries(t *testing.T) {
	documents := map[string]string{
		"README.md":                filepath.Join("..", "..", "README.md"),
		"docs/usage.md":            filepath.Join("..", "..", "docs", "usage.md"),
		"docs/generation-modes.md": filepath.Join("..", "..", "docs", "generation-modes.md"),
	}

	requiredSnippets := []string{
		"Dry-run remains the default",
		"`generic`",
		"`codex -> codex`",
		"`claude -> claude`",
		"`devin -> devin`",
		"`cursor -> cursor`",
		"`copilot -> copilot`",
		"`gemini -> gemini`",
		"`roo -> roo`",
		"`windsurf -> windsurf`",
		"`aider -> aider`",
		"run-and-report",
		"does not parse or apply",
		"does not write",
		"does not modify production code",
		"does not auto-commit, auto-push, or auto-merge",
	}

	for name, path := range documents {
		t.Run(name, func(t *testing.T) {
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", path, err)
			}
			source := string(contents)
			for _, snippet := range requiredSnippets {
				if !strings.Contains(source, snippet) {
					t.Fatalf("%s missing documentation snippet %q", name, snippet)
				}
			}
		})
	}
}
