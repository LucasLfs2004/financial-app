# Release 2 — Validação do Marco A

**Data:** 2026-09-01  
**Branch:** `feature/r2-milestone-a`  
**Baseline:** `origin/develop` em `c9c0a28`

## Resultado

O Marco A foi validado localmente e está apto para revisão por pull request.

- R2-T01 concluiu o contrato HTTP da Release 2;
- R2-T02 concluiu os tipos de domínio e o projetor puro de fatura;
- a suíte da Release 1 permaneceu compatível;
- nenhuma migration ou alteração de ambiente foi introduzida neste marco.

## Contrato OpenAPI

- 28 paths;
- 36 `operationId` únicos;
- 85 schemas;
- 109 referências internas resolvidas;
- contrato OpenAPI 3.1 validado pelo Redocly CLI;
- exemplos de cartão, forma de pagamento, ajuste, movimentação e composição de
  fatura incluídos.

O lint recomendado do Redocly não encontrou erros. Permaneceram quatro avisos
não bloqueantes já relacionados à fundação do contrato:

- licença ainda não declarada no `info`;
- servidor local apontando para `localhost`;
- `/health` sem resposta 4xx;
- `/ready` sem resposta 4xx.

Esses avisos não alteram a validade do OpenAPI nem o comportamento contratado.

## Domínio

Foram validados:

- enums e parsing estrito;
- dia nominal de vencimento entre 1 e 31;
- offset entre 0 e 12;
- virada de ano no cálculo da fatura;
- dia inexistente sem data inventada;
- seleção temporal de configuração do cartão;
- fallback de pagamento direto;
- compatibilidade entre método e cartão;
- identidade por item e mês de referência;
- movimentação para outra fatura;
- ajustes com referência conhecida ou ausente;
- ordenação determinística;
- duplicidade e overflow.

## Evidências executadas

```text
go vet ./...                         sem problemas
go test ./...                        aprovado
go test -race ./...                  aprovado
go test ./internal/cardinvoice       aprovado
supabase db reset --local --yes       aprovado
supabase test db supabase/tests       70 testes aprovados
supabase db lint --local              sem erros
testes integrados com Postgres local  aprovados
teste local com Auth e Data API       aprovado
```

## Compatibilidade e riscos

- itens existentes continuam com pagamento direto quando não possuem período
  explícito;
- nenhuma regra da Release 1 foi removida;
- `reference_month` passa a aceitar nulo somente para ajuste sem competência;
- o breakdown ganha `card_invoice_adjustments_cents`;
- os handlers e a persistência da Release 2 ainda não existem e serão
  implementados nos próximos marcos;
- as alterações locais sobre Open Finance permaneceram fora dos commits do
  Marco A.
