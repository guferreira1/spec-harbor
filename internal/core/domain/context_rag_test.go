package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestContextRAGProviderRequestBoundsSourcesAndRendersAttribution(t *testing.T) {
	limits := DefaultContextRAGLimits()
	limits.MaxSources = 2
	limits.MaxSnippetChars = 40
	limits.MaxTotalContextChars = 70

	query, err := NewContextRAGQuery("architecture", limits)
	if err != nil {
		t.Fatalf("NewContextRAGQuery() error = %v", err)
	}
	sources := []ContextRAGSource{
		mustContextRAGSource(t, ContextRAGSourceInput{
			Kind:           ContextRAGSourceLocal,
			Path:           "README.md",
			SourceCategory: ContextSourceCategoryReadme,
			LineStart:      1,
			LineEnd:        3,
			Score:          50,
			Text:           "Hexagonal Architecture keeps the core independent from adapters.",
			SelectionOrder: 0,
			SourceRank:     1,
		}, limits),
		mustContextRAGSource(t, ContextRAGSourceInput{
			Kind:           ContextRAGSourceGitHub,
			Repository:     "owner/repo",
			ResolvedRef:    "main",
			CommitSHA:      "abc123",
			Path:           "docs/usage.md",
			SourceCategory: ContextSourceCategoryDocumentation,
			LineStart:      10,
			LineEnd:        12,
			Score:          40,
			Text:           "Remote context retrieval is explicit and bounded.",
			SelectionOrder: 1,
			SourceRank:     1,
		}, limits),
		mustContextRAGSource(t, ContextRAGSourceInput{
			Kind:           ContextRAGSourceLocal,
			Path:           "docs/workflow.md",
			SourceCategory: ContextSourceCategoryDocumentation,
			LineStart:      5,
			LineEnd:        6,
			Score:          20,
			Text:           "This third source should be omitted by the max source limit.",
			SelectionOrder: 2,
			SourceRank:     1,
		}, limits),
	}

	request, truncated, err := NewContextRAGProviderRequest(
		ContextRAGProviderOpenAI,
		"gpt-5.4-mini",
		query,
		sources,
		limits,
	)
	if err != nil {
		t.Fatalf("NewContextRAGProviderRequest() error = %v", err)
	}
	if !truncated {
		t.Fatalf("truncated = false, want true")
	}
	if len(request.Sources) != 2 {
		t.Fatalf("sources = %d, want 2", len(request.Sources))
	}
	for _, want := range []string{
		"Answer only from the supplied sources.",
		"[S1]",
		"Path: README.md",
		"Lines: 1-3",
		"[S2]",
		"Source: github",
		"Remote: yes",
		"Repository: owner/repo",
		"Resolved SHA: abc123",
	} {
		if !strings.Contains(request.Instructions+"\n"+request.SourceContext, want) {
			t.Fatalf("provider request missing %q:\n%s\n%s", want, request.Instructions, request.SourceContext)
		}
	}
	if strings.Contains(request.SourceContext, "docs/workflow.md") {
		t.Fatalf("provider request included source beyond max sources:\n%s", request.SourceContext)
	}
}

func TestContextRAGRejectsUnsafeSourcesAndUnsupportedProvider(t *testing.T) {
	limits := DefaultContextRAGLimits()
	if _, err := NewContextRAGProviderName("anthropic"); err == nil {
		t.Fatalf("NewContextRAGProviderName() error = nil, want unsupported provider")
	}
	_, err := NewContextRAGSource(ContextRAGSourceInput{
		Kind:           ContextRAGSourceLocal,
		Path:           ".env",
		SourceCategory: ContextSourceCategoryDocumentation,
		Text:           "secret",
	}, limits)
	if err == nil {
		t.Fatalf("NewContextRAGSource() error = nil, want skipped sensitive path")
	}
}

func TestContextRAGProviderResponseBoundsAnswerAndMapsErrors(t *testing.T) {
	limits := DefaultContextRAGLimits()
	limits.MaxAnswerChars = 20
	response, err := NewContextRAGProviderResponse(
		ContextRAGProviderOpenAI,
		"gpt-5.4-mini",
		strings.Repeat("answer ", 20),
		false,
		"",
		limits,
	)
	if err != nil {
		t.Fatalf("NewContextRAGProviderResponse() error = %v", err)
	}
	if !response.OutputTruncated || len(response.Answer) > limits.MaxAnswerChars {
		t.Fatalf("response = %+v, want truncated answer within limit", response)
	}

	tests := []struct {
		err  error
		want ContextRAGStatus
	}{
		{ContextRAGProviderError{Code: ContextRAGProviderErrorMissingCredentials}, ContextRAGStatusMissingCredentials},
		{ContextRAGProviderError{Code: ContextRAGProviderErrorUnauthorized}, ContextRAGStatusProviderUnauthorized},
		{ContextRAGProviderError{Code: ContextRAGProviderErrorRateLimited}, ContextRAGStatusProviderRateLimited},
		{ContextRAGProviderError{Code: ContextRAGProviderErrorTimeout}, ContextRAGStatusProviderTimeout},
		{ContextRAGProviderError{Code: ContextRAGProviderErrorInvalidResponse}, ContextRAGStatusProviderResponseInvalid},
		{ContextRAGProviderError{Code: ContextRAGProviderErrorOversizedResponse}, ContextRAGStatusProviderResponseOversized},
		{errors.New("plain failure"), ContextRAGStatusProviderFailed},
	}
	for _, test := range tests {
		if got := ContextRAGStatusForProviderError(test.err); got != test.want {
			t.Fatalf("ContextRAGStatusForProviderError(%v) = %s, want %s", test.err, got, test.want)
		}
	}
}

func mustContextRAGSource(t *testing.T, input ContextRAGSourceInput, limits ContextRAGLimits) ContextRAGSource {
	t.Helper()
	source, err := NewContextRAGSource(input, limits)
	if err != nil {
		t.Fatalf("NewContextRAGSource(%+v) error = %v", input, err)
	}
	return source
}
