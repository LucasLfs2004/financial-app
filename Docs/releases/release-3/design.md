# Release 3 — Design técnico — Dívidas e parcelas projetadas

## 1. Direção arquitetural

A API permanece um monólito modular em Go. A release introduz um módulo
vertical de dívidas e amplia os projetores existentes.

```text
internal/
  debt/
    domain/
    application/
    repository/
    transport/http/
  financialitem/
  paymentmethod/
  cardinvoice/
  monthlysummary/
  planning/
```

`debt` responde por cadastro, cronograma, término, liberação e quitação. Ele
reutiliza `Money`, `YearMonth` e `MonthInterval`. Seu projetor puro produz
ocorrências normalizadas consumidas por `cardinvoice` e `monthlysummary`; esses
módulos não recalculam numeração nem corte por quitação.

## 2. Decisões técnicas

### 2.1 Dívida como item financeiro especializado

Uma dívida é um `financial_item` de kind `debt_installment`, acompanhado por
metadados obrigatórios em `debts`.

Isso reaproveita vigências, métodos de pagamento, identidade de movimentação,
ownership e snapshots. A criação genérica rejeita esse kind; o caso de uso de
dívida cria item, período e metadados na mesma transação.

### 2.2 Cronograma derivado

Parcelas não são materializadas:

```text
remaining_count = total_installments - first_projected_installment + 1
scheduled_end = start_month + remaining_count - 1
installment_number(month)
  = first_projected_installment + months_between(start_month, month)
```

O período financeiro aplicável fornece o valor. Assim, mudanças temporais não
duplicam `amount_cents` nos metadados da dívida.

### 2.3 Término estrutural e efetivo

- `scheduled_end_month`: última parcela da estrutura original;
- `effective_end_month`: quitação antecipada ou término estrutural.

O cronograma original continua reconstruível após uma quitação. A projeção
atual substitui seu mês e suprime meses posteriores.

### 2.4 Quitação como evento autoral

`debt_early_settlements` persiste somente o fato não derivável. O evento é
imutável e único.

```text
month < settlement: parcela regular
month = settlement: ocorrência de quitação
month > settlement: sem ocorrência
```

Períodos originais não são apagados.

### 2.5 Identidade compartilhada

Uma ocorrência continua identificada por `financial_item_id +
reference_month`. Número e tipo são metadados. Movimentações R2 permanecem
válidas sem nova identidade.

### 2.6 Valor liberado não é lançamento

O valor liberado aparece em respostas e agregados de liberação, nunca como
renda ou economia. No término natural ele usa a última parcela regular; na
quitação usa a primeira parcela regular suprimida. A ausência futura da
parcela já produz o efeito financeiro.

### 2.7 Compatibilidade temporal

Mudanças de valor reutilizam períodos financeiros. O caso de uso especializado
preserva o término e rejeita mudanças fora do cronograma. Métodos reutilizam
`financial_item_payment_periods`, cuja validação passa a aceitar dívida.

## 3. Modelo de dados proposto

### 3.1 Extensão de `financial_item_kind`

Adicionar `debt_installment` com efeito de despesa, recorrência mensal, período
inicial delimitado e criação exclusiva pelo módulo `debt`.

O enum deve ser alterado em migration própria anterior às migrations que usem
o valor novo.

### 3.2 `debts`

```text
financial_item_id uuid PK
user_id uuid NOT NULL
original_total_cents bigint NULL
total_installments integer NOT NULL
first_projected_installment integer NOT NULL
scheduled_start_month date NOT NULL
scheduled_end_month date NOT NULL
created_at timestamptz NOT NULL
updated_at timestamptz NOT NULL
```

Invariantes:

- FK composta para item do mesmo proprietário;
- item obrigatoriamente `debt_installment`;
- total positivo e primeira parcela dentro do total;
- meses normalizados e término consistente com a fórmula;
- cronograma independente do horizonte, com duração estrutural máxima de 120
  meses;
- total original nulo ou não negativo;
- RLS por `user_id` e índices por usuário/término.

Nome, descrição, status e períodos permanecem nas tabelas de item.

### 3.3 `debt_early_settlements`

```text
id uuid PK
financial_item_id uuid NOT NULL
user_id uuid NOT NULL
reference_month date NOT NULL
amount_cents bigint NOT NULL
reason text NULL
recorded_by uuid NOT NULL
recorded_at timestamptz NOT NULL
created_at timestamptz NOT NULL
```

Invariantes:

- FK composta para dívida do mesmo proprietário;
- uma quitação por dívida;
- mês normalizado entre início e mês anterior ao término;
- valor positivo e motivo até 500 caracteres;
- evento append-only;
- RLS e escrita somente pela API Go.

### 3.4 Períodos da dívida

O período inicial usa o início e término derivados, valor da parcela,
recorrência `monthly` e offset informado ou zero. Mudança de valor fecha o
período aplicável e cria outro até o término estrutural.

Banco e aplicação garantem recorrência mensal, fim obrigatório, ausência de
sobreposição, contenção no cronograma e cobertura contínua sob transação.

## 4. Tipos de domínio

```go
type Debt struct {
    ID                        string
    Name                      string
    OriginalTotal             *Money
    TotalInstallments         int
    FirstProjectedInstallment int
    ScheduledStart            YearMonth
    ScheduledEnd              YearMonth
    Periods                   []InstallmentPeriod
    Settlement                *EarlySettlement
    Status                    FinancialItemStatus
}

type Occurrence struct {
    DebtID            string
    SourceID          string
    ReferenceMonth    YearMonth
    InstallmentNumber int
    InstallmentsTotal int
    Amount             Money
    Kind               OccurrenceKind
}

type Projection struct {
    ScheduledEnd    YearMonth
    EffectiveEnd    YearMonth
    ReleaseFrom     YearMonth
    ReleasedMonthly Money
    Occurrences     []Occurrence
}
```

`OccurrenceKind` aceita `scheduled` e `early_settlement`. Parcela regular usa
o período como `SourceID`; quitação usa o evento.

`remaining_installments` inclui a ocorrência do próprio `as_of_month` no
término natural. Antes de uma quitação, conta as ocorrências projetadas até o
evento substitutivo; a partir do mês em que a quitação já está registrada,
retorna zero. Dívidas arquivadas também retornam zero, sem perder a capacidade
de reconstruir seu cronograma para auditoria.

## 5. Fluxos

### 5.1 Criar dívida

1. autenticar e resolver a moeda padrão do cadastro financeiro;
2. validar nome, valores, numeração e início;
3. derivar término e validar o limite estrutural de 120 meses;
4. criar item, período, dívida e método opcional na mesma transação;
5. retornar dívida e resumo do cronograma.

### 5.2 Projetar cronograma

1. carregar dívida, períodos e quitação consistentemente;
2. derivar números por diferença mensal;
3. selecionar valor aplicável;
4. substituir quitação e eliminar meses posteriores;
5. resolver método, caixa e movimentação;
6. ordenar e calcular liberação com `Money` seguro.

Depois da leitura, o núcleo é puro:

```text
ProjectDebt(input) -> Projection
```

### 5.3 Alterar valor futuro

1. bloquear dívida e períodos;
2. rejeitar recurso arquivado ou mudança depois da quitação;
3. validar competência;
4. fechar período anterior e criar novo até o término;
5. validar cobertura e confirmar a transação.

### 5.4 Alterar método de pagamento

O fluxo R2 permanece. `IsExpense` passa a considerar `debt_installment`; o
projetor usa apenas meses com ocorrência.

### 5.5 Quitar antecipadamente

1. bloquear a dívida;
2. garantir item ativo e ausência de quitação;
3. validar mês e valor;
4. criar evento imutável sem remover períodos;
5. reprojetar e retornar término efetivo.

### 5.6 Integrar à fatura

`cardinvoice` carrega dívidas sem N+1, recebe ocorrências do módulo `debt`,
resolve método/cartão, aplica movimentação e transporta metadados. Ele não
calcula numeração, término ou quitação.

### 5.7 Integrar ao resumo

Em `reference`, entram ocorrências cuja competência é o mês. Em `cash`, entram
dívidas diretas no mês de caixa e dívidas de cartão pelos componentes da
fatura. Todas somam em `debt_installments_cents`, sem duplicidade.

### 5.8 Projetar liberações

Listar dívidas do intervalo, calcular término efetivo, selecionar liberações,
ordenar e agregar por mês com soma segura. O agregado não altera o resumo.

### 5.9 Snapshot v4

Novas ativações incluem dívidas, períodos, quitações e recursos R2 alcançáveis,
em ordem estável. Leituras v1/v2/v3 e reativação idempotente permanecem
intactas.

## 6. Contratos de resposta

### 6.1 Resumo de dívida

```json
{
  "id": "uuid",
  "name": "Transplante",
  "original_total_cents": 720000,
  "total_installments": 12,
  "first_projected_installment": 5,
  "scheduled_start_month": "2026-09",
  "scheduled_end_month": "2027-04",
  "effective_end_month": "2027-04",
  "release_from_month": "2027-05",
  "released_monthly_cents": 60000,
  "as_of_month": "2026-09",
  "projection_status": "active",
  "remaining_installments": 8,
  "status": "active"
}
```

### 6.2 Ocorrência

```json
{
  "debt_id": "uuid",
  "source_id": "uuid",
  "reference_month": "2026-09",
  "installment_number": 5,
  "installments_total": 12,
  "amount_cents": 60000,
  "debt_occurrence_kind": "scheduled",
  "payment_method": "credit_card",
  "cash_month": "2026-10",
  "credit_card_id": "uuid",
  "invoice_payment_month": "2026-10"
}
```

### 6.3 Liberação

```json
{
  "debt_id": "uuid",
  "scheduled_end_month": "2027-04",
  "effective_end_month": "2027-04",
  "release_from_month": "2027-05",
  "released_monthly_cents": 60000,
  "reason": "scheduled_completion"
}
```

`reason` também aceita `early_settlement`.

## 7. Concorrência e consistência

- criação é transacional;
- mudanças e quitação usam row lock;
- unicidade e exclusion constraints encerram corridas;
- projeções compostas usam `repeatable read`;
- somas usam `Money` e propagam overflow;
- `GET` nunca materializa ocorrências.

## 8. Segurança

- `ownerID` vem do principal;
- queries filtram `user_id`; o horizonte é parâmetro de projeção, não de
  ownership;
- FKs compostas e RLS impedem relações cruzadas;
- acesso alheio retorna `404`;
- Data API recebe somente leitura segura;
- logs não registram nomes, valores, cronogramas ou motivos.

## 9. Observabilidade

Logs podem incluir request ID, rota, status, duração, quantidade de dívidas ou
ocorrências, código de erro e indicação sem valor de quitação/inconsistência.

Métricas recomendadas: duração da projeção, ocorrências processadas, conflitos
de período, quitações rejeitadas e inconsistências item/dívida.

## 10. Testes

### Unidade

- parcelas restantes, numeração e término;
- início diferente de 1 e virada de ano;
- valor temporal, quitação, liberação e status;
- ordenação, overflow e ausência após término.

### Repositório e SQL

- relação item/dívida, kind, fórmula e limite estrutural;
- período delimitado, cobertura e não sobreposição;
- quitação única/imutável;
- ownership, RLS e grants.

### Integração HTTP

- CRUD, cronograma, mudança de valor e arquivamento;
- pagamento direto/cartão, movimentação e quitação;
- liberações, resumo, snapshot e isolamento.

### Compatibilidade

- suítes R1/R2 verdes;
- faturas sem dívida inalteradas;
- snapshots v1/v2/v3 legíveis;
- OpenAPI válido.

## 11. Erros adicionais

- `invalid_debt_schedule`;
- `debt_schedule_too_long`;
- `debt_period_gap`;
- `debt_archived`;
- `debt_already_settled`;
- `invalid_early_settlement_month`;
- `debt_occurrence_not_found`;
- `debt_projection_inconsistent`.

Validação retorna `422`, ausência/ownership `404` e conflito `409`.

## 12. Migration e rollback

Migrations aditivas:

1. adicionar `debt_installment` em migration isolada;
2. criar `debts`;
3. criar quitações, triggers, RLS e índices;
4. ampliar validações de método.

Rollback de aplicação exige desabilitar escrita R3 antes de retornar à R2. A
versão anterior ignora as tabelas, mas não entende novos itens. Dados e valor
do enum são preservados; remoção exige migration posterior aprovada.
