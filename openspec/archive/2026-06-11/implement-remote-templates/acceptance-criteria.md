# Acceptance Criteria: Implement Remote Templates

## Config Schema

- `.specharbor/config.yml` supports `templates.aliases.<alias>.source: remote`.
- A remote alias requires `url`, `checksum`, and `format`.
- `format` supports exactly `zip`.
- `checksum` supports exactly `sha256:<64 hex characters>`.
- Remote aliases reject unsupported fields, including `template`, `path`, `headers`, `auth`, `command`, `script`, `branch`, `tag`, and `ref`.
- Builtin and custom aliases remain source-specific and do not accept remote-only fields unless a later change explicitly adds them.
- Omitted `templates`, omitted `templates.aliases`, config version handling, and existing builtin/custom alias parsing remain unchanged.

## URL And Checksum Validation

- Remote URL validation happens before network access.
- Valid remote URLs must use HTTPS and include a host and path.
- HTTP, file, SSH, git, git+ssh, FTP, SCP-style git targets, missing hosts, missing paths, whitespace, control characters, and over-length URLs are rejected.
- URLs with userinfo credentials are rejected.
- URLs with query strings are rejected.
- URLs with fragments are rejected.
- Missing checksums are rejected.
- Unsupported checksum algorithms are rejected.
- Malformed SHA-256 digests are rejected.
- Downloaded bytes are verified against the configured checksum before archive parsing.
- Checksum mismatch produces a clear error and writes nothing.

## Network Behavior

- Network access occurs only when the requested `--config-template <alias>` resolves to `source: remote`.
- Direct built-in template generation performs no network access.
- Direct custom template generation performs no network access.
- Config-template aliases with `source: builtin` perform no network access.
- Config-template aliases with `source: custom` perform no network access.
- Missing OpenSpec project structure fails before network access.
- The HTTP adapter uses `GET` only.
- The HTTP adapter sends no credentials, auth headers, cookies, user-provided headers, or environment-expanded tokens.
- Redirects are not followed and produce a clear error.
- Non-`200` responses produce a clear error.
- Timeouts produce a clear error.
- Max response size failures produce a clear error.

## Remote Format

- Only zip archives are supported.
- The zip archive contains the five required OpenSpec files as root-level regular file entries:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- Missing required files are rejected.
- Empty or whitespace-only required files are rejected.
- Nested paths are rejected.
- Absolute paths are rejected.
- Path traversal is rejected.
- Windows drive paths are rejected.
- Symlink entries are rejected.
- Executable entries are rejected.
- Duplicate required file entries are rejected.
- Extra file entries are rejected.
- Malformed zip archives are rejected.
- Downloaded and uncompressed size limits are enforced.

## Command Behavior

- `specharbor generate <change-id> --config-template <alias>` resolves a remote alias from `.specharbor/config.yml`.
- No separate remote-template command or flag is introduced.
- Existing config-template conflict behavior remains intact.
- Existing direct built-in, direct custom, config builtin, and config custom behavior remains unchanged.
- Same alias names across builtin, custom, and remote source references remain controlled by the explicit config alias source.
- Remote template generation does not auto-run validation.
- Generated remote-template changes are compatible with `specharbor validate <change-id>`.

## Write Safety

- Generated files are written only under `openspec/changes/<change-id>/`.
- Generated filenames are limited to the five required OpenSpec filenames.
- Existing files are skipped and never overwritten.
- Remote archive paths never influence output paths.
- Invalid config, invalid URL, missing checksum, network errors, timeout, size-limit errors, checksum mismatch, malformed archive, unsafe archive paths, symlink entries, executable entries, missing files, and empty files write nothing.
- The command writes no production code, documentation, config, CI files, prompt files, archive files, source-control files, or arbitrary output paths.
- The command performs no source-control automation, workflow automation, auto-commit, auto-push, PR creation, merge, or archive.

## CLI Output

- Success output includes the change id.
- Success output includes the config alias.
- Success output identifies the resolved source as `remote`.
- Success output includes the sanitized remote host.
- Success output includes the remote format.
- Success output includes the checksum algorithm.
- Success output includes the change path and directory status.
- Success output lists created files.
- Success output lists skipped existing files when present.
- Success output includes safety notes that remote access used only the explicit configured alias, checksum verification happened before archive parsing, and only OpenSpec change files were written.
- Error output clearly distinguishes URL validation, checksum validation, network, timeout, response size, checksum mismatch, archive, missing-file, empty-file, and write failures.
- Output never displays credentials, query-token values, auth headers, cookies, or environment-derived secrets.

## Architecture

- Remote template source models live in `internal/core/domain`.
- URL validation rules that are pure live in `internal/core/domain`.
- Checksum parsing and verification over bytes live in core.
- Archive file policy concepts live in core.
- Remote generation orchestration lives in `internal/core/usecase`.
- Remote fetching is accessed through a core-owned port.
- Zip bundle reading is accessed through a core-owned port or equivalent narrow abstraction.
- HTTP client behavior lives in adapters.
- Zip decoding lives in adapters unless implementation documents a safer pure-core alternative.
- CLI parsing, dependency wiring, output formatting, and exit-code mapping live in `internal/adapters/cli`.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, concrete HTTP clients, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.
- Use cases depend on interfaces.
- CLI code contains no URL, checksum, archive, or generation business rules.

## Documentation

- `README.md`, `docs/usage.md`, `docs/generation-modes.md`, and any config/template-specific docs describe remote templates after implementation.
- Documentation includes the `.specharbor/config.yml` schema for remote aliases.
- Documentation explains the checksum requirement and checksum mismatch behavior.
- Documentation explains HTTPS-only behavior and redirect rejection.
- Documentation explains the supported zip archive structure.
- Documentation explains command usage through `--config-template`.
- Documentation explains skip-existing write behavior and validation guidance.
- Documentation states that no credentials, OAuth, environment token expansion, marketplace search, git clone, script execution, shell execution, provider APIs, production code writes, source-control automation, PR creation, merge, or archive automation are introduced.

## Tests And Regression

- Domain tests cover remote source model, URL validation, HTTPS-only behavior, non-HTTPS rejection, credential/query/fragment rejection, missing checksum, checksum parser, unsupported checksum algorithm, checksum mismatch, source type validation, and archive path policy.
- Use case tests cover successful remote alias generation, missing URL, invalid URL, non-HTTPS URL, missing checksum, checksum mismatch, network failure, timeout, max size exceeded, unsafe archive path, symlink archive entry rejection, executable entry rejection, missing required files, empty required files, write boundaries, existing-file skip behavior, no network access for builtin/custom aliases, and no production code writes.
- Adapter tests cover HTTP timeout, max download size, no credentials sent, redirect rejection, non-`200` responses, zip parsing, traversal rejection, absolute path rejection, nested path rejection, symlink rejection, executable entry rejection, duplicate/extra file rejection, missing files, and checksum-before-parsing orchestration.
- CLI tests cover remote alias output, sanitized host output, checksum output, network/checksum/archive errors, and unchanged existing generation modes.
- Regression tests prove built-in template generation, custom template generation, config-template builtin/custom generation, AI-assisted generation, validate, prompt, review, archive, config, workflow, scan, init, help, version, and agent-assisted behavior remain unchanged.
- Architecture tests prove import boundaries and absence of provider/source-control/workflow/agent SDKs, shell/script execution, arbitrary output path writes, and core HTTP client imports.

## Verification

- `go test ./...` passes after implementation.
- `go run ./cmd/specharbor validate implement-remote-templates` passes for this OpenSpec change.
- Implementation updates `tasks.md` only for work actually completed.
