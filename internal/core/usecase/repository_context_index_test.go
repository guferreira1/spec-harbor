package usecase

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

func TestRepositoryContextIndexBuildsSupportedMetadataOnlyInventory(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\nsecret prose must not be stored\n"
	fileSystem.files["go.mod"] = "module example.com/project\n"
	fileSystem.files["openspec/project.md"] = "# Project\n"
	fileSystem.directories["docs"] = []ports.RepositoryContextIndexDirectoryEntry{
		{Name: "usage.md", IsRegular: true},
		{Name: "secrets.md", IsRegular: true},
		{Name: "node_modules", IsDirectory: true},
		{Name: "linked.md", IsSymlink: true},
	}
	fileSystem.files["docs/usage.md"] = "# Usage\n"
	fileSystem.files["docs/secrets.md"] = "# Secret\n"
	fileSystem.directories["docs/node_modules"] = []ports.RepositoryContextIndexDirectoryEntry{{Name: "README.md", IsRegular: true}}
	fileSystem.files["docs/node_modules/README.md"] = "# Generated\n"

	result, err := NewBuildRepositoryContextIndex(fileSystem).Execute(RepositoryContextIndexInput{
		ProjectRoot: "/project",
		Mode:        domain.RepositoryContextIndexModeReport,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.RepositoryContextIndexStatusBuilt {
		t.Fatalf("Status = %q, want built", result.Status)
	}
	assertIndexEntry(t, result.Index, "README.md", domain.ContextSourceCategoryReadme, domain.RepositoryContextIndexFileTypeMarkdown)
	assertIndexEntry(t, result.Index, "go.mod", domain.ContextSourceCategoryPackageManifest, domain.RepositoryContextIndexFileTypeGoModule)
	assertIndexEntry(t, result.Index, "docs/usage.md", domain.ContextSourceCategoryDocumentation, domain.RepositoryContextIndexFileTypeMarkdown)
	assertNoIndexEntry(t, result.Index, "docs/secrets.md")
	assertNoIndexEntry(t, result.Index, "docs/node_modules/README.md")
	assertNoIndexEntry(t, result.Index, "docs/linked.md")
	if strings.Contains(indexTextForTest(result.Index), "secret prose") || strings.Contains(indexTextForTest(result.Index), "module example.com/project") {
		t.Fatalf("index stored raw file contents: %+v", result.Index)
	}
	assertSkipReason(t, result.Index, "docs/secrets.md", domain.RepositoryContextIndexSkipSensitiveFile)
	assertSkipReason(t, result.Index, "docs/node_modules", domain.RepositoryContextIndexSkipGeneratedDirectory)
	assertSkipReason(t, result.Index, "docs/linked.md", domain.RepositoryContextIndexSkipSymlink)
}

func TestRepositoryContextIndexEnforcesLimits(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = strings.Repeat("x", 20)
	fileSystem.files["AGENTS.md"] = "agents"

	result, err := NewBuildRepositoryContextIndexWithLimits(fileSystem, domain.RepositoryContextIndexLimits{
		MaxIndexedFiles:   10,
		MaxFileSizeBytes:  8,
		MaxTotalFileBytes: 64,
		MaxSkippedRecords: 10,
		MaxDirectoryDepth: 1,
	}).Execute(RepositoryContextIndexInput{ProjectRoot: "/project"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Index.Truncated {
		t.Fatalf("Truncated = false, want true")
	}
	assertSkipReason(t, result.Index, "README.md", domain.RepositoryContextIndexSkipFileTooLarge)
}

func TestRepositoryContextIndexWritePersistsStableJSON(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\n"

	result, err := NewBuildRepositoryContextIndex(fileSystem).Execute(RepositoryContextIndexInput{
		ProjectRoot: "/project",
		Mode:        domain.RepositoryContextIndexModeWrite,
	})
	if err != nil {
		t.Fatalf("Execute(write) error = %v", err)
	}
	if result.Status != domain.RepositoryContextIndexStatusWritten {
		t.Fatalf("Status = %q, want written", result.Status)
	}
	written := fileSystem.files[domain.RepositoryContextIndexPath]
	for _, want := range []string{`"schema_version": 1`, `"mode": "deterministic"`, `"path": "README.md"`} {
		if !strings.Contains(written, want) {
			t.Fatalf("written index = %q, want %q", written, want)
		}
	}
	if strings.Contains(written, "# Project") {
		t.Fatalf("written index stored raw file contents: %s", written)
	}
}

func TestRepositoryContextIndexCheckReportsCurrentMissingInvalidAndStale(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\n"
	useCase := NewBuildRepositoryContextIndex(fileSystem)

	missing, err := useCase.Execute(RepositoryContextIndexInput{ProjectRoot: "/project", Mode: domain.RepositoryContextIndexModeCheck})
	if err != nil {
		t.Fatalf("missing check error = %v", err)
	}
	if missing.Status != domain.RepositoryContextIndexStatusMissing {
		t.Fatalf("missing Status = %q, want missing", missing.Status)
	}

	if _, err := useCase.Execute(RepositoryContextIndexInput{ProjectRoot: "/project", Mode: domain.RepositoryContextIndexModeWrite}); err != nil {
		t.Fatalf("write error = %v", err)
	}
	current, err := useCase.Execute(RepositoryContextIndexInput{ProjectRoot: "/project", Mode: domain.RepositoryContextIndexModeCheck})
	if err != nil {
		t.Fatalf("current check error = %v", err)
	}
	if current.Status != domain.RepositoryContextIndexStatusCurrent {
		t.Fatalf("current Status = %q, want current", current.Status)
	}

	fileSystem.files["README.md"] = "# Project changed\n"
	stale, err := useCase.Execute(RepositoryContextIndexInput{ProjectRoot: "/project", Mode: domain.RepositoryContextIndexModeCheck})
	if err != nil {
		t.Fatalf("stale check error = %v", err)
	}
	if stale.Status != domain.RepositoryContextIndexStatusStale {
		t.Fatalf("stale Status = %q, want stale", stale.Status)
	}
	if len(stale.StaleReasons) == 0 {
		t.Fatalf("StaleReasons = none, want changed hash/size reasons")
	}

	fileSystem.files[domain.RepositoryContextIndexPath] = "{not json"
	invalid, err := useCase.Execute(RepositoryContextIndexInput{ProjectRoot: "/project", Mode: domain.RepositoryContextIndexModeCheck})
	if err != nil {
		t.Fatalf("invalid check error = %v", err)
	}
	if invalid.Status != domain.RepositoryContextIndexStatusInvalid {
		t.Fatalf("invalid Status = %q, want invalid", invalid.Status)
	}
}

func TestRepositoryContextIndexWriteFailurePreservesOriginal(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\n"
	fileSystem.files[domain.RepositoryContextIndexPath] = "original"
	fileSystem.writeErrors[domain.RepositoryContextIndexPath] = errors.New("disk full")

	_, err := NewBuildRepositoryContextIndex(fileSystem).Execute(RepositoryContextIndexInput{
		ProjectRoot: "/project",
		Mode:        domain.RepositoryContextIndexModeWrite,
	})
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("Execute(write) error = %v, want disk full", err)
	}
	if fileSystem.files[domain.RepositoryContextIndexPath] != "original" {
		t.Fatalf("index changed after write failure: %q", fileSystem.files[domain.RepositoryContextIndexPath])
	}
}

func TestRepositoryContextIndexEmptyRepositoryIsSafe(t *testing.T) {
	result, err := NewBuildRepositoryContextIndex(newFakeRepositoryContextIndexFileSystem()).Execute(RepositoryContextIndexInput{
		ProjectRoot: "/project",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(result.Index.Entries) != 0 {
		t.Fatalf("Entries = %+v, want empty", result.Index.Entries)
	}
	if result.Index.Project.RootMarker != "none" {
		t.Fatalf("RootMarker = %q, want none", result.Index.Project.RootMarker)
	}
}

type fakeRepositoryContextIndexFileSystem struct {
	files       map[string]string
	directories map[string][]ports.RepositoryContextIndexDirectoryEntry
	mtimes      map[string]time.Time
	writeErrors map[string]error
}

func newFakeRepositoryContextIndexFileSystem() *fakeRepositoryContextIndexFileSystem {
	return &fakeRepositoryContextIndexFileSystem{
		files:       make(map[string]string),
		directories: make(map[string][]ports.RepositoryContextIndexDirectoryEntry),
		mtimes:      make(map[string]time.Time),
		writeErrors: make(map[string]error),
	}
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) FileExists(_ string, relativePath string) (bool, error) {
	_, exists := fileSystem.files[relativePath]
	return exists, nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) DirectoryExists(_ string, relativePath string) (bool, error) {
	_, exists := fileSystem.directories[relativePath]
	return exists, nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) ListDirectory(_ string, relativePath string) ([]ports.RepositoryContextIndexDirectoryEntry, error) {
	return append([]ports.RepositoryContextIndexDirectoryEntry(nil), fileSystem.directories[relativePath]...), nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) FileMetadata(_ string, relativePath string) (ports.RepositoryContextIndexFileMetadata, error) {
	contents, exists := fileSystem.files[relativePath]
	if !exists {
		return ports.RepositoryContextIndexFileMetadata{}, errors.New("missing file")
	}
	mtime := fileSystem.mtimes[relativePath]
	if mtime.IsZero() {
		mtime = time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	}
	return ports.RepositoryContextIndexFileMetadata{SizeBytes: int64(len(contents)), ModifiedTime: mtime}, nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) ReadFileBytes(_ string, relativePath string, maxBytes int64) ([]byte, error) {
	contents, exists := fileSystem.files[relativePath]
	if !exists {
		return nil, errors.New("missing file")
	}
	if int64(len(contents)) > maxBytes {
		return nil, errors.New("file exceeds limit")
	}
	return []byte(contents), nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) CreateDirectory(_ string, relativePath string) error {
	fileSystem.directories[relativePath] = fileSystem.directories[relativePath]
	return nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) ReadFileSafely(_ string, relativePath string) (string, error) {
	contents, exists := fileSystem.files[relativePath]
	if !exists {
		return "", errors.New("missing file")
	}
	return contents, nil
}

func (fileSystem *fakeRepositoryContextIndexFileSystem) WriteFileSafely(_ string, relativePath string, contents string) error {
	if err := fileSystem.writeErrors[relativePath]; err != nil {
		return err
	}
	fileSystem.files[relativePath] = contents
	fileSystem.mtimes[relativePath] = time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
	return nil
}

func assertIndexEntry(
	t *testing.T,
	index domain.RepositoryContextIndex,
	relativePath string,
	category domain.ContextSourceCategory,
	fileType domain.RepositoryContextIndexFileType,
) {
	t.Helper()
	for _, entry := range index.Entries {
		if entry.Path == relativePath {
			if entry.SourceCategory != category || entry.FileType != fileType {
				t.Fatalf("entry %+v, want category %q file type %q", entry, category, fileType)
			}
			return
		}
	}
	t.Fatalf("missing index entry %q in %+v", relativePath, index.Entries)
}

func assertNoIndexEntry(t *testing.T, index domain.RepositoryContextIndex, relativePath string) {
	t.Helper()
	for _, entry := range index.Entries {
		if entry.Path == relativePath {
			t.Fatalf("unexpected index entry %+v", entry)
		}
	}
}

func assertSkipReason(
	t *testing.T,
	index domain.RepositoryContextIndex,
	relativePath string,
	reason domain.RepositoryContextIndexSkipReason,
) {
	t.Helper()
	for _, skipped := range index.Skipped {
		if skipped.Path == relativePath && skipped.Reason == reason {
			return
		}
	}
	t.Fatalf("missing skip %q %q in %+v", relativePath, reason, index.Skipped)
}

func indexTextForTest(index domain.RepositoryContextIndex) string {
	var builder strings.Builder
	for _, entry := range index.Entries {
		builder.WriteString(entry.Path)
		builder.WriteString(entry.ContentHash)
		builder.WriteString(entry.SourceEvidenceCategory)
	}
	return builder.String()
}
