package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectBriefCLIAdapterDoesNotIntroduceExecutionRemoteAPIsOrDirectWrites(t *testing.T) {
	sourcePath := filepath.Join("..", "adapters", "cli", "project_brief.go")
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
		"go-openai",
		"anthropic",
		"generative-ai-go",
		"go-github",
		"go-gitlab",
		"go-git",
		"WriteFile(",
		"WriteFileIfAbsent",
		"CreateDirectory",
		"RepositoryIndex",
		"Embedding",
		"VectorStore",
		"WorkflowDispatcher",
		"CommitChanges",
		"PushChanges",
		"MergeChanges",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden briefing adapter behavior %q", sourcePath, forbidden)
		}
	}
}
