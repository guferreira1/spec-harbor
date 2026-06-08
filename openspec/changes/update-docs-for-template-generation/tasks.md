# Tasks: Update Docs For Template Generation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/update-docs-for-template-generation/`.
- [x] Inspect `git status --short` before editing and identify unrelated in-progress work that must not be modified.
- [x] Confirm this is a documentation-only change for the already implemented built-in template generation feature.
- [x] Confirm the future implementation diff may touch only `README.md`, Markdown files under `docs/`, and `openspec/changes/update-docs-for-template-generation/tasks.md`.
- [x] Confirm no Go code, Go tests, CLI behavior, `.github/workflows/ci.yml`, CI configuration, `.specharbor/config.yml`, init templates, or unrelated OpenSpec changes should be modified.
- [x] Inspect `implement-template-generation` outputs or current CLI behavior enough to verify the implemented command shape.
- [x] Confirm `go run ./cmd/specharbor generate <change-id> --blank` remains implemented and should stay documented.
- [x] Confirm `go run ./cmd/specharbor generate <change-id> --template <template-name>` is implemented for built-in templates.
- [x] Confirm the implemented built-in templates are exactly `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Do not document guided generation, AI-assisted generation, agent-assisted generation, custom templates, remote templates, config-driven templates, or interactive prompts as implemented.

## Phase 1: Documentation Inventory

- [x] Inventory `README.md` for command lists, implemented status sections, in-progress sections, and planned roadmap wording.
- [x] Inventory `docs/usage.md` for generation command examples and mode descriptions.
- [x] Inventory `docs/generation-modes.md` for implemented and planned generation mode language.
- [x] Check other Markdown files under `docs/` for references to template generation, generation modes, or planned roadmap items.
- [x] Identify every place that still says template generation is planned behavior.
- [x] Identify every duplicated generated-file structure that should mention `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Identify where to describe deterministic local generic starter content without overstating generated requirements.
- [x] Identify where to document that existing files are not overwritten.
- [x] Identify where to document that partially existing change directories are recoverable by creating only missing required files.
- [x] Keep planned roadmap wording clearly separated from implemented command behavior.

## Phase 2: README Updates

- [x] Update the README command list, if needed, to include `go run ./cmd/specharbor generate <change-id> --template <template-name>`.
- [x] Keep the README example for blank generation using `go run ./cmd/specharbor generate <change-id> --blank`.
- [x] Add or adjust a template generation example using a concrete built-in template name.
- [x] List or mention exactly the implemented built-in templates: `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Remove template generation from planned behavior if it is still listed as planned.
- [x] Keep future template capabilities, if mentioned, clearly labeled as planned and separate from implemented built-in template generation.
- [x] Keep README examples copy-pasteable from the repository root.
- [x] Keep the README concise and avoid duplicating all details from `docs/usage.md` and `docs/generation-modes.md`.

## Phase 3: Usage Documentation Updates

- [x] Update `docs/usage.md` to document `go run ./cmd/specharbor generate <change-id> --template <template-name>`.
- [x] Document `go run ./cmd/specharbor generate <change-id> --blank` as still implemented.
- [x] Document the four supported built-in templates exactly: `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Include copy-pasteable examples from the repository root for blank generation and template generation.
- [x] Explain that template generation creates the same required OpenSpec change files as blank generation.
- [x] List the required files as `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Explain that generated template content is deterministic, local, generic starter content.
- [x] Explain that generated template content is safe to edit and does not mean SpecHarbor inferred project-specific requirements.
- [x] Explain that existing files are skipped and not overwritten.
- [x] Explain that a partially existing change directory can be recovered by creating only missing required files.
- [x] Ensure usage docs do not claim guided, AI-assisted, agent-assisted, custom, remote, config-driven, or interactive generation is implemented.
- [x] Ensure usage docs do not introduce provider API key setup, agent setup, remote registry setup, or custom template setup as current behavior.

## Phase 4: Generation Modes Documentation Updates

- [x] Update `docs/generation-modes.md` so blank generation is listed as implemented.
- [x] Update `docs/generation-modes.md` so built-in template generation is listed as implemented.
- [x] Document the implemented template command shape with `go run ./cmd/specharbor generate <change-id> --template <template-name>`.
- [x] List exactly the implemented built-in templates: `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Explain that built-in template generation creates `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Explain that built-in template generation uses deterministic, local, generic starter content.
- [x] Explain the no-overwrite and partial-directory recovery behavior.
- [x] Remove template generation from the planned generation list, except for clearly labeled future template capabilities such as custom, remote, or config-driven templates.
- [x] Keep guided generation listed only as planned if it appears.
- [x] Keep AI-assisted generation listed only as planned if it appears.
- [x] Keep agent-assisted generation listed only as planned if it appears.
- [x] Keep hybrid generation listed only as planned if it appears.
- [x] Keep roadmap wording clearly separated from implemented behavior.

## Phase 5: Verification

- [x] Review changed documentation for stale claims that built-in template generation is planned.
- [x] Review changed documentation for inaccurate claims that custom templates, remote templates, guided generation, AI-assisted generation, agent-assisted generation, config-driven templates, or interactive prompts are implemented.
- [x] Review changed documentation to confirm it lists exactly `feature`, `bugfix`, `docs`, and `refactor` as built-in templates.
- [x] Review changed documentation to confirm `--blank` remains documented as implemented.
- [x] Review changed documentation to confirm `--template <template-name>` is documented as implemented.
- [x] Review changed documentation to confirm generated files are listed as `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Review changed documentation to confirm examples are copy-pasteable from the repository root using `go run ./cmd/specharbor ...`.
- [x] Verify the final diff modifies only `README.md`, Markdown files under `docs/`, and `openspec/changes/update-docs-for-template-generation/tasks.md`.
- [x] Verify no Go code was modified.
- [x] Verify no Go tests were modified.
- [x] Verify `.github/workflows/ci.yml` and CI configuration were not modified.
- [x] Verify `.specharbor/config.yml` was not modified.
- [x] Verify init templates were not modified.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate update-docs-for-template-generation`.
- [x] Inspect `git status --short`, `git diff --stat`, `git diff --name-only`, and `git diff`.

## Phase 6: Task Updates

- [x] Update this `tasks.md` by checking off only documentation implementation tasks actually completed.
- [x] Leave tasks unchecked for any work not performed.
- [x] Record any verification command that could not be run in the final implementation response.
- [x] Confirm the final implementation remains documentation-only.
- [x] Confirm no unrelated OpenSpec changes were modified.
