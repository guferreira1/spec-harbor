package templates

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestBuiltInChangeTemplatesProvideEveryRequiredFile(t *testing.T) {
	content := NewBuiltInChangeTemplates()
	expectedMarkdown := map[domain.TemplateName]map[string][]string{
		domain.FeatureTemplate: {
			"proposal.md": {
				"# Proposal",
				"## Summary",
				"## Problem",
				"## Proposed Solution",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design",
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
				"# Tasks",
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
				"# Acceptance Criteria",
				"The requested feature behavior is implemented within the approved scope.",
				"Existing behavior outside the feature scope remains unchanged.",
			},
			"risks.md": {
				"# Risks",
				"## Scope Creep",
				"## Architecture Boundaries",
				"## Backwards Compatibility",
			},
		},
		domain.BugfixTemplate: {
			"proposal.md": {
				"# Proposal",
				"## Current Behavior",
				"## Expected Behavior",
				"## Impact",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design",
				"## Root Cause",
				"## Fix Approach",
				"## Boundaries",
				"## Regression Testing",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks",
				"## Reproduce",
				"## Test",
				"## Fix",
				"## Regression",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria",
				"The incorrect behavior is corrected.",
				"Regression coverage proves the bug does not return.",
			},
			"risks.md": {
				"# Risks",
				"## Regression Risk",
				"## Over-Fixing",
			},
		},
		domain.DocsTemplate: {
			"proposal.md": {
				"# Proposal",
				"## Documentation Goal",
				"## Audience",
				"## Files to Update",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design",
				"## Documentation Structure",
				"## Source of Truth",
				"## Accuracy Rules",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks",
				"## Inventory",
				"## README Or Docs Updates",
				"## Command Verification",
				"## Markdown Verification",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria",
				"The change is limited to approved Markdown documentation files.",
				"Command examples match current CLI behavior.",
			},
			"risks.md": {
				"# Risks",
				"## Stale Documentation",
				"## Overstating Planned Behavior",
			},
		},
		domain.RefactorTemplate: {
			"proposal.md": {
				"# Proposal",
				"## Refactor Goal",
				"## Current Pain",
				"## Non-Functional Goal",
				"## Scope",
				"## Out of Scope",
				"## Success Criteria",
			},
			"design.md": {
				"# Design",
				"## Boundaries",
				"## Migration Plan",
				"## Compatibility",
				"## Testing",
				"## Validation",
			},
			"tasks.md": {
				"# Tasks",
				"## Baseline Tests",
				"## Small Refactor Steps",
				"## Regression Tests",
				"## Verification",
			},
			"acceptance-criteria.md": {
				"# Acceptance Criteria",
				"External behavior remains unchanged.",
				"The refactor stays within the approved internal scope.",
			},
			"risks.md": {
				"# Risks",
				"## Accidental Behavior Changes",
				"## Broad Diffs",
			},
		},
	}

	for _, templateName := range domain.SupportedTemplateNames() {
		t.Run(string(templateName), func(t *testing.T) {
			for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
				starterContent, err := content.ContentFor(templateName, requiredFile)
				if err != nil {
					t.Fatalf("ContentFor(%q, %q) error = %v", templateName, requiredFile, err)
				}

				reloadedContent, err := content.ContentFor(templateName, requiredFile)
				if err != nil {
					t.Fatalf("ContentFor(%q, %q) reload error = %v", templateName, requiredFile, err)
				}
				if reloadedContent != starterContent {
					t.Fatalf("ContentFor(%q, %q) returned nondeterministic content", templateName, requiredFile)
				}
				if strings.TrimSpace(starterContent) == "" {
					t.Fatalf("ContentFor(%q, %q) returned empty content", templateName, requiredFile)
				}
				if !strings.HasPrefix(starterContent, "# ") {
					t.Fatalf("ContentFor(%q, %q) = %q, want Markdown title", templateName, requiredFile, starterContent)
				}
				for _, forbidden := range forbiddenBuiltInTemplateContentPatterns() {
					if strings.Contains(strings.ToLower(starterContent), forbidden) {
						t.Fatalf("ContentFor(%q, %q) contains forbidden content pattern %q", templateName, requiredFile, forbidden)
					}
				}

				expectedSections, exists := expectedMarkdown[templateName][requiredFile]
				if !exists {
					t.Fatalf("no content expectations for %s %q", templateName, requiredFile)
				}
				for _, want := range expectedSections {
					if !strings.Contains(starterContent, want) {
						t.Fatalf("ContentFor(%q, %q) = %q, want to contain %q", templateName, requiredFile, starterContent, want)
					}
				}
			}
		})
	}
}

func TestBuiltInChangeTemplatesTasksUseUncheckedTasksOnly(t *testing.T) {
	content := NewBuiltInChangeTemplates()

	for _, templateName := range domain.SupportedTemplateNames() {
		t.Run(string(templateName), func(t *testing.T) {
			starterContent, err := content.ContentFor(templateName, "tasks.md")
			if err != nil {
				t.Fatalf("ContentFor(%q, tasks.md) error = %v", templateName, err)
			}
			assertBuiltInTemplateUncheckedTasksOnly(t, starterContent)
		})
	}
}

func TestBuiltInChangeTemplatesRejectUnknownTemplateName(t *testing.T) {
	content := NewBuiltInChangeTemplates()

	_, err := content.ContentFor(domain.TemplateName("maintenance"), "proposal.md")
	if err == nil {
		t.Fatalf("ContentFor() error = nil, want unknown template error")
	}
	if err.Error() != "unknown template name: maintenance" {
		t.Fatalf("ContentFor() error = %q, want unknown template context", err.Error())
	}
}

func TestBuiltInChangeTemplatesRejectUnknownPath(t *testing.T) {
	content := NewBuiltInChangeTemplates()

	_, err := content.ContentFor(domain.FeatureTemplate, "unknown.md")
	if err == nil {
		t.Fatalf("ContentFor() error = nil, want unknown path error")
	}
	if !strings.Contains(err.Error(), "no template content for unknown.md in feature template") {
		t.Fatalf("ContentFor() error = %q, want template and path context", err.Error())
	}
}

func forbiddenBuiltInTemplateContentPatterns() []string {
	return []string{
		"http://",
		"https://",
		"api key",
		"secret",
		"credential",
		"token",
		"password",
		"provider setup",
		"openai",
		"anthropic",
		"gemini",
		"ollama",
		"localhost",
		"127.0.0.1",
		"/home/",
		"/users/",
		"/tmp/",
		"~/",
		"c:\\",
		"remote registry",
		"template registry",
		"registry.npmjs.org",
		"ghcr.io",
		"docker.io",
	}
}

func assertBuiltInTemplateUncheckedTasksOnly(t *testing.T, contents string) {
	t.Helper()

	checkboxes := 0
	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- [") {
			continue
		}

		checkboxes++
		if !strings.HasPrefix(line, "- [ ]") {
			t.Fatalf("tasks.md checkbox line = %q, want unchecked task", line)
		}
	}
	if checkboxes == 0 {
		t.Fatalf("tasks.md content = %q, want unchecked tasks", contents)
	}
}
