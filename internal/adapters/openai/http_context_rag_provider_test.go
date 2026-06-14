package openai

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestHTTPContextRAGProviderCallsResponsesAPI(t *testing.T) {
	var requestBody string
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", request.Method)
		}
		if request.URL.Path != "/responses" {
			t.Fatalf("path = %s, want /responses", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer secret-token" {
			t.Fatalf("Authorization header was not set")
		}
		contents, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("ReadAll(request.Body) error = %v", err)
		}
		requestBody = string(contents)
		response.Header().Set("Content-Type", "application/json")
		fmt.Fprint(response, `{"output_text":"Use source-attributed context. [S1]"}`)
	}))
	defer server.Close()

	provider := NewHTTPContextRAGProviderWithClient(
		"secret-token",
		domain.DefaultContextRAGOpenAIModel,
		server.URL,
		server.Client(),
		4096,
	)
	result, err := provider.GenerateContextAnswer(contextRAGProviderRequestForOpenAITest())
	if err != nil {
		t.Fatalf("GenerateContextAnswer() error = %v", err)
	}
	if result.Answer != "Use source-attributed context. [S1]" {
		t.Fatalf("answer = %q", result.Answer)
	}
	for _, want := range []string{
		`"model":"gpt-5.4-mini"`,
		`"store":false`,
		`"max_output_tokens":`,
		"Answer only from the supplied sources.",
		"Bounded source snippets:",
		"README.md",
	} {
		if !strings.Contains(requestBody, want) {
			t.Fatalf("request body missing %q: %s", want, requestBody)
		}
	}
	if strings.Contains(requestBody, "secret-token") {
		t.Fatalf("request body leaked API token: %s", requestBody)
	}
}

func TestHTTPContextRAGProviderParsesNestedOutputText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, `{"output":[{"content":[{"type":"output_text","text":"Nested answer. [S1]"}]}]}`)
	}))
	defer server.Close()

	provider := NewHTTPContextRAGProviderWithClient("secret-token", "gpt-test", server.URL, server.Client(), 4096)
	result, err := provider.GenerateContextAnswer(contextRAGProviderRequestForOpenAITest())
	if err != nil {
		t.Fatalf("GenerateContextAnswer() error = %v", err)
	}
	if result.Answer != "Nested answer. [S1]" {
		t.Fatalf("answer = %q, want nested output text", result.Answer)
	}
}

func TestHTTPContextRAGProviderMissingAPIKeyDoesNotCallServer(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		called = true
	}))
	defer server.Close()

	provider := NewHTTPContextRAGProviderWithClient("", "gpt-test", server.URL, server.Client(), 4096)
	_, err := provider.GenerateContextAnswer(contextRAGProviderRequestForOpenAITest())
	if err == nil {
		t.Fatalf("GenerateContextAnswer() error = nil, want missing credentials")
	}
	var providerErr domain.ContextRAGProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != domain.ContextRAGProviderErrorMissingCredentials {
		t.Fatalf("error = %T %v, want missing credentials", err, err)
	}
	if called {
		t.Fatalf("server was called despite missing API key")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked token-like value: %v", err)
	}
}

func TestHTTPContextRAGProviderMapsStatusFailures(t *testing.T) {
	tests := []struct {
		status int
		want   domain.ContextRAGProviderErrorCode
	}{
		{http.StatusUnauthorized, domain.ContextRAGProviderErrorUnauthorized},
		{http.StatusForbidden, domain.ContextRAGProviderErrorUnauthorized},
		{http.StatusTooManyRequests, domain.ContextRAGProviderErrorRateLimited},
		{http.StatusInternalServerError, domain.ContextRAGProviderErrorFailed},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("status_%d", test.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.WriteHeader(test.status)
				fmt.Fprint(response, `{"error":{"message":"provider detail omitted"}}`)
			}))
			defer server.Close()

			provider := NewHTTPContextRAGProviderWithClient("secret-token", "gpt-test", server.URL, server.Client(), 4096)
			_, err := provider.GenerateContextAnswer(contextRAGProviderRequestForOpenAITest())
			var providerErr domain.ContextRAGProviderError
			if !errors.As(err, &providerErr) || providerErr.Code != test.want {
				t.Fatalf("error = %T %v, want %s", err, err, test.want)
			}
		})
	}
}

func TestHTTPContextRAGProviderRejectsOversizedAndMalformedResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
		max  int64
		want domain.ContextRAGProviderErrorCode
	}{
		{"oversized", strings.Repeat("x", 32), 8, domain.ContextRAGProviderErrorOversizedResponse},
		{"malformed", `not-json`, 4096, domain.ContextRAGProviderErrorInvalidResponse},
		{"missing text", `{"output":[]}`, 4096, domain.ContextRAGProviderErrorInvalidResponse},
		{"incomplete", `{"output_text":"partial","incomplete_details":{"reason":"max_output_tokens"}}`, 4096, domain.ContextRAGProviderErrorInvalidResponse},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				fmt.Fprint(response, test.body)
			}))
			defer server.Close()

			provider := NewHTTPContextRAGProviderWithClient("secret-token", "gpt-test", server.URL, server.Client(), test.max)
			_, err := provider.GenerateContextAnswer(contextRAGProviderRequestForOpenAITest())
			var providerErr domain.ContextRAGProviderError
			if !errors.As(err, &providerErr) || providerErr.Code != test.want {
				t.Fatalf("error = %T %v, want %s", err, err, test.want)
			}
		})
	}
}

func TestHTTPContextRAGProviderMapsTimeoutTransportFailure(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return nil, timeoutError{}
	})}
	provider := NewHTTPContextRAGProviderWithClient("secret-token", "gpt-test", "https://example.invalid", client, 4096)

	_, err := provider.GenerateContextAnswer(contextRAGProviderRequestForOpenAITest())
	var providerErr domain.ContextRAGProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != domain.ContextRAGProviderErrorTimeout {
		t.Fatalf("error = %T %v, want timeout", err, err)
	}
}

func contextRAGProviderRequestForOpenAITest() domain.ContextRAGProviderRequest {
	return domain.ContextRAGProviderRequest{
		Provider:       domain.ContextRAGProviderOpenAI,
		Model:          domain.DefaultContextRAGOpenAIModel,
		Query:          domain.ContextRAGQuery{DisplayQuery: "architecture"},
		Instructions:   domain.ContextRAGInstructions(),
		SourceContext:  "[S1]\nPath: README.md\nContent:\nHexagonal Architecture keeps boundaries clear.",
		MaxAnswerChars: 1000,
	}
}

type roundTripFunc func(request *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }
