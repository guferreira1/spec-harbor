package domain

import (
	"strings"
	"testing"
)

func validateSingleFile(fileName string, content string) []ValidationFinding {
	return ValidateChangeFileContents([]ChangeFileContent{
		{
			FileName:     fileName,
			RelativePath: "openspec/changes/example/" + fileName,
			Content:      content,
		},
	})
}

func findingCodes(findings []ValidationFinding) []string {
	var codes []string
	for _, finding := range findings {
		codes = append(codes, string(finding.Code))
	}
	return codes
}

func hasFindingCode(findings []ValidationFinding, code ValidationFindingCode) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func findingByCode(t *testing.T, findings []ValidationFinding, code ValidationFindingCode) ValidationFinding {
	t.Helper()
	for _, finding := range findings {
		if finding.Code == code {
			return finding
		}
	}
	t.Fatalf("findings %v missing code %q", findingCodes(findings), code)
	return ValidationFinding{}
}

func TestEmptyFileSuppressesOtherContentFindings(t *testing.T) {
	for _, content := range []string{"", "   ", "\n\n\t\n"} {
		findings := validateSingleFile("proposal.md", content)
		if len(findings) != 1 {
			t.Fatalf("findings = %v, want only file_empty", findingCodes(findings))
		}
		finding := findings[0]
		if finding.Code != ValidationFindingCodeFileEmpty {
			t.Fatalf("code = %q, want %q", finding.Code, ValidationFindingCodeFileEmpty)
		}
		if finding.Severity != ValidationFindingSeverityError {
			t.Fatalf("severity = %q, want error", finding.Severity)
		}
		if finding.RelativePath != "openspec/changes/example/proposal.md" {
			t.Fatalf("RelativePath = %q, want file path", finding.RelativePath)
		}
	}
}

func TestMissingHeadingAndBodyDetection(t *testing.T) {
	findings := validateSingleFile("design.md", "Some prose without any heading.\n")
	if !hasFindingCode(findings, ValidationFindingCodeFileMissingHeading) {
		t.Fatalf("findings = %v, want file_missing_heading", findingCodes(findings))
	}
	if hasFindingCode(findings, ValidationFindingCodeFileMissingBody) {
		t.Fatalf("findings = %v, want no file_missing_body for prose content", findingCodes(findings))
	}

	findings = validateSingleFile("design.md", "# Design\n\n## Overview\n")
	if !hasFindingCode(findings, ValidationFindingCodeFileMissingBody) {
		t.Fatalf("findings = %v, want file_missing_body for heading-only file", findingCodes(findings))
	}
	if hasFindingCode(findings, ValidationFindingCodeFileMissingHeading) {
		t.Fatalf("findings = %v, want no file_missing_heading", findingCodes(findings))
	}
	if findingByCode(t, findings, ValidationFindingCodeFileMissingBody).Severity != ValidationFindingSeverityError {
		t.Fatalf("file_missing_body severity is not error")
	}
}

func TestPlaceholderDetection(t *testing.T) {
	flagged := []struct {
		name   string
		line   string
		marker string
	}{
		{name: "standalone TBD", line: "The rollout plan is TBD for now.", marker: `"TBD"`},
		{name: "standalone TODO", line: "TODO write the design.", marker: `"TODO"`},
		{name: "standalone FIXME", line: "FIXME before merging.", marker: `"FIXME"`},
		{name: "list item N/A", line: "- N/A", marker: `"N/A"`},
		{name: "list item ellipsis", line: "- ...", marker: `"..."`},
		{name: "list item question mark", line: "* ?", marker: `"?"`},
		{name: "lorem ipsum", line: "Lorem Ipsum dolor sit amet.", marker: `"lorem ipsum"`},
		{name: "checkbox with TBD text", line: "- [ ] TBD", marker: `"TBD"`},
	}

	for _, test := range flagged {
		t.Run(test.name, func(t *testing.T) {
			content := "# Title\n\nIntro body line.\n" + test.line + "\n"
			findings := validateSingleFile("design.md", content)
			finding := findingByCode(t, findings, ValidationFindingCodePlaceholderContent)
			if finding.Severity != ValidationFindingSeverityWarning {
				t.Fatalf("severity = %q, want warning", finding.Severity)
			}
			if !strings.Contains(finding.Message, test.marker) || !strings.Contains(finding.Message, "(line 4)") {
				t.Fatalf("message = %q, want marker %s and line 4", finding.Message, test.marker)
			}
		})
	}

	notFlagged := []struct {
		name string
		line string
	}{
		{name: "prose question", line: "Should the validator support custom rules later?"},
		{name: "todo substring", line: "The mastodon integration is unrelated."},
		{name: "lowercase todo word", line: "We keep a todo list elsewhere."},
		{name: "checkbox syntax", line: "- [ ] Write the usage documentation."},
		{name: "na inside prose", line: "The N/A label only matters as a whole item."},
	}

	for _, test := range notFlagged {
		t.Run(test.name, func(t *testing.T) {
			content := "# Title\n\nIntro body line.\n" + test.line + "\n"
			findings := validateSingleFile("design.md", content)
			if hasFindingCode(findings, ValidationFindingCodePlaceholderContent) {
				t.Fatalf("findings = %v, want no placeholder_content for %q", findingCodes(findings), test.line)
			}
		})
	}
}

func TestProposalSectionDetection(t *testing.T) {
	withSection := "# Proposal\n\n## Problem\n\nUsers cannot validate quality.\n"
	if findings := validateSingleFile("proposal.md", withSection); hasFindingCode(findings, ValidationFindingCodeProposalSectionMissing) {
		t.Fatalf("findings = %v, want no proposal_section_missing", findingCodes(findings))
	}

	caseInsensitive := "# Proposal\n\n### GOALS AND CONTEXT\n\nDeliver validation.\n"
	if findings := validateSingleFile("proposal.md", caseInsensitive); hasFindingCode(findings, ValidationFindingCodeProposalSectionMissing) {
		t.Fatalf("findings = %v, want goal heading matched case-insensitively", findingCodes(findings))
	}

	withoutSection := "# Proposal\n\n## Context\n\nSome context only.\n"
	findings := validateSingleFile("proposal.md", withoutSection)
	finding := findingByCode(t, findings, ValidationFindingCodeProposalSectionMissing)
	if finding.Severity != ValidationFindingSeverityWarning {
		t.Fatalf("severity = %q, want warning", finding.Severity)
	}
}

func TestDesignSectionDetection(t *testing.T) {
	withSection := "# Design\n\n## Technical Decisions\n\nUse pure functions.\n"
	if findings := validateSingleFile("design.md", withSection); hasFindingCode(findings, ValidationFindingCodeDesignSectionMissing) {
		t.Fatalf("findings = %v, want no design_section_missing", findingCodes(findings))
	}

	withoutSection := "# Design\n\n## Context\n\nBackground only.\n"
	findings := validateSingleFile("design.md", withoutSection)
	finding := findingByCode(t, findings, ValidationFindingCodeDesignSectionMissing)
	if finding.Severity != ValidationFindingSeverityWarning {
		t.Fatalf("severity = %q, want warning", finding.Severity)
	}

	titleOnlyLevelOne := "# Design Overview\n\nBody text.\n"
	if findings := validateSingleFile("design.md", titleOnlyLevelOne); !hasFindingCode(findings, ValidationFindingCodeDesignSectionMissing) {
		t.Fatalf("findings = %v, want design_section_missing when only a level-1 heading exists", findingCodes(findings))
	}
}

func TestTasksCheckboxGrammar(t *testing.T) {
	valid := "# Tasks\n\n## Phase 1\n\n- [ ] Implement the rules.\n- [x] Read the spec.\n* [X] Review the design.\n"
	findings := validateSingleFile("tasks.md", valid)
	for _, code := range []ValidationFindingCode{
		ValidationFindingCodeTasksCheckboxMissing,
		ValidationFindingCodeTasksCheckboxMalformed,
		ValidationFindingCodeTasksPhaseHeadingMissing,
		ValidationFindingCodeTasksAllCompleted,
	} {
		if hasFindingCode(findings, code) {
			t.Fatalf("findings = %v, want no %q for valid tasks", findingCodes(findings), code)
		}
	}

	malformed := "# Tasks\n\n## Phase 1\n\n- [] missing space\n-[ ] missing gap\n- [y] wrong mark\n- [x]\n- [ ] One valid task.\n"
	findings = validateSingleFile("tasks.md", malformed)
	var malformedFindings []ValidationFinding
	for _, finding := range findings {
		if finding.Code == ValidationFindingCodeTasksCheckboxMalformed {
			malformedFindings = append(malformedFindings, finding)
		}
	}
	if len(malformedFindings) != 4 {
		t.Fatalf("malformed findings = %v, want 4", findingCodes(findings))
	}
	for index, wantLine := range []string{"(line 5)", "(line 6)", "(line 7)", "(line 8)"} {
		if !strings.Contains(malformedFindings[index].Message, wantLine) {
			t.Fatalf("malformed message %d = %q, want %q", index, malformedFindings[index].Message, wantLine)
		}
		if malformedFindings[index].Severity != ValidationFindingSeverityError {
			t.Fatalf("malformed severity = %q, want error", malformedFindings[index].Severity)
		}
	}
	if hasFindingCode(findings, ValidationFindingCodeTasksCheckboxMissing) {
		t.Fatalf("findings = %v, want no tasks_checkbox_missing when one valid task exists", findingCodes(findings))
	}

	noCheckboxes := "# Tasks\n\n## Phase 1\n\nWrite things down.\n"
	findings = validateSingleFile("tasks.md", noCheckboxes)
	if findingByCode(t, findings, ValidationFindingCodeTasksCheckboxMissing).Severity != ValidationFindingSeverityError {
		t.Fatalf("tasks_checkbox_missing severity is not error")
	}

	noPhase := "# Tasks\n\n- [ ] Implement the rules.\n"
	findings = validateSingleFile("tasks.md", noPhase)
	if findingByCode(t, findings, ValidationFindingCodeTasksPhaseHeadingMissing).Severity != ValidationFindingSeverityWarning {
		t.Fatalf("tasks_phase_heading_missing severity is not warning")
	}

	allCompleted := "# Tasks\n\n## Phase 1\n\n- [x] Implement the rules.\n- [X] Test the rules.\n"
	findings = validateSingleFile("tasks.md", allCompleted)
	if findingByCode(t, findings, ValidationFindingCodeTasksAllCompleted).Severity != ValidationFindingSeverityWarning {
		t.Fatalf("tasks_all_completed severity is not warning")
	}
}

func TestAcceptanceCriteriaItemDetection(t *testing.T) {
	meaningful := []string{
		"# Acceptance Criteria\n\n- Validation reports findings.\n",
		"# Acceptance Criteria\n\n* Output stays deterministic.\n",
		"# Acceptance Criteria\n\n1. Exit codes follow severity.\n",
		"# Acceptance Criteria\n\n- [ ] Warnings keep exit code zero.\n",
	}
	for _, content := range meaningful {
		if findings := validateSingleFile("acceptance-criteria.md", content); hasFindingCode(findings, ValidationFindingCodeAcceptanceCriteriaItemMissing) {
			t.Fatalf("findings = %v, want no acceptance_criteria_item_missing for %q", findingCodes(findings), content)
		}
	}

	missing := []string{
		"# Acceptance Criteria\n\nProse without items.\n",
		"# Acceptance Criteria\n\n- N/A\n",
		"# Acceptance Criteria\n\n- ...\n- ?\n",
		"# Acceptance Criteria\n\n- TBD\n* TODO\n1. FIXME\n",
	}
	for _, content := range missing {
		findings := validateSingleFile("acceptance-criteria.md", content)
		finding := findingByCode(t, findings, ValidationFindingCodeAcceptanceCriteriaItemMissing)
		if finding.Severity != ValidationFindingSeverityError {
			t.Fatalf("severity = %q, want error", finding.Severity)
		}
	}
}

func TestRisksMitigationDetection(t *testing.T) {
	withHeading := "# Risks\n\n## Risks\n\n- Strict rules.\n\n## Mitigations\n\n- Warnings only.\n"
	if findings := validateSingleFile("risks.md", withHeading); hasFindingCode(findings, ValidationFindingCodeRisksMitigationMissing) {
		t.Fatalf("findings = %v, want no risks_mitigation_missing with Mitigations heading", findingCodes(findings))
	}

	withLine := "# Risks\n\n- Strict rules. Mitigation: keep them warnings.\n"
	if findings := validateSingleFile("risks.md", withLine); hasFindingCode(findings, ValidationFindingCodeRisksMitigationMissing) {
		t.Fatalf("findings = %v, want no risks_mitigation_missing with mitigation line", findingCodes(findings))
	}

	withoutMitigation := "# Risks\n\n- Strict rules could reject changes.\n"
	findings := validateSingleFile("risks.md", withoutMitigation)
	finding := findingByCode(t, findings, ValidationFindingCodeRisksMitigationMissing)
	if finding.Severity != ValidationFindingSeverityWarning {
		t.Fatalf("severity = %q, want warning", finding.Severity)
	}

	headingOnly := "# Risks\n\n## Risks\n"
	findings = validateSingleFile("risks.md", headingOnly)
	if hasFindingCode(findings, ValidationFindingCodeRisksMitigationMissing) {
		t.Fatalf("findings = %v, want no mitigation warning for heading-only file", findingCodes(findings))
	}
	if !hasFindingCode(findings, ValidationFindingCodeFileMissingBody) {
		t.Fatalf("findings = %v, want file_missing_body for heading-only file", findingCodes(findings))
	}
}

func TestBoilerplateOnlyDetectionUsesDomainOwnedMarkers(t *testing.T) {
	starterProposal := `# Proposal

## Problem

Describe the problem this change should solve and who is affected.

## Goal

Describe the outcome this change should deliver.

## Scope

- List the behavior, files, or interfaces included in this change.
`
	findings := validateSingleFile("proposal.md", starterProposal)
	finding := findingByCode(t, findings, ValidationFindingCodeBoilerplateOnlyContent)
	if finding.Severity != ValidationFindingSeverityWarning {
		t.Fatalf("severity = %q, want warning", finding.Severity)
	}
	for _, code := range []ValidationFindingCode{ValidationFindingCodeFileEmpty, ValidationFindingCodeFileMissingHeading, ValidationFindingCodeFileMissingBody} {
		if hasFindingCode(findings, code) {
			t.Fatalf("findings = %v, want no %q for starter content", findingCodes(findings), code)
		}
	}

	starterTasks := `# Tasks

## Implementation

- [ ] Read the project context, architecture spec, and active OpenSpec change.
- [ ] Keep implementation limited to the approved scope.
`
	findings = validateSingleFile("tasks.md", starterTasks)
	if !hasFindingCode(findings, ValidationFindingCodeBoilerplateOnlyContent) {
		t.Fatalf("findings = %v, want boilerplate_only_content for starter tasks", findingCodes(findings))
	}
	if hasFindingCode(findings, ValidationFindingCodeTasksCheckboxMissing) {
		t.Fatalf("findings = %v, want starter checkboxes recognized as valid", findingCodes(findings))
	}

	edited := starterProposal + "\nValidation must report severities for every finding.\n"
	findings = validateSingleFile("proposal.md", edited)
	if hasFindingCode(findings, ValidationFindingCodeBoilerplateOnlyContent) {
		t.Fatalf("findings = %v, want no boilerplate_only_content for edited content", findingCodes(findings))
	}

	authored := "# Proposal\n\n## Problem\n\nUsers cannot tell whether changes are ready.\n"
	findings = validateSingleFile("proposal.md", authored)
	if hasFindingCode(findings, ValidationFindingCodeBoilerplateOnlyContent) {
		t.Fatalf("findings = %v, want no boilerplate_only_content for authored content", findingCodes(findings))
	}
}

func TestBoilerplateNormalizationToleratesMarkersAndWhitespace(t *testing.T) {
	variants := []string{
		"- Describe the outcome this change should deliver.",
		"* Describe the outcome this change should deliver.",
		"1. Describe the outcome this change should deliver.",
		"- [ ] Describe the outcome this change should deliver.",
		"- [x] Describe the outcome this change should deliver.",
		"   Describe the outcome   this change should deliver.",
	}
	for _, variant := range variants {
		content := "# Proposal\n\n" + variant + "\n"
		findings := validateSingleFile("proposal.md", content)
		if !hasFindingCode(findings, ValidationFindingCodeBoilerplateOnlyContent) {
			t.Fatalf("findings = %v, want boilerplate_only_content for variant %q", findingCodes(findings), variant)
		}
	}
}

func TestCrossFileArchitectureRule(t *testing.T) {
	proposal := ChangeFileContent{
		FileName:     "proposal.md",
		RelativePath: "openspec/changes/example/proposal.md",
		Content:      "# Proposal\n\n## Goal\n\nMove rules into internal/core packages.\n",
	}
	designWithoutArchitecture := ChangeFileContent{
		FileName:     "design.md",
		RelativePath: "openspec/changes/example/design.md",
		Content:      "# Design\n\n## Overview\n\nKeep rules pure.\n",
	}

	findings := ValidateChangeFileContents([]ChangeFileContent{proposal, designWithoutArchitecture})
	finding := findingByCode(t, findings, ValidationFindingCodeDesignArchitectureSectionMissing)
	if finding.Severity != ValidationFindingSeverityWarning {
		t.Fatalf("severity = %q, want warning", finding.Severity)
	}
	if finding.RelativePath != designWithoutArchitecture.RelativePath {
		t.Fatalf("RelativePath = %q, want design.md path", finding.RelativePath)
	}

	designWithArchitecture := designWithoutArchitecture
	designWithArchitecture.Content += "\n## Architecture\n\nDomain owns the rules.\n"
	findings = ValidateChangeFileContents([]ChangeFileContent{proposal, designWithArchitecture})
	if hasFindingCode(findings, ValidationFindingCodeDesignArchitectureSectionMissing) {
		t.Fatalf("findings = %v, want no architecture warning with Architecture heading", findingCodes(findings))
	}

	plainProposal := proposal
	plainProposal.Content = "# Proposal\n\n## Goal\n\nImprove docs only.\n"
	findings = ValidateChangeFileContents([]ChangeFileContent{plainProposal, designWithoutArchitecture})
	if hasFindingCode(findings, ValidationFindingCodeDesignArchitectureSectionMissing) {
		t.Fatalf("findings = %v, want no architecture warning without internal package mentions", findingCodes(findings))
	}
}

func TestCrossFileDocumentationTaskRule(t *testing.T) {
	proposal := ChangeFileContent{
		FileName:     "proposal.md",
		RelativePath: "openspec/changes/example/proposal.md",
		Content:      "# Proposal\n\n## Goal\n\nImprove the `specharbor validate` command output.\n",
	}
	tasksWithoutDocs := ChangeFileContent{
		FileName:     "tasks.md",
		RelativePath: "openspec/changes/example/tasks.md",
		Content:      "# Tasks\n\n## Phase 1\n\n- [ ] Implement the new report format.\n",
	}

	findings := ValidateChangeFileContents([]ChangeFileContent{proposal, tasksWithoutDocs})
	finding := findingByCode(t, findings, ValidationFindingCodeTasksDocumentationTaskMissing)
	if finding.Severity != ValidationFindingSeverityWarning {
		t.Fatalf("severity = %q, want warning", finding.Severity)
	}
	if finding.RelativePath != tasksWithoutDocs.RelativePath {
		t.Fatalf("RelativePath = %q, want tasks.md path", finding.RelativePath)
	}

	tasksWithDocs := tasksWithoutDocs
	tasksWithDocs.Content += "- [ ] Update README and docs for the new output.\n"
	findings = ValidateChangeFileContents([]ChangeFileContent{proposal, tasksWithDocs})
	if hasFindingCode(findings, ValidationFindingCodeTasksDocumentationTaskMissing) {
		t.Fatalf("findings = %v, want no documentation warning with docs task", findingCodes(findings))
	}

	plainProposal := proposal
	plainProposal.Content = "# Proposal\n\n## Goal\n\nInternal refactor only.\n"
	findings = ValidateChangeFileContents([]ChangeFileContent{plainProposal, tasksWithoutDocs})
	if hasFindingCode(findings, ValidationFindingCodeTasksDocumentationTaskMissing) {
		t.Fatalf("findings = %v, want no documentation warning without CLI mention", findingCodes(findings))
	}
}

func TestCrossFileRulesSkipEmptyFiles(t *testing.T) {
	proposal := ChangeFileContent{
		FileName:     "proposal.md",
		RelativePath: "openspec/changes/example/proposal.md",
		Content:      "# Proposal\n\n## Goal\n\nMove rules into internal/core packages.\n",
	}
	emptyDesign := ChangeFileContent{
		FileName:     "design.md",
		RelativePath: "openspec/changes/example/design.md",
		Content:      "",
	}

	findings := ValidateChangeFileContents([]ChangeFileContent{proposal, emptyDesign})
	if hasFindingCode(findings, ValidationFindingCodeDesignArchitectureSectionMissing) {
		t.Fatalf("findings = %v, want cross-file rules skipped for empty design.md", findingCodes(findings))
	}
	if !hasFindingCode(findings, ValidationFindingCodeFileEmpty) {
		t.Fatalf("findings = %v, want file_empty for empty design.md", findingCodes(findings))
	}
}
