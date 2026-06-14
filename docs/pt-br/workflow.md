# Workflow

O SpecHarbor é baseado em um pacote de mudança OpenSpec. O objetivo é deixar o trabalho explícito antes de um agente ou colaborador iniciar alteração de código.

```text
Ideia -> Mudança OpenSpec -> Tasks -> Prompt de agente -> Implementação -> Review -> Archive
```

## Workflow recomendado

Execute o guia read-only com:

```bash
go run ./cmd/specharbor workflow
```

A forma instalada é `specharbor workflow`. Ele imprime o fluxo advisory de nove etapas do OpenSpec/SDD:

1. Spec Author Agent
2. Architecture Reviewer Agent
3. Implementer Agent
4. Test Engineer Agent
5. Change Reviewer Agent
6. Commit
7. Pull Request
8. Merge
9. Archive

Modelo de saída resumida:

```text
SpecHarbor recommended workflow.
Title: OpenSpec/SDD agent-driven workflow

Steps:
1. spec-author - Spec Author Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: none
   Purpose: Create or refine the OpenSpec change package.
   Commands:
   - specharbor generate <change-id> --guided ...
   - specharbor prompt <change-id> --role spec-author

6. commit - Commit
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Commands:
   - none
```

As sugestões do comando são advisory. `specharbor workflow` não executa comandos, não inspeciona status local de workflow e não decide próximo passo. A decisão de próximo passo fica intencionalmente a cargo da equipe.

O workflow se relaciona com comandos existentes da seguinte forma:

- `context discover` inspeciona opcionalmente fontes locais limitadas e reporta sinais de contexto classificados antes de briefing ou authoring.
- `context index` opcionalmente grava inventário metadado apenas local para fontes de contexto suportadas.
- `context retrieve` opcionalmente busca snippets contextuais de fonte e com limites quando há índice local atual.
- `context github` opcionalmente recupera snippets de contexto remoto com source attribution apenas quando chamado explicitamente.
- `context rag` opcionalmente pede resposta de provedor com base em contexto bounded e atribuído, apenas quando chamado explicitamente.
- `brief` opcionalmente coleta contexto confirmado antes da authoring quando o contexto local está ausente ou ambíguo.
- `generate` cria ou inicia o pacote de mudança OpenSpec para o Spec Author Agent.
- `validate` verifica arquivos OpenSpec obrigatórios antes de revisão ou implementação.
- `prompt --role ...` imprime prompts para `spec-author`, `architecture-reviewer`, `implementer`, `test-engineer` e `change-reviewer`.
- `review` verifica o checklist de tarefas e arquivos da mudança.
- `archive` arquiva uma mudança aceita.

`specharbor context discover` é descoberta local read-only. Ele classifica evidência de repositório em:

detected facts, suggested assumptions e contexto confirmado de projeto. Não executa comandos, não chama APIs de provedor, não faz descoberta remota, não gera índice e não executa RAG.

`specharbor context index` é inventário metadado local. Ele pode imprimir relatório, gravar `.specharbor/context-index.json` ou checar se esse índice gerado está obsoleto. Não recupera snippets, não ranqueia contexto, não cria embeddings/vetores, não executa RAG, não chama APIs remotas ou de provedor, não executa comandos/prompt, não roda agentes e não automatiza source control. O índice não é contexto confirmado do projeto e não substitui `brief` ou `context discover`.

`specharbor context retrieve --query "<query>"` é recuperação local/offline de um índice atual `.specharbor/context-index.json`. Ele lê apenas fontes indexadas locais com limites bounded de arquivo/fonte/resultado/tamanho e imprime correspondências atribuídas por fonte. Os resultados não são contexto confirmado do projeto, não entram automaticamente em prompts nesta versão e não usam embeddings, vetores, geração RAG, contexto remoto, APIs de provedor, execução de comando, execução de prompt, execução de agente ou automação de source control.

`specharbor context github --repo owner/name --query "<query>"` é recuperação remota explícita e read-only no GitHub. Ele usa rede apenas nesse comando, pode usar `SPECHARBOR_GITHUB_TOKEN` opcionalmente, busca apenas fontes aprovadas e imprime snippets/summary com attribution. Resultados remotos não são contexto confirmado do projeto, não são gravados em `.specharbor/context-index.json`, não entram automaticamente em prompts e não usam embeddings, vetores, geração RAG, APIs de provedor, execução de comando, execução de agente, mutation de GitHub ou automação de source control.

`specharbor context rag --query "<query>" --provider openai` é a resposta assistida por provedor sobre contexto bounded. Por padrão usa recuperação local, só inclui GitHub se `--from github --repo owner/name` for passado, lê `SPECHARBOR_OPENAI_API_KEY` apenas nesse comando e imprime resposta gerada com lista de fontes. As respostas RAG não são contexto confirmado do projeto, não são persistidas, não entram automaticamente em prompts e não fazem automação de source control, execução de shell, execução de agente, mutation de GitHub, embeddings, vetores ou escrita automática de arquivos.

`specharbor brief` grava `.specharbor/project-brief.md` somente após confirmação interativa. `specharbor brief --update` é o fluxo de manutenção explícita para um brief existente: preserva valores confirmados por padrão, mostra fatos detectados e suposições como itens de revisão, antecipa mudanças e grava apenas após confirmação final. Briefing é coleta/m manutenção explícita de contexto, não indexação, não RAG, não integração com provedor, não execução de agente, não automação de source control ou remota.

Commit, Pull Request e Merge permanecem manuais. Fora do `context github` explicitamente read-only e do `context rag` com chamada de provedor explícita, o SpecHarbor não commita, não faz push, não cria PR, não faz merge, não chama GitHub, não chama GitLab, não inspeciona CI, não chama APIs de provedor, não chama CLIs de agente, não executa automação de source-control, não executa workflow e não faz automação remota.

## Pacote de mudança

Uma mudança vive em `openspec/changes/<change-id>/`:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Responsabilidades dos arquivos:

- `proposal.md`: descreve problema, meta, escopo, fora de escopo e critérios de sucesso.
- `design.md`: descreve abordagem técnica e tradeoffs importantes.
- `tasks.md`: lista implementação e verificação.
- `acceptance-criteria.md`: define condições observáveis de conclusão.
- `risks.md`: registra riscos conhecidos e mitigações.

Os arquivos não servem para burocracia. Eles restringem escopo de implementação, facilitam revisão e dão trabalho concreto para agentes de codificação seguirem.

## Da ideia à mudança

Comece com um id de mudança que descreva o trabalho:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

Preencha os arquivos gerados antes da implementação significativa. O gerador blank cria apenas estrutura; ele não infere requisitos.

## Validar antes da implementação

```bash
go run ./cmd/specharbor validate add-example-feature
```

A validação checa se os arquivos OpenSpec obrigatórios existem e sua qualidade de estrutura. Não prova que design está correto.

## Gerar prompt por papel

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

Os prompts de papel apontam agente para regras do repositório, arquitetura, mudança ativa e contexto classificado disponível. A seção de contexto mantém contexto confirmado, fatos detectados e suposições separadas; suposições continuam suposições.

Workflows de agent-assisted não exigem chaves de API de provedor, pois o SpecHarbor apenas gera prompts para agentes externos consumirem, sem executar comandos ou agentes.

## Implementar e revisar

A implementação deve permanecer dentro do escopo da mudança ativa. Para este repositório, código Go deve seguir as fronteiras da arquitetura descritas em `openspec/specs/architecture/spec.md`.

Depois da implementação:

```bash
go run ./cmd/specharbor review add-example-feature
go test ./...
```

Revise o diff antes de finalizar. Atualize `tasks.md` apenas para trabalho realmente concluído.

## Arquivar

Arquive apenas após a mudança estar completa:

```bash
go run ./cmd/specharbor archive add-example-feature
```

Arquivar move a mudança ativa para área datada de archive para não confundir trabalho concluído com trabalho ativo.

## Dogfooding

O SpecHarbor usa OpenSpec changes para seu próprio desenvolvimento. Trabalho significativo nesse repositório deve iniciar em `openspec/changes/` e permanecer restrito ao pacote da mudança.
