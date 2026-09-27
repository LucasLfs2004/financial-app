# Deploy da API no Render

## Arquitetura dos ambientes

| Ambiente | API | Banco e Auth | Dados |
| --- | --- | --- | --- |
| desenvolvimento | máquina local | Supabase local | descartáveis |
| produção | Render Free em Oregon | Supabase Cloud em Oregon | persistentes |

O Render executa apenas a API. O filesystem do contêiner é efêmero e não deve
armazenar dados da aplicação.

## Criação do serviço

1. Promova para `main` o commit validado em `develop` pelo fluxo de release.
2. No Render, crie um Blueprint a partir deste repositório.
3. Autorize a leitura do repositório e selecione o arquivo `render.yaml`.
4. Preencha os três valores marcados como secretos.
5. Confirme a criação do serviço `financial-api`.

O Blueprint configura Docker, plano gratuito, região Oregon, branch `main` e
health check em `/ready`. Cada commit novo em `main` inicia um deploy.

## Secrets de produção

Cadastre os valores somente no painel do Render. Não grave esses valores em
arquivos versionados, logs, tickets ou documentação.

| Variável | Origem |
| --- | --- |
| `DATABASE_URL` | Supabase > Connect > Session pooler |
| `SUPABASE_URL` | Supabase > Settings > API |
| `SUPABASE_PUBLISHABLE_KEY` | Supabase > Settings > API > Publishable key |

Use a connection string do **Session pooler**, adequada para um serviço de API
persistente com saída IPv4. A URL precisa exigir TLS (`sslmode=require`). Não
use a chave `service_role`: a API valida o token do usuário com a chave pública
e acessa o Postgres por uma credencial própria.

As demais variáveis são definidas pelo Blueprint:

```text
APP_ENV=production
PORT=10000
DATABASE_MAX_CONNECTIONS=3
SUPABASE_AUTH_TIMEOUT=5s
```

## Publicação de schema

O schema cloud é atualizado exclusivamente por migrations versionadas:

```bash
supabase db push --linked --dry-run
supabase db push --linked
supabase migration list --linked
```

Nunca execute `supabase db reset --linked`. Resets pertencem somente ao banco
local descartável. Testes de integração que criam ou removem registros também
devem usar apenas o Supabase local.

## Smoke test

Depois do primeiro deploy, substitua `<render-url>` pelo endereço entregue pelo
Render:

```bash
curl --fail --show-error https://<render-url>/health
curl --fail --show-error https://<render-url>/ready
curl --fail --show-error https://<render-url>/
```

Resultados esperados:

- `/health`: HTTP 200 com `{"status":"ok"}`;
- `/ready`: HTTP 200 com `{"status":"ready"}`;
- `/`: HTTP 200 identificando `financial-api`.

No plano gratuito, a API é suspensa após um período sem tráfego. A primeira
requisição depois disso pode demorar enquanto o contêiner inicia.

## Rollback

Para falha somente na API, use o rollback do Render para um dos deploys
anteriores. Para migrations, siga a nota de rollback da release; não reverta o
banco com reset e não altere migrations já aplicadas.
