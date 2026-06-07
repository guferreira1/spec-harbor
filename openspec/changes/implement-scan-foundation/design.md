# Design: Implement Scan Foundation

## Overview

`specharbor scan` inspects the current project root and reports a deterministic, stack-agnostic project context.

This first scan foundation detects only the presence of known top-level or conventional signal files and directories. It does not parse manifests deeply, read source code, traverse the tree recursively, detect frameworks, call AI providers, call external tools, or access the network.

The implementation should be small enough to review easily while establishing the product shape needed for future scanner extensions: a structured domain result, a deterministic detection catalog and assembler in the domain, a use case orchestration boundary, a scan-specific filesystem port, and CLI report formatting outside the core.

The scanner must be stack-agnostic. It must not assume the host project is written in Go, Node, Java, Python, Rust, .NET, or any specific stack. It detects known signals when present and reports unknown or undetected areas through notes and `- none detected` lines.

## CLI Contract

Supported command shape:

```text
specharbor scan
```

Reject:

- `specharbor scan <anything>` (no positional arguments are supported in this version)
- any flag such as `--json`, `--format`, `--path`, `--deep`, `--ai`, `--github`, or `--all`

On argument errors, return an error from the CLI adapter so `cmd/specharbor/main.go` handles the existing error flow.

The CLI obtains the project root from the current working directory.

Scan is informational. When the scan completes, the CLI prints the report and returns zero. A non-zero exit is reserved for execution errors that prevent scanning, such as an unavailable or unreadable project root, a current-working-directory lookup failure, or a filesystem failure. There is no `invalid` scan status; a project with no detected signals is a successful scan with an empty result and explanatory notes.

## Expected CLI Output

Populated output:

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

Empty output:

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

Output rules:

- Always print the completion line, the project root line, and all seven sections in the order shown: Detected ecosystems, Package managers, Test command hints, CI, Containers/deployment, SpecHarbor/OpenSpec, Notes.
- Separate the project root line and each section with a single blank line.
- When a section has no entries, print a single `- none detected` line under it. This applies to every section, including Notes.
- The `Project root:` line prints the absolute current working directory provided by the CLI adapter. This is the only place an absolute local path appears; it is expected and required so the user can confirm what was scanned.
- The report must not print debug output, provider names, agent names, source-control details, network output, manifest contents, source code, or future metadata fields.

## Detection Categories

The scan result groups detections into these categories:

- ecosystems (language/ecosystem signals);
- package managers;
- test command hints (derived from detected ecosystems);
- CI providers;
- container/deployment artifacts;
- OpenSpec/SpecHarbor signals;
- notes (deterministic observations about undetected or noteworthy-absent areas).

## Detection Catalog

Detection is driven by a single deterministic, ordered catalog of signal rules in the domain. Each rule describes one detectable signal: its category, a human-readable name, the relative path or suffix to probe, the probe kind, and an optional test command hint.

Recommended rule shape:

```text
Category    one of: ecosystem, package-manager, ci, container-deployment, specharbor
Name        human-readable label, may be empty for path-only categories
Path        relative path to probe, or a suffix for suffix rules
Kind        one of: file, directory, file-suffix
TestCommand optional test command hint emitted when the rule matches
```

Probe kinds:

- `file`: the rule matches when `FileExists(projectRoot, Path)` is true.
- `directory`: the rule matches when `DirectoryExists(projectRoot, Path)` is true.
- `file-suffix`: the rule matches when any top-level entry name ends with `Path`. This is the only rule kind that requires the single non-recursive root listing, and exists to support extension-based detection such as `.csproj` and `.sln`.

Recommended initial catalog (deterministic order):

| Category             | Name           | Path                    | Kind        | Test command hint |
| -------------------- | -------------- | ----------------------- | ----------- | ----------------- |
| ecosystem            | Go             | `go.mod`                | file        | `go test ./...`   |
| ecosystem            | Node           | `package.json`          | file        | `npm test`        |
| ecosystem            | Node           | `tsconfig.json`         | file        |                   |
| ecosystem            | Java           | `pom.xml`               | file        | `mvn test`        |
| ecosystem            | Java           | `build.gradle`          | file        | `gradle test`     |
| ecosystem            | Java           | `build.gradle.kts`      | file        | `gradle test`     |
| ecosystem            | Java           | `settings.gradle`       | file        |                   |
| ecosystem            | Java           | `settings.gradle.kts`   | file        |                   |
| ecosystem            | Python         | `pyproject.toml`        | file        | `pytest`          |
| ecosystem            | Python         | `requirements.txt`      | file        | `pytest`          |
| ecosystem            | Python         | `Pipfile`               | file        | `pytest`          |
| ecosystem            | Python         | `poetry.lock`           | file        |                   |
| ecosystem            | Rust           | `Cargo.toml`            | file        | `cargo test`      |
| ecosystem            | .NET           | `.csproj`               | file-suffix | `dotnet test`     |
| ecosystem            | .NET           | `.sln`                  | file-suffix | `dotnet test`     |
| package-manager      | npm            | `package-lock.json`     | file        |                   |
| package-manager      | pnpm           | `pnpm-lock.yaml`        | file        |                   |
| package-manager      | yarn           | `yarn.lock`             | file        |                   |
| ci                   | GitHub Actions | `.github/workflows`     | directory   |                   |
| ci                   | GitLab CI      | `.gitlab-ci.yml`        | file        |                   |
| ci                   | Jenkins        | `Jenkinsfile`           | file        |                   |
| container-deployment | (empty)        | `Dockerfile`            | file        |                   |
| container-deployment | (empty)        | `docker-compose.yml`    | file        |                   |
| container-deployment | (empty)        | `docker-compose.yaml`   | file        |                   |
| container-deployment | (empty)        | `compose.yml`           | file        |                   |
| container-deployment | (empty)        | `compose.yaml`          | file        |                   |
| container-deployment | (empty)        | `kubernetes`            | directory   |                   |
| container-deployment | (empty)        | `k8s`                   | directory   |                   |
| container-deployment | (empty)        | `helm`                  | directory   |                   |
| specharbor           | (empty)        | `openspec/project.md`   | file        |                   |
| specharbor           | (empty)        | `openspec/changes`      | directory   |                   |
| specharbor           | (empty)        | `openspec/specs`        | directory   |                   |
| specharbor           | (empty)        | `.specharbor/config.yml`| file        |                   |
| specharbor           | (empty)        | `.specharbor/rules`     | directory   |                   |

This catalog is suggested. The implementer may adjust labels or ordering for clarity, but the catalog must remain a single deterministic source of truth, must stay stack-agnostic, and must only probe top-level or conventional paths.

The catalog is the scan presence-detection source of truth. It is intentionally separate from `domain.RequiredOpenSpecChangeFiles()`, which defines the required change-file policy for generation, validation, and review. Scan reports OpenSpec/SpecHarbor presence; it does not enforce change-file policy, so it must not reuse or duplicate the required-change-file list as a detection catalog.

## Signal Display

Each matched rule becomes a detection with a display `Signal`:

- `file` rule: `Signal` is the path, for example `go.mod`.
- `directory` rule: `Signal` is the path followed by a trailing slash, for example `.github/workflows/`, `openspec/changes/`, `kubernetes/`.
- `file-suffix` rule: `Signal` is the suffix pattern itself, for example `.csproj`. Displaying the pattern rather than a discovered filename keeps output deterministic.

The CLI formatter prints each detection as:

- `- {Name}: {Signal}` when `Name` is non-empty (ecosystems, package managers, CI);
- `- {Signal}` when `Name` is empty (container/deployment, OpenSpec/SpecHarbor).

This uniform rule reproduces the expected output exactly, so the formatter needs no per-category special cases beyond the empty-name check.

## Test Command Hints

Test command hints are derived deterministically from matched ecosystem rules that carry a `TestCommand`. Hints are collected in catalog order and de-duplicated, preserving first-seen order. A hint is never inferred from package manager rules or by running any command; it is a static, conventional suggestion only.

For example, a project with `go.mod` and `package.json` yields:

```text
- go test ./...
- npm test
```

Test command hints are best-effort conventions, not verified commands. The scan never executes them.

## Notes

Notes are deterministic observations produced by the assembler from the set of matched rules. The initial note rules are intentionally small:

- When no rules match at all, emit exactly one note: `No known project signals detected.`
- When at least one container/deployment signal is detected but no Kubernetes or Helm directory signal is present (none of `kubernetes/`, `k8s/`, `helm/` matched), emit: `No Kubernetes manifests detected.`

These two rules reproduce both expected output examples. Additional note rules are out of scope for this change and may be added later. When no note applies but signals were detected, the Notes section prints `- none detected`.

## Domain Model

Add scan domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- a detection value with a human-readable name and a display signal;
- a structured scan result containing project root, ecosystems, package managers, test command hints, CI providers, container/deployment artifacts, OpenSpec/SpecHarbor signals, and notes;
- the deterministic detection catalog;
- a pure assembler that turns the project root plus a set of matched rules into a scan result, including signal display, de-duplicated test command hints, and notes.

A possible result shape:

```text
ProjectRoot          string
Ecosystems           []ScanDetection
PackageManagers      []ScanDetection
TestCommandHints     []string
CIProviders          []ScanDetection
ContainerDeployments []ScanDetection
SpecHarborSignals    []ScanDetection
Notes                []string
```

A possible detection shape:

```text
Name   string
Signal string
```

The result constructor should defensively copy slices, following the convention already used by `domain.NewReviewResult` and `domain.NewGenerationResult`.

The deterministic catalog, signal display, test command hint de-duplication, and note rules are pure domain/application logic, so they belong in the domain and must be testable without mocks.

The domain package must not import adapters, CLI packages, ports, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.

## Existing ProjectContext Placeholder

`internal/core/domain/project_context.go` currently defines an unused `ProjectContext` placeholder. It is dead code with no references and predates concrete scan behavior.

This change introduces the concrete scan result domain model, so the placeholder must not be left in place. Replace it with the explicit scan result model or remove it once superseded. Do not keep `ProjectContext` as an unused abstraction, because the global rules forbid placeholder abstractions kept only to satisfy a target architecture. This is the only file outside the scan change that this change may modify or delete, and only because it is the scan/project-context placeholder being replaced.

## Ports

Add a scan-specific filesystem port under:

```text
internal/core/ports
```

Expected contract:

```text
FileExists(root string, relativePath string) (bool, error)
DirectoryExists(root string, relativePath string) (bool, error)
ListDirectoryNames(root string, relativePath string) ([]string, error)
```

Use a behavior-specific name such as `ScanFileSystem`.

`ListDirectoryNames` returns the names of the immediate entries of a directory, without recursion. It is used to validate project root availability and to support extension-based detection (`.csproj`, `.sln`). The port must contain only the operations needed by this change.

Do not reuse initialization, validation, generation, archive, review, prompt rendering, AI provider, workflow dispatcher, source-control, or agent contracts directly, even if the local filesystem adapter can satisfy overlapping methods.

## Use Case

Add a scan use case under:

```text
internal/core/usecase
```

Expected input:

- project root.

Expected behavior:

- validate that the use case dependency is present;
- validate that the filesystem dependency is present;
- trim and validate that project root is non-empty;
- list the top-level entries of the project root through the scan filesystem port to validate root availability and to support suffix rules;
- return an execution error when the project root cannot be listed (unavailable or unreadable root);
- iterate `domain.ScanSignalRules()` in catalog order;
- for `file` rules, probe `FileExists`; for `directory` rules, probe `DirectoryExists`; for `file-suffix` rules, match the suffix against the top-level entry names already listed;
- return an execution error when a filesystem existence check fails;
- collect the matched rules;
- call the pure domain assembler with the project root and matched rules to build the structured scan result;
- return the structured scan result;
- never print, call `os`, access terminal IO, call provider APIs, call source-control APIs, run external tools or processes, access the network, write files, import adapters, or import workflow SDKs.

The use case performs filesystem probing and delegates all grouping, signal display, test command hint derivation, and note computation to the pure domain assembler.

## Filesystem Adapter

Use `internal/adapters/filesystem` as the concrete implementation of the scan filesystem port.

The local filesystem adapter already supports directory existence and file existence checks that distinguish directories from files. Add only the listing support required by the scan port:

- list the immediate entry names of a directory without recursion;
- return an error when the directory does not exist or cannot be read, so the use case can treat an unavailable root as an execution error.

The adapter must not know:

- which signal paths or suffixes are detected;
- how detections are categorized;
- how signals are displayed;
- how test command hints are derived;
- how notes are computed;
- how CLI reports are formatted;
- future AI, agent, workflow, source-control, security, license, or framework-detection policy.

Add the new local filesystem method as a thin, non-recursive listing. The adapter must not recurse into subdirectories.

## CLI Adapter

Update `internal/adapters/cli` so the `scan` command:

- parses `specharbor scan` with no positional arguments and no flags;
- rejects any positional argument;
- rejects any flag;
- obtains the current working directory for the project root;
- constructs the scan use case with the local filesystem adapter;
- invokes the use case with the project root;
- prints a human-readable scan report from the structured result;
- returns argument and execution errors to `cmd/specharbor/main.go` without panicking;
- returns zero when the report is printed successfully.

The CLI adapter may format human-readable output, but it must not contain detection catalog data, categorization, signal display policy, test command hint derivation, note computation, filesystem policy, provider logic, source-control logic, or workflow logic.

`cmd/specharbor/main.go` should remain limited to existing process bootstrapping; no change is expected there.

Update the help text only if it is already inaccurate; the existing help already lists `scan`. Replace only the `scan` placeholder entry in the command registry.

## Extensibility Direction

Prepare for future scan capabilities through the structured result, the deterministic catalog, and the pure assembler, but do not add unused extension frameworks.

Acceptable first implementation shapes:

- a single scan use case with small private helper methods;
- an explicit scan result and detection domain type;
- a single deterministic detection catalog plus a pure assembler in the domain;
- a dedicated scan filesystem port;
- direct CLI wiring similar to `validate`, `review`, and `archive`.

Avoid:

- recursive directory walkers;
- manifest parsers;
- framework detectors;
- AI clients, provider clients, external-agent dispatchers, or workflow connectors;
- source-control clients or network calls;
- exported scanner registries, factories, chains, strategy interfaces, plugin abstractions, or provider abstractions;
- `--json`, `--deep`, `--path`, `--ai`, or other flag plumbing for future modes.

Future changes can introduce additional collaborators, richer detection, deeper parsing, and machine-readable output once concrete behavior requires them.

## Testing Strategy

Add focused tests for:

- the detection catalog being non-empty, deterministic, and stack-agnostic across categories;
- the assembler grouping matched ecosystem, package manager, CI, container/deployment, and OpenSpec/SpecHarbor rules into the correct result fields in catalog order;
- the assembler formatting `directory` signals with a trailing slash, `file` signals as the path, and `file-suffix` signals as the suffix pattern;
- the assembler emitting `- {Name}: {Signal}` semantics through non-empty names for ecosystems, package managers, and CI, and empty names for container/deployment and OpenSpec/SpecHarbor;
- the assembler de-duplicating test command hints in first-seen order;
- the assembler emitting `No known project signals detected.` when no rules match;
- the assembler emitting `No Kubernetes manifests detected.` when container/deployment signals exist without Kubernetes or Helm directories;
- the assembler emitting no Kubernetes note when a Kubernetes or Helm directory signal is present;
- the scan result constructor defensively copying slices;
- the use case returning a populated scan result for a fake filesystem with Go, Node, npm, GitHub Actions, Docker, and OpenSpec signals present;
- the use case returning an empty scan result with the no-signals note when nothing is detected;
- the use case detecting `.csproj` and `.sln` through the top-level listing;
- the use case rejecting an empty project root;
- the use case returning an execution error when the project root cannot be listed;
- the use case returning an execution error when a filesystem existence check fails;
- the use case never reading nested paths beyond the conventional catalog paths and never recursing;
- the local filesystem adapter satisfying the scan filesystem port;
- the local filesystem adapter listing immediate entry names without recursion;
- the local filesystem adapter returning an error for a missing directory listing;
- the CLI printing a populated scan report in the expected shape;
- the CLI printing the empty scan report in the expected shape, including `- none detected` under every section;
- the CLI rejecting positional arguments and flags;
- the CLI returning zero when the report is printed;
- existing `help`, `version`, `init`, `generate`, `prompt`, `validate`, `review`, `archive`, `config`, and unknown command behavior is preserved.

Use fake ports for use case tests. Use temporary directories for filesystem adapter and CLI integration-style tests. Tests must not depend on the real host stack of the SpecHarbor repository.

## Validation

Run:

```text
gofmt
go test ./...
```

Do not require network access, provider credentials, local model credentials, source-control credentials, external-agent tools, workflow credentials, package installation, or external processes for this change.
