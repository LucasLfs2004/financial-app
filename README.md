# Financial API

API de planejamento financeiro pessoal em Go e Supabase.

## Estado

- Release 0: concluída no modelo local-first;
- Release 1: em implementação; Marcos A a E concluídos, com contrato, domínio,
  persistência, planejamento, premissas, ativação e resumo mensal disponíveis.

Documentos:

- [visão do produto](./Docs/documentacao-produto-planejador-financeiro.md);
- [releases](./Docs/releases/README.md);
- [Release 0](./Docs/releases/release-0/README.md);
- [decisão Supabase local-first](./Docs/decisao-ambiente-supabase-local-first.md);
- [Release 1 — spec](./Docs/releases/release-1/spec.md);
- [Release 1 — design](./Docs/releases/release-1/design.md);
- [Release 1 — tasks](./Docs/releases/release-1/tasks.md);
- [fluxo Git e releases](./Docs/git-release-flow.md).

## Requisitos

- Go 1.24+;
- Docker Desktop;
- Supabase CLI.

## Ambiente local

```bash
cp .env.example .env
supabase start
supabase status
```

Atualize `SUPABASE_PUBLISHABLE_KEY` em `.env` com a chave local exibida pelo
CLI. Depois:

```bash
supabase db reset
set -a
source .env
set +a
go run ./cmd/api
```

Endpoints:

```text
GET http://localhost:8080/
GET http://localhost:8080/health
GET http://localhost:8080/ready
GET http://localhost:8080/v1/me
POST http://localhost:8080/v1/plans
GET http://localhost:8080/v1/plans/current
PATCH http://localhost:8080/v1/plans/current
POST http://localhost:8080/v1/plans/current/activate
GET http://localhost:8080/v1/plans/current/original
POST http://localhost:8080/v1/plans/current/items
GET http://localhost:8080/v1/plans/current/items
GET http://localhost:8080/v1/plans/current/items/{item_id}
PATCH http://localhost:8080/v1/plans/current/items/{item_id}
POST http://localhost:8080/v1/plans/current/items/{item_id}/changes
POST http://localhost:8080/v1/plans/current/items/{item_id}/archive
GET http://localhost:8080/v1/plans/current/savings
PUT http://localhost:8080/v1/plans/current/savings
GET http://localhost:8080/v1/plans/current/months/{month}/summary?basis=cash
```

`/v1/me` exige:

```text
Authorization: Bearer <supabase-access-token>
```

## Testes

```bash
go test ./...
```

O cenário integrado dos Marcos C a E usa o banco local:

```bash
TEST_DATABASE_URL='postgresql://postgres:postgres@127.0.0.1:54322/postgres?sslmode=disable' \
  go test -v ./internal/integration
```

Exemplos de resumo para julho de 2026:

```bash
curl -H 'Authorization: Bearer <supabase-access-token>' \
  'http://localhost:8080/v1/plans/current/months/2026-07/summary?basis=cash'

curl -H 'Authorization: Bearer <supabase-access-token>' \
  'http://localhost:8080/v1/plans/current/months/2026-07/summary?basis=reference'
```

Para recriar o banco:

```bash
supabase db reset
```

Para validar migrations e regras do banco:

```bash
supabase test db supabase/tests --local
supabase db lint --local --level warning
```

## Docker da API

O Supabase local deve estar ativo no host.

```bash
docker compose up --build
```

O compose utiliza `host.docker.internal` para alcançar o Supabase local.

## Estrutura

- `api`: contrato OpenAPI;
- `cmd/api`: ponto de entrada;
- `internal/platform`: infraestrutura compartilhada;
- `internal/profile`: perfil associado ao Supabase Auth;
- `supabase`: configuração, migrations e seed;
- `Docs/releases`: spec, design e tasks das releases.
