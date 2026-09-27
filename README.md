# Financial API

API de planejamento financeiro pessoal em Go e Supabase.

## Estado

- Release 0: fundação da API concluída;
- Release 1: planejamento, premissas, resumo mensal e snapshot original;
- Release 2: concluída localmente, com instituições, cartões, formas de
  pagamento temporais, faturas projetadas, ajustes, movimentações auditáveis,
  resumo integrado e snapshot v2;
- Release 2.1: registros financeiros contínuos e snapshot v3;
- Release 3: dívidas projetadas, parcelas, quitação antecipada, integração com
  cartão e resumo mensal, além de snapshot v4. Validação local concluída.

O smoke test em Supabase Cloud permanece pendente até o provisionamento do
ambiente. A validação local da Release 3 está em
[`Docs/releases/release-3/quality-report.md`](Docs/releases/release-3/quality-report.md).

## Estrutura

- `cmd/api`: ponto de entrada da API;
- `internal`: módulos de domínio, aplicação, persistência e transporte;
- `internal/platform`: autenticação, configuração, banco e servidor HTTP;
- `supabase`: migrations, seed e testes pgTAP;
- `api/openapi.yaml`: contrato HTTP OpenAPI 3.1;
- `postman`: collections e ambiente de validação manual;
- `Docs/releases`: spec, design, tasks e evidências por release.

## Ambiente local

Requisitos: Go 1.25+ com o toolchain indicado em `go.mod`, Docker Desktop e
Supabase CLI.

```bash
cp .env.example .env
supabase start
supabase db reset --local --yes
```

Atualize `SUPABASE_PUBLISHABLE_KEY` no `.env` com a chave exibida por
`supabase status`. Depois carregue o ambiente e inicie a API:

```bash
set -a
. ./.env
set +a
go run ./cmd/api
```

Endpoints de saúde:

```text
GET http://localhost:8080/health
GET http://localhost:8080/ready
```

Os endpoints `/v1` protegidos exigem `Authorization: Bearer <access-token>`.
Consulte [`api/openapi.yaml`](api/openapi.yaml) para o contrato completo.

## Validação

```bash
go fmt ./...
go vet ./...
go test ./...
go test -race ./...
supabase test db supabase/tests --local
supabase db lint --local --level warning
```

A validação manual da Release 3 usa:

- [`Financial API - Release 3.postman_collection.json`](postman/Financial%20API%20-%20Release%203.postman_collection.json);
- [`Financial API - Local.postman_environment.json`](postman/Financial%20API%20-%20Local.postman_environment.json);
- [`Docs/postman-validation.md`](Docs/postman-validation.md).

## Releases

- [mapa de releases](Docs/releases/README.md);
- [Release 2 — especificação](Docs/releases/release-2/spec.md);
- [Release 2 — design](Docs/releases/release-2/design.md);
- [Release 2 — tasks](Docs/releases/release-2/tasks.md);
- [Release 3 — spec](Docs/releases/release-3/spec.md);
- [Release 3 — design](Docs/releases/release-3/design.md);
- [Release 3 — tasks](Docs/releases/release-3/tasks.md);
- [fluxo Git e releases](Docs/git-release-flow.md).
