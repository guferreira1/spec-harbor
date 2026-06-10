package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowCoreFilesStayInsideArchitectureBoundaries(t *testing.T) {
	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "workflow.go"),
		filepath.Join("..", "core", "usecase", "show_workflow.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			parsedFile, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("ParseFile(%q) error = %v", sourcePath, err)
			}

			for _, imported := range parsedFile.Imports {
				importPath := strings.Trim(imported.Path.Value, `"`)
				if forbiddenCoreImport(importPath) {
					t.Fatalf("%s imports forbidden dependency %q", sourcePath, importPath)
				}
			}
		})
	}
}

func TestWorkflowFeatureDoesNotIntroduceRemoteAutomationOrExecutionDependencies(t *testing.T) {
	forbiddenSnippets := []string{
		`"os"`,
		`"os/exec"`,
		`"net"`,
		`"net/http"`,
		"exec.Command",
		"http.Get",
		"http.Post",
		"go-github",
		"go-gitlab",
		"go-git",
		"WorkflowDispatcher",
		"CredentialStore",
	}

	for _, sourcePath := range []string{
		filepath.Join("..", "core", "domain", "workflow.go"),
		filepath.Join("..", "core", "usecase", "show_workflow.go"),
	} {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}

			source := string(contents)
			for _, forbidden := range forbiddenSnippets {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden workflow dependency %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestWorkflowConnectorAdapterIsNotIntroduced(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "adapters", "workflow"))
	if err == nil {
		t.Fatalf("workflow adapter directory exists, but workflow connectors are out of scope")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Stat(adapters/workflow) error = %v", err)
	}
}

func TestWorkflowDocumentationDescribesReadOnlyAdvisoryBoundaries(t *testing.T) {
	documents := map[string]string{
		"README.md":        filepath.Join("..", "..", "README.md"),
		"docs/usage.md":    filepath.Join("..", "..", "docs", "usage.md"),
		"docs/workflow.md": filepath.Join("..", "..", "docs", "workflow.md"),
	}

	requiredSnippets := []string{
		"`specharbor workflow`",
		"Spec Author Agent",
		"Architecture Reviewer Agent",
		"Implementer Agent",
		"Test Engineer Agent",
		"Change Reviewer Agent",
		"Commit",
		"Pull Request",
		"Merge",
		"Archive",
		"advisory",
		"read-only",
		"does not execute",
		"does not commit",
		"does not create PRs",
		"does not merge",
		"GitHub",
		"GitLab",
		"CI",
		"provider APIs",
		"agent CLIs",
		"remote automation",
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
					t.Fatalf("%s missing workflow documentation snippet %q", name, snippet)
				}
			}
		})
	}
}
