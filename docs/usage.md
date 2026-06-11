# Usage

Run examples from the repository root while developing SpecHarbor:

```bash
go run ./cmd/specharbor help
```

You can also build a local binary:

```bash
go build ./cmd/specharbor
```

That creates a local `specharbor` binary in the repository root.

## Implemented Commands

### Check Version Metadata

```bash
go run ./cmd/specharbor version
specharbor version
```

`specharbor version` prints deterministic multiline build metadata:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Fields:

- `version`: product version metadata.
- `commit`: source commit supplied by the build.
- `date`: build date supplied by the build.
- `dirty`: working tree state supplied by the build.

`dev` means no release version was injected. `unknown` means the build did not provide that metadata field. Git release tags use `vX.Y.Z`, for example `v0.1.0`, while release binary metadata uses plain `X.Y.Z`, for example `0.1.0`.

Plain `go install` without `-ldflags` uses the same development fallback metadata. An installed binary built that way is expected to print:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

This is expected behavior. To get release metadata, the binary must be built with injected `-ldflags` values.

Release builds inject metadata through Go `-ldflags -X` variables in `github.com/guferreira1/spec-harbor/internal/platform/version`:

```bash
go build \
  -ldflags "
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Version=0.1.0
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Commit=abc1234
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Date=2026-06-10T19:00:00Z
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Dirty=false
  " \
  ./cmd/specharbor
```

Runtime displays the injected version string as-is and does not normalize it. It does not inspect Git tags, read `.git`, run Git, or normalize versions. GoReleaser injects release metadata when building GitHub Release assets from tags such as `v0.1.0`, and those binaries display plain metadata such as `0.1.0`. Installation options are documented in [Install](install.md): `install.sh` and the npm wrapper package are implemented in-repo, while npm registry publishing, the Homebrew tap, native Linux packages, Windows package managers, signing, SBOMs, and Docker images remain future steps.

### Initialize a Project

```bash
go run ./cmd/specharbor init
```

`init` creates the OpenSpec and SpecHarbor project files that are missing in the current directory. Existing files are skipped.

### Scan a Project

```bash
go run ./cmd/specharbor scan
```

`scan` performs an informational local project scan. It is stack-agnostic: user projects do not need to be Go projects.

The report can include detected ecosystems, package managers, test command hints, CI signals, container or deployment signals, SpecHarbor/OpenSpec signals, and notes. The command does not require flags or arguments.

### Show the Recommended Workflow

```bash
go run ./cmd/specharbor workflow
```

The installed command form is `specharbor workflow`. It prints the recommended OpenSpec/SDD workflow as read-only advisory text:

```text
SpecHarbor recommended workflow.
Title: OpenSpec/SDD agent-driven workflow

Steps:
1. spec-author - Spec Author Agent
2. architecture-reviewer - Architecture Reviewer Agent
3. implementer - Implementer Agent
4. test-engineer - Test Engineer Agent
5. change-reviewer - Change Reviewer Agent
6. commit - Commit
7. pull-request - Pull Request
8. merge - Merge
9. archive - Archive
```

The full output includes each step id, display name, purpose, mode, supported/advisory indicators, predecessor step ids, command suggestions, and safety notes. The suggestions connect the workflow to existing commands:

- `generate` creates or starts an OpenSpec change package.
- `validate` checks required OpenSpec change files and their content quality.
- `prompt --role ...` prints prompts for Spec Author Agent, Architecture Reviewer Agent, Implementer Agent, Test Engineer Agent, and Change Reviewer Agent.
- `review` checks the local change package and task checkbox completion.
- `archive` explicitly moves an accepted change to the archive.

Command suggestions are advisory and `specharbor workflow` does not execute them. Commit, Pull Request, and Merge remain manual steps; SpecHarbor does not commit, does not create PRs, and does not merge. This command is read-only and does not call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control automation, workflow execution, external commands, or remote automation.

### Generate a Change

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

`generate <change-id> --blank` creates the expected OpenSpec change structure with blank/manual starter content.

Interactive generation is a prompt layer over existing deterministic generation paths:

```bash
go run ./cmd/specharbor generate <change-id> --interactive
```

`<change-id>` remains required on the command line; interactive mode does not prompt for it. `--interactive` cannot be combined with direct generation or input flags such as `--blank`, `--template`, `--custom-template`, `--config-template`, `--guided`, `--hybrid`, `--ai-assisted`, `--agent-assisted`, `--from-file`, `--overwrite`, `--agent`, `--execute`, `--type`, `--title`, or `--summary`.

Interactive mode requires an interactive terminal. In CI, piped input, or other non-TTY contexts it fails immediately with:

```text
interactive mode requires a TTY
```

It does not prompt, hang, or write files in that case.

Supported interactive paths in this version are exactly:

- `blank`
- built-in template
- custom template
- config template
- hybrid

The first menu accepts the numbered choices `1` through `5` and stable keywords such as `blank`, `template`, `custom`, `config`, and `hybrid`. Direct guided generation remains available only through non-interactive flags. AI-assisted generation, agent-assisted generation, local agent runner execution, live runner output application, and raw remote URL entry are not offered by interactive prompts.

Prompt sequence:

- Blank asks only for the generation path.
- Built-in template asks for the built-in template name (`feature`, `bugfix`, `docs`, or `refactor`).
- Custom template asks for a custom template name, optional title, and optional summary.
- Config template asks for a config alias, optional title, and optional summary.
- Hybrid asks for exactly one source namespace (built-in template, custom template, or config template), the source value, required title, required summary, and optional type.

Invalid required answers retry up to three attempts. Empty required answers are invalid. Invalid non-empty hybrid type answers retry up to three attempts. Empty optional title, summary, and hybrid type answers are treated as omitted. Retry exhaustion exits non-zero and writes nothing.

Before any generation use case runs, interactive mode prints a deterministic summary:

```text
Interactive generation summary:
Change: add-login
Generation path: hybrid
Hybrid source: built-in template
Template: feature
Title: Add login
Summary: Add login support
Expected write target: openspec/changes/add-login/
Files: proposal.md, design.md, tasks.md, acceptance-criteria.md, risks.md
Validation: automatic yes
Safety:
- Writes are limited to OpenSpec change files.
- Production code will not be modified.
- Source-control commands will not be run.
- Workflow automation will not be triggered.
- Provider, LLM, and agent APIs will not be called.
- No auto-commit, auto-push, PR creation, merge, or archive will be performed.

Proceed? [y/N]:
```

Blank, built-in template, custom template, and config template summaries show `Validation: automatic no`. Hybrid summaries show `Validation: automatic yes`, preserving hybrid's existing automatic validation after generation. Interactive mode does not add a validation prompt and never auto-fixes validation findings.

Confirmation is trimmed and case-insensitive. `y` and `yes` proceed in any casing, including `Y`, `YES`, and `Yes`. `n` and `no` cancel in any casing, including `N`, `NO`, and `No`. Empty confirmation and EOF also cancel. Cancellation exits non-zero with `operation cancelled` and writes nothing. Unsupported confirmation answers retry up to three attempts; confirmation retry exhaustion writes nothing.

On confirmation, interactive mode delegates to the same behavior as the equivalent direct command:

```bash
go run ./cmd/specharbor generate add-blank --interactive
go run ./cmd/specharbor generate add-feature --interactive
go run ./cmd/specharbor generate add-payment-flow --interactive
go run ./cmd/specharbor generate add-configured-feature --interactive
go run ./cmd/specharbor generate add-login --interactive
```

Write behavior, existing-file preservation, template rendering, config alias lookup, remote-template safeguards, and validation behavior remain owned by the selected generation mode. Remote templates are reachable only through existing config aliases after confirmation; interactive mode does not ask for URLs or checksums and does not print credentials, query strings, fragments, auth headers, cookies, OAuth material, or environment-derived secrets.

Built-in template generation uses the same command with `--template <template-name>` and deterministic built-in starter content:

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

For example:

```bash
go run ./cmd/specharbor generate add-example-feature --template feature
```

Supported built-in templates are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Custom template generation uses project-local templates with `--custom-template <template-name>`:

```bash
go run ./cmd/specharbor generate <change-id> --custom-template <template-name>
```

For example:

```bash
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature
```

A custom template is a plain directory under the fixed project-local root `.specharbor/templates/<template-name>/` containing all five required OpenSpec change files:

```text
.specharbor/templates/<template-name>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Custom template behavior:

- All five files are required and must be non-empty; a missing template directory, missing required files, or an empty file fails with a clear error before anything is written.
- Unknown extra files and subdirectories in the template directory are ignored and never copied.
- Template content supports minimal deterministic variable substitution: `{{change_id}}` is always replaced; `{{title}}` and `{{summary}}` are replaced only when the optional `--title` and `--summary` flags are provided. Unresolved and unknown `{{...}}` tokens remain in the output verbatim.
- There are no conditionals, loops, includes, or any other templating language features, and templates are never executed.
- `--custom-template` is mutually exclusive with `--blank`, `--template`, `--guided`, and `--agent-assisted`.
- Custom template names must be safe single path segments (characters `[A-Za-z0-9._-]`, no `/` or `\`, no `..` sequences, no leading `.` or `-`, at most 128 characters); invalid names are rejected before any filesystem access.
- Built-in, custom, and config-driven templates are disjoint: `--template` resolves only the built-in set, `--custom-template` resolves only `.specharbor/templates/`, and `--config-template` resolves only `.specharbor/config.yml` aliases.
- Files are written only under `openspec/changes/<change-id>/`; existing files are skipped, never overwritten, and any template validation failure produces zero writes.
- Direct custom templates are project-local only: no marketplace, no arbitrary local paths, no template script execution, no shell execution, no network/provider behavior, and no production code writes.

After generation, run `go run ./cmd/specharbor validate <change-id>` to check the generated change; validation findings depend on the template's content quality, exactly as for hand-authored changes.

Custom template title/summary example:

```bash
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature --title "Add payments" --summary "Adds a payment flow."
```

Config-driven template generation uses aliases declared in `.specharbor/config.yml`:

```bash
go run ./cmd/specharbor generate <change-id> --config-template <alias>
go run ./cmd/specharbor generate <change-id> --config-template <alias> --title "<title>" --summary "<summary>"
```

Schema:

```yaml
version: 1

templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature

    default-feature:
      source: builtin
      template: feature

    service-feature:
      source: remote
      url: https://example.com/specharbor/templates/service-feature.zip
      checksum: sha256:<64-hex>
      format: zip
```

Rules:

- `version: 1` is required when `--config-template` is used; missing config, missing version, unsupported versions, invalid YAML, invalid alias entries, and missing aliases fail clearly.
- Supported source kinds are exactly `builtin`, `custom`, and `remote`.
- `builtin` resolves only supported built-in templates: `feature`, `bugfix`, `docs`, and `refactor`.
- `custom` resolves only `.specharbor/templates/<template-name>/` and uses the same required-file validation and `{{change_id}}`, `{{title}}`, and `{{summary}}` substitution as direct `--custom-template`.
- `remote` fetches one explicitly configured HTTPS ZIP URL, verifies the configured `sha256:<64-hex>` checksum before archive parsing, and writes the decoded OpenSpec files without rendering or executing template scripts.
- Remote aliases require `url`, `checksum`, and `format`; only `format: zip` is supported. `template` is invalid for remote aliases, and `url`, `checksum`, and `format` are invalid for `builtin` and `custom` aliases.
- Remote URLs must use HTTPS and include a host and path. HTTP, file, SSH, git, git+ssh, FTP, SCP-style targets, credentials/userinfo, query strings, fragments, whitespace/control characters, over-length URLs, and redirects are rejected.
- Remote ZIP bundles must contain exactly five non-empty root-level regular files: `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`. Nested paths, absolute paths, traversal, Windows drive paths, symlinks, executable entries, duplicate files, extra files, missing files, empty files, malformed ZIPs, oversized downloads, and oversized uncompressed content are rejected.
- Alias names must be safe single path segments: non-empty, at most 128 characters, characters `[A-Za-z0-9._-]`, no `/` or `\`, no absolute paths, no traversal or `..` sequence, no leading `.` or `-`.
- `--title` and `--summary` are optional with `--config-template`; they are passed through to the resolved custom template path and do not change built-in template output.
- `--config-template` is mutually exclusive with `--blank`, `--template`, `--custom-template`, `--guided`, `--agent-assisted`, `--ai-assisted`, and `--execute`.
- `--template`, `--custom-template`, and `--config-template` use separate namespaces. A built-in template, custom template, and config alias may share the same name without shadowing, fallback, or guessing.
- Remote templates have no persistent cache in this first version and do not support credentials, OAuth, auth headers, cookies, environment token expansion, git clone, marketplace search, provider APIs, script execution, shell execution, production code writes, source-control automation, auto-commit, PR, merge, or archive automation.
- Generated files are still limited to `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md` under `openspec/changes/<change-id>/`; existing files are skipped.

Hybrid generation combines one deterministic template source with required title and summary metadata:

```bash
go run ./cmd/specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
go run ./cmd/specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
go run ./cmd/specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
```

Hybrid source selection is explicit:

- `--template <name>` resolves only built-in templates.
- `--custom-template <name>` resolves only `.specharbor/templates/<name>/`.
- `--config-template <alias>` resolves only `.specharbor/config.yml` aliases.
- Exactly one source selector is required. Missing or multiple selectors fail before writes.
- There is no source guessing, fallback, or namespace shadowing. The same name may exist in all three namespaces because the flag selects the namespace.

Hybrid metadata rules:

- `--title` is required and must be non-empty after trimming.
- `--summary` is required and must be non-empty after trimming.
- `--type` is optional, but when provided must be exactly `feature`, `bugfix`, `docs`, or `refactor`.
- Direct built-in sources derive omitted type from the selected template. For example, `--hybrid --template feature` derives `type=feature`.
- Config aliases resolving to built-in templates derive omitted type from the resolved built-in template.
- A provided type must match a direct or resolved built-in template. `--hybrid --template feature --type feature` succeeds; `--hybrid --template feature --type bugfix` fails clearly and writes nothing.
- Custom sources, config custom aliases, and config remote aliases do not infer omitted type. If `--type` is omitted, `{{type}}` remains unresolved in those template contents. If `--type` is provided, `{{type}}` is replaced with that value.

Hybrid rendering replaces `{{change_id}}`, `{{title}}`, and `{{summary}}`. It replaces `{{type}}` only when a provided or built-in-derived effective type exists. Unknown or unresolved `{{...}}` tokens remain verbatim. Hybrid does not add conditionals, loops, functions, includes, hooks, shell commands, scripts, or executable template behavior.

Hybrid examples:

```bash
go run ./cmd/specharbor generate add-login \
  --hybrid \
  --template feature \
  --title "Add login" \
  --summary "Add an OpenSpec change for login"
```

The built-in example derives `type=feature`.

```bash
go run ./cmd/specharbor generate add-login \
  --hybrid \
  --template feature \
  --type bugfix \
  --title "Add login" \
  --summary "Add an OpenSpec change for login"
```

The mismatch example fails because `bugfix` does not match built-in template `feature`.

```bash
go run ./cmd/specharbor generate add-payment-flow \
  --hybrid \
  --custom-template api-feature \
  --title "Add payments" \
  --summary "Adds a payment flow."
```

The custom example does not infer type, so `{{type}}` remains unresolved unless `--type` is provided.

Config built-in example:

```bash
go run ./cmd/specharbor generate add-login \
  --hybrid \
  --config-template default-feature \
  --title "Add login" \
  --summary "Add login support"
```

Config custom example:

```bash
go run ./cmd/specharbor generate add-payment-flow \
  --hybrid \
  --config-template api-feature \
  --title "Add payments" \
  --summary "Adds a payment flow." \
  --type feature
```

Config remote example:

```bash
go run ./cmd/specharbor generate add-service \
  --hybrid \
  --config-template service-feature \
  --title "Add service" \
  --summary "Adds a service workflow."
```

Remote templates are available to hybrid only through `--config-template <alias>`. The alias must resolve to `source: remote` and keeps the existing remote safeguards unchanged: HTTPS only, no credentials, no query strings, no fragments, no redirects, checksum required, checksum verified before ZIP parsing, ZIP only, strict archive safety, no cache, no shell or script execution, no production code writes, and only the five OpenSpec change files are written.

After successful hybrid writes or a skip-only rerun, SpecHarbor runs the existing validation logic. The hybrid report includes validation status, required file count, error count, warning count, and findings. Validation warnings keep exit code `0`; validation errors are printed after the generation report and then the command exits non-zero. Validation never auto-fixes files.

Hybrid rejects `--blank`, `--guided`, `--ai-assisted`, `--agent-assisted`, `--from-file`, `--overwrite`, `--agent`, and `--execute`. AI overlay and live runner output application are intentionally out of scope for this first version. Hybrid does not call provider APIs, LLM APIs, local model APIs, agents, source-control tools, workflow tools, shell commands, or scripts. It writes no production code and performs no auto-commit, auto-push, pull request, merge, or archive automation.

Guided generation uses explicit CLI flags:

```bash
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

Guided generation is deterministic, local, and non-interactive. It does not prompt during command execution; it uses the supplied `--type`, `--title`, and `--summary` values.

Supported guided types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

AI-assisted generation imports a local file containing AI-authored OpenSpec Markdown in a strict delimiter format:

```bash
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file>
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --overwrite
```

The source file is local text only. It may be output that a user saved from an AI or agent tool, but SpecHarbor only reads the file from disk. It does not call provider APIs, remote AI services, local model APIs, OAuth, credentials, agents, shell commands, source-control tools, workflow tools, or network services.

The file must contain exactly these five file blocks, using exact delimiter lines:

```text
---FILE: proposal.md---
# Proposal

...
---END FILE---
---FILE: design.md---
# Design

...
---END FILE---
---FILE: tasks.md---
# Tasks

## Phase 1

- [ ] Implement the approved work.
---END FILE---
---FILE: acceptance-criteria.md---
# Acceptance Criteria

- The approved behavior is observable.
---END FILE---
---FILE: risks.md---
# Risks

## Risks

- A concrete risk is identified.

## Mitigations

- A mitigation is defined.
---END FILE---
```

Allowed filenames are exactly `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`. Unknown filenames, duplicate blocks, missing blocks, empty content, absolute paths, `..` traversal, nested paths, malformed block syntax, fenced wrapper formats, patch/diff formats, orphan end markers, unclosed blocks, and non-whitespace text outside file blocks are rejected before any writes.

AI-assisted generation writes only these files under:

```text
openspec/changes/<change-id>/
```

Existing files are skipped by default and reported as skipped. `--overwrite` is explicit and replaces existing required files only; symlink output targets are rejected instead of followed. All parsing and target preflight checks happen before file writes; malformed AI output writes nothing. After successful writes or skips, SpecHarbor runs the existing `validate <change-id>` logic and prints the validation status, error count, warning count, and findings. Validation warnings keep exit code `0`; validation errors are printed and then the command exits non-zero. Validation never auto-fixes files and AI-assisted generation never modifies production code.

Safety boundaries printed by the command are part of the contract: provider APIs called `no`, remote AI services called `no`, agent commands executed `no`, production code modified `no`, source-control commands run `no`, and auto-commit, auto-push, PR, merge, or archive `no`. Direct live runner application such as `--agent-assisted --execute --apply` is not implemented.

Agent-assisted spec authoring uses explicit CLI flags:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

For example:

```bash
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support"
```

Supported agent-assisted authoring types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Dry-run remains the default. Without `--execute`, agent-assisted spec authoring prints a deterministic authoring plan to stdout and prints a deterministic, copy-pasteable prompt to stdout.

The generated prompt is meant to help an external agent author or refine only the OpenSpec change package. Implementation remains a later step through the normal SpecHarbor workflow.

Dry-run agent-assisted spec authoring:

- writes no files;
- writes no prompt file;
- does not create or modify OpenSpec files;
- does not create or modify production code;
- does not execute agents or local agent commands;
- does not require a runner;
- does not resolve executable command mappings;
- does not call provider APIs;
- does not call local models;
- does not call network APIs;
- does not call source-control APIs;
- does not call workflow tools.

Recognized agent targets for dry-run are:

- `codex` - Codex
- `claude` - Claude Code
- `devin` - Devin
- `cursor` - Cursor
- `copilot` - GitHub Copilot
- `gemini` - Gemini CLI
- `roo` - Roo Code
- `windsurf` - Windsurf
- `aider` - Aider
- `generic` - Generic Agent

Unknown dry-run agents are rejected as an intentional validation tightening.

`--execute` is explicit and is supported only with `--agent-assisted`:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

Execute mode sends the same deterministic OpenSpec authoring prompt through stdin to a supported local command. The runner working directory is the current project root.

Supported executable local command mappings are:

- `codex -> codex`
- `claude -> claude`
- `devin -> devin`
- `cursor -> cursor`
- `copilot -> copilot`
- `gemini -> gemini`
- `roo -> roo`
- `windsurf -> windsurf`
- `aider -> aider`

`generic` is recognized for dry-run only. `--execute --agent generic` fails because generic execution requires a future config-driven command mapping.

Execute mode is run-and-report only:

- missing local commands produce startup errors with no runner result and no exit code;
- started commands with non-zero exit codes produce a full report, then SpecHarbor exits non-zero;
- stdout, stderr, exit code, and execution status are captured for started processes;
- SpecHarbor does not parse or apply output;
- stdout and stderr are displayed only, not parsed or applied;
- SpecHarbor does not write OpenSpec files from runner output;
- SpecHarbor does not modify production code from runner output;
- SpecHarbor does not auto-commit, auto-push, or auto-merge.

Provider APIs, IDE automation, OAuth, credentials, marketplace integrations, remote execution, source-control automation, and workflow automation remain out of scope. Local agent command behavior is controlled by the installed local tool.

Blank, built-in template, custom template, config-template, hybrid, guided, and AI-assisted generation create the same required OpenSpec change files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Built-in template content is deterministic, local, and generic starter content. Hybrid content starts from exactly one deterministic template source and applies explicit metadata substitution before validation. Guided content is deterministic, local starter content that includes the supplied title and summary. AI-assisted content comes only from the explicit local `--from-file` source after strict parsing.

Generated content is safe to edit after generation. Guided output does not mean SpecHarbor inferred project-specific requirements beyond the provided type, title, and summary.

Blank, built-in template, custom template, config-template, hybrid, and guided generation skip existing files and do not overwrite them. AI-assisted generation also skips existing files by default, and overwrites only with explicit `--overwrite`. If a change directory partially exists, running generation again recovers it by creating only the missing required files.

Copy-pasteable examples from the repository root:

```bash
go run ./cmd/specharbor generate add-interactive-change --interactive
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
go run ./cmd/specharbor generate fix-example-bug --template bugfix
go run ./cmd/specharbor generate update-example-docs --template docs
go run ./cmd/specharbor generate refactor-example-flow --template refactor
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature --title "Add payments" --summary "Adds a payment flow."
go run ./cmd/specharbor generate add-payment-flow --config-template api-feature
go run ./cmd/specharbor generate add-payment-flow --config-template api-feature --title "Add payments" --summary "Adds a payment flow."
go run ./cmd/specharbor generate add-login --hybrid --template feature --title "Add login" --summary "Add an OpenSpec change for login"
go run ./cmd/specharbor generate add-payment-flow --hybrid --custom-template api-feature --title "Add payments" --summary "Adds a payment flow."
go run ./cmd/specharbor generate add-service --hybrid --config-template service-feature --title "Add service" --summary "Adds a service workflow."
go run ./cmd/specharbor generate add-guided-feature --guided --type feature --title "Add guided feature" --summary "Create a guided OpenSpec change from explicit CLI inputs."
go run ./cmd/specharbor generate fix-guided-bug --guided --type bugfix --title "Fix guided bug" --summary "Describe the bugfix using deterministic guided starter content."
go run ./cmd/specharbor generate update-guided-docs --guided --type docs --title "Update guided docs" --summary "Document guided generation as implemented behavior."
go run ./cmd/specharbor generate refactor-guided-flow --guided --type refactor --title "Refactor guided flow" --summary "Describe a behavior-preserving refactor with explicit context."
go run ./cmd/specharbor generate add-ai-assisted-change --ai-assisted --from-file agent-output.txt
go run ./cmd/specharbor generate add-ai-assisted-change --ai-assisted --from-file agent-output.txt --overwrite
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support"
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support" --execute
```

### Validate a Change

```bash
go run ./cmd/specharbor validate add-example-feature
```

`validate <change-id>` runs deterministic, local, read-only validation over the change package under `openspec/changes/<change-id>/`. It never writes files, never modifies the change, and never calls the network, AI providers, agents, or source-control tooling.

It checks the five required OpenSpec files:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

Every finding carries a severity, a stable snake_case rule code, a message (with line numbers where relevant), and the file path.

Error findings (the change package is not usable downstream):

- `project_root_unavailable` - the OpenSpec project structure is missing.
- `change_directory_missing` - the change id is well formed but unknown.
- `required_file_missing` - one of the five required files is missing.
- `file_empty` - a required file is empty or whitespace-only. Other content findings for that file are suppressed.
- `file_missing_heading` - a required file has no markdown heading.
- `file_missing_body` - a required file contains only headings.
- `tasks_checkbox_missing` - `tasks.md` has no valid checkbox task.
- `tasks_checkbox_malformed` - a checkbox-like line (for example `- []`, `-[ ]`, `- [y]`, or `- [x]` without text) breaks the `- [ ] text` grammar; reported with line numbers.
- `acceptance_criteria_item_missing` - `acceptance-criteria.md` has no list or checkbox item with meaningful text; placeholder-only items (`N/A`, `...`, `?`, `TBD`, `TODO`, `FIXME`) do not count.

Warning findings (quality gaps that never fail the command):

- `placeholder_content` - standalone `TBD`/`TODO`/`FIXME`, placeholder-only list items (`N/A`, `...`, `?`), or `lorem ipsum`.
- `boilerplate_only_content` - the file still contains only known starter/boilerplate guidance lines and was never meaningfully edited.
- `proposal_section_missing` - no Problem, Goal, or Summary section.
- `design_section_missing` - no Overview, Approach, Design, Architecture, Technical Decisions, or Decisions section.
- `tasks_phase_heading_missing` - checkbox tasks exist but no level-2 phase heading.
- `tasks_all_completed` - every checkbox task is checked; confirm implementation evidence before review.
- `risks_mitigation_missing` - risks are listed without mitigation notes.
- `design_architecture_section_missing` - change files mention `internal/core`, `internal/adapters`, or `internal/platform` but `design.md` has no Architecture section.
- `tasks_documentation_task_missing` - `proposal.md` or `design.md` references the `specharbor` CLI but `tasks.md` has no documentation task.

Severity drives status and exit codes:

- Zero error findings: status `valid`, exit code `0`. Warnings alone never fail the command.
- One or more error findings: status `invalid`, exit code `1`.
- Missing or unsafe change ids (`..`, separators, absolute paths, leading `.` or `-`, characters outside `[A-Za-z0-9._-]`, more than 128 characters) are rejected with a clear command error before any filesystem access. Internal single dots such as `change.v1` are accepted.

Example valid output:

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 0
```

Example valid output with warnings (exit code `0`):

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 2

Warnings:
- [warning] placeholder_content: Placeholder marker "TBD" found (line 12) (openspec/changes/add-example-feature/design.md)
- [warning] risks_mitigation_missing: Risks are listed without mitigation notes. (openspec/changes/add-example-feature/risks.md)
```

Example invalid output (exit code `1`):

```text
SpecHarbor change is invalid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature

Errors:
- [error] required_file_missing: Missing required file: design.md (openspec/changes/add-example-feature/design.md)
- [error] tasks_checkbox_missing: No checkbox tasks found. (openspec/changes/add-example-feature/tasks.md)

Warnings:
- [warning] proposal_section_missing: No Problem, Goal, or Summary section found. (openspec/changes/add-example-feature/proposal.md)
```

Run validation before implementation (to confirm the change package is structurally ready), before review (to catch content gaps cheaply), and before opening a PR (to keep low-quality packages out of the shared workflow).

Intentional behavior changes from the earlier presence-only validation:

- Unusable change packages now fail: empty required files, files without a heading or body, tasks files without valid checkboxes, malformed checkbox lines, and acceptance-criteria files without a meaningful item produce errors and a non-zero exit.
- A freshly generated `--blank` or template change now validates as valid with `boilerplate_only_content` (and applicable `placeholder_content`) warnings instead of zero findings; the exit code stays `0`, so the documented `generate -> validate` flow keeps working.
- Change ids are validated more strictly and rejected before any filesystem access.
- The report replaces the single `Findings:` count with `Errors:` and `Warnings:` counts and groups findings by severity with file paths appended.

### Generate a Role Prompt

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

`prompt <change-id> --role <role>` prints an agent prompt for an existing OpenSpec change.

Supported roles:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

Use `--role` for prompt roles. Agent-target flags are not implemented.

### Review a Change

```bash
go run ./cmd/specharbor review add-example-feature
```

`review <change-id>` reviews the change package and task completion state. It exits non-zero when the review status is not approved.

### Archive a Change

```bash
go run ./cmd/specharbor archive add-example-feature
```

`archive <change-id>` moves a completed change from `openspec/changes/<change-id>/` to a dated archive path under `openspec/archive/<date>/<change-id>/`.

### Show Local Config

```bash
go run ./cmd/specharbor config show
go run ./cmd/specharbor config
```

`config show` and `config` are read-only. They read `.specharbor/config.yml` from the current project, support config version `1`, and print the local config report.

Current config behavior does not write config files and does not implement:

- `config get`
- `config set`
- `config unset`

## Normal Flow

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

For the detailed recommended nine-step workflow, run:

```bash
go run ./cmd/specharbor workflow
```

A typical local sequence is:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
```

Do not archive a change until the implementation is complete and reviewed.

## Planned Behavior

The following items are product direction, not implemented command behavior:

- config-driven generic runner commands;
- AI provider setup;
- provider API key management;
- config mutation commands;
- external workflow connectors.
