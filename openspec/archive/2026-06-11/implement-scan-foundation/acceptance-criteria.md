# Acceptance Criteria: Implement Scan Foundation

- Running `specharbor scan` inside a project returns a structured scan result.
- The structured result contains the project root, detected ecosystems, package managers, test command hints, CI providers, container/deployment artifacts, OpenSpec/SpecHarbor signals, and notes.
- The scan is stack-agnostic and does not assume Go, Node, Java, Python, Rust, .NET, or any specific stack.
- The scan inspects only top-level or conventional paths.
- The scan performs at most a single non-recursive listing of the project root and never recurses into subdirectories.
- The project root is obtained from the current working directory in the CLI adapter.
- The project root is validated before scanning, and an unavailable or unreadable root returns an execution error.
- Detection is driven by a single deterministic, ordered catalog of signal rules in the domain.
- Each catalog rule has a category, a human-readable name, a relative path or suffix, a probe kind, and an optional test command hint.
- `file` rules are detected through file existence checks.
- `directory` rules are detected through directory existence checks.
- `file-suffix` rules are detected by matching top-level entry names against the suffix using the single root listing.
- Ecosystem detection covers Go (`go.mod`), Node (`package.json`, `tsconfig.json`), Java (`pom.xml`, `build.gradle`, `build.gradle.kts`, `settings.gradle`, `settings.gradle.kts`), Python (`pyproject.toml`, `requirements.txt`, `Pipfile`, `poetry.lock`), Rust (`Cargo.toml`), and .NET (`.csproj`, `.sln`).
- Package manager detection covers npm (`package-lock.json`), pnpm (`pnpm-lock.yaml`), and yarn (`yarn.lock`).
- CI detection covers GitHub Actions (`.github/workflows/`), GitLab CI (`.gitlab-ci.yml`), and Jenkins (`Jenkinsfile`).
- Container/deployment detection covers `Dockerfile`, `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`, `kubernetes/`, `k8s/`, and `helm/`.
- OpenSpec/SpecHarbor detection covers `openspec/project.md`, `openspec/changes/`, `openspec/specs/`, `.specharbor/config.yml`, and `.specharbor/rules/`.
- The detection catalog is separate from `domain.RequiredOpenSpecChangeFiles()` and does not duplicate or reuse the required-change-file policy.
- `directory` signals are displayed with a trailing slash, `file` signals are displayed as the path, and `file-suffix` signals are displayed as the suffix pattern.
- Ecosystem, package manager, and CI detections display as `- {Name}: {Signal}`.
- Container/deployment and OpenSpec/SpecHarbor detections display as `- {Signal}` with no name.
- Test command hints are derived deterministically from matched ecosystem rules.
- Test command hints are de-duplicated and preserve first-seen order.
- Test command hints are never executed.
- The scan never runs external commands, tools, or processes.
- The note `No known project signals detected.` is produced when no rules match.
- The note `No Kubernetes manifests detected.` is produced when at least one container/deployment signal is detected but no `kubernetes/`, `k8s/`, or `helm/` directory signal is present.
- No Kubernetes note is produced when a Kubernetes or Helm directory signal is present.
- The CLI prints a populated scan report following this output shape:

```text
SpecHarbor project scan completed.
Project root: /path/to/project

Detected ecosystems:
- Go: go.mod
- Node: package.json

Package managers:
- npm: package-lock.json

Test command hints:
- go test ./...
- npm test

CI:
- GitHub Actions: .github/workflows/

Containers/deployment:
- Dockerfile
- docker-compose.yml

SpecHarbor/OpenSpec:
- openspec/project.md
- openspec/changes/
- .specharbor/config.yml

Notes:
- No Kubernetes manifests detected.
```

- The CLI prints an empty scan report following this output shape:

```text
SpecHarbor project scan completed.
Project root: /path/to/project

Detected ecosystems:
- none detected

Package managers:
- none detected

Test command hints:
- none detected

CI:
- none detected

Containers/deployment:
- none detected

SpecHarbor/OpenSpec:
- none detected

Notes:
- No known project signals detected.
```

- All seven sections are always printed in order: Detected ecosystems, Package managers, Test command hints, CI, Containers/deployment, SpecHarbor/OpenSpec, Notes.
- The project root line and each section are separated by a single blank line.
- Any section with no entries prints a single `- none detected` line, including the Notes section.
- The `Project root:` line is the only absolute local path in the report and is required.
- Scan exits zero when the report is printed successfully.
- Scan exits non-zero only on execution errors that prevent scanning, such as an unavailable or unreadable project root, a current-working-directory failure, or a filesystem failure.
- Running `specharbor scan` with any positional argument returns a clear argument error.
- Running `specharbor scan` with any flag returns a clear argument error.
- Filesystem checks and listing are performed through a scan-specific port owned by `internal/core/ports`.
- Concrete filesystem behavior lives in `internal/adapters/filesystem`.
- Scan orchestration lives in `internal/core/usecase`.
- Scan result concepts, the detection catalog, and the pure assembler live in `internal/core/domain`.
- The CLI adapter handles argument parsing, current-working-directory lookup, dependency construction, and human-readable report formatting only.
- Detection catalog data, categorization, signal display policy, test command hint derivation, and note computation do not live in the CLI adapter or filesystem adapter.
- The unused `internal/core/domain/project_context.go` placeholder is replaced by the scan result domain model or removed; no unused placeholder abstraction remains.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.
- The domain package does not import ports.
- The implementation does not write files.
- The implementation does not parse manifests deeply, parse source code, or detect frameworks.
- The implementation does not perform recursive dependency analysis or recursive traversal.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, or source-control host APIs.
- The implementation does not access the network or install packages.
- The implementation does not perform security, license, or vulnerability scanning.
- The implementation does not generate OpenSpec changes from scan output.
- The implementation does not update `.specharbor/config.yml` or `openspec/project.md`.
- The implementation does not change init, generate, validate, prompt, review, archive, or config behavior.
- The implementation does not modify existing OpenSpec changes.
- The implementation does not require provider API keys, local model credentials, agent credentials, source-control credentials, or workflow credentials.
- The implementation does not add unused scanner registries, factories, chains, strategy interfaces, plugin abstractions, provider abstractions, AI abstractions, agent abstractions, or workflow abstractions.
- Existing `help`, `version`, `init`, `generate`, `prompt`, `validate`, `review`, `archive`, `config`, and unknown command behavior is preserved.
- Focused tests cover domain detection behavior, the catalog and assembler, use case orchestration, filesystem adapter compatibility, CLI behavior, CLI exit behavior, and existing command regressions.
- Tests do not depend on the real host stack of the SpecHarbor repository.
- `go test ./...` succeeds.
