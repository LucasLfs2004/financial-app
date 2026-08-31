# Release 2 — Design técnico — Cartões e faturas projetadas

## 1. Direção arquitetural

A API continua como monólito modular em Go. A release adiciona os módulos
abaixo e integra a projeção ao cálculo mensal existente.

```text
internal/
  financialinstitution/
    domain/
    application/
    repository/
    transport/http/
  creditcard/
    domain/
    application/
    repository/
    transport/http/
  paymentmethod/
    domain/
    application/
    repository/
    transport/http/
  cardinvoice/
    domain/
    application/
    repository/
    transport/http/
  monthlysummary/
```

O domínio de fatura depende de portas de leitura de itens, métodos de pagamento
e cartões. Ele não depende de HTTP nem de Postgres. `monthlysummary` reutiliza
o mesmo seletor de ocorrências de fatura para evitar duas implementações da
regra financeira.

## 2. Decisões técnicas

### 2.1 Fatura como projeção, não como tabela mensal

A identidade pública da fatura é `(card_id, payment_month)`. Não existe uma
linha `card_invoices` criada para cada mês.

Motivos:

- o total é derivado e muda quando premissas mudam;
- um `GET` não deve materializar estado;
- não haverá sincronização de milhares de linhas futuras;
- cartão e mês já formam uma identidade estável para rotas e cache.

Somente dados autorais da fatura são persistidos: ajustes e movimentações.

### 2.2 Configurações temporais

Vencimento, offset e método de pagamento usam períodos mensais inclusivos. O
mesmo mecanismo conceitual da Release 1 é mantido:

- alteração encerra o período anterior;
- cria um novo período;
- registra `recorded_at`;
- não sobrescreve o passado;
- Postgres impede sobreposição.

O período de cartão aplicável à referência fornece o offset. O período de
cartão aplicável ao mês de pagamento fornece o vencimento nominal exibido.

### 2.3 Seleção única de ocorrências

Uma ocorrência de despesa é identificada por:

```text
financial_item_id + reference_month
```

O período financeiro aplicável fornece valor e tipo. O período de método de
pagamento aplicável fornece `direct` ou `credit_card`. Na ausência deste,
assume-se `direct`.

Para cartão:

```text
default_payment_month
= reference_month + card_offset(reference_month)
```

A última movimentação válida, se existir, substitui cartão e mês padrão.

### 2.4 Componentes, não um total duplicado

O projetor retorna componentes normalizados:

```go
type Component struct {
    SourceID        string // UUID do período financeiro ou do ajuste
    SourceType      ComponentType
    ItemID          *string
    AdjustmentID    *string
    Name            string
    ReferenceMonth  *YearMonth
    PaymentMonth    YearMonth
    CardID           string
    Amount           Money
    Allocation       AllocationOrigin
}
```

`cardinvoice` soma os componentes para mostrar a fatura. `monthlysummary` usa
os mesmos componentes como sources de compromisso na base `cash`. Não cria
uma source separada para o total da fatura.

Na base `reference`, ocorrências de item usam a seleção já existente e ajustes
somente entram quando possuem referência explícita.

### 2.5 Vencimento nominal

O domínio não chama calendário bancário. Para o mês da fatura:

- se o dia existe, retorna a data nominal e resolução `exact`;
- se não existe, retorna data nula e resolução `invalid_for_month`.

Não existe fallback para último dia útil, fim do mês ou próximo dia útil.

### 2.6 Instituições e cartões fora do plano

Instituições e cartões pertencem ao usuário, não ao planejamento. Isso permite
preservar o cadastro quando um planejamento for arquivado e outro criado.

Métodos, ajustes e movimentações carregam `plan_id`, pois participam de uma
projeção específica e precisam respeitar seu horizonte.

## 3. Modelo de dados proposto

### 3.1 `financial_institutions`

```text
id uuid PK
user_id uuid NOT NULL
name text NOT NULL
status financial_resource_status NOT NULL
archived_at timestamptz NULL
created_at timestamptz NOT NULL
updated_at timestamptz NOT NULL
```

Invariantes:

- nome normalizado não vazio e com até 120 caracteres;
- unicidade case-insensitive entre instituições ativas do usuário;
- RLS por `user_id`;
- índice por usuário e status.

### 3.2 `credit_cards`

```text
id uuid PK
user_id uuid NOT NULL
institution_id uuid NOT NULL
name text NOT NULL
status financial_resource_status NOT NULL
archived_at timestamptz NULL
created_at timestamptz NOT NULL
updated_at timestamptz NOT NULL
```

Invariantes:

- FK composta garante instituição do mesmo usuário;
- nome não vazio e com até 120 caracteres;
- unicidade case-insensitive de nome ativo por instituição e usuário;
- cartão arquivado não volta a `active` nesta release;
- RLS por `user_id`.

### 3.3 `credit_card_periods`

```text
id uuid PK
credit_card_id uuid NOT NULL
user_id uuid NOT NULL
start_month date NOT NULL
end_month date NULL
nominal_due_day smallint NOT NULL
payment_month_offset smallint NOT NULL DEFAULT 1
context text NULL
recorded_at timestamptz NOT NULL
created_at timestamptz NOT NULL
```

Invariantes:

- meses no primeiro dia;
- dia entre 1 e 31;
- offset entre 0 e 12;
- intervalos inclusivos e não sobrepostos por cartão;
- FK composta preserva ownership;
- GiST `daterange` para não sobreposição;
- índices por cartão e intervalo;
- RLS por `user_id`.

### 3.4 `financial_item_payment_periods`

```text
id uuid PK
financial_item_id uuid NOT NULL
plan_id uuid NOT NULL
user_id uuid NOT NULL
start_month date NOT NULL
end_month date NULL
method payment_method_kind NOT NULL
credit_card_id uuid NULL
context text NULL
recorded_at timestamptz NOT NULL
created_at timestamptz NOT NULL
```

Invariantes:

- somente itens de despesa;
- `credit_card` exige cartão, `direct` exige nulo;
- período contido no horizonte do plano;
- no máximo um período explícito por item e competência;
- ausência de linha significa `direct`;
- FKs compostas garantem mesmo plano/proprietário do item e proprietário do
  cartão;
- GiST contra sobreposição;
- RLS por `user_id`.

### 3.5 `card_invoice_adjustments`

```text
id uuid PK
plan_id uuid NOT NULL
user_id uuid NOT NULL
credit_card_id uuid NOT NULL
payment_month date NOT NULL
reference_month date NULL
name text NOT NULL
amount_cents bigint NOT NULL
context text NULL
status invoice_adjustment_status NOT NULL
archived_at timestamptz NULL
created_at timestamptz NOT NULL
updated_at timestamptz NOT NULL
```

Invariantes:

- `amount_cents >= 0`;
- meses normalizados;
- pagamento no horizonte operacional;
- referência, quando presente, dentro do plano;
- cartão, plano e ajuste do mesmo usuário;
- RLS por `user_id`;
- índices por cartão/mês e plano/referência.

O primeiro escopo não mantém versões completas de texto e valor. Alterações
registram evento de auditoria na tabela seguinte antes do update.

### 3.6 `card_invoice_audit_events`

```text
id uuid PK
plan_id uuid NOT NULL
user_id uuid NOT NULL
event_type card_invoice_event_type NOT NULL
financial_item_id uuid NULL
reference_month date NULL
adjustment_id uuid NULL
from_credit_card_id uuid NULL
from_payment_month date NULL
to_credit_card_id uuid NULL
to_payment_month date NULL
before_document jsonb NULL
after_document jsonb NULL
reason text NULL
recorded_at timestamptz NOT NULL
```

Eventos suportados:

- `occurrence_moved`;
- `adjustment_changed`;
- `adjustment_archived`.

Para `occurrence_moved`, item, referência, origem e destino são obrigatórios.
O estado atual é a última movimentação ordenada por `recorded_at, id`. A
aplicação valida que a origem informada coincide com a alocação atual e grava o
evento em transação.

Eventos são append-only. Update e delete são bloqueados; o cliente autenticado
não recebe grants diretos de escrita.

## 4. Fluxos

### 4.1 Criar cartão

1. autentica e resolve proprietário;
2. carrega instituição ativa do usuário;
3. valida nome, mês inicial, vencimento e offset;
4. cria cartão e período inicial na mesma transação;
5. retorna cartão com configuração atual.

A instituição do cartão é imutável depois da criação. Uma correção de emissor
exige arquivar o cartão e criar outro, preservando a filiação histórica.

### 4.2 Alterar configuração do cartão

1. carrega cartão e períodos com lock;
2. valida `effective_from`;
3. encerra o período anterior no mês precedente;
4. cria novo período;
5. rejeita sobreposição ou lacuna inválida;
6. preserva faturas anteriores à vigência.

### 4.3 Alterar forma de pagamento

1. carrega plano, item e períodos com lock;
2. garante que o item é despesa;
3. valida cartão quando necessário;
4. encerra o método explícito anterior;
5. cria período `direct` ou `credit_card`;
6. mantém ausência anterior como pagamento direto implícito;
7. retorna histórico ordenado.

### 4.4 Projetar uma fatura

1. valida cartão, mês e horizonte operacional;
2. carrega itens e períodos financeiros candidatos;
3. seleciona ocorrências vinculadas ao cartão;
4. deriva a fatura padrão pela configuração da competência;
5. aplica a última movimentação válida;
6. mantém somente componentes destinados ao cartão e mês consultados;
7. adiciona ajustes ativos;
8. resolve vencimento nominal pela configuração do mês de pagamento;
9. ordena componentes;
10. soma com operações seguras e valida consistência.

O núcleo é puro depois do carregamento:

```text
ProjectInvoice(input) -> Invoice
```

### 4.5 Mover ocorrência

1. resolve a ocorrência por item e referência;
2. calcula sua alocação atual;
3. valida cartão e mês de destino;
4. rejeita destino igual à origem;
5. grava evento append-only com origem calculada;
6. retorna a nova alocação e histórico.

### 4.6 Calcular resumo mensal

Base `reference`:

1. seleciona ocorrências financeiras pela competência;
2. adiciona ajustes com referência explícita;
3. ignora ajustes sem referência;
4. calcula totais uma vez.

Base `cash`:

1. seleciona rendas e despesas diretas com as regras da Release 1;
2. exclui dessa seleção ocorrências vinculadas a cartão;
3. projeta componentes de todas as faturas pagas no mês;
4. converte cada componente em uma source de compromisso;
5. calcula os totais sem adicionar source de fatura agregada.

Sources de cartão adicionam campos opcionais:

```text
payment_method
credit_card_id
invoice_payment_month
invoice_allocation
reference_known
```

O breakdown adiciona `card_invoice_adjustments_cents`. Componentes originados
em itens continuam somando em `fixed_expenses_cents` ou
`projected_variable_expenses_cents`; apenas ajustes usam o novo bucket.

### 4.7 Ativar plano e gerar snapshot v2

1. mantém leitura de snapshot v1;
2. usa versão 2 para toda nova ativação após a implantação;
3. ordena todos os recursos e períodos por chaves estáveis;
4. inclui somente cartões e instituições alcançáveis pelo plano;
5. inclui ajustes e eventos vigentes no momento da ativação;
6. representa capacidades ainda não usadas com coleções vazias;
7. grava snapshot e ativa na mesma transação existente.

## 5. Contratos de resposta

### 5.1 Fatura reduzida

```json
{
  "card_id": "uuid",
  "card_name": "Nubank principal",
  "institution": {"id": "uuid", "name": "Nubank"},
  "payment_month": "2026-12",
  "nominal_due_day": 6,
  "nominal_due_date": "2026-12-06",
  "nominal_due_date_resolution": "exact",
  "currency_code": "BRL",
  "projected_total_cents": 88000,
  "component_count": 2
}
```

### 5.2 Componente

```json
{
  "source_id": "9f3cba5a-621f-4bfa-b4a0-62341ff02a54",
  "source_type": "financial_item_occurrence",
  "item_id": "uuid",
  "adjustment_id": null,
  "name": "Gasolina",
  "reference_month": "2026-11",
  "reference_known": true,
  "payment_month": "2026-12",
  "amount_cents": 70000,
  "allocation": "calculated_from_reference"
}
```

Para ajuste sem referência, `reference_month` é nulo,
`reference_known = false` e `allocation = selected_invoice`.

## 6. Concorrência e cache

- alterações temporais usam row lock;
- exclusion constraints encerram corridas de sobreposição;
- eventos append-only têm índice para busca da última alocação;
- faturas podem receber ETag derivado da versão máxima das fontes;
- cache, se introduzido, é invalidado por cartão e intervalo afetado;
- a primeira implementação pode operar sem cache.

## 7. Segurança

- todas as consultas incluem `user_id` vindo do principal autenticado;
- acesso cruzado retorna `404`;
- RLS replica a fronteira em todas as tabelas novas;
- Data API recebe somente `SELECT` onde leitura direta for segura;
- ajustes, períodos e auditoria são escritos somente pela API Go;
- payloads de auditoria nunca armazenam token ou dados de autenticação.

## 8. Observabilidade

Logs estruturados podem incluir:

- request ID;
- rota, status e duração;
- quantidade de componentes;
- quantidade de cartões consultados;
- código de erro;
- indicação de overflow ou inconsistência.

Não registrar:

- nomes de itens, cartões ou instituições;
- valores financeiros;
- documentos de auditoria;
- token, e-mail ou payload completo.

Métricas recomendadas:

- duração da projeção;
- componentes processados por fatura;
- inconsistências rejeitadas;
- conflitos de período;
- movimentações por resultado.

## 9. Testes

### Unidade

- vencimento nominal válido e inválido;
- offset 0, 1 e 12;
- seleção temporal de cartão e método;
- alocação padrão e movimentada;
- ajustes com e sem referência;
- ordenação e total da fatura;
- overflow;
- não duplicação no resumo.

### Repositório

- ownership composto;
- RLS;
- não sobreposição;
- auditoria append-only;
- unicidade de nomes ativos;
- cartão arquivado;
- queries de componentes e histórico.

### Integração HTTP

- CRUD e arquivamento;
- alterações por vigência;
- listagem e detalhe de fatura;
- ajuste e movimentação;
- erros e idempotência;
- isolamento entre dois usuários;
- resumo nas duas bases;
- fatura de transbordo.

### Compatibilidade

- suíte completa da Release 1 permanece verde;
- item sem método explícito mantém o cálculo anterior;
- snapshot v1 continua legível;
- OpenAPI continua válido.

## 10. Erros adicionais

- `institution_archived`;
- `card_archived`;
- `invalid_payment_method`;
- `payment_period_overlap`;
- `card_configuration_missing`;
- `invoice_outside_operational_horizon`;
- `occurrence_not_found`;
- `occurrence_not_card_linked`;
- `same_invoice_destination`;
- `stale_invoice_allocation`;
- `adjustment_archived`;
- `invoice_inconsistent`.

Validação retorna `422`, ausência/ownership retorna `404` e conflitos de estado
ou concorrência retornam `409`, seguindo o envelope da Release 1.

## 11. Migration e rollback

As migrations são aditivas. Nenhuma coluna da Release 1 é removida ou
reinterpretada no banco.

Rollback de aplicação:

- versão anterior ignora as tabelas novas;
- itens sem uso dos endpoints novos continuam funcionando;
- dados novos permanecem preservados até nova implantação.

Rollback de schema em produção não apaga imediatamente tabelas com dados. A
remoção, se necessária, ocorre em migration posterior e explicitamente
aprovada.
