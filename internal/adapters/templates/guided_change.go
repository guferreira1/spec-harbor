package templates

import (
	"fmt"
	"strings"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type GuidedChangeTemplates struct{}

func NewGuidedChangeTemplates() *GuidedChangeTemplates {
	return &GuidedChangeTemplates{}
}

func (templates *GuidedChangeTemplates) ContentFor(
	guidedType domain.GuidedType,
	title string,
	summary string,
	relativePath string,
) (string, error) {
	normalizedType, err := domain.ParseGuidedType(string(guidedType))
	if err != nil {
		return "", err
	}

	source, err := guidedTemplateSourceFor(normalizedType, relativePath)
	if err != nil {
		return "", err
	}

	return renderGuidedTemplate(source, strings.TrimSpace(title), strings.TrimSpace(summary)), nil
}

func guidedTemplateSourceFor(guidedType domain.GuidedType, relativePath string) (string, error) {
	switch guidedType {
	case domain.FeatureGuidedType:
		return featureGuidedSourceFor(relativePath)
	case domain.BugfixGuidedType:
		return bugfixGuidedSourceFor(relativePath)
	case domain.DocsGuidedType:
		return docsGuidedSourceFor(relativePath)
	case domain.RefactorGuidedType:
		return refactorGuidedSourceFor(relativePath)
	default:
		return "", fmt.Errorf("unknown guided type: %s", guidedType)
	}
}

func featureGuidedSourceFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return featureGuidedProposalContent, nil
	case "design.md":
		return featureGuidedDesignContent, nil
	case "tasks.md":
		return featureGuidedTasksContent, nil
	case "acceptance-criteria.md":
		return featureGuidedAcceptanceCriteriaContent, nil
	case "risks.md":
		return featureGuidedRisksContent, nil
	default:
		return "", unknownGuidedFileError(domain.FeatureGuidedType, relativePath)
	}
}

func bugfixGuidedSourceFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return bugfixGuidedProposalContent, nil
	case "design.md":
		return bugfixGuidedDesignContent, nil
	case "tasks.md":
		return bugfixGuidedTasksContent, nil
	case "acceptance-criteria.md":
		return bugfixGuidedAcceptanceCriteriaContent, nil
	case "risks.md":
		return bugfixGuidedRisksContent, nil
	default:
		return "", unknownGuidedFileError(domain.BugfixGuidedType, relativePath)
	}
}

func docsGuidedSourceFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return docsGuidedProposalContent, nil
	case "design.md":
		return docsGuidedDesignContent, nil
	case "tasks.md":
		return docsGuidedTasksContent, nil
	case "acceptance-criteria.md":
		return docsGuidedAcceptanceCriteriaContent, nil
	case "risks.md":
		return docsGuidedRisksContent, nil
	default:
		return "", unknownGuidedFileError(domain.DocsGuidedType, relativePath)
	}
}

func refactorGuidedSourceFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return refactorGuidedProposalContent, nil
	case "design.md":
		return refactorGuidedDesignContent, nil
	case "tasks.md":
		return refactorGuidedTasksContent, nil
	case "acceptance-criteria.md":
		return refactorGuidedAcceptanceCriteriaContent, nil
	case "risks.md":
		return refactorGuidedRisksContent, nil
	default:
		return "", unknownGuidedFileError(domain.RefactorGuidedType, relativePath)
	}
}

func unknownGuidedFileError(guidedType domain.GuidedType, relativePath string) error {
	return fmt.Errorf("no guided content for %s in %s guided change", relativePath, guidedType)
}

func renderGuidedTemplate(source string, title string, summary string) string {
	replacer := strings.NewReplacer(
		"{{title}}", title,
		"{{summary}}", summary,
	)
	return replacer.Replace(source)
}

const featureGuidedProposalContent = `# Proposal: {{title}}

## Summary

{{summary}}

## Problem

Describe the user or system problem this feature should solve.

## Proposed Solution

Describe the intended behavior for "{{title}}" and the main implementation approach.

## Scope

- List the behavior, files, interfaces, or paths included in this feature.

## Out of Scope

- List related behavior that should not be implemented in this change.

## Success Criteria

- Describe the observable outcomes that prove "{{title}}" is complete.
`

const featureGuidedDesignContent = `# Design: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Architecture Notes

Describe the affected layers, boundaries, and dependencies for this feature.

## Domain

Describe any domain concepts or value objects this feature needs.

## Ports

Describe any interfaces the use case needs from adapters.

## Use Case

Describe the orchestration flow and error handling.

## Adapters

Describe the concrete implementations needed outside the core.

## CLI

Describe command parsing, output, and user-facing errors when applicable.

## Testing

- List unit and integration tests for success, error, and idempotency paths.

## Validation

- List formatting and verification commands to run before completion.
`

const featureGuidedTasksContent = `# Tasks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Baseline

- [ ] Read the project context, architecture spec, and active OpenSpec change.
- [ ] Run the baseline test suite before implementation.

## Domain

- [ ] Add or update domain concepts needed for this feature.

## Ports

- [ ] Add or update small core interfaces needed by the use case.

## Use Case

- [ ] Implement the feature orchestration in the core use case.

## Adapters

- [ ] Add or update concrete adapter implementations.

## CLI

- [ ] Wire command parsing, execution, and reporting when applicable.

## Tests

- [ ] Add focused tests for success, error, and regression paths.

## Verification

- [ ] Run formatting and verification commands.
- [ ] Update this task list after completing each task.
`

const featureGuidedAcceptanceCriteriaContent = `# Acceptance Criteria: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Criteria

- The feature behavior described by the summary is implemented within the approved scope.
- Existing behavior outside the feature scope remains unchanged.
- User-facing errors are clear and deterministic where applicable.
- Required files, interfaces, and paths behave as described.
- Automated tests cover the important success and failure paths.
- Required verification commands pass.
`

const featureGuidedRisksContent = `# Risks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Scope Creep

- The feature may expand beyond the approved behavior.

## Architecture Boundaries

- Logic may leak into the wrong layer or bypass existing ports.

## Backwards Compatibility

- Existing commands, files, or integrations may change unexpectedly.

## Mitigations

- Keep changes focused, reuse existing policies, and add regression tests.
`

const bugfixGuidedProposalContent = `# Proposal: {{title}}

## Summary

{{summary}}

## Current Behavior

Describe the incorrect behavior and how it appears to users or maintainers.

## Expected Behavior

Describe the corrected behavior for "{{title}}".

## Impact

Describe who is affected and the severity of the issue.

## Scope

- List the behavior, files, or tests included in this bug fix.

## Out of Scope

- List adjacent behavior that should not be changed.

## Success Criteria

- Describe how reviewers can confirm the issue is fixed.
`

const bugfixGuidedDesignContent = `# Design: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Root Cause

Describe the likely cause of the incorrect behavior.

## Fix Approach

Describe the smallest change that corrects the behavior.

## Boundaries

Describe what must remain unchanged.

## Regression Testing

- List tests that reproduce the issue and prove the fix.

## Validation

- List formatting and verification commands to run before completion.
`

const bugfixGuidedTasksContent = `# Tasks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Reproduce

- [ ] Reproduce or characterize the current incorrect behavior.

## Test

- [ ] Add a failing test or focused coverage for the bug.

## Fix

- [ ] Implement the smallest correction within the approved scope.

## Regression

- [ ] Add or update regression coverage for related behavior.

## Verification

- [ ] Run formatting and verification commands.
- [ ] Update this task list after completing each task.
`

const bugfixGuidedAcceptanceCriteriaContent = `# Acceptance Criteria: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Criteria

- The incorrect behavior described by the summary is corrected.
- Existing supported behavior continues to work.
- Regression coverage proves the bug does not return.
- The fix stays within the approved scope.
- Required verification commands pass.
`

const bugfixGuidedRisksContent = `# Risks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Regression Risk

- The fix may change nearby behavior unexpectedly.

## Incomplete Reproduction

- The implementation may address a symptom instead of the cause.

## Over-Fixing

- The change may expand into unrelated cleanup or redesign.

## Mitigations

- Reproduce the issue, keep the fix small, and add regression tests.
`

const docsGuidedProposalContent = `# Proposal: {{title}}

## Summary

{{summary}}

## Documentation Goal

Describe what the documentation should clarify or add.

## Audience

Describe who will read this documentation.

## Files to Update

- List the approved Markdown files included in this change.

## Scope

- List the documentation-only updates included in this change.

## Out of Scope

- List behavior, code, or configuration changes that should not be included.

## Success Criteria

- Describe how reviewers can confirm the documentation is accurate.
`

const docsGuidedDesignContent = `# Design: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Documentation Structure

Describe the sections, order, and cross-references to update.

## Source of Truth

Describe the code, commands, or specs used to verify the documentation.

## Accuracy Rules

- Avoid claiming behavior that is not implemented.
- Keep command examples aligned with current CLI behavior.
- Keep this change limited to approved Markdown documentation unless explicitly approved.

## Validation

- List documentation checks and command verification to run before completion.
`

const docsGuidedTasksContent = `# Tasks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Inventory

- [ ] Identify the documentation files and source material to update.

## Markdown Updates

- [ ] Update the approved Markdown files.

## Command Verification

- [ ] Verify command examples against current behavior.

## Markdown Verification

- [ ] Review formatting, links, headings, and terminology.

## Verification

- [ ] Run required verification commands.
- [ ] Update this task list after completing each task.
`

const docsGuidedAcceptanceCriteriaContent = `# Acceptance Criteria: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Criteria

- The change is limited to approved Markdown documentation files.
- Command examples match current CLI behavior.
- The documentation does not overstate planned or unsupported behavior.
- Existing product behavior remains unchanged.
- Required verification commands pass.
`

const docsGuidedRisksContent = `# Risks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Stale Documentation

- Documentation may diverge from current behavior.

## Overstating Planned Behavior

- Documentation may describe future behavior as already available.

## Mixed Scope

- Documentation work may accidentally include behavior changes.

## Mitigations

- Verify examples, cite the current source of truth, and keep edits Markdown-only.
`

const refactorGuidedProposalContent = `# Proposal: {{title}}

## Summary

{{summary}}

## Refactor Goal

Describe the internal structure improvement this refactor should deliver.

## Current Pain

Describe the duplication, coupling, complexity, or maintenance cost being addressed.

## Non-Functional Goal

Describe the quality improvement for "{{title}}" without changing external behavior.

## Scope

- List the internal files, boundaries, or patterns included in this refactor.

## Out of Scope

- List behavior changes, redesigns, or cleanup that should not be included.

## Success Criteria

- Describe how reviewers can confirm behavior stayed the same.
`

const refactorGuidedDesignContent = `# Design: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Boundaries

Describe the package, layer, or ownership boundaries that must be preserved.

## Migration Plan

Describe the small steps used to move from the current structure to the target structure.

## Compatibility

Describe how external behavior, command output, and file formats remain unchanged.

## Testing

- List baseline and regression tests that protect existing behavior.

## Validation

- List formatting and verification commands to run before completion.
`

const refactorGuidedTasksContent = `# Tasks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Baseline Tests

- [ ] Run baseline tests before changing structure.

## Small Refactor Steps

- [ ] Make the first small internal refactor.
- [ ] Make the next small internal refactor.
- [ ] Remove obsolete internal code only after replacement is covered.

## Regression Tests

- [ ] Add or update tests that prove external behavior is unchanged.

## Verification

- [ ] Run formatting and verification commands.
- [ ] Review the diff for accidental behavior changes.
- [ ] Update this task list after completing each task.
`

const refactorGuidedAcceptanceCriteriaContent = `# Acceptance Criteria: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Criteria

- External behavior remains unchanged.
- Public command output and file formats remain compatible.
- The refactor stays within the approved internal scope.
- Tests cover the behavior protected by the refactor.
- Required verification commands pass.
`

const refactorGuidedRisksContent = `# Risks: {{title}}

## Context

Title: {{title}}

Summary: {{summary}}

## Accidental Behavior Changes

- Internal changes may alter observable behavior.

## Broad Diffs

- The refactor may become difficult to review if it changes too much at once.

## Architecture Boundary Drift

- Code may move across layers without preserving dependencies.

## Mitigations

- Work in small steps, keep behavior tests close, and review dependency direction.
`
