# Release 2 — Spec — Por que minha fatura é esse valor?

**Status:** aprovada para implementação  
**Versão da decisão:** 1  
**Aprovada em:** 2026-08-31

Esta versão congela o escopo funcional da Release 2. Mudanças posteriores que
alterem regras, contrato ou critérios de aceite devem ser registradas como nova
decisão e refletidas em `design.md`, `tasks.md` e `api/openapi.yaml` antes da
implementação correspondente.

## 1. Objetivo

Entregar a segunda capacidade financeira completa da API:

> Quais compromissos formam a fatura projetada de cada cartão e em qual mês
> ela reduz o dinheiro livre?

A release deve permitir que um usuário autenticado cadastre instituições e
cartões, vincule despesas por vigência, consulte faturas projetadas e entenda
cada componente sem registrar compras individuais.

## 2. Fontes

- visão geral:
  [`documentacao-produto-planejador-financeiro.md`](../../documentacao-produto-planejador-financeiro.md);
- decisão de competência e pagamento:
  [`decisao-competencia-e-pagamento-cartao.md`](../../decisao-competencia-e-pagamento-cartao.md);
- base funcional:
  [`Release 1`](../release-1/spec.md).

Requisitos diretamente relacionados:

- RF-009 a RF-011;
- RF-013, RF-014, RF-021, RF-033 e RF-040;
- HU-E01 a HU-E04;
- Fatia 2 — “Por que minha fatura é esse valor?”.

## 3. Resultado para o usuário

O usuário consegue:

1. cadastrar um banco ou outra instituição financeira;
2. cadastrar mais de um cartão na mesma instituição;
3. definir nome, dia nominal de vencimento e offset mensal do cartão;
4. alterar essas configurações a partir de um mês sem apagar o passado;
5. vincular uma despesa a um cartão a partir de uma competência;
6. trocar o cartão ou voltar ao pagamento direto por vigência;
7. consultar faturas projetadas;
8. abrir uma fatura e reconciliar seu total com os componentes;
9. lançar um ajuste consolidado diretamente em uma fatura;
10. mover uma ocorrência excepcional para outra fatura com auditoria;
11. ver a despesa na competência do consumo e no caixa do pagamento.

## 4. Escopo

### 4.1 Instituições financeiras

- recurso pertencente ao usuário, separado do cartão;
- nome obrigatório;
- uma instituição pode possuir vários cartões;
- status `active` ou `archived`;
- instituições arquivadas permanecem no histórico;
- não é permitido criar cartão ativo em instituição arquivada.

### 4.2 Cartões

Todo cartão possui:

- proprietário;
- instituição financeira;
- nome;
- status `active` ou `archived`;
- configurações mensais com vigência;
- data de criação e atualização.

Cada período de configuração possui:

- mês inicial;
- mês final opcional;
- dia nominal de vencimento entre 1 e 31;
- offset entre competência e mês de pagamento, de 0 a 12;
- instante em que a configuração foi registrada.

O offset padrão de criação é `1`: uma ocorrência referente a novembro forma,
por padrão, a fatura paga em dezembro.

O dia de vencimento é nominal. Quando o dia não existe no mês, a API retorna o
dia configurado, `nominal_due_date = null` e
`nominal_due_date_resolution = invalid_for_month`. Ela não antecipa nem
prorroga silenciosamente a data.

### 4.3 Forma de pagamento por vigência

Somente despesas podem receber forma de pagamento nesta release.

Métodos suportados:

- `direct`;
- `credit_card`.

Na ausência de configuração explícita, a despesa continua `direct` e utiliza o
`cash_month_offset` do período financeiro, preservando o comportamento da
Release 1.

Uma mudança de forma de pagamento possui:

- mês inicial;
- mês final opcional;
- método;
- cartão obrigatório quando o método for `credit_card`;
- instante do registro;
- contexto opcional.

Períodos explícitos da mesma despesa não podem se sobrepor. Uma mudança futura
encerra o período anterior e preserva competências passadas.

Quando uma ocorrência está em `credit_card`, seu mês de caixa é derivado do
offset do cartão aplicável à competência. O `cash_month_offset` do período
financeiro não participa desse cálculo.

### 4.4 Fatura projetada

A fatura é um recurso de domínio identificado por:

```text
cartão + mês de pagamento
```

Ela é calculada a partir de:

- ocorrências de despesas vinculadas ao cartão;
- ajustes consolidados associados diretamente à fatura;
- movimentações excepcionais registradas pelo usuário.

A fatura não é uma segunda despesa. Seu total é a soma de seus componentes e o
mesmo valor não pode aparecer duas vezes no resumo.

A resposta da fatura contém:

- cartão e instituição;
- mês de pagamento;
- dia nominal de vencimento;
- data nominal derivada, quando válida;
- total projetado;
- quantidade de componentes;
- componentes ordenados de forma determinística;
- origem da alocação de cada componente.

Origens possíveis:

- `calculated_from_reference`;
- `selected_invoice`;
- `moved_by_user`.

### 4.5 Ajustes consolidados

Um ajuste permite representar um valor conhecido da fatura sem cadastrar cada
compra. Ele possui:

- nome;
- valor não negativo em centavos;
- fatura selecionada;
- mês de referência opcional;
- contexto opcional;
- status e timestamps.

Referência ausente significa competência desconhecida. O sistema não inventa
uma competência. Nesta release, ajustes representam cobranças; créditos,
estornos e pagamentos realizados ficam fora do escopo.

O ajuste é aditivo: ele representa somente a parcela ainda não explicada por
itens vinculados. Informar o total integral da fatura como ajuste enquanto seus
componentes já estão vinculados produziria duplicidade e deve ser evitado pelo
cliente com ajuda do detalhamento retornado pela API.

Alterar ou arquivar um ajuste preserva os registros necessários para auditoria.

### 4.6 Movimentação excepcional

O usuário pode mover uma ocorrência de despesa vinculada a cartão para outra
fatura, inclusive de outro cartão próprio.

Cada movimentação registra:

- item e competência da ocorrência;
- cartão e fatura de origem;
- cartão e fatura de destino;
- motivo opcional;
- usuário responsável;
- instante do registro.

Uma nova movimentação considera a alocação atual como origem. O histórico não
é apagado. O destino deve pertencer ao usuário e respeitar o horizonte
operacional de faturas.

### 4.7 Competência, caixa e resumo mensal

Na base `reference`:

- a despesa vinculada ao cartão aparece no mês de referência;
- ela continua identificada pelo item financeiro original;
- ajustes com referência conhecida aparecem nessa competência;
- ajustes sem referência não são atribuídos a um mês inventado.

Na base `cash`:

- cada componente aparece no mês da fatura;
- as sources retornam metadados de cartão e fatura;
- os componentes formam `commitments_cents` diretamente;
- não existe uma source adicional para o total da fatura.

O breakdown preserva os tipos financeiros dos itens e adiciona
`card_invoice_adjustments_cents` para ajustes diretos. Cada source de cartão
informa `source_type`, `credit_card_id`, `invoice_payment_month` e
`invoice_allocation`.

Assim, a soma das sources continua igual ao total do resumo e uma ocorrência
participa no máximo uma vez em cada base.

### 4.8 Horizonte operacional

O resumo mensal continua aceitando somente meses entre `plan.start_month` e
`plan.end_month`.

Uma fatura pode, entretanto, vencer após o fim do planejamento quando uma
competência válida é deslocada pelo cartão. A consulta de faturas admite o
intervalo:

```text
plan.start_month até plan.end_month + 12 meses
```

Essas faturas de transbordo são consultáveis e explicáveis, mas não produzem
resumo de dinheiro livre fora do horizonte. O cliente deve ampliar o horizonte
do planejamento quando quiser calcular esse mês.

### 4.9 Fotografia original

Planos ativados antes desta release mantêm seu snapshot original na versão
existente.

Toda ativação realizada pela versão desta release cria snapshot schema 2. Ele
inclui as instituições e cartões vinculados, configurações temporais, formas
de pagamento, ajustes e movimentações existentes; coleções ainda não usadas
ficam vazias. O snapshot continua imutável.

## 5. Fora de escopo

- registro de compras individuais;
- data de fechamento do cartão;
- ciclo diário de compras;
- importação de fatura;
- Open Finance ou integração bancária;
- limite disponível do cartão;
- melhor dia de compra;
- pagamento, atraso, juros ou parcelamento da fatura;
- créditos, cashback e estornos;
- valores realizados e conciliação bancária;
- fechamento mensal;
- dívidas e compras parceladas;
- cartões adicionais com titulares distintos;
- idempotência genérica por chave de requisição;
- exclusão física de histórico;
- cálculo mensal fora do horizonte do plano.

## 6. Regras de negócio

### RN-R2-001 — Propriedade

Instituição, cartão, item, ajuste e movimentação devem pertencer ao usuário
autenticado. Recurso alheio é indistinguível de recurso inexistente.

### RN-R2-002 — Instituição pai

Todo cartão pertence exatamente a uma instituição ativa no momento da criação.

### RN-R2-003 — Configuração temporal

Para um cartão e um mês, no máximo um período de configuração é aplicável.

### RN-R2-004 — Vencimento nominal

O dia configurado não prova a data efetiva de vencimento. Datas inexistentes
não são corrigidas silenciosamente.

### RN-R2-005 — Offset padrão

O offset inicial é `1`, aceita valores de 0 a 12 e deriva o mês padrão da fatura
a partir da referência.

### RN-R2-006 — Compatibilidade do método

Rendas não podem ser vinculadas a cartão. `credit_card` exige cartão ativo na
competência inicial; `direct` exige `card_id = null`.

### RN-R2-007 — Precedência do cartão

Para despesa vinculada a cartão, a configuração do cartão substitui o
`cash_month_offset` do período financeiro somente no cálculo de caixa.

### RN-R2-008 — Identidade da ocorrência

Uma ocorrência projetada é identificada por item e mês de referência. Alterar
o valor de um período não cria duas ocorrências na mesma competência.

### RN-R2-009 — Total da fatura

O total projetado é exatamente a soma dos componentes vigentes, ajustes ativos
e movimentações aplicáveis, com verificação de overflow.

### RN-R2-010 — Sem dupla contagem

Na base de caixa, um componente de fatura substitui a ocorrência direta. O
total da fatura nunca é subtraído novamente.

### RN-R2-011 — Referência desconhecida

Ajuste sem referência participa da fatura e da base de caixa, mas não de um
resumo por competência.

### RN-R2-012 — Movimentação auditável

Uma exceção registra origem, destino, autor e instante. A movimentação atual é
a última válida e o histórico anterior permanece consultável.

### RN-R2-013 — Arquivamento

Recursos com histórico não são apagados. Cartão arquivado não aceita novos
vínculos nem ajustes, mas faturas passadas permanecem consultáveis.

### RN-R2-014 — Determinismo

Mesmos dados e mês produzem a mesma composição, total e ordenação.

### RN-R2-015 — Isolamento da fotografia

Adicionar a versão 2 do snapshot não altera snapshots versão 1 já criados.

## 7. Casos de borda

- cartão com vencimento nominal no dia 31 em fevereiro;
- offset zero;
- referência de dezembro formando fatura em janeiro do ano seguinte;
- fatura de transbordo após o horizonte;
- despesa mudando de cartão a partir de agosto;
- despesa voltando a pagamento direto;
- tentativa de sobrepor métodos de pagamento;
- cartão arquivado com faturas históricas;
- instituição arquivada com cartão ativo;
- ajuste sem referência;
- ajuste com valor zero;
- movimentação para a mesma fatura atual;
- duas movimentações sucessivas da mesma ocorrência;
- item alterado depois de uma movimentação;
- item arquivado com ocorrência histórica;
- total da fatura superior à renda;
- usuário tentando usar cartão de outro usuário;
- valor total excedendo `int64`;
- listagem de intervalo inválido ou maior que 24 meses.

## 8. Critérios de aceite

### CA-R2-001 — Fatura padrão

Dado um cartão com offset `1` e vencimento nominal no dia `6`, uma despesa de
R$ 700 referente a novembro aparece:

- em novembro no resumo `reference`;
- na fatura de dezembro;
- em dezembro no resumo `cash`.

### CA-R2-002 — Composição

Uma fatura com aluguel de R$ 1.000, combustível de R$ 700 e ajuste de R$ 180
retorna total de R$ 1.880 e exatamente esses três componentes.

### CA-R2-003 — Sem dupla contagem

No resumo `cash`, a soma das sources da fatura é R$ 1.880 e
`commitments_cents` aumenta em R$ 1.880, não em R$ 3.760.

### CA-R2-004 — Mudança futura

Trocar o cartão de uma despesa a partir de agosto não altera faturas formadas
por competências até julho.

### CA-R2-005 — Exceção

Mover a ocorrência de novembro da fatura de dezembro para janeiro remove o
componente de dezembro, adiciona-o em janeiro e preserva origem, destino e
instante no histórico.

### CA-R2-006 — Referência desconhecida

Um ajuste sem referência aparece na fatura e no resumo `cash`, sem aparecer em
nenhum resumo `reference`.

### CA-R2-007 — Dia inexistente

Para vencimento nominal `31` e fatura `2027-02`, a resposta mantém
`nominal_due_day = 31`, retorna `nominal_due_date = null` e informa
`invalid_for_month`.

### CA-R2-008 — Isolamento

Dois usuários com instituições, cartões e faturas diferentes não conseguem
ler, vincular ou mover recursos entre si pela API ou Data API.

### CA-R2-009 — Transbordo

Uma ocorrência de dezembro com offset `1` aparece na fatura de janeiro, ainda
que janeiro esteja fora do plano; a fatura é consultável e o resumo de janeiro
retorna `month_outside_horizon`.

### CA-R2-010 — Snapshot

Ativar um rascunho cria snapshot versão 2 determinístico, com ou sem cartões.
Snapshot versão 1 existente permanece byte a byte inalterado.

## 9. Contrato HTTP planejado

Os nomes serão congelados em `api/openapi.yaml` pela primeira task.

```text
POST   /v1/financial-institutions
GET    /v1/financial-institutions
PATCH  /v1/financial-institutions/{institution_id}
POST   /v1/financial-institutions/{institution_id}/archive

POST   /v1/credit-cards
GET    /v1/credit-cards
GET    /v1/credit-cards/{card_id}
PATCH  /v1/credit-cards/{card_id}
POST   /v1/credit-cards/{card_id}/changes
POST   /v1/credit-cards/{card_id}/archive

POST   /v1/plans/current/items/{item_id}/payment-changes
GET    /v1/plans/current/items/{item_id}/payment-history

GET    /v1/credit-cards/{card_id}/invoices?from=AAAA-MM&to=AAAA-MM
GET    /v1/credit-cards/{card_id}/invoices/{payment_month}
POST   /v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments
PATCH  /v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}
POST   /v1/credit-cards/{card_id}/invoices/{payment_month}/adjustments/{adjustment_id}/archive

POST   /v1/plans/current/items/{item_id}/occurrences/{reference_month}/invoice-moves
GET    /v1/plans/current/items/{item_id}/occurrences/{reference_month}/invoice-moves

GET    /v1/plans/current/months/{month}/summary?basis=cash
GET    /v1/plans/current/months/{month}/summary?basis=reference
```

Listagens de fatura exigem `from` e `to`, com intervalo inclusivo máximo de 24
meses. Por padrão, meses sem componentes são omitidos; `include_empty=true`
permite retorná-los. O detalhe de uma fatura válida sem componentes retorna
`200` com total zero, facilitando a navegação mensal do cliente.

## 10. Compatibilidade

- os endpoints da Release 1 permanecem válidos;
- itens existentes continuam com pagamento direto por padrão;
- nenhum dado precisa ser migrado pelo usuário;
- o resumo adiciona metadados opcionais de cartão/fatura às sources;
- novos tipos de componente são aditivos;
- `reference_month` passa a aceitar `null` exclusivamente para ajuste de fatura
  sem competência conhecida;
- o breakdown adiciona `card_invoice_adjustments_cents`;
- snapshots antigos continuam legíveis por `schema_version`;
- essa ampliação do contrato `/v1` é aceita enquanto não existe consumidor
  externo; se surgir um consumidor antes do merge, R2-T01 deve introduzir a
  compatibilidade por nova versão em vez de quebrá-lo;
- qualquer outra mudança incompatível encontrada na execução exige ADR e
  revisão da spec antes da implementação.
