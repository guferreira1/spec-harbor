package domain

import "testing"

func TestBlankGenerationModeValue(t *testing.T) {
	if BlankMode != "blank" {
		t.Fatalf("BlankMode = %q, want blank", BlankMode)
	}
}

func TestNewGenerationResultCopiesFileSlices(t *testing.T) {
	created := []string{"proposal.md"}
	skipped := []string{"design.md"}

	result := NewGenerationResult("change", BlankMode, "openspec/changes/change", true, created, skipped)
	created[0] = "mutated.md"
	skipped[0] = "mutated.md"

	if result.CreatedFiles()[0] != "proposal.md" {
		t.Fatalf("CreatedFiles()[0] = %q, want proposal.md", result.CreatedFiles()[0])
	}
	if result.SkippedExistingFiles()[0] != "design.md" {
		t.Fatalf("SkippedExistingFiles()[0] = %q, want design.md", result.SkippedExistingFiles()[0])
	}
}

func TestGenerationResultAccessorsReturnCopies(t *testing.T) {
	result := NewGenerationResult(
		"change",
		BlankMode,
		"openspec/changes/change",
		false,
		[]string{"proposal.md"},
		[]string{"design.md"},
	)

	created := result.CreatedFiles()
	skipped := result.SkippedExistingFiles()
	created[0] = "mutated.md"
	skipped[0] = "mutated.md"

	if result.CreatedFiles()[0] != "proposal.md" {
		t.Fatalf("CreatedFiles()[0] = %q, want proposal.md", result.CreatedFiles()[0])
	}
	if result.SkippedExistingFiles()[0] != "design.md" {
		t.Fatalf("SkippedExistingFiles()[0] = %q, want design.md", result.SkippedExistingFiles()[0])
	}
}
