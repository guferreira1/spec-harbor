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
