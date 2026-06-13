package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalContextRetrievalDoesNotIntroduceExecutionRemoteAPIsOrRAG(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "local_context_retrieval.go"),
		filepath.Join("..", "core", "usecase", "retrieve_local_context.go"),
		filepath.Join("..", "adapters", "cli", "context_discovery.go"),
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
				"SemanticSearch",
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
					t.Fatalf("%s contains forbidden local context retrieval behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}
