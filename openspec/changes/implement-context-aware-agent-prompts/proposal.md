# Proposal: Implement Context-Aware Agent Prompts

## Problem

SpecHarbor can collect confirmed project context in `.specharbor/project-brief.md` and can discover local context with `specharbor context discover`, but existing `specharbor prompt` output does not consume either source. Role prompts point agents at repository rules and the active OpenSpec change, but they do not summarize confirmed stack, architecture, commands, project purpose, documentation sources, or classified discovery signals.

That gap leaves receiving agents to infer project facts from many files. Inference is risky when repository evidence is incomplete, ambiguous, conventional, or in conflict with user-confirmed project context. Agents must not invent stack, architecture, commands, persistence decisions, workflow decisions, or project direction.

## Goal

Make existing `specharbor prompt <change-id> --role <role>` output context-aware while preserving current role behavior, workflow order, safety boundaries, and deterministic rendering.

Generated prompts should include a dedicated `## Project Context` section when useful context is available. The section should prefer confirmed project context from `.specharbor/project-brief.md`, include selected local discovery signals with source evidence and confidence, and keep detected facts, suggested assumptions, and user-confirmed context clearly separated.

## Scope

- Update existing role prompt generation for `specharbor prompt`.
- Preserve the existing command shape:

```text
specharbor prompt <change-id> --role <role>
```

- Preserve currently supported prompt roles:
  - `spec-author`
  - `architecture-reviewer`
  - `implementer`
  - `test-engineer`
  - `change-reviewer`
- Add role-aware Project Context rendering for the existing supported roles.
- Prefer `user_confirmed_context` parsed from `.specharbor/project-brief.md`.
- Treat only known `.specharbor/project-brief.md` sections parsed by the existing discovery boundary as user-confirmed context.
- Include useful `detected_fact` signals from local context discovery with source evidence and confidence.
- Include `suggested_assumption` signals only when clearly labeled as assumptions.
- Ensure suggested assumptions are never rendered as facts.
- Handle conflicts by preferring user-confirmed context and safely noting conflicting detected facts when useful.
- Add missing-context instructions telling agents to ask or explicitly label assumptions instead of inventing project facts.
- Include `.specharbor/project-brief.md` in `Read first` when present or relevant.
- Continue to include existing important read-first sources:
  - `AGENTS.md`
  - `.specharbor/rules/global.md`
  - the role-specific rule file
  - `README.md`
  - `docs/`
  - `openspec/project.md`
  - `openspec/specs/`
  - the active OpenSpec change directory
- Reuse existing context discovery output through a stable core/domain or use-case boundary.
- Keep prompt generation deterministic and testable.
- Avoid dumping raw file contents into generated prompts.
- Add sensible prompt context size limits and deterministic truncation or summarization.
- Preserve existing final decision labels and role-specific output expectations.
- Preserve existing agent workflow order.
- Preserve existing behavior for `brief`, `context discover`, `scan`, `generate`, `validate`, `review`, `archive`, `config`, `workflow`, and `version`.
- Update public documentation during implementation to describe context-aware prompt behavior.

## Role Context Policy

The existing five prompt roles should receive Project Context when context is available:

- Spec Author Agent: full context, because it shapes scope, assumptions, and requested OpenSpec content.
- Architecture Reviewer Agent: full context, because confirmed architecture and detected architecture hints are review inputs.
- Implementer Agent: full context, because stack, commands, architecture, and agent behavior preferences affect implementation.
- Test Engineer Agent: full context, because test/build commands and stack evidence affect verification work.
- Change Reviewer Agent: full context, because reviewers need to compare implementation claims against confirmed and detected context.

Pull Request and Archive workflow steps are currently manual workflow steps, not supported `specharbor prompt --role` values. This change must not introduce new PR Agent or Archive Housekeeping Agent prompt roles unless such roles already exist at implementation time through another accepted change. If a future or concurrent prompt surface exists for those manual steps, it should receive only minimal context such as confirmed project identity, purpose, and safety instructions; it should not receive full detected command or architecture assumptions unless that is explicitly specified by that role's accepted change.

## Out Of Scope

- Implementing code in this spec-authoring task.
- Modifying production code in this spec-authoring task.
- Modifying documentation outside this OpenSpec change in this spec-authoring task.
- Adding new supported prompt roles.
- Changing the `specharbor prompt` CLI shape.
- Changing project brief merge, update, overwrite, append, or questionnaire behavior.
- Changing context discovery detection rules except where strictly needed to consume existing output.
- Changing `specharbor context discover` report formatting.
- Changing `specharbor scan` behavior.
- RAG.
- Embeddings.
- Vector databases.
- Repository-wide indexing.
- Local retrieval or snippet ranking.
- GitHub remote discovery.
- GitLab remote discovery.
- Bitbucket remote discovery.
- Remote source-control API calls.
- Automatic prompt execution.
- Automatic agent execution.
- Provider APIs, local model APIs, or external API requirements.
- Running package managers, tests, builds, run commands, scripts, shells, agent CLIs, or workflow tools from prompt generation.
- Modifying source files based on context.
- Release automation.
- npm changes.
- Homebrew changes.
- `install.sh` changes.
- Publishing flows.
- Archiving, tagging, committing, pushing, opening pull requests, merging, or source-control automation.

## Success Criteria

- `specharbor prompt` is specified as context-aware for existing supported roles.
- Generated prompts include a dedicated Project Context section when context is available.
- Confirmed context from `.specharbor/project-brief.md` is preferred over detected facts.
- Unknown project brief sections are not treated as confirmed context.
- Detected facts can be included with source evidence and confidence.
- Suggested assumptions can be included only when clearly labeled as assumptions.
- Suggested assumptions are never rendered as facts or confirmed context.
- Conflicts between confirmed context and detected facts prefer confirmed context.
- Missing or ambiguous context produces safe instructions to ask or label assumptions.
- Prompt generation does not silently promote detected facts or assumptions into confirmed context.
- Prompt generation does not execute project commands.
- Prompt generation does not require external APIs or provider API keys.
- Existing role-specific prompt behavior, workflow order, and final decision labels are preserved.
- Prompt rendering avoids raw file-content dumping and uses bounded deterministic context summaries.
- Public docs are planned for the implementation change.
- OpenSpec validation passes with zero errors.
