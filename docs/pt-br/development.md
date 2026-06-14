# Desenvolvimento

O SpecHarbor é uma CLI em Go. A versão do Go do repositório está declarada em `go.mod`:

```text
go 1.23
```

Use Go 1.23 ou uma toolchain Go compatível e mais nova.

## Comandos locais

Execute testes:

```bash
go test ./...
```

Execute testes sem cache quando quiser o mesmo modo do CI:

```bash
go test -count=1 ./...
```

Construa a CLI:

```bash
go build ./cmd/specharbor
```

Rode a CLI durante o desenvolvimento:

```bash
go run ./cmd/specharbor help
```

Verifique formatação:

```bash
find . -name '*.go' -print0 | xargs -0 gofmt -l
```

Formate arquivos Go alterados com `gofmt` antes de finalizar mudanças.

## CI

O CI do SpecHarbor é específico para Go porque o SpecHarbor é implementado em Go.

O GitHub Actions roda em pull requests e pushes para `main`. O fluxo:

- usa o módulo Go do repositório via `go.mod`;
- verifica `gofmt`;
- executa `go test -count=1 ./...`.

Isso não torna `specharbor scan` específico de Go. `specharbor scan` foi feito para ser stack-agnostic e não deve assumir que projetos de usuário são Go.

## Mudanças somente de documentação

Para OpenSpecs apenas de documentação, mantenha o diff em arquivos Markdown e em `tasks.md` da mudança ativa. Não modifique código Go, testes Go, `.github/workflows/ci.yml`, comportamento da CLI ou templates de init a menos que a change ativa exija explicitamente.
