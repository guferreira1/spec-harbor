package usecase

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestRetrieveGitHubRemoteContextReturnsSourceAttributedResults(t *testing.T) {
	reader := newFakeGitHubRemoteReader()
	reader.files["README.md"] = "# Architecture\n\nSpecHarbor follows Hexagonal Architecture.\n"

	report, err := NewRetrieveGitHubRemoteContext(reader).Execute(RetrieveGitHubRemoteContextInput{
		Repository: "guferreira1/spec-harbor",
		Query:      "hexagonal architecture",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if report.Status != domain.GitHubRemoteContextStatusCurrent {
		t.Fatalf("Status = %q, want current", report.Status)
	}
	if report.Repository != "guferreira1/spec-harbor" {
		t.Fatalf("Repository = %q", report.Repository)
	}
	if report.DefaultBranch != "main" || report.ResolvedRef != "main" || report.CommitSHA != "abc123" {
		t.Fatalf("ref metadata = default %q resolved %q sha %q", report.DefaultBranch, report.ResolvedRef, report.CommitSHA)
	}
	if len(report.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(report.Results))
	}
	result := report.Results[0]
	if result.Rank != 1 || result.Path != "README.md" || !result.Remote {
		t.Fatalf("result = %#v", result)
	}
	if result.SourceCategory != domain.ContextSourceCategoryReadme || result.SourceEvidenceCategory != "readme" {
		t.Fatalf("source attribution = %s/%s", result.SourceCategory, result.SourceEvidenceCategory)
	}
	if result.Snippet.LineStart == 0 || !strings.Contains(strings.ToLower(result.Snippet.Text), "hexagonal architecture") {
		t.Fatalf("snippet = %#v", result.Snippet)
	}
}

func TestRetrieveGitHubRemoteContextSupportsURLRepoRefAndPathFilter(t *testing.T) {
	reader := newFakeGitHubRemoteReader()
	reader.directories["docs"] = []domain.GitHubRemoteEntry{
		{Path: "docs/architecture.md", Type: "file", SizeBytes: 40},
		{Path: "docs/usage.md", Type: "file", SizeBytes: 30},
	}
	reader.files["docs/architecture.md"] = "# Architecture\nRemote architecture context\n"
	reader.files["docs/usage.md"] = "# Usage\nNo matching term\n"

	report, err := NewRetrieveGitHubRemoteContext(reader).Execute(RetrieveGitHubRemoteContextInput{
		Repository:  "https://github.com/guferreira1/spec-harbor",
		Ref:         "feature/context",
		Query:       "remote architecture",
		PathFilters: []string{"docs/architecture.md"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if report.RequestedRef != "feature/context" || report.ResolvedRef != "feature/context" {
		t.Fatalf("ref = requested %q resolved %q", report.RequestedRef, report.ResolvedRef)
	}
	if len(report.Results) != 1 || report.Results[0].Path != "docs/architecture.md" {
		t.Fatalf("results = %#v", report.Results)
	}
	if len(reader.reads) != 1 || reader.reads[0] != "docs/architecture.md" {
		t.Fatalf("reads = %#v, want only docs/architecture.md", reader.reads)
	}
}

func TestRetrieveGitHubRemoteContextSkipsSensitiveGeneratedOversizedAndUnsupported(t *testing.T) {
	reader := newFakeGitHubRemoteReader()
	reader.directories["docs"] = []domain.GitHubRemoteEntry{
		{Path: "docs/secrets.md", Type: "file", SizeBytes: 20},
		{Path: "docs/private.pem", Type: "file", SizeBytes: 20},
		{Path: "docs/architecture.txt", Type: "file", SizeBytes: 20},
		{Path: "docs/large.md", Type: "file", SizeBytes: 200 * 1024},
		{Path: "docs/current.md", Type: "file", SizeBytes: 40},
		{Path: "docs/node_modules", Type: "dir"},
	}
	reader.files["docs/current.md"] = "architecture context"

	report, err := NewRetrieveGitHubRemoteContext(reader).Execute(RetrieveGitHubRemoteContextInput{
		Repository:  "owner/repo",
		Query:       "architecture",
		PathFilters: []string{"docs"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(report.Results) != 1 || report.Results[0].Path != "docs/current.md" {
		t.Fatalf("results = %#v", report.Results)
	}
	assertGitHubRemoteSkipped(t, report.Skipped, "docs/secrets.md", "sensitive_file")
	assertGitHubRemoteSkipped(t, report.Skipped, "docs/private.pem", "sensitive_file")
	assertGitHubRemoteSkipped(t, report.Skipped, "docs/architecture.txt", "unsupported_file_type")
	assertGitHubRemoteSkipped(t, report.Skipped, "docs/large.md", "file_too_large")
	assertGitHubRemoteSkipped(t, report.Skipped, "docs/node_modules", "generated_directory")
}

func TestRetrieveGitHubRemoteContextMapsRemoteFailuresToReports(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus domain.GitHubRemoteContextStatus
	}{
		{
			name: "rate limit",
			err: domain.GitHubRemoteContextError{
				Code:    domain.GitHubRemoteContextErrorRateLimit,
				Message: "GitHub API rate limit exceeded",
			},
			wantStatus: domain.GitHubRemoteContextStatusRateLimited,
		},
		{
			name: "not found",
			err: domain.GitHubRemoteContextError{
				Code:    domain.GitHubRemoteContextErrorNotFound,
				Message: "GitHub repository not found",
			},
			wantStatus: domain.GitHubRemoteContextStatusNotFound,
		},
		{
			name: "invalid token",
			err: domain.GitHubRemoteContextError{
				Code:    domain.GitHubRemoteContextErrorInvalidToken,
				Message: "GitHub token is invalid",
			},
			wantStatus: domain.GitHubRemoteContextStatusUnauthorized,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := newFakeGitHubRemoteReader()
			reader.resolveRepositoryErr = test.err
			report, err := NewRetrieveGitHubRemoteContext(reader).Execute(RetrieveGitHubRemoteContextInput{
				Repository: "owner/repo",
				Query:      "architecture",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if report.Status != test.wantStatus {
				t.Fatalf("Status = %q, want %q", report.Status, test.wantStatus)
			}
			if strings.Contains(report.Message, "secret-token") {
				t.Fatalf("report leaked token-like content: %q", report.Message)
			}
		})
	}
}

func TestRetrieveGitHubRemoteContextReturnsNoResultsSafely(t *testing.T) {
	reader := newFakeGitHubRemoteReader()
	reader.files["README.md"] = "no matching content"

	report, err := NewRetrieveGitHubRemoteContext(reader).Execute(RetrieveGitHubRemoteContextInput{
		Repository: "owner/repo",
		Query:      "architecture",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if report.Status != domain.GitHubRemoteContextStatusNoResults {
		t.Fatalf("Status = %q, want no_results", report.Status)
	}
	if len(report.Results) != 0 {
		t.Fatalf("Results = %#v, want none", report.Results)
	}
}

func TestRetrieveGitHubRemoteContextRejectsInvalidInput(t *testing.T) {
	reader := newFakeGitHubRemoteReader()
	for _, input := range []RetrieveGitHubRemoteContextInput{
		{Repository: "../repo", Query: "architecture"},
		{Repository: "owner/repo", Query: ""},
		{Repository: "owner/repo", Query: "architecture", Ref: "../main"},
		{Repository: "owner/repo", Query: "architecture", PathFilters: []string{"../README.md"}},
	} {
		t.Run(input.Repository+input.Query+input.Ref, func(t *testing.T) {
			if _, err := NewRetrieveGitHubRemoteContext(reader).Execute(input); err == nil {
				t.Fatalf("Execute(%#v) error = nil", input)
			}
		})
	}
}

func TestRetrieveGitHubRemoteContextBoundsResultsAndRenderedOutput(t *testing.T) {
	reader := newFakeGitHubRemoteReader()
	reader.directories["docs"] = []domain.GitHubRemoteEntry{
		{Path: "docs/a.md", Type: "file", SizeBytes: 100},
		{Path: "docs/b.md", Type: "file", SizeBytes: 100},
	}
	reader.files["docs/a.md"] = "architecture " + strings.Repeat("a", 200)
	reader.files["docs/b.md"] = "architecture " + strings.Repeat("b", 200)
	limits := domain.DefaultGitHubRemoteContextLimits()
	limits.MaxResults = 1
	limits.MaxRenderedContentChars = 80
	limits.MaxSnippetChars = 120

	report, err := NewRetrieveGitHubRemoteContextWithLimits(reader, limits).Execute(RetrieveGitHubRemoteContextInput{
		Repository:  "owner/repo",
		Query:       "architecture",
		PathFilters: []string{"docs"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !report.OutputTruncated {
		t.Fatalf("OutputTruncated = false, want true")
	}
	if len(report.Results) > 1 {
		t.Fatalf("len(Results) = %d, want at most 1", len(report.Results))
	}
}

type fakeGitHubRemoteReader struct {
	defaultBranch        string
	files                map[string]string
	directories          map[string][]domain.GitHubRemoteEntry
	reads                []string
	resolveRepositoryErr error
	resolveRefErr        error
	listErr              error
	readErr              error
}

func newFakeGitHubRemoteReader() *fakeGitHubRemoteReader {
	return &fakeGitHubRemoteReader{
		defaultBranch: "main",
		files:         make(map[string]string),
		directories:   make(map[string][]domain.GitHubRemoteEntry),
	}
}

func (reader *fakeGitHubRemoteReader) ResolveRepository(
	locator domain.GitHubRepositoryLocator,
) (domain.GitHubRemoteRepository, error) {
	if reader.resolveRepositoryErr != nil {
		return domain.GitHubRemoteRepository{}, reader.resolveRepositoryErr
	}
	return domain.GitHubRemoteRepository{Locator: locator, DefaultBranch: reader.defaultBranch}, nil
}

func (reader *fakeGitHubRemoteReader) ResolveRef(
	_ domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteRef,
) (domain.GitHubRemoteResolvedRef, error) {
	if reader.resolveRefErr != nil {
		return domain.GitHubRemoteResolvedRef{}, reader.resolveRefErr
	}
	return domain.GitHubRemoteResolvedRef{
		RequestedRef:  ref.Value(),
		DefaultBranch: reader.defaultBranch,
		ResolvedRef:   ref.Value(),
		CommitSHA:     "abc123",
	}, nil
}

func (reader *fakeGitHubRemoteReader) ListDirectory(
	_ domain.GitHubRepositoryLocator,
	_ domain.GitHubRemoteResolvedRef,
	relativePath string,
) ([]domain.GitHubRemoteEntry, error) {
	if reader.listErr != nil {
		return nil, reader.listErr
	}
	entries, ok := reader.directories[relativePath]
	if !ok {
		return nil, domain.GitHubRemoteContextError{Code: domain.GitHubRemoteContextErrorNotFound, Message: "not found"}
	}
	return append([]domain.GitHubRemoteEntry(nil), entries...), nil
}

func (reader *fakeGitHubRemoteReader) ReadFile(
	_ domain.GitHubRepositoryLocator,
	_ domain.GitHubRemoteResolvedRef,
	relativePath string,
	maxBytes int64,
) (domain.GitHubRemoteFile, error) {
	if reader.readErr != nil {
		return domain.GitHubRemoteFile{}, reader.readErr
	}
	reader.reads = append(reader.reads, relativePath)
	contents, ok := reader.files[relativePath]
	if !ok {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorNotFound,
			Message: "not found",
		}
	}
	if int64(len(contents)) > maxBytes {
		return domain.GitHubRemoteFile{}, domain.GitHubRemoteContextError{
			Code:    domain.GitHubRemoteContextErrorOversizedResponse,
			Message: "file too large",
		}
	}
	return domain.GitHubRemoteFile{Path: relativePath, SizeBytes: int64(len(contents)), Contents: []byte(contents)}, nil
}

func assertGitHubRemoteSkipped(
	t *testing.T,
	skipped []domain.GitHubRemoteContextSkipped,
	path string,
	reason string,
) {
	t.Helper()
	for _, skip := range skipped {
		if skip.Path == path && skip.Reason == reason {
			return
		}
	}
	t.Fatalf("skipped = %#v, want %s:%s", skipped, path, reason)
}
