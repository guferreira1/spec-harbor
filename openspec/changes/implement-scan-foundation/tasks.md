# Tasks: Implement Scan Foundation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-scan-foundation/`.
- [x] Inspect the current CLI command registry and error flow before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing domain, use case, port, filesystem adapter, and test patterns before editing.
- [x] Inspect the existing review and archive implementations for report formatting, dependency wiring, and filesystem port patterns.
- [x] Inspect the unused `internal/core/domain/project_context.go` placeholder and confirm it has no references.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to `specharbor scan` and the deterministic, stack-agnostic scan foundation described by this change.
- [x] Do not assume the host project is written in Go, Node, Java, Python, Rust, .NET, or any specific stack.
- [x] Do not implement recursive traversal, deep manifest parsing, source-code parsing, framework detection, AI-assisted scanning, dependency analysis, external command execution, network access, package installation, file writing, machine-readable output, or changes to init, generate, validate, prompt, review, archive, or config behavior.

## Phase 1: Domain Concepts

- [x] Add scan domain types under `internal/core/domain`.
- [x] Represent a detection value with a human-readable name and a display signal.
- [x] Represent a structured scan result containing project root, ecosystems, package managers, test command hints, CI providers, container/deployment artifacts, OpenSpec/SpecHarbor signals, and notes.
- [x] Add a result constructor that defensively copies slices, following the convention used by `domain.NewReviewResult` and `domain.NewGenerationResult`.
- [x] Replace the unused `internal/core/domain/project_context.go` placeholder with the explicit scan result model, or remove it once superseded, so no placeholder abstraction remains.
- [x] Keep domain code free of adapters, CLI packages, ports, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, and external process execution.

## Phase 2: Ports

- [x] Add a small scan-specific filesystem port under `internal/core/ports`.
- [x] Include only the operations needed by this change: file existence, directory existence, and non-recursive listing of immediate directory entry names.
- [x] Use a behavior-specific name such as `ScanFileSystem`.
- [x] Keep the scan filesystem port separate from initialization, validation, generation, archive, review, prompt rendering, AI provider, workflow dispatcher, source-control, and agent contracts.
- [x] Ensure the scan use case depends on the scan filesystem port instead of `internal/adapters/filesystem`.

## Phase 3: Scan Detection Rules

- [x] Add the deterministic detection catalog in the domain as pure, ordered data.
- [x] Represent each rule with a category, a human-readable name, a relative path or suffix, a probe kind (`file`, `directory`, or `file-suffix`), and an optional test command hint.
- [x] Populate ecosystem rules for Go (`go.mod`), Node (`package.json`, `tsconfig.json`), Java (`pom.xml`, `build.gradle`, `build.gradle.kts`, `settings.gradle`, `settings.gradle.kts`), Python (`pyproject.toml`, `requirements.txt`, `Pipfile`, `poetry.lock`), Rust (`Cargo.toml`), and .NET (`.csproj`, `.sln` as file-suffix rules).
- [x] Populate package manager rules for npm (`package-lock.json`), pnpm (`pnpm-lock.yaml`), and yarn (`yarn.lock`).
- [x] Populate CI rules for GitHub Actions (`.github/workflows` directory), GitLab CI (`.gitlab-ci.yml`), and Jenkins (`Jenkinsfile`).
- [x] Populate container/deployment rules for `Dockerfile`, `docker-compose.yml`, `docker-compose.yaml`, `compose.yml`, `compose.yaml`, `kubernetes` directory, `k8s` directory, and `helm` directory, with empty names.
- [x] Populate OpenSpec/SpecHarbor rules for `openspec/project.md`, `openspec/changes` directory, `openspec/specs` directory, `.specharbor/config.yml`, and `.specharbor/rules` directory, with empty names.
- [x] Attach deterministic test command hints to the relevant ecosystem rules (for example `go test ./...`, `npm test`, `mvn test`, `gradle test`, `pytest`, `cargo test`, `dotnet test`).
- [x] Keep the detection catalog separate from `domain.RequiredOpenSpecChangeFiles()`; the catalog reports presence and must not duplicate or reuse the required-change-file policy.
- [x] Add a pure assembler that turns the project root and the matched rules into a scan result.
- [x] Build the display signal in the assembler: `file` rules use the path, `directory` rules use the path with a trailing slash, and `file-suffix` rules use the suffix pattern.
- [x] Group matched rules into the correct result fields in catalog order.
- [x] Collect test command hints from matched ecosystem rules and de-duplicate them preserving first-seen order.
- [x] Emit the note `No known project signals detected.` when no rules match.
- [x] Emit the note `No Kubernetes manifests detected.` when at least one container/deployment signal is detected but no `kubernetes`, `k8s`, or `helm` directory signal is present.
- [x] Keep the catalog, signal display, hint derivation, and note rules pure and deterministic, with no filesystem access, terminal IO, AI calls, network access, or external libraries.

## Phase 4: Scan Use Case

- [x] Add the scan use case under `internal/core/usecase`.
- [x] Make the use case accept the project root as input.
- [x] Validate that the use case dependency is present.
- [x] Validate that the filesystem dependency is present.
- [x] Validate that the project root is non-empty after trimming whitespace.
- [x] List the top-level entries of the project root through the scan filesystem port to validate root availability and to support suffix rules.
- [x] Return an execution error when the project root cannot be listed (unavailable or unreadable root).
- [x] Iterate `domain.ScanSignalRules()` in catalog order.
- [x] Probe `file` rules with `FileExists`, `directory` rules with `DirectoryExists`, and `file-suffix` rules against the already-listed top-level entry names.
- [x] Return an execution error when a filesystem existence check fails.
- [x] Collect the matched rules and call the pure domain assembler with the project root and matched rules.
- [x] Return the structured scan result.
- [x] Detect only top-level or conventional catalog paths; do not recurse and do not list any directory other than the project root.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, provider SDKs, network APIs, source-control APIs, external agents, external processes, or workflow tools; do not write files.
- [x] Keep the structure ready for future scan collaborators without adding unused exported scanner registries, factories, chains, strategy interfaces, plugin abstractions, provider abstractions, AI abstractions, agent abstractions, or workflow abstractions.

## Phase 5: Filesystem Adapter

- [x] Use `internal/adapters/filesystem` as the concrete implementation of the scan filesystem port.
- [x] Add a non-recursive method that lists the immediate entry names of a directory.
- [x] Ensure the listing method returns an error when the directory does not exist or cannot be read so the use case can treat an unavailable root as an execution error.
- [x] Reuse the existing directory and file existence checks that distinguish directories from files.
- [x] Do not recurse into subdirectories in the listing method.
- [x] Add or update adapter tests for non-recursive listing behavior and the missing-directory error.
- [x] Ensure the filesystem adapter does not contain detection catalog data, categorization, signal display policy, test command hint derivation, note computation, CLI report formatting, provider logic, source-control logic, workflow logic, or framework detection.

## Phase 6: CLI Wiring and Reporting

- [x] Replace the `scan` placeholder in `internal/adapters/cli/cli.go` command registry.
- [x] Parse `specharbor scan` with no positional arguments and no flags.
- [x] Reject any positional argument.
- [x] Reject any flag.
- [x] Obtain the current working directory for the project root.
- [x] Construct the scan use case with the local filesystem adapter.
- [x] Invoke the use case with the project root.
- [x] Print a human-readable scan report from the structured result.
- [x] Print the completion line, the project root line, and all seven sections in order: Detected ecosystems, Package managers, Test command hints, CI, Containers/deployment, SpecHarbor/OpenSpec, Notes.
- [x] Separate the project root line and each section with a single blank line.
- [x] Print `- {Name}: {Signal}` for detections with a non-empty name and `- {Signal}` for detections with an empty name.
- [x] Print `- none detected` under any section that has no entries, including Notes.
- [x] Match the expected populated and empty report shapes from `design.md`.
- [x] Return zero when the report is printed successfully.
- [x] Return argument and execution errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping.
- [x] Preserve existing `help`, `version`, `init`, `generate`, `prompt`, `validate`, `review`, `archive`, `config`, and unknown command behavior.
- [x] Keep the CLI adapter free of detection catalog data, categorization, signal display policy, test command hint derivation, note computation, filesystem policy, provider logic, source-control logic, and workflow logic.

## Phase 7: Tests

- [x] Add domain tests proving the detection catalog is non-empty and covers ecosystems, package managers, CI, container/deployment, and OpenSpec/SpecHarbor categories.
- [x] Add domain tests proving the assembler groups matched rules into the correct result fields in catalog order.
- [x] Add domain tests proving `directory` signals render with a trailing slash, `file` signals render as the path, and `file-suffix` signals render as the suffix pattern.
- [x] Add domain tests proving ecosystems, package managers, and CI detections carry non-empty names while container/deployment and OpenSpec/SpecHarbor detections carry empty names.
- [x] Add domain tests proving test command hints are collected and de-duplicated in first-seen order.
- [x] Add domain tests proving `No known project signals detected.` is emitted when no rules match.
- [x] Add domain tests proving `No Kubernetes manifests detected.` is emitted when container/deployment signals exist without a Kubernetes or Helm directory.
- [x] Add domain tests proving no Kubernetes note is emitted when a Kubernetes or Helm directory signal is present.
- [x] Add domain tests proving the scan result constructor defensively copies slices.
- [x] Add use case tests with a fake scan filesystem for a populated project with Go, Node, npm, GitHub Actions, Docker, and OpenSpec signals.
- [x] Add use case tests for an empty project returning an empty result with the no-signals note.
- [x] Add use case tests proving `.csproj` and `.sln` are detected through the top-level listing.
- [x] Add use case tests proving an empty project root is rejected.
- [x] Add use case tests proving an unlistable project root returns an execution error.
- [x] Add use case tests proving a filesystem existence-check failure returns an execution error.
- [x] Add use case tests proving only the project root is listed and no recursion occurs.
- [x] Add filesystem adapter tests proving the local adapter satisfies the scan filesystem port.
- [x] Add filesystem adapter tests proving non-recursive listing of immediate entry names.
- [x] Add filesystem adapter tests proving a missing-directory listing returns an error.
- [x] Add CLI tests for a populated scan report matching the expected shape.
- [x] Add CLI tests for the empty scan report matching the expected shape, including `- none detected` under every section.
- [x] Add CLI tests proving positional arguments and flags are rejected.
- [x] Add CLI tests proving scan returns zero when the report is printed.
- [x] Ensure CLI tests build a temporary project directory and do not depend on the real host stack of the SpecHarbor repository.
- [x] Preserve or add regression coverage for `help`, `version`, `init`, `generate`, `prompt`, `validate`, `review`, `archive`, `config`, and unknown command behavior.

## Phase 8: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor scan` prints a populated report and exits zero inside a project with known signals.
- [x] Manually verify `specharbor scan` prints the empty report with `- none detected` sections and the no-signals note inside a directory with no known signals.
- [x] Manually verify `specharbor scan` is stack-agnostic by detecting non-Go signals such as `package.json`, `Dockerfile`, and `.gitlab-ci.yml`.
- [x] Manually verify positional arguments and flags are rejected with clear errors.
- [x] Manually verify the command performs at most one non-recursive listing of the project root and does not recurse.
- [x] Manually verify the command does not write files, call AI providers, call external agents, call workflow tools, call source-control APIs, access the network, run external processes, or install packages.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
