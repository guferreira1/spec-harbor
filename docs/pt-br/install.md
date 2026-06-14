# Instalação do SpecHarbor

SpecHarbor é distribuído por meio de ativos oficiais do GitHub Release construídos
pelas builds do `GoReleaser` para `guferreira1/spec-harbor`. Cada canal de
instalação baixa esses ativos por HTTPS e verifica checksums SHA-256 antes de
instalar.

## Status dos canais

| Canal | Status |
| --- | --- |
| GitHub Releases | Disponível para `v0.1.0` |
| `install.sh` | Disponível para Linux e macOS usando ativos de release reais |
| npm | Disponível como pacote sem escopo `specharbor@0.1.0` |
| Homebrew | Disponível como `brew install guferreira1/tap/specharbor` |
| `go install` from source | Opção de fallback/desenvolvedor; imprime metadados de fallback de desenvolvimento |
| automação de publicação de pacote | Automatizada no push de tags para npm e Homebrew |
| Pacotes Linux `.deb` / `.rpm` | Apenas futuro |
| Scoop / Winget do Windows | Apenas futuro |
| assinatura | Apenas futuro |
| SBOM | Apenas futuro |
| Docker | Apenas futuro |

Canais de instalação de binário exigem um GitHub Release publicado com ativos e
checksums correspondentes. A primeira release publicada é `v0.1.0`.

## Assets de release

Cada tag de release `vX.Y.Z` publica estes ativos:

```text
specharbor_Linux_x86_64.tar.gz
specharbor_Linux_arm64.tar.gz
specharbor_Darwin_x86_64.tar.gz
specharbor_Darwin_arm64.tar.gz
specharbor_Windows_x86_64.zip
specharbor_Windows_arm64.zip
checksums.txt
```

Os ativos compactados contém o binário `specharbor` (`specharbor.exe` no
Windows). `checksums.txt` contém uma linha SHA-256 por ativo. As URLs de
download dos ativos seguem este padrão:

```text
https://github.com/guferreira1/spec-harbor/releases/download/vX.Y.Z/<asset>
```

Os ativos de release cobrem Linux `amd64`, Linux `arm64`, macOS `amd64`, macOS
`arm64`, Windows `amd64` e Windows `arm64`. O `install.sh` suporta Linux e
macOS. O wrapper npm suporta Linux, macOS e Windows em `x64` e `arm64`. A
fórmula atual do Homebrew é somente para macOS.

## Instalação manual via GitHub Releases

1. Baixe o arquivo compactado para seu SO/arquitetura e `checksums.txt` a partir da
   [página de releases](https://github.com/guferreira1/spec-harbor/releases).

2. Verifique o checksum. No Linux:

   ```bash
   grep specharbor_Linux_x86_64.tar.gz checksums.txt | sha256sum -c -
   ```

   No macOS:

   ```bash
   grep specharbor_Darwin_arm64.tar.gz checksums.txt | shasum -a 256 -c -
   ```

   Não instale um arquivo compactado cujo checksum falhe na verificação.

3. Extraia e coloque o binário no seu `PATH`:

   ```bash
   tar -xzf specharbor_Linux_x86_64.tar.gz specharbor
   mkdir -p "$HOME/.local/bin"
   install -m 0755 specharbor "$HOME/.local/bin/specharbor"
   ```

   No Windows, extraia `specharbor.exe` do `.zip` e coloque-o em um diretório
   presente no seu `PATH`.

4. Verifique a instalação:

   ```bash
   specharbor version
   ```

## install.sh (Linux e macOS)

A raiz do repositório contém um script de instalação POSIX `sh` que automatiza o
fluxo manual: ele detecta SO e arquitetura, resolve a release mais recente (ou
uma versão fixada), baixa o arquivo correspondente e `checksums.txt` por HTTPS,
verifica o checksum SHA-256 e instala o binário em um diretório local do usuário.

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
```

Se preferir revisar o script antes de executá-lo (recomendado):

```bash
curl -sSLO https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh
less install.sh
sh install.sh
```

Opções:

```bash
# Fixe uma versão (variável de ambiente ou flag):
SPECHARBOR_VERSION=v0.1.0 sh install.sh
sh install.sh --version v0.1.0

# Substitua o diretório de instalação (padrão: $HOME/.local/bin):
SPECHARBOR_INSTALL_DIR="$HOME/bin" sh install.sh
sh install.sh --install-dir "$HOME/bin"

# Dry run: imprime SO, arquitetura, versão, URL do ativo e destino de
# instalação sem baixar o arquivo ou gravar algo:
sh install.sh --dry-run
```

Comportamento e garantias:

- Suporta Linux e macOS (Darwin) em `x86_64`/`amd64` e
  `aarch64`/`arm64`. Outras plataformas falham com erro claro apontando para isso.
- Strings de versão são validadas estritamente (`X.Y.Z` ou `vX.Y.Z`) antes de
  serem usadas em URLs.
- Os downloads são somente por HTTPS e restritos a
  URLs `https://github.com/guferreira1/spec-harbor/releases/`.
- A verificação SHA-256 contra `checksums.txt` é obrigatória. Se não houver
  ferramentas `sha256sum` ou `shasum`, o script falha ao invés de pular a
  verificação.
- Nunca invoca `sudo`. Se o diretório de instalação não permite escrita, o script
  falha com orientação para escolher um diretório local ao usuário.
- Nunca executa conteúdo baixado e escreve apenas em seu diretório temporário e no
  destino de instalação. Falhas removem arquivos temporários e não deixam binário
  parcial no destino de instalação.

Teste o script sem conexão com:

```bash
sh scripts/test-install-sh.sh
```

## Pacote npm global

O pacote wrapper npm está neste repositório em
`packages/npm/specharbor/` e é publicado no registro npm como `specharbor`.
Atualizações de versão são publicadas automaticamente quando uma tag `vX.Y.Z` é
enviada, após a existência dos ativos do GitHub Release; veja
[Metadados de release](release.md) para o workflow e segredos obrigatórios.

Instale com:

```bash
npm install -g specharbor
specharbor version
```

Você também pode executar o pacote publicado sem instalação global:

```bash
npx specharbor version
```

Como o wrapper funciona:

- A versão do pacote `X.Y.Z` fixa exatamente uma tag de release `vX.Y.Z`.
- Em `postinstall`, o wrapper detecta `process.platform`/`process.arch`,
  baixa o ativo de release correspondente e `checksums.txt` somente por HTTPS de
  `https://github.com/guferreira1/spec-harbor/releases/download/`,
  valida o checksum SHA-256 com `crypto` do Node e extrai o binário no diretório
  `native/` do pacote.
- A instalação com `--ignore-scripts` pula o `postinstall`; o launcher faz o
  mesmo download verificado com checksum na primeira execução.
- O launcher encaminha argumentos e stdio para o binário nativo usando APIs de
  chamada por array (sem strings de shell) e preserva o código de saída.
- Plataformas não suportadas falham com erro determinístico informando a
  plataforma e apontando para este documento, tanto em postinstall quanto na
  execução.

Consulte [packages/npm/specharbor/README.md](../packages/npm/specharbor/README.md)
para mais detalhes e execute `npm test` para a suíte de testes offline.

## Tap do Homebrew

O suporte do Homebrew está disponível pelo repositório externo de tap pessoal
`guferreira1/homebrew-tap`, com a fórmula chamada `specharbor`. Instale com:

```bash
brew install guferreira1/tap/specharbor
```

A fórmula satisfaz estas expectativas no repositório de tap:

- `url` aponta para um ativo oficial e fixado do GitHub Release.
- O valor `sha256`, copiado de `checksums.txt` da release, é obrigatório para
  cada ativo referenciado.
- A fórmula instala o binário pré-compilado; não faz build da fonte.
- O bloco `test do` da fórmula executa `specharbor version` e valida se a saída
  contém a versão esperada.

O tap valida `brew audit --strict --online specharbor`, instalação da fórmula,
`specharbor version`, `brew test specharbor` e o comando de instalação do usuário
nos runners macOS do GitHub Actions. A fórmula é atualizada automaticamente a
cada release de tag `vX.Y.Z` pelo job `homebrew-publish`, que gera
`Formula/specharbor.rb` a partir de `checksums.txt`; veja
[Metadados de release](release.md). Nenhum arquivo de fórmula Homebrew é commitado
neste repositório.

## Canais somente futuros

Os itens a seguir estão planejados, mas ainda não implementados:

- Pacotes nativos Linux (`.deb` e `.rpm`), provavelmente via nfpm em uma mudança
  futura.
- Gerenciadores de pacote do Windows: Scoop e Winget.
- Assinatura de binário (por exemplo, cosign), geração de SBOM, imagens Docker e
  mecanismos de atualização automática.
- Automação de publicação de pacote para pacotes Linux, gerenciadores Windows,
  assinatura, SBOMs e Docker.

A publicação de npm e Homebrew é automatizada em releases de tag; veja
[Metadados de release](release.md). Caminhos intermediários: usuários Linux têm
`install.sh`, npm e instalação manual; usuários Windows têm npm e instalação
manual via `.zip`.

## Verificando uma instalação

```bash
specharbor version
```

Um binário de release imprime metadados de release injetados:

```text
SpecHarbor 0.1.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

Uma build de fonte sem metadados injetados imprime o fallback de desenvolvimento:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Veja [Metadados de release](release.md) para a convenção completa de versão.

## Solução de problemas

### `specharbor: command not found`

Diretórios de instalação locais do usuário, como `$HOME/.local/bin`, podem não
estar no seu `PATH`. Adicione o diretório ao perfil do shell:

```bash
# bash: ~/.bashrc — zsh: ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```

Em seguida, reinicie o shell ou execute `source` no perfil. Verifique com:

```bash
command -v specharbor
```

Se um shell ainda resolver um binário antigo após a substituição, limpe o cache
de comandos e tente novamente:

```bash
hash -r
specharbor version
```

### Permissão negada

`install.sh` nunca invoca `sudo`. Se o diretório de instalação não permitir
escrita, escolha um diretório gravável pelo usuário:

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh -s -- --install-dir "$HOME/bin"
```

Para instalações manuais, crie você mesmo o diretório de destino e utilize um
diretório que já esteja no `PATH`.

### Incompatibilidade de checksum

Não instale um arquivo compactado cujo checksum não valide. Exclua o download
parcial, baixe novamente o arquivo e `checksums.txt` do GitHub Release oficial e
tente novamente. Se a incompatibilidade persistir, trate o artefato como não
confiável e abra uma issue com a URL da release e a saída de checksum.

### Plataforma ou arquitetura não suportada

Existem ativos de release para Linux, macOS e Windows em `amd64`/`x64` e
`arm64`. `install.sh` suporta apenas Linux e macOS. Para sistemas sem suporte,
use uma máquina suportada, teste o pacote npm se o Node suportar seu par
SO/arquitetura, ou faça build da fonte com Go.

### `postinstall` do npm ignorado ou falha no download da primeira execução

`npm install --ignore-scripts -g specharbor` pula o download de `postinstall`.
Isso é suportado: `npx specharbor version` ou `specharbor version` fazem o mesmo
download verificado por checksum na primeira execução. Falhas na primeira
execução geralmente indicam uso offline, restrições de proxy, restrições de acesso
ao GitHub ou uma plataforma não suportada. Reexecute com acesso de rede aos GitHub
Releases, ou use uma instalação manual por GitHub Release.

### Problemas do tap/instalação Homebrew

Use o atalho da tap exatamente:

```bash
brew install guferreira1/tap/specharbor
```

Se o Homebrew não encontrar ou atualizar a fórmula, atualize os taps e tente
novamente:

```bash
brew update
brew untap guferreira1/tap
brew tap guferreira1/tap
brew install guferreira1/tap/specharbor
```

O repositório de tap externo é `guferreira1/homebrew-tap`; nenhum arquivo de
fórmula vive neste repositório.

### Os metadados de versão parecem inesperados

Verifique o binário instalado com:

```bash
specharbor version
```

Binários de release de `v0.1.0` imprimem `SpecHarbor 0.1.0` e incluem metadados
de commit, date e dirty. Saída `dev`/`unknown` geralmente significa que o binário
foi construído a partir da fonte sem metadados de release injetados, como com `go
install`.

## fallback `go install`

Build a partir da fonte é o fallback manual documentado e exige Go:

```bash
go install github.com/guferreira1/spec-harbor/cmd/specharbor@latest
```

Fixe uma tag com `@vX.Y.Z` quando já existirem releases. Binários construídos
dessa forma usam metadados de fallback de desenvolvimento (`dev`/`unknown`) porque
nenhum metadado de release é injetado; essa saída é esperada e documentada em
[Metadados de release](release.md).

## Modelo de segurança

Todos os canais de instalação seguem as mesmas regras:

- Somente HTTPS; os downloads são restritos a URLs oficiais
  `https://github.com/guferreira1/spec-harbor/releases/`.
- A verificação do checksum SHA-256 contra `checksums.txt` da release é
  obrigatória para cada pacote baixado; falha na verificação aborta a instalação
  e remove arquivos parciais.
- Checksums protegem contra corrupção e adulteração em trânsito; são servidos a
  partir da mesma release dos ativos, portanto não protegem contra uma release
  totalmente comprometida. Assinatura é trabalho futuro.
- Nenhum canal executa scripts baixados, monta strings de shell a partir da
  entrada do usuário, exige tokens ou autenticação, envia telemetria, altera estado
  do Git, ou grava fora do diretório de instalação e pastas cache/temporárias.
- A publicação de npm e Homebrew é automatizada em releases de tag `vX.Y.Z` por
  meio de um fluxo de trabalho somente de tag que valida consistência de versão e
  executa testes de pacote antes de publicar; veja
  [Metadados de release](release.md).
