# Release 2.1 — Relatório de qualidade

## Resultado

A Release 2.1 foi validada localmente em 23/09/2026.

O planejamento agora funciona como janela de projeção. Itens, períodos,
métodos de pagamento, ajustes e movimentações pertencem ao usuário e podem
ultrapassar ou sobreviver ao planejamento atual. Economia planejada e snapshots
continuam vinculados ao plano.

## Evidências

- reset completo do banco aplicou todas as migrations desde zero;
- pgTAP: 6 arquivos e 129 testes aprovados;
- `go fmt ./...`: aprovado;
- `go vet ./...`: aprovado;
- `go test ./... -count=1`: aprovado, incluindo integração local;
- `go test -race ./... -count=1`: aprovado;
- `supabase db lint --local --level warning`: nenhum erro;
- OpenAPI 3.1 válido no Redocly CLI, com quatro warnings preexistentes de
  governança geral do contrato;
- `git diff --check`: aprovado.

## Cenário de aceite

O teste `TestRelease21RecordsOutlivePlanHorizon` comprova:

1. plano de setembro a dezembro de 2026;
2. salário e seguro válidos até setembro de 2027;
3. resumo de dezembro usando apenas a janela do primeiro plano;
4. método de pagamento válido até setembro de 2027;
5. snapshot v3 contendo o período completo do seguro;
6. arquivamento do primeiro plano;
7. novo planejamento em 2027 reutilizando os mesmos registros;
8. resumo de setembro de 2027 sem recadastro ou cópia.

## Compatibilidade

- rotas canônicas: `/v1/financial-items`;
- aliases das Releases 1 e 2 sob `/v1/plans/current/items` continuam ativos;
- `plan_id` foi removido das respostas de registros;
- `currency_code` passou a ser explícito em itens e ajustes;
- snapshots v1 e v2 permanecem legíveis; novas ativações usam schema v3;
- a Release 3 foi atualizada para dívidas user-scoped e snapshot v4.

## Riscos residuais

- rollback do schema exige backup para ser lossless;
- clientes que validam rigidamente `FinancialItem.plan_id` precisam migrar para
  `currency_code`;
- conversão cambial e soma de moedas diferentes permanecem fora do escopo;
- deploy e smoke no Supabase Cloud ainda dependem do ambiente remoto.

