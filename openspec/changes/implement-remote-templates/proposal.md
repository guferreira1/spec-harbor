# Proposal: Implement Remote Templates

## Summary

Add a safe remote template source for config-driven OpenSpec change generation:

```bash
specharbor generate <change-id> --config-template <alias>
```

Projects may declare explicit remote template aliases in `.specharbor/config.yml`. A remote alias points to one HTTPS zip bundle, pins the exact bundle bytes with a required `sha256` checksum, and generates the same five OpenSpec change files as every other generation path.

This change is intentionally narrow. It is not a marketplace, package manager, source-control integration, provider integration, arbitrary downloader, script runner, or remote code execution system.

## Problem

SpecHarbor supports built-in templates, project-local custom templates, and config-driven aliases for local template references. Teams may also want to reuse approved templates that are hosted outside a single repository, such as a company-maintained template bundle. Without a controlled remote-template model, users either copy templates manually into each project or invent ad hoc download steps outside SpecHarbor's safety boundaries.

The next product step is to define a deterministic remote template foundation that keeps all of SpecHarbor's existing OpenSpec-only write guarantees.

## Goal

A project can configure a remote template alias explicitly:

```yaml
version: 1

templates:
  aliases:
    service-feature:
      source: remote
      url: https://example.com/specharbor/templates/service-feature.zip
      checksum: sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
      format: zip
```

and then generate a change with:

```bash
specharbor generate add-service-endpoint --config-template service-feature
```

SpecHarbor resolves the alias, validates the remote reference, fetches the pinned HTTPS bundle, verifies the checksum before parsing the zip, extracts only the approved OpenSpec template files, and writes only under:

```text
openspec/changes/<change-id>/
```

## Scope

- Extend config-driven template aliases with a new `source: remote` kind.
- Support remote templates only through explicit `.specharbor/config.yml` aliases.
- Require `url`, `checksum`, and `format` for every remote alias.
- Support only `format: zip` in the first version.
- Require HTTPS URLs with no credentials, no query string, no fragment, no local file URLs, no git URLs, no SSH URLs, and no redirects.
- Require checksum strings in the form `sha256:<64 hex characters>`.
- Verify the checksum over the downloaded zip bytes before archive parsing.
- Fetch remote templates only when a requested `--config-template <alias>` resolves to `source: remote`.
- Do not use a persistent cache in the first version.
- Decode zip bundles with strict archive safety rules.
- Require the zip to contain exactly the five required OpenSpec change files as root-level regular files:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- Reject archive path traversal, absolute paths, nested paths, symlinks, executable entries, duplicate files, extra files, missing files, empty files, and oversized bundles.
- Reuse existing config-template command behavior, write-if-absent conflict behavior, and OpenSpec project structure requirements.
- Keep generated output limited to `openspec/changes/<change-id>/` and the five approved filenames.
- Add domain, use case, adapter, CLI, regression, architecture, and documentation tasks for implementation.

## Out of Scope

- Implementing code in this spec-authoring task.
- Modifying files outside `openspec/changes/implement-remote-templates/` in this task.
- A separate `--remote-template` command or flag.
- Automatic remote template discovery.
- Marketplace search or registry browsing.
- Package-manager behavior.
- Git clone, GitHub/GitLab source-control APIs, SSH URLs, git URLs, or repository automation.
- Credentials, OAuth, auth headers, cookies, userinfo URLs, query-token credentials, or environment token expansion.
- HTTP URLs, file URLs, local path URLs, or non-HTTPS transports.
- Redirect following.
- Unpinned remote templates.
- Checksum algorithms other than `sha256`.
- Tar archives, directory indexes, remote manifests, individual file URLs, or remote include graphs.
- Persistent template caching.
- Template scripts, hooks, shell execution, external command execution, or executable archive behavior.
- Provider APIs, local model APIs, AI-assisted generation changes, or agent runner changes.
- Hybrid generation beyond using a remote template as the template source.
- Production code writes.
- Documentation, config, CI, prompt, archive, or arbitrary path writes by generation.
- Auto-commit, auto-push, PR creation, merge, archive, task checkbox updates, or workflow automation.

## Compatibility

- Existing `--template <name>` behavior remains unchanged and never accesses the network.
- Existing `--custom-template <name>` behavior remains unchanged and never accesses the network.
- Existing config aliases with `source: builtin` and `source: custom` remain unchanged.
- Built-in, custom, and remote aliases remain disjoint by the explicit `--config-template` alias source declared in config.
- A project without a requested remote alias never performs remote template network access.
- Generated remote-template changes remain ordinary OpenSpec change packages and work with `specharbor validate <change-id>`.
- Generation does not auto-run validation, matching current config-template generation behavior.

## Success Criteria

- `specharbor generate <change-id> --config-template <alias>` can resolve a configured `source: remote` alias and generate the five required files from a pinned HTTPS zip bundle.
- Invalid remote config fails before network access where possible.
- Network access occurs only for the requested remote alias and only through adapter-owned fetch behavior.
- Downloaded bytes must match the configured `sha256` checksum before zip parsing.
- Checksum mismatch, network failure, timeout, size-limit failure, unsafe archive content, or missing required files produces no OpenSpec writes.
- Remote generation writes only the five approved filenames under `openspec/changes/<change-id>/`.
- Existing files are skipped and never overwritten.
- CLI success output identifies the change id, config alias, remote source kind, sanitized host, checksum algorithm, created files, skipped files, and safety boundaries without displaying credentials.
- Documentation explains the config schema, checksum requirement, HTTPS-only behavior, supported zip format, `--config-template` usage, and safety boundaries.
- Tests prove existing generation modes, AI-assisted generation, validation, prompt, review, archive, config, workflow, and agent-runner behavior remain unchanged.
- Core packages do not import adapters, CLI packages, concrete HTTP clients, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, `os`, terminal IO, or process execution packages for this feature.
