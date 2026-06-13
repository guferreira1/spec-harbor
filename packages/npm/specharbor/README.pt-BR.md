# specharbor

O pacote `specharbor` é a distribuição do npm do CLI SpecHarbor.

Consulte a documentação principal em inglês: [README.md](README.md)

Este pacote instala e executa o binário oficial do SpecHarbor para a sua plataforma.
Ele não é apenas um lançador de comandos: ele disponibiliza o ponto de entrada do
fluxo OpenSpec/SDD usado no repositório:

- estruturação explícita de mudanças;
- validação local;
- prompts por função;
- revisão e arquivamento da mudança;
- fluxos do CLI compatíveis com trabalho com agentes de IA.

## Instalação

```bash
npm install -g specharbor
specharbor version
```

Também é possível executar sem instalação global:

```bash
npx specharbor version
```

## O que o SpecHarbor faz

SpecHarbor é uma CLI em Go para fluxos de trabalho com agentes de programação assistidos por IA baseados em OpenSpec.
Ela ajuda a transformar uma ideia em uma mudança com escopo explícito e trilha de contexto: OpenSpec, validação, implementação, revisão e arquivamento.

## Comandos comuns

- `specharbor init`
- `specharbor workflow`
- `specharbor generate ...`
- `specharbor validate ...`
- `specharbor prompt ...`
- `specharbor review ...`
- `specharbor archive ...`
- `specharbor brief`
- `specharbor context discover`
- `specharbor context index`

## Como o wrapper npm funciona

Este pacote é um launcher Node pequeno para o binário nativo.

- A versão do pacote `X.Y.Z` mapeia para exatamente uma tag de release do GitHub `vX.Y.Z`.
- No `postinstall` (ou na primeira execução quando os scripts são pulados), baixa o binário oficial da versão correspondente dos lançamentos do GitHub.
- A integridade do arquivo é validada por SHA-256 contra `checksums.txt` antes da extração.
- O binário extraído é armazenado no diretório `native/` do pacote.
- Os argumentos da CLI são repassados como array para o processo do binário com `stdio` herdado.
- O código de saída do binário é preservado.
- Nenhum comando é construído por string de shell para repasse de argumentos.

## Plataformas suportadas

| `process.platform` | `process.arch` | Asset de release |
| --- | --- | --- |
| `linux` | `x64` | `specharbor_Linux_x86_64.tar.gz` |
| `linux` | `arm64` | `specharbor_Linux_arm64.tar.gz` |
| `darwin` | `x64` | `specharbor_Darwin_x86_64.tar.gz` |
| `darwin` | `arm64` | `specharbor_Darwin_arm64.tar.gz` |
| `win32` | `x64` | `specharbor_Windows_x86_64.zip` |
| `win32` | `arm64` | `specharbor_Windows_arm64.zip` |

Plataformas não suportadas retornam erro determinístico.

## Mapeamento de versão

| Versão do npm | Tag de release no GitHub |
| --- | --- |
| `0.1.0` | `v0.1.0` |

O mapeamento é de um para um pela versão do pacote.

## Solução de problemas

- **scripts de instalação ignorados**
  `npm install --ignore-scripts -g specharbor` ignora o postinstall por design.
  A primeira execução (`specharbor version` ou `npx specharbor version`) mantém o mesmo fluxo com download e verificação de checksum.

- **falha offline/proxy/acesso ao GitHub**
  Falhas na primeira execução podem ocorrer com acesso de rede restrito.
  Refaça em ambiente com acesso ao GitHub ou use uma opção manual em [docs/install.md](https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md).

- **plataforma ou arquitetura não suportada**
  Apenas Linux, macOS e Windows em `x64` e `arm64` são suportados.
  A mensagem de erro aponta as opções manuais.

- **falha de verificação de checksum**
  O pacote recusa instalação/execução com mismatch de checksum.
  Repita a instalação e, se o erro persistir, abra uma issue com a versão do pacote e o log do erro.

- **verificar versão**

```bash
specharbor version
```

O release oficial imprime metadados injetados. Binários construídos sem metadados de release podem imprimir `dev`/`unknown`.

## Modelo de segurança

- URLs de release apenas por HTTPS em
  `https://github.com/guferreira1/spec-harbor/releases/download/`
- Verificação SHA-256 obrigatória usando `checksums.txt`.
- Sem fluxo de autenticação por token no runtime.
- Nenhuma execução de shell para repassar argumentos.
- Nenhum caminho gravável fora do diretório do pacote (`native/`) e diretórios temporários necessários.

## Documentação relacionada

- [Instalação e verificação](https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md)
- [README principal do SpecHarbor](https://github.com/guferreira1/spec-harbor/blob/main/README.md)
- [Metadados de release](https://github.com/guferreira1/spec-harbor/blob/main/docs/release.md)

## Testes

```bash
npm test
```

Os testes usam fixtures locais e não publicam nada.
