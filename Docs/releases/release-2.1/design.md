# Release 2.1 — Design técnico — Plano como janela de projeção

## 1. Direção arquitetural

A Release 2.1 corrige a fronteira entre cadastro financeiro e planejamento:

```text
usuário
├── perfil/moeda base
├── itens financeiros ─ períodos ─ métodos de pagamento
├── instituições ─ cartões
├── ajustes e movimentações de fatura
└── planejamentos
    ├── horizonte de projeção
    ├── economia planejada
    └── snapshot original
```

O monólito modular permanece. Não é criado um novo serviço nem um segundo
modelo paralelo de lançamentos.

## 2. Decisões técnicas

### 2.1 Ownership por usuário

`financial_items` passa a ter identidade composta `(id, user_id)`. Tabelas
dependentes repetem `user_id` e usam FKs compostas para garantir ownership sem
consultar o planejamento.

O `plan_id` é removido das tabelas operacionais após o backfill de moeda e a
substituição das constraints. Ele permanece apenas onde o dado é realmente uma
configuração ou fotografia do plano.

### 2.2 Moeda explícita nos registros monetários

Adicionar `currency_code text not null` a `financial_items` e
`card_invoice_adjustments`.

Na migration:

1. adicionar a coluna inicialmente nula;
2. copiar `plans.currency_code` pelo `plan_id` legado;
3. validar formato;
4. tornar a coluna obrigatória;
5. remover o vínculo legado.

Na criação, a aplicação lê o plano atual apenas para obter a moeda padrão. As
demais operações do item não dependem do plano.

### 2.3 Seleção por janela

Repositórios carregam registros por `user_id` e, quando possível, aplicam
predicados de interseção para reduzir dados:

```sql
period.start_month <= :window_end
and coalesce(period.end_month, 'infinity') >= :window_start
```

Projetores puros continuam responsáveis por competência, caixa, cartão e
offset. O plano entra como parâmetro da projeção, não como owner do item.

### 2.4 Economia permanece no plano

`saving_periods` não muda de ownership. Sua semântica é uma intenção específica
do plano, e não um registro financeiro contínuo.

### 2.5 Faturas e eventos

`financial_item_payment_periods`, `card_invoice_adjustments` e
`card_invoice_audit_events` passam a usar FKs por usuário.

Movimentações são selecionadas por `(user_id, financial_item_id,
reference_month)`. Ajustes são selecionados por `(user_id, credit_card_id,
payment_month)` ou por referência. O horizonte continua validado no caso de uso
de consulta para evitar projeções acidentais ilimitadas.

### 2.6 Snapshots

O schema v2 é preservado para snapshots existentes. Novas ativações passam a
usar schema v3 da Release 2.1. A futura Release 3 deverá usar schema v4 ao
adicionar dívidas.

O documento v3 inclui `currency_code` em cada item e não usa `plan_id` como
ownership. A query captura itens que interceptem o horizonte e os recursos
alcançáveis por eles, com ordenação determinística.

## 3. Migração de dados

A migration é progressiva dentro de uma transação:

1. adicionar `currency_code` e chaves `(id, user_id)`;
2. preencher moeda a partir do plano legado;
3. remover triggers que validam horizonte;
4. substituir FKs das tabelas dependentes;
5. persistir moeda também nos ajustes independentes de fatura;
6. substituir índices baseados em plano por índices baseados em usuário e data;
7. remover `plan_id` de períodos, métodos, ajustes, eventos e itens;
8. recriar funções/constraints dependentes;
9. manter RLS por `user_id`.

Rollback operacional exige restaurar o schema anterior e repovoar `plan_id` a
partir de um mapeamento preservado externamente. Por isso, o deploy deve gerar
backup antes da migration e usar expansão/contração em produção se houver
clientes antigos escrevendo diretamente no banco. Hoje as escritas passam pela
API Go, reduzindo o risco de compatibilidade.

## 4. Alterações por módulo

### `financialitem`

- remover `PlanID` do modelo público;
- adicionar `CurrencyCode`;
- repositório recebe owner, sem plan ID;
- criação recebe moeda resolvida pela aplicação;
- validação temporal deixa de consultar horizonte.

### `paymentmethod`

- validar owner/item/cartão;
- remover contenção no horizonte;
- repositório filtra por owner e item.

### `invoiceadjustment` e `invoiceallocation`

- remover plano das identidades e portas;
- preservar validações de competência e destino;
- usar o plano somente para autorizar a janela operacional exposta pela API.

### `cardinvoice`

- carregar itens, métodos, ajustes e movimentos por usuário;
- manter leitura `repeatable read`;
- projetar apenas a janela solicitada.

### `monthlysummary`

- validar o mês contra o plano;
- carregar itens por usuário e moeda;
- carregar economia por plano;
- projetar faturas com dados user-scoped.

### `planning`

- ativação verifica renda que intercepte o horizonte;
- snapshot v3 captura dados user-scoped alcançáveis;
- snapshots v1/v2 não sofrem alteração.

## 5. Contrato HTTP

As rotas canônicas de itens, pagamentos e movimentações usam
`/v1/financial-items`. Os caminhos antigos sob `/v1/plans/current/items`
permanecem como aliases de compatibilidade nas Releases 2.1 e 3, sem alterar o
ownership dos recursos.

Mudanças de representação:

- `FinancialItem.plan_id` é removido;
- `FinancialItem.currency_code` é adicionado;
- erros por `month_outside_horizon` deixam de ocorrer no CRUD de itens e
  métodos, mas continuam em resumos e consultas de projeção;
- exemplos passam a demonstrar períodos que ultrapassam o plano.

Essa remoção é uma mudança incompatível para clientes que leem `plan_id`. Se o
cliente atual depender do campo, a implementação pode mantê-lo como nullable e
deprecated por uma release sem restaurar o vínculo no banco.

## 6. Concorrência e segurança

- mudanças temporais continuam usando locks e exclusion constraints;
- relações usam `(resource_id, user_id)`;
- RLS permanece baseada em `auth.uid() = user_id`;
- buscas alheias retornam `404`;
- snapshots usam leitura consistente;
- consultas não podem somar moedas diferentes.

## 7. Estratégia de testes

- domínio: vigência fora do plano, período aberto e moeda;
- repositórios: ownership sem `plan_id` e não sobreposição;
- migration: backfill e constraints após remoção;
- resumo: interseção, moeda e ausência de dupla contagem;
- fatura: métodos e movimentos depois do fim do plano;
- ativação: renda somente fora do horizonte não habilita o plano;
- snapshot: schema v3, ordem, alcance e leitura v1/v2;
- integração: cenário do seguro até setembro de 2027;
- isolamento: dois usuários e tentativas de vínculo cruzado.
