package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIAssistedGenerationDocumentationDescribesFromFileSafetyBoundaries(t *testing.T) {
	documents := map[string]string{
		"README.md":                filepath.Join("..", "..", "README.md"),
		"docs/usage.md":            filepath.Join("..", "..", "docs", "usage.md"),
		"docs/generation-modes.md": filepath.Join("..", "..", "docs", "generation-modes.md"),
	}

	requiredSnippets := []string{
		"--ai-assisted --from-file",
		"--overwrite",
		"local",
		"strict",
		"proposal.md",
		"design.md",
		"tasks.md",
		"acceptance-criteria.md",
		"risks.md",
		"validation",
		"symlink output targets",
		"provider APIs",
		"remote AI services",
		"production code",
		"source-control",
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
					t.Fatalf("%s missing AI-assisted generation documentation snippet %q", name, snippet)
				}
			}
		})
	}
}
