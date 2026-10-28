# Frontend 0–3.1 — Design de interfaces e integração

**Data:** 04/10/2026. **Estado:** design proposto, não implementado.
Fontes: [spec](../../../IA/financial-api/Docs/frontend/spec-cobertura-api-0-3.1.md), [catálogo completo de contratos](integration-contracts.md), OpenAPI e handlers Go atuais. Execução: [tasks](tasks.md).

## 1. Decisões de arquitetura

**D-01 — Transporte.** Manter Next.js App Router e Luke UI. Server Components consultam a API para o carregamento inicial. Interações do browser usam `/api/financial/[...path]`, que encaminha método, query, body e Content-Type para `FINANCIAL_API_URL + /v1/...`. O token vem da sessão do servidor; a interface não envia user_id nem acessa tabelas financeiras do Supabase. `financialApiFetch` precisa oferecer acesso ao status HTTP para que o proxy preserve 200/201 e códigos de erro, sem quebrar os helpers que retornam JSON. Multipart não recebe um Content-Type manual sem boundary.

**D-02 — Organização.** Reutilizar `lib/api/financial-api.ts`, `browser-api.ts`, `types.ts` e componentes de domínio. Extrair modelos e adaptadores por domínio se o arquivo de tipos crescer; não criar um segundo cliente concorrente. Separar DTO da API, view model da tela e valores de formulário. Regras de projeção, conciliação, alocação e totalização oficial pertencem à API; helpers locais só tratam datas, janelas, formatação e seleção de vigência para apresentação.

**D-03 — Históricos.** Integrar a resposta atual em array sem quebra de contrato. O view model de pagamento é `{itemId, defaultMethod: 'direct', periods}` montado pelo adaptador, não um objeto recebido da API. Movimentações são eventos em `data`; não há alocação atual nesse handler. Corrigir OpenAPI para documentar arrays. Acrescentar `current_allocation` como campo irmão de `data` no GET de movimentações, mantendo o array e resolvendo alocação pela lógica existente da API. Essa extensão é proposta e bloqueia o painel completo de alocação atual até FE-T04. Não inferir alocação atual apenas pelo último evento quando não houver eventos ou houver configuração temporal relevante.

**D-04 — Snapshots.** Ler documento congelado por schema_version. A versão 4 real contém `plan.release_2` e `plan.release_3`; não buscar recursos atuais para preencher o snapshot. Schemas 1–3 precisam de fixtures reais/sintéticas fiéis às versões, extraídas das migrations/tests. Coleções históricas podem conter datas completas de banco (`YYYY-MM-DD`), IDs/campos diferentes das representações operacionais e ausência de propriedades recentes. Normalizar só no view model, preservando o documento original. Versão desconhecida usa visualização estruturada somente leitura.

**D-05 — Ajustes.** Criar GET autenticado `/v1/invoice-adjustments/{adjustment_id}` para leitura por identidade independente da alocação exibida. Implementar PATCH parcial na rota existente de manutenção e validação de que cartão/mês da URL correspondem à localização persistida. Campos omitidos preservam valor; `null` limpa somente referência/contexto; nome/valor não aceitam null. Essa é uma correção/extensão proposta, ainda não disponível. Não liberar editor completo antes de FE-T02/03.

**D-06 — Estado e atualização.** Consultas usam `loading | success | empty | error | unauthorized | not_found`. Cada formulário mantém valores, erros e `idle | submitting | success | error`; criação de assinatura admite `created_pending_payment`. Mutação bem-sucedida recarrega recursos afetados. Consultas têm AbortController/request ID para impedir retorno antigo sobrescrever filtros atuais. Sem retries automáticos de POST: falha de rede após envio pode significar que a mutação ocorreu; recarregar/verificar estado antes de repetir. A API não possui idempotência genérica.

## 2. Rotas de interface e carregamento

Query parameters da interface são validados antes do uso. IDs são codificados ao construir rotas. Valores `all` dos filtros representam ausência do filtro na API; não enviar `status=all`.

| Interface | Query de interface | Consultas iniciais | Ações/painéis |
| --- | --- | --- | --- |
| `/dashboard` | `month`, `basis`, `composition=income|commitments|savings` | Perfil, plano e resumo do mês | Composição por categoria e origem |
| `/planning` | `from`, `basis` | Plano, economia, resumos dentro da janela | Edição do rascunho, economia, ativação e link original |
| `/planning/original` | nenhuma | Original do plano | Seções por versão, somente leitura |
| `/entries` | `kind`, `status` | Itens e cartões | Cadastro e filtros |
| `/entries/[itemId]` | `month`, `tab=values|payments` | Item; pagamento em despesas; cartões quando necessário | Metadados, valores, método e arquivamento |
| `/cards` | `status`, `institutionStatus` | Instituições e cartões | Criação, edição de instituição e arquivamentos |
| `/cards/[cardId]` | `month` | Cartão | Nome, configuração vigente, histórico e mudança futura |
| `/invoices/[cardId]/all` | `from`, `to` | Plano, cartão e faturas do intervalo | Navegação, gráfico e importação |
| `/invoices/[cardId]/[month]` | `adjustment`, `moveItem`, `referenceMonth` | Fatura e cartões; leitura de ajuste sob demanda | Editar/arquivar cobrança; movimentar/ver histórico |
| `/debts` | `as_of`, `status`, `projection_status`, `releaseFrom`, `releaseTo` | Dívidas, cartões e liberações | Cadastro, filtros e período de liberações |
| `/debts/[debtId]` | `as_of`, `from`, `to` | Dívida; cronograma da janela; pagamentos; cartões | Reajuste, pagamento, quitação e arquivamento |

Painéis possuem estado na URL para retorno/compartilhamento, mas não dados financeiros completos, tokens, nomes ou payloads. Quando um item for `debt_installment`, encaminhar ao detalhe da dívida preservando mês de análise. GET separado de quitação só é necessário quando o detalhe não entrega settlement.

## 3. Composição das telas

### 3.1 Dashboard e planejamento

Manter cards/gráfico atuais. Clicar no total abre `MonthlyCompositionPanel`: resumo da categoria, buckets, tabela de fontes, competência, caixa e link de origem. Total oficial vem da API; soma local das fontes pode verificar a consistência, mas não substituir o total. Categorias de pagamento e ajuste respeitam sinais e classificação retornados. `breakdown` e `sources` são duas representações do mesmo valor, não parcelas aditivas entre si.

`PlanningProjection` recebe plano/limites e janela efetiva, não apenas mês inicial. A janela de cinco meses termina em `min(from + 4, plan.end_month)`. Se a entrada estiver fora do horizonte, corrigir a URL para um mês válido e explicar a seleção. Plano ativo mostra parâmetros somente leitura; não chamar PATCH para editar seu horizonte. Economia mostra lista de vigências, inclusive zero configurado.

`OriginalSnapshotView` mostra captura/versão e seções Plano, Itens, Economia, Cartões e Métodos, Ajustes e Movimentações, Dívidas e Quitações conforme os dados daquela versão. Não calcular comparação com o plano atual nesta entrega.

### 3.2 Registros e cartões

`FinancialItemDetail` usa metadados e tabelas `ValuePeriodHistory`/`PaymentPeriodHistory`. Valor destacado pertence à vigência que contém `month`; se não houver, mostrar ausência nesse mês. Períodos futuros aparecem com indicação de vigência futura. Arquivamento de item genérico possui escolha explícita: apenas arquivar status ou interromper ocorrências a partir da competência informada.

`InstitutionList` em `/cards` tem filtros e ações por linha. `CardConfigurationHistory` no detalhe mostra início/fim, dia nominal, offset, contexto e registro. `CardConfigurationForm` separa nome de configuração. Não permitir selecionar instituição arquivada para novo cartão; arquivar instituição não provoca cascata visual de arquivamento de cartões.

### 3.3 Faturas e ajustes

`InvoicePeriodNavigator` navega janelas de até 24 meses, dentro do intervalo operacional. Usar “Pagamento em mês/ano” e vencimento nominal, evitando “ciclo atual” inferido. Datas inexistentes mantêm explicação de vencimento nominal.

`InvoiceComponentRow` diferencia: ocorrência financeira, parcela de dívida, quitação, ajuste manual, cobrança CSV, pagamento recebido e parcela inferida. A origem de cobrança CSV/manual e capacidades de manutenção vêm da leitura proposta do ajuste; antes dela, não presumir origem pelo nome. Componente com `item_id` e `adjustment_id` representa substituição de ocorrência pela cobrança no projector atual; conferir detalhes do vínculo no GET de ajuste antes de habilitar manutenção.

O diálogo de edição carrega metadados reais antes de abrir o formulário. Arquivar uma cobrança conciliada exige confirmação com impacto “A previsão da assinatura voltará a compor esse mês”. Depois da resposta, recarregar fatura/resumo. Link reservado permanece preservado, sem botão “reimportar para desfazer”.

`InvoiceMoveHistoryPanel` consulta eventos pela identidade item + competência e mostra alocação atual quando a extensão estiver disponível. Datas/nomes de cartões arquivados continuam legíveis; não fabricar autor ausente no DTO.

### 3.4 Dívidas e importação

Cronograma usa janela padrão de 12 meses, seleção anterior/próximo e limite máximo de 120 meses por chamada. Seleção inicial: mês de análise limitado ao cronograma efetivo; dívida concluída abre janela final, dívida futura abre janela inicial. Não buscar 180 parcelas numa única chamada. Quantidades/resumo vêm de GET da dívida, não do tamanho da janela.

Quitação usa input month com mínimo estrutural e máximo mês anterior ao término natural, consulta/verificação pontual quando necessário e rejeição oficial da API. Não depende de `regularMonths.slice(0,-1)` da janela carregada. Com evento existente, mostrar dados somente leitura. Expor claramente que registro é projetado e imutável.

Importação conserva arquivo/revisão/resultado. FormData fica no browser; URL não contém CSV/token. Toda mudança de seleção invalida prévia. Cobranças ambíguas ou pagamentos sem fatura bloqueiam confirmar. Backfill de vínculo existente continua confirmável com zero novos lançamentos. Parcelas inferidas mostram badge “Projeção de parcela importada”, não “Dívida cadastrada”.

## 4. Contratos críticos — requests e responses reais

O [catálogo](integration-contracts.md) define todas as rotas, parâmetros, modelos, constraints e status documentados. Exemplos abaixo são sintéticos; UUIDs/tokens devem ser substituídos por valores retornados pelo ambiente. Formas abreviadas são identificadas como view models ou exemplos parciais, nunca como responses completas.

### 4.1 Criar assinatura e vincular cartão

POST `/v1/financial-items` → **201** `{data: FinancialItem}`:

```json
{
  "name": "TotalPass",
  "kind": "fixed_expense",
  "description": null,
  "invoice_match_title": "Totalpass",
  "renewal_day": 28,
  "period": {
    "start_month": "2026-10",
    "end_month": null,
    "amount_cents": 10000,
    "recurrence": "monthly",
    "cash_month_offset": 0,
    "context": "Assinatura"
  }
}
```

Salvar `data.id` antes da segunda chamada. POST `/v1/financial-items/{item_id}/payment-changes` → **201** `{data: PaymentMethodPeriod}`:

```json
{
  "effective_from": "2026-10",
  "end_month": null,
  "method": "credit_card",
  "credit_card_id": "22222222-2222-4222-8222-222222222222",
  "context": null
}
```

Para pagamento direto, usar `method: direct` e `credit_card_id: null`. Os responses da segunda chamada não são FinancialItem; refrescar item/histórico depois. PATCH do item envia somente campos modificados; limpar descrição/título/dia envia null. Mudança de valor requer `effective_from`, `end_month` explícito, `amount_cents`, `cash_month_offset` e contexto opcional; o response é **201** `{data: FinancialItem}`. Arquivamento de item retorna **200** o mesmo modelo e aceita `effective_from`/`reason` opcionais.

### 4.2 Configuração do cartão

POST `/v1/credit-cards/{card_id}/changes` → **201** `{data: CreditCard}`:

```json
{
  "effective_from": "2027-01",
  "end_month": null,
  "nominal_due_day": 6,
  "payment_month_offset": 1,
  "context": "Nova configuração"
}
```

Response contém o cartão inteiro com `institution`, status e `configurations[]`; não apenas o novo período. `end_month` deve existir mesmo quando null. GET cartão/listagem já entregam configurações. Instituições: POST `{name}` → 201; PATCH `{name}` → 200; arquivar `{reason: null}` → 200; envelopes `{data: FinancialInstitution}`. Cartão PATCH aceita somente `{name}`; criar cartão usa `{institution_id,name,configuration}` com o mesmo payload de configuração.

### 4.3 Históricos atuais e view models

GET `/v1/financial-items/{item_id}/payment-history` → **200**, exemplo completo de um período:

```json
{
  "data": [
    {
      "id": "33333333-3333-4333-8333-333333333333",
      "start_month": "2026-10",
      "end_month": null,
      "method": "credit_card",
      "credit_card_id": "22222222-2222-4222-8222-222222222222",
      "context": null,
      "recorded_at": "2026-10-04T12:00:00Z",
      "created_at": "2026-10-04T12:00:00Z"
    }
  ]
}
```

`financial_item_id` não está no elemento do handler atual. View model usa itemId da rota; ausência de períodos tem defaultMethod direct.

GET `/v1/financial-items/{item_id}/occurrences/{reference_month}/invoice-moves` → **200** `{data: InvoiceOccurrenceMove[]}`. Campos do evento: `id`, `item_id`, `reference_month`, `from_credit_card_id`, `from_payment_month`, `to_credit_card_id`, `to_payment_month`, `reason`, `recorded_at`. Nenhum autor nesse DTO. POST da mesma rota envia:

```json
{
  "target_credit_card_id": "22222222-2222-4222-8222-222222222222",
  "target_payment_month": "2026-12",
  "reason": "Ajuste da fatura selecionada"
}
```

POST → **201** `{data: InvoiceOccurrenceMove}`. GET alvo após FE-T04 mantém array e adiciona `current_allocation: {credit_card_id,payment_month,origin}` no topo; alocação ausente/null precisa ser definida/testada para uma ocorrência sem caminho de cartão válido, sem fabricar um período. Response atual não tem esse campo.

### 4.4 Dívidas, cronograma e quitação

POST `/v1/debts` → **201** `{data: Debt}`:

```json
{
  "name": "Financiamento",
  "description": null,
  "original_total_cents": null,
  "total_installments": 180,
  "first_projected_installment": 1,
  "scheduled_start_month": "2026-10",
  "installment_amount_cents": 60000,
  "cash_month_offset": 1,
  "context": null,
  "payment_method": "direct",
  "credit_card_id": null
}
```

`original_total_cents` é obrigatório anulável; zero explícito permanece zero. GET `/v1/debts?as_of=2026-10&status=active&projection_status=active` → **200** `{data: Debt[]}`. GET detalhe admite as_of e inclui settlement, status projetado, parcelas restantes, término/liberação derivados.

GET `/v1/debts/{debt_id}/schedule?from=2026-10&to=2027-09` → **200** `{data: {debt: Debt, range: {from,to}, occurrences: DebtOccurrence[]}}`. Occurrence inclui dívida, source_id, competência, n/N, valor, método, caixa, cartão/fatura e origem scheduled/early_settlement. Janela seguinte começa em 2027-10, sem sobreposição.

POST `/changes` envia `{effective_from,installment_amount_cents,context}` → **201** `{data: Debt}`. PATCH detalhe altera apenas nome/descrição → **200** `{data: Debt}`. POST `/archive` envia `{reason}` opcional → **200** `{data: Debt}`; não aceitar effective_from por analogia com itens genéricos.

POST `/v1/debts/{debt_id}/early-settlement`:

```json
{"reference_month":"2027-02","amount_cents":150000,"reason":"Quitação projetada"}
```

→ **201** `{data: DebtEarlySettlement}`, não Debt. Settlement: `id,debt_id,reference_month,amount_cents,reason,recorded_by,recorded_at,created_at`. Depois buscar novamente dívida/cronograma. GET separado do evento → 200 o mesmo envelope. Não criar edição/remover evento.

GET `/v1/debt-releases?from=2026-10&to=2027-09` → **200** `{data: {releases: DebtRelease[],monthly_totals: [{month,released_monthly_cents}]},range: {from,to}}`. Preservar range no topo; não presumir que está em data. Liberações não são renda.

### 4.5 Importação multipart

POST `/v1/credit-cards/{card_id}/invoices/{payment_month}/imports/preview` → **200**; confirmar em `/imports` → **201**. Ambos retornam `{data: InvoiceImportResult}` da implementação atual; confirmar adiciona imported_count quando aplicável. O token retornado deve ser tratado como opaco.

| Parte multipart | Prévia | Confirmação | Valor |
| --- | --- | --- | --- |
| `file` | obrigatório | obrigatório | mesmo CSV selecionado; até 2 MiB, até 5.000 linhas |
| `payment_allocations` | enviar | enviar | JSON string de linha CSV → fatura: `{"3":"2026-11"}`; sem alocações: `{}` |
| `preview_token` | não enviar | obrigatório | token da última prévia compatível com a seleção |

Resultado contém `card_id,payment_month,source,rows,new_count,existing_count,skipped_count,needs_invoice_count,needs_subscription_match_count,new_total_cents,new_charges_cents,new_payments_cents,new_payment_count,projected_installments,new_projected_count,preview_token` e imported_count opcional. Linhas incluem descrição/data/valor/kind/status/reason, fatura alocada, metadados de parcelas e de assinatura quando aplicáveis. Não exigir campos omitempty nas linhas sem associação.

Pagamentos nas linhas/componentes têm amount_cents negativo. Os agregados `new_payments_cents` da importação e `payments_total_cents` da fatura são magnitudes positivas; saldo líquido é compras menos pagamentos. Não somar a magnitude positiva dos pagamentos ao saldo nem trocar o sinal das linhas. Exemplo: compras 11000 e pagamento recebido -5000 resultam em new_charges_cents 11000, new_payments_cents 5000 e new_total_cents 6000.

Estados new/existing/skipped/imported/needs_invoice/needs_subscription_match guiam apresentação. Uma linha existing com matched_item_id e planned_amount_cents pode representar backfill pendente. Não confundir cobranças brutas do CSV com diferença líquida da projeção.

### 4.6 Plano, resumo e original

Criar plano `{name,start_month,end_month,currency_code}` → **201** `{data: Plan}`. GET/PATCH current → 200 `{data: Plan}`; PATCH somente em draft. Ativar POST sem body → **200** `{data: {plan: Plan,original: OriginalPlanSnapshot}}`. PUT savings envia `{effective_from,end_month,amount_cents,context}` → **200** `{data:{configured,periods}}`; não interpretar zero como não configurado.

GET `/v1/plans/current/months/{month}/summary?basis=cash` → **200** `{data: MonthlySummary}`. Usar todos os buckets/sources do catálogo; incluir metadados source_type, payment_method, credit_card_id, invoice_payment_month, invoice_allocation, reference_known, is_projected e dívida no DTO do frontend.

GET original → **200** `{data: {id,plan_id,kind:'original',schema_version,captured_at,plan}}`. Na versão 4, plan inclui:

```text
name, start_month, end_month, currency_code
items[], savings {configured,periods[]}
release_2 {
  financial_institutions[], credit_cards[] (periods, não configurations),
  payment_methods[], invoice_adjustments[], invoice_moves[]
}
release_3 {
  debts[] {financial_item_id, original_total_cents, total_installments,
    first_projected_installment, scheduled_start_month, scheduled_end_month,
    created_at, updated_at, periods[], payment_methods[], early_settlement}
}
```

Esse documento não é o DTO operacional Debt: não presumir remaining_installments/as_of_month no snapshot. Normalizar datas integrais históricas para apresentação e manter o JSON original imutável.

## 5. Contratos propostos para ajustes — ainda não publicados

### 5.1 Leitura por identidade

GET `/v1/invoice-adjustments/{adjustment_id}` → **200** `{data: InvoiceAdjustmentDetail}`. Sem query/body. 401 sem sessão; 404 ID inexistente/alheio; 500 indisponibilidade interna. Verificar ownership em todas as relações. O envelope proposto:

```json
{
  "data": {
    "id": "44444444-4444-4444-8444-444444444444",
    "currency_code": "BRL",
    "credit_card_id": "22222222-2222-4222-8222-222222222222",
    "payment_month": "2026-11",
    "reference_month": "2026-10",
    "name": "Totalpass",
    "amount_cents": 11000,
    "context": null,
    "status": "active",
    "archived_at": null,
    "created_at": "2026-10-04T12:00:00Z",
    "updated_at": "2026-10-04T12:00:00Z",
    "entry_kind": "charge",
    "import_source": "nubank_csv",
    "transaction_date": "2026-10-28",
    "is_projected": false,
    "installment_number": null,
    "installments_total": null,
    "reconciled_item_id": "11111111-1111-4111-8111-111111111111",
    "reconciled_reference_month": "2026-10",
    "capabilities": {
      "can_edit": true,
      "editable_fields": ["name", "amount_cents", "context"],
      "can_archive": true,
      "archive_effect": "restore_subscription_forecast"
    }
  }
}
```

`import_source`: null, nubank_csv ou nubank_csv_projection; `entry_kind`: charge/payment; `archive_effect`: remove_charge/restore_subscription_forecast/none. Capabilities dependem do recurso/estado, não de permissão fornecida pelo browser. Para arquivado: nenhuma mutação. Pagamento importado: nenhuma edição/arquivamento nesta interface até existir fluxo financeiro específico. Parcela inferida: permitir arquivamento apenas se a semântica de reserva/reimportação for validada em FE-T03; não presumir que reimportar restaura uma projeção arquivada.

Cobrança manual: permitir nome/valor/referência/contexto. Importada sem vínculo: preservar competência da importação e fingerprint; campos editáveis definidos pelo backend. Conciliada: referência ligada ao item é somente leitura; alteração de valor afeta apenas a ocorrência conciliada. Não permitir alterar cartão/mês/import fingerprint/entry kind/vínculo pelo PATCH. Reimportação não sobrescreve uma edição do valor real silenciosamente; comportamento deve ser garantido pelos testes da API.

### 5.2 PATCH parcial e arquivamento

PATCH da rota existente `/v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}` → **200** `{data: CardInvoiceAdjustment}`. Request proposto `{name?,amount_cents?,reference_month?,context?}` com ao menos um campo, preservação de omissão e null explícito em referência/contexto. Exemplo:

```json
{"amount_cents":11500}
```

Essa alteração preserva nome/contexto/vínculo. Atualizar a linha e auditoria na mesma transação, validando identidade da URL antes de alterar. Nenhuma alteração de localização pelo request. Após resposta, GET por identidade se forem necessárias capabilities/metadados completos.

POST `/archive` da mesma rota → **200** `{data: CardInvoiceAdjustment}`; `{}`/sem body é compatível com handler atual. Não anunciar reason auditável sem implementação explícita. 400 payload inválido; 404 ID/rota inexistente ou alheio; 409 arquivado/não editável/estado conflitante. Mapear erros sem vazar detalhes de banco. A confirmação local usa archive_effect; o cálculo do efeito continua backend.

## 6. Erros e consequências por fluxo

| HTTP/código | Tratamento |
| --- | --- |
| 401 unauthorized | Exigir autenticação; limpar dados da sessão na interface |
| 400 validation_error | Manter formulário/seleção; mostrar rejeição; revalidar entrada |
| 404 not_found | Recurso não encontrado; plano/original ausente somente no contexto apropriado |
| 409 period_overlap/payment_period_overlap | Explicar conflito de vigência, manter valores |
| 409 card_archived/debt_archived | Recarregar estado e desabilitar ação incompatível |
| 409 debt_already_settled/debt_completed | Recarregar dívida; não repetir registro |
| 409 invoice_move_rejected/adjustment_rejected | Mensagem de rejeição sem inventar causa; pedir nova consulta antes de repetir |
| 422 invoice_outside_operational_horizon/month_outside_horizon | Corrigir intervalo, não apresentar ausência de despesas |
| 422 debt_schedule_too_long | Dividir janela; não limitar duração da dívida |
| 422 invalid_settlement_month | Corrigir competência; manter formulário |
| 422 card_configuration_missing | Informar configuração ausente e link do cartão |
| 5xx/timeout/rede | Indisponibilidade; manter estado; retry explícito para GET |

Na importação, token/previsão desatualizados podem voltar como validation_error genérico: oferecer gerar nova prévia e revisar, sem afirmar que o CSV foi necessariamente malformado. O cliente preserva status/code/message; não transforma toda falha em api_request_failed sem manter o erro original.

## 7. Atualização e testes

| Mutação | Reconsultar/invalidate |
| --- | --- |
| Instituição/cartão | Detalhe/listas; faturas e seleção de cartões quando afetadas |
| Período/configuração/método | Histórico, detalhe, faturas e resumos/projeções relevantes |
| Ajuste/importação | Fatura, lista/gráfico de faturas, resumo e projeção; dados de vínculo |
| Movimentação | Históricos, fatura de origem/destino e resumos correspondentes |
| Quitação | Dívida, cronograma, liberações, faturas e resumos |
| Plano/economia | Plano/economia e resumos; original apenas consultar, nunca reconstruir |

Testes de contrato conferem bodies/envelopes/status reais, omissão/null, arrays de histórico, fonte de timestamps e GET proposto. Testes financeiros conferem antes/depois com idempotência de importação, reserva após arquivamento e contexto preservado. UI cobre janelas longas, filtro em URL, mobile/teclado, falha parcial e retorno fora de ordem. Nenhuma evidência local substitui smoke test autenticado depois da publicação.
