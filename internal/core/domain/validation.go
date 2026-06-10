package domain

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationFindingSeverity string

const (
	ValidationFindingSeverityError   ValidationFindingSeverity = "error"
	ValidationFindingSeverityWarning ValidationFindingSeverity = "warning"
)

type ValidationFindingCode string

const (
	ValidationFindingCodeProjectRootUnavailable ValidationFindingCode = "project_root_unavailable"
	ValidationFindingCodeChangeDirectoryMissing ValidationFindingCode = "change_directory_missing"
	ValidationFindingCodeRequiredFileMissing    ValidationFindingCode = "required_file_missing"

	ValidationFindingCodeFileEmpty                        ValidationFindingCode = "file_empty"
	ValidationFindingCodeFileMissingHeading               ValidationFindingCode = "file_missing_heading"
	ValidationFindingCodeFileMissingBody                  ValidationFindingCode = "file_missing_body"
	ValidationFindingCodePlaceholderContent               ValidationFindingCode = "placeholder_content"
	ValidationFindingCodeBoilerplateOnlyContent           ValidationFindingCode = "boilerplate_only_content"
	ValidationFindingCodeProposalSectionMissing           ValidationFindingCode = "proposal_section_missing"
	ValidationFindingCodeDesignSectionMissing             ValidationFindingCode = "design_section_missing"
	ValidationFindingCodeTasksCheckboxMissing             ValidationFindingCode = "tasks_checkbox_missing"
	ValidationFindingCodeTasksCheckboxMalformed           ValidationFindingCode = "tasks_checkbox_malformed"
	ValidationFindingCodeTasksPhaseHeadingMissing         ValidationFindingCode = "tasks_phase_heading_missing"
	ValidationFindingCodeTasksAllCompleted                ValidationFindingCode = "tasks_all_completed"
	ValidationFindingCodeAcceptanceCriteriaItemMissing    ValidationFindingCode = "acceptance_criteria_item_missing"
	ValidationFindingCodeRisksMitigationMissing           ValidationFindingCode = "risks_mitigation_missing"
	ValidationFindingCodeDesignArchitectureSectionMissing ValidationFindingCode = "design_architecture_section_missing"
	ValidationFindingCodeTasksDocumentationTaskMissing    ValidationFindingCode = "tasks_documentation_task_missing"
)

type ValidationFinding struct {
	Severity     ValidationFindingSeverity
	Code         ValidationFindingCode
	Message      string
	RelativePath string
	Subject      string
}

type ValidationResult struct {
	ChangeID      string
	CheckedPath   string
	Status        ValidationStatus
	RequiredFiles []string
	Findings      []ValidationFinding
}

func NewValidationResult(changeID string, checkedPath string, requiredFiles []string, findings []ValidationFinding) ValidationResult {
	result := ValidationResult{
		ChangeID:      changeID,
		CheckedPath:   checkedPath,
		Status:        ValidationStatusValid,
		RequiredFiles: append([]string(nil), requiredFiles...),
		Findings:      append([]ValidationFinding(nil), findings...),
	}

	if result.ErrorCount() > 0 {
		result.Status = ValidationStatusInvalid
	}

	return result
}

func (result ValidationResult) ErrorCount() int {
	return result.countBySeverity(ValidationFindingSeverityError)
}

func (result ValidationResult) WarningCount() int {
	return result.countBySeverity(ValidationFindingSeverityWarning)
}

func (result ValidationResult) countBySeverity(severity ValidationFindingSeverity) int {
	count := 0
	for _, finding := range result.Findings {
		if finding.Severity == severity {
			count++
		}
	}
	return count
}

func RequiredOpenSpecChangeFiles() []string {
	return []string{
		"proposal.md",
		"design.md",
		"tasks.md",
		"acceptance-criteria.md",
		"risks.md",
	}
}
