# Proposal: Implement Project Briefing Foundation

## Problem

SpecHarbor can scan a repository for deterministic local signals, generate and validate OpenSpec changes, render role prompts, and show the recommended workflow. Those capabilities help once a change is described, but they do not give users a structured way to provide missing project context when repository signals are incomplete, ambiguous, or absent.

Coding agents are especially risky when project purpose, users, stack, architecture, commands, or expectations for missing context are unclear. In that situation SpecHarbor should ask the user directly instead of presenting inferred stack, architecture, commands, or project decisions as facts.

## Goal

Add the first interactive project briefing workflow:

```text
specharbor brief
```

The command collects explicit project context through multiple-choice terminal questions, asks for confirmation, and writes a human-readable Markdown brief to:

```text
.specharbor/project-brief.md
```

The brief becomes a project-owned context artifact that future SpecHarbor workflows can read without requiring repository-wide indexing, embeddings, RAG, GitHub remote discovery, or provider API keys.

## Scope

- Add a new `specharbor brief` CLI command.
- Accept no positional arguments and no flags in the first version.
- Require an interactive TTY and fail clearly without prompting or writing when no TTY is available.
- Ask interactive multiple-choice questions with three to five suggested options per question.
- Ensure the final option for every question is `Other / custom`.
- Collect at least:
  - project type;
  - project purpose;
  - target users;
  - stack;
  - architecture;
  - install command;
  - test command;
  - build command;
  - run command;
  - preferred agent behavior when context is missing.
- Allow custom text when `Other / custom` is selected.
- Use detected local context, if any, only as suggestions that require user confirmation.
- Treat all unconfirmed, missing, or ambiguous context as unconfirmed.
- Never invent stack, architecture, commands, or project decisions as facts.
- Print a deterministic pre-write summary showing the target file and safety boundaries.
- Ask for explicit confirmation before writing.
- Treat empty confirmation, EOF, `n`, and `no` as cancellation.
- Write nothing when cancelled or when confirmation retry limits are exceeded.
- Create `.specharbor/` when needed after confirmation, using safe filesystem write rules.
- Write `.specharbor/project-brief.md` only when it does not already exist.
- Refuse to merge, update, overwrite, or append to an existing project brief in this first version.
- Render deterministic, human-readable Markdown with the expected structure:

```text
# Project Brief

## Project type

## Purpose

## Target users

## Stack

## Architecture

## Commands

### Install

### Test

### Build

### Run

## Agent behavior

## Context sources

## Assumptions
```

- Clearly identify user-provided answers, detected context, and assumptions in the rendered Markdown.
- Preserve existing behavior for `version`, `init`, `scan`, `generate`, `validate`, `prompt`, `review`, `archive`, `config`, and `workflow`.
- Add focused tests for the brief domain model, Markdown rendering, use case orchestration, CLI prompt behavior, cancellation behavior, safe writes, and command regressions.
- Update public usage documentation in the implementation change to describe the command, prompt flow, output file, cancellation behavior, and safety boundaries.

## Out of Scope

- Implementing code in this spec-authoring task.
- Modifying files outside `openspec/changes/implement-project-briefing-foundation/` in this spec-authoring task.
- Repository-wide context indexing.
- Embeddings.
- Vector databases.
- RAG providers.
- GitHub remote discovery.
- GitLab remote discovery.
- Source-control host API calls.
- Prompt injection into generated agent prompts.
- Reading `.specharbor/project-brief.md` from role prompt generation.
- Modifying generated agent prompts.
- Merging, updating, overwriting, or appending to an existing project brief.
- Automatic deep repository analysis.
- Recursive project scanning beyond existing scan behavior.
- Running package managers, test commands, build commands, run commands, scripts, or shell commands.
- Verifying that user-provided commands work.
- Calling AI providers, LLM APIs, local model APIs, or agent CLIs.
- Agent execution or workflow execution.
- Provider API keys or credential management.
- Config mutation.
- Changing release, npm, Homebrew, `install.sh`, GoReleaser, publishing, package-manager, or release workflow behavior.
- Archiving, publishing, releasing, tagging, committing, pushing, opening pull requests, merging, or source-control automation.

## Success Criteria

- `specharbor brief` is specified as a new command.
- The command collects explicit project context through interactive multiple-choice questions.
- Every question has three to five options and the last option is `Other / custom`.
- Custom answers are accepted for every question.
- The collected context covers project type, purpose, target users, stack, architecture, install, test, build, run, and agent missing-context behavior.
- The command writes `.specharbor/project-brief.md` only after user confirmation.
- Cancellation writes no file.
- Existing project briefs are not merged, overwritten, appended, or updated.
- The generated project brief is deterministic, human-readable Markdown.
- User-provided answers, detected context, and assumptions are clearly separated.
- Missing or ambiguous context is handled through suggestions that require confirmation.
- Assumptions are never treated as confirmed facts.
- No RAG, embeddings, repository indexing, GitHub remote discovery, provider APIs, agent execution, prompt injection, or release/publishing behavior is introduced.
- Existing CLI commands continue to work.
- `go run ./cmd/specharbor validate implement-project-briefing-foundation` passes with zero errors.
