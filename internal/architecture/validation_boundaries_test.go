package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The validation filesystem port must stay read-only: validation never writes
// files, so the port must expose no write operations.
func TestValidationFileSystemPortStaysReadOnly(t *testing.T) {
	portPath := filepath.Join("..", "core", "ports", "validation.go")
	contents, err := os.ReadFile(portPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", portPath, err)
	}
	source := string(contents)

	for _, required := range []string{
		"DirectoryExists(root string, relativePath string) (bool, error)",
		"FileExists(root string, relativePath string) (bool, error)",
		"ReadFile(root string, relativePath string) (string, error)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("%s missing required read-only method %q", portPath, required)
		}
	}

	for _, forbidden := range []string{
		"Write",
		"Create",
		"Move",
		"Remove",
		"Delete",
		"Rename",
		"Mkdir",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("%s contains forbidden write operation %q", portPath, forbidden)
		}
	}
}

// Domain validation rules must stay pure: no filesystem, terminal, network,
// or adapter/template imports. Boilerplate detection must rely only on
// domain-owned markers, never on adapter template files.
func TestDomainValidationStaysPureAndTemplateFree(t *testing.T) {
	domainDirectory := filepath.Join("..", "core", "domain")
	allowedImports := map[string]bool{
		"errors":  true,
		"fmt":     true,
		"regexp":  true,
		"sort":    true,
		"strings": true,
		"unicode": true,
	}

	entries, err := os.ReadDir(domainDirectory)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", domainDirectory, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		sourcePath := filepath.Join(domainDirectory, entry.Name())
		parsedFile, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%q) error = %v", sourcePath, err)
		}

		for _, imported := range parsedFile.Imports {
			importPath := strings.Trim(imported.Path.Value, `"`)
			if strings.Contains(importPath, "internal/adapters") {
				t.Fatalf("%s imports adapter package %q", sourcePath, importPath)
			}
			if !allowedImports[importPath] {
				t.Fatalf("%s imports %q, not in the allowed pure-domain import list", sourcePath, importPath)
			}
		}
	}
}

// Validation rule logic must live in domain, not in the use case or CLI: the
// use case orchestrates and the CLI formats, so neither may embed starter
// marker wording.
func TestStarterMarkersAreOwnedByDomainOnly(t *testing.T) {
	markerSample := "Describe the problem this change should solve and who is affected."

	domainContentPath := filepath.Join("..", "core", "domain", "validation_content.go")
	domainSource, err := os.ReadFile(domainContentPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", domainContentPath, err)
	}
	if !strings.Contains(string(domainSource), markerSample) {
		t.Fatalf("%s does not define the canonical starter markers", domainContentPath)
	}

	for _, sourcePath := range []string{
		filepath.Join("..", "core", "usecase", "validate_change.go"),
		filepath.Join("..", "adapters", "cli", "cli.go"),
	} {
		contents, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
		}
		if strings.Contains(string(contents), markerSample) {
			t.Fatalf("%s duplicates domain-owned starter markers", sourcePath)
		}
	}
}
