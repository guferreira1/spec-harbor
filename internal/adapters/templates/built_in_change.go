package templates

import (
	"fmt"

	"github.com/guferreira1/spec-harbor/internal/core/domain"
)

type BuiltInChangeTemplates struct{}

func NewBuiltInChangeTemplates() *BuiltInChangeTemplates {
	return &BuiltInChangeTemplates{}
}

func (templates *BuiltInChangeTemplates) ContentFor(templateName domain.TemplateName, relativePath string) (string, error) {
	switch templateName {
	case domain.FeatureTemplate:
		return featureContentFor(relativePath)
	case domain.BugfixTemplate:
		return bugfixContentFor(relativePath)
	case domain.DocsTemplate:
		return docsContentFor(relativePath)
	case domain.RefactorTemplate:
		return refactorContentFor(relativePath)
	default:
		return "", fmt.Errorf("unknown template name: %s", templateName)
	}
}

func featureContentFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return featureProposalContent, nil
	case "design.md":
		return featureDesignContent, nil
	case "tasks.md":
		return featureTasksContent, nil
	case "acceptance-criteria.md":
		return featureAcceptanceCriteriaContent, nil
	case "risks.md":
		return featureRisksContent, nil
	default:
		return "", unknownTemplateFileError(domain.FeatureTemplate, relativePath)
	}
}

func bugfixContentFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return bugfixProposalContent, nil
	case "design.md":
		return bugfixDesignContent, nil
	case "tasks.md":
		return bugfixTasksContent, nil
	case "acceptance-criteria.md":
		return bugfixAcceptanceCriteriaContent, nil
	case "risks.md":
		return bugfixRisksContent, nil
	default:
		return "", unknownTemplateFileError(domain.BugfixTemplate, relativePath)
	}
}

func docsContentFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return docsProposalContent, nil
	case "design.md":
		return docsDesignContent, nil
	case "tasks.md":
		return docsTasksContent, nil
	case "acceptance-criteria.md":
		return docsAcceptanceCriteriaContent, nil
	case "risks.md":
		return docsRisksContent, nil
	default:
		return "", unknownTemplateFileError(domain.DocsTemplate, relativePath)
	}
}

func refactorContentFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return refactorProposalContent, nil
	case "design.md":
		return refactorDesignContent, nil
	case "tasks.md":
		return refactorTasksContent, nil
	case "acceptance-criteria.md":
		return refactorAcceptanceCriteriaContent, nil
	case "risks.md":
		return refactorRisksContent, nil
	default:
		return "", unknownTemplateFileError(domain.RefactorTemplate, relativePath)
	}
}

func unknownTemplateFileError(templateName domain.TemplateName, relativePath string) error {
	return fmt.Errorf("no template content for %s in %s template", relativePath, templateName)
}

const featureProposalContent = `# Proposal

## Summary

Describe the capability this feature should add.

## Problem

Describe the user or system problem this feature should solve.

## Proposed Solution

Describe the intended behavior and the main implementation approach.

## Scope

- List the behavior, files, interfaces, or workflows included in this feature.

## Out of Scope

- List related behavior that should not be implemented in this change.

## Success Criteria

- Describe the observable outcomes that prove the feature is complete.
`

const featureDesignContent = `# Design

## Architecture Notes

Describe the affected layers, boundaries, and dependencies.

## Domain

Describe any domain concepts or value objects this feature needs.

## Ports

Describe any interfaces the use case needs from adapters.

## Use Case

Describe the orchestration flow and error handling.

## Adapters

Describe the concrete implementations needed outside the core.

## CLI

Describe command parsing, output, and user-facing errors.

## Testing

- List unit and integration tests for success, error, and idempotency paths.

## Validation

- List formatting and verification commands to run before completion.
`

const featureTasksContent = `# Tasks

## Baseline

- [ ] Read the project context, architecture spec, and active OpenSpec change.
- [ ] Run the baseline test suite before implementation.

## Domain

- [ ] Add or update domain concepts needed for the feature.

## Ports

- [ ] Add or update small core interfaces needed by the use case.

## Use Case

- [ ] Implement the feature orchestration in the core use case.

## Adapters

- [ ] Add or update concrete adapter implementations.

## CLI

- [ ] Wire command parsing, execution, and reporting.

## Tests

- [ ] Add focused tests for success, error, and regression paths.

## Verification

- [ ] Run formatting and verification commands.
- [ ] Update this task list after completing each task.
`

const featureAcceptanceCriteriaContent = `# Acceptance Criteria

- The requested feature behavior is implemented within the approved scope.
- Existing behavior outside the feature scope remains unchanged.
- User-facing errors are clear and deterministic.
- Required files, interfaces, and workflows behave as described.
- Automated tests cover the important success and failure paths.
- Required verification commands pass.
`

const featureRisksContent = `# Risks

## Scope Creep

- The feature may expand beyond the approved behavior.

## Architecture Boundaries

- Logic may leak into the wrong layer or bypass existing ports.

## Backwards Compatibility

- Existing commands, files, or integrations may change unexpectedly.

## Mitigations

- Keep changes focused, reuse existing policies, and add regression tests.
`

const bugfixProposalContent = `# Proposal

## Current Behavior

Describe the incorrect behavior and how it appears to users or maintainers.

## Expected Behavior

Describe the corrected behavior.

## Impact

Describe who is affected and the severity of the issue.

## Scope

- List the behavior, files, or tests included in this bug fix.

## Out of Scope

- List adjacent behavior that should not be changed.

## Success Criteria

- Describe how reviewers can confirm the issue is fixed.
`

const bugfixDesignContent = `# Design

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

const bugfixTasksContent = `# Tasks

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

const bugfixAcceptanceCriteriaContent = `# Acceptance Criteria

- The incorrect behavior is corrected.
- Existing supported behavior continues to work.
- Regression coverage proves the bug does not return.
- The fix stays within the approved scope.
- Required verification commands pass.
`

const bugfixRisksContent = `# Risks

## Regression Risk

- The fix may change nearby behavior unexpectedly.

## Incomplete Reproduction

- The implementation may address a symptom instead of the cause.

## Over-Fixing

- The change may expand into unrelated cleanup or redesign.

## Mitigations

- Reproduce the issue, keep the fix small, and add regression tests.
`

const docsProposalContent = `# Proposal

## Documentation Goal

Describe what the documentation should clarify or add.

## Audience

Describe who will read this documentation.

## Files to Update

- List the README or documentation files included in this change.

## Scope

- List the documentation-only updates included in this change.

## Out of Scope

- List behavior, code, or configuration changes that should not be included.

## Success Criteria

- Describe how reviewers can confirm the documentation is accurate.
`

const docsDesignContent = `# Design

## Documentation Structure

Describe the sections, order, and cross-references to update.

## Source of Truth

Describe the code, commands, or specs used to verify the documentation.

## Accuracy Rules

- Avoid claiming behavior that is not implemented.
- Keep command examples aligned with current CLI behavior.
- Keep this change limited to Markdown documentation unless explicitly approved.

## Validation

- List documentation checks and command verification to run before completion.
`

const docsTasksContent = `# Tasks

## Inventory

- [ ] Identify the documentation files and source material to update.

## README Or Docs Updates

- [ ] Update the approved Markdown files.

## Command Verification

- [ ] Verify command examples against current behavior.

## Markdown Verification

- [ ] Review formatting, links, headings, and terminology.

## Verification

- [ ] Run required verification commands.
- [ ] Update this task list after completing each task.
`

const docsAcceptanceCriteriaContent = `# Acceptance Criteria

- The change is limited to approved Markdown documentation files.
- Command examples match current CLI behavior.
- The documentation does not overstate planned or unsupported behavior.
- Existing product behavior remains unchanged.
- Required verification commands pass.
`

const docsRisksContent = `# Risks

## Stale Documentation

- Documentation may diverge from current behavior.

## Overstating Planned Behavior

- Documentation may describe future behavior as already available.

## Mixed Scope

- Documentation work may accidentally include behavior changes.

## Mitigations

- Verify examples, cite the current source of truth, and keep edits Markdown-only.
`

const refactorProposalContent = `# Proposal

## Refactor Goal

Describe the internal structure improvement this refactor should deliver.

## Current Pain

Describe the duplication, coupling, complexity, or maintenance cost being addressed.

## Non-Functional Goal

Describe the quality improvement without changing external behavior.

## Scope

- List the internal files, boundaries, or patterns included in this refactor.

## Out of Scope

- List behavior changes, redesigns, or cleanup that should not be included.

## Success Criteria

- Describe how reviewers can confirm behavior stayed the same.
`

const refactorDesignContent = `# Design

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

const refactorTasksContent = `# Tasks

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

const refactorAcceptanceCriteriaContent = `# Acceptance Criteria

- External behavior remains unchanged.
- Public command output and file formats remain compatible.
- The refactor stays within the approved internal scope.
- Tests cover the behavior protected by the refactor.
- Required verification commands pass.
`

const refactorRisksContent = `# Risks

## Accidental Behavior Changes

- Internal changes may alter observable behavior.

## Broad Diffs

- The refactor may become difficult to review if it changes too much at once.

## Architecture Boundary Drift

- Code may move across layers without preserving dependencies.

## Mitigations

- Work in small steps, keep behavior tests close, and review dependency direction.
`
