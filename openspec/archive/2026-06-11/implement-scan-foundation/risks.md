# Risks: Implement Scan Foundation

## Architecture leakage

Scan touches CLI parsing, project-root discovery, filesystem access, a detection catalog, categorization, signal display, test command hint derivation, note computation, and report formatting. The main risk is placing detection rules or report formatting directly in the CLI adapter or filesystem adapter.

Mitigation:

- Keep CLI responsibilities limited to argument parsing, current-working-directory lookup, dependency construction, and report formatting.
- Keep scan orchestration in `internal/core/usecase`.
- Keep the scan result, detection catalog, signal display, hint derivation, and note rules in `internal/core/domain`.
- Use a scan-specific filesystem port from `internal/core/ports`.
- Keep concrete filesystem behavior in `internal/adapters/filesystem`.
- Keep detection catalog data, categorization, and note computation out of the CLI and filesystem adapter.

## Stack-specific assumptions

The scanner runs inside a Go repository, so it is tempting to special-case Go or to test against the SpecHarbor repository's own files. That would make the scanner biased and the tests fragile.

Mitigation:

- Treat Go as one ecosystem among many in the deterministic catalog.
- Do not branch behavior on whether the host project is Go.
- Build scan use case and CLI tests against temporary directories with controlled signal files.
- Do not assert against the real SpecHarbor repository layout in tests.
- Cover non-Go signals such as `package.json`, `Dockerfile`, and `.gitlab-ci.yml`.

## Accidental recursive or deep scanning

The feature must inspect only top-level or conventional paths, but extension-based detection (`.csproj`, `.sln`) tempts a directory walk, and signal detection tempts recursive search.

Mitigation:

- Probe fixed catalog paths with file and directory existence checks.
- Perform exactly one non-recursive listing of the project root, used only for suffix rules and root validation.
- Do not list any directory other than the project root.
- Do not recurse into subdirectories.
- Add tests proving only the project root is listed and no recursion occurs.

## Detection catalog drift and policy confusion

Validation, generation, and review already use `domain.RequiredOpenSpecChangeFiles()`. Reusing that list as a scan catalog, or duplicating signal paths across the CLI and domain, could blur scan presence-detection with required-change-file policy.

Mitigation:

- Keep the scan detection catalog as a single deterministic source of truth in the domain.
- Do not reuse `domain.RequiredOpenSpecChangeFiles()` as a detection catalog.
- Report OpenSpec/SpecHarbor presence without enforcing change-file policy.
- Do not put catalog data in the CLI or filesystem adapter.

## Non-determinism in output

A scan report that depends on filesystem iteration order, discovered filenames, or map ordering would produce unstable output and flaky tests.

Mitigation:

- Drive detection and grouping from the ordered catalog, not from filesystem listing order.
- Display `file-suffix` matches as the suffix pattern rather than a discovered filename.
- Collect and de-duplicate test command hints in first-seen catalog order.
- Print sections in a fixed order.
- Add tests asserting exact populated and empty output shapes.

## Scan treated as a pass/fail gate

Review and validate return non-zero for non-approved or invalid results. Scan is informational, so reusing a pass/fail mental model could make a project with no signals look like a failure.

Mitigation:

- Treat scan as informational with no `invalid` status.
- Return zero whenever a report is produced, including for empty results.
- Reserve non-zero exit and execution errors for an unavailable or unreadable root, a current-working-directory failure, or a filesystem failure.
- Always print every section, using `- none detected` for empty sections and a note for the no-signals case.

## Note rules growing unbounded

Notes are a natural place to add many "missing X" observations, which could become noisy, opinionated, or stack-biased.

Mitigation:

- Limit the initial note rules to the two documented cases: no signals at all, and container/deployment present without Kubernetes or Helm.
- Keep note rules deterministic and pure.
- Document richer notes as future work.
- Test both note rules and the case where no note applies.

## Overbuilding future scanner behavior

SpecHarbor will eventually need richer scanning: deep manifest parsing, framework detection, dependency analysis, AI-assisted scanning, and machine-readable output. Adding scanner registries, strategy interfaces, plugin chains, or provider abstractions now would add unused surface area.

Mitigation:

- Implement only `specharbor scan`.
- Model a structured result, a deterministic catalog, and a pure assembler that future features can extend.
- Avoid exported scanner registries, factories, chains, strategy interfaces, plugin abstractions, provider abstractions, AI abstractions, agent abstractions, and workflow connectors until concrete behavior requires them.
- Do not add `--json`, `--deep`, `--path`, `--ai`, or other flag plumbing for future modes.

## Underbuilding the scan foundation

A single CLI function that checks a few files directly could satisfy the first output, but it would make future scan behavior harder to add cleanly and would leak policy into the adapter.

Mitigation:

- Introduce the scan result domain concept and a deterministic detection catalog.
- Add a scan-specific filesystem port.
- Keep scan execution in a use case with focused input validation and probing.
- Format the CLI report from the structured result.
- Cover use case behavior with fake ports.

## Leftover placeholder abstraction

The unused `internal/core/domain/project_context.go` placeholder predates concrete scan behavior. Leaving it alongside a new scan result type would keep dead code and two competing project-context models.

Mitigation:

- Replace the `ProjectContext` placeholder with the explicit scan result model, or remove it once superseded.
- Limit out-of-change file edits to this single placeholder, which is directly the scan/project-context concept being replaced.
- Confirm no references to `ProjectContext` remain after the change.

## Filesystem error classification

The scan must distinguish an absent signal (a normal, expected outcome) from a filesystem failure (an execution error). Mapping every error to an empty detection could hide real problems; mapping every absence to an error would make scan unusable.

Mitigation:

- Treat a missing file or directory as a non-match, not an error.
- Treat an unreadable or unavailable project root as an execution error.
- Treat a failed existence check as an execution error.
- Cover both non-match and execution-error paths in use case tests.

## Accidental external integration

Scan is naturally adjacent to AI, agents, source control, workflow systems, security scanning, and dependency analysis. Accidentally pulling those concerns into this first change would increase complexity and credentials risk.

Mitigation:

- Do not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, source-control host APIs, network APIs, or external processes.
- Do not write files, install packages, or perform security, license, or vulnerability scanning.
- Do not require provider API keys, local model credentials, agent credentials, source-control credentials, or workflow credentials.
- Keep tests local and deterministic.
- Document richer scan modes as future OpenSpec changes.

## Accidental behavior changes

Scan is adjacent to init, generate, validate, prompt, review, and archive because all operate through the CLI registry and filesystem adapter. Broad helper extraction or CLI refactoring could accidentally change existing commands.

Mitigation:

- Keep this change scoped to scan.
- Replace only the `scan` placeholder in the command registry.
- Add the listing method to the filesystem adapter without changing existing methods.
- Do not modify init, generate, validate, prompt, review, archive, or config behavior.
- Preserve existing tests and add regression coverage for existing CLI commands.
- Run `go test ./...`.
