# Release 2 — Validação de segurança do Marco G

Data: 17/09/2026

## Escopo

A R2-T15 foi validada com dois usuários reais do Supabase Auth, cada um com
plano, instituição, cartão e premissas próprios. O cenário percorre os handlers
HTTP e os repositórios PostgreSQL usados pela aplicação.

## Casos aprovados

- leitura cruzada de cartão e fatura retorna `404`;
- alteração cruzada de instituição retorna `404`;
- vínculo de item próprio com cartão alheio retorna `404`;
- movimentação para cartão alheio é rejeitada com `409`;
- consultas equivalentes no Data API retornam `[]` pelas policies RLS;
- logs estruturados não contêm tokens, nomes de recursos nem valores
  financeiros usados no cenário.

## Evidência

```text
TEST_DATABASE_URL=<local> \
TEST_SUPABASE_URL=http://127.0.0.1:54321 \
TEST_SUPABASE_PUBLISHABLE_KEY=<local> \
go test ./internal/integration \
  -run TestRelease2EndToEndIsolation -count=1 -v

PASS
```

O teste usa emails únicos por execução e não depende de IDs compartilhados
entre usuários.
