package domain

import (
	"strings"
	"testing"
)

func TestRequiredAIGeneratedFileNames(t *testing.T) {
	got := RequiredAIGeneratedFileNames()
	want := []string{"proposal.md", "design.md", "tasks.md", "acceptance-criteria.md", "risks.md"}
	assertDomainStringsEqual(t, got, want)

	got[0] = "mutated.md"
	assertDomainStringsEqual(t, RequiredAIGeneratedFileNames(), want)

	for _, fileName := range want {
		if !IsAllowedAIGeneratedFileName(fileName) {
			t.Fatalf("IsAllowedAIGeneratedFileName(%q) = false, want true", fileName)
		}
	}
	if IsAllowedAIGeneratedFileName("other.md") {
		t.Fatalf("IsAllowedAIGeneratedFileName(other.md) = true, want false")
	}
}

func TestParseAIOutputBlocksSuccessOrdersFilesByRequiredList(t *testing.T) {
	source := strings.Join([]string{
		block("risks.md", "# Risks\n\n## Risks\n\n- Risk.\n\n## Mitigations\n\n- Mitigation."),
		block("proposal.md", "# Proposal\n\n## Problem\n\nProblem."),
		block("acceptance-criteria.md", "# Acceptance Criteria\n\n- Criteria."),
		block("tasks.md", "# Tasks\n\n## Phase 1\n\n- [ ] Do work."),
		block("design.md", "# Design\n\n## Overview\n\nApproach."),
	}, "\n")

	result := ParseAIOutputBlocks(source)
	if result.HasErrors() {
		t.Fatalf("HasErrors() = true, findings = %v", result.Findings())
	}

	var got []string
	for _, file := range result.Files() {
		got = append(got, file.FileName)
		if strings.TrimSpace(file.Content) == "" {
			t.Fatalf("file %s has empty content", file.FileName)
		}
	}
	assertDomainStringsEqual(t, got, RequiredOpenSpecChangeFiles())

	content, ok := result.ContentFor("proposal.md")
	if !ok || !strings.Contains(content, "Problem.") {
		t.Fatalf("ContentFor(proposal.md) = %q, %v; want proposal content", content, ok)
	}
}

func TestParseAIOutputBlocksRejectsUnsafeAndUnsupportedFileNames(t *testing.T) {
	tests := []struct {
		name string
		file string
		code AIOutputParseFindingCode
	}{
		{name: "unknown", file: "notes.md", code: AIOutputParseFindingCodeUnknownFileBlock},
		{name: "path traversal", file: "../proposal.md", code: AIOutputParseFindingCodePathTraversalName},
		{name: "absolute unix", file: "/tmp/proposal.md", code: AIOutputParseFindingCodeAbsoluteFileName},
		{name: "absolute windows", file: `C:\tmp\proposal.md`, code: AIOutputParseFindingCodeAbsoluteFileName},
		{name: "nested slash", file: "nested/proposal.md", code: AIOutputParseFindingCodeNestedFileName},
		{name: "nested backslash", file: `nested\proposal.md`, code: AIOutputParseFindingCodeNestedFileName},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseAIOutputBlocks(block(test.file, "# Bad\n\nContent."))
			assertDomainFindingCode(t, result, test.code)
			assertDomainFindingCode(t, result, AIOutputParseFindingCodeMissingFileBlock)
			if len(result.Files()) != 0 {
				t.Fatalf("Files() = %v, want no accepted files", result.Files())
			}
		})
	}
}

func TestParseAIOutputBlocksRejectsDuplicateMissingAndEmptyBlocks(t *testing.T) {
	duplicate := strictAIOutputFixture()
	duplicate += "\n" + block("proposal.md", "# Proposal\n\nDuplicate.")
	assertDomainFindingCode(t, ParseAIOutputBlocks(duplicate), AIOutputParseFindingCodeDuplicateFileBlock)

	missing := strings.Replace(strictAIOutputFixture(), block("risks.md", validAIFileContent("risks.md")), "", 1)
	assertDomainFindingCode(t, ParseAIOutputBlocks(missing), AIOutputParseFindingCodeMissingFileBlock)

	empty := strings.Replace(strictAIOutputFixture(), block("design.md", validAIFileContent("design.md")), block("design.md", "  \n\t"), 1)
	result := ParseAIOutputBlocks(empty)
	assertDomainFindingCode(t, result, AIOutputParseFindingCodeEmptyFileContent)
	finding := firstDomainFindingByCode(t, result, AIOutputParseFindingCodeEmptyFileContent)
	if finding.FileName != "design.md" || finding.Line == 0 {
		t.Fatalf("empty finding = %+v, want design.md with line", finding)
	}
}

func TestParseAIOutputBlocksRejectsMalformedSyntax(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   AIOutputParseFindingCode
	}{
		{name: "malformed start", source: "---FILE:proposal.md---\n# Proposal\n---END FILE---", code: AIOutputParseFindingCodeMalformedFileBlock},
		{name: "orphan end", source: "---END FILE---", code: AIOutputParseFindingCodeMalformedFileBlock},
		{name: "unclosed", source: "---FILE: proposal.md---\n# Proposal\n", code: AIOutputParseFindingCodeUnclosedFileBlock},
		{name: "text outside", source: "Here are the files:\n" + strictAIOutputFixture(), code: AIOutputParseFindingCodeTextOutsideFileBlock},
		{name: "fenced wrapper", source: "```text\n" + strictAIOutputFixture() + "\n```", code: AIOutputParseFindingCodeTextOutsideFileBlock},
		{name: "diff format", source: "diff --git a/proposal.md b/proposal.md\n+content\n", code: AIOutputParseFindingCodeTextOutsideFileBlock},
		{name: "nested start", source: "---FILE: proposal.md---\n---FILE: design.md---\n# Design\n---END FILE---", code: AIOutputParseFindingCodeMalformedFileBlock},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertDomainFindingCode(t, ParseAIOutputBlocks(test.source), test.code)
		})
	}
}

func TestAIOutputParseResultDefensivelyCopies(t *testing.T) {
	files := []AIOutputFileBlock{{FileName: "proposal.md", Content: "# Proposal"}}
	findings := []AIOutputParseFinding{{
		Severity: AIOutputParseFindingSeverityError,
		Code:     AIOutputParseFindingCodeMissingFileBlock,
		FileName: "design.md",
	}}

	result := NewAIOutputParseResult(files, findings)
	files[0].FileName = "mutated.md"
	findings[0].FileName = "mutated.md"

	if result.Files()[0].FileName != "proposal.md" {
		t.Fatalf("Files() did not preserve constructor copy: %v", result.Files())
	}
	if result.Findings()[0].FileName != "design.md" {
		t.Fatalf("Findings() did not preserve constructor copy: %v", result.Findings())
	}

	copiedFiles := result.Files()
	copiedFindings := result.Findings()
	copiedFiles[0].FileName = "mutated.md"
	copiedFindings[0].FileName = "mutated.md"

	if result.Files()[0].FileName != "proposal.md" || result.Findings()[0].FileName != "design.md" {
		t.Fatalf("accessors did not defensively copy")
	}
}

func TestAIAssistedGenerationResultModelDefensivelyCopies(t *testing.T) {
	generated := []string{"proposal.md"}
	skipped := []string{"design.md"}
	overwritten := []string{"tasks.md"}
	validation := NewValidationResult("change", "openspec/changes/change", []string{"proposal.md"}, []ValidationFinding{{
		Severity:     ValidationFindingSeverityWarning,
		Code:         ValidationFindingCodeRisksMitigationMissing,
		RelativePath: "openspec/changes/change/risks.md",
	}})

	result := NewAIAssistedGenerationResult(
		"change",
		"agent-output.txt",
		"openspec/changes/change",
		true,
		true,
		generated,
		skipped,
		overwritten,
		validation,
	)

	generated[0] = "mutated.md"
	skipped[0] = "mutated.md"
	overwritten[0] = "mutated.md"
	validation.RequiredFiles[0] = "mutated.md"
	validation.Findings[0].RelativePath = "mutated.md"

	assertDomainStringsEqual(t, result.GeneratedFiles(), []string{"proposal.md"})
	assertDomainStringsEqual(t, result.SkippedFiles(), []string{"design.md"})
	assertDomainStringsEqual(t, result.OverwrittenFiles(), []string{"tasks.md"})
	gotValidation, ok := result.ValidationResult()
	if !ok {
		t.Fatalf("ValidationResult() ok = false, want true")
	}
	if gotValidation.RequiredFiles[0] != "proposal.md" || gotValidation.Findings[0].RelativePath != "openspec/changes/change/risks.md" {
		t.Fatalf("validation copy = %+v, want original values", gotValidation)
	}

	result.GeneratedFiles()[0] = "mutated.md"
	result.SkippedFiles()[0] = "mutated.md"
	result.OverwrittenFiles()[0] = "mutated.md"
	gotValidation.RequiredFiles[0] = "mutated.md"
	gotValidation.Findings[0].RelativePath = "mutated.md"

	if result.GeneratedFiles()[0] != "proposal.md" ||
		result.SkippedFiles()[0] != "design.md" ||
		result.OverwrittenFiles()[0] != "tasks.md" {
		t.Fatalf("result accessors did not defensively copy")
	}
	gotValidation, _ = result.ValidationResult()
	if gotValidation.RequiredFiles[0] != "proposal.md" || gotValidation.Findings[0].RelativePath != "openspec/changes/change/risks.md" {
		t.Fatalf("validation accessor did not defensively copy")
	}
	if result.ProviderAPIsCalled ||
		result.RemoteAIServicesCalled ||
		result.AgentCommandsExecuted ||
		result.ProductionCodeModified ||
		result.VCSCommandsRun ||
		result.AutomationPerformed {
		t.Fatalf("safety flags = %+v, want all false", result)
	}
}

func strictAIOutputFixture() string {
	var parts []string
	for _, fileName := range RequiredOpenSpecChangeFiles() {
		parts = append(parts, block(fileName, validAIFileContent(fileName)))
	}
	return strings.Join(parts, "\n")
}

func block(fileName string, content string) string {
	return "---FILE: " + fileName + "---\n" + content + "\n---END FILE---"
}

func validAIFileContent(fileName string) string {
	switch fileName {
	case "proposal.md":
		return "# Proposal\n\n## Problem\n\nUsers need a safe import path.\n\n## Goal\n\nImport strict file blocks."
	case "design.md":
		return "# Design\n\n## Overview\n\nParse first, then write approved files.\n\n## Architecture\n\nDomain parser and use case orchestration."
	case "tasks.md":
		return "# Tasks\n\n## Phase 1\n\n- [ ] Implement parser.\n- [ ] Update docs."
	case "acceptance-criteria.md":
		return "# Acceptance Criteria\n\n- Strict blocks are imported safely."
	case "risks.md":
		return "# Risks\n\n## Risks\n\n- Malformed output could be ambiguous.\n\n## Mitigations\n\n- Reject before writes."
	default:
		return "# " + fileName + "\n\nContent."
	}
}

func assertDomainFindingCode(t *testing.T, result AIOutputParseResult, code AIOutputParseFindingCode) {
	t.Helper()
	firstDomainFindingByCode(t, result, code)
}

func firstDomainFindingByCode(t *testing.T, result AIOutputParseResult, code AIOutputParseFindingCode) AIOutputParseFinding {
	t.Helper()
	for _, finding := range result.Findings() {
		if finding.Code == code {
			if finding.Severity != AIOutputParseFindingSeverityError {
				t.Fatalf("finding severity = %q, want error", finding.Severity)
			}
			if finding.Message == "" {
				t.Fatalf("finding message is empty for code %s", code)
			}
			return finding
		}
	}
	t.Fatalf("findings = %v, want code %s", result.Findings(), code)
	return AIOutputParseFinding{}
}

func assertDomainStringsEqual(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings = %v, want %v", got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("strings = %v, want %v", got, want)
		}
	}
}
