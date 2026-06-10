package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHybridGenerationCoreDoesNotIntroduceExecutionOrAutomation(t *testing.T) {
	sources := []string{
		filepath.Join("..", "core", "domain", "hybrid_generation.go"),
		filepath.Join("..", "core", "usecase", "generate_hybrid_change.go"),
		filepath.Join("..", "core", "ports", "generation.go"),
	}
	forbiddenSnippets := []string{
		`"os"`,
		`"os/exec"`,
		`"net"`,
		`"net/http"`,
		"exec.Command",
		"http.Get",
		"http.Post",
		"ParseAIOutput",
		"WorkflowDispatcher",
		"SourceControl",
		"GitCommit",
		"AutoCommitter",
		"AutoPusher",
		"AutoMerger",
		"go-openai",
		"anthropic",
		"generative-ai-go",
		"go-github",
		"go-gitlab",
		"go-git",
	}

	for _, sourcePath := range sources {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range forbiddenSnippets {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden hybrid dependency or behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestHybridGenerationWriteSurfaceStaysOpenSpecOnly(t *testing.T) {
	useCasePath := filepath.Join("..", "core", "usecase", "generate_hybrid_change.go")
	contents, err := os.ReadFile(useCasePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", useCasePath, err)
	}
	source := string(contents)

	for _, required := range []string{
		"domain.RequiredOpenSpecChangeFiles()",
		"changePath+\"/\"+requiredFile",
		"openspecChangesDirectory + \"/\" + changeID",
		"WriteFileIfAbsent",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("%s missing expected bounded write snippet %q", useCasePath, required)
		}
	}

	for _, forbidden := range []string{
		"cmd/",
		".github/",
		".git/",
		"go.mod",
		"go.sum",
		"agent-prompts",
		"production/",
		"WriteFile(",
		"CreatePullRequest",
		"CommitChanges",
		"PushChanges",
		"MergeChanges",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden production or arbitrary write snippet %q", useCasePath, forbidden)
		}
	}
}
