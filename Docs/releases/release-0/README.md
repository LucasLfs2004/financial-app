# Release 0 — Fundação da API

## Objetivo

Preparar uma base executável, segura e reproduzível para que as releases de
domínio possam começar sem precisar reinventar autenticação, conexão com banco,
migrations ou convenções HTTP.

## Estado

A Release 0 está concluída no modelo local-first.

A fundação local está implementada, validada e pronta para sustentar a Release
1. O provisionamento cloud foi conscientemente adiado até existir necessidade
de homologação externa, frontend hospedado ou publicação.

Decisão relacionada:
[`decisao-ambiente-supabase-local-first.md`](../../decisao-ambiente-supabase-local-first.md).

## Entregas

- projeto Supabase inicializado em `supabase/`;
- Postgres 17 configurado para desenvolvimento local;
- migrations versionadas;
- perfil criado automaticamente para usuários do Supabase Auth;
- RLS de perfil por `auth.uid()`;
- configuração por variáveis de ambiente;
- pool de conexões Postgres;
- validação de access token pelo Supabase Auth;
- identidade extraída do token, nunca de `user_id` enviado pelo cliente;
- endpoint autenticado `GET /v1/me`;
- `GET /health` para vivacidade;
- `GET /ready` com verificação real do banco;
- erros JSON padronizados;
- request ID, logs estruturados, recuperação de panic e headers básicos;
- testes unitários;
- contrato OpenAPI inicial;
- comandos locais no `Makefile`.

## Fronteira de segurança

O identificador do proprietário de qualquer recurso deve vir exclusivamente do
access token validado.

```text
Authorization: Bearer <access-token>
             ↓
Supabase Auth valida o token
             ↓
API obtém o subject do usuário
             ↓
repositórios filtram por owner_id
```

Regras obrigatórias para as próximas releases:

1. handlers não recebem `user_id` como autoridade;
2. consultas por identificador também incluem o proprietário;
3. recursos de outro usuário não são revelados;
4. tabelas expostas pelo Supabase recebem RLS;
5. chaves secretas ou `service_role` nunca são enviadas ao frontend;
6. logs não registram tokens, senhas ou URLs de banco completas.

RLS é defesa adicional. A API continua responsável por aplicar o proprietário
em todas as operações, pois conexões administrativas do backend podem ignorar
RLS.

## Ambiente local

### Dependências

- Go 1.24 ou superior;
- Docker Desktop;
- Supabase CLI;
- `curl` e `jq` para testes manuais.

### Inicialização

```bash
cp .env.example .env
supabase start
supabase status
```

Copie a chave `Publishable` exibida por `supabase status` para:

```text
SUPABASE_PUBLISHABLE_KEY
```

Depois:

```bash
supabase db reset
go test ./...
go run ./cmd/api
```

Endpoints:

```text
GET /health
GET /ready
GET /v1/me
```

`/v1/me` exige um access token emitido pelo Supabase local.

### Parada

```bash
supabase stop
```

O comando preserva o volume local. Não utilizar `--no-backup` sem intenção
explícita de apagar os dados locais.

## Provisionamento cloud futuro

O provisionamento cloud não faz parte do critério de encerramento da Release 0.
Ele permanece documentado para ser executado sem redesenhar o banco.

### Decisões necessárias

Antes de criar o projeto:

- organização Supabase;
- nome do projeto;
- região;
- tamanho ou plano;
- senha forte do banco;
- política de confirmação de e-mail;
- URLs permitidas quando o frontend existir.

Para o uso inicial no Brasil, `sa-east-1` é a região específica de São Paulo.
A escolha deve considerar também onde a API será hospedada.

### Criação

Não colocar a senha no histórico do shell. Preferir prompt interativo ou
variável de ambiente temporária.

```bash
supabase projects create financial-api \
  --org-id <organization-id> \
  --region sa-east-1
```

Depois de obter o project reference:

```bash
supabase link --project-ref <project-ref>
supabase db push
supabase config push
```

Validar o histórico:

```bash
supabase migration list
```

### Variáveis da API cloud

```text
APP_ENV=production
DATABASE_URL=<connection-string>
DATABASE_MAX_CONNECTIONS=<pool-size>
SUPABASE_URL=https://<project-ref>.supabase.co
SUPABASE_PUBLISHABLE_KEY=<publishable-key>
SUPABASE_AUTH_TIMEOUT=5s
```

Para um backend persistente:

- usar conexão direta quando o ambiente alcançar IPv6;
- usar Supavisor em modo session quando o ambiente for somente IPv4;
- não utilizar transaction mode como primeira escolha para um processo
  persistente sem validar as limitações do driver e da aplicação.

As credenciais devem ficar no gerenciador de segredos do ambiente de
hospedagem.

## Critérios de encerramento

- [x] Supabase local inicializa;
- [x] banco pode ser recriado somente pelas migrations;
- [x] cadastro no Auth cria um perfil;
- [x] perfil possui RLS por usuário;
- [x] API falha cedo quando configuração obrigatória está ausente;
- [x] readiness falha quando o banco não está disponível;
- [x] endpoint protegido rejeita token inválido;
- [x] endpoint protegido usa a identidade do token;
- [x] RLS impede leitura cruzada entre usuários;
- [x] testes Go passam;
- [x] fluxo local foi validado ponta a ponta;
- [x] imagem Docker da API compila com Go 1.24;
- [x] estratégia local-first registrada;
- [x] procedimento de publicação cloud documentado.

Itens adiados até a publicação:

- projeto cloud;
- vínculo entre repositório e cloud;
- aplicação das migrations no cloud;
- secrets no ambiente hospedado;
- smoke test externo.

## Evidência da validação local

Em 24/07/2026:

- `supabase db reset` aplicou
  `20260724000100_release_0_profiles.sql`;
- 9 testes Go passaram;
- `GET /ready` retornou `200`;
- token inválido em `GET /v1/me` retornou `401`;
- um usuário do Auth recebeu perfil automaticamente;
- `GET /v1/me` retornou o perfil do subject autenticado;
- um `user_id` arbitrário na query não alterou a identidade consultada.
- dois usuários foram criados e a RLS retornou uma linha própria e zero linhas
  do outro usuário;
- `go vet ./...` não encontrou problemas;
- `docker compose build` concluiu com sucesso.
