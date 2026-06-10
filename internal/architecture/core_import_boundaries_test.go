package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCoreProductionImportsStayInsideArchitectureBoundaries(t *testing.T) {
	for _, directory := range []string{
		filepath.Join("..", "core", "domain"),
		filepath.Join("..", "core", "ports"),
		filepath.Join("..", "core", "usecase"),
	} {
		t.Run(filepath.ToSlash(directory), func(t *testing.T) {
			entries, err := os.ReadDir(directory)
			if err != nil {
				t.Fatalf("ReadDir(%q) error = %v", directory, err)
			}

			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
					continue
				}

				sourcePath := filepath.Join(directory, entry.Name())
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
			}
		})
	}
}

func TestAgentRunnerFoundationDoesNotIntroduceOutputApplicationOrAutomationPorts(t *testing.T) {
	forbiddenSnippets := []string{
		"AgentCommandRunner",
		"AgentOutputWriter",
		"AgentOutputApplier",
		"WriteAgentOutput",
		"ApplyAgentOutput",
		"WorkflowDispatcher",
		"SourceControl",
		"GitCommit",
		"AutoCommitter",
		"AutoPusher",
		"AutoMerger",
		"CommitChanges",
		"PushChanges",
		"MergeChanges",
		"CredentialStore",
		"OAuth",
	}

	for _, directory := range []string{
		filepath.Join("..", "core", "ports"),
		filepath.Join("..", "adapters"),
	} {
		t.Run(filepath.ToSlash(directory), func(t *testing.T) {
			err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
					return nil
				}

				contents, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				source := string(contents)
				for _, forbidden := range forbiddenSnippets {
					if strings.Contains(source, forbidden) {
						t.Fatalf("%s contains deferred execution/write abstraction %q", path, forbidden)
					}
				}
				return nil
			})
			if err != nil {
				t.Fatalf("WalkDir(%q) error = %v", directory, err)
			}
		})
	}
}

func TestAgentRunnerPortAndAdapterStayNarrow(t *testing.T) {
	portPath := filepath.Join("..", "core", "ports", "generation.go")
	portContents, err := os.ReadFile(portPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", portPath, err)
	}
	if !strings.Contains(string(portContents), "type AgentRunner interface") {
		t.Fatalf("%s does not define the core-owned AgentRunner port", portPath)
	}
	if strings.Contains(string(portContents), "WriteAgentOutput") ||
		strings.Contains(string(portContents), "ApplyAgentOutput") ||
		strings.Contains(string(portContents), "WorkflowDispatcher") {
		t.Fatalf("%s contains forbidden output application or workflow port", portPath)
	}

	adapterPath := filepath.Join("..", "adapters", "agentrunner", "local_command_runner.go")
	adapterContents, err := os.ReadFile(adapterPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", adapterPath, err)
	}
	adapterSource := string(adapterContents)
	for _, want := range []string{
		`"os/exec"`,
		"exec.Command(commandName, request.FixedArgs()...)",
		"command.Stdin = strings.NewReader(request.Prompt())",
		"domain.NewAgentRunResult",
	} {
		if !strings.Contains(adapterSource, want) {
			t.Fatalf("%s missing expected narrow local runner behavior %q", adapterPath, want)
		}
	}
	for _, forbidden := range []string{
		"ResolveExecutableAgentCommand",
		"RecognizedAgentTarget",
		"WriteAgentOutput",
		"ApplyAgentOutput",
		"WorkflowDispatcher",
		"git commit",
		"git push",
		"git merge",
	} {
		if strings.Contains(adapterSource, forbidden) {
			t.Fatalf("%s contains forbidden runner adapter behavior %q", adapterPath, forbidden)
		}
	}
}

func forbiddenCoreImport(importPath string) bool {
	for _, exact := range []string{
		"os",
		"os/exec",
		"net",
		"net/http",
	} {
		if importPath == exact {
			return true
		}
	}

	for _, prefix := range []string{
		"github.com/guferreira1/spec-harbor/internal/adapters",
		"github.com/guferreira1/spec-harbor/cmd",
		"github.com/openai/",
		"github.com/anthropics/",
		"github.com/google/generative-ai-go",
		"google.golang.org/genai",
		"github.com/sashabaranov/go-openai",
		"github.com/charmbracelet/",
		"github.com/spf13/",
		"github.com/AlecAivazis/survey",
		"github.com/google/go-github",
		"github.com/xanzy/go-gitlab",
		"github.com/go-git/go-git",
	} {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	return false
}
