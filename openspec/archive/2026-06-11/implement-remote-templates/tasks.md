# Tasks: Implement Remote Templates

## Phase 0 - Baseline and scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-remote-templates/`.
- [x] Inspect current config-template, custom-template, generation, filesystem, config parser, CLI, validation, and architecture test patterns before editing.
- [x] Keep implementation limited to `source: remote` config aliases resolved through `specharbor generate <change-id> --config-template <alias>`.
- [x] Preserve existing `--template`, `--custom-template`, config-template `builtin`/`custom`, blank, guided, AI-assisted, agent-assisted, validate, prompt, review, archive, config, workflow, scan, init, help, and version behavior.
- [x] Do not implement a `--remote-template` flag, marketplace search, automatic discovery, git clone, SSH URLs, credentials, OAuth, provider APIs, shell execution, script execution, source-control automation, workflow automation, production code writes, auto-commit, PR creation, merge, or automatic archive.

## Phase 1 - Domain model

- [x] Extend the config template source kind model with `remote` while preserving existing `builtin` and `custom` behavior.
- [x] Add a remote template reference model for `url`, `checksum`, and `format`.
- [x] Add pure remote URL validation for HTTPS-only URLs with host and path, no credentials, no query string, no fragment, no whitespace/control characters, no local file URLs, no git URLs, no SSH URLs, and bounded length.
- [x] Add a remote template format model that accepts only `zip`.
- [x] Add a checksum model that accepts only `sha256:<64 hex characters>`, normalizes hex, and rejects missing, malformed, or unsupported algorithms.
- [x] Add pure checksum verification over downloaded bytes.
- [x] Add archive path policy concepts for the five required root-level filenames, rejecting traversal, absolute paths, nested paths, Windows drive paths, duplicates, extra files, symlinks, executable entries, missing files, and empty files.
- [x] Add a remote template bundle/domain model holding exactly the five required OpenSpec file contents with defensive copying.
- [x] Extend generation result concepts to report config alias, remote source kind, remote host, remote format, checksum algorithm, created files, skipped files, and change path.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, concrete HTTP clients, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, and process execution packages.

## Phase 2 - Domain tests

- [x] Test remote source type validation accepts `remote` and still accepts `builtin` and `custom`.
- [x] Test unsupported source types fail clearly.
- [x] Test remote URL validation accepts valid HTTPS URLs with host and path.
- [x] Test remote URL validation rejects missing URL, invalid URL, `http`, `file`, `ssh`, `git`, `git+ssh`, SCP-style git targets, missing host, missing path, credentials, query string, fragment, whitespace, control characters, and over-length values.
- [x] Test remote format validation accepts only `zip`.
- [x] Test missing format and unsupported formats fail clearly.
- [x] Test checksum parser accepts valid `sha256:<hex>` values and normalizes uppercase hex if accepted.
- [x] Test missing checksum, malformed checksum, short/long digest, non-hex digest, and unsupported algorithms fail clearly.
- [x] Test checksum verification succeeds for matching bytes and fails with expected and actual digest facts for mismatches.
- [x] Test archive path policy accepts only the five required root-level filenames.
- [x] Test archive path policy rejects traversal, absolute paths, nested paths, Windows drive paths, duplicate files, extra files, symlinks, executable entries, missing required files, and empty required files.
- [x] Test remote bundle and result models defensively copy slices, maps, and byte-derived values where applicable.

## Phase 3 - Config parsing

- [x] Extend `.specharbor/config.yml` parsing so `templates.aliases.<alias>.source: remote` is supported.
- [x] Require `url`, `checksum`, and `format` for remote alias entries.
- [x] Reject `template` and any unknown fields for remote alias entries.
- [x] Preserve existing strict source-specific validation for `builtin` and `custom` aliases.
- [x] Reject `url`, `checksum`, `format`, path-like, auth-like, script-like, git-like, or command-like fields for `builtin` and `custom` aliases unless this change explicitly supports them.
- [x] Keep omitted `templates` and omitted `templates.aliases` behavior unchanged.
- [x] Keep supported config version behavior unchanged.
- [x] Add config adapter tests for valid remote aliases and each invalid remote schema case.
- [x] Add regression tests proving existing config parsing for `builtin`, `custom`, omitted aliases, invalid YAML, and unsupported versions remains unchanged.

## Phase 4 - Ports and adapters

- [x] Add a small core-owned remote template fetch port that returns downloaded bytes and safe response metadata without exposing HTTP client types to core.
- [x] Add a small core-owned remote template bundle reader port that decodes a zip byte slice into the approved file map without exposing zip implementation types to core.
- [x] Implement an HTTP remote template fetch adapter.
- [x] Use only `GET` requests.
- [x] Enforce a total request timeout, recommended default `15s`.
- [x] Disable redirect following and return a clear unsupported-redirect error for `3xx` responses.
- [x] Enforce the max HTTP response body size, recommended default `5 MiB`, while reading.
- [x] Reject non-`200` responses clearly.
- [x] Send no credentials, auth headers, cookies, custom user-provided headers, or environment-expanded tokens.
- [x] Implement zip bundle decoding in an adapter.
- [x] Enforce exact required regular file entries and the uncompressed size limits from `design.md`.
- [x] Reject malformed zips, path traversal, absolute paths, nested paths, Windows drive paths, duplicate files, extra files, symlinks, executable entries, missing required files, and empty files.
- [x] Ensure checksum verification happens in the use case before the zip bundle reader is invoked.
- [x] Add adapter tests for timeout, max download size, no redirects, no credentials sent, non-`200` responses, network errors, malformed zip, unsafe zip entries, size limits, and successful zip decoding.

## Phase 5 - Use case orchestration

- [x] Add remote-template dependencies to the config-template generation use case without affecting direct built-in/custom generation paths.
- [x] Validate change id and config alias before loading config or accessing network.
- [x] Load `.specharbor/config.yml` only for `--config-template`.
- [x] Resolve the requested alias from parsed config.
- [x] For `source: builtin` and `source: custom`, delegate to existing behavior unchanged and do not call the remote fetch port.
- [x] For `source: remote`, validate URL, checksum, and format before network access.
- [x] Require `openspec/project.md` and `openspec/changes/` before remote network access.
- [x] Fetch downloaded bytes through the remote template fetch port.
- [x] Verify the checksum over downloaded bytes before archive parsing.
- [x] Decode the zip through the bundle reader port only after checksum verification succeeds.
- [x] Validate the decoded bundle provides exactly the five required non-empty files.
- [x] Write generated files only under `openspec/changes/<change-id>/` using the existing write-if-absent flow.
- [x] Skip existing files and never overwrite.
- [x] Return structured result data for config alias, resolved source `remote`, remote host, format, checksum algorithm, change path, created files, and skipped files.
- [x] Ensure invalid config, missing URL, invalid URL, non-HTTPS URL, missing checksum, unsupported checksum, network failure, timeout, size-limit error, checksum mismatch, malformed archive, unsafe archive path, symlink entry, missing files, and empty files produce clear errors.
- [x] Ensure all remote fetch, checksum, archive, and bundle validation failures produce zero writes, including no change-directory creation.
- [x] Ensure no production code, docs, config, CI, source-control, prompt, archive, or arbitrary output files are written by generation.

## Phase 6 - Use case tests

- [x] Generate from a config alias with `source: remote` and a valid pinned zip.
- [x] Verify the remote result includes change id, config alias, remote source kind, remote host, format, checksum algorithm, created files, skipped files, and change path.
- [x] Verify missing URL returns a clear error and writes nothing.
- [x] Verify invalid URL returns a clear error and writes nothing.
- [x] Verify non-HTTPS URL returns a clear error and performs no network call.
- [x] Verify URL credentials, query strings, and fragments return clear errors and perform no network call.
- [x] Verify missing checksum returns a clear error and performs no network call.
- [x] Verify unsupported checksum algorithm returns a clear error and performs no network call.
- [x] Verify checksum mismatch returns a clear error, does not parse the archive, and writes nothing.
- [x] Verify network failure returns a clear error and writes nothing.
- [x] Verify timeout returns a clear error and writes nothing.
- [x] Verify max download size exceeded returns a clear error and writes nothing.
- [x] Verify malformed zip returns a clear error and writes nothing.
- [x] Verify unsafe archive path returns a clear error and writes nothing.
- [x] Verify symlink archive entry rejection writes nothing.
- [x] Verify executable archive entry rejection writes nothing.
- [x] Verify duplicate archive entries and extra archive entries write nothing.
- [x] Verify missing required files returns a clear error and writes nothing.
- [x] Verify empty required files return a clear error and write nothing.
- [x] Verify generated files are written only under `openspec/changes/<change-id>/`.
- [x] Verify existing files are skipped and preserved.
- [x] Verify remote generation does not write production code, docs, config, CI, source-control, prompt, archive, or arbitrary output files.
- [x] Verify no network access occurs for direct built-in, direct custom, config `builtin`, or config `custom` aliases.
- [x] Verify missing OpenSpec project structure fails before network access.

## Phase 7 - CLI

- [x] Wire the remote fetch adapter and zip bundle reader adapter into `generateCommand` only for config-template generation use case construction.
- [x] Keep `--config-template <alias>` parsing and existing conflict rules unchanged unless this change explicitly documents a necessary error message addition.
- [x] Do not add `--remote-template`.
- [x] Print remote config-template success output with change id, config alias, resolved source `remote`, sanitized remote host, format, checksum algorithm, change path, directory status, created files, skipped files, and safety notes.
- [x] Never display credentials, query strings, fragments, auth headers, cookies, or environment-derived tokens in output.
- [x] Print clear errors for missing URL, invalid URL, non-HTTPS URL, URL credentials/query/fragment, missing checksum, unsupported checksum, missing format, unsupported format, network failure, timeout, max size exceeded, checksum mismatch, malformed archive, unsafe archive path, symlink entry, duplicate/extra entries, missing files, and empty files.
- [x] Preserve existing config-template builtin/custom success output and error output where possible.
- [x] Preserve existing CLI exit-code behavior for generation errors.

## Phase 8 - CLI, regression, and architecture tests

- [x] Test remote config-template success output.
- [x] Test remote output displays the sanitized host and checksum algorithm.
- [x] Test remote output lists created files and skipped existing files.
- [x] Test remote error output for URL, checksum, network, timeout, size, archive, and missing-file failures.
- [x] Test `--config-template` conflicts with `--blank`, `--template`, `--custom-template`, `--guided`, `--agent-assisted`, `--ai-assisted`, `--execute`, `--type`, `--agent`, `--from-file`, and `--overwrite` remain correct.
- [x] Regression test built-in template generation unchanged.
- [x] Regression test custom template generation unchanged.
- [x] Regression test config-template `builtin` generation unchanged and network-free.
- [x] Regression test config-template `custom` generation unchanged and network-free.
- [x] Regression test AI-assisted generation unchanged.
- [x] Regression test validate, prompt, review, archive, config, workflow, scan, init, help, and version unchanged.
- [x] Architecture tests confirm core does not import adapters.
- [x] Architecture tests confirm core does not import CLI packages.
- [x] Architecture tests confirm core does not import `os`, concrete HTTP client packages, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.
- [x] Architecture tests confirm no shell or script execution is introduced for remote templates.
- [x] Architecture tests confirm no arbitrary output path write path is introduced.

## Phase 9 - Documentation

- [x] Update `README.md` to list remote templates as implemented after code lands.
- [x] Update `README.md` with the remote alias schema, checksum requirement, HTTPS-only behavior, and safety boundary summary.
- [x] Update `docs/usage.md` with command usage through `--config-template`, config examples, supported zip format, error behavior, and validation guidance.
- [x] Update `docs/generation-modes.md` to move remote templates from planned to implemented after code lands.
- [x] Update any config/template-specific docs present at implementation time.
- [x] Document that `source: remote` requires `url`, `checksum`, and `format`.
- [x] Document that only `format: zip` is supported.
- [x] Document the required five root-level files in the zip.
- [x] Document that checksum verification happens before archive parsing.
- [x] Document that no persistent cache exists in this first version.
- [x] Document that no credentials, OAuth, auth headers, cookies, environment token expansion, git clone, marketplace search, provider APIs, script execution, shell execution, production code writes, source-control automation, auto-commit, PR, merge, or archive automation are introduced.
- [x] Verify documented command examples match implemented CLI behavior.

## Phase 10 - Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-remote-templates`.
- [x] Manually verify successful remote generation from a test HTTPS server or controlled test adapter serving a pinned zip.
- [x] Manually verify checksum mismatch writes nothing.
- [x] Manually verify non-HTTPS URLs fail before network access.
- [x] Manually verify redirects are rejected.
- [x] Manually verify oversized responses are rejected.
- [x] Manually verify unsafe zip paths and symlink entries are rejected.
- [x] Manually verify existing target files are skipped and preserved.
- [x] Run `git status --short`.
- [x] Inspect `git diff -- openspec/changes/implement-remote-templates/`.
- [x] Update this `tasks.md` only for implementation work actually completed.

## Test Engineer follow-up - Coverage gaps

- [x] Expanded CLI remote config-template error-output coverage for missing URL, invalid URL syntax, credential/query/fragment rejection, unsupported checksum algorithm, missing/unsupported format, malformed archive, unsafe archive path, symlink, executable, duplicate, extra, missing-file, and empty-file failures.
- [x] Completed config/domain source-specific alias field rejection coverage for remote-only fields on `builtin` and `custom`, required `template` on both local source kinds, required remote fields, remote `template` rejection, and unsupported path/auth/script/git/command-like fields.
- [x] Strengthened checksum mismatch tests to assert `sha256`, expected digest, and actual digest facts while preserving checksum verification in core and network/archive behavior in adapters.
