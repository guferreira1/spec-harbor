# Acceptance Criteria: Implement Custom Templates

## Command behavior

- `specharbor generate <change-id> --custom-template <template-name>` generates `openspec/changes/<change-id>/` with the five required OpenSpec change files rendered from `.specharbor/templates/<template-name>/`.
- `--title` and `--summary` are optionally accepted with `--custom-template`; their existing required behavior with `--guided` and `--agent-assisted` is unchanged, and they remain rejected with `--blank` and `--template`.
- `--custom-template` combined with `--blank`, `--template`, `--guided`, `--agent-assisted`, `--type`, `--agent`, or `--execute` fails with a clear error; duplicate flags, missing flag values, unsupported flags, extra positional arguments, and a missing change id fail with clear errors.
- Running the command outside an initialized OpenSpec project keeps the existing guidance to run `specharbor init` first.

## Template identity and discovery

- Custom template names are validated before any filesystem access: empty names, names with `/` or `\`, `.` or `..` or embedded `..` sequences, leading `.` or `-`, characters outside `[A-Za-z0-9._-]`, and names longer than 128 characters are rejected with clear errors naming the custom template concept.
- Templates resolve only from `.specharbor/templates/<template-name>/` under the project root; there is no search path, fallback, registry, or network lookup.
- A missing template directory fails with an error naming the expected `.specharbor/templates/<template-name>` path.

## Template content requirements

- A valid custom template contains all five files: `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, `risks.md`; missing files fail with one error listing every missing filename.
- Empty or whitespace-only template files fail with a clear error naming the template and file.
- Unknown extra files and subdirectories in the template directory are ignored and never read or copied.
- `{{change_id}}` is substituted in every generated file; `{{title}}` and `{{summary}}` are substituted when provided and remain verbatim when omitted; unknown `{{...}}` tokens remain verbatim; no other content is altered.

## Write safety

- Files are written only under `openspec/changes/<change-id>/` and only with the five known filenames; template content cannot influence output paths.
- Existing files in the change directory are never overwritten; they are reported as skipped, matching existing generation behavior.
- Any template validation failure (invalid name, missing directory, missing files, empty files) produces zero filesystem writes, including no change-directory creation.
- The command writes no production code, no documentation, no configuration, and no CI files.

## CLI output

- Success output includes the change id, the template name labeled as custom, the relative template source path, the change path, the directory status, created files, skipped existing files when present, and a note that only OpenSpec change files were written.
- Error cases exit non-zero with a single clear error message; success exits zero.

## Compatibility and regression

- Built-in `--template` behavior is byte-identical to before this change: same names, same content, same output, and the same unknown-template error for non-built-in names, even when a custom template directory with the same name exists.
- Blank, guided, and agent-assisted generation, `validate`, `prompt`, `review`, `archive`, and `config` behavior are unchanged, proven by regression tests.
- Changes generated from custom templates work with `specharbor validate <change-id>` without validator changes.

## Architecture

- The custom template name, model, allowed-file list, and rendering rules live in `internal/core/domain` as pure logic; orchestration lives in `internal/core/usecase`; the new read port lives in `internal/core/ports`; `LocalFileSystem` in `internal/adapters/filesystem` satisfies it without changes; parsing and report formatting live in `internal/adapters/cli`.
- Core packages gain no imports of adapters, CLI packages, `os`, network clients, or provider/source-control/workflow/agent SDKs; existing architecture boundary tests keep passing.
- The feature performs no external command execution and no network calls.

## Documentation

- `README.md`, `docs/usage.md`, and `docs/generation-modes.md` document the template directory structure, required files, command usage, built-in vs custom flag model, variable substitution behavior, validation expectations, safety boundaries, and the explicit non-goals: no remote templates, no config-driven registry, and no template or script execution.
- All documented command examples match implemented behavior.

## Verification

- `gofmt` reports no unformatted files and `go test ./...` passes.
- Manual verification covers successful generation, skip-existing re-runs, each documented error case, and validating a generated change.
