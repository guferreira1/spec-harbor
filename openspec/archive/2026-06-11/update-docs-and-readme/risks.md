# Risks: Update Docs and README

## Stale feature claims

Parallel scan, config, and CI work may not be merged when the documentation update is implemented. The docs could accidentally describe in-progress behavior as complete.

Mitigation:

- Verify command behavior on the documentation branch before writing implemented status.
- Use implemented, in-progress, and planned labels conservatively.
- Mention scan/config/CI as roadmap when merge status is uncertain.

## Command syntax drift

Old documentation examples may use obsolete syntax such as prompt `--agent` instead of the current role-based prompt command.

Mitigation:

- Inspect the CLI command registry and tests before editing docs.
- Prefer examples using verified command syntax.
- Keep one authoritative command reference in `docs/usage.md` and link to it from the README.

## Documentation sprawl

Adding many files can make the docs harder to maintain and easier to contradict.

Mitigation:

- Use the file plan in `design.md`.
- Reconcile existing `docs/getting-started.md` and `docs/generation-modes.md` instead of leaving stale duplicates.
- Avoid creating extra docs files unless a concrete need is documented.

## Mixing documentation with implementation

A documentation polish change could expand into CLI, test, scan, config, CI, or release work.

Mitigation:

- Keep this change Markdown-only except for task updates.
- Do not modify Go code, Go tests, or `.github/workflows/ci.yml`.
- Open separate OpenSpec changes for behavior, CI, release, or distribution work.

## Confusing SpecHarbor CI with user project scanning

Because SpecHarbor is a Go CLI, docs may accidentally imply that user project scanning is Go-specific.

Mitigation:

- State that SpecHarbor's own CI is Go-specific.
- State separately that user project scanning must remain stack-agnostic.
- Avoid Go-only scan examples unless explicitly identified as examples of scanning the SpecHarbor repository itself.

## Parallel agent conflicts

Multiple coding agents may update overlapping docs, OpenSpec changes, or workflow files on separate branches.

Mitigation:

- Document branch hygiene in contributor docs.
- Recommend separate branches or worktrees for separate OpenSpec changes.
- Tell agents to check `git status --short` before starting and before finalizing.
- Warn that unrelated dirty worktree changes must not be reverted or overwritten.

## Over-documenting AI providers

The product direction includes AI-provider and agent-assisted workflows, but detailed provider setup is not part of this change.

Mitigation:

- Keep provider notes high-level and roadmap-oriented.
- Do not include API key examples, provider credentials, provider-specific setup, or integration walkthroughs.
- Make clear that agent-assisted workflows should not require provider API keys.
