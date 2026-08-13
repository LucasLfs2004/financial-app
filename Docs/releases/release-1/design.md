# Release 1 — Design técnico — Quanto está livre?

## 1. Direção arquitetural

A API permanece um monólito modular em Go.

Cada feature concentra regras, casos de uso, portas e adaptadores. Infraestrutura
compartilhada permanece em `internal/platform`.

```text
cmd/api
internal/
  platform/
    auth/
    config/
    database/
    httpserver/
  profile/
  planning/
    domain/
    application/
    repository/
    transport/http/
  financialitem/
    domain/
    application/
    repository/
    transport/http/
  monthlysummary/
    application/
    transport/http/
```

As pastas podem ser simplificadas enquanto pequenas, mas dependências devem
apontar para o domínio, nunca do domínio para HTTP ou Postgres.

## 2. Decisões técnicas

### 2.1 Monólito modular

Não serão criados microserviços. Planejamento, itens e resumo participam da
mesma consistência transacional.

### 2.2 Postgres como fonte persistente

- Supabase Postgres;
- migrations SQL versionadas;
- `pgxpool` para conexões;
- transações explícitas em alterações temporais e ativação;
- nenhuma regra financeira escondida em trigger.

Triggers ficam restritos a invariantes técnicas, como `updated_at` e criação de
perfil.

### 2.3 Dinheiro

No domínio:

```go
type Money struct {
	cents int64
}

func NewMoney(cents int64) Money
func (m Money) Cents() int64
```

No banco:

```text
bigint
```

Regras:

- sem `float32` ou `float64`;
- centavos ficam encapsulados e não podem ser modificados após a criação;
- operações verificam overflow;
- moeda pertence ao planejamento;
- respostas retornam centavos e código da moeda.

### 2.4 Mês

No domínio:

```go
type YearMonth struct {
    Year  int
    Month time.Month
}
```

No banco:

- coluna `date`;
- sempre primeiro dia do mês;
- constraint garantindo `date_trunc('month', value) = value`;
- conversão HTTP estrita de `AAAA-MM`.

### 2.5 Autorização

O middleware da Release 0 fornece `auth.Principal`.

Casos de uso recebem `ownerID` dessa identidade. Repositórios utilizam:

```sql
where id = $1 and user_id = $2
```

Para recursos ausentes ou pertencentes a outro usuário, a API retorna `404`,
evitando revelar existência.

RLS replica a fronteira nas tabelas expostas:

```sql
to authenticated
using ((select auth.uid()) = user_id)
with check ((select auth.uid()) = user_id)
```

Na Release 1, `authenticated` recebe somente `SELECT` em `plans` e
`plan_snapshots`. Escritas passam pela API Go para que ativação, snapshots e
demais mudanças obedeçam às transações e regras de aplicação. A RLS continua
isolando leituras feitas pelo Data API.

### 2.6 Erros

Formato:

```json
{
  "error": {
    "code": "period_overlap",
    "message": "The new period overlaps an existing period",
    "details": {
      "item_id": "..."
    }
  }
}
```

Códigos previstos:

- `validation_error`;
- `not_found`;
- `plan_already_exists`;
- `plan_not_activatable`;
- `period_overlap`;
- `month_outside_horizon`;
- `unsupported_basis`;
- `conflict`;
- `internal_error`.

## 3. Modelo de dados proposto

### 3.1 `plans`

```text
id uuid PK
user_id uuid NOT NULL
name text NOT NULL
status plan_status NOT NULL
start_month date NOT NULL
end_month date NOT NULL
currency_code text NOT NULL
activated_at timestamptz NULL
archived_at timestamptz NULL
created_at timestamptz NOT NULL
updated_at timestamptz NOT NULL
```

Índices e invariantes:

- índice por `user_id`;
- unicidade parcial de planejamento não arquivado por usuário;
- início menor ou igual ao fim;
- meses normalizados para o primeiro dia;
- RLS por `user_id`.

### 3.2 `plan_snapshots`

```text
id uuid PK
plan_id uuid NOT NULL
user_id uuid NOT NULL
kind snapshot_kind NOT NULL
schema_version integer NOT NULL
document jsonb NOT NULL
created_at timestamptz NOT NULL
```

Para a Release 1:

- apenas `kind = original`;
- um original por planejamento;
- `schema_version` positivo identifica o formato do documento;
- update é bloqueado por trigger e delete não é concedido ao cliente;
- ativação cria snapshot e altera status na mesma transação.

O JSON é apropriado para a fotografia imutável. Dados operacionais continuam
normalizados.

### 3.3 `financial_items`

```text
id uuid PK
plan_id uuid NOT NULL
user_id uuid NOT NULL
name text NOT NULL
kind financial_item_kind NOT NULL
description text NULL
status financial_item_status NOT NULL
archived_at timestamptz NULL
created_at timestamptz NOT NULL
updated_at timestamptz NOT NULL
```

Kinds:

- `recurring_income`;
- `one_time_income`;
- `fixed_expense`;
- `projected_variable_expense`.

### 3.4 `financial_item_periods`

```text
id uuid PK
financial_item_id uuid NOT NULL
plan_id uuid NOT NULL
user_id uuid NOT NULL
start_month date NOT NULL
end_month date NULL
amount_cents bigint NOT NULL
recurrence recurrence_kind NOT NULL
cash_month_offset smallint NOT NULL DEFAULT 0
context text NULL
recorded_at timestamptz NOT NULL
created_at timestamptz NOT NULL
```

Invariantes:

- `amount_cents >= 0`;
- offset inicial entre 0 e 12;
- `monthly` permite intervalo;
- `once` exige início igual ao fim;
- períodos do mesmo item não se sobrepõem;
- chaves estrangeiras compostas garantem que plano, item e período tenham o
  mesmo proprietário;
- RLS por `user_id`;
- índice por item e intervalo;
- índice por planejamento e intervalo.

Postgres reforça a não sobreposição com uma exclusion constraint GiST sobre
`daterange` inclusivo. `end_month = null` é tratado como infinito. A aplicação
continua validando para retornar erro de domínio legível.

A compatibilidade entre `financial_item_kind` e `recurrence_kind`, além da
vigência dentro do horizonte do plano, permanece na aplicação. Essas regras
dependem de dados entre agregados e não serão escondidas em triggers.

### 3.5 `saving_periods`

```text
id uuid PK
plan_id uuid NOT NULL
user_id uuid NOT NULL
start_month date NOT NULL
end_month date NULL
amount_cents bigint NOT NULL
context text NULL
created_at timestamptz NOT NULL
```

Invariantes equivalentes aos períodos de item, incluindo não sobreposição.

## 4. Fluxos

### 4.1 Criar rascunho

1. middleware autentica;
2. caso de uso verifica planejamento não arquivado;
3. valida horizonte, nome e moeda;
4. persiste `draft`;
5. retorna recurso.

### 4.2 Adicionar item

1. valida proprietário e planejamento;
2. valida kind;
3. cria item e período inicial na mesma transação;
4. verifica vigência dentro do horizonte;
5. retorna item com período.

### 4.3 Alterar a partir de um mês

1. carrega item e períodos com lock;
2. identifica período aplicável;
3. valida mês efetivo;
4. encerra período anterior;
5. cria novo período;
6. registra `recorded_at`;
7. mantém competências anteriores;
8. invalida apenas caches futuros, caso existam futuramente.

### 4.4 Ativar

1. bloqueia planejamento;
2. se já ativo, retorna snapshot existente;
3. valida renda e configuração de economia;
4. lê premissas em ordem determinística;
5. cria JSON versionado;
6. insere snapshot original;
7. atualiza planejamento para `active`;
8. commit.

### 4.5 Calcular resumo

1. valida mês e base;
2. carrega planejamento do usuário;
3. seleciona períodos aplicáveis;
4. expande somente a ocorrência solicitada;
5. desloca para caixa quando `basis=cash`;
6. separa rendas e compromissos;
7. carrega economia aplicável;
8. calcula totais em `int64`;
9. compara totais com componentes;
10. monta resposta explicável.

O cálculo deve ser uma função de aplicação pura após o carregamento:

```text
CalculateMonthlySummary(input) → summary
```

## 5. Resposta do resumo

Exemplo reduzido:

```json
{
  "data": {
    "month": "2026-07",
    "basis": "cash",
    "result_kind": "planned_free",
    "currency_code": "BRL",
    "plan_status": "draft",
    "income_cents": 600000,
    "commitments_cents": 280000,
    "planned_savings_cents": 120000,
    "result_cents": 200000,
    "is_negative": false,
    "completeness": "projected",
    "breakdown": {
      "fixed_expenses_cents": 210000,
      "projected_variable_expenses_cents": 70000
    },
    "sources": []
  }
}
```

`result_kind` vale `planned_free` na base `cash` e
`planned_reference_result` na base `reference`. O campo neutro `result_cents`
evita apresentar o resultado por competência como dinheiro disponível em
conta.

Cada source contém:

- item ID;
- nome;
- kind;
- período aplicado;
- mês de referência;
- mês de caixa;
- valor;
- participação no cálculo.

## 6. Concorrência e idempotência

- criação de plano protegida por índice único;
- ativação protegida por transação e unicidade do snapshot;
- alteração temporal usa row lock;
- conflitos retornam `409`;
- endpoints de comando que possam sofrer retry devem aceitar
  `Idempotency-Key` em evolução posterior;
- ativação já é naturalmente idempotente pelo estado e índice.

## 7. Observabilidade

Logs estruturados incluem:

- request ID;
- rota;
- status;
- duração;
- user ID apenas quando necessário e sem e-mail;
- código de erro;
- duração do cálculo;
- quantidade de fontes consideradas.

Não registrar:

- token;
- renda;
- valores financeiros;
- nome dos itens;
- documentos de snapshot.

## 8. Testes

### Unidade

- `Money`;
- `YearMonth`;
- vigências;
- deslocamento de caixa;
- fórmulas;
- resultado negativo;
- zero versus ausência;
- ativação.

### Repositório

- ownership obrigatório;
- constraints;
- não sobreposição;
- snapshot imutável;
- unicidade de plano.

### Integração HTTP

- autenticação;
- validações;
- erros;
- isolamento entre dois usuários;
- resumo nas duas bases;
- mudança futura preservando passado.

### Aceitação

Cenário mínimo:

```text
salário
despesas fixas
gasolina projetada
valor para guardar
resultado mensal explicável
```

## 9. OpenAPI

O contrato será atualizado antes dos handlers. Cada operação deve definir:

- autenticação;
- request;
- response;
- erros;
- exemplos;
- formato de dinheiro;
- formato de mês;
- semântica da base.

Mudanças incompatíveis exigem nova versão ou decisão explícita.
