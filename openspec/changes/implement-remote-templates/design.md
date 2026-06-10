# Design: Implement Remote Templates

## Overview

Remote templates extend config-driven template aliases with one new explicit source kind:

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

The command surface remains:

```bash
specharbor generate <change-id> --config-template <alias>
```

The execution flow is:

1. Parse `--config-template <alias>` with the existing config-template flag rules.
2. Validate the change id and config alias.
3. Load `.specharbor/config.yml` through the existing config port path.
4. Resolve the alias.
5. If the alias is `builtin` or `custom`, keep existing behavior unchanged.
6. If the alias is `remote`, validate the URL, checksum, and format.
7. Require the OpenSpec project structure before network access.
8. Fetch the remote bundle through a remote-template fetch port.
9. Enforce timeout and max download size in the HTTP adapter.
10. Verify the `sha256` checksum over the downloaded zip bytes in core before archive parsing.
11. Decode the zip through a bundle reader port using strict archive policy.
12. Validate the resulting five file contents.
13. Write with existing write-if-absent behavior under `openspec/changes/<change-id>/`.
14. Return a structured generation result for CLI reporting.

The feature intentionally adds one controlled remote content source. It does not add remote code execution, script execution, provider calls, source-control automation, or arbitrary output paths.

## Remote Source Schema

Use source-specific fields for remote aliases:

```yaml
templates:
  aliases:
    api-feature:
      source: remote
      url: https://example.com/templates/api-feature.zip
      checksum: sha256:<64-hex-characters>
      format: zip
```

Required fields for `source: remote`:

- `source`, exactly `remote`;
- `url`, an HTTPS URL value;
- `checksum`, exactly `sha256:<hex>`;
- `format`, exactly `zip`.

No optional metadata fields are included in the first version. Fields such as `name`, `description`, `template`, `version`, `path`, `headers`, `auth`, `branch`, `tag`, `ref`, `command`, or `script` are unsupported for remote aliases and must fail clearly when the alias is parsed for config-template generation.

### URL Rules

The domain owns pure URL validation. It may use standard URL parsing, but it must not use HTTP clients or perform network access.

Rules:

- scheme must be exactly `https`;
- host must be present;
- path must be present and must not end at only `/`;
- URL must not contain username or password userinfo;
- URL must not contain a query string;
- URL must not contain a fragment;
- URL must not contain control characters or whitespace;
- URL must not be a local `file:` URL;
- URL must not be `http:`, `ssh:`, `git:`, `git+ssh:`, `ftp:`, or any other non-HTTPS scheme;
- URL must not be an SCP-style git target such as `git@example.com:org/repo`;
- URL string length must be bounded to a small implementation constant, recommended maximum `2048` characters.

Redirects are not followed in the first version. Any `3xx` response is a remote fetch error that reports redirects are unsupported.

### Checksum Rules

The checksum field is required and supports only:

```text
sha256:<64 hex characters>
```

Rules:

- algorithm must be exactly `sha256`;
- digest must be exactly 64 hexadecimal characters;
- uppercase hex input may be accepted but must be normalized for comparison and output;
- missing checksum, empty checksum, malformed checksum, or unsupported algorithms fail before network access;
- checksum verification happens over the exact downloaded zip bytes;
- archive parsing never runs when checksum verification fails.

## Supported Remote Format

Support only zip archives in the first implementation. `format: zip` is required even when the URL path ends with `.zip`; no extension or content-type inference is used for source selection.

The zip bundle must contain exactly these five root-level regular file entries:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

Archive rules:

- required files must appear at the archive root, not inside a directory;
- nested paths are rejected;
- absolute paths are rejected;
- path traversal is rejected;
- Windows drive paths are rejected;
- duplicate required file entries are rejected;
- extra file entries are rejected;
- symlink entries are rejected;
- directory entries are rejected except that implementation may ignore harmless root directory metadata if the zip library exposes it separately from file entries;
- executable mode bits are rejected;
- empty or whitespace-only required file contents are rejected;
- archive comments, file comments, and metadata are ignored and never written;
- file contents are interpreted as text bytes and written as strings without executing or rendering scripts.

Recommended size limits:

- maximum HTTP response body: `5 MiB`;
- maximum total uncompressed template content: `1 MiB`;
- maximum uncompressed bytes per required file: `256 KiB`;
- maximum regular file entries: exactly `5`.

These constants are intentionally conservative for Markdown OpenSpec templates. Making them configurable is future scope.

## Checksum And Trust

Remote template trust is based on explicit configuration plus checksum pinning.

Decision:

- No unsigned or unpinned remote templates in the first version.
- No trust-on-first-use.
- No checksum discovery from the remote server.
- No detached signature support in this change.
- No provider, registry, or source-control trust model.

The use case verifies checksums after download and before archive parsing. The checksum model and verification logic are pure core logic over bytes. HTTP fetching remains adapter-owned.

Checksum mismatch must produce a clear error such as:

```text
remote template checksum mismatch for alias api-feature: expected sha256:<expected>, got sha256:<actual>
```

No change directory or file write happens on checksum mismatch.

## Network Behavior

Only the requested config alias can trigger network access.

Network rules:

- built-in template generation never accesses the network;
- custom template generation never accesses the network;
- config-template aliases with `source: builtin` or `source: custom` never access the network;
- config-template aliases with `source: remote` access the network only after config validation, alias resolution, checksum parsing, URL validation, change-id validation, and OpenSpec project structure validation succeed;
- network calls happen only in adapter code;
- core uses a remote-template fetch port;
- no provider APIs are called;
- no source-control APIs are called;
- no git clone is performed;
- no shell command is executed;
- no environment token expansion is performed.

HTTP adapter requirements:

- use only `GET`;
- send no credentials;
- send no auth headers;
- send no cookies;
- send no custom user-provided headers;
- do not expand environment variables into URLs or headers;
- disable redirect following;
- require a total request timeout, recommended default `15s`;
- enforce max response size while reading;
- reject non-`200` responses with a clear error;
- reject redirects with a clear error;
- report DNS, TLS, connection, timeout, and read errors clearly;
- use normal platform TLS verification.

CLI and error output must sanitize remote source information. Since userinfo and query strings are rejected, the preferred success output is the remote URL host only, not the full URL.

## Cache Behavior

No persistent cache is included in the first version.

Every remote generation:

1. fetches the configured HTTPS URL;
2. verifies the checksum;
3. parses the zip;
4. generates the change files.

There is no `.specharbor/cache/templates/` directory, no cache invalidation behavior, no cleanup command, no offline mode, and no cache fallback in this change.

A future cache may be checksum-keyed and safe by construction, but that requires a separate OpenSpec change covering storage, invalidation, pruning, corruption handling, and security review.

## Config Integration

Remote templates integrate only through config-driven template aliases:

```bash
specharbor generate <change-id> --config-template <alias>
```

There is no `--remote-template` command or flag in the first version.

Rationale:

- existing config-template aliases already provide an explicit project-owned indirection layer;
- remote URLs should be reviewed and pinned in config, not typed ad hoc at the command line;
- keeping one command path preserves disjoint direct template modes;
- remote behavior remains opt-in and discoverable in `.specharbor/config.yml`.

Existing source kinds remain valid:

- `source: builtin` with `template`;
- `source: custom` with `template`;
- `source: remote` with `url`, `checksum`, and `format`.

The parser must enforce source-specific required and unsupported fields. A remote alias must not accept `template`, and builtin/custom aliases must not accept `url`, `checksum`, or `format` unless a future spec changes that contract.

## Write Behavior

Remote generation follows existing config-template write behavior:

- output directory is always `openspec/changes/<change-id>/`;
- output filenames are exactly the five required OpenSpec filenames;
- existing files are skipped and reported as skipped;
- existing files are never overwritten;
- downloaded archive filenames never influence output paths;
- archive directory structure never influences output paths;
- all remote fetching, checksum verification, archive parsing, and template validation complete before the change directory is created or any file is written;
- invalid config, invalid URL, missing checksum, checksum mismatch, network failure, timeout, size-limit failure, unsafe archive path, symlink archive entry, extra archive file, missing required file, or empty file produces zero writes.

The command writes no production code, docs, config, CI files, prompt files, archive files, source-control files, or arbitrary output paths.

## Validation Integration

Generated remote-template changes are standard OpenSpec change packages and are compatible with:

```bash
specharbor validate <change-id>
```

Generation does not auto-run validation. This follows existing config-template generation behavior and avoids making remote templates behave differently from built-in or custom aliases.

Documentation should recommend running validation after generation.

## CLI Output

Success output should extend the existing config-template report shape:

```text
SpecHarbor config template change generated.
Change: add-service-endpoint
Config template: service-feature
Resolved source: remote
Remote host: example.com
Remote format: zip
Checksum: sha256
Change path: openspec/changes/add-service-endpoint
Change directory: created
Created files:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
Skipped existing files:
- proposal.md
Safety:
- Remote access used only the explicit configured alias.
- Checksum was verified before archive parsing.
- Only OpenSpec change files under openspec/changes/add-service-endpoint/ were written.
```

If there are no skipped files, omit the skipped section. The report must not display credentials. Because credentials, query strings, and fragments are rejected, the host is sufficient and safer than displaying the full URL.

Required error cases:

- missing URL;
- invalid URL;
- non-HTTPS URL;
- URL with credentials, query string, or fragment;
- redirects unsupported;
- missing checksum;
- malformed checksum;
- unsupported checksum algorithm;
- missing format;
- unsupported format;
- network failure;
- timeout;
- non-`200` response;
- max download size exceeded;
- checksum mismatch;
- unsupported archive format;
- malformed zip archive;
- unsafe archive path;
- absolute archive path;
- nested archive path;
- path traversal archive path;
- duplicate archive file;
- extra archive file;
- executable archive entry;
- symlink archive entry;
- archive size limit exceeded;
- missing required template files;
- empty required template file.

## Architecture

Layer responsibilities follow the existing dependency rule.

### Domain

`internal/core/domain` owns pure concepts:

- `ConfigTemplateSourceRemote` source kind;
- remote template reference model;
- remote template URL value object and pure URL validation;
- remote template format model with only `zip`;
- checksum model and `sha256` parser;
- checksum verification over bytes;
- remote template bundle model holding the five files;
- archive file policy concepts such as allowed filenames and safe root-level path rules;
- remote generation result facts needed by CLI output.

Domain must not perform filesystem access, network access, terminal IO, environment reads, process execution, or adapter calls.

### Use Case

`internal/core/usecase` owns orchestration:

- detect config-template remote aliases after config resolution;
- validate change id and config alias before config and network work;
- load `.specharbor/config.yml` through existing config ports;
- require OpenSpec project structure before network access;
- invoke the remote fetch port only for `source: remote`;
- verify checksum before archive parsing;
- invoke the bundle reader port after checksum verification;
- validate that the bundle provides the five required non-empty files;
- delegate to the same write-if-absent generation flow;
- return structured result fields for remote source, remote host, checksum algorithm, created files, skipped files, and change path.

Use cases must not import adapters, CLI packages, `os`, concrete HTTP packages, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.

### Ports

Add small core-owned ports:

```text
RemoteTemplateFetcher
  Fetch(request) -> downloaded bytes and safe response metadata

RemoteTemplateBundleReader
  ReadZipBundle(bytes, policy) -> remote template bundle files
```

The exact Go names may differ, but the contracts must be consumer-owned and narrow. The fetcher must not expose HTTP client types to core. The bundle reader must not expose `archive/zip` concrete types to core.

Existing filesystem write abstractions remain the only write path.

### Adapters

Adapters own concrete technical behavior:

- HTTP client implementation;
- request timeout;
- redirect disabling;
- max response size enforcement;
- status-code mapping;
- TLS and network errors;
- zip archive decoding;
- symlink and executable mode detection in zip metadata;
- archive size enforcement;
- mapping technical errors to clear use-case errors.

Likely locations are new focused adapter packages such as `internal/adapters/remote` or `internal/adapters/templates`, chosen during implementation to match repository conventions.

### CLI

`internal/adapters/cli` owns:

- argument parsing;
- dependency wiring for the new remote fetcher and bundle reader;
- output formatting;
- sanitized remote-host display;
- exit-code mapping.

The CLI must not contain checksum, URL, archive, or template business rules.

## Documentation Plan

Implementation must update:

- `README.md`;
- `docs/usage.md`;
- `docs/generation-modes.md`;
- any config/template-specific docs present at implementation time.

Documentation must explain:

- the purpose of remote templates;
- the `.specharbor/config.yml` schema for `source: remote`;
- the required `url`, `checksum`, and `format` fields;
- HTTPS-only behavior;
- no redirects;
- no credentials, OAuth, cookies, auth headers, or environment token expansion;
- the required `sha256` checksum and how mismatch behaves;
- the supported zip bundle format;
- root-level five-file archive requirement;
- command usage through `--config-template`;
- generated write path and skip-existing behavior;
- validation guidance;
- no marketplace, discovery, git clone, provider APIs, source-control automation, script execution, shell execution, production code writes, auto-commit, PR, merge, or archive automation.

## Testing Strategy

Testing should be layered:

- Domain tests for remote source model, URL validation, HTTPS-only behavior, non-HTTPS rejection, credential/query/fragment rejection, format validation, checksum parsing, unsupported checksum algorithms, checksum mismatch, source kind validation, and archive path policy concepts.
- Use case tests for remote alias generation, error ordering, missing URL, invalid URL, non-HTTPS URL, missing checksum, checksum mismatch, network failure, timeout, max size exceeded, unsafe archive path, symlink entry rejection, missing required files, empty files, write boundaries, skip-existing behavior, no production code writes, and no network access for builtin/custom aliases.
- Adapter tests for HTTP timeout, redirect rejection, max response size, no credentials/headers/cookies, non-`200` responses, zip traversal rejection, absolute/nested path rejection, symlink rejection, executable entry rejection, extra file rejection, duplicate file rejection, missing files, and checksum-before-parsing behavior at the use-case boundary.
- CLI tests for parsing unchanged `--config-template`, remote success output, sanitized host output, checksum algorithm output, remote error output, and unchanged builtin/custom config-template behavior.
- Regression tests for built-in templates, custom templates, config builtin/custom aliases, AI-assisted generation, validation, prompt, review, archive, config, workflow, scan, init, help, version, and agent-assisted run-and-report.
- Architecture tests for import boundaries and absence of provider/source-control/workflow/agent SDKs, shell execution, script execution, arbitrary output writes, and core HTTP client imports.

## Validation

Implementation verification should include:

```bash
go test ./...
go run ./cmd/specharbor validate implement-remote-templates
```

Manual verification should include successful generation from a local test HTTP server serving a pinned zip, then clear failures for checksum mismatch, redirects, HTTP URLs, oversized responses, unsafe zip paths, symlink entries, extra files, missing files, and existing target files.
