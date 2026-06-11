# Risks: Implement Remote Templates

## Remote Trust

- A remote server can change the bytes served at a stable URL.
- A user may copy a checksum from an untrusted source.
- A pinned checksum protects deterministic bytes, but it does not prove the template content is well-designed or safe to use.

## Network Reliability

- Generation can fail because of DNS, TLS, connectivity, server downtime, timeouts, or non-`200` responses.
- No persistent cache means every remote-template generation depends on current network availability.
- Disabling redirects may surprise users whose hosting provider uses redirects by default.

## Archive Safety

- Zip archives can contain path traversal, absolute paths, symlinks, duplicate entries, executable bits, nested directories, or decompression bombs.
- Archive metadata can be inconsistent across platforms and zip tools.
- Rejecting extra or nested files is safer but less convenient for teams that want to package notes beside templates.

## Config Complexity

- Adding `source: remote` increases the config parser's source-specific validation surface.
- Users may confuse remote aliases with direct `--template` or `--custom-template` flags.
- Users may expect `template:` to remain the source-specific field for every alias kind, while remote aliases use `url`, `checksum`, and `format`.

## Output Safety

- URLs can accidentally contain secrets in userinfo, query strings, fragments, or paths.
- Even sanitized output can disclose internal hostnames.
- Error messages may include remote technical details that should not include credentials or tokens.

## Scope Creep

- Remote templates can invite marketplace search, unauthenticated arbitrary URLs, git clone, credentials, private registries, source-control automation, cache management, template metadata, includes, scripts, hooks, and package-manager semantics.
- Adding cache behavior now would require storage, invalidation, corruption handling, cleanup, and security rules beyond the first safe foundation.
- Supporting more archive formats or remote manifests would expand parsing and safety rules before there is a proven need.

## Behavior Drift

- Extending config-template generation could accidentally alter existing builtin/custom alias behavior.
- Adding network dependencies to CLI wiring could accidentally create network access for non-remote generation paths.
- Changing result or report models could break output expectations for current modes.

## Error Quality

- Remote failures can be layered: invalid config, URL validation, network errors, timeouts, non-`200` responses, size limits, checksum mismatch, malformed zip, unsafe archive entries, missing files, and write conflicts.
- Poorly ordered errors can make troubleshooting difficult.

## Mitigations

- Require explicit config aliases only; do not accept ad hoc remote URLs on the CLI.
- Require HTTPS and reject credentials, query strings, fragments, git URLs, SSH URLs, and local file URLs.
- Require `sha256` checksums for every remote template.
- Verify downloaded bytes before zip parsing.
- Disable redirects in the first version.
- Enforce strict response and uncompressed size limits.
- Support only zip archives with exactly the five required root-level OpenSpec files.
- Reject symlinks, executable entries, duplicate entries, extra file entries, nested paths, absolute paths, path traversal, and empty files.
- Use no persistent cache in the first version.
- Keep network access behind a remote-template fetch port and call it only for resolved `source: remote` aliases.
- Require OpenSpec project structure before network access.
- Keep all writes restricted to the existing change path and fixed required filenames.
- Add regression tests proving builtin/custom template generation and config-template aliases remain unchanged and network-free.
- Add architecture tests for core import boundaries, absence of concrete HTTP clients in core, and absence of shell/script execution.
- Sanitize CLI output by displaying the remote host and checksum algorithm, not credentials or query data.
- Document the checksum workflow and remote safety boundaries clearly.

## Open Questions

- A future cache should probably be checksum-keyed, but cache storage, invalidation, pruning, and corruption handling require a separate design.
- Private remote templates may be useful, but credentials, OAuth, tokens, and auth headers are intentionally excluded until a separate trust and secret-handling model exists.
- Redirect support may be useful for common hosting providers, but it requires explicit host-policy and checksum-before-parse tests before enabling.
