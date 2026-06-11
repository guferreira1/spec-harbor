# Proposal: Implement Config-Driven Templates

## Summary

Add a safe config-driven template alias layer for OpenSpec change generation:

```bash
specharbor generate <change-id> --config-template <alias>
```

Projects can declare named template aliases in `.specharbor/config.yml` under `templates.aliases`. Each alias resolves to either an existing built-in template or an existing project-local custom template. Generation then delegates to the same built-in or custom template behavior already used by `--template` and `--custom-template`.

This change is intentionally local and deterministic. It does not add remote templates, arbitrary local paths, marketplace lookup, provider APIs, template execution, shell execution, source-control automation, or writes outside `openspec/changes/<change-id>/`.

## Problem

SpecHarbor now supports built-in templates and project-local custom templates, but users still have to remember the concrete source flag and template name for each recurring change type. A project may want stable team-facing names such as `api-feature`, `default-feature`, or `service-refactor` without requiring every contributor or agent prompt to know whether the underlying source is built-in or custom.

The next product step is a configuration layer that maps project-owned aliases to approved template references. That layer should prepare for future remote or hybrid template resolution without introducing those higher-risk behaviors now.

## Goal

A user can add a versioned local config entry like:

```yaml
version: 1

templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature

    default-feature:
      source: builtin
      template: feature
```

and generate a change with:

```bash
specharbor generate add-payment-flow --config-template api-feature
```

SpecHarbor resolves the alias from `.specharbor/config.yml`, validates the alias and referenced template, and generates only the standard OpenSpec change files under `openspec/changes/<change-id>/`.

## Scope

- Add a config-driven template generation mode to `specharbor generate <change-id>` through `--config-template <alias>`.
- Define the `.specharbor/config.yml` schema for configured template aliases under `templates.aliases`.
- Support exactly two source kinds in this change:
  - `builtin`, referencing existing built-in templates such as `feature`, `bugfix`, `docs`, and `refactor`;
  - `custom`, referencing existing project-local custom templates under `.specharbor/templates/<template-name>/`.
- Reject unsupported source kinds and remote/path-style fields with clear config validation errors.
- Add a domain-owned config template alias value object with safe single-path-segment validation.
- Add domain-owned config template reference and source models.
- Load `.specharbor/config.yml` through core-owned ports and adapter-owned filesystem/YAML behavior, using the existing config loading path where practical.
- Require config version `1` for template alias resolution.
- Update generated default `.specharbor/config.yml` so new projects include `version: 1` and an empty `templates.aliases` section.
- Resolve aliases only when `--config-template` is used.
- Keep built-in, custom, blank, guided, agent-assisted, and AI-assisted generation behavior unchanged when their existing flags are used.
- Keep namespaces disjoint by flag:
  - `--template <name>` resolves only built-in templates;
  - `--custom-template <name>` resolves only `.specharbor/templates/<name>/`;
  - `--config-template <alias>` resolves only `templates.aliases.<alias>`.
- Allow the same text to exist as a built-in template name, a custom template name, and a config alias because the flag disambiguates it.
- Reuse existing write-if-absent conflict behavior and custom template validation behavior after an alias resolves.
- Update public docs in the implementation work: `README.md`, `docs/usage.md`, `docs/generation-modes.md`, and any config/template-specific docs that exist at implementation time.
- Add the domain, use case, adapter, CLI, regression, and architecture tests listed in `tasks.md`.

## Out of Scope

- Implementing code in this spec-authoring task.
- Remote template URLs or template downloads.
- Marketplace templates.
- Arbitrary local template paths from config.
- A `local` source kind.
- Hybrid generation.
- AI provider API calls.
- Local model API calls.
- Agent execution.
- Template scripts, hooks, shell commands, or external process execution.
- Template front matter, includes, conditionals, loops, or additional templating language features.
- Template aliases controlling output paths.
- Writes outside `openspec/changes/<change-id>/`.
- Production code writes.
- Documentation/config/CI writes by the generation command.
- Source-control automation, including commit, push, pull request creation, merge, archive, or task checkbox automation.
- Remote credential storage, OAuth, or provider configuration.
- Config mutation commands such as `config set`.
- Validation of remote template metadata.
- A documentation-only separate change.

## Compatibility

- Existing `--template <name>` behavior remains unchanged and never consults config.
- Existing `--custom-template <name>` behavior remains unchanged and never consults config.
- Existing blank, guided, agent-assisted, AI-assisted, validate, prompt, review, archive, scan, workflow, and config behavior remains unchanged except for any explicitly documented `config show` display addition if the implementation chooses to show template aliases.
- A project without `.specharbor/config.yml` is unaffected unless `--config-template` is used. In that mode, missing config is a clear error.
- A project with no `templates.aliases` entries can still use every existing generation mode.
- Existing initialized projects may need to add `version: 1` manually before using config-driven templates if their generated config predates this change.

## Success Criteria

- `specharbor generate <change-id> --config-template <alias>` resolves `<alias>` from `.specharbor/config.yml` and delegates to the referenced built-in or custom template generation behavior.
- Alias validation rejects empty values, separators, absolute/path-like values, traversal, `..` sequences, leading `.` or `-`, unsupported characters, and over-length values before template filesystem access.
- Config schema errors, missing config, unsupported config version, missing alias, unsupported source kind, unknown built-in templates, missing custom templates, and missing custom template files all produce clear errors.
- Config-driven generation writes only under `openspec/changes/<change-id>/` and only the five standard OpenSpec change files.
- Existing files are skipped and never overwritten, matching built-in/custom generation behavior.
- CLI success output identifies the change id, config alias, resolved source kind, resolved template name, generated files, skipped files, and the OpenSpec-only write boundary.
- Documentation explains the schema, command usage, source kinds, disjoint flag behavior, validation expectations, examples, and safety boundaries.
- Tests prove existing generation modes and non-generation commands remain unchanged.
- Core packages do not import adapters, CLI packages, `os`, network/provider/source-control/workflow/agent SDKs, or process execution packages.
