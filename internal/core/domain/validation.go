package domain

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

type ValidationFindingSeverity string

const (
	ValidationFindingSeverityError ValidationFindingSeverity = "error"
)

type ValidationFindingCode string

const (
	ValidationFindingCodeProjectRootUnavailable ValidationFindingCode = "project_root_unavailable"
	ValidationFindingCodeChangeDirectoryMissing ValidationFindingCode = "change_directory_missing"
	ValidationFindingCodeRequiredFileMissing    ValidationFindingCode = "required_file_missing"
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

	if len(result.Findings) > 0 {
		result.Status = ValidationStatusInvalid
	}

	return result
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
