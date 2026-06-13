package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type RetrieveLocalContextInput struct {
	ProjectRoot string
	Query       string
}

type RetrieveLocalContext struct {
	fileSystem ports.RepositoryContextIndexFileSystem
	limits     domain.LocalContextRetrievalLimits
}

func NewRetrieveLocalContext(
	fileSystem ports.RepositoryContextIndexFileSystem,
) *RetrieveLocalContext {
	return &RetrieveLocalContext{
		fileSystem: fileSystem,
		limits:     domain.DefaultLocalContextRetrievalLimits(),
	}
}

func NewRetrieveLocalContextWithLimits(
	fileSystem ports.RepositoryContextIndexFileSystem,
	limits domain.LocalContextRetrievalLimits,
) *RetrieveLocalContext {
	return &RetrieveLocalContext{
		fileSystem: fileSystem,
		limits:     limits,
	}
}

func (useCase *RetrieveLocalContext) Execute(
	input RetrieveLocalContextInput,
) (domain.LocalContextRetrievalReport, error) {
	if useCase == nil {
		return domain.LocalContextRetrievalReport{}, errors.New("local context retrieval use case is required")
	}
	if useCase.fileSystem == nil {
		return domain.LocalContextRetrievalReport{}, errors.New("local context retrieval filesystem is required")
	}
	if err := useCase.limits.Validate(); err != nil {
		return domain.LocalContextRetrievalReport{}, err
	}
	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.LocalContextRetrievalReport{}, errors.New("project root is required")
	}
	query, err := domain.NewLocalContextRetrievalQuery(input.Query, useCase.limits)
	if err != nil {
		return domain.LocalContextRetrievalReport{}, err
	}

	stored, dependencyReport, ok, err := useCase.loadAndCheckIndex(projectRoot, query)
	if err != nil || !ok {
		return dependencyReport, err
	}

	results, outputTruncated, err := useCase.retrieveFromIndex(projectRoot, query, stored)
	if err != nil {
		return domain.LocalContextRetrievalReport{}, err
	}
	status := domain.LocalContextRetrievalStatusCurrent
	if len(results) == 0 {
		status = domain.LocalContextRetrievalStatusNoResults
	}
	return domain.NewLocalContextRetrievalReport(
		status,
		query,
		domain.RepositoryContextIndexPath,
		results,
		nil,
		"",
		outputTruncated,
	), nil
}

func (useCase *RetrieveLocalContext) loadAndCheckIndex(
	projectRoot string,
	query domain.LocalContextRetrievalQuery,
) (domain.RepositoryContextIndex, domain.LocalContextRetrievalReport, bool, error) {
	exists, err := useCase.fileSystem.FileExists(projectRoot, domain.RepositoryContextIndexPath)
	if err != nil {
		return domain.RepositoryContextIndex{}, domain.LocalContextRetrievalReport{}, false, err
	}
	if !exists {
		return domain.RepositoryContextIndex{}, useCase.indexFailureReport(
			domain.LocalContextRetrievalStatusMissingIndex,
			query,
			"repository context index is missing; run specharbor context index --write",
			nil,
		), false, nil
	}

	contents, err := useCase.fileSystem.ReadFileSafely(projectRoot, domain.RepositoryContextIndexPath)
	if err != nil {
		return domain.RepositoryContextIndex{}, useCase.indexFailureReport(
			domain.LocalContextRetrievalStatusInvalidIndex,
			query,
			fmt.Sprintf("repository context index is unreadable: %v; run specharbor context index --write", err),
			nil,
		), false, nil
	}

	var stored domain.RepositoryContextIndex
	if err := json.Unmarshal([]byte(contents), &stored); err != nil {
		return domain.RepositoryContextIndex{}, useCase.indexFailureReport(
			domain.LocalContextRetrievalStatusInvalidIndex,
			query,
			fmt.Sprintf("repository context index is invalid JSON: %v; run specharbor context index --write", err),
			nil,
		), false, nil
	}
	stored, err = domain.NormalizeRepositoryContextIndex(stored)
	if err != nil {
		return domain.RepositoryContextIndex{}, useCase.indexFailureReport(
			domain.LocalContextRetrievalStatusInvalidIndex,
			query,
			fmt.Sprintf("repository context index is invalid: %v; run specharbor context index --write", err),
			nil,
		), false, nil
	}

	currentReport, err := NewBuildRepositoryContextIndex(useCase.fileSystem).Execute(RepositoryContextIndexInput{
		ProjectRoot: projectRoot,
		Mode:        domain.RepositoryContextIndexModeReport,
	})
	if err != nil {
		return domain.RepositoryContextIndex{}, domain.LocalContextRetrievalReport{}, false, err
	}
	current := currentReport.Index
	if stored.Truncated || current.Truncated {
		return domain.RepositoryContextIndex{}, useCase.indexFailureReport(
			domain.LocalContextRetrievalStatusTruncatedIndex,
			query,
			"repository context index is truncated; run specharbor context index --write after reducing indexed sources",
			nil,
		), false, nil
	}

	reasons := domain.CompareRepositoryContextIndexes(stored, current)
	if len(reasons) > 0 {
		return domain.RepositoryContextIndex{}, useCase.indexFailureReport(
			domain.LocalContextRetrievalStatusStaleIndex,
			query,
			"repository context index is stale; run specharbor context index --write",
			limitLocalContextStaleReasons(reasons, useCase.limits.MaxStaleReasonsInReport),
		), false, nil
	}
	return stored, domain.LocalContextRetrievalReport{}, true, nil
}

func (useCase *RetrieveLocalContext) retrieveFromIndex(
	projectRoot string,
	query domain.LocalContextRetrievalQuery,
	index domain.RepositoryContextIndex,
) ([]domain.LocalContextRetrievalResult, bool, error) {
	var results []domain.LocalContextRetrievalResult
	var totalRead int64
	for _, entry := range index.Entries {
		if !entry.SupportedForRetrieval {
			if result, ok := metadataOnlyLocalContextResult(query, entry); ok {
				results = append(results, result)
			}
			continue
		}
		if !localContextRetrievalEntryEligible(entry, useCase.limits) {
			continue
		}
		if totalRead+entry.SizeBytes > useCase.limits.MaxTotalSourceReadBytes {
			break
		}
		contents, err := useCase.fileSystem.ReadFileBytes(projectRoot, entry.Path, useCase.limits.MaxSourceReadBytes)
		if err != nil {
			continue
		}
		totalRead += int64(len(contents))

		fileResults, err := extractLocalContextResults(query, entry, string(contents), useCase.limits)
		if err != nil {
			return nil, false, err
		}
		results = append(results, fileResults...)
	}

	domain.SortLocalContextRetrievalResults(results)
	outputTruncated := false
	if len(results) > useCase.limits.MaxResults {
		results = results[:useCase.limits.MaxResults]
		outputTruncated = true
	}
	results, truncatedByOutput := limitLocalContextRenderedOutput(results, useCase.limits.MaxRenderedContentChars)
	if truncatedByOutput {
		outputTruncated = true
	}
	return results, outputTruncated, nil
}

func (useCase *RetrieveLocalContext) indexFailureReport(
	status domain.LocalContextRetrievalStatus,
	query domain.LocalContextRetrievalQuery,
	message string,
	reasons []domain.RepositoryContextIndexStaleReason,
) domain.LocalContextRetrievalReport {
	return domain.NewLocalContextRetrievalReport(
		status,
		query,
		domain.RepositoryContextIndexPath,
		nil,
		reasons,
		message,
		false,
	)
}

func localContextRetrievalEntryEligible(
	entry domain.RepositoryContextIndexEntry,
	limits domain.LocalContextRetrievalLimits,
) bool {
	if entry.Path == domain.RepositoryContextIndexPath {
		return false
	}
	if _, err := domain.SafeRepositoryContextIndexPath(entry.Path); err != nil {
		return false
	}
	if _, skip := domain.ShouldSkipRepositoryContextIndexPath(entry.Path); skip {
		return false
	}
	if entry.SizeBytes < 0 || entry.SizeBytes > limits.MaxSourceReadBytes {
		return false
	}
	return true
}

func extractLocalContextResults(
	query domain.LocalContextRetrievalQuery,
	entry domain.RepositoryContextIndexEntry,
	contents string,
	limits domain.LocalContextRetrievalLimits,
) ([]domain.LocalContextRetrievalResult, error) {
	lines := strings.Split(contents, "\n")
	windows := matchingLocalContextWindows(query, lines, limits)
	results := make([]domain.LocalContextRetrievalResult, 0, len(windows))
	for _, window := range windows {
		text := strings.Join(lines[window.start-1:window.end], "\n")
		text = domain.TrimLocalContextRetrievalSnippet(text, limits.MaxSnippetChars)
		score := domain.ScoreLocalContextRetrievalCandidate(query, entry, text)
		if !score.Matched {
			continue
		}
		result, err := domain.NewLocalContextRetrievalResult(entry, score, domain.LocalContextRetrievalSnippet{
			LineStart: window.start,
			LineEnd:   window.end,
			Text:      text,
		}, "")
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		if result, ok := metadataOnlyLocalContextResult(query, entry); ok {
			return []domain.LocalContextRetrievalResult{result}, nil
		}
		return nil, nil
	}
	domain.SortLocalContextRetrievalResults(results)
	if len(results) > limits.MaxSnippetsPerFile {
		results = results[:limits.MaxSnippetsPerFile]
	}
	return results, nil
}

type localContextWindow struct {
	start int
	end   int
}

func matchingLocalContextWindows(
	query domain.LocalContextRetrievalQuery,
	lines []string,
	limits domain.LocalContextRetrievalLimits,
) []localContextWindow {
	var windows []localContextWindow
	for index, line := range lines {
		if !localContextLineMatches(query, line) {
			continue
		}
		lineNumber := index + 1
		start := lineNumber - limits.MaxContextWindowLines
		if start < 1 {
			start = 1
		}
		end := lineNumber + limits.MaxContextWindowLines
		if end > len(lines) {
			end = len(lines)
		}
		if start == 1 && end == len(lines) && len(lines) > 1 {
			start = lineNumber
			end = lineNumber
		}
		if len(windows) > 0 && start <= windows[len(windows)-1].end+1 {
			if end > windows[len(windows)-1].end {
				windows[len(windows)-1].end = end
			}
			continue
		}
		windows = append(windows, localContextWindow{start: start, end: end})
	}
	for index, window := range windows {
		if window.start != 1 || window.end != len(lines) || len(lines) <= 1 {
			continue
		}
		for lineIndex := window.start - 1; lineIndex < window.end; lineIndex++ {
			if localContextLineMatches(query, lines[lineIndex]) {
				lineNumber := lineIndex + 1
				windows[index] = localContextWindow{start: lineNumber, end: lineNumber}
				break
			}
		}
	}
	return windows
}

func localContextLineMatches(query domain.LocalContextRetrievalQuery, line string) bool {
	normalized := strings.Join(strings.FieldsFunc(strings.ToLower(line), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), " ")
	if query.NormalizedPhrase != "" && strings.Contains(normalized, query.NormalizedPhrase) {
		return true
	}
	for _, term := range query.Terms {
		for _, token := range strings.Fields(normalized) {
			if token == term {
				return true
			}
		}
	}
	return false
}

func metadataOnlyLocalContextResult(
	query domain.LocalContextRetrievalQuery,
	entry domain.RepositoryContextIndexEntry,
) (domain.LocalContextRetrievalResult, bool) {
	score := domain.ScoreLocalContextRetrievalCandidate(query, entry, "")
	if !score.Matched {
		return domain.LocalContextRetrievalResult{}, false
	}
	summary := fmt.Sprintf("Indexed source metadata matched query terms. File type: %s.", entry.FileType)
	result, err := domain.NewLocalContextRetrievalResult(entry, score, domain.LocalContextRetrievalSnippet{}, summary)
	if err != nil {
		return domain.LocalContextRetrievalResult{}, false
	}
	return result, true
}

func limitLocalContextRenderedOutput(
	results []domain.LocalContextRetrievalResult,
	maxChars int,
) ([]domain.LocalContextRetrievalResult, bool) {
	var used int
	truncated := false
	limited := make([]domain.LocalContextRetrievalResult, 0, len(results))
	for _, result := range results {
		contentLength := len(result.Snippet.Text) + len(result.Summary)
		if used+contentLength > maxChars {
			truncated = true
			break
		}
		used += contentLength
		limited = append(limited, result)
	}
	return limited, truncated
}

func limitLocalContextStaleReasons(
	reasons []domain.RepositoryContextIndexStaleReason,
	limit int,
) []domain.RepositoryContextIndexStaleReason {
	copied := append([]domain.RepositoryContextIndexStaleReason(nil), reasons...)
	sort.SliceStable(copied, func(i, j int) bool {
		if copied[i].Path != copied[j].Path {
			return copied[i].Path < copied[j].Path
		}
		return copied[i].Code < copied[j].Code
	})
	if limit >= 0 && len(copied) > limit {
		return copied[:limit]
	}
	return copied
}
