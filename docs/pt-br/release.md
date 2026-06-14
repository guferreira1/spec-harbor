# Metadados de release

O SpecHarbor usa GoReleaser para construir ativos do GitHub Release a partir de
tags de versão enviadas.

## Release pública atual

A release pública atual é o GitHub Release `v0.1.0`, construído do commit
`e6faff91feef07e5c1e47181243286268daf17b5`. Binários de release exibem a versão
`0.1.0` simples.

Os canais de distribuição pública validados para esta release são:

- GitHub Releases.
- `install.sh`.
- npm package `specharbor@0.1.0`.
- Homebrew tap install command `brew install guferreira1/tap/specharbor`.

A fórmula Homebrew está no repositório externo de tap
`guferreira1/homebrew-tap`; o commit de tap validado é
`a61783bcfa44f7eafdce72c70043b76e6f80df9c`.

## Verificar versão

Execute:

```bash
specharbor version
```

Saída padrão de desenvolvimento:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Campos:

- `version`: metadado de versão do produto exibido na primeira linha.
- `commit`: commit de origem fornecido pelo build.
- `date`: timestamp de build fornecido pelo build.
- `dirty`: estado do working tree fornecido pelo build.

`dev` significa que nenhuma versão de release foi injetada. `unknown` significa que
o build não forneceu aquele campo de metadado.

`go install` puro sem `-ldflags` usa o mesmo metadado de fallback de
desenvolvimento. Um binário instalado construído dessa forma deve imprimir:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Este é o comportamento esperado.

## Convenção de versão

Tags de release do Git usam `vX.Y.Z`, por exemplo `v0.1.0`.

Metadados de versão de release usam `X.Y.Z` simples, por exemplo `0.1.0`. O
GoReleaser injeta o valor simples, então uma release construída a partir da tag
`v0.1.0` exibe:

```text
SpecHarbor 0.1.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

O runtime exibe a string de versão injetada como está e não a normaliza. Se um build
manual injetar `v0.1.0`, `specharbor version` pode exibir `v0.1.0`.

## Injeção em tempo de build

Builds de release injetam metadados usando variáveis Go `-ldflags -X` em:

```text
github.com/guferreira1/spec-harbor/internal/platform/version
```

GoReleaser injeta exatamente:

- `Version={{ .Version }}`
- `Commit={{ .FullCommit }}`
- `Date={{ .Date }}`
- `Dirty={{ .IsGitDirty }}`

O comando runtime não inspeciona tags Git, não lê `.git`, não executa comandos Git,
não executa comandos shell, não chama rede, não escreve arquivos e não normaliza
versões.

## Workflow de release

Mantenedores publicam uma release enviando uma tag que bate com `v*`, por exemplo:

```bash
git tag v0.1.0
git push origin v0.1.0
```

O workflow de release do GitHub Actions roda apenas para tags enviadas que batem com
`v*`. Ele não roda em pushes de branch normais nem em pull requests, então um PR
não pode publicar. O workflow segue menor privilégio: permissão de topo `contents:
read` e apenas jobs que precisam escalam mais.

O workflow executa os jobs em ordem:

1. `validate-release-inputs` — executa `scripts/validate-release-version.sh` para
   rejeitar qualquer tag que não seja exatamente `vX.Y.Z` e falhar se `X.Y.Z` da
   tag não coincidir com a versão em `package.json` do npm.
2. `goreleaser` (`contents: write`) — executa `go test ./...` e
   `release --clean` com `GITHUB_TOKEN` do repositório para criar o GitHub
   Release, fazer upload dos ativos e publicar `checksums.txt`.
3. `npm-publish` (`contents: read`, `id-token: write`, `needs: goreleaser`) —
   revalida a versão, executa testes do pacote npm e validação de conteúdo
   empacotado, depois publica `specharbor@X.Y.Z`.
4. `homebrew-publish` (`contents: read`, `needs: goreleaser`) — renderiza a
   fórmula a partir de `checksums.txt` da release e atualiza o repositório
   `guferreira1/homebrew-tap`.
5. `release-summary` — registra um resumo por canal.

`npm-publish` e `homebrew-publish` esperam `goreleaser` porque ambos consomem os
ativos publicados do GitHub Release e `checksums.txt`. Este workflow baseado em
tag agora publica ativos do GitHub Release, `checksums.txt`, o pacote npm e o tap
Homebrew. Não publica pacotes nativos Linux, manifests de gerenciadores Windows,
assinatura de binários, geração de SBOMs ou imagens Docker.

## Ativos da release

GoReleaser constrói um binário `specharbor` de `./cmd/specharbor` para estes arquivos:

- `specharbor_Linux_x86_64.tar.gz`
- `specharbor_Linux_arm64.tar.gz`
- `specharbor_Darwin_x86_64.tar.gz`
- `specharbor_Darwin_arm64.tar.gz`
- `specharbor_Windows_x86_64.zip`
- `specharbor_Windows_arm64.zip`

Ativos de Linux e macOS usam `.tar.gz`. Ativos de Windows usam `.zip`. O
GoReleaser também gera `checksums.txt` com checksums SHA-256.

As opções de instalação que consomem esses ativos — download manual,
`install.sh`, pacote wrapper npm e tap Homebrew — estão documentadas em
[Instalação](install.md).

Para `v0.1.0`, esses ativos cobrem:

- Linux amd64.
- Linux arm64.
- macOS amd64.
- macOS arm64.
- Windows amd64.
- Windows arm64.
- `checksums.txt`.

## Canais de pacote

O pacote npm `specharbor@X.Y.Z` é publicado automaticamente pelo job
`npm-publish` e mapeia a versão do pacote `X.Y.Z` para a tag GitHub Release
`vX.Y.Z`. Antes de publicar, o job executa `scripts/validate-release-version.sh`,
os testes do pacote npm (`npm test`) e uma validação de conteúdo com
`npm pack --dry-run` que exige `bin/`, `lib/`, `scripts/`, `README.md`,
`README.pt-BR.md` e `package.json` e rejeita `native/`, `node_modules/` e
fixtures de teste. A publicação usa `npm publish --provenance --access public`.

Homebrew está disponível por:

```bash
brew install guferreira1/tap/specharbor
```

A fórmula macOS no repositório externo `guferreira1/homebrew-tap` é atualizada
automaticamente pelo job `homebrew-publish`, que renderiza
`Formula/specharbor.rb` a partir de `checksums.txt` da release com
`scripts/render-homebrew-formula.sh` e faz commit no tap. A fórmula fixa cada
download em SHA-256 e mantém um bloco `test do` executando `specharbor version`.

### Segredos obrigatórios e configuração de publicador confiável

| Segredo/configuração | Propósito |
| --- | --- |
| `GITHUB_TOKEN` (built in) | Cria o GitHub Release e lê seus ativos. |
| npm trusted publisher (recommended) | Configura um publisher confiável para `specharbor` em npmjs.com apontando para `guferreira1/spec-harbor` e `.github/workflows/release.yml`. Com `id-token: write` e npm >= 11.5, o job `npm-publish` publica sem token de longa duração. |
| `NPM_TOKEN` (fallback) | Token granular de automação npm usado como `NODE_AUTH_TOKEN` quando não há publisher confiável configurado. |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Token com acesso de escrita em `guferreira1/homebrew-tap` para que o job `homebrew-publish` possa fazer commit da fórmula. O `GITHUB_TOKEN` padrão não escreve em repositório separado. |

Os tokens são lidos apenas do contexto `secrets` e nunca impressos.

## Checklist de release para mantenedores

1. Atualize `packages/npm/specharbor/package.json` `version` para `X.Y.Z`.
2. Garanta que a tag enviada seja `vX.Y.Z` e corresponda à versão do pacote
   (localmente: `sh scripts/validate-release-version.sh vX.Y.Z`).
3. Configure o publisher confiável npm ou `NPM_TOKEN`.
4. Configure `HOMEBREW_TAP_GITHUB_TOKEN`.
5. Envie a tag: `git tag vX.Y.Z && git push origin vX.Y.Z`.
6. Verifique os ativos do GitHub Release e `checksums.txt`.
7. Verifique o pacote npm: `npm view specharbor@X.Y.Z`.
8. Verifique a fórmula Homebrew atualizada em `guferreira1/homebrew-tap`.
9. Verifique se `install.sh` resolve a nova release.
10. Verifique se `specharbor version` imprime `X.Y.Z`.

## Falhas e rollback

- **GitHub Release funciona, mas npm falha.** Reexecute o job `npm-publish`; ele
  é idempotente até a versão existir no npm. O npm proíbe republicar a mesma
  versão, então nunca tente sobrescrever — se a versão do pacote já foi publicada,
  atualize para novo `X.Y.Z`, retague e publique novamente.
- **npm funciona, mas Homebrew falha.** Reexecute o job `homebrew-publish`; ele
  re-renderiza a fórmula com o `checksums.txt` da release existente e faz push
  apenas se a fórmula mudou, por isso reapresentações são seguras.
- **Versão incompatível.** `validate-release-inputs` falha a release antes de
  publicar; corrija `package.json` ou a tag e publique novamente.
- **Recuperação manual.** A fórmula Homebrew pode ser regenerada localmente com
  `scripts/render-homebrew-formula.sh` usando um `checksums.txt` baixado e
  enviada manualmente ao tap se a CI estiver indisponível.

## Snapshot local e verificação dry-run

As mudanças do workflow de release podem ser revisadas sem enviar uma tag real.
Pull requests nunca publicam porque o workflow é só por tag, e cada etapa tem
equivalente local com dry-run.

Releases locais de snapshot são apenas para verificação. Elas escrevem artefatos
gerados em `dist/`, que é ignorado pelo Git.

Use `goreleaser check` e `goreleaser release --snapshot --clean` antes de
publicar mudanças de release.

Execute:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

Em seguida, execute um binário snapshot gerado:

```bash
./dist/specharbor_linux_amd64_v1/specharbor version
```

As versões snapshot podem incluir metadados de snapshot do GoReleaser em vez de uma
versão de release normal. Elas devem ainda exibir valores injetados de `commit`,
`date` e `dirty` em vez dos valores padrão `unknown` de fallback.

Valide o restante do caminho de publicação sem publicar:

```bash
# Version consistency gate (and its tests).
sh scripts/validate-release-version.sh v0.1.0
sh scripts/test-validate-release-version.sh

# Homebrew formula rendering (and its tests).
sh scripts/test-render-homebrew-formula.sh

# npm package tests and contents.
cd packages/npm/specharbor
npm test
npm pack --dry-run
cd -
```

## Trabalho futuro

GitHub Releases, `checksums.txt`, `install.sh`, o pacote wrapper npm e o tap
externo Homebrew são automatizados e documentados em [Instalação](install.md).
A automação de publicação não implementa estes itens de futuro:

- Pacotes nativos Linux como `.deb`, `.rpm` ou `.apk`.
- Manifests de gerenciadores Windows como Winget, Scoop ou Chocolatey.
- Assinatura, cosign, atestações ou geração de SBOM.
- Imagens Docker ou manifests Docker.

Esses canais e recursos de supply chain permanecem para o futuro e exigem
mudanças OpenSpec separadas ou ação manual de mantenedores.
