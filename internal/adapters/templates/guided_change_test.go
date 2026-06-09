package templates

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
	"github.com/guferreira1/spec-harbor/internal/core/ports"
)

var _ ports.GuidedChangeContent = (*GuidedChangeTemplates)(nil)

func TestGuidedChangeTemplatesProvideEveryRequiredFile(t *testing.T) {
	content := NewGuidedChangeTemplates()
	expectedMarkdown := map[domain.GuidedType]map[string][]string{
		domain.FeatureGuidedType: {
			"proposal.md": {
				"# Proposal: Add reports",
				"## Summary",
				"## Problem",
				"## Proposed Solution",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design: Add reports",
				"## Context",
				"## Architecture Notes",
				"## Domain",
				"## Ports",
				"## Use Case",
				"## Adapters",
				"## CLI",
				"## Testing",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks: Add reports",
				"## Context",
				"## Baseline",
				"## Domain",
				"## Ports",
				"## Use Case",
				"## Adapters",
				"## CLI",
				"## Tests",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria: Add reports",
				"The feature behavior described by the summary is implemented within the approved scope.",
				"Existing behavior outside the feature scope remains unchanged.",
			},
			"risks.md": {
				"# Risks: Add reports",
				"## Scope Creep",
				"## Architecture Boundaries",
				"## Backwards Compatibility",
			},
		},
		domain.BugfixGuidedType: {
			"proposal.md": {
				"# Proposal: Add reports",
				"## Current Behavior",
				"## Expected Behavior",
				"## Impact",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design: Add reports",
				"## Root Cause",
				"## Fix Approach",
				"## Boundaries",
				"## Regression Testing",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks: Add reports",
				"## Reproduce",
				"## Test",
				"## Fix",
				"## Regression",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria: Add reports",
				"The incorrect behavior described by the summary is corrected.",
				"Regression coverage proves the bug does not return.",
			},
			"risks.md": {
				"# Risks: Add reports",
				"## Regression Risk",
				"## Incomplete Reproduction",
				"## Over-Fixing",
			},
		},
		domain.DocsGuidedType: {
			"proposal.md": {
				"# Proposal: Add reports",
				"## Documentation Goal",
				"## Audience",
				"## Files to Update",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design: Add reports",
				"## Documentation Structure",
				"## Source of Truth",
				"## Accuracy Rules",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks: Add reports",
				"## Inventory",
				"## Markdown Updates",
				"## Command Verification",
				"## Markdown Verification",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria: Add reports",
				"The change is limited to approved Markdown documentation files.",
				"Command examples match current CLI behavior.",
			},
			"risks.md": {
				"# Risks: Add reports",
				"## Stale Documentation",
				"## Overstating Planned Behavior",
				"## Mixed Scope",
			},
		},
		domain.RefactorGuidedType: {
			"proposal.md": {
				"# Proposal: Add reports",
				"## Refactor Goal",
				"## Current Pain",
				"## Non-Functional Goal",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design: Add reports",
				"## Boundaries",
				"## Migration Plan",
				"## Compatibility",
				"## Testing",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks: Add reports",
				"## Baseline Tests",
				"## Small Refactor Steps",
				"## Regression Tests",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria: Add reports",
				"External behavior remains unchanged.",
				"The refactor stays within the approved internal scope.",
			},
			"risks.md": {
				"# Risks: Add reports",
				"## Accidental Behavior Changes",
				"## Broad Diffs",
				"## Architecture Boundary Drift",
			},
		},
	}

	for _, guidedType := range domain.SupportedGuidedTypes() {
		t.Run(string(guidedType), func(t *testing.T) {
			for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
				starterContent, err := content.ContentFor(
					guidedType,
					" Add reports ",
					" Create report generation support ",
					requiredFile,
				)
				if err != nil {
					t.Fatalf("ContentFor(%q, %q) error = %v", guidedType, requiredFile, err)
				}

				reloadedContent, err := content.ContentFor(
					guidedType,
					" Add reports ",
					" Create report generation support ",
					requiredFile,
				)
				if err != nil {
					t.Fatalf("ContentFor(%q, %q) reload error = %v", guidedType, requiredFile, err)
				}
				if reloadedContent != starterContent {
					t.Fatalf("ContentFor(%q, %q) returned nondeterministic content", guidedType, requiredFile)
				}
				if strings.TrimSpace(starterContent) == "" {
					t.Fatalf("ContentFor(%q, %q) returned empty content", guidedType, requiredFile)
				}
				if !strings.HasPrefix(starterContent, "# ") {
					t.Fatalf("ContentFor(%q, %q) = %q, want Markdown title", guidedType, requiredFile, starterContent)
				}
				if !strings.Contains(starterContent, "Add reports") {
					t.Fatalf("ContentFor(%q, %q) = %q, want title", guidedType, requiredFile, starterContent)
				}
				if !strings.Contains(starterContent, "Create report generation support") {
					t.Fatalf("ContentFor(%q, %q) = %q, want summary", guidedType, requiredFile, starterContent)
				}
				if strings.Contains(starterContent, " Add reports ") ||
					strings.Contains(starterContent, " Create report generation support ") {
					t.Fatalf("ContentFor(%q, %q) = %q, want trimmed title and summary", guidedType, requiredFile, starterContent)
				}
				for _, forbidden := range forbiddenBuiltInTemplateContentPatterns() {
					if strings.Contains(strings.ToLower(starterContent), forbidden) {
						t.Fatalf("ContentFor(%q, %q) contains forbidden content pattern %q", guidedType, requiredFile, forbidden)
					}
				}

				expectedSections, exists := expectedMarkdown[guidedType][requiredFile]
				if !exists {
					t.Fatalf("no content expectations for %s %q", guidedType, requiredFile)
				}
				for _, want := range expectedSections {
					if !strings.Contains(starterContent, want) {
						t.Fatalf("ContentFor(%q, %q) = %q, want to contain %q", guidedType, requiredFile, starterContent, want)
					}
				}
			}
		})
	}
}

func TestGuidedChangeTemplatesTasksUseUncheckedTasksOnly(t *testing.T) {
	content := NewGuidedChangeTemplates()

	for _, guidedType := range domain.SupportedGuidedTypes() {
		t.Run(string(guidedType), func(t *testing.T) {
			starterContent, err := content.ContentFor(
				guidedType,
				"Add reports",
				"Create report generation support",
				"tasks.md",
			)
			if err != nil {
				t.Fatalf("ContentFor(%q, tasks.md) error = %v", guidedType, err)
			}
			assertBuiltInTemplateUncheckedTasksOnly(t, starterContent)
		})
	}
}

func TestGuidedChangeTemplatesRejectUnknownGuidedType(t *testing.T) {
	content := NewGuidedChangeTemplates()

	_, err := content.ContentFor(domain.GuidedType("maintenance"), "Title", "Summary", "proposal.md")
	if err == nil {
		t.Fatalf("ContentFor() error = nil, want unknown guided type error")
	}
	if err.Error() != "unknown guided type: maintenance" {
		t.Fatalf("ContentFor() error = %q, want unknown guided type context", err.Error())
	}
}

func TestGuidedChangeTemplatesRejectUnknownPath(t *testing.T) {
	content := NewGuidedChangeTemplates()

	_, err := content.ContentFor(domain.FeatureGuidedType, "Title", "Summary", "unknown.md")
	if err == nil {
		t.Fatalf("ContentFor() error = nil, want unknown path error")
	}
	if !strings.Contains(err.Error(), "no guided content for unknown.md in feature guided change") {
		t.Fatalf("ContentFor() error = %q, want guided type and path context", err.Error())
	}
}
