# Prompt inicial para o Codex — Planejador Financeiro

Quero iniciar um projeto de planejador financeiro pessoal.

Antes de propor arquitetura, escrever código ou criar tarefas, leia integralmente o arquivo:

`documentacao-produto-planejador-financeiro.md`

Esse documento é a fonte de verdade sobre a visão do produto, conceitos, regras, módulos, fluxos, critérios de aceite, roadmap e decisões já tomadas.

## Contexto essencial

O produto nasceu de planilhas anuais utilizadas durante três anos para:

- projetar renda e compromissos;
- acompanhar dívidas;
- calcular quanto guardar;
- calcular quanto permanece livre;
- consolidar custos previsíveis por cartão;
- projetar patrimônio;
- comparar meses;
- ajustar o planejamento quando a vida muda.

O aplicativo deve tornar essa lógica mais fácil de cadastrar, manter e explicar.

## Princípios inegociáveis

1. Todo número deve ser explicável.
2. Cartão é forma de pagamento e agrupador, não uma despesa.
3. Projetado e realizado são informações distintas.
4. Informar um realizado não altera automaticamente o futuro.
5. Mudanças de valor criam períodos de vigência e não apagam o passado.
6. O aplicativo apresenta números; o usuário interpreta.
7. Não adicionar IA, aconselhamento ou julgamento como parte do núcleo.
8. Não exigir o cadastro de cada transação individual.
9. Priorizar valores mensais consolidados.
10. Construir por fatias verticais, sem tentar implementar a visão inteira de uma vez.

## Primeiro objetivo funcional

A primeira versão deve responder:

> Com minha renda, meus compromissos e o valor que quero guardar, quanto tenho realmente livre neste mês?

Ela deve permitir, inicialmente:

- cadastrar renda;
- cadastrar despesas fixas;
- cadastrar despesas variáveis projetadas;
- cadastrar dívidas parceladas;
- vincular itens a cartões;
- definir valor mensal para guardar;
- calcular compromissos;
- calcular dinheiro livre;
- detalhar a composição das faturas;
- projetar os próximos meses;
- mostrar quando um compromisso terminará.

## Cenário principal para validar o produto

Use como cenário de referência:

- salário que muda a partir de determinado mês;
- períodos anteriores com outra renda ou sem renda;
- seguro veicular com mudança de valor e descrição do período;
- gasolina com valor projetado e realizado;
- dívida temporária que termina em dezembro;
- múltiplos cartões;
- valor planejado para guardar;
- cálculo do dinheiro livre;
- projeção até o fim do ano.

## Forma de trabalho solicitada

1. Comece resumindo o entendimento do domínio.
2. Identifique somente as decisões de produto que realmente bloqueiam a primeira fatia.
3. Proponha uma divisão em pequenas entregas verticais.
4. Para cada entrega, liste:
   - objetivo do usuário;
   - regras envolvidas;
   - dados necessários;
   - casos de borda;
   - critérios de aceite.
5. Não defina stack sem uma solicitação explícita.
6. Não implemente toda a aplicação de uma vez.
7. Ao encontrar ambiguidade, consulte as seções “Questões de produto ainda abertas” e “Decisões de produto já tomadas”.
8. Atualize a documentação sempre que uma decisão nova alterar o domínio.

O primeiro resultado esperado é um backlog refinado da Fatia 1: “Quanto está livre?”, sem código.
