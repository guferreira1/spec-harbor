package domain

import "testing"

func TestNewReviewResultSetsStatusFromFindingsAndTasks(t *testing.T) {
	requiredFiles := RequiredOpenSpecChangeFiles()

	approved := NewReviewResult("change", "openspec/changes/change", requiredFiles, NewReviewTaskSummary(2, 0), nil)
	if approved.Status != ReviewStatusApproved {
		t.Fatalf("approved Status = %q, want %q", approved.Status, ReviewStatusApproved)
	}

	needsWork := NewReviewResult("change", "openspec/changes/change", requiredFiles, NewReviewTaskSummary(1, 1), []ReviewFinding{
		{
			Severity: ReviewFindingSeverityWarning,
			Code:     ReviewFindingCodeIncompleteTask,
			Message:  "Task is not completed: Run tests",
			Subject:  "Run tests",
		},
	})
	if needsWork.Status != ReviewStatusNeedsWork {
		t.Fatalf("needs-work Status = %q, want %q", needsWork.Status, ReviewStatusNeedsWork)
	}

	invalid := NewReviewResult("change", "openspec/changes/change", requiredFiles, ReviewTaskSummary{}, []ReviewFinding{
		{
			Severity: ReviewFindingSeverityError,
			Code:     ReviewFindingCodeRequiredFileMissing,
			Message:  "Missing required file: risks.md",
			Subject:  "risks.md",
		},
	})
	if invalid.Status != ReviewStatusInvalid {
		t.Fatalf("invalid Status = %q, want %q", invalid.Status, ReviewStatusInvalid)
	}

	noTasks := NewReviewResult("change", "openspec/changes/change", requiredFiles, ReviewTaskSummary{}, nil)
	if noTasks.Status != ReviewStatusInvalid {
		t.Fatalf("no-task Status = %q, want %q", noTasks.Status, ReviewStatusInvalid)
	}
}

func TestNewReviewResultCopiesSlices(t *testing.T) {
	requiredFiles := []string{"proposal.md", "tasks.md"}
	findings := []ReviewFinding{
		{
			Severity: ReviewFindingSeverityWarning,
			Code:     ReviewFindingCodeIncompleteTask,
			Message:  "Task is not completed: Run tests",
			Subject:  "Run tests",
		},
	}

	result := NewReviewResult("change", "openspec/changes/change", requiredFiles, NewReviewTaskSummary(1, 1), findings)
	requiredFiles[0] = "mutated.md"
	findings[0].Subject = "mutated"

	if result.RequiredFiles[0] != "proposal.md" {
		t.Fatalf("RequiredFiles[0] = %q, want proposal.md", result.RequiredFiles[0])
	}
	if result.Findings[0].Subject != "Run tests" {
		t.Fatalf("Findings[0].Subject = %q, want Run tests", result.Findings[0].Subject)
	}
}

func TestNewReviewTaskSummaryCountsTotalTasks(t *testing.T) {
	summary := NewReviewTaskSummary(3, 2)
	if summary.Total != 5 {
		t.Fatalf("Total = %d, want 5", summary.Total)
	}
	if summary.Completed != 3 {
		t.Fatalf("Completed = %d, want 3", summary.Completed)
	}
	if summary.Incomplete != 2 {
		t.Fatalf("Incomplete = %d, want 2", summary.Incomplete)
	}
}
