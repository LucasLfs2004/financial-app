# Releases da Financial API

Este diretório organiza as entregas incrementais da API.

## Sequência

1. [Release 0 — Fundação da API](./release-0/README.md)
2. [Release 1 — Quanto está livre?](./release-1/spec.md)
3. [Release 2 — Por que minha fatura é esse valor?](./release-2/spec.md) —
   concluída e validada localmente em 17/09/2026;
4. Release 3 — Quando essa dívida termina?
5. Release 4 — Como ficam os próximos meses?
6. Release 5 — Quanto realmente aconteceu?
7. Release 6 — Como minha vida financeira mudou?
8. Release 7 — Quanto terei acumulado?
9. Release 8 — E se alguma coisa mudar?

Cada release funcional deve seguir:

```text
spec → design → tasks → implementação → validação
```

As regras gerais continuam tendo como fonte de verdade:

- [`documentacao-produto-planejador-financeiro.md`](../documentacao-produto-planejador-financeiro.md);
- decisões específicas registradas em `Docs/`;
- a especificação da release em andamento.

As evidências finais da Release 2 estão no
[`quality-report.md`](./release-2/quality-report.md). O deploy e smoke test em
Supabase Cloud permanecem condicionados ao provisionamento do ambiente.
