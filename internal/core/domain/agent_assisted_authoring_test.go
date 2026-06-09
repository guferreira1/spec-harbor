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
