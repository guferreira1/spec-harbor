# Proposal: Update Docs and README

## Problem

SpecHarbor's public documentation no longer matches the product direction or the command set that has been implemented through the core workflow.

The current README and docs still describe several planned commands as if they are the main user flow, while the implemented workflow is now centered on explicit OpenSpec change ids and role-based prompts:

```text
specharbor init
specharbor generate <change-id> --blank
specharbor validate <change-id>
specharbor prompt <change-id> --role <role>
specharbor review <change-id>
specharbor archive <change-id>
```

Additional work is in progress or recently worked on in parallel for scan, config, and CI. Public documentation must not claim those features are complete until their branches are stabilized and merged.

SpecHarbor is also being built with OpenSpec/SDD and dogfooding its own workflow. Contributor documentation should make that explicit so future contributors and coding agents create an OpenSpec change before implementing meaningful features.

## Goal

Prepare the implementation plan for updating SpecHarbor's documentation so it accurately explains:

- what SpecHarbor is;
- why it exists;
- the OpenSpec/SDD workflow it supports;
- the implemented command set;
- how to install or build locally;
- how to run tests;
- how to initialize a project;
- how to generate a blank change;
- how to validate a change;
- how to generate role prompts;
- how to review a change;
- how to archive a change;
- how scan and config fit into the roadmap depending on merge status;
- how agent roles work;
- how contributors should use OpenSpec changes before implementing features;
- how to avoid parallel-branch conflicts when multiple agents are working.

The documentation should be direct, developer-focused, and useful for both users and contributors. It should avoid marketing fluff and avoid overstating unfinished functionality.

## Scope

- Update public project documentation in a future implementation of this change.
- Review and update `README.md`.
- Review and update existing docs under `docs/`.
- Add new docs only where the design justifies them and where existing files cannot cover the topic cleanly.
- Clearly separate implemented commands from planned or roadmap commands.
- Keep command examples simple and copy-pasteable.
- Prefer examples that work from the repository root.
- Document local build and test commands for the Go CLI.
- Document the OpenSpec/SDD contribution workflow.
- Document role-based agent prompt generation at a high level.
- Document branch hygiene expectations for parallel coding-agent work.
- Explain that SpecHarbor's own CI is Go-specific, while user project scanning must remain stack-agnostic.
- Mention scan, config, and CI only according to their actual merge status at implementation time.
- Keep this OpenSpec change implementation-ready for a future documentation update.

## Out of Scope

- Changing CLI behavior.
- Changing Go code.
- Changing tests.
- Changing `.github/workflows/ci.yml`.
- Implementing scan behavior.
- Implementing config behavior.
- Implementing CI behavior.
- Adding release automation.
- Adding install scripts.
- Publishing documentation.
- Creating a website or docs site.
- Adding package manager distribution.
- Adding badges unless the documentation implementation explicitly justifies them.
- Adding detailed AI provider setup documentation beyond high-level roadmap notes.
- Updating the living architecture spec unless a concrete documentation implementation need appears and is justified before editing.
- Creating unnecessary documentation files or duplicating the same content across multiple docs.

## Success Criteria

- The future documentation update accurately reflects the implemented command set on the stabilized branch.
- Unfinished or in-progress features are identified as planned, roadmap, or recently worked on, not complete.
- README provides a concise product overview, current command flow, local build/test instructions, and links to deeper docs.
- Usage documentation includes simple examples for init, blank generation, validation, role prompts, review, and archive.
- Contributor documentation explains that meaningful implementation work should start from an OpenSpec change.
- Agent role documentation explains how roles fit the workflow without duplicating all rule files.
- Parallel-agent branch hygiene guidance is clear and practical.
- Documentation explains the distinction between SpecHarbor's Go-specific repository CI and stack-agnostic user project scanning.
- Documentation remains Markdown-only and does not change code, tests, or workflow behavior.
- `go test ./...` succeeds after the documentation implementation unless an unrelated pre-existing failure is documented.
