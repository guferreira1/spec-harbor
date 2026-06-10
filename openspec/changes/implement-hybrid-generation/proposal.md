# Proposal: Implement Hybrid Generation

## Summary

Add a safe first hybrid OpenSpec generation mode:

```bash
specharbor generate <change-id> \
  --hybrid \
  --template <name> \
  --title "<title>" \
  --summary "<summary>"
```

The same hybrid mode also accepts exactly one deterministic template source selected by flag:

```bash
specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>"
```

`--type <feature|bugfix|docs|refactor>` is optional metadata. When present, it is validated and made available to hybrid rendering and CLI output. When the selected or resolved source is a built-in template and `--type` is omitted, hybrid derives the effective type from the selected or resolved built-in template name. When the selected or resolved source is a built-in template and `--type` is provided, it must match the built-in template name. Custom sources, config custom aliases, and config remote aliases do not infer type; if `--type` is omitted, `{{type}}` remains unresolved/verbatim.

Hybrid generation in this first version composes one existing deterministic template source with explicit metadata substitution, writes only the five OpenSpec change files, skips existing files by default, and then runs validation. It does not call provider APIs, does not execute agents, does not import AI-authored files, and does not automate source control or workflow steps.

## Problem

SpecHarbor already supports several safe generation paths:

- blank generation;
- built-in templates;
- project-local custom templates;
- config-driven aliases for built-in, custom, and pinned remote templates;
- guided generation from explicit metadata;
- AI-assisted generation from a strict local file;
- agent-assisted prompt generation and run-and-report execution.

Users often need a single command that starts from an approved template source while also applying guided title and summary metadata. Today they must choose either direct template generation, which may not receive metadata consistently across all source kinds, or guided generation, which does not reuse team templates. The gap is a deterministic composition path that creates a complete OpenSpec change without adding provider calls, runner application, shell execution, or source-control automation.

## Goal

Implement a hybrid generation foundation that:

- requires explicit `--hybrid` mode;
- requires exactly one deterministic template source:
  - `--template <name>`;
  - `--custom-template <name>`;
  - `--config-template <alias>`;
- requires non-empty `--title` and `--summary`;
- optionally accepts validated `--type <feature|bugfix|docs|refactor>`;
- derives omitted type metadata from direct built-in templates and config aliases resolving to built-in templates;
- does not infer omitted type metadata for custom sources, config custom aliases, or config remote aliases;
- resolves the selected source using the existing source-specific safety rules;
- applies deterministic metadata substitution to generated OpenSpec Markdown;
- writes only `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md` under `openspec/changes/<change-id>/`;
- skips existing files and preserves user-authored content;
- runs existing validation after successful generation or skips;
- reports generated files, skipped files, validation status, and safety boundaries.

## Scope

- Add `--hybrid` to `specharbor generate <change-id>`.
- Treat `--template`, `--custom-template`, and `--config-template` as source selectors when `--hybrid` is present.
- Require exactly one source selector with `--hybrid`.
- Disallow `--blank` for hybrid generation.
- Disallow `--guided`, `--ai-assisted`, `--agent-assisted`, `--from-file`, `--overwrite`, `--agent`, and `--execute` with `--hybrid`.
- Require `--title` and `--summary` for hybrid generation and trim them before use.
- Accept optional `--type` for hybrid generation and validate it against the existing supported type set.
- Make `{{change_id}}`, `{{title}}`, `{{summary}}`, and `{{type}}` available to hybrid metadata substitution, with `{{type}}` using a provided type or a built-in-derived effective type when available.
- Derive omitted type from direct built-in template sources and config aliases that resolve to built-in template sources.
- Do not infer omitted type for direct custom sources, config custom aliases, or config remote aliases.
- Preserve unknown or unresolved template tokens verbatim.
- Resolve remote templates only through an explicit `--config-template <alias>` whose config entry has `source: remote`.
- Reuse existing remote template HTTPS, checksum, ZIP, and archive safety behavior unchanged.
- Run existing `validate <change-id>` logic after successful hybrid generation or skips.
- Add domain, use case, CLI, adapter regression, architecture, and documentation tests as described in `tasks.md`.
- Update public docs in the implementation PR: `README.md`, `docs/usage.md`, `docs/generation-modes.md`, and any generation/template docs present at implementation time.

## Out of Scope

- Implementing code in this spec-authoring task.
- AI overlay from `--from-file`.
- Provider API calls.
- LLM API calls.
- Local model API calls.
- Live agent runner output application.
- `--agent-assisted --execute --apply`.
- Script execution.
- Shell execution.
- Template hooks.
- Arbitrary local paths.
- New remote flags such as `--remote-template`.
- Remote behavior beyond already-approved config-template remote aliases.
- Bypassing remote checksum, HTTPS, ZIP, size, or archive safeguards.
- Marketplace behavior.
- Git clone.
- Production code writes.
- Documentation, config, CI, prompt, archive, source-control, or arbitrary path writes by the hybrid command.
- Validation auto-fix.
- Auto-commit.
- Auto-push.
- Pull request creation.
- Merge automation.
- Automatic archive.
- Source-control or workflow automation.
- Documentation-only separate changes.

## Success Criteria

- `specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>"` generates a hybrid change from a built-in template.
- `specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>"` generates a hybrid change from a project-local custom template.
- `specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>"` generates a hybrid change from a config alias resolving to built-in, custom, or remote.
- Hybrid generation requires exactly one source selector and fails clearly for missing or multiple selectors.
- Hybrid generation requires non-empty title and summary.
- Optional type metadata is validated; if omitted for a selected or resolved built-in source, the effective type is derived from the selected or resolved built-in template.
- If a built-in source is selected or resolved, a provided type must match that built-in template.
- If a custom source, config custom alias, or config remote alias omits type, `{{type}}` remains unresolved/verbatim.
- Source resolution, custom template loading, config alias validation, and remote template safety behavior reuse the existing approved rules.
- Hybrid generation writes only the five required OpenSpec files under the active change directory.
- Existing files are skipped and preserved.
- Validation runs after successful generation or skips, warnings keep exit code `0`, and errors are reported before a non-zero exit.
- Existing blank, direct built-in template, direct custom template, direct config-template, remote config-template, guided, AI-assisted, agent-assisted, validate, prompt, review, archive, config, workflow, scan, init, help, and version behavior remains unchanged.
- Core packages do not import adapters, CLI packages, `os`, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, shell execution, or script execution packages for this feature.
