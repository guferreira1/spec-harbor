# Proposal: Implement AI-Assisted Generation

## Summary

Add the first safe AI-assisted OpenSpec generation path:

```text
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> [--overwrite]
```

This command imports a local file that contains AI-authored OpenSpec Markdown in a strict SpecHarbor block format. It parses the file locally, writes only the five required OpenSpec change files under `openspec/changes/<change-id>/`, then runs validation and reports the result.

This change moves SpecHarbor beyond prompt rendering and run-and-report output capture, but keeps a strict boundary: SpecHarbor does not call provider APIs, does not call remote AI services, does not execute agents, does not apply live runner output, does not modify production code, and does not run source-control or workflow automation.

## Problem

SpecHarbor currently supports deterministic blank, template, guided, and agent-assisted authoring flows. Agent-assisted authoring can print a prompt and can explicitly run a local agent in run-and-report mode, but runner output is intentionally not parsed or applied.

Users still need a safe bridge from AI-authored content to actual OpenSpec change files. Copying generated Markdown by hand is error-prone, while applying arbitrary agent output directly would be unsafe. The first useful foundation must accept only a local, user-controlled file, parse a deterministic format, reject ambiguous or unsafe content, limit writes to the known OpenSpec change files, and validate the generated package after writing.

## Goal

Support AI-assisted generation of exactly these files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

The feature must:

- validate `<change-id>` with the domain `ChangeID` value object before filesystem writes;
- require a local `--from-file` source containing AI-authored output;
- parse only a strict deterministic file block format;
- reject malformed output before any write;
- reject unknown filenames, duplicate blocks, missing required blocks, path traversal, absolute paths, nested paths, and empty file content before any write;
- write only known required OpenSpec files under the resolved change directory;
- create the change directory if it is missing and the input is otherwise valid;
- skip existing required files by default;
- overwrite existing required files only with explicit `--overwrite`;
- run validation after successful generation;
- report generated, skipped, overwritten, parse, write, and validation outcomes clearly;
- keep parser/write errors non-zero;
- keep validation warnings exit-zero when there are no validation errors;
- make validation errors exit non-zero after reporting them.

## Command Surface Decision

Choose Option A with an explicit source-file flag:

```text
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> [--overwrite]
```

Rejected alternatives:

- Option B, `--agent-assisted --apply-output`, would blur the existing agent-assisted runner boundary. `--agent-assisted` currently means prompt authoring and optional run-and-report execution, not file application.
- Option C, `--agent-assisted --execute --apply`, is out of scope because it applies live runner output directly. That is a higher-risk workflow needing confirmation controls and stronger runner integration design.

`--ai-assisted` is the safer first public mode name because the source is AI-authored content, not necessarily live output from a SpecHarbor agent runner. The local source file may contain content a user saved from any AI or agent tool, but SpecHarbor only reads it as local text.

## Scope

- Add AI-assisted from-file generation under the existing `generate` command.
- Add `--ai-assisted`.
- Add required `--from-file <agent-output-file>` for `--ai-assisted`.
- Add optional `--overwrite` for `--ai-assisted`.
- Define and implement the strict AI output block parser.
- Define domain concepts for allowed AI-generated filenames, parsed file blocks, parse findings, and generation results.
- Add use case orchestration for safe parse, preflight, writes, validation, and result reporting.
- Add small core-owned ports for reading the source AI output file and safely checking/writing approved OpenSpec target files.
- Add local filesystem adapter behavior for the new ports.
- Add CLI parsing, output formatting, and exit-code mapping.
- Add focused tests for domain, use case, ports/adapters, CLI, regression, and architecture boundaries.
- Update README and docs in the implementation PR because this is public CLI behavior.

## Out of Scope

- Provider API calls.
- Remote AI service calls.
- Local model API calls.
- OAuth.
- Credential management.
- Secret storage.
- Remote execution.
- Cloud execution.
- IDE automation.
- Marketplace integrations.
- Live runner output application.
- `--agent-assisted --execute --apply`.
- Patch application.
- Shell command interpretation.
- Production code generation or modification.
- Writing docs outside the active OpenSpec change.
- Writing config files.
- Writing CI files.
- Writing source-control files.
- Writing arbitrary paths from AI output.
- Auto-commit.
- Auto-push.
- Pull request creation.
- Merge automation.
- Automatic archive.
- Workflow dispatch.
- Documentation-only separate changes.

## Success Criteria

- `specharbor generate <change-id> --ai-assisted --from-file <file>` imports strict local AI output and writes only missing required OpenSpec files under `openspec/changes/<change-id>/`.
- `--overwrite` is required before existing required files are replaced.
- Malformed AI output writes nothing and exits non-zero with clear parse findings.
- Unknown, duplicate, missing, absolute, traversal, nested, or empty file blocks write nothing and exit non-zero with clear findings.
- Successful writes are followed by `validate <change-id>` behavior using the existing validation semantics.
- Validation warnings are printed and keep exit code `0` when there are no errors.
- Validation errors are printed and make the command exit non-zero.
- CLI output includes change id, source file, target path, generated files, skipped files, overwritten files, validation status, finding counts, and safety notes.
- Existing blank, template, guided, agent-assisted, validate, prompt, review, archive, config, scan, init, and workflow behavior remains unchanged.
- Core packages do not import adapters, CLI packages, `os`, `os/exec`, provider SDKs, network APIs, source-control SDKs, workflow SDKs, or external-agent SDKs.
- Documentation explains the from-file workflow, strict output format, overwrite behavior, validation behavior, examples, and safety boundaries.
