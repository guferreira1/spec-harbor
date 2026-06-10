package domain

import "testing"

func TestNewValidationResultSetsStatusFromFindings(t *testing.T) {
	validResult := NewValidationResult("change", "openspec/changes/change", RequiredOpenSpecChangeFiles(), nil)
	if validResult.Status != ValidationStatusValid {
		t.Fatalf("valid Status = %q, want %q", validResult.Status, ValidationStatusValid)
	}

	invalidResult := NewValidationResult("change", "openspec/changes/change", RequiredOpenSpecChangeFiles(), []ValidationFinding{
		{
			Severity: ValidationFindingSeverityError,
			Code:     ValidationFindingCodeRequiredFileMissing,
			Message:  "Missing required file: proposal.md",
			Subject:  "proposal.md",
		},
	})
	if invalidResult.Status != ValidationStatusInvalid {
		t.Fatalf("invalid Status = %q, want %q", invalidResult.Status, ValidationStatusInvalid)
	}
}

func TestNewValidationResultKeepsWarningsOnlyResultValid(t *testing.T) {
	result := NewValidationResult("change", "openspec/changes/change", RequiredOpenSpecChangeFiles(), []ValidationFinding{
		{Severity: ValidationFindingSeverityWarning, Code: ValidationFindingCodeBoilerplateOnlyContent, Message: "Only starter boilerplate content found."},
		{Severity: ValidationFindingSeverityWarning, Code: ValidationFindingCodePlaceholderContent, Message: `Placeholder marker "TBD" found (line 3)`},
	})

	if result.Status != ValidationStatusValid {
		t.Fatalf("warnings-only Status = %q, want %q", result.Status, ValidationStatusValid)
	}
	if result.ErrorCount() != 0 {
		t.Fatalf("ErrorCount() = %d, want 0", result.ErrorCount())
	}
	if result.WarningCount() != 2 {
		t.Fatalf("WarningCount() = %d, want 2", result.WarningCount())
	}
}

func TestNewValidationResultCountsMixedSeverities(t *testing.T) {
	result := NewValidationResult("change", "openspec/changes/change", RequiredOpenSpecChangeFiles(), []ValidationFinding{
		{Severity: ValidationFindingSeverityError, Code: ValidationFindingCodeFileEmpty},
		{Severity: ValidationFindingSeverityWarning, Code: ValidationFindingCodeRisksMitigationMissing},
		{Severity: ValidationFindingSeverityError, Code: ValidationFindingCodeTasksCheckboxMissing},
	})

	if result.Status != ValidationStatusInvalid {
		t.Fatalf("mixed Status = %q, want %q", result.Status, ValidationStatusInvalid)
	}
	if result.ErrorCount() != 2 {
		t.Fatalf("ErrorCount() = %d, want 2", result.ErrorCount())
	}
	if result.WarningCount() != 1 {
		t.Fatalf("WarningCount() = %d, want 1", result.WarningCount())
	}
}

func TestValidationFindingCodesAreStable(t *testing.T) {
	codes := map[ValidationFindingCode]string{
		ValidationFindingCodeProjectRootUnavailable:           "project_root_unavailable",
		ValidationFindingCodeChangeDirectoryMissing:           "change_directory_missing",
		ValidationFindingCodeRequiredFileMissing:              "required_file_missing",
		ValidationFindingCodeFileEmpty:                        "file_empty",
		ValidationFindingCodeFileMissingHeading:               "file_missing_heading",
		ValidationFindingCodeFileMissingBody:                  "file_missing_body",
		ValidationFindingCodePlaceholderContent:               "placeholder_content",
		ValidationFindingCodeBoilerplateOnlyContent:           "boilerplate_only_content",
		ValidationFindingCodeProposalSectionMissing:           "proposal_section_missing",
		ValidationFindingCodeDesignSectionMissing:             "design_section_missing",
		ValidationFindingCodeTasksCheckboxMissing:             "tasks_checkbox_missing",
		ValidationFindingCodeTasksCheckboxMalformed:           "tasks_checkbox_malformed",
		ValidationFindingCodeTasksPhaseHeadingMissing:         "tasks_phase_heading_missing",
		ValidationFindingCodeTasksAllCompleted:                "tasks_all_completed",
		ValidationFindingCodeAcceptanceCriteriaItemMissing:    "acceptance_criteria_item_missing",
		ValidationFindingCodeRisksMitigationMissing:           "risks_mitigation_missing",
		ValidationFindingCodeDesignArchitectureSectionMissing: "design_architecture_section_missing",
		ValidationFindingCodeTasksDocumentationTaskMissing:    "tasks_documentation_task_missing",
	}

	for code, want := range codes {
		if string(code) != want {
			t.Fatalf("finding code = %q, want %q", string(code), want)
		}
	}

	if string(ValidationFindingSeverityError) != "error" {
		t.Fatalf("error severity = %q, want %q", string(ValidationFindingSeverityError), "error")
	}
	if string(ValidationFindingSeverityWarning) != "warning" {
		t.Fatalf("warning severity = %q, want %q", string(ValidationFindingSeverityWarning), "warning")
	}
}

func TestRequiredOpenSpecChangeFiles(t *testing.T) {
	files := RequiredOpenSpecChangeFiles()
	want := []string{"proposal.md", "design.md", "tasks.md", "acceptance-criteria.md", "risks.md"}

	if len(files) != len(want) {
		t.Fatalf("required file count = %d, want %d", len(files), len(want))
	}
	for index, file := range want {
		if files[index] != file {
			t.Fatalf("required file %d = %q, want %q", index, files[index], file)
		}
	}

	files[0] = "mutated.md"
	reloadedFiles := RequiredOpenSpecChangeFiles()
	if reloadedFiles[0] != "proposal.md" {
		t.Fatalf("RequiredOpenSpecChangeFiles() returned mutable policy, first file = %q", reloadedFiles[0])
	}
}
