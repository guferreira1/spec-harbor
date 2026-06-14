# Uso

Para executar exemplos da raiz do repositório durante o desenvolvimento do SpecHarbor:

```bash
go run ./cmd/specharbor help
```

Também é possível construir o binário local:

```bash
go build ./cmd/specharbor
```

Isso cria um binário local `specharbor` na raiz do repositório.

## Comandos implementados

### Verificar metadados de versão

```bash
go run ./cmd/specharbor version
specharbor version
```

`specharbor version` imprime metadados de build determinísticos em múltiplas linhas:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Campos:

- `version`: metadado da versão do produto.
- `commit`: commit da origem fornecido pelo build.
- `date`: data de build fornecida pelo build.
- `dirty`: estado da árvore de trabalho fornecido pelo build.

`dev` significa que nenhuma versão de release foi injetada. `unknown` significa que o build não forneceu esse campo de metadado. Tags de release do Git usam `vX.Y.Z`, por exemplo `v0.2.0`, enquanto o metadado de binário de release usa `X.Y.Z`, por exemplo `0.2.0`.

`go install` puro sem `-ldflags` usa o mesmo metadado de fallback de desenvolvimento. Um binário instalado dessa forma deve imprimir:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Esse é o comportamento esperado. Para obter metadados de release, o binário precisa ser compilado com valores `-ldflags` injetados.

Builds de release injetam metadados por meio de variáveis Go `-ldflags -X` em `github.com/guferreira1/spec-harbor/internal/platform/version`:

```bash
go build \
  -ldflags "
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Version=0.2.0
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Commit=abc1234
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Date=2026-06-10T19:00:00Z
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Dirty=false
  " \
  ./cmd/specharbor
```

A execução exibe a string de versão injetada como recebida e não a normaliza. Não inspeciona tags Git, não lê `.git`, não executa Git e não normaliza versões. O GoReleaser injeta metadados de release ao construir assets do GitHub Release a partir de tags como `v0.2.0`, e esses binários exibem metadados simples como `0.2.0`. As opções de instalação estão documentadas em [Instalação](install.md): `install.sh`, o pacote wrapper npm e o tap do Homebrew estão disponíveis, enquanto pacotes Linux nativos, gerenciadores Windows, assinatura, SBOMs e imagens Docker seguem como próximos passos.

### Inicializar um projeto

```bash
go run ./cmd/specharbor init
```

`init` cria os arquivos de projeto OpenSpec e SpecHarbor que faltam no diretório atual. Arquivos existentes são ignorados.

### Escanear um projeto

```bash
go run ./cmd/specharbor scan
```

`scan` executa uma varredura informativa de projeto local. É agnóstico de stack: projetos de usuário não precisam ser projetos Go.

O relatório pode incluir ecossistemas detectados, gerenciadores de pacotes, dicas de comandos de teste, sinais de CI, sinais de container/deploy, sinais SpecHarbor/OpenSpec e observações. O comando não exige flags ou argumentos.

### Descobrir contexto do projeto

```bash
go run ./cmd/specharbor context discover
specharbor context discover
```

`context discover` executa descoberta de contexto de projeto local/offline determinística. Lê apenas um conjunto limitado de fontes de repositório suportadas e imprime um relatório estruturado na seguinte ordem:

```text
Detected project context:

User-confirmed context:
- none detected

Detected facts:
- Stack: Go
  Source: go.mod
  Classification: detected_fact
  Confidence: high

Suggested assumptions:
- Test command: go test ./...
  Source: go.mod (go.mod convention)
  Classification: suggested_assumption
  Confidence: medium

Notes:
- none detected
```

Classificações de sinal:

- `user_confirmed_context`: analisado a partir de seções conhecidas em `.specharbor/project-brief.md`.
- `detected_fact`: existe evidência explícita em uma fonte de repositório suportada.
- `suggested_assumption`: inferência convencional ou incompleta; nunca é tratada como fato.

Níveis de confiança são `high`, `medium` e `low`. A confiança não muda a classificação: um fato com alta confiança ainda não é contexto confirmado pelo usuário, e uma suposição ainda é uma suposição.

Fontes de descoberta suportadas incluem:

- `AGENTS.md`, `.specharbor/rules/` e `.specharbor/project-brief.md`;
- `README.md`, `CONTRIBUTING.md` e arquivos Markdown sob limite em `docs/`;
- `openspec/project.md` e arquivos Markdown sob limite em `openspec/specs/`;
- `package.json`, `go.mod`, `pom.xml`, `build.gradle`, `build.gradle.kts`, `Cargo.toml`, `pyproject.toml` e `requirements.txt`;
- `Dockerfile`, `docker-compose.yml`, `docker-compose.yaml`, `Makefile`, `Taskfile.yml`, `Taskfile.yaml` e `.github/workflows/`;
- validação de layout de repositório com limite para pontos de entrada de CLI em `cmd/`.

O comando pode reportar tipo de projeto, resumo de propósito, stack, linguagens, frameworks, dicas de arquitetura, gerenciadores de pacotes, comandos de teste/build/run, fontes de documentação, fontes de instrução de agente, fontes OpenSpec, sinais de container e de workflow, e observações. Não despeja o conteúdo bruto de arquivos.

Limites de segurança:

- nenhum comando de projeto, gerenciador de pacotes, testes, builds, scripts, agentes, shells, APIs de provedor, APIs de modelo local, APIs de rede, APIs de controle de fonte ou ferramentas de workflow são executados;
- não há indexação de repositório, embeddings, bancos de dados vetoriais, RAG, recuperação, ranqueamento de snippet, descoberta remota, injeção de prompt, comportamento de merge/update ou sobrescrita de brief;
- arquivos sensíveis como `.env`, `.env.*`, `*.pem`, `*.key`, `id_rsa`, `id_ed25519`, `secrets.*` e `credentials.*` são ignorados;
- pastas pesadas/geradas como `.git/`, `node_modules/`, `dist/`, `build/`, `target/`, `vendor/`, `coverage/`, `.tmp/`, `.cache/`, `.next/`, `.nuxt/`, `out/`, `bin/` e `obj/` são ignoradas;
- links simbólicos não são atravessados.

`specharbor context` sem `discover`, `index`, `retrieve`, `github` ou `rag`, subcomandos não suportados como `context update`, argumentos posicionais após `discover` e flags como `--json`, `--path`, `--deep`, `--github` ou `--rag` são rejeitados.

Quando `specharbor brief` é executado, a descoberta pode fornecer apenas sugestões de menu e registros de contexto detectado. O usuário ainda deve selecionar ou informar todas as respostas, a confirmação ainda é obrigatória antes da escrita e um arquivo `.specharbor/project-brief.md` existente ainda é recusado em vez de ser mesclado, atualizado, sobrescrito ou anexado.

### Indexar contexto do repositório

```bash
go run ./cmd/specharbor context index
go run ./cmd/specharbor context index --write
go run ./cmd/specharbor context index --check
specharbor context index
specharbor context index --write
specharbor context index --check
```

`context index` cria um índice de contexto de repositório local/offline determinístico para fontes de contexto suportadas. O índice é metadado de inventário limitado para trabalho futuro de recuperação local. Ele não é retrieval, ranqueamento de snippets, RAG, banco semântico, vector store, nem contexto de projeto confirmado.

Modos do comando:

- sem flag: gera o índice atual em memória e imprime um relatório conciso sem gravar arquivos;
- `--write`: gera o índice atual e grava com segurança `.specharbor/context-index.json`;
- `--check`: lê `.specharbor/context-index.json`, reconstrói os metadados atuais e informa se o índice armazenado está atual, obsoleto, ausente ou inválido.

`--write` e `--check` são mutuamente exclusivos. Flags não suportadas, argumentos posicionais não suportados, `--json`, `--path`, `--deep`, `--github` e `--rag` são rejeitados.

O caminho do índice gerado é:

```text
.specharbor/context-index.json
```

O arquivo é estado local gerado e é ignorado pelo controle de fonte. Não é para ser commitado. O arquivo de índice gerado não é incluído como entrada de índice.

O JSON armazenado usa a schema version `1` e metadados de geração determinísticos. Ele registra:

- limites selecionados;
- um marcador seguro da raiz do projeto, como `openspec/project.md`, quando presente;
- entradas com caminho relativo, categoria de fonte, tipo de arquivo, pista de linguagem/ecossistema, tamanho de arquivo, hash de conteúdo SHA-256, metadados de horário de modificação, flag de suporte a recuperação, pistas de classificação e categoria de evidência da fonte;
- registros de skip com caminho relativo e código de motivo apenas;
- estado de truncamento.

O índice não armazena conteúdo bruto de arquivos, snippets, segredos, saída de comandos, saída de provedor, embeddings, vetores, dados de API remota, caminhos locais absolutos ou contexto de projeto confirmado.

Fontes indexadas suportadas incluem:

- `AGENTS.md`, `.specharbor/rules/` e `.specharbor/project-brief.md`;
- `README.md`, `CONTRIBUTING.md` e arquivos Markdown sob limite em `docs/`;
- `openspec/project.md` e arquivos Markdown sob limite em `openspec/specs/`;
- `package.json`, `go.mod`, `pom.xml`, `build.gradle`, `build.gradle.kts`, `Cargo.toml`, `pyproject.toml` e `requirements.txt`;
- `Dockerfile`, `docker-compose.yml`, `docker-compose.yaml`, `Makefile`, `Taskfile.yml`, `Taskfile.yaml` e `.github/workflows/`.

Limites de segurança:

- nenhum comando de projeto, gerenciador de pacotes, testes, builds, scripts, shells, agentes, prompts, APIs de provedor, APIs de modelo local, APIs de rede, APIs de controle de fonte ou ferramentas de workflow são executados;
- não há recuperação de snippets por ranking, embeddings, bancos de dados vetoriais, RAG, contexto remoto, integração com provedor, execução de prompt, execução de agente, verificação de comando, automação de controle de fonte, automação de release, mudanças de npm, mudanças de Homebrew, mudanças de `install.sh`, mudanças de GoReleaser ou comportamento de publicação;
- arquivos sensíveis como `.env`, `.env.*`, `*.pem`, `*.key`, `id_rsa`, `id_ed25519`, `secrets.*` e `credentials.*` são ignorados;
- pastas pesadas/geradas como `.git/`, `node_modules/`, `dist/`, `build/`, `target/`, `vendor/`, `coverage/`, `.tmp/`, `.cache/`, `.next/`, `.nuxt/`, `out/`, `bin/` e `obj/` são ignoradas;
- links simbólicos não são atravessados;
- traversal de caminho, caminhos absolutos, caminhos de unidades Windows, caminhos com byte nulo e caminhos fora da raiz do projeto são rejeitados.

Os limites padrão são 500 arquivos indexados, 256 KiB por arquivo indexado, 5 MiB no total de bytes indexados, 200 registros de skip persistidos e profundidade limitada para diretórios de contexto suportados. Quando os limites são atingidos, o relatório marca o índice como truncado e registra motivos de skip estáveis sem despejar conteúdos.

### Recuperar contexto local

```bash
go run ./cmd/specharbor context retrieve --query "architecture"
specharbor context retrieve --query "architecture"
```

`context retrieve` executa recuperação local/offline determinística sobre fontes suportadas representadas por `.specharbor/context-index.json`. O comando exige um índice existente válido, atual e não truncado gerado por:

```bash
go run ./cmd/specharbor context index --write
```

Índices faltantes, inválidos, obsoletos, ilegíveis, com esquema não suportado ou truncados falham de forma segura e orientam o usuário a rodar `specharbor context index --write`. A recuperação local nunca cria, atualiza ou persiste o índice em silêncio, e não grava cache de recuperação nem arquivo de saída.

A consulta deve ser explícita via `--query`. Consultas vazias são rejeitadas, consultas com mais de 512 caracteres são rejeitadas e os termos normalizados são limitados. Consultas posicionais e flags não suportadas como `--json`, `--github`, `--remote`, `--rag`, `--embed`, `--provider`, `--execute`, `--agent` e `--deep` são rejeitadas.

A recuperação lê apenas fontes locais indexadas com marcação de recuperação suportadas, incluindo `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, arquivos Markdown em `docs/`, `openspec/project.md`, arquivos Markdown em `openspec/specs/`, arquivos Markdown em `.specharbor/rules/`, `.specharbor/project-brief.md`, manifests de pacote/build/dependência, arquivos Docker e compose, fontes Makefile/Taskfile e arquivos YAML de workflow em `.github/workflows/`.

Os limites padrão de retrieval são 128 KiB por arquivo de fonte, 1 MiB de leituras totais, 10 resultados totais, 2 snippets por arquivo, 600 caracteres por snippet, 2 linhas de contexto antes e depois do termo e 8.000 caracteres de snippet/resumo renderizados. Os resultados incluem ranking, caminho, categoria ou evidência, score, pistas de classificação quando presentes, intervalo de linhas quando aplicável e snippet/resumo limitado.

A retrieval usa pontuação lexical determinística com reforços por caminho, nome de arquivo, frase, categoria e pista de classificação. Não usa embeddings, vector databases, provedores semânticos, reranking de LLM, APIs de provedor, busca remota, saída de comandos ou geração de resposta via RAG.

Limites de segurança:

- nenhum comando de projeto, gerenciador de pacotes, teste, build, script, shell, agente, prompt, API de provedor, API de modelo local, API de rede, API de controle de fonte ou ferramenta de workflow é executado;
- não há embeddings, bancos de dados vetoriais, geração de resposta por RAG, contexto remoto, integração com provedor, execução de prompt, execução de agente, verificação de comando, automação de controle de fonte, automação de release, mudanças em npm, mudanças em Homebrew, mudanças em `install.sh`, mudanças em GoReleaser ou comportamento de publicação;
- arquivos sensíveis como `.env`, `.env.*`, `*.pem`, `*.key`, `id_rsa`, `id_ed25519`, `secrets.*` e `credentials.*` são ignorados;
- pastas pesadas/geradas como `.git/`, `node_modules/`, `dist/`, `build/`, `target/`, `vendor/`, `coverage/`, `.tmp/`, `.cache/`, `.next/`, `.nuxt/`, `out/`, `bin/` e `obj/` são ignoradas;
- links simbólicos não são atravessados;
- traversal de caminho, caminhos absolutos, caminhos de unidades Windows, caminhos com byte nulo e caminhos fora da raiz do projeto são rejeitados;
- não é imprimido o dump bruto completo de arquivo.

### Recuperar contexto remoto no GitHub

```bash
go run ./cmd/specharbor context github --repo owner/name --query "architecture"
specharbor context github --repo owner/name --query "architecture"
```

`context github` executa recuperação de contexto remota explícita, limitada e somente leitura no GitHub. É separado da recuperação local e não lê nem escreve `.specharbor/context-index.json`.

Formas suportadas:

```bash
specharbor context github --repo owner/name --query "architecture"
specharbor context github --repo https://github.com/owner/name --query "architecture"
specharbor context github --repo owner/name --ref main --query "architecture"
specharbor context github --repo owner/name --query "architecture" --path docs
specharbor context github --repo owner/name --query "architecture" --path docs/usage.md --path README.md
```

Entrada de repositório aceita `owner/name` e `https://github.com/<owner>/<repo>`, normalizada para `owner/name`. Hosts não suportados, URLs do GitHub Enterprise, credenciais, query strings, fragments, caminhos de sistema de arquivos, traversal, espaços e caracteres de controle são rejeitados.

`--query` é obrigatório. Consultas vazias são rejeitadas, consultas com mais de 512 caracteres são rejeitadas e os termos normalizados de consulta têm limites. Consultas posicionais e flags não suportadas como `--json`, `--rag`, `--embed`, `--provider`, `--execute`, `--agent` e `--deep` são rejeitadas.

`--ref` é opcional. Quando omitido, o SpecHarbor resolve a branch padrão do repositório pelo GitHub. A entrada de `ref` é limitada e rejeita traversal, bytes nulos, URLs, credenciais, query strings, fragments e formas suspeitas de barra.

`--path` é opcional e repetível. Filtros de caminho estreitam apenas o conjunto de fontes aprovado; eles não expandem recuperação para arquivos arbitrários. Filtros devem ser caminhos relativos ao repositório e rejeitam traversal, caminhos absolutos, unidades Windows, bytes nulos, query strings, fragments e expansão de wildcard.

Fontes remotas suportadas são limitadas a `README.md`, `AGENTS.md`, `CONTRIBUTING.md`, arquivos Markdown sob `docs/`, `openspec/project.md`, arquivos Markdown sob `openspec/specs/`, arquivos Markdown sob `.specharbor/rules/`, `.specharbor/project-brief.md`, manifestos/configurações suportados como `go.mod`, `package.json`, `Cargo.toml`, `pyproject.toml`, `Dockerfile`, `Makefile`, `Taskfile.yml`, arquivos compose e arquivos YAML de workflow em `.github/workflows/`.

Limites padrão: 50 arquivos buscados, 128 KiB por arquivo, 1 MiB de conteúdo total, 10 resultados totais, 2 snippets por arquivo, 600 caracteres por snippet, 8.000 caracteres de snippet/resumo renderizado, 500 entradas de árvore ou diretório escaneadas, profundidade de diretório limitada e timeout de HTTP de 10 segundos. Os resultados incluem repositório, detalhes de `ref` padrão/solicitada/resolvida quando disponíveis, SHA resolvido quando disponível, ranking, caminho, categoria ou evidência, score, intervalo de linhas quando aplicável, snippet ou resumo limitado e `Remote: yes`.

Autenticação é opcional. Acesso a repositório público funciona sem token quando o GitHub permite. Se `SPECHARBOR_GITHUB_TOKEN` estiver definido, o SpecHarbor envia como token Bearer do GitHub somente para este comando. O token não é impresso, persistido, incluído em relatórios ou mensagens de erro.

Erros seguros são retornados para entrada inválida, hosts não suportados, falhas de rede, timeouts, rate limits, respostas não autorizadas ou proibidas, respostas de not found, tokens inválidos, arquivos muito grandes, respostas muito grandes, conteúdo não suportado e muitos candidatos.

Limites de segurança:

- o acesso à rede é usado apenas quando `context github` é chamado explicitamente;
- solicitações ao GitHub são HTTPS somente leitura para `api.github.com`;
- não há APIs de escrita do GitHub, commits, branches, PRs, issues, comentários, labels, releases, tags, execução de workflow ou mutação de repositório;
- nenhum `git`, `gh`, comando de shell, gerenciador de pacotes, script, comando de projeto, prompt ou agente local é executado;
- por padrão, contexto remoto não é cacheado nem persistido;
- `.specharbor/context-index.json` não é gravado nem modificado;
- resultados remotos não são contexto confirmado pelo usuário e não são injetados automaticamente em prompts;
- arquivos sensíveis como `.env`, `.env.*`, `*.pem`, `*.key`, `id_rsa`, `id_ed25519`, `secrets.*`, `credentials.*`, `.npmrc`, `.pypirc` e `.netrc` são ignorados;
- pastas pesadas/geradas como `.git/`, `node_modules/`, `dist/`, `build/`, `target/`, `vendor/`, `coverage/`, `.tmp/`, `.cache/`, `.next/`, `.nuxt/`, `out/`, `bin/` e `obj/` são ignoradas;
- arquivos binários e não suportados são ignorados;
- o ranking é lexical determinístico com reforços por caminho, nome de arquivo, frase, categoria e cabeçalho;
- não são introduzidos embeddings, bancos vetoriais, geração de resposta RAG, APIs de provedor, reranking por LLM, execução de prompt, execução de agente ou automação de controle de fonte.

### Gerar resposta de contexto do provedor

```bash
SPECHARBOR_OPENAI_API_KEY=... go run ./cmd/specharbor context rag --query "architecture" --provider openai
SPECHARBOR_OPENAI_API_KEY=... specharbor context rag --query "architecture" --provider openai
```

`context rag` é o caminho explícito de resposta de contexto com provedor. Ele não é executado a partir de `context discover`, `context index`, `context retrieve`, `context github`, `brief`, `prompt`, `validate`, `review` ou `scan`.

O primeiro provedor suportado é `openai`. O comando lê `SPECHARBOR_OPENAI_API_KEY` apenas quando `context rag` é invocado. Se a variável estiver ausente ou vazia, SpecHarbor emite um relatório seguro `missing_credentials` e sai com código não zero. O token nunca é exibido, persistido, incluído no contexto do provedor ou nos detalhes de erro. `SPECHARBOR_OPENAI_MODEL` pode sobrescrever o modelo padrão OpenAI; quando não definido, SpecHarbor usa `gpt-5.4-mini`.

Formas suportadas:

```bash
specharbor context rag --query "architecture" --provider openai
specharbor context rag --query "architecture" --provider openai --from local
specharbor context rag --query "architecture" --provider openai --from github --repo owner/name
specharbor context rag --query "architecture" --provider openai --from local --from github --repo owner/name
specharbor context rag --query "architecture" --provider openai --from github --repo owner/name --ref main --path README.md
specharbor context rag --query "architecture" --provider openai --max-sources 4 --max-answer-chars 2000
```

Comportamento de fonte:

- sem `--from`, somente recuperação local é usada;
- `--from local` usa recuperação local determinística sobre o `.specharbor/context-index.json` atual;
- `--from github` só é usado quando passado explicitamente e exige `--repo owner/name`;
- `--path` e `--ref` se aplicam apenas à fonte GitHub explícita;
- snippets locais e remotos selecionados são limitados, têm fonte atribuída e são enviados ao provedor com caminho, tipo da fonte, marcador remoto e intervalo de linhas quando disponíveis.

A saída inclui resposta gerada, nome do provedor, modelo, status, quantidade de fontes, lista de fontes, marcador local/remoto, detalhes de repositório/ref para fontes GitHub, intervalo de linhas quando disponível e marcadores de truncamento quando aplicável. A saída do provedor é tratada como texto de resposta gerada, não como contexto de projeto confirmado.

Limites de segurança:

- nenhuma chamada a provedor acontece, exceto quando `context rag` for explicitamente invocado;
- nenhum token de provedor é lido por comandos locais/offline;
- não é impresso dump bruto de requisição/resposta do provedor;
- nenhum prompt de provedor, resposta de provedor, embeddings, vetores, cache ou saída de retrieval é persistido;
- `.specharbor/context-index.json` é lido para recuperação local, mas nunca gravado ou modificado;
- GitHub é somente leitura quando `--from github --repo owner/name` é explícito, e mutações no GitHub não são executadas;
- não é introduzida automação de controle de fonte, execução de shell, execução de prompt, execução de agente, execução de comando de projeto, automação de release, mudanças de npm, mudanças de Homebrew, mudanças em `install.sh`, mudanças em GoReleaser ou comportamento de publicação;
- respostas geradas não são injetadas automaticamente em prompts de papel, briefs de projeto, specs ou arquivos fonte.

### Briefing de projeto

```bash
go run ./cmd/specharbor brief
specharbor brief
```

`brief` inicia um fluxo interativo de briefing de projeto para casos em que o contexto do repositório está faltando, incompleto ou ambíguo. Sem flags, não aceita argumentos posicionais e só escreve quando `.specharbor/project-brief.md` está ausente. Flags não suportadas como `--force`, `--overwrite`, `--json`, `--from-scan`, `--github` ou `--rag` são rejeitadas.

O comando exige um TTY interativo. Em CI, entrada por pipe ou outros contextos sem TTY ele falha antes de pedir entrada ou escrever:

```text
brief requires an interactive TTY
```

O contexto solicitado no briefing inclui:

- tipo de projeto;
- propósito do projeto;
- usuários-alvo;
- stack;
- arquitetura;
- comando de instalação;
- comando de testes;
- comando de build;
- comando de execução;
- comportamento de agente preferido quando contexto está faltante.

Cada pergunta é de múltipla escolha com três a cinco opções. A última opção é sempre `Other / custom`; escolhê-la solicita uma resposta customizada não vazia. Escolhas de menu inválidas e respostas customizadas vazias são repetidas até três tentativas e, se falharem, termina sem escrever.

Antes de gravar, `brief` imprime o arquivo alvo e limites de segurança:

```text
SpecHarbor will create:

.specharbor/project-brief.md

Segurança:
- Stack, arquitetura e comandos vêm apenas de respostas confirmadas.
- Contexto detectado permanece separado das respostas do usuário.
- Suposições não são fatos confirmados.
- Nenhuma indexação de repositório, RAG, API de provedor, execução de agente, automação de controle de fonte, release ou publicação será executada.

Confirm? [y/N]:
```

A confirmação aceita `y` e `yes` normalizados, independente de maiúsculas/minúsculas. `n`, `no`, confirmação vazia e EOF cancelam com `operation cancelled` e não gravam nenhum arquivo. Respostas de confirmação não suportadas são repetidas até três tentativas; esgotar tentativas não grava nada.

Após confirmação, `brief` cria `.specharbor/` quando necessário e grava:

```text
.specharbor/project-brief.md
```

O arquivo é determinístico e em Markdown legível com estas seções:

```text
# Project Brief

## Project type
## Objetivo
## Target users
## Stack
## Architecture
## Commands
### Install
### Test
### Build
### Run
## Agent behavior
## Context sources
## Assumptions
```

Respostas fornecidas pelo usuário, contexto detectado e suposições são rotulados separadamente. Sugestões de descoberta e contexto detectado podem ser registrados separadamente nos briefs gerados, mas as respostas fornecidas permanecem separadas do contexto detectado e das suposições. O comando nunca converte silenciosamente stack, arquitetura, comando ou decisões de projeto faltantes ou ambíguas em fatos.

Se `.specharbor/project-brief.md` já existir, `brief` sem `--update` se recusa a sobrescrever ou anexar. A geração de prompt contextualizado por papel pode ler o brief existente pelo limite de descoberta local, mas não mescla, atualiza, sobrescreve ou anexa no brief.

Atualize explicitamente um brief de projeto existente com:

```bash
go run ./cmd/specharbor brief --update
specharbor brief --update
```

`brief --update` exige um TTY interativo e um `.specharbor/project-brief.md` existente. Ele lê o brief atual, reutiliza o caso de uso local `context discover` para fatos detectados e suposições sugeridas, e monta uma proposta de atualização. Não duplica a lógica de descoberta e não executa comandos.

O fluxo de atualização é confirmação primeiro:

- valores confirmados existentes pelo usuário são mantidos por padrão;
- fatos detectados são apenas evidência até aceitação explícita;
- suposições sugeridas permanecem suposições até aceitação explícita;
- conflitos preferem contexto confirmado existente por padrão;
- valores confirmados antigos e suposições antigas são expostos, mas não removidos automaticamente;
- o usuário pode manter um valor existente, inserir substituição customizada, aceitar um fato detectado, aceitar uma suposição sugerida, ignorar fatos detectados para um campo, manter suposições antigas, remover suposições antigas ou cancelar;
- uma prévia revisável é impressa antes da gravação;
- confirmação final é obrigatória antes de gravar;
- cancelamento ou EOF mantém o brief existente inalterado.

O Markdown atualizado permanece determinístico e mantém contexto confirmado pelo usuário, fatos detectados e suposições sugeridas separados. O comportamento seguro de gravação evita atualizações parciais. O comando não executa indexação de repositório inteiro, retrieval, ranqueamento de snippets, embeddings, bancos de dados vetoriais, RAG, contexto remoto GitHub, APIs de provedor, verificação de comando, execução de comando de projeto, execução de agente, execução de prompt, automação de controle de fonte, automação de release, mudanças de npm, mudanças de Homebrew, mudanças de `install.sh`, mudanças de GoReleaser ou fluxos de publicação.

### Mostrar o fluxo de trabalho recomendado

```bash
go run ./cmd/specharbor workflow
```

O formato de comando instalado é `specharbor workflow`. Ele imprime o fluxo de trabalho OpenSpec/SDD recomendado como texto de orientação apenas leitura:

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

A saída completa inclui cada `step id`, nome de exibição, propósito, modo, indicadores de suportado/aconselhável, IDs dos passos predecessores, sugestões de comandos e notas de segurança. As sugestões conectam o fluxo às rotinas existentes:

- `generate` cria ou inicia um pacote de mudança OpenSpec.
- `validate` valida os arquivos OpenSpec obrigatórios e a qualidade do conteúdo.
- `prompt --role ...` imprime prompts para `spec-author`, `architecture-reviewer`, `implementer`, `test-engineer` e `change-reviewer`.
- `review` valida o pacote de mudança local e a conclusão das tasks por checkbox.
- `archive` move explicitamente uma mudança aceita para o `archive`.

As sugestões de comando são somente de consulta e `specharbor workflow` não as executa. Commit, Pull Request e Merge permanecem etapas manuais; SpecHarbor não faz commit, não cria PRs e não faz merge. Este comando é somente leitura e não chama GitHub, GitLab, CI, APIs de provedor, CLIs de agente, automação de controle de fonte, execução de workflow, comandos externos ou automação remota.

### Gerar uma mudança

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

`generate <change-id> --blank` cria a estrutura esperada de mudança OpenSpec com conteúdo inicial em branco/manual.

A geração interativa é uma camada de prompt sobre os caminhos de geração determinísticos existentes:

```bash
go run ./cmd/specharbor generate <change-id> --interactive
```

`<change-id>` continua obrigatório na linha de comando; o modo interativo não solicita esse valor. `--interactive` não pode ser combinado com geração direta ou flags de entrada como `--blank`, `--template`, `--custom-template`, `--config-template`, `--guided`, `--hybrid`, `--ai-assisted`, `--agent-assisted`, `--from-file`, `--overwrite`, `--agent`, `--execute`, `--type`, `--title` ou `--summary`.

O modo interativo exige um terminal interativo. Em CI, entrada por pipe ou outros contextos sem TTY, falha imediatamente com:

```text
interactive mode requires a TTY
```

Nesse caso, não faz perguntas, não trava e não grava arquivos.

Os caminhos interativos suportados nesta versão são exatamente:

- `blank`
- built-in template
- custom template
- config template
- hybrid

O primeiro menu aceita as escolhas numeradas de `1` a `5` e palavras-chave estáveis como `blank`, `template`, `custom`, `config` e `hybrid`. A geração guiada direta permanece disponível apenas por flags não interativas. Geração assistida por IA, geração com agente, execução local de runner de agente, aplicação de saída de runner ao vivo e entrada de URL remota bruta não estão disponíveis nos prompts interativos.

Sequência de prompt:

- Blank solicita apenas o caminho de geração.
- Template built-in solicita o nome do template integrado (`feature`, `bugfix`, `docs` ou `refactor`).
- Template custom solicita nome do template custom, título opcional e resumo opcional.
- Template config solicita um alias de configuração, título opcional e resumo opcional.
- Hybrid solicita exatamente um namespace de fonte (template built-in, template custom ou config template), o valor da fonte, título obrigatório, resumo obrigatório e tipo opcional.

Respostas obrigatórias inválidas repetem até três tentativas. Respostas obrigatórias vazias são inválidas. Tipos de híbrido não vazios inválidos repetem até três tentativas. Título, resumo e tipo de híbrido opcionais vazios são tratados como omitidos. Ao esgotar tentativas, termina com código não-zero e não escreve nada.

Antes de executar qualquer caso de uso de geração, o modo interativo imprime um resumo determinístico:

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

Resumos de blank, template built-in, template custom e config template mostram `Validation: automatic no`. Resumos de hybrid mostram `Validation: automatic yes`, preservando a validação automática existente do hybrid após geração. O modo interativo não adiciona prompt de validação e nunca corrige automaticamente achados de validação.

A confirmação é normalizada e insensível a maiúsculas/minúsculas. `y` e `yes` confirmam em qualquer caixa, incluindo `Y`, `YES` e `Yes`. `n` e `no` cancelam em qualquer caixa, incluindo `N`, `NO` e `No`. Confirmação vazia e EOF também cancelam. Cancelamento sai com código não-zero e não grava nada. Respostas de confirmação não suportadas repetem até três tentativas; esgotar tentativas não grava nada.

Com confirmação, o modo interativo delega ao mesmo comportamento do comando direto equivalente:

```bash
go run ./cmd/specharbor generate add-blank --interactive
go run ./cmd/specharbor generate add-feature --interactive
go run ./cmd/specharbor generate add-payment-flow --interactive
go run ./cmd/specharbor generate add-configured-feature --interactive
go run ./cmd/specharbor generate add-login --interactive
```

Comportamento de gravação, preservação de arquivos existentes, renderização de templates, busca de aliases de configuração, proteções de template remoto e validação permanecem sob o modo de geração selecionado. Templates remotos são acessados apenas por aliases de configuração existentes após confirmação; o modo interativo não pede URLs ou checksums e não imprime credenciais, query strings, fragments, headers de autenticação, cookies, material OAuth ou segredos derivados do ambiente.

A geração com template built-in usa o mesmo comando com `--template <template-name>` e conteúdo inicial integrado determinístico:

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

Por exemplo:

```bash
go run ./cmd/specharbor generate add-example-feature --template feature
```

Os templates integrados suportados são exatamente:

- `feature`
- `bugfix`
- `docs`
- `refactor`

A geração com template customizado usa templates locais de projeto com `--custom-template <template-name>`:

```bash
go run ./cmd/specharbor generate <change-id> --custom-template <template-name>
```

Por exemplo:

```bash
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature
```

Um template customizado é um diretório simples sob a raiz local de projeto fixa `.specharbor/templates/<template-name>/` com os cinco arquivos obrigatórios de mudança OpenSpec:

```text
.specharbor/templates/<template-name>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Comportamento do template customizado:

Todos os cinco arquivos são obrigatórios e devem não ficar vazios; a ausência de diretório de template, de arquivos obrigatórios ou de um arquivo vazio falha com erro claro antes de qualquer escrita.
Arquivos extra desconhecidos e subdiretórios no diretório de template são ignorados e nunca copiados.
O conteúdo do template suporta substituição determinística mínima de variáveis: `{{change_id}}` é sempre substituído; `{{title}}` e `{{summary}}` são substituídos apenas quando as flags opcionais `--title` e `--summary` forem informadas. Tokens `{{...}}` desconhecidos ou não resolvidos permanecem na saída literalmente.
Não há condicionais, loops, `includes` ou qualquer outro recurso de linguagem de template; templates nunca são executados.
`--custom-template` é mutuamente exclusivo com `--blank`, `--template`, `--guided` e `--agent-assisted`.
Nomes de template customizado devem ser segmentos únicos de caminho seguros (caracteres `[A-Za-z0-9._-]`, sem `/` ou `\`, sem sequências `..`, sem início `.` ou `-`, no máximo 128 caracteres); nomes inválidos são rejeitados antes de qualquer acesso ao sistema de arquivos.
Templates integrados, customizados e orientados por configuração são disjuntos: `--template` resolve apenas os templates integrados, `--custom-template` resolve apenas `.specharbor/templates/`, e `--config-template` resolve apenas aliases em `.specharbor/config.yml`.
Os arquivos são gravados apenas em `openspec/changes/<change-id>/`; arquivos existentes são ignorados, nunca sobrescritos, e qualquer falha de validação de template gera escrita zero.
Templates customizados diretos são locais ao projeto: sem marketplace, sem caminhos locais arbitrários, sem execução de script de template, sem execução de shell, sem comportamento de rede/provedor e sem escrita de código de produção.

Após geração, execute `go run ./cmd/specharbor validate <change-id>` para validar a mudança gerada; os achados de validação dependem da qualidade do conteúdo do template, exatamente como nas mudanças escritas manualmente.

Exemplo de título/resumo de template customizado:

```bash
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature --title "Add payments" --summary "Adds a payment flow."
```

A geração orientada por configuração via template usa aliases declarados em `.specharbor/config.yml`:

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

Regras:

- `version: 1` é obrigatória quando `--config-template` é usado; config faltante, versão faltante, versões sem suporte, YAML inválido, entradas de alias inválidas e aliases faltantes falham com erro claro.
- Os tipos de fonte suportados são exatamente `builtin`, `custom` e `remote`.
- `builtin` resolve apenas templates integrados suportados: `feature`, `bugfix`, `docs` e `refactor`.
- `custom` resolve apenas `.specharbor/templates/<template-name>/` e usa a mesma validação de arquivos obrigatórios e substituição de `{{change_id}}`, `{{title}}` e `{{summary}}` do `--custom-template` direto.
- `remote` busca uma URL HTTPS ZIP explicitamente configurada, verifica o checksum `sha256:<64-hex>` antes de parsear o arquivo e grava os arquivos OpenSpec decodificados sem renderizar ou executar scripts de template.
- Aliases remotos exigem `url`, `checksum` e `format`; apenas `format: zip` é suportado. `template` é inválido para aliases remotos, e `url`, `checksum` e `format` são inválidos para aliases `builtin` e `custom`.
- URLs remotas devem usar HTTPS e incluir host e caminho. São rejeitados HTTP, file, SSH, git, git+ssh, FTP, alvos estilo SCP, `credentials/userinfo`, query strings, fragments, caracteres em branco/controle, URLs muito longas e redirects.
- Pacotes ZIP remotos devem conter exatamente cinco arquivos regulares não vazios no nível raiz: `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md` e `risks.md`. Caminhos aninhados, absolutos, traversal, unidades Windows, symlinks, entradas executáveis, arquivos duplicados, arquivos extras, arquivos faltantes, arquivos vazios, ZIP malformados, downloads muito grandes e conteúdo descompactado muito grande são rejeitados.
- Nomes de alias devem ser segmentos de caminho seguros: não vazios, no máximo 128 caracteres, caracteres `[A-Za-z0-9._-]`, sem `/` ou `\`, sem caminhos absolutos, sem traversal ou `..`, sem início `.` ou `-`.
- `--title` e `--summary` são opcionais com `--config-template`; são repassados para o caminho do template customizado resolvido e não mudam o resultado do template integrado.
- `--config-template` é mutuamente exclusivo com `--blank`, `--template`, `--custom-template`, `--guided`, `--agent-assisted`, `--ai-assisted` e `--execute`.
- `--template`, `--custom-template` e `--config-template` usam namespaces separados. Um template integrado, template customizado e alias de configuração podem ter o mesmo nome sem sombreamento, fallback ou adivinhação.
- Templates remotos não têm cache persistente nesta primeira versão e não suportam credenciais, OAuth, cabeçalhos de auth, cookies, expansão de token de ambiente, git clone, busca de marketplace, APIs de provedor, execução de script, execução de shell, escrita de código de produção, automação de controle de fonte, auto-commit, PR, merge ou automação de archive.
- Arquivos gerados permanecem limitados a `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md` e `risks.md` sob `openspec/changes/<change-id>/`; arquivos existentes são ignorados.

A geração híbrida combina uma fonte de template determinística com metadados obrigatórios de título e resumo:

```bash
go run ./cmd/specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
go run ./cmd/specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
go run ./cmd/specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
```

Seleção de fonte híbrida é explícita:

- `--template <name>` resolve apenas templates integrados.
- `--custom-template <name>` resolve apenas `.specharbor/templates/<name>/`.
- `--config-template <alias>` resolve apenas aliases de `.specharbor/config.yml`.
- É necessário exatamente um seletor de fonte. A ausência ou múltiplos seletores falham antes da escrita.
- Não há adivinhação de fonte, fallback ou `namespace shadowing`. O mesmo nome pode existir nos três namespaces porque a flag seleciona o namespace.

Regras de metadados do híbrido:

- `--title` é obrigatório e não pode ficar vazio após `trim`.
- `--summary` é obrigatório e não pode ficar vazio após `trim`.
- `--type` é opcional, mas quando informado deve ser exatamente `feature`, `bugfix`, `docs` ou `refactor`.
- Fontes integradas diretas derivam o tipo omitido do template selecionado. Por exemplo, `--hybrid --template feature` deriva `type=feature`.
- Aliases de configuração que resolvem para templates integrados derivam o tipo omitido do template integrado resolvido.
- O tipo informado deve corresponder ao template integrado direto ou resolvido. `--hybrid --template feature --type feature` tem sucesso; `--hybrid --template feature --type bugfix` falha claramente e não grava nada.
- Fontes customizadas, aliases custom de configuração e aliases remotos de configuração não inferem o tipo omitido. Se `--type` for omitido, `{{type}}` permanece não resolvido nesse conteúdo de template. Se `--type` for informado, `{{type}}` é substituído por esse valor.

A renderização `hybrid` substitui `{{change_id}}`, `{{title}}` e `{{summary}}`. Ele substitui `{{type}}` apenas quando existe um tipo efetivo informado ou derivado de template integrado. Tokens desconhecidos ou não resolvidos `{{...}}` permanecem literais. `hybrid` não adiciona condicionais, loops, funções, `includes`, hooks, comandos de shell, scripts ou comportamento executável de template.

Exemplos de híbrido:

```bash
go run ./cmd/specharbor generate add-login \
  --hybrid \
  --template feature \
  --title "Add login" \
  --summary "Add an OpenSpec change for login"
```

O exemplo integrado deriva `type=feature`.

```bash
go run ./cmd/specharbor generate add-login \
  --hybrid \
  --template feature \
  --type bugfix \
  --title "Add login" \
  --summary "Add an OpenSpec change for login"
```

O exemplo de incompatibilidade falha porque `bugfix` não combina com o template integrado `feature`.

```bash
go run ./cmd/specharbor generate add-payment-flow \
  --hybrid \
  --custom-template api-feature \
  --title "Add payments" \
  --summary "Adds a payment flow."
```

O exemplo customizado não infere tipo, então `{{type}}` permanece não resolvido a menos que `--type` seja informado.

Exemplo de template integrado por configuração:

```bash
go run ./cmd/specharbor generate add-login \
  --hybrid \
  --config-template default-feature \
  --title "Add login" \
  --summary "Add login support"
```

Exemplo de template customizado por configuração:

```bash
go run ./cmd/specharbor generate add-payment-flow \
  --hybrid \
  --config-template api-feature \
  --title "Add payments" \
  --summary "Adds a payment flow." \
  --type feature
```

Exemplo de configuração remota:

```bash
go run ./cmd/specharbor generate add-service \
  --hybrid \
  --config-template service-feature \
  --title "Add service" \
  --summary "Adds a service workflow."
```

Templates remotos ficam disponíveis para `hybrid` apenas por `--config-template <alias>`. O alias deve resolver para `source: remote` e mantém as mesmas salvaguardas remotas existentes: apenas HTTPS, sem credenciais, sem query strings, sem fragments, sem redirects, checksum obrigatório, checksum verificado antes do parse do ZIP, apenas ZIP, segurança estrita de arquivo, sem cache, sem execução de shell ou script, sem escrita de código de produção, e apenas os cinco arquivos OpenSpec são escritos.

Após gravações híbridas com sucesso ou rerun apenas de `skip`, o SpecHarbor executa a lógica de validação existente. O relatório de `hybrid` inclui status de validação, total de arquivos obrigatórios, quantidade de erros, quantidade de avisos e achados. Avisos de validação mantêm código de saída `0`; erros de validação são impressos após o relatório de geração e então o comando sai com código não-zero. A validação não corrige arquivos automaticamente.

`hybrid` rejeita `--blank`, `--guided`, `--ai-assisted`, `--agent-assisted`, `--from-file`, `--overwrite`, `--agent` e `--execute`. A sobreposição de IA e a aplicação de saída de runner ao vivo estão intencionalmente fora do escopo desta versão inicial. `hybrid` não chama APIs de provedor, APIs de LLM, APIs de modelo local, agentes, ferramentas de controle de fonte, ferramentas de workflow, comandos de shell ou scripts. Não grava código de produção e não executa auto-commit, auto-push, pull request, merge ou automação de archive.

A geração guiada usa flags explícitas de CLI:

```bash
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

A geração guiada é determinística, local e não interativa. Não solicita prompt durante a execução do comando; usa os valores informados de `--type`, `--title` e `--summary`.

Os tipos de geração guiada suportados são exatamente:

- `feature`
- `bugfix`
- `docs`
- `refactor`

A geração assistida por IA importa um arquivo local contendo OpenSpec Markdown criado por IA em formato de delimitador estrito:

```bash
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file>
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --overwrite
```

O arquivo de origem é somente texto local. Ele pode ser uma saída salva por um usuário de uma ferramenta de IA ou de agente, mas o SpecHarbor só lê o arquivo do disco. Não chama APIs de provedor, serviços de IA remotos, APIs de modelo local, OAuth, credenciais, agentes, comandos de shell, ferramentas de controle de fonte, ferramentas de workflow ou serviços de rede.

O arquivo deve conter exatamente estes cinco blocos de arquivo, com linhas de delimitador exatas:

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

São permitidos exatamente os nomes de arquivo `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md` e `risks.md`. Nomes desconhecidos, blocos duplicados, blocos ausentes, conteúdo vazio, caminhos absolutos, `..` traversal, caminhos aninhados, sintaxe de bloco malformada, formatos de wrapper com fence, formatos de patch/diff, marcadores de fim órfãos, blocos não fechados e texto não em branco fora dos blocos de arquivo são rejeitados antes de qualquer gravação.

A geração assistida por IA grava apenas estes arquivos em:

```text
openspec/changes/<change-id>/
```

Arquivos existentes são ignorados por padrão e reportados como ignorados. `--overwrite` é explícito e substitui apenas os arquivos obrigatórios existentes; alvos de saída symlink são rejeitados em vez de seguidos. Todo parsing e verificações prévias de destino ocorrem antes da escrita de arquivos; saída de IA malformada não grava nada. Após gravações ou skips bem sucedidos, o SpecHarbor executa a lógica existente de `validate <change-id>` e imprime o status de validação, quantidade de erros, quantidade de avisos e os achados. Avisos de validação mantêm código de saída `0`; erros de validação são exibidos e então o comando sai com código não-zero. A validação nunca corrige automaticamente os arquivos e a geração assistida por IA nunca modifica código de produção.

As fronteiras de segurança exibidas pelo comando fazem parte do contrato: chamadas de APIs de provedor `no`, serviços de IA remotos `no`, comandos de agente executados `no`, código de produção modificado `no`, comandos de controle de fonte executados `no`, e auto-commit, auto-push, PR, merge ou archive `no`. Aplicação direta de runner ao vivo, como `--agent-assisted --execute --apply`, não está implementada.

A autoria assistida por agente usa flags explícitas de CLI:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

Por exemplo:

```bash
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support"
```

Os tipos de autoria assistida por agente suportados são exatamente:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Dry-run permanece como padrão. Sem `--execute`, a autoria assistida por agente imprime no stdout um plano de autoria determinístico e imprime um prompt determinístico e copiável no stdout.

O prompt gerado tem o objetivo de ajudar um agente externo a escrever ou refinar apenas o pacote de mudança OpenSpec. A implementação continua sendo uma etapa posterior pelo fluxo normal do SpecHarbor.

Autoria assistida por agente em dry-run:

- não grava arquivos;
- não grava arquivo de prompt;
- não cria ou modifica arquivos OpenSpec;
- não cria ou modifica código de produção;
- não executa agentes ou comandos locais de agente;
- não exige runner;
- não resolve mapeamentos de comando executável;
- não chama APIs de provedor;
- não chama modelos locais;
- não chama APIs de rede;
- não chama APIs de controle de fonte;
- não chama ferramentas de workflow.

Os alvos de agente reconhecidos para dry-run são:

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

Agentes de dry-run desconhecidos são rejeitados como um endurecimento intencional de validação.

`--execute` é explícito e é suportado apenas com `--agent-assisted`:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

O modo execute envia o mesmo prompt de autoria OpenSpec determinístico via stdin para um comando local suportado. O diretório de trabalho do runner é a raiz do projeto atual.

Mapeamentos de comandos locais executáveis suportados são:

- `codex -> codex`
- `claude -> claude`
- `devin -> devin`
- `cursor -> cursor`
- `copilot -> copilot`
- `gemini -> gemini`
- `roo -> roo`
- `windsurf -> windsurf`
- `aider -> aider`

`generic` é reconhecido apenas para dry-run. `--execute --agent generic` falha porque a execução genérica exige, no futuro, um mapeamento de comando orientado por configuração.

O modo execute é apenas executar e reportar:

- comandos locais ausentes produzem erros de inicialização sem resultado de runner e sem código de saída;
- comandos iniciados com código de saída não-zero produzem relatório completo, então o SpecHarbor termina com código não-zero;
- stdout, stderr, código de saída e status de execução são capturados para processos iniciados;
- o SpecHarbor não faz parse nem aplica saída;
- stdout e stderr são exibidos apenas, sem parse ou aplicação;
- o SpecHarbor não grava arquivos OpenSpec a partir da saída do runner;
- o SpecHarbor não modifica código de produção com a saída do runner;
- o SpecHarbor não faz auto-commit, auto-push ou auto-merge.

APIs de provedor, automação de IDE, OAuth, credenciais, integrações de marketplace, execução remota, automação de controle de fonte e automação de workflow permanecem fora do escopo. O comportamento de comando do agente local é controlado pela ferramenta local instalada.

As gerações blank, template integrado, template customizado, config-template, híbrido, guiado e assistido por IA criam os mesmos arquivos obrigatórios de mudança OpenSpec:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

O conteúdo de template integrado é determinístico, local e de conteúdo inicial genérico. O conteúdo híbrido parte de exatamente uma fonte de template determinística e aplica substituição explícita de metadados antes da validação. O conteúdo guiado é determinístico e local com conteúdo inicial que inclui o título e resumo fornecidos. O conteúdo assistido por IA vem apenas da fonte local explícita de `--from-file` após parsing estrito.

O conteúdo gerado pode ser editado com segurança após a geração. A saída guiada não significa que o SpecHarbor inferiu requisitos específicos do projeto além do tipo, título e resumo fornecidos.

As gerações blank, template integrado, template customizado, config-template, híbrido e guiado ignoram arquivos existentes e não os sobrescrevem. A geração assistida por IA também ignora arquivos existentes por padrão e sobrescreve apenas com `--overwrite` explícito. Se o diretório da change existir parcialmente, executar geração novamente o recupera criando apenas os arquivos obrigatórios ausentes.

Exemplos copiáveis da raiz do repositório:

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

### Validar uma Mudança

```bash
go run ./cmd/specharbor validate add-example-feature
```

`validate <change-id>` executa validação determinística, local e somente leitura sobre o pacote de mudança em `openspec/changes/<change-id>/`. Ele nunca grava arquivos, nunca modifica a change e nunca chama rede, provedores de IA, agentes ou ferramentas de controle de fonte.

Verifica os cinco arquivos OpenSpec obrigatórios:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

Cada achado contém severidade, um código de regra estável em `snake_case`, uma mensagem (com números de linha quando relevante) e o caminho do arquivo.

Achados de erro (o pacote de mudança não é utilizável a jusante):

- `project_root_unavailable` - a estrutura de projeto OpenSpec está ausente.
- `change_directory_missing` - o ID da change está bem formado, mas é desconhecido.
- `required_file_missing` - um dos cinco arquivos obrigatórios está ausente.
- `file_empty` - um arquivo obrigatório está vazio ou contém apenas espaços. Outros achados de conteúdo desse arquivo são suprimidos.
- `file_missing_heading` - um arquivo obrigatório não tem cabeçalho markdown.
- `file_missing_body` - um arquivo obrigatório contém apenas cabeçalhos.
- `tasks_checkbox_missing` - `tasks.md` não possui uma tarefa de checkbox válida.
- `tasks_checkbox_malformed` - uma linha com aparência de checkbox (por exemplo, `- []`, `-[ ]`, `- [y]` ou `- [x]` sem texto) quebra a gramática `- [ ] text`; reportada com número de linha.
- `acceptance_criteria_item_missing` - `acceptance-criteria.md` não possui lista ou item de checkbox com texto significativo; itens somente placeholder (`N/A`, `...`, `?`, `TBD`, `TODO`, `FIXME`) não contam.

Achados de aviso (lacunas de qualidade que nunca fazem o comando falhar):

- `placeholder_content` - `TBD`/`TODO`/`FIXME` isolados, itens de lista que são apenas placeholder (`N/A`, `...`, `?`) ou `lorem ipsum`.
- `boilerplate_only_content` - o arquivo ainda contém apenas linhas de orientação de início conhecidas e nunca foi editado de forma significativa.
- `proposal_section_missing` - falta seção Problem, Goal ou Summary.
- `design_section_missing` - falta seção Overview, Approach, Design, Architecture, Technical Decisions ou Decisions.
- `tasks_phase_heading_missing` - existem tarefas de checkbox, mas não há cabeçalho de fase nível 2.
- `tasks_all_completed` - todas as tasks de checkbox estão marcadas; confirme evidências de implementação antes da revisão.
- `risks_mitigation_missing` - os riscos são listados sem notas de mitigação.
- `design_architecture_section_missing` - os arquivos de mudança mencionam `internal/core`, `internal/adapters` ou `internal/platform`, mas `design.md` não tem seção Architecture.
- `tasks_documentation_task_missing` - `proposal.md` ou `design.md` referenciam o CLI `specharbor`, mas `tasks.md` não tem tarefa de documentação.

A severidade determina status e códigos de saída:

- Sem achados de erro: status `valid`, código de saída `0`. Apenas avisos nunca fazem o comando falhar.
- Um ou mais achados de erro: status `invalid`, código de saída `1`.
- IDs de change ausentes ou inseguros (`..`, separadores, caminhos absolutos, início com `.` ou `-`, caracteres fora de `[A-Za-z0-9._-]`, mais de 128 caracteres) são rejeitados com erro de comando claro antes de qualquer acesso ao sistema de arquivos. Pontos internos únicos como `change.v1` são aceitos.

Exemplo de saída válida:

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 0
```

Exemplo de saída válida com avisos (código de saída `0`):

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

Exemplo de saída inválida (código de saída `1`):

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

Execute validação antes da implementação (para confirmar que o pacote de mudança está estruturalmente pronto), antes da revisão (para detectar lacunas de conteúdo com baixo custo) e antes de abrir PR (para manter pacotes de baixa qualidade fora do workflow compartilhado).

Alterações de comportamento intencionais em relação à validação anterior baseada apenas em presença:

- Mudanças de change não utilizáveis agora falham: arquivos obrigatórios vazios, arquivos sem cabeçalho ou sem corpo, arquivos de tasks sem checkboxes válidas, linhas de checkbox malformadas e arquivos acceptance-criteria sem item significativo produzem erros e saída não-zero.
- Uma change recém-gerada com `--blank` ou template agora valida como válida com avisos `boilerplate_only_content` (e `placeholder_content` aplicável) em vez de zero achados; o código de saída permanece `0`, portanto o fluxo documentado `generate -> validate` continua funcionando.
- IDs de change são validados com mais rigor e rejeitados antes de qualquer acesso ao sistema de arquivos.
- O relatório substitui a única contagem `Findings:` por contagens `Errors:` e `Warnings:` e agrupa os achados por severidade com caminhos de arquivo anexados.

### Gerar um Prompt por Papel

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

`prompt <change-id> --role <role>` imprime um prompt de agente para uma mudança OpenSpec existente. Para os papéis suportados, a geração do prompt é ciente de contexto: ela pode incluir uma seção dedicada `## Project Context` após as orientações iniciais de leitura e antes da tarefa.

Papéis suportados:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

Use `--role` para papéis de prompt. Flags de `agent-target` não são implementadas.

O contexto do projeto pode incluir:

- `User-confirmed context` de seções conhecidas de `.specharbor/project-brief.md`.
- `Detected facts` da descoberta local de contexto delimitada, com caminho de origem e confiança.
- `Suggested assumptions` da descoberta local de contexto delimitada, sempre rotuladas como suposições com caminho de origem e confiança.

A precedência é conservadora: contexto confirmado pelo usuário vence fatos detectados, fatos detectados vencem suposições sugeridas e suposições nunca são renderizadas como fatos. Se o contexto confirmado conflitar com fatos detectados, o prompt prefere o valor confirmado e pode incluir uma nota curta de conflito. Se o contexto estiver ausente ou ambíguo, o prompt gerado orienta o agente receptor a perguntar ou rotular explicitamente suposições em vez de inventar stack, arquitetura, comandos, decisões de persistência, decisões de workflow ou direção do projeto.

A geração de prompt lê apenas o resultado de descoberta classificado e renderiza resumos delimitados. Não despeja conteúdo bruto de arquivos, não executa comandos, não roda testes ou builds, não chama APIs de provedor, não executa agentes, não realiza RAG, não cria embeddings ou vector databases, não indexa o repositório, não faz descoberta remota e não automatiza controle de fonte.

### Revisar uma Mudança

```bash
go run ./cmd/specharbor review add-example-feature
```

`review <change-id>` revisa o pacote de mudança e o estado de conclusão de tarefas. Ele sai com código não-zero quando o status de revisão não está aprovado.

### Arquivar uma Mudança

```bash
go run ./cmd/specharbor archive add-example-feature
```

`archive <change-id>` move uma change concluída de `openspec/changes/<change-id>/` para o caminho de archive datado em `openspec/archive/<date>/<change-id>/`.

### Mostrar configuração local

```bash
go run ./cmd/specharbor config show
go run ./cmd/specharbor config
```

`config show` e `config` são somente leitura. Eles leem `.specharbor/config.yml` do projeto atual, suportam `version: 1` e imprimem o relatório de configuração local.

O comportamento atual de configuração não grava arquivos de config e não implementa:

- `config get`
- `config set`
- `config unset`

## Fluxo normal

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

Para executar o fluxo recomendado de nove etapas em detalhes, rode:

```bash
go run ./cmd/specharbor workflow
```

Uma sequência local típica é:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
```

Não archive uma change até que a implementação esteja concluída e revisada.

## Comportamento planejado

Os itens seguintes são direção de produto, não comportamento implementado de comando:

- comandos de runner genéricos orientados por configuração;
- configuração de provedor de IA;
- gerenciamento de API key de provedor;
- comandos de mutação de configuração;
- conectores de fluxo de trabalho externos.
