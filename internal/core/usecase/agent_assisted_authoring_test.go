package usecase

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestAgentAssistedAuthoringDryRunReturnsPlanAndPromptForSupportedTypes(t *testing.T) {
	tests := []struct {
		name          string
		authoringType domain.AgentAssistedAuthoringType
	}{
		{name: "feature", authoringType: domain.FeatureAgentAssistedAuthoringType},
		{name: "bugfix", authoringType: domain.BugfixAgentAssistedAuthoringType},
		{name: "docs", authoringType: domain.DocsAgentAssistedAuthoringType},
		{name: "refactor", authoringType: domain.RefactorAgentAssistedAuthoringType},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := newFakeAgentAssistedPromptRenderer()
			result, err := NewAgentAssistedAuthoring(renderer).Execute(AgentAssistedAuthoringInput{
				ProjectRoot:   " /project ",
				ChangeID:      " implement-payment-retry ",
				AgentName:     " codex ",
				AuthoringType: " " + string(test.authoringType) + " ",
				Title:         " Add payment retry policy ",
				Summary:       " Create retry policy ",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			wantChangePath := openspecChangesDirectory + "/implement-payment-retry"
			if result.ChangeID != "implement-payment-retry" {
				t.Fatalf("ChangeID = %q, want trimmed change id", result.ChangeID)
			}
			if result.AgentName != domain.AgentName("codex") {
				t.Fatalf("AgentName = %q, want codex", result.AgentName)
			}
			if result.AuthoringType != test.authoringType {
				t.Fatalf("AuthoringType = %q, want %q", result.AuthoringType, test.authoringType)
			}
			if result.Title != "Add payment retry policy" {
				t.Fatalf("Title = %q, want trimmed title", result.Title)
			}
			if result.Summary != "Create retry policy" {
				t.Fatalf("Summary = %q, want trimmed summary", result.Summary)
			}
			if result.ChangePath != wantChangePath {
				t.Fatalf("ChangePath = %q, want %q", result.ChangePath, wantChangePath)
			}
			if result.Prompt != "prompt:implement-payment-retry:"+string(test.authoringType)+":Add payment retry policy:Create retry policy" {
				t.Fatalf("Prompt = %q, want rendered prompt", result.Prompt)
			}
			assertStringSlicesEqual(t, result.RequiredFiles(), domain.RequiredOpenSpecChangeFiles())
			assertStringSlicesEqual(t, result.Plan(), []string{
				"Validate the agent-assisted authoring request.",
				"Build the OpenSpec authoring plan for " + wantChangePath + ".",
				"Render a deterministic, copy-pasteable Markdown authoring prompt.",
				"Stop before writing files, executing agents, or running external commands.",
			})
			assertAgentAssistedDryRunStatuses(t, result)
			if len(renderer.requests) != 1 {
				t.Fatalf("renderer requests = %v, want one request", renderer.requests)
			}
			renderRequest := renderer.requests[0]
			if renderRequest.ChangeID != "implement-payment-retry" {
				t.Fatalf("render request ChangeID = %q, want trimmed change id", renderRequest.ChangeID)
			}
			if renderRequest.AgentName != domain.AgentName("codex") {
				t.Fatalf("render request AgentName = %q, want codex", renderRequest.AgentName)
			}
			if renderRequest.AuthoringType != test.authoringType {
				t.Fatalf("render request AuthoringType = %q, want %q", renderRequest.AuthoringType, test.authoringType)
			}
			if renderRequest.Title != "Add payment retry policy" {
				t.Fatalf("render request Title = %q, want trimmed title", renderRequest.Title)
			}
			if renderRequest.Summary != "Create retry policy" {
				t.Fatalf("render request Summary = %q, want trimmed summary", renderRequest.Summary)
			}
			if renderRequest.ChangePath != wantChangePath {
				t.Fatalf("render request ChangePath = %q, want %q", renderRequest.ChangePath, wantChangePath)
			}
			assertStringSlicesEqual(t, renderRequest.RequiredFiles, domain.RequiredOpenSpecChangeFiles())
		})
	}
}

func TestAgentAssistedAuthoringDryRunIsDeterministic(t *testing.T) {
	renderer := newFakeAgentAssistedPromptRenderer()
	useCase := NewAgentAssistedAuthoring(renderer)
	input := AgentAssistedAuthoringInput{
		ProjectRoot:   "/project",
		ChangeID:      "implement-payment-retry",
		AgentName:     "codex",
		AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
		Title:         "Add payment retry policy",
		Summary:       "Create retry policy",
	}

	first, err := useCase.Execute(input)
	if err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	second, err := useCase.Execute(input)
	if err != nil {
		t.Fatalf("second Execute() error = %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Execute() results differ:\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func TestAgentAssistedAuthoringRejectsInvalidInputBeforePromptRendering(t *testing.T) {
	tests := []struct {
		name  string
		input AgentAssistedAuthoringInput
		want  string
	}{
		{
			name: "missing project root",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   " ",
				ChangeID:      "change",
				AgentName:     "codex",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
			},
			want: "project root is required",
		},
		{
			name: "missing change id",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      " ",
				AgentName:     "codex",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
			},
			want: "change id is required",
		},
		{
			name: "missing agent",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     " ",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
			},
			want: "agent name is required",
		},
		{
			name: "missing type",
			input: AgentAssistedAuthoringInput{
				ProjectRoot: "/project",
				ChangeID:    "change",
				AgentName:   "codex",
				Title:       "Title",
				Summary:     "Summary",
			},
			want: "agent-assisted authoring type is required",
		},
		{
			name: "unknown type",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     "codex",
				AuthoringType: "maintenance",
				Title:         "Title",
				Summary:       "Summary",
			},
			want: "unknown agent-assisted authoring type: maintenance",
		},
		{
			name: "missing title",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     "codex",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         " ",
				Summary:       "Summary",
			},
			want: "agent-assisted title is required",
		},
		{
			name: "missing summary",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     "codex",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       " ",
			},
			want: "agent-assisted summary is required",
		},
		{
			name: "unsafe change id",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "../unsafe",
				AgentName:     "codex",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
			},
			want: "change id must be a safe single path segment",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := newFakeAgentAssistedPromptRenderer()
			_, err := NewAgentAssistedAuthoring(renderer).Execute(test.input)
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if len(renderer.requests) != 0 {
				t.Fatalf("renderer requests = %v, want none before rejecting input", renderer.requests)
			}
		})
	}
}

func TestAgentAssistedAuthoringRejectsExecuteWithoutRenderingPrompt(t *testing.T) {
	renderer := newFakeAgentAssistedPromptRenderer()
	_, err := NewAgentAssistedAuthoring(renderer).Execute(AgentAssistedAuthoringInput{
		ProjectRoot:   "/project",
		ChangeID:      "change",
		AgentName:     "codex",
		AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
		Title:         "Title",
		Summary:       "Summary",
		Execute:       true,
	})
	if !errors.Is(err, ErrAgentAssistedExecuteUnsupported) {
		t.Fatalf("Execute() error = %v, want unsupported execute error", err)
	}
	if len(renderer.requests) != 0 {
		t.Fatalf("renderer requests = %v, want none for unsupported execute", renderer.requests)
	}
}

func TestAgentAssistedAuthoringReturnsPromptRenderingErrorsWithoutSafetyStatus(t *testing.T) {
	wantErr := errors.New("template unavailable")
	renderer := newFakeAgentAssistedPromptRenderer()
	renderer.err = wantErr

	_, err := NewAgentAssistedAuthoring(renderer).Execute(AgentAssistedAuthoringInput{
		ProjectRoot:   "/project",
		ChangeID:      "change",
		AgentName:     "codex",
		AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
		Title:         "Title",
		Summary:       "Summary",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapping %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "render agent-assisted authoring prompt") {
		t.Fatalf("Execute() error = %q, want render context", err.Error())
	}
}

func TestAgentAssistedAuthoringRejectsMissingPromptRenderer(t *testing.T) {
	_, err := NewAgentAssistedAuthoring(nil).Execute(AgentAssistedAuthoringInput{
		ProjectRoot:   "/project",
		ChangeID:      "change",
		AgentName:     "codex",
		AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
		Title:         "Title",
		Summary:       "Summary",
	})
	if err == nil {
		t.Fatalf("Execute() error = nil, want missing prompt renderer error")
	}
	if !strings.Contains(err.Error(), "agent-assisted authoring prompt renderer is required") {
		t.Fatalf("Execute() error = %q, want prompt renderer context", err.Error())
	}
}

func assertAgentAssistedDryRunStatuses(t *testing.T, result domain.AgentAssistedAuthoringResult) {
	t.Helper()

	for _, status := range []struct {
		name string
		got  bool
	}{
		{name: "DryRun", got: result.DryRun},
		{name: "NoFilesWritten", got: result.NoFilesWritten},
		{name: "NoPromptFileWritten", got: result.NoPromptFileWritten},
		{name: "NoAgentExecuted", got: result.NoAgentExecuted},
		{name: "NoExternalCommandExecuted", got: result.NoExternalCommandExecuted},
		{name: "NoAgentOutputParsedOrApplied", got: result.NoAgentOutputParsedOrApplied},
	} {
		if !status.got {
			t.Fatalf("%s = false, want true", status.name)
		}
	}
}

type fakeAgentAssistedPromptRenderer struct {
	requests []domain.AgentAssistedAuthoringPromptRequest
	err      error
}

func newFakeAgentAssistedPromptRenderer() *fakeAgentAssistedPromptRenderer {
	return &fakeAgentAssistedPromptRenderer{}
}

func (renderer *fakeAgentAssistedPromptRenderer) Render(
	request domain.AgentAssistedAuthoringPromptRequest,
) (string, error) {
	renderer.requests = append(renderer.requests, request)
	if renderer.err != nil {
		return "", renderer.err
	}
	return strings.Join([]string{
		"prompt",
		request.ChangeID,
		string(request.AuthoringType),
		request.Title,
		request.Summary,
	}, ":"), nil
}
