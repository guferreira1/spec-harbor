# Proposal: Implement Custom Templates

## Summary

Let users define reusable, project-local OpenSpec change templates under `.specharbor/templates/<template-name>/` and generate changes from them with a new explicit flag:

```bash
specharbor generate <change-id> --custom-template <template-name>
```

A custom template is a plain directory containing the five standard OpenSpec change files (`proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, `risks.md`). Generation reads those files through the existing filesystem port, performs minimal deterministic variable substitution (`{{change_id}}`, `{{title}}`, `{{summary}}`), and writes the rendered files only under `openspec/changes/<change-id>/` using the existing write-if-absent behavior.

No remote templates, no config-driven registry, no marketplace, no network calls, no template execution, and no shell or script execution. Templates are static Markdown content only.

## Problem

SpecHarbor ships exactly four built-in templates (`feature`, `bugfix`, `docs`, `refactor`) compiled into the binary. Teams that repeatedly author similar OpenSpec changes — for example an API feature with house conventions, a data-migration change, or a compliance-review change — cannot reuse their own starter content without forking SpecHarbor or copy-pasting previous change directories by hand. Hand-copying old changes is error-prone: stale ids leak into new files, required files get missed, and content drifts from team conventions.

The product flow `Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive` starts faster and more consistently when teams can encode their own starter packages once and reuse them deterministically.

## Goal

A user can create a directory under `.specharbor/templates/<template-name>/` containing the five standard OpenSpec change files and generate a new change from it with one command, with the same safety, determinism, and skip-existing-files behavior as built-in template generation — without any change to SpecHarbor code, configuration, or built-in template behavior.

## Scope

- Add a `--custom-template <template-name>` flag to `specharbor generate <change-id>` as a new, mutually exclusive generation mode flag alongside `--blank`, `--template`, `--guided`, and `--agent-assisted`.
- Discover custom templates only from the fixed project-local root `.specharbor/templates/<template-name>/`, resolved against the project root exactly as existing generation resolves it.
- Add a domain-owned `CustomTemplateName` value object mirroring the `ChangeID` safety rules: non-empty, single path segment, no `/` or `\`, no `..` sequences, no leading `.` or `-`, allowed characters `[A-Za-z0-9._-]`, maximum length 128, with clear errors raised before any filesystem access.
- Require all five standard OpenSpec change files in a valid custom template; missing or empty (whitespace-only) template files fail with clear errors before any write occurs.
- Ignore unknown extra files and subdirectories inside a template directory; only the five known filenames are read and copied.
- Support minimal deterministic variable substitution in template content: `{{change_id}}` always, plus `{{title}}` and `{{summary}}` when the optional `--title` and `--summary` flags are provided with `--custom-template`. Unresolved variables remain in the output as-is.
- Write rendered files only under `openspec/changes/<change-id>/`, reusing the existing required-file list, the existing write-if-absent conflict behavior, and the existing requirement that the OpenSpec project structure exists.
- Keep built-in `--template` behavior, names, content, output, and errors unchanged; a custom template may share a built-in name without affecting `--template` resolution because the two flags resolve from disjoint sources.
- Extend the generation result and CLI report so custom-template output shows the change id, the template name, that the template is custom, the relative template source path, created files, skipped existing files, and a note that only OpenSpec change files were written.
- Route all template reads through a small core-owned port implemented by the existing local filesystem adapter; core packages gain no `os`, adapter, or CLI imports.
- Add domain, use case, adapter, CLI, regression, and architecture tests as described in `tasks.md`.
- Update `README.md`, `docs/usage.md`, and `docs/generation-modes.md` inside this change's implementation work because public CLI behavior changes.

## Out of Scope

- Implementing code in this spec-authoring task.
- Remote templates, template downloads, or any network access.
- A config-driven template registry; discovery is fixed-path only (`.specharbor/templates/`), and no YAML configuration keys are added or read for templates.
- Marketplace or shared template distribution.
- Hybrid generation combining custom templates with other modes.
- A templating language: no conditionals, no loops, no functions, no includes, no front matter, and no template-defined variables.
- Template execution of any kind: no scripts, no hooks, no shell commands, no external processes.
- Provider APIs, local model calls, credentials, or OAuth.
- Templates controlling output paths; output filenames are exactly the five known OpenSpec change filenames.
- Writes outside `openspec/changes/<change-id>/`: no production code writes, no config writes, no CI writes, and no documentation writes other than the documentation updates listed in this change.
- Overriding or shadowing built-in templates through `--template`.
- Changes to `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, or `workflow` behavior.
- Validation of custom template source directories by `specharbor validate` (it validates generated changes, not template sources).
- Interactive prompts.

## Compatibility

- `--template <name>` keeps resolving only the four built-in templates with unchanged content, output, and errors; a name that is not built-in keeps failing with the existing unknown-template error even when a custom template with that name exists.
- Blank, guided, and agent-assisted generation are unchanged.
- `--title` and `--summary` remain required with `--guided` and `--agent-assisted` exactly as today; they become optionally accepted with `--custom-template` only.
- Projects without a `.specharbor/templates/` directory are unaffected; the directory is only inspected when `--custom-template` is used.
- Generated changes remain compatible with `specharbor validate <change-id>`; validation findings depend on the template's content quality, exactly as for hand-authored changes.

## Success Criteria

- A user can create `.specharbor/templates/<template-name>/` with the five standard files and run `specharbor generate <change-id> --custom-template <template-name>` to produce `openspec/changes/<change-id>/` containing the five rendered files.
- `{{change_id}}` is replaced with the change id in every generated file; `{{title}}` and `{{summary}}` are replaced when provided and left verbatim when not.
- Invalid custom template names (path traversal, separators, absolute paths, leading `.` or `-`, unsafe characters, over-length, empty) are rejected with clear errors before any filesystem access.
- A missing template directory, missing required template files, and empty template files each produce a distinct, clear error, and no file or directory is created under `openspec/changes/` when the template is invalid.
- Extra files in the template directory are ignored and never copied.
- Existing files under `openspec/changes/<change-id>/` are never overwritten; they are reported as skipped, matching existing generation behavior.
- CLI output identifies the template as custom, shows its relative source path, and lists created and skipped files.
- All existing generation modes, `--template` included, behave identically to before this change, proven by regression tests.
- Core packages gain no imports of adapters, CLI packages, `os`, network, or external SDKs; architecture tests keep passing.
- `README.md`, `docs/usage.md`, and `docs/generation-modes.md` document the directory structure, required files, command usage, variables, safety boundaries, and the explicit non-goals (no remote templates, no config registry, no template execution).
