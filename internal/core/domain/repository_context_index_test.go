package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestRepositoryContextIndexEntryValidatesMetadataOnlyFields(t *testing.T) {
	entry, err := NewRepositoryContextIndexEntry(RepositoryContextIndexEntryInput{
		Path:                   "README.md",
		SourceCategory:         ContextSourceCategoryReadme,
		FileType:               RepositoryContextIndexFileTypeMarkdown,
		LanguageOrEcosystem:    "Markdown",
		SizeBytes:              14,
		ContentHash:            "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ModifiedTime:           "2026-06-13T12:00:00Z",
		SupportedForRetrieval:  true,
		ClassificationHints:    []RepositoryContextIndexClassificationHint{RepositoryContextIndexHintInventoryMetadata, RepositoryContextIndexHintDetectedFact},
		SourceEvidenceCategory: "readme",
	})
	if err != nil {
		t.Fatalf("NewRepositoryContextIndexEntry() error = %v", err)
	}
	if entry.Path != "README.md" {
		t.Fatalf("Path = %q, want README.md", entry.Path)
	}
	if strings.Contains(entry.ContentHash, "project secret") {
		t.Fatalf("entry stored raw content in hash field: %+v", entry)
	}

	tests := []struct {
		name  string
		input RepositoryContextIndexEntryInput
		want  string
	}{
		{
			name: "unsafe path",
			input: RepositoryContextIndexEntryInput{
				Path:                   "../README.md",
				SourceCategory:         ContextSourceCategoryReadme,
				FileType:               RepositoryContextIndexFileTypeMarkdown,
				SizeBytes:              1,
				ContentHash:            "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				ModifiedTime:           "2026-06-13T12:00:00Z",
				SourceEvidenceCategory: "readme",
			},
			want: "path must not contain path traversal",
		},
		{
			name: "bad hash",
			input: RepositoryContextIndexEntryInput{
				Path:                   "README.md",
				SourceCategory:         ContextSourceCategoryReadme,
				FileType:               RepositoryContextIndexFileTypeMarkdown,
				SizeBytes:              1,
				ContentHash:            "project secret content",
				ModifiedTime:           "2026-06-13T12:00:00Z",
				SourceEvidenceCategory: "readme",
			},
			want: "content hash must be sha256",
		},
		{
			name: "unsupported hint",
			input: RepositoryContextIndexEntryInput{
				Path:                   "README.md",
				SourceCategory:         ContextSourceCategoryReadme,
				FileType:               RepositoryContextIndexFileTypeMarkdown,
				SizeBytes:              1,
				ContentHash:            "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
				ModifiedTime:           "2026-06-13T12:00:00Z",
				ClassificationHints:    []RepositoryContextIndexClassificationHint{"raw_content"},
				SourceEvidenceCategory: "readme",
			},
			want: "unsupported repository context index classification hint",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewRepositoryContextIndexEntry(test.input)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewRepositoryContextIndexEntry() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRepositoryContextIndexOrdersEntriesAndSkippedRecords(t *testing.T) {
	readme := mustRepositoryContextIndexEntry(t, "README.md", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	agents := mustRepositoryContextIndexEntry(t, "AGENTS.md", "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	skipB := mustRepositoryContextIndexSkipped(t, "docs/node_modules", RepositoryContextIndexSkipGeneratedDirectory)
	skipA := mustRepositoryContextIndexSkipped(t, "docs/secrets.md", RepositoryContextIndexSkipSensitiveFile)

	index, err := NewRepositoryContextIndex("openspec/project.md", DefaultRepositoryContextIndexLimits(), []RepositoryContextIndexEntry{readme, agents}, []RepositoryContextIndexSkipped{skipB, skipA}, false)
	if err != nil {
		t.Fatalf("NewRepositoryContextIndex() error = %v", err)
	}

	if got := []string{index.Entries[0].Path, index.Entries[1].Path}; !reflect.DeepEqual(got, []string{"AGENTS.md", "README.md"}) {
		t.Fatalf("entry order = %v, want deterministic path order", got)
	}
	if got := []string{index.Skipped[0].Path, index.Skipped[1].Path}; !reflect.DeepEqual(got, []string{"docs/node_modules", "docs/secrets.md"}) {
		t.Fatalf("skip order = %v, want deterministic path order", got)
	}
}

func TestRepositoryContextIndexSkipPolicyMatchesSensitiveAndGeneratedPaths(t *testing.T) {
	for _, relativePath := range []string{
		".env",
		".env.local",
		"docs/private.pem",
		"docs/private.key",
		"id_rsa",
		"id_ed25519",
		"docs/secrets.local",
		"docs/credentials.json",
	} {
		t.Run(relativePath, func(t *testing.T) {
			reason, skip := ShouldSkipRepositoryContextIndexPath(relativePath)
			if !skip || reason != RepositoryContextIndexSkipSensitiveFile {
				t.Fatalf("ShouldSkipRepositoryContextIndexPath(%q) = %q, %t; want sensitive skip", relativePath, reason, skip)
			}
		})
	}

	for _, relativePath := range []string{
		".git/config",
		"docs/node_modules/README.md",
		"docs/dist/output.md",
		"docs/.cache/file.md",
		"docs/bin/tool.md",
	} {
		t.Run(relativePath, func(t *testing.T) {
			reason, skip := ShouldSkipRepositoryContextIndexPath(relativePath)
			if !skip || reason != RepositoryContextIndexSkipGeneratedDirectory {
				t.Fatalf("ShouldSkipRepositoryContextIndexPath(%q) = %q, %t; want generated skip", relativePath, reason, skip)
			}
		})
	}
}

func TestCompareRepositoryContextIndexesReportsStaleReasons(t *testing.T) {
	storedEntry := mustRepositoryContextIndexEntry(t, "README.md", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	currentEntry := storedEntry
	currentEntry.SizeBytes = 20
	currentEntry.ContentHash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	currentEntry.ModifiedTime = "2026-06-13T12:01:00Z"
	currentNew := mustRepositoryContextIndexEntry(t, "go.mod", "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")

	stored, err := NewRepositoryContextIndex("openspec/project.md", DefaultRepositoryContextIndexLimits(), []RepositoryContextIndexEntry{storedEntry}, nil, false)
	if err != nil {
		t.Fatalf("stored index error = %v", err)
	}
	current, err := NewRepositoryContextIndex("openspec/project.md", DefaultRepositoryContextIndexLimits(), []RepositoryContextIndexEntry{currentEntry, currentNew}, nil, true)
	if err != nil {
		t.Fatalf("current index error = %v", err)
	}

	reasons := CompareRepositoryContextIndexes(stored, current)
	for _, want := range []string{"content_hash_changed", "file_size_changed", "modified_time_changed", "entry_added", "truncation_changed"} {
		if !hasRepositoryContextIndexStaleReason(reasons, want) {
			t.Fatalf("stale reasons = %+v, want %s", reasons, want)
		}
	}
}

func TestSafeRepositoryContextIndexPathRejectsUnsafeInputs(t *testing.T) {
	for _, relativePath := range []string{"../README.md", "/tmp/README.md", `C:\tmp\README.md`, "docs/\x00secret.md"} {
		t.Run(relativePath, func(t *testing.T) {
			_, err := SafeRepositoryContextIndexPath(relativePath)
			if err == nil {
				t.Fatalf("SafeRepositoryContextIndexPath(%q) error = nil, want unsafe path error", relativePath)
			}
		})
	}
}

func mustRepositoryContextIndexEntry(t *testing.T, relativePath string, hash string) RepositoryContextIndexEntry {
	t.Helper()
	entry, err := NewRepositoryContextIndexEntry(RepositoryContextIndexEntryInput{
		Path:                   relativePath,
		SourceCategory:         ContextSourceCategoryReadme,
		FileType:               RepositoryContextIndexFileTypeMarkdown,
		LanguageOrEcosystem:    "Markdown",
		SizeBytes:              10,
		ContentHash:            hash,
		ModifiedTime:           "2026-06-13T12:00:00Z",
		SupportedForRetrieval:  true,
		ClassificationHints:    []RepositoryContextIndexClassificationHint{RepositoryContextIndexHintInventoryMetadata},
		SourceEvidenceCategory: "readme",
	})
	if err != nil {
		t.Fatalf("NewRepositoryContextIndexEntry() error = %v", err)
	}
	return entry
}

func mustRepositoryContextIndexSkipped(
	t *testing.T,
	relativePath string,
	reason RepositoryContextIndexSkipReason,
) RepositoryContextIndexSkipped {
	t.Helper()
	skipped, err := NewRepositoryContextIndexSkipped(relativePath, reason)
	if err != nil {
		t.Fatalf("NewRepositoryContextIndexSkipped() error = %v", err)
	}
	return skipped
}

func hasRepositoryContextIndexStaleReason(reasons []RepositoryContextIndexStaleReason, code string) bool {
	for _, reason := range reasons {
		if reason.Code == code {
			return true
		}
	}
	return false
}
