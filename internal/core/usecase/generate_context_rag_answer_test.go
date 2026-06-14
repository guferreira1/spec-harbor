package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestGenerateContextRAGAnswerUsesLocalSourcesAndProvider(t *testing.T) {
	localRetriever := &contextRAGFakeLocalRetriever{
		report: contextRAGLocalReport(t, domain.LocalContextRetrievalStatusCurrent, []domain.LocalContextRetrievalResult{{
			Path:           "README.md",
			SourceCategory: domain.ContextSourceCategoryReadme,
			Score:          50,
			Snippet: domain.LocalContextRetrievalSnippet{
				LineStart: 1,
				LineEnd:   2,
				Text:      "Hexagonal Architecture keeps adapters outside core.",
			},
		}}, ""),
	}
	githubRetriever := &contextRAGFakeGitHubRetriever{}
	provider := &contextRAGFakeProvider{response: domain.ContextRAGProviderResponse{Answer: "Use hexagonal boundaries. [S1]"}}
	useCase := NewGenerateContextRAGAnswer(localRetriever, githubRetriever, provider)

	result, err := useCase.Execute(GenerateContextRAGAnswerInput{
		ProjectRoot: "/repo",
		Query:       "architecture",
		Provider:    "openai",
		Model:       domain.DefaultContextRAGOpenAIModel,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.ContextRAGStatusAnswered {
		t.Fatalf("status = %s, want answered", result.Status)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.calls)
	}
	if githubRetriever.calls != 0 {
		t.Fatalf("github calls = %d, want 0 for default local-only source", githubRetriever.calls)
	}
	if !strings.Contains(provider.request.SourceContext, "Path: README.md") ||
		!strings.Contains(provider.request.SourceContext, "Remote: no") {
		t.Fatalf("provider source context = %q, want local source attribution", provider.request.SourceContext)
	}
}

func TestGenerateContextRAGAnswerUsesGitHubOnlyWhenExplicit(t *testing.T) {
	localRetriever := &contextRAGFakeLocalRetriever{}
	githubRetriever := &contextRAGFakeGitHubRetriever{
		report: contextRAGGitHubReport(t, domain.GitHubRemoteContextStatusCurrent, []domain.GitHubRemoteContextResult{{
			Repository:     "owner/repo",
			DefaultBranch:  "main",
			ResolvedRef:    "main",
			CommitSHA:      "abc123",
			Path:           "README.md",
			SourceCategory: domain.ContextSourceCategoryReadme,
			Score:          40,
			Snippet: domain.GitHubRemoteSnippet{
				LineStart: 3,
				LineEnd:   4,
				Text:      "Remote snippets are source attributed.",
			},
			Remote: true,
		}}, ""),
	}
	provider := &contextRAGFakeProvider{response: domain.ContextRAGProviderResponse{Answer: "Remote retrieval is explicit. [S1]"}}
	useCase := NewGenerateContextRAGAnswer(localRetriever, githubRetriever, provider)

	result, err := useCase.Execute(GenerateContextRAGAnswerInput{
		ProjectRoot:      "/repo",
		Query:            "architecture",
		Provider:         "openai",
		Model:            domain.DefaultContextRAGOpenAIModel,
		SourceKinds:      []string{"github"},
		GitHubRepository: "owner/repo",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.ContextRAGStatusAnswered {
		t.Fatalf("status = %s, want answered", result.Status)
	}
	if localRetriever.calls != 0 {
		t.Fatalf("local calls = %d, want 0 for github-only source", localRetriever.calls)
	}
	if githubRetriever.calls != 1 {
		t.Fatalf("github calls = %d, want 1", githubRetriever.calls)
	}
	if !strings.Contains(provider.request.SourceContext, "Remote: yes") ||
		!strings.Contains(provider.request.SourceContext, "Repository: owner/repo") {
		t.Fatalf("provider source context = %q, want remote source attribution", provider.request.SourceContext)
	}
}

func TestGenerateContextRAGAnswerMissingSourcesDoesNotCallProvider(t *testing.T) {
	localRetriever := &contextRAGFakeLocalRetriever{
		report: contextRAGLocalReport(t, domain.LocalContextRetrievalStatusNoResults, nil, "no matching local context found"),
	}
	provider := &contextRAGFakeProvider{response: domain.ContextRAGProviderResponse{Answer: "should not be called"}}
	useCase := NewGenerateContextRAGAnswer(localRetriever, &contextRAGFakeGitHubRetriever{}, provider)

	result, err := useCase.Execute(GenerateContextRAGAnswerInput{
		ProjectRoot: "/repo",
		Query:       "architecture",
		Provider:    "openai",
		Model:       domain.DefaultContextRAGOpenAIModel,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Status != domain.ContextRAGStatusMissingSources {
		t.Fatalf("status = %s, want missing_sources", result.Status)
	}
	if provider.calls != 0 {
		t.Fatalf("provider calls = %d, want 0", provider.calls)
	}
}

func TestGenerateContextRAGAnswerMapsProviderFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want domain.ContextRAGStatus
	}{
		{"missing credentials", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorMissingCredentials, Message: "missing key"}, domain.ContextRAGStatusMissingCredentials},
		{"unauthorized", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorUnauthorized, Message: "not authorized"}, domain.ContextRAGStatusProviderUnauthorized},
		{"rate limited", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorRateLimited, Message: "rate limited"}, domain.ContextRAGStatusProviderRateLimited},
		{"timeout", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorTimeout, Message: "timeout"}, domain.ContextRAGStatusProviderTimeout},
		{"oversized", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorOversizedResponse, Message: "too large"}, domain.ContextRAGStatusProviderResponseOversized},
		{"malformed", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorInvalidResponse, Message: "malformed"}, domain.ContextRAGStatusProviderResponseInvalid},
		{"network", domain.ContextRAGProviderError{Code: domain.ContextRAGProviderErrorNetwork, Message: "network"}, domain.ContextRAGStatusProviderFailed},
		{"plain", errors.New("plain failure"), domain.ContextRAGStatusProviderFailed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			localRetriever := &contextRAGFakeLocalRetriever{
				report: contextRAGLocalReport(t, domain.LocalContextRetrievalStatusCurrent, []domain.LocalContextRetrievalResult{{
					Path:           "README.md",
					SourceCategory: domain.ContextSourceCategoryReadme,
					Score:          50,
					Snippet: domain.LocalContextRetrievalSnippet{
						LineStart: 1,
						LineEnd:   1,
						Text:      "Architecture source.",
					},
				}}, ""),
			}
			provider := &contextRAGFakeProvider{err: test.err}
			useCase := NewGenerateContextRAGAnswer(localRetriever, &contextRAGFakeGitHubRetriever{}, provider)

			result, err := useCase.Execute(GenerateContextRAGAnswerInput{
				ProjectRoot: "/repo",
				Query:       "architecture",
				Provider:    "openai",
				Model:       domain.DefaultContextRAGOpenAIModel,
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.Status != test.want {
				t.Fatalf("status = %s, want %s", result.Status, test.want)
			}
			if result.Message == "" {
				t.Fatalf("message is empty, want safe provider detail")
			}
		})
	}
}

func TestGenerateContextRAGAnswerHandlesInsufficientAndInvalidProviderResponses(t *testing.T) {
	tests := []struct {
		name     string
		response domain.ContextRAGProviderResponse
		want     domain.ContextRAGStatus
	}{
		{
			name:     "insufficient",
			response: domain.ContextRAGProviderResponse{Answer: "The provided sources are insufficient to answer this question.", Insufficient: true},
			want:     domain.ContextRAGStatusInsufficientSources,
		},
		{
			name:     "empty answer",
			response: domain.ContextRAGProviderResponse{},
			want:     domain.ContextRAGStatusProviderResponseInvalid,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			localRetriever := &contextRAGFakeLocalRetriever{
				report: contextRAGLocalReport(t, domain.LocalContextRetrievalStatusCurrent, []domain.LocalContextRetrievalResult{{
					Path:           "README.md",
					SourceCategory: domain.ContextSourceCategoryReadme,
					Score:          50,
					Snippet: domain.LocalContextRetrievalSnippet{
						LineStart: 1,
						LineEnd:   1,
						Text:      "Architecture source.",
					},
				}}, ""),
			}
			provider := &contextRAGFakeProvider{response: test.response}
			useCase := NewGenerateContextRAGAnswer(localRetriever, &contextRAGFakeGitHubRetriever{}, provider)

			result, err := useCase.Execute(GenerateContextRAGAnswerInput{
				ProjectRoot: "/repo",
				Query:       "architecture",
				Provider:    "openai",
				Model:       domain.DefaultContextRAGOpenAIModel,
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if result.Status != test.want {
				t.Fatalf("status = %s, want %s", result.Status, test.want)
			}
		})
	}
}

type contextRAGFakeLocalRetriever struct {
	calls  int
	report domain.LocalContextRetrievalReport
	err    error
}

func (retriever *contextRAGFakeLocalRetriever) Execute(
	input RetrieveLocalContextInput,
) (domain.LocalContextRetrievalReport, error) {
	retriever.calls++
	return retriever.report, retriever.err
}

type contextRAGFakeGitHubRetriever struct {
	calls  int
	report domain.GitHubRemoteContextReport
	err    error
}

func (retriever *contextRAGFakeGitHubRetriever) Execute(
	input RetrieveGitHubRemoteContextInput,
) (domain.GitHubRemoteContextReport, error) {
	retriever.calls++
	return retriever.report, retriever.err
}

type contextRAGFakeProvider struct {
	calls    int
	request  domain.ContextRAGProviderRequest
	response domain.ContextRAGProviderResponse
	err      error
}

func (provider *contextRAGFakeProvider) GenerateContextAnswer(
	request domain.ContextRAGProviderRequest,
) (domain.ContextRAGProviderResponse, error) {
	provider.calls++
	provider.request = request
	return provider.response, provider.err
}

func contextRAGLocalReport(
	t *testing.T,
	status domain.LocalContextRetrievalStatus,
	results []domain.LocalContextRetrievalResult,
	message string,
) domain.LocalContextRetrievalReport {
	t.Helper()
	query, err := domain.NewLocalContextRetrievalQuery("architecture", domain.DefaultLocalContextRetrievalLimits())
	if err != nil {
		t.Fatalf("NewLocalContextRetrievalQuery() error = %v", err)
	}
	return domain.NewLocalContextRetrievalReport(
		status,
		query,
		domain.RepositoryContextIndexPath,
		results,
		nil,
		message,
		false,
	)
}

func contextRAGGitHubReport(
	t *testing.T,
	status domain.GitHubRemoteContextStatus,
	results []domain.GitHubRemoteContextResult,
	message string,
) domain.GitHubRemoteContextReport {
	t.Helper()
	limits := domain.DefaultGitHubRemoteContextLimits()
	locator, err := domain.NewGitHubRepositoryLocator("owner/repo", limits)
	if err != nil {
		t.Fatalf("NewGitHubRepositoryLocator() error = %v", err)
	}
	query, err := domain.NewGitHubRemoteContextQuery("architecture", limits)
	if err != nil {
		t.Fatalf("NewGitHubRemoteContextQuery() error = %v", err)
	}
	return domain.NewGitHubRemoteContextReport(
		status,
		locator,
		domain.GitHubRemoteResolvedRef{DefaultBranch: "main", ResolvedRef: "main", CommitSHA: "abc123"},
		query,
		nil,
		results,
		nil,
		message,
		false,
	)
}
