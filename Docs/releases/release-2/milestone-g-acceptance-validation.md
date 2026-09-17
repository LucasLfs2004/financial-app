# Release 2 — Validação de aceite do Marco G

Data: 17/09/2026

## Cenário

O teste `TestRelease2AcceptanceScenario` executa a jornada completa por HTTP
com um usuário real do Supabase Auth:

1. cria plano, renda, aluguel, gasolina e poupança explícita;
2. cria instituição e cartões;
3. configura offset `1` e vencimento nominal no dia `6`;
4. altera o cartão da gasolina a partir de agosto sem reescrever julho;
5. projeta dezembro com aluguel de R$ 1.000, gasolina de R$ 700 e ajuste de
   R$ 180;
6. compara novembro em `reference` e dezembro em `cash`;
7. soma as sources de compromisso para provar ausência de dupla contagem;
8. move gasolina de novembro para janeiro e valida o histórico append-only;
9. consulta a fatura de transbordo e rejeita o resumo fora do horizonte;
10. valida vencimento nominal no dia 31 em fevereiro;
11. ativa o plano e inspeciona o snapshot v2 com cartões alcançáveis, ajuste e
    evento de movimentação.

## Resultado

```text
TEST_DATABASE_URL=<local> \
TEST_SUPABASE_URL=http://127.0.0.1:54321 \
TEST_SUPABASE_PUBLISHABLE_KEY=<local> \
go test ./internal/integration \
  -run TestRelease2AcceptanceScenario -count=1 -v

PASS
```

A fatura de dezembro totalizou `188000` centavos e três componentes. O resumo
`cash` apresentou o mesmo total de compromissos, sem source agregada de fatura.
Após a movimentação, dezembro passou a `118000` centavos e janeiro recebeu a
ocorrência movida preservando origem, destino, motivo e instante.
