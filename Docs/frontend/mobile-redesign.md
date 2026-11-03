# Resumo, cartões e faturas no mobile

Implementação de 02/11/2026, baseada no conceito visual aprovado pelo mantenedor.

## Experiência

- O resumo prioriza o valor livre previsto e a conta receitas menos compromissos
  menos reserva. Os grupos permitem abrir as fontes e acessar seus registros.
- O gráfico mostra o dinheiro livre previsto por mês, com valores, ano, zero,
  sinal negativo e indisponibilidade distintos. Mantém a janela de cinco meses.
- Cartões mostram as faturas do mês escolhido em linhas compactas. Cada linha
  abre diretamente o detalhe. Configurações, edição e arquivamento ficam no menu
  do cartão; cadastros e instituições deixam de ocupar o início da consulta.
- A fatura explica compras, pagamentos e composição. Lançamentos expandem seus
  detalhes e ações existentes. A importação Nubank abre pelo próprio detalhe.
- O mobile usa Início, Lançamentos, Cartões e Planejar. Dívidas têm acesso em
  Lançamentos e no resumo. Links antigos de faturas continuam por redirecionamento.
- O cabeçalho deixa de consultar um segundo saldo independente do mês/filtro
  selecionado, evitando números divergentes e uma chamada desnecessária.
- O resumo deixa de consultar contagens de cartões e dívidas usadas pelo antigo
  painel de onboarding.

## Contratos e limites

Totais oficiais, regras de pagamento, projeção, conciliação e importação continuam
na API. Não há migration, dependência nova ou alteração de autenticação.

O caixa conserva o filtro existente de previsões variáveis. A visão selecionada
fica explícita e as estimativas excluídas têm aviso e ação para inclusão.

`is_projected = false` não comprova realização: ocorrências de itens financeiros
continuam identificadas como compromissos previstos. Ajustes são cobranças
registradas; pagamentos e créditos preservam seus sinais. Fontes desconhecidas
não ganham um rótulo de realização por inferência. A composição é explicativa e
nunca substitui o saldo oficial da fatura.

Faturas indisponíveis não viram zero nem geram um total parcial apresentado como
completo. A consolidação exige todas as faturas e a moeda do perfil. O mês de
consulta respeita os limites do plano e o horizonte operacional de faturas.

## Validação

- TypeScript, ESLint, build de produção e 23 testes Node aprovados.
- Regressões cobrem sinais de pagamentos/créditos, classificação de previsões,
  zero versus indisponibilidade, redirecionamentos de autenticação e limites de mês.
- Conferência visual dos componentes reais com fixtures sintéticas em 390 px,
  fatura/gráfico em 320 px e resumo em 1440 px. Sem overflow horizontal da página;
  o gráfico admite rolagem própria com instrução no mobile.
- Expansão da composição, detalhe do lançamento, abertura/fechamento da importação
  e painel de novo cartão conferidos sem enviar mutações.
- Corrigido o vínculo entre o título do painel de cartão e seu diálogo acessível.
- O harness temporário de preview foi removido; não há rota pública de fixtures.

Não foi executado smoke autenticado da branch publicada nem realizado deploy.
As capturas locais usam números ilustrativos. Antes de produção, validar em
staging a consulta com dados reais, troca de mês/filtro, importação e manutenção
já previstas no produto. Não interpretar esta revisão visual como conciliação
ou confirmação de pagamento.

Publicação pelo fluxo de branch/PR para develop e posterior promoção por PR para
main. Rollback pelo retorno ao frontend anterior; não há estado de banco novo.
