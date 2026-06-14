# Papéis dos Agentes

O SpecHarbor usa templates de prompt por papel para que sessões de agentes diferentes possam trabalhar na mesma mudança OpenSpec sem assumir as mesmas responsabilidades.

Gere o prompt de papel na raiz do repositório:

```bash
go run ./cmd/specharbor prompt implement-config-foundation --role implementer
```

Papéis suportados:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

Esses cinco IDs também aparecem como etapas de `agent-assisted` no `specharbor workflow`.

Os prompts de papel podem incluir uma seção `## Project Context` quando houver contexto local confirmado ou descoberto. A seção separa contexto confirmado pelo usuário, fatos detectados e suposições sugeridas. O contexto confirmado em `.specharbor/project-brief.md` tem precedência sobre fatos detectados; suposições continuam suposições e não viram fatos. A seção é limitada, inclui evidências de fonte/confiança e não executa comandos, não roda agentes, não chama APIs de provedor, não faz RAG, não gera índice do repositório e não usa descoberta remota.

## Agente Autor de Especificação

Cria ou refina os arquivos de mudança OpenSpec. Esse papel foca em `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md` e `risks.md` e não deve implementar a mudança.

## Agente Revisor de Arquitetura

Revisa specs ou diffs contra o contrato arquitetural. Esse papel valida se o trabalho proposto respeita fronteiras como domínio, ports, use cases, adapters e responsabilidades da CLI.

## Agente Implementador

Aplica uma mudança OpenSpec aprovada. Esse papel deve ler a mudança ativa, manter as alterações no escopo descrito, atualizar `tasks.md` apenas para trabalho concluído e executar as verificações solicitadas.

## Agente Engenheiro de Teste

Adiciona ou atualiza testes focados quando a mudança ativa solicitar trabalho de teste. Esse papel deve manter os testes alinhados com o comportamento especificado e evitar churn fora de escopo.

## Agente Revisor de Mudanças

Revisa o diff final contra a mudança OpenSpec ativa. Esse papel deve priorizar desvio de escopo, violações de arquitetura, status de tarefas obsoletas, ausência de verificação e alegações que não batem com a implementação.

## Fluxo recomendado

```text
Spec Author -> Architecture Reviewer -> Implementer -> Test Engineer -> Change Reviewer
```

O fluxo recomendado mais amplo continua manualmente por Commit, Pull Request, Merge e Archive. `specharbor workflow` mostra a sequência completa de orientação sem executar comandos ou automatizar etapas de controle de fonte.

Pull Request e Archive são etapas de workflow, não papéis de prompt suportados nesta versão.

Para mudanças pequenas:

```text
Spec Author -> Implementer -> Change Reviewer
```

Regras globais e por papel estão em `.specharbor/rules/`. A documentação deve linkar para essas regras em vez de copiar todas as instruções nesta página.

Workflows com `agent-assisted` não exigem chaves de API de provedor. O SpecHarbor apenas gera prompts para serem colados em uma ferramenta de agente externa.
