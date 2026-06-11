# Design: Implement Interactive Prompts

## Overview

Interactive generation is a CLI-owned prompt layer in front of existing generation use cases.

The flow is:

1. Parse `specharbor generate <change-id> --interactive`.
2. Reject unsupported flag combinations.
3. Validate `<change-id>` with the existing change id rules.
4. Verify stdin is an interactive terminal.
5. Prompt for a supported generation path and required path-specific values.
6. Validate each answer with existing value objects or generation validators.
7. Print a deterministic summary with validation behavior and safety notes.
8. Ask for confirmation.
9. On confirmation, build the same request shape used by non-interactive generation.
10. Delegate to the existing generation path and print the existing generation report as much as possible.

Interactive mode is not a new authoring engine. It is a guided CLI input mechanism for existing deterministic generation behavior.

## Interactive Mode Definition

Interactive generation means: terminal prompts collect generation options that a user would otherwise pass as flags.

Supported first-version paths:

- blank generation;
- built-in template generation;
- project-local custom template generation;
- config template generation;
- hybrid generation with a built-in, custom, or config source.

Not supported in the first interactive version:

- direct guided generation;
- AI-assisted from-file import;
- agent-assisted authoring;
- agent-assisted execution;
- remote template URL entry;
- source-control or workflow automation.

Direct `--guided` remains available through existing non-interactive flags:

```bash
specharbor generate <change-id> --guided --type feature --title "Title" --summary "Summary"
```

The first interactive version does not include it in the menu because the interactive feature already asks guided questions for template-backed and blank workflows, and adding a direct guided branch would broaden the scope without changing the safety foundation.

## Command Contract

Supported:

```bash
specharbor generate <change-id> --interactive
```

Rejected:

```bash
specharbor generate --interactive
specharbor generate <change-id> --interactive --blank
specharbor generate <change-id> --interactive --template feature
specharbor generate <change-id> --interactive --custom-template api-feature
specharbor generate <change-id> --interactive --config-template api-feature
specharbor generate <change-id> --interactive --guided
specharbor generate <change-id> --interactive --hybrid
specharbor generate <change-id> --interactive --ai-assisted --from-file output.txt
specharbor generate <change-id> --interactive --agent-assisted --agent codex
specharbor generate <change-id> --interactive --title "Title"
specharbor generate <change-id> --interactive --summary "Summary"
specharbor generate <change-id> --interactive --type feature
specharbor generate <change-id> --interactive --execute
specharbor generate <change-id> --interactive --overwrite
```

Flag decisions:

- `--interactive` specified more than once fails clearly.
- `--interactive` requires exactly one positional change id.
- `--interactive` is mutually exclusive with every direct generation mode flag and every direct generation input flag.
- Pre-seeding answers through flags is deferred. The first version avoids precedence questions and keeps prompt summaries authoritative.

## Non-Interactive Compatibility

Interactive mode must check terminal availability before prompting.

Behavior when stdin is not a TTY:

```text
interactive mode requires a TTY
```

The command exits non-zero and writes nothing.

Non-interactive automation remains supported through existing flags. CI jobs and scripts should continue to use direct commands such as:

```bash
specharbor generate add-login --blank
specharbor generate add-login --template feature
specharbor generate add-login --hybrid --template feature --title "Add login" --summary "Add login support"
```

Interactive mode must never block indefinitely waiting for input in CI.

## Prompt Flow

The initial prompt flow is intentionally small.

### Common Start

The CLI validates `<change-id>` before showing prompts. If invalid, it fails before reading from stdin.

Then it asks:

```text
Select generation path:
1. blank
2. built-in template
3. custom template
4. config template
5. hybrid
```

The CLI accepts either the listed number or the exact lower-case path name. Matching is trimmed and case-insensitive for user convenience, but the internal summary and request values are normalized and deterministic.

### Blank Path

Prompts:

- no additional prompts.

Summary:

```text
Interactive generation summary:
Change: <change-id>
Generation path: blank
Expected write target: openspec/changes/<change-id>/
Files: proposal.md, design.md, tasks.md, acceptance-criteria.md, risks.md
Validation: automatic no
Safety:
- Writes are limited to OpenSpec change files.
- Production code will not be modified.
- Source-control commands will not be run.
- Workflow automation will not be triggered.
- Provider, LLM, and agent APIs will not be called.
- No auto-commit, auto-push, PR creation, merge, or archive will be performed.

Proceed? [y/N]:
```

Generated request:

```text
generate <change-id> --blank
```

### Built-In Template Path

Prompts:

```text
Select built-in template:
1. feature
2. bugfix
3. docs
4. refactor
```

The CLI accepts either the listed number or the supported template name.

Generated request:

```text
generate <change-id> --template <template-name>
```

### Custom Template Path

Prompts:

- custom template name;
- optional title, where an empty answer means omitted;
- optional summary, where an empty answer means omitted.

Generated request:

```text
generate <change-id> --custom-template <template-name> [--title "<title>"] [--summary "<summary>"]
```

The optional title and summary prompts preserve the existing direct custom-template behavior. They are not required, and empty answers are valid omissions.

### Config Template Path

Prompts:

- config template alias;
- optional title, where an empty answer means omitted;
- optional summary, where an empty answer means omitted.

Generated request:

```text
generate <change-id> --config-template <alias> [--title "<title>"] [--summary "<summary>"]
```

The CLI does not resolve the alias before confirmation. Config parsing, alias lookup, remote-template fetches, and source-specific validation remain delegated to existing generation behavior after confirmation.

### Hybrid Path

Prompts:

```text
Select hybrid source:
1. built-in template
2. custom template
3. config template
```

Then prompt for the source value:

- built-in template name, selected from `feature`, `bugfix`, `docs`, `refactor`;
- custom template name;
- config template alias.

Then prompt for:

- title, required;
- summary, required;
- optional type, where an empty answer means omitted.

Generated requests:

```text
generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>" [--type <type>]
generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>" [--type <type>]
generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>" [--type <type>]
```

Hybrid keeps its existing semantics:

- built-in sources derive omitted type from the built-in template;
- config aliases resolving to built-in sources derive omitted type from the resolved built-in template;
- custom, config custom, and config remote sources do not infer type when omitted;
- provided type must match selected or resolved built-in templates;
- validation runs after successful hybrid writes or skip-only completion.

## Confirmation And Cancellation

Before invoking generation, interactive mode prints a deterministic summary with normalized values:

```text
Interactive generation summary:
Change: <change-id>
Generation path: <blank|built-in template|custom template|config template|hybrid>
Hybrid source: <built-in template|custom template|config template>
Template: <template-name>
Custom template: <template-name>
Config alias: <alias>
Title: <title>
Summary: <summary>
Type: <type>
Expected write target: openspec/changes/<change-id>/
Files: proposal.md, design.md, tasks.md, acceptance-criteria.md, risks.md
Validation: automatic <yes|no>
Safety:
- Writes are limited to OpenSpec change files.
- Production code will not be modified.
- Source-control commands will not be run.
- Workflow automation will not be triggered.
- Provider, LLM, and agent APIs will not be called.
- No auto-commit, auto-push, PR creation, merge, or archive will be performed.

Proceed? [y/N]:
```

Rules:

- Omit `Hybrid source`, `Template`, `Custom template`, `Config alias`, `Title`, `Summary`, or `Type` lines when they do not apply.
- The `Validation: automatic <yes|no>` line is always included.
- `Validation: automatic no` is used for blank, built-in template, custom template, and config template paths.
- `Validation: automatic yes` is used for the hybrid path.
- The `Safety:` section and its six bullet lines are always included exactly as shown.
- Safety notes appear after the validation line and before the confirmation prompt.
- Confirmation matching is case-insensitive after trimming whitespace.
- Trimmed `y` and `yes` in any casing proceed.
- Trimmed `n` and `no` in any casing cancel.
- Empty confirmation input cancels.
- EOF cancels.
- Unsupported confirmation answers retry up to three attempts using the same retry policy as other invalid required answers.
- After three unsupported confirmation answers, the command fails clearly with confirmation retry exhaustion and writes nothing.
- Cancellation exits non-zero with a clear `operation cancelled` message.
- Cancellation writes nothing because generation use cases are not invoked before confirmation.

## Input Validation

Prompted values use existing validation rules.

Required values:

- change id: existing `domain.NewChangeID` behavior;
- built-in template name: existing `domain.ParseTemplateName`;
- custom template name: existing `domain.NewCustomTemplateName`;
- config template alias: existing `domain.NewConfigTemplateAlias`;
- hybrid type when provided: existing hybrid type validation;
- hybrid title: existing hybrid metadata validation, non-empty after trimming;
- hybrid summary: existing hybrid metadata validation, non-empty after trimming.

Optional values:

- custom-template title may be omitted;
- custom-template summary may be omitted;
- config-template title may be omitted;
- config-template summary may be omitted;
- hybrid type may be omitted.

Retry rules:

- Required answers retry up to three invalid attempts.
- Menu selections retry up to three invalid attempts.
- Optional title and summary accept empty input without retry.
- Optional hybrid type accepts empty input without retry; invalid non-empty values retry up to three attempts.
- Retry exhaustion returns a clear error and writes nothing.

The prompt layer may perform early syntactic validation with domain value objects, but the generation use cases remain the final authority for source existence, config alias resolution, remote fetch safety, rendered file validation, write policy, and validation integration.

## Write Behavior

Interactive mode must not perform writes itself.

After confirmation, it delegates to existing generation behavior:

- blank path delegates to blank generation;
- built-in template path delegates to built-in template generation;
- custom template path delegates to custom template generation;
- config template path delegates to config template generation;
- hybrid path delegates to hybrid generation.

The selected use case controls:

- requiring OpenSpec project structure;
- creating `openspec/changes/<change-id>/`;
- writing only approved OpenSpec files;
- skipping existing files;
- preserving existing files;
- remote template safety through config aliases;
- validation behavior.

Interactive mode must not write before confirmation and must not add output paths.

## Validation Integration

Interactive mode follows the selected mode's existing validation behavior.

- Blank generation does not add automatic validation.
- Built-in template generation does not add automatic validation.
- Custom template generation does not add automatic validation.
- Config template generation does not add automatic validation.
- Hybrid generation keeps its existing automatic validation behavior.

The pre-confirmation summary exposes this behavior with exactly one validation line:

- blank path: `Validation: automatic no`;
- built-in template path: `Validation: automatic no`;
- custom template path: `Validation: automatic no`;
- config template path: `Validation: automatic no`;
- hybrid path: `Validation: automatic yes`.

Interactive mode does not offer to run validation as an additional prompt in the first version. That avoids changing direct mode semantics and keeps the prompt flow short. Users can run validation manually:

```bash
specharbor validate <change-id>
```

A future change may add an explicit post-generation validation prompt, but it must specify exit behavior and avoid changing automation contracts.

## Terminal Abstraction

Core must not know about terminals.

Layer responsibilities:

- `internal/adapters/cli` owns TTY detection, prompt text, answer parsing, retry loops, summary formatting, cancellation mapping, and fake terminal tests.
- `internal/adapters/cli` converts confirmed prompt answers into the existing internal generation argument/request shapes.
- `internal/core/domain` may own only pure value objects if an interactive-specific cancellation/result model becomes necessary, but no terminal behavior belongs in domain.
- `internal/core/usecase` remains non-interactive and continues to accept structured inputs.
- `internal/core/ports` should not gain a terminal port unless implementation proves it is needed by a core use case. The preferred first implementation uses an adapter-owned prompt abstraction.

Recommended adapter-owned abstraction:

```go
type interactiveTerminal interface {
    IsInputTerminal() bool
    ReadLine() (string, error)
    WriteString(string) error
}
```

The exact shape can vary, but it must support deterministic tests with fake input and captured output.

Forbidden in core:

- `os.Stdin`;
- `os.Stdout`;
- `os.Stderr`;
- terminal packages;
- TTY detection;
- prompt formatting;
- readline behavior.

## CLI Output

Prompt text should be user-friendly but deterministic.

Output requirements:

- Prompt labels and menu ordering are stable.
- Final pre-confirmation summary is stable.
- Cancellation message is clear.
- Non-TTY error is clear.
- Invalid-answer retry messages are clear and bounded.
- Successful generation reuses the selected mode's existing report as much as possible.
- Hybrid output keeps its existing validation and safety report behavior.
- Error output follows existing CLI error style.
- Remote credentials, query strings, auth headers, cookies, OAuth material, environment-derived secrets, and secret-bearing URLs are never printed.

Interactive prompt text may be more conversational than non-interactive reports, but generated summaries must be suitable for tests and documentation.

## Safety Boundaries

Interactive prompts must not:

- call provider APIs;
- call LLM APIs;
- call local model APIs;
- execute agents;
- execute local commands;
- execute remote commands;
- run shell commands;
- run scripts;
- parse or apply live runner output;
- import AI-authored files;
- write production code;
- write arbitrary paths;
- write docs, config, CI, prompt, archive, or source-control files;
- mutate config;
- run validation auto-fix;
- run source-control commands;
- commit;
- push;
- create pull requests;
- merge;
- archive;
- run workflow automation.

Remote templates remain reachable only when the user selects config-template or hybrid config-template and the existing config alias resolves to `source: remote`. The existing remote safeguards remain unchanged and run only after confirmation through the existing generation use case.

## Architecture

Expected implementation shape:

- Extend the generate parser to recognize `--interactive`.
- Add a small CLI prompt orchestrator in `internal/adapters/cli`.
- Add adapter-owned terminal IO abstractions for TTY detection and tests.
- Reuse existing domain value objects for prompt validation.
- Reuse existing use cases for generation.
- Reuse existing CLI report functions where practical.
- Add architecture tests proving core stays terminal-free.

Core packages must not import:

- adapters;
- CLI packages;
- `os`;
- `os/exec`;
- terminal packages;
- concrete HTTP clients for interactive behavior;
- provider SDKs;
- source-control SDKs;
- workflow SDKs;
- external-agent SDKs;
- shell or script execution packages;
- external TUI frameworks.

CLI code must not contain source resolution, remote safety, write-policy, validation-rule, or template-rendering business rules beyond prompt answer validation and mapping.

## Documentation

Implementation must update:

- `README.md`;
- `docs/usage.md`;
- `docs/generation-modes.md`;
- any CLI or interactive docs present at implementation time.

Docs must explain:

- what interactive mode does;
- command syntax;
- why `<change-id>` remains required;
- supported prompt paths;
- unsupported first-version paths;
- prompt sequence;
- validation and retry behavior;
- that interactive mode prints a deterministic summary before writing;
- that the summary includes whether validation is automatic;
- that blank, built-in template, custom template, and config template paths show `Validation: automatic no`;
- that hybrid paths show `Validation: automatic yes`;
- that the summary includes the deterministic safety notes;
- confirmation and cancellation behavior;
- that confirmation is case-insensitive;
- that `y` and `yes` proceed;
- that `n` and `no` cancel;
- that empty confirmation and EOF cancel;
- that cancellation writes nothing and exits non-zero;
- non-TTY behavior;
- selected-mode write behavior;
- selected-mode validation behavior;
- config-template remote behavior through aliases only;
- safety boundaries;
- examples.

Required examples:

```bash
specharbor generate add-login --interactive
```

Docs should include one transcript-style example for built-in template generation and one for hybrid generation. Transcript examples must include the validation line and safety section, and must not imply provider calls, agent execution, source-control automation, or arbitrary writes.

## Testing Strategy

Testing must cover domain reuse, CLI parsing and prompt behavior, request mapping, integration with existing generation use cases, regressions, architecture boundaries, documentation expectations, and safety boundaries. The detailed test list lives in `tasks.md`.
