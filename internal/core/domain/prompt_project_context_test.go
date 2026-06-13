package domain

import (
	"strings"
	"testing"
)

func TestPromptProjectContextRendersSeparatedContextSections(t *testing.T) {
	result := NewContextDiscoveryResult([]ContextSignal{
		newPromptContextSignal(t, ContextSignalKindStack, "Go", ContextSignalClassificationUserConfirmedContext, ContextConfidenceHigh, ContextSource{
			Path:     ".specharbor/project-brief.md",
			Category: ContextSourceCategoryProjectBrief,
			Evidence: "Stack",
		}),
		newPromptContextSignal(t, ContextSignalKindLanguage, "Go", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     "go.mod",
			Category: ContextSourceCategoryPackageManifest,
		}),
		newPromptContextSignal(t, ContextSignalKindTestCommand, "go test ./...", ContextSignalClassificationSuggestedAssumption, ContextConfidenceMedium, ContextSource{
			Path:     "go.mod",
			Category: ContextSourceCategoryPackageManifest,
			Evidence: "go.mod convention",
		}),
	}, nil)

	context := NewPromptProjectContext(result, DefaultPromptContextRenderPolicy())
	rendered := context.RenderMarkdown()

	for _, want := range []string{
		"## Project Context",
		"Use the context below as guidance, but do not treat assumptions as facts.",
		"### User-confirmed context",
		"- Stack: Go",
		"### Detected facts",
		"- Language: Go\n  Source: go.mod\n  Confidence: high",
		"### Suggested assumptions",
		"- Test command may be `go test ./...`\n  Source: go.mod (go.mod convention)\n  Confidence: medium",
		"Rules:",
		"Prefer user-confirmed context over detected facts.",
		"Do not treat suggested assumptions as facts.",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("RenderMarkdown() =\n%s\nwant to contain %q", rendered, want)
		}
	}

	beforeAssumptions := strings.Split(rendered, "### Suggested assumptions")[0]
	if strings.Contains(beforeAssumptions, "go test ./...") {
		t.Fatalf("assumption rendered before Suggested assumptions:\n%s", rendered)
	}
}

func TestPromptProjectContextRendersConfirmedProjectBriefAgentBehavior(t *testing.T) {
	result := NewContextDiscoveryResult([]ContextSignal{
		newPromptContextSignal(t, ContextSignalKindAgentInstructionSource, "Ask before assuming", ContextSignalClassificationUserConfirmedContext, ContextConfidenceHigh, ContextSource{
			Path:     ".specharbor/project-brief.md",
			Category: ContextSourceCategoryProjectBrief,
			Evidence: "Agent behavior",
		}),
		newPromptContextSignal(t, ContextSignalKindAgentInstructionSource, "AGENTS.md", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     "AGENTS.md",
			Category: ContextSourceCategoryAgentInstruction,
		}),
		newPromptContextSignal(t, ContextSignalKindAgentInstructionSource, ".specharbor/rules/", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     ".specharbor/rules",
			Category: ContextSourceCategorySpecHarborRules,
		}),
		newPromptContextSignal(t, ContextSignalKindAgentInstructionSource, "Proceed without asking", ContextSignalClassificationSuggestedAssumption, ContextConfidenceLow, ContextSource{
			Path:     "README.md",
			Category: ContextSourceCategoryReadme,
			Evidence: "unsupported assumption",
		}),
	}, nil)

	rendered := NewPromptProjectContext(result, DefaultPromptContextRenderPolicy()).RenderMarkdown()

	confirmedStart := strings.Index(rendered, "### User-confirmed context")
	detectedStart := strings.Index(rendered, "### Detected facts")
	conflictStart := strings.Index(rendered, "### Conflict notes")
	if confirmedStart == -1 || detectedStart == -1 || conflictStart == -1 || confirmedStart > detectedStart {
		t.Fatalf("RenderMarkdown() =\n%s\nwant confirmed context before detected facts and conflict notes", rendered)
	}

	confirmedSection := rendered[confirmedStart:detectedStart]
	detectedSection := rendered[detectedStart:conflictStart]
	for _, want := range []string{
		"### User-confirmed context",
		"- Agent behavior: Ask before assuming",
	} {
		if !strings.Contains(confirmedSection, want) {
			t.Fatalf("confirmed section =\n%s\nwant %q", confirmedSection, want)
		}
	}
	for _, want := range []string{
		"### Detected facts",
		"- Agent rules: AGENTS.md",
		"- Agent rules: .specharbor/rules/",
	} {
		if !strings.Contains(detectedSection, want) {
			t.Fatalf("detected section =\n%s\nwant %q", detectedSection, want)
		}
	}
	if strings.Contains(detectedSection, "Ask before assuming") {
		t.Fatalf("confirmed agent behavior rendered as detected fact:\n%s", rendered)
	}
	if strings.Contains(rendered, "### Suggested assumptions") || strings.Contains(rendered, "Proceed without asking") {
		t.Fatalf("suggested agent behavior rendered despite confirmed context:\n%s", rendered)
	}
}

func TestPromptProjectContextPrefersConfirmedContextAndNotesConflicts(t *testing.T) {
	result := NewContextDiscoveryResult([]ContextSignal{
		newPromptContextSignal(t, ContextSignalKindStack, "Go", ContextSignalClassificationUserConfirmedContext, ContextConfidenceHigh, ContextSource{
			Path:     ".specharbor/project-brief.md",
			Category: ContextSourceCategoryProjectBrief,
			Evidence: "Stack",
		}),
		newPromptContextSignal(t, ContextSignalKindStack, "Go", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     "go.mod",
			Category: ContextSourceCategoryPackageManifest,
		}),
		newPromptContextSignal(t, ContextSignalKindStack, "Node.js", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     "package.json",
			Category: ContextSourceCategoryPackageManifest,
		}),
	}, nil)

	rendered := NewPromptProjectContext(result, DefaultPromptContextRenderPolicy()).RenderMarkdown()

	if !strings.Contains(rendered, "### User-confirmed context\n\n- Stack: Go") {
		t.Fatalf("RenderMarkdown() =\n%s\nwant confirmed stack first", rendered)
	}
	if strings.Contains(rendered, "- Stack: Go\n  Source: go.mod") {
		t.Fatalf("detected duplicate of confirmed stack rendered:\n%s", rendered)
	}
	for _, want := range []string{
		"- Stack: Node.js\n  Source: package.json\n  Confidence: high",
		"### Conflict notes",
		"Confirmed Stack is Go from .specharbor/project-brief.md; detected Stack includes Node.js from package.json.",
		"Prefer the confirmed value unless the user updates the brief.",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("RenderMarkdown() =\n%s\nwant to contain %q", rendered, want)
		}
	}
}

func TestPromptProjectContextDoesNotRenderAssumptionsAsFacts(t *testing.T) {
	result := NewContextDiscoveryResult([]ContextSignal{
		newPromptContextSignal(t, ContextSignalKindTestCommand, "make test", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     "Makefile",
			Category: ContextSourceCategoryTaskRunner,
			Evidence: "test target",
		}),
		newPromptContextSignal(t, ContextSignalKindTestCommand, "go test ./...", ContextSignalClassificationSuggestedAssumption, ContextConfidenceMedium, ContextSource{
			Path:     "go.mod",
			Category: ContextSourceCategoryPackageManifest,
			Evidence: "go.mod convention",
		}),
	}, nil)

	rendered := NewPromptProjectContext(result, DefaultPromptContextRenderPolicy()).RenderMarkdown()

	if !strings.Contains(rendered, "### Detected facts") || !strings.Contains(rendered, "- Test command: make test") {
		t.Fatalf("RenderMarkdown() =\n%s\nwant detected test command", rendered)
	}
	if strings.Contains(rendered, "### Suggested assumptions") || strings.Contains(rendered, "go test ./...") {
		t.Fatalf("assumption rendered despite detected fact taking precedence:\n%s", rendered)
	}
}

func TestPromptProjectContextRendersMissingContextInstructions(t *testing.T) {
	rendered := NewPromptProjectContext(NewContextDiscoveryResult(nil, nil), DefaultPromptContextRenderPolicy()).RenderMarkdown()

	for _, want := range []string{
		"## Project Context",
		promptContextMissingInstructions,
		"Ask before making major architecture, persistence, or workflow decisions when context is missing, ambiguous, or conflicting.",
		"Do not invent stack, architecture, commands, persistence decisions, workflow decisions, or project direction.",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("RenderMarkdown() =\n%s\nwant to contain %q", rendered, want)
		}
	}
}

func TestPromptProjectContextAppliesDeterministicSizeLimits(t *testing.T) {
	result := NewContextDiscoveryResult([]ContextSignal{
		newPromptContextSignal(t, ContextSignalKindLanguage, "Go with an intentionally long description", ContextSignalClassificationDetectedFact, ContextConfidenceHigh, ContextSource{
			Path:     "go.mod",
			Category: ContextSourceCategoryPackageManifest,
		}),
		newPromptContextSignal(t, ContextSignalKindFramework, "VeryLongFrameworkName", ContextSignalClassificationDetectedFact, ContextConfidenceMedium, ContextSource{
			Path:     "package.json",
			Category: ContextSourceCategoryPackageManifest,
		}),
		newPromptContextSignal(t, ContextSignalKindRunCommand, "go run ./cmd/example", ContextSignalClassificationSuggestedAssumption, ContextConfidenceMedium, ContextSource{
			Path:     "cmd/example",
			Category: ContextSourceCategoryRepositoryLayout,
			Evidence: "Go CLI layout convention",
		}),
		newPromptContextSignal(t, ContextSignalKindBuildCommand, "go build ./...", ContextSignalClassificationSuggestedAssumption, ContextConfidenceMedium, ContextSource{
			Path:     "go.mod",
			Category: ContextSourceCategoryPackageManifest,
			Evidence: "go.mod convention",
		}),
	}, nil)
	policy := DefaultPromptContextRenderPolicy()
	policy.MaxDetectedFactItems = 1
	policy.MaxSuggestedAssumptionItems = 1
	policy.MaxValueLength = 12

	rendered := NewPromptProjectContext(result, policy).RenderMarkdown()

	if !strings.Contains(rendered, "- Language: Go with a...") {
		t.Fatalf("RenderMarkdown() =\n%s\nwant truncated detected value", rendered)
	}
	if strings.Contains(rendered, "VeryLongFrameworkName") {
		t.Fatalf("RenderMarkdown() =\n%s\nwant second detected fact omitted", rendered)
	}
	if strings.Count(rendered, " may be `") != 1 {
		t.Fatalf("RenderMarkdown() =\n%s\nwant one suggested assumption", rendered)
	}
	if !strings.Contains(rendered, promptContextOmittedNotice) {
		t.Fatalf("RenderMarkdown() =\n%s\nwant omitted notice", rendered)
	}
}

func newPromptContextSignal(
	t *testing.T,
	kind ContextSignalKind,
	value string,
	classification ContextSignalClassification,
	confidence ContextConfidence,
	source ContextSource,
) ContextSignal {
	t.Helper()

	signal, err := NewContextSignal(ContextSignalInput{
		Kind:           kind,
		Value:          value,
		Classification: classification,
		Confidence:     confidence,
		Source:         source,
	})
	if err != nil {
		t.Fatalf("NewContextSignal() error = %v", err)
	}
	return signal
}
