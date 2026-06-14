package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ContextRAGOpenAIAPIKeyEnvVar = "SPECHARBOR_OPENAI_API_KEY"
	ContextRAGOpenAIModelEnvVar  = "SPECHARBOR_OPENAI_MODEL"
	DefaultContextRAGOpenAIModel = "gpt-5.4-mini"
)

type ContextRAGProviderName string

const (
	ContextRAGProviderOpenAI ContextRAGProviderName = "openai"
)

type ContextRAGSourceKind string

const (
	ContextRAGSourceLocal  ContextRAGSourceKind = "local"
	ContextRAGSourceGitHub ContextRAGSourceKind = "github"
)

type ContextRAGStatus string

const (
	ContextRAGStatusAnswered                  ContextRAGStatus = "answered"
	ContextRAGStatusInsufficientSources       ContextRAGStatus = "insufficient_sources"
	ContextRAGStatusMissingSources            ContextRAGStatus = "missing_sources"
	ContextRAGStatusMissingCredentials        ContextRAGStatus = "missing_credentials"
	ContextRAGStatusProviderFailed            ContextRAGStatus = "provider_failed"
	ContextRAGStatusProviderTimeout           ContextRAGStatus = "provider_timeout"
	ContextRAGStatusProviderRateLimited       ContextRAGStatus = "provider_rate_limited"
	ContextRAGStatusProviderUnauthorized      ContextRAGStatus = "provider_unauthorized"
	ContextRAGStatusProviderResponseInvalid   ContextRAGStatus = "provider_response_invalid"
	ContextRAGStatusProviderResponseOversized ContextRAGStatus = "provider_response_oversized"
)

type ContextRAGProviderErrorCode string

const (
	ContextRAGProviderErrorMissingCredentials ContextRAGProviderErrorCode = "missing_credentials"
	ContextRAGProviderErrorUnauthorized       ContextRAGProviderErrorCode = "unauthorized"
	ContextRAGProviderErrorRateLimited        ContextRAGProviderErrorCode = "rate_limited"
	ContextRAGProviderErrorTimeout            ContextRAGProviderErrorCode = "timeout"
	ContextRAGProviderErrorNetwork            ContextRAGProviderErrorCode = "network"
	ContextRAGProviderErrorInvalidResponse    ContextRAGProviderErrorCode = "invalid_response"
	ContextRAGProviderErrorOversizedResponse  ContextRAGProviderErrorCode = "oversized_response"
	ContextRAGProviderErrorFailed             ContextRAGProviderErrorCode = "failed"
)

type ContextRAGProviderError struct {
	Code    ContextRAGProviderErrorCode
	Message string
}

func (err ContextRAGProviderError) Error() string {
	if strings.TrimSpace(err.Message) == "" {
		return string(err.Code)
	}
	return err.Message
}

type ContextRAGLimits struct {
	MaxQueryChars            int
	MaxSources               int
	HardMaxSources           int
	MaxSnippetChars          int
	MaxTotalContextChars     int
	MaxAnswerChars           int
	HardMaxAnswerChars       int
	MaxProviderResponseBytes int64
	ProviderTimeoutSeconds   int
	MaxRenderedOutputChars   int
}

type ContextRAGQuery struct {
	DisplayQuery string
}

type ContextRAGSourceInput struct {
	Kind                   ContextRAGSourceKind
	Repository             string
	RequestedRef           string
	DefaultBranch          string
	ResolvedRef            string
	CommitSHA              string
	Path                   string
	SourceCategory         ContextSourceCategory
	SourceEvidenceCategory string
	LineStart              int
	LineEnd                int
	Score                  int
	Text                   string
	SelectionOrder         int
	SourceRank             int
}

type ContextRAGSource struct {
	ID                     int
	Kind                   ContextRAGSourceKind
	Repository             string
	RequestedRef           string
	DefaultBranch          string
	ResolvedRef            string
	CommitSHA              string
	Path                   string
	SourceCategory         ContextSourceCategory
	SourceEvidenceCategory string
	LineStart              int
	LineEnd                int
	Score                  int
	Text                   string
	Truncated              bool
	SelectionOrder         int
	SourceRank             int
}

type ContextRAGProviderRequest struct {
	Provider       ContextRAGProviderName
	Model          string
	Query          ContextRAGQuery
	Instructions   string
	SourceContext  string
	Sources        []ContextRAGSource
	MaxAnswerChars int
}

type ContextRAGProviderResponse struct {
	Provider        ContextRAGProviderName
	Model           string
	Answer          string
	Insufficient    bool
	OutputTruncated bool
	FinishReason    string
}

type ContextRAGReport struct {
	Status          ContextRAGStatus
	Provider        ContextRAGProviderName
	Model           string
	Query           ContextRAGQuery
	Answer          string
	Sources         []ContextRAGSource
	Message         string
	OutputTruncated bool
}

func DefaultContextRAGLimits() ContextRAGLimits {
	return ContextRAGLimits{
		MaxQueryChars:            512,
		MaxSources:               8,
		HardMaxSources:           20,
		MaxSnippetChars:          600,
		MaxTotalContextChars:     6000,
		MaxAnswerChars:           4000,
		HardMaxAnswerChars:       8000,
		MaxProviderResponseBytes: 65536,
		ProviderTimeoutSeconds:   20,
		MaxRenderedOutputChars:   12000,
	}
}

func NewContextRAGProviderName(raw string) (ContextRAGProviderName, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "", errors.New("context rag provider is required")
	}
	switch ContextRAGProviderName(value) {
	case ContextRAGProviderOpenAI:
		return ContextRAGProviderOpenAI, nil
	default:
		return "", fmt.Errorf("unsupported context rag provider: %s", value)
	}
}

func NewContextRAGSourceKinds(rawSources []string) ([]ContextRAGSourceKind, error) {
	if len(rawSources) == 0 {
		return []ContextRAGSourceKind{ContextRAGSourceLocal}, nil
	}
	seen := make(map[ContextRAGSourceKind]bool, len(rawSources))
	sources := make([]ContextRAGSourceKind, 0, len(rawSources))
	for _, raw := range rawSources {
		value := strings.ToLower(strings.TrimSpace(raw))
		if value == "" {
			return nil, errors.New("context rag source is required")
		}
		var source ContextRAGSourceKind
		switch ContextRAGSourceKind(value) {
		case ContextRAGSourceLocal:
			source = ContextRAGSourceLocal
		case ContextRAGSourceGitHub:
			source = ContextRAGSourceGitHub
		default:
			return nil, fmt.Errorf("unsupported context rag source: %s", value)
		}
		if seen[source] {
			continue
		}
		seen[source] = true
		sources = append(sources, source)
	}
	return sources, nil
}

func NewContextRAGQuery(rawQuery string, limits ContextRAGLimits) (ContextRAGQuery, error) {
	if err := limits.Validate(); err != nil {
		return ContextRAGQuery{}, err
	}
	display := strings.TrimSpace(rawQuery)
	if display == "" {
		return ContextRAGQuery{}, errors.New("context rag query is required")
	}
	if len(display) > limits.MaxQueryChars {
		return ContextRAGQuery{}, fmt.Errorf("context rag query must be at most %d characters", limits.MaxQueryChars)
	}
	return ContextRAGQuery{DisplayQuery: display}, nil
}

func NewContextRAGSource(input ContextRAGSourceInput, limits ContextRAGLimits) (ContextRAGSource, error) {
	if err := limits.Validate(); err != nil {
		return ContextRAGSource{}, err
	}
	if input.Kind != ContextRAGSourceLocal && input.Kind != ContextRAGSourceGitHub {
		return ContextRAGSource{}, fmt.Errorf("unsupported context rag source kind: %s", input.Kind)
	}
	relativePath, err := safeContextRAGSourcePath(input.Kind, input.Path)
	if err != nil {
		return ContextRAGSource{}, err
	}
	if input.SourceCategory != "" && !input.SourceCategory.IsSupported() {
		return ContextRAGSource{}, fmt.Errorf("unsupported context rag source category: %s", input.SourceCategory)
	}
	if input.LineStart < 0 || input.LineEnd < 0 || (input.LineStart == 0 && input.LineEnd > 0) ||
		(input.LineStart > 0 && input.LineEnd < input.LineStart) {
		return ContextRAGSource{}, errors.New("context rag source line range is invalid")
	}
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return ContextRAGSource{}, errors.New("context rag source text is required")
	}
	truncated := false
	if len(text) > limits.MaxSnippetChars {
		text = TrimLocalContextRetrievalSnippet(text, limits.MaxSnippetChars)
		truncated = true
	}
	return ContextRAGSource{
		Kind:                   input.Kind,
		Repository:             strings.TrimSpace(input.Repository),
		RequestedRef:           strings.TrimSpace(input.RequestedRef),
		DefaultBranch:          strings.TrimSpace(input.DefaultBranch),
		ResolvedRef:            strings.TrimSpace(input.ResolvedRef),
		CommitSHA:              strings.TrimSpace(input.CommitSHA),
		Path:                   relativePath,
		SourceCategory:         input.SourceCategory,
		SourceEvidenceCategory: strings.TrimSpace(input.SourceEvidenceCategory),
		LineStart:              input.LineStart,
		LineEnd:                input.LineEnd,
		Score:                  input.Score,
		Text:                   text,
		Truncated:              truncated,
		SelectionOrder:         input.SelectionOrder,
		SourceRank:             input.SourceRank,
	}, nil
}

func NewContextRAGProviderRequest(
	provider ContextRAGProviderName,
	model string,
	query ContextRAGQuery,
	sources []ContextRAGSource,
	limits ContextRAGLimits,
) (ContextRAGProviderRequest, bool, error) {
	if err := limits.Validate(); err != nil {
		return ContextRAGProviderRequest{}, false, err
	}
	if provider != ContextRAGProviderOpenAI {
		return ContextRAGProviderRequest{}, false, fmt.Errorf("unsupported context rag provider: %s", provider)
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return ContextRAGProviderRequest{}, false, errors.New("context rag provider model is required")
	}
	if strings.TrimSpace(query.DisplayQuery) == "" {
		return ContextRAGProviderRequest{}, false, errors.New("context rag query is required")
	}
	prepared, truncated := PrepareContextRAGSources(sources, limits)
	if len(prepared) == 0 {
		return ContextRAGProviderRequest{}, truncated, errors.New("context rag sources are required")
	}
	return ContextRAGProviderRequest{
		Provider:       provider,
		Model:          model,
		Query:          query,
		Instructions:   ContextRAGInstructions(),
		SourceContext:  RenderContextRAGSourceContext(prepared),
		Sources:        prepared,
		MaxAnswerChars: limits.MaxAnswerChars,
	}, truncated, nil
}

func NewContextRAGProviderResponse(
	provider ContextRAGProviderName,
	model string,
	answer string,
	insufficient bool,
	finishReason string,
	limits ContextRAGLimits,
) (ContextRAGProviderResponse, error) {
	if err := limits.Validate(); err != nil {
		return ContextRAGProviderResponse{}, err
	}
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" {
		return ContextRAGProviderResponse{}, ContextRAGProviderError{
			Code:    ContextRAGProviderErrorInvalidResponse,
			Message: "provider response did not include answer text",
		}
	}
	truncated := false
	if len(trimmed) > limits.MaxAnswerChars {
		trimmed = TrimLocalContextRetrievalSnippet(trimmed, limits.MaxAnswerChars)
		truncated = true
	}
	return ContextRAGProviderResponse{
		Provider:        provider,
		Model:           strings.TrimSpace(model),
		Answer:          trimmed,
		Insufficient:    insufficient,
		OutputTruncated: truncated,
		FinishReason:    strings.TrimSpace(finishReason),
	}, nil
}

func NewContextRAGReport(
	status ContextRAGStatus,
	provider ContextRAGProviderName,
	model string,
	query ContextRAGQuery,
	answer string,
	sources []ContextRAGSource,
	message string,
	outputTruncated bool,
) ContextRAGReport {
	copiedSources := append([]ContextRAGSource(nil), sources...)
	SortContextRAGSources(copiedSources)
	for index := range copiedSources {
		copiedSources[index].ID = index + 1
	}
	return ContextRAGReport{
		Status:          status,
		Provider:        provider,
		Model:           strings.TrimSpace(model),
		Query:           query,
		Answer:          strings.TrimSpace(answer),
		Sources:         copiedSources,
		Message:         strings.TrimSpace(message),
		OutputTruncated: outputTruncated,
	}
}

func (limits ContextRAGLimits) Validate() error {
	if limits.MaxQueryChars <= 0 {
		return errors.New("context rag max query chars must be positive")
	}
	if limits.MaxSources <= 0 || limits.HardMaxSources <= 0 || limits.MaxSources > limits.HardMaxSources {
		return errors.New("context rag source limits are invalid")
	}
	if limits.MaxSnippetChars <= 0 || limits.MaxTotalContextChars <= 0 {
		return errors.New("context rag context limits must be positive")
	}
	if limits.MaxAnswerChars <= 0 || limits.HardMaxAnswerChars <= 0 || limits.MaxAnswerChars > limits.HardMaxAnswerChars {
		return errors.New("context rag answer limits are invalid")
	}
	if limits.MaxProviderResponseBytes <= 0 {
		return errors.New("context rag provider response bytes must be positive")
	}
	if limits.ProviderTimeoutSeconds <= 0 {
		return errors.New("context rag provider timeout must be positive")
	}
	if limits.MaxRenderedOutputChars <= 0 {
		return errors.New("context rag rendered output chars must be positive")
	}
	return nil
}

func (limits ContextRAGLimits) WithMaxSources(maxSources int) (ContextRAGLimits, error) {
	if maxSources <= 0 {
		return ContextRAGLimits{}, errors.New("context rag max sources must be positive")
	}
	if maxSources > limits.HardMaxSources {
		return ContextRAGLimits{}, fmt.Errorf("context rag max sources must be at most %d", limits.HardMaxSources)
	}
	limits.MaxSources = maxSources
	return limits, limits.Validate()
}

func (limits ContextRAGLimits) WithMaxAnswerChars(maxAnswerChars int) (ContextRAGLimits, error) {
	if maxAnswerChars <= 0 {
		return ContextRAGLimits{}, errors.New("context rag max answer chars must be positive")
	}
	if maxAnswerChars > limits.HardMaxAnswerChars {
		return ContextRAGLimits{}, fmt.Errorf("context rag max answer chars must be at most %d", limits.HardMaxAnswerChars)
	}
	limits.MaxAnswerChars = maxAnswerChars
	return limits, limits.Validate()
}

func PrepareContextRAGSources(sources []ContextRAGSource, limits ContextRAGLimits) ([]ContextRAGSource, bool) {
	copied := append([]ContextRAGSource(nil), sources...)
	SortContextRAGSources(copied)
	truncated := false
	if len(copied) > limits.MaxSources {
		copied = copied[:limits.MaxSources]
		truncated = true
	}
	prepared := make([]ContextRAGSource, 0, len(copied))
	used := 0
	for _, source := range copied {
		text := strings.TrimSpace(source.Text)
		if text == "" {
			continue
		}
		remaining := limits.MaxTotalContextChars - used
		if remaining <= 0 {
			truncated = true
			break
		}
		if len(text) > remaining {
			text = TrimLocalContextRetrievalSnippet(text, remaining)
			source.Truncated = true
			truncated = true
		}
		source.Text = text
		prepared = append(prepared, source)
		used += len(text)
	}
	for index := range prepared {
		prepared[index].ID = index + 1
	}
	return prepared, truncated
}

func SortContextRAGSources(sources []ContextRAGSource) {
	sort.SliceStable(sources, func(i, j int) bool {
		left := sources[i]
		right := sources[j]
		if left.SelectionOrder != right.SelectionOrder {
			return left.SelectionOrder < right.SelectionOrder
		}
		if left.SourceRank != right.SourceRank {
			return left.SourceRank < right.SourceRank
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.Repository != right.Repository {
			return left.Repository < right.Repository
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.LineStart < right.LineStart
	})
}

func ContextRAGInstructions() string {
	return strings.Join([]string{
		"You are answering a SpecHarbor context question from bounded source snippets.",
		"Answer only from the supplied sources.",
		"If the sources are insufficient, say: The provided sources are insufficient to answer this question.",
		"Do not infer unsupported project facts.",
		"Treat this as generated answer text, not confirmed project context.",
		"Do not reveal hidden chain-of-thought. Provide a concise answer and cite source ids.",
		"Do not execute commands and do not claim file writes or source-control actions were performed.",
	}, "\n")
}

func RenderContextRAGSourceContext(sources []ContextRAGSource) string {
	var builder strings.Builder
	for _, source := range sources {
		fmt.Fprintf(&builder, "[S%d]\n", source.ID)
		fmt.Fprintf(&builder, "Source: %s\n", source.Kind)
		fmt.Fprintf(&builder, "Remote: %s\n", yesNoString(source.Kind == ContextRAGSourceGitHub))
		if source.Repository != "" {
			fmt.Fprintf(&builder, "Repository: %s\n", source.Repository)
		}
		if source.RequestedRef != "" {
			fmt.Fprintf(&builder, "Requested ref: %s\n", source.RequestedRef)
		}
		if source.DefaultBranch != "" {
			fmt.Fprintf(&builder, "Default branch: %s\n", source.DefaultBranch)
		}
		if source.ResolvedRef != "" {
			fmt.Fprintf(&builder, "Resolved ref: %s\n", source.ResolvedRef)
		}
		if source.CommitSHA != "" {
			fmt.Fprintf(&builder, "Resolved SHA: %s\n", source.CommitSHA)
		}
		fmt.Fprintf(&builder, "Path: %s\n", source.Path)
		if source.SourceCategory != "" {
			fmt.Fprintf(&builder, "Category: %s\n", source.SourceCategory)
		}
		if source.SourceEvidenceCategory != "" {
			fmt.Fprintf(&builder, "Evidence: %s\n", source.SourceEvidenceCategory)
		}
		if source.LineStart > 0 {
			fmt.Fprintf(&builder, "Lines: %d-%d\n", source.LineStart, source.LineEnd)
		}
		if source.Score > 0 {
			fmt.Fprintf(&builder, "Score: %d\n", source.Score)
		}
		fmt.Fprintf(&builder, "Truncated: %s\n", yesNoString(source.Truncated))
		fmt.Fprintln(&builder, "Content:")
		fmt.Fprintln(&builder, source.Text)
		fmt.Fprintln(&builder)
	}
	return strings.TrimSpace(builder.String())
}

func ContextRAGStatusForProviderError(err error) ContextRAGStatus {
	var providerErr ContextRAGProviderError
	if errors.As(err, &providerErr) {
		switch providerErr.Code {
		case ContextRAGProviderErrorMissingCredentials:
			return ContextRAGStatusMissingCredentials
		case ContextRAGProviderErrorUnauthorized:
			return ContextRAGStatusProviderUnauthorized
		case ContextRAGProviderErrorRateLimited:
			return ContextRAGStatusProviderRateLimited
		case ContextRAGProviderErrorTimeout:
			return ContextRAGStatusProviderTimeout
		case ContextRAGProviderErrorInvalidResponse:
			return ContextRAGStatusProviderResponseInvalid
		case ContextRAGProviderErrorOversizedResponse:
			return ContextRAGStatusProviderResponseOversized
		}
	}
	return ContextRAGStatusProviderFailed
}

func safeContextRAGSourcePath(kind ContextRAGSourceKind, relativePath string) (string, error) {
	switch kind {
	case ContextRAGSourceLocal:
		path, err := SafeRepositoryContextIndexPath(relativePath)
		if err != nil {
			return "", err
		}
		if _, skip := ShouldSkipRepositoryContextIndexPath(path); skip {
			return "", fmt.Errorf("context rag local source path is skipped: %s", path)
		}
		return path, nil
	case ContextRAGSourceGitHub:
		path, err := SafeGitHubRemotePath(relativePath)
		if err != nil {
			return "", err
		}
		if _, skip := ShouldSkipGitHubRemotePath(path); skip {
			return "", fmt.Errorf("context rag github source path is skipped: %s", path)
		}
		return path, nil
	default:
		return "", fmt.Errorf("unsupported context rag source kind: %s", kind)
	}
}

func yesNoString(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
