# Contribuição

O SpecHarbor usa changes OpenSpec para manter o trabalho de implementação com escopo e revisável. Recursos de recurso, arquitetura, comportamento, testes, CI ou documentação devem começar com uma mudança OpenSpec em `openspec/changes/<change-id>/`.

## Required Reading

Antes de implementar uma mudança significativa, contribuidores e agentes devem ler:

- `AGENTS.md`
- `.specharbor/rules/global.md`
- the relevant role rule under `.specharbor/rules/`
- `openspec/project.md`
- `openspec/specs/architecture/spec.md`
- the active change under `openspec/changes/<change-id>/`

## Arquivos OpenSpec

Cada mudança ativa deve explicar o trabalho com:

- `proposal.md`: problem, goal, scope, out-of-scope work, and success criteria.
- `design.md`: technical approach and tradeoffs.
- `tasks.md`: implementation and verification checklist.
- `acceptance-criteria.md`: observable completion criteria.
- `risks.md`: risks and mitigations.

Mantenha esses arquivos alinhados. Atualize `tasks.md` apenas para trabalho que realmente foi concluído.

## Escopo

A implementação deve permanecer dentro da change ativa. Não adicione comandos, refatorações, docs, testes ou mudanças de CI fora do escopo só por conveniência.

Não misture documentação, código, testes e CI em uma única change ampla, a menos que a OpenSpec ativa exija explicitamente essa mistura.

## Fronteiras de Arquitetura

O SpecHarbor segue a arquitetura hexagonal:

- Domain code belongs in `internal/core/domain`.
- Ports belong in `internal/core/ports`.
- Use cases belong in `internal/core/usecase`.
- Concrete implementations belong in `internal/adapters`.
- Core must not import adapters.
- Use cases must depend on interfaces.
- CLI code must not contain business rules.

Use `openspec/specs/architecture/spec.md` as the source of detail instead of duplicating the full architecture spec here.

## Verificação

Execute os testes após a implementação:

```bash
go test ./...
```

Inspect diffs before finalizing work:

```bash
git status --short
git diff --stat
git diff --name-only
git diff
```

Check that the diff matches the active change and that `tasks.md` does not claim work that was not completed.

## Higiene de Branches

Use branches ou worktrees separados para changes OpenSpec separados. Prefira um `change-id` ativo por branch.

Avoid multiple agents editing the same files without explicit coordination. If parallel feature branches are merged, reconcile the docs afterward so command lists and status labels match the stabilized branch.

Before starting and before finalizing, run:

```bash
git status --short
```

Do not revert, overwrite, or clean up unrelated dirty worktree changes. Treat unrelated local changes as someone else's work unless coordination says otherwise.

Do not describe scan, config, CI, or other recently worked features as complete until the relevant branch is merged into the branch being documented and the behavior is verified there.
