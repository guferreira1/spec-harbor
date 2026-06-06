package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.ValidationFileSystem = (*LocalFileSystem)(nil)

func TestLocalFileSystemCreatesDirectoriesAndFiles(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	exists, err := fileSystem.DirectoryExists(root, "openspec")
	if err != nil {
		t.Fatalf("DirectoryExists() error = %v", err)
	}
	if exists {
		t.Fatalf("DirectoryExists() = true, want false")
	}

	if err := fileSystem.CreateDirectory(root, "openspec"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}

	exists, err = fileSystem.DirectoryExists(root, "openspec")
	if err != nil {
		t.Fatalf("DirectoryExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("DirectoryExists() = false, want true")
	}

	created, err := fileSystem.WriteFileIfAbsent(root, "openspec/project.md", "project")
	if err != nil {
		t.Fatalf("WriteFileIfAbsent() error = %v", err)
	}
	if !created {
		t.Fatalf("WriteFileIfAbsent() created = false, want true")
	}

	contents, err := os.ReadFile(filepath.Join(root, "openspec", "project.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "project" {
		t.Fatalf("file contents = %q, want %q", string(contents), "project")
	}

	exists, err = fileSystem.FileExists(root, "openspec/project.md")
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("FileExists() = false, want true")
	}
}

func TestLocalFileSystemDoesNotOverwriteExistingFiles(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := fileSystem.CreateDirectory(root, ".specharbor"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	filePath := filepath.Join(root, ".specharbor", "config.yml")
	if err := os.WriteFile(filePath, []byte("original"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	created, err := fileSystem.WriteFileIfAbsent(root, ".specharbor/config.yml", "replacement")
	if err != nil {
		t.Fatalf("WriteFileIfAbsent() error = %v", err)
	}
	if created {
		t.Fatalf("WriteFileIfAbsent() created = true, want false")
	}

	contents, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "original" {
		t.Fatalf("file contents = %q, want %q", string(contents), "original")
	}
}

func TestLocalFileSystemDistinguishesMissingFilesAndDirectories(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	fileExists, err := fileSystem.FileExists(root, "openspec/project.md")
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if fileExists {
		t.Fatalf("FileExists() = true, want false")
	}

	directoryExists, err := fileSystem.DirectoryExists(root, "openspec/changes")
	if err != nil {
		t.Fatalf("DirectoryExists() error = %v", err)
	}
	if directoryExists {
		t.Fatalf("DirectoryExists() = true, want false")
	}
}

func TestLocalFileSystemDistinguishesFilesFromDirectories(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "openspec", "project.md"), []byte("project"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	fileExists, err := fileSystem.FileExists(root, "openspec/changes")
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if fileExists {
		t.Fatalf("FileExists() for directory = true, want false")
	}

	directoryExists, err := fileSystem.DirectoryExists(root, "openspec/project.md")
	if err != nil {
		t.Fatalf("DirectoryExists() error = %v", err)
	}
	if directoryExists {
		t.Fatalf("DirectoryExists() for file = true, want false")
	}
}
