package usecase

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type RetrieveGitHubRemoteContextInput struct {
	Repository  string
	Ref         string
	Query       string
	PathFilters []string
}

type RetrieveGitHubRemoteContext struct {
	reader ports.GitHubRemoteContextReader
	limits domain.GitHubRemoteContextLimits
}

func NewRetrieveGitHubRemoteContext(
	reader ports.GitHubRemoteContextReader,
) *RetrieveGitHubRemoteContext {
	return &RetrieveGitHubRemoteContext{
		reader: reader,
		limits: domain.DefaultGitHubRemoteContextLimits(),
	}
}

func NewRetrieveGitHubRemoteContextWithLimits(
	reader ports.GitHubRemoteContextReader,
	limits domain.GitHubRemoteContextLimits,
) *RetrieveGitHubRemoteContext {
	return &RetrieveGitHubRemoteContext{
		reader: reader,
		limits: limits,
	}
}

func (useCase *RetrieveGitHubRemoteContext) Execute(
	input RetrieveGitHubRemoteContextInput,
) (domain.GitHubRemoteContextReport, error) {
	if useCase == nil {
		return domain.GitHubRemoteContextReport{}, errors.New("github remote context use case is required")
	}
	if useCase.reader == nil {
		return domain.GitHubRemoteContextReport{}, errors.New("github remote context reader is required")
	}
	if err := useCase.limits.Validate(); err != nil {
		return domain.GitHubRemoteContextReport{}, err
	}

	locator, err := domain.NewGitHubRepositoryLocator(input.Repository, useCase.limits)
	if err != nil {
		return domain.GitHubRemoteContextReport{}, err
	}
	query, err := domain.NewGitHubRemoteContextQuery(input.Query, useCase.limits)
	if err != nil {
		return domain.GitHubRemoteContextReport{}, err
	}
	ref, err := useCase.parseRef(input.Ref)
	if err != nil {
		return domain.GitHubRemoteContextReport{}, err
	}
	filters, err := useCase.parsePathFilters(input.PathFilters)
	if err != nil {
		return domain.GitHubRemoteContextReport{}, err
	}

	repository, err := useCase.reader.ResolveRepository(locator)
	if err != nil {
		return useCase.failureReport(locator, domain.GitHubRemoteResolvedRef{}, query, filters, err), nil
	}
	resolvedRefInput := ref
	if !resolvedRefInput.Provided() {
		resolvedRefInput, err = domain.NewGitHubRemoteRef(repository.DefaultBranch, useCase.limits)
		if err != nil {
			return domain.GitHubRemoteContextReport{}, err
		}
	}
	resolvedRef, err := useCase.reader.ResolveRef(locator, resolvedRefInput)
	if err != nil {
		return useCase.failureReport(locator, domain.GitHubRemoteResolvedRef{
			RequestedRef:  ref.Value(),
			DefaultBranch: repository.DefaultBranch,
		}, query, filters, err), nil
	}
	if resolvedRef.DefaultBranch == "" {
		resolvedRef.DefaultBranch = repository.DefaultBranch
	}
	if !ref.Provided() {
		resolvedRef.RequestedRef = ""
	} else if resolvedRef.RequestedRef == "" {
		resolvedRef.RequestedRef = ref.Value()
	}

	candidates, skipped, truncated, err := useCase.collectCandidates(locator, resolvedRef, filters)
	if err != nil {
		return useCase.failureReport(locator, resolvedRef, query, filters, err), nil
	}
	results, outputTruncated, err := useCase.retrieveResults(locator, resolvedRef, query, candidates, &skipped)
	if err != nil {
		return useCase.failureReport(locator, resolvedRef, query, filters, err), nil
	}
	if truncated {
		outputTruncated = true
	}
	status := domain.GitHubRemoteContextStatusCurrent
	message := ""
	if len(results) == 0 {
		status = domain.GitHubRemoteContextStatusNoResults
		message = "no matching GitHub remote context found"
	}
	return domain.NewGitHubRemoteContextReport(
		status,
		locator,
		resolvedRef,
		query,
		filters,
		results,
		skipped,
		message,
		outputTruncated,
	), nil
}

func (useCase *RetrieveGitHubRemoteContext) parseRef(raw string) (domain.GitHubRemoteRef, error) {
	if strings.TrimSpace(raw) == "" {
		return domain.DefaultGitHubRemoteRef(), nil
	}
	return domain.NewGitHubRemoteRef(raw, useCase.limits)
}

func (useCase *RetrieveGitHubRemoteContext) parsePathFilters(rawFilters []string) ([]domain.GitHubRemotePathFilter, error) {
	if len(rawFilters) > useCase.limits.MaxPathFilters {
		return nil, fmt.Errorf("github path filters must be at most %d", useCase.limits.MaxPathFilters)
	}
	filters := make([]domain.GitHubRemotePathFilter, 0, len(rawFilters))
	seen := make(map[string]bool)
	for _, raw := range rawFilters {
		filter, err := domain.NewGitHubRemotePathFilter(raw, useCase.limits)
		if err != nil {
			return nil, err
		}
		if seen[filter.String()] {
			continue
		}
		seen[filter.String()] = true
		filters = append(filters, filter)
	}
	sort.SliceStable(filters, func(i, j int) bool { return filters[i].String() < filters[j].String() })
	return filters, nil
}

func (useCase *RetrieveGitHubRemoteContext) collectCandidates(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	filters []domain.GitHubRemotePathFilter,
) ([]domain.GitHubRemoteCandidate, []domain.GitHubRemoteContextSkipped, bool, error) {
	collector := &githubRemoteCandidateCollector{
		useCase: useCase,
		locator: locator,
		ref:     ref,
		filters: filters,
	}
	for _, spec := range domain.DefaultGitHubRemoteSourceSpecs() {
		if spec.Directory {
			if !githubRemoteDirectoryRelevantToFilters(spec.Path, filters) {
				continue
			}
			if err := collector.collectDirectory(spec, spec.Path, useCase.limits.MaxDirectoryDepth); err != nil {
				return nil, nil, false, err
			}
			continue
		}
		if !githubRemotePathAllowedByFilters(spec.Path, filters) {
			continue
		}
		collector.addCandidate(spec, spec.Path, 0)
	}
	return collector.candidates, collector.skipped, collector.truncated, nil
}

type githubRemoteCandidateCollector struct {
	useCase *RetrieveGitHubRemoteContext
	locator domain.GitHubRepositoryLocator
	ref     domain.GitHubRemoteResolvedRef
	filters []domain.GitHubRemotePathFilter

	candidates  []domain.GitHubRemoteCandidate
	skipped     []domain.GitHubRemoteContextSkipped
	treeEntries int
	truncated   bool
}

func (collector *githubRemoteCandidateCollector) collectDirectory(
	spec domain.GitHubRemoteSourceSpec,
	relativePath string,
	depthRemaining int,
) error {
	if collector.treeEntries >= collector.useCase.limits.MaxTreeEntriesScanned {
		collector.addSkip(relativePath, string(domain.GitHubRemoteContextErrorTooManyCandidates))
		collector.truncated = true
		return nil
	}
	entries, err := collector.useCase.reader.ListDirectory(collector.locator, collector.ref, relativePath)
	if err != nil {
		if githubRemoteErrorCode(err) == domain.GitHubRemoteContextErrorNotFound {
			return nil
		}
		return err
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	for _, entry := range entries {
		collector.treeEntries++
		if collector.treeEntries > collector.useCase.limits.MaxTreeEntriesScanned {
			collector.addSkip(relativePath, string(domain.GitHubRemoteContextErrorTooManyCandidates))
			collector.truncated = true
			return nil
		}
		entryPath, err := domain.SafeGitHubRemotePath(cleanGitHubRemoteEntryPath(relativePath, entry.Path))
		if err != nil {
			collector.addSkip(entry.Path, "unsafe_path")
			continue
		}
		if reason, skip := domain.ShouldSkipGitHubRemotePath(entryPath); skip {
			collector.addSkip(entryPath, reason)
			continue
		}
		if entry.Type == "dir" {
			if depthRemaining <= 0 {
				collector.addSkip(entryPath, "max_directory_depth")
				continue
			}
			if !githubRemoteDirectoryRelevantToFilters(entryPath, collector.filters) {
				continue
			}
			if err := collector.collectDirectory(spec, entryPath, depthRemaining-1); err != nil {
				return err
			}
			continue
		}
		if entry.Type != "file" {
			collector.addSkip(entryPath, "unsupported_file_type")
			continue
		}
		if !githubRemoteHasSupportedExtension(entryPath, spec.Extensions) {
			collector.addSkip(entryPath, "unsupported_file_type")
			continue
		}
		if !githubRemotePathAllowedByFilters(entryPath, collector.filters) {
			continue
		}
		collector.addCandidate(spec, entryPath, entry.SizeBytes)
	}
	return nil
}

func (collector *githubRemoteCandidateCollector) addCandidate(
	spec domain.GitHubRemoteSourceSpec,
	relativePath string,
	sizeBytes int64,
) {
	if len(collector.candidates) >= collector.useCase.limits.MaxFilesFetched {
		collector.addSkip(relativePath, string(domain.GitHubRemoteContextErrorTooManyCandidates))
		collector.truncated = true
		return
	}
	if reason, skip := domain.ShouldSkipGitHubRemotePath(relativePath); skip {
		collector.addSkip(relativePath, reason)
		return
	}
	if sizeBytes > collector.useCase.limits.MaxFileSizeBytes {
		collector.addSkip(relativePath, "file_too_large")
		collector.truncated = true
		return
	}
	candidate, err := domain.NewGitHubRemoteCandidate(domain.GitHubRemoteCandidate{
		Path:                   relativePath,
		SourceCategory:         spec.SourceCategory,
		FileType:               spec.FileType,
		LanguageOrEcosystem:    spec.LanguageOrEcosystem,
		SizeBytes:              sizeBytes,
		SourceEvidenceCategory: spec.SourceEvidenceCategory,
	})
	if err != nil {
		collector.addSkip(relativePath, "unsafe_path")
		return
	}
	collector.candidates = append(collector.candidates, candidate)
}

func (collector *githubRemoteCandidateCollector) addSkip(relativePath string, reason string) {
	if len(collector.skipped) >= collector.useCase.limits.MaxSkippedRecords {
		collector.truncated = true
		return
	}
	pathValue := strings.TrimSpace(relativePath)
	if safePath, err := domain.SafeGitHubRemotePath(relativePath); err == nil {
		pathValue = safePath
	}
	if pathValue == "" {
		pathValue = "unknown"
	}
	collector.skipped = append(collector.skipped, domain.GitHubRemoteContextSkipped{
		Path:   pathValue,
		Reason: reason,
	})
}

func (useCase *RetrieveGitHubRemoteContext) retrieveResults(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	query domain.GitHubRemoteContextQuery,
	candidates []domain.GitHubRemoteCandidate,
	skipped *[]domain.GitHubRemoteContextSkipped,
) ([]domain.GitHubRemoteContextResult, bool, error) {
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].Path < candidates[j].Path })
	var results []domain.GitHubRemoteContextResult
	var totalBytes int64
	for _, candidate := range candidates {
		if totalBytes+candidate.SizeBytes > useCase.limits.MaxTotalFileBytes && candidate.SizeBytes > 0 {
			appendGitHubRemoteSkip(skipped, candidate.Path, "max_total_bytes_reached", useCase.limits)
			break
		}
		file, err := useCase.reader.ReadFile(locator, ref, candidate.Path, useCase.limits.MaxFileSizeBytes)
		if err != nil {
			code := githubRemoteErrorCode(err)
			if code == domain.GitHubRemoteContextErrorNotFound ||
				code == domain.GitHubRemoteContextErrorOversizedResponse ||
				code == domain.GitHubRemoteContextErrorUnsupportedContent {
				appendGitHubRemoteSkip(skipped, candidate.Path, string(code), useCase.limits)
				continue
			}
			return nil, false, err
		}
		if int64(len(file.Contents)) > useCase.limits.MaxFileSizeBytes {
			appendGitHubRemoteSkip(skipped, candidate.Path, "file_too_large", useCase.limits)
			continue
		}
		if totalBytes+int64(len(file.Contents)) > useCase.limits.MaxTotalFileBytes {
			appendGitHubRemoteSkip(skipped, candidate.Path, "max_total_bytes_reached", useCase.limits)
			break
		}
		if !utf8.Valid(file.Contents) {
			appendGitHubRemoteSkip(skipped, candidate.Path, "unsupported_content", useCase.limits)
			continue
		}
		totalBytes += int64(len(file.Contents))
		fileResults, err := extractGitHubRemoteContextResults(locator, ref, query, candidate, string(file.Contents), useCase.limits)
		if err != nil {
			return nil, false, err
		}
		results = append(results, fileResults...)
	}

	domain.SortGitHubRemoteContextResults(results)
	outputTruncated := false
	if len(results) > useCase.limits.MaxResults {
		results = results[:useCase.limits.MaxResults]
		outputTruncated = true
	}
	results, truncatedByOutput := limitGitHubRemoteRenderedOutput(results, useCase.limits.MaxRenderedContentChars)
	if truncatedByOutput {
		outputTruncated = true
	}
	return results, outputTruncated, nil
}

func extractGitHubRemoteContextResults(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	query domain.GitHubRemoteContextQuery,
	candidate domain.GitHubRemoteCandidate,
	contents string,
	limits domain.GitHubRemoteContextLimits,
) ([]domain.GitHubRemoteContextResult, error) {
	lines := strings.Split(contents, "\n")
	windows := matchingGitHubRemoteWindows(query, lines, 2)
	results := make([]domain.GitHubRemoteContextResult, 0, len(windows))
	for _, window := range windows {
		text := strings.Join(lines[window.start-1:window.end], "\n")
		text = domain.TrimLocalContextRetrievalSnippet(text, limits.MaxSnippetChars)
		score := domain.ScoreGitHubRemoteContextCandidate(query, candidate, text)
		if !score.Matched {
			continue
		}
		result, err := domain.NewGitHubRemoteContextResult(locator, ref, candidate, score, domain.GitHubRemoteSnippet{
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
		if result, ok := metadataOnlyGitHubRemoteContextResult(locator, ref, query, candidate); ok {
			return []domain.GitHubRemoteContextResult{result}, nil
		}
		return nil, nil
	}
	domain.SortGitHubRemoteContextResults(results)
	if len(results) > limits.MaxSnippetsPerFile {
		results = results[:limits.MaxSnippetsPerFile]
	}
	return results, nil
}

type githubRemoteWindow struct {
	start int
	end   int
}

func matchingGitHubRemoteWindows(
	query domain.GitHubRemoteContextQuery,
	lines []string,
	contextWindowLines int,
) []githubRemoteWindow {
	var windows []githubRemoteWindow
	for index, line := range lines {
		if !githubRemoteLineMatches(query, line) {
			continue
		}
		lineNumber := index + 1
		start := lineNumber - contextWindowLines
		if start < 1 {
			start = 1
		}
		end := lineNumber + contextWindowLines
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
		windows = append(windows, githubRemoteWindow{start: start, end: end})
	}
	return windows
}

func githubRemoteLineMatches(query domain.GitHubRemoteContextQuery, line string) bool {
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

func metadataOnlyGitHubRemoteContextResult(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	query domain.GitHubRemoteContextQuery,
	candidate domain.GitHubRemoteCandidate,
) (domain.GitHubRemoteContextResult, bool) {
	score := domain.ScoreGitHubRemoteContextCandidate(query, candidate, "")
	if !score.Matched {
		return domain.GitHubRemoteContextResult{}, false
	}
	summary := fmt.Sprintf("Remote source metadata matched query terms. File type: %s.", candidate.FileType)
	result, err := domain.NewGitHubRemoteContextResult(locator, ref, candidate, score, domain.GitHubRemoteSnippet{}, summary)
	if err != nil {
		return domain.GitHubRemoteContextResult{}, false
	}
	return result, true
}

func (useCase *RetrieveGitHubRemoteContext) failureReport(
	locator domain.GitHubRepositoryLocator,
	ref domain.GitHubRemoteResolvedRef,
	query domain.GitHubRemoteContextQuery,
	filters []domain.GitHubRemotePathFilter,
	err error,
) domain.GitHubRemoteContextReport {
	status := domain.GitHubRemoteContextStatusFailed
	switch githubRemoteErrorCode(err) {
	case domain.GitHubRemoteContextErrorRateLimit:
		status = domain.GitHubRemoteContextStatusRateLimited
	case domain.GitHubRemoteContextErrorUnauthorized, domain.GitHubRemoteContextErrorInvalidToken:
		status = domain.GitHubRemoteContextStatusUnauthorized
	case domain.GitHubRemoteContextErrorForbidden:
		status = domain.GitHubRemoteContextStatusForbidden
	case domain.GitHubRemoteContextErrorNotFound:
		status = domain.GitHubRemoteContextStatusNotFound
	}
	return domain.NewGitHubRemoteContextReport(status, locator, ref, query, filters, nil, nil, err.Error(), false)
}

func githubRemotePathAllowedByFilters(relativePath string, filters []domain.GitHubRemotePathFilter) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		value := filter.String()
		if value == relativePath || strings.HasPrefix(relativePath, value+"/") {
			return true
		}
	}
	return false
}

func githubRemoteDirectoryRelevantToFilters(relativePath string, filters []domain.GitHubRemotePathFilter) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		value := filter.String()
		if value == relativePath ||
			strings.HasPrefix(value, relativePath+"/") ||
			strings.HasPrefix(relativePath, value+"/") {
			return true
		}
	}
	return false
}

func githubRemoteHasSupportedExtension(relativePath string, extensions []string) bool {
	lower := strings.ToLower(relativePath)
	for _, extension := range extensions {
		if strings.HasSuffix(lower, strings.ToLower(extension)) {
			return true
		}
	}
	return false
}

func githubRemoteErrorCode(err error) domain.GitHubRemoteContextErrorCode {
	var remoteErr domain.GitHubRemoteContextError
	if errors.As(err, &remoteErr) {
		return remoteErr.Code
	}
	return ""
}

func appendGitHubRemoteSkip(
	skipped *[]domain.GitHubRemoteContextSkipped,
	relativePath string,
	reason string,
	limits domain.GitHubRemoteContextLimits,
) {
	if skipped == nil || len(*skipped) >= limits.MaxSkippedRecords {
		return
	}
	pathValue := strings.TrimSpace(relativePath)
	if safePath, err := domain.SafeGitHubRemotePath(relativePath); err == nil {
		pathValue = safePath
	}
	if pathValue == "" {
		pathValue = "unknown"
	}
	*skipped = append(*skipped, domain.GitHubRemoteContextSkipped{Path: pathValue, Reason: reason})
}

func limitGitHubRemoteRenderedOutput(
	results []domain.GitHubRemoteContextResult,
	maxChars int,
) ([]domain.GitHubRemoteContextResult, bool) {
	var used int
	truncated := false
	limited := make([]domain.GitHubRemoteContextResult, 0, len(results))
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

func cleanGitHubRemoteEntryPath(parent string, entryPath string) string {
	if strings.TrimSpace(entryPath) == "" {
		return parent
	}
	if strings.Contains(entryPath, "/") {
		return entryPath
	}
	return path.Join(parent, entryPath)
}
