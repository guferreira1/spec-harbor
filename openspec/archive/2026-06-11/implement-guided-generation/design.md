# Design: Implement Guided Generation

## Overview

This change adds a third local generation mode:

```text
specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

Guided generation is non-interactive in this first implementation. The command is guided by explicit CLI flags, not by runtime prompts. The mode creates the standard OpenSpec change package and writes deterministic Markdown based on:

- guided type;
- title;
- summary.

The implementation must preserve the existing dependency direction:

```text
CLI adapter -> generation use case -> ports + domain
adapters -> ports
```

Core packages must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external process execution, or concrete template engines.

## CLI Contract

Supported command shapes after this change:

```text
specharbor generate <change-id> --blank
specharbor generate <change-id> --template feature
specharbor generate <change-id> --template bugfix
specharbor generate <change-id> --template docs
specharbor generate <change-id> --template refactor
specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --guided --type bugfix --title "<title>" --summary "<summary>"
specharbor generate <change-id> --guided --type docs --title "<title>" --summary "<summary>"
specharbor generate <change-id> --guided --type refactor --title "<title>" --summary "<summary>"
```

The CLI adapter must:

- parse exactly one change id;
- parse `--guided` as a generation mode flag;
- require `--type`, `--title`, and `--summary` when `--guided` is present;
- reject `--guided` without `--type`;
- reject `--guided` without `--title`;
- reject `--guided` without `--summary`;
- reject empty guided type, title, or summary values when representable by the parser;
- reject unknown guided types with a clear error;
- reject `--guided` combined with `--blank`;
- reject `--guided` combined with `--template`;
- preserve the existing `--blank` and `--template` parser behavior;
- reject unsupported flags;
- reject extra positional arguments;
- reject duplicate conflicting flags where the current parser can detect them;
- preserve existing unsafe change id validation behavior;
- obtain the project root from the current working directory;
- construct concrete dependencies and invoke the generation use case;
- print a deterministic human-readable report from the structured result.

The order of flags should follow existing generate parser behavior and may allow flags before or after the change id, as long as the behavior is covered by tests.

## Guided Inputs

Guided generation input should be represented explicitly at the use case boundary. The current `GenerateChangeInput` already carries project root, change id, generation mode, and template name. This change should extend it conservatively to carry guided input, such as:

- guided type;
- title;
- summary.

The exact field names may differ, but the use case boundary must not depend on CLI-specific argument structs.

Input validation rules:

- project root is required;
- change id is required and must remain a safe single path segment;
- guided type is required for guided mode;
- title is required for guided mode;
- summary is required for guided mode;
- guided type must be one of `feature`, `bugfix`, `docs`, or `refactor`;
- guided mode must reject missing guided content dependencies before generation;
- guided type and content must be validated before target directory or file writes.

Whitespace should be trimmed for validation. Generated content should use the normalized title and summary that passed validation.

## Guided Types

The supported guided types are intentionally the same first four built-in template categories:

```text
feature
bugfix
docs
refactor
```

The implementation may reuse the existing `domain.TemplateName` value if that keeps the model simple and accurate, or introduce a separate `domain.GuidedType` value if that avoids confusing guided content with built-in template selection. Either choice must keep the meaning clear:

- `--template <template-name>` selects generic built-in template content.
- `--guided --type <type> --title <title> --summary <summary>` selects guided content enriched with user-provided context.

Do not add custom guided types, provider-backed type discovery, config-backed type selection, marketplace concepts, or remote type registries.

## Required Files

Guided generation creates the same required files as blank and built-in template generation:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

The implementation must reuse the shared required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`, instead of duplicating the required file list in CLI or guided-specific policy code.

Guided content may be keyed by required filename, but the filenames to create must come from the shared domain policy.

## Guided Starter Content

All guided starter content must be deterministic, local, human-readable, generic, and safe to commit. It must include the supplied title and summary in generated output.

The content strategy should reuse existing built-in template conventions where appropriate:

- `feature` guided content should resemble feature template structure and include the supplied title and summary in relevant proposal, design, tasks, acceptance, and risk context.
- `bugfix` guided content should resemble bugfix template structure and include the supplied title and summary in relevant current behavior, expected behavior, impact, testing, and risk context.
- `docs` guided content should resemble docs template structure and include the supplied title and summary in relevant documentation goal, audience, source-of-truth, validation, and risk context.
- `refactor` guided content should resemble refactor template structure and include the supplied title and summary in relevant refactor goal, boundaries, compatibility, testing, and risk context.

Generated `tasks.md` files must contain unchecked tasks only. Guided content must not claim implementation, validation, review, or documentation updates have already happened.

Because the first guided mode is non-interactive, generated content must not include instructions that suggest SpecHarbor asked follow-up questions during command execution.

## Domain Model

Keep generation concepts under:

```text
internal/core/domain
```

Expected domain updates:

- support `GuidedMode` as an implemented generation mode;
- add or reuse a supported type value for `feature`, `bugfix`, `docs`, and `refactor`;
- add a guided generation result shape or extend the existing structured result so CLI reporting can identify guided mode and selected type;
- preserve existing template result behavior for `TemplateMode`;
- reuse the shared required OpenSpec change file policy.

Domain code must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, external-agent tooling, workflow SDKs, external process execution, or concrete template engines.

## Ports

Filesystem writes must continue through the generation filesystem port under:

```text
internal/core/ports
```

Guided content must be provided through a core-owned port. A possible shape is:

```text
ContentFor(input domain.GuidedChangeInput, relativePath string) (string, error)
```

or:

```text
ContentFor(guidedType domain.GuidedType, title string, summary string, relativePath string) (string, error)
```

The exact interface may differ if it is smaller or better aligned with existing code, but it must:

- be consumed by the generation use case;
- provide deterministic content for every required file;
- support `feature`, `bugfix`, `docs`, and `refactor`;
- allow unknown guided types to be rejected clearly before writes;
- return clear errors for unknown guided types or unknown filenames;
- avoid mixing guided content with prompt generation, validation, project initialization, AI providers, source-control integrations, workflow dispatchers, config repositories, external execution, or terminal interaction.

## Guided Content Adapter

Provide concrete guided content through an adapter, likely under:

```text
internal/adapters/templates
```

This is appropriate if the implementation reuses current built-in template generation conventions. Another adapter package is acceptable only if it has a clear reason and still implements a core-owned port.

Acceptable implementation choices:

- constants in Go files;
- small local render helpers for title and summary interpolation;
- embedded local files if they are clearer and remain deterministic repository assets.

Avoid runtime filesystem template discovery, custom template paths, remote templates, provider prompts, agent prompts, semantic generation, network access, source-control access, workflow access, external command execution, and config integration.

The adapter must return deterministic Markdown for every supported guided type and required file. It must return clear errors for unknown guided types and unknown filenames.

## Generation Use Case

Update generation orchestration under:

```text
internal/core/usecase
```

Expected behavior for guided generation:

- validate required use case dependencies;
- trim and validate project root;
- trim and validate change id;
- validate requested generation mode;
- validate guided type, title, and summary for guided mode;
- reject unknown guided types before target directory or file writes;
- reject unsafe change ids before filesystem writes;
- preserve existing OpenSpec project availability checks;
- preserve the existing clear project-structure error when `openspec/project.md` or `openspec/changes/` is missing;
- preserve the existing behavior that generation does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`;
- build the target relative path as `openspec/changes/<change-id>`;
- create the target change directory when missing;
- continue without error when the target change directory already exists;
- treat partially existing change directories as recoverable;
- obtain required filenames from the shared domain required-file policy;
- obtain starter content from the guided content port for each required file;
- write each missing required file through write-if-absent behavior;
- skip existing files without overwriting content;
- record created files and skipped existing files in the structured result;
- return selected guided type in the structured result when guided generation succeeds.

The use case must not print output, call `os`, perform terminal IO, call provider SDKs, call network APIs, run external agents, run external processes, access source-control APIs, import adapters, import workflow tools, or read config.

## Idempotency and Overwrite Policy

Guided generation must preserve existing generation idempotency:

- if the change directory does not exist, create it;
- if the change directory already exists, continue;
- if a required file does not exist, create it;
- if a required file already exists, skip it;
- never overwrite existing file contents;
- report created and skipped files deterministically.

If the directory already exists and some files exist, guided generation must create only missing required files using the provided title and summary. Existing files must remain unchanged even if their previous content came from blank or template generation.

## CLI Reporting

Output should be concise, deterministic, and human-readable. A successful guided generation report should identify:

- operation status;
- change id;
- generation mode;
- selected guided type;
- title;
- relative change path;
- whether the directory was created or already existed;
- created file count;
- skipped existing file count;
- created filenames when files were created;
- skipped filenames when files already existed.

The report may include the summary if it remains concise, but generated files must include the summary regardless of report formatting.

Do not print absolute local paths, debug output, provider details, agent details, workflow details, source-control details, network details, prompts, or validation summaries.

## Testing Strategy

Add focused tests for:

- guided mode domain representation;
- guided type validation for `feature`, `bugfix`, `docs`, and `refactor`;
- missing guided type validation;
- missing title validation;
- missing summary validation;
- guided content adapter support for every supported type and every required file;
- guided output includes the supplied title and summary;
- generated guided `tasks.md` files contain unchecked tasks only;
- content adapter errors for unknown guided types and unknown filenames;
- use case creates a new change with each guided type;
- use case fills in missing files in an existing change directory;
- use case skips existing files without overwriting content;
- use case rejects unknown guided types before target writes;
- use case rejects missing guided inputs before target writes;
- use case keeps unsafe change id behavior consistent with blank and template generation;
- CLI accepts `generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`;
- CLI accepts flag ordering consistent with current generate parser behavior;
- CLI rejects `--guided` without `--type`;
- CLI rejects `--guided` without `--title`;
- CLI rejects `--guided` without `--summary`;
- CLI rejects unknown guided types;
- CLI rejects `--guided` combined with `--blank`;
- CLI rejects `--guided` combined with `--template`;
- CLI rejects unsupported flags and extra positional arguments;
- CLI preserves existing blank generation behavior;
- CLI preserves existing template generation behavior;
- unrelated commands retain existing behavior.

Use temporary directories and fake ports where they keep tests deterministic.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```
