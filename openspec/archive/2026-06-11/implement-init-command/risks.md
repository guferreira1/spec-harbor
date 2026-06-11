# Risks: Implement Init Command

## Accidental overwrite

The command writes files into an existing project. The primary risk is replacing user-authored OpenSpec or SpecHarbor files.

Mitigation: the use case must skip existing files, and the filesystem adapter should use absent-only file creation so existing files are protected even if checks race.

## Architecture leakage

Initialization touches filesystem paths and CLI output, which can easily pull business rules into the CLI adapter.

Mitigation: keep required file lists, initialization status, and overwrite policy in the core use case. Keep `os`, path handling, and concrete writes in adapters. Keep the CLI responsible only for current-directory lookup, use case invocation, and output formatting.

## Template runtime dependency

If default rule files are copied from repository-relative paths at runtime, installed binaries may fail outside the source checkout.

Mitigation: use embedded templates or generated constants in an adapter so defaults are available from the compiled CLI.

## Overbuilding the first feature

There is a risk of adding flags, interactive setup, validation, migration, or repair behavior before users need it.

Mitigation: limit this change to no-overwrite initialization in the current working directory. Defer force overwrite, custom paths, validation, and interactive flows to later OpenSpec changes.

## Ambiguous partial initialization

Projects may already contain some OpenSpec or SpecHarbor files. Treating any existing marker as fully initialized could leave the structure incomplete.

Mitigation: consider the project already initialized only when every required directory and file exists. Otherwise create missing items and report existing items as skipped.

## Default content drift

Generated defaults may diverge from the repository's current `.specharbor` example configuration and role rules.

Mitigation: implementation tests should verify all required default outputs exist, and the implementation should source defaults from a single embedded template set.
