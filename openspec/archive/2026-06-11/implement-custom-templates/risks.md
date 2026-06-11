# Risks: Implement Custom Templates

## Path safety regressions

- User-supplied template names become part of filesystem paths for the first time on the read side. A validation gap could allow reads outside the project root through traversal sequences or separators.
- Residual risk: a symlink placed inside `.specharbor/templates/` by someone with write access to the project could point outside the project root; this change relies on name validation and root-relative resolution, consistent with how existing change-id paths are handled, and does not add symlink resolution defenses.

## Confusion between built-in and custom templates

- Users may expect `--template <name>` to find their custom template, or expect a custom template named `feature` to replace the built-in one, and be surprised by the disjoint-flag model.
- A future config-driven registry (`implement-config-driven-templates`) could be constrained by the flag split if it later wants unified resolution.

## Template content quality

- Custom templates are user-authored: low-quality, placeholder-heavy, or stale templates will deterministically reproduce their problems in every generated change, and `specharbor validate` will flag the generated changes rather than the template source.
- Unresolved `{{title}}`/`{{summary}}` tokens left verbatim when flags are omitted may read as broken output to users who expected substitution or an error.

## Scope creep toward a templating language

- Variable substitution invites requests for conditionals, loops, includes, per-template variables, or front matter, which would turn static templates into an execution surface.

## Strictness trade-offs

- Requiring all five files and rejecting empty files may frustrate users who want partial or skeleton templates.
- Ignoring extra files silently may hide typos (for example a template author writing `proposals.md` and not noticing the real `proposal.md` is missing until the missing-file error appears).

## Behavior drift in shared code paths

- Extending `GenerateChange`, `GenerationResult`, and the CLI argument parser risks accidental changes to blank, built-in template, guided, or agent-assisted behavior, which downstream users and scripts depend on.

## Mitigations

- Validate the `CustomTemplateName` value object in the domain before any filesystem access, mirroring the proven `ChangeID` rules, and route every read through the port with root-relative paths only; adapter tests pin that reads resolve under the project root.
- Keep namespaces disjoint by design and make output self-describing: the custom report labels the template as custom and prints its source path, while `--template` output and errors stay byte-identical; documentation states explicitly that custom templates never shadow built-ins and that no registry exists yet.
- Document that generated changes should be checked with `specharbor validate <change-id>` and that template quality is the template author's responsibility; document the leave-as-is behavior for unresolved variables so it is expected, not surprising.
- Hold the line on scope in the spec: the design names conditionals, loops, functions, includes, and execution as explicit non-goals, so any richer templating requires a new OpenSpec change rather than incremental additions.
- Make the missing-file error aggregate all missing filenames and name the expected directory, so typo'd extra files surface immediately through the missing required file they failed to provide.
- Protect shared code paths with regression tests for every existing generation mode and CLI surface, plus architecture boundary tests, before and after the implementation work.
- Sequence safely: load, validate, and render the template fully before any write, so failures leave no partial change directory behind.
