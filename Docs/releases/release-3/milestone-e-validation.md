# Release 3 — Validação do Marco E

## Escopo

O Marco E conclui a integração financeira das dívidas nas tasks R3-T11 e
R3-T12:

- parcelas e quitações antecipadas nas faturas de cartão;
- parcelas diretas e de cartão no resumo mensal;
- reconciliação do novo bucket de dívidas com totais e sources.

## Evidências funcionais

### R3-T11 — Faturas

- o carregamento `repeatable read` lê dívidas, períodos e quitações em lote,
  sem N+1;
- o projetor puro de dívidas continua responsável por numeração, valor temporal
  e corte após quitação;
- `cardinvoice` resolve método, cartão, mês de pagamento e movimentação;
- parcelas e quitações viram componentes com identidade da competência e
  metadados da dívida;
- troca temporal de cartão, movimentação manual e transbordo de fatura não
  duplicam componentes.

### R3-T12 — Resumo mensal

- a base `reference` recebe ocorrências pela competência;
- a base `cash` recebe dívidas diretas pelo offset do período e dívidas de
  cartão exclusivamente pelos componentes da fatura;
- a quitação substitutiva participa das duas bases sem recriar parcelas
  posteriores;
- `debt_installments_cents` integra commitments e é reconciliado com os
  sources;
- os sources transportam dívida, número da parcela, total, tipo da ocorrência,
  método e metadados de cartão quando aplicáveis;
- valores liberados permanecem informativos e não se tornam renda ou source.

## Validações executadas

- `gofmt` nos arquivos alterados;
- `go vet ./...` sem ocorrências;
- suíte Go completa com 186 testes em 41 pacotes;
- suíte Go com race detector: 186 testes em 41 pacotes;
- 12 testes de integração reais após reconstrução do banco;
- reset completo do banco a partir de todas as migrations;
- 170 assertions pgTAP em 9 arquivos;
- lint do schema Supabase sem erros.

## Compatibilidade

- itens e resumos das Releases 1 e 2 mantêm o comportamento anterior;
- componentes de cartão que não pertencem a dívidas continuam reconciliados
  pela identidade item/competência;
- o contrato HTTP preserva campos existentes e acrescenta apenas o bucket e
  metadados de dívida previstos no OpenAPI;
- não há migration nova no Marco E.
