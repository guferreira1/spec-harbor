# Contribuição

O SpecHarbor usa changes OpenSpec para manter o trabalho de implementação com escopo e revisável. Recursos de recurso, arquitetura, comportamento, testes, CI ou documentação devem começar com uma mudança OpenSpec em `openspec/changes/<change-id>/`.

## Leitura obrigatória

Antes de implementar uma mudança significativa, contribuidores e agentes devem ler:

- `AGENTS.md`
- `.specharbor/rules/global.md`
- a regra de papel relevante em `.specharbor/rules/`
- `openspec/project.md`
- `openspec/specs/architecture/spec.md`
- a change ativa em `openspec/changes/<change-id>/`

## Arquivos OpenSpec

Cada mudança ativa deve explicar o trabalho com:

- `proposal.md`: problema, meta, escopo, out-of-scope e critérios de sucesso.
- `design.md`: abordagem técnica e trade-offs.
- `tasks.md`: lista de implementação e verificação.
- `acceptance-criteria.md`: critérios de conclusão observáveis.
- `risks.md`: riscos e mitigação.

Mantenha esses arquivos alinhados. Atualize `tasks.md` apenas para trabalho que realmente foi concluído.

## Escopo

A implementação deve permanecer dentro da change ativa. Não adicione comandos, refatorações, docs, testes ou mudanças de CI fora do escopo só por conveniência.

Não misture documentação, código, testes e CI em uma única change ampla, a menos que a OpenSpec ativa exija explicitamente essa mistura.

## Fronteiras de Arquitetura

O SpecHarbor segue a arquitetura hexagonal:

- O código de domínio pertence a `internal/core/domain`.
- Ports pertencem a `internal/core/ports`.
- Casos de uso pertencem a `internal/core/usecase`.
- Implementações concretas pertencem a `internal/adapters`.
- Core não deve importar adapters.
- Casos de uso devem depender de interfaces.
- O código da CLI não deve conter regras de negócio.

Use `openspec/specs/architecture/spec.md` como fonte de detalhe em vez de duplicar toda a especificação de arquitetura aqui.

## Verificação

Execute os testes após a implementação:

```bash
go test ./...
```

Revise os diff antes de finalizar:

```bash
git status --short
git diff --stat
git diff --name-only
git diff
```

Verifique se o diff corresponde à mudança ativa e que `tasks.md` não afirma trabalho não concluído.

## Higiene de Branches

Use branches ou worktrees separados para `changes` OpenSpec separadas. Prefira um `change-id` ativo por branch.

Evite que múltiplos agentes editem os mesmos arquivos sem coordenação explícita. Se branches de feature paralelas forem mescladas, reconcilie a documentação depois para que listas de comandos e rótulos de status correspondam à branch estabilizada.

Antes de começar e antes de finalizar, execute:

```bash
git status --short
```

Não reverta, sobrescreva ou limpe mudanças sujas no worktree que não estão relacionadas. Trate mudanças locais não relacionadas como trabalho de outra pessoa, a menos que a coordenação diga o contrário.

Não descreva `scan`, `config`, CI ou outras features recentemente trabalhadas como completas até que a branch relevante seja mesclada na branch em documentação e o comportamento esteja verificado ali.
