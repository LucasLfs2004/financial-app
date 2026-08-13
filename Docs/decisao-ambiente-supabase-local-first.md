# Decisão técnica — Supabase local-first

> **Status:** aceita  
> **Data:** 30/07/2026  
> **Escopo:** desenvolvimento, autenticação, banco e publicação futura

## Contexto

A organização Supabase inicialmente considerada já possui outros projetos. Um
novo projeto cloud poderia consumir a cota gratuita ou gerar custo adicional.

O Supabase CLI permite executar localmente Postgres, Auth, Data API, Studio,
Storage e os demais serviços necessários ao desenvolvimento.

## Decisão

O desenvolvimento da Financial API será local-first.

Durante as releases iniciais:

- Supabase roda localmente via Docker;
- schema, funções, triggers e políticas são definidos por migrations;
- autenticação e RLS são testadas no stack local;
- a API utiliza as URLs e chaves locais;
- nenhuma release funcional depende de um projeto cloud ativo.

O ambiente cloud será provisionado quando houver necessidade de:

- publicar a API;
- integrar um frontend hospedado;
- executar homologação externa;
- testar infraestrutura próxima da produção.

## Fluxo local

```text
supabase start
       ↓
supabase db reset
       ↓
go test ./...
       ↓
go run ./cmd/api
```

## Publicação futura

Quando o cloud for necessário:

```text
criar projeto cloud
       ↓
supabase link --project-ref <ref>
       ↓
supabase db push
       ↓
supabase config push
       ↓
configurar variáveis da API
       ↓
executar smoke tests
```

As migrations versionadas são a fonte de verdade para o schema.

## O que é transferido

Pelo fluxo de migrations:

- tabelas;
- constraints;
- índices;
- funções;
- triggers;
- políticas RLS;
- demais objetos versionados em SQL.

Configurações suportadas pelo CLI podem ser aplicadas com
`supabase config push`.

## O que não é transferido automaticamente

- usuários locais de teste;
- senhas e sessões;
- dados locais;
- arquivos do Storage;
- secrets;
- credenciais;
- configurações externas de SMTP ou OAuth não versionadas.

Dados reais, se existirem no momento da publicação, exigirão um plano de
migração explícito. Durante o desenvolvimento inicial, dados locais são
descartáveis.

## Consequências

### Positivas

- ausência de custo cloud durante o desenvolvimento;
- ambiente reproduzível;
- testes de migrations desde o zero;
- Auth e RLS disponíveis localmente;
- liberdade para resetar dados de teste;
- publicação posterior baseada nos mesmos artefatos.

### Cuidados

- `supabase db reset` apaga usuários e dados locais;
- `supabase stop` preserva o volume por padrão;
- chaves locais nunca devem ser usadas em produção;
- o stack local não deve ser exposto à internet;
- divergências feitas manualmente no Studio precisam virar migration.

