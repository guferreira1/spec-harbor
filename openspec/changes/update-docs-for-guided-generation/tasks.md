# Tasks: Update Docs For Guided Generation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/update-docs-for-guided-generation/`.
- [x] Inspect `git status --short` before editing and identify unrelated in-progress work that must not be modified.
- [x] Confirm this is a documentation-only change for the already implemented guided generation feature.
- [x] Confirm the implementation diff may touch only `README.md`, Markdown files under `docs/`, and `openspec/changes/update-docs-for-guided-generation/tasks.md`.
- [x] Confirm no Go code, Go tests, CLI behavior, `.github/workflows/ci.yml`, CI configuration, `.specharbor/config.yml`, init templates, or unrelated OpenSpec changes should be modified.
- [x] Inspect `implement-guided-generation` outputs or current CLI behavior enough to verify the implemented guided command shape.
- [x] Confirm `go run ./cmd/specharbor generate <change-id> --blank` remains implemented and should stay documented.
- [x] Confirm `go run ./cmd/specharbor generate <change-id> --template <template-name>` remains implemented and should stay documented.
- [x] Confirm `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"` is implemented.
- [x] Confirm the supported guided types are exactly `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Confirm guided generation is deterministic, local, non-interactive, and driven by explicit CLI flags.
- [x] Confirm guided generation does not prompt during command execution.
- [x] Do not document AI-assisted generation, agent-assisted generation, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts as implemented.

## Phase 1: Documentation Inventory

- [x] Inventory `README.md` for command lists, implemented status sections, in-progress sections, and planned roadmap wording.
- [x] Inventory `docs/usage.md` for generation command examples and mode descriptions.
- [x] Inventory `docs/generation-modes.md` for implemented and planned generation mode language.
- [x] Check other Markdown files under `docs/` for references to guided generation, generation modes, or planned roadmap items.
- [x] Identify every place that still says guided generation is planned behavior.
- [x] Identify every place that should continue to describe blank generation as implemented.
- [x] Identify every place that should continue to describe built-in template generation as implemented.
- [x] Identify where to document the guided command syntax with `--guided --type <type> --title "<title>" --summary "<summary>"`.
- [x] Identify where to list the supported guided types exactly as `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Identify every duplicated generated-file structure that should mention `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Identify where to describe deterministic local non-interactive guided generation without implying an interactive wizard.
- [x] Identify where to explain that guided generation uses explicit CLI flags and does not prompt during command execution.
- [x] Identify where to document that guided generated content includes the supplied title and summary.
- [x] Identify where to document that existing files are not overwritten.
- [x] Identify where to document that partially existing change directories are recoverable by creating only missing required files.
- [x] Keep planned roadmap wording clearly separated from implemented command behavior.

## Phase 2: README Updates

- [x] Update the README command list, if needed, to include `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`.
- [x] Keep the README example for blank generation using `go run ./cmd/specharbor generate <change-id> --blank`.
- [x] Keep or adjust a built-in template generation example using `go run ./cmd/specharbor generate <change-id> --template <template-name>`.
- [x] Add or adjust a guided generation example using a concrete supported guided type, title, and summary.
- [x] List or mention exactly the supported guided types: `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Describe guided generation as deterministic, local, and non-interactive if the README status text needs that clarification.
- [x] Remove guided generation from planned behavior if it is still listed as planned.
- [x] Keep AI-assisted, agent-assisted, hybrid, custom template, remote template, config-driven template, and interactive prompt roadmap items clearly labeled as planned when mentioned.
- [x] Keep README examples copy-pasteable from the repository root.
- [x] Keep the README concise and avoid duplicating all details from `docs/usage.md` and `docs/generation-modes.md`.

## Phase 3: Usage Documentation Updates

- [x] Update `docs/usage.md` to document `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`.
- [x] Document `go run ./cmd/specharbor generate <change-id> --blank` as still implemented.
- [x] Document `go run ./cmd/specharbor generate <change-id> --template <template-name>` as still implemented.
- [x] Document the four supported guided types exactly: `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Include copy-pasteable examples from the repository root for blank generation, built-in template generation, and guided generation.
- [x] Explain that guided generation is deterministic, local, and non-interactive.
- [x] Explain that guided generation uses explicit CLI flags and does not prompt during command execution.
- [x] Explain that guided generated content includes the supplied title and summary.
- [x] Explain that guided generation creates the same required OpenSpec change files as blank and built-in template generation.
- [x] List the required files as `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Explain that existing files are skipped and not overwritten.
- [x] Explain that a partially existing change directory can be recovered by creating only missing required files.
- [x] Ensure usage docs do not claim AI-assisted, agent-assisted, custom, remote, config-driven, hybrid, or interactive generation is implemented.
- [x] Ensure usage docs do not introduce provider API key setup, agent setup, remote registry setup, custom template setup, config-backed template selection, or interactive prompt flows as current behavior.

## Phase 4: Generation Modes Documentation Updates

- [x] Update `docs/generation-modes.md` so blank generation is listed as implemented.
- [x] Update `docs/generation-modes.md` so built-in template generation is listed as implemented.
- [x] Update `docs/generation-modes.md` so guided generation is listed as implemented.
- [x] Document the guided command shape with `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`.
- [x] List exactly the supported guided types: `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Explain that guided generation creates `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Explain that guided generation uses deterministic, local starter content based on explicit CLI inputs.
- [x] Explain that guided generation is non-interactive and does not prompt during command execution.
- [x] Explain that guided generated content includes the supplied title and summary.
- [x] Explain the no-overwrite and partial-directory recovery behavior.
- [x] Remove guided generation from the planned generation list.
- [x] Keep AI-assisted generation listed only as planned if it appears.
- [x] Keep agent-assisted generation listed only as planned if it appears.
- [x] Keep hybrid generation listed only as planned if it appears.
- [x] Keep custom templates listed only as planned if they appear.
- [x] Keep remote templates listed only as planned if they appear.
- [x] Keep config-driven templates listed only as planned if they appear.
- [x] Keep interactive prompts listed only as planned if they appear.
- [x] Keep roadmap wording clearly separated from implemented behavior.

## Phase 5: Verification

- [x] Review changed documentation for stale claims that guided generation is planned.
- [x] Review changed documentation for inaccurate claims that AI-assisted generation, agent-assisted generation, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts are implemented.
- [x] Review changed documentation to confirm it lists exactly `feature`, `bugfix`, `docs`, and `refactor` as supported guided types.
- [x] Review changed documentation to confirm `--blank` remains documented as implemented.
- [x] Review changed documentation to confirm `--template <template-name>` remains documented as implemented.
- [x] Review changed documentation to confirm `--guided --type <type> --title "<title>" --summary "<summary>"` is documented as implemented.
- [x] Review changed documentation to confirm guided generation is described as deterministic, local, and non-interactive.
- [x] Review changed documentation to confirm guided generation is described as explicit-flag driven and does not prompt during command execution.
- [x] Review changed documentation to confirm guided generated content includes the supplied title and summary.
- [x] Review changed documentation to confirm generated files are listed as `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Review changed documentation to confirm existing files are not overwritten.
- [x] Review changed documentation to confirm partially existing change directories are recoverable by creating only missing required files.
- [x] Review changed documentation to confirm examples are copy-pasteable from the repository root using `go run ./cmd/specharbor ...`.
- [x] Verify the final diff modifies only `README.md`, Markdown files under `docs/`, and `openspec/changes/update-docs-for-guided-generation/tasks.md`.
- [x] Verify no Go code was modified.
- [x] Verify no Go tests were modified.
- [x] Verify no CLI behavior was modified.
- [x] Verify `.github/workflows/ci.yml` and CI configuration were not modified.
- [x] Verify `.specharbor/config.yml` was not modified.
- [x] Verify init templates were not modified.
- [x] Verify unrelated OpenSpec changes were not modified.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate update-docs-for-guided-generation`.
- [x] Inspect `git status --short`, `git diff --stat`, `git diff --name-only`, and `git diff`.

## Phase 6: Task Updates

- [x] Update this `tasks.md` by checking off only documentation implementation tasks actually completed.
- [x] Leave tasks unchecked for any work not performed.
- [x] Record any verification command that could not be run in the final implementation response.
- [x] Confirm the final implementation remains documentation-only.
- [x] Confirm no unrelated OpenSpec changes were modified.
