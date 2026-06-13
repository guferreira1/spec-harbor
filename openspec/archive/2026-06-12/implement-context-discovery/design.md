# Design: Implement Context Discovery

## Overview

Context discovery adds a deterministic local use case that inspects common repository files and produces structured context signals.

The first command surface should be:

```text
specharbor context discover
```

This keeps richer context detection separate from the existing `specharbor scan` presence report. `scan` should remain stack-agnostic and shallow. `context discover` can read selected project files, parse common manifests, and report stronger context signals with source evidence, classification, and confidence.

`specharbor brief` may call the same discovery use case before prompting so detected values can appear as suggestions. This integration must preserve the existing briefing contract: discovered values are not confirmed until the user selects or enters them, and `.specharbor/project-brief.md` is still written only after confirmation and only when absent.

## Command Contract

Supported command:

```text
specharbor context discover
```

Rejected in this first version:

```text
specharbor context
specharbor context discover <anything>
specharbor context discover --json
specharbor context discover --path <path>
specharbor context discover --deep
specharbor context discover --github
specharbor context discover --rag
specharbor context update
specharbor context index
```

The CLI adapter obtains the project root from the current working directory, constructs the discovery use case with the local filesystem adapter, and formats the returned structured result.

The command is read-only. A repository with no supported context files is a successful command with empty or ambiguous context sections and explanatory notes.

## Expected CLI Output

The report should be deterministic and readable:

```text
Detected project context:

User-confirmed context:
- none detected

Detected facts:
- Stack: Go (confidence: high, source: go.mod)
- CLI entrypoint: cmd/specharbor (confidence: medium, source: repository layout)
- Architecture: Hexagonal Architecture (confidence: high, source: openspec/specs/architecture/spec.md)
- Agent rules: .specharbor/rules/ (confidence: high, source: .specharbor/rules/)

Suggested assumptions:
- Test command: go test ./... (confidence: medium, source: go.mod convention)
```

If `.specharbor/project-brief.md` exists, confirmed values must be grouped first:

```text
User-confirmed context:
- Stack: Go (confidence: high, source: .specharbor/project-brief.md)
```

Output rules:

- Always print grouped sections in this order: user-confirmed context, detected facts, suggested assumptions, notes.
- Print `- none detected` for empty groups.
- Keep ordering deterministic by signal category, then source path, then display value.
- Do not print raw file contents, secrets, debug output, network metadata, source-control state, or absolute paths other than the project root if the existing CLI style requires it.

## Signal Model

Add context discovery domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- `ContextSignal`
- `ContextSignalKind`
- `ContextSignalClassification`
- `ContextConfidence`
- `ContextSource`
- `ContextSourceCategory`
- `ContextDiscoveryResult`
- `ContextDiscoveryNote`

Recommended signal kinds:

- `project_type`
- `purpose_summary`
- `stack`
- `language`
- `framework`
- `architecture_hint`
- `package_manager`
- `test_command`
- `build_command`
- `run_command`
- `documentation_source`
- `agent_instruction_source`
- `openspec_source`
- `cli_entrypoint`
- `container_signal`
- `workflow_signal`

Required classifications:

- `detected_fact`: explicit repository evidence supports the value.
- `suggested_assumption`: repository evidence supports a possible suggestion, but not a confirmed fact.
- `user_confirmed_context`: `.specharbor/project-brief.md` contains confirmed project context.

Required confidence levels:

- `high`: direct, explicit evidence in a supported source.
- `medium`: explicit evidence exists, but it is indirect, conventional, or less specific.
- `low`: weak evidence that may be useful as a prompt suggestion only.

Confidence is not a replacement for classification. A high-confidence detected fact is still not user-confirmed context. A low-confidence assumption must never be rendered as a fact.

## Source Categories

Each signal must retain source evidence with at least:

- relative path;
- source category;
- optional evidence label such as a manifest field, script key, heading, or rule name.

Recommended source categories:

- `project_brief`
- `readme`
- `contributing`
- `documentation`
- `agent_instruction`
- `openspec_project`
- `openspec_spec`
- `specharbor_rules`
- `package_manifest`
- `build_manifest`
- `dependency_manifest`
- `task_runner`
- `container_config`
- `workflow_config`
- `repository_layout`

Domain constructors should defensively copy slices and validate that every signal has a non-empty kind, value, classification, confidence level, and source.

## Discovery Sources

The implementation must inspect only the supported local sources when present:

| Source | Expected signal use |
| --- | --- |
| `AGENTS.md` | agent instruction source, architecture hints, command guidance when explicit |
| `README.md` | purpose summary, project type, documentation source, commands when explicit |
| `CONTRIBUTING.md` | documentation source, test/build guidance when explicit |
| `docs/` | documentation source and explicit command or architecture guidance from bounded Markdown files |
| `openspec/project.md` | purpose, stack, architecture, agent rules, OpenSpec source |
| `openspec/specs/` | architecture hints and OpenSpec source |
| `.specharbor/rules/` | agent instruction source |
| `.specharbor/project-brief.md` | user-confirmed context with precedence |
| `package.json` | Node stack, package manager field, scripts for test/build/run |
| `go.mod` | Go stack, module evidence |
| `pom.xml` | Java/Maven stack and build manifest |
| `build.gradle` | JVM/Gradle stack and build manifest |
| `Cargo.toml` | Rust stack and Cargo manifest |
| `pyproject.toml` | Python stack and tool/build metadata when explicit |
| `requirements.txt` | Python dependency manifest |
| `Dockerfile` | container signal |
| `docker-compose.yml` | container/run signal when explicit enough |
| `Makefile` | explicit test/build/run targets |
| `Taskfile.yml` | explicit test/build/run tasks |
| `.github/workflows/` | workflow signal and explicit test/build commands only when safely readable from workflow YAML |

The implementation may read a bounded set of Markdown files under `docs/` and `openspec/specs/`, but it must not perform repository-wide indexing or recursive semantic analysis. It should use stable traversal ordering and reasonable file-size limits to keep output deterministic and fast.

## Fact And Assumption Rules

Detected facts must come from explicit repository evidence.

Examples of `detected_fact`:

- `go.mod` exists, so language/stack includes Go.
- `package.json` has a `scripts.test` value, so that exact script command can be reported as a test command fact.
- `openspec/specs/architecture/spec.md` states Hexagonal Architecture, so architecture can be reported from that source.
- `.specharbor/rules/` exists, so agent instruction sources include that directory.
- `.specharbor/project-brief.md` has a confirmed Stack section, so the value is `user_confirmed_context`.

Examples of `suggested_assumption`:

- `go.mod` exists and no explicit test command is found, so `go test ./...` may be suggested as a conventional Go test command.
- `pom.xml` exists and no explicit command is found, so `mvn test` may be suggested as a Maven convention.
- `package.json` exists without lockfiles or a `packageManager` field, so npm may be suggested as a package-manager assumption.

Rules:

- Suggested assumptions must be labeled as assumptions and must include their evidence source.
- Suggested assumptions must never be passed to project brief creation as confirmed answers unless the user confirms them through `specharbor brief`.
- Discovery must not invent stack, architecture, commands, frameworks, or project decisions without at least one supported source.
- Ambiguous signals should either produce multiple detected facts with sources or a note explaining ambiguity. They should not collapse into a single unsupported decision.

## Project Brief Precedence

When `.specharbor/project-brief.md` exists:

- parse the supported brief sections conservatively;
- keep parsing bounded to the known project brief headings and file-size limits;
- emit parsed values as `user_confirmed_context`;
- assign high confidence to successfully parsed confirmed sections;
- prefer confirmed values over conflicting detected facts in summaries;
- keep conflicting detected facts visible as detected facts with their own sources, unless doing so would make output misleading;
- do not modify, merge, update, overwrite, or append to the existing brief.

When `specharbor brief` uses discovery:

- confirmed project brief data should not trigger an update flow in this change;
- existing brief write-if-absent refusal remains unchanged;
- detected facts and assumptions may appear as prompt suggestions only when no existing brief will be overwritten;
- the final saved brief must still contain user-confirmed answers as `user_provided` answers according to the existing brief model.

## Security And Ignore Rules

Discovery must skip obvious sensitive files even if they are otherwise reachable:

- `.env`
- `.env.*`
- `*.pem`
- `*.key`
- `id_rsa`
- `id_ed25519`
- `secrets.*`
- `credentials.*`

Discovery must skip heavy or generated folders:

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

The filesystem adapter should enforce path safety, avoid following symlinks, and avoid returning file contents for skipped paths. The domain should own the skip policy catalog so it can be tested deterministically without the concrete filesystem.

## Architecture

The implementation must preserve the existing Hexagonal Architecture boundary:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

Responsibilities:

- Domain owns context signal models, classifications, source categories, confidence levels, source evidence, skip policy, and deterministic ordering.
- Ports expose only the filesystem operations the discovery use case consumes.
- Use cases orchestrate source collection, detector execution, project brief precedence, and structured result assembly.
- Classification and confidence decisions must come from small deterministic core/domain detector rules, never from CLI formatting.
- Adapters perform concrete filesystem traversal and file reading through safe local path handling.
- CLI parses `context discover`, invokes the use case, and formats the returned structured result.

Core packages must not import adapters, CLI packages, `os`, terminal IO, network APIs, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.

## Ports And Adapters

Add a small discovery filesystem port under:

```text
internal/core/ports
```

Expected operations:

```text
FileExists(root string, relativePath string) (bool, error)
DirectoryExists(root string, relativePath string) (bool, error)
ListDirectory(root string, relativePath string) ([]DirectoryEntry, error)
ReadFile(root string, relativePath string) (string, error)
```

The exact names may differ, but the port must be owned by the discovery use case and must expose only what deterministic local discovery needs.

Concrete traversal and reading belongs in `internal/adapters/filesystem`. The adapter should reuse existing safe path handling where possible.

Do not add provider, RAG, vector-store, GitHub, GitLab, workflow connector, agent runner, command execution, or prompt injection ports for this change.

## Use Case

Add context discovery orchestration under:

```text
internal/core/usecase
```

Expected behavior:

- validate required dependencies;
- validate project root is non-empty;
- read only supported source files and directories;
- apply skip rules before reads;
- build a deterministic source snapshot;
- run small deterministic detectors;
- classify each signal as `detected_fact`, `suggested_assumption`, or `user_confirmed_context`;
- assign confidence and source evidence to every signal;
- prefer project brief data when present;
- return a structured `ContextDiscoveryResult`;
- never print output;
- never call `os` directly;
- never call provider APIs, network APIs, source-control APIs, workflow tools, agent tools, package managers, shell commands, or repository indexing.

Discovery rules should be small, deterministic, and testable. A simple ordered detector list is sufficient. Avoid large framework dependencies or placeholder plugin registries.

## CLI Adapter

Update `internal/adapters/cli` to register a `context` command with the supported `discover` subcommand.

The CLI adapter should:

- reject missing subcommands, unsupported subcommands, positional arguments after `discover`, and all flags;
- obtain the current working directory as project root;
- instantiate the discovery use case with filesystem dependencies;
- print a deterministic human-readable report;
- return clear errors for filesystem failures;
- keep all discovery rules, classifications, source categories, confidence decisions, and skip policy out of CLI code.

## Brief Integration

`specharbor brief` may use context discovery to enrich the existing question flow:

- run discovery before prompting, after TTY validation;
- show detected facts and assumptions as suggestions, not confirmed answers;
- keep the final option as `Other / custom`;
- keep each question within the existing three-to-five option rule;
- require user selection or custom input before a value becomes a confirmed answer;
- include detected context sources in the generated project brief only as detected context records;
- render assumptions separately from confirmed answers;
- preserve cancellation behavior and write-if-absent behavior.

If discovery fails because supported context is missing or ambiguous, `brief` should continue with the existing interactive questions whenever possible. It should not block abruptly solely because no context was discovered.

## Documentation Plan

The implementation change should update public documentation after code exists:

- `README.md`: add `specharbor context discover` to implemented commands and clarify how it differs from `scan` and `brief`.
- `docs/usage.md`: document command syntax, output groups, classifications, confidence levels, supported sources, skip rules, and safety boundaries.
- `docs/workflow.md`: mention context discovery as an optional preparation step before briefing or authoring.
- `docs/agent-roles.md` or generation docs only if needed to state that prompt injection remains out of scope.

Documentation must not claim RAG, indexing, prompt injection, external APIs, remote discovery, command execution, or brief update behavior.

## Testing Strategy

Add focused tests for:

- context signal construction and validation;
- classification validation for `detected_fact`, `suggested_assumption`, and `user_confirmed_context`;
- confidence-level validation and deterministic ordering;
- source category validation;
- detecting Go project context from `go.mod`;
- detecting Node project context from `package.json`;
- detecting Java project context from `pom.xml` or `build.gradle`;
- detecting Python project context from `pyproject.toml` or `requirements.txt`;
- detecting Rust project context from `Cargo.toml`;
- detecting explicit commands from `package.json` scripts;
- detecting explicit commands from `Makefile`;
- detecting explicit commands from `Taskfile.yml`;
- detecting documentation sources from `README.md`, `CONTRIBUTING.md`, and `docs/`;
- detecting agent instruction sources from `AGENTS.md` and `.specharbor/rules/`;
- detecting OpenSpec sources from `openspec/project.md` and `openspec/specs/`;
- skipping sensitive files;
- skipping generated and heavy folders;
- avoiding symlink traversal where existing filesystem safety patterns apply;
- deterministic output for the same source snapshot;
- project brief precedence when `.specharbor/project-brief.md` exists;
- use case behavior when context is missing;
- use case behavior when context is ambiguous;
- CLI report formatting for populated and empty results;
- `specharbor brief` suggestion behavior with detected facts;
- `specharbor brief` behavior with ambiguous or missing context;
- regression coverage for existing `specharbor brief` behavior;
- regression coverage for existing CLI commands.

Verification after implementation:

```text
gofmt
go test ./...
go run ./cmd/specharbor validate implement-context-discovery
```
