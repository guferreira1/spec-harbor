package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInteractiveGenerationCoreStaysTerminalFree(t *testing.T) {
	for _, directory := range []string{
		filepath.Join("..", "core", "domain"),
		filepath.Join("..", "core", "ports"),
		filepath.Join("..", "core", "usecase"),
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
				for _, forbidden := range []string{
					`"os"`,
					`"os/exec"`,
					`"github.com/charmbracelet/`,
					`"github.com/AlecAivazis/survey`,
					`"golang.org/x/term`,
					"os.Stdin",
					"os.Stdout",
					"os.Stderr",
					"ReadLine(",
					"IsInputTerminal(",
					"Proceed? [y/N]:",
					"Select generation path:",
					"interactiveTerminal",
				} {
					if strings.Contains(source, forbidden) {
						t.Fatalf("%s contains forbidden interactive terminal behavior %q", path, forbidden)
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

func TestInteractiveGenerationAdapterDoesNotIntroduceExecutionOrAutomation(t *testing.T) {
	sourcePath := filepath.Join("..", "adapters", "cli", "interactive_generation.go")
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
		"WorkflowDispatcher",
		"SourceControl",
		"GitCommit",
		"AutoCommitter",
		"AutoPusher",
		"AutoMerger",
		"CreatePullRequest",
		"CommitChanges",
		"PushChanges",
		"MergeChanges",
		"ArchiveChange",
		"WriteFile(",
		"WriteFileIfAbsent",
		"CreateDirectory",
		"Remove(",
		"shell",
		"script",
		"provider API",
		"LLM API",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden execution, automation, or write behavior %q", sourcePath, forbidden)
		}
	}
}
