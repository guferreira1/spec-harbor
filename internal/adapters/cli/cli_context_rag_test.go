package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

func TestContextRagReportsMissingOpenAIKeyWithoutProviderCall(t *testing.T) {
	provider := &cliFakeContextRagProvider{}
	withFakeContextRagProvider(t, provider)
	t.Setenv(domain.ContextRAGOpenAIAPIKeyEnvVar, "")

	var output bytes.Buffer
	err := execute([]string{"context", "rag", "--provider", "openai", "--query", "architecture"}, &output)
	var exitErr ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 1 {
		t.Fatalf("error = %T %v, want ExitError 1", err, err)
	}
	report := output.String()
	for _, want := range []string{
		"Context RAG answer:",
		"Provider: openai",
		"Status: missing_credentials",
		"Sources: 0",
		domain.ContextRAGOpenAIAPIKeyEnvVar,
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("missing-token output = %q, want %q", report, want)
		}
	}
	if provider.calls != 0 {
		t.Fatalf("provider calls = %d, want 0", provider.calls)
	}
}

func TestContextRagPrintsSourceAttributedAnswerWithFakeProvider(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\n\nHexagonal Architecture keeps boundaries clear.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	writeContextIndexForCLITest(t)
	indexPath := filepath.Join(root, domain.RepositoryContextIndexPath)
	beforeIndex, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile(index before) error = %v", err)
	}

	provider := &cliFakeContextRagProvider{
		response: domain.ContextRAGProviderResponse{Answer: "SpecHarbor uses hexagonal boundaries. [S1]"},
	}
	captured := withFakeContextRagProvider(t, provider)
	capturedGitHubToken := withFakeGitHubRemoteContextReader(t, newCLIFakeGitHubRemoteReader())
	t.Setenv(domain.ContextRAGOpenAIAPIKeyEnvVar, "secret-token")
	t.Setenv(domain.ContextRAGOpenAIModelEnvVar, "test-model")
	t.Setenv(domain.GitHubRemoteContextTokenEnvVar, "github-token")

	var output bytes.Buffer
	if err := execute([]string{
		"context", "rag",
		"--provider", "openai",
		"--query", "hexagonal architecture",
		"--max-sources", "3",
		"--max-answer-chars", "500",
	}, &output); err != nil {
		t.Fatalf("execute(context rag) error = %v\noutput:\n%s", err, output.String())
	}
	report := output.String()
	for _, want := range []string{
		"Context RAG answer:",
		"Provider: openai",
		"Model: test-model",
		"Status: answered",
		"Sources: 1",
		"Answer:",
		"SpecHarbor uses hexagonal boundaries. [S1]",
		"Source list:",
		"README.md",
		"Source: local",
		"Remote: no",
		"Lines:",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("context rag output = %q, want %q", report, want)
		}
	}
	if strings.Contains(report, "secret-token") {
		t.Fatalf("context rag output leaked token: %q", report)
	}
	if captured.apiKey != "secret-token" {
		t.Fatalf("captured apiKey = %q, want secret-token", captured.apiKey)
	}
	if *capturedGitHubToken != "" {
		t.Fatalf("github token was read for local-only context rag: %q", *capturedGitHubToken)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.calls)
	}
	if provider.request.MaxAnswerChars != 500 {
		t.Fatalf("MaxAnswerChars = %d, want 500", provider.request.MaxAnswerChars)
	}
	afterIndex, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile(index after) error = %v", err)
	}
	if !bytes.Equal(beforeIndex, afterIndex) {
		t.Fatalf("context rag modified %s", domain.RepositoryContextIndexPath)
	}
}

func TestContextRagCanUseExplicitGitHubSourceWithFakeProvider(t *testing.T) {
	reader := newCLIFakeGitHubRemoteReader()
	reader.files["README.md"] = "# Architecture\n\nRemote snippets are explicit and bounded.\n"
	withFakeGitHubRemoteContextReader(t, reader)

	provider := &cliFakeContextRagProvider{
		response: domain.ContextRAGProviderResponse{Answer: "Remote context is explicit. [S1]"},
	}
	withFakeContextRagProvider(t, provider)
	t.Setenv(domain.ContextRAGOpenAIAPIKeyEnvVar, "secret-token")
	t.Setenv(domain.GitHubRemoteContextTokenEnvVar, "github-token")

	var output bytes.Buffer
	if err := execute([]string{
		"context", "rag",
		"--provider", "openai",
		"--query", "remote snippets",
		"--from", "github",
		"--repo", "owner/repo",
		"--path", "README.md",
	}, &output); err != nil {
		t.Fatalf("execute(context rag github) error = %v\noutput:\n%s", err, output.String())
	}
	report := output.String()
	for _, want := range []string{
		"Status: answered",
		"Sources: 1",
		"Remote context is explicit. [S1]",
		"Source: github",
		"Remote: yes",
		"Repository: owner/repo",
		"Resolved SHA: abc123",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("context rag github output = %q, want %q", report, want)
		}
	}
	if strings.Contains(report, "secret-token") || strings.Contains(report, "github-token") {
		t.Fatalf("context rag github output leaked token: %q", report)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.calls)
	}
	if len(provider.request.Sources) != 1 || provider.request.Sources[0].Kind != domain.ContextRAGSourceGitHub {
		t.Fatalf("provider sources = %+v, want one github source", provider.request.Sources)
	}
}

func TestContextRagRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing provider", args: []string{"context", "rag", "--query", "architecture"}, want: "context rag provider is required"},
		{name: "unsupported provider", args: []string{"context", "rag", "--provider", "other", "--query", "architecture"}, want: "unsupported context rag provider: other"},
		{name: "missing query", args: []string{"context", "rag", "--provider", "openai"}, want: "context rag query is required"},
		{name: "duplicate provider", args: []string{"context", "rag", "--provider", "openai", "--provider", "openai", "--query", "architecture"}, want: "context rag provider flag specified more than once"},
		{name: "github repo required", args: []string{"context", "rag", "--provider", "openai", "--query", "architecture", "--from", "github"}, want: "context rag repo is required when github sources are selected"},
		{name: "repo requires github", args: []string{"context", "rag", "--provider", "openai", "--query", "architecture", "--repo", "owner/repo"}, want: "context rag repo requires --from github"},
		{name: "bad max sources", args: []string{"context", "rag", "--provider", "openai", "--query", "architecture", "--max-sources", "0"}, want: "context rag max-sources must be a positive integer"},
		{name: "too many sources", args: []string{"context", "rag", "--provider", "openai", "--query", "architecture", "--max-sources", "21"}, want: "context rag max sources must be at most 20"},
		{name: "path requires github", args: []string{"context", "rag", "--provider", "openai", "--query", "architecture", "--path", "README.md"}, want: "context rag path requires --from github"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			err := execute(test.args, &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("execute(%v) error = %v, want %q", test.args, err, test.want)
			}
			if output.String() != "" {
				t.Fatalf("output = %q, want empty", output.String())
			}
		})
	}
}

func TestOfflineContextCommandsDoNotConstructContextRagProvider(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Project\n\nArchitecture boundaries.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}
	provider := &cliFakeContextRagProvider{}
	withFakeContextRagProvider(t, provider)

	commands := [][]string{
		{"context", "discover"},
		{"context", "index", "--write"},
		{"context", "retrieve", "--query", "architecture"},
	}
	for _, command := range commands {
		var output bytes.Buffer
		if err := execute(command, &output); err != nil {
			t.Fatalf("execute(%v) error = %v\noutput:\n%s", command, err, output.String())
		}
	}
	if provider.calls != 0 {
		t.Fatalf("provider calls = %d, want 0", provider.calls)
	}
}

func TestOfflineWorkflowCommandsDoNotConstructContextRagProvider(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	createOpenSpecProject(t, root)
	createAuthoredOpenSpecChange(t, root, "ready-change", map[string]string{
		"tasks.md": "# Tasks\n\n- [x] Implement the change.\n",
	})
	provider := &cliFakeContextRagProvider{}
	withFakeContextRagProvider(t, provider)

	commands := []struct {
		args       []string
		allowError bool
	}{
		{args: []string{"scan"}},
		{args: []string{"validate", "ready-change"}},
		{args: []string{"review", "ready-change"}},
		{args: []string{"prompt", "ready-change", "--role", "implementer"}},
		{args: []string{"brief"}, allowError: true},
	}
	for _, command := range commands {
		var output bytes.Buffer
		err := execute(command.args, &output)
		if err != nil && !command.allowError {
			t.Fatalf("execute(%v) error = %v\noutput:\n%s", command.args, err, output.String())
		}
	}
	if provider.calls != 0 {
		t.Fatalf("provider calls = %d, want 0", provider.calls)
	}
}

type cliFakeContextRagProvider struct {
	calls    int
	request  domain.ContextRAGProviderRequest
	response domain.ContextRAGProviderResponse
	err      error
}

func (provider *cliFakeContextRagProvider) GenerateContextAnswer(
	request domain.ContextRAGProviderRequest,
) (domain.ContextRAGProviderResponse, error) {
	provider.calls++
	provider.request = request
	return provider.response, provider.err
}

type cliContextRagProviderCapture struct {
	apiKey string
	model  string
	limits domain.ContextRAGLimits
}

func withFakeContextRagProvider(
	t *testing.T,
	provider ports.ContextRAGProvider,
) *cliContextRagProviderCapture {
	t.Helper()
	original := newContextRagProvider
	capture := &cliContextRagProviderCapture{}
	newContextRagProvider = func(apiKey string, model string, limits domain.ContextRAGLimits) ports.ContextRAGProvider {
		capture.apiKey = apiKey
		capture.model = model
		capture.limits = limits
		return provider
	}
	t.Cleanup(func() {
		newContextRagProvider = original
	})
	return capture
}
