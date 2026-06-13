package usecase

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

func TestRetrieveLocalContextReturnsBoundedAttributedSnippets(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\n\nSpecHarbor uses Hexagonal Architecture for boundaries.\nDo not dump this unrelated line.\n"
	fileSystem.files["openspec/project.md"] = "# Project\n\nArchitecture: Hexagonal Architecture\n"
	writeTestRepositoryContextIndex(t, fileSystem)

	result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
		ProjectRoot: "/project",
		Query:       "hexagonal architecture",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.LocalContextRetrievalStatusCurrent {
		t.Fatalf("Status = %q, want current", result.Status)
	}
	if len(result.Results) == 0 {
		t.Fatalf("Results = none, want matches")
	}
	first := result.Results[0]
	if first.Rank != 1 {
		t.Fatalf("Rank = %d, want 1", first.Rank)
	}
	if first.Path == "" || first.SourceCategory == "" || first.Score <= 0 {
		t.Fatalf("result missing attribution or score: %+v", first)
	}
	if first.Snippet.LineStart <= 0 || first.Snippet.LineEnd < first.Snippet.LineStart {
		t.Fatalf("result missing line range: %+v", first.Snippet)
	}
	if !strings.Contains(strings.ToLower(first.Snippet.Text), "hexagonal architecture") {
		t.Fatalf("Snippet = %q, want matching text", first.Snippet.Text)
	}
}

func TestRetrieveLocalContextReportsNoResultsSafely(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\nNo matching words here.\n"
	writeTestRepositoryContextIndex(t, fileSystem)

	result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
		ProjectRoot: "/project",
		Query:       "nonexistentterm",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.LocalContextRetrievalStatusNoResults {
		t.Fatalf("Status = %q, want no_results", result.Status)
	}
	if len(result.Results) != 0 {
		t.Fatalf("Results = %+v, want none", result.Results)
	}
}

func TestRetrieveLocalContextFailsSafelyForIndexDependencyStates(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		fileSystem := newFakeRepositoryContextIndexFileSystem()
		result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
			ProjectRoot: "/project",
			Query:       "architecture",
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != domain.LocalContextRetrievalStatusMissingIndex {
			t.Fatalf("Status = %q, want missing_index", result.Status)
		}
		if !strings.Contains(result.Message, "specharbor context index --write") {
			t.Fatalf("Message = %q, want actionable index command", result.Message)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		fileSystem := newFakeRepositoryContextIndexFileSystem()
		fileSystem.files[domain.RepositoryContextIndexPath] = "{not json"
		result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
			ProjectRoot: "/project",
			Query:       "architecture",
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != domain.LocalContextRetrievalStatusInvalidIndex {
			t.Fatalf("Status = %q, want invalid_index", result.Status)
		}
	})

	t.Run("stale changed source", func(t *testing.T) {
		fileSystem := newFakeRepositoryContextIndexFileSystem()
		fileSystem.files["README.md"] = "# Project\nArchitecture\n"
		writeTestRepositoryContextIndex(t, fileSystem)
		fileSystem.files["README.md"] = "# Project\nArchitecture changed\n"

		result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
			ProjectRoot: "/project",
			Query:       "architecture",
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != domain.LocalContextRetrievalStatusStaleIndex {
			t.Fatalf("Status = %q, want stale_index", result.Status)
		}
		if len(result.StaleReasons) == 0 {
			t.Fatalf("StaleReasons = none, want stale details")
		}
	})

	t.Run("stale missing source", func(t *testing.T) {
		fileSystem := newFakeRepositoryContextIndexFileSystem()
		fileSystem.files["README.md"] = "# Project\nArchitecture\n"
		writeTestRepositoryContextIndex(t, fileSystem)
		delete(fileSystem.files, "README.md")

		result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
			ProjectRoot: "/project",
			Query:       "architecture",
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != domain.LocalContextRetrievalStatusStaleIndex {
			t.Fatalf("Status = %q, want stale_index", result.Status)
		}
	})

	t.Run("truncated", func(t *testing.T) {
		fileSystem := newFakeRepositoryContextIndexFileSystem()
		entry := mustRetrievalRepositoryContextIndexEntry(t, "README.md", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		index, err := domain.NewRepositoryContextIndex("none", domain.DefaultRepositoryContextIndexLimits(), []domain.RepositoryContextIndexEntry{entry}, nil, true)
		if err != nil {
			t.Fatalf("NewRepositoryContextIndex() error = %v", err)
		}
		contents, err := encodeRepositoryContextIndex(index)
		if err != nil {
			t.Fatalf("encodeRepositoryContextIndex() error = %v", err)
		}
		fileSystem.files[domain.RepositoryContextIndexPath] = contents
		fileSystem.files["README.md"] = "# Project\nArchitecture\n"

		result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
			ProjectRoot: "/project",
			Query:       "architecture",
		})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if result.Status != domain.LocalContextRetrievalStatusTruncatedIndex {
			t.Fatalf("Status = %q, want truncated_index", result.Status)
		}
	})
}

func TestRetrieveLocalContextSkipsSensitiveSymlinkUnsafeAndOversizedSources(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = "# Project\nArchitecture\n"
	fileSystem.files["docs/secrets.md"] = "secret token architecture"
	fileSystem.directories["docs"] = []ports.RepositoryContextIndexDirectoryEntry{
		{Name: "secrets.md", IsRegular: true},
		{Name: "linked.md", IsSymlink: true},
	}
	writeTestRepositoryContextIndex(t, fileSystem)

	result, err := NewRetrieveLocalContext(fileSystem).Execute(RetrieveLocalContextInput{
		ProjectRoot: "/project",
		Query:       "secret architecture",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, retrieved := range result.Results {
		if strings.Contains(retrieved.Path, "secrets") || strings.Contains(retrieved.Snippet.Text, "secret token") {
			t.Fatalf("retrieved sensitive source: %+v", retrieved)
		}
		if strings.Contains(retrieved.Path, "linked") {
			t.Fatalf("retrieved symlink source: %+v", retrieved)
		}
	}

	limits := domain.DefaultLocalContextRetrievalLimits()
	limits.MaxSourceReadBytes = 8
	limited, err := NewRetrieveLocalContextWithLimits(fileSystem, limits).Execute(RetrieveLocalContextInput{
		ProjectRoot: "/project",
		Query:       "architecture",
	})
	if err != nil {
		t.Fatalf("limited Execute() error = %v", err)
	}
	if limited.Status != domain.LocalContextRetrievalStatusNoResults {
		t.Fatalf("limited Status = %q, want no_results", limited.Status)
	}
}

func TestRetrieveLocalContextEnforcesResultSnippetAndOutputLimits(t *testing.T) {
	fileSystem := newFakeRepositoryContextIndexFileSystem()
	fileSystem.files["README.md"] = strings.Join([]string{
		"# Project",
		"architecture first",
		"architecture second",
		"architecture third",
		"architecture fourth",
	}, "\n")
	fileSystem.files["docs/a.md"] = "architecture docs"
	fileSystem.files["docs/b.md"] = "architecture more docs"
	fileSystem.directories["docs"] = []ports.RepositoryContextIndexDirectoryEntry{
		{Name: "a.md", IsRegular: true},
		{Name: "b.md", IsRegular: true},
	}
	writeTestRepositoryContextIndex(t, fileSystem)

	limits := domain.DefaultLocalContextRetrievalLimits()
	limits.MaxResults = 1
	limits.MaxSnippetsPerFile = 1
	limits.MaxRenderedContentChars = 1000
	result, err := NewRetrieveLocalContextWithLimits(fileSystem, limits).Execute(RetrieveLocalContextInput{
		ProjectRoot: "/project",
		Query:       "architecture",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(result.Results) != 1 {
		t.Fatalf("Results = %d, want 1", len(result.Results))
	}
	if !result.OutputTruncated {
		t.Fatalf("OutputTruncated = false, want true")
	}
	if strings.Contains(result.Results[0].Snippet.Text, "fourth") {
		t.Fatalf("snippet appears to dump too much file content: %q", result.Results[0].Snippet.Text)
	}
}

func writeTestRepositoryContextIndex(t *testing.T, fileSystem *fakeRepositoryContextIndexFileSystem) {
	t.Helper()
	_, err := NewBuildRepositoryContextIndex(fileSystem).Execute(RepositoryContextIndexInput{
		ProjectRoot: "/project",
		Mode:        domain.RepositoryContextIndexModeWrite,
	})
	if err != nil {
		t.Fatalf("write index error = %v", err)
	}
}

func mustRetrievalRepositoryContextIndexEntry(
	t *testing.T,
	relativePath string,
	hash string,
) domain.RepositoryContextIndexEntry {
	t.Helper()
	entry, err := domain.NewRepositoryContextIndexEntry(domain.RepositoryContextIndexEntryInput{
		Path:                   relativePath,
		SourceCategory:         domain.ContextSourceCategoryReadme,
		FileType:               domain.RepositoryContextIndexFileTypeMarkdown,
		LanguageOrEcosystem:    "Markdown",
		SizeBytes:              10,
		ContentHash:            hash,
		ModifiedTime:           "2026-06-13T12:00:00Z",
		SupportedForRetrieval:  true,
		ClassificationHints:    []domain.RepositoryContextIndexClassificationHint{domain.RepositoryContextIndexHintInventoryMetadata},
		SourceEvidenceCategory: "readme",
	})
	if err != nil {
		t.Fatalf("NewRepositoryContextIndexEntry() error = %v", err)
	}
	return entry
}
