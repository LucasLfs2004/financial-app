# Financial API

API de planejamento financeiro pessoal em Go e Supabase.

## Estado

- Release 0: concluída no modelo local-first;
- Release 1: em implementação; contrato, domínio base e persistência de
  planejamento, itens e vigências concluídos.

Documentos:

- [visão do produto](./Docs/documentacao-produto-planejador-financeiro.md);
- [releases](./Docs/releases/README.md);
- [Release 0](./Docs/releases/release-0/README.md);
- [decisão Supabase local-first](./Docs/decisao-ambiente-supabase-local-first.md);
- [Release 1 — spec](./Docs/releases/release-1/spec.md);
- [Release 1 — design](./Docs/releases/release-1/design.md);
- [Release 1 — tasks](./Docs/releases/release-1/tasks.md).

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
```

`/v1/me` exige:

```text
Authorization: Bearer <supabase-access-token>
```

## Testes

```bash
go test ./...
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
