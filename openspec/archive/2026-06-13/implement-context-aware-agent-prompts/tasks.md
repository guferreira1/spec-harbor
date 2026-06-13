# Tasks: Implement Context-Aware Agent Prompts

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, the relevant role-specific rule file, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm implementation scope is limited to context-aware `specharbor prompt` behavior, tests, docs, and this change's task updates.
- [x] Confirm no new prompt roles, RAG, indexing, remote discovery, external API calls, command execution, agent execution, source-control automation, release, npm, Homebrew, `install.sh`, or publishing behavior is introduced.
- [x] Inspect current prompt rendering, role templates, context discovery, project brief, workflow, and CLI wiring before editing.

## Domain And Context Modeling

- [x] Reuse existing context discovery domain models where appropriate.
- [x] Add prompt-specific context domain types only if they remove real rendering complexity.
- [x] Model Project Context sections for user-confirmed context, detected facts, suggested assumptions, and conflict notes.
- [x] Validate that detected facts include source evidence and confidence before rendering.
- [x] Validate that suggested assumptions include source evidence and confidence before rendering.
- [x] Ensure unknown `.specharbor/project-brief.md` sections are not rendered as user-confirmed context.
- [x] Ensure suggested assumptions cannot be rendered as detected facts or user-confirmed context.
- [x] Add deterministic ordering for prompt context items.
- [x] Add deterministic conflict detection between confirmed context and detected facts of the same kind.
- [x] Add bounded prompt context rendering policy and truncation notices.

## Ports And Use Cases

- [x] Add or extend a small core-owned context provider boundary for prompt generation.
- [x] Reuse context discovery output through a stable domain/use-case boundary instead of duplicating discovery rules in CLI or templates.
- [x] Ensure prompt generation does not create a dependency cycle with context discovery.
- [x] Keep filesystem reads for project brief and discovery sources behind adapter/port boundaries.
- [x] Extend the render prompt use case to obtain classified context for the project root.
- [x] Extend the render prompt use case to derive role-aware Project Context content.
- [x] Preserve current validation for project root, change id, and role.
- [x] Preserve current unsupported role behavior.
- [x] Preserve current deterministic default task behavior unless explicitly changed by this active change.
- [x] Ensure the use case never prints output, calls `os` directly, executes commands, or calls external APIs.
- [x] Ensure missing context still renders a safe prompt with ask-or-label-assumptions instructions.

## Prompt Templates And Rendering

- [x] Add a deterministic Project Context insertion point to role prompt templates.
- [x] Render Project Context after read-first sources and before task instructions, unless a role-specific template requires a clearer equivalent location.
- [x] Include `.specharbor/project-brief.md` in `Read first` when present or when confirmed context is rendered.
- [x] Continue including `AGENTS.md`, `.specharbor/rules/global.md`, the role-specific rule file, `README.md`, `docs/`, `openspec/project.md`, `openspec/specs/`, and the active OpenSpec change directory as read-first guidance.
- [x] Render user-confirmed context separately from detected facts.
- [x] Render detected facts with source evidence and confidence.
- [x] Render suggested assumptions only under `Suggested assumptions`.
- [x] Render conflict notes only as safe guidance that prefers confirmed context.
- [x] Render missing-context instructions that tell agents not to invent stack, architecture, commands, persistence decisions, workflow decisions, or project direction.
- [x] Avoid dumping raw file contents from README, docs, OpenSpec, manifests, workflows, rules, or project brief files.
- [x] Avoid absolute local paths, secrets, command output, timestamps, and nondeterministic formatting.
- [x] Preserve existing final decision labels and role-specific output expectations.

## Role Behavior

- [x] Add Project Context rendering for `spec-author` when context is available.
- [x] Add Project Context rendering for `architecture-reviewer` when context is available.
- [x] Add Project Context rendering for `implementer` when context is available.
- [x] Add Project Context rendering for `test-engineer` when context is available.
- [x] Add Project Context rendering for `change-reviewer` when context is available.
- [x] Preserve the existing supported prompt role list unless another accepted change has already added roles.
- [x] Preserve the existing recommended workflow order.
- [x] Do not add PR Agent or Archive Housekeeping Agent prompt roles in this change.
- [x] If PR or Archive prompt roles already exist at implementation time, apply only the minimal context policy described in `design.md`.

## CLI

- [x] Keep `specharbor prompt <change-id> --role <role>` as the documented command shape.
- [x] Preserve existing argument parsing errors for missing change id, missing role, missing role value, unsupported role, unsupported flags, and extra arguments.
- [x] Keep CLI code limited to parsing, dependency construction, invoking the use case, and printing the rendered prompt.
- [x] Ensure successful stdout contains only the rendered prompt.
- [x] Do not add prompt output files, prompt execution, agent dispatch, provider calls, or command execution.

## Documentation

- [x] Update `README.md` after implementation to describe `specharbor prompt` as context-aware.
- [x] Update `docs/usage.md` with Project Context section behavior, context precedence, classifications, assumptions, conflict handling, missing-context instructions, and safety boundaries.
- [x] Update `docs/workflow.md` to mention that role prompts can use confirmed and discovered local context without executing commands or agents.
- [x] Update `docs/agent-roles.md` if needed to explain context-aware role prompts without changing role responsibilities.
- [x] Keep docs clear that this feature is not RAG, indexing, remote discovery, provider integration, prompt execution, agent execution, command execution, or source-control automation.
- [x] Do not update release, npm, Homebrew, `install.sh`, package-manager, or publishing docs for this feature.

## Tests

- [x] Add unit tests for prompt context model construction.
- [x] Add unit tests for deterministic Project Context rendering.
- [x] Add tests for `.specharbor/project-brief.md` confirmed context inclusion.
- [x] Add tests proving confirmed context is preferred over detected facts.
- [x] Add tests for detected facts inclusion with source evidence and confidence.
- [x] Add tests proving unknown project brief sections are not rendered as confirmed context.
- [x] Add tests for suggested assumption labeling.
- [x] Add tests ensuring assumptions are not rendered as facts.
- [x] Add tests for conflict handling between confirmed context and detected facts.
- [x] Add tests for missing context behavior.
- [x] Add tests for prompt size limits and deterministic truncation notices.
- [x] Add tests for `spec-author` context-aware prompt output.
- [x] Add tests for `architecture-reviewer` context-aware prompt output.
- [x] Add tests for `implementer` context-aware prompt output.
- [x] Add tests for `test-engineer` context-aware prompt output.
- [x] Add tests for `change-reviewer` context-aware prompt output.
- [x] Add regression tests for existing `specharbor prompt` output structure.
- [x] Add regression tests for existing role final decision labels, if present.
- [x] Add regression tests for existing recommended workflow order.
- [x] Add regression tests for existing `specharbor brief` behavior.
- [x] Add regression tests for existing `specharbor context discover` behavior.
- [x] Add regression tests for existing `specharbor scan` behavior.
- [x] Add architecture boundary tests or extend existing tests if needed.

## Tester Follow-Up

- [x] Restore final decision labels in generated prompts for all five supported context-aware roles.
- [x] Render confirmed Agent behavior from `.specharbor/project-brief.md` as user-confirmed Project Context.
- [x] Add regression tests for real role prompt labels and confirmed Agent behavior precedence.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-context-aware-agent-prompts`.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff -- openspec/changes/implement-context-aware-agent-prompts/`.
