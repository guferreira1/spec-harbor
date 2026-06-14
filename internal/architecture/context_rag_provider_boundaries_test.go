package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextRAGCoreDoesNotImportNetworkExecutionAdaptersOrProviderSDKs(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "context_rag.go"),
		filepath.Join("..", "core", "ports", "context_rag_provider.go"),
		filepath.Join("..", "core", "usecase", "generate_context_rag_answer.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			parsedFile, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("ParseFile(%q) error = %v", sourcePath, err)
			}
			for _, imported := range parsedFile.Imports {
				importPath := strings.Trim(imported.Path.Value, `"`)
				for _, forbidden := range []string{
					"net",
					"net/http",
					"os",
					"os/exec",
				} {
					if importPath == forbidden {
						t.Fatalf("%s imports forbidden dependency %q", sourcePath, importPath)
					}
				}
				for _, prefix := range []string{
					"github.com/guferreira1/spec-harbor/internal/adapters",
					"github.com/openai/",
					"github.com/anthropics/",
					"github.com/sashabaranov/go-openai",
					"github.com/google/generative-ai-go",
					"google.golang.org/genai",
				} {
					if strings.HasPrefix(importPath, prefix) {
						t.Fatalf("%s imports forbidden provider or adapter dependency %q", sourcePath, importPath)
					}
				}
			}
		})
	}
}

func TestContextRAGAdapterStaysProviderOnlyAndNoAutomation(t *testing.T) {
	adapterPath := filepath.Join("..", "adapters", "openai", "http_context_rag_provider.go")
	contents, err := os.ReadFile(adapterPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", adapterPath, err)
	}
	source := string(contents)
	for _, want := range []string{
		`"net/http"`,
		"http.MethodPost",
		"/responses",
		"Authorization",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("%s missing expected provider adapter behavior %q", adapterPath, want)
		}
	}
	for _, forbidden := range []string{
		`"os/exec"`,
		"exec.Command",
		"CreatePullRequest",
		"CreateIssue",
		"CreateComment",
		"CreateRelease",
		"CreateTag",
		"CreateBranch",
		"CreateCommit",
		"UpdateRef",
		"DeleteRef",
		"MergePullRequest",
		"DispatchWorkflow",
		"git commit",
		"git push",
		"gh pr",
		"WorkflowDispatcher",
		"AgentExecution",
		"VectorStore",
		"EmbeddingStore",
		"github.com/sashabaranov/go-openai",
		"github.com/openai/",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden provider behavior %q", adapterPath, forbidden)
		}
	}
}

func TestContextRAGDoesNotModifyPromptGenerationPath(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "prompt_project_context.go"),
		filepath.Join("..", "adapters", "cli", "cli.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range []string{
				"ContextRAG",
				"context rag",
				"SPECHARBOR_OPENAI_API_KEY",
				"NewHTTPContextRAGProvider",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains provider context prompt coupling %q", sourcePath, forbidden)
				}
			}
		})
	}
}
