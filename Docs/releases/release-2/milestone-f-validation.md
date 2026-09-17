# Validação do Marco F — integração financeira

Data: 17/09/2026

## Escopo concluído

- resumo mensal com componentes de cartão nas bases `cash` e `reference`;
- exclusão do caminho direto em caixa para impedir dupla contagem;
- ajustes com e sem competência no bucket apropriado;
- metadados de cartão, fatura e origem de alocação nas sources;
- cálculos monetários centralizados no value object `Money`;
- fotografia original v2 determinística e limitada a recursos alcançáveis;
- leitura preservada para snapshots v1 e imutabilidade mantida.

## Evidências automatizadas

```text
go test ./internal/monthlysummary ./internal/cardinvoice/... ./internal/platform/httpserver ./cmd/api
PASS

TEST_DATABASE_URL=<local> go test ./internal/integration \
  -run TestSnapshotV2IsReachableOrderedImmutableAndV1RemainsReadable -count=1 -v
PASS
```

Os testes do resumo cobrem competência, caixa, ajuste sem referência,
movimentação para outra fatura e ausência de dupla contagem. O teste de
integração do snapshot valida alcance, arrays vazios, ordenação, reativação
idempotente, bloqueio de atualização e leitura de schema v1.

## Estratégia de evolução

`plan_snapshots.schema_version` é o discriminador canônico. Novas ativações
gravam versão 2; snapshots existentes não são migrados nem reescritos. A
consulta de leitura permanece neutra à versão e retorna o JSON persistido junto
do discriminador, deixando a interpretação para o consumidor apropriado.

O reset integral não fazia parte da execução original do Marco F. Ele foi
posteriormente autorizado e aprovado no gate final da R2-T17, junto da suíte
SQL completa e do lint do banco.
