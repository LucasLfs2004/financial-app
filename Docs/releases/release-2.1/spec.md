# Release 2.1 — Spec — Registros financeiros contínuos

## 1. Objetivo

Separar a vida útil dos registros financeiros do horizonte de um planejamento.

O planejamento passa a responder somente quais competências serão projetadas e
visualizadas. Receitas, despesas, dívidas, formas de pagamento e ocorrências de
fatura pertencem ao usuário e podem começar antes ou terminar depois desse
intervalo.

Exemplo:

```text
planejamento: 2026-09 até 2026-12
seguro:        2026-09 até 2027-09
```

O seguro participa dos quatro meses do planejamento atual e permanece disponível
para uma projeção futura, sem recadastro e sem ser transferido entre planos.

## 2. Resultado para o usuário

O usuário consegue:

1. cadastrar um registro cuja vigência ultrapasse o planejamento atual;
2. manter receitas, despesas e dívidas ao arquivar ou substituir um planejamento;
3. visualizar em cada planejamento apenas as ocorrências que interceptam seu
   horizonte;
4. preservar mudanças de valor e de pagamento em uma linha temporal contínua;
5. reutilizar os mesmos compromissos em planejamentos futuros;
6. distinguir registros da vida financeira de decisões específicas de um plano.

## 3. Modelo conceitual

### 3.1 Registros do usuário

Pertencem ao usuário, sem propriedade do planejamento:

- itens financeiros e seus períodos;
- dívidas e quitações futuras;
- métodos de pagamento por vigência;
- ajustes de fatura;
- movimentações excepcionais de ocorrências;
- instituições e cartões, como já ocorre na Release 2.

Esses recursos usam `user_id` como fronteira de ownership. Seus períodos são
validados pela própria consistência temporal, não pelo horizonte do plano.

### 3.2 Configurações do planejamento

Continuam pertencendo ao planejamento:

- nome, status, horizonte e moeda da projeção;
- valor planejado para guardar;
- fotografia original do planejamento.

O valor planejado para guardar permanece no plano porque representa uma decisão
para aquele cenário, não um compromisso financeiro independente.

### 3.3 Projeção por interseção

Um registro participa da projeção quando uma ocorrência de sua vigência pertence
ao intervalo consultado. O vínculo é calculado, não persistido.

```text
ocorrências do plano
= ocorrências dos registros do usuário
  cuja competência ou efeito de caixa pertence à janela consultada
```

Na base de referência, a seleção usa a competência. Na base de caixa, usa o mês
de recebimento, pagamento direto ou fatura aplicável, preservando as regras de
offset e transbordo da Release 2.

## 4. Escopo

### 4.1 Itens financeiros

- criação não exige que início ou fim estejam dentro do plano;
- período aberto continua válido até encerramento explícito;
- item pontual continua exigindo início igual ao fim;
- períodos do mesmo item continuam sem sobreposição;
- alteração e arquivamento podem ocorrer em qualquer competência válida;
- listagem e detalhe não dependem da existência de planejamento atual.

### 4.2 Moeda

Cada item e ajuste de fatura registra `currency_code`, preenchido com a moeda do
planejamento atual no momento da criação nesta release. O planejamento consome
apenas registros da mesma moeda, pois conversão cambial permanece fora do escopo.

Essa cópia explícita evita reinterpretar um registro antigo caso um planejamento
futuro use outra moeda.

### 4.3 Métodos de pagamento

- pertencem ao item e ao usuário;
- podem cobrir toda a vigência do item, inclusive depois do plano atual;
- continuam restritos a itens de despesa;
- cartão vinculado continua pertencendo ao mesmo usuário;
- ausência de configuração continua significando pagamento direto.

### 4.4 Faturas

Ajustes e movimentações deixam de pertencer ao planejamento. São identificados
por usuário, cartão ou item e competência.

A consulta de uma fatura ainda exige um planejamento atual para definir a janela
de projeção exposta nesta release. A persistência, contudo, não perde os eventos
quando esse plano é arquivado.

### 4.5 Resumo mensal

- permanece limitado ao horizonte do planejamento atual;
- carrega itens do usuário compatíveis com a moeda do plano;
- seleciona somente ocorrências relevantes ao mês e à base solicitada;
- mantém economia planejada específica do plano;
- mantém explicabilidade e ausência de dupla contagem.

### 4.6 Ativação e fotografia original

Um plano é ativável quando existe ao menos uma ocorrência de renda compatível
com sua moeda e que intercepte seu horizonte, além da configuração explícita de
economia.

O snapshot captura deterministicamente:

- o plano e sua economia;
- itens cujos períodos interceptem o horizonte;
- os períodos completos desses itens, inclusive a parte que ultrapasse o fim;
- métodos, cartões e eventos necessários para explicar suas ocorrências;
- ajustes alcançáveis pela janela de referência ou caixa.

Snapshots v1 e v2 existentes permanecem imutáveis.

### 4.7 Compatibilidade e migração

- dados existentes são migrados a partir do `plan_id` atual;
- `currency_code` é copiado do plano de origem;
- identificadores públicos são preservados;
- rotas canônicas passam a usar `/v1/financial-items`; aliases legados sob
  `/v1/plans/current/items` permanecem operacionais nesta release;
- `plan_id` deixa de ser obrigatório nas representações de registros;
- `plan_id` é removido das respostas de registros; clientes devem usar
  `currency_code` e a identidade do próprio item.

## 5. Fora de escopo

- múltiplas moedas somadas no mesmo planejamento;
- conversão cambial;
- cenários que incluem ou excluem manualmente registros específicos;
- compartilhamento de registros entre usuários;
- resumo sem planejamento atual;
- valores realizados e conciliação bancária;
- alteração retroativa de snapshots existentes.

## 6. Regras de negócio

### RN-R21-001 — Ownership contínuo

Todo registro financeiro pertence ao usuário autenticado e sobrevive ao ciclo de
vida de qualquer planejamento.

### RN-R21-002 — Plano como janela

O horizonte limita a projeção e a consulta do plano, não a vigência persistida de
um registro.

### RN-R21-003 — Interseção temporal

Somente ocorrências relevantes ao mês e à base consultada participam dos totais.

### RN-R21-004 — Moeda compatível

Um plano não agrega itens de moeda diferente e nunca realiza conversão implícita.

### RN-R21-005 — Economia específica

O valor planejado para guardar continua vinculado ao plano e não é reutilizado
automaticamente por outro planejamento.

### RN-R21-006 — Histórico

Trocar ou arquivar um planejamento não apaga, encerra nem duplica registros.

### RN-R21-007 — Segurança

Queries, chaves estrangeiras e RLS usam `user_id` para impedir relações entre
recursos de usuários diferentes.

### RN-R21-008 — Snapshot

A fotografia original contém as premissas alcançáveis pela projeção no instante
da ativação e continua imutável.

### RN-R21-009 — Compatibilidade

Sem registros fora do horizonte, os resultados das Releases 1 e 2 permanecem
inalterados.

## 7. Critérios de aceite

1. Um plano `2026-09..2026-12` aceita seguro `2026-09..2027-09`.
2. O resumo de setembro a dezembro inclui o seguro uma vez por mês.
3. O item continua consultável depois do fim ou arquivamento do plano.
4. Um plano futuro da mesma moeda reutiliza o seguro sem cópia.
5. Método de pagamento pode vigorar até setembro de 2027.
6. Faturas e movimentações preservam a identidade `item + competência`.
7. Itens de outra moeda não entram no resumo.
8. Economia continua isolada por plano.
9. Snapshots anteriores continuam legíveis e imutáveis.
10. RLS e API rejeitam vínculos com recursos de outro usuário.
11. Suítes das Releases 1 e 2 permanecem verdes após adaptação dos contratos.
