# Tasks: Update Docs For Agent-Assisted Spec Authoring

## Phase 0: Baseline and Scope

- [ ] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/update-docs-for-agent-assisted-spec-authoring/`.
- [ ] Inspect `git status --short` before editing and identify unrelated in-progress work that must not be modified.
- [ ] Confirm this is a documentation-only change for the already implemented dry-run agent-assisted spec authoring feature.
- [ ] Confirm the implementation diff may touch only `README.md`, Markdown files under `docs/`, and `openspec/changes/update-docs-for-agent-assisted-spec-authoring/tasks.md`.
- [ ] Confirm no Go code, Go tests, CLI behavior, `.github/workflows/ci.yml`, CI configuration, `.specharbor/config.yml`, init templates, or unrelated OpenSpec changes should be modified.
- [ ] Inspect `implement-agent-assisted-spec-authoring` outputs or current CLI behavior enough to verify the implemented agent-assisted command shape.
- [ ] Confirm `go run ./cmd/specharbor generate <change-id> --blank` remains implemented and should stay documented.
- [ ] Confirm `go run ./cmd/specharbor generate <change-id> --template <template-name>` remains implemented and should stay documented.
- [ ] Confirm `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"` remains implemented and should stay documented.
- [ ] Confirm `go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"` is implemented as dry-run agent-assisted spec authoring.
- [ ] Confirm the supported agent-assisted authoring types are exactly `feature`, `bugfix`, `docs`, and `refactor`.
- [ ] Confirm dry-run prints a deterministic authoring plan to stdout.
- [ ] Confirm dry-run prints a deterministic, copy-pasteable authoring prompt to stdout.
- [ ] Confirm dry-run writes no files.
- [ ] Confirm dry-run writes no prompt file.
- [ ] Confirm dry-run does not create or modify OpenSpec files.
- [ ] Confirm dry-run does not create or modify production code.
- [ ] Confirm dry-run does not execute agents or local agent commands.
- [ ] Confirm dry-run does not call provider APIs, local models, network APIs, source-control APIs, or workflow tools.
- [ ] Confirm `--execute` is currently unsupported and returns a clear error.
- [ ] Do not document AI-assisted generation, agent execution, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts as implemented.

## Phase 1: Documentation Inventory

- [ ] Inventory `README.md` for command lists, implemented status sections, in-progress sections, and planned roadmap wording.
- [ ] Inventory `docs/usage.md` for generation command examples and mode descriptions.
- [ ] Inventory `docs/generation-modes.md` for implemented and planned generation mode language.
- [ ] Check other Markdown files under `docs/` for references to agent-assisted generation, agent-assisted spec authoring, generation modes, dry-run behavior, or planned roadmap items.
- [ ] Identify every place that still says agent-assisted generation is planned behavior.
- [ ] Identify every place that should continue to describe blank generation as implemented.
- [ ] Identify every place that should continue to describe built-in template generation as implemented.
- [ ] Identify every place that should continue to describe guided generation as implemented.
- [ ] Identify where to document the agent-assisted command syntax with `--agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`.
- [ ] Identify where to list the supported agent-assisted authoring types exactly as `feature`, `bugfix`, `docs`, and `refactor`.
- [ ] Identify where to explain that agent-assisted spec authoring is dry-run only in this first version.
- [ ] Identify where to explain that dry-run prints a deterministic authoring plan and copy-pasteable prompt to stdout.
- [ ] Identify where to explain that dry-run writes no files and writes no prompt file.
- [ ] Identify where to explain that dry-run does not create or modify OpenSpec files.
- [ ] Identify where to explain that dry-run does not create or modify production code.
- [ ] Identify where to explain that dry-run does not execute agents.
- [ ] Identify where to explain that dry-run does not call provider APIs, local models, network APIs, source-control APIs, or workflow tools.
- [ ] Identify where to document that `--execute` is currently unsupported and returns a clear error.
- [ ] Identify where to explain that the generated prompt is meant to help an external agent author or refine only the OpenSpec change package.
- [ ] Identify where to explain that implementation remains a later step through the normal SpecHarbor workflow.
- [ ] Keep planned roadmap wording clearly separated from implemented command behavior.

## Phase 2: README Updates

- [ ] Update the README command list, if needed, to include `go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`.
- [ ] Keep the README example for blank generation using `go run ./cmd/specharbor generate <change-id> --blank`.
- [ ] Keep or adjust a built-in template generation example using `go run ./cmd/specharbor generate <change-id> --template <template-name>`.
- [ ] Keep or adjust a guided generation example using `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`.
- [ ] Add or adjust an agent-assisted example using a concrete agent name, supported type, title, and summary.
- [ ] List or mention exactly the supported agent-assisted authoring types: `feature`, `bugfix`, `docs`, and `refactor`.
- [ ] Describe agent-assisted spec authoring as dry-run only if the README status text needs that clarification.
- [ ] Describe dry-run stdout behavior if the README status text needs that clarification.
- [ ] Describe that dry-run does not write files, execute agents, or call provider/network/workflow APIs if the README status text needs that clarification.
- [ ] Remove generic agent-assisted generation from planned behavior if it is still listed as planned.
- [ ] Keep future AI-assisted generation, agent execution, hybrid generation, custom templates, remote templates, config-driven templates, and interactive prompts clearly labeled as planned or deferred when mentioned.
- [ ] Keep README examples copy-pasteable from the repository root.
- [ ] Keep the README concise and avoid duplicating all details from `docs/usage.md` and `docs/generation-modes.md`.

## Phase 3: Usage Documentation Updates

- [ ] Update `docs/usage.md` to document `go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`.
- [ ] Document `go run ./cmd/specharbor generate <change-id> --blank` as still implemented.
- [ ] Document `go run ./cmd/specharbor generate <change-id> --template <template-name>` as still implemented.
- [ ] Document `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"` as still implemented.
- [ ] Document the four supported agent-assisted authoring types exactly: `feature`, `bugfix`, `docs`, and `refactor`.
- [ ] Include copy-pasteable examples from the repository root for blank generation, built-in template generation, guided generation, and agent-assisted spec authoring.
- [ ] Explain that agent-assisted spec authoring is dry-run only in this first version.
- [ ] Explain that dry-run prints a deterministic authoring plan to stdout.
- [ ] Explain that dry-run prints a deterministic, copy-pasteable prompt to stdout.
- [ ] Explain that dry-run writes no files.
- [ ] Explain that dry-run writes no prompt file.
- [ ] Explain that dry-run does not create or modify OpenSpec files.
- [ ] Explain that dry-run does not create or modify production code.
- [ ] Explain that dry-run does not execute agents or local agent commands.
- [ ] Explain that dry-run does not call provider APIs.
- [ ] Explain that dry-run does not call local models.
- [ ] Explain that dry-run does not call network APIs.
- [ ] Explain that dry-run does not call source-control APIs.
- [ ] Explain that dry-run does not call workflow tools.
- [ ] Explain that `--execute` is currently unsupported and returns a clear error.
- [ ] Explain that the generated prompt is meant to help an external agent author or refine only the OpenSpec change package.
- [ ] Explain that implementation remains a later step through the normal SpecHarbor workflow.
- [ ] Ensure usage docs do not claim AI-assisted generation, agent execution, custom, remote, config-driven, hybrid, or interactive generation is implemented.
- [ ] Ensure usage docs do not introduce provider API key setup, agent setup, remote registry setup, custom template setup, config-backed template selection, interactive prompt flows, source-control automation, or workflow automation as current behavior.

## Phase 4: Generation Modes Documentation Updates

- [ ] Update `docs/generation-modes.md` so blank generation is listed as implemented.
- [ ] Update `docs/generation-modes.md` so built-in template generation is listed as implemented.
- [ ] Update `docs/generation-modes.md` so guided generation is listed as implemented.
- [ ] Update `docs/generation-modes.md` so dry-run agent-assisted spec authoring is listed as implemented.
- [ ] Document the agent-assisted command shape with `go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`.
- [ ] List exactly the supported agent-assisted authoring types: `feature`, `bugfix`, `docs`, and `refactor`.
- [ ] Explain that agent-assisted spec authoring is dry-run only in this first version.
- [ ] Explain that dry-run prints a deterministic authoring plan and copy-pasteable prompt to stdout.
- [ ] Explain that dry-run writes no files and writes no prompt file.
- [ ] Explain that dry-run does not create or modify OpenSpec files.
- [ ] Explain that dry-run does not create or modify production code.
- [ ] Explain that dry-run does not execute agents.
- [ ] Explain that dry-run does not call provider APIs, local models, network APIs, source-control APIs, or workflow tools.
- [ ] Explain that `--execute` is currently unsupported and returns a clear error.
- [ ] Explain that the generated prompt is meant to help an external agent author or refine only the OpenSpec change package.
- [ ] Explain that implementation remains a later step through the normal SpecHarbor workflow.
- [ ] Remove generic agent-assisted generation from the planned generation list.
- [ ] Keep AI-assisted generation listed only as planned if it appears.
- [ ] Keep future agent execution or non-dry-run agent-assisted behavior listed only as planned or deferred if it appears.
- [ ] Keep hybrid generation listed only as planned if it appears.
- [ ] Keep custom templates listed only as planned if they appear.
- [ ] Keep remote templates listed only as planned if they appear.
- [ ] Keep config-driven templates listed only as planned if they appear.
- [ ] Keep interactive prompts listed only as planned if they appear.
- [ ] Keep roadmap wording clearly separated from implemented behavior.

## Phase 5: Verification

- [ ] Review changed documentation for stale claims that agent-assisted generation is planned.
- [ ] Review changed documentation for inaccurate claims that AI-assisted generation, agent execution, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts are implemented.
- [ ] Review changed documentation to confirm it lists exactly `feature`, `bugfix`, `docs`, and `refactor` as supported agent-assisted authoring types.
- [ ] Review changed documentation to confirm `--blank` remains documented as implemented.
- [ ] Review changed documentation to confirm `--template <template-name>` remains documented as implemented.
- [ ] Review changed documentation to confirm `--guided --type <type> --title "<title>" --summary "<summary>"` remains documented as implemented.
- [ ] Review changed documentation to confirm `--agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"` is documented as implemented dry-run behavior.
- [ ] Review changed documentation to confirm agent-assisted spec authoring is described as dry-run only in this first version.
- [ ] Review changed documentation to confirm dry-run stdout authoring plan behavior is documented.
- [ ] Review changed documentation to confirm dry-run stdout prompt behavior is documented.
- [ ] Review changed documentation to confirm dry-run no-file-write behavior is documented.
- [ ] Review changed documentation to confirm dry-run no-prompt-file behavior is documented.
- [ ] Review changed documentation to confirm dry-run no-OpenSpec-file-create-or-modify behavior is documented.
- [ ] Review changed documentation to confirm dry-run no-production-code-create-or-modify behavior is documented.
- [ ] Review changed documentation to confirm dry-run no-agent-execution behavior is documented.
- [ ] Review changed documentation to confirm dry-run no-provider-API, no-local-model, no-network-API, no-source-control-API, and no-workflow-tool behavior is documented.
- [ ] Review changed documentation to confirm `--execute` is documented as currently unsupported with a clear error.
- [ ] Review changed documentation to confirm the generated prompt is described as helping an external agent author or refine only the OpenSpec change package.
- [ ] Review changed documentation to confirm implementation remains a later step through the normal SpecHarbor workflow.
- [ ] Review changed documentation to confirm examples are copy-pasteable from the repository root using `go run ./cmd/specharbor ...`.
- [ ] Verify the final diff modifies only `README.md`, Markdown files under `docs/`, and `openspec/changes/update-docs-for-agent-assisted-spec-authoring/tasks.md`.
- [ ] Verify no Go code was modified.
- [ ] Verify no Go tests were modified.
- [ ] Verify no CLI behavior was modified.
- [ ] Verify `.github/workflows/ci.yml` and CI configuration were not modified.
- [ ] Verify `.specharbor/config.yml` was not modified.
- [ ] Verify init templates were not modified.
- [ ] Verify unrelated OpenSpec changes were not modified.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/specharbor validate update-docs-for-agent-assisted-spec-authoring`.
- [ ] Inspect `git status --short`, `git diff --stat`, `git diff --name-only`, and `git diff`.

## Phase 6: Task Updates

- [ ] Update this `tasks.md` by checking off only documentation implementation tasks actually completed.
- [ ] Leave tasks unchecked for any work not performed.
- [ ] Record any verification command that could not be run in the final implementation response.
- [ ] Confirm the final implementation remains documentation-only.
- [ ] Confirm no unrelated OpenSpec changes were modified.
