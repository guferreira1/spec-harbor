package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.ValidationFileSystem = (*LocalFileSystem)(nil)
var _ ports.GenerationFileSystem = (*LocalFileSystem)(nil)
var _ ports.ArchiveFileSystem = (*LocalFileSystem)(nil)

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

func TestLocalFileSystemChecksAnyPathExistence(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "openspec", "project.md"), []byte("project"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tests := []struct {
		name         string
		relativePath string
		want         bool
	}{
		{name: "directory", relativePath: "openspec/changes", want: true},
		{name: "file", relativePath: "openspec/project.md", want: true},
		{name: "missing", relativePath: "openspec/archive", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exists, err := fileSystem.PathExists(root, test.relativePath)
			if err != nil {
				t.Fatalf("PathExists() error = %v", err)
			}
			if exists != test.want {
				t.Fatalf("PathExists() = %v, want %v", exists, test.want)
			}
		})
	}
}

func TestLocalFileSystemMovesDirectoriesAndPreservesNestedContents(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()
	sourcePath := filepath.Join(root, "openspec", "changes", "change", "nested")
	if err := os.MkdirAll(sourcePath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "openspec", "archive", "2026-06-06"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcePath, "proposal.md"), []byte("proposal"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := fileSystem.MoveDirectory(
		root,
		"openspec/changes/change",
		"openspec/archive/2026-06-06/change",
	)
	if err != nil {
		t.Fatalf("MoveDirectory() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "openspec", "changes", "change")); !os.IsNotExist(err) {
		t.Fatalf("source directory stat error = %v, want not exist", err)
	}

	contents, err := os.ReadFile(filepath.Join(root, "openspec", "archive", "2026-06-06", "change", "nested", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(contents) != "proposal" {
		t.Fatalf("moved file contents = %q, want proposal", string(contents))
	}
}

func TestLocalFileSystemMoveDirectoryRejectsFileSource(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()
	sourcePath := filepath.Join(root, "openspec", "changes", "change")
	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatalf("MkdirAll(source parent) error = %v", err)
	}
	if err := os.WriteFile(sourcePath, []byte("source file"), 0o644); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "openspec", "archive", "2026-06-06"), 0o755); err != nil {
		t.Fatalf("MkdirAll(archive parent) error = %v", err)
	}

	err := fileSystem.MoveDirectory(
		root,
		"openspec/changes/change",
		"openspec/archive/2026-06-06/change",
	)
	if err == nil {
		t.Fatalf("MoveDirectory() error = nil, want source file rejection")
	}

	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile(source) error = %v", err)
	}
	if string(contents) != "source file" {
		t.Fatalf("source contents = %q, want source file", string(contents))
	}
	if _, err := os.Stat(filepath.Join(root, "openspec", "archive", "2026-06-06", "change")); !os.IsNotExist(err) {
		t.Fatalf("destination stat error = %v, want not exist", err)
	}
}

func TestLocalFileSystemMoveDirectoryDoesNotOverwriteExistingDestination(t *testing.T) {
	tests := []struct {
		name            string
		createTarget    func(t *testing.T, root string)
		assertPreserved func(t *testing.T, root string)
	}{
		{
			name: "destination file",
			createTarget: func(t *testing.T, root string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(root, "openspec", "archive", "2026-06-06", "change"), []byte("existing file"), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			assertPreserved: func(t *testing.T, root string) {
				t.Helper()
				contents, err := os.ReadFile(filepath.Join(root, "openspec", "archive", "2026-06-06", "change"))
				if err != nil {
					t.Fatalf("ReadFile() error = %v", err)
				}
				if string(contents) != "existing file" {
					t.Fatalf("destination file contents = %q, want existing file", string(contents))
				}
			},
		},
		{
			name: "destination directory",
			createTarget: func(t *testing.T, root string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Join(root, "openspec", "archive", "2026-06-06", "change"), 0o755); err != nil {
					t.Fatalf("MkdirAll() error = %v", err)
				}
				if err := os.WriteFile(filepath.Join(root, "openspec", "archive", "2026-06-06", "change", "existing.md"), []byte("existing directory"), 0o644); err != nil {
					t.Fatalf("WriteFile() error = %v", err)
				}
			},
			assertPreserved: func(t *testing.T, root string) {
				t.Helper()
				contents, err := os.ReadFile(filepath.Join(root, "openspec", "archive", "2026-06-06", "change", "existing.md"))
				if err != nil {
					t.Fatalf("ReadFile() error = %v", err)
				}
				if string(contents) != "existing directory" {
					t.Fatalf("destination directory contents = %q, want existing directory", string(contents))
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			fileSystem := NewLocalFileSystem()
			if err := os.MkdirAll(filepath.Join(root, "openspec", "changes", "change"), 0o755); err != nil {
				t.Fatalf("MkdirAll() error = %v", err)
			}
			if err := os.WriteFile(filepath.Join(root, "openspec", "changes", "change", "proposal.md"), []byte("source"), 0o644); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			if err := os.MkdirAll(filepath.Join(root, "openspec", "archive", "2026-06-06"), 0o755); err != nil {
				t.Fatalf("MkdirAll() error = %v", err)
			}
			test.createTarget(t, root)

			err := fileSystem.MoveDirectory(
				root,
				"openspec/changes/change",
				"openspec/archive/2026-06-06/change",
			)
			if err == nil {
				t.Fatalf("MoveDirectory() error = nil, want existing destination error")
			}

			sourceContents, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "change", "proposal.md"))
			if err != nil {
				t.Fatalf("ReadFile(source) error = %v", err)
			}
			if string(sourceContents) != "source" {
				t.Fatalf("source contents = %q, want source", string(sourceContents))
			}
			test.assertPreserved(t, root)
		})
	}
}
