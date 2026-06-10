package domain

import (
	"reflect"
	"testing"
)

func TestParseAgentNameRejectsEmptyValues(t *testing.T) {
	for _, value := range []string{"", " "} {
		t.Run(value, func(t *testing.T) {
			_, err := ParseAgentName(value)
			if err == nil {
				t.Fatalf("ParseAgentName(%q) error = nil, want required error", value)
			}
			if err.Error() != "agent name is required" {
				t.Fatalf("ParseAgentName(%q) error = %q, want required error", value, err.Error())
			}
		})
	}
}

func TestParseAgentNameTrimsDisplayName(t *testing.T) {
	got, err := ParseAgentName(" codex ")
	if err != nil {
		t.Fatalf("ParseAgentName() error = %v", err)
	}
	if got != AgentName("codex") {
		t.Fatalf("ParseAgentName() = %q, want codex", got)
	}
}

func TestParseAgentNameRejectsUnsafeDisplayText(t *testing.T) {
	for _, value := range []string{"codex\nrunner", "codex\rrunner"} {
		t.Run(value, func(t *testing.T) {
			_, err := ParseAgentName(value)
			if err == nil {
				t.Fatalf("ParseAgentName(%q) error = nil, want safe display text error", value)
			}
			if err.Error() != "agent name must be safe display text" {
				t.Fatalf("ParseAgentName(%q) error = %q, want safe display text error", value, err.Error())
			}
		})
	}
}

func TestParseAgentNameRejectsUnknownAgentTargets(t *testing.T) {
	_, err := ParseAgentName("unknown")
	if err == nil {
		t.Fatalf("ParseAgentName() error = nil, want unknown target error")
	}
	if err.Error() != "unknown agent target: unknown" {
		t.Fatalf("ParseAgentName() error = %q, want unknown target error", err.Error())
	}
}

func TestRecognizedAgentTargetsAreStable(t *testing.T) {
	got := RecognizedAgentTargets()
	want := []RecognizedAgentTargetInfo{
		{ID: "codex", DisplayName: "Codex"},
		{ID: "claude", DisplayName: "Claude Code"},
		{ID: "devin", DisplayName: "Devin"},
		{ID: "cursor", DisplayName: "Cursor"},
		{ID: "copilot", DisplayName: "GitHub Copilot"},
		{ID: "gemini", DisplayName: "Gemini CLI"},
		{ID: "roo", DisplayName: "Roo Code"},
		{ID: "windsurf", DisplayName: "Windsurf"},
		{ID: "aider", DisplayName: "Aider"},
		{ID: "generic", DisplayName: "Generic Agent"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RecognizedAgentTargets() = %#v, want %#v", got, want)
	}
	got[0].ID = "mutated"
	if RecognizedAgentTargets()[0].ID != "codex" {
		t.Fatalf("RecognizedAgentTargets() returned mutable policy")
	}
}

func TestExecutableAgentCommandsAreStable(t *testing.T) {
	got := ExecutableAgentCommands()
	wantIDs := []AgentName{"codex", "claude", "devin", "cursor", "copilot", "gemini", "roo", "windsurf", "aider"}
	if len(got) != len(wantIDs) {
		t.Fatalf("ExecutableAgentCommands() length = %d, want %d", len(got), len(wantIDs))
	}

	for index, wantID := range wantIDs {
		command := got[index]
		if command.AgentID != wantID {
			t.Fatalf("ExecutableAgentCommands()[%d].AgentID = %q, want %q", index, command.AgentID, wantID)
		}
		if command.CommandName != string(wantID) {
			t.Fatalf("ExecutableAgentCommands()[%d].CommandName = %q, want %q", index, command.CommandName, wantID)
		}
		if len(command.FixedArgs()) != 0 {
			t.Fatalf("ExecutableAgentCommands()[%d].FixedArgs() = %v, want none", index, command.FixedArgs())
		}
		if command.AgentDisplayName == "" {
			t.Fatalf("ExecutableAgentCommands()[%d].AgentDisplayName is empty", index)
		}
	}
}

func TestResolveExecutableAgentCommand(t *testing.T) {
	tests := []struct {
		agentName   AgentName
		displayName string
	}{
		{agentName: "codex", displayName: "Codex"},
		{agentName: "claude", displayName: "Claude Code"},
		{agentName: "devin", displayName: "Devin"},
		{agentName: "cursor", displayName: "Cursor"},
		{agentName: "copilot", displayName: "GitHub Copilot"},
		{agentName: "gemini", displayName: "Gemini CLI"},
		{agentName: "roo", displayName: "Roo Code"},
		{agentName: "windsurf", displayName: "Windsurf"},
		{agentName: "aider", displayName: "Aider"},
	}

	for _, test := range tests {
		t.Run(string(test.agentName), func(t *testing.T) {
			command, err := ResolveExecutableAgentCommand(test.agentName)
			if err != nil {
				t.Fatalf("ResolveExecutableAgentCommand(%q) error = %v", test.agentName, err)
			}
			if command.AgentID != test.agentName {
				t.Fatalf("AgentID = %q, want %q", command.AgentID, test.agentName)
			}
			if command.AgentDisplayName != test.displayName {
				t.Fatalf("AgentDisplayName = %q, want %q", command.AgentDisplayName, test.displayName)
			}
			if command.CommandName != string(test.agentName) {
				t.Fatalf("CommandName = %q, want %q", command.CommandName, test.agentName)
			}
			if len(command.FixedArgs()) != 0 {
				t.Fatalf("FixedArgs() = %v, want none", command.FixedArgs())
			}
		})
	}
}

func TestResolveExecutableAgentCommandRejectsGenericAndUnknownTargets(t *testing.T) {
	tests := []struct {
		name      string
		agentName AgentName
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
			_, err := ResolveExecutableAgentCommand(test.agentName)
			if err == nil {
				t.Fatalf("ResolveExecutableAgentCommand() error = nil, want %q", test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ResolveExecutableAgentCommand() error = %q, want %q", err.Error(), test.want)
			}
		})
	}
}

func TestParseAgentAssistedAuthoringTypeAcceptsSupportedTypes(t *testing.T) {
	tests := []struct {
		value string
		want  AgentAssistedAuthoringType
	}{
		{value: "feature", want: FeatureAgentAssistedAuthoringType},
		{value: "bugfix", want: BugfixAgentAssistedAuthoringType},
		{value: "docs", want: DocsAgentAssistedAuthoringType},
		{value: "refactor", want: RefactorAgentAssistedAuthoringType},
		{value: " feature ", want: FeatureAgentAssistedAuthoringType},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, err := ParseAgentAssistedAuthoringType(test.value)
			if err != nil {
				t.Fatalf("ParseAgentAssistedAuthoringType(%q) error = %v", test.value, err)
			}
			if got != test.want {
				t.Fatalf("ParseAgentAssistedAuthoringType(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestParseAgentAssistedAuthoringTypeRejectsEmptyAndUnknownValues(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "", want: "agent-assisted authoring type is required"},
		{value: " ", want: "agent-assisted authoring type is required"},
		{value: "maintenance", want: "unknown agent-assisted authoring type: maintenance"},
		{value: "Feature", want: "unknown agent-assisted authoring type: Feature"},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			_, err := ParseAgentAssistedAuthoringType(test.value)
			if err == nil {
				t.Fatalf("ParseAgentAssistedAuthoringType(%q) error = nil, want %q", test.value, test.want)
			}
			if err.Error() != test.want {
				t.Fatalf("ParseAgentAssistedAuthoringType(%q) error = %q, want %q", test.value, err.Error(), test.want)
			}
		})
	}
}

func TestSupportedAgentAssistedAuthoringTypesReturnsOnlySupportedTypes(t *testing.T) {
	got := SupportedAgentAssistedAuthoringTypes()
	want := []AgentAssistedAuthoringType{
		FeatureAgentAssistedAuthoringType,
		BugfixAgentAssistedAuthoringType,
		DocsAgentAssistedAuthoringType,
		RefactorAgentAssistedAuthoringType,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedAgentAssistedAuthoringTypes() = %v, want %v", got, want)
	}
	for _, supportedType := range got {
		if !supportedType.IsSupported() {
			t.Fatalf("%q IsSupported() = false, want true", supportedType)
		}
	}
	for _, unsupported := range []AgentAssistedAuthoringType{"", "maintenance", "Feature", "test"} {
		if unsupported.IsSupported() {
			t.Fatalf("%q IsSupported() = true, want false", unsupported)
		}
	}

	got[0] = AgentAssistedAuthoringType("mutated")
	if SupportedAgentAssistedAuthoringTypes()[0] != FeatureAgentAssistedAuthoringType {
		t.Fatalf("SupportedAgentAssistedAuthoringTypes() returned mutable policy")
	}
}

func TestAgentAssistedPromptRequestCopiesRequiredFiles(t *testing.T) {
	requiredFiles := []string{"proposal.md", "design.md"}
	request := NewAgentAssistedAuthoringPromptRequest(
		"change",
		AgentName("codex"),
		FeatureAgentAssistedAuthoringType,
		"Title",
		"Summary",
		"openspec/changes/change",
		requiredFiles,
	)

	requiredFiles[0] = "mutated.md"
	if request.RequiredFiles[0] != "proposal.md" {
		t.Fatalf("RequiredFiles = %v, want copied required files", request.RequiredFiles)
	}
}

func TestAgentAssistedAuthoringResultCopiesSlicesAndSetsDryRunStatuses(t *testing.T) {
	requiredFiles := []string{"proposal.md", "design.md"}
	plan := []string{"first", "second"}

	result := NewAgentAssistedAuthoringResult(
		"change",
		AgentName("codex"),
		FeatureAgentAssistedAuthoringType,
		"Title",
		"Summary",
		"openspec/changes/change",
		requiredFiles,
		plan,
		"prompt",
	)

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

	requiredFiles[0] = "mutated.md"
	plan[0] = "mutated"
	if !reflect.DeepEqual(result.RequiredFiles(), []string{"proposal.md", "design.md"}) {
		t.Fatalf("RequiredFiles() = %v, want copied files", result.RequiredFiles())
	}
	if !reflect.DeepEqual(result.Plan(), []string{"first", "second"}) {
		t.Fatalf("Plan() = %v, want copied plan", result.Plan())
	}

	result.RequiredFiles()[0] = "mutated.md"
	result.Plan()[0] = "mutated"
	if result.RequiredFiles()[0] != "proposal.md" {
		t.Fatalf("RequiredFiles() returned mutable result")
	}
	if result.Plan()[0] != "first" {
		t.Fatalf("Plan() returned mutable result")
	}
}

func TestAgentRunRequestCopiesResolvedCommandData(t *testing.T) {
	command := ResolvedAgentCommand{
		AgentID:          "codex",
		AgentDisplayName: "Codex",
		CommandName:      "codex",
		fixedArgs:        []string{"--fixed"},
	}

	request := NewAgentRunRequest(command, "prompt", " /project ")
	command.fixedArgs[0] = "mutated"

	if request.AgentID != "codex" {
		t.Fatalf("AgentID = %q, want codex", request.AgentID)
	}
	if request.AgentDisplayName != "Codex" {
		t.Fatalf("AgentDisplayName = %q, want Codex", request.AgentDisplayName)
	}
	if request.CommandName != "codex" {
		t.Fatalf("CommandName = %q, want codex", request.CommandName)
	}
	if request.Prompt() != "prompt" {
		t.Fatalf("Prompt() = %q, want prompt", request.Prompt())
	}
	if request.WorkingDirectory() != "/project" {
		t.Fatalf("WorkingDirectory() = %q, want /project", request.WorkingDirectory())
	}
	if !reflect.DeepEqual(request.FixedArgs(), []string{"--fixed"}) {
		t.Fatalf("FixedArgs() = %v, want copied args", request.FixedArgs())
	}
	request.FixedArgs()[0] = "mutated"
	if request.FixedArgs()[0] != "--fixed" {
		t.Fatalf("FixedArgs() returned mutable result")
	}
}

func TestAgentRunResultRepresentsStartedProcessStatus(t *testing.T) {
	success := NewAgentRunResult("out", "err", 0)
	if success.Status != AgentRunStatusSuccess {
		t.Fatalf("success Status = %q, want %q", success.Status, AgentRunStatusSuccess)
	}
	if success.Stdout != "out" || success.Stderr != "err" || success.ExitCode != 0 {
		t.Fatalf("success result = %#v, want captured output and exit code", success)
	}

	nonZero := NewAgentRunResult("out", "err", 7)
	if nonZero.Status != AgentRunStatusNonZeroExit {
		t.Fatalf("nonZero Status = %q, want %q", nonZero.Status, AgentRunStatusNonZeroExit)
	}
	if nonZero.ExitCode != 7 {
		t.Fatalf("nonZero ExitCode = %d, want 7", nonZero.ExitCode)
	}
}

func TestExecutedAgentAssistedAuthoringResultReportsRunAndSafetyFacts(t *testing.T) {
	requiredFiles := []string{"proposal.md"}
	plan := []string{"plan"}
	command := ResolvedAgentCommand{
		AgentID:          "codex",
		AgentDisplayName: "Codex",
		CommandName:      "codex",
		fixedArgs:        []string{"--fixed"},
	}
	runResult := NewAgentRunResult("stdout", "stderr", 0)

	result := NewExecutedAgentAssistedAuthoringResult(
		"change",
		command,
		FeatureAgentAssistedAuthoringType,
		"Title",
		"Summary",
		"openspec/changes/change",
		" /project ",
		requiredFiles,
		plan,
		"prompt",
		runResult,
	)

	if result.DryRun {
		t.Fatalf("DryRun = true, want execute result")
	}
	if !result.Execute {
		t.Fatalf("Execute = false, want true")
	}
	if !result.PromptSentToRunner {
		t.Fatalf("PromptSentToRunner = false, want true")
	}
	if result.WorkingDirectory != "/project" {
		t.Fatalf("WorkingDirectory = %q, want /project", result.WorkingDirectory)
	}
	for _, status := range []struct {
		name string
		got  bool
	}{
		{name: "NoFilesWritten", got: result.NoFilesWritten},
		{name: "NoPromptFileWritten", got: result.NoPromptFileWritten},
		{name: "NoAgentOutputParsedOrApplied", got: result.NoAgentOutputParsedOrApplied},
		{name: "NoOpenSpecFilesWrittenFromOutput", got: result.NoOpenSpecFilesWrittenFromOutput},
		{name: "NoProductionCodeModifiedBySpecHarbor", got: result.NoProductionCodeModifiedBySpecHarbor},
		{name: "NoAutoCommitPushMerge", got: result.NoAutoCommitPushMerge},
	} {
		if !status.got {
			t.Fatalf("%s = false, want true", status.name)
		}
	}
	if !reflect.DeepEqual(result.ResolvedCommandFixedArgs(), []string{"--fixed"}) {
		t.Fatalf("ResolvedCommandFixedArgs() = %v, want copied args", result.ResolvedCommandFixedArgs())
	}
	gotRunResult, ok := result.RunnerResult()
	if !ok {
		t.Fatalf("RunnerResult() ok = false, want true")
	}
	if gotRunResult != runResult {
		t.Fatalf("RunnerResult() = %#v, want %#v", gotRunResult, runResult)
	}

	requiredFiles[0] = "mutated.md"
	plan[0] = "mutated"
	command.fixedArgs[0] = "mutated"
	if result.RequiredFiles()[0] != "proposal.md" {
		t.Fatalf("RequiredFiles() returned mutable result")
	}
	if result.Plan()[0] != "plan" {
		t.Fatalf("Plan() returned mutable result")
	}
	if result.ResolvedCommandFixedArgs()[0] != "--fixed" {
		t.Fatalf("ResolvedCommandFixedArgs() returned mutable result")
	}
}
