# Risks: Implement Config-Driven Templates

## Path Safety

- Config values are user-authored and may try to smuggle path traversal, absolute paths, remote URLs, or shell-like content through aliases or references.
- A future arbitrary local path feature could weaken the safe fixed-root custom template model if it is added without strict containment and symlink handling.

## Namespace Confusion

- Users may expect `--template <alias>` to resolve config aliases or expect config aliases to override built-in templates.
- The same name can legally exist as a built-in template, custom template, and config alias, which is safe by flag but may need clear documentation.

## Config Compatibility

- Current generated default config lacks a `version` field, while config reading expects supported version `1`.
- Existing projects with older config files may need manual edits before `--config-template` works.
- Tightening template-specific config validation could accidentally affect `config show` if implementation shares parsing paths without care.

## Scope Creep

- Config-driven aliases can invite requests for remote URLs, local paths, marketplace templates, variables, inheritance, includes, template execution, or source-control automation.
- Adding `local` path support now would expand the security surface beyond the established custom template root.

## Behavior Drift

- Reusing the existing generation use case and CLI parser may accidentally alter built-in, custom, blank, guided, agent-assisted, or AI-assisted generation behavior.
- Adding optional `--title` and `--summary` pass-through for config aliases could create ambiguity for built-in aliases if not documented and tested.

## Error Quality

- Config errors can be nested: invalid YAML, invalid alias names, unsupported source kinds, invalid referenced template names, and missing files can all appear similar to users.
- Poor error messages would make the feature hard to troubleshoot and could push users back to direct template flags.

## Mitigations

- Keep first scope limited to `builtin` and `custom`; reject `remote`, `local`, `url`, and `path` fields clearly.
- Validate aliases and source-specific references in the domain before custom template filesystem access.
- Keep namespaces disjoint by flag and document the model in README, usage docs, and generation-mode docs.
- Update generated default config for new projects with `version: 1` and `templates.aliases: {}` while avoiding automatic rewrites of existing user config.
- Treat `config show` changes as optional and low-risk only; if output changes, test and document it explicitly.
- Preserve direct built-in and custom generation paths with regression tests before and after the config-driven path is added.
- Use structured use case errors with config context so users can distinguish missing config, invalid config, missing alias, unsupported source, invalid reference, and missing template files.
- Keep all writes restricted to the existing OpenSpec change output path and fixed required filenames.
- Require architecture tests to prevent core imports of adapters, CLI packages, `os`, provider/network/source-control/workflow/agent SDKs, and process execution packages.
