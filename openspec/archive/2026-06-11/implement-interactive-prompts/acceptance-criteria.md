# Acceptance Criteria: Implement Interactive Prompts

## Command Behavior

- `specharbor generate <change-id> --interactive` starts the interactive prompt flow when stdin is a TTY.
- `specharbor generate --interactive` fails clearly because change id is required.
- `--interactive` specified more than once fails clearly.
- `--interactive` rejects extra positional arguments.
- `--interactive` cannot be combined with `--blank`, `--template`, `--custom-template`, `--config-template`, `--guided`, `--hybrid`, `--ai-assisted`, or `--agent-assisted`.
- `--interactive` cannot be combined with `--type`, `--title`, `--summary`, `--from-file`, `--agent`, `--execute`, or `--overwrite`.
- Existing non-interactive generation commands keep their current behavior.

## Non-TTY Behavior

- Interactive mode checks TTY availability before showing prompts.
- When stdin is not a terminal, the command fails with `interactive mode requires a TTY`.
- Non-TTY failure exits non-zero.
- Non-TTY failure writes nothing.
- Non-TTY failure does not hang.
- Automation remains available through existing non-interactive flags.

## Supported Paths

- Interactive mode supports blank generation.
- Interactive mode supports built-in template generation.
- Interactive mode supports custom template generation.
- Interactive mode supports config template generation.
- Interactive mode supports hybrid generation with a built-in source.
- Interactive mode supports hybrid generation with a custom source.
- Interactive mode supports hybrid generation with a config source.
- Interactive mode does not offer direct guided generation in the first version.
- Interactive mode does not offer AI-assisted generation in the first version.
- Interactive mode does not offer agent-assisted generation in the first version.

## Prompt Flow

- The generation path menu has deterministic ordering.
- The built-in template menu has deterministic ordering: feature, bugfix, docs, refactor.
- Hybrid source selection has deterministic ordering: built-in template, custom template, config template.
- Blank generation asks no unnecessary path-specific questions.
- Built-in template generation asks for a template name only.
- Custom template generation asks for a custom template name and optional title/summary.
- Config template generation asks for a config alias and optional title/summary.
- Hybrid generation asks for source namespace, source value, required title, required summary, and optional type.
- Prompt answers are trimmed before validation and request mapping.
- Prompt output is deterministic enough for tests and documentation.

## Input Validation

- Interactive change id validation uses existing `ChangeID` behavior.
- Built-in template validation uses existing `TemplateName` behavior.
- Custom template validation uses existing `CustomTemplateName` behavior.
- Config alias validation uses existing `ConfigTemplateAlias` behavior.
- Hybrid title validation requires a non-empty trimmed value.
- Hybrid summary validation requires a non-empty trimmed value.
- Hybrid type validation uses the existing supported set: `feature`, `bugfix`, `docs`, and `refactor`.
- Empty optional custom-template title is treated as omitted.
- Empty optional custom-template summary is treated as omitted.
- Empty optional config-template title is treated as omitted.
- Empty optional config-template summary is treated as omitted.
- Empty optional hybrid type is treated as omitted.
- Invalid required answers retry up to three attempts.
- Invalid non-empty optional hybrid type answers retry up to three attempts.
- Retry exhaustion fails clearly and writes nothing.

## Confirmation And Cancellation

- Interactive mode prints a summary before generation.
- The summary includes change id.
- The summary includes selected generation path.
- The summary includes selected template when applicable.
- The summary includes selected custom template when applicable.
- The summary includes selected config alias when applicable.
- The summary includes selected hybrid source when applicable.
- The summary includes title, summary, and type only when applicable.
- The summary includes the target write directory.
- The summary includes the five approved OpenSpec filenames.
- The summary includes `Validation: automatic no` for blank generation.
- The summary includes `Validation: automatic no` for built-in template generation.
- The summary includes `Validation: automatic no` for custom template generation.
- The summary includes `Validation: automatic no` for config template generation.
- The summary includes `Validation: automatic yes` for hybrid generation.
- The summary includes the deterministic safety notes before the confirmation prompt.
- Generation does not start until the user confirms.
- `y` confirms.
- `Y` confirms.
- `yes` confirms.
- `YES` confirms.
- `Yes` confirms.
- `n` cancels.
- `N` cancels.
- `no` cancels.
- `NO` cancels.
- `No` cancels.
- Confirmation answers are trimmed and matched case-insensitively.
- Empty confirmation input cancels.
- EOF cancels.
- Cancellation exits non-zero with `operation cancelled`.
- Cancellation writes nothing.
- Unsupported confirmation answers re-prompt within the same three-attempt retry policy.
- Confirmation retry exhaustion fails clearly and writes nothing.

## Request Mapping

- Blank prompt answers delegate to existing blank generation behavior.
- Built-in template prompt answers delegate to existing built-in template generation behavior.
- Custom template prompt answers delegate to existing custom template generation behavior.
- Config template prompt answers delegate to existing config template generation behavior.
- Hybrid built-in prompt answers delegate to existing hybrid built-in generation behavior.
- Hybrid custom prompt answers delegate to existing hybrid custom generation behavior.
- Hybrid config prompt answers delegate to existing hybrid config generation behavior.
- Source existence, config alias lookup, remote fetch, checksum validation, archive validation, rendering, write policy, and validation integration remain owned by existing use cases and adapters.

## Write Safety

- Interactive mode performs no writes before confirmation.
- Interactive generation writes only under `openspec/changes/<change-id>/`.
- Interactive generation writes only `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- Existing files are preserved according to the selected mode's existing behavior.
- Template content, config values, remote archive paths, and prompt answers cannot choose arbitrary output paths.
- Interactive mode writes no production code.
- Interactive mode writes no docs files.
- Interactive mode writes no config files.
- Interactive mode writes no CI files.
- Interactive mode writes no prompt files.
- Interactive mode writes no archive files.
- Interactive mode writes no source-control files.
- Interactive mode writes no arbitrary files.

## Validation Behavior

- Interactive blank generation follows blank generation's existing validation behavior.
- Interactive built-in template generation follows built-in template generation's existing validation behavior.
- Interactive custom template generation follows custom template generation's existing validation behavior.
- Interactive config template generation follows config template generation's existing validation behavior.
- Interactive hybrid generation follows hybrid generation's existing validation behavior.
- Blank, built-in template, custom template, and config template summaries show `Validation: automatic no`.
- Hybrid summaries show `Validation: automatic yes`.
- Interactive mode does not add a post-generation validation prompt in the first version.
- Interactive mode does not auto-fix validation findings.

## Remote Template Behavior

- Interactive mode does not prompt for remote URLs.
- Interactive mode does not prompt for checksums.
- Interactive mode does not add `--remote-template`.
- Remote templates are reachable only through existing config aliases.
- Existing remote template HTTPS, checksum, no-redirect, no-credential, no-query, no-fragment, ZIP, archive path, symlink, executable-bit, duplicate, extra-file, missing-file, empty-file, and size-limit safeguards remain unchanged.
- Remote fetch behavior can happen only after user confirmation through existing config-template or hybrid config-template generation behavior.
- Remote output never displays credentials, query-token values, auth headers, cookies, OAuth material, or environment-derived secrets.

## CLI Output

- Non-TTY error output is clear.
- Cancellation output is clear.
- Retry exhaustion output is clear.
- Pre-confirmation summary is deterministic.
- Pre-confirmation summary includes the validation behavior line for every supported path.
- Pre-confirmation summary includes the deterministic safety section before confirmation.
- Successful output reuses existing selected-mode output as much as possible.
- Hybrid output keeps its validation and safety report behavior.
- Error output follows existing CLI style.
- Prompt output does not include secrets.

## Architecture

- TTY detection lives outside core.
- Prompt rendering lives outside core.
- Answer parsing lives outside core.
- Retry loops live outside core.
- Confirmation handling lives outside core.
- Core packages do not import terminal packages.
- Core packages do not import adapters.
- Core packages do not import CLI packages.
- Core packages do not import `os`.
- Core packages do not import `os/exec`.
- Core packages do not read stdin.
- Core packages do not write stdout or stderr.
- Use cases remain non-interactive and accept structured inputs.
- CLI prompt code maps confirmed answers to existing use case inputs.
- No provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, shell execution, script execution, or external TUI frameworks are introduced for this feature.

## Safety Boundaries

- Interactive mode does not call provider APIs.
- Interactive mode does not call LLM APIs.
- Interactive mode does not call local model APIs.
- Interactive mode does not execute agents.
- Interactive mode does not run shell commands.
- Interactive mode does not run scripts.
- Interactive mode does not parse or apply live runner output.
- Interactive mode does not import AI-authored files.
- Interactive mode does not run source-control commands.
- Interactive mode does not run workflow commands.
- Interactive mode does not commit.
- Interactive mode does not push.
- Interactive mode does not create pull requests.
- Interactive mode does not merge.
- Interactive mode does not archive automatically.

## Documentation

- `README.md` documents interactive generation after implementation.
- `docs/usage.md` documents command syntax, prompt sequence, deterministic pre-write summary, confirmation, cancellation, non-TTY behavior, validation behavior, write behavior, and examples.
- `docs/generation-modes.md` explains interactive generation as a prompt layer over existing modes.
- Any CLI or interactive docs present at implementation time are updated.
- Documentation states that `<change-id>` remains required.
- Documentation states that before writing, interactive mode prints a deterministic summary.
- Documentation states that the summary includes whether validation is automatic.
- Documentation states that blank, built-in template, custom template, and config template paths show `Validation: automatic no`.
- Documentation states that hybrid paths show `Validation: automatic yes`.
- Documentation states that the summary includes deterministic safety notes.
- Documentation states that confirmation is case-insensitive.
- Documentation states that `y` and `yes` proceed.
- Documentation states that `n` and `no` cancel.
- Documentation states that empty confirmation and EOF cancel.
- Documentation states that cancellation writes nothing and exits non-zero.
- Documentation states that direct guided generation remains a non-interactive flag mode in the first version.
- Documentation states that AI-assisted and agent-assisted generation are not offered by interactive prompts in the first version.
- Documentation states that remote templates are reachable only through existing config aliases.
- Documentation states all provider, agent, shell, workflow, source-control, and production-code write exclusions.

## Tests And Regression

- Domain tests cover reuse of value objects and absence of terminal logic in domain.
- CLI tests cover parsing, non-TTY behavior, prompt sequences, invalid input retries, retry exhaustion, cancellation, confirmation, EOF, output summaries, and no-write guarantees before confirmation or on cancellation.
- Integration tests cover prompt-answer mapping to existing generation requests and selected-mode delegation.
- Regression tests prove existing generation modes and non-generation commands remain unchanged.
- Architecture tests prove core import boundaries and safety exclusions.
- Documentation tests or checks cover the new public CLI behavior if the project has documentation architecture tests for related features.

## Verification

- `go test ./...` passes after implementation.
- `go run ./cmd/specharbor validate implement-interactive-prompts` passes for this OpenSpec change.
- Implementation updates `tasks.md` only for work actually completed.
