# SpecHarbor

**SpecHarbor** é uma CLI em Go para fluxos de trabalho com agentes de IA para programação baseados em OpenSpec.

Ela ajuda uma equipe a transformar uma ideia solta em um pacote de mudança com escopo explícito, que humanos e agentes podem seguir com trilhas de validação e revisão.

## O que é o SpecHarbor

SpecHarbor organiza o fluxo OpenSpec/SDD para agentes e colaboradores:

- iniciar mudanças com pacotes OpenSpec estruturados;
- coletar contexto local de forma explicitada;
- gerar prompts por função para execução de implementação;
- validar conteúdo e revisar checklists;
- arquivar mudanças concluídas.

Esse pacote não executa implementação automaticamente; ele prepara e coordena a parte operacional de planejamento e segurança do fluxo.

## Por que ele existe

Agentes de IA funcionam melhor quando recebem instruções explícitas.
Instruções vagas geram risco de:

- editar arquivos não relacionados;
- ignorar arquitetura;
- pular testes;
- inventar requisitos;
- introduzir comportamento inconsistente.

O SpecHarbor reduz esse risco ao concentrar o trabalho em um pacote OpenSpec auditável para cada mudança.

## Para quem é

- desenvolvedores usando Codex, Claude Code, Cursor, Devin, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, Aider ou agentes genéricos;
- equipes que adotam OpenSpec/SDD;
- mantenedores que precisam de trilha de revisão para trabalho com agentes;
- desenvolvedores individuais que querem reduzir risco em fluxos com IA.

## Fluxo principal

```text
Ideia -> Mudança OpenSpec -> tasks.md -> Prompt de agente -> Implementação -> Revisão -> Arquivamento
```

Atalhos de comando equivalentes:

```text
specharbor generate <change-id> ... -> specharbor validate <change-id> -> specharbor prompt <change-id> --role <role> -> specharbor review <change-id> -> specharbor archive <change-id>
```

`specharbor workflow` exibe o fluxo de nove etapas com os papéis recomendados.

## Conceitos principais

- **OpenSpec change**
  Cada trabalho está em `openspec/changes/<change-id>/` com cinco arquivos obrigatórios:
  `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md` e `risks.md`.

- **tasks.md**
  Lista as tarefas de implementação e acompanha o progresso da mudança.

- **agent roles**
  Roles suportadas: `spec-author`, `architecture-reviewer`, `implementer`, `test-engineer`, `change-reviewer`.

- **project brief**
  `specharbor brief` cria `.specharbor/project-brief.md` com contexto confirmado e decisões do projeto.

- **context discover**
  `specharbor context discover` lê fontes locais limitadas e retorna contexto confirmado, fatos detectados e suposições sugeridas com fonte e confiança.

- **repository context index**
  `specharbor context index` gera o inventário local de metadados em `.specharbor/context-index.json`, sem guardar conteúdo bruto dos arquivos.

- **validação**
  `specharbor validate <change-id>` verifica estrutura e regras de qualidade para os cinco arquivos da mudança.
  Erros bloqueiam o comando; avisos não falham por padrão.

- **review**
  `specharbor review <change-id>` revisa o status para evitar que mudanças incompletas avancem no fluxo.

- **arquivamento**
  `specharbor archive <change-id>` move a mudança concluída para `openspec/archive/` com data.

## Instalação

Canais disponíveis:

- **install.sh**

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
```

- **npm**

```bash
npm install -g specharbor
```

- **Homebrew**

```bash
brew install guferreira1/tap/specharbor
```

- **GitHub Releases**
- **go install** (opção de desenvolvimento)

Verifique com:

```bash
specharbor version
```

Consulte [docs/install.md](../install.md) para suporte por plataforma, checksums e troubleshooting.

## Quickstart

```bash
cd /caminho/do/projeto
specharbor init
specharbor generate add-login-feature --guided --type feature --title "Adicionar login" --summary "Adicionar fluxo de login seguro"
specharbor validate add-login-feature
specharbor prompt add-login-feature --role implementer
specharbor review add-login-feature
specharbor archive add-login-feature
```

Extras úteis de contexto:

```bash
specharbor context discover
specharbor context index --write
specharbor brief
```

## Recursos atuais

- inicialização de projeto OpenSpec (`init`);
- descoberta de contexto e índice de contexto do repositório;
- briefing de contexto (`brief`, `brief --update`);
- geração de mudanças (blank, guided, templates, custom templates, config templates, hybrid, ai-assisted de arquivo, agent-assisted);
- validação (`validate`);
- prompts por função (`prompt --role`);
- revisão (`review`);
- arquivamento (`archive`);
- orientação de workflow (`workflow`);
- suporte de instalação e canais `install.sh`, GitHub Releases, npm e Homebrew.

## Modelo de segurança

- comportamento principal local-first para descoberta, índice, validação e revisão;
- confirmação explícita para escrita de contexto;
- sem auto-commit, auto-push, criação automática de PR, merge ou arquivamento;
- sem modificação de código de produção durante geração/revisão/validação;
- sem chamadas de API de provedores de modelo no fluxo atual;
- paths e symlinks tratados de forma segura;
- wrapper npm sem construção de shell string;
- separação clara entre contexto confirmado, fatos detectados e suposições;
- instalação sempre com verificação de checksum antes da execução.

## Navegação da documentação

- [Instalação](../install.md)
- [Uso](../usage.md)
- [Fluxo de trabalho](../workflow.md)
- [Funções de agente](../agent-roles.md)
- [Modos de geração](../generation-modes.md)
- [Metadados de release](../release.md)
- [Contribuição](../contributing.md)

## Status e roadmap

Status atual:

- OpenSpec inicializado e fluxo de mudança estruturado;
- descoberta de contexto e índice local;
- múltiplos modos de geração;
- validação e revisão;
- distribuição via wrapper npm e canais oficiais.

Itens planejados (já documentados):

- pacotes Linux `.deb`/`.rpm` (futuro);
- integração com gerenciadores Windows (Scoop/Winget);
- comandos de mutação de config (`config get/set/unset`);
- integração com provedores de IA e conectores de workflow mais amplos.

Veja [context-initiative-remaining-plan.md](planning/context-initiative-remaining-plan.md).

## Contribuição

- comece mudanças significativas com um pacote em `openspec/changes/<change-id>/`;
- mantenha o trabalho alinhado à proposta e atualize `tasks.md`;
- execute validação (`go test ./...`) antes de fechar;
- respeite os limites de arquitetura do projeto.

Consulte [docs/contributing.md](../contributing.md) para mais detalhes.

## Licença

MIT
