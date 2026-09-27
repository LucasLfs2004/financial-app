# Validação do Marco C — Cadastro e cronograma

Data: 26/09/2026

## Resultado

O Marco C da Release 3 foi concluído. A API agora cadastra, consulta, altera,
arquiva e projeta dívidas parceladas por meio do módulo vertical `debt`, sem
materializar ocorrências mensais no banco.

## Entregas

### R3-T05 — Cadastro e consulta

- criação transacional de `financial_items`, `financial_item_periods` e
  `debts`;
- moeda obtida do perfil do proprietário;
- `as_of` explícito ou derivado do timezone do perfil;
- listagem por status persistido e projetado, ordenada por término efetivo;
- detalhe com término, liberação e parcelas restantes;
- edição de nome/descrição e arquivamento sem apagar cronograma;
- criação genérica de `debt_installment` rejeitada;
- carregamento agregado de períodos e quitação, sem N+1;
- ownership divergente tratado como recurso inexistente.

### R3-T06 — Mudança de valor

- endpoint especializado `POST /v1/debts/{debt_id}/changes`;
- dívida e período aplicável bloqueados durante a mudança;
- período anterior fechado e projeções posteriores substituídas atomicamente;
- offset de caixa preservado;
- término e numeração estrutural mantidos;
- competências fora do cronograma ou posteriores à quitação rejeitadas;
- primeira, intermediária e última parcela cobertas por integração real.

### R3-T07 — Cronograma projetado

- endpoint `GET /v1/debts/{debt_id}/schedule` com intervalo inclusivo;
- limite de 120 meses e recorte pelo término efetivo;
- ocorrências ordenadas com parcela, total, valor, origem e completude;
- fallback direto com `cash_month_offset` do período financeiro;
- resolução temporal de cartão e mês da fatura preparada e testada;
- quitação substitui sua competência e elimina ocorrências posteriores;
- respostas não persistem nem interpretam parcelas como pagamentos realizados.

## Evidências automatizadas

```text
supabase db reset --local --yes             aprovado desde banco vazio
supabase test db supabase/tests --local     167 testes aprovados em 8 arquivos
supabase db lint --local --level warning    nenhum erro de schema
go test ./...                               todos os pacotes aprovados
go test -race ./...                         todos os pacotes aprovados
go vet ./...                                nenhum problema encontrado
```

Os testes Go foram executados com `TEST_DATABASE_URL` apontando para o
PostgreSQL local, portanto o ciclo de cadastro, mudança e cronograma do Marco C
foi exercitado contra as migrations reais. Nenhum ambiente Supabase Cloud foi
alterado.

## Segurança e integridade verificadas

- criação grava item, período e dívida na mesma transação;
- consultas e mutações sempre filtram `user_id`;
- identificador de outro usuário retorna `not found`;
- dívida arquivada não aceita nova edição ou mudança;
- alterações temporais não criam lacuna nem sobreposição;
- mudança de valor preserva parcelas anteriores e término;
- intervalos inválidos e cronogramas acima de 120 meses são rejeitados;
- cronograma mantém uma única ocorrência lógica por dívida e competência.

## Observação de integração

As branches foram empilhadas sobre o Marco B na ordem T05, T06 e T07. A
resolução de cartão já existe no projetor e possui cobertura unitária; a
persistência de métodos de pagamento para itens `debt_installment` continua
deliberadamente bloqueada até a R3-T08, responsável por ampliar a validação do
banco e concluir a integração de escrita.
