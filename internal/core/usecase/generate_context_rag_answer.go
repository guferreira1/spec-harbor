package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

type GenerateContextRAGAnswerInput struct {
	ProjectRoot       string
	Query             string
	Provider          string
	Model             string
	SourceKinds       []string
	GitHubRepository  string
	GitHubRef         string
	GitHubPathFilters []string
}

type GenerateContextRAGAnswer struct {
	localRetriever  contextRAGLocalRetriever
	githubRetriever contextRAGGitHubRetriever
	provider        ports.ContextRAGProvider
	limits          domain.ContextRAGLimits
}

type contextRAGLocalRetriever interface {
	Execute(input RetrieveLocalContextInput) (domain.LocalContextRetrievalReport, error)
}

type contextRAGGitHubRetriever interface {
	Execute(input RetrieveGitHubRemoteContextInput) (domain.GitHubRemoteContextReport, error)
}

func NewGenerateContextRAGAnswer(
	localRetriever contextRAGLocalRetriever,
	githubRetriever contextRAGGitHubRetriever,
	provider ports.ContextRAGProvider,
) *GenerateContextRAGAnswer {
	return &GenerateContextRAGAnswer{
		localRetriever:  localRetriever,
		githubRetriever: githubRetriever,
		provider:        provider,
		limits:          domain.DefaultContextRAGLimits(),
	}
}

func NewGenerateContextRAGAnswerWithLimits(
	localRetriever contextRAGLocalRetriever,
	githubRetriever contextRAGGitHubRetriever,
	provider ports.ContextRAGProvider,
	limits domain.ContextRAGLimits,
) *GenerateContextRAGAnswer {
	return &GenerateContextRAGAnswer{
		localRetriever:  localRetriever,
		githubRetriever: githubRetriever,
		provider:        provider,
		limits:          limits,
	}
}

func (useCase *GenerateContextRAGAnswer) Execute(
	input GenerateContextRAGAnswerInput,
) (domain.ContextRAGReport, error) {
	if useCase == nil {
		return domain.ContextRAGReport{}, errors.New("context rag use case is required")
	}
	if useCase.provider == nil {
		return domain.ContextRAGReport{}, errors.New("context rag provider is required")
	}
	if err := useCase.limits.Validate(); err != nil {
		return domain.ContextRAGReport{}, err
	}
	projectRoot := strings.TrimSpace(input.ProjectRoot)
	if projectRoot == "" {
		return domain.ContextRAGReport{}, errors.New("project root is required")
	}
	providerName, err := domain.NewContextRAGProviderName(input.Provider)
	if err != nil {
		return domain.ContextRAGReport{}, err
	}
	query, err := domain.NewContextRAGQuery(input.Query, useCase.limits)
	if err != nil {
		return domain.ContextRAGReport{}, err
	}
	sourceKinds, err := domain.NewContextRAGSourceKinds(input.SourceKinds)
	if err != nil {
		return domain.ContextRAGReport{}, err
	}

	sources, sourceFailure, err := useCase.collectSources(projectRoot, query, sourceKinds, input)
	if err != nil || sourceFailure.Status != "" {
		return sourceFailure, err
	}
	if len(sources) == 0 {
		return domain.NewContextRAGReport(
			domain.ContextRAGStatusMissingSources,
			providerName,
			input.Model,
			query,
			"",
			nil,
			"no local or GitHub context sources matched the query",
			false,
		), nil
	}

	request, contextTruncated, err := domain.NewContextRAGProviderRequest(
		providerName,
		input.Model,
		query,
		sources,
		useCase.limits,
	)
	if err != nil {
		return domain.NewContextRAGReport(
			domain.ContextRAGStatusMissingSources,
			providerName,
			input.Model,
			query,
			"",
			sources,
			err.Error(),
			contextTruncated,
		), nil
	}

	response, err := useCase.provider.GenerateContextAnswer(request)
	if err != nil {
		status := domain.ContextRAGStatusForProviderError(err)
		return domain.NewContextRAGReport(
			status,
			providerName,
			request.Model,
			query,
			"",
			request.Sources,
			err.Error(),
			contextTruncated,
		), nil
	}
	response, err = domain.NewContextRAGProviderResponse(
		providerName,
		request.Model,
		response.Answer,
		response.Insufficient,
		response.FinishReason,
		useCase.limits,
	)
	if err != nil {
		status := domain.ContextRAGStatusForProviderError(err)
		return domain.NewContextRAGReport(
			status,
			providerName,
			request.Model,
			query,
			"",
			request.Sources,
			err.Error(),
			contextTruncated,
		), nil
	}

	status := domain.ContextRAGStatusAnswered
	if response.Insufficient {
		status = domain.ContextRAGStatusInsufficientSources
	}
	return domain.NewContextRAGReport(
		status,
		providerName,
		request.Model,
		query,
		response.Answer,
		request.Sources,
		"",
		contextTruncated || response.OutputTruncated,
	), nil
}

func (useCase *GenerateContextRAGAnswer) collectSources(
	projectRoot string,
	query domain.ContextRAGQuery,
	sourceKinds []domain.ContextRAGSourceKind,
	input GenerateContextRAGAnswerInput,
) ([]domain.ContextRAGSource, domain.ContextRAGReport, error) {
	var sources []domain.ContextRAGSource
	for order, sourceKind := range sourceKinds {
		switch sourceKind {
		case domain.ContextRAGSourceLocal:
			localSources, failure, err := useCase.collectLocalSources(projectRoot, query, order)
			if err != nil || failure.Status != "" {
				return nil, failure, err
			}
			sources = append(sources, localSources...)
		case domain.ContextRAGSourceGitHub:
			githubSources, failure, err := useCase.collectGitHubSources(query, input, order)
			if err != nil || failure.Status != "" {
				return nil, failure, err
			}
			sources = append(sources, githubSources...)
		default:
			return nil, domain.ContextRAGReport{}, fmt.Errorf("unsupported context rag source: %s", sourceKind)
		}
	}
	return sources, domain.ContextRAGReport{}, nil
}

func (useCase *GenerateContextRAGAnswer) collectLocalSources(
	projectRoot string,
	query domain.ContextRAGQuery,
	order int,
) ([]domain.ContextRAGSource, domain.ContextRAGReport, error) {
	if useCase.localRetriever == nil {
		return nil, domain.ContextRAGReport{}, errors.New("context rag local retriever is required")
	}
	report, err := useCase.localRetriever.Execute(RetrieveLocalContextInput{
		ProjectRoot: projectRoot,
		Query:       query.DisplayQuery,
	})
	if err != nil {
		return nil, domain.ContextRAGReport{}, err
	}
	if report.Status != domain.LocalContextRetrievalStatusCurrent &&
		report.Status != domain.LocalContextRetrievalStatusNoResults {
		return nil, domain.NewContextRAGReport(
			domain.ContextRAGStatusMissingSources,
			domain.ContextRAGProviderOpenAI,
			"",
			query,
			"",
			nil,
			report.Message,
			false,
		), nil
	}
	sources := make([]domain.ContextRAGSource, 0, len(report.Results))
	for _, result := range report.Results {
		text := result.Snippet.Text
		if strings.TrimSpace(text) == "" {
			text = result.Summary
		}
		source, err := domain.NewContextRAGSource(domain.ContextRAGSourceInput{
			Kind:                   domain.ContextRAGSourceLocal,
			Path:                   result.Path,
			SourceCategory:         result.SourceCategory,
			SourceEvidenceCategory: result.SourceEvidenceCategory,
			LineStart:              result.Snippet.LineStart,
			LineEnd:                result.Snippet.LineEnd,
			Score:                  result.Score,
			Text:                   text,
			SelectionOrder:         order,
			SourceRank:             result.Rank,
		}, useCase.limits)
		if err != nil {
			continue
		}
		sources = append(sources, source)
	}
	return sources, domain.ContextRAGReport{}, nil
}

func (useCase *GenerateContextRAGAnswer) collectGitHubSources(
	query domain.ContextRAGQuery,
	input GenerateContextRAGAnswerInput,
	order int,
) ([]domain.ContextRAGSource, domain.ContextRAGReport, error) {
	if useCase.githubRetriever == nil {
		return nil, domain.ContextRAGReport{}, errors.New("context rag github retriever is required")
	}
	report, err := useCase.githubRetriever.Execute(RetrieveGitHubRemoteContextInput{
		Repository:  input.GitHubRepository,
		Ref:         input.GitHubRef,
		Query:       query.DisplayQuery,
		PathFilters: input.GitHubPathFilters,
	})
	if err != nil {
		return nil, domain.ContextRAGReport{}, err
	}
	if report.Status != domain.GitHubRemoteContextStatusCurrent &&
		report.Status != domain.GitHubRemoteContextStatusNoResults {
		return nil, domain.NewContextRAGReport(
			domain.ContextRAGStatusMissingSources,
			domain.ContextRAGProviderOpenAI,
			"",
			query,
			"",
			nil,
			report.Message,
			report.OutputTruncated,
		), nil
	}
	sources := make([]domain.ContextRAGSource, 0, len(report.Results))
	for _, result := range report.Results {
		text := result.Snippet.Text
		if strings.TrimSpace(text) == "" {
			text = result.Summary
		}
		source, err := domain.NewContextRAGSource(domain.ContextRAGSourceInput{
			Kind:                   domain.ContextRAGSourceGitHub,
			Repository:             result.Repository,
			RequestedRef:           result.RequestedRef,
			DefaultBranch:          result.DefaultBranch,
			ResolvedRef:            result.ResolvedRef,
			CommitSHA:              result.CommitSHA,
			Path:                   result.Path,
			SourceCategory:         result.SourceCategory,
			SourceEvidenceCategory: result.SourceEvidenceCategory,
			LineStart:              result.Snippet.LineStart,
			LineEnd:                result.Snippet.LineEnd,
			Score:                  result.Score,
			Text:                   text,
			SelectionOrder:         order,
			SourceRank:             result.Rank,
		}, useCase.limits)
		if err != nil {
			continue
		}
		sources = append(sources, source)
	}
	return sources, domain.ContextRAGReport{}, nil
}
