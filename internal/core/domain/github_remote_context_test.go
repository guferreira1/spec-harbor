package domain

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewGitHubRepositoryLocatorValidatesAndNormalizes(t *testing.T) {
	limits := DefaultGitHubRemoteContextLimits()

	for _, test := range []struct {
		name      string
		raw       string
		wantOwner string
		wantName  string
	}{
		{name: "owner name", raw: "guferreira1/spec-harbor", wantOwner: "guferreira1", wantName: "spec-harbor"},
		{name: "url", raw: "https://github.com/guferreira1/spec-harbor", wantOwner: "guferreira1", wantName: "spec-harbor"},
	} {
		t.Run(test.name, func(t *testing.T) {
			locator, err := NewGitHubRepositoryLocator(test.raw, limits)
			if err != nil {
				t.Fatalf("NewGitHubRepositoryLocator() error = %v", err)
			}
			if locator.Owner() != test.wantOwner || locator.Name() != test.wantName {
				t.Fatalf("locator = %s/%s, want %s/%s", locator.Owner(), locator.Name(), test.wantOwner, test.wantName)
			}
		})
	}
}

func TestNewGitHubRepositoryLocatorRejectsUnsafeInput(t *testing.T) {
	limits := DefaultGitHubRemoteContextLimits()
	tests := []string{
		"",
		"owner",
		"owner/name/extra",
		"owner/",
		"/owner/name",
		`C:\owner\name`,
		"owner/../repo",
		"https://github.com/owner/name?token=secret",
		"https://github.com/user:token@github.com/owner/name",
		"https://example.com/owner/name",
		"git@github.com:owner/name",
		"owner/name#main",
		"owner/name withspace",
	}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			if _, err := NewGitHubRepositoryLocator(raw, limits); err == nil {
				t.Fatalf("NewGitHubRepositoryLocator(%q) error = nil", raw)
			}
		})
	}
}

func TestNewGitHubRemoteRefRejectsUnsafeInput(t *testing.T) {
	limits := DefaultGitHubRemoteContextLimits()
	valid, err := NewGitHubRemoteRef("feature/context", limits)
	if err != nil {
		t.Fatalf("NewGitHubRemoteRef(valid) error = %v", err)
	}
	if !valid.Provided() || valid.Value() != "feature/context" {
		t.Fatalf("ref = %#v, want provided feature/context", valid)
	}

	for _, raw := range []string{
		"",
		"../main",
		"feature//context",
		"/main",
		"main/",
		"https://github.com/owner/repo",
		"user:token@main",
		"main?token=secret",
		"main#fragment",
		"main branch",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := NewGitHubRemoteRef(raw, limits); err == nil {
				t.Fatalf("NewGitHubRemoteRef(%q) error = nil", raw)
			}
		})
	}
}

func TestGitHubRemoteQueryNormalizesTerms(t *testing.T) {
	query, err := NewGitHubRemoteContextQuery("  Hexagonal architecture, HEXAGONAL ports!  ", DefaultGitHubRemoteContextLimits())
	if err != nil {
		t.Fatalf("NewGitHubRemoteContextQuery() error = %v", err)
	}
	if query.DisplayQuery != "Hexagonal architecture, HEXAGONAL ports!" {
		t.Fatalf("DisplayQuery = %q", query.DisplayQuery)
	}
	if query.NormalizedPhrase != "hexagonal architecture ports" {
		t.Fatalf("NormalizedPhrase = %q", query.NormalizedPhrase)
	}
	if !reflect.DeepEqual(query.Terms, []string{"hexagonal", "architecture", "ports"}) {
		t.Fatalf("Terms = %#v", query.Terms)
	}
	if !reflect.DeepEqual(query.SortedTerms, []string{"architecture", "hexagonal", "ports"}) {
		t.Fatalf("SortedTerms = %#v", query.SortedTerms)
	}
}

func TestGitHubRemoteQueryRejectsEmptyOversizedAndTermless(t *testing.T) {
	limits := DefaultGitHubRemoteContextLimits()
	for _, raw := range []string{"", "   ", "!!!", strings.Repeat("a", limits.MaxQueryChars+1)} {
		t.Run(raw, func(t *testing.T) {
			if _, err := NewGitHubRemoteContextQuery(raw, limits); err == nil {
				t.Fatalf("NewGitHubRemoteContextQuery(%q) error = nil", raw)
			}
		})
	}
}

func TestNewGitHubRemotePathFilterValidatesSafeRelativePaths(t *testing.T) {
	limits := DefaultGitHubRemoteContextLimits()
	filter, err := NewGitHubRemotePathFilter("./docs//usage.md", limits)
	if err != nil {
		t.Fatalf("NewGitHubRemotePathFilter(valid) error = %v", err)
	}
	if filter.String() != "docs/usage.md" {
		t.Fatalf("filter = %q, want docs/usage.md", filter.String())
	}

	for _, raw := range []string{
		"",
		"../README.md",
		"/README.md",
		`C:\README.md`,
		"docs/*.md",
		"docs/readme.md?raw=1",
		"docs/readme.md#fragment",
		".env",
		"node_modules/package.json",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := NewGitHubRemotePathFilter(raw, limits); err == nil {
				t.Fatalf("NewGitHubRemotePathFilter(%q) error = nil", raw)
			}
		})
	}
}

func TestShouldSkipGitHubRemotePathIncludesSensitiveAndGeneratedRules(t *testing.T) {
	for _, path := range []string{
		".env",
		".env.local",
		"deploy/private.pem",
		"deploy/private.key",
		"id_rsa",
		"id_ed25519",
		"secrets.yml",
		"credentials.json",
		".npmrc",
		".pypirc",
		".netrc",
		"node_modules/package.json",
		"dist/app.js",
	} {
		t.Run(path, func(t *testing.T) {
			if _, skip := ShouldSkipGitHubRemotePath(path); !skip {
				t.Fatalf("ShouldSkipGitHubRemotePath(%q) skip = false", path)
			}
		})
	}
}

func TestGitHubRemoteContextScoringAndTieBreakingAreDeterministic(t *testing.T) {
	query, err := NewGitHubRemoteContextQuery("architecture", DefaultGitHubRemoteContextLimits())
	if err != nil {
		t.Fatalf("NewGitHubRemoteContextQuery() error = %v", err)
	}
	readme := mustGitHubRemoteCandidate(t, "README.md", ContextSourceCategoryReadme)
	openspec := mustGitHubRemoteCandidate(t, "openspec/project.md", ContextSourceCategoryOpenSpecProject)

	readmeScore := ScoreGitHubRemoteContextCandidate(query, readme, "# Architecture\nHexagonal architecture")
	openspecScore := ScoreGitHubRemoteContextCandidate(query, openspec, "# Architecture\nHexagonal architecture")
	readmeResult := mustGitHubRemoteResult(t, query, readme, readmeScore, "# Architecture\nHexagonal architecture", 1)
	openspecResult := mustGitHubRemoteResult(t, query, openspec, openspecScore, "# Architecture\nHexagonal architecture", 1)

	results := []GitHubRemoteContextResult{readmeResult, openspecResult}
	SortGitHubRemoteContextResults(results)
	if results[0].Path != "openspec/project.md" {
		t.Fatalf("first result = %q, want openspec/project.md", results[0].Path)
	}
}

func TestDefaultGitHubRemoteSourceSpecsIncludeRequiredSources(t *testing.T) {
	specs := DefaultGitHubRemoteSourceSpecs()
	seen := make(map[string]bool)
	for _, spec := range specs {
		seen[spec.Path] = true
	}
	for _, want := range []string{
		"README.md",
		"AGENTS.md",
		"CONTRIBUTING.md",
		"docs",
		"openspec/project.md",
		"openspec/specs",
		".specharbor/rules",
		".specharbor/project-brief.md",
		"package.json",
		"go.mod",
		".github/workflows",
	} {
		if !seen[want] {
			t.Fatalf("DefaultGitHubRemoteSourceSpecs missing %q", want)
		}
	}
}

func mustGitHubRemoteCandidate(
	t *testing.T,
	path string,
	category ContextSourceCategory,
) GitHubRemoteCandidate {
	t.Helper()
	candidate, err := NewGitHubRemoteCandidate(GitHubRemoteCandidate{
		Path:                   path,
		SourceCategory:         category,
		FileType:               RepositoryContextIndexFileTypeMarkdown,
		LanguageOrEcosystem:    "Markdown",
		SizeBytes:              100,
		SourceEvidenceCategory: string(category),
	})
	if err != nil {
		t.Fatalf("NewGitHubRemoteCandidate() error = %v", err)
	}
	return candidate
}

func mustGitHubRemoteResult(
	t *testing.T,
	query GitHubRemoteContextQuery,
	candidate GitHubRemoteCandidate,
	score GitHubRemoteContextScore,
	text string,
	line int,
) GitHubRemoteContextResult {
	t.Helper()
	locator, err := NewGitHubRepositoryLocator("owner/repo", DefaultGitHubRemoteContextLimits())
	if err != nil {
		t.Fatalf("NewGitHubRepositoryLocator() error = %v", err)
	}
	ref := GitHubRemoteResolvedRef{ResolvedRef: "main", CommitSHA: "abc123"}
	if !score.Matched {
		t.Fatalf("score did not match query %#v", query)
	}
	result, err := NewGitHubRemoteContextResult(locator, ref, candidate, score, GitHubRemoteSnippet{
		LineStart: line,
		LineEnd:   line + strings.Count(text, "\n"),
		Text:      text,
	}, "")
	if err != nil {
		t.Fatalf("NewGitHubRemoteContextResult() error = %v", err)
	}
	return result
}
