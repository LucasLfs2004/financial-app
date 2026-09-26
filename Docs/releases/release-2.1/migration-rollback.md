# Release 2.1 — Migração e rollback

## Migração

A migration `20260923000100_release_2_1_user_owned_financial_records.sql`:

1. copia a moeda do plano de origem para itens e ajustes;
2. cria FKs de ownership diretamente por `user_id`;
3. substitui relações compostas que dependiam de `plan_id`;
4. remove triggers de contenção no horizonte;
5. recria a validação de métodos de pagamento sem consultar planos;
6. remove `plan_id` das tabelas operacionais;
7. cria índices por usuário, moeda e intervalo.

As tabelas `plans`, `saving_periods` e `plan_snapshots` continuam plan-scoped.

## Compatibilidade de deploy

As escritas financeiras devem passar pela versão nova da API assim que a
migration for aplicada. A versão anterior tenta escrever colunas removidas e
não é compatível com o novo schema.

Sequência recomendada:

1. gerar backup/point-in-time recovery;
2. pausar writers antigos;
3. aplicar a migration;
4. publicar a API nova;
5. executar smoke de criação, resumo, fatura e snapshot;
6. reabrir writers.

## Rollback

O rollback lossless exige restauração do backup anterior à migration. Depois
que um usuário cria mais de um planejamento usando os mesmos registros, não
existe uma associação única capaz de reconstruir o antigo `plan_id`.

Em incidente antes de novas escritas:

1. interromper a API;
2. restaurar o backup;
3. publicar a versão anterior;
4. validar contagens e snapshots;
5. reabrir tráfego.

Em incidente depois de novas escritas, preferir roll-forward. Reintroduzir
`plan_id` escolhendo arbitrariamente um plano perderia a semântica contínua e
não é um rollback seguro.

## Smoke mínimo

- criar plano `2026-09..2026-12`;
- criar seguro `2026-09..2027-09`;
- confirmar resumo de dezembro;
- configurar pagamento até setembro de 2027;
- ativar e verificar snapshot v3;
- arquivar o plano, criar outro em 2027 e confirmar o mesmo seguro;
- tentar acesso cruzado com outro usuário.

