# Risks: Implement Advanced Validation

## Risks

### False positives in content heuristics

Deterministic text rules can still misfire: legitimate prose containing the word `TODO` in a quoted command, an acceptance criterion legitimately reading `N/A`, or a tasks file that intentionally has no phases would be flagged.

**Mitigation:** every heuristic rule is a warning, never an error, so a misfire cannot block validation or change exit codes. Trigger definitions are exact (standalone uppercase tokens, whole-item matches) and covered by negative test cases. If real-world misfires accumulate, a future change can add config-driven rule toggles.

### Breaking existing workflows with stricter errors

Packages that previously validated (empty files, checkbox-free tasks) now fail. Users or CI scripts depending on the old lenient behavior will see new non-zero exits.

**Mitigation:** the error set is limited to packages that are unusable downstream anyway; the change is documented in the proposal's Compatibility section, README, and docs/usage.md. Quality findings stay warnings, so the documented `generate --blank` then `validate` flow keeps exiting `0`. Regression tests assert every existing well-authored change in this repository still passes.

### Output-format coupling

Scripts that parse the current text output (`Findings: N` line, ungrouped list) will need updating, and the new format becomes a de facto contract.

**Mitigation:** the format change is intentional and documented with before/after examples. Stable rule codes give scripts a durable anchor. Machine consumers are explicitly pointed to the future JSON/reporting change rather than encouraged to scrape text.

### Boilerplate detection drifting from templates

`boilerplate_only_content` compares against the canonical starter-marker set owned by `internal/core/domain` — deliberately not derived from the adapter template files, because domain must not import `internal/adapters/templates` or read template files. If template wording changes later and the domain markers are not updated, unedited generated content could stop being detected as boilerplate.

**Mitigation:** the duplication is intentional and guarded. Domain tests pin the marker set, and an adapter-layer drift-guard test (adapters may import domain) feeds freshly generated blank starter content to the domain rule and asserts it is recognized as boilerplate. A template wording change therefore fails that test immediately, forcing an intentional domain-marker update through domain tests and a spec change instead of a silent behavior change. The trade-off — keeping the architecture boundary clean at the cost of one deliberately duplicated, small, stable string set — is accepted.

### Rule sprawl in the use case

Adding many rules risks a bloated `ValidateChange` with embedded business logic, eroding the hexagonal boundary.

**Mitigation:** rules live in domain as an ordered rule chain; the use case only loads content and iterates. Architecture tests and the existing import-boundary tests guard the dependency direction.

### Cross-file rules over-warning on meta-changes

Spec-authoring or docs-only changes that merely mention `internal/core` in prose (like this one) will trigger the architecture-section warning even when no code is touched.

**Mitigation:** warning severity only; the warning is satisfied by adding an Architecture heading to design.md, which is good authoring practice for such changes anyway. The rule's trigger and rationale are documented so authors understand the nudge.

### Port extension ripples through tests

Adding `ReadFile` to `ValidationFileSystem` forces every existing mock of the port to grow a method.

**Mitigation:** the port has few implementors (local filesystem adapter plus test doubles in the validate use case tests); the adapter already implements `ReadFile`, so the change is mechanical and caught at compile time.

## Trade-offs accepted

- Two severity levels instead of three: simpler model now; `info` can be added later if a rule genuinely needs it.
- Domain-owned boilerplate markers intentionally duplicate starter wording that also exists in adapter templates: a small, stable, test-guarded string set in exchange for keeping core free of adapter and template dependencies.
- No JSON output in this change: CI consumers wait for a dedicated reporting change, in exchange for a smaller, safer scope here.
- Markdown-only architecture awareness: weaker than parsing Go code, but deterministic, fast, and non-overlapping with the existing architecture tests.
