package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.ContextDiscoveryFileSystem = (*ContextDiscoveryFileSystem)(nil)

func TestContextDiscoveryFileSystemReadsFilesWithLimit(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewContextDiscoveryFileSystem()
	path := filepath.Join(root, "README.md")
	if err := os.WriteFile(path, []byte("project context\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	exists, err := fileSystem.FileExists(root, "README.md")
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("FileExists() = false, want true")
	}

	contents, err := fileSystem.ReadFile(root, "README.md", 64)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if contents != "project context\n" {
		t.Fatalf("ReadFile() = %q, want README contents", contents)
	}

	_, err = fileSystem.ReadFile(root, "README.md", 4)
	if err == nil || !strings.Contains(err.Error(), "context discovery file exceeds 4 bytes") {
		t.Fatalf("ReadFile(over limit) error = %v, want size limit error", err)
	}
}

func TestContextDiscoveryFileSystemListsDirectoryEntriesWithoutFollowingSymlinks(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewContextDiscoveryFileSystem()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatalf("MkdirAll(docs) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "usage.md"), []byte("# Usage\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(usage.md) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "docs", "nested"), 0o755); err != nil {
		t.Fatalf("MkdirAll(nested) error = %v", err)
	}
	if err := os.Symlink(filepath.Join(root, "docs", "usage.md"), filepath.Join(root, "docs", "linked.md")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	entries, err := fileSystem.ListDirectory(root, "docs")
	if err != nil {
		t.Fatalf("ListDirectory() error = %v", err)
	}

	var sawFile, sawDirectory, sawSymlink bool
	for _, entry := range entries {
		switch entry.Name {
		case "usage.md":
			sawFile = entry.IsRegular && !entry.IsSymlink
		case "nested":
			sawDirectory = entry.IsDirectory && !entry.IsSymlink
		case "linked.md":
			sawSymlink = entry.IsSymlink && !entry.IsRegular && !entry.IsDirectory
		}
	}
	if !sawFile || !sawDirectory || !sawSymlink {
		t.Fatalf("entries = %+v, want regular file, directory, and symlink marker", entries)
	}

	exists, err := fileSystem.FileExists(root, "docs/linked.md")
	if err != nil {
		t.Fatalf("FileExists(symlink) error = %v, want symlink skipped without error", err)
	}
	if exists {
		t.Fatalf("FileExists(symlink) = true, want false")
	}

	_, err = fileSystem.ReadFile(root, "docs/linked.md", 64)
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("ReadFile(symlink) error = %v, want symlink rejection", err)
	}
}

func TestContextDiscoveryFileSystemSkipsSymlinkDirectories(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewContextDiscoveryFileSystem()
	target := filepath.Join(root, "target-docs")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll(target) error = %v", err)
	}
	if err := os.Symlink(target, filepath.Join(root, "docs")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	exists, err := fileSystem.DirectoryExists(root, "docs")
	if err != nil {
		t.Fatalf("DirectoryExists(symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("DirectoryExists(symlink) = true, want false")
	}
}

func TestContextDiscoveryFileSystemSkipsIntermediateSymlinkDirectories(t *testing.T) {
	root := t.TempDir()
	outsideOpenSpec := t.TempDir()
	fileSystem := NewContextDiscoveryFileSystem()
	if err := os.WriteFile(filepath.Join(outsideOpenSpec, "project.md"), []byte("# Outside\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside project.md) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(outsideOpenSpec, "specs"), 0o755); err != nil {
		t.Fatalf("MkdirAll(outside specs) error = %v", err)
	}
	if err := os.Symlink(outsideOpenSpec, filepath.Join(root, "openspec")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	exists, err := fileSystem.FileExists(root, "openspec/project.md")
	if err != nil {
		t.Fatalf("FileExists(intermediate symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists(intermediate symlink) = true, want false")
	}

	exists, err = fileSystem.DirectoryExists(root, "openspec/specs")
	if err != nil {
		t.Fatalf("DirectoryExists(intermediate symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("DirectoryExists(intermediate symlink) = true, want false")
	}

	_, err = fileSystem.ReadFile(root, "openspec/project.md", 64)
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("ReadFile(intermediate symlink) error = %v, want symlink rejection", err)
	}
}

func TestContextDiscoveryFileSystemDoesNotTraverseSymlinkedDocsDirectory(t *testing.T) {
	root := t.TempDir()
	outsideDocs := t.TempDir()
	fileSystem := NewContextDiscoveryFileSystem()
	if err := os.WriteFile(filepath.Join(outsideDocs, "usage.md"), []byte("# Outside docs\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside docs) error = %v", err)
	}
	if err := os.Symlink(outsideDocs, filepath.Join(root, "docs")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	exists, err := fileSystem.DirectoryExists(root, "docs")
	if err != nil {
		t.Fatalf("DirectoryExists(docs symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("DirectoryExists(docs symlink) = true, want false")
	}

	_, err = fileSystem.ListDirectory(root, "docs")
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("ListDirectory(docs symlink) error = %v, want symlink rejection", err)
	}
}

func TestContextDiscoveryFileSystemSkipsSupportedSymlinkFiles(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	fileSystem := NewContextDiscoveryFileSystem()
	outsideReadme := filepath.Join(outside, "README.md")
	if err := os.WriteFile(outsideReadme, []byte("# Outside README\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside README) error = %v", err)
	}
	if err := os.Symlink(outsideReadme, filepath.Join(root, "README.md")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	exists, err := fileSystem.FileExists(root, "README.md")
	if err != nil {
		t.Fatalf("FileExists(README symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists(README symlink) = true, want false")
	}

	_, err = fileSystem.ReadFile(root, "README.md", 64)
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("ReadFile(README symlink) error = %v, want symlink rejection", err)
	}
}
