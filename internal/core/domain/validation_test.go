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
