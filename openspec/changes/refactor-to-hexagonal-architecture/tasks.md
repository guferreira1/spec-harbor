# Tasks: Refactor to Hexagonal Architecture

## Phase 0: Baseline and Guardrails

- [ ] Read `AGENTS.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, `openspec/changes/refactor-to-hexagonal-architecture/proposal.md`, and `openspec/changes/refactor-to-hexagonal-architecture/design.md`.
- [ ] Confirm the current implementation files before moving anything: `cmd/specharbor/main.go`, `internal/cli/cli.go`, `internal/scanner/project.go`, `internal/generator/mode.go`, `internal/prompt/agent.go`, `internal/ai/provider.go`, and `internal/version/version.go`.
- [ ] Run `go test ./...` to establish the pre-migration baseline.
- [ ] Keep the implementation limited to package reorganization, import updates, and package declarations for the files listed above.
- [ ] Do not add new CLI commands, command behavior, providers, agents, scanners, validators, config storage, workflow connectors, templates, or OpenSpec file operations in this change.
- [ ] Do not create placeholder abstractions or empty packages only to satisfy the target architecture.

## Phase 1: Move Pure Domain Values

- [ ] Move `internal/scanner/project.go` to `internal/core/domain/project_context.go` and keep `ProjectContext` as a pure domain value.
- [ ] Move `internal/generator/mode.go` to `internal/core/domain/generation_mode.go` and keep `GenerationMode` plus its existing constants unchanged.
- [ ] Move `internal/prompt/agent.go` to `internal/core/domain/agent.go` and keep `Agent` plus its existing constants unchanged.
- [ ] Move `internal/ai/provider.go` to `internal/core/domain/ai_provider.go` and keep `Provider` plus its existing constants unchanged.
- [ ] Update package declarations for moved domain files to `package domain`.
- [ ] Update any imports that reference `internal/scanner`, `internal/generator`, `internal/prompt`, or `internal/ai`.
- [ ] After imports are updated, ensure no Go files remain in the old package directories.

## Phase 2: Move Platform Metadata

- [ ] Move `internal/version/version.go` to `internal/platform/version/version.go`.
- [ ] Keep the package name `version` and keep `Version` unchanged.
- [ ] Update imports from `github.com/guferreira1/spec-harbor/internal/version` to `github.com/guferreira1/spec-harbor/internal/platform/version`.
- [ ] Verify `internal/platform/version` contains only technical version metadata and no business rules.

## Phase 3: Move CLI Delivery Adapter

- [ ] Move `internal/cli/cli.go` to `internal/adapters/cli/cli.go`.
- [ ] Keep the package name `cli` so the existing `Execute(args []string) error` entrypoint remains stable.
- [ ] Update `cmd/specharbor/main.go` to import `github.com/guferreira1/spec-harbor/internal/adapters/cli`.
- [ ] Keep `cmd/specharbor/main.go` limited to process bootstrapping: pass `os.Args[1:]`, print top-level errors to stderr, and set the process exit code.
- [ ] Preserve current CLI output and error behavior for `help`, `version`, known placeholder commands, and unknown commands.
- [ ] Do not introduce use cases while commands still only print placeholder output.

## Phase 4: Avoid Premature Ports and Use Cases

- [ ] Do not create `internal/core/ports` interfaces until an existing use case needs them.
- [ ] Do not create `internal/core/usecase` orchestration until behavior exists outside CLI parsing and placeholder command output.
- [ ] Do not create `internal/adapters/scanner`, `internal/adapters/specauthor`, `internal/adapters/prompt`, or `internal/adapters/ai` until concrete adapter behavior exists; the current scanner, generator, prompt, and AI files contain domain values only.
- [ ] If a moved package has no remaining code, prefer removing the old empty directory instead of leaving compatibility wrappers.

## Phase 5: Verification

- [ ] Run `gofmt` on all moved or edited Go files.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/specharbor version` and verify it still prints the same version value.
- [ ] Run `go run ./cmd/specharbor help` and verify the command list is unchanged.
- [ ] Run `go run ./cmd/specharbor unknown-command` and verify the unknown-command error behavior is unchanged.
- [ ] Verify no Go files import the old package paths: `internal/cli`, `internal/scanner`, `internal/generator`, `internal/prompt`, `internal/ai`, or `internal/version`.
- [ ] Verify `internal/core/domain` does not import adapters, CLI, file system, terminal, network, environment, provider SDKs, or platform packages.
- [ ] Update this `tasks.md` by checking off only the tasks completed during the implementation.
