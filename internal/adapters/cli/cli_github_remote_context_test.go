package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

func TestContextGitHubPrintsRemoteResults(t *testing.T) {
	reader := newCLIFakeGitHubRemoteReader()
	reader.files["README.md"] = "# Architecture\n\nHexagonal Architecture guides this repository.\n"
	capturedToken := withFakeGitHubRemoteContextReader(t, reader)
	t.Setenv(domain.GitHubRemoteContextTokenEnvVar, "secret-token")

	var output bytes.Buffer
	if err := execute([]string{"context", "github", "--repo", "owner/repo", "--query", "hexagonal architecture"}, &output); err != nil {
		t.Fatalf("execute(context github) error = %v\noutput:\n%s", err, output.String())
	}
	report := output.String()
	for _, want := range []string{
		"GitHub remote context:",
		"Repository: owner/repo",
		"Query: hexagonal architecture",
		"Normalized terms: hexagonal, architecture",
		"Default branch: main",
		"Resolved ref: main",
		"Resolved SHA: abc123",
		"Remote: yes",
		"Results: 1",
		"1. README.md",
		"Category: readme",
		"Lines:",
		"Hexagonal Architecture",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("context github output = %q, want %q", report, want)
		}
	}
	if strings.Contains(report, "secret-token") {
		t.Fatalf("context github output leaked token: %q", report)
	}
	if *capturedToken != "secret-token" {
		t.Fatalf("captured token = %q, want secret-token", *capturedToken)
	}
}

func TestContextGitHubPrintsNoResults(t *testing.T) {
	reader := newCLIFakeGitHubRemoteReader()
	reader.files["README.md"] = "no matching content"
	withFakeGitHubRemoteContextReader(t, reader)

	var output bytes.Buffer
	if err := execute([]string{"context", "github", "--repo", "owner/repo", "--query", "architecture"}, &output); err != nil {
		t.Fatalf("execute(context github no results) error = %v\noutput:\n%s", err, output.String())
	}
	report := output.String()
	if !strings.Contains(report, "Status: no_results") ||
		!strings.Contains(report, "No matching GitHub remote context found.") {
		t.Fatalf("context github no-results output = %q", report)
	}
}

func TestContextGitHubRejectsUnsupportedArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing repo", args: []string{"context", "github", "--query", "architecture"}, want: "context github repo is required"},
		{name: "missing query", args: []string{"context", "github", "--repo", "owner/repo"}, want: "context github query is required"},
		{name: "positional repo", args: []string{"context", "github", "owner/repo", "--query", "architecture"}, want: "unexpected argument: owner/repo"},
		{name: "duplicate repo", args: []string{"context", "github", "--repo", "owner/repo", "--repo", "owner/other", "--query", "architecture"}, want: "context github repo flag specified more than once"},
		{name: "unsupported flag", args: []string{"context", "github", "--repo", "owner/repo", "--query", "architecture", "--rag"}, want: "unsupported flag: --rag"},
		{name: "missing path value", args: []string{"context", "github", "--repo", "owner/repo", "--query", "architecture", "--path"}, want: "context github path value is required"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil {
				t.Fatalf("execute(%v) error = nil", test.args)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func TestContextGitHubReturnsExitErrorForRemoteFailure(t *testing.T) {
	reader := newCLIFakeGitHubRemoteReader()
	reader.resolveRepositoryErr = domain.GitHubRemoteContextError{
		Code:    domain.GitHubRemoteContextErrorRateLimit,
		Message: "GitHub API rate limit exceeded",
	}
	withFakeGitHubRemoteContextReader(t, reader)

	var output bytes.Buffer
	err := execute([]string{"context", "github", "--repo", "owner/repo", "--query", "architecture"}, &output)
	var exitErr ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("error = %T %v, want ExitError 1", err, err)
	}
	report := output.String()
	if !strings.Contains(report, "Status: rate_limited") ||
		!strings.Contains(report, "Detail: GitHub API rate limit exceeded") {
		t.Fatalf("remote failure output = %q", report)
	}
}

func withFakeGitHubRemoteContextReader(
	t *testing.T,
	reader ports.GitHubRemoteContextReader,
) *string {
	t.Helper()
	original := newGitHubRemoteContextReader
	capturedToken := ""
	newGitHubRemoteContextReader = func(token string) ports.GitHubRemoteContextReader {
		capturedToken = token
		return reader
	}
	t.Cleanup(func() {
		newGitHubRemoteContextReader = original
	})
	return &capturedToken
}

type cliFakeGitHubRemoteReader struct {
	defaultBranch        string
	files                map[string]string
	directories          map[string][]domain.GitHubRemoteEntry
	resolveRepositoryErr error
}

func newCLIFakeGitHubRemoteReader() *cliFakeGitHubRemoteReader {
	return &cliFakeGitHubRemoteReader{
		defaultBranch: "main",
		files:         make(map[string]string),
		directories:   make(map[string][]domain.GitHubRemoteEntry),
	}
}

func (reader *cliFakeGitHubRemoteReader) ResolveRepository(
	locator domain.GitHubRepositoryLocator,
) (domain.GitHubRemoteRepository, error) {
	if reader.resolveRepositoryErr != nil {
		return domain.GitHubRemoteRepository{}, reader.resolveRepositoryErr
	}
	return domain.GitHubRemoteRepository{Locator: locator, DefaultBranch: reader.defaultBranch}, nil
}

func (reader *cliFakeGitHubRemoteReader) ResolveRef(
	_ domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteRef,
) (domain.GitHubRemoteResolvedRef, error) {
	return domain.GitHubRemoteResolvedRef{
		RequestedRef:  ref.Value(),
		DefaultBranch: reader.defaultBranch,
		ResolvedRef:   ref.Value(),
		CommitSHA:     "abc123",
	}, nil
}

func (reader *cliFakeGitHubRemoteReader) ListDirectory(
	_ domain.GitHubRepositoryLocator,
	_ domain.GitHubRemoteResolvedRef,
	relativePath string,
) ([]domain.GitHubRemoteEntry, error) {
	entries, ok := reader.directories[relativePath]
	if !ok {
		return nil, domain.GitHubRemoteContextError{Code: domain.GitHubRemoteContextErrorNotFound, Message: "not found"}
	}
	return append([]domain.GitHubRemoteEntry(nil), entries...), nil
}

func (reader *cliFakeGitHubRemoteReader) ReadFile(
	_ domain.GitHubRepositoryLocator,
	_ domain.GitHubRemoteResolvedRef,
	relativePath string,
	maxBytes int64,
) (domain.GitHubRemoteFile, error) {
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
