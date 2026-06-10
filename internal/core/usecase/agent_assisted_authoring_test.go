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
			name: "unknown agent",
			input: AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     "unknown",
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
			},
			want: "unknown agent target: unknown",
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

func TestAgentAssistedAuthoringDryRunWorksForEveryRecognizedTargetWithoutRunner(t *testing.T) {
	for _, target := range domain.RecognizedAgentTargets() {
		t.Run(string(target.ID), func(t *testing.T) {
			renderer := newFakeAgentAssistedPromptRenderer()
			result, err := NewAgentAssistedAuthoring(renderer).Execute(AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     string(target.ID),
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
			})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !result.DryRun {
				t.Fatalf("DryRun = false, want true")
			}
			if result.AgentName != target.ID {
				t.Fatalf("AgentName = %q, want %q", result.AgentName, target.ID)
			}
			if result.AgentDisplayName != target.DisplayName {
				t.Fatalf("AgentDisplayName = %q, want %q", result.AgentDisplayName, target.DisplayName)
			}
			if len(renderer.requests) != 1 {
				t.Fatalf("renderer requests = %d, want one", len(renderer.requests))
			}
			if _, ok := result.RunnerResult(); ok {
				t.Fatalf("RunnerResult() ok = true, want false")
			}
		})
	}
}

func TestAgentAssistedAuthoringExecuteSupportsMappedTargets(t *testing.T) {
	for _, command := range domain.ExecutableAgentCommands() {
		t.Run(string(command.AgentID), func(t *testing.T) {
			renderer := newFakeAgentAssistedPromptRenderer()
			runner := newFakeAgentRunner(domain.NewAgentRunResult("stdout", "stderr", 0))
			input := AgentAssistedAuthoringInput{
				ProjectRoot:   " /project ",
				ChangeID:      "change",
				AgentName:     string(command.AgentID),
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
				Execute:       true,
			}

			dryRun, err := NewAgentAssistedAuthoring(renderer).Execute(AgentAssistedAuthoringInput{
				ProjectRoot:   input.ProjectRoot,
				ChangeID:      input.ChangeID,
				AgentName:     input.AgentName,
				AuthoringType: input.AuthoringType,
				Title:         input.Title,
				Summary:       input.Summary,
			})
			if err != nil {
				t.Fatalf("dry-run Execute() error = %v", err)
			}
			renderer.requests = nil

			result, err := NewAgentAssistedAuthoringWithRunner(renderer, runner).Execute(input)
			if err != nil {
				t.Fatalf("execute Execute() error = %v", err)
			}

			if result.DryRun {
				t.Fatalf("DryRun = true, want false")
			}
			if !result.Execute {
				t.Fatalf("Execute = false, want true")
			}
			if result.AgentName != command.AgentID {
				t.Fatalf("AgentName = %q, want %q", result.AgentName, command.AgentID)
			}
			if result.AgentDisplayName != command.AgentDisplayName {
				t.Fatalf("AgentDisplayName = %q, want %q", result.AgentDisplayName, command.AgentDisplayName)
			}
			if result.ResolvedCommandName != command.CommandName {
				t.Fatalf("ResolvedCommandName = %q, want %q", result.ResolvedCommandName, command.CommandName)
			}
			if result.WorkingDirectory != "/project" {
				t.Fatalf("WorkingDirectory = %q, want /project", result.WorkingDirectory)
			}
			if result.Prompt != dryRun.Prompt {
				t.Fatalf("execute prompt = %q, want dry-run prompt %q", result.Prompt, dryRun.Prompt)
			}
			if len(renderer.requests) != 1 {
				t.Fatalf("renderer requests = %d, want one", len(renderer.requests))
			}
			if len(runner.requests) != 1 {
				t.Fatalf("runner requests = %d, want one", len(runner.requests))
			}
			runRequest := runner.requests[0]
			if runRequest.CommandName != command.CommandName {
				t.Fatalf("runner CommandName = %q, want %q", runRequest.CommandName, command.CommandName)
			}
			if runRequest.WorkingDirectory() != "/project" {
				t.Fatalf("runner WorkingDirectory() = %q, want /project", runRequest.WorkingDirectory())
			}
			if runRequest.Prompt() != dryRun.Prompt {
				t.Fatalf("runner Prompt() = %q, want dry-run prompt", runRequest.Prompt())
			}
			runResult, ok := result.RunnerResult()
			if !ok {
				t.Fatalf("RunnerResult() ok = false, want true")
			}
			if runResult.Status != domain.AgentRunStatusSuccess {
				t.Fatalf("runner status = %q, want success", runResult.Status)
			}
		})
	}
}

func TestAgentAssistedAuthoringExecuteReportsNonZeroRunnerExit(t *testing.T) {
	renderer := newFakeAgentAssistedPromptRenderer()
	runner := newFakeAgentRunner(domain.NewAgentRunResult("stdout", "stderr", 5))

	result, err := NewAgentAssistedAuthoringWithRunner(renderer, runner).Execute(AgentAssistedAuthoringInput{
		ProjectRoot:   "/project",
		ChangeID:      "change",
		AgentName:     "codex",
		AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
		Title:         "Title",
		Summary:       "Summary",
		Execute:       true,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	runResult, ok := result.RunnerResult()
	if !ok {
		t.Fatalf("RunnerResult() ok = false, want true")
	}
	if runResult.Status != domain.AgentRunStatusNonZeroExit {
		t.Fatalf("Status = %q, want non_zero_exit", runResult.Status)
	}
	if runResult.Stdout != "stdout" || runResult.Stderr != "stderr" || runResult.ExitCode != 5 {
		t.Fatalf("RunnerResult() = %#v, want captured stdout/stderr/exit code", runResult)
	}
}

func TestAgentAssistedAuthoringExecuteReturnsStartupFailureWithoutNormalResult(t *testing.T) {
	wantErr := errors.New("start runner")
	renderer := newFakeAgentAssistedPromptRenderer()
	runner := newFakeAgentRunner(domain.AgentRunResult{})
	runner.err = wantErr

	result, err := NewAgentAssistedAuthoringWithRunner(renderer, runner).Execute(AgentAssistedAuthoringInput{
		ProjectRoot:   "/project",
		ChangeID:      "change",
		AgentName:     "codex",
		AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
		Title:         "Title",
		Summary:       "Summary",
		Execute:       true,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want startup failure", err)
	}
	if _, ok := result.RunnerResult(); ok {
		t.Fatalf("RunnerResult() ok = true, want false")
	}
	if len(runner.requests) != 1 {
		t.Fatalf("runner requests = %d, want one", len(runner.requests))
	}
}

func TestAgentAssistedAuthoringExecuteRejectsGenericAndUnknownTargetsBeforeRunner(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		want      string
	}{
		{
			name:      "generic",
			agentName: "generic",
			want:      "agent target has no executable local runner mapping in this change: generic",
		},
		{
			name:      "unknown",
			agentName: "unknown",
			want:      "unknown agent target: unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := newFakeAgentAssistedPromptRenderer()
			runner := newFakeAgentRunner(domain.NewAgentRunResult("", "", 0))
			_, err := NewAgentAssistedAuthoringWithRunner(renderer, runner).Execute(AgentAssistedAuthoringInput{
				ProjectRoot:   "/project",
				ChangeID:      "change",
				AgentName:     test.agentName,
				AuthoringType: string(domain.FeatureAgentAssistedAuthoringType),
				Title:         "Title",
				Summary:       "Summary",
				Execute:       true,
			})
			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("Execute() error = %q, want %q", err.Error(), test.want)
			}
			if len(renderer.requests) != 0 {
				t.Fatalf("renderer requests = %d, want none", len(renderer.requests))
			}
			if len(runner.requests) != 0 {
				t.Fatalf("runner requests = %d, want none", len(runner.requests))
			}
		})
	}
}

func TestAgentAssistedAuthoringExecuteRequiresRunnerForMappedTargets(t *testing.T) {
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
	if err == nil {
		t.Fatalf("Execute() error = nil, want missing runner error")
	}
	if err.Error() != "agent runner is required for execute mode" {
		t.Fatalf("Execute() error = %q, want missing runner error", err.Error())
	}
	if len(renderer.requests) != 0 {
		t.Fatalf("renderer requests = %d, want none", len(renderer.requests))
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

type fakeAgentRunner struct {
	requests []domain.AgentRunRequest
	result   domain.AgentRunResult
	err      error
}

func newFakeAgentRunner(result domain.AgentRunResult) *fakeAgentRunner {
	return &fakeAgentRunner{result: result}
}

func (runner *fakeAgentRunner) Run(request domain.AgentRunRequest) (domain.AgentRunResult, error) {
	runner.requests = append(runner.requests, request)
	if runner.err != nil {
		return domain.AgentRunResult{}, runner.err
	}
	return runner.result, nil
}
