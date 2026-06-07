package domain

type ReviewStatus string

const (
	ReviewStatusApproved  ReviewStatus = "approved"
	ReviewStatusNeedsWork ReviewStatus = "needs-work"
	ReviewStatusInvalid   ReviewStatus = "invalid"
)

type ReviewFindingSeverity string

const (
	ReviewFindingSeverityError   ReviewFindingSeverity = "error"
	ReviewFindingSeverityWarning ReviewFindingSeverity = "warning"
)

type ReviewFindingCode string

const (
	ReviewFindingCodeProjectRootUnavailable ReviewFindingCode = "project_root_unavailable"
	ReviewFindingCodeChangeDirectoryMissing ReviewFindingCode = "change_directory_missing"
	ReviewFindingCodeRequiredFileMissing    ReviewFindingCode = "required_file_missing"
	ReviewFindingCodeTasksFileUnreadable    ReviewFindingCode = "tasks_file_unreadable"
	ReviewFindingCodeTasksNotFound          ReviewFindingCode = "tasks_not_found"
	ReviewFindingCodeIncompleteTask         ReviewFindingCode = "incomplete_task"
)

type ReviewFinding struct {
	Severity     ReviewFindingSeverity
	Code         ReviewFindingCode
	Message      string
	RelativePath string
	Subject      string
}

type ReviewTaskSummary struct {
	Total      int
	Completed  int
	Incomplete int
}

func NewReviewTaskSummary(completed int, incomplete int) ReviewTaskSummary {
	return ReviewTaskSummary{
		Total:      completed + incomplete,
		Completed:  completed,
		Incomplete: incomplete,
	}
}

type ReviewResult struct {
	ChangeID      string
	CheckedPath   string
	Status        ReviewStatus
	RequiredFiles []string
	Tasks         ReviewTaskSummary
	Findings      []ReviewFinding
}

func NewReviewResult(changeID string, checkedPath string, requiredFiles []string, tasks ReviewTaskSummary, findings []ReviewFinding) ReviewResult {
	result := ReviewResult{
		ChangeID:      changeID,
		CheckedPath:   checkedPath,
		Status:        calculateReviewStatus(tasks, findings),
		RequiredFiles: append([]string(nil), requiredFiles...),
		Tasks:         tasks,
		Findings:      append([]ReviewFinding(nil), findings...),
	}

	return result
}

func calculateReviewStatus(tasks ReviewTaskSummary, findings []ReviewFinding) ReviewStatus {
	for _, finding := range findings {
		if finding.Severity == ReviewFindingSeverityError {
			return ReviewStatusInvalid
		}
	}

	if tasks.Total == 0 {
		return ReviewStatusInvalid
	}

	if tasks.Incomplete > 0 {
		return ReviewStatusNeedsWork
	}

	for _, finding := range findings {
		if finding.Severity == ReviewFindingSeverityWarning {
			return ReviewStatusNeedsWork
		}
	}

	return ReviewStatusApproved
}
