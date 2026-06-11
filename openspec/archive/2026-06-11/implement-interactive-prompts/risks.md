# Risks: Implement Interactive Prompts

## Risks

### Automation Regressions

Interactive prompts can accidentally make scripted generation hang if the command prompts when stdin is not a terminal or if existing flag combinations change behavior.

### Mode Confusion

SpecHarbor already has `--guided`, `--hybrid`, `--ai-assisted`, and `--agent-assisted`. A new `--interactive` flag could be misunderstood as a replacement for those modes instead of a prompt layer over selected existing paths.

### Flag Precedence Ambiguity

Allowing `--interactive` together with direct generation flags would raise questions about which value wins: the flag value or the prompt answer.

### Prompt Scope Creep

Interactive mode could become a broad wizard that mutates config, fetches arbitrary remote templates, runs agents, starts workflows, or edits production code.

### Terminal Logic In Core

Prompting is user-interface behavior. If stdin, stdout, TTY detection, retry loops, or prompt formatting enter core packages, the architecture boundary becomes harder to test and maintain.

### Business Rules In CLI

The CLI adapter could accidentally duplicate source resolution, remote safety, rendering, write policy, or validation logic while trying to show smarter prompts.

### No-Write Guarantee Before Confirmation

If the implementation invokes generation or source resolution too early, it could create directories, fetch remote content, or write files before the user confirms.

### Ambiguous Pre-Confirmation Summary

If the summary omits validation behavior or safety boundaries, users and tests may not be able to tell whether validation will run automatically or whether side-effect exclusions still apply.

### Ambiguous Confirmation Matching

If confirmation parsing differs from menu parsing, users may reasonably expect uppercase or mixed-case answers to work but receive inconsistent behavior.

### Non-TTY Detection Portability

TTY detection differs by platform and test environment. Incorrect detection can block valid interactive use or allow CI hangs.

### Retry Loop Friction

A retry limit protects against infinite loops, but overly strict parsing or unclear retry messages can frustrate users.

### Cancellation Exit Code Surprise

Users may expect cancellation to exit zero because it is an intentional action. The design chooses non-zero because no generation happened.

### Remote Alias Surprise

Selecting a config alias can resolve to a remote template. Users may not realize that network access can happen after confirmation when an existing alias points to `source: remote`.

### Documentation Drift

Interactive mode is public CLI behavior. Incomplete docs could imply that interactive prompts run agents, call providers, validate every generated change, or automate source-control steps.

## Mitigations

### Fail Fast Without A TTY

Check TTY availability before prompts and return `interactive mode requires a TTY` with a non-zero exit. Add CLI tests that simulate non-TTY input and prove no writes occur.

### Keep The Command Surface Small

Implement only:

```bash
specharbor generate <change-id> --interactive
```

Do not add `specharbor interactive generate <change-id>` or prompt for change id in the first version.

### Reject Mixed Flags

Make `--interactive` mutually exclusive with direct generation mode flags and direct input flags. This removes precedence questions and makes the prompt summary the single source of selected options.

### Limit Supported Paths

Support blank, built-in template, custom template, config template, and hybrid generation only. Do not offer direct guided, AI-assisted, or agent-assisted paths in the first interactive menu.

### Keep Prompting In The CLI Adapter

Use an adapter-owned terminal abstraction with fake input/output for tests. Keep domain and use cases non-interactive.

### Delegate Business Rules

Use existing value objects for answer validation, then delegate confirmed generation to existing use cases. Do not duplicate source existence checks, config alias resolution, remote safety, rendering, write policy, or validation behavior in prompt code.

### Require Confirmation Before Generation

Print a deterministic summary and ask for confirmation before invoking any generation use case. Cancellation, EOF, and retry exhaustion all return before writes.

### Make Summary Side Effects Deterministic

Always include exactly one validation line in the pre-confirmation summary. Use `Validation: automatic no` for blank, built-in template, custom template, and config template paths, and `Validation: automatic yes` for hybrid paths. Always include the deterministic safety section before the confirmation prompt.

### Make Confirmation Parsing Explicit

Trim confirmation answers and match `y`, `yes`, `n`, and `no` case-insensitively. Treat empty confirmation and EOF as cancellation. Retry unsupported confirmation answers up to three attempts, then fail clearly with no writes.

### Bound Retry Loops

Use three attempts for invalid required answers, invalid menu answers, invalid non-empty optional type, and unsupported confirmation answers. After exhaustion, return a clear error and write nothing.

### Document Cancellation Semantics

Document that cancellation exits non-zero with `operation cancelled` because no generation occurred. Keep this behavior consistent for `n`, `no`, empty confirmation input, and EOF.

### Make Remote Alias Behavior Explicit

Do not prompt for arbitrary URLs or checksums. Document that remote network behavior is possible only through existing config aliases after confirmation and uses existing remote safeguards.

### Protect Existing Modes With Regression Tests

Add regression tests for all existing generation modes and non-generation commands so parser changes do not alter current behavior.

### Update Public Docs In The Same Change

Update README and generation docs with command syntax, prompt examples, non-TTY behavior, confirmation/cancellation behavior, selected-mode validation behavior, write boundaries, and all safety exclusions.

## Trade-Offs Accepted

- The first version requires `<change-id>` on the command line instead of prompting for it. This keeps output paths explicit and avoids changing the current positional contract.
- The first version rejects direct flags with `--interactive`. This is less flexible but avoids precedence ambiguity and keeps prompt summaries deterministic.
- Direct guided generation is not offered in the first interactive menu. Existing `--guided` remains available through flags, and a future change can add a guided prompt branch if needed.
- Cancellation exits non-zero. This makes cancelled generation easy to detect in wrappers, even though the user intentionally cancelled.
- Interactive mode does not add a validation prompt. The selected mode's validation behavior remains unchanged, and users can run `specharbor validate <change-id>` manually.
- Optional title and summary are offered for custom and config templates to preserve existing direct flag behavior, even though some config aliases may resolve to sources that ignore them.

## Open Questions

- A future interactive guided branch may be useful, but it should be specified separately so its relationship to existing non-interactive `--guided` is clear.
- A future `specharbor generate --interactive` form could prompt for change id, but it should address path identity, retry behavior, and automation expectations explicitly.
- A future validation prompt may be useful after non-hybrid generation, but it should specify whether validation errors affect exit code after files are written.
- A future richer interactive UI may be useful, but external TUI frameworks should not be introduced unless the benefits justify the dependency and testing cost.
