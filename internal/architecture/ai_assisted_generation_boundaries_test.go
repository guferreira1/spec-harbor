package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIAssistedGenerationCoreDoesNotIntroduceExternalAutomation(t *testing.T) {
	sources := []string{
		filepath.Join("..", "core", "domain", "ai_assisted_generation.go"),
		filepath.Join("..", "core", "usecase", "generate_ai_assisted_change.go"),
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
		"WorkflowDispatcher",
		"SourceControl",
		"GitCommit",
		"AutoCommitter",
		"AutoPusher",
		"AutoMerger",
		"CredentialStore",
		"OAuth",
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
					t.Fatalf("%s contains forbidden AI-assisted dependency or automation %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestAIAssistedGenerationDoesNotIntroduceProviderWorkflowOrSourceControlAdapters(t *testing.T) {
	for _, directory := range []string{
		filepath.Join("..", "adapters", "aiprovider"),
		filepath.Join("..", "adapters", "provider"),
		filepath.Join("..", "adapters", "workflow"),
		filepath.Join("..", "adapters", "sourcecontrol"),
		filepath.Join("..", "adapters", "git"),
	} {
		t.Run(filepath.ToSlash(directory), func(t *testing.T) {
			_, err := os.Stat(directory)
			if err == nil {
				t.Fatalf("%s exists, but AI-assisted from-file generation must not add this adapter surface", directory)
			}
			if !os.IsNotExist(err) {
				t.Fatalf("Stat(%q) error = %v", directory, err)
			}
		})
	}
}

func TestAIAssistedGenerationWriteSurfaceStaysOpenSpecOnly(t *testing.T) {
	useCasePath := filepath.Join("..", "core", "usecase", "generate_ai_assisted_change.go")
	contents, err := os.ReadFile(useCasePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", useCasePath, err)
	}
	source := string(contents)

	for _, required := range []string{
		"domain.RequiredAIGeneratedFileNames()",
		"changePath + \"/\" + fileName",
		"openspecChangesDirectory + \"/\" + changeID",
		"WriteFileIfAbsent",
		"WriteFile(projectRoot",
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
		"production",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden production or arbitrary write path snippet %q", useCasePath, forbidden)
		}
	}
}

func TestAIAssistedGenerationSymlinkHandlingStaysInFilesystemAdapter(t *testing.T) {
	coreSources := []string{
		filepath.Join("..", "core", "domain", "ai_assisted_generation.go"),
		filepath.Join("..", "core", "usecase", "generate_ai_assisted_change.go"),
		filepath.Join("..", "core", "ports", "generation.go"),
	}
	for _, sourcePath := range coreSources {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range []string{"Lstat", "ModeSymlink", "EvalSymlinks"} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains filesystem symlink handling %q; this must stay in adapters", sourcePath, forbidden)
				}
			}
		})
	}

	adapterPath := filepath.Join("..", "adapters", "filesystem", "local.go")
	contents, err := os.ReadFile(adapterPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", adapterPath, err)
	}
	adapterSource := string(contents)
	for _, want := range []string{"os.Lstat", "os.ModeSymlink", "EnsureSafeWriteTarget"} {
		if !strings.Contains(adapterSource, want) {
			t.Fatalf("%s missing expected symlink safety implementation %q", adapterPath, want)
		}
	}
}
