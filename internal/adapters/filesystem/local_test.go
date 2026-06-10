package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.ValidationFileSystem = (*LocalFileSystem)(nil)
var _ ports.GenerationFileSystem = (*LocalFileSystem)(nil)
var _ ports.AIAssistedGenerationFileSystem = (*LocalFileSystem)(nil)
var _ ports.ArchiveFileSystem = (*LocalFileSystem)(nil)
var _ ports.ReviewFileSystem = (*LocalFileSystem)(nil)
var _ ports.ScanFileSystem = (*LocalFileSystem)(nil)
var _ ports.ConfigFileSystem = (*LocalFileSystem)(nil)

func TestLocalFileSystemListsImmediateEntryNamesWithoutRecursion(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example"), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "Dockerfile"), []byte("FROM scratch"), 0o644); err != nil {
		t.Fatalf("WriteFile(Dockerfile) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "openspec", "project.md"), []byte("project"), 0o644); err != nil {
		t.Fatalf("WriteFile(openspec/project.md) error = %v", err)
	}

	names, err := fileSystem.ListDirectoryNames(root, ".")
	if err != nil {
		t.Fatalf("ListDirectoryNames() error = %v", err)
	}

	want := map[string]bool{"openspec": true, "go.mod": true, "Dockerfile": true}
	if len(names) != len(want) {
		t.Fatalf("ListDirectoryNames() = %v, want immediate entries %v", names, want)
	}
	for _, name := range names {
		if !want[name] {
			t.Fatalf("ListDirectoryNames() returned unexpected entry %q (names = %v)", name, names)
		}
		if strings.Contains(name, "/") || strings.Contains(name, string(filepath.Separator)) {
			t.Fatalf("ListDirectoryNames() returned a nested path %q, want immediate entry name", name)
		}
	}
}

func TestLocalFileSystemListDirectoryNamesReturnsErrorForMissingDirectory(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	_, err := fileSystem.ListDirectoryNames(root, "missing")
	if err == nil {
		t.Fatalf("ListDirectoryNames() error = nil, want missing directory error")
	}
}

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

func TestLocalFileSystemReadsReviewFiles(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()
	tasksPath := filepath.Join(root, "openspec", "changes", "change")
	if err := os.MkdirAll(tasksPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(tasksPath, "tasks.md"), []byte("- [x] Done\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	contents, err := fileSystem.ReadFile(root, "openspec/changes/change/tasks.md")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if contents != "- [x] Done\n" {
		t.Fatalf("ReadFile() = %q, want task contents", contents)
	}
}

func TestLocalFileSystemReadsValidationFilesUnderProjectRoot(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()
	changePath := filepath.Join(root, "openspec", "changes", "change")
	if err := os.MkdirAll(changePath, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(changePath, "proposal.md"), []byte("# Proposal\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	relativePath := "openspec/changes/change/proposal.md"
	fullPath := fileSystem.fullPath(root, relativePath)
	if !strings.HasPrefix(fullPath, root+string(filepath.Separator)) {
		t.Fatalf("fullPath(%q) = %q, want path under project root %q", relativePath, fullPath, root)
	}

	contents, err := fileSystem.ReadFile(root, relativePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if contents != "# Proposal\n" {
		t.Fatalf("ReadFile() = %q, want proposal contents", contents)
	}
}

func TestLocalFileSystemReportsMissingFilesDistinctlyFromReadErrors(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	exists, err := fileSystem.FileExists(root, "openspec/changes/change/proposal.md")
	if err != nil {
		t.Fatalf("FileExists() error = %v, want missing file reported without error", err)
	}
	if exists {
		t.Fatalf("FileExists() = true, want false")
	}

	_, err = fileSystem.ReadFile(root, "openspec/changes/change/proposal.md")
	if err == nil {
		t.Fatalf("ReadFile() error = nil, want read error for missing file")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("ReadFile() error = %v, want not-exist error", err)
	}
}

func TestLocalFileSystemReadsAIAssistedSourceFileLocally(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "agent-output.txt")
	if err := os.WriteFile(sourcePath, []byte("strict blocks"), 0o644); err != nil {
		t.Fatalf("WriteFile(source) error = %v", err)
	}

	contents, err := NewLocalFileSystem().ReadSourceFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadSourceFile() error = %v", err)
	}
	if contents != "strict blocks" {
		t.Fatalf("ReadSourceFile() = %q, want strict blocks", contents)
	}
}

func TestLocalFileSystemReportsMissingAIAssistedSourceFileClearly(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "missing-output.txt")

	_, err := NewLocalFileSystem().ReadSourceFile(sourcePath)
	if err == nil {
		t.Fatalf("ReadSourceFile() error = nil, want missing source error")
	}
	if !strings.Contains(err.Error(), "source file not found") || !strings.Contains(err.Error(), sourcePath) {
		t.Fatalf("ReadSourceFile() error = %q, want clear missing source path", err.Error())
	}
}

func TestLocalFileSystemWritesAIAssistedTargetFilesAndOverwritesExplicitly(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := fileSystem.CreateDirectory(root, "openspec/changes/ai-change"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}

	created, err := fileSystem.WriteFileIfAbsent(root, "openspec/changes/ai-change/proposal.md", "first")
	if err != nil {
		t.Fatalf("WriteFileIfAbsent() error = %v", err)
	}
	if !created {
		t.Fatalf("WriteFileIfAbsent() created = false, want true")
	}

	created, err = fileSystem.WriteFileIfAbsent(root, "openspec/changes/ai-change/proposal.md", "second")
	if err != nil {
		t.Fatalf("WriteFileIfAbsent(existing) error = %v", err)
	}
	if created {
		t.Fatalf("WriteFileIfAbsent(existing) created = true, want false")
	}

	if err := fileSystem.WriteFile(root, "openspec/changes/ai-change/proposal.md", "overwritten"); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(root, "openspec", "changes", "ai-change", "proposal.md"))
	if err != nil {
		t.Fatalf("ReadFile(proposal.md) error = %v", err)
	}
	if string(contents) != "overwritten" {
		t.Fatalf("proposal.md = %q, want overwritten", string(contents))
	}

	if err := fileSystem.WriteFile(root, "openspec/changes/ai-change/design.md", "new overwrite target"); err != nil {
		t.Fatalf("WriteFile(missing) error = %v", err)
	}
	contents, err = os.ReadFile(filepath.Join(root, "openspec", "changes", "ai-change", "design.md"))
	if err != nil {
		t.Fatalf("ReadFile(design.md) error = %v", err)
	}
	if string(contents) != "new overwrite target" {
		t.Fatalf("design.md = %q, want newly written content", string(contents))
	}
}

func TestLocalFileSystemRejectsAIAssistedSymlinkTargetPaths(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := fileSystem.CreateDirectory(root, "openspec/changes/ai-change"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	outsidePath := filepath.Join(t.TempDir(), "outside-proposal.md")
	if err := os.WriteFile(outsidePath, []byte("outside original"), 0o644); err != nil {
		t.Fatalf("WriteFile(outside) error = %v", err)
	}
	targetPath := filepath.Join(root, "openspec", "changes", "ai-change", "proposal.md")
	if err := os.Symlink(outsidePath, targetPath); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	relativePath := "openspec/changes/ai-change/proposal.md"
	if exists, err := fileSystem.FileExists(root, relativePath); err == nil || !strings.Contains(err.Error(), "symlink paths are not allowed") {
		t.Fatalf("FileExists() = %v, %v; want symlink rejection", exists, err)
	}
	if err := fileSystem.EnsureSafeWriteTarget(root, relativePath); err == nil || !strings.Contains(err.Error(), "symlink target paths are not allowed for generated OpenSpec files") {
		t.Fatalf("EnsureSafeWriteTarget() error = %v, want generated symlink target rejection", err)
	}
	if created, err := fileSystem.WriteFileIfAbsent(root, relativePath, "replacement"); err == nil || created {
		t.Fatalf("WriteFileIfAbsent() = %v, %v; want symlink rejection without creation", created, err)
	}
	if err := fileSystem.WriteFile(root, relativePath, "replacement"); err == nil || !strings.Contains(err.Error(), "symlink target paths are not allowed") {
		t.Fatalf("WriteFile() error = %v, want symlink rejection", err)
	}

	contents, err := os.ReadFile(outsidePath)
	if err != nil {
		t.Fatalf("ReadFile(outside) error = %v", err)
	}
	if string(contents) != "outside original" {
		t.Fatalf("outside target = %q, want unchanged", string(contents))
	}
}

func TestLocalFileSystemRejectsAIAssistedSymlinkParentDirectories(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("MkdirAll(changes) error = %v", err)
	}
	outsideDirectory := t.TempDir()
	changeLink := filepath.Join(root, "openspec", "changes", "ai-change")
	if err := os.Symlink(outsideDirectory, changeLink); err != nil {
		t.Skipf("Symlink() unavailable: %v", err)
	}

	relativePath := "openspec/changes/ai-change/proposal.md"
	if err := fileSystem.EnsureSafeWriteTarget(root, relativePath); err == nil || !strings.Contains(err.Error(), "symlink parent directories are not allowed for generated OpenSpec files") {
		t.Fatalf("EnsureSafeWriteTarget() error = %v, want symlink parent rejection", err)
	}
	if err := fileSystem.WriteFile(root, relativePath, "replacement"); err == nil || !strings.Contains(err.Error(), "symlink parent directories are not allowed") {
		t.Fatalf("WriteFile() error = %v, want symlink parent rejection", err)
	}
	if _, err := os.Stat(filepath.Join(outsideDirectory, "proposal.md")); !os.IsNotExist(err) {
		t.Fatalf("outside proposal stat error = %v, want no file created through symlink parent", err)
	}
}

func TestLocalFileSystemDistinguishesMissingAIAssistedTargetsFromUnsafeSymlinks(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := fileSystem.CreateDirectory(root, "openspec/changes/ai-change"); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	missingPath := "openspec/changes/ai-change/proposal.md"
	if err := fileSystem.EnsureSafeWriteTarget(root, missingPath); err != nil {
		t.Fatalf("EnsureSafeWriteTarget(missing) error = %v, want nil", err)
	}
	exists, err := fileSystem.FileExists(root, missingPath)
	if err != nil {
		t.Fatalf("FileExists(missing) error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists(missing) = true, want false")
	}
}

func TestLocalFileSystemRejectsUnsafeAIAssistedTargetPaths(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	tests := []struct {
		name         string
		relativePath string
	}{
		{name: "path traversal", relativePath: "../outside.md"},
		{name: "nested traversal", relativePath: "openspec/../outside.md"},
		{name: "absolute unix", relativePath: "/tmp/outside.md"},
		{name: "absolute windows", relativePath: `C:\tmp\outside.md`},
		{name: "backslash traversal", relativePath: `openspec\..\outside.md`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := fileSystem.PathExists(root, test.relativePath); err == nil {
				t.Fatalf("PathExists(%q) error = nil, want unsafe path rejection", test.relativePath)
			}
			if err := fileSystem.EnsureSafeWriteTarget(root, test.relativePath); err == nil {
				t.Fatalf("EnsureSafeWriteTarget(%q) error = nil, want unsafe path rejection", test.relativePath)
			}
			if _, err := fileSystem.WriteFileIfAbsent(root, test.relativePath, "content"); err == nil {
				t.Fatalf("WriteFileIfAbsent(%q) error = nil, want unsafe path rejection", test.relativePath)
			}
			if err := fileSystem.WriteFile(root, test.relativePath, "content"); err == nil {
				t.Fatalf("WriteFile(%q) error = nil, want unsafe path rejection", test.relativePath)
			}
		})
	}
}

func TestLocalFileSystemReportsAIAssistedWriteErrors(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	_, err := fileSystem.WriteFileIfAbsent(root, "openspec/changes/missing/proposal.md", "content")
	if err == nil {
		t.Fatalf("WriteFileIfAbsent() error = nil, want missing directory write error")
	}
	if !os.IsNotExist(err) {
		t.Fatalf("WriteFileIfAbsent() error = %v, want not-exist write error", err)
	}
}

func TestLocalFileSystemReadsLocalConfigFile(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := os.MkdirAll(filepath.Join(root, ".specharbor"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.specharbor) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".specharbor", "config.yml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(config.yml) error = %v", err)
	}

	exists, err := fileSystem.FileExists(root, ".specharbor/config.yml")
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("FileExists() = false, want true")
	}

	contents, err := fileSystem.ReadFile(root, ".specharbor/config.yml")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if contents != "version: 1\n" {
		t.Fatalf("ReadFile() = %q, want version config", contents)
	}
}

func TestLocalFileSystemDoesNotTreatConfigDirectoryAsFile(t *testing.T) {
	root := t.TempDir()
	fileSystem := NewLocalFileSystem()

	if err := os.MkdirAll(filepath.Join(root, ".specharbor", "config.yml"), 0o755); err != nil {
		t.Fatalf("MkdirAll(config.yml directory) error = %v", err)
	}

	exists, err := fileSystem.FileExists(root, ".specharbor/config.yml")
	if err != nil {
		t.Fatalf("FileExists() error = %v", err)
	}
	if exists {
		t.Fatalf("FileExists() for config directory = true, want false")
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
