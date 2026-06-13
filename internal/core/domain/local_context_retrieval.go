package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type LocalContextRetrievalStatus string

const (
	LocalContextRetrievalStatusCurrent        LocalContextRetrievalStatus = "current"
	LocalContextRetrievalStatusNoResults      LocalContextRetrievalStatus = "no_results"
	LocalContextRetrievalStatusMissingIndex   LocalContextRetrievalStatus = "missing_index"
	LocalContextRetrievalStatusInvalidIndex   LocalContextRetrievalStatus = "invalid_index"
	LocalContextRetrievalStatusStaleIndex     LocalContextRetrievalStatus = "stale_index"
	LocalContextRetrievalStatusTruncatedIndex LocalContextRetrievalStatus = "truncated_index"
)

type LocalContextRetrievalLimits struct {
	MaxQueryChars           int
	MaxQueryTerms           int
	MaxSourceReadBytes      int64
	MaxTotalSourceReadBytes int64
	MaxResults              int
	MaxSnippetsPerFile      int
	MaxSnippetChars         int
	MaxContextWindowLines   int
	MaxRenderedContentChars int
	MaxStaleReasonsInReport int
}

type LocalContextRetrievalQuery struct {
	DisplayQuery     string
	NormalizedPhrase string
	Terms            []string
	SortedTerms      []string
}

type LocalContextRetrievalSnippet struct {
	LineStart int
	LineEnd   int
	Text      string
}

type LocalContextRetrievalResult struct {
	Rank                   int
	Path                   string
	SourceCategory         ContextSourceCategory
	SourceEvidenceCategory string
	Score                  int
	CategoryPriority       int
	ClassificationHints    []RepositoryContextIndexClassificationHint
	Snippet                LocalContextRetrievalSnippet
	Summary                string
}

type LocalContextRetrievalReport struct {
	Status          LocalContextRetrievalStatus
	Query           LocalContextRetrievalQuery
	IndexPath       string
	Results         []LocalContextRetrievalResult
	StaleReasons    []RepositoryContextIndexStaleReason
	Message         string
	OutputTruncated bool
}

type LocalContextRetrievalScore struct {
	Value            int
	CategoryPriority int
	Matched          bool
}

func DefaultLocalContextRetrievalLimits() LocalContextRetrievalLimits {
	return LocalContextRetrievalLimits{
		MaxQueryChars:           512,
		MaxQueryTerms:           32,
		MaxSourceReadBytes:      128 * 1024,
		MaxTotalSourceReadBytes: 1024 * 1024,
		MaxResults:              10,
		MaxSnippetsPerFile:      2,
		MaxSnippetChars:         600,
		MaxContextWindowLines:   2,
		MaxRenderedContentChars: 8000,
		MaxStaleReasonsInReport: 10,
	}
}

func NewLocalContextRetrievalQuery(
	rawQuery string,
	limits LocalContextRetrievalLimits,
) (LocalContextRetrievalQuery, error) {
	if err := limits.Validate(); err != nil {
		return LocalContextRetrievalQuery{}, err
	}
	display := strings.TrimSpace(rawQuery)
	if display == "" {
		return LocalContextRetrievalQuery{}, errors.New("retrieval query is required")
	}
	if len(display) > limits.MaxQueryChars {
		return LocalContextRetrievalQuery{}, fmt.Errorf("retrieval query must be at most %d characters", limits.MaxQueryChars)
	}

	terms := deduplicateLocalContextTerms(tokenizeLocalContextText(display))
	if len(terms) == 0 {
		return LocalContextRetrievalQuery{}, errors.New("retrieval query must contain at least one letter or digit term")
	}
	if len(terms) > limits.MaxQueryTerms {
		terms = terms[:limits.MaxQueryTerms]
	}
	sortedTerms := append([]string(nil), terms...)
	sort.Strings(sortedTerms)

	return LocalContextRetrievalQuery{
		DisplayQuery:     display,
		NormalizedPhrase: strings.Join(terms, " "),
		Terms:            append([]string(nil), terms...),
		SortedTerms:      sortedTerms,
	}, nil
}

func NewLocalContextRetrievalResult(
	entry RepositoryContextIndexEntry,
	score LocalContextRetrievalScore,
	snippet LocalContextRetrievalSnippet,
	summary string,
) (LocalContextRetrievalResult, error) {
	relativePath, err := SafeRepositoryContextIndexPath(entry.Path)
	if err != nil {
		return LocalContextRetrievalResult{}, err
	}
	if score.Value <= 0 {
		return LocalContextRetrievalResult{}, errors.New("local context retrieval score must be positive")
	}
	hints := append([]RepositoryContextIndexClassificationHint(nil), entry.ClassificationHints...)
	sort.SliceStable(hints, func(i, j int) bool { return hints[i] < hints[j] })
	if snippet.Text == "" && strings.TrimSpace(summary) == "" {
		return LocalContextRetrievalResult{}, errors.New("local context retrieval result requires a snippet or summary")
	}
	if snippet.Text != "" && (snippet.LineStart <= 0 || snippet.LineEnd < snippet.LineStart) {
		return LocalContextRetrievalResult{}, errors.New("local context retrieval snippet line range is invalid")
	}
	return LocalContextRetrievalResult{
		Path:                   relativePath,
		SourceCategory:         entry.SourceCategory,
		SourceEvidenceCategory: entry.SourceEvidenceCategory,
		Score:                  score.Value,
		CategoryPriority:       score.CategoryPriority,
		ClassificationHints:    hints,
		Snippet:                snippet,
		Summary:                strings.TrimSpace(summary),
	}, nil
}

func NewLocalContextRetrievalReport(
	status LocalContextRetrievalStatus,
	query LocalContextRetrievalQuery,
	indexPath string,
	results []LocalContextRetrievalResult,
	staleReasons []RepositoryContextIndexStaleReason,
	message string,
	outputTruncated bool,
) LocalContextRetrievalReport {
	copiedResults := append([]LocalContextRetrievalResult(nil), results...)
	SortLocalContextRetrievalResults(copiedResults)
	for index := range copiedResults {
		copiedResults[index].Rank = index + 1
	}
	return LocalContextRetrievalReport{
		Status:          status,
		Query:           query,
		IndexPath:       indexPath,
		Results:         copiedResults,
		StaleReasons:    append([]RepositoryContextIndexStaleReason(nil), staleReasons...),
		Message:         strings.TrimSpace(message),
		OutputTruncated: outputTruncated,
	}
}

func ScoreLocalContextRetrievalCandidate(
	query LocalContextRetrievalQuery,
	entry RepositoryContextIndexEntry,
	snippetText string,
) LocalContextRetrievalScore {
	content := normalizeLocalContextMatchText(snippetText)
	entryPath := normalizeLocalContextMatchText(entry.Path)
	filename := normalizeLocalContextMatchText(localContextPathBase(entry.Path))

	score := 0
	matched := false
	if query.NormalizedPhrase != "" && content != "" && strings.Contains(content, query.NormalizedPhrase) {
		score += 50
		matched = true
	}
	if query.NormalizedPhrase != "" && strings.Contains(entryPath, query.NormalizedPhrase) {
		score += 30
		matched = true
	}

	for _, term := range query.Terms {
		if contentCount := countLocalContextTerm(content, term); contentCount > 0 {
			if contentCount > 10 {
				contentCount = 10
			}
			score += contentCount * 3
			matched = true
		}
		if containsLocalContextTerm(entryPath, term) {
			score += 8
			matched = true
		}
		if containsLocalContextTerm(filename, term) {
			score += 12
			matched = true
		}
	}

	if localContextSnippetHasHeadingMatch(snippetText, query.Terms) {
		score += 10
	}

	categoryPriority := LocalContextRetrievalSourceCategoryPriority(entry.SourceCategory)
	if matched {
		score += categoryPriority
		score += localContextRetrievalClassificationPriority(entry.ClassificationHints)
	}
	return LocalContextRetrievalScore{
		Value:            score,
		CategoryPriority: categoryPriority,
		Matched:          matched,
	}
}

func SortLocalContextRetrievalResults(results []LocalContextRetrievalResult) {
	sort.SliceStable(results, func(i, j int) bool {
		left := results[i]
		right := results[j]
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if left.CategoryPriority != right.CategoryPriority {
			return left.CategoryPriority > right.CategoryPriority
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Snippet.LineStart != right.Snippet.LineStart {
			return left.Snippet.LineStart < right.Snippet.LineStart
		}
		return left.Snippet.Text < right.Snippet.Text
	})
}

func (limits LocalContextRetrievalLimits) Validate() error {
	if limits.MaxQueryChars <= 0 {
		return errors.New("local context retrieval max query chars must be positive")
	}
	if limits.MaxQueryTerms <= 0 {
		return errors.New("local context retrieval max query terms must be positive")
	}
	if limits.MaxSourceReadBytes <= 0 {
		return errors.New("local context retrieval max source read bytes must be positive")
	}
	if limits.MaxTotalSourceReadBytes <= 0 {
		return errors.New("local context retrieval max total source read bytes must be positive")
	}
	if limits.MaxResults <= 0 {
		return errors.New("local context retrieval max results must be positive")
	}
	if limits.MaxSnippetsPerFile <= 0 {
		return errors.New("local context retrieval max snippets per file must be positive")
	}
	if limits.MaxSnippetChars <= 0 {
		return errors.New("local context retrieval max snippet chars must be positive")
	}
	if limits.MaxContextWindowLines < 0 {
		return errors.New("local context retrieval context window lines must not be negative")
	}
	if limits.MaxRenderedContentChars <= 0 {
		return errors.New("local context retrieval max rendered content chars must be positive")
	}
	if limits.MaxStaleReasonsInReport < 0 {
		return errors.New("local context retrieval max stale reasons must not be negative")
	}
	return nil
}

func LocalContextRetrievalSourceCategoryPriority(category ContextSourceCategory) int {
	switch category {
	case ContextSourceCategoryProjectBrief:
		return 80
	case ContextSourceCategoryOpenSpecProject:
		return 75
	case ContextSourceCategoryOpenSpecSpec:
		return 70
	case ContextSourceCategorySpecHarborRules, ContextSourceCategoryAgentInstruction:
		return 65
	case ContextSourceCategoryReadme, ContextSourceCategoryContributing:
		return 60
	case ContextSourceCategoryDocumentation:
		return 55
	case ContextSourceCategoryPackageManifest,
		ContextSourceCategoryDependencyManifest,
		ContextSourceCategoryBuildManifest:
		return 45
	case ContextSourceCategoryWorkflowConfig,
		ContextSourceCategoryTaskRunner,
		ContextSourceCategoryContainerConfig:
		return 35
	default:
		return 10
	}
}

func TrimLocalContextRetrievalSnippet(text string, maxChars int) string {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) <= maxChars {
		return trimmed
	}
	if maxChars <= 3 {
		return trimmed[:maxChars]
	}
	return strings.TrimSpace(trimmed[:maxChars-3]) + "..."
}

func tokenizeLocalContextText(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		tokens = append(tokens, field)
	}
	return tokens
}

func deduplicateLocalContextTerms(terms []string) []string {
	seen := make(map[string]bool, len(terms))
	result := make([]string, 0, len(terms))
	for _, term := range terms {
		if seen[term] {
			continue
		}
		seen[term] = true
		result = append(result, term)
	}
	return result
}

func normalizeLocalContextMatchText(value string) string {
	return strings.Join(tokenizeLocalContextText(value), " ")
}

func countLocalContextTerm(normalizedText string, term string) int {
	count := 0
	for _, token := range strings.Fields(normalizedText) {
		if token == term {
			count++
		}
	}
	return count
}

func containsLocalContextTerm(normalizedText string, term string) bool {
	return countLocalContextTerm(normalizedText, term) > 0
}

func localContextSnippetHasHeadingMatch(snippetText string, terms []string) bool {
	for _, line := range strings.Split(snippetText, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		normalized := normalizeLocalContextMatchText(trimmed)
		for _, term := range terms {
			if containsLocalContextTerm(normalized, term) {
				return true
			}
		}
	}
	return false
}

func localContextRetrievalClassificationPriority(
	hints []RepositoryContextIndexClassificationHint,
) int {
	score := 0
	for _, hint := range hints {
		switch hint {
		case RepositoryContextIndexHintUserConfirmedContext:
			score += 20
		case RepositoryContextIndexHintDetectedFact:
			score += 10
		case RepositoryContextIndexHintSuggestedAssumption:
			score += 5
		case RepositoryContextIndexHintInventoryMetadata:
			score += 3
		}
	}
	return score
}

func localContextPathBase(value string) string {
	normalized := normalizeContextPath(value)
	segments := strings.Split(normalized, "/")
	if len(segments) == 0 {
		return normalized
	}
	return segments[len(segments)-1]
}
