package templates

import "fmt"

type DefaultBlankChangeContent struct{}

func NewDefaultBlankChangeContent() *DefaultBlankChangeContent {
	return &DefaultBlankChangeContent{}
}

func (content *DefaultBlankChangeContent) ContentFor(relativePath string) (string, error) {
	switch relativePath {
	case "proposal.md":
		return proposalStarterContent, nil
	case "design.md":
		return designStarterContent, nil
	case "tasks.md":
		return tasksStarterContent, nil
	case "acceptance-criteria.md":
		return acceptanceCriteriaStarterContent, nil
	case "risks.md":
		return risksStarterContent, nil
	default:
		return "", fmt.Errorf("no blank change content for %s", relativePath)
	}
}

const proposalStarterContent = `# Proposal

## Problem

Describe the problem this change should solve and who is affected.

## Goal

Describe the outcome this change should deliver.

## Scope

- List the behavior, files, or interfaces included in this change.

## Out of Scope

- List related work that should not be implemented in this change.

## Success Criteria

- Describe how reviewers can tell the change is complete.
`

const designStarterContent = `# Design

## Overview

Describe the proposed approach at a high level.

## Architecture

Describe the affected layers, boundaries, and dependencies.

## Technical Decisions

- Record important implementation choices and tradeoffs.

## Testing Strategy

- Describe the tests needed to cover the change.

## Validation

- List the commands or checks that should pass before completion.
`

const tasksStarterContent = `# Tasks

## Implementation

- [ ] Read the project context, architecture spec, and active OpenSpec change.
- [ ] Keep implementation limited to the approved scope.
- [ ] Add or update domain, use case, port, adapter, CLI, and test code as needed.
- [ ] Run required formatting and verification commands.
- [ ] Update this task list only after implementation work is complete.
`

const acceptanceCriteriaStarterContent = `# Acceptance Criteria

- The requested behavior is implemented within the approved scope.
- Existing behavior outside the scope remains unchanged.
- Errors are clear and actionable for users.
- Automated tests cover the important success and failure paths.
- Required verification commands pass.
`

const risksStarterContent = `# Risks

## Risks

- Identify technical, product, security, or delivery risks introduced by this change.

## Mitigations

- Describe how each risk will be reduced, tested, or monitored.
`
