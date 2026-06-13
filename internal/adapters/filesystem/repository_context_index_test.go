package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.RepositoryContextIndexFileSystem = (*RepositoryContextIndexFileSystem)(nil)

func TestRepositoryContextIndexFileSystemReadsMetadataAndBytesWithoutRawSafetyBypass(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewRepositoryContextIndexFileSystem()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	metadata, err := fileSystem.FileMetadata(root, "README.md")
	if err != nil {
		t.Fatalf("FileMetadata() error = %v", err)
	}
	if metadata.SizeBytes != int64(len("# Project\n")) {
		t.Fatalf("SizeBytes = %d, want README size", metadata.SizeBytes)
	}
	if metadata.ModifiedTime.IsZero() {
		t.Fatalf("ModifiedTime is zero")
	}

	contents, err := fileSystem.ReadFileBytes(root, "README.md", 64)
	if err != nil {
		t.Fatalf("ReadFileBytes() error = %v", err)
	}
	if string(contents) != "# Project\n" {
		t.Fatalf("ReadFileBytes() = %q, want README bytes", string(contents))
	}

	_, err = fileSystem.ReadFileBytes(root, "README.md", 4)
	if err == nil || !strings.Contains(err.Error(), "repository context index file exceeds 4 bytes") {
		t.Fatalf("ReadFileBytes(over limit) error = %v, want size limit", err)
	}
}

func TestRepositoryContextIndexFileSystemRejectsUnsafePaths(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewRepositoryContextIndexFileSystem()

	for _, relativePath := range []string{"../README.md", "/tmp/README.md", `C:\tmp\README.md`, "docs/\x00secret.md"} {
		t.Run(relativePath, func(t *testing.T) {
			_, err := fileSystem.FileMetadata(root, relativePath)
			if err == nil {
				t.Fatalf("FileMetadata(%q) error = nil, want unsafe path error", relativePath)
			}
		})
	}
}

func TestRepositoryContextIndexFileSystemDoesNotFollowSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	fileSystem := NewRepositoryContextIndexFileSystem()
	if err := os.WriteFile(filepath.Join(outside, "README.md"), []byte("# Outside\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside) error = %v", err)
	}
	if err := os.Symlink(filepath.Join(outside, "README.md"), filepath.Join(root, "README.md")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	exists, err := fileSystem.FileExists(root, "README.md")
	if err != nil {
		t.Fatalf("FileExists(symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists(symlink) = true, want false")
	}
	_, err = fileSystem.FileMetadata(root, "README.md")
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("FileMetadata(symlink) error = %v, want symlink rejection", err)
	}
	_, err = fileSystem.ReadFileBytes(root, "README.md", 64)
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("ReadFileBytes(symlink) error = %v, want symlink rejection", err)
	}
}

func TestRepositoryContextIndexFileSystemRejectsIntermediateSymlinkDirectories(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	fileSystem := NewRepositoryContextIndexFileSystem()
	if err := os.WriteFile(filepath.Join(outside, "project.md"), []byte("# Outside\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside project) error = %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "openspec")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	exists, err := fileSystem.FileExists(root, "openspec/project.md")
	if err != nil {
		t.Fatalf("FileExists(intermediate symlink) error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists(intermediate symlink) = true, want false")
	}
	_, err = fileSystem.ReadFileBytes(root, "openspec/project.md", 64)
	if err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("ReadFileBytes(intermediate symlink) error = %v, want symlink rejection", err)
	}
}

func TestRepositoryContextIndexFileSystemSafeWritePreservesOriginalOnSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	fileSystem := NewRepositoryContextIndexFileSystem()
	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	outsideIndex := filepath.Join(outside, "context-index.json")
	if err := os.WriteFile(outsideIndex, []byte("outside"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside index) error = %v", err)
	}
	if err := os.Symlink(outsideIndex, filepath.Join(root, ".specharbor", "context-index.json")); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	err := fileSystem.WriteFileSafely(root, ".specharbor/context-index.json", "{}\n")
	if err == nil || !strings.Contains(err.Error(), "symlink target paths are not allowed") {
		t.Fatalf("WriteFileSafely(symlink) error = %v, want symlink target rejection", err)
	}
	contents, err := os.ReadFile(outsideIndex)
	if err != nil {
		t.Fatalf("ReadFile(outside index) error = %v", err)
	}
	if string(contents) != "outside" {
		t.Fatalf("outside index = %q, want unchanged", string(contents))
	}
}
