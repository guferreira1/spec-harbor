package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionPlatformPackageAvoidsRuntimeDiscoveryDependencies(t *testing.T) {
	versionDirectory := filepath.Join("..", "platform", "version")

	entries, err := os.ReadDir(versionDirectory)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", versionDirectory, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		sourcePath := filepath.Join(versionDirectory, entry.Name())
		t.Run(filepath.ToSlash(sourcePath), func(t *testing.T) {
			parsedFile, err := parser.ParseFile(token.NewFileSet(), sourcePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("ParseFile(%q) error = %v", sourcePath, err)
			}

			for _, imported := range parsedFile.Imports {
				importPath := strings.Trim(imported.Path.Value, `"`)
				if forbiddenVersionRuntimeImport(importPath) {
					t.Fatalf("%s imports forbidden version runtime dependency %q", sourcePath, importPath)
				}
			}

			source := mustReadArchitectureFile(t, sourcePath)
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
					t.Fatalf("%s contains forbidden version runtime behavior %q", sourcePath, forbidden)
				}
			}
		})
	}
}

func TestVersionCommandDoesNotWriteFilesOrUseRuntimeDiscovery(t *testing.T) {
	sourcePath := filepath.Join("..", "adapters", "cli", "cli.go")
	source := mustReadArchitectureFile(t, sourcePath)
	versionCommand := sourceBetween(t, source, "func versionCommand", "\nfunc parseVersionArguments")

	for _, forbidden := range []string{
		"os.",
		"WriteFile",
		"Mkdir",
		"Create",
		"Remove",
		"Rename",
		"exec.Command",
		"http.Get",
		"http.Post",
		".git",
		"git ",
	} {
		if strings.Contains(versionCommand, forbidden) {
			t.Fatalf("versionCommand contains forbidden behavior %q", forbidden)
		}
	}
}

func TestReleaseVersioningDoesNotIntroduceReleaseAutomationOrPackageArtifacts(t *testing.T) {
	root := filepath.Join("..", "..")

	for _, relativePath := range []string{
		".goreleaser.yaml",
		".goreleaser.yml",
		".github/workflows/release.yml",
		".github/workflows/release.yaml",
		"install.sh",
		"publish.sh",
		"release.sh",
		"scripts/install.sh",
		"scripts/publish",
		"scripts/publish.sh",
		"scripts/release",
		"scripts/release.sh",
		"package.json",
		"package-lock.json",
		"npm",
		"packages/npm",
		"Formula",
		"homebrew",
		"nfpm.yaml",
		".nfpm.yaml",
		"packaging",
		"debian",
		"rpm",
		"winget",
		"scoop",
	} {
		assertArchitecturePathDoesNotExist(t, filepath.Join(root, filepath.FromSlash(relativePath)))
	}

	workflowsDirectory := filepath.Join(root, ".github", "workflows")
	workflowEntries, err := os.ReadDir(workflowsDirectory)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("ReadDir(%q) error = %v", workflowsDirectory, err)
	}
	for _, entry := range workflowEntries {
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "release") {
			t.Fatalf("release-specific workflow file exists: %s", filepath.Join(workflowsDirectory, entry.Name()))
		}

		contents := mustReadArchitectureFile(t, filepath.Join(workflowsDirectory, entry.Name()))
		for _, forbidden := range []string{
			"goreleaser",
			"softprops/action-gh-release",
			"gh release",
			"npm publish",
			"brew tap",
		} {
			if strings.Contains(strings.ToLower(contents), forbidden) {
				t.Fatalf("%s contains release publishing behavior %q", entry.Name(), forbidden)
			}
		}
	}

	assertNoReleaseArtifacts(t, root)
}

func TestReleaseVersioningDocumentationDescribesImplementedScopeOnly(t *testing.T) {
	documents := map[string]string{
		"README.md":       filepath.Join("..", "..", "README.md"),
		"docs/usage.md":   filepath.Join("..", "..", "docs", "usage.md"),
		"docs/release.md": filepath.Join("..", "..", "docs", "release.md"),
	}

	requiredSnippets := []string{
		"specharbor version",
		"SpecHarbor dev",
		"commit: unknown",
		"date: unknown",
		"dirty: unknown",
		"`dev`",
		"`unknown`",
		"`v0.1.0`",
		"`0.1.0`",
		"github.com/guferreira1/spec-harbor/internal/platform/version",
		"-ldflags",
		"displays the injected version string as-is",
		"does not normalize",
		"GitHub Releases",
		"install scripts",
		"npm",
		"Homebrew",
		"future",
	}

	for name, path := range documents {
		t.Run(name, func(t *testing.T) {
			source := mustReadArchitectureFile(t, path)
			for _, snippet := range requiredSnippets {
				if !strings.Contains(source, snippet) {
					t.Fatalf("%s missing release versioning documentation snippet %q", name, snippet)
				}
			}
		})
	}
}

func forbiddenVersionRuntimeImport(importPath string) bool {
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

func assertArchitecturePathDoesNotExist(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err == nil {
		t.Fatalf("out-of-scope release/package artifact exists: %s", path)
	}
	if !os.IsNotExist(err) {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
}

func assertNoReleaseArtifacts(t *testing.T, root string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		lowerName := strings.ToLower(entry.Name())
		for _, suffix := range []string{".tar.gz", ".tgz", ".zip", ".sha256", ".sha512"} {
			if strings.HasSuffix(lowerName, suffix) {
				t.Fatalf("generated release artifact exists: %s", path)
			}
		}
		for _, name := range []string{"checksums.txt", "sha256sums", "sha256sums.txt"} {
			if lowerName == name {
				t.Fatalf("generated checksum artifact exists: %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) error = %v", root, err)
	}
}

func mustReadArchitectureFile(t *testing.T, path string) string {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(contents)
}

func sourceBetween(t *testing.T, source string, start string, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source missing start marker %q", start)
	}
	remaining := source[startIndex:]
	endIndex := strings.Index(remaining, end)
	if endIndex < 0 {
		t.Fatalf("source missing end marker %q after %q", end, start)
	}
	return remaining[:endIndex]
}
