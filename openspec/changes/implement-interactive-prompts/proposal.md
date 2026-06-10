# Proposal: Implement Interactive Prompts

## Summary

Add a safe first interactive generation mode for SpecHarbor:

```bash
specharbor generate <change-id> --interactive
```

Interactive mode guides a user through selecting one existing deterministic generation path, collecting only the values required by that path, showing a deterministic summary with validation behavior and safety notes, and asking for confirmation before any OpenSpec files are written.

The first version supports local terminal interaction only. It does not call provider APIs, LLM APIs, local model APIs, agent runners, shell commands, scripts, source-control tools, workflow connectors, or remote execution systems. It does not write production code, docs, config, CI files, prompt files, archive files, source-control files, or arbitrary paths.

## Problem

SpecHarbor already supports several generation paths, but users must currently know and provide the correct flags up front:

- `--blank`
- `--template`
- `--custom-template`
- `--config-template`
- `--guided`
- `--agent-assisted`
- `--ai-assisted`
- `--hybrid`

This is efficient for automation and repeatable scripts, but it is less friendly when a user is starting from an idea and does not remember the exact flag combination. A prompt-driven path can improve discoverability, but it must not weaken SpecHarbor's deterministic behavior, CI compatibility, or architecture boundaries.

## Goal

Implement an interactive prompt foundation that:

- uses `specharbor generate <change-id> --interactive` as the first command surface;
- requires an interactive TTY and fails clearly when one is unavailable;
- prompts the user to choose one supported existing generation path;
- supports blank, built-in template, custom template, config template, and hybrid generation;
- excludes AI-assisted import, agent-assisted authoring, and guided generation from the first interactive menu;
- collects only the values needed for the selected path;
- validates prompt answers using the existing domain value objects and generation validation rules where applicable;
- limits invalid-answer retries to three attempts;
- prints a deterministic summary before generation, including selected validation behavior and safety notes;
- requires explicit confirmation before calling generation use cases;
- allows cancellation and EOF before writes;
- delegates actual generation to existing use cases;
- preserves existing non-interactive commands and flags unchanged;
- keeps terminal input and output outside core packages;
- remains testable with fake terminal input and output.

## Chosen Command Surface

Use:

```bash
specharbor generate <change-id> --interactive
```

Reasons:

- It extends the existing `generate` command instead of adding a new command tree.
- It keeps the change id explicit, matching existing write destinations and avoiding a prompt for output path identity.
- It is easy to discover from existing `generate` help and docs.
- It keeps automation users on the current non-interactive flag surface.
- It avoids a separate `specharbor interactive ...` namespace before there are multiple interactive workflows.

Rejected for the first version:

```bash
specharbor interactive generate <change-id>
specharbor generate --interactive
```

`specharbor interactive generate <change-id>` adds a command namespace before there is a clear need for more interactive subcommands. `specharbor generate --interactive` would require prompting for change id or changing the current positional contract, so it is deferred.

## Scope

- Add `--interactive` parsing to `specharbor generate <change-id>`.
- Reject duplicate `--interactive` flags clearly.
- Reject `--interactive` when combined with direct generation mode flags or direct generation input flags:
  - `--blank`
  - `--template`
  - `--custom-template`
  - `--config-template`
  - `--guided`
  - `--ai-assisted`
  - `--agent-assisted`
  - `--hybrid`
  - `--type`
  - `--title`
  - `--summary`
  - `--from-file`
  - `--agent`
  - `--execute`
  - `--overwrite`
- Require exactly one positional `<change-id>` with `--interactive`.
- Validate the change id using the existing `ChangeID` behavior before prompting.
- Detect non-interactive stdin before showing prompts.
- Prompt for one of the supported first-version paths:
  - blank;
  - built-in template;
  - custom template;
  - config template;
  - hybrid.
- For built-in template paths, prompt for one supported built-in template name.
- For custom template paths, prompt for one project-local custom template name and optional title/summary values.
- For config template paths, prompt for one config alias and optional title/summary values.
- For hybrid paths, prompt for exactly one hybrid source namespace, its source value, required title, required summary, and optional type.
- Print a deterministic summary of selected options, selected validation behavior, and safety notes before generation.
- Show `Validation: automatic no` for blank, built-in template, custom template, and config template paths.
- Show `Validation: automatic yes` for hybrid paths.
- Ask for confirmation before invoking any generation use case.
- Treat trimmed `y` and `yes` in any casing as confirmation.
- Treat trimmed `n` and `no` in any casing as cancellation.
- Treat EOF before confirmation as cancellation.
- Treat empty confirmation as cancellation.
- Retry unsupported confirmation answers up to three attempts, then fail clearly with no writes.
- Return non-zero on cancellation with a clear `operation cancelled` message.
- Delegate confirmed requests to existing blank, built-in template, custom template, config template, and hybrid generation behavior.
- Follow the selected mode's existing write behavior, existing-file preservation, and validation behavior.
- Update public docs in the implementation change: `README.md`, `docs/usage.md`, `docs/generation-modes.md`, and any CLI/interactive docs present at implementation time.
- Add domain, CLI, use case/integration, regression, architecture, safety, and documentation tests as described in `tasks.md`.

## Out of Scope

- Implementing code in this spec-authoring task.
- Modifying files outside `openspec/changes/implement-interactive-prompts/` in this spec-authoring task.
- Adding `specharbor interactive generate <change-id>`.
- Adding `specharbor generate --interactive` without a change id.
- Prompting for change id in the first version.
- Supporting direct interactive guided generation in the first version.
- Supporting AI-assisted from-file import in interactive mode.
- Supporting agent-assisted authoring in interactive mode.
- Executing local or remote agents.
- Calling provider APIs.
- Calling LLM APIs.
- Calling local model APIs.
- Calling remote APIs except existing remote-template fetch behavior reached through an existing config alias after user confirmation.
- Adding direct remote template prompts, URL prompts, checksum prompts, marketplace prompts, or git clone behavior.
- Running shell commands.
- Running scripts.
- Adding external TUI frameworks.
- Adding workflow automation.
- Adding source-control automation.
- Writing production code.
- Writing docs, config, CI, prompt, archive, source-control, or arbitrary files from interactive generation.
- Adding validation auto-fix.
- Auto-commit.
- Auto-push.
- Creating pull requests.
- Merging.
- Archiving automatically.
- Redesigning existing generation modes.
- Changing existing non-interactive generation behavior.

## Success Criteria

- `specharbor generate <change-id> --interactive` starts a local terminal prompt flow only when stdin is a TTY.
- Non-TTY execution fails immediately with a clear `interactive mode requires a TTY` error and does not hang.
- Interactive mode rejects mixed direct generation flags before prompting.
- Prompt answers map to existing generation requests.
- Blank, built-in template, custom template, config template, and hybrid prompt flows work.
- AI-assisted, agent-assisted, and direct guided generation are not offered in the first interactive menu.
- Invalid prompt answers retry up to three attempts, then fail clearly with no writes.
- Confirmation is required before generation and is matched case-insensitively.
- The pre-confirmation summary always includes whether validation is automatic.
- The pre-confirmation summary always includes deterministic safety notes before the confirmation prompt.
- Cancellation and EOF write nothing and exit non-zero with `operation cancelled`.
- Unsupported confirmation answers retry up to three attempts, then fail clearly with no writes.
- Writes remain limited to `openspec/changes/<change-id>/` and the five required OpenSpec files.
- Existing files are preserved according to the selected mode.
- Validation behavior follows the selected mode: hybrid keeps its validation integration; direct blank/template/custom/config behavior remains unchanged.
- Core packages do not read stdin, write stdout/stderr, import terminal packages, import CLI packages, import adapters, or depend on concrete terminal implementations.
- Existing non-interactive commands remain unchanged.
- Documentation explains command syntax, prompt flow, non-TTY behavior, cancellation, confirmation, validation/write behavior, safety boundaries, and examples.
