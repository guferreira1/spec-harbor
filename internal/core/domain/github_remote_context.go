package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const GitHubRemoteContextTokenEnvVar = "SPECHARBOR_GITHUB_TOKEN"

type GitHubRemoteContextStatus string

const (
	GitHubRemoteContextStatusCurrent      GitHubRemoteContextStatus = "current"
	GitHubRemoteContextStatusNoResults    GitHubRemoteContextStatus = "no_results"
	GitHubRemoteContextStatusFailed       GitHubRemoteContextStatus = "failed"
	GitHubRemoteContextStatusRateLimited  GitHubRemoteContextStatus = "rate_limited"
	GitHubRemoteContextStatusUnauthorized GitHubRemoteContextStatus = "unauthorized"
	GitHubRemoteContextStatusForbidden    GitHubRemoteContextStatus = "forbidden"
	GitHubRemoteContextStatusNotFound     GitHubRemoteContextStatus = "not_found"
)

type GitHubRemoteContextErrorCode string

const (
	GitHubRemoteContextErrorNetwork            GitHubRemoteContextErrorCode = "network_unavailable"
	GitHubRemoteContextErrorTimeout            GitHubRemoteContextErrorCode = "timeout"
	GitHubRemoteContextErrorRateLimit          GitHubRemoteContextErrorCode = "rate_limit"
	GitHubRemoteContextErrorUnauthorized       GitHubRemoteContextErrorCode = "unauthorized"
	GitHubRemoteContextErrorForbidden          GitHubRemoteContextErrorCode = "forbidden"
	GitHubRemoteContextErrorNotFound           GitHubRemoteContextErrorCode = "not_found"
	GitHubRemoteContextErrorInvalidToken       GitHubRemoteContextErrorCode = "invalid_token"
	GitHubRemoteContextErrorOversizedResponse  GitHubRemoteContextErrorCode = "oversized_response"
	GitHubRemoteContextErrorUnsupportedContent GitHubRemoteContextErrorCode = "unsupported_content"
	GitHubRemoteContextErrorTooManyCandidates  GitHubRemoteContextErrorCode = "too_many_candidates"
	GitHubRemoteContextErrorInvalidResponse    GitHubRemoteContextErrorCode = "invalid_response"
)

type GitHubRemoteContextError struct {
	Code    GitHubRemoteContextErrorCode
	Message string
}

func (err GitHubRemoteContextError) Error() string {
	if strings.TrimSpace(err.Message) == "" {
		return string(err.Code)
	}
	return err.Message
}

type GitHubRemoteContextLimits struct {
	MaxRepositoryChars      int
	MaxRefChars             int
	MaxQueryChars           int
	MaxQueryTerms           int
	MaxPathFilters          int
	MaxPathChars            int
	MaxFilesFetched         int
	MaxFileSizeBytes        int64
	MaxTotalFileBytes       int64
	MaxSnippetsPerFile      int
	MaxSnippetChars         int
	MaxResults              int
	MaxRenderedContentChars int
	MaxTreeEntriesScanned   int
	MaxDirectoryDepth       int
	MaxSkippedRecords       int
}

type GitHubRepositoryLocator struct {
	owner string
	name  string
}

type GitHubRemoteRef struct {
	value    string
	provided bool
}

type GitHubRemotePathFilter struct {
	path string
}

type GitHubRemoteContextQuery struct {
	DisplayQuery     string
	NormalizedPhrase string
	Terms            []string
	SortedTerms      []string
}

type GitHubRemoteRepository struct {
	Locator       GitHubRepositoryLocator
	DefaultBranch string
}

type GitHubRemoteResolvedRef struct {
	RequestedRef  string
	DefaultBranch string
	ResolvedRef   string
	CommitSHA     string
}

type GitHubRemoteEntry struct {
	Path      string
	Type      string
	SizeBytes int64
}

type GitHubRemoteFile struct {
	Path      string
	SizeBytes int64
	Contents  []byte
}

type GitHubRemoteSourceSpec struct {
	Path                   string
	Directory              bool
	Extensions             []string
	SourceCategory         ContextSourceCategory
	FileType               RepositoryContextIndexFileType
	LanguageOrEcosystem    string
	SourceEvidenceCategory string
	MaxDepth               int
}

type GitHubRemoteCandidate struct {
	Path                   string
	SourceCategory         ContextSourceCategory
	FileType               RepositoryContextIndexFileType
	LanguageOrEcosystem    string
	SizeBytes              int64
	SourceEvidenceCategory string
}

type GitHubRemoteSnippet struct {
	LineStart int
	LineEnd   int
	Text      string
}

type GitHubRemoteContextResult struct {
	Rank                   int
	Repository             string
	RequestedRef           string
	DefaultBranch          string
	ResolvedRef            string
	CommitSHA              string
	Path                   string
	SourceCategory         ContextSourceCategory
	SourceEvidenceCategory string
	Score                  int
	CategoryPriority       int
	Snippet                GitHubRemoteSnippet
	Summary                string
	Remote                 bool
}

type GitHubRemoteContextSkipped struct {
	Path   string
	Reason string
}

type GitHubRemoteContextReport struct {
	Status          GitHubRemoteContextStatus
	Repository      string
	RequestedRef    string
	DefaultBranch   string
	ResolvedRef     string
	CommitSHA       string
	Query           GitHubRemoteContextQuery
	PathFilters     []string
	Results         []GitHubRemoteContextResult
	Skipped         []GitHubRemoteContextSkipped
	Message         string
	OutputTruncated bool
}

type GitHubRemoteContextScore struct {
	Value            int
	CategoryPriority int
	Matched          bool
}

func DefaultGitHubRemoteContextLimits() GitHubRemoteContextLimits {
	return GitHubRemoteContextLimits{
		MaxRepositoryChars:      200,
		MaxRefChars:             200,
		MaxQueryChars:           512,
		MaxQueryTerms:           32,
		MaxPathFilters:          20,
		MaxPathChars:            512,
		MaxFilesFetched:         50,
		MaxFileSizeBytes:        128 * 1024,
		MaxTotalFileBytes:       1024 * 1024,
		MaxSnippetsPerFile:      2,
		MaxSnippetChars:         600,
		MaxResults:              10,
		MaxRenderedContentChars: 8000,
		MaxTreeEntriesScanned:   500,
		MaxDirectoryDepth:       4,
		MaxSkippedRecords:       50,
	}
}

func NewGitHubRepositoryLocator(raw string, limits GitHubRemoteContextLimits) (GitHubRepositoryLocator, error) {
	if err := limits.Validate(); err != nil {
		return GitHubRepositoryLocator{}, err
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return GitHubRepositoryLocator{}, errors.New("github repository is required")
	}
	if len(value) > limits.MaxRepositoryChars {
		return GitHubRepositoryLocator{}, fmt.Errorf("github repository must be at most %d characters", limits.MaxRepositoryChars)
	}
	if containsControlOrWhitespace(value) {
		return GitHubRepositoryLocator{}, errors.New("github repository must not contain whitespace or control characters")
	}
	if strings.ContainsRune(value, 0) {
		return GitHubRepositoryLocator{}, errors.New("github repository contains a null byte")
	}
	if strings.Contains(value, "?") || strings.Contains(value, "#") {
		return GitHubRepositoryLocator{}, errors.New("github repository must not include query strings or fragments")
	}
	if strings.Contains(value, "@") {
		return GitHubRepositoryLocator{}, errors.New("github repository must not include credentials")
	}
	if strings.HasPrefix(value, "/") || isRepositoryContextIndexWindowsDrivePath(value) {
		return GitHubRepositoryLocator{}, errors.New("github repository must not be a filesystem path")
	}

	if strings.HasPrefix(value, "https://") {
		const prefix = "https://github.com/"
		if !strings.HasPrefix(value, prefix) {
			return GitHubRepositoryLocator{}, errors.New("github repository URL host must be github.com")
		}
		value = strings.TrimPrefix(value, prefix)
	} else if strings.Contains(value, "://") {
		return GitHubRepositoryLocator{}, errors.New("github repository URL must use https://github.com")
	}

	value = strings.Trim(value, "/")
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return GitHubRepositoryLocator{}, errors.New("github repository must use owner/name")
	}
	if parts[0] == "." || parts[1] == "." || parts[0] == ".." || parts[1] == ".." {
		return GitHubRepositoryLocator{}, errors.New("github repository must not contain traversal")
	}
	for _, part := range parts {
		if !isGitHubLocatorSegment(part) {
			return GitHubRepositoryLocator{}, fmt.Errorf("github repository segment contains unsupported characters: %s", part)
		}
	}
	return GitHubRepositoryLocator{owner: parts[0], name: parts[1]}, nil
}

func NewGitHubRemoteRef(raw string, limits GitHubRemoteContextLimits) (GitHubRemoteRef, error) {
	if err := limits.Validate(); err != nil {
		return GitHubRemoteRef{}, err
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return GitHubRemoteRef{}, errors.New("github ref is required")
	}
	if len(value) > limits.MaxRefChars {
		return GitHubRemoteRef{}, fmt.Errorf("github ref must be at most %d characters", limits.MaxRefChars)
	}
	if strings.ContainsRune(value, 0) {
		return GitHubRemoteRef{}, errors.New("github ref contains a null byte")
	}
	if containsControlOrWhitespace(value) {
		return GitHubRemoteRef{}, errors.New("github ref must not contain whitespace or control characters")
	}
	if strings.Contains(value, "://") || strings.Contains(value, "@") {
		return GitHubRemoteRef{}, errors.New("github ref must not include URLs or credentials")
	}
	if strings.Contains(value, "?") || strings.Contains(value, "#") {
		return GitHubRemoteRef{}, errors.New("github ref must not include query strings or fragments")
	}
	normalized := strings.ReplaceAll(value, "\\", "/")
	if strings.HasPrefix(normalized, "/") || strings.HasSuffix(normalized, "/") || strings.Contains(normalized, "//") {
		return GitHubRemoteRef{}, errors.New("github ref contains invalid slashes")
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == "." || segment == ".." {
			return GitHubRemoteRef{}, errors.New("github ref must not contain traversal")
		}
	}
	return GitHubRemoteRef{value: normalized, provided: true}, nil
}

func DefaultGitHubRemoteRef() GitHubRemoteRef {
	return GitHubRemoteRef{}
}

func NewGitHubRemotePathFilter(raw string, limits GitHubRemoteContextLimits) (GitHubRemotePathFilter, error) {
	if err := limits.Validate(); err != nil {
		return GitHubRemotePathFilter{}, err
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return GitHubRemotePathFilter{}, errors.New("github path filter is required")
	}
	if len(value) > limits.MaxPathChars {
		return GitHubRemotePathFilter{}, fmt.Errorf("github path filter must be at most %d characters", limits.MaxPathChars)
	}
	if strings.Contains(value, "?") || strings.Contains(value, "#") {
		return GitHubRemotePathFilter{}, errors.New("github path filter must not include query strings or fragments")
	}
	if strings.ContainsAny(value, "*?[]{}") {
		return GitHubRemotePathFilter{}, errors.New("github path filter must not include wildcards")
	}
	path, err := SafeGitHubRemotePath(value)
	if err != nil {
		return GitHubRemotePathFilter{}, err
	}
	if _, skip := ShouldSkipGitHubRemotePath(path); skip {
		return GitHubRemotePathFilter{}, fmt.Errorf("github path filter targets a skipped path: %s", path)
	}
	return GitHubRemotePathFilter{path: path}, nil
}

func NewGitHubRemoteContextQuery(rawQuery string, limits GitHubRemoteContextLimits) (GitHubRemoteContextQuery, error) {
	if err := limits.Validate(); err != nil {
		return GitHubRemoteContextQuery{}, err
	}
	display := strings.TrimSpace(rawQuery)
	if display == "" {
		return GitHubRemoteContextQuery{}, errors.New("github context query is required")
	}
	if len(display) > limits.MaxQueryChars {
		return GitHubRemoteContextQuery{}, fmt.Errorf("github context query must be at most %d characters", limits.MaxQueryChars)
	}
	terms := deduplicateLocalContextTerms(tokenizeLocalContextText(display))
	if len(terms) == 0 {
		return GitHubRemoteContextQuery{}, errors.New("github context query must contain at least one letter or digit term")
	}
	if len(terms) > limits.MaxQueryTerms {
		terms = terms[:limits.MaxQueryTerms]
	}
	sortedTerms := append([]string(nil), terms...)
	sort.Strings(sortedTerms)
	return GitHubRemoteContextQuery{
		DisplayQuery:     display,
		NormalizedPhrase: strings.Join(terms, " "),
		Terms:            append([]string(nil), terms...),
		SortedTerms:      sortedTerms,
	}, nil
}

func NewGitHubRemoteCandidate(input GitHubRemoteCandidate) (GitHubRemoteCandidate, error) {
	relativePath, err := SafeGitHubRemotePath(input.Path)
	if err != nil {
		return GitHubRemoteCandidate{}, err
	}
	if !input.SourceCategory.IsSupported() {
		return GitHubRemoteCandidate{}, fmt.Errorf("unsupported github remote source category: %s", input.SourceCategory)
	}
	if !input.FileType.IsSupported() {
		return GitHubRemoteCandidate{}, fmt.Errorf("unsupported github remote file type: %s", input.FileType)
	}
	if input.SizeBytes < 0 {
		return GitHubRemoteCandidate{}, errors.New("github remote candidate size must not be negative")
	}
	evidence := strings.TrimSpace(input.SourceEvidenceCategory)
	if evidence == "" {
		return GitHubRemoteCandidate{}, errors.New("github remote source evidence category is required")
	}
	return GitHubRemoteCandidate{
		Path:                   relativePath,
		SourceCategory:         input.SourceCategory,
		FileType:               input.FileType,
		LanguageOrEcosystem:    strings.TrimSpace(input.LanguageOrEcosystem),
		SizeBytes:              input.SizeBytes,
		SourceEvidenceCategory: evidence,
	}, nil
}

func NewGitHubRemoteContextResult(
	repository GitHubRepositoryLocator,
	ref GitHubRemoteResolvedRef,
	candidate GitHubRemoteCandidate,
	score GitHubRemoteContextScore,
	snippet GitHubRemoteSnippet,
	summary string,
) (GitHubRemoteContextResult, error) {
	if score.Value <= 0 {
		return GitHubRemoteContextResult{}, errors.New("github remote context score must be positive")
	}
	if snippet.Text == "" && strings.TrimSpace(summary) == "" {
		return GitHubRemoteContextResult{}, errors.New("github remote context result requires a snippet or summary")
	}
	if snippet.Text != "" && (snippet.LineStart <= 0 || snippet.LineEnd < snippet.LineStart) {
		return GitHubRemoteContextResult{}, errors.New("github remote context snippet line range is invalid")
	}
	return GitHubRemoteContextResult{
		Repository:             repository.String(),
		RequestedRef:           ref.RequestedRef,
		DefaultBranch:          ref.DefaultBranch,
		ResolvedRef:            ref.ResolvedRef,
		CommitSHA:              ref.CommitSHA,
		Path:                   candidate.Path,
		SourceCategory:         candidate.SourceCategory,
		SourceEvidenceCategory: candidate.SourceEvidenceCategory,
		Score:                  score.Value,
		CategoryPriority:       score.CategoryPriority,
		Snippet:                snippet,
		Summary:                strings.TrimSpace(summary),
		Remote:                 true,
	}, nil
}

func NewGitHubRemoteContextReport(
	status GitHubRemoteContextStatus,
	repository GitHubRepositoryLocator,
	ref GitHubRemoteResolvedRef,
	query GitHubRemoteContextQuery,
	filters []GitHubRemotePathFilter,
	results []GitHubRemoteContextResult,
	skipped []GitHubRemoteContextSkipped,
	message string,
	outputTruncated bool,
) GitHubRemoteContextReport {
	copiedResults := append([]GitHubRemoteContextResult(nil), results...)
	SortGitHubRemoteContextResults(copiedResults)
	for index := range copiedResults {
		copiedResults[index].Rank = index + 1
	}
	filterValues := make([]string, 0, len(filters))
	for _, filter := range filters {
		filterValues = append(filterValues, filter.String())
	}
	copiedSkipped := append([]GitHubRemoteContextSkipped(nil), skipped...)
	sort.SliceStable(copiedSkipped, func(i, j int) bool {
		if copiedSkipped[i].Path != copiedSkipped[j].Path {
			return copiedSkipped[i].Path < copiedSkipped[j].Path
		}
		return copiedSkipped[i].Reason < copiedSkipped[j].Reason
	})
	return GitHubRemoteContextReport{
		Status:          status,
		Repository:      repository.String(),
		RequestedRef:    ref.RequestedRef,
		DefaultBranch:   ref.DefaultBranch,
		ResolvedRef:     ref.ResolvedRef,
		CommitSHA:       ref.CommitSHA,
		Query:           query,
		PathFilters:     filterValues,
		Results:         copiedResults,
		Skipped:         copiedSkipped,
		Message:         strings.TrimSpace(message),
		OutputTruncated: outputTruncated,
	}
}

func ScoreGitHubRemoteContextCandidate(
	query GitHubRemoteContextQuery,
	candidate GitHubRemoteCandidate,
	snippetText string,
) GitHubRemoteContextScore {
	content := normalizeLocalContextMatchText(snippetText)
	entryPath := normalizeLocalContextMatchText(candidate.Path)
	filename := normalizeLocalContextMatchText(localContextPathBase(candidate.Path))
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
	categoryPriority := LocalContextRetrievalSourceCategoryPriority(candidate.SourceCategory)
	if matched {
		score += categoryPriority
	}
	return GitHubRemoteContextScore{
		Value:            score,
		CategoryPriority: categoryPriority,
		Matched:          matched,
	}
}

func SortGitHubRemoteContextResults(results []GitHubRemoteContextResult) {
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

func DefaultGitHubRemoteSourceSpecs() []GitHubRemoteSourceSpec {
	specs := []GitHubRemoteSourceSpec{
		githubRemoteFileSource("README.md", ContextSourceCategoryReadme, RepositoryContextIndexFileTypeMarkdown, "Markdown", "readme"),
		githubRemoteFileSource("AGENTS.md", ContextSourceCategoryAgentInstruction, RepositoryContextIndexFileTypeMarkdown, "Markdown", "agent_instruction"),
		githubRemoteFileSource("CONTRIBUTING.md", ContextSourceCategoryContributing, RepositoryContextIndexFileTypeMarkdown, "Markdown", "contributing"),
		githubRemoteDirectorySource("docs", []string{".md"}, ContextSourceCategoryDocumentation, RepositoryContextIndexFileTypeMarkdown, "Markdown", "documentation"),
		githubRemoteFileSource("openspec/project.md", ContextSourceCategoryOpenSpecProject, RepositoryContextIndexFileTypeMarkdown, "OpenSpec", "openspec_project"),
		githubRemoteDirectorySource("openspec/specs", []string{".md"}, ContextSourceCategoryOpenSpecSpec, RepositoryContextIndexFileTypeMarkdown, "OpenSpec", "openspec_spec"),
		githubRemoteDirectorySource(".specharbor/rules", []string{".md"}, ContextSourceCategorySpecHarborRules, RepositoryContextIndexFileTypeMarkdown, "SpecHarbor", "specharbor_rules"),
		githubRemoteFileSource(".specharbor/project-brief.md", ContextSourceCategoryProjectBrief, RepositoryContextIndexFileTypeMarkdown, "SpecHarbor", "project_brief"),
		githubRemoteFileSource("package.json", ContextSourceCategoryPackageManifest, RepositoryContextIndexFileTypeJSON, "Node.js", "package_manifest"),
		githubRemoteFileSource("go.mod", ContextSourceCategoryPackageManifest, RepositoryContextIndexFileTypeGoModule, "Go", "package_manifest"),
		githubRemoteFileSource("pom.xml", ContextSourceCategoryBuildManifest, RepositoryContextIndexFileTypeXML, "JVM", "build_manifest"),
		githubRemoteFileSource("build.gradle", ContextSourceCategoryBuildManifest, RepositoryContextIndexFileTypeGradle, "JVM", "build_manifest"),
		githubRemoteFileSource("build.gradle.kts", ContextSourceCategoryBuildManifest, RepositoryContextIndexFileTypeGradle, "JVM", "build_manifest"),
		githubRemoteFileSource("Cargo.toml", ContextSourceCategoryPackageManifest, RepositoryContextIndexFileTypeTOML, "Rust", "package_manifest"),
		githubRemoteFileSource("pyproject.toml", ContextSourceCategoryBuildManifest, RepositoryContextIndexFileTypeTOML, "Python", "build_manifest"),
		githubRemoteFileSource("requirements.txt", ContextSourceCategoryDependencyManifest, RepositoryContextIndexFileTypeText, "Python", "dependency_manifest"),
		githubRemoteFileSource("Dockerfile", ContextSourceCategoryContainerConfig, RepositoryContextIndexFileTypeDockerfile, "Docker", "container_config"),
		githubRemoteFileSource("docker-compose.yml", ContextSourceCategoryContainerConfig, RepositoryContextIndexFileTypeYAML, "Docker", "container_config"),
		githubRemoteFileSource("docker-compose.yaml", ContextSourceCategoryContainerConfig, RepositoryContextIndexFileTypeYAML, "Docker", "container_config"),
		githubRemoteFileSource("Makefile", ContextSourceCategoryTaskRunner, RepositoryContextIndexFileTypeMakefile, "Make", "task_runner"),
		githubRemoteFileSource("Taskfile.yml", ContextSourceCategoryTaskRunner, RepositoryContextIndexFileTypeYAML, "Task", "task_runner"),
		githubRemoteFileSource("Taskfile.yaml", ContextSourceCategoryTaskRunner, RepositoryContextIndexFileTypeYAML, "Task", "task_runner"),
		githubRemoteDirectorySource(".github/workflows", []string{".yml", ".yaml"}, ContextSourceCategoryWorkflowConfig, RepositoryContextIndexFileTypeYAML, "GitHub Actions", "workflow_config"),
	}
	return append([]GitHubRemoteSourceSpec(nil), specs...)
}

func SafeGitHubRemotePath(relativePath string) (string, error) {
	trimmed := strings.TrimSpace(relativePath)
	if trimmed == "" {
		return "", errors.New("github remote path is required")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", errors.New("github remote path contains a null byte")
	}
	if strings.Contains(trimmed, "?") || strings.Contains(trimmed, "#") {
		return "", errors.New("github remote path must not include query strings or fragments")
	}
	normalized := normalizeContextPath(trimmed)
	if strings.HasPrefix(normalized, "/") || isRepositoryContextIndexWindowsDrivePath(normalized) {
		return "", fmt.Errorf("github remote path must be relative: %s", relativePath)
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return "", fmt.Errorf("github remote path must not contain traversal: %s", relativePath)
		}
	}
	cleaned := cleanRepositoryContextIndexPath(normalized)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("github remote path must be a safe relative path: %s", relativePath)
	}
	return cleaned, nil
}

func ShouldSkipGitHubRemotePath(relativePath string) (string, bool) {
	normalized := normalizeContextPath(relativePath)
	if normalized == "" || normalized == "." {
		return "", false
	}
	segments := strings.Split(normalized, "/")
	for _, segment := range segments[:len(segments)-1] {
		if isGeneratedContextDirectoryName(segment) {
			return "generated_directory", true
		}
	}
	base := segments[len(segments)-1]
	if isSensitiveGitHubRemoteFileName(base) {
		return "sensitive_file", true
	}
	if isGeneratedContextDirectoryName(base) {
		return "generated_directory", true
	}
	return "", false
}

func (limits GitHubRemoteContextLimits) Validate() error {
	if limits.MaxRepositoryChars <= 0 {
		return errors.New("github remote max repository chars must be positive")
	}
	if limits.MaxRefChars <= 0 {
		return errors.New("github remote max ref chars must be positive")
	}
	if limits.MaxQueryChars <= 0 || limits.MaxQueryTerms <= 0 {
		return errors.New("github remote query limits must be positive")
	}
	if limits.MaxPathFilters < 0 || limits.MaxPathChars <= 0 {
		return errors.New("github remote path limits are invalid")
	}
	if limits.MaxFilesFetched <= 0 || limits.MaxFileSizeBytes <= 0 || limits.MaxTotalFileBytes <= 0 {
		return errors.New("github remote file limits must be positive")
	}
	if limits.MaxSnippetsPerFile <= 0 || limits.MaxSnippetChars <= 0 || limits.MaxResults <= 0 {
		return errors.New("github remote result limits must be positive")
	}
	if limits.MaxRenderedContentChars <= 0 || limits.MaxTreeEntriesScanned <= 0 || limits.MaxDirectoryDepth < 0 {
		return errors.New("github remote traversal limits are invalid")
	}
	if limits.MaxSkippedRecords < 0 {
		return errors.New("github remote max skipped records must not be negative")
	}
	return nil
}

func (locator GitHubRepositoryLocator) Owner() string {
	return locator.owner
}

func (locator GitHubRepositoryLocator) Name() string {
	return locator.name
}

func (locator GitHubRepositoryLocator) String() string {
	if locator.owner == "" || locator.name == "" {
		return ""
	}
	return locator.owner + "/" + locator.name
}

func (ref GitHubRemoteRef) Value() string {
	return ref.value
}

func (ref GitHubRemoteRef) Provided() bool {
	return ref.provided
}

func (filter GitHubRemotePathFilter) String() string {
	return filter.path
}

func githubRemoteFileSource(
	relativePath string,
	category ContextSourceCategory,
	fileType RepositoryContextIndexFileType,
	language string,
	evidence string,
) GitHubRemoteSourceSpec {
	return GitHubRemoteSourceSpec{
		Path:                   relativePath,
		SourceCategory:         category,
		FileType:               fileType,
		LanguageOrEcosystem:    language,
		SourceEvidenceCategory: evidence,
	}
}

func githubRemoteDirectorySource(
	relativePath string,
	extensions []string,
	category ContextSourceCategory,
	fileType RepositoryContextIndexFileType,
	language string,
	evidence string,
) GitHubRemoteSourceSpec {
	return GitHubRemoteSourceSpec{
		Path:                   relativePath,
		Directory:              true,
		Extensions:             append([]string(nil), extensions...),
		SourceCategory:         category,
		FileType:               fileType,
		LanguageOrEcosystem:    language,
		SourceEvidenceCategory: evidence,
		MaxDepth:               DefaultGitHubRemoteContextLimits().MaxDirectoryDepth,
	}
}

func isGitHubLocatorSegment(value string) bool {
	for _, char := range value {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func containsControlOrWhitespace(value string) bool {
	for _, char := range value {
		if char <= 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}

func isSensitiveGitHubRemoteFileName(name string) bool {
	if isSensitiveContextFileName(name) {
		return true
	}
	switch name {
	case ".npmrc", ".pypirc", ".netrc":
		return true
	default:
		return false
	}
}
