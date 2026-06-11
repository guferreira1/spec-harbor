# Design: Implement Template Generation

## Overview

This change adds a second local generation mode alongside blank generation:

```text
specharbor generate <change-id> --template <template-name>
```

Template generation creates the standard OpenSpec change package and writes deterministic starter Markdown tailored to a selected built-in template. It does not infer requirements from the project, call providers, run agents, read custom template files, access the network, integrate with source control, or use configuration.

The implementation must preserve the architecture established by blank generation:

```text
CLI adapter -> generation use case -> ports + domain
adapters -> ports
```

## CLI Contract

Supported command shapes:

```text
specharbor generate <change-id> --blank
specharbor generate <change-id> --template feature
specharbor generate <change-id> --template bugfix
specharbor generate <change-id> --template docs
specharbor generate <change-id> --template refactor
```

Existing blank generation behavior must remain unchanged.

For template generation, the CLI adapter must:

- parse exactly one change id;
- parse exactly one `--template` flag when template generation is requested;
- require a non-empty template name after `--template`;
- reject `--template` when no template name is provided;
- reject unknown template names with a clear error;
- reject using `--blank` and `--template` together;
- reject unsupported flags;
- reject extra positional arguments;
- reject duplicate generation mode flags;
- preserve existing unsafe change id validation behavior;
- obtain the project root from the current working directory;
- construct concrete dependencies and invoke the generation use case;
- print a deterministic human-readable report from the structured result.

The CLI adapter may format output, but it must not contain template body content, required-file policy, overwrite policy, filesystem write policy, provider logic, agent logic, workflow logic, or source-control logic.

## Built-In Templates

The only supported template names in this change are:

```text
feature
bugfix
docs
refactor
```

Unknown template names must produce a clear error before any target change directory or files are created.

Template names should be represented as domain concepts or validated domain values so use case inputs remain explicit. Do not add custom template paths, discovery behavior, template marketplace concepts, remote registry concepts, or config-backed template selection.

## Required Files

Template generation creates the same required files as blank generation:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

The implementation should reuse the existing shared required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`, instead of duplicating the required file list in CLI or generation-specific policy code.

Template starter content may be keyed by required filename, but the files to create must come from the shared required-file policy.

## Starter Content

All template content must be deterministic, local, generic, and safe to commit.

The `feature` template is for normal product capability changes.

- `proposal.md` should include sections such as Summary, Problem, Proposed Solution, Scope, Out of Scope, and Success Criteria.
- `design.md` should include Architecture Notes, Domain, Ports, Use Case, Adapters, CLI, Testing, and Validation sections.
- `tasks.md` should include baseline, domain, ports, use case, adapters, CLI, tests, and verification tasks.
- `acceptance-criteria.md` should focus on observable behavior and scope control.
- `risks.md` should focus on scope creep, architecture boundaries, and backwards compatibility.

The `bugfix` template is for correcting incorrect behavior.

- `proposal.md` should include Current Behavior, Expected Behavior, Impact, Scope, Out of Scope, and Success Criteria.
- `design.md` should include Root Cause, Fix Approach, Boundaries, Regression Testing, and Validation sections.
- `tasks.md` should include reproduce, test, fix, regression, and verification tasks.
- `acceptance-criteria.md` should focus on corrected behavior and no regressions.
- `risks.md` should focus on regression risk, incomplete reproduction, and over-fixing.

The `docs` template is for documentation-only changes.

- `proposal.md` should include Documentation Goal, Audience, Files to Update, Scope, Out of Scope, and Success Criteria.
- `design.md` should include Documentation Structure, Source of Truth, Accuracy Rules, and Validation sections.
- `tasks.md` should include inventory, README/docs updates, command verification, and Markdown verification tasks.
- `acceptance-criteria.md` should focus on Markdown-only scope and command accuracy.
- `risks.md` should focus on stale docs, overstating planned behavior, and mixing documentation with behavior changes.

The `refactor` template is for internal structure improvements without behavior changes.

- `proposal.md` should include Refactor Goal, Current Pain, Non-Functional Goal, Scope, Out of Scope, and Success Criteria.
- `design.md` should include Boundaries, Migration Plan, Compatibility, Testing, and Validation sections.
- `tasks.md` should include baseline tests, small refactor steps, regression tests, and verification tasks.
- `acceptance-criteria.md` should focus on unchanged external behavior.
- `risks.md` should focus on accidental behavior changes, broad diffs, and architecture boundary drift.

Generated `tasks.md` files must contain unchecked tasks only. Template content must not claim implementation has already happened.

## Domain Model

Keep generation concepts under:

```text
internal/core/domain
```

Expected domain concepts:

- generation mode value for template generation;
- supported built-in template name values;
- generation result fields that can include selected template name when applicable;
- reuse of the existing required OpenSpec change file policy.

Template names should be validated before filesystem writes. A possible domain representation is a small value type for built-in template names with constants for `feature`, `bugfix`, `docs`, and `refactor`.

Domain code must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, source-control SDKs, workflow SDKs, or concrete template engines.

## Ports

Keep filesystem writes behind the existing generation filesystem port or a compatible small generation-specific port under:

```text
internal/core/ports
```

Template content must be accessed through a port owned by the core, such as:

```text
SupportedTemplates() []domain.TemplateName
ContentFor(templateName domain.TemplateName, filename string) (string, error)
```

The exact interface may differ if it is smaller or better aligned with existing code, but it must:

- be consumed by the generation use case;
- allow unknown template names to be rejected clearly before writes;
- provide content for every required file;
- return clear errors for unknown template names or unknown filenames;
- avoid mixing template content with prompt generation, validation, AI providers, source-control integrations, workflow dispatch, project initialization, or config.

## Template Content Adapter

Provide built-in template content through a concrete adapter, likely under:

```text
internal/adapters/templates
```

Acceptable implementation choices:

- constants in Go files;
- embedded local files if the design proves they are clearer and remain part of the repository.

Avoid external template files loaded at runtime unless there is a concrete reason and test coverage for deterministic behavior. Do not add filesystem template discovery, user template paths, remote template registries, provider prompts, agent prompts, or semantic rendering in this change.

The adapter must return deterministic Markdown for every supported template and required file. It must return clear errors for unknown template names and unknown filenames.

## Generation Use Case

Update generation orchestration under:

```text
internal/core/usecase
```

Expected behavior for template generation:

- validate dependencies;
- trim and validate project root;
- trim and validate change id;
- validate requested generation mode;
- validate the selected template name;
- reject unsafe change ids before filesystem writes;
- verify OpenSpec project availability the same way blank generation does;
- return the existing clear project-structure error when `openspec/project.md` or `openspec/changes/` is missing;
- build the target relative path as `openspec/changes/<change-id>`;
- create the target change directory when missing;
- continue without error when the target change directory already exists;
- treat partially existing change directories as recoverable;
- obtain required filenames from the shared domain required-file policy;
- obtain starter content from the template content port for each required file;
- write each required file through write-if-absent behavior;
- record created files and skipped existing files;
- never overwrite existing files;
- return a structured generation result.

Unknown template names and missing template names should be rejected before target writes occur. CLI argument errors may be rejected in the CLI adapter before invoking the use case, but the use case must still protect its own boundary against invalid template values.

The use case must not print output, call `os`, perform terminal IO, call provider SDKs, call network APIs, run external agents, run external processes, access source-control APIs, import adapters, or import workflow tools.

## Idempotency and Overwrite Policy

Template generation must preserve current generation idempotency:

- if the change directory does not exist, create it;
- if the change directory already exists, continue;
- if a required file does not exist, create it;
- if a required file already exists, skip it;
- never overwrite existing file contents;
- report created and skipped files deterministically.

This behavior must match blank generation and must be covered by tests.

## CLI Reporting

Output should be concise, deterministic, and human-readable. A successful template generation report should identify:

- operation status;
- change id;
- selected template;
- relative change path;
- whether the directory was created or already existed;
- created file count;
- skipped existing file count;
- created filenames when files were created;
- skipped filenames when files already existed.

Do not print absolute local paths, debug output, provider details, agent details, workflow details, source-control details, network details, or validation summaries.

## Testing Strategy

Add focused tests for:

- domain template name validation;
- template generation mode representation;
- template content adapter support for `feature`, `bugfix`, `docs`, and `refactor`;
- content adapter errors for unknown templates and unknown filenames;
- use case creates a new change with each built-in template;
- use case fills in missing files in an existing change directory;
- use case skips existing files without overwriting content;
- use case rejects unknown template names before writes;
- use case keeps unsafe change id behavior consistent with blank generation;
- use case does not call provider, agent, source-control, workflow, network, config, or external execution dependencies;
- CLI accepts `generate <change-id> --template <template-name>`;
- CLI rejects missing template name;
- CLI rejects `--blank` and `--template` together;
- CLI rejects unsupported flags and extra arguments;
- CLI preserves existing blank generation behavior;
- unrelated commands retain existing behavior.

Use temporary directories and fake ports where they keep tests deterministic.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```
