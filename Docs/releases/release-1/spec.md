# Release 1 — Spec — Quanto está livre?

## 1. Objetivo

Entregar a primeira capacidade financeira completa da API:

> Com o que está previsto para entrar, os compromissos do mês e o valor
> planejado para guardar, quanto permanece livre?

A release deve permitir que um usuário autenticado crie seu planejamento
principal, cadastre as premissas mínimas e consulte um resumo mensal
explicável.

## 2. Fontes

- visão geral:
  [`documentacao-produto-planejador-financeiro.md`](../../documentacao-produto-planejador-financeiro.md);
- competência e caixa:
  [`decisao-competencia-e-pagamento-cartao.md`](../../decisao-competencia-e-pagamento-cartao.md);
- fundação:
  [`Release 0`](../release-0/README.md).

Requisitos de produto diretamente relacionados:

- RF-001 a RF-007;
- RF-012 a RF-014;
- RF-018;
- RF-021;
- RF-033;
- RF-040;
- HU-A01, HU-B01 a HU-B03, HU-C01 a HU-C04 e HU-F01 a HU-F03.

## 3. Resultado para o usuário

O usuário consegue:

1. criar um planejamento principal em rascunho;
2. definir período e moeda;
3. cadastrar rendas recorrentes e pontuais;
4. cadastrar despesas fixas;
5. cadastrar despesas variáveis projetadas;
6. definir quanto pretende guardar;
7. registrar quando cada valor começa e termina;
8. consultar qualquer mês do horizonte;
9. entender a composição do total;
10. confirmar o planejamento e preservar a referência original.

## 4. Escopo

### 4.1 Planejamento principal

- um planejamento principal por usuário;
- status `draft`, `active` ou `archived`;
- horizonte definido por mês inicial e final;
- moeda principal, inicialmente `BRL`;
- timezone herdada do perfil;
- criação e edição livre enquanto estiver em rascunho;
- ativação explícita;
- fotografia original imutável criada na primeira ativação;
- ativação idempotente.

### 4.2 Itens financeiros

Tipos da release:

- renda recorrente;
- renda pontual;
- despesa fixa;
- despesa variável projetada.

Todo item possui:

- identificador;
- planejamento;
- proprietário;
- nome;
- tipo;
- descrição opcional;
- status;
- data de criação e atualização.

### 4.3 Períodos

Valores temporais pertencem a períodos, não diretamente ao item.

Cada período possui:

- mês inicial;
- mês final opcional;
- valor projetado em centavos;
- recorrência mensal ou pontual;
- descrição contextual opcional;
- deslocamento entre referência e caixa.

Regras:

- meses são representados por `AAAA-MM`;
- início e fim são inclusivos;
- período sem fim vale até ser encerrado ou até o horizonte consultado;
- períodos incompatíveis do mesmo item não podem se sobrepor;
- mudança de valor cria novo período;
- o período anterior é encerrado no mês imediatamente anterior;
- meses históricos não são sobrescritos silenciosamente;
- alteração retroativa registra quando foi realizada;
- item pontual possui apenas uma ocorrência;
- valor monetário não utiliza ponto flutuante.

### 4.4 Competência e caixa

Cada ocorrência mensal possui:

- mês de referência;
- mês de recebimento ou pagamento.

Na Release 1:

- renda pode ser recebida no mesmo mês ou em mês posterior à referência;
- despesa paga diretamente usa deslocamento configurável, normalmente zero;
- cartões ainda não existem;
- resumo aceita `reference` ou `cash`;
- a base padrão para dinheiro livre é `cash`;
- todo resumo retorna a base utilizada.

### 4.5 Valor planejado para guardar

- pode existir sem destino;
- possui períodos de vigência;
- reduz o dinheiro livre;
- destinos específicos não são obrigatórios;
- alocações opcionais poderão ser adicionadas sem alterar o total;
- uma alocação nunca representa compromisso adicional.

Nesta release, a API entrega o total planejado. Cadastro estruturado de metas,
contas de investimento e múltiplas alocações fica fora do escopo.

### 4.6 Resumo mensal

Fórmula na base de caixa:

```text
dinheiro livre planejado
= rendas previstas para recebimento
− despesas com pagamento previsto
− valor planejado para guardar
```

Fórmula na base de referência:

```text
resultado planejado por competência
= rendas referentes ao mês
− despesas referentes ao mês
− valor planejado para guardar referente ao mês
```

O resultado por referência não deve ser apresentado como saldo disponível em
conta.

O resumo retorna:

- mês;
- moeda;
- base;
- renda total;
- compromissos totais;
- valor planejado para guardar;
- dinheiro livre ou resultado por competência;
- indicador de resultado negativo;
- composição por tipo;
- itens de origem;
- informação de completude.

## 5. Fora de escopo

- cartões e faturas;
- dívidas e parcelas;
- valores realizados;
- fechamento mensal;
- patrimônio e investimentos;
- metas financeiras estruturadas;
- categorias customizáveis;
- importação;
- múltiplos planejamentos independentes;
- múltiplas moedas no mesmo planejamento;
- cenários;
- transações individuais;
- edição da fotografia original;
- dashboard anual agregado.

Embora não exista dashboard anual, o usuário pode consultar individualmente
qualquer mês pertencente ao horizonte.

## 6. Regras de negócio

### RN-001 — Propriedade

Todo recurso pertence ao usuário autenticado. Identificadores de outro usuário
não concedem acesso.

### RN-002 — Planejamento único

Um usuário não pode possuir dois planejamentos principais não arquivados.

### RN-003 — Horizonte

O mês final deve ser igual ou posterior ao mês inicial.

### RN-004 — Moeda

Todos os itens de um planejamento utilizam a moeda do planejamento.

### RN-005 — Dinheiro

Valores são inteiros em centavos, maiores ou iguais a zero. O sinal financeiro
é determinado pelo tipo do item.

### RN-006 — Dinheiro livre negativo

O resultado pode ser negativo e deve ser retornado sem julgamento.

### RN-007 — Ausência e zero

Ausência de configuração não é igual a valor zero.

### RN-008 — Explicabilidade

Os totais do resumo devem corresponder exatamente à soma dos itens retornados
no detalhamento.

### RN-009 — Vigência

Para uma competência, no máximo um período do mesmo item pode ser aplicável.

### RN-010 — Mudança

Alterar um valor vigente cria novo período e preserva o anterior.

### RN-011 — Arquivamento

Item utilizado em uma competência não é apagado. Ele é encerrado ou arquivado.

### RN-012 — Ativação

O planejamento somente pode ser ativado quando possuir:

- horizonte válido;
- pelo menos uma renda;
- configuração explícita do valor planejado para guardar, inclusive zero.

### RN-013 — Fotografia original

A primeira ativação cria uma fotografia imutável das premissas. Repetir a
operação retorna a mesma referência e não cria outra fotografia.

### RN-014 — Rascunho

Resumos de rascunho são permitidos e identificados como prévia.

### RN-015 — Base

`cash` utiliza recebimento e pagamento. `reference` utiliza competência
econômica. Uma ocorrência participa no máximo uma vez em cada base.

## 7. Casos de borda

- mês sem renda;
- resultado livre negativo;
- renda pontual no mesmo mês de uma recorrente;
- salário referente a novembro e recebido em dezembro;
- período começando depois do início do planejamento;
- período terminando antes do fim do planejamento;
- alteração retroativa;
- tentativa de criar períodos sobrepostos;
- item arquivado com histórico;
- valor zero explícito;
- ausência de meta para guardar;
- meta para guardar igual a zero;
- mês fora do horizonte;
- usuário tentando acessar recurso de outro usuário;
- ativação repetida;
- ativação sem renda;
- mês inválido, como `2026-13`.

## 8. Critérios de aceite

### CA-001 — Primeiro cálculo

Dado um planejamento com:

```text
Renda: R$ 6.000
Despesas fixas: R$ 2.100
Despesas variáveis projetadas: R$ 700
Guardar: R$ 1.200
```

o resumo retorna:

```text
Compromissos: R$ 2.800
Livre: R$ 2.000
```

### CA-002 — Explicação

Abrir o resumo permite identificar cada renda e despesa que formou o cálculo.

### CA-003 — Vigência

Uma mudança válida a partir de agosto não altera julho.

### CA-004 — Renda defasada

Uma renda referente a novembro e recebida em dezembro aparece:

- em novembro na base `reference`;
- em dezembro na base `cash`.

### CA-005 — Isolamento

Dois usuários com dados diferentes recebem somente seus próprios recursos e
totais.

### CA-006 — Original

Ativar o rascunho cria uma única fotografia original. Mudanças posteriores não
alteram essa fotografia.

### CA-007 — Negativo

Compromissos superiores à renda retornam valor negativo e sua composição.

### CA-008 — Consistência

A soma dos componentes é igual aos totais em todas as respostas.

## 9. Contrato HTTP

Os nomes definitivos estão publicados em
[`api/openapi.yaml`](../../../api/openapi.yaml).

```text
POST   /v1/plans
GET    /v1/plans/current
PATCH  /v1/plans/current
POST   /v1/plans/current/activate
GET    /v1/plans/current/original

POST   /v1/plans/current/items
GET    /v1/plans/current/items
GET    /v1/plans/current/items/{item_id}
PATCH  /v1/plans/current/items/{item_id}
POST   /v1/plans/current/items/{item_id}/changes
POST   /v1/plans/current/items/{item_id}/archive

PUT    /v1/plans/current/savings
GET    /v1/plans/current/savings

GET    /v1/plans/current/months/{month}/summary?basis=cash
GET    /v1/plans/current/months/{month}/summary?basis=reference
```

`month` utiliza o formato `AAAA-MM`. Todos exigem autenticação.

## 10. Questões não bloqueantes

- paginação da lista de itens;
- categorias padrão;
- alocações do valor guardado;
- criação de nova referência original no futuro.
