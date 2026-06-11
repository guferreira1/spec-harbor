# Proposal: Implement Guided Generation

## Summary

Add deterministic guided OpenSpec change generation:

```text
specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

The first supported guided types are:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Guided generation creates the same standard OpenSpec change files as blank and built-in template generation, but enriches the starter content with explicit command-line inputs supplied by the user. It must remain non-interactive, local, deterministic, safe to commit, and independent of AI providers, network access, source-control integrations, workflow integrations, external execution, and provider credentials.

## Problem

SpecHarbor currently supports:

```text
specharbor generate <change-id> --blank
specharbor generate <change-id> --template <template-name>
```

Blank generation creates a structure for manual authoring. Built-in template generation creates generic starter content for common change types. Both are useful, but neither captures the user's concrete title and summary at generation time.

The documentation still lists guided generation as planned behavior. Without a deterministic guided mode, users who already know the change type, title, and summary must generate a blank or generic template and then manually add the same basic context to each change package.

## Goal

Implement a first guided generation mode that accepts all required guidance through CLI flags:

```text
specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --guided --type bugfix --title "<title>" --summary "<summary>"
specharbor generate <change-id> --guided --type docs --title "<title>" --summary "<summary>"
specharbor generate <change-id> --guided --type refactor --title "<title>" --summary "<summary>"
```

For each command, SpecHarbor must:

- create `openspec/changes/<change-id>/` when missing;
- create the required OpenSpec change files when missing;
- populate missing files with deterministic starter content based on guided type, title, and summary;
- include the supplied title and summary in the generated guided content;
- skip existing files without overwriting content;
- preserve existing partial-directory recovery behavior;
- preserve existing unsafe change id validation;
- preserve existing `--blank` behavior;
- preserve existing `--template <template-name>` behavior;
- report created and skipped files in clear human-readable CLI output.

## Scope

- Add `--guided` generation mode to `specharbor generate`.
- Require `--type <type>`, `--title "<title>"`, and `--summary "<summary>"` when `--guided` is used.
- Support exactly these guided types: `feature`, `bugfix`, `docs`, and `refactor`.
- Create the same required OpenSpec change files:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- Reuse existing required-file, write-if-absent, idempotency, partial-directory recovery, and skip-existing-file reporting behavior.
- Reuse built-in template content strategy where appropriate, while enriching guided output with the supplied title and summary.
- Return clear errors for missing guided arguments, unknown guided type, conflicting generation modes, unsupported flags, extra positional arguments, and duplicate conflicting flags where the current parser supports that validation.
- Keep CLI parsing and report formatting in `internal/adapters/cli`.
- Keep generation orchestration in `internal/core/usecase`.
- Keep generation domain concepts in `internal/core/domain`.
- Keep filesystem behavior behind generation ports in `internal/core/ports`.
- Provide guided content through a core-owned port/interface.
- Implement concrete deterministic guided content in `internal/adapters/templates` or another adapter package if the implementation justifies it.
- Add focused tests for domain concepts, ports, guided content, use case orchestration, CLI parsing/reporting, idempotency, error cases, unsafe change ids, and regressions for blank and template generation.

## Out of Scope

- Interactive terminal wizard.
- Runtime prompts that ask questions during command execution.
- AI-assisted generation.
- Agent-assisted generation.
- Hybrid generation.
- Remote templates.
- Custom templates.
- Config-driven templates.
- Template marketplace.
- Provider API setup.
- Provider API keys.
- Network access.
- Source-control integration.
- Workflow integration.
- Updating README or docs outside this OpenSpec change.
- Changing `init`.
- Changing `scan`.
- Changing `validate`.
- Changing `prompt`.
- Changing `review`.
- Changing `archive`.
- Changing `config`.
- Changing CI.
- Modifying `.github/workflows/ci.yml`.
- Modifying `.specharbor/config.yml`.

## Success Criteria

- `specharbor generate <change-id> --blank` continues to behave as it does today.
- `specharbor generate <change-id> --template feature` and the other existing built-in templates continue to behave as they do today.
- `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"` creates or completes a change package with feature-oriented guided content.
- `specharbor generate <change-id> --guided --type bugfix --title "<title>" --summary "<summary>"` creates or completes a change package with bugfix-oriented guided content.
- `specharbor generate <change-id> --guided --type docs --title "<title>" --summary "<summary>"` creates or completes a change package with documentation-oriented guided content.
- `specharbor generate <change-id> --guided --type refactor --title "<title>" --summary "<summary>"` creates or completes a change package with refactor-oriented guided content.
- Guided output includes the supplied title and summary.
- Existing files are skipped and never overwritten.
- Partially existing change directories are completed by creating only missing required files.
- Missing guided arguments, unknown guided types, conflicting generation modes, unsupported flags, extra positional arguments, duplicate conflicting flags, and unsafe change ids return clear errors.
- Guided generation does not prompt interactively and does not call AI providers, network APIs, source-control APIs, workflow connectors, external processes, provider SDKs, or config-backed template systems.
- No unrelated README, docs, CI, config, or command behavior changes are included.
- `go test ./...` succeeds after implementation.
