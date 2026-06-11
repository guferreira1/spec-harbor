# Acceptance Criteria: Implement AI-Assisted Generation

## Command Behavior

- `specharbor generate <change-id> --ai-assisted --from-file <agent-output-file>` imports strict local AI-authored output and writes approved OpenSpec files.
- `specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --overwrite` explicitly permits replacing existing required OpenSpec files.
- `--from-file` is required with `--ai-assisted`.
- `--from-file` without `--ai-assisted` is rejected clearly.
- `--overwrite` without `--ai-assisted` is rejected clearly.
- `--ai-assisted` cannot be combined with `--blank`, `--template`, `--guided`, or `--agent-assisted`.
- `--ai-assisted --execute` is rejected clearly.
- Unsupported flags, duplicate flags, extra positional arguments, missing change id, and unsafe change id return clear errors.

## AI Output Format

- The parser accepts blocks whose start line exactly matches `---FILE: <filename>---`.
- The parser accepts block end lines that exactly match `---END FILE---`.
- The parser accepts only `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- Unknown filenames are rejected.
- Duplicate file blocks are rejected.
- Missing required file blocks are rejected.
- Empty or whitespace-only generated content is rejected.
- Path traversal filenames are rejected.
- Absolute filenames are rejected.
- Nested filenames and filenames containing `/` or `\` are rejected.
- Text outside file blocks is rejected unless it is whitespace.
- Unclosed blocks and orphan end markers are rejected.
- Parser findings include stable codes and clear messages.
- Parser findings include filenames and line numbers where useful.
- Malformed AI output writes nothing.
- The parser is deterministic and local.
- The parser does not execute content.
- The parser does not interpret shell commands.
- The parser does not apply patches.

## Write Behavior

- The change id is validated with the domain `ChangeID` value object before source reads and filesystem writes.
- The command requires existing OpenSpec project structure.
- The command creates `openspec/changes/<change-id>/` if it is missing and the parsed source is valid.
- If the target path exists and is not a directory, the command fails before writes.
- Writes are limited to `openspec/changes/<change-id>/`.
- Writes are limited to the five required OpenSpec filenames.
- AI-provided paths are never used as write paths.
- Existing files are skipped by default and reported as skipped.
- Existing files are overwritten only when `--overwrite` is present.
- Generated files are reported.
- Overwritten files are reported.
- Parser errors, source read errors, unsafe change ids, target preflight errors, and unsupported flags write nothing.
- Runtime write failures are reported clearly.
- The implementation does not automatically delete or roll back user files after a runtime write failure.

## Validation Integration

- Validation runs after all planned writes complete successfully.
- Validation uses the existing validation semantics and finding model.
- Validation status is printed.
- Validation error and warning counts are printed.
- Validation findings are printed with severity, code, message, and path.
- Validation warnings alone keep the command exit code `0`.
- Validation errors make the command exit non-zero after the generation and validation report is printed.
- Parser, source read, argument, and write errors make the command exit non-zero.
- Malformed AI output does not trigger target validation because no files are written.

## CLI Output

- Success output includes the change id.
- Success output includes the source AI output file path supplied by the user.
- Success output includes the target change path.
- Success output includes whether the target directory was created or already existed.
- Success output includes overwrite status.
- Success output includes generated files.
- Success output includes skipped files when any exist.
- Success output includes overwritten files when `--overwrite` replaces files.
- Failure output for parse errors includes stable parse codes and clear findings.
- Output includes validation status and validation finding summary after successful writes.
- Output includes safety notes stating that no provider APIs were called.
- Output includes safety notes stating that no remote AI services were called.
- Output includes safety notes stating that no agent commands were executed.
- Output includes safety notes stating that no production code was modified.
- Output includes safety notes stating that no source-control commands were run.
- Output includes safety notes stating that no auto-commit, auto-push, PR creation, merge, or archive was performed.
- Output ordering is deterministic.

## Architecture

- AI output block models and parser rules live in `internal/core/domain`.
- Allowed generated filename policy lives in `internal/core/domain`.
- AI-assisted generation result concepts live in `internal/core/domain`.
- Orchestration lives in `internal/core/usecase`.
- Source reads and OpenSpec target writes go through small core-owned ports.
- Local filesystem reads and writes live in adapters.
- CLI parsing, dependency wiring, output formatting, and exit-code mapping live in `internal/adapters/cli`.
- Core packages do not import adapters.
- Core packages do not import CLI packages.
- Core packages do not import `os`.
- Core packages do not import `os/exec`.
- Core packages do not perform terminal IO.
- Core packages do not import provider SDKs, network APIs, source-control SDKs, workflow SDKs, or external-agent SDKs.
- No provider API adapter is introduced.
- No workflow connector adapter is introduced.
- No source-control adapter is introduced.
- No patch application abstraction is introduced.
- No arbitrary path write abstraction is introduced.

## Safety Boundaries

- No provider APIs are called.
- No remote AI services are called.
- No local model APIs are called.
- No OAuth is introduced.
- No credentials are introduced or stored.
- No remote execution is introduced.
- No cloud execution is introduced.
- No IDE automation is introduced.
- No marketplace integration is introduced.
- No live agent runner output is applied.
- No production code is modified by this feature.
- No docs outside the active OpenSpec change are modified by this command.
- No config files are modified by this command.
- No CI files are modified by this command.
- No source-control files are modified by this command.
- No source-control commands are run.
- No workflow commands are run.
- No auto-commit, auto-push, PR creation, merge, or automatic archive is performed.

## Documentation

- `README.md` documents AI-assisted from-file generation after implementation.
- `docs/usage.md` documents syntax, examples, strict output format, skip behavior, overwrite behavior, validation behavior, exit codes, and safety notes.
- `docs/generation-modes.md` identifies AI-assisted from-file generation as implemented after this feature is implemented.
- Documentation does not claim provider API integration.
- Documentation does not claim remote AI service integration.
- Documentation does not claim production code generation.
- Documentation does not claim source-control automation.

## Tests and Regression

- Domain tests cover allowed generated file names, rejected unknown file names, rejected duplicate file blocks, rejected path traversal, rejected absolute paths, parser success, parser malformed input, missing required file blocks, empty content behavior, generation result model, and defensive copying.
- Use case tests cover valid generation, malformed output writing nothing, unknown filenames writing nothing, duplicate blocks writing nothing, missing blocks writing nothing, default no-overwrite behavior, explicit overwrite behavior, write scope, validation integration, validation warnings, validation errors, and unsafe change id behavior.
- Ports and adapter tests cover source file reads, missing source file errors, approved target writes, traversal prevention, and write errors.
- CLI tests cover command parsing, required arguments, unsupported flags, extra args, success output, parse error output, overwrite behavior, validation report integration, and exit-code behavior.
- Regression tests prove existing generate modes remain unchanged.
- Regression tests prove validate remains unchanged except being invoked by this new command.
- Regression tests prove prompt, review, archive, config, scan, init, workflow, help, and version remain unchanged.
- Architecture tests prove core import boundaries and absence of provider, network, source-control, workflow, external-agent, production-code-write, arbitrary-path-write, and patch-application behavior.
- `go test ./...` succeeds after implementation.
