# Acceptance Criteria: Refactor to Hexagonal Architecture

- The OpenSpec change includes proposal, design, tasks, acceptance criteria, and risks documents.
- The design maps every current package to the target Hexagonal Architecture structure.
- The implemented scope is limited to package reorganization, import updates, and package declaration updates for the files listed in `tasks.md`.
- Pure value definitions for project context, generation modes, agent targets, and AI provider identities are located in `internal/core/domain`.
- Version metadata is located in `internal/platform/version`.
- CLI delivery code is located in `internal/adapters/cli`, and `cmd/specharbor` remains limited to process bootstrapping.
- Existing CLI behavior is preserved for `help`, `version`, known placeholder commands, and unknown commands.
- The dependency rule is preserved: core code must not import adapters, CLI, terminal IO, file system access, HTTP clients, provider SDKs, or external APIs.
- Separation between AI providers, agent targets, and workflow connectors is preserved without adding new implementations.
- Agent-assisted authoring remains a workflow that does not require provider API keys.
- This change does not add new behavior, new OpenSpec file operations, new providers, new agents, new scanners, new validators, new templates, new workflow connectors, or config storage.
- This change does not add placeholder ports, use cases, adapters, compatibility wrappers, or empty packages.
