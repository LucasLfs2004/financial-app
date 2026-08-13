# Documentação de Produto — Planejador Financeiro Pessoal

> **Nome do produto:** a definir  
> **Nome provisório neste documento:** Planejador Financeiro  
> **Status:** visão consolidada do produto  
> **Origem:** histórico de três anos de planejamento financeiro pessoal em planilhas, conversa de descoberta do produto e planilha de projeção de 2026  
> **Objetivo deste documento:** registrar a visão completa do aplicativo, seus conceitos, regras, módulos, fluxos, requisitos e fases de entrega, servindo como fonte principal de contexto para produto, design e implementação  
> **Escopo técnico:** este documento não define linguagem, framework, banco de dados, infraestrutura ou arquitetura de software

---

## 1. Resumo executivo

O Planejador Financeiro é um aplicativo pessoal para organizar a vida financeira a partir de projeções mensais e anuais.

O produto nasce de uma necessidade prática: entender, antes de gastar, quanto da renda já está comprometido com despesas, dívidas, parcelas e objetivos de economia; quanto pode ser guardado; e quanto permanece realmente livre para uso durante o mês.

A proposta não é substituir um banco, importar automaticamente cada compra ou atuar como consultor financeiro. O aplicativo deve permitir que o próprio usuário forneça premissas simples sobre sua vida financeira e receba, em troca, uma visão organizada, rastreável e explicável de seu passado, presente e futuro.

O sistema transforma informações como:

- salário mensal;
- rendas extras;
- despesas fixas;
- estimativas de despesas variáveis;
- dívidas e parcelas;
- cartões de crédito;
- investimentos;
- metas de economia;
- mudanças de valores ao longo do tempo;
- valores efetivamente realizados;

em visualizações como:

- total comprometido no mês;
- valor planejado para guardar;
- valor realmente livre para gastar;
- composição projetada das faturas;
- comparação entre projetado e realizado;
- evolução histórica de custos;
- despesas que terminarão em breve;
- projeção de patrimônio no fim do ano;
- impacto numérico de mudanças de renda ou estilo de vida.

O princípio central do produto é:

> **Todo número exibido deve ser explicável.**

O segundo princípio central é:

> **O aplicativo apresenta os dados; o usuário interpreta a própria vida.**

O produto não deve dizer que uma decisão foi certa ou errada, não deve transformar números em sermões e não deve depender de inteligência artificial para gerar valor. Seu papel é organizar dados pessoais com clareza suficiente para que o usuário chegue às próprias conclusões.

---

## 2. Origem do produto

Durante aproximadamente três anos, o idealizador do produto criou planilhas anuais para projetar sua vida financeira.

Essas planilhas buscavam responder perguntas como:

- Quanto da minha renda já está comprometido?
- Quais dívidas ainda estão ativas?
- Quando cada parcela terminará?
- Quanto preciso guardar por mês?
- Quanto posso gastar livremente sem comprometer minha meta?
- Quanto devo ter acumulado ao fim do ano?
- Quanto meus investimentos podem render?
- Qual será o impacto de uma mudança de salário?
- Qual cartão concentrará cada custo?
- Quanto cada fatura deverá representar nos próximos meses?

A planilha de 2026, criada em dezembro de 2025, já continha elementos como:

- salário mensal;
- renda extra;
- total a pagar;
- valor a guardar;
- balanço do mês;
- taxas de CDI e Selic;
- aportes em diferentes plataformas;
- percentuais do CDI;
- rendimentos projetados;
- patrimônio acumulado;
- meta financeira;
- despesas como seguro, pedágio, academia e combustível;
- valores consolidados por cartão;
- compromissos com término previsto, como IPVA e dívidas.

A planilha conseguia realizar os cálculos, mas apresentava uma limitação importante: as relações entre os números estavam escondidas nas fórmulas.

Uma linha como “Cartão Itaú” podia ser a soma de vários custos previsíveis, mas somente quem criou a planilha sabia quais despesas formavam aquele valor. Uma alteração de salário, dívida, parcela ou estimativa exigia encontrar e revisar células, fórmulas e meses relacionados.

Mudanças reais de vida também tornavam a manutenção difícil. Por exemplo:

- perda inesperada de emprego;
- entrada em uma nova empresa;
- alteração de salário;
- novas despesas;
- mudança de prioridades;
- fim ou início de dívidas;
- troca de veículo;
- alteração de seguro;
- mudança na capacidade de poupança.

O produto surge para transformar a lógica dessas planilhas em um sistema mais simples de alimentar, manter, explicar e analisar.

---

## 3. Problema do usuário

### 3.1 Problema principal

O saldo disponível em uma conta não representa necessariamente o dinheiro que a pessoa pode gastar.

Parte desse saldo já pode estar destinada a:

- contas fixas;
- parcelas;
- dívidas;
- faturas futuras;
- despesas variáveis previstas;
- investimentos;
- metas pessoais;
- compromissos que ainda vencerão.

Sem uma visão clara dessas destinações, o usuário pode gastar um dinheiro que parecia disponível, mas já estava comprometido com seu planejamento.

### 3.2 Problemas das planilhas

Planilhas conseguem resolver o problema matemático, mas frequentemente criam dificuldades operacionais:

- fórmulas pouco legíveis;
- dependências invisíveis entre células;
- duplicação manual de meses;
- dificuldade de alterar uma premissa;
- ausência de histórico estruturado;
- risco de sobrescrever valores anteriores;
- dificuldade de explicar a origem de um total;
- dificuldade de distinguir projeção de valor realizado;
- dificuldade de comparar períodos;
- grande esforço para manter o planejamento atualizado.

### 3.3 Problemas de aplicativos financeiros tradicionais

Muitos aplicativos financeiros concentram-se em registrar o passado:

- quanto foi gasto;
- em qual estabelecimento;
- em qual categoria;
- qual era o saldo.

Esse acompanhamento pode ser útil, mas não responde integralmente às perguntas de planejamento:

- Quanto posso gastar neste mês?
- Quanto preciso guardar?
- O que já está comprometido?
- Quando minha capacidade de poupança aumentará?
- Quanto devo ter ao fim do ano?
- Como uma decisão tomada hoje altera os próximos meses?
- Qual parte da fatura é recorrente, parcelada ou variável?
- Como meus custos mudaram durante diferentes períodos da vida?

Outros aplicativos tentam interpretar os dados pelo usuário, com alertas e avaliações prescritivas. A proposta deste produto é diferente: oferecer dados objetivos e permitir que o próprio usuário faça a análise.

---

## 4. Visão do produto

### 4.1 Visão resumida

Permitir que uma pessoa monte seu planejamento financeiro como montaria uma planilha, mas sem precisar criar fórmulas, duplicar meses ou memorizar quais dados formam cada total.

### 4.2 Proposta de valor

> **Planeje quanto guardar, entenda para onde seu dinheiro vai e descubra quanto pode gastar sem comprometer seus objetivos.**

Outra forma de expressar a proposta:

> **Saiba quanto do seu dinheiro está comprometido, quanto deve ser guardado e quanto está realmente livre.**

### 4.3 Transformação prometida

Antes do produto:

- o usuário conhece seu saldo, mas não sabe quanto está realmente livre;
- totais dependem de fórmulas difíceis de manter;
- mudanças quebram ou invalidam projeções;
- o histórico é perdido ao editar valores;
- valores esperados e reais se misturam;
- decisões financeiras dependem de memória.

Depois do produto:

- o usuário visualiza sua renda distribuída;
- compromissos possuem origem rastreável;
- faturas são formadas por itens conhecidos;
- alterações futuras recalculam o planejamento;
- mudanças não apagam o passado;
- projeções e valores realizados coexistem;
- o usuário pode analisar sua trajetória com seus próprios dados.

---

## 5. Objetivos do produto

### 5.1 Objetivo principal

Responder com clareza:

> **Depois de considerar renda, compromissos e objetivo de economia, quanto posso gastar livremente neste mês sem comprometer meu planejamento?**

### 5.2 Objetivos secundários

1. Permitir a construção de uma projeção mensal e anual.
2. Manter uma visão completa das dívidas ativas.
3. Mostrar quando cada dívida ou compromisso terminará.
4. Consolidar custos projetados por cartão sem esconder sua composição.
5. Comparar valores projetados com valores efetivamente realizados.
6. Preservar o histórico de alterações de renda e despesas.
7. Permitir análises históricas baseadas em fatos e números.
8. Projetar capacidade de economia e patrimônio ao fim do período.
9. Reduzir o esforço necessário para atualizar um planejamento.
10. Permitir que o usuário investigue qualquer total até chegar aos itens de origem.

### 5.3 Objetivos que não pertencem ao produto inicial

O produto não nasce para:

- substituir uma conta bancária;
- executar pagamentos;
- fornecer crédito;
- recomendar investimentos;
- classificar decisões como boas ou ruins;
- importar cada transação automaticamente;
- oferecer aconselhamento financeiro por inteligência artificial;
- calcular impostos;
- atuar como sistema contábil;
- exigir registro de cada compra cotidiana.

---

## 6. Princípios do produto

### 6.1 Todo número deve ser explicável

Qualquer total deve permitir detalhamento.

Se a dashboard mostrar R$ 2.300 de compromissos, o usuário deve conseguir abrir o valor e visualizar os itens que o formam.

Se um cartão apresentar fatura projetada de R$ 1.400, deve ser possível identificar exatamente quais despesas, dívidas e parcelas compõem esse total.

### 6.2 O sistema não julga

O aplicativo pode mostrar:

- aumento de R$ 380 por mês;
- impacto anual de R$ 4.560;
- custo equivalente a 24% da renda;
- realizado 12% acima do projetado.

O aplicativo não deve afirmar:

- “Você errou.”
- “Você gastou demais.”
- “Essa compra foi irresponsável.”
- “Você precisa reduzir seu padrão de vida.”
- “Sua saúde financeira está ruim.”

O contexto da vida pertence ao usuário.

### 6.3 Projeção e realidade devem coexistir

A previsão original não deve desaparecer quando o valor real for informado.

O produto deve preservar:

- quanto o usuário esperava gastar;
- quanto realmente gastou;
- qual foi a diferença;
- se a projeção futura foi mantida ou alterada.

### 6.4 O passado não deve ser reescrito silenciosamente

Alterar o valor atual de um seguro, salário ou assinatura não deve substituir os valores históricos.

Mudanças devem criar novos períodos de vigência.

### 6.5 Simples na superfície, profundo no detalhamento

A dashboard inicial deve ser direta.

O detalhamento deve estar disponível por meio de navegação progressiva:

1. total;
2. categoria;
3. item;
4. período;
5. origem ou contexto da mudança.

### 6.6 O usuário controla as premissas

O sistema pode apresentar médias e comparações, mas não deve alterar automaticamente projeções futuras.

Se o realizado de gasolina foi maior durante três meses, o produto pode oferecer a opção de atualizar a estimativa, mas a decisão pertence ao usuário.

### 6.7 O produto deve respeitar a vida real

A vida financeira não é estática.

O produto deve representar:

- mudanças de emprego;
- mudanças salariais;
- períodos sem renda;
- novas dívidas;
- quitação antecipada;
- despesas temporárias;
- troca de veículo;
- mudanças de cartão;
- eventos extraordinários;
- meses atípicos.

---

## 7. Público e perfil de uso

### 7.1 Usuário principal

Pessoa que:

- recebe renda mensal;
- possui custos fixos e variáveis;
- utiliza um ou mais cartões;
- possui ou pode possuir dívidas parceladas;
- deseja guardar dinheiro;
- quer planejar o ano;
- aceita informar manualmente valores consolidados;
- prefere entender os números a receber conselhos automáticos;
- atualmente utiliza planilhas, anotações ou cálculos mentais.

### 7.2 Perfil comportamental

O produto atende especialmente pessoas que pensam:

- “Quero saber quanto posso gastar sem culpa.”
- “Preciso visualizar todas as minhas dívidas.”
- “Quero chegar ao fim do ano com determinado valor.”
- “Não quero controlar cada cafezinho.”
- “Quero entender como meus custos mudaram.”
- “Quero saber por que minha fatura terá esse valor.”
- “Quero ajustar uma premissa e ver o restante recalculado.”

### 7.3 Uso inicial pessoal

O produto pode nascer como ferramenta de uso individual do criador, sem necessidade imediata de monetização.

Mesmo assim, o domínio deve ser documentado como produto reutilizável, evitando regras rígidas que funcionem apenas para uma pessoa.

---

## 8. Trabalhos que o usuário deseja realizar

### 8.1 Planejar um mês

Quando iniciar ou atualizar seu planejamento, o usuário quer informar renda, compromissos e objetivo de economia para saber quanto permanece livre.

### 8.2 Planejar um ano

Quando definir metas anuais, o usuário quer visualizar os resultados de cada mês e o patrimônio projetado ao fim do ano.

### 8.3 Entender uma fatura

Quando visualizar o total de um cartão, o usuário quer saber quais custos previsíveis formam esse valor.

### 8.4 Acompanhar dívidas

Quando possuir parcelas ou compromissos temporários, o usuário quer saber:

- saldo ou quantidade restante;
- parcelas ativas;
- mês de término;
- valor que será liberado após o término.

### 8.5 Registrar o mês real

Quando um mês avançar ou terminar, o usuário quer informar valores realizados de despesas variáveis e comparar com o que havia projetado.

### 8.6 Alterar o futuro sem apagar o passado

Quando um custo mudar, o usuário quer registrar a nova realidade a partir de uma data, preservando os períodos anteriores.

### 8.7 Analisar sua trajetória

Quando olhar para anos anteriores, o usuário quer enxergar mudanças financeiras e seus impactos numéricos.

### 8.8 Testar um plano

Quando considerar uma mudança, o usuário quer entender se sua renda e seus compromissos comportam aquela decisão.

---

## 9. Modelo mental do produto

O produto deve representar a vida financeira por meio de cinco camadas.

### 9.1 Premissas fornecidas pelo usuário

Exemplos:

- salário líquido de R$ 6.000 por mês;
- transplante de R$ 600 até dezembro;
- gasolina projetada em R$ 700;
- seguro do carro de R$ 350;
- meta de guardar R$ 1.500;
- dívida paga no cartão Nubank;
- investimento com aporte mensal.

### 9.2 Períodos de validade

Cada premissa pode valer somente durante determinado intervalo.

Exemplos:

- salário antigo até março;
- novo salário a partir de julho;
- seguro do Civic até julho;
- seguro do Golf a partir de agosto;
- transplante até dezembro;
- IPVA durante cinco meses.

### 9.3 Ocorrências mensais

O sistema interpreta cada premissa e identifica o que pertence a cada mês.

Exemplo:

- transplante aparece de julho a dezembro;
- seguro aparece durante todos os meses de vigência;
- IPVA aparece apenas nas parcelas aplicáveis;
- renda extra aparece em um mês específico;
- gasolina aparece mensalmente enquanto a projeção estiver ativa.

### 9.4 Consolidações

As ocorrências são agrupadas por:

- mês;
- categoria;
- cartão;
- tipo;
- dívida;
- situação;
- planejado e realizado.

### 9.5 Visualizações

Os mesmos dados alimentam:

- dashboard mensal;
- planejamento anual;
- faturas projetadas;
- histórico;
- gráficos;
- metas;
- fechamento mensal;
- comparações.

---

## 10. Conceitos fundamentais e glossário

### 10.1 Item financeiro

Entidade que representa algo relevante na vida financeira do usuário.

Exemplos:

- salário;
- seguro veicular;
- gasolina;
- transplante;
- academia;
- financiamento;
- aporte;
- renda extra.

Um item possui identidade própria e pode ter diferentes valores ao longo do tempo.

### 10.2 Período financeiro

Intervalo em que determinadas características de um item são válidas.

Pode armazenar:

- valor esperado;
- data inicial;
- data final;
- descrição;
- cartão;
- recorrência;
- contexto;
- observação.

Exemplo:

| Item | Início | Fim | Valor | Descrição |
|---|---:|---:|---:|---|
| Seguro veicular | out/2024 | jul/2025 | R$ 220 | Seguro do Civic |
| Seguro veicular | ago/2025 | jan/2026 | R$ 350 | Seguro do Golf |
| Seguro veicular | fev/2026 | atual | R$ 315 | Renovação |

### 10.3 Valor projetado

Valor que o usuário espera para determinado mês.

Não significa que o dinheiro já foi efetivamente gasto ou recebido.

### 10.4 Valor realizado

Valor que efetivamente ocorreu em determinado mês.

Pode ser informado manualmente e deve coexistir com o projetado.

### 10.5 Diferença

Resultado de:

```text
valor realizado − valor projetado
```

- resultado positivo em uma despesa: realizado acima do projetado;
- resultado negativo em uma despesa: realizado abaixo do projetado;
- para receitas, a interpretação pode ser apresentada explicitamente para evitar ambiguidade.

### 10.6 Variação percentual

```text
(realizado − projetado) ÷ projetado × 100
```

Quando o projetado for zero, o percentual não deve ser calculado como número comum. O sistema deve apresentar “não aplicável” ou informação equivalente.

### 10.7 Renda

Valor recebido pelo usuário.

Pode ser:

- recorrente;
- temporária;
- pontual;
- fixa;
- variável;
- salário;
- renda extra;
- bônus;
- PLR;
- reembolso;
- outra entrada.

### 10.8 Compromisso

Valor que reduz a disponibilidade financeira do usuário.

Pode ser:

- despesa fixa;
- estimativa variável;
- dívida;
- parcela;
- pagamento pontual;
- reserva planejada;
- aporte.

### 10.9 Despesa fixa

Custo previsível e geralmente recorrente.

Exemplos:

- seguro;
- academia;
- assinatura;
- aluguel;
- plano de saúde.

“Fixa” não significa que o valor nunca muda. Significa apenas que o custo é recorrente e conhecido dentro de um período.

### 10.10 Despesa variável projetada

Custo cujo valor exato não é conhecido antecipadamente, mas para o qual o usuário reserva uma estimativa.

Exemplos:

- gasolina;
- mercado;
- lazer;
- estacionamento;
- alimentação.

### 10.11 Dívida

Compromisso com prazo, saldo ou quantidade de parcelas conhecida.

Pode conter:

- valor total;
- valor mensal;
- número total de parcelas;
- parcela atual;
- início;
- término;
- pagamento;
- descrição;
- possibilidade de quitação antecipada.

### 10.12 Forma de pagamento

Meio utilizado para pagar uma despesa.

Exemplos:

- cartão de crédito;
- débito em conta;
- Pix;
- boleto;
- dinheiro.

### 10.13 Cartão de crédito

Forma de pagamento e agrupador de compromissos.

O cartão não deve ser tratado como categoria de despesa. O total da fatura deve ser derivado dos itens vinculados a ele.

### 10.14 Categoria

Classificação econômica do item.

Exemplos:

- moradia;
- veículo;
- saúde;
- alimentação;
- lazer;
- educação;
- assinaturas;
- dívidas;
- investimentos.

### 10.15 Valor planejado para guardar

Valor que o usuário deseja separar no mês para economia ou investimento.

Esse valor reduz o dinheiro livre, mesmo que não seja classificado como despesa de consumo.

### 10.16 Dinheiro livre

Valor realmente disponível para uso sem comprometer as premissas atuais.

Fórmula conceitual:

```text
renda prevista
+ renda extra prevista
− compromissos previstos
− valor planejado para guardar
= dinheiro livre planejado
```

### 10.17 Patrimônio projetado

Estimativa de quanto o usuário terá acumulado no fim de um período, considerando saldos iniciais, aportes e rendimentos projetados.

### 10.18 Planejamento original

Versão das premissas consideradas quando o plano foi criado ou formalmente salvo como referência.

### 10.19 Estimativa atualizada

Combinação de:

```text
valores realizados nos meses encerrados
+ projeções atuais dos meses em aberto e futuros
```

### 10.20 Fechamento mensal

Processo em que o usuário revisa o mês, registra valores realizados e confirma o resultado daquele período.

### 10.21 Evento de mudança

Registro de alteração relevante em uma premissa.

Exemplos:

- aumento do seguro;
- mudança de salário;
- troca de veículo;
- fim de uma dívida;
- mudança de cartão;
- atualização de projeção.

### 10.22 Contexto do período

Descrição que explica a realidade representada por um período.

Exemplos:

- “Seguro do Civic”;
- “Seguro do Golf”;
- “Salário Entrepay”;
- “Salário Appmax”;
- “Ajuste após aumento de deslocamento”;
- “Parcela do transplante”.

---

## 11. Fórmulas conceituais do produto

As fórmulas abaixo descrevem o comportamento esperado. Não definem implementação técnica.

### 11.1 Compromissos projetados do mês

```text
soma das despesas fixas vigentes
+ soma das despesas variáveis projetadas vigentes
+ soma das parcelas de dívidas vigentes
+ soma dos pagamentos pontuais daquele mês
```

### 11.2 Fatura projetada de um cartão

```text
soma dos compromissos do mês vinculados ao cartão
```

O total não deve ser armazenado como informação independente quando puder ser derivado dos itens.

### 11.3 Dinheiro livre planejado

```text
renda prevista do mês
+ renda extra prevista
− compromissos projetados
− valor planejado para guardar
```

### 11.4 Dinheiro livre realizado

```text
renda realizada
+ renda extra realizada
− compromissos realizados
− valor efetivamente guardado
```

### 11.5 Valor restante para gastar

Durante o mês:

```text
dinheiro livre planejado
− gastos livres já realizados
```

A definição de “gastos livres” deverá ser consistente com a forma de fechamento adotada pelo usuário.

### 11.6 Estimativa anual atualizada

```text
realizado dos meses fechados
+ projeção dos meses abertos e futuros
```

### 11.7 Patrimônio acumulado projetado

```text
saldo acumulado anterior
+ aporte do mês
+ rendimento projetado do mês
```

### 11.8 Impacto anual aproximado de uma mudança recorrente

```text
diferença mensal × quantidade de meses aplicáveis
```

### 11.9 Percentual da renda comprometida

```text
compromissos do mês ÷ renda do mês × 100
```

Quando a renda for zero, o sistema deve evitar divisões inválidas e apresentar o dado com contexto.

---

## 12. Módulos funcionais

## 12.1 Configuração inicial do planejamento

### Objetivo

Permitir que o usuário crie rapidamente uma primeira projeção utilizável.

### Informações mínimas

- mês inicial;
- horizonte de planejamento;
- renda principal;
- valor que deseja guardar;
- principais compromissos;
- cartões utilizados.

### Resultado esperado

Ao terminar a configuração inicial, o usuário deve visualizar:

- renda do mês;
- compromissos;
- valor planejado para guardar;
- dinheiro livre;
- composição dos cartões;
- próximos meses projetados.

### Requisito de experiência

O usuário não deve precisar configurar categorias, regras avançadas ou investimentos para obter valor inicial.

---

## 12.2 Planejamentos e horizontes

### Objetivo

Organizar projeções em períodos coerentes, normalmente anuais.

### Comportamentos

O usuário poderá:

- criar um planejamento anual;
- selecionar o ano ou período;
- visualizar meses anteriores e futuros;
- manter histórico de anos;
- copiar premissas para um novo ano;
- ajustar valores sem perder o planejamento original;
- comparar planejamento original e estimativa atualizada.

### Observação

O aplicativo deve evitar transformar cada ano em uma ilha sem relação histórica. Itens como seguro e salário podem atravessar diferentes planejamentos.

---

## 12.3 Receitas

### Tipos suportados

- salário;
- renda extra;
- bônus;
- PLR;
- décimo terceiro;
- reembolso;
- venda de bem;
- outra receita.

### Comportamentos

O usuário poderá:

- cadastrar receita recorrente;
- cadastrar receita pontual;
- definir início e término;
- alterar o valor a partir de uma data;
- registrar o valor efetivamente recebido;
- adicionar descrição do período;
- encerrar uma renda;
- visualizar histórico.

### Exemplo

```text
Salário

Outubro/2024 a março/2026
Empresa anterior
R$ X

Julho/2026 em diante
Nova empresa
R$ Y
```

### Regras

- uma mudança salarial cria novo período;
- meses sem renda devem ser representáveis;
- bônus e PLR não devem exigir recorrência;
- receitas projetadas e realizadas devem ser comparáveis.

---

## 12.4 Despesas fixas

### Exemplos

- seguro;
- aluguel;
- academia;
- assinatura;
- pedágio;
- internet;
- plano de saúde.

### Comportamentos

O usuário poderá:

- cadastrar valor recorrente;
- informar frequência;
- definir período;
- vincular forma de pagamento;
- alterar valor a partir de determinada data;
- registrar descrição por período;
- encerrar a despesa;
- pausar quando necessário;
- informar realizado mensal.

### Regra principal

Alterar uma despesa não deve apagar seu histórico.

---

## 12.5 Despesas variáveis projetadas

### Exemplos

- gasolina;
- mercado;
- alimentação;
- lazer;
- estacionamento;
- manutenção.

### Comportamentos

O usuário poderá:

- definir uma estimativa padrão;
- aplicar a estimativa aos meses futuros;
- registrar valor real de cada mês;
- comparar projetado e realizado;
- manter a projeção futura mesmo após variação;
- atualizar meses futuros por decisão própria;
- criar novo período de projeção;
- adicionar observação para um mês atípico.

### Exemplo

```text
Gasolina

Projeção: R$ 700 por mês

Julho
Projetado: R$ 700
Realizado: R$ 845
Diferença: +R$ 145
Variação: +20,7%

Agosto
Projetado: R$ 700
Realizado: não informado
```

### Regra principal

O sistema não altera automaticamente a projeção futura com base no realizado.

Pode oferecer:

```text
Média realizada dos últimos 3 meses: R$ 765
Projeção atual: R$ 700

[Manter R$ 700]
[Atualizar os próximos meses]
```

---

## 12.6 Dívidas e parcelas

### Objetivo

Dar visibilidade completa sobre compromissos temporários.

### Informações

- nome;
- descrição;
- valor total, quando conhecido;
- valor da parcela;
- quantidade total;
- parcela atual;
- data inicial;
- data final;
- forma de pagamento;
- categoria;
- status;
- observações.

### Comportamentos

O usuário poderá:

- cadastrar dívida;
- visualizar parcelas futuras;
- registrar pagamento;
- corrigir parcela;
- quitar antecipadamente;
- suspender;
- renegociar;
- alterar cartão;
- visualizar quanto será liberado após o término.

### Exemplo

```text
Transplante
R$ 600 por mês
Julho a dezembro

A partir de janeiro:
R$ 600 deixam de estar comprometidos
```

### Estados possíveis

- planejada;
- ativa;
- concluída;
- quitada antecipadamente;
- suspensa;
- renegociada;
- cancelada.

### Regra de renegociação

Uma renegociação não deve apagar a dívida anterior. Deve encerrar a condição antiga e criar uma nova condição relacionada.

---

## 12.7 Cartões de crédito

### Papel no produto

Cartão é uma forma de pagamento e uma dimensão de análise.

### Informações

- nome;
- instituição;
- identificador visual opcional;
- limite opcional;
- dia de fechamento opcional;
- dia de vencimento opcional;
- status;
- período de uso.

### Comportamentos

O usuário poderá:

- vincular itens ao cartão;
- visualizar fatura projetada;
- abrir composição;
- comparar fatura projetada e realizada;
- alterar um item de cartão a partir de uma data;
- encerrar cartão preservando histórico;
- visualizar participação de cada categoria.

### Exemplo de composição

```text
Nubank — agosto/2026

Transplante              R$ 600
Academia                 R$ 150
Assinaturas              R$  90
Gasolina projetada       R$ 500
Parcela de compra        R$ 180
──────────────────────────────
Fatura projetada       R$ 1.520
```

### Regras

- não cadastrar “fatura” como categoria quando ela puder ser derivada;
- permitir ajustes manuais quando a fatura real incluir valores não modelados;
- diferenciar composição explicada de diferença não categorizada;
- o usuário deve conseguir chegar do total ao item de origem.

### Ciclo de cartão

O tratamento completo de fechamento, vencimento e mês da compra pode ser evoluído por fases.

Na primeira versão, o usuário deve conseguir definir em qual mês financeiro um compromisso compõe a fatura. A evolução futura poderá modelar ciclos de fechamento com maior precisão.

---

## 12.8 Formas de pagamento

### Tipos

- cartão;
- débito em conta;
- boleto;
- Pix;
- dinheiro;
- transferência;
- outra.

### Comportamentos

- um item pode trocar de forma de pagamento ao longo do tempo;
- a troca cria novo período;
- formas encerradas permanecem no histórico;
- o usuário pode analisar despesas por forma.

---

## 12.9 Categorias

### Objetivo

Permitir agrupamento sem esconder a identidade do item.

### Categorias iniciais sugeridas

- moradia;
- veículo;
- saúde;
- alimentação;
- lazer;
- educação;
- assinaturas;
- mobilidade;
- família;
- dívidas;
- investimentos;
- impostos;
- outros.

### Comportamentos

- categorias editáveis;
- possibilidade de subcategoria futuramente;
- histórico não deve ser quebrado ao renomear;
- item e categoria são conceitos diferentes;
- cartão nunca deve ser categoria.

---

## 12.10 Meta mensal de economia

### Objetivo

Reservar parte da renda antes de calcular o dinheiro livre.

### Comportamentos

O usuário poderá:

- definir valor mensal;
- alterar por período;
- registrar valor efetivamente guardado;
- vincular a uma ou mais metas;
- visualizar diferença;
- suspender em meses específicos;
- definir valor pontual.

### Exemplo

```text
Planejado para guardar: R$ 1.500
Efetivamente guardado: R$ 1.280
Diferença: -R$ 220
```

---

## 12.11 Metas financeiras

### Exemplos

- patrimônio no fim do ano;
- reserva de emergência;
- viagem;
- entrada de imóvel;
- troca de carro;
- quitação de dívida.

### Informações

- nome;
- valor-alvo;
- data-alvo;
- saldo inicial;
- aportes relacionados;
- descrição;
- status.

### Visualizações

- valor atual;
- valor projetado;
- valor faltante;
- progresso;
- aporte necessário, apenas como cálculo objetivo;
- data estimada com as premissas atuais.

### Linguagem

Evitar avaliações emocionais. Preferir:

```text
Meta: R$ 30.000
Projeção atual: R$ 27.400
Diferença: R$ 2.600
```

---

## 12.12 Investimentos e patrimônio

### Origem da funcionalidade

A planilha original projetava:

- aportes;
- percentuais do CDI;
- taxa CDI;
- taxa Selic;
- rendimento mensal;
- saldo acumulado;
- meta.

### Escopo funcional

O usuário poderá cadastrar:

- conta ou plataforma;
- saldo inicial;
- tipo de referência;
- percentual de uma taxa;
- taxa esperada;
- aportes;
- retiradas;
- rendimento projetado;
- rendimento realizado;
- período;
- objetivo relacionado.

### Exemplos

- Mercado Pago a 120% do CDI;
- caixinha a 115% do CDI;
- caixinha a 100% do CDI.

### Regras

- taxas podem ser informadas manualmente;
- projeção não deve ser apresentada como garantia;
- o sistema deve distinguir aporte de rendimento;
- aportes reduzem dinheiro livre e aumentam patrimônio;
- retiradas aumentam disponibilidade e reduzem patrimônio;
- o usuário deve conseguir comparar projeção e rendimento real.

### Resultado

```text
Saldo anterior
+ aporte
+ rendimento
− retirada
= saldo acumulado
```

---

## 12.13 Valores realizados

### Objetivo

Registrar o que realmente aconteceu sem transformar o aplicativo em rastreador de cada transação.

### Abordagem

O usuário pode informar o total consolidado mensal de um item.

Exemplo:

```text
Gasolina de julho: R$ 530
```

Não é necessário cadastrar:

```text
Posto A: R$ 150
Posto B: R$ 200
Posto C: R$ 180
```

### Comportamentos

- registrar realizado por item e mês;
- manter projetado original;
- adicionar observação;
- corrigir realizado;
- informar realizado igual ao projetado;
- identificar valores pendentes no fechamento;
- comparar total realizado por categoria e cartão.

---

## 12.14 Fechamento mensal

### Objetivo

Consolidar o mês de maneira simples.

### Fluxo

1. usuário abre o fechamento;
2. sistema mostra receitas e despesas esperadas;
3. itens variáveis sem realizado são destacados;
4. usuário informa valores consolidados;
5. usuário confirma valores fixos;
6. sistema calcula diferenças;
7. usuário registra quanto guardou;
8. usuário revisa o dinheiro livre real;
9. mês é fechado.

### Opções para um item sem realizado

- informar valor;
- usar projetado como realizado;
- deixar pendente;
- marcar como não ocorrido;
- mover para outro mês, quando aplicável.

### Depois do fechamento

O mês fechado passa a utilizar valores realizados nas visões atuais, mantendo a possibilidade de consultar o planejamento original.

### Reabertura

O usuário deve conseguir reabrir um mês, com registro claro de que houve alteração posterior.

---

## 12.15 Histórico temporal

### Objetivo

Mostrar como a vida financeira mudou.

### Princípio

Mudanças criam períodos, não sobrescritas.

### Exemplo de seguro

```text
Seguro veicular

Outubro/2024 a julho/2025
R$ 220 por mês
Seguro do Civic

Agosto/2025 a janeiro/2026
R$ 350 por mês
Seguro do Golf

Fevereiro/2026 em diante
R$ 315 por mês
Renovação
```

### Visualizações

- linha do tempo;
- gráfico em degraus;
- tabela de períodos;
- antes e depois;
- impacto mensal;
- impacto anual;
- percentual da renda;
- eventos relacionados no mesmo período.

### Aplicações

- seguro;
- salário;
- aluguel;
- academia;
- carro;
- internet;
- plano de saúde;
- qualquer custo recorrente.

---

## 12.16 Linha do tempo financeira

### Objetivo

Apresentar acontecimentos financeiros de forma cronológica.

### Exemplos de evento

```text
Agosto/2025
Seguro veicular: R$ 220 → R$ 350
Descrição: troca do Civic pelo Golf

Julho/2026
Início de nova renda

Dezembro/2026
Última parcela do transplante
```

### Regras

- exibir fatos, não julgamento;
- permitir filtros;
- permitir abrir origem;
- relacionar alterações simultâneas;
- mostrar impacto quando calculável.

---

## 12.17 Comparação entre períodos

### Objetivo

Permitir que o usuário analise mudanças da própria vida.

### Exemplos

```text
Custo mensal com veículo

Antes: R$ 1.050
Depois: R$ 1.430
Diferença: +R$ 380
Impacto anual: +R$ 4.560
```

```text
Participação do veículo na renda

Antes: 18%
Depois: 24%
```

### Seleções possíveis

- dois meses;
- dois intervalos;
- antes e depois de uma mudança;
- dois anos;
- dois contextos descritos pelo usuário.

---

## 12.18 Cenários e simulações

### Objetivo

Permitir análise do futuro sem alterar o planejamento principal.

### Exemplos

- troca de carro;
- novo aluguel;
- aumento salarial;
- nova dívida;
- antecipação de quitação;
- redução de custo;
- aumento da meta de economia.

### Comportamentos futuros

- duplicar plano atual como cenário;
- alterar premissas;
- comparar resultados;
- aplicar cenário ao plano principal;
- descartar cenário.

### Regra

Cenários devem ser claramente identificados como simulação.

---

## 13. Dashboard mensal

### 13.1 Objetivo

Responder imediatamente:

> Quanto entra, quanto está comprometido, quanto quero guardar e quanto está livre?

### 13.2 Indicadores principais

- renda prevista;
- renda realizada;
- compromissos projetados;
- compromissos realizados;
- valor planejado para guardar;
- valor efetivamente guardado;
- dinheiro livre planejado;
- dinheiro livre realizado;
- valor livre restante;
- faturas projetadas;
- dívidas ativas.

### 13.3 Mensagem principal

Exemplo:

```text
Você possui R$ 1.240 livres neste mês
sem comprometer as premissas atuais.
```

A frase deve ser tratada como resultado matemático, não como recomendação.

### 13.4 Detalhamento progressivo

```text
Compromissos: R$ 3.270
```

Ao abrir:

```text
Moradia       R$ 1.200
Saúde         R$   700
Veículo       R$   650
Assinaturas   R$   120
Outros        R$   600
```

Ao abrir veículo:

```text
Seguro          R$ 350
Gasolina        R$ 250
Estacionamento  R$  50
```

Ao abrir seguro:

```text
Histórico de períodos
R$ 220 — out/2024 a jul/2025
R$ 350 — ago/2025 em diante
```

### 13.5 Informações complementares

- compromissos que terminam em breve;
- variações relevantes em números;
- itens sem valor realizado;
- faturas com maior composição;
- diferença para a meta;
- meses seguintes.

### 13.6 Ausência de linguagem prescritiva

Não utilizar:

- “bom”;
- “ruim”;
- “saudável”;
- “perigoso”;
- “excessivo”;
- “irresponsável”.

Utilizar:

- acima;
- abaixo;
- diferença;
- aumento;
- redução;
- participação;
- projeção;
- realizado;
- estimativa.

---

## 14. Dashboard anual

### 14.1 Objetivo

Mostrar o caminho do ano até a meta.

### Indicadores

- renda anual projetada;
- compromissos anuais projetados;
- total planejado para guardar;
- total efetivamente guardado;
- patrimônio projetado;
- patrimônio realizado;
- meta;
- valor faltante;
- estimativa atualizada;
- planejamento original;
- meses fechados;
- meses futuros.

### Tabela mensal

| Mês | Renda | Compromissos | Guardar | Livre | Realizado | Situação |
|---|---:|---:|---:|---:|---:|---|
| Janeiro | ... | ... | ... | ... | ... | Fechado |
| Fevereiro | ... | ... | ... | ... | ... | Fechado |
| Julho | ... | ... | ... | ... | ... | Atual |
| Dezembro | ... | ... | ... | ... | ... | Projetado |

### Visões

- planejamento original;
- estimativa atualizada;
- realizado;
- comparação.

### Exemplo de meta

```text
Meta para dezembro: R$ 30.000
Projeção atual: R$ 27.400
Diferença: R$ 2.600
```

---

## 15. Visualizações e gráficos

Os gráficos devem complementar os números, não substituí-los.

### 15.1 Distribuição de compromissos por categoria

Objetivo: mostrar composição do mês.

### 15.2 Composição por cartão

Objetivo: explicar faturas.

### 15.3 Projetado versus realizado

Objetivo: comparar expectativa e realidade ao longo dos meses.

### 15.4 Evolução do dinheiro livre

Objetivo: mostrar quanto permaneceu livre em cada período.

### 15.5 Evolução do patrimônio

Objetivo: acompanhar saldo acumulado, aportes e rendimentos.

### 15.6 Linha histórica de um item

Objetivo: mostrar mudanças de valor.

Para custos por período, gráfico em degraus tende a representar melhor mudanças de vigência do que uma linha suavizada.

### 15.7 Dívidas ao longo do tempo

Objetivo: mostrar queda de compromissos e valores que serão liberados.

### 15.8 Renda comprometida

Objetivo: mostrar percentuais numéricos de comprometimento.

### Requisitos

- sempre exibir valores acessíveis fora do gráfico;
- permitir detalhamento;
- evitar excesso de cores;
- não usar cores como único meio de transmitir significado;
- diferenciar projetado e realizado visualmente;
- apresentar unidade, período e origem.

---

## 16. Fluxos principais

## 16.1 Criar primeiro planejamento

1. criar planejamento;
2. escolher período;
3. cadastrar renda;
4. definir quanto deseja guardar;
5. cadastrar principais compromissos;
6. vincular cartões;
7. revisar projeção;
8. visualizar dinheiro livre.

### Critério de sucesso

O usuário chega à primeira projeção sem precisar compreender todos os módulos.

---

## 16.2 Cadastrar uma dívida temporária

1. criar item;
2. selecionar tipo dívida;
3. informar nome e descrição;
4. informar valor mensal;
5. informar início e término ou parcelas;
6. vincular forma de pagamento;
7. revisar ocorrências;
8. salvar;
9. visualizar impacto mensal e data de liberação.

---

## 16.3 Cadastrar gasolina

1. criar item “Gasolina”;
2. selecionar despesa variável;
3. definir projeção mensal;
4. vincular cartão;
5. definir início;
6. aplicar aos meses futuros;
7. salvar;
8. visualizar composição da fatura e do custo veicular.

---

## 16.4 Registrar gasolina realizada

1. abrir mês;
2. localizar gasolina;
3. informar realizado;
4. adicionar observação opcional;
5. visualizar diferença;
6. escolher manter ou alterar projeção futura;
7. salvar.

---

## 16.5 Alterar seguro após troca de carro

1. abrir seguro veicular;
2. selecionar “alterar a partir de uma data”;
3. informar novo valor;
4. informar início;
5. descrever “Seguro do Golf”;
6. sistema encerra período anterior;
7. sistema cria novo período;
8. projeções futuras são recalculadas;
9. histórico permanece disponível.

---

## 16.6 Fechar um mês

1. abrir fechamento;
2. revisar receitas;
3. revisar despesas fixas;
4. informar despesas variáveis;
5. revisar dívidas;
6. informar valor guardado;
7. verificar pendências;
8. confirmar;
9. visualizar comparação final.

---

## 16.7 Investigar um total

1. clicar no total;
2. visualizar agrupamento;
3. abrir categoria ou cartão;
4. abrir item;
5. consultar período e valor;
6. consultar histórico ou realizado.

### Critério

Nenhum total relevante deve terminar em uma tela sem origem explicável.

---

## 16.8 Simular uma decisão

1. criar cenário;
2. selecionar data da mudança;
3. adicionar ou alterar premissa;
4. recalcular meses;
5. comparar plano principal e cenário;
6. interpretar diferenças;
7. aplicar ou descartar.

---

## 17. Regras temporais

### 17.1 Datas distintas

O produto deve distinguir:

- data em que a informação foi cadastrada;
- data em que começa a valer;
- data em que deixa de valer;
- mês financeiro de referência;
- data em que o realizado foi informado.

### 17.2 Mudança futura

Exemplo:

```text
Informado em 21/07:
Seguro passará para R$ 350 em agosto.
```

A projeção de julho não muda. Agosto em diante é recalculado.

### 17.3 Mudança retroativa

Exemplo:

```text
Informado em julho:
Seguro mudou em junho.
```

O sistema ajusta junho e meses seguintes, mantendo registro de que a alteração foi cadastrada posteriormente.

### 17.4 Sobreposição de períodos

O sistema deve impedir ou resolver períodos incompatíveis para o mesmo item.

Não pode existir, sem intenção explícita:

```text
R$ 220 de janeiro a setembro
R$ 350 de agosto em diante
```

### 17.5 Períodos sem fim

Um item pode permanecer ativo indefinidamente até ser encerrado.

### 17.6 Alteração no meio do mês

A primeira versão pode trabalhar por competência mensal.

Quando uma mudança ocorrer no meio do mês, o usuário poderá:

- aplicar ao mês inteiro;
- aplicar a partir do mês seguinte;
- registrar um valor específico realizado.

Rateios diários podem ser considerados futuramente.

### 17.7 Exclusão

Itens que já participaram de meses históricos não devem desaparecer sem rastreabilidade.

Preferir:

- arquivar;
- cancelar;
- encerrar;
- corrigir com histórico.

---

## 18. Planejado, realizado e estimativa atualizada

### 18.1 Três visões obrigatórias

O produto deve conseguir distinguir:

1. **planejamento original** — o que se esperava no início;
2. **realizado** — o que efetivamente aconteceu;
3. **estimativa atualizada** — realizado passado mais projeção futura atual.

### 18.2 Exemplo anual

```text
Planejamento original: R$ 24.000 de custos
Estimativa atualizada: R$ 26.400
Realizado até julho: R$ 15.200
```

### 18.3 Meses abertos

Em meses ainda não fechados, o sistema pode combinar:

- realizados já informados;
- projeções ainda não realizadas.

A interface deve indicar claramente a composição.

### 18.4 Ausência de realizado

A ausência não deve ser automaticamente interpretada como zero.

Estados possíveis:

- não informado;
- igual ao projetado;
- não ocorreu;
- valor informado;
- pendente;
- movido.

---

## 19. Histórico e análise pessoal

### 19.1 Visão do produto

O histórico não existe apenas para auditoria. Ele é uma ferramenta de autoconhecimento baseada em dados.

### 19.2 Perguntas que o histórico deve permitir

- Quanto meu custo veicular mudou?
- Quando meu seguro aumentou?
- Qual era minha renda naquele período?
- Quanto uma troca de carro alterou meus compromissos?
- Em quais meses consegui guardar mais?
- Quanto minhas projeções de gasolina se aproximavam do real?
- Quando uma dívida terminou?
- Como minha capacidade de gastar livremente evoluiu?
- Qual era a participação de determinada categoria na renda?

### 19.3 Descrições humanas

Os dados devem aceitar contexto em linguagem do usuário.

Exemplo:

```text
Seguro do Civic
Seguro do Golf
Período sem emprego
Novo salário
Mudança de casa
Aumento de deslocamento
```

### 19.4 Correlação sem causalidade automática

O aplicativo pode mostrar que dois eventos ocorreram próximos.

Não deve afirmar causalidade sem informação explícita.

Exemplo permitido:

```text
No mesmo período:
- seguro aumentou R$ 130;
- gasolina projetada aumentou R$ 250;
- renda permaneceu igual.
```

Exemplo a evitar:

```text
A troca de carro prejudicou sua vida financeira.
```

---

## 20. Requisitos funcionais detalhados

### RF-001 — Criar planejamento

O usuário deve poder criar um planejamento para um intervalo mensal.

### RF-002 — Selecionar mês

O usuário deve poder navegar entre meses passados, atual e futuros.

### RF-003 — Cadastrar receita recorrente

O usuário deve poder cadastrar renda mensal com início e término opcional.

### RF-004 — Cadastrar receita pontual

O usuário deve poder cadastrar renda em um único mês.

### RF-005 — Alterar receita por vigência

O usuário deve poder criar novo período salarial sem apagar o anterior.

### RF-006 — Cadastrar despesa fixa

O usuário deve poder cadastrar custo recorrente previsível.

### RF-007 — Cadastrar despesa variável

O usuário deve poder cadastrar estimativa recorrente e posteriormente informar valor real.

### RF-008 — Cadastrar dívida

O usuário deve poder representar parcelas e término.

### RF-009 — Vincular forma de pagamento

O usuário deve poder associar um item a cartão ou outra forma.

### RF-010 — Derivar fatura projetada

O sistema deve somar itens vinculados ao cartão no mês.

### RF-011 — Explicar fatura

O usuário deve poder abrir a composição do total.

### RF-012 — Definir meta de economia mensal

O usuário deve poder reservar um valor antes do cálculo de dinheiro livre.

### RF-013 — Calcular dinheiro livre

O sistema deve calcular renda menos compromissos e economia planejada.

### RF-014 — Projetar meses futuros

O sistema deve aplicar períodos e recorrências.

### RF-015 — Registrar realizado

O usuário deve poder informar valor real por item e mês.

### RF-016 — Comparar projetado e realizado

O sistema deve calcular diferença absoluta e percentual.

### RF-017 — Manter projeção após realizado

Informar valor real não altera automaticamente meses futuros.

### RF-018 — Atualizar projeção futura

O usuário deve poder iniciar novo período com valor atualizado.

### RF-019 — Fechar mês

O usuário deve poder consolidar um mês.

### RF-020 — Reabrir mês

O usuário deve poder corrigir fechamento.

### RF-021 — Preservar histórico

Mudanças não devem sobrescrever períodos anteriores.

### RF-022 — Visualizar linha do tempo

O usuário deve poder visualizar eventos financeiros.

### RF-023 — Comparar períodos

O usuário deve poder comparar valores e percentuais.

### RF-024 — Visualizar término de compromissos

O sistema deve mostrar quando despesas temporárias terminam.

### RF-025 — Calcular valor liberado

O sistema deve mostrar redução de compromisso após término.

### RF-026 — Cadastrar investimentos

O usuário deve poder cadastrar saldos, aportes e projeções.

### RF-027 — Projetar rendimento

O sistema deve calcular rendimento conforme premissas informadas.

### RF-028 — Registrar rendimento real

O usuário deve poder informar valor realizado.

### RF-029 — Criar meta financeira

O usuário deve poder definir valor e data-alvo.

### RF-030 — Calcular valor faltante

O sistema deve comparar patrimônio e meta.

### RF-031 — Visualizar estimativa anual atualizada

O sistema deve combinar realizado e projetado.

### RF-032 — Consultar planejamento original

O usuário deve poder recuperar a expectativa inicial.

### RF-033 — Arquivar item

O usuário deve poder retirar item do uso atual sem apagar histórico.

### RF-034 — Adicionar contexto

O usuário deve poder descrever períodos e eventos.

### RF-035 — Filtrar análises

O usuário deve poder filtrar por período, categoria, cartão, item e tipo.

### RF-036 — Criar cenário

Em fase posterior, o usuário deve poder simular mudanças sem alterar o plano.

### RF-037 — Aplicar cenário

O usuário deve poder transformar uma simulação em plano ativo.

### RF-038 — Exportar dados

Em fase posterior, o usuário deve poder exportar seus dados e projeções.

### RF-039 — Importar base inicial

Em fase posterior, o usuário poderá acelerar a configuração por importação estruturada.

### RF-040 — Identificar números incompletos

O sistema deve deixar claro quando um total contém projeções, realizados e pendências.

---

## 21. Requisitos de experiência

### RX-001 — Primeiro valor rápido

O usuário deve obter uma projeção útil com poucas informações.

### RX-002 — Detalhamento progressivo

A interface não deve apresentar todas as complexidades na primeira tela.

### RX-003 — Linguagem objetiva

Textos devem usar conceitos como:

- projetado;
- realizado;
- diferença;
- período;
- compromisso;
- livre;
- guardado.

### RX-004 — Valores legíveis

Todos os valores devem apresentar moeda, sinal e período de referência.

### RX-005 — Edição contextual

Ao alterar um item, o sistema deve perguntar a partir de quando a mudança vale.

### RX-006 — Consequência visível

Antes de confirmar uma mudança, o usuário deve poder visualizar meses afetados.

### RX-007 — Sem surpresa histórica

A interface deve informar quando uma ação altera meses anteriores.

### RX-008 — Ausência de culpa

O design e o texto não devem utilizar punição, vergonha, nota de comportamento ou mascote triste.

### RX-009 — Dados antes de opinião

Comparações numéricas devem preceder qualquer texto interpretativo.

### RX-010 — Mobile utilizável

O produto deve permitir consulta e fechamento mensal em telas menores, mesmo que análises extensas sejam mais confortáveis em desktop.

---

## 22. Requisitos de qualidade do produto

### 22.1 Exatidão

Cálculos financeiros devem ser consistentes e reproduzíveis.

### 22.2 Rastreabilidade

Todo total deve ser decomponível.

### 22.3 Integridade histórica

Alterações não devem apagar realidade anterior.

### 22.4 Clareza temporal

Toda informação deve possuir período de referência claro.

### 22.5 Privacidade

Dados financeiros são sensíveis e devem ser tratados como privados por padrão.

### 22.6 Portabilidade

O usuário deve ter caminho para obter seus próprios dados.

### 22.7 Confiabilidade

O produto deve evitar alterações implícitas de premissas.

### 22.8 Acessibilidade

Números, gráficos e estados não devem depender somente de cor.

### 22.9 Transparência de projeção

Projeções devem ser identificadas como estimativas, não garantias.

---

## 23. Casos de borda

### 23.1 Renda zerada

O sistema deve representar períodos sem renda sem quebrar percentuais.

### 23.2 Dinheiro livre negativo

O valor pode ser negativo.

O produto deve mostrar o número e sua composição, sem classificar emocionalmente.

### 23.3 Projetado igual a zero e realizado positivo

A diferença absoluta é válida; variação percentual não é aplicável.

### 23.4 Despesa que não ocorreu

O usuário deve poder registrar zero como realizado com estado explícito “não ocorreu”.

### 23.5 Despesa movida

Uma parcela ou gasto pode mudar de mês.

A movimentação deve preservar origem.

### 23.6 Dívida quitada antecipadamente

Meses futuros devem ser removidos da projeção a partir da quitação, preservando o histórico.

### 23.7 Renegociação

Condição anterior é encerrada; nova condição é criada.

### 23.8 Troca de cartão

Novo período de pagamento é criado.

### 23.9 Cartão encerrado

Itens futuros devem ser revisados; histórico permanece.

### 23.10 Reembolso

Pode ser modelado como receita relacionada ou redução de despesa, desde que a interface torne o efeito claro.

### 23.11 Estorno

Deve preservar referência ao item original quando possível.

### 23.12 Receita variável

Pode possuir projeção e realizado, assim como despesa variável.

### 23.13 Décimo terceiro ou PLR

Deve ser representado como receita pontual ou programada.

### 23.14 Despesa anual parcelada

IPVA pode possuir meses específicos ou parcelamento.

### 23.15 Mudança retroativa em mês fechado

O sistema deve alertar que o fechamento será afetado e registrar a revisão.

### 23.16 Meses incompletos

O sistema deve indicar quando o mês ainda possui valores pendentes.

### 23.17 Duplicidade

O usuário deve ser avisado sobre possíveis itens duplicados, sem impedir casos legítimos.

### 23.18 Moeda

A primeira experiência pode trabalhar com uma moeda principal por planejamento. Suporte multimoeda pode ser posterior.

### 23.19 Arredondamento

Valores exibidos e totais devem seguir regra consistente.

### 23.20 Valores inesperados em fatura

O usuário pode registrar ajuste de fatura não categorizado e posteriormente detalhá-lo.

---

## 24. Exemplo completo de funcionamento

### 24.1 Premissas

```text
Salário: R$ 6.000
Renda extra: R$ 0
Guardar: R$ 1.200

Seguro: R$ 350
Gasolina projetada: R$ 700
Academia: R$ 150
Transplante: R$ 600 até dezembro
Moradia: R$ 1.500
Assinaturas: R$ 100
```

### 24.2 Pagamentos

```text
Nubank
- transplante: R$ 600
- academia: R$ 150
- assinaturas: R$ 100

Santander
- gasolina: R$ 700

Débito
- seguro: R$ 350
- moradia: R$ 1.500
```

### 24.3 Resultado projetado

```text
Renda: R$ 6.000
Compromissos: R$ 3.400
Guardar: R$ 1.200
Livre: R$ 1.400
```

### 24.4 Faturas projetadas

```text
Nubank: R$ 850
Santander: R$ 700
```

### 24.5 Realizado de gasolina

```text
Projetado: R$ 700
Realizado: R$ 845
Diferença: +R$ 145
```

### 24.6 Resultado atualizado do mês

```text
Compromissos realizados/estimados: R$ 3.545
Livre atualizado: R$ 1.255
```

### 24.7 Término da dívida

Em janeiro, o transplante deixa de compor o planejamento.

```text
Valor liberado: R$ 600 por mês
```

O usuário poderá decidir:

- aumentar dinheiro livre;
- aumentar economia;
- criar nova meta;
- assumir novo compromisso;
- não alterar nada.

O aplicativo não escolhe.

---

## 25. Exemplo de histórico de veículo

### 25.1 Período Civic

```text
Outubro/2024 a julho/2025

Seguro: R$ 220
Gasolina média: R$ 650
Manutenção média: R$ 180
Total mensal observado: R$ 1.050
```

### 25.2 Período Golf

```text
Agosto/2025 em diante

Seguro: R$ 350
Gasolina média: R$ 900
Manutenção média: R$ 180
Total mensal observado: R$ 1.430
```

### 25.3 Comparação

```text
Diferença mensal: +R$ 380
Impacto anual aproximado: +R$ 4.560
```

### 25.4 Relação com renda

```text
Antes: 18% da renda
Depois: 24% da renda
```

O sistema apresenta esses dados sem concluir se a troca foi adequada.

---

## 26. Mapa de telas

### 26.1 Início

- seleção de mês;
- dinheiro livre;
- renda;
- compromissos;
- economia;
- cartões;
- dívidas;
- alertas objetivos;
- acesso ao detalhamento.

### 26.2 Planejamento anual

- meses;
- original;
- atualizado;
- realizado;
- meta;
- patrimônio;
- gráficos.

### 26.3 Itens financeiros

- lista;
- filtros;
- status;
- valor atual;
- próximo término;
- tipo;
- categoria;
- forma de pagamento.

### 26.4 Detalhe do item

- valor atual;
- período;
- projetado e realizado;
- gráfico;
- histórico;
- descrições;
- forma de pagamento;
- alterações.

### 26.5 Cartões

- lista;
- fatura projetada;
- realizado;
- composição;
- histórico.

### 26.6 Dívidas

- lista;
- valor mensal;
- parcelas;
- término;
- valor a liberar;
- status.

### 26.7 Fechamento mensal

- receitas;
- despesas;
- pendências;
- realizado;
- diferenças;
- economia;
- confirmação.

### 26.8 Histórico

- linha do tempo;
- filtros;
- eventos;
- impactos;
- comparações.

### 26.9 Investimentos

- contas;
- aportes;
- rendimento;
- saldo;
- patrimônio;
- meta.

### 26.10 Metas

- valor-alvo;
- valor atual;
- projeção;
- diferença;
- data.

### 26.11 Cenários

- plano principal;
- mudanças simuladas;
- comparação;
- aplicação.

### 26.12 Configurações

- categorias;
- formas de pagamento;
- preferências de moeda;
- dados;
- exportação.

---

## 27. Estados e status

### 27.1 Item financeiro

- planejado;
- ativo;
- pausado;
- encerrado;
- cancelado;
- arquivado.

### 27.2 Mês

- futuro;
- aberto;
- pendente;
- fechado;
- reaberto.

### 27.3 Valor mensal

- projetado;
- parcialmente realizado;
- realizado;
- não ocorrido;
- pendente;
- movido.

### 27.4 Dívida

- planejada;
- ativa;
- suspensa;
- concluída;
- quitada antecipadamente;
- renegociada;
- cancelada.

### 27.5 Meta

- planejada;
- ativa;
- atingida;
- pausada;
- cancelada.

### 27.6 Cenário

- rascunho;
- comparado;
- aplicado;
- descartado.

---

## 28. Roadmap de produto

O roadmap descreve uma sequência recomendada. Não limita a visão completa.

## Fase 0 — Protótipo conceitual

### Objetivo

Validar o modelo mental e as telas principais.

### Entregas

- dashboard com dados fictícios;
- visão mensal;
- composição de cartão;
- lista de compromissos;
- fluxo de cadastro;
- exemplo de histórico.

### Critério

A interface deve comunicar a proposta antes da implementação completa.

---

## Fase 1 — Planejamento mensal utilizável

### Objetivo

Responder à principal pergunta do produto.

### Entregas

- receitas;
- despesas fixas;
- despesas variáveis projetadas;
- dívidas;
- cartões;
- meta mensal de economia;
- projeção do mês;
- dinheiro livre;
- detalhamento de totais.

### Critério

O usuário consegue substituir a parte principal da planilha para o mês atual.

---

## Fase 2 — Projeção anual

### Objetivo

Representar todos os meses e compromissos temporários.

### Entregas

- períodos;
- recorrências;
- término de dívidas;
- tabela anual;
- meses futuros;
- valor liberado;
- meta anual.

### Critério

O usuário consegue visualizar sua trajetória até dezembro.

---

## Fase 3 — Projetado versus realizado

### Objetivo

Fechar o ciclo de planejamento.

### Entregas

- realizado mensal;
- diferença;
- variação;
- fechamento;
- reabertura;
- estimativa atualizada.

### Critério

Meses passados deixam de ser apenas projeções.

---

## Fase 4 — Histórico temporal

### Objetivo

Preservar e analisar mudanças de vida.

### Entregas

- períodos versionados;
- descrições;
- linha do tempo;
- comparação antes/depois;
- gráficos históricos.

### Critério

Alterar um item nunca apaga seu contexto anterior.

---

## Fase 5 — Investimentos e patrimônio

### Objetivo

Reproduzir e melhorar a projeção de rendimentos da planilha.

### Entregas

- contas;
- saldo inicial;
- aportes;
- taxas;
- percentuais;
- rendimento;
- patrimônio;
- metas.

### Critério

O usuário acompanha valor guardado e estimativa de patrimônio.

---

## Fase 6 — Cenários

### Objetivo

Permitir testes sem comprometer o plano.

### Entregas

- criar cenário;
- comparar;
- aplicar;
- descartar.

### Critério

O usuário testa decisões futuras com números.

---

## Fase 7 — Portabilidade e automações

### Possibilidades

- importação;
- exportação;
- lembretes;
- modelos;
- duplicação de planejamento;
- integração financeira opcional.

### Observação

Integrações não devem descaracterizar a proposta de simplicidade e controle consciente.

---

## 29. Definição do primeiro produto mínimo

O primeiro produto mínimo não precisa conter toda a visão.

Ele deve responder:

> **Com minha renda, meus compromissos e o valor que quero guardar, quanto tenho realmente livre neste mês?**

### Escopo mínimo

- uma renda recorrente;
- renda extra pontual;
- despesas fixas;
- despesas variáveis projetadas;
- dívidas parceladas;
- cartões;
- meta de economia;
- visão mensal;
- detalhamento de fatura;
- dinheiro livre;
- projeção dos próximos meses;
- término de compromisso.

### Fora do primeiro produto mínimo

- Open Finance;
- importação automática;
- inteligência artificial;
- cenários avançados;
- recomendações;
- transações individuais;
- múltiplas moedas;
- colaboração;
- investimentos sofisticados;
- conciliação bancária.

### Critério de encerramento

A primeira versão está funcional quando o usuário consegue cadastrar o transplante, seguro, gasolina, cartões e salário e entender quanto pode gastar e guardar até o fim do ano.

---

## 30. Épicos e histórias de usuário

## Épico A — Planejamento

### HU-A01

Como usuário, quero criar um planejamento mensal para visualizar minha distribuição financeira.

**Aceite:**

- período definido;
- totais calculados;
- valores explicáveis.

### HU-A02

Como usuário, quero visualizar meses futuros para saber como compromissos afetam o ano.

**Aceite:**

- recorrências aplicadas;
- términos respeitados;
- navegação mensal.

---

## Épico B — Receitas

### HU-B01

Como usuário, quero cadastrar meu salário para incluí-lo nas projeções.

### HU-B02

Como usuário, quero alterar meu salário a partir de uma data sem perder o histórico.

### HU-B03

Como usuário, quero cadastrar renda extra em um mês específico.

---

## Épico C — Despesas

### HU-C01

Como usuário, quero cadastrar uma despesa fixa recorrente.

### HU-C02

Como usuário, quero cadastrar uma estimativa variável.

### HU-C03

Como usuário, quero alterar uma despesa a partir de determinado mês.

### HU-C04

Como usuário, quero encerrar uma despesa.

---

## Épico D — Dívidas

### HU-D01

Como usuário, quero cadastrar uma dívida parcelada para visualizar todos os meses afetados.

### HU-D02

Como usuário, quero saber quando a dívida termina.

### HU-D03

Como usuário, quero visualizar quanto ficará livre após a última parcela.

### HU-D04

Como usuário, quero quitar antecipadamente preservando o histórico.

---

## Épico E — Cartões

### HU-E01

Como usuário, quero vincular uma despesa a um cartão.

### HU-E02

Como usuário, quero visualizar a fatura projetada.

### HU-E03

Como usuário, quero abrir a fatura e visualizar sua composição.

### HU-E04

Como usuário, quero alterar o cartão de uma despesa a partir de uma data.

---

## Épico F — Dinheiro livre

### HU-F01

Como usuário, quero definir quanto desejo guardar.

### HU-F02

Como usuário, quero visualizar quanto permanece livre.

### HU-F03

Como usuário, quero abrir o cálculo e entender cada valor.

---

## Épico G — Realizado

### HU-G01

Como usuário, quero informar quanto realmente gastei com gasolina.

### HU-G02

Como usuário, quero comparar realizado e projetado.

### HU-G03

Como usuário, quero manter minha projeção futura mesmo que um mês seja atípico.

### HU-G04

Como usuário, quero atualizar a projeção futura quando decidir.

---

## Épico H — Fechamento

### HU-H01

Como usuário, quero revisar os valores do mês.

### HU-H02

Como usuário, quero preencher pendências.

### HU-H03

Como usuário, quero fechar o mês e consolidar os resultados.

### HU-H04

Como usuário, quero corrigir um fechamento.

---

## Épico I — Histórico

### HU-I01

Como usuário, quero visualizar todos os períodos de um custo.

### HU-I02

Como usuário, quero adicionar descrição para lembrar o contexto.

### HU-I03

Como usuário, quero comparar antes e depois de uma mudança.

### HU-I04

Como usuário, quero consultar minha linha do tempo financeira.

---

## Épico J — Investimentos

### HU-J01

Como usuário, quero cadastrar uma conta de investimento.

### HU-J02

Como usuário, quero projetar aportes e rendimentos.

### HU-J03

Como usuário, quero comparar rendimento projetado e real.

### HU-J04

Como usuário, quero acompanhar patrimônio acumulado.

---

## Épico K — Metas

### HU-K01

Como usuário, quero definir um objetivo financeiro.

### HU-K02

Como usuário, quero visualizar valor faltante.

### HU-K03

Como usuário, quero comparar meta e projeção.

---

## Épico L — Cenários

### HU-L01

Como usuário, quero simular uma mudança sem alterar meu planejamento.

### HU-L02

Como usuário, quero comparar cenário e plano atual.

### HU-L03

Como usuário, quero aplicar uma simulação aprovada.

---

## 31. Critérios globais de aceite

Uma funcionalidade financeira somente deve ser considerada concluída quando:

1. o valor pode ser rastreado até sua origem;
2. o período está explícito;
3. projetado e realizado não se confundem;
4. alterações não apagam histórico;
5. casos de ausência e zero são distintos;
6. resultados negativos são suportados;
7. meses afetados são previsíveis;
8. a interface não julga;
9. totais permanecem consistentes entre telas;
10. há exemplos ou testes de regra suficientes para validar o comportamento.

---

## 32. Métricas de sucesso do produto

Como projeto pessoal, o sucesso inicial pode ser avaliado por utilidade real.

### 32.1 Ativação

- usuário consegue montar o primeiro mês;
- usuário entende o dinheiro livre;
- usuário consegue explicar uma fatura.

### 32.2 Manutenção

- tempo necessário para atualizar salário ou custo;
- tempo necessário para fechar o mês;
- quantidade de fórmulas manuais evitadas;
- quantidade de meses atualizados automaticamente.

### 32.3 Clareza

- qualquer total pode ser detalhado;
- o usuário identifica dívidas ativas;
- o usuário identifica compromissos que terminarão;
- o usuário consegue comparar projetado e realizado.

### 32.4 Valor pessoal

- planejamento substitui a planilha principal;
- produto é consultado durante o mês;
- histórico é útil para decisões futuras;
- projeção anual permanece atualizada após mudanças.

---

## 33. Riscos de produto

### 33.1 Escopo excessivo

A visão completa é grande. Tentar implementar tudo antes da primeira versão pode impedir a entrega.

**Resposta:** desenvolver por fatias verticais.

### 33.2 Complexidade temporal

Vigências, retroatividade e fechamento podem gerar comportamentos difíceis.

**Resposta:** definir regras mensais claras antes de permitir casos avançados.

### 33.3 Registro manual cansativo

Se o produto exigir cada transação, poderá perder a proposta.

**Resposta:** priorizar valores consolidados e planejamento.

### 33.4 Gráficos sem utilidade

Dashboards podem ficar bonitos e pouco explicativos.

**Resposta:** números e drill-down são obrigatórios.

### 33.5 Inteligência artificial prematura

Uma camada de IA pode desviar o produto para recomendações e julgamentos.

**Resposta:** manter valor central independente de IA.

### 33.6 Totais não explicáveis

A existência de ajustes manuais pode recriar células misteriosas.

**Resposta:** todo ajuste deve possuir origem, descrição e período.

### 33.7 Histórico inconsistente

Edições comuns podem apagar versões.

**Resposta:** tratar vigência como conceito central.

---

## 34. Decisões de produto já tomadas

1. O foco principal é projeção, não somente controle do passado.
2. Cartão é forma de pagamento e agrupador, não despesa.
3. Faturas projetadas são formadas pela soma de itens vinculados.
4. Dívidas devem ser especificadas individualmente.
5. O sistema deve mostrar quanto pode ser guardado.
6. O sistema deve mostrar quanto está realmente livre.
7. Valores variáveis possuem projeção e realizado.
8. Informar realizado não altera automaticamente o futuro.
9. Alterações criam períodos e preservam histórico.
10. Descrições podem explicar o contexto de cada período.
11. O usuário deve poder analisar a própria trajetória.
12. O aplicativo apresenta números e evita julgamentos.
13. Inteligência artificial não é parte necessária da proposta.
14. Não é obrigatório registrar cada transação.
15. Fechamento pode trabalhar com valores consolidados.
16. Investimentos e CDI fazem parte da visão completa.
17. O produto deve suportar planejamento anual.
18. Todo total precisa ser explicável.
19. A visão simples deve permitir aprofundamento.
20. A entrega deve ser dividida em fases.

---

## 35. Questões de produto ainda abertas

As questões abaixo não impedem a primeira versão, mas devem ser decididas conscientemente.

### 35.1 Competência de cartão

- o item pertence ao mês da compra, fechamento ou vencimento?
- o usuário escolhe manualmente na primeira versão?

### 35.2 Limite do registro realizado

- somente total mensal por item;
- possibilidade opcional de registrar ocorrências individuais;
- importação futura.

### 35.3 Planejamento original

- será criado automaticamente no início do ano;
- poderá ter versões nomeadas;
- quais alterações atualizam apenas a estimativa?

### 35.4 Economia e investimento

- guardar dinheiro é um compromisso independente;
- todo valor guardado precisa ir para uma conta;
- pode existir reserva sem destino?

### 35.5 Realizado de despesas fixas

- assumir projetado como realizado por padrão;
- exigir confirmação no fechamento;
- permitir regras por item.

### 35.6 Comparação entre contextos

- descrições livres;
- etiquetas;
- entidades como veículo, emprego ou moradia;
- evolução gradual.

### 35.7 Múltiplos planejamentos

- um planejamento ativo;
- vários cenários;
- vários usuários futuramente.

### 35.8 Privacidade e autenticação

- quais garantias serão necessárias antes de utilizar dados reais;
- quando habilitar múltiplos usuários.

### 35.9 Taxas de investimento

- entrada manual;
- histórico mensal;
- atualização automática futura.

---

## 36. Recomendação de ordem para implementação funcional

Sem definir tecnologia, a ordem de produto recomendada é:

1. modelo de mês e período;
2. renda;
3. despesas fixas;
4. despesas variáveis projetadas;
5. dívidas;
6. cartões;
7. cálculo de compromissos;
8. meta de economia;
9. dinheiro livre;
10. dashboard mensal;
11. projeção dos meses seguintes;
12. realizado;
13. fechamento;
14. estimativa anual atualizada;
15. histórico por vigência;
16. linha do tempo;
17. investimentos;
18. metas anuais;
19. cenários;
20. importações e integrações.

Cada etapa deve produzir uma experiência utilizável.

---

## 37. Fatias verticais recomendadas

### Fatia 1 — “Quanto está livre?”

- renda;
- despesas;
- economia;
- cálculo;
- dashboard.

### Fatia 2 — “Por que minha fatura é esse valor?”

- cartões;
- vínculo;
- composição;
- detalhamento.

### Fatia 3 — “Quando essa dívida termina?”

- parcelas;
- término;
- projeção;
- valor liberado.

### Fatia 4 — “Quanto eu realmente gastei?”

- realizado;
- diferença;
- fechamento.

### Fatia 5 — “Como esse custo mudou?”

- períodos;
- histórico;
- descrição;
- comparação.

### Fatia 6 — “Quanto terei no fim do ano?”

- planejamento anual;
- economia;
- patrimônio;
- meta.

---

## 38. Cenário de validação principal

A primeira validação completa deve conseguir representar:

- novo salário a partir de determinado mês;
- meses anteriores com outra renda ou sem renda;
- seguro veicular com mudança de valor;
- combustível variável;
- dívida do transplante até dezembro;
- múltiplos cartões;
- meta mensal para guardar;
- valor livre mensal;
- fechamento de um mês;
- comparação projetado versus realizado;
- projeção até o fim do ano.

Esse cenário cobre o núcleo real que originou o produto.

---

## 39. Tom e identidade conceitual

O produto deve parecer:

- organizado;
- confiável;
- claro;
- calmo;
- técnico na medida certa;
- pessoal;
- transparente.

Não deve parecer:

- banco vendendo produto;
- aplicativo infantil;
- coach;
- sistema contábil empresarial;
- ferramenta punitiva;
- dashboard carregada de indicadores sem contexto.

### Exemplos de linguagem apropriada

- “Projetado”
- “Realizado”
- “Diferença”
- “Livre no mês”
- “Compromissos”
- “Termina em dezembro”
- “R$ 600 deixam de estar comprometidos em janeiro”
- “Média realizada nos últimos três meses”
- “Atualizar projeção futura”

### Exemplos a evitar

- “Você falhou em economizar”
- “Gasto ruim”
- “Seu comportamento está inadequado”
- “Você deveria cancelar”
- “Parabéns, você foi disciplinado”
- “Nota financeira”

---

## 40. Resumo final do produto

O Planejador Financeiro transforma premissas pessoais em uma visão explicável da vida financeira.

Seu núcleo é formado por:

```text
Renda
Compromissos
Dívidas
Cartões
Valor para guardar
Dinheiro livre
Projetado
Realizado
Histórico
Projeção futura
```

O produto deve ajudar a pessoa a:

- viver o mês de maneira organizada;
- entender quanto pode gastar;
- proteger o valor que deseja guardar;
- acompanhar dívidas;
- projetar o fim do ano;
- adaptar o plano quando a vida muda;
- olhar para trás e compreender sua trajetória;
- tomar decisões usando seus próprios números.

O aplicativo não deve decidir pela pessoa.

Ele deve oferecer algo que uma boa planilha tenta oferecer, mas com relações visíveis, histórico preservado, atualização simples e números sempre explicáveis.

---

# Apêndice A — Exemplo de estrutura conceitual de dados

Este apêndice não define banco ou implementação. Serve para tornar as entidades do domínio explícitas.

## Usuário

- identificação;
- preferências;
- moeda principal;
- configurações de visualização.

## Planejamento

- nome;
- início;
- fim;
- status;
- versão de referência;
- data de criação.

## Item financeiro

- nome;
- tipo;
- categoria;
- status;
- descrição geral.

## Período do item

- item;
- início;
- fim;
- valor projetado;
- recorrência;
- forma de pagamento;
- descrição do período;
- contexto.

## Valor mensal realizado

- item;
- mês;
- valor;
- observação;
- estado;
- data de registro.

## Dívida

- item relacionado;
- valor total;
- parcela;
- quantidade;
- início;
- término;
- status.

## Forma de pagamento

- tipo;
- nome;
- período;
- dados de fechamento opcionais.

## Meta

- nome;
- valor;
- data;
- status.

## Conta de investimento

- nome;
- saldo inicial;
- referência;
- percentual;
- período.

## Aporte ou retirada

- conta;
- mês;
- valor;
- tipo.

## Evento histórico

- data;
- tipo;
- item;
- descrição;
- valores anterior e novo;
- impacto.

## Fechamento mensal

- mês;
- estado;
- data;
- totais;
- pendências;
- revisões.

---

# Apêndice B — Checklist para novas funcionalidades

Antes de adicionar uma funcionalidade, responder:

1. Qual pergunta do usuário ela resolve?
2. O dado é projetado, realizado ou ambos?
3. Qual é o período de validade?
4. A alteração preserva o histórico?
5. O número pode ser explicado?
6. Como afeta o dinheiro livre?
7. Como afeta cartões?
8. Como afeta o planejamento anual?
9. A funcionalidade está mostrando dados ou julgando?
10. Pode ser entregue em uma fatia menor?
11. O usuário precisa registrar cada transação ou um total consolidado basta?
12. O comportamento em mês fechado está definido?
13. O comportamento retroativo está definido?
14. Zero e ausência são diferentes?
15. Existe risco de alterar o futuro sem consentimento?

---

# Apêndice C — Checklist de uma tela financeira

Uma tela que exibe valores deve responder:

- qual período;
- qual moeda;
- projetado ou realizado;
- total de quê;
- origem do total;
- itens incluídos;
- itens pendentes;
- comparação disponível;
- ação possível;
- consequência da ação.

---

# Apêndice D — Instrução de uso como fonte do projeto

Este documento deve ser tratado como fonte de verdade para a visão do produto.

Ao propor uma funcionalidade, deve-se:

1. localizar o princípio ou requisito relacionado;
2. declarar eventuais conflitos;
3. não inventar comportamento que apague histórico;
4. não transformar cartão em despesa;
5. não misturar projetado e realizado;
6. não adicionar aconselhamento prescritivo;
7. manter totais explicáveis;
8. respeitar a entrega por fases;
9. registrar novas decisões;
10. atualizar as questões abertas quando forem resolvidas.
