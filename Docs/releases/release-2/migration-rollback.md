# Release 2 — Migração, rollback e smoke cloud

## Migração

A Release 2 é aditiva e depende da aplicação sequencial das migrations:

```text
20260901000100_release_2_institutions_cards.sql
20260901000200_release_2_payment_methods.sql
20260901000300_release_2_invoice_adjustments.sql
```

Antes do deploy:

1. gerar backup verificável do banco;
2. confirmar que as migrations da Release 1 já foram aplicadas;
3. aplicar as migrations da Release 2 na ordem do repositório;
4. executar lint e smoke de leitura antes de liberar escrita;
5. publicar a API somente depois da confirmação do schema.

As mudanças criam tabelas, enums, índices, constraints, triggers e policies
novos. Não há transformação destrutiva de dados da Release 1 nem backfill
obrigatório: ausência de forma de pagamento continua significando `direct`.

## Rollback

O rollback operacional preferencial é da aplicação:

1. interromper novas escritas da Release 2;
2. redeployar o artefato/tag anterior conhecido;
3. manter as estruturas aditivas no banco, pois a aplicação da Release 1 as
   ignora;
4. investigar e corrigir por migration forward-only.

Não remover enums, tabelas ou eventos de auditoria em um rollback rotineiro.
Caso seja indispensável restaurar o banco, entrar em manutenção e recuperar o
backup anterior às migrations. Eventos append-only e snapshots não devem ser
apagados ou reescritos manualmente.

## Smoke test no Supabase Cloud

O ambiente cloud ainda não está provisionado. Quando estiver disponível:

1. validar secrets e conexão sem registrá-los em logs;
2. aplicar migrations e executar `supabase db lint` no projeto alvo;
3. cadastrar dois usuários temporários;
4. criar plano, instituição, cartão offset `1`/dia `6`, aluguel e gasolina;
5. validar fatura, resumos `reference`/`cash`, ajuste e movimentação;
6. confirmar RLS com leitura cruzada pelo Data API;
7. ativar o plano e verificar snapshot schema v2;
8. revisar logs, métricas e erros;
9. remover os usuários temporários conforme a política do ambiente;
10. registrar commit/tag, horário, executor e resultado do smoke.

Até essa execução, o smoke cloud é o único gate externo pendente e não reduz a
conclusão da validação local da Release 2.
