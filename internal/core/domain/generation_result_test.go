package domain

import "testing"

func TestBlankGenerationModeValue(t *testing.T) {
	if BlankMode != "blank" {
		t.Fatalf("BlankMode = %q, want blank", BlankMode)
	}
}

func TestTemplateGenerationModeValue(t *testing.T) {
	if TemplateMode != "template" {
		t.Fatalf("TemplateMode = %q, want template", TemplateMode)
	}
}

func TestGuidedGenerationModeValue(t *testing.T) {
	if GuidedMode != "guided" {
		t.Fatalf("GuidedMode = %q, want guided", GuidedMode)
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

func TestNewTemplateGenerationResultSetsTemplateModeAndName(t *testing.T) {
	result := NewTemplateGenerationResult(
		"change",
		FeatureTemplate,
		"openspec/changes/change",
		true,
		[]string{"proposal.md"},
		[]string{"design.md"},
	)

	if result.Mode != TemplateMode {
		t.Fatalf("Mode = %q, want %q", result.Mode, TemplateMode)
	}
	if result.TemplateName != FeatureTemplate {
		t.Fatalf("TemplateName = %q, want %q", result.TemplateName, FeatureTemplate)
	}
}

func TestNewGuidedGenerationResultSetsGuidedModeTypeAndTitle(t *testing.T) {
	result := NewGuidedGenerationResult(
		"change",
		FeatureGuidedType,
		"Add reports",
		"openspec/changes/change",
		true,
		[]string{"proposal.md"},
		[]string{"design.md"},
	)

	if result.Mode != GuidedMode {
		t.Fatalf("Mode = %q, want %q", result.Mode, GuidedMode)
	}
	if result.GuidedType != FeatureGuidedType {
		t.Fatalf("GuidedType = %q, want %q", result.GuidedType, FeatureGuidedType)
	}
	if result.GuidedTitle != "Add reports" {
		t.Fatalf("GuidedTitle = %q, want Add reports", result.GuidedTitle)
	}
}
