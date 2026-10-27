# Contratos de integração — Frontend 0–3.1

**Data:** 04/10/2026. Documentação de implementação; não altera o OpenAPI nem implementa handlers.

Leia o [design](design.md) para os contratos atuais divergentes e as extensões propostas. Catálogo das 48 operações canônicas do OpenAPI: 46 autenticadas de produto e 2 de saúde. Aliases legados excluídos.

## Convenções

- Browser: `/api/financial/<caminho sem /v1>`; servidor: `FINANCIAL_API_URL + /v1/...`.
- API de domínio exige `Authorization: Bearer <access-token>`; o proxy extrai a sessão e transmite o token somente à API configurada.
- JSON usa `Content-Type: application/json`; importação usa multipart gerado por FormData.
- Obrigatório anulável deve estar presente com null; omissão não equivale a null.
- Schemas abaixo refletem o OpenAPI nesta data. Divergências dos handlers têm precedência para integrar a versão atual; propostas não estão publicadas.
- O proxy frontend preserva status/body do upstream (inclusive 201 e erros), query e multipart. FE-T01 validada localmente; smoke autenticado pendente.

## Estado da integração frontend

As rotas de detalhe, históricos, composição mensal e snapshot foram implementadas. Evidências em [quality-report](quality-report.md); dependências em [pending-items](pending-items.md). Payment-history e invoice-moves são consumidos como arrays conforme os handlers atuais. `current_allocation` é opcional até publicação da extensão. A manutenção de ajustes segue como contrato proposto: o frontend está preparado, mas `NEXT_PUBLIC_ENABLE_INVOICE_ADJUSTMENT_MAINTENANCE` permanece desabilitado por padrão. Os schemas abaixo preservam o catálogo original e suas divergências explícitas; não equivalem a um OpenAPI atualizado na API.

Fixtures históricas v1–v4 e desconhecida: [tests/fixtures](../../tests/fixtures/README.md). O visualizador usa somente o JSON congelado.

## Rotas, requests e responses

### GET `/health`

Operação: `getHealth`. Verificar se o processo está vivo

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: Status |

### GET `/ready`

Operação: `getReadiness`. Verificar as dependências da API

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: Status |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/me`

Operação: `getCurrentUser`. Consultar o perfil do usuário autenticado

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CurrentUserResponse |
| 401 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/plans`

Operação: `createPlan`. Criar o planejamento principal em rascunho

Request: `application/json`, body obrigatório; `CreatePlanRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: PlanResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/plans/current`

Operação: `getCurrentPlan`. Consultar o planejamento principal

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: PlanResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PATCH `/v1/plans/current`

Operação: `updateCurrentPlan`. Alterar os metadados ou horizonte do rascunho

Request: `application/json`, body obrigatório; `UpdatePlanRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: PlanResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/plans/current/activate`

Operação: `activateCurrentPlan`. Ativar o planejamento e criar a referência original

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: PlanActivationResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/plans/current/original`

Operação: `getOriginalPlanSnapshot`. Consultar a fotografia original imutável

**Divergência/atenção:** SnapshotPlanData é incompleto para schemas 2–4; documento real inclui release_2/release_3. Ver D-04/FE-T05.

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: OriginalPlanSnapshotResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-items`

Operação: `createFinancialItem`. Cadastrar uma renda ou despesa projetada

Request: `application/json`, body obrigatório; `CreateFinancialItemRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: FinancialItemResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/financial-items`

Operação: `listFinancialItems`. Listar os registros financeiros do usuário

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `kind` | query | não | `FinancialItemKind` | enum=["recurring_income", "one_time_income", "fixed_expense", "projected_variable_expense", "debt_installment"]  |
| `status` | query | não | `FinancialItemStatus` | enum=["active", "archived"]  |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialItemListResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/financial-items/{item_id}`

Operação: `getFinancialItem`. Consultar um item e seu histórico de períodos

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialItemResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PATCH `/v1/financial-items/{item_id}`

Operação: `updateFinancialItem`. Alterar nome ou descrição de um item

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `UpdateFinancialItemRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialItemResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-items/{item_id}/changes`

Operação: `changeFinancialItemFromMonth`. Alterar um item a partir de uma competência

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `CreateFinancialItemChangeRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: FinancialItemResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-items/{item_id}/archive`

Operação: `archiveFinancialItem`. Arquivar um item sem apagar seu histórico

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |

Request: `application/json`, body opcional; `ArchiveFinancialItemRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialItemResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/debts`

Operação: `createDebt`. Cadastrar uma dívida parcelada

Request: `application/json`, body obrigatório; `CreateDebtRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: DebtResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/debts`

Operação: `listDebts`. Listar dívidas do usuário

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `as_of` | query | não | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês usado para calcular status e parcelas restantes. Quando ausente, usa o mês atual no timezone do perfil do usuário.  |
| `projection_status` | query | não | `DebtProjectionStatus` | enum=["planned", "active", "completed", "settled_early", "archived"]  |
| `status` | query | não | `FinancialItemStatus` | enum=["active", "archived"]  |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtListResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/debts/{debt_id}`

Operação: `getDebt`. Consultar uma dívida e seu término projetado

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |
| `as_of` | query | não | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês usado para calcular status e parcelas restantes. Quando ausente, usa o mês atual no timezone do perfil do usuário.  |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PATCH `/v1/debts/{debt_id}`

Operação: `updateDebt`. Alterar nome ou descrição da dívida

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `UpdateDebtRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/debts/{debt_id}/changes`

Operação: `changeDebtAmountFromMonth`. Alterar o valor projetado a partir de uma competência

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `CreateDebtChangeRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: DebtResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/debts/{debt_id}/archive`

Operação: `archiveDebt`. Arquivar uma dívida sem apagar seu histórico

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |

Request: `application/json`, body opcional; `ArchiveFinancialItemRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/debts/{debt_id}/schedule`

Operação: `getDebtSchedule`. Projetar o cronograma da dívida em um intervalo

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |
| `from` | query | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Primeiro mês do intervalo inclusivo, limitado a 120 meses. |
| `to` | query | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Último mês do intervalo inclusivo, limitado a 120 meses. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtScheduleResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/debts/{debt_id}/early-settlement`

Operação: `createDebtEarlySettlement`. Registrar quitação antecipada projetada

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `CreateDebtEarlySettlementRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: DebtEarlySettlementResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/debts/{debt_id}/early-settlement`

Operação: `getDebtEarlySettlement`. Consultar a quitação antecipada da dívida

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `debt_id` | path | sim | `UUID` | — Identificador da dívida pertencente ao usuário autenticado. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtEarlySettlementResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/debt-releases`

Operação: `listDebtReleases`. Listar valores mensais liberados pelo término de dívidas

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `from` | query | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Primeiro mês do intervalo inclusivo, limitado a 120 meses. |
| `to` | query | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Último mês do intervalo inclusivo, limitado a 120 meses. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: DebtReleaseListResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/plans/current/savings`

Operação: `getPlannedSavings`. Consultar os períodos do valor planejado para guardar

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: PlannedSavingsResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PUT `/v1/plans/current/savings`

Operação: `putPlannedSavings`. Definir o valor para guardar a partir de um mês

Request: `application/json`, body obrigatório; `PutPlannedSavingsRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: PlannedSavingsResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-institutions`

Operação: `createFinancialInstitution`. Cadastrar uma instituição financeira

Request: `application/json`, body obrigatório; `CreateFinancialInstitutionRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: FinancialInstitutionResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/financial-institutions`

Operação: `listFinancialInstitutions`. Listar instituições financeiras

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `status` | query | não | `FinancialResourceStatus` | enum=["active", "archived"]  |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialInstitutionListResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PATCH `/v1/financial-institutions/{institution_id}`

Operação: `updateFinancialInstitution`. Alterar o nome de uma instituição ativa

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `institution_id` | path | sim | `UUID` | — Instituição pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `UpdateFinancialInstitutionRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialInstitutionResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-institutions/{institution_id}/archive`

Operação: `archiveFinancialInstitution`. Arquivar uma instituição sem apagar seu histórico

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `institution_id` | path | sim | `UUID` | — Instituição pertencente ao usuário autenticado. |

Request: `application/json`, body opcional; `ArchiveResourceRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: FinancialInstitutionResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/credit-cards`

Operação: `createCreditCard`. Cadastrar um cartão e sua configuração inicial

Request: `application/json`, body obrigatório; `CreateCreditCardRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: CreditCardResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/credit-cards`

Operação: `listCreditCards`. Listar cartões do usuário autenticado

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `status` | query | não | `FinancialResourceStatus` | enum=["active", "archived"]  |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CreditCardListResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/credit-cards/{card_id}`

Operação: `getCreditCard`. Consultar um cartão e suas configurações

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CreditCardResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PATCH `/v1/credit-cards/{card_id}`

Operação: `updateCreditCard`. Alterar o nome de um cartão ativo

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `UpdateCreditCardRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CreditCardResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/credit-cards/{card_id}/changes`

Operação: `changeCreditCardConfiguration`. Alterar vencimento ou offset a partir de um mês

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `CreateCreditCardConfigurationChangeRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: CreditCardResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/credit-cards/{card_id}/archive`

Operação: `archiveCreditCard`. Arquivar um cartão sem apagar suas faturas históricas

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |

Request: `application/json`, body opcional; `ArchiveResourceRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CreditCardResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-items/{item_id}/payment-changes`

Operação: `changeFinancialItemPaymentMethod`. Alterar a forma de pagamento de uma despesa por vigência

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |

Request: `application/json`, body obrigatório; `CreatePaymentMethodChangeRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: PaymentMethodHistoryResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/financial-items/{item_id}/payment-history`

Operação: `getFinancialItemPaymentHistory`. Consultar o histórico de formas de pagamento

**Divergência/atenção:** Atual: {data: PaymentMethodPeriod[]}; o objeto PaymentMethodHistory do OpenAPI não é retornado. Ver D-03/FE-T04.

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: PaymentMethodHistoryResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/credit-cards/{card_id}/invoices`

Operação: `listCreditCardInvoices`. Listar faturas projetadas em um intervalo

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `from` | query | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Primeiro mês do intervalo inclusivo de faturas. |
| `to` | query | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Último mês do intervalo inclusivo de faturas. |
| `include_empty` | query | não | `boolean` | default=false Inclui meses sem componentes quando verdadeiro. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CardInvoiceListResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/credit-cards/{card_id}/invoices/{payment_month}`

Operação: `getCreditCardInvoice`. Consultar uma fatura projetada e sua composição

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `payment_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês de pagamento que identifica a fatura projetada. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CardInvoiceResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/credit-cards/{card_id}/invoices/{payment_month}/imports/preview`

Operação: `previewNubankInvoiceImport`. Pré-visualizar importação de CSV Nubank

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `payment_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês de pagamento que identifica a fatura projetada. |

Request: `multipart/form-data`, body obrigatório; `{ file: string (binary), payment_allocations: string }`.

```json
{
  "type": "object",
  "required": [
    "file"
  ],
  "properties": {
    "file": {
      "type": "string",
      "format": "binary"
    },
    "payment_allocations": {
      "type": "string",
      "description": "JSON que associa o número da linha CSV ao mês da fatura do pagamento, por exemplo `{\"29\":\"2026-09\"}`."
    }
  }
}
```

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: InvoiceImportResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |

### POST `/v1/credit-cards/{card_id}/invoices/{payment_month}/imports`

Operação: `confirmNubankInvoiceImport`. Confirmar importação de CSV Nubank

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `payment_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês de pagamento que identifica a fatura projetada. |

Request: `multipart/form-data`, body obrigatório; `{ file: string (binary), preview_token: string, payment_allocations: string }`.

```json
{
  "type": "object",
  "required": [
    "file",
    "preview_token"
  ],
  "properties": {
    "file": {
      "type": "string",
      "format": "binary"
    },
    "preview_token": {
      "type": "string"
    },
    "payment_allocations": {
      "type": "string",
      "description": "Obrigatório quando o CSV contém `Pagamento recebido`; deve repetir o JSON usado na prévia final."
    }
  }
}
```

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: InvoiceImportResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | Cartão arquivado |
| 422 | application/json: Error |
| 500 | application/json: Error |

### POST `/v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments`

Operação: `createCardInvoiceAdjustment`. Adicionar uma cobrança consolidada diretamente à fatura

**Divergência/atenção:** Handler atual reduz erros de serviço a 400 validation_error. Ver FE-T03.

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `payment_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês de pagamento que identifica a fatura projetada. |

Request: `application/json`, body obrigatório; `CreateCardInvoiceAdjustmentRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: CardInvoiceAdjustmentResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### PATCH `/v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}`

Operação: `updateCardInvoiceAdjustment`. Alterar um ajuste preservando auditoria

**Divergência/atenção:** PATCH parcial é documentado, mas o handler atual exige dados completos e pode apagar campos omitidos. Bloqueado por FE-T03.

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `payment_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês de pagamento que identifica a fatura projetada. |
| `adjustment_id` | path | sim | `UUID` | — Ajuste pertencente à fatura indicada na rota. |

Request: `application/json`, body obrigatório; `UpdateCardInvoiceAdjustmentRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CardInvoiceAdjustmentResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}/archive`

Operação: `archiveCardInvoiceAdjustment`. Arquivar um ajuste preservando auditoria

**Divergência/atenção:** Handler atual ignora reason do body e devolve 409 adjustment_rejected para erros. Não prometer auditoria do motivo.

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `card_id` | path | sim | `UUID` | — Cartão pertencente ao usuário autenticado. |
| `payment_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Mês de pagamento que identifica a fatura projetada. |
| `adjustment_id` | path | sim | `UUID` | — Ajuste pertencente à fatura indicada na rota. |

Request: `application/json`, body opcional; `ArchiveResourceRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: CardInvoiceAdjustmentResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### POST `/v1/financial-items/{item_id}/occurrences/{reference_month}/invoice-moves`

Operação: `moveFinancialItemOccurrenceInvoice`. Mover uma ocorrência para outra fatura

**Divergência/atenção:** Rejeição atual: 409 invoice_move_rejected.

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |
| `reference_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Competência que identifica a ocorrência do item. |

Request: `application/json`, body obrigatório; `CreateInvoiceOccurrenceMoveRequest`.

| HTTP | Response / modelo |
| --- | --- |
| 201 | application/json: InvoiceOccurrenceMoveHistoryResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 409 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/financial-items/{item_id}/occurrences/{reference_month}/invoice-moves`

Operação: `getFinancialItemOccurrenceInvoiceMoveHistory`. Consultar o histórico de movimentações da ocorrência

**Divergência/atenção:** Atual: {data: InvoiceOccurrenceMove[]}, sem current_allocation. A extensão aditiva é proposta em D-03.

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `item_id` | path | sim | `string (uuid)` | — Identificador do item pertencente ao usuário autenticado. |
| `reference_month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Competência que identifica a ocorrência do item. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: InvoiceOccurrenceMoveHistoryResponse |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

### GET `/v1/plans/current/months/{month}/summary`

Operação: `getMonthlySummary`. Calcular o resumo projetado de uma competência

| Parâmetro | Local | Obrigatório | Tipo | Regras |
| --- | --- | --- | --- | --- |
| `month` | path | sim | `YearMonth` | pattern="^[0-9]{4}-(0[1-9] / 1[0-2])$" Competência mensal no formato `AAAA-MM`. |
| `basis` | query | não | `SummaryBasis` | enum=["cash", "reference"]; default="cash" Base de consolidação. O padrão é `cash`. |

Request: sem body.

| HTTP | Response / modelo |
| --- | --- |
| 200 | application/json: MonthlySummaryResponse |
| 400 | application/json: Error |
| 401 | application/json: Error |
| 404 | application/json: Error |
| 422 | application/json: Error |
| 500 | application/json: Error |
| 503 | application/json: Error |

## Modelos de payload e response

Nomes das tabelas definidos abaixo, com constraints e required preservados. Referências locais são aos modelos deste catálogo/OpenAPI. Consulte o design para os modelos corrigidos/propostos.

### Model: AmountCents

```json
{
  "type": "integer",
  "format": "int64",
  "minimum": 0,
  "description": "Valor monetário não negativo representado em centavos.",
  "examples": [
    120000
  ]
}
```

### Model: ArchiveFinancialItemRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "effective_from": {
      "description": "Primeira competência na qual o item deixa de gerar ocorrências. Se\nausente, apenas o status atual é arquivado.\n",
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: ArchiveResourceRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CardInvoice

```json
{
  "unevaluatedProperties": false,
  "allOf": [
    {
      "$ref": "#/components/schemas/CardInvoiceSummary"
    },
    {
      "type": "object",
      "required": [
        "components"
      ],
      "properties": {
        "components": {
          "type": "array",
          "items": {
            "$ref": "#/components/schemas/CardInvoiceComponent"
          }
        }
      }
    }
  ]
}
```

### Model: CardInvoiceAdjustment

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "currency_code",
    "credit_card_id",
    "payment_month",
    "reference_month",
    "name",
    "amount_cents",
    "context",
    "status",
    "archived_at",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "credit_card_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "reference_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "status": {
      "$ref": "#/components/schemas/InvoiceAdjustmentStatus"
    },
    "archived_at": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: CardInvoiceAdjustmentResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/CardInvoiceAdjustment"
    }
  }
}
```

### Model: CardInvoiceComponent

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "source_id",
    "source_type",
    "item_id",
    "adjustment_id",
    "name",
    "reference_month",
    "reference_known",
    "payment_month",
    "amount_cents",
    "payment_date",
    "is_projected",
    "allocation"
  ],
  "properties": {
    "source_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "source_type": {
      "$ref": "#/components/schemas/CardInvoiceComponentType"
    },
    "item_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "adjustment_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "reference_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "reference_known": {
      "type": "boolean"
    },
    "payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "amount_cents": {
      "$ref": "#/components/schemas/SignedAmountCents"
    },
    "payment_date": {
      "type": [
        "string",
        "null"
      ],
      "format": "date"
    },
    "is_projected": {
      "type": "boolean",
      "description": "Parcela futura inferida de um CSV; substituída quando a parcela real é importada."
    },
    "allocation": {
      "$ref": "#/components/schemas/InvoiceAllocationOrigin"
    },
    "debt_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "installment_number": {
      "type": [
        "integer",
        "null"
      ],
      "format": "int32",
      "minimum": 1
    },
    "installments_total": {
      "type": [
        "integer",
        "null"
      ],
      "format": "int32",
      "minimum": 1
    },
    "debt_occurrence_kind": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/DebtOccurrenceKind"
        },
        {
          "type": "null"
        }
      ]
    }
  }
}
```

### Model: CardInvoiceComponentType

```json
{
  "type": "string",
  "enum": [
    "financial_item_occurrence",
    "invoice_adjustment",
    "invoice_payment"
  ]
}
```

### Model: CardInvoiceListResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data",
    "range"
  ],
  "properties": {
    "data": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/CardInvoiceSummary"
      }
    },
    "range": {
      "$ref": "#/components/schemas/MonthRange"
    }
  }
}
```

### Model: CardInvoiceResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/CardInvoice"
    }
  }
}
```

### Model: CardInvoiceSummary

```json
{
  "type": "object",
  "required": [
    "card_id",
    "card_name",
    "institution",
    "payment_month",
    "nominal_due_day",
    "nominal_due_date",
    "nominal_due_date_resolution",
    "currency_code",
    "projected_total_cents",
    "charges_total_cents",
    "payments_total_cents",
    "component_count"
  ],
  "properties": {
    "card_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "card_name": {
      "type": "string"
    },
    "institution": {
      "$ref": "#/components/schemas/FinancialInstitution"
    },
    "payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "nominal_due_day": {
      "type": "integer",
      "format": "int32",
      "minimum": 1,
      "maximum": 31
    },
    "nominal_due_date": {
      "type": [
        "string",
        "null"
      ],
      "format": "date"
    },
    "nominal_due_date_resolution": {
      "$ref": "#/components/schemas/NominalDueDateResolution"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "projected_total_cents": {
      "$ref": "#/components/schemas/SignedAmountCents"
    },
    "charges_total_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "payments_total_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "component_count": {
      "type": "integer",
      "format": "int32",
      "minimum": 0
    }
  }
}
```

### Model: Completeness

```json
{
  "type": "string",
  "enum": [
    "projected"
  ]
}
```

### Model: CreatableFinancialItemKind

```json
{
  "type": "string",
  "description": "Tipos aceitos pela criação genérica de itens financeiros.",
  "enum": [
    "recurring_income",
    "one_time_income",
    "fixed_expense",
    "projected_variable_expense"
  ]
}
```

### Model: CreateCardInvoiceAdjustmentRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name",
    "amount_cents",
    "reference_month"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "reference_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreateCreditCardConfigurationChangeRequest

```json
{
  "$ref": "#/components/schemas/CreditCardConfigurationInput"
}
```

### Model: CreateCreditCardRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "institution_id",
    "name",
    "configuration"
  ],
  "properties": {
    "institution_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "configuration": {
      "$ref": "#/components/schemas/CreditCardConfigurationInput"
    }
  }
}
```

### Model: CreateDebtChangeRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "effective_from",
    "installment_amount_cents"
  ],
  "properties": {
    "effective_from": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "installment_amount_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreateDebtEarlySettlementRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "reference_month",
    "amount_cents"
  ],
  "properties": {
    "reference_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "amount_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreateDebtRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name",
    "original_total_cents",
    "total_installments",
    "first_projected_installment",
    "scheduled_start_month",
    "installment_amount_cents"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "description": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 1000
    },
    "original_total_cents": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/AmountCents"
        },
        {
          "type": "null"
        }
      ],
      "description": "Informativo; não é usado para inferir saldo ou parcela."
    },
    "total_installments": {
      "type": "integer",
      "format": "int32",
      "minimum": 1
    },
    "first_projected_installment": {
      "type": "integer",
      "format": "int32",
      "minimum": 1
    },
    "scheduled_start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "installment_amount_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "cash_month_offset": {
      "type": "integer",
      "format": "int32",
      "minimum": 0,
      "maximum": 12,
      "default": 0
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "payment_method": {
      "allOf": [
        {
          "$ref": "#/components/schemas/PaymentMethod"
        }
      ],
      "default": "direct"
    },
    "credit_card_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    }
  }
}
```

### Model: CreateFinancialInstitutionRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    }
  }
}
```

### Model: CreateFinancialItemChangeRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "effective_from",
    "end_month",
    "amount_cents",
    "cash_month_offset"
  ],
  "properties": {
    "effective_from": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "cash_month_offset": {
      "type": "integer",
      "format": "int32",
      "minimum": 0,
      "maximum": 12
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreateFinancialItemRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name",
    "kind",
    "period"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "kind": {
      "$ref": "#/components/schemas/CreatableFinancialItemKind"
    },
    "invoice_match_title": {
      "type": [
        "string",
        "null"
      ],
      "minLength": 1,
      "maxLength": 120,
      "description": "Descrição completa do campo title do CSV. Somente despesa fixa; compara ignorando caixa e espaços externos. null desativa novos vínculos."
    },
    "renewal_day": {
      "type": [
        "integer",
        "null"
      ],
      "minimum": 1,
      "maximum": 31,
      "description": "Dia nominal de renovação da despesa fixa; não determina o ciclo da fatura."
    },
    "description": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 1000
    },
    "period": {
      "$ref": "#/components/schemas/FinancialPeriodInput"
    }
  }
}
```

### Model: CreateInvoiceOccurrenceMoveRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "target_credit_card_id",
    "target_payment_month"
  ],
  "properties": {
    "target_credit_card_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "target_payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreatePaymentMethodChangeRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "effective_from",
    "end_month",
    "method",
    "credit_card_id"
  ],
  "properties": {
    "effective_from": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "method": {
      "$ref": "#/components/schemas/PaymentMethod"
    },
    "credit_card_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreatePlanRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name",
    "start_month",
    "end_month",
    "currency_code"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    }
  }
}
```

### Model: CreditCard

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "institution",
    "name",
    "status",
    "archived_at",
    "configurations",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "institution": {
      "$ref": "#/components/schemas/FinancialInstitution"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "status": {
      "$ref": "#/components/schemas/FinancialResourceStatus"
    },
    "archived_at": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    },
    "configurations": {
      "type": "array",
      "minItems": 1,
      "items": {
        "$ref": "#/components/schemas/CreditCardConfigurationPeriod"
      }
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: CreditCardConfigurationInput

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "effective_from",
    "end_month",
    "nominal_due_day",
    "payment_month_offset"
  ],
  "properties": {
    "effective_from": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "nominal_due_day": {
      "type": "integer",
      "format": "int32",
      "minimum": 1,
      "maximum": 31
    },
    "payment_month_offset": {
      "type": "integer",
      "format": "int32",
      "minimum": 0,
      "maximum": 12,
      "default": 1
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: CreditCardConfigurationPeriod

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "start_month",
    "end_month",
    "nominal_due_day",
    "payment_month_offset",
    "context",
    "recorded_at",
    "created_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "nominal_due_day": {
      "type": "integer",
      "format": "int32",
      "minimum": 1,
      "maximum": 31
    },
    "payment_month_offset": {
      "type": "integer",
      "format": "int32",
      "minimum": 0,
      "maximum": 12
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "recorded_at": {
      "type": "string",
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: CreditCardListResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/CreditCard"
      }
    }
  }
}
```

### Model: CreditCardResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/CreditCard"
    }
  }
}
```

### Model: CurrencyCode

```json
{
  "type": "string",
  "pattern": "^[A-Z]{3}$",
  "examples": [
    "BRL"
  ]
}
```

### Model: CurrentUser

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "email",
    "display_name",
    "timezone",
    "currency_code"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "email": {
      "type": "string",
      "format": "email"
    },
    "display_name": {
      "type": [
        "string",
        "null"
      ]
    },
    "timezone": {
      "type": "string",
      "examples": [
        "America/Sao_Paulo"
      ]
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    }
  }
}
```

### Model: CurrentUserResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/CurrentUser"
    }
  }
}
```

### Model: Debt

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "currency_code",
    "name",
    "description",
    "original_total_cents",
    "total_installments",
    "first_projected_installment",
    "scheduled_start_month",
    "scheduled_end_month",
    "effective_end_month",
    "release_from_month",
    "released_monthly_cents",
    "as_of_month",
    "projection_status",
    "remaining_installments",
    "status",
    "settlement",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "description": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 1000
    },
    "original_total_cents": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/AmountCents"
        },
        {
          "type": "null"
        }
      ]
    },
    "total_installments": {
      "type": "integer",
      "format": "int32",
      "minimum": 1
    },
    "first_projected_installment": {
      "type": "integer",
      "format": "int32",
      "minimum": 1
    },
    "scheduled_start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "scheduled_end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "effective_end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "release_from_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "released_monthly_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "as_of_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "projection_status": {
      "$ref": "#/components/schemas/DebtProjectionStatus"
    },
    "remaining_installments": {
      "type": "integer",
      "format": "int32",
      "minimum": 0
    },
    "status": {
      "$ref": "#/components/schemas/FinancialItemStatus"
    },
    "settlement": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/DebtEarlySettlement"
        },
        {
          "type": "null"
        }
      ]
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: DebtEarlySettlement

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "debt_id",
    "reference_month",
    "amount_cents",
    "reason",
    "recorded_by",
    "recorded_at",
    "created_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "debt_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "reference_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "amount_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "recorded_by": {
      "$ref": "#/components/schemas/UUID"
    },
    "recorded_at": {
      "type": "string",
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: DebtEarlySettlementResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/DebtEarlySettlement"
    }
  }
}
```

### Model: DebtListResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/Debt"
      }
    }
  }
}
```

### Model: DebtOccurrence

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "debt_id",
    "source_id",
    "reference_month",
    "installment_number",
    "installments_total",
    "amount_cents",
    "debt_occurrence_kind",
    "payment_method",
    "cash_month",
    "credit_card_id",
    "invoice_payment_month",
    "completeness"
  ],
  "properties": {
    "debt_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "source_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "reference_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "installment_number": {
      "type": "integer",
      "format": "int32",
      "minimum": 1
    },
    "installments_total": {
      "type": "integer",
      "format": "int32",
      "minimum": 1
    },
    "amount_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "debt_occurrence_kind": {
      "$ref": "#/components/schemas/DebtOccurrenceKind"
    },
    "payment_method": {
      "$ref": "#/components/schemas/PaymentMethod"
    },
    "cash_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "credit_card_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "invoice_payment_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "completeness": {
      "$ref": "#/components/schemas/Completeness"
    }
  }
}
```

### Model: DebtOccurrenceKind

```json
{
  "type": "string",
  "enum": [
    "scheduled",
    "early_settlement"
  ]
}
```

### Model: DebtProjectionStatus

```json
{
  "type": "string",
  "enum": [
    "planned",
    "active",
    "completed",
    "settled_early",
    "archived"
  ]
}
```

### Model: DebtRelease

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "debt_id",
    "name",
    "currency_code",
    "scheduled_end_month",
    "effective_end_month",
    "release_from_month",
    "released_monthly_cents",
    "reason"
  ],
  "properties": {
    "debt_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "name": {
      "type": "string"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "scheduled_end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "effective_end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "release_from_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "released_monthly_cents": {
      "$ref": "#/components/schemas/PositiveAmountCents"
    },
    "reason": {
      "$ref": "#/components/schemas/DebtReleaseReason"
    }
  }
}
```

### Model: DebtReleaseList

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "releases",
    "monthly_totals"
  ],
  "properties": {
    "releases": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/DebtRelease"
      }
    },
    "monthly_totals": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/DebtReleaseMonthlyTotal"
      }
    }
  }
}
```

### Model: DebtReleaseListResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data",
    "range"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/DebtReleaseList"
    },
    "range": {
      "$ref": "#/components/schemas/MonthRange"
    }
  }
}
```

### Model: DebtReleaseMonthlyTotal

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "month",
    "released_monthly_cents"
  ],
  "properties": {
    "month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "released_monthly_cents": {
      "$ref": "#/components/schemas/AmountCents"
    }
  }
}
```

### Model: DebtReleaseReason

```json
{
  "type": "string",
  "enum": [
    "scheduled_completion",
    "early_settlement"
  ]
}
```

### Model: DebtResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/Debt"
    }
  }
}
```

### Model: DebtSchedule

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "debt",
    "range",
    "occurrences"
  ],
  "properties": {
    "debt": {
      "$ref": "#/components/schemas/Debt"
    },
    "range": {
      "$ref": "#/components/schemas/MonthRange"
    },
    "occurrences": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/DebtOccurrence"
      }
    }
  }
}
```

### Model: DebtScheduleResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/DebtSchedule"
    }
  }
}
```

### Model: FinancialInstitution

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "name",
    "status",
    "archived_at",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "status": {
      "$ref": "#/components/schemas/FinancialResourceStatus"
    },
    "archived_at": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: FinancialInstitutionListResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/FinancialInstitution"
      }
    }
  }
}
```

### Model: FinancialInstitutionResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/FinancialInstitution"
    }
  }
}
```

### Model: FinancialItem

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "currency_code",
    "name",
    "kind",
    "description",
    "status",
    "archived_at",
    "periods",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "kind": {
      "$ref": "#/components/schemas/FinancialItemKind"
    },
    "invoice_match_title": {
      "type": [
        "string",
        "null"
      ],
      "minLength": 1,
      "maxLength": 120,
      "description": "Descrição completa do campo title do CSV. Somente despesa fixa; compara ignorando caixa e espaços externos. null desativa novos vínculos."
    },
    "renewal_day": {
      "type": [
        "integer",
        "null"
      ],
      "minimum": 1,
      "maximum": 31,
      "description": "Dia nominal de renovação da despesa fixa; não determina o ciclo da fatura."
    },
    "description": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 1000
    },
    "status": {
      "$ref": "#/components/schemas/FinancialItemStatus"
    },
    "archived_at": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    },
    "periods": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/FinancialPeriod"
      }
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: FinancialItemKind

```json
{
  "type": "string",
  "enum": [
    "recurring_income",
    "one_time_income",
    "fixed_expense",
    "projected_variable_expense",
    "debt_installment"
  ]
}
```

### Model: FinancialItemListResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/FinancialItem"
      }
    }
  }
}
```

### Model: FinancialItemResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/FinancialItem"
    }
  }
}
```

### Model: FinancialItemStatus

```json
{
  "type": "string",
  "enum": [
    "active",
    "archived"
  ]
}
```

### Model: FinancialPeriod

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "start_month",
    "end_month",
    "amount_cents",
    "recurrence",
    "cash_month_offset",
    "context",
    "recorded_at",
    "created_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "recurrence": {
      "$ref": "#/components/schemas/Recurrence"
    },
    "cash_month_offset": {
      "type": "integer",
      "format": "int32",
      "minimum": 0,
      "maximum": 12
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "recorded_at": {
      "type": "string",
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: FinancialPeriodInput

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "start_month",
    "end_month",
    "amount_cents",
    "recurrence",
    "cash_month_offset"
  ],
  "properties": {
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "recurrence": {
      "$ref": "#/components/schemas/Recurrence"
    },
    "cash_month_offset": {
      "type": "integer",
      "format": "int32",
      "minimum": 0,
      "maximum": 12,
      "default": 0
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: FinancialResourceStatus

```json
{
  "type": "string",
  "enum": [
    "active",
    "archived"
  ]
}
```

### Model: InvoiceAdjustmentStatus

```json
{
  "type": "string",
  "enum": [
    "active",
    "archived"
  ]
}
```

### Model: InvoiceAllocation

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "credit_card_id",
    "payment_month",
    "origin"
  ],
  "properties": {
    "credit_card_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "origin": {
      "$ref": "#/components/schemas/InvoiceAllocationOrigin"
    }
  }
}
```

### Model: InvoiceAllocationOrigin

```json
{
  "type": "string",
  "enum": [
    "calculated_from_reference",
    "selected_invoice",
    "moved_by_user"
  ]
}
```

### Model: InvoiceImportResponse

```json
{
  "type": "object",
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "type": "object",
      "required": [
        "card_id",
        "payment_month",
        "source",
        "rows",
        "projected_installments",
        "new_projected_count",
        "new_count",
        "existing_count",
        "skipped_count",
        "needs_invoice_count",
        "new_total_cents",
        "new_charges_cents",
        "new_payments_cents",
        "new_payment_count",
        "preview_token"
      ],
      "properties": {
        "card_id": {
          "$ref": "#/components/schemas/UUID"
        },
        "payment_month": {
          "$ref": "#/components/schemas/YearMonth"
        },
        "source": {
          "type": "string",
          "const": "nubank_csv"
        },
        "rows": {
          "type": "array",
          "items": {
            "type": "object",
            "required": [
              "line",
              "date",
              "title",
              "amount_cents",
              "kind",
              "status"
            ],
            "properties": {
              "line": {
                "type": "integer"
              },
              "date": {
                "type": "string",
                "format": "date"
              },
              "title": {
                "type": "string"
              },
              "amount_cents": {
                "type": "integer",
                "format": "int64"
              },
              "kind": {
                "type": "string",
                "enum": [
                  "expense",
                  "payment",
                  "unsupported"
                ]
              },
              "invoice_payment_month": {
                "$ref": "#/components/schemas/YearMonth"
              },
              "installment_number": {
                "type": "integer"
              },
              "installments_total": {
                "type": "integer"
              },
              "status": {
                "type": "string",
                "enum": [
                  "new",
                  "existing",
                  "skipped",
                  "needs_invoice",
                  "needs_subscription_match",
                  "imported"
                ]
              },
              "matched_item_id": {
                "$ref": "#/components/schemas/UUID"
              },
              "matched_reference_month": {
                "$ref": "#/components/schemas/YearMonth"
              },
              "planned_amount_cents": {
                "$ref": "#/components/schemas/AmountCents"
              },
              "reason": {
                "type": "string",
                "enum": [
                  "non_expense",
                  "already_imported_to_other_invoice",
                  "projected_installment",
                  "subscription_match",
                  "ambiguous_subscription"
                ]
              }
            }
          }
        },
        "projected_installments": {
          "type": "array",
          "description": "Parcelas futuras inferidas das linhas do CSV, sem data de transação conhecida.",
          "items": {
            "type": "object",
            "required": [
              "source_line",
              "title",
              "amount_cents",
              "invoice_payment_month",
              "installment_number",
              "installments_total",
              "status"
            ],
            "properties": {
              "source_line": {
                "type": "integer"
              },
              "title": {
                "type": "string"
              },
              "amount_cents": {
                "$ref": "#/components/schemas/AmountCents"
              },
              "invoice_payment_month": {
                "$ref": "#/components/schemas/YearMonth"
              },
              "installment_number": {
                "type": "integer"
              },
              "installments_total": {
                "type": "integer"
              },
              "status": {
                "type": "string",
                "enum": [
                  "new",
                  "existing",
                  "imported"
                ]
              }
            }
          }
        },
        "new_projected_count": {
          "type": "integer",
          "description": "Número de parcelas futuras novas; não incluído em `new_total_cents`."
        },
        "new_count": {
          "type": "integer"
        },
        "existing_count": {
          "type": "integer"
        },
        "skipped_count": {
          "type": "integer"
        },
        "needs_subscription_match_count": {
          "type": "integer",
          "description": "Linhas ambíguas que impedem a confirmação até corrigir cadastro ou arquivo."
        },
        "needs_invoice_count": {
          "type": "integer"
        },
        "new_total_cents": {
          "$ref": "#/components/schemas/SignedAmountCents"
        },
        "new_charges_cents": {
          "$ref": "#/components/schemas/AmountCents"
        },
        "new_payments_cents": {
          "$ref": "#/components/schemas/AmountCents"
        },
        "new_payment_count": {
          "type": "integer"
        },
        "imported_count": {
          "type": "integer",
          "description": "Presente quando houve novas inserções na confirmação."
        },
        "preview_token": {
          "type": "string",
          "description": "Envie na confirmação junto ao mesmo CSV."
        }
      }
    }
  }
}
```

### Model: InvoiceOccurrenceMove

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "item_id",
    "reference_month",
    "from_credit_card_id",
    "from_payment_month",
    "to_credit_card_id",
    "to_payment_month",
    "reason",
    "recorded_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "item_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "reference_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "from_credit_card_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "from_payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "to_credit_card_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "to_payment_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "recorded_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: InvoiceOccurrenceMoveHistory

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "item_id",
    "reference_month",
    "current_allocation",
    "events"
  ],
  "properties": {
    "item_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "reference_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "current_allocation": {
      "$ref": "#/components/schemas/InvoiceAllocation"
    },
    "events": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/InvoiceOccurrenceMove"
      }
    }
  }
}
```

### Model: InvoiceOccurrenceMoveHistoryResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/InvoiceOccurrenceMoveHistory"
    }
  }
}
```

### Model: MonthRange

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "from",
    "to"
  ],
  "properties": {
    "from": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "to": {
      "$ref": "#/components/schemas/YearMonth"
    }
  }
}
```

### Model: MonthlySummary

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "month",
    "basis",
    "result_kind",
    "currency_code",
    "plan_status",
    "income_cents",
    "commitments_cents",
    "planned_savings_cents",
    "result_cents",
    "is_negative",
    "completeness",
    "breakdown",
    "sources"
  ],
  "properties": {
    "month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "basis": {
      "$ref": "#/components/schemas/SummaryBasis"
    },
    "result_kind": {
      "$ref": "#/components/schemas/SummaryResultKind"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "plan_status": {
      "$ref": "#/components/schemas/PlanStatus"
    },
    "income_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "commitments_cents": {
      "$ref": "#/components/schemas/SignedAmountCents"
    },
    "planned_savings_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "result_cents": {
      "$ref": "#/components/schemas/SignedAmountCents"
    },
    "is_negative": {
      "type": "boolean"
    },
    "completeness": {
      "$ref": "#/components/schemas/Completeness"
    },
    "breakdown": {
      "$ref": "#/components/schemas/MonthlySummaryBreakdown"
    },
    "sources": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/MonthlySummarySource"
      }
    }
  }
}
```

### Model: MonthlySummaryBreakdown

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "recurring_income_cents",
    "one_time_income_cents",
    "fixed_expenses_cents",
    "projected_variable_expenses_cents",
    "debt_installments_cents",
    "card_invoice_adjustments_cents",
    "card_invoice_payments_cents"
  ],
  "properties": {
    "recurring_income_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "one_time_income_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "fixed_expenses_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "projected_variable_expenses_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "debt_installments_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "card_invoice_adjustments_cents": {
      "$ref": "#/components/schemas/SignedAmountCents"
    },
    "card_invoice_payments_cents": {
      "$ref": "#/components/schemas/AmountCents"
    }
  }
}
```

### Model: MonthlySummaryResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/MonthlySummary"
    }
  }
}
```

### Model: MonthlySummarySource

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "source_id",
    "item_id",
    "name",
    "kind",
    "effect",
    "reference_month",
    "cash_month",
    "amount_cents",
    "is_projected"
  ],
  "properties": {
    "source_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "item_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "name": {
      "type": "string"
    },
    "kind": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/FinancialItemKind"
        },
        {
          "type": "string",
          "const": "planned_savings"
        },
        {
          "type": "string",
          "const": "card_invoice_adjustment"
        },
        {
          "type": "string",
          "const": "card_invoice_payment"
        }
      ]
    },
    "effect": {
      "$ref": "#/components/schemas/SummarySourceEffect"
    },
    "reference_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "cash_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "amount_cents": {
      "$ref": "#/components/schemas/SignedAmountCents"
    },
    "is_projected": {
      "type": "boolean",
      "description": "Identifica uma parcela futura inferida de um CSV Nubank."
    },
    "source_type": {
      "$ref": "#/components/schemas/SummarySourceType"
    },
    "payment_method": {
      "$ref": "#/components/schemas/PaymentMethod"
    },
    "credit_card_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "invoice_payment_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "invoice_allocation": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/InvoiceAllocationOrigin"
        },
        {
          "type": "null"
        }
      ]
    },
    "reference_known": {
      "type": "boolean"
    },
    "debt_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "installment_number": {
      "type": [
        "integer",
        "null"
      ],
      "format": "int32",
      "minimum": 1
    },
    "installments_total": {
      "type": [
        "integer",
        "null"
      ],
      "format": "int32",
      "minimum": 1
    },
    "debt_occurrence_kind": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/DebtOccurrenceKind"
        },
        {
          "type": "null"
        }
      ]
    }
  }
}
```

### Model: NominalDueDateResolution

```json
{
  "type": "string",
  "enum": [
    "exact",
    "invalid_for_month"
  ]
}
```

### Model: OriginalPlanSnapshot

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "plan_id",
    "kind",
    "schema_version",
    "captured_at",
    "plan"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "plan_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "kind": {
      "type": "string",
      "const": "original"
    },
    "schema_version": {
      "type": "integer",
      "minimum": 1
    },
    "captured_at": {
      "type": "string",
      "format": "date-time"
    },
    "plan": {
      "$ref": "#/components/schemas/SnapshotPlanData"
    }
  }
}
```

### Model: OriginalPlanSnapshotResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/OriginalPlanSnapshot"
    }
  }
}
```

### Model: PaymentMethod

```json
{
  "type": "string",
  "enum": [
    "direct",
    "credit_card"
  ]
}
```

### Model: PaymentMethodHistory

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "financial_item_id",
    "default_method",
    "periods"
  ],
  "properties": {
    "financial_item_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "default_method": {
      "type": "string",
      "const": "direct"
    },
    "periods": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/PaymentMethodPeriod"
      }
    }
  }
}
```

### Model: PaymentMethodHistoryResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/PaymentMethodHistory"
    }
  }
}
```

### Model: PaymentMethodPeriod

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "financial_item_id",
    "start_month",
    "end_month",
    "method",
    "credit_card_id",
    "context",
    "recorded_at",
    "created_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "financial_item_id": {
      "$ref": "#/components/schemas/UUID"
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "method": {
      "$ref": "#/components/schemas/PaymentMethod"
    },
    "credit_card_id": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/UUID"
        },
        {
          "type": "null"
        }
      ]
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "recorded_at": {
      "type": "string",
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: Plan

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "name",
    "status",
    "start_month",
    "end_month",
    "currency_code",
    "activated_at",
    "archived_at",
    "created_at",
    "updated_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "status": {
      "$ref": "#/components/schemas/PlanStatus"
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "activated_at": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    },
    "archived_at": {
      "type": [
        "string",
        "null"
      ],
      "format": "date-time"
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    },
    "updated_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: PlanActivation

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "plan",
    "original"
  ],
  "properties": {
    "plan": {
      "$ref": "#/components/schemas/Plan"
    },
    "original": {
      "$ref": "#/components/schemas/OriginalPlanSnapshot"
    }
  }
}
```

### Model: PlanActivationResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/PlanActivation"
    }
  }
}
```

### Model: PlanResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/Plan"
    }
  }
}
```

### Model: PlanStatus

```json
{
  "type": "string",
  "enum": [
    "draft",
    "active",
    "archived"
  ]
}
```

### Model: PlannedSavings

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "configured",
    "periods"
  ],
  "properties": {
    "configured": {
      "type": "boolean",
      "description": "Diferencia ausência de configuração de um valor zero explícito."
    },
    "periods": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/SavingPeriod"
      }
    }
  }
}
```

### Model: PlannedSavingsResponse

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "data"
  ],
  "properties": {
    "data": {
      "$ref": "#/components/schemas/PlannedSavings"
    }
  }
}
```

### Model: PositiveAmountCents

```json
{
  "type": "integer",
  "format": "int64",
  "minimum": 1,
  "description": "Valor monetário positivo representado em centavos.",
  "examples": [
    60000
  ]
}
```

### Model: PutPlannedSavingsRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "effective_from",
    "end_month",
    "amount_cents"
  ],
  "properties": {
    "effective_from": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: Recurrence

```json
{
  "type": "string",
  "enum": [
    "monthly",
    "once"
  ]
}
```

### Model: SavingPeriod

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "start_month",
    "end_month",
    "amount_cents",
    "context",
    "created_at"
  ],
  "properties": {
    "id": {
      "$ref": "#/components/schemas/UUID"
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "created_at": {
      "type": "string",
      "format": "date-time"
    }
  }
}
```

### Model: SignedAmountCents

```json
{
  "type": "integer",
  "format": "int64",
  "description": "Resultado monetário em centavos; pode ser negativo.",
  "examples": [
    200000,
    -50000
  ]
}
```

### Model: SnapshotPlanData

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name",
    "start_month",
    "end_month",
    "currency_code",
    "items",
    "savings"
  ],
  "properties": {
    "name": {
      "type": "string"
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    },
    "items": {
      "type": "array",
      "items": {
        "$ref": "#/components/schemas/FinancialItem"
      }
    },
    "savings": {
      "$ref": "#/components/schemas/PlannedSavings"
    }
  }
}
```

### Model: Status

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "status"
  ],
  "properties": {
    "status": {
      "type": "string"
    }
  }
}
```

### Model: SummaryBasis

```json
{
  "type": "string",
  "enum": [
    "cash",
    "reference"
  ],
  "default": "cash"
}
```

### Model: SummaryResultKind

```json
{
  "type": "string",
  "enum": [
    "planned_free",
    "planned_reference_result"
  ]
}
```

### Model: SummarySourceEffect

```json
{
  "type": "string",
  "enum": [
    "income",
    "commitment",
    "savings"
  ]
}
```

### Model: SummarySourceType

```json
{
  "type": "string",
  "enum": [
    "financial_item_occurrence",
    "planned_savings",
    "invoice_adjustment",
    "invoice_payment",
    "invoice_payment_cash"
  ]
}
```

### Model: UUID

```json
{
  "type": "string",
  "format": "uuid"
}
```

### Model: UpdateCardInvoiceAdjustmentRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "amount_cents": {
      "$ref": "#/components/schemas/AmountCents"
    },
    "reference_month": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/YearMonth"
        },
        {
          "type": "null"
        }
      ]
    },
    "context": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    },
    "reason": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 500
    }
  }
}
```

### Model: UpdateCreditCardRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    }
  }
}
```

### Model: UpdateDebtRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "description": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 1000
    }
  }
}
```

### Model: UpdateFinancialInstitutionRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "name"
  ],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    }
  }
}
```

### Model: UpdateFinancialItemRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "invoice_match_title": {
      "type": [
        "string",
        "null"
      ],
      "minLength": 1,
      "maxLength": 120,
      "description": "Descrição completa do campo title do CSV. Somente despesa fixa; compara ignorando caixa e espaços externos. null desativa novos vínculos."
    },
    "renewal_day": {
      "type": [
        "integer",
        "null"
      ],
      "minimum": 1,
      "maximum": 31,
      "description": "Dia nominal de renovação da despesa fixa; não determina o ciclo da fatura."
    },
    "description": {
      "type": [
        "string",
        "null"
      ],
      "maxLength": 1000
    }
  }
}
```

### Model: UpdatePlanRequest

```json
{
  "type": "object",
  "additionalProperties": false,
  "minProperties": 1,
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1,
      "maxLength": 120
    },
    "start_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "end_month": {
      "$ref": "#/components/schemas/YearMonth"
    },
    "currency_code": {
      "$ref": "#/components/schemas/CurrencyCode"
    }
  }
}
```

### Model: YearMonth

```json
{
  "type": "string",
  "pattern": "^[0-9]{4}-(0[1-9]|1[0-2])$",
  "examples": [
    "2026-07"
  ]
}
```

## Modelos compartilhados adicionais

Inclui erros, saúde e definições referenciadas por respostas comuns.

### Model: Error

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "error"
  ],
  "properties": {
    "error": {
      "$ref": "#/components/schemas/ErrorBody"
    }
  }
}
```

### Model: ErrorBody

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": [
    "code",
    "message"
  ],
  "properties": {
    "code": {
      "type": "string"
    },
    "message": {
      "type": "string"
    },
    "details": {
      "oneOf": [
        {
          "$ref": "#/components/schemas/ErrorDetail"
        },
        {
          "type": "null"
        }
      ]
    }
  }
}
```

### Model: ErrorDetail

```json
{
  "type": "object",
  "additionalProperties": true
}
```
