# Proposal: Implement Scan Foundation

## Problem

SpecHarbor can initialize OpenSpec projects, generate blank change packages, validate required change structure, render role-specific prompts, review active changes, and archive completed changes. It still cannot inspect a project and report the development context it is operating in.

Project scanning is a core product capability. Future authoring, generation, and review behavior will benefit from a deterministic understanding of the host project: which language ecosystems, package managers, test tooling, CI systems, container/deployment artifacts, and OpenSpec/SpecHarbor structures are present. SpecHarbor must support real-world projects across many stacks, so scanning must be stack-agnostic.

Future scanning behavior may parse manifests in depth, read source code, detect frameworks, analyze dependency graphs, call AI providers, or feed scan output into generation. The first implementation should establish a deterministic, stack-agnostic scan foundation without mixing those future concerns into the initial command.

## Goal

Implement the first scan capability:

```text
specharbor scan
```

The command scans the current project root and produces a deterministic, human-readable project context report. It detects known signals when present and reports unknown or undetected areas clearly.

The first version detects only top-level or conventional paths:

- project root availability;
- known language/ecosystem signal files;
- package manager signal files;
- test command hints derived from detected ecosystems;
- CI configuration files and directories;
- container/deployment files and directories;
- OpenSpec/SpecHarbor presence.

The implementation must return a structured scan result and print a concise human-readable CLI report. The scanner must not assume the host project is written in Go or any other specific stack.

## Scope

- Replace the `scan` placeholder with real behavior for `specharbor scan`.
- Accept no positional arguments and no flags in this first version.
- Reject any positional argument.
- Reject any flag.
- Obtain the project root from the current working directory in the CLI adapter.
- Validate that the project root is available before scanning.
- Scan only the current project root.
- Detect top-level or conventional paths only.
- Avoid recursive deep scanning. Perform at most a single non-recursive listing of the project root, justified solely by extension-based detection (for example `.csproj` and `.sln`).
- Detect language/ecosystem signals when present:
  - Go: `go.mod`
  - Node/JavaScript/TypeScript: `package.json`, `tsconfig.json`
  - Java: `pom.xml`, `build.gradle`, `build.gradle.kts`, `settings.gradle`, `settings.gradle.kts`
  - Python: `pyproject.toml`, `requirements.txt`, `Pipfile`, `poetry.lock`
  - Rust: `Cargo.toml`
  - .NET: any top-level `*.csproj` or `*.sln`
- Detect package manager signals when present:
  - npm: `package-lock.json`
  - pnpm: `pnpm-lock.yaml`
  - yarn: `yarn.lock`
- Derive deterministic test command hints from detected ecosystems.
- Detect CI configuration when present:
  - GitHub Actions: `.github/workflows/`
  - GitLab CI: `.gitlab-ci.yml`
  - Jenkins: `Jenkinsfile`
- Detect container/deployment artifacts when present:
  - `Dockerfile`
  - `docker-compose.yml`, `docker-compose.yaml`
  - `compose.yml`, `compose.yaml`
  - `kubernetes/`, `k8s/`, `helm/`
- Detect OpenSpec/SpecHarbor signals when present:
  - `openspec/project.md`
  - `openspec/changes/`
  - `openspec/specs/`
  - `.specharbor/config.yml`
  - `.specharbor/rules/`
- Produce deterministic notes for undetected or noteworthy-absent areas.
- Return a structured scan result containing project root, ecosystems, package managers, test command hints, CI providers, container/deployment artifacts, OpenSpec/SpecHarbor signals, and notes.
- Print a human-readable CLI report from the structured scan result, including the empty-result shape.
- Keep scan orchestration in `internal/core/usecase`.
- Keep scan result concepts and deterministic detection rules in `internal/core/domain`.
- Add a small scan-specific filesystem port under `internal/core/ports`.
- Perform concrete filesystem behavior through `internal/adapters/filesystem`.
- Keep detection rules pure and deterministic.
- Replace the unused `internal/core/domain/project_context.go` placeholder with the explicit scan result domain model, or remove it once it is superseded, so the project does not keep a placeholder abstraction.
- Add focused tests for domain detection behavior, the deterministic catalog and assembler, use case orchestration, filesystem adapter compatibility, CLI parsing, CLI reporting, and existing command regressions.

## Out of Scope

- AI-assisted scanning.
- Recursive dependency analysis.
- Recursive directory traversal beyond a single top-level listing.
- Running external commands or tools.
- Parsing full source code.
- Parsing package manager manifests deeply (only presence is detected).
- Detecting every possible framework.
- Generating OpenSpec changes from scan output.
- Writing scan results to files.
- Machine-readable output formats.
- Updating `.specharbor/config.yml`.
- Updating `openspec/project.md`.
- GitHub or GitLab API calls.
- CI modification.
- Docker or Kubernetes validation.
- Security scanning.
- License scanning.
- Vulnerability scanning.
- Network access.
- Package installation.
- Changing init, generate, validate, prompt, review, archive, or config behavior.
- Modifying existing OpenSpec changes.
- Provider API keys, local model credentials, source-control credentials, workflow credentials, or external-agent credentials.
- Public scanner registries, factories, chains, strategy interfaces, plugin abstractions, AI abstractions, provider abstractions, source-control abstractions, or workflow abstractions that are not directly used by this first command.
- Updating the living architecture spec.

## Success Criteria

- Running `specharbor scan` inside a project prints a deterministic project context report and exits zero.
- The report lists detected ecosystems, package managers, test command hints, CI configuration, container/deployment artifacts, and OpenSpec/SpecHarbor signals, plus notes.
- When nothing relevant is detected, the report still prints every section with `- none detected` and a `No known project signals detected.` note.
- The scan is stack-agnostic and does not assume Go, Node, Java, Python, Rust, .NET, or any specific stack.
- Detection only inspects top-level or conventional paths and performs at most one non-recursive listing of the project root.
- The use case returns a structured scan result containing project root, ecosystems, package managers, test command hints, CI providers, container/deployment artifacts, OpenSpec/SpecHarbor signals, and notes.
- The CLI report follows the expected populated and empty output shapes.
- Scan follows Hexagonal Architecture: CLI parses and reports, the use case orchestrates, domain models scan concepts and detection rules, ports define filesystem dependencies, and adapters perform concrete filesystem behavior.
- The implementation does not write files, call AI providers, call external agents, call workflow tools, call source-control host APIs, access the network, run external processes, or install packages.
- Existing `help`, `version`, `init`, `generate`, `prompt`, `validate`, `review`, `archive`, `config`, and unknown command behavior is preserved.
- `go test ./...` succeeds.
