# Risks: Implement Template Generation

## Blank Generation Regression

Adding `--template` could accidentally change existing `--blank` parsing, output, project-structure checks, unsafe change id validation, or overwrite behavior.

Mitigation:

- Keep `--blank` behavior as the baseline contract.
- Add regression tests for blank generation success, idempotency, and argument errors.
- Keep template generation as an additive mode.
- Avoid broad CLI parser rewrites unless required.

## Architecture Leakage

Template generation touches CLI flags, template selection, filesystem writes, required-file policy, starter content, overwrite behavior, and reporting. A common failure mode is placing generation rules or template bodies directly in the CLI adapter.

Mitigation:

- Keep CLI responsibilities limited to parsing, dependency construction, current-working-directory lookup, and report formatting.
- Keep generation orchestration in `internal/core/usecase`.
- Keep generation concepts in `internal/core/domain`.
- Keep filesystem behavior behind ports in `internal/core/ports`.
- Keep concrete filesystem writes in `internal/adapters/filesystem`.
- Keep built-in template content in a concrete template/content adapter.
- Reuse the shared domain required OpenSpec change file policy.

## Overbuilding Template Support

SpecHarbor will eventually support more generation modes and possibly custom templates, but this change only needs deterministic built-in templates. Adding registries, marketplaces, runtime discovery, remote templates, provider integrations, or config-backed template selection now would create unused surface area.

Mitigation:

- Support only `feature`, `bugfix`, `docs`, and `refactor`.
- Use simple local deterministic content.
- Do not add custom template paths, filesystem discovery, remote registries, marketplace concepts, or config integration.
- Do not add provider, agent, source-control, workflow, or external execution abstractions.
- Add future extension points only where the current behavior requires a real port.

## Accidental Overwrite

Users may run template generation against a partially authored change directory. Overwriting any required file would destroy work.

Mitigation:

- Reuse the existing write-if-absent generation behavior.
- Never overwrite existing files.
- Treat existing change directories as recoverable.
- Create only missing required files.
- Report skipped existing files.
- Add tests proving existing content is preserved exactly.

## Unknown Template Writes

If unknown template validation happens too late, the command could create a directory before failing on template content lookup.

Mitigation:

- Validate template names before target directory or file writes.
- Make the template content port or domain model expose enough information for early validation.
- Add tests proving unknown templates are rejected before writes occur.

## CLI Parsing Ambiguity

Adding `--template <name>` introduces new argument combinations: missing names, both generation modes together, duplicate flags, unsupported flags, and extra arguments.

Mitigation:

- Define one clear generation mode per command invocation.
- Reject `--blank` and `--template` together.
- Reject missing template names explicitly.
- Reject unsupported flags and extra positional arguments explicitly.
- Add focused CLI tests for each error path.

## Unsafe Change IDs

The change id is used to build a path under `openspec/changes/`. Template generation must not weaken existing path-safety rules.

Mitigation:

- Reuse the existing blank generation change id validation behavior.
- Validate change ids before filesystem writes.
- Keep target paths constrained to `openspec/changes/<change-id>`.
- Add tests that unsafe ids are rejected before writes occur.

## Starter Content Too Prescriptive

Template content should help authors start from a useful structure without pretending that SpecHarbor inferred requirements or completed implementation work.

Mitigation:

- Use generic headings and starter prompts.
- Keep generated tasks unchecked.
- Avoid project-specific claims.
- Avoid generated claims that implementation, validation, or review has happened.
- Keep output deterministic and safe to commit.

## Documentation Drift

The docs currently describe template generation as planned. Updating public docs in this same change could make documentation claim behavior before implementation is merged and verified.

Mitigation:

- Do not modify README or docs outside this OpenSpec change.
- Track public documentation updates as a separate follow-up after implementation is merged.
- Include acceptance criteria that explicitly prohibit docs, README, and CI changes in this implementation change.

## Unrelated Command Regression

CLI command parsing changes can affect unrelated commands even when the feature is scoped to generation.

Mitigation:

- Keep command changes scoped to `generate`.
- Preserve existing command dispatch behavior.
- Add regression coverage for `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `help`, `version`, and unknown commands.
