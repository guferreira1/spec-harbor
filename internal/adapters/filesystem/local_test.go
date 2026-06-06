package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

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
