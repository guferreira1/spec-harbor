# Desenvolvimento

# Desenvolvimento

O SpecHarbor é uma CLI em Go. A versão do Go do repositório está declarada em `go.mod`:

```text
go 1.23
```

Use Go 1.23 or a compatible newer Go toolchain.

## Local Commands

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

O CI do SpecHarbor é específico para Go porque o SpecHarbor é escrito em Go.

GitHub Actions runs on pull requests and pushes to `main`. The workflow:

- uses the repository Go module through `go.mod`;
- checks `gofmt`;
- runs `go test -count=1 ./...`.

Isso não torna `specharbor scan` específico de Go. `specharbor scan` foi feito para ser stack-agnostic e não deve assumir que projetos de usuário são Go.

## Mudanças somente de documentação

Para OpenSpecs apenas de documentação, mantenha o diff em arquivos Markdown e em `tasks.md` da mudança ativa. Não modifique código Go, testes Go, `.github/workflows/ci.yml`, comportamento da CLI ou templates de init a menos que a change ativa exija explicitamente.
