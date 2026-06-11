# Risks: Implement AI-Assisted Generation

## Applying Untrusted AI Output

AI output is untrusted text. If SpecHarbor accepts arbitrary filenames, paths, patches, or shell-like content, the feature could overwrite unrelated files or hide production-code changes inside generated output.

Mitigation:

- Parse only the strict `---FILE: <filename>---` block format.
- Allow only the five required OpenSpec filenames.
- Reject unknown filenames, duplicate blocks, missing blocks, nested paths, absolute paths, traversal, and empty content before writes.
- Never execute content.
- Never parse or apply patches.
- Never interpret shell commands.
- Write only use-case-constructed target paths under `openspec/changes/<change-id>/`.

## Confusing AI-Assisted With Agent Runner Apply

Users may assume `--ai-assisted` applies live output from `--agent-assisted --execute`. That would weaken the runner foundation, which intentionally reports stdout and stderr without parsing or applying them.

Mitigation:

- Use a separate `--ai-assisted --from-file` command surface.
- Document that the source is a local user-controlled file.
- Keep `--agent-assisted --execute --apply` out of scope.
- Keep the existing runner behavior unchanged.
- Add regression tests proving runner output is still not parsed or applied.

## Accidental Overwrites

AI-generated Markdown may replace hand-authored OpenSpec files if overwrite behavior is too permissive.

Mitigation:

- Skip existing files by default.
- Require explicit `--overwrite` before replacing required files.
- Report skipped and overwritten files separately.
- Add CLI and use case tests for default skip behavior and explicit overwrite behavior.

## Partial Writes

Even after parsing and preflight, an unexpected write failure can occur after some files have been written. Automatic rollback could delete or corrupt user-authored files, especially when existing files are skipped or overwritten.

Mitigation:

- Run all parser and target preflight checks before writes.
- Keep the approved write set small and deterministic.
- Do not use temp files in the first version, avoiding additional filenames in the write surface.
- Report runtime write failures clearly.
- Do not automatically delete or roll back user files.
- Run validation only after planned writes complete successfully.

## Validation Failure After Writes

Generated files may be syntactically accepted by the parser but still fail OpenSpec validation. Rolling back on validation errors could surprise users and remove useful drafts.

Mitigation:

- Always run validation after successful writes.
- Report validation errors and warnings clearly.
- Keep validation warnings exit-zero.
- Make validation errors exit non-zero.
- Do not undo writes after validation findings.
- Document that validation errors leave generated files in place for review and editing.

## Format Friction

Strict delimiter blocks require the user or AI tool to produce an exact format. This can reject output that a human can understand.

Mitigation:

- Provide clear docs and examples.
- Return precise parse findings with codes, filenames, and line numbers.
- Reject ambiguous output before writes.
- Prefer safety over permissive parsing in the first version.

## Marker Collision

Generated Markdown could need the literal line `---END FILE---`, which would terminate a block.

Mitigation:

- Reserve the exact delimiter lines in the first version.
- Document the reserved marker behavior.
- Defer escaped markers or alternate encodings to a future format change if real use cases require them.

## Source File Path Misunderstanding

`--from-file` could be mistaken for a URL, provider name, or remote AI output source.

Mitigation:

- Treat `--from-file` as a local filesystem path only.
- Do not fetch URLs.
- Do not call provider APIs.
- Do not call remote AI services.
- Report missing local files clearly.
- Document local-only behavior.

## Architecture Leakage

Path validation, parser rules, write policy, or validation result semantics could leak into the CLI adapter or filesystem adapter, making the behavior harder to test and evolve.

Mitigation:

- Put parser and allowed filename policy in domain.
- Put orchestration and write decisions in the use case.
- Put filesystem effects behind small core-owned ports.
- Keep CLI code limited to parsing, wiring, formatting, and exit mapping.
- Add architecture tests for core import boundaries and absence of provider, workflow, source-control, and external-agent dependencies.

## Documentation Drift

Public docs currently list AI-assisted generation as planned. After implementation, stale docs could either hide the feature or overstate it as provider-backed generation.

Mitigation:

- Update README and docs in the implementation PR.
- Document the exact from-file command.
- Document the strict block format.
- Document default skip and explicit overwrite behavior.
- Document validation behavior.
- Document no provider APIs, remote AI services, production-code writes, source-control automation, PR creation, merge, or archive behavior.

## Regression in Existing Generation Modes

Adding new flags to `generate` can accidentally change existing blank, template, guided, or agent-assisted parsing and reporting.

Mitigation:

- Keep `--ai-assisted` parsing additive and mode-conflict checks explicit.
- Add regression tests for every existing generation mode.
- Preserve the existing `--agent-assisted --execute` run-and-report behavior.
- Preserve existing validation command behavior except for reuse by the new command.

## Overbuilding

Future versions may need provider-backed AI generation, live runner apply, confirmation flows, diff previews, rollback, JSON reports, or hybrid generation. Adding those now would expand risk and architecture surface before the safe foundation is proven.

Mitigation:

- Implement only local from-file AI-assisted generation.
- Avoid provider ports, workflow connectors, source-control ports, confirmation systems, patch application, and live runner apply.
- Keep the parser and write contracts small.
- Defer advanced workflows to separate OpenSpec changes.
