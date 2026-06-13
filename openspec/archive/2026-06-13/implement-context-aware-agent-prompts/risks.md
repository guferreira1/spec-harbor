# Risks: Implement Context-Aware Agent Prompts

## Assumptions Becoming Facts

Suggested assumptions such as conventional test commands can look authoritative when placed inside an agent prompt.

Mitigation: render assumptions only under `Suggested assumptions`, use assumption wording such as `may be`, include source and confidence, and add tests proving assumptions never appear under user-confirmed context or detected facts.

## Detected Facts Overriding Confirmed Context

Repository files can conflict with `.specharbor/project-brief.md`, especially in monorepos or stale projects.

Mitigation: prefer `user_confirmed_context` over `detected_fact`, render confirmed context first, and add conflict notes only as safe context. Do not update or merge the project brief in this change.

## Stale Project Brief Data

Confirmed project context can become stale when the repository evolves after `.specharbor/project-brief.md` was written.

Mitigation: show the project brief as the source for confirmed context, keep conflicting detected facts separately labeled, include safe conflict notes when useful, and never auto-update or merge the brief during prompt generation.

## User Confusion Between Confirmed And Detected Context

Users and receiving agents may misread detected repository facts as user-confirmed project decisions.

Mitigation: keep `User-confirmed context`, `Detected facts`, and `Suggested assumptions` as separate prompt sections, include source and confidence for detected facts, and document that only known project brief sections are confirmed context.

## Prompt Bloat

Discovery can produce many signals. Dumping all of them into every prompt would make prompts noisy and could hide the active OpenSpec change.

Mitigation: use role-aware selection, item limits, value length limits, total context budgets, and deterministic truncation notices. Keep the active change and role task prominent.

## Raw File Content Exposure

Prompt rendering could accidentally include large documentation excerpts, project brief contents, workflow contents, or sensitive information.

Mitigation: consume structured context signals only, render relative source paths and short evidence labels, enforce skip policies through discovery, and test that raw private prose is not copied into prompts.

## CLI Business Logic Drift

The CLI adapter could start deciding which signals are confirmed, detected, assumed, or conflicting.

Mitigation: keep context selection, precedence, conflict handling, and rendering policy in core/domain or core/usecase. CLI code should parse arguments, wire dependencies, call the use case, and print the returned prompt.

## Context Discovery Duplication

Prompt generation could duplicate discovery rules, causing `specharbor context discover` and `specharbor prompt` to disagree.

Mitigation: reuse existing context discovery output through a stable core/domain or use-case boundary. If needed, extract shared context collection rather than copying detector logic into templates or CLI formatting.

## Dependency Cycles

Adding context-aware prompts could couple prompt generation and context discovery in both directions.

Mitigation: keep discovery independent from prompt generation. Prompt generation may consume discovery results, but discovery must not import or call prompt rendering.

## Role Scope Creep

The mention of Pull Request and Archive agents could lead implementation to add new supported prompt roles or source-control automation.

Mitigation: preserve the current supported role list unless another accepted OpenSpec change has already added roles. Keep Pull Request manual and Archive explicit. Apply only minimal context if such roles already exist at implementation time.

## Existing Prompt Regression

Adding Project Context could break existing role prompt structure, final decision labels, read-first guidance, or tests expecting stable prompts.

Mitigation: insert Project Context in a controlled location, preserve existing role-specific text and final labels, and add regression tests for current output structure and workflow order.

## Missing Context Overconfidence

When no brief exists and discovery finds little evidence, prompts may still sound confident.

Mitigation: render missing-context instructions that require the receiving agent to ask or explicitly label assumptions before large architecture, persistence, workflow, stack, or command decisions.

## Command Execution Expectations

Including detected commands in a prompt could imply that SpecHarbor ran or verified those commands.

Mitigation: label command source and classification, never run commands during prompt generation, and document that detected commands and assumptions are context signals, not verified execution results.

## Documentation Drift

Docs could describe the feature as RAG, indexing, remote discovery, provider integration, or automatic agent execution.

Mitigation: update docs only during implementation, describe prompt context as bounded local classified context, and repeat the no-execution and no-external-API boundaries.
