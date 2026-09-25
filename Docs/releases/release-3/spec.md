# Release 3 — Spec — Quando essa dívida termina?

## 1. Objetivo

Responder, para cada compromisso parcelado:

> Quantas parcelas ainda estão projetadas, em que mês a dívida termina e
> quanto deixa de ficar comprometido depois disso?

A release transforma dívidas temporárias em fontes financeiras explicáveis,
integradas ao resumo mensal e às faturas projetadas, sem confundir projeção de
parcela com pagamento efetivamente realizado.

## 2. Fontes

- [`documentacao-produto-planejador-financeiro.md`](../../documentacao-produto-planejador-financeiro.md);
- Release 1 — itens, vigências, competência, caixa, resumo e fotografia;
- Release 2 — métodos de pagamento, cartões, faturas e movimentações;
- RF-008, RF-009, RF-014, RF-021, RF-024 e RF-025;
- HU-D01 a HU-D04;
- Fatia 3 — “Quando essa dívida termina?”.

## 3. Resultado para o usuário

O usuário consegue:

1. cadastrar uma dívida com parcelas conhecidas;
2. informar qual parcela inicia a projeção atual;
3. consultar o cronograma futuro com número, competência e valor;
4. identificar a última parcela e o primeiro mês sem o compromisso;
5. visualizar o valor mensal liberado após o término;
6. pagar a parcela diretamente ou vinculá-la a um cartão;
7. ver parcelas de cartão dentro da composição da fatura;
8. alterar o valor projetado a partir de um mês sem apagar o passado;
9. quitar antecipadamente substituindo a parcela do mês por um valor de
   quitação e removendo ocorrências posteriores da projeção;
10. reconciliar dívida, fatura e resumo mensal sem dupla contagem.

## 4. Escopo

### 4.1 Dívida parcelada

Uma dívida pertence ao usuário e é também um item financeiro especializado.
Ela não pertence ao planejamento: participa de cada projeção pela interseção
entre seu cronograma e o horizonte consultado, conforme a Release 2.1. Sua
identidade pública é o próprio `financial_item_id`, exposto como `debt_id` nas
rotas de dívida.

Ela possui nome, descrição opcional, valor total original opcional, quantidade
total de parcelas, número da primeira parcela projetada, mês dessa parcela,
valor mensal projetado, contexto, status e timestamps.

O valor total original é informativo. Ele não serve para inferir juros, saldo
devedor ou parcela e pode ser diferente da soma das parcelas projetadas.

### 4.2 Cronograma projetado

O cronograma é derivado; não existe uma linha persistida para cada parcela.

Para uma dívida com:

```text
total_installments = 12
first_projected_installment = 5
start_month = 2026-09
```

a projeção contém oito ocorrências, de `5/12` a `12/12`, e termina em
`2027-04`.

Cada ocorrência contém:

- identidade lógica `debt_id + reference_month`;
- número e total de parcelas;
- mês de referência e valor projetado;
- método e mês de pagamento;
- cartão e fatura, quando aplicável;
- origem `scheduled` ou `early_settlement`.

O mês de referência representa a competência. Pagamento direto usa o
`cash_month_offset` do período; cartão usa o offset aplicável à competência;
uma movimentação excepcional continua prevalecendo sobre a alocação padrão.

### 4.3 Alteração do valor projetado

O valor pode mudar a partir de uma competência do cronograma. A alteração
encerra o período anterior, cria outro até o término estrutural e preserva
parcelas anteriores, números e data final.

Alterar quantidade, início ou numeração estrutural depois da criação não faz
parte desta release. Uma dívida criada incorretamente pode ser arquivada e
recadastrada.

### 4.4 Forma de pagamento

Dívidas aceitam `direct` e `credit_card`. Sem configuração explícita, usam
pagamento direto. Mudanças de método preservam competências anteriores.

Na base `reference`, a parcela aparece em sua competência. Na base `cash`, ela
aparece no pagamento direto ou na fatura correspondente.

### 4.5 Término e valor liberado

Sem quitação antecipada:

- `scheduled_end_month` é o mês da última parcela;
- `release_from_month` é o mês seguinte;
- `released_monthly_cents` é o valor da última parcela regular projetada.

Quando existe quitação, `released_monthly_cents` é o valor da primeira parcela
regular suprimida. Assim, o campo explica a redução imediata a partir de
`release_from_month`, mesmo quando a parcela teve mudanças de valor.

O valor liberado descreve uma redução de compromisso. Ele não é renda, economia
ou source adicional do resumo, e o sistema não decide seu destino.

### 4.6 Quitação antecipada

O usuário pode registrar uma única quitação antecipada, contendo mês de
referência, valor projetado, motivo opcional, autor e instante.

A quitação:

- ocorre entre a primeira e a penúltima parcela projetada;
- substitui a parcela regular do mês;
- usa o método de pagamento aplicável;
- remove parcelas posteriores da projeção atual;
- preserva cronograma original e evento para auditoria;
- antecipa término efetivo e liberação;
- não confirma pagamento realizado.

O evento é append-only e não pode ser editado ou removido nesta release.

### 4.7 Status

Além do status persistido `active` ou `archived`, a API expõe
`projection_status` para um `as_of_month`:

- `planned`: antes da primeira parcela;
- `active`: entre início e término efetivo;
- `completed`: depois do término natural;
- `settled_early`: no mês da quitação e depois dele;
- `archived`: item arquivado.

Status projetado não significa pagamento confirmado.

### 4.8 Resumo mensal

O breakdown adiciona `debt_installments_cents`. Parcelas e quitações entram em
`commitments_cents` e reduzem o resultado.

Sources podem adicionar `debt_id`, `installment_number`,
`installments_total` e `debt_occurrence_kind`, além dos metadados de pagamento
existentes. A soma das sources continua igual ao total. Valor liberado nunca é
somado ao resultado.

### 4.9 Fatura projetada

Parcela paga por cartão é componente `financial_item_occurrence`, com
metadados de dívida. A quitação substitui a ocorrência regular da mesma
competência. Movimentações continuam identificadas por item e competência.

### 4.10 Horizonte operacional

- início e término estrutural independem do horizonte do plano;
- resumos permanecem limitados ao horizonte;
- cartão pode gerar fatura de transbordo até `plan.end_month + 12`;
- cronograma aceita intervalo inclusivo máximo de 120 meses, limitado à dívida;
- um planejamento inclui somente as parcelas que alcançam sua janela.

### 4.11 Fotografia original

Snapshots v1, v2 e v3 permanecem intactos. Novas ativações geram schema 4,
incluindo dívidas, períodos, métodos, quitações e recursos R2 alcançáveis. O
documento permanece determinístico e imutável.

## 5. Fora de escopo

- pagamento realizado, parcelas pagas, atraso ou inadimplência;
- vencimento diário;
- juros, multas, correção e sistemas Price/SAC;
- cálculo automático de saldo devedor;
- baixa bancária ou conciliação;
- renegociação, suspensão ou relacionamento entre contratos;
- quitação parcial e amortização extraordinária;
- reversão pública de quitação;
- alteração estrutural após criação;
- importação;
- valores realizados e fechamento mensal;
- recomendação para o valor liberado;
- dashboard anual agregado.

## 6. Regras de negócio

### RN-R3-001 — Propriedade

Todo recurso pertence ao usuário autenticado. Identificadores alheios retornam
`404`.

### RN-R3-002 — Item especializado

Toda dívida possui exatamente um item `debt_installment`; todo item desse tipo
possui exatamente uma dívida.

### RN-R3-003 — Parcela inicial

`first_projected_installment` fica entre `1` e `total_installments`.

### RN-R3-004 — Término derivado

```text
scheduled_end_month
= start_month + (total_installments - first_projected_installment) meses
```

O cliente não envia término independente.

### RN-R3-005 — Dinheiro

Valores usam centavos inteiros. Parcela e quitação são maiores que zero; total
original, quando informado, é não negativo.

### RN-R3-006 — Ausência e zero

Total original ausente significa desconhecido. Zero explícito não vira nulo.

### RN-R3-007 — Projeção, não realizado

Ocorrências da R3 mantêm `completeness = projected`.

### RN-R3-008 — Numeração estável

Alterar valor ou método não muda identidade nem número da parcela.

### RN-R3-009 — Sem dupla contagem

Uma parcela participa no máximo uma vez por base. Cartão substitui o caminho
direto em `cash`, e o total da fatura não é somado novamente.

### RN-R3-010 — Quitação substitutiva

No mês da quitação existe uma ocorrência: a quitação. A parcela regular do mês
e todas as posteriores deixam de participar da projeção atual.

### RN-R3-011 — Quitação imutável

O evento é append-only e único por dívida.

### RN-R3-012 — Valor liberado informativo

`released_monthly_cents` nunca é source positiva.

### RN-R3-013 — Explicabilidade

Cronograma, fatura e resumo apontam para a mesma ocorrência lógica.

### RN-R3-014 — Determinismo

Cronogramas, sources, componentes e snapshots usam ordenação estável.

### RN-R3-015 — Compatibilidade

Sem dívidas, itens, resumos, faturas e snapshots preservam a R2.

## 7. Casos de borda

- dívida `1/1`;
- dívida preexistente iniciando em `17/24`;
- virada de ano;
- mudança de valor antes da última parcela;
- pagamento direto com offset 12;
- fatura após o horizonte;
- troca de cartão e movimentação excepcional;
- quitação na primeira competência;
- tentativa de quitação na última parcela ou segunda quitação;
- dívida concluída ou arquivada;
- total original desconhecido ou diferente da soma projetada;
- overflow;
- acesso cruzado.

## 8. Critérios de aceite

### CA-R3-001 — Cronograma restante

Dívida de 12 parcelas iniciando em setembro na `5/12` retorna oito ocorrências
e termina em abril do ano seguinte.

### CA-R3-002 — Valor liberado

Dívida de R$ 600 terminando em dezembro retorna término `2026-12`, liberação
em `2027-01` e `released_monthly_cents = 60000`, sem source positiva em janeiro.

### CA-R3-003 — Competência

Parcela de R$ 600 referente a novembro aparece em novembro em `reference`.

### CA-R3-004 — Cartão e caixa

Com cartão de offset 1, parcela de novembro aparece em novembro em
`reference`, na fatura de dezembro e em dezembro em `cash`, uma única vez.

### CA-R3-005 — Mudança futura

Alterar R$ 600 para R$ 650 a partir de janeiro não muda dezembro nem a
numeração.

### CA-R3-006 — Quitação antecipada

Dívida até dezembro quitada em outubro por R$ 1.500 mantém o passado, substitui
outubro, remove novembro/dezembro da projeção e libera a partir de novembro.

### CA-R3-007 — Quitação em cartão

Quitação em cartão participa uma vez da fatura e de `cash`, mantendo sua
competência em `reference`.

### CA-R3-008 — Movimentação

Mover parcela de dezembro para fatura de janeiro altera apenas sua alocação de
caixa e preserva dívida, competência e número.

### CA-R3-009 — Isolamento

Dois usuários não leem, alteram, quitam ou projetam dívidas alheias pela API Go
ou pelo Data API.

### CA-R3-010 — Snapshot

Ativação com dívida cria snapshot v4 determinístico; v1/v2/v3 permanecem
intactos.

### CA-R3-011 — Compatibilidade

Sem dívidas, a suíte de aceitação da R2 não muda.

## 9. Contrato HTTP planejado

Os nomes serão congelados em `api/openapi.yaml` pela primeira task.

```text
POST   /v1/debts
GET    /v1/debts?as_of=AAAA-MM&status=active
GET    /v1/debts/{debt_id}?as_of=AAAA-MM
PATCH  /v1/debts/{debt_id}
POST   /v1/debts/{debt_id}/changes
POST   /v1/debts/{debt_id}/archive

GET    /v1/debts/{debt_id}/schedule?from=AAAA-MM&to=AAAA-MM
POST   /v1/debts/{debt_id}/early-settlement
GET    /v1/debts/{debt_id}/early-settlement

GET    /v1/debt-releases?from=AAAA-MM&to=AAAA-MM
```

As rotas de método de pagamento, movimentação de ocorrência e resumo continuam
válidas. `as_of` usa o mês atual no timezone do perfil quando omitido.

As rotas são centradas no usuário, e não no planejamento. O planejamento
consome somente as ocorrências que intersectam seu horizonte.

## 10. Compatibilidade

- `financial_item_kind` recebe `debt_installment` aditivamente;
- item genérico pode listar o tipo, mas não criá-lo;
- métodos de pagamento passam a aceitar dívida;
- componentes e sources ganham campos opcionais;
- o resumo ganha um bucket sem reinterpretar os anteriores;
- snapshots antigos continuam legíveis por `schema_version`;
- incompatibilidade encontrada em R3-T01 exige revisão da spec.

## 11. Questões não bloqueantes

- categorias e credor estruturados;
- correção estrutural auditável;
- reversão de quitação;
- suspensão e renegociação;
- saldo devedor manual;
- quitação parcial;
- vencimento diário;
- agrupamento de contratos.
