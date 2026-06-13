package domain

import (
	"strings"
	"testing"
)

func TestParseProjectBriefMarkdownParsesKnownSections(t *testing.T) {
	detected, err := NewDetectedProjectBriefContextSource("Stack from go.mod", "Go")
	if err != nil {
		t.Fatalf("NewDetectedProjectBriefContextSource() error = %v", err)
	}
	assumption, err := NewProjectBriefAssumption("Run command: go run ./cmd/specharbor (Source: cmd/specharbor)")
	if err != nil {
		t.Fatalf("NewProjectBriefAssumption() error = %v", err)
	}
	brief, err := NewProjectBrief(sampleProjectBriefAnswers(t), []ProjectBriefContextSource{detected}, []ProjectBriefAssumption{assumption})
	if err != nil {
		t.Fatalf("NewProjectBrief() error = %v", err)
	}

	parsed, err := ParseProjectBriefMarkdown(brief.RenderMarkdown())
	if err != nil {
		t.Fatalf("ParseProjectBriefMarkdown() error = %v", err)
	}

	if parsed.Answers.Stack.Value != "Go" {
		t.Fatalf("parsed stack = %q, want Go", parsed.Answers.Stack.Value)
	}
	if len(parsed.ContextSources) != 1 || parsed.ContextSources[0].Source != ProjectBriefAnswerSourceDetectedContext {
		t.Fatalf("parsed context sources = %+v, want detected context", parsed.ContextSources)
	}
	if len(parsed.Assumptions) != 1 || parsed.Assumptions[0].Source != ProjectBriefAnswerSourceAssumption {
		t.Fatalf("parsed assumptions = %+v, want assumption source", parsed.Assumptions)
	}
}

func TestProjectBriefUpdateProposalDetectsConflictsAndStaleRecords(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	discovery := NewContextDiscoveryResult([]ContextSignal{
		mustProjectBriefUpdateSignal(t, ContextSignalKindStack, "Node.js", ContextSignalClassificationDetectedFact, "package.json"),
	}, nil)

	proposal := NewProjectBriefUpdateProposal(parsed, discovery)

	conflicts := proposal.Conflicts()
	if len(conflicts) != 1 {
		t.Fatalf("conflicts = %+v, want one stack conflict", conflicts)
	}
	if conflicts[0].Field != ProjectBriefFieldStack || conflicts[0].ExistingValue != "Go" || conflicts[0].CandidateValue != "Node.js" {
		t.Fatalf("conflict = %+v, want Go versus Node.js stack conflict", conflicts[0])
	}
	if len(proposal.StaleAssumptions) != 0 {
		t.Fatalf("stale assumptions = %+v, want none", proposal.StaleAssumptions)
	}
}

func TestProjectBriefUpdateDoesNotPromoteDetectedFactWithoutDecision(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	discovery := NewContextDiscoveryResult([]ContextSignal{
		mustProjectBriefUpdateSignal(t, ContextSignalKindStack, "Node.js", ContextSignalClassificationDetectedFact, "package.json"),
	}, nil)
	proposal := NewProjectBriefUpdateProposal(parsed, discovery)

	updated, err := ApplyProjectBriefUpdateDecisions(proposal, ProjectBriefUpdateDecisions{})
	if err != nil {
		t.Fatalf("ApplyProjectBriefUpdateDecisions() error = %v", err)
	}

	if updated.Stack.Value != "Go" {
		t.Fatalf("stack = %q, want existing confirmed Go", updated.Stack.Value)
	}
	if updated.Stack.Source != ProjectBriefAnswerSourceUserProvided {
		t.Fatalf("stack source = %q, want user provided", updated.Stack.Source)
	}
}

func TestProjectBriefUpdateAcceptsDetectedFactOnlyWithDecision(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	discovery := NewContextDiscoveryResult([]ContextSignal{
		mustProjectBriefUpdateSignal(t, ContextSignalKindStack, "Node.js", ContextSignalClassificationDetectedFact, "package.json"),
	}, nil)
	proposal := NewProjectBriefUpdateProposal(parsed, discovery)

	updated, err := ApplyProjectBriefUpdateDecisions(proposal, ProjectBriefUpdateDecisions{
		FieldDecisions: []ProjectBriefMergeDecision{{
			Field: ProjectBriefFieldStack,
			Kind:  ProjectBriefMergeDecisionAcceptDetectedFact,
			Value: "Node.js",
		}},
	})
	if err != nil {
		t.Fatalf("ApplyProjectBriefUpdateDecisions() error = %v", err)
	}

	if updated.Stack.Value != "Node.js" {
		t.Fatalf("stack = %q, want accepted detected fact", updated.Stack.Value)
	}
	if updated.Stack.Source != ProjectBriefAnswerSourceUserProvided {
		t.Fatalf("accepted stack source = %q, want user provided after explicit confirmation", updated.Stack.Source)
	}
}

func TestProjectBriefUpdateDoesNotPromoteSuggestedAssumptionWithoutDecision(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	discovery := NewContextDiscoveryResult([]ContextSignal{
		mustProjectBriefUpdateSignal(t, ContextSignalKindRunCommand, "go run ./cmd/specharbor", ContextSignalClassificationSuggestedAssumption, "cmd/specharbor"),
	}, nil)
	proposal := NewProjectBriefUpdateProposal(parsed, discovery)

	updated, err := ApplyProjectBriefUpdateDecisions(proposal, ProjectBriefUpdateDecisions{})
	if err != nil {
		t.Fatalf("ApplyProjectBriefUpdateDecisions() error = %v", err)
	}

	if updated.Commands.Run.Value != "go run ./cmd/specharbor" {
		t.Fatalf("run command = %q, want existing confirmed run command", updated.Commands.Run.Value)
	}
	if len(updated.Assumptions()) == 0 {
		t.Fatalf("assumptions = none, want suggested assumption recorded separately")
	}
	rendered := updated.RenderMarkdown()
	beforeAssumptions := strings.Split(rendered, "## Assumptions")[0]
	if strings.Contains(beforeAssumptions, "suggested_assumption") || strings.Contains(beforeAssumptions, "Source: assumption") {
		t.Fatalf("assumption was rendered before assumptions section:\n%s", rendered)
	}
}

func TestProjectBriefUpdateSupportsCustomAndIgnoreDetectedFactDecisions(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	discovery := NewContextDiscoveryResult([]ContextSignal{
		mustProjectBriefUpdateSignal(t, ContextSignalKindBuildCommand, "make build", ContextSignalClassificationDetectedFact, "Makefile"),
	}, nil)
	proposal := NewProjectBriefUpdateProposal(parsed, discovery)

	updated, err := ApplyProjectBriefUpdateDecisions(proposal, ProjectBriefUpdateDecisions{
		FieldDecisions: []ProjectBriefMergeDecision{
			{Field: ProjectBriefFieldPurpose, Kind: ProjectBriefMergeDecisionReplaceWithCustom, Value: "Spec workflow automation"},
			{Field: ProjectBriefFieldBuild, Kind: ProjectBriefMergeDecisionIgnoreDetectedFact},
		},
	})
	if err != nil {
		t.Fatalf("ApplyProjectBriefUpdateDecisions() error = %v", err)
	}

	if updated.Purpose.Value != "Spec workflow automation" {
		t.Fatalf("purpose = %q, want custom replacement", updated.Purpose.Value)
	}
	for _, contextSource := range updated.ContextSources() {
		if contextSource.Value == "make build" {
			t.Fatalf("ignored detected fact was recorded as context source: %+v", updated.ContextSources())
		}
	}
}

func TestProjectBriefUpdateCanRemoveStaleAssumptions(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	assumption, err := NewProjectBriefAssumption("Run command: old command")
	if err != nil {
		t.Fatalf("NewProjectBriefAssumption() error = %v", err)
	}
	parsed.Assumptions = []ProjectBriefAssumption{assumption}
	proposal := NewProjectBriefUpdateProposal(parsed, NewContextDiscoveryResult(nil, nil))
	if len(proposal.StaleAssumptions) != 1 {
		t.Fatalf("stale assumptions = %+v, want one", proposal.StaleAssumptions)
	}

	updated, err := ApplyProjectBriefUpdateDecisions(proposal, ProjectBriefUpdateDecisions{RemoveStaleAssumptions: true})
	if err != nil {
		t.Fatalf("ApplyProjectBriefUpdateDecisions() error = %v", err)
	}
	if len(updated.Assumptions()) != 0 {
		t.Fatalf("assumptions = %+v, want stale assumption removed", updated.Assumptions())
	}
}

func TestProjectBriefUpdateRenderingIsDeterministic(t *testing.T) {
	parsed := parsedSampleProjectBrief(t)
	discovery := NewContextDiscoveryResult([]ContextSignal{
		mustProjectBriefUpdateSignal(t, ContextSignalKindStack, "Node.js", ContextSignalClassificationDetectedFact, "package.json"),
	}, nil)
	proposal := NewProjectBriefUpdateProposal(parsed, discovery)
	decisions := ProjectBriefUpdateDecisions{
		FieldDecisions: []ProjectBriefMergeDecision{{
			Field: ProjectBriefFieldStack,
			Kind:  ProjectBriefMergeDecisionAcceptDetectedFact,
			Value: "Node.js",
		}},
	}

	updated, err := ApplyProjectBriefUpdateDecisions(proposal, decisions)
	if err != nil {
		t.Fatalf("ApplyProjectBriefUpdateDecisions() error = %v", err)
	}
	first := updated.RenderMarkdown()
	second := updated.RenderMarkdown()
	if first != second {
		t.Fatalf("RenderMarkdown() changed between calls:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	previewFirst := RenderProjectBriefUpdatePreview(proposal, decisions)
	previewSecond := RenderProjectBriefUpdatePreview(proposal, decisions)
	if previewFirst != previewSecond {
		t.Fatalf("RenderProjectBriefUpdatePreview() changed between calls")
	}
}

func parsedSampleProjectBrief(t *testing.T) ParsedProjectBrief {
	t.Helper()

	brief, err := NewProjectBrief(sampleProjectBriefAnswers(t), nil, nil)
	if err != nil {
		t.Fatalf("NewProjectBrief() error = %v", err)
	}
	parsed, err := ParseProjectBriefMarkdown(brief.RenderMarkdown())
	if err != nil {
		t.Fatalf("ParseProjectBriefMarkdown() error = %v", err)
	}
	return parsed
}

func mustProjectBriefUpdateSignal(
	t *testing.T,
	kind ContextSignalKind,
	value string,
	classification ContextSignalClassification,
	sourcePath string,
) ContextSignal {
	t.Helper()

	signal, err := NewContextSignal(ContextSignalInput{
		Kind:           kind,
		Value:          value,
		Classification: classification,
		Confidence:     ContextConfidenceHigh,
		Source: ContextSource{
			Path:     sourcePath,
			Category: ContextSourceCategoryPackageManifest,
		},
	})
	if err != nil {
		t.Fatalf("NewContextSignal() error = %v", err)
	}
	return signal
}
