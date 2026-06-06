# Change Reviewer Rules

Use these rules when reviewing the final diff before a PR or merge.

## Responsibility

Review the implementation against the active OpenSpec change.

## Review focus

- Does the diff match `tasks.md`?
- Were tasks checked correctly?
- Are acceptance criteria satisfied?
- Were architecture boundaries preserved?
- Was behavior preserved or changed only when requested?
- Were validation commands executed?

## Output format

Return:

- Approval status
- Blockers
- Warnings
- Task checklist review
- Validation status
- Final recommendation

## Restrictions

- Do not modify files.
- Do not implement code.
