package domain

import (
	"strings"
	"testing"
)

func TestHybridSourceSelectionAcceptsExactlyOneSource(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		customName   string
		alias        string
		wantKind     HybridSelectedSourceKind
		wantName     string
	}{
		{name: "built-in", templateName: "feature", wantKind: HybridSelectedSourceBuiltIn, wantName: "feature"},
		{name: "custom", customName: "api-feature", wantKind: HybridSelectedSourceCustom, wantName: "api-feature"},
		{name: "config", alias: "default-feature", wantKind: HybridSelectedSourceConfig, wantName: "default-feature"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selection, err := NewHybridSourceSelection(test.templateName, test.customName, test.alias)
			if err != nil {
				t.Fatalf("NewHybridSourceSelection() error = %v", err)
			}
			if selection.Kind() != test.wantKind {
				t.Fatalf("Kind() = %q, want %q", selection.Kind(), test.wantKind)
			}
			if selection.Name() != test.wantName {
				t.Fatalf("Name() = %q, want %q", selection.Name(), test.wantName)
			}
		})
	}
}

func TestHybridSourceSelectionRejectsMissingAndMultipleSources(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		customName   string
		alias        string
		want         string
	}{
		{name: "missing", want: "hybrid source selector is required"},
		{name: "multiple", templateName: "feature", customName: "api-feature", want: "hybrid requires exactly one source selector"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewHybridSourceSelection(test.templateName, test.customName, test.alias)
			if err == nil || err.Error() != test.want {
				t.Fatalf("NewHybridSourceSelection() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestHybridSourceSelectionUsesExistingValueObjectValidation(t *testing.T) {
	tests := []struct {
		name         string
		templateName string
		customName   string
		alias        string
		want         string
	}{
		{name: "built-in", templateName: "maintenance", want: "unknown template name: maintenance"},
		{name: "custom", customName: "../escape", want: "custom template name must be a single path segment"},
		{name: "config", alias: "../escape", want: "config template alias must be a single path segment"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewHybridSourceSelection(test.templateName, test.customName, test.alias)
			if err == nil || err.Error() != test.want {
				t.Fatalf("NewHybridSourceSelection() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestHybridMetadataRequiresTitleAndSummaryAndValidatesOptionalType(t *testing.T) {
	metadata, err := NewHybridMetadata("  Add login  ", "  Add an OpenSpec change.  ", "feature")
	if err != nil {
		t.Fatalf("NewHybridMetadata() error = %v", err)
	}
	if metadata.Title() != "Add login" {
		t.Fatalf("Title() = %q, want Add login", metadata.Title())
	}
	if metadata.Summary() != "Add an OpenSpec change." {
		t.Fatalf("Summary() = %q, want trimmed summary", metadata.Summary())
	}
	if providedType, ok := metadata.ProvidedType(); !ok || providedType != HybridTypeFeature {
		t.Fatalf("ProvidedType() = %q, %t; want feature, true", providedType, ok)
	}

	tests := []struct {
		name    string
		title   string
		summary string
		value   string
		want    string
	}{
		{name: "missing title", title: " ", summary: "Summary", want: "hybrid title is required"},
		{name: "missing summary", title: "Title", summary: " ", want: "hybrid summary is required"},
		{name: "unsupported type", title: "Title", summary: "Summary", value: "maintenance", want: "unsupported hybrid type: maintenance"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewHybridMetadata(test.title, test.summary, test.value)
			if err == nil || err.Error() != test.want {
				t.Fatalf("NewHybridMetadata() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestParseHybridTypeAcceptsSupportedValues(t *testing.T) {
	for _, value := range []struct {
		raw  string
		want HybridType
	}{
		{raw: "feature", want: HybridTypeFeature},
		{raw: "bugfix", want: HybridTypeBugfix},
		{raw: "docs", want: HybridTypeDocs},
		{raw: "refactor", want: HybridTypeRefactor},
	} {
		t.Run(value.raw, func(t *testing.T) {
			got, err := ParseHybridType(value.raw)
			if err != nil {
				t.Fatalf("ParseHybridType(%q) error = %v", value.raw, err)
			}
			if got != value.want {
				t.Fatalf("ParseHybridType(%q) = %q, want %q", value.raw, got, value.want)
			}
		})
	}
}

func TestHybridMetadataEffectiveTypeRules(t *testing.T) {
	metadata, err := NewHybridMetadata("Title", "Summary", "")
	if err != nil {
		t.Fatalf("NewHybridMetadata() error = %v", err)
	}
	metadata, err = metadata.WithBuiltInEffectiveType(FeatureTemplate)
	if err != nil {
		t.Fatalf("WithBuiltInEffectiveType() error = %v", err)
	}
	if effectiveType, ok := metadata.EffectiveType(); !ok || effectiveType != HybridTypeFeature {
		t.Fatalf("EffectiveType() = %q, %t; want feature, true", effectiveType, ok)
	}
	if !metadata.TypeIsDerived() {
		t.Fatalf("TypeIsDerived() = false, want true")
	}

	configMetadata, err := NewHybridMetadata("Title", "Summary", "")
	if err != nil {
		t.Fatalf("NewHybridMetadata() error = %v", err)
	}
	configMetadata, err = configMetadata.WithBuiltInEffectiveType(DocsTemplate)
	if err != nil {
		t.Fatalf("WithBuiltInEffectiveType(config built-in) error = %v", err)
	}
	if effectiveType, ok := configMetadata.EffectiveType(); !ok || effectiveType != HybridTypeDocs {
		t.Fatalf("config EffectiveType() = %q, %t; want docs, true", effectiveType, ok)
	}

	matching, err := NewHybridMetadata("Title", "Summary", "bugfix")
	if err != nil {
		t.Fatalf("NewHybridMetadata(matching) error = %v", err)
	}
	if _, err := matching.WithBuiltInEffectiveType(BugfixTemplate); err != nil {
		t.Fatalf("WithBuiltInEffectiveType(matching) error = %v", err)
	}

	mismatched, err := NewHybridMetadata("Title", "Summary", "bugfix")
	if err != nil {
		t.Fatalf("NewHybridMetadata(mismatched) error = %v", err)
	}
	_, err = mismatched.WithBuiltInEffectiveType(FeatureTemplate)
	if err == nil || err.Error() != "hybrid type bugfix does not match built-in template feature" {
		t.Fatalf("WithBuiltInEffectiveType(mismatched) error = %v", err)
	}
}

func TestHybridMetadataNonBuiltInTypeRulesAndRendering(t *testing.T) {
	withoutType, err := NewHybridMetadata("Add login", "Add authentication.", "")
	if err != nil {
		t.Fatalf("NewHybridMetadata() error = %v", err)
	}
	withoutType = withoutType.WithoutDerivedType()
	if _, ok := withoutType.EffectiveType(); ok {
		t.Fatalf("EffectiveType() exists for non-built-in omitted type")
	}
	rendered := withoutType.RenderContent("{{change_id}} {{title}} {{summary}} {{type}} {{unknown}}", "add-login")
	want := "add-login Add login Add authentication. {{type}} {{unknown}}"
	if rendered != want {
		t.Fatalf("RenderContent() = %q, want %q", rendered, want)
	}

	withType, err := NewHybridMetadata("Add login", "Add authentication.", "refactor")
	if err != nil {
		t.Fatalf("NewHybridMetadata() error = %v", err)
	}
	withType = withType.WithoutDerivedType()
	rendered = withType.RenderContent("{{change_id}} {{title}} {{summary}} {{type}} {{unknown}}", "add-login")
	want = "add-login Add login Add authentication. refactor {{unknown}}"
	if rendered != want {
		t.Fatalf("RenderContent(with type) = %q, want %q", rendered, want)
	}
}

func TestHybridGenerationResultMappingAndDefensiveCopies(t *testing.T) {
	selection, err := NewHybridSourceSelection("feature", "", "")
	if err != nil {
		t.Fatalf("NewHybridSourceSelection() error = %v", err)
	}
	metadata, err := NewHybridMetadata("Title", "Summary", "")
	if err != nil {
		t.Fatalf("NewHybridMetadata() error = %v", err)
	}
	metadata, err = metadata.WithBuiltInEffectiveType(FeatureTemplate)
	if err != nil {
		t.Fatalf("WithBuiltInEffectiveType() error = %v", err)
	}
	validation := NewValidationResult("change", "openspec/changes/change", []string{"proposal.md"}, []ValidationFinding{{
		Severity:     ValidationFindingSeverityWarning,
		Code:         ValidationFindingCodeRisksMitigationMissing,
		Message:      "warning",
		RelativePath: "risks.md",
	}})
	created := []string{"proposal.md"}
	skipped := []string{"design.md"}

	result := NewHybridGenerationResult(
		"change",
		selection,
		HybridResolvedSourceBuiltin,
		"feature",
		metadata,
		"openspec/changes/change",
		true,
		created,
		skipped,
		validation,
		HybridRemoteFacts{},
	)

	if result.Mode != HybridMode || result.SelectedSourceKind != HybridSelectedSourceBuiltIn || result.ResolvedSourceKind != HybridResolvedSourceBuiltin {
		t.Fatalf("result mapping = %+v, want hybrid built-in mapping", result)
	}
	if result.EffectiveType != HybridTypeFeature || !result.TypeAvailable {
		t.Fatalf("result type = %q %t, want feature true", result.EffectiveType, result.TypeAvailable)
	}

	created[0] = "mutated.md"
	skipped[0] = "mutated.md"
	validation.RequiredFiles[0] = "mutated.md"

	gotCreated := result.CreatedFiles()
	gotSkipped := result.SkippedExistingFiles()
	gotValidation, ok := result.ValidationResult()
	if !ok {
		t.Fatalf("ValidationResult() ok = false, want true")
	}
	if gotCreated[0] != "proposal.md" || gotSkipped[0] != "design.md" || gotValidation.RequiredFiles[0] != "proposal.md" {
		t.Fatalf("defensive copies failed: created=%v skipped=%v validation=%v", gotCreated, gotSkipped, gotValidation.RequiredFiles)
	}

	gotCreated[0] = "mutated.md"
	gotSkipped[0] = "mutated.md"
	gotValidation.RequiredFiles[0] = "mutated.md"
	if strings.Join(result.CreatedFiles(), ",") != "proposal.md" ||
		strings.Join(result.SkippedExistingFiles(), ",") != "design.md" {
		t.Fatalf("accessor defensive copies failed")
	}
	gotValidationAgain, _ := result.ValidationResult()
	if gotValidationAgain.RequiredFiles[0] != "proposal.md" {
		t.Fatalf("validation accessor defensive copy failed")
	}
}
