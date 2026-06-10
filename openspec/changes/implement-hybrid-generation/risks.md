# Risks: Implement Hybrid Generation

## Risks

### Mode Confusion

Hybrid reuses `--template`, `--custom-template`, and `--config-template` as source selectors. Users may confuse direct template behavior with hybrid behavior, especially because hybrid requires metadata and runs validation while direct template generation does not.

### Source Conflict Ambiguity

The same text may exist as a built-in template name, a custom template name, and a config alias. If the implementation guesses a source or adds fallback behavior, a command could silently use the wrong template source.

### Metadata Contradictions

Optional type metadata can conflict with built-in template names, for example a bugfix template paired with `--type feature`. Contradictory metadata would create confusing OpenSpec content and downstream prompts.

### Omitted Type Ambiguity

If omitted type behavior differs by implementation path, direct built-in sources, config aliases resolving to built-ins, custom templates, and remote templates could render `{{type}}` inconsistently.

### Remote Template Trust

Hybrid can use remote templates through config aliases. Remote template content is still external content, even when HTTPS, checksum, and ZIP safeguards are enforced. Applying metadata substitution to verified remote content must not become remote code execution or a bypass around remote safeguards.

### AI Overlay Scope Creep

Hybrid is adjacent to AI-assisted and agent-assisted workflows. Adding `--from-file` overlay in the first version would raise overwrite, partial overlay, provenance, validation, and malformed-output questions that are already handled differently by AI-assisted generation.

### Validation Exit Surprise

Direct template and config-template generation do not auto-run validation. Hybrid intentionally does. Users may be surprised when generation writes files but exits non-zero because validation found errors.

### Partial Runtime Writes

The design requires source and target preflight before writes, but a runtime filesystem error can still occur after some files are created.

### Parser Regression

Adding `--hybrid` to the existing `generate` parser could accidentally alter blank, direct template, custom template, config-template, guided, AI-assisted, or agent-assisted command behavior.

### Architecture Leakage

Hybrid touches source resolution, metadata rendering, validation, remote templates, and CLI reporting. There is a risk of putting business rules in CLI code or concrete filesystem, config, remote, or template details in core.

### Documentation Drift

Hybrid will become public CLI behavior. Incomplete docs could make it look like provider-backed generation, live runner application, remote marketplace behavior, or source-control automation.

## Mitigations

### Keep The Command Explicit

Require `--hybrid` and exactly one source selector. Preserve direct behavior when `--hybrid` is absent. Document that direct template commands remain unchanged and that hybrid is a separate composition mode.

### Keep Source Namespaces Disjoint

Use the existing flag-selected namespace model:

- `--template` means built-in;
- `--custom-template` means project-local custom;
- `--config-template` means config alias.

Do not add fallback, source guessing, or precedence. Test same-name scenarios across all three namespaces.

### Validate Metadata Early

Require title and summary before source writes. Validate optional type against the supported set. Reject built-in source and type mismatches before writes.

### Make Omitted Type Deterministic

Derive omitted type only from direct built-in templates and config aliases resolving to built-in templates. Do not infer type for custom sources, config custom aliases, or config remote aliases; leave `{{type}}` unresolved unless `--type` is provided. Print `Type: <type>` only when an effective type exists.

### Delegate Remote Safety

Allow remote templates only through existing config aliases. Reuse the existing URL, checksum, fetch, ZIP, archive, and size-limit safeguards unchanged. Apply only deterministic string substitution to verified Markdown content and never execute remote content.

### Exclude AI Overlay From Version One

Reject `--from-file` with `--hybrid`. Keep AI-assisted from-file import as a separate mode. Require any future overlay behavior to have its own OpenSpec change with strict parser reuse, preflight, overwrite, and validation semantics.

### Make Validation Behavior Visible

Print validation status, error count, warning count, and findings in the hybrid report. Document that validation errors leave generated files in place for editing but make the command exit non-zero.

### Preflight Before Writes

Resolve and validate the selected source, metadata, config alias, remote reference, checksum, archive, and rendered required-file set before creating the change directory. Skip existing files by default and do not add overwrite in the first version. Report unexpected runtime write failures clearly without deleting user files.

### Protect Existing Modes With Regression Tests

Add CLI and use case regression tests for every existing generation mode and non-generation command. Keep parser changes localized and explicit.

### Preserve Hexagonal Boundaries

Put hybrid source selection, metadata, rendering policy, and result concepts in domain. Put orchestration in use cases. Reuse core-owned ports. Keep CLI limited to parsing, wiring, output formatting, and exit-code mapping. Extend architecture tests for core import boundaries and safety exclusions.

### Update Public Docs In The Same Change

Update README and generation docs with command examples, source-selection rules, metadata behavior, validation exit behavior, remote-template relationship, and safety boundaries. State that provider APIs, LLM APIs, live runner application, shell execution, production code writes, and source-control automation are not part of hybrid generation.

## Trade-Offs Accepted

- Hybrid runs validation even though direct template generation does not. This gives the explicit hybrid mode a more complete authoring pass at the cost of a possible non-zero exit after files are written.
- AI overlay is deferred. This makes the first version less powerful, but keeps the safety model clear and implementation scope smaller.
- No overwrite flag is included. Existing files are preserved, which supports safe reruns but requires manual cleanup or a future explicit overwrite design for replacement workflows.
- Type is derived for built-in sources when omitted. This makes built-in hybrid output deterministic and keeps built-in template type aligned with rendered metadata.
- Type is optional and not inferred for custom and remote sources. This avoids forcing teams to classify every custom template while still supporting type metadata when users provide it.

## Open Questions

- A future AI overlay may be useful, but it should be specified separately with strict parser reuse, partial versus complete overlay rules, overwrite controls, preflight guarantees, and validation behavior.
- A future `--validate=false` escape hatch may be requested, but this first version keeps validation mandatory to make hybrid behavior predictable.
- A future richer templating language may be requested, but this change deliberately keeps rendering to static string substitution only.
