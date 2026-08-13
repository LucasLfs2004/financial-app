# Decisão de Produto — Competência e pagamento de gastos no cartão

> **Status:** decisão inicial para as primeiras releases  
> **Contexto:** planejamento financeiro mensal com valores consolidados  
> **Escopo:** cartões de crédito, faturas, relação com a renda e cálculo do dinheiro livre

## 1. Problema

Um gasto realizado no cartão pode envolver meses diferentes:

- o mês ao qual o consumo se refere;
- o mês em que o gasto compõe uma fatura;
- o mês em que a fatura é paga.

Exemplo:

```text
Consumo realizado em novembro
Fatura paga em dezembro
Renda referente a novembro recebida em dezembro
Renda recebida em dezembro utilizada no pagamento da fatura
```

O produto precisa preservar o contexto do consumo sem comprometer o cálculo de
disponibilidade financeira. Para responder quanto está livre no mês, o momento
relevante é quando a fatura reduz a renda disponível.

## 2. Decisão principal

Os compromissos pagos por cartão reduzem o dinheiro livre no mês de pagamento
da fatura.

O mês de referência do consumo deve ser armazenado sempre que for conhecido.
Ele pode permanecer ausente somente quando a origem não fornecer essa
informação, como em um ajuste cadastrado diretamente na fatura.

```text
Mês de referência: quando o consumo aconteceu ou ao que ele se refere
Mês de pagamento: quando a fatura compromete a renda disponível
```

As análises por competência utilizam o mês de referência. Os cálculos de
dinheiro livre e fluxo de caixa utilizam o mês de pagamento.

O mesmo princípio se aplica à renda:

```text
Mês de referência da renda: período ao qual a renda se refere
Mês de recebimento: quando a renda fica disponível
```

Assim, o produto pode representar uma renda referente a novembro que foi
recebida no primeiro dia útil de dezembro.

## 3. Modelo mensal inicial

O produto não exige o registro de cada compra individual. Por isso, a primeira
versão utilizará uma regra mensal, em vez de tentar reproduzir com precisão o
ciclo diário de cada cartão.

No comportamento padrão:

```text
Gasto de referência de novembro
→ fatura com vencimento em dezembro
→ compromisso financeiro de dezembro
```

Essa relação representa o fluxo principal em que o cartão fecha no fim de um
mês e é pago com a renda do mês seguinte.

Nesse exemplo, a visão por competência de novembro relaciona a renda e os
consumos referentes a novembro. A visão de caixa de dezembro relaciona a renda
recebida e a fatura paga em dezembro. São duas leituras dos mesmos fatos, sem
duplicar valores.

O deslocamento padrão entre referência e pagamento deve ser representado como
uma regra do cartão. Inicialmente, o padrão será de um mês.

```text
payment_month = reference_month + 1 mês
```

Essa regra poderá ser configurável futuramente caso outros comportamentos
sejam necessários.

## 4. Dados do cartão

Para a primeira versão, o cartão deve possuir:

- proprietário;
- banco ou instituição;
- nome;
- dia nominal de vencimento;
- status;
- início e término de utilização, quando aplicável;
- deslocamento mensal padrão entre referência e pagamento.

Exemplo:

```text
Banco: Nubank
Cartão: Ultravioleta
Dia nominal de vencimento: 6
Deslocamento padrão: 1 mês
```

O banco e o cartão são conceitos distintos. Um banco pode possuir mais de um
cartão.

## 5. Dia de fechamento e vencimento

O dia de fechamento não será obrigatório na primeira versão.

O fechamento real pode variar por:

- instituição;
- mês;
- finais de semana;
- feriados;
- alterações feitas pelo banco;
- regras específicas do ciclo.

Sem integração com a instituição financeira, a API não consegue descobrir com
segurança as datas exatas de fechamento e vencimento. Finais de semana,
feriados e regras da instituição podem alterar as datas efetivas. Por isso, a
API não deve apresentar uma data calculada como se fosse uma informação
confirmada pelo banco.

Na configuração do cartão, o produto armazena inicialmente o dia nominal de
vencimento. O mês da fatura é suficiente para os cálculos mensais.

```text
Dia nominal de vencimento: configuração do cartão
Mês da fatura: competência mensal do pagamento
Data efetiva de vencimento: opcional
Data efetiva de pagamento: opcional
```

Se o dia nominal não existir em determinado mês, como dia 31 em fevereiro, a
API não deve inventar silenciosamente uma data efetiva. A especificação técnica
da release deverá definir como apresentar a previsão nominal, sem tratá-la como
confirmação da instituição.

O produto trabalhará inicialmente com competência mensal consolidada. Caso uma
integração futura forneça ciclos reais, poderão ser armazenadas datas exatas
por fatura:

```text
início do ciclo
fim do ciclo
data de fechamento
data de vencimento
```

## 6. Mês de referência

Ao cadastrar um gasto no cartão, o mês de referência deve ser informado ou
gerado sempre que for conhecido.

Quando informado:

```text
reference_month + deslocamento do cartão
→ payment_month
→ mês da fatura
```

Exemplo:

```text
Referência: novembro/2026
Deslocamento: 1 mês
Vencimento nominal do cartão: dia 6

Fatura calculada: dezembro/2026
Previsão nominal: dia 6
```

Quando a referência não for conhecida:

- a referência permanece ausente;
- a API não inventa uma data ou mês de consumo;
- o gasto deve ser associado a uma fatura de pagamento;
- o cálculo de dinheiro livre utiliza o mês dessa fatura.

A ausência de referência não impede o gasto de participar da fatura nem dos
cálculos financeiros. Entretanto, ela não deve ser utilizada para itens
recorrentes ou pontuais cuja competência seja conhecida.

## 7. Cadastro direto em uma fatura

O usuário poderá conhecer a fatura correta mesmo sem informar o mês de
referência.

Exemplo:

```text
Ajuste de fatura: R$ 180
Fatura: dezembro/2026
Referência: não informada
```

Nesse caso, o mês de pagamento é explícito porque o gasto foi associado
diretamente à fatura. Isso não significa que o usuário informou manualmente
uma data de pagamento arbitrária.

A fatura continua sendo uma entidade do cartão, com mês de pagamento e
vencimento nominal derivados da configuração:

```text
Fatura: dezembro/2026
Dia nominal de vencimento: 6
Data efetiva de vencimento: não informada
```

## 8. Gastos recorrentes

Um compromisso recorrente vinculado ao cartão deve gerar participações nas
faturas futuras conforme sua vigência e o deslocamento do cartão.

Exemplo:

```text
Gasolina projetada: R$ 700 por mês
Vigência inicial: novembro/2026
Cartão: Nubank
Deslocamento: 1 mês
```

Resultado:

```text
Referência novembro → fatura dezembro
Referência dezembro → fatura janeiro
Referência janeiro → fatura fevereiro
```

O total da fatura é derivado dos compromissos vinculados. A fatura não é uma
despesa adicional e não pode provocar dupla contagem.

Para compromissos recorrentes, o mês de referência faz parte da ocorrência
mensal gerada pelo sistema e não deve permanecer ausente.

## 9. Competência e recebimento da renda

A renda pode possuir um mês de referência diferente do mês em que fica
disponível.

Exemplo:

```text
Salário referente a novembro/2026
Recebido no primeiro dia útil de dezembro/2026
```

Na visão por competência, essa renda participa de novembro. Na visão de caixa,
ela participa de dezembro.

```text
Análise por competência → mês de referência da renda
Fluxo de caixa → mês de recebimento
```

Quando referência e recebimento ocorrerem no mesmo mês, os dois campos poderão
ter a mesma competência. A separação não obriga o usuário a pensar de uma única
forma; ela permite que a API apresente ambas as leituras.

## 10. Exceções e correções

O usuário deve poder mover um gasto para outra fatura quando a alocação padrão
não representar a realidade.

Exemplos:

- compra processada somente no ciclo seguinte;
- ajuste feito pela instituição;
- cobrança antecipada;
- lançamento cadastrado diretamente em uma fatura;
- correção de informação anterior.

A movimentação deve preservar:

- fatura calculada ou selecionada anteriormente;
- nova fatura;
- data da alteração;
- motivo opcional;
- usuário responsável.

Uma correção não deve apagar silenciosamente o histórico.

## 11. Regras de cálculo

### Dinheiro livre

O dinheiro livre representa disponibilidade e, por isso, utiliza a visão de
caixa:

```text
Dinheiro livre planejado do mês
= rendas previstas para recebimento no mês
− compromissos com pagamento previsto no mês
− faturas com pagamento previsto no mês
− valor planejado para guardar no mês

Dinheiro livre realizado do mês
= rendas efetivamente recebidas no mês
− compromissos efetivamente pagos no mês
− faturas efetivamente pagas no mês
− valor efetivamente guardado no mês
```

Os itens de uma fatura participam do detalhamento, mas o mesmo valor não pode
ser subtraído novamente como uma despesa independente.

### Análise por competência

Quando houver mês de referência:

```text
rendas da competência → mês de referência da renda
consumos da competência → mês de referência do gasto
```

Essa visão pode apresentar um resultado planejado por competência, mas não deve
nomeá-lo como saldo disponível em conta.

Quando não houver mês de referência, o produto deve indicar que a análise
daquele item não possui competência informada.

### Bases de visualização

Sempre que houver informação suficiente, a API deve permitir consolidação por:

- `reference`: competência econômica da renda e do consumo;
- `cash`: recebimento da renda e pagamento dos compromissos.

Todo resumo deve informar explicitamente qual base foi utilizada. A mesma
ocorrência pode aparecer nas duas análises, mas só pode participar uma vez de
cada total.

## 12. Comportamento esperado da API

A API deve:

- calcular o mês padrão da fatura a partir da referência e do deslocamento
  mensal do cartão;
- apresentar o vencimento nominal sem tratá-lo como data efetiva confirmada;
- exigir ou gerar referência quando ela for conhecida;
- permitir referência ausente em ajustes cuja competência seja desconhecida;
- permitir associação direta a uma fatura;
- retornar de forma explícita a origem da alocação;
- permitir movimentação consciente entre faturas;
- preservar histórico;
- impedir dupla contagem;
- utilizar o mês de pagamento no cálculo do dinheiro livre;
- distinguir referência e recebimento de renda;
- permitir consolidações por competência e por caixa;
- informar a base utilizada em cada resumo;
- distinguir vencimento nominal de datas efetivas.

Exemplo conceitual:

```json
{
  "name": "Gasolina",
  "amount": 70000,
  "reference_month": "2026-11",
  "card_id": "card-id",
  "invoice": {
    "month": "2026-12",
    "nominal_due_day": 6,
    "effective_due_date": null,
    "payment_date": null,
    "allocation": "calculated_from_reference"
  }
}
```

Exemplo sem referência:

```json
{
  "name": "Ajuste da fatura",
  "amount": 18000,
  "reference_month": null,
  "card_id": "card-id",
  "invoice": {
    "month": "2026-12",
    "nominal_due_day": 6,
    "effective_due_date": null,
    "payment_date": null,
    "allocation": "selected_invoice"
  }
}
```

Exemplo conceitual de renda:

```json
{
  "name": "Salário",
  "amount": 600000,
  "reference_month": "2026-11",
  "receipt_month": "2026-12"
}
```

Exemplo conceitual de resumo:

```json
{
  "month": "2026-11",
  "basis": "reference"
}
```

Os nomes definitivos dos campos e endpoints serão definidos na especificação
técnica da release correspondente.

## 13. Evolução futura

O modelo poderá evoluir para ciclos diários quando houver uma fonte confiável,
como:

- informação manual detalhada do ciclo;
- importação de fatura;
- integração com instituição financeira;
- Open Finance;
- registro opcional de compras individuais.

Essa evolução não deve alterar a regra principal: o dinheiro livre é afetado
no mês em que o pagamento compromete a disponibilidade financeira.

## 14. Resumo da decisão

1. O mês de referência deve ser armazenado quando for conhecido.
2. A referência pode permanecer ausente em ajustes cuja competência seja
   desconhecida.
3. A ausência de referência não é convertida em uma data inventada.
4. O mês de pagamento afeta o dinheiro livre.
5. A renda distingue mês de referência e mês de recebimento.
6. A API oferece análises por competência e por caixa.
7. Todo resumo informa explicitamente a base utilizada.
8. O cartão possui dia nominal de vencimento.
9. Datas efetivas de vencimento e pagamento são opcionais.
10. O fechamento não é obrigatório na primeira versão.
11. A regra padrão desloca a referência para a fatura do mês seguinte.
12. O usuário pode associar um gasto diretamente a uma fatura.
13. Exceções podem ser movidas com rastreabilidade.
14. Fatura é agrupador, não uma segunda despesa.
15. Ciclos exatos poderão ser incorporados futuramente com dados confiáveis.
