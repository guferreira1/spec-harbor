package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHubRemoteContextCoreDoesNotImportHTTPExecutionProvidersOrRAG(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "github_remote_context.go"),
		filepath.Join("..", "core", "ports", "github_remote_context.go"),
		filepath.Join("..", "core", "usecase", "retrieve_github_remote_context.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range []string{
				`"net/http"`,
				`"os/exec"`,
				"exec.Command",
				"http.Get",
				"http.Post",
				"Embedding",
				"VectorStore",
				"SemanticSearch",
				"RAG",
				"PromptExecution",
				"AgentExecution",
				"WorkflowDispatcher",
				"go-github",
				"go-gitlab",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden GitHub remote context core behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestGitHubRemoteContextPortAndAdapterRemainReadOnly(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "ports", "github_remote_context.go"),
		filepath.Join("..", "adapters", "github", "http_remote_context.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range []string{
				"CreatePullRequest",
				"CreateIssue",
				"CreateComment",
				"CreateLabel",
				"CreateRelease",
				"CreateTag",
				"CreateBranch",
				"CreateCommit",
				"CreateRef",
				"UpdateRef",
				"DeleteRef",
				"MergePullRequest",
				"DispatchWorkflow",
				"TriggerWorkflow",
				"http.MethodPost",
				"http.MethodPatch",
				"http.MethodPut",
				"http.MethodDelete",
				`"os/exec"`,
				"exec.Command",
				"git commit",
				"git push",
				"gh pr",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden GitHub mutation behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestGitHubRemoteContextAdapterIsOnlyFeatureFileImportingHTTP(t *testing.T) {
	adapterPath := filepath.Join("..", "adapters", "github", "http_remote_context.go")
	contents, err := os.ReadFile(adapterPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", adapterPath, err)
	}
	if !strings.Contains(string(contents), `"net/http"`) || !strings.Contains(string(contents), "http.MethodGet") {
		t.Fatalf("%s does not contain the expected bounded GET-only HTTP adapter behavior", adapterPath)
	}

	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "github_remote_context.go"),
		filepath.Join("..", "core", "ports", "github_remote_context.go"),
		filepath.Join("..", "core", "usecase", "retrieve_github_remote_context.go"),
	} {
		contents, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
		}
		if strings.Contains(string(contents), `"net/http"`) {
			t.Fatalf("%s imports net/http outside the GitHub adapter", sourcePath)
		}
	}
}
