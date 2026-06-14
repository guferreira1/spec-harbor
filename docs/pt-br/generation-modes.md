# Modos de geração

O SpecHarbor implementa atualmente prompts de geração interativa, geração em
branco, geração por template integrado, geração por template customizado local ao
projeto, aliases de template orientados por configuração, geração híbrida,
geração guiada, geração assistida por IA a partir de arquivo, autoria de
especificação com `agent-assisted` em dry-run e execução local explícita de runner
com `agent-assisted` no modo run-and-report.

## Implementado

### Geração interativa

```bash
go run ./cmd/specharbor generate <change-id> --interactive
```

A geração interativa é uma camada de prompt de CLI sobre o comportamento de geração
existente. Ela mantém `<change-id>` explícito na linha de comando, exige um TTY,
coleta apenas valores de caminho específicos, valida respostas, imprime um
resumo de pré-confirmação determinístico e delega para o caminho de geração
existente apenas após confirmação.

Caminhos interativos suportados nesta versão:

- blank
- template integrado
- template customizado
- template por configuração
- hybrid

Excluídos dos prompts interativos nesta versão:

- geração guiada direta
- geração assistida por IA
- geração com `agent-assisted`
- execução local de runner de agente
- aplicação de saída de runner ao vivo
- URLs de template remoto em bruto ou checksums

O primeiro menu possui ordenação determinística:

```text
Selecione o caminho de geração:
1. blank
2. built-in template
3. custom template
4. config template
5. hybrid
```

Prompts específicos de caminho:

- Blank não faz perguntas adicionais.
- Template integrado pede um dos templates integrados suportados: `feature`, `bugfix`, `docs` ou `refactor`.
- Template customizado pede o nome do template customizado e título/resumo opcionais.
- Template por configuração pede um alias de configuração e título/resumo opcionais.
- Híbrido pede exatamente um namespace de fonte (`built-in template`, `custom template` ou `config template`), o valor da fonte, título obrigatório, resumo obrigatório e tipo opcional.

Exemplos de sequência de respostas:

```text
blank:           1 -> y
built-in:        2 -> feature -> y
custom:          3 -> api-feature -> Add payments -> Adds a payment flow. -> y
config:          4 -> default-feature -> empty title -> empty summary -> y
hybrid built-in: 5 -> 1 -> feature -> Add login -> Add login support -> empty type -> y
```

Antes da escrita, cada fluxo interativo imprime um resumo determinístico com id da
mudança, caminho selecionado, valores de fonte selecionados, destino de escrita
esperado, nomes de arquivos aprovados, comportamento de validação e notas de
segurança.
Resumos de blank, template integrado, template customizado e template por
configuração mostram:

```text
Validation: automatic no
```

Resumos híbridos mostram:

```text
Validation: automatic yes
```

A seção de segurança é sempre exibida antes da confirmação:

```text
Safety:
- A escrita está limitada a arquivos de mudança OpenSpec.
- O código de produção não será modificado.
- Comandos de controle de fonte não serão executados.
- A automação de workflow não será acionada.
- APIs de provedor, LLM e agente não serão chamadas.
- Nenhum auto-commit, auto-push, criação de PR, merge ou archive será realizado.
```

A confirmação é com trim e insensível a maiúsculas/minúsculas. `y` e `yes`
avançam em qualquer caixa; `n` e `no` cancelam em qualquer caixa. Confirmação
vazia e EOF também cancelam. O cancelamento encerra com código não-zero,
`operation cancelled` e sem escrita. Respostas obrigatórias inválidas e respostas
de confirmação inválidas tentam até 3 vezes, depois falham claramente e não escrevem.

O modo interativo preserva o comportamento de escrita e validação do modo
selecionado. Ele não resolve aliases de configuração, não busca templates remotos
e não chama casos de uso de geração antes da confirmação. Templates remotos
continuam alcançáveis apenas por aliases de configuração existentes e mantêm as
salvaguardas existentes de HTTPS, checksum, ZIP, sem credencial, sem query,
sem fragmento, sem script, sem código de produção e escrita apenas de OpenSpec.

O modo interativo não chama APIs de provedor, APIs LLM, APIs de modelo local,
agentes, ferramentas de controle de fonte, ferramentas de workflow, comandos shell
ou scripts. Não realiza escrita de código de produção, mutação de configuração,
auto-commit, auto-push, criação de pull request, merge ou automação de
`archive`.

### Geração em branco

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

A geração em branco cria a estrutura de arquivos de mudança OpenSpec para o usuário
preencher o conteúdo manualmente.

### Geração por template built-in

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

Os templates integrados implementados são exatamente:

- `feature`
- `bugfix`
- `docs`
- `refactor`

A geração por template integrado escreve conteúdo inicial genérico, local e
independente, para o template selecionado.

### Geração com template customizado

```bash
go run ./cmd/specharbor generate <change-id> --custom-template <template-name>
```

A geração com template customizado renderiza templates locais reutilizáveis de OpenSpec
no projeto. Um template customizado é um diretório simples em `.specharbor/templates/`
com todos os cinco arquivos obrigatórios da mudança OpenSpec:

```text
.specharbor/templates/<template-name>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Todos os cinco arquivos são obrigatórios e devem ser não vazios. Um diretório de
template ausente, arquivos obrigatórios ausentes ou arquivo vazio (`whitespace`)
causam erro claro e geram zero escritas — nem mesmo o diretório da mudança é criado.
Arquivos extras e subdiretórios desconhecidos dentro do diretório do template são
ignorados e nunca copiados.

Nomes de template customizado são validados antes de qualquer acesso ao sistema de
arquivos: caracteres permitidos `[A-Za-z0-9._-]`, nome com um único segmento,
sem `/` ou `\`, sem sequências `..`, sem início com `.` ou `-`, e no máximo
128 caracteres.

O conteúdo do template suporta apenas substituição determinística mínima de variáveis:

- `{{change_id}}` é sempre substituído pelo id da mudança.
- `{{title}}` é substituído apenas quando a flag opcional `--title` é informada.
- `{{summary}}` é substituído apenas quando a flag opcional `--summary` é informada.
- Tokens `{{...}}` não reconhecidos ou não resolvidos permanecem sem erros na saída.

```bash
go run ./cmd/specharbor generate <change-id> --custom-template <template-name> --title "<title>" --summary "<summary>"
```

Não existe linguagem de template: sem condicionais, sem laços, sem funções, sem
`includes` e sem variáveis definidas no template. Templates são conteúdo Markdown
estático e nunca são executados; não há scripts, hooks, comandos shell ou
processos externos durante a geração.

Templates integrados e customizados resolvem de fontes distintas:

- `--template <name>` resolve apenas os quatro templates integrados; seu
  comportamento, conteúdo, saída e erros por nome desconhecido não mudam.
- `--custom-template <name>` resolve apenas `.specharbor/templates/<name>/`.
- `--config-template <alias>` resolve apenas aliases declarados em
  `.specharbor/config.yml`.
- Um template customizado com nome igual a um integrado (ex.: `feature`) não sobrescreve
  nem oculta o template integrado.

`--custom-template` é mutuamente exclusivo com `--blank`, `--template`, `--guided`
e `--agent-assisted`. O relatório de sucesso identifica o template como custom,
mostra o caminho relativo da fonte e lista arquivos criados e ignorados, além de
registrar que apenas arquivos OpenSpec foram gravados.

Gerações personalizadas funcionam com `specharbor validate <change-id>` como qualquer
outra mudança; a geração não executa validação automaticamente, e as falhas de validação
dependem da qualidade do conteúdo do template.

Limites de segurança para templates customizados diretos: templates são locais e de
escopo do projeto; não há marketplace, não há chamadas de provedor, não há
credenciais e não há gravações fora de `openspec/changes/<change-id>/`.
Arquivos existentes são ignorados e nunca sobrescritos.

### Aliases de template por configuração

```bash
go run ./cmd/specharbor generate <change-id> --config-template <alias>
go run ./cmd/specharbor generate <change-id> --config-template <alias> --title "<title>" --summary "<summary>"
```

Templates orientados por configuração permitem que o projeto defina aliases estáveis
para templates integrados, customizados locais ou remotos HTTPS fixos. Templates
remotos são usados apenas por `--config-template`; não existe `--remote-template`.
Os aliases ficam em `.specharbor/config.yml`:

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

`version: 1` é obrigatória para geração orientada por configuração. A ausência de
`templates` ou `templates.aliases` significa que não há aliases. Arquivo de config
ausente, versão ausente/inválida, YAML inválido, entrada de alias inválida ou alias
solicitado inexistente retornam erro claro.

Os tipos de fonte suportados são exatamente:

- `builtin`, resolve apenas o template integrado informado.
- `custom`, resolve apenas `.specharbor/templates/<template-name>/`.
- `remote`, busca uma única URL HTTPS de ZIP com checksum obrigatório
  `sha256:<64-hex>`.

Aliases remotos exigem `url`, `checksum` e `format`; somente `format: zip` é
suportado. Aliases remotos rejeitam `template` e campos desconhecidos. Aliases
integrados e customizados continuam exigindo `template` e rejeitam campos exclusivos de
remoto, como `url`, `checksum` e `format`.

As regras de URL remota são estritas: apenas HTTPS, host e caminho obrigatórios,
sem credenciais/userinfo, sem query string, sem fragmento, sem redirects, sem URLs
locais, sem SSH/git/git+ssh/FTP e sem alvos no estilo SCP. O checksum é
verificado sobre os bytes do ZIP baixado antes de qualquer parsing de arquivo.

Pacotes ZIP remotos devem conter exatamente estes cinco arquivos raiz não vazios:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

Caminhos aninhados, caminhos absolutos, traversal, caminhos de disco Windows,
symlinks, entradas executáveis, arquivos duplicados, arquivos extras, arquivos
faltantes, arquivos vazios, ZIPs malformados, HTTP oversized e conteúdo não
comprimido oversized são rejeitados e não gravam nada.

Aliases são segmentos de caminho seguros: não vazios, no máximo 128 caracteres,
caracteres permitidos `[A-Za-z0-9._-]`, sem `/` ou `\`, sem caminhos absolutos,
sem traversal nem sequência `..`, sem `.` ou `-` no início. Aliases CLI inválidos
falham antes da resolução do template, e aliases inválidos no config falham durante
validação.

A geração orientada por configuração delega ao comportamento da origem resolvida:

- Alias integrado gera os mesmos arquivos de `--template <name>`.
- Alias customizado usa o mesmo diretório do template customizado, validação de
  arquivos obrigatórios e substituição determinística `{{change_id}}`,
  `{{title}}` e `{{summary}}` de `--custom-template <name>`.
- Alias remoto grava o conteúdo verificado do ZIP como Markdown OpenSpec; caminhos
  de arquivo do arquivo remoto nunca influenciam caminhos de saída, e não há scripts
  ou comandos shell remotos executados.
- `--title` e `--summary` opcionais passam para o comportamento resolvido; templates
  integrados não usam esses valores.

Namespaces são disjuntos por flag. `--template feature`, `--custom-template feature`
e `--config-template feature` são três buscas distintas. Não existe shadowing,
fallback ou inferência de fonte.

`--config-template` é mutuamente exclusivo com `--blank`, `--template`,
`--custom-template`, `--guided`, `--agent-assisted`, `--ai-assisted` e `--execute`.
Também não aceita `--type`, `--agent`, `--from-file` ou `--overwrite`.

Limites de segurança: aliases remotos orientadas por configuração não possuem cache
persistente nesta primeira versão e não suportam credenciais, OAuth, cabeçalhos de
autenticação, cookies, expansão de token de ambiente, git clone, descoberta de
marketplace, APIs de provedor, scripts de template, execução de shell, escrita de
código de produção, automação de controle de fonte, auto-commit, criação de PR,
automação de merge ou archive. Arquivos gerados permanecem em
`openspec/changes/<change-id>/`, usam os cinco nomes obrigatórios do OpenSpec e não
sobrescrevem arquivos existentes.

### Geração híbrida

```bash
go run ./cmd/specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
go run ./cmd/specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
go run ./cmd/specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
```

A geração híbrida combina exatamente uma fonte de template determinística com metadados
obrigatórios guiados e depois valida a mudança OpenSpec gerada. É útil quando a equipe
quer uma fonte de template reutilizável e metadados explícitos de título/resumo em
um único comando seguro.

Seletores de fonte suportados são exatamente:

- `--template <name>` para templates integrados.
- `--custom-template <name>` para `.specharbor/templates/<name>/`.
- `--config-template <alias>` para aliases de `.specharbor/config.yml`.

Exatamente um seletor de fonte é obrigatório. Falta ou múltipla seleção falham antes
de escrita. Não há tentativa de fonte, fallback, shadowing ou ordem de precedência.

Híbrido exige `--title` e `--summary`; ambos sofrem trim e precisam ser não vazios.
`--type` é opcional e deve ser exatamente `feature`, `bugfix`, `docs` ou `refactor`
quando informado.

Comportamento de tipo é específico por fonte:

- `--template feature`, `bugfix`, `docs` ou `refactor` diretas derivam tipo omitido do
  template integrado selecionado.
- Um alias de configuração que resolva para template integrado deriva o tipo omitido
  a partir do template integrado resolvido.
- O tipo informado deve corresponder a um template integrado selecionado ou resolvido.
  `--hybrid --template feature --type feature` passa; `--hybrid --template feature --type bugfix`
  falha e não escreve.
- Templates customizados, aliases customizados por config e aliases remotos por config
  não inferem tipo. Se `--type` for omitido, `{{type}}` permanece não resolvido.
  Se informado, é renderizado.

O render híbrido substitui `{{change_id}}`, `{{title}}` e `{{summary}}` e substitui
`{{type}}` apenas quando existe tipo efetivo provido. Tokens desconhecidos ou não
resolvidos permanecem sem alteração. Híbrido não adiciona condicionais, laços,
funções, includes, hooks, scripts, comandos shell ou comportamento de template
executável.

Exemplos:

```bash
go run ./cmd/specharbor generate add-login --hybrid --template feature --title "Add login" --summary "Add an OpenSpec change for login"
go run ./cmd/specharbor generate add-login --hybrid --template feature --type bugfix --title "Add login" --summary "Add an OpenSpec change for login"
go run ./cmd/specharbor generate add-payment-flow --hybrid --custom-template api-feature --title "Add payments" --summary "Adds a payment flow."
go run ./cmd/specharbor generate add-login --hybrid --config-template default-feature --title "Add login" --summary "Add login support"
go run ./cmd/specharbor generate add-payment-flow --hybrid --config-template api-feature --title "Add payments" --summary "Adds a payment flow." --type feature
go run ./cmd/specharbor generate add-service --hybrid --config-template service-feature --title "Add service" --summary "Adds a service workflow."
```

O primeiro exemplo deriva `type=feature`.

O segundo exemplo falha porque `bugfix` não combina com template integrado `feature`.

O exemplo customizado não infere tipo, então `{{type}}` permanece não resolvido,
exceto se `--type` for informado.

Geração híbrida remota está disponível apenas com `--config-template <alias>` que
resolve para `source: remote`. Mantém as salvaguardas já existentes de template
remoto: apenas HTTPS, sem credenciais, sem query strings, sem fragmentos, sem
redirects, checksum obrigatório e validado antes do parsing do ZIP, apenas ZIP, segurança
rígida de arquivo, sem cache, sem execução de shell/script, sem escrita de código de
produção e sem caminhos arbitrários de saída.

Híbrido grava somente `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md` e
`risks.md` em `openspec/changes/<change-id>/`. Arquivos existentes são ignorados e
preservados; `--overwrite` é rejeitado nesta versão inicial.

A validação ocorre após escritas bem-sucedidas ou conclusão com apenas pulos. O
relatório inclui status, quantidade de arquivos obrigatórios, quantidade de erros,
quantidade de avisos e achados. Achados com severidade de warning mantêm código de saída `0`; erros de validação
geram saída não-zero após exibir o relatório. Híbrido não corrige achados
automaticamente.

Fora de escopo para geração híbrida: `--blank`, overlay de IA, `--from-file`,
aplicação de saída de runner ao vivo, `--agent`, `--execute`, APIs de provedor,
APIs de LLM, APIs de modelo local, ferramentas de controle de fonte, ferramentas de
workflow, comandos shell, scripts, escrita de código de produção, auto-commit,
auto-push, criação de PR, automação de merge e archive.

### Geração guiada

```bash
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

Os tipos guiados implementados são exatamente:

- `feature`
- `bugfix`
- `docs`
- `refactor`

A geração guiada escreve conteúdo inicial local e determinístico baseado nas entradas
explícitas de CLI. É não interativa e não pede input durante a execução; usa os
valores de `--type`, `--title` e `--summary` informados.

O conteúdo gerado guiado inclui título e resumo fornecidos. O conteúdo gerado pode
ser editado com segurança e não significa que o SpecHarbor tenha inferido requisitos
específicos do projeto além dos inputs fornecidos.

### Geração assistida por IA a partir de arquivo

```bash
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file>
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --overwrite
```

A geração assistida por IA importa conteúdo OpenSpec gerado por IA de um arquivo local
fornecido explicitamente pelo usuário. O arquivo deve usar blocos delimitadores
estritos:

```text
---FILE: proposal.md---
...
---END FILE---
---FILE: design.md---
...
---END FILE---
---FILE: tasks.md---
...
---END FILE---
---FILE: acceptance-criteria.md---
...
---END FILE---
---FILE: risks.md---
...
---END FILE---
```

Nomes de bloco permitidos são exatamente `proposal.md`, `design.md`, `tasks.md`,
`acceptance-criteria.md` e `risks.md`. Todos os cinco são obrigatórios. Nomes
desconhecidos, duplicados, blocos ausentes, conteúdo vazio, caminhos absolutos,
traversal, caminhos aninhados, delimitadores malformados, wrappers com fence,
formatos patch/diff e texto fora dos blocos são rejeitados antes de escrita.

O comando parseia a fonte completa antes de qualquer escrita-alvo, cria
`openspec/changes/<change-id>/` quando necessário, grava somente os cinco arquivos
aprovados sob esse diretório e executa validação após escritas ou skips. Arquivos
existentes são ignorados por padrão; `--overwrite` é obrigatório para substituí-los.
Targets de symlink de saída existentes são rejeitados em vez de seguidos. Warnings
de validação mantêm exit code `0`; erros de validação são exibidos e fazem o comando
encerrar com não-zero.

A geração assistida por arquivo não é geração com provedor e não é aplicação de saída
do runner ao vivo. Não chama APIs de provedor, serviços remotos de IA, APIs de
modelo local, OAuth, credenciais, comandos shell, automação de controle de fonte,
automatização de workflow ou runners de agente. Não modifica código de produção,
não aplica patches, não auto-commit, não auto-push, não cria PRs, não merge e não
arquiva.

### Autoria de especificação com agente

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

Os tipos de autoria com agente implementados são exatamente:

- `feature`
- `bugfix`
- `docs`
- `refactor`

O dry-run continua sendo padrão. Sem `--execute`, a autoria de especificação com
agente imprime um plano determinístico e um prompt copy-pasteable em stdout.

O prompt gerado foi feito para ajudar um agente externo a autorar ou refinar apenas o
pacote de mudança OpenSpec. A implementação permanece para etapa posterior no fluxo
normal do SpecHarbor.

Dry-run de autoria com agente:

- não grava arquivos;
- não grava arquivo de prompt;
- não cria nem modifica arquivos OpenSpec;
- não cria nem modifica código de produção;
- não executa agentes;
- não requer runner;
- não resolve mapeamentos de comando executável;
- não chama APIs de provedor;
- não chama modelos locais;
- não chama APIs de rede;
- não chama APIs de controle de fonte;
- não chama ferramentas de workflow.

Alvos de agente reconhecidos são:

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

Dry-run com agente desconhecido é rejeitado como endurecimento intencional de validação.

O modo de execução é explícito:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

Modo execução envia o mesmo prompt determinístico de autoria OpenSpec via stdin para
um comando local suportado. O diretório de trabalho é a raiz do projeto atual.

Mapeamentos de comando local executável são:

- `codex -> codex`
- `claude -> claude`
- `devin -> devin`
- `cursor -> cursor`
- `copilot -> copilot`
- `gemini -> gemini`
- `roo -> roo`
- `windsurf -> windsurf`
- `aider -> aider`

`generic` é reconhecido apenas para dry-run até uma futura funcionalidade de runner/
templates orientada por configuração definir um mapeamento determinístico. `--execute
--agent generic` falha com erro claro de target não mapeado.

Modo execução continua run-and-report apenas:

- comandos locais ausentes geram erros de inicialização sem resultado de runner e sem exit code;
- comandos iniciados com código não-zero geram relatório e depois SpecHarbor sai não-zero;
- stdout, stderr, código de saída e status de execução são capturados e reportados
  para processos iniciados;
- SpecHarbor não faz parse nem aplica saída;
- SpecHarbor não grava arquivos a partir da saída;
- SpecHarbor não modifica código de produção a partir da saída;
- SpecHarbor não auto-commita, auto-pusha ou auto-mescla.

APIs de provedor, automação de IDE, OAuth, credenciais, integrações com marketplace,
execução remota, automação de controle de fonte e automação de workflow continuam fora
de escopo. O comportamento local do comando de agente é controlado pela ferramenta
instalada.

Blank, template integrado, template customizado, config-template, híbrido, guiado e
assistido por IA criam os arquivos OpenSpec obrigatórios:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Arquivos existentes são ignorados e não sobrescritos para blank, template integrado,
template customizado, config-template, híbrido e geração guiada. Geração assistida por
IA também ignora arquivos existentes por padrão e substitui apenas com `--overwrite`
explícito. Diretórios de mudança parcialmente existentes podem ser recuperados porque
a geração cria apenas arquivos obrigatórios ausentes.

## Planejado

Os itens a seguir são direção de produto, não comportamento de comando implementado:

- comandos de runner genéricos orientados por configuração;
- prompts interativos.

Configuração detalhada de provedor, automação de IDE, integrações de marketplace,
execução remota, automação de workflow e comportamento de aplicação de arquivo
não fazem parte do conjunto de geração implementado atualmente.
