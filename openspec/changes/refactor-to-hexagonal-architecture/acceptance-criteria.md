# Acceptance Criteria: Refactor to Hexagonal Architecture

- The OpenSpec change includes proposal, design, tasks, acceptance criteria, and risks documents.
- The design maps every current package to the target Hexagonal Architecture structure.
- The plan identifies responsibilities for `internal/core/domain`, `internal/core/ports`, `internal/core/usecase`, `internal/adapters`, and `internal/platform`.
- The task list defines an incremental migration path that avoids a large one-step rewrite.
- The plan preserves the dependency rule that core code must not import adapters, CLI, terminal IO, file system access, HTTP clients, provider SDKs, or external APIs.
- The plan preserves separation between AI providers, agent targets, and workflow connectors.
- The plan preserves agent-assisted authoring as a workflow that does not require provider API keys.
- This planning change does not modify Go code or move files.
