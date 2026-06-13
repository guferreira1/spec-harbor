package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryContextIndexDoesNotIntroduceExecutionRemoteAPIsOrRAG(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "repository_context_index.go"),
		filepath.Join("..", "core", "ports", "repository_context_index.go"),
		filepath.Join("..", "core", "usecase", "repository_context_index.go"),
		filepath.Join("..", "adapters", "cli", "context_discovery.go"),
		filepath.Join("..", "adapters", "filesystem", "repository_context_index.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range []string{
				`"os/exec"`,
				`"net/http"`,
				"exec.Command",
				"http.Get",
				"http.Post",
				"Embedding",
				"VectorStore",
				"RAG",
				"PromptInjection",
				"WorkflowDispatcher",
				"CommitChanges",
				"PushChanges",
				"MergeChanges",
				"go-github",
				"go-gitlab",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden repository context index behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}
