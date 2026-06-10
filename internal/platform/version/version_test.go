package version

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentUsesDefaultMetadataValues(t *testing.T) {
	metadata := Current()

	if metadata.Version != defaultVersion {
		t.Fatalf("Version = %q, want %q", metadata.Version, defaultVersion)
	}
	if metadata.Commit != defaultCommit {
		t.Fatalf("Commit = %q, want %q", metadata.Commit, defaultCommit)
	}
	if metadata.Date != defaultDate {
		t.Fatalf("Date = %q, want %q", metadata.Date, defaultDate)
	}
	if metadata.Dirty != defaultDirty {
		t.Fatalf("Dirty = %q, want %q", metadata.Dirty, defaultDirty)
	}
}

func TestNewMetadataFallsBackToDefaultsForEmptyValues(t *testing.T) {
	metadata := NewMetadata("", "", "", "")

	if metadata != (Metadata{
		Version: defaultVersion,
		Commit:  defaultCommit,
		Date:    defaultDate,
		Dirty:   defaultDirty,
	}) {
		t.Fatalf("metadata = %#v, want defaults", metadata)
	}
}

func TestMetadataFormatDefaultOutput(t *testing.T) {
	metadata := NewMetadata("", "", "", "")

	want := `SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown`
	if metadata.Format() != want {
		t.Fatalf("Format() = %q, want %q", metadata.Format(), want)
	}
}

func TestMetadataFormatInjectedPlainSemVerLikeValues(t *testing.T) {
	metadata := NewMetadata("0.1.0", "abc1234", "2026-06-10T19:00:00Z", "false")

	want := `SpecHarbor 0.1.0
commit: abc1234
date: 2026-06-10T19:00:00Z
dirty: false`
	if metadata.Format() != want {
		t.Fatalf("Format() = %q, want %q", metadata.Format(), want)
	}
}

func TestMetadataPreservesInjectedVersionValueAsIs(t *testing.T) {
	metadata := NewMetadata("v0.1.0", "abc1234", "2026-06-10T19:00:00Z", "false")

	if metadata.Version != "v0.1.0" {
		t.Fatalf("Version = %q, want preserved injected value", metadata.Version)
	}
	if got, want := firstLine(metadata.Format()), "SpecHarbor v0.1.0"; got != want {
		t.Fatalf("first line = %q, want %q", got, want)
	}
}

func TestProductionVersionPackageAvoidsRuntimeGitShellNetworkAndFilesystemDependencies(t *testing.T) {
	for _, sourcePath := range productionGoFiles(t) {
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			parsedFile, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("ParseFile(%q) error = %v", sourcePath, err)
			}

			for _, imported := range parsedFile.Imports {
				importPath := strings.Trim(imported.Path.Value, `"`)
				if forbiddenVersionImport(importPath) {
					t.Fatalf("%s imports forbidden runtime dependency %q", sourcePath, importPath)
				}
			}

			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", sourcePath, err)
			}
			source := string(contents)
			for _, forbidden := range []string{
				".git",
				"git tag",
				"git describe",
				"exec.Command",
				"http.Get",
				"http.Post",
				"net.Dial",
				"TrimPrefix",
				"TrimSuffix",
			} {
				if strings.Contains(source, forbidden) {
					t.Fatalf("%s contains forbidden runtime behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(value, "\n")
	return line
}

func productionGoFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir(.) error = %v", err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		paths = append(paths, entry.Name())
	}
	return paths
}

func forbiddenVersionImport(importPath string) bool {
	for _, exact := range []string{
		"io/fs",
		"net",
		"net/http",
		"os",
		"os/exec",
		"path",
		"path/filepath",
	} {
		if importPath == exact {
			return true
		}
	}

	for _, prefix := range []string{
		"github.com/go-git/",
		"github.com/google/go-github",
		"github.com/xanzy/go-gitlab",
		"gopkg.in/src-d/go-git",
	} {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	return false
}
