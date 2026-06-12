# Risks: Implement Project Briefing Foundation

## Suggestions Becoming Facts

Detected context such as ecosystem files or test command hints may look authoritative even when they are only shallow scan signals.

Mitigation: label detected context as suggestions until user confirmation, store confirmed answers with a user-provided source, and render detected context in a separate section from confirmed answers.

## Scope Creep Into Repository Intelligence

Project briefing could expand into repository indexing, embeddings, RAG, dependency graph analysis, source-code parsing, or GitHub remote discovery.

Mitigation: keep this change limited to interactive questions, optional shallow scan suggestions, deterministic Markdown rendering, and write-if-absent persistence. Add tests or static checks as needed to ensure no vector store, RAG provider, remote API, or repository indexing code is introduced.

## CLI Business Rules

The CLI prompt layer could accidentally own decisions about what counts as confirmed project context.

Mitigation: keep source categories and project brief construction in domain/use case code. The CLI may collect answers and render prompts, but the use case should validate required answers and source classification before rendering and writing.

## Existing Brief Overwrite

Overwriting or appending to an existing `.specharbor/project-brief.md` could destroy user-maintained context or create confusing duplicated sections.

Mitigation: use write-if-absent behavior in this first version and return a clear error when the file already exists. Defer merge/update behavior to a future OpenSpec change.

## Overconfident Command Capture

Users may select conventional commands such as `go test ./...` or `npm test` even though the command has not been verified in the project.

Mitigation: record commands as user-provided answers, not verified execution results. Do not run commands, package managers, scripts, or shell probes during briefing.

## Prompt Injection Expectations

Because the brief is intended for future context-aware workflows, users may expect `specharbor prompt` to automatically include it.

Mitigation: explicitly keep prompt injection and generated prompt changes out of scope. Document that future prompt usage requires a separate OpenSpec change.

## Documentation Drift

Documentation could describe `specharbor brief` before the implementation is merged, or describe it as an indexing or agent-intelligence feature.

Mitigation: update public docs only in the implementation change after behavior exists, and describe briefing as explicit user-provided context collection with shallow suggestions, not RAG or repository intelligence.

## Architecture Boundary Drift

Briefing touches CLI prompts, filesystem writes, scan suggestions, and Markdown rendering, which could blur responsibilities across layers.

Mitigation: keep terminal IO in adapters, persistence policy in the use case, source classification in domain, concrete filesystem behavior in adapters, and dependencies directed inward.

## Partial Writes

Failures after confirmation could leave a partial or empty project brief.

Mitigation: render and validate the full Markdown content before writing, create the directory only after confirmation, use safe filesystem write behavior, and add cancellation and failure tests proving no brief file is written before confirmation.
