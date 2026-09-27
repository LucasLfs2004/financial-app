# Release 3 — Migration e rollback

## Ordem de implantação

Aplicar as migrations R3 na ordem dos timestamps:

1. `20260926000100_release_3_debt_item_kind.sql`: novo kind;
2. `20260926000200_release_3_debts.sql`: metadados, integridade e RLS;
3. `20260926000300_release_3_debt_early_settlements.sql`: quitações imutáveis;
4. `20260926000400_release_3_debt_payment_methods.sql`: métodos de pagamento.

As alterações são aditivas. O reset local a partir do banco vazio aplicou toda
a sequência R0–R3 e passou nos testes pgTAP e no lint do schema.

## Compatibilidade e rollback

O código R3 lê snapshots v1/v2/v3 e cria v4 apenas em novas ativações. O
snapshot original é imutável; não há atualização retroativa dos documentos.

Antes de voltar a aplicação para R2.1, bloquear novas escritas de dívidas e
quitação. A versão anterior não entende `debt_installment`; portanto, o
rollback de aplicação exige avaliar os usuários que já criaram dívidas, testar
consultas legadas com esses dados e impedir que a versão antiga processe seus
itens. Não apagar o kind ou as tabelas em produção: enum PostgreSQL e dados
persistidos não têm reversão segura automática. Qualquer remoção requer
migration própria, aprovada e com plano de preservação de dados.

O rollback operacional preferido após uma falha de implantação é restaurar o
último artefato R3 saudável, mantendo o schema aditivo, ou corrigir a falha em
uma nova versão. Não reutilizar nem mover uma tag publicada.

## Smoke em cloud

Após aplicar as migrations R3 no Supabase Cloud e publicar a API, criar
usuário de teste, dívida direta e dívida em cartão; conferir cronograma,
resumo, fatura, quitação, RLS e snapshot v4. Registrar a tag e o artefato
efetivamente promovidos antes de considerar a release publicada.
