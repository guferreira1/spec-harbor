# Risks: Implement Guided Generation

## Blank and Template Regression

Adding `--guided`, `--type`, `--title`, and `--summary` could accidentally change existing `--blank` or `--template` parsing, output, project-structure checks, unsafe change id validation, idempotency, or overwrite behavior.

Mitigation:

- Keep blank and template behavior as baseline contracts.
- Add regression tests for blank generation success, idempotency, and argument errors.
- Add regression tests for template generation success, idempotency, supported templates, and argument errors.
- Keep guided generation as an additive mode.
- Avoid broad CLI parser rewrites unless required.

## Interactive Prompt Creep

The word "guided" can imply a terminal wizard. This change explicitly requires guidance through command-line flags only.

Mitigation:

- Require `--type`, `--title`, and `--summary`.
- Return clear errors for missing guided inputs.
- Do not read from stdin or terminal prompts.
- Do not add question flows, prompt sessions, or wizard state.
- Add tests or code review checks proving no interactive prompts are introduced.

## Architecture Leakage

Guided generation touches CLI flags, guided type selection, title and summary validation, starter content, filesystem writes, overwrite behavior, and reporting. A common failure mode is placing generation rules or guided content directly in the CLI adapter.

Mitigation:

- Keep CLI responsibilities limited to parsing, dependency construction, current-working-directory lookup, and report formatting.
- Keep generation orchestration in `internal/core/usecase`.
- Keep generation concepts in `internal/core/domain`.
- Keep filesystem behavior behind ports in `internal/core/ports`.
- Keep concrete filesystem writes in `internal/adapters/filesystem`.
- Provide guided content through a core-owned port.
- Keep concrete guided content in an adapter.
- Reuse the shared domain required OpenSpec change file policy.

## Overbuilding Guided Support

SpecHarbor may eventually support interactive, AI-assisted, agent-assisted, hybrid, custom template, remote, or config-driven generation. Adding those abstractions now would increase surface area without implementing the requested behavior.

Mitigation:

- Support only deterministic local guided generation through explicit flags.
- Support only `feature`, `bugfix`, `docs`, and `refactor`.
- Do not add interactive wizard abstractions.
- Do not add provider, agent, source-control, workflow, remote template, marketplace, custom template, or external execution abstractions.
- Add future extension points only where current guided behavior requires a real port.

## Accidental Overwrite

Users may run guided generation against an existing or partially authored change directory. Overwriting any required file would destroy work.

Mitigation:

- Reuse existing write-if-absent generation behavior.
- Never overwrite existing files.
- Treat existing change directories as recoverable.
- Create only missing required files.
- Report skipped existing files.
- Add tests proving existing content is preserved exactly.

## Unknown Guided Type Writes

If guided type validation happens too late, the command could create a directory before failing on guided content lookup.

Mitigation:

- Validate guided type before target directory or file writes.
- Make the guided content port or domain model expose enough information for early validation.
- Add tests proving unknown guided types are rejected before writes occur.

## Missing Input Ambiguity

Guided mode requires multiple value flags. Ambiguous parsing could produce unclear errors when values are missing, flags appear out of order, flags are duplicated, or another flag appears where a value is expected.

Mitigation:

- Define explicit parser errors for missing `--type`, `--title`, and `--summary`.
- Reject unsupported flags and extra positional arguments.
- Reject `--guided` combined with `--blank` or `--template`.
- Reject duplicate conflicting flags where current parser patterns support it.
- Add focused CLI tests for each error path.

## Unsafe Change IDs

The change id is used to build a path under `openspec/changes/`. Guided generation must not weaken existing path-safety rules.

Mitigation:

- Reuse existing blank and template change id validation behavior.
- Validate change ids before filesystem writes.
- Keep target paths constrained to `openspec/changes/<change-id>`.
- Add tests that unsafe ids are rejected before writes occur.

## Starter Content Too Prescriptive

Guided content should help authors start from concrete title and summary context without pretending SpecHarbor inferred requirements, asked follow-up questions, or completed implementation work.

Mitigation:

- Use the supplied title and summary as explicit context.
- Keep content generic and editable.
- Keep generated tasks unchecked.
- Avoid project-specific claims.
- Avoid generated claims that implementation, validation, review, documentation updates, or archiving has happened.
- Keep output deterministic and safe to commit.

## Documentation Drift

Public docs currently describe guided generation as planned. Updating docs in this implementation change could claim behavior before the code is merged and verified.

Mitigation:

- Do not modify README or docs outside this OpenSpec change.
- Track public documentation updates as a separate follow-up after guided generation is implemented and verified.
- Include acceptance criteria that explicitly prohibit docs, README, CI, and config changes in this implementation change.

## Unrelated Command Regression

CLI command parsing changes can affect unrelated commands even when the feature is scoped to generation.

Mitigation:

- Keep command changes scoped to `generate`.
- Preserve existing command dispatch behavior.
- Add regression coverage for `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `help`, `version`, and unknown commands.
