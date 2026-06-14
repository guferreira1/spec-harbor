# Proposal: Implement GitHub Remote Context

## Problem

SpecHarbor has confirmed project briefing, local context discovery, a metadata-only local context index, and deterministic local context retrieval. Users still have no explicit way to inspect bounded context from a GitHub repository without cloning it or leaving the SpecHarbor context workflow.

Remote context is useful when a user wants source-attributed repository evidence from GitHub, but it is risky if it becomes automatic, unbounded, credential-leaking, or confused with local/offline context.

## Goal

Add explicit, bounded, read-only GitHub remote context retrieval:

```text
specharbor context github --repo owner/name --query "<query>"
specharbor context github --repo owner/name --query "<query>" --ref <branch|tag|sha>
specharbor context github --repo owner/name --query "<query>" --path <relative-path>
```

GitHub remote context means explicit retrieval of source-attributed snippets or metadata from a GitHub repository. It is not local context discovery, local repository indexing, RAG answer generation, embeddings, vector storage, provider or LLM behavior, source-control automation, GitHub mutation, prompt execution, or agent execution.

## Command Shape

The command shape is:

```text
specharbor context github --repo owner/name --query "<query>" [--ref <ref>] [--path <relative-path>...]
```

This fits the existing `specharbor context <subcommand>` convention used by `discover`, `index`, and `retrieve`, while the `github` subcommand makes the network and provider boundary explicit. It avoids changing `context retrieve`, which remains local/offline and index-backed.

Existing commands remain local/offline:

- `specharbor context discover`
- `specharbor context index`
- `specharbor context retrieve`
- `specharbor brief`
- `specharbor prompt`

Those commands must not call GitHub.

## Scope

- Add the explicit `context github` CLI path.
- Validate and normalize GitHub repository locators.
- Support `owner/name` repository locators.
- Support optional `https://github.com/<owner>/<repo>` input only if normalized to `owner/name`.
- Support optional `--ref <branch|tag|sha>`.
- Resolve the default branch through GitHub API when `--ref` is omitted.
- Require explicit `--query`.
- Support repeatable `--path <relative-path>` filters as bounded allowlist filters.
- Fetch only the approved source set from GitHub through a read-only port and adapter.
- Support public repositories without a token when GitHub allows it.
- Support an optional token read from `SPECHARBOR_GITHUB_TOKEN`.
- Return concise, bounded, source-attributed remote results.
- Use deterministic local scoring after bounded candidate collection.
- Add focused tests with fake ports/transports.
- Update `README.md`, `docs/usage.md`, and `docs/workflow.md` after implementation.

## Out Of Scope

- GitHub write APIs or mutations.
- Creating commits, branches, PRs, issues, comments, labels, releases, tags, or workflow runs.
- CI inspection beyond read-only file metadata.
- Local `git`, `gh`, shell, package-manager, script, project-command, prompt, or agent execution.
- Cloning repositories.
- Automatic injection of remote context into prompts.
- Writing `.specharbor/context-index.json`.
- Persisting a remote index or cache by default.
- Treating remote context as user-confirmed context.
- RAG answer generation.
- Embeddings.
- Vector databases.
- Provider or LLM APIs.
- GitHub Enterprise, GitLab, Bitbucket, or generic forge support.
- Release automation, npm files, Homebrew files, `install.sh`, GoReleaser files, or publishing behavior.
- Archiving this OpenSpec change before merge.

## Repository Input

The supported first-version locator is:

```text
owner/name
```

Rules:

- owner and repository are required;
- exactly one `/` separates owner and repository;
- reject empty segments;
- reject path traversal;
- reject full filesystem paths;
- reject query strings and fragments;
- reject credentials or userinfo;
- reject unsupported hosts;
- reject whitespace and control characters;
- reject values over 200 characters.

If URL input is accepted, only this host and scheme are allowed:

```text
https://github.com/<owner>/<repo>
```

The URL form must normalize to `owner/name`. GitHub Enterprise is out of scope.

## Ref Behavior

`--ref` is optional. When omitted, SpecHarbor resolves the repository default branch through GitHub API and uses that ref.

Ref rules:

- maximum length: 200 characters;
- reject empty refs;
- reject null bytes;
- reject traversal segments;
- reject leading or trailing `/`;
- reject `//`;
- reject URL-like refs and credentials;
- reject query strings and fragments;
- allow branch names, tag names, and commit SHAs that satisfy the rules;
- output includes the requested ref when provided and the resolved commit SHA when available.

## Query And Path Behavior

`--query` is required.

Query rules:

- trim whitespace;
- reject empty queries;
- maximum length: 512 characters;
- lower-case and split terms on non-letter and non-digit boundaries;
- keep at most 32 unique terms in first-seen order;
- reject queries with no usable terms;
- do not execute regex, shell, glob, provider, or remote semantic queries.

`--path` is optional and repeatable. It narrows the approved remote source set; it never expands beyond that source set.

Path rules:

- repository-relative paths only;
- maximum length: 512 characters;
- reject traversal;
- reject absolute paths;
- reject Windows drive paths;
- reject null bytes;
- reject query strings and fragments;
- reject wildcard expansion;
- normalize `./` and duplicate slashes;
- skip sensitive and generated paths.

## Source Set

The first version fetches only bounded supported sources:

- `README.md`
- `AGENTS.md`
- `CONTRIBUTING.md`
- Markdown files under `docs/` within bounded traversal
- `openspec/project.md`
- Markdown files under `openspec/specs/` within bounded traversal
- Markdown files under `.specharbor/rules/`
- `.specharbor/project-brief.md`
- `package.json`
- `go.mod`
- `pom.xml`
- `build.gradle`
- `build.gradle.kts`
- `Cargo.toml`
- `pyproject.toml`
- `requirements.txt`
- `Dockerfile`
- `docker-compose.yml`
- `docker-compose.yaml`
- `Makefile`
- `Taskfile.yml`
- `Taskfile.yaml`
- `.github/workflows/*.yml`
- `.github/workflows/*.yaml`

Do not fetch the entire repository. Do not fetch arbitrary large trees, binary files, generated folders, or sensitive files.

## Sensitive And Generated Skipping

Remote context must skip at least these sensitive files:

- `.env`
- `.env.*`
- `*.pem`
- `*.key`
- `id_rsa`
- `id_ed25519`
- `secrets.*`
- `credentials.*`
- `.npmrc`
- `.pypirc`
- `.netrc`

Remote context must skip at least these generated or heavy folders:

- `.git/`
- `node_modules/`
- `dist/`
- `build/`
- `target/`
- `vendor/`
- `coverage/`
- `.tmp/`
- `.cache/`
- `.next/`
- `.nuxt/`
- `out/`
- `bin/`
- `obj/`

## Network And Authentication

Only HTTPS requests to `api.github.com` are in scope.

Authentication is optional. SpecHarbor reads `SPECHARBOR_GITHUB_TOKEN` when present because it is project-scoped and avoids implicitly consuming a broadly shared `GITHUB_TOKEN`. The token must never be printed, persisted, included in errors, included in reports, or used outside the GitHub read-only HTTP adapter.

Network behavior:

- bounded HTTP timeout: 10 seconds by default;
- bounded response bodies;
- bounded files, tree entries, and total bytes;
- no background work;
- no unbounded retries;
- no redirects to unsupported hosts;
- clear safe errors for offline/network failure, rate limit, unauthorized, forbidden, not found, invalid token, oversized file, unsupported file type, and too many candidates.

## Remote Retrieval Model

This change implements remote retrieval only.

It fetches bounded remote candidate metadata and file contents, extracts bounded snippets or metadata summaries, scores them locally, and renders a report. It does not persist a remote index, does not cache by default, does not write `.specharbor/context-index.json`, and does not mix remote results into local indexes or prompts.

## Output

Output must be concise and source-attributed. Each result includes:

- repository owner/name;
- requested ref when provided;
- default branch when used;
- resolved commit SHA when available;
- path;
- source category or evidence category;
- score and rank;
- line range when practical;
- bounded snippet or metadata summary;
- `Remote: yes`.

Output must not include tokens, credentials, credential-bearing URLs, huge file contents, binary contents, generated contents, or sensitive file contents.

## Compatibility

Preserve behavior for:

- `specharbor context discover`
- `specharbor context index`
- `specharbor context retrieve`
- `specharbor brief`
- `specharbor brief --update`
- context-aware `specharbor prompt`
- `specharbor scan`
- release, npm, Homebrew, `install.sh`, and GoReleaser behavior

## Success Criteria

- `specharbor context github --repo owner/name --query "<query>"` returns bounded GitHub remote context.
- Local/offline commands remain local and do not require network or tokens.
- Repository, ref, query, and path validation reject unsafe inputs.
- GitHub access is read-only, HTTPS-only, token-safe, and bounded.
- Results are deterministic, source-attributed, and marked remote.
- Sensitive/generated paths, oversized files, binary files, and unsupported sources are skipped safely.
- Ranking is deterministic and does not use embeddings, vectors, LLM reranking, provider APIs, or RAG.
- Tests use fake adapters/transports and require no real GitHub network or token.
- OpenSpec validation and `go test ./...` pass.
