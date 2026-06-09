package templates

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestAgentAssistedAuthoringPromptIncludesRequiredContextAndInputs(t *testing.T) {
	request := agentAssistedPromptRequest(domain.FeatureAgentAssistedAuthoringType)
	prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"# Agent-Assisted OpenSpec Authoring Prompt",
		"You are `codex`",
		"SpecHarbor is an open source CLI",
		"OpenSpec-based development workflows",
		"copy-pasteable",
		"Change id: `implement-payment-retry`",
		"Authoring type: `feature`",
		"Title: Add payment retry policy",
		"Summary: Create a controlled retry policy for failed payment attempts",
		"Target OpenSpec path: `openspec/changes/implement-payment-retry`",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt = %q, want to contain %q", prompt, want)
		}
	}
}

func TestAgentAssistedAuthoringPromptAvoidsUnsafeOperationalArtifacts(t *testing.T) {
	request := agentAssistedPromptRequest(domain.FeatureAgentAssistedAuthoringType)
	prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, forbidden := range []string{
		"http://",
		"https://",
		"localhost",
		"127.0.0.1",
		"/home/",
		"/users/",
		"/tmp/",
		"~/",
		"c:\\",
		"api_key",
		"api key:",
		"secret:",
		"token:",
		"password",
		"sk-",
		"github.com/",
		"gitlab.com/",
		"registry.npmjs.org",
		"ghcr.io",
		"docker.io",
		"git commit",
		"git push",
		"git merge",
		"gh workflow",
		"workflow run",
		"go test",
		"go run",
		"curl ",
		"docker build",
		"kubectl",
		"terraform apply",
		"deploy to",
		"apply patch",
		"edit internal/",
		"modify internal/",
	} {
		if strings.Contains(lowerPrompt, forbidden) {
			t.Fatalf("prompt contains unsafe operational artifact %q:\n%s", forbidden, prompt)
		}
	}
}

func TestAgentAssistedAuthoringPromptListsEveryRequiredOpenSpecFile(t *testing.T) {
	request := agentAssistedPromptRequest(domain.FeatureAgentAssistedAuthoringType)
	prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		if !strings.Contains(prompt, "`"+requiredFile+"`") {
			t.Fatalf("prompt = %q, want required file %q", prompt, requiredFile)
		}
	}
}

func TestAgentAssistedAuthoringPromptContainsScopeSafetyInstructions(t *testing.T) {
	request := agentAssistedPromptRequest(domain.FeatureAgentAssistedAuthoringType)
	prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"Create or refine only files under `openspec/changes/implement-payment-retry/`",
		"Do not implement production code.",
		"Do not modify unrelated files.",
		"Leave implementation tasks unchecked in `tasks.md`",
		"Do not run implementation, tests, source-control commands, workflow commands",
		"Output Markdown-only OpenSpec content.",
		"Do not depend on a prompt file or on files written by SpecHarbor.",
		"Run or recommend `specharbor validate implement-payment-retry`",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt = %q, want to contain %q", prompt, want)
		}
	}
}

func TestAgentAssistedAuthoringPromptPreservesArchitectureBoundaries(t *testing.T) {
	request := agentAssistedPromptRequest(domain.FeatureAgentAssistedAuthoringType)
	prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"Domain code belongs in `internal/core/domain`.",
		"Ports belong in `internal/core/ports`.",
		"Use cases belong in `internal/core/usecase`.",
		"Concrete implementations belong in `internal/adapters`.",
		"Core must not import adapters.",
		"CLI must not contain business rules.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt = %q, want architecture boundary %q", prompt, want)
		}
	}
}

func TestAgentAssistedAuthoringPromptAppliesReadmeDocsBoundaryForNonDocsTypes(t *testing.T) {
	for _, authoringType := range []domain.AgentAssistedAuthoringType{
		domain.FeatureAgentAssistedAuthoringType,
		domain.BugfixAgentAssistedAuthoringType,
		domain.RefactorAgentAssistedAuthoringType,
	} {
		t.Run(string(authoringType), func(t *testing.T) {
			request := agentAssistedPromptRequest(authoringType)
			prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}

			want := "Do not change README or documentation files; this authoring type is not documentation scope."
			if !strings.Contains(prompt, want) {
				t.Fatalf("prompt = %q, want non-docs README/docs prohibition", prompt)
			}
		})
	}
}

func TestAgentAssistedAuthoringPromptAllowsDocsScopeWithoutProductionCode(t *testing.T) {
	request := agentAssistedPromptRequest(domain.DocsAgentAssistedAuthoringType)
	prompt, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"Documentation scope may be described only when it comes from the title and summary",
		"do not edit README or docs files in this authoring step",
		"Do not implement production code.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt = %q, want docs boundary %q", prompt, want)
		}
	}
}

func TestAgentAssistedAuthoringPromptIsDeterministic(t *testing.T) {
	request := agentAssistedPromptRequest(domain.FeatureAgentAssistedAuthoringType)
	first, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("first Render() error = %v", err)
	}
	second, err := NewAgentAssistedAuthoringPromptTemplate().Render(request)
	if err != nil {
		t.Fatalf("second Render() error = %v", err)
	}

	if first != second {
		t.Fatalf("Render() returned nondeterministic prompt:\nfirst: %q\nsecond: %q", first, second)
	}
}

func TestAgentAssistedAuthoringPromptRejectsMissingRequiredTemplateData(t *testing.T) {
	tests := []struct {
		name    string
		request domain.AgentAssistedAuthoringPromptRequest
		want    string
	}{
		{
			name:    "missing change id",
			request: domain.AgentAssistedAuthoringPromptRequest{AgentName: "codex", AuthoringType: domain.FeatureAgentAssistedAuthoringType, Title: "Title", Summary: "Summary", ChangePath: "path", RequiredFiles: []string{"proposal.md"}},
			want:    "change id is required",
		},
		{
			name:    "missing agent",
			request: domain.AgentAssistedAuthoringPromptRequest{ChangeID: "change", AuthoringType: domain.FeatureAgentAssistedAuthoringType, Title: "Title", Summary: "Summary", ChangePath: "path", RequiredFiles: []string{"proposal.md"}},
			want:    "agent name is required",
		},
		{
			name:    "missing type",
			request: domain.AgentAssistedAuthoringPromptRequest{ChangeID: "change", AgentName: "codex", Title: "Title", Summary: "Summary", ChangePath: "path", RequiredFiles: []string{"proposal.md"}},
			want:    "agent-assisted authoring type is required",
		},
		{
			name:    "missing required files",
			request: domain.AgentAssistedAuthoringPromptRequest{ChangeID: "change", AgentName: "codex", AuthoringType: domain.FeatureAgentAssistedAuthoringType, Title: "Title", Summary: "Summary", ChangePath: "path"},
			want:    "required OpenSpec files are required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewAgentAssistedAuthoringPromptTemplate().Render(test.request)
			if err == nil {
				t.Fatalf("Render() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Render() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func agentAssistedPromptRequest(authoringType domain.AgentAssistedAuthoringType) domain.AgentAssistedAuthoringPromptRequest {
	return domain.NewAgentAssistedAuthoringPromptRequest(
		"implement-payment-retry",
		domain.AgentName("codex"),
		authoringType,
		"Add payment retry policy",
		"Create a controlled retry policy for failed payment attempts",
		"openspec/changes/implement-payment-retry",
		domain.RequiredOpenSpecChangeFiles(),
	)
}
