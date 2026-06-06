package templates

import (
	"strings"
	"testing"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

func TestDefaultBlankChangeContentProvidesEveryRequiredFile(t *testing.T) {
	content := NewDefaultBlankChangeContent()
	expectedMarkdown := map[string][]string{
		"proposal.md": {
			"# Proposal",
			"## Problem",
			"## Goal",
			"## Scope",
			"## Out of Scope",
			"## Success Criteria",
		},
		"design.md": {
			"# Design",
			"## Overview",
			"## Architecture",
			"## Technical Decisions",
			"## Testing Strategy",
			"## Validation",
		},
		"tasks.md": {
			"# Tasks",
			"## Implementation",
			"- [ ] Read the project context",
			"- [ ] Keep implementation limited to the approved scope.",
			"- [ ] Run required formatting and verification commands.",
		},
		"acceptance-criteria.md": {
			"# Acceptance Criteria",
			"- The requested behavior is implemented within the approved scope.",
			"- Existing behavior outside the scope remains unchanged.",
			"- Required verification commands pass.",
		},
		"risks.md": {
			"# Risks",
			"## Risks",
			"## Mitigations",
		},
	}

	for _, requiredFile := range domain.RequiredOpenSpecChangeFiles() {
		starterContent, err := content.ContentFor(requiredFile)
		if err != nil {
			t.Fatalf("ContentFor(%q) error = %v", requiredFile, err)
		}
		reloadedContent, err := content.ContentFor(requiredFile)
		if err != nil {
			t.Fatalf("ContentFor(%q) reload error = %v", requiredFile, err)
		}
		if reloadedContent != starterContent {
			t.Fatalf("ContentFor(%q) returned nondeterministic content", requiredFile)
		}
		if strings.TrimSpace(starterContent) == "" {
			t.Fatalf("ContentFor(%q) returned empty content", requiredFile)
		}
		if !strings.HasPrefix(starterContent, "# ") {
			t.Fatalf("ContentFor(%q) = %q, want Markdown title", requiredFile, starterContent)
		}
		expectedSections, exists := expectedMarkdown[requiredFile]
		if !exists {
			t.Fatalf("no content expectations for required file %q", requiredFile)
		}
		for _, want := range expectedSections {
			if !strings.Contains(starterContent, want) {
				t.Fatalf("ContentFor(%q) = %q, want to contain %q", requiredFile, starterContent, want)
			}
		}
	}
}

func TestDefaultBlankChangeTasksContentUsesUncheckedTasksOnly(t *testing.T) {
	content := NewDefaultBlankChangeContent()

	starterContent, err := content.ContentFor("tasks.md")
	if err != nil {
		t.Fatalf("ContentFor(tasks.md) error = %v", err)
	}
	checkboxes := 0
	for _, line := range strings.Split(starterContent, "\n") {
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
		t.Fatalf("tasks.md content = %q, want unchecked tasks", starterContent)
	}
}

func TestDefaultBlankChangeContentRejectsUnknownPath(t *testing.T) {
	content := NewDefaultBlankChangeContent()

	if _, err := content.ContentFor("unknown.md"); err == nil {
		t.Fatalf("ContentFor() error = nil, want error")
	}
}
