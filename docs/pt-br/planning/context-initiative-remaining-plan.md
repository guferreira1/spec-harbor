# Plano restante da Context Initiative

## Objetivo

Este documento descreve o plano restante do Context Initiative. É apenas um artefato de planejamento. Ele não cria uma mudança OpenSpec ativa, não implementa código de produto e não altera comportamento de produção.

A base já concluída já oferece:

- `specharbor brief` para contexto de projeto confirmado em `.specharbor/project-brief.md`.
- `specharbor context discover` para descoberta local/offline com contexto classificado.
- saída de `specharbor prompt <change-id> --role <role>` ciente de contexto para os cinco papéis suportados.

Os recursos restantes devem estender esse ciclo sem enfraquecer a distinção atual entre contexto confirmado pelo usuário, fatos detectados e suposições sugeridas.

## Estratégia de Branches

Use uma branch para este documento de planejamento:

```text
chore/plan-remaining-context-initiative
```

Use branches separadas para implementações reais:

```text
feat/implement-project-brief-merge-and-update
feat/implement-repository-context-index
feat/implement-local-context-retrieval
feat/implement-github-remote-context
feat/implement-context-rag-provider
```

Use branches de arquivo (archive) separadas após cada feature mesclada:

```text
chore/archive-implement-project-brief-merge-and-update
chore/archive-implement-repository-context-index
chore/archive-implement-local-context-retrieval
chore/archive-implement-github-remote-context
chore/archive-implement-context-rag-provider
```

Uma única branch é aceitável para este plano porque ele é um artefato de documentação sem alteração de produção. Uma branch única não é ideal para implementar todas as cinco features juntas porque elas competiriam pelos mesmos modelos de domínio, ports, use cases, comandos de contexto do CLI, pontos de integração de prompts, documentação e testes. Combiná-las tornaria mais difícil manter escopo, revisar, isolar falhas e preservar histórico de archive.

Cada branch de implementação deve conter exatamente uma mudança OpenSpec ativa e as edições de produto da feature. Após a integração dessa feature, deve-se arquivá-la a partir de uma branch fresca baseada no `main` atualizado, para que o housekeeping de archive fique separado da revisão de implementação.

## Estratégia de subagentes

Use um agente principal (coordenador) com subagentes de planejamento.

O Agente Principal / Coordenador tem responsabilidade de:

- sequenciamento entre as cinco features;
- validação de branch/worktree;
- controle de escopo OpenSpec;
- decisões arquiteturais e de dependências;
- integração das recomendações de subagentes;
- aceite final de cada spec antes da implementação;
- validação final antes do PR;
- sequenciamento de arquivo (archive) após merge.

Subagentes podem:

- propor orientações de spec OpenSpec;
- identificar riscos e critérios de aceitação;
- propor quebra de tarefas;
- inspecionar dependências e sobreposição prováveis;
- sugerir testes focados e mudanças de documentação.

Subagentes não devem:

- implementar código de produção simultaneamente nas cinco features;
- editar arquivos de produção compartilhados fora do sequenciamento do coordenador;
- criar mudanças OpenSpec ativas sem o coordenador designar explicitamente uma feature;
- alterar release, npm, Homebrew, `install.sh`, publicação, tag, merge ou automação de controle de fonte.

Cada subagente deve seguir disciplina OpenSpec/SDD. O coordenador valida toda saída antes de iniciar implementação. A paralelização é para planejamento e revisão; edições de produção devem seguir a sequência por feature.

## Ordem de execução recomendada

Use esta ordem:

1. `implement-project-brief-merge-and-update`
2. `implement-repository-context-index`
3. `implement-local-context-retrieval`
4. `implement-github-remote-context`
5. `implement-context-rag-provider`

Raciocínio de dependência:

- Merge/update de brief vem primeiro porque estabiliza o ciclo de contexto confirmado. Features seguintes precisam de atualização segura de dados de `.specharbor/project-brief.md` sem sobrescrever intenção do usuário.
- O índice de contexto de repositório vem em segundo porque retrieval e recursos remoto/RAG precisam de metadados de inventário estáveis antes de selecionar, ranquear ou enriquecer contexto.
- Retrieval local vem em terceiro porque pode se basear no índice/inventário e preservar comportamento local/offline.
- Contexto remoto do GitHub vem em quarto porque deve reutilizar as mesmas abstrações de contexto sem se tornar obrigatório para uso local.
- Provedor de RAG vem por último porque depende de registros contextuais estáveis e limites de retrieval bem definidos, e precisa manter fallback local/offline intacto.

## Feature Brief: implement-project-brief-merge-and-update

ID da mudança: `implement-project-brief-merge-and-update`

Objetivo: adicionar um modo seguro, de confirmação-first, para mesclar/atualizar `.specharbor/project-brief.md` sem sobrescrever contexto pertencente ao usuário nem tratar sinais detectados como fatos confirmados.

Por que vem nessa ordem: o `brief` atual recusa, por design, merge, atualização, sobrescrita e append. Isso estava correto para a fundação, mas o ciclo de contexto restante precisa de um caminho controlado de atualização antes de indexação, retrieval, remoto ou RAG dependerem de contexto confirmado.

Escopo principal:

- definir comportamento explícito de atualização para briefs existentes;
- preservar seções conhecidas do brief e valores já confirmados;
- oferecer fatos detectados e suposições apenas como sugestões que exigem confirmação;
- adicionar tratamento de conflitos entre valores confirmados existentes e evidências atuais do repositório;
- manter escrita determinística, auditável e segura;
- atualizar documentação somente quando o comportamento existir.

Fronteiras explícitas de out-of-scope:

- sem indexação de repositório inteiro;
- sem retrieval, ranking de trechos, embeddings, vector store ou RAG;
- sem descoberta GitHub ou remota;
- sem APIs de provedor, APIs de modelo local ou execução de agente;
- sem verificação automática de comando;
- sem mudanças em release, npm, Homebrew, `install.sh`, publicação, tag, merge ou automação de controle de fonte.

Fronteiras arquiteturais esperadas:

- o domínio deve possuir modelos de campos do brief, categorias confirmed/detected/assumption, decisões de merge e registros de conflito;
- ports expõem apenas operações de sistema de arquivos consumidas pelo use case de atualização;
- use cases orquestram parsing, renderização da proposta de atualização, input de confirmação e política de escrita;
- CLI só controla prompts e formatação para usuário;
- comportamento concreto de filesystem permanece em adapters.

Foco de testes:

- parsing e preservação do brief existente;
- renderização da proposta de atualização;
- confirmação e cancelamento;
- resolução de conflitos;
- suposições nunca virando fatos;
- ausência de escrita parcial em falha;
- cobertura de regressão para `brief`, `context discover` e `prompt` existentes.

Impacto em documentação:

- atualizar `README.md`, `../usage.md` e `../workflow.md` apenas se a implementação expuser comando/flag de atualização.
- documentar que o comportamento de atualização é confirmation-first e não envolve indexação, RAG, descoberta remota ou verificação de comando.

Riscos principais:

- sobrescrever contexto mantido pelo usuário;
- transformar fatos/suposições detectadas em valores confirmados sem ação explícita;
- mascarar contexto confirmado obsoleto em vez de expor conflitos;
- deixar regras de merge no CLI em vez de no domínio.

Responsabilidade do subagente:

- propor orientação de spec para semântica de atualização, casos de conflito e fluxo de confirmação, critérios de aceite e testes;
- identificar como manter comportamento existente de `brief/discovery/prompt` estável;
- não escrever código.

Checklist de validação do coordenador:

- a spec possui exatamente cinco arquivos OpenSpec;
- o escopo está limitado a merge/update de brief;
- o comportamento de escrita condicional permanece disponível ou foi evoluído com nota explícita de compatibilidade;
- não aparecem indexação, retrieval, contexto remoto, embeddings, RAG ou arquivos de release no escopo;
- testes cobrem confirmação, conflitos, dados obsoletos e ausência de promoção silenciosa de suposições.

## Feature Brief: implement-repository-context-index

ID da mudança: `implement-repository-context-index`

Objetivo: adicionar um índice de contexto de repositório seguro que registre metadados de inventário delimitados para fontes de contexto suportadas sem executar retrieval, embeddings, RAG ou descoberta remota.

Por que vem nessa ordem: com contexto confirmado já atualizável com segurança, um índice fornece metadados de inventário local estáveis para retrieval futura. O índice ainda não seleciona nem ranqueia trechos.

Escopo principal:

- definir modelo de índice de contexto e comportamento de persistência/atualização;
- inventariar fontes locais suportadas, categorias, metadados e informações de frescor;
- reutilizar regras já existentes de skip para arquivos sensíveis e diretórios pesados/gerados;
- manter conteúdo de índice delimitado e determinístico;
- validar ou reportar estado para ajudar o usuário a entender o que foi indexado.

Fronteiras explícitas de out-of-scope:

- sem retrieval local nem ranking de snippets;
- sem embeddings, vector store ou RAG;
- sem coleta de fontes GitHub/remo­ta;
- sem alteração de prompt exceto leitura de metadados estáveis;
- sem análise de dependências de projeto;
- sem mudanças de release ou publicação.

Fronteiras arquiteturais esperadas:

- domínio controla registros de índice, categorias, marcadores de frescor, referências de política de skip e ordenação determinística;
- ports expõem operações de inventário de filesystem e persistência/validação de índice;
- use cases montam, leem, validam e reportam estado de índice;
- adapters implementam travessia de filesystem segura e persistência.

CLI apenas parseia comandos e formata relatórios sem possuir regras de índice.

Foco de testes:

- geração determinística do índice;
- tratamento seguro de caminhos e symlinks;
- políticas de pasta sensível/pesada;
- detecção de frescor e índices obsoletos;
- comportamento com repositório vazio ou ambíguo;
- ausência de retrieval ou embedding.

Impacto em documentação:

- documentar formato do comando e localização do arquivo de índice após definir a implementação;
- explicar que o índice é metadado/inventário, não banco semântico, vector store ou sistema RAG.

Riscos principais:

- indexar conteúdo excessivo;
- armazenar segredos brutos ou arquivos grandes integrais;
- criar formato de persistência que retrieval não consegue reutilizar.

- usuários interpretarem metadados de inventário como contexto confirmado.

Responsabilidade do subagente:

- propor opções de schema de índice, restrições de segurança, critérios de aceite e riscos de migração/obsolescência;
- identificar sobreposição com fontes de `context discover`;
- não escrever código.

Checklist de validação do coordenador:

- a spec diz explicitamente que indexação não implementa retrieval;
- o modelo de índice preserva evidência de fonte e distinção de classificação;
- políticas de skip herdadas ou claramente estendidas;
- caminho e formato de persistência justificadas;
- testes provam determinismo e limites.

## Feature Brief: implement-local-context-retrieval

ID da mudança: `implement-local-context-retrieval`

Objetivo: adicionar retrieval local/offline determinístico sobre o índice/inventário delimitado para que SpecHarbor selecione contexto local relevante sem embeddings, RAG, APIs de provedor ou serviços remotos.

Por que vem nessa ordem: retrieval precisa de inventário seguro antes. Ele deve validar a abstração local antes de incluir providers remotos.

Escopo principal:

- definir queries e modelos de resultado locais;
- recuperar trechos limitados ou registros estruturados de fontes locais suportadas;
- usar ranking determinístico lexical/metadados/regra;
- preservar evidência de fonte, classificação e confiança;
- manter retrieval seguro para uso futuro em prompts/spec sem dump bruto de arquivos.

Fronteiras explícitas de out-of-scope:

- sem embeddings;
- sem vector databases;
- sem provedor RAG;
- sem contexto remoto do GitHub;
- sem atualização obrigatória de project brief;
- sem execução de comandos do projeto;
- sem mudanças de release ou publicação.

Fronteiras arquiteturais esperadas:

- domínio controla modelos de query, result, rank, snippet e evidência;
- ports expõem acesso a fontes indexadas locais e leitura delimitada de arquivos;
- use cases orquestram recuperação e limitação de resultados;
- adapters fazem leitura segura.

CLI ou integração com prompt apenas formata resultados retornados e não implementa regras de ranking.

Foco de testes:

- ranking determinístico e desempate;
- limites e truncamento;
- caminhos relativos válidos;
- enforcement da política de skip;
- ausência de promoção de suposição;
- comportamento com índice faltante/obsoleto;
- regressão de discovery e prompts.

Impacto em documentação:

- documentar comportamento de retrieval, limites e limites locais/offline;
- deixar claro que retrieval não é RAG e não chama provedores.

Riscos principais:

- ranking de retrieval virando regra de negócio escondida no CLI;
- exposição de conteúdo bruto;
- confundir trecho recuperado com contexto confirmado;
- abstrações RAG prematuras vazando para retrieval local.

Responsabilidade do subagente:

- propor use cases e regras de ranking, limites e casos de teste;
- identificar campos de índice necessários;
- não escrever código.

Checklist de validação do coordenador:

- a spec diz explicitamente que retrieval não implementa embeddings ou RAG;
- o modelo preserva separação entre confirmed, detected e assumption;
- limites de resultado e evidência de fonte especificados;
- testes cobrem ranking determinístico e truncamento seguro;
- não exige dependência remota de provider para funcionamento básico.

## Feature Brief: implement-github-remote-context

ID da mudança: `implement-github-remote-context`

Objetivo: adicionar contexto remoto GitHub opcional por meio de abstrações de controle de repositório, preservando uso local/offline quando GitHub estiver indisponível ou não configurado.

Por que vem nessa ordem: contexto remoto deve reutilizar as mesmas abstrações de contexto, indexação e retrieval após essas abstrações ficarem estáveis. Ele não deve redefini-las.

Escopo principal:

- definir fontes remotas opcionais (metadados do repositório, arquivos da branch padrão, issues, pull requests, discussões e metadados de workflow apenas quando aprovado explicitamente na spec);
- adicionar port no domínio para leitura de contexto remoto;
- implementar adapter GitHub para essa port;
- manter credenciais opcionais e explícitas;
- mapear contexto remoto nas classificações/fonte já existentes;
- manter fallback local/offline.

Fronteiras explícitas de out-of-scope:

- sem requisito de credenciais GitHub para fluxos locais;
- sem automação de controle de fonte (commit, push, PR, merge, tag, release);
- sem chamadas remotas obrigatórias de comandos locais;
- sem GitLab, Bitbucket ou forge genérica salvo especificação própria;
- sem provedor RAG ou embeddings;
- sem mudanças de release ou publicação.

Fronteiras arquiteturais esperadas:

- domínio mantém registros e mapeamento de classificação de contexto remoto;
- ports definem pequenas interfaces de leitura remota para use cases;
- use cases orquestram coleta opcional e fallback local;
- detalhes GitHub (auth, paginação, rate limit, parsing) ficam em adapters;
- CLI/config não deve vazar tipos de SDK GitHub para o core.

Foco de testes:

- testes de port com stubs/fakes para comportamento de use case;
- testes de adapter com respostas HTTP mockadas ou fixtures;
- comportamento sem credencial;
- rate limit e erros de API;
- fallback local quando remoto indisponível;
- ausência de efeitos colaterais de controle de fonte.

Impacto em documentação:

- documentar contexto GitHub como opcional;
- documentar credenciais, falhas, rate limit e fallback;
- deixar claro que não há automação de PR/merge/release/workflow.

Riscos principais:

- tornar fluxos locais dependentes de rede/credencial;
- coletar dados remotos demais;
- vazar conteúdo remoto privado sem limites;
- confundir declarações remotas de issues/PR com fatos confirmados.

Responsabilidade do subagente:

- propor limites de fontes remotas, credenciais, fallback e testes;
- identificar abstrações locais a serem reutilizadas;
- não escrever código.

Checklist de validação do coordenador:

- a spec afirma que GitHub é opcional e fallback local permanece;
- dados remotos são classificados e rastreáveis, não confirmados por padrão;
- não incluir automação de controle de fonte;
- responsabilidades de adapter mantêm GitHub fora do domínio.

## Feature Brief: implement-context-rag-provider

ID da mudança: `implement-context-rag-provider`

Objetivo: adicionar suporte opcional a provedor de RAG após retrieval local e contexto remoto estarem estáveis, preservando fallback determinístico local/offline e fluxos de agentes sem exigir API keys de provedor.

Por que vem nessa ordem: RAG é o recurso mais dependente de infraestrutura. Deve depender de registros de contexto estáveis e limites de retrieval já definidos.

Escopo principal:

- definir limites de provider e augmentação de retrieval;
- suportar embeddings ou providers de busca apenas por ports pequenos no domínio;
- manter retrieval local disponível quando configuração de provider estiver ausente;
- preservar classificação, evidência de fonte e suposições;
- limitar tamanho de contexto em prompt/espec;
- adicionar tratamento de erros e fallback do provider.

Fronteiras explícitas de out-of-scope:

- não remover retrieval/local/offline;
- sem API keys de provedor obrigatórias para fluxos agent-assisted;
- sem chamadas silenciosas de provider de comandos não relacionados;
- sem automação de controle de fonte;
- sem mudanças de release/publish;
- sem tratar sumários RAG como contexto confirmado.

Fronteiras arquiteturais esperadas:

- domínio controla resultados de retrieval aumentada, evidência de fonte, confiança e status de fallback;
- ports definem pequenas interfaces para embedding, busca vetorial e augmentação quando necessário;
- use cases escolhem entre retrieval local e augmentação opcional por provider;
- autenticação, payload, erros, limites e mapping de resposta do provider ficam nos adapters;
- CLI/config apenas conectam configuração explícita de provider.

Foco de testes:

- testes com providers fake para sucesso, timeout, auth error e rate-limit;
- fallback local quando provider indisponível ou desativado;
- nenhuma chamada sem configuração explícita;
- preservação de classificação;
- limite de tamanho de prompt/contexto;
- tratamento de segredos e segurança de configuração.

Impacto em documentação:

- documentar RAG como augmentação opcional;
- documentar fallback local e comportamento sem provider;
- documentar que saída de RAG não é contexto confirmado;
- evitar prometer suporte além do implementado.

Riscos principais:

- tornar RAG obrigatório;
- exigir chaves de provider para agentes;
- permitir sobrescrever contexto confirmado com saída de provider;
- vazar segredos ou conteúdo privado;
- introduzir abstrações genéricas amplas antes de necessidade.

Responsabilidade do subagente:

- propor fronteiras de provider, fallback e regras de erro;
- confirmar como retrieval local continua base;
- não escrever código.

Checklist de validação do coordenador:

- a spec confirma que fallback local/offline é obrigatório;
- chamadas de provider exigem configuração explícita;
- workflows agent-assisted não exigem API key;
- saída de RAG é classificada e com evidência, não confirmada;
- testes cobrem falha de provider e fallback.

## Regras de conflito e dependência

Para evitar sobreposição:

- `implement-project-brief-merge-and-update` não implementa indexação.
- `implement-repository-context-index` não implementa retrieval.
- `implement-local-context-retrieval` não implementa embeddings nem RAG.
- `implement-github-remote-context` não deve se tornar obrigatório para uso local.
- `implement-context-rag-provider` não pode remover fallback local/offline.
- Nenhuma feature deve alterar release, npm, Homebrew, `install.sh`, GoReleaser, publicação, tag, merge ou automação de controle de fonte sem spec de release separado.
- Nenhuma feature deve promover suposições silenciosamente em fatos.
- Fatos detectados não são contexto confirmado sem confirmação explícita em fluxo.
- Contexto remoto não é contexto confirmado por padrão.
- Resumos de RAG não são contexto confirmado por padrão.
- O contexto do prompt deve permanecer delimitado e não pode despejar arquivos brutos.
- Arquivos compartilhados precisam ser sequenciados pelo coordenador quando múltiplas features dependam deles.

## Workflow OpenSpec/SDD

Cada feature real deve seguir este fluxo:

```text
Spec Author -> Architecture Reviewer -> Implementer -> Tester -> Change Reviewer -> PR -> merge -> Archive Housekeeping
```

Cada implementação real deve ainda criar exatamente cinco arquivos OpenSpec:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Cada implementação requer:

- um worktree dedicado;
- a branch esperada;
- árvore inicial limpa;
- verificação de branch/status antes do início;
- staging explícito apenas dos arquivos pretendidos;
- sem `git add -A`;
- testes focados e validação antes do PR;
- PR para a branch de implementação;
- merge antes do archive;
- branch de archive separada após merge;
- validação de arquivo após mover a mudança para `openspec/archive/<date>/<change-id>/`.

Implementação exige que os agentes leiam `AGENTS.md`, `.specharbor/rules/global.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md` e a mudança ativa antes de editar produção.

## Esqueletos de prompt de subagentes

### Subagente de planejamento de merge/atualização de brief

```text
You are the Brief Merge/Update Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Read the completed brief, discovery, and context-aware prompt archives. Plan the OpenSpec scope for implement-project-brief-merge-and-update. Define goals, out-of-scope boundaries, architecture responsibilities, risks, acceptance criteria, and tests. Preserve confirmation-first behavior and never promote detected facts or assumptions into confirmed context without explicit user confirmation.
```

### Subagente de planejamento de index de repositório

```text
You are the Repository Index Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-repository-context-index. Focus on bounded inventory metadata, deterministic index behavior, safe paths, skip rules, persistence choices, stale index handling, and tests. Do not include retrieval, embeddings, RAG, GitHub remote context, provider calls, or release changes.
```

### Subagente de planejamento de retrieval local

```text
You are the Local Retrieval Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-local-context-retrieval. Define local/offline retrieval models, ranking constraints, result limits, source evidence, and tests. Build on the repository index or bounded inventory. Do not include embeddings, vector databases, RAG providers, GitHub remote context, or provider APIs.
```

### Subagente de planejamento de contexto remoto GitHub

```text
You are the GitHub Remote Context Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-github-remote-context. Define optional remote context sources, source-control context ports, GitHub adapter responsibilities, credential handling, API failure behavior, rate-limit behavior, classification mapping, and local/offline fallback. Do not include source-control automation, PR creation, merge, push, release, embeddings, or RAG.
```

### Subagente de planejamento de provedor RAG

```text
You are the RAG Provider Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-context-rag-provider. Define optional provider boundaries, fallback to local retrieval, classification preservation, provider configuration, error handling, and tests. Do not require provider API keys for agent-assisted workflows. Do not remove local/offline behavior or treat RAG output as confirmed context.
```

### Subagente de revisão do coordenador principal

```text
You are the Main Coordinator Review Agent for SpecHarbor.

Review subagent planning/spec guidance only. Do not write production code.

Validate sequence, scope boundaries, architecture boundaries, OpenSpec file completeness, conflict rules, testing focus, documentation impact, branch/worktree requirements, and archive sequencing. Reject guidance that overlaps features, silently promotes assumptions into facts, makes remote/RAG mandatory, or touches release/publishing/source-control automation without a separate spec.
```

## Portões de decisão

Não comece implementação até todos os portões aprovados:

- Este documento de planejamento foi revisado.
- A ordem de features foi aceita ou alterada com justificativa.
- A primeira OpenSpec change foi selecionada.
- Não existem mudanças ativas sobrepostas sem planejamento.
- As fronteiras do escopo foram aprovadas.
- A branch principal está limpa antes de criar branch/worktree.
- O worktree de implementação está na branch esperada.
- O worktree de implementação inicia limpo.
- O estado de archive está limpo e não há mudança ativa obsoleta para a feature.
- Os cinco arquivos OpenSpec da feature estão completos antes da implementação.
- A revisão arquitetural não tem bloqueios.
- Testes e validações acordados estão definidos antes da implementação.

## Próxima implementação recomendada

A próxima feature concreta deve ser:

```text
implement-project-brief-merge-and-update
```

Ela é indicada porque o projeto já tem criação de contexto confirmado, descoberta local e prompts contextualizados, mas ainda não tem um caminho seguro para reconciliar um brief existente com novas evidências ou evidências alteradas do repositório. Estabilizar esse ciclo de contexto confirmado primeiro impede que index, retrieval, GitHub e RAG operem sobre briefs desatualizados ou difíceis de atualizar.
