package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.CustomTemplateFileSystem = (*LocalFileSystem)(nil)

func TestLocalFileSystemReadsCustomTemplateFilesUnderProjectRoot(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	templateDirectory := filepath.Join(root, ".specharbor", "templates", "api-feature")
	if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		contents := "# Template " + requiredFile + "\n"
		if err := os.WriteFile(filepath.Join(templateDirectory, requiredFile), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%q) error = %v", requiredFile, err)
		}
	}

	exists, err := fileSystem.DirectoryExists(root, ".specharbor/templates/api-feature")
	if err != nil {
		t.Fatalf("DirectoryExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("DirectoryExists() = false, want true")
	}

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		relativePath := ".specharbor/templates/api-feature/" + requiredFile
		fileExists, err := fileSystem.FileExists(root, relativePath)
		if err != nil {
			t.Fatalf("FileExists(%q) error = %v", relativePath, err)
		}
		if !fileExists {
			t.Fatalf("FileExists(%q) = false, want true", relativePath)
		}

		contents, err := fileSystem.ReadFile(root, relativePath)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", relativePath, err)
		}
		if contents != "# Template "+requiredFile+"\n" {
			t.Fatalf("ReadFile(%q) = %q, want template contents", relativePath, contents)
		}
	}
}

func TestLocalFileSystemCustomTemplateReadsResolveOnlyUnderProvidedRoot(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "project")
	templateDirectory := filepath.Join(root, ".specharbor", "templates", "api-feature")
	if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	outsideFile := filepath.Join(parent, "secret.md")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fileSystem := NewLocalFileSystem()

	insidePath := filepath.Join(root, ".specharbor", "templates", "api-feature", "proposal.md")
	if err := os.WriteFile(insidePath, []byte("inside"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	contents, err := fileSystem.ReadFile(root, ".specharbor/templates/api-feature/proposal.md")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if contents != "inside" {
		t.Fatalf("ReadFile() = %q, want inside", contents)
	}

	exists, err := fileSystem.FileExists(root, "secret.md")
	if err != nil {
		t.Fatalf("FileExists(secret.md) error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists(secret.md) = true, want false: read resolved outside the project root")
	}
}

func TestLocalFileSystemReportsMissingCustomTemplateFilesDistinctlyFromReadErrors(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	templateDirectory := filepath.Join(root, ".specharbor", "templates", "api-feature")
	if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	exists, err := fileSystem.FileExists(root, ".specharbor/templates/api-feature/proposal.md")
	if err != nil {
		t.Fatalf("FileExists() error = %v, want nil for a missing file", err)
	}
	if exists {
		t.Fatalf("FileExists() = true, want false for a missing file")
	}

	if _, err := fileSystem.ReadFile(root, ".specharbor/templates/api-feature/proposal.md"); err == nil {
		t.Fatalf("ReadFile() error = nil, want error for a missing file")
	}

	missingDirectory, err := fileSystem.DirectoryExists(root, ".specharbor/templates/unknown-template")
	if err != nil {
		t.Fatalf("DirectoryExists() error = %v, want nil for a missing directory", err)
	}
	if missingDirectory {
		t.Fatalf("DirectoryExists() = true, want false for a missing directory")
	}
}

func TestLocalFileSystemCustomTemplateLoadingDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	templateDirectory := filepath.Join(root, ".specharbor", "templates", "api-feature")
	if err := os.MkdirAll(templateDirectory, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDirectory, "proposal.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	entriesBefore := listAllPaths(t, root)

	if _, err := fileSystem.DirectoryExists(root, ".specharbor/templates/api-feature"); err != nil {
		t.Fatalf("DirectoryExists() error = %v", err)
	}
	if _, err := fileSystem.FileExists(root, ".specharbor/templates/api-feature/proposal.md"); err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if _, err := fileSystem.ReadFile(root, ".specharbor/templates/api-feature/proposal.md"); err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	entriesAfter := listAllPaths(t, root)
	if len(entriesBefore) != len(entriesAfter) {
		t.Fatalf("filesystem entries changed during template loading: before %v, after %v", entriesBefore, entriesAfter)
	}
	for index := range entriesBefore {
		if entriesBefore[index] != entriesAfter[index] {
			t.Fatalf("filesystem entries changed during template loading: before %v, after %v", entriesBefore, entriesAfter)
		}
	}
}

func listAllPaths(t *testing.T, root string) []string {
	t.Helper()

	var paths []string
	err := filepath.WalkDir(root, func(path string, _ os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir() error = %v", err)
	}
	return paths
}
