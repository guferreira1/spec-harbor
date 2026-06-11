# Acceptance Criteria: Implement Agent-Assisted Spec Authoring

## Supported Commands

- `specharbor generate <change-id> --agent-assisted --agent <agent-name> --type feature --title "<title>" --summary "<summary>"` runs in dry-run mode.
- `specharbor generate <change-id> --agent-assisted --agent <agent-name> --type bugfix --title "<title>" --summary "<summary>"` runs in dry-run mode.
- `specharbor generate <change-id> --agent-assisted --agent <agent-name> --type docs --title "<title>" --summary "<summary>"` runs in dry-run mode.
- `specharbor generate <change-id> --agent-assisted --agent <agent-name> --type refactor --title "<title>" --summary "<summary>"` runs in dry-run mode.
- Supported agent-assisted authoring types are exactly `feature`, `bugfix`, `docs`, and `refactor`.

## Dry-Run Output

- Dry-run validates inputs before producing a report.
- Dry-run prints a deterministic authoring plan.
- Dry-run always prints the generated deterministic authoring prompt to stdout.
- Dry-run does not report a prompt-output path.
- Dry-run does not write a prompt file.
- The generated prompt is directly copy-pasteable into the selected agent.
- Dry-run reports that no external command was executed.
- Dry-run reports that no files were written.
- Dry-run reports that no prompt file was written.
- Dry-run reports that no agent output was parsed or applied.

## Dry-Run Safety

- Dry-run never executes external commands.
- Dry-run never executes local agent commands.
- Dry-run never calls AI providers.
- Dry-run never calls local models.
- Dry-run never calls external agents.
- Dry-run never calls network APIs.
- Dry-run never calls source-control APIs.
- Dry-run never calls workflow tools.
- Dry-run writes no files.
- Dry-run does not create `openspec/changes/<change-id>/`.
- Dry-run does not write `proposal.md`.
- Dry-run does not write `design.md`.
- Dry-run does not write `tasks.md`.
- Dry-run does not write `acceptance-criteria.md`.
- Dry-run does not write `risks.md`.
- Dry-run does not create or modify OpenSpec files.
- Dry-run does not modify production code.
- Dry-run does not modify README files.
- Dry-run does not modify docs.
- Dry-run does not modify CI files.
- Dry-run does not modify `.github/workflows/ci.yml`.
- Dry-run does not modify `.specharbor/config.yml`.
- Dry-run never parses, applies, or writes agent output, including OpenSpec files.
- No write/apply port is introduced.
- No confirmation flow is introduced.
- No `AgentRunner`, local command adapter, workflow connector, source-control automation, or execute use case path is introduced.

## Unsupported Execute

- `--execute` is explicitly unsupported in this first version.
- `--execute` returns a clear unsupported flag/mode error.
- `--execute` does not execute external commands.
- `--execute` does not execute local agent commands.
- `--execute` does not call provider APIs, local model APIs, network APIs, source-control APIs, or workflow tools.
- `--execute` does not parse, apply, or write agent output.
- `--execute` does not write OpenSpec files, production code, docs, README files, prompt files, CI files, or config files.

## Input Errors

- Missing `--agent` returns a clear error.
- `--agent` without a following value returns a clear error.
- Missing `--type` returns a clear error.
- `--type` without a following value returns a clear error.
- Missing `--title` returns a clear error.
- `--title` without a following value returns a clear error.
- Missing `--summary` returns a clear error.
- `--summary` without a following value returns a clear error.
- Empty agent, type, title, or summary values return clear errors when representable by the parser.
- Unknown authoring type returns a clear error.
- Unsafe change id returns a clear error.
- Unsafe change ids are rejected before prompt rendering.
- `--agent-assisted` with `--blank` returns a clear mode-conflict error.
- `--agent-assisted` with `--template` returns a clear mode-conflict error.
- `--agent-assisted` with `--guided` returns a clear mode-conflict error.
- Unsupported flags return clear errors.
- Extra positional arguments return clear errors.
- Duplicate conflicting flags return clear errors where the current parser supports duplicate validation.
- Guided-only input flags continue to require the correct generation mode where existing parser behavior requires that.

## Prompt Content

- The generated authoring prompt includes project context.
- The generated authoring prompt includes the change id.
- The generated authoring prompt includes the authoring type.
- The generated authoring prompt includes the title.
- The generated authoring prompt includes the summary.
- The generated authoring prompt includes the required OpenSpec files from `domain.RequiredOpenSpecChangeFiles()`.
- The generated authoring prompt includes `proposal.md`.
- The generated authoring prompt includes `design.md`.
- The generated authoring prompt includes `tasks.md`.
- The generated authoring prompt includes `acceptance-criteria.md`.
- The generated authoring prompt includes `risks.md`.
- The generated authoring prompt instructs the agent to create or refine only OpenSpec files under `openspec/changes/<change-id>/`.
- The generated authoring prompt explicitly instructs the agent not to implement production code.
- The generated authoring prompt explicitly instructs the agent not to modify unrelated files.
- For non-docs types, the generated authoring prompt prohibits README/docs changes.
- For `docs`, the generated authoring prompt permits documentation scope from the title and summary while still forbidding production code.
- The generated authoring prompt explicitly instructs the agent to leave implementation tasks unchecked.
- The generated authoring prompt explicitly instructs the agent to preserve architecture boundaries.
- The generated authoring prompt states that domain code belongs in `internal/core/domain`.
- The generated authoring prompt states that ports belong in `internal/core/ports`.
- The generated authoring prompt states that use cases belong in `internal/core/usecase`.
- The generated authoring prompt states that concrete implementations belong in `internal/adapters`.
- The generated authoring prompt states that core must not import adapters.
- The generated authoring prompt states that CLI must not contain business rules.
- The generated authoring prompt instructs the agent to run or recommend `specharbor validate <change-id>` when available.
- The generated authoring prompt defines clear Markdown-only OpenSpec output expectations.
- The generated authoring prompt does not ask the agent to run implementation, tests, source-control commands, workflow commands, provider setup, credential setup, commits, pushes, merges, deployment, or production code edits.
- The generated authoring prompt does not depend on a prompt file or on files written by SpecHarbor.

## Architecture

- CLI parsing remains in `internal/adapters/cli`.
- CLI report formatting remains in `internal/adapters/cli`.
- Agent-assisted spec authoring orchestration lives in `internal/core/usecase`.
- Agent-assisted domain concepts live in `internal/core/domain`.
- Prompt rendering goes through a core-owned port/interface.
- Concrete prompt template content lives in `internal/adapters/templates` or another justified adapter package.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, or concrete process execution packages.
- The command does not require provider API keys.
- The command does not store credentials.

## Regressions and Verification

- Existing `specharbor generate <change-id> --blank` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template feature` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template bugfix` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template docs` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template refactor` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type bugfix --title "<title>" --summary "<summary>"` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type docs --title "<title>" --summary "<summary>"` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type refactor --title "<title>" --summary "<summary>"` behavior remains unchanged.
- `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `help`, `version`, and unknown command behavior remain unchanged.
- No README changes are included.
- No docs changes outside this OpenSpec change are included.
- No CI changes are included.
- `.github/workflows/ci.yml` is not modified.
- `.specharbor/config.yml` is not modified.
- Focused tests cover domain concepts, prompt rendering, use case behavior, CLI parsing/reporting, dry-run safety, unsupported `--execute`, prompt copy-pasteability, README/docs prompt boundaries, mode conflicts, unsafe change ids, error cases, absence of write/apply ports, and regressions for existing generation modes.
- `go test ./...` succeeds.
