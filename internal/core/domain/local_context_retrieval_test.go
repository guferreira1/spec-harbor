package domain

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestLocalContextRetrievalQueryValidatesAndNormalizes(t *testing.T) {
	query, err := NewLocalContextRetrievalQuery("  Hexagonal architecture, HEXAGONAL ports!  ", DefaultLocalContextRetrievalLimits())
	if err != nil {
		t.Fatalf("NewLocalContextRetrievalQuery() error = %v", err)
	}
	if query.DisplayQuery != "Hexagonal architecture, HEXAGONAL ports!" {
		t.Fatalf("DisplayQuery = %q", query.DisplayQuery)
	}
	if query.NormalizedPhrase != "hexagonal architecture ports" {
		t.Fatalf("NormalizedPhrase = %q", query.NormalizedPhrase)
	}
	if !reflect.DeepEqual(query.Terms, []string{"hexagonal", "architecture", "ports"}) {
		t.Fatalf("Terms = %v", query.Terms)
	}
	if !reflect.DeepEqual(query.SortedTerms, []string{"architecture", "hexagonal", "ports"}) {
		t.Fatalf("SortedTerms = %v", query.SortedTerms)
	}
}

func TestLocalContextRetrievalQueryRejectsMissingEmptyAndOversizedInput(t *testing.T) {
	limits := DefaultLocalContextRetrievalLimits()
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{name: "empty", query: "", want: "query is required"},
		{name: "punctuation only", query: "??? ---", want: "at least one letter or digit"},
		{name: "oversized", query: strings.Repeat("a", limits.MaxQueryChars+1), want: "at most 512 characters"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewLocalContextRetrievalQuery(test.query, limits)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("NewLocalContextRetrievalQuery() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLocalContextRetrievalQueryBoundsTermCount(t *testing.T) {
	limits := DefaultLocalContextRetrievalLimits()
	rawTerms := make([]string, 0, limits.MaxQueryTerms+5)
	for index := 0; index < limits.MaxQueryTerms+5; index++ {
		rawTerms = append(rawTerms, fmt.Sprintf("term%d", index))
	}
	query, err := NewLocalContextRetrievalQuery(strings.Join(rawTerms, " "), limits)
	if err != nil {
		t.Fatalf("NewLocalContextRetrievalQuery() error = %v", err)
	}
	if len(query.Terms) != limits.MaxQueryTerms {
		t.Fatalf("Terms = %d, want %d", len(query.Terms), limits.MaxQueryTerms)
	}
}

func TestLocalContextRetrievalScoringAndTieBreakingAreDeterministic(t *testing.T) {
	query, err := NewLocalContextRetrievalQuery("architecture", DefaultLocalContextRetrievalLimits())
	if err != nil {
		t.Fatalf("query error = %v", err)
	}
	readme := mustLocalContextRetrievalResult(t, query, "README.md", ContextSourceCategoryReadme, "# Architecture\nHexagonal architecture")
	openspec := mustLocalContextRetrievalResult(t, query, "openspec/project.md", ContextSourceCategoryOpenSpecProject, "# Architecture\nHexagonal architecture")
	agent := mustLocalContextRetrievalResult(t, query, "AGENTS.md", ContextSourceCategoryAgentInstruction, "# Architecture\nHexagonal architecture")

	results := []LocalContextRetrievalResult{readme, agent, openspec}
	SortLocalContextRetrievalResults(results)
	got := []string{results[0].Path, results[1].Path, results[2].Path}
	want := []string{"openspec/project.md", "AGENTS.md", "README.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result order = %v, want %v", got, want)
	}
}

func TestLocalContextRetrievalResultRejectsUnsafePath(t *testing.T) {
	query, err := NewLocalContextRetrievalQuery("secret", DefaultLocalContextRetrievalLimits())
	if err != nil {
		t.Fatalf("query error = %v", err)
	}
	entry := localContextEntry("../secret.md", ContextSourceCategoryReadme)
	score := ScoreLocalContextRetrievalCandidate(query, entry, "secret")
	_, err = NewLocalContextRetrievalResult(entry, score, LocalContextRetrievalSnippet{LineStart: 1, LineEnd: 1, Text: "secret"}, "")
	if err == nil || !strings.Contains(err.Error(), "path must not contain path traversal") {
		t.Fatalf("NewLocalContextRetrievalResult() error = %v, want unsafe path", err)
	}
}

func mustLocalContextRetrievalResult(
	t *testing.T,
	query LocalContextRetrievalQuery,
	relativePath string,
	category ContextSourceCategory,
	snippet string,
) LocalContextRetrievalResult {
	t.Helper()
	entry := localContextEntry(relativePath, category)
	score := ScoreLocalContextRetrievalCandidate(query, entry, snippet)
	result, err := NewLocalContextRetrievalResult(entry, score, LocalContextRetrievalSnippet{
		LineStart: 1,
		LineEnd:   2,
		Text:      snippet,
	}, "")
	if err != nil {
		t.Fatalf("NewLocalContextRetrievalResult() error = %v", err)
	}
	return result
}

func localContextEntry(relativePath string, category ContextSourceCategory) RepositoryContextIndexEntry {
	return RepositoryContextIndexEntry{
		Path:                   relativePath,
		SourceCategory:         category,
		FileType:               RepositoryContextIndexFileTypeMarkdown,
		LanguageOrEcosystem:    "Markdown",
		SizeBytes:              20,
		ContentHash:            "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ModifiedTime:           "2026-06-13T12:00:00Z",
		SupportedForRetrieval:  true,
		ClassificationHints:    []RepositoryContextIndexClassificationHint{RepositoryContextIndexHintDetectedFact, RepositoryContextIndexHintInventoryMetadata},
		SourceEvidenceCategory: "test",
	}
}
