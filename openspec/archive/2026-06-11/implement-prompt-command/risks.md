# Risks: Implement Prompt Command

## Architecture leakage

Prompt generation touches CLI parsing, filesystem/template loading, and rendering. The main risk is placing template lookup, supported-role policy, or rendering decisions directly inside the CLI adapter.

Mitigation:

- Keep role validation and prompt orchestration in a core use case.
- Keep concrete template loading and rendering in adapters.
- Keep the CLI limited to argument parsing, use case invocation, and stdout formatting.

## Accidental AI or agent integration

Because prompts are intended for coding agents, the implementation could prematurely call an AI provider, execute an external agent, or require credentials.

Mitigation:

- Treat this command as a pure local renderer.
- Do not add provider SDKs, network calls, external process execution, or workflow connectors.
- Cover behavior with tests that exercise local rendering only.

## Template runtime dependency

Loading templates from repository-relative paths may fail when the CLI is installed and run outside the source checkout.

Mitigation:

- Prefer embedded templates if the project is expected to support installed-binary usage.
- If repository-local loading is used for this change, keep the adapter isolated so a future change can replace it with embedded or configured loading without changing the use case.

## Placeholder drift

Templates may gain placeholders that the use case does not provide, producing prompts with raw `{{...}}` markers.

Mitigation:

- Render the known placeholders required by existing role templates: `change_id` and `task`.
- Add tests that successful output contains no raw `{{change_id}}` or `{{task}}`.
- Return render errors for unresolved placeholders if the selected renderer supports that behavior.

## Role name ambiguity

Display names such as "Spec Author" and CLI role identifiers such as `spec-author` may diverge.

Mitigation:

- Use the kebab-case identifiers as stable command/API values.
- Validate roles against a single supported-role list.
- Map role identifiers directly to role template filenames.

## Overbuilding prompt workflows

There is a risk of adding task flags, custom template paths, output formats, validation, prompt dispatch, or agent integrations before the basic local renderer exists.

Mitigation:

- Limit this change to `specharbor prompt <change-id> --role <role>`.
- Defer custom template directories, richer prompt data, validation, and external integrations to separate OpenSpec changes.
