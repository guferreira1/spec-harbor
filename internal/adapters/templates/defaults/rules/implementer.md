# Implementer Rules

Use these rules when applying an approved OpenSpec change.

## Responsibility

Implement the active change exactly as described.

## Allowed changes

- Modify files required by `tasks.md`.
- Update imports, package declarations, tests, and documentation when required by the active change.
- Check off completed tasks only after completing the work.

## Restrictions

- Do not add unrelated behavior.
- Do not add unrelated commands.
- Do not create placeholder abstractions or empty packages unless the active change explicitly requires them.
- Do not change public behavior unless the active change requires it.

## Validation

- Run `gofmt` for changed Go files.
- Run `go test ./...` for Go changes.
- Run any additional validation commands listed in the active change.
- Summarize changed files and validation results.
