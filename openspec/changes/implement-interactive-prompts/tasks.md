# Tasks: Implement Interactive Prompts

## Phase 1 - Command Parsing

- [x] Add `--interactive` parsing to `specharbor generate <change-id>`.
- [x] Reject duplicate `--interactive` flags with a clear error.
- [x] Require exactly one positional change id when `--interactive` is present.
- [x] Reject `specharbor generate --interactive` without a change id.
- [x] Reject extra positional arguments with `--interactive`.
- [x] Reject `--interactive` combined with `--blank`.
- [x] Reject `--interactive` combined with `--template`.
- [x] Reject `--interactive` combined with `--custom-template`.
- [x] Reject `--interactive` combined with `--config-template`.
- [x] Reject `--interactive` combined with `--guided`.
- [x] Reject `--interactive` combined with `--hybrid`.
- [x] Reject `--interactive` combined with `--ai-assisted`.
- [x] Reject `--interactive` combined with `--agent-assisted`.
- [x] Reject `--interactive` combined with `--type`.
- [x] Reject `--interactive` combined with `--title`.
- [x] Reject `--interactive` combined with `--summary`.
- [x] Reject `--interactive` combined with `--from-file`.
- [x] Reject `--interactive` combined with `--agent`.
- [x] Reject `--interactive` combined with `--execute`.
- [x] Reject `--interactive` combined with `--overwrite`.
- [x] Preserve all existing non-interactive parse behavior.

## Phase 2 - CLI Terminal Abstraction

- [x] Add an adapter-owned terminal abstraction for interactive CLI input, output, and TTY detection.
- [x] Keep the abstraction local to `internal/adapters/cli` unless implementation proves a narrower package is needed.
- [x] Ensure production interactive execution checks real stdin TTY state.
- [x] Ensure tests can use fake terminal input and captured output.
- [x] Return `interactive mode requires a TTY` before prompting when stdin is not a terminal.
- [x] Ensure non-TTY interactive execution exits non-zero and writes nothing.
- [x] Ensure EOF before confirmation is treated as cancellation and writes nothing.
- [x] Keep terminal IO out of `internal/core/domain`.
- [x] Keep terminal IO out of `internal/core/ports`.
- [x] Keep terminal IO out of `internal/core/usecase`.

## Phase 3 - Prompt Flow

- [x] Implement the common generation path menu with stable ordering: blank, built-in template, custom template, config template, hybrid.
- [x] Accept menu choices by number.
- [x] Accept menu choices by normalized path name.
- [x] Reject unsupported menu answers and retry up to three attempts.
- [x] Fail clearly after menu retry exhaustion and write nothing.
- [x] Implement the blank prompt flow with no extra questions.
- [x] Implement the built-in template prompt flow with stable choices: feature, bugfix, docs, refactor.
- [x] Implement the custom template prompt flow with custom template name, optional title, and optional summary.
- [x] Implement the config template prompt flow with config alias, optional title, and optional summary.
- [x] Implement the hybrid prompt flow with source namespace, source value, required title, required summary, and optional type.
- [x] For hybrid source namespace, support built-in template, custom template, and config template.
- [x] Do not offer direct guided generation in the first interactive menu.
- [x] Do not offer AI-assisted generation in the first interactive menu.
- [x] Do not offer agent-assisted generation in the first interactive menu.
- [x] Do not ask questions that are irrelevant to the selected path.

## Phase 4 - Input Validation

- [x] Validate the interactive change id using the existing `ChangeID` value object before prompting.
- [x] Validate built-in template answers using the existing `TemplateName` behavior.
- [x] Validate custom template answers using the existing `CustomTemplateName` behavior.
- [x] Validate config alias answers using the existing `ConfigTemplateAlias` behavior.
- [x] Validate hybrid source selection using existing hybrid source selection rules.
- [x] Validate hybrid title as required and non-empty after trimming.
- [x] Validate hybrid summary as required and non-empty after trimming.
- [x] Validate non-empty hybrid type answers using existing hybrid type validation.
- [x] Treat empty optional custom-template title as omitted.
- [x] Treat empty optional custom-template summary as omitted.
- [x] Treat empty optional config-template title as omitted.
- [x] Treat empty optional config-template summary as omitted.
- [x] Treat empty optional hybrid type as omitted.
- [x] Retry required invalid answers up to three attempts.
- [x] Retry invalid non-empty optional type answers up to three attempts.
- [x] Fail clearly after retry exhaustion and write nothing.
- [x] Ensure source existence, config alias lookup, remote fetch, remote checksum, archive validation, write policy, and validation behavior remain delegated to existing generation use cases.

## Phase 5 - Confirmation And Cancellation

- [x] Print a deterministic interactive generation summary before generation.
- [x] Include change id in the summary.
- [x] Include selected generation path in the summary.
- [x] Include selected template in the summary when applicable.
- [x] Include selected custom template in the summary when applicable.
- [x] Include selected config alias in the summary when applicable.
- [x] Include selected hybrid source in the summary when applicable.
- [x] Include title, summary, and type only when applicable.
- [x] Include target write directory `openspec/changes/<change-id>/`.
- [x] Include the five approved OpenSpec files in the summary.
- [x] Include `Validation: automatic no` in the blank path summary.
- [x] Include `Validation: automatic no` in the built-in template path summary.
- [x] Include `Validation: automatic no` in the custom template path summary.
- [x] Include `Validation: automatic no` in the config template path summary.
- [x] Include `Validation: automatic yes` in the hybrid path summary.
- [x] Include the deterministic safety section in the summary before the confirmation prompt.
- [x] Ask for confirmation before invoking any generation use case.
- [x] Treat `y` and `yes` as confirmation.
- [x] Treat `Y`, `YES`, and `Yes` as confirmation.
- [x] Treat any casing of trimmed `y` or `yes` as confirmation.
- [x] Treat `n` and `no` as cancellation.
- [x] Treat `N`, `NO`, and `No` as cancellation.
- [x] Treat any casing of trimmed `n` or `no` as cancellation.
- [x] Treat empty confirmation input as cancellation.
- [x] Treat EOF as cancellation.
- [x] Retry unsupported confirmation answers up to three attempts.
- [x] Fail clearly after three unsupported confirmation answers.
- [x] Return non-zero on cancellation with `operation cancelled`.
- [x] Ensure cancellation writes nothing.
- [x] Ensure retry exhaustion at confirmation writes nothing.
- [x] Ensure no generation use case is called before confirmation.

## Phase 6 - Request Mapping And Delegation

- [x] Map the blank prompt flow to the existing blank generation request.
- [x] Map the built-in template prompt flow to the existing built-in template generation request.
- [x] Map the custom template prompt flow to the existing custom template generation request with optional title and summary.
- [x] Map the config template prompt flow to the existing config template generation request with optional title and summary.
- [x] Map the hybrid built-in prompt flow to the existing hybrid built-in generation request.
- [x] Map the hybrid custom prompt flow to the existing hybrid custom generation request.
- [x] Map the hybrid config prompt flow to the existing hybrid config generation request.
- [x] Reuse existing generation report output as much as possible after confirmed generation.
- [x] Preserve hybrid validation report and exit-code behavior.
- [x] Preserve direct blank, built-in template, custom template, and config template validation behavior.
- [x] Do not add a post-generation validation prompt in the first version.
- [x] Do not add source resolution, remote safety, write-policy, or validation business rules to CLI prompt code.

## Phase 7 - Domain Tests

- [x] Add or update tests proving existing `ChangeID` validation is reused for interactive change ids.
- [x] Add or update tests proving existing `TemplateName` validation is reused for interactive built-in template answers.
- [x] Add or update tests proving existing `CustomTemplateName` validation is reused for interactive custom template answers.
- [x] Add or update tests proving existing `ConfigTemplateAlias` validation is reused for interactive config alias answers.
- [x] Add or update tests proving existing hybrid type/title/summary validation is reused where applicable.
- [x] Add tests for any interactive-specific result or cancellation model if implementation introduces one.
- [x] Do not add terminal logic to domain tests because no terminal logic belongs in domain.

## Phase 8 - CLI Tests

- [x] Test `--interactive` parsing.
- [x] Test duplicate `--interactive` flag behavior.
- [x] Test `--interactive` requires a change id.
- [x] Test `--interactive` rejects extra positional arguments.
- [x] Test conflicts between `--interactive` and direct generation mode flags.
- [x] Test conflicts between `--interactive` and direct input flags.
- [x] Test non-TTY failure behavior.
- [x] Test prompt sequence for blank generation.
- [x] Test prompt sequence for built-in template generation.
- [x] Test prompt sequence for custom template generation.
- [x] Test prompt sequence for config template generation.
- [x] Test prompt sequence for hybrid generation.
- [x] Test invalid menu input retry behavior.
- [x] Test invalid value retry behavior.
- [x] Test retry exhaustion behavior.
- [x] Test summary includes `Validation: automatic no` for blank generation.
- [x] Test summary includes `Validation: automatic no` for built-in template generation.
- [x] Test summary includes `Validation: automatic no` for custom template generation.
- [x] Test summary includes `Validation: automatic no` for config template generation.
- [x] Test summary includes `Validation: automatic yes` for hybrid generation.
- [x] Test summary includes the deterministic safety section.
- [x] Test cancellation with `n`.
- [x] Test cancellation with `N`.
- [x] Test cancellation with `no`.
- [x] Test cancellation with `NO`.
- [x] Test cancellation with `No`.
- [x] Test cancellation with empty confirmation input.
- [x] Test EOF confirmation cancellation.
- [x] Test confirmation with `y`.
- [x] Test confirmation with `Y`.
- [x] Test confirmation with `yes`.
- [x] Test confirmation with `YES`.
- [x] Test confirmation with `Yes`.
- [x] Test invalid confirmation retry and exhaustion behavior.
- [x] Test successful blank generation flow.
- [x] Test successful built-in template flow.
- [x] Test successful custom template flow.
- [x] Test successful config template flow.
- [x] Test successful hybrid built-in flow.
- [x] Test successful hybrid custom flow.
- [x] Test successful hybrid config flow.
- [x] Test unsupported mode handling.
- [x] Test output summary.
- [x] Test error output.
- [x] Test no writes before confirmation.
- [x] Test no writes on cancellation.
- [x] Test no writes on confirmation retry exhaustion.
- [x] Test no writes on invalid input exhaustion.

## Phase 9 - Use Case And Integration Tests

- [x] Test selected prompt answers map to the same request as `generate <change-id> --blank`.
- [x] Test selected prompt answers map to the same request as `generate <change-id> --template <name>`.
- [x] Test selected prompt answers map to the same request as `generate <change-id> --custom-template <name>`.
- [x] Test selected prompt answers map to the same request as `generate <change-id> --config-template <alias>`.
- [x] Test selected prompt answers map to the same request as hybrid built-in generation.
- [x] Test selected prompt answers map to the same request as hybrid custom generation.
- [x] Test selected prompt answers map to the same request as hybrid config generation.
- [x] Test existing files are preserved according to the selected mode.
- [x] Test validation behavior follows the selected mode.
- [x] Test interactive blank generation writes only OpenSpec change files.
- [x] Test interactive built-in template generation writes only OpenSpec change files.
- [x] Test interactive custom template generation writes only OpenSpec change files.
- [x] Test interactive config template generation writes only OpenSpec change files.
- [x] Test interactive hybrid generation writes only OpenSpec change files.
- [x] Test no production files are written.
- [x] Test no docs files are written.
- [x] Test no config files are written.
- [x] Test no CI files are written.
- [x] Test no source-control files are written.
- [x] Test no prompt files are written.
- [x] Test no archive files are written.
- [x] Test no arbitrary files are written.
- [x] Test remote template behavior is reachable only through an existing config alias after confirmation.

## Phase 10 - Regression Tests

- [x] Confirm blank generation is unchanged.
- [x] Confirm built-in template generation is unchanged.
- [x] Confirm custom template generation is unchanged.
- [x] Confirm config-template builtin generation is unchanged.
- [x] Confirm config-template custom generation is unchanged.
- [x] Confirm config-template remote generation is unchanged.
- [x] Confirm hybrid generation is unchanged except for being callable after interactive confirmation.
- [x] Confirm guided generation is unchanged.
- [x] Confirm AI-assisted generation is unchanged.
- [x] Confirm agent-assisted dry-run generation is unchanged.
- [x] Confirm agent-assisted run-and-report generation is unchanged.
- [x] Confirm validate behavior is unchanged.
- [x] Confirm prompt behavior is unchanged.
- [x] Confirm review behavior is unchanged.
- [x] Confirm archive behavior is unchanged.
- [x] Confirm config behavior is unchanged.
- [x] Confirm workflow behavior is unchanged.
- [x] Confirm scan behavior is unchanged.
- [x] Confirm init behavior is unchanged.
- [x] Confirm help behavior is unchanged.
- [x] Confirm version behavior is unchanged.

## Phase 11 - Architecture And Safety Tests

- [x] Add or update architecture tests proving core does not import terminal packages.
- [x] Add or update architecture tests proving core does not import adapters.
- [x] Add or update architecture tests proving core does not import CLI packages.
- [x] Add or update architecture tests proving core does not import `os`.
- [x] Add or update architecture tests proving core does not import `os/exec`.
- [x] Add or update architecture tests proving core does not read stdin.
- [x] Add or update architecture tests proving core does not write stdout or stderr.
- [x] Add or update architecture tests proving no provider SDKs are introduced.
- [x] Add or update architecture tests proving no network SDKs are introduced for interactive behavior.
- [x] Add or update architecture tests proving no source-control SDKs are introduced.
- [x] Add or update architecture tests proving no workflow SDKs are introduced.
- [x] Add or update architecture tests proving no external-agent SDKs are introduced.
- [x] Add or update architecture tests proving no shell or script execution is introduced.
- [x] Add or update architecture tests proving no arbitrary output path write abstraction is introduced.
- [x] Add or update tests proving no production-code write path is introduced by interactive generation.
- [x] Add or update tests proving no provider, LLM, local model, source-control, workflow, or live agent output application behavior is introduced.
- [x] Add or update tests proving no auto-commit, auto-push, pull request creation, merge, or archive automation is introduced.
- [x] Do not introduce external TUI frameworks unless a later design explicitly justifies them.

## Phase 12 - Documentation

- [x] Update `README.md` to list interactive generation after implementation.
- [x] Update `README.md` with command syntax and safety boundaries.
- [x] Update `docs/usage.md` with interactive command syntax.
- [x] Update `docs/usage.md` with prompt sequence examples.
- [x] Ensure transcript-style examples include the summary validation line and safety section.
- [x] Update `docs/usage.md` to explain that interactive mode prints a deterministic summary before writing.
- [x] Update `docs/usage.md` to document the summary validation line.
- [x] Update `docs/usage.md` to document the deterministic summary safety notes.
- [x] Update `docs/usage.md` with confirmation and cancellation behavior.
- [x] Update `docs/usage.md` with non-TTY behavior.
- [x] Update `docs/usage.md` with validation and write behavior.
- [x] Update `docs/generation-modes.md` to explain interactive generation as a prompt layer over existing modes.
- [x] Update `docs/generation-modes.md` with supported interactive paths.
- [x] Update any CLI or interactive docs present at implementation time.
- [x] Document that `<change-id>` remains required.
- [x] Document that the summary includes whether validation is automatic.
- [x] Document that blank, built-in template, custom template, and config template paths show `Validation: automatic no`.
- [x] Document that hybrid paths show `Validation: automatic yes`.
- [x] Document that the summary includes safety notes before confirmation.
- [x] Document that confirmation is case-insensitive.
- [x] Document that `y` and `yes` proceed.
- [x] Document that `n` and `no` cancel.
- [x] Document that empty confirmation and EOF cancel.
- [x] Document that cancellation writes nothing and exits non-zero.
- [x] Document that direct guided generation remains non-interactive in the first version.
- [x] Document that AI-assisted and agent-assisted modes are not offered by interactive prompts.
- [x] Document that remote templates are reachable only through existing config aliases.
- [x] Document that interactive mode does not call provider APIs, LLM APIs, local model APIs, agents, source-control tools, workflow tools, shell commands, or scripts.
- [x] Document that interactive mode writes no production code and performs no auto-commit, auto-push, pull request creation, merge, or archive automation.
- [x] Include examples for blank, built-in template, custom template, config template, and hybrid interactive flows.

## Phase 13 - Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-interactive-prompts`.
- [x] Manually verify interactive blank generation in a TTY.
- [x] Manually verify interactive built-in template generation in a TTY.
- [x] Manually verify interactive custom template generation in a TTY.
- [x] Manually verify interactive config template generation in a TTY.
- [x] Manually verify interactive hybrid generation in a TTY.
- [x] Manually verify non-TTY interactive execution fails without hanging.
- [x] Manually verify cancellation writes nothing.
- [x] Manually verify invalid input retry exhaustion writes nothing.
- [x] Run `git status --short`.
- [x] Inspect `git diff -- openspec/changes/implement-interactive-prompts/`.
- [x] Update this `tasks.md` only for implementation work actually completed.

## Phase 14 - Test Engineer Defect Fixes

- [x] Preserve direct hybrid type validation behavior in interactive hybrid prompts.
- [x] Reject mixed-case and uppercase interactive hybrid type answers.
- [x] Keep lowercase supported interactive hybrid types accepted.
- [x] Move production TTY detection behind CLI-owned portable build-tagged implementations.
- [x] Add focused tests for hybrid title retry and retry exhaustion.
- [x] Add focused tests for hybrid summary retry and retry exhaustion.
- [x] Add focused tests for whitespace-trimmed confirmation proceed variants.
- [x] Add focused tests for whitespace-trimmed confirmation cancellation variants.
