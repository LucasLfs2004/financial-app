# Qualidade e cobertura frontend — 04/10/2026

**Resultado:** implementação frontend viável entregue e checks locais aprovados. Aceitação integrada parcial; sem deploy ou smoke de produção. [Tasks](tasks.md) mantém `[~]` quando ainda exige API/ambiente real. Próximos passos: [pending-items](pending-items.md).

## Ambiente e execução

- Node 22.21.0; Next 15.5.25; React 19.3.0; TypeScript 5.9.3 (versões instaladas).
- `pnpm test`: 14 testes aprovados; módulos TypeScript reais compilados em memória, fetch/sessão simulados.
- `pnpm type-check`, `pnpm lint`, `pnpm build`: aprovados. Rotas `/cards/[cardId]`, `/entries/[itemId]`, `/planning/original` presentes no build.
- Browser local Next, desktop 1280×720 e mobile 390×844, dados sintéticos: detalhe/configuração de cartão, instituições, pagamentos, composição e snapshot v4. Instituição: abrir edição, Tab até nome, Escape, confirmação contextual e cancelar arquivo. Sem envio de mutações; documento sem transbordamento horizontal na fixture móvel.
- Rota temporária de inspeção removida. Sem criação/alteração de dados financeiros de usuário, testes da API, migrations ou publicação.

## Evidências

| ID | Caso |
| --- | --- |
| E01 | `tests/integration.test.mjs`: calendário/timezone, limites inclusivos 24/120, projeção no fim e vigência de zero. |
| E02 | Mesmo arquivo: array payment-history, timestamps e default direto. |
| E03 | Mesmo arquivo: estrutura de 180 parcelas, janela final, quitação independente e 1/1. Não executa dívida real na API. |
| E04 | Mesmo arquivo: proxy 200/201/400/401/404/409/422/500, query/ID/multipart, Headers, abort, POST único e sessão server. |
| E05 | `tests/frontend-contracts.test.mjs`: pagamento completo/null/cartão obrigatório/intervalo invertido; metadata omitida/null/zero. |
| E06 | Mesmo arquivo: ajuste proposto respeita capabilities/rota persistida/contexto ausente; pagamento/arquivo sem editor. Sem chamada real. |
| E07 | Mesmo arquivo + fixtures: snapshots v1–v4/unknown, datas integrais/coleções ausentes/zero/imutabilidade. Dados sintéticos baseados nas queries históricas. |
| E08 | Mesmo arquivo: CSV magnitude/saldo, backfill zero e ambiguidade. Não verifica transação/idempotência da API. |
| E09 | Mesmo arquivo: fontes dívida/fatura/item/economia sem inventar competência. |
| E10 | Conferência browser descrita acima; não é E2E autenticado nem auditoria completa de acessibilidade. |
| E11 | Tipos/lint/build e inspeção de componentes/rotas. Evidência de implementação, não de efeitos financeiros. |

## Matriz CA-FE

| CA | Entrega / evidência local | Aceitação restante |
| --- | --- | --- |
| 01 | Instituições/filtros/nome/arquivo contextual; E10/E11. | Mutações e seleções autenticadas, instituição arquivada sem novo cartão. |
| 02 | Cartão/configuração temporal/payload completo; E01/E05/E10/E11. | Mudança e sobreposição reais. |
| 03 | Configuração ausente/dia nominal sem correção local; E10/E11. | Fatura com dia inexistente e mudança temporal. |
| 04 | Faturas <=24 meses/histórico/transbordo/retorno; E01/E11. | Faturas reais e intervalo vazio. |
| 05 | Editor proposto por ID/capabilities/rota persistida; E06/E11. | T02/03 API; flag desabilitada por padrão. |
| 06 | PATCH só diferenças e omissão/null; E06. | Handler/transação/auditoria/rejeições. |
| 07 | Vínculo persistido e efeito de arquivo preparados; E06/E11. | GET/capabilities, arquivo restaura previsão. |
| 08 | Reserva explicada apenas via capability; E11. | Identidade e reimportação sem duplicação na API. |
| 09 | Multipart/ambiguidade/token/preview invalidada; E04/E08/E11. | Preview/confirm e matching autenticados. |
| 10 | Backfill sem novos lançamentos/zero; E08. | Reimportação após edição/arquivo/idempotência. |
| 11 | Falha parcial preserva ID/link; conciliação preparada; E11. | Reload real e forecast/matching/transação da API. |
| 12 | n/N/projeção CSV/pagamentos em magnitude; E08/E11. | Séries/reimportação/substituição no backend. |
| 13 | Registro/detalhe/metadata/vigências/arquivo por mês; E05/E11. | Efeitos temporais e arquivo reais. |
| 14 | Arrays/contexto/instante de pagamento e movimentos; E02/E05/E11. | Trocas futuras, dois eventos e current_allocation da API. |
| 15 | Composição oficial/buckets/fontes/links; E09/E10/E11. | Resumos reais com todas as origens/bases. |
| 16 | Dívida longa em janelas e quitação independente; E03. | 180 parcelas reais/contagens/liberações oficiais. |
| 17 | Evento readonly, 1/1 sem ação, lock sem retry; E03/E04/E11. | POST fora da janela/rejeição/refresh reais. |
| 18 | Quitação/n/N/link somente com debt_id; E09/E11. | Fatura de dívida quitada e parcela CSV. |
| 19 | Perfil/horizonte/economia zero/plano readonly; E01/E07/E11. | Ativação/economia autenticadas. |
| 20 | Snapshot congelado v1–v4/unknown sem GET atual; E07/E10/E11. | Formalizar contrato histórico; atualizar plano e reler original em ambiente real. |
| 21 | Erros/sessão/abort/request ID/locks; E04/E11. | Sessão real, falha parcial, fora de ordem pela UI e acessibilidade completa. |
| 22 | Rotas/query/abas/retorno/IDs/Tab/Escape; E10/E11. | Navegação protegida/reload/voltar em todos os filtros. |
| 23 | Composição/projeção, caixa padrão e competência sob flag; E09/E10/E11. | Valores oficiais e paridade de bases reais. |
| 24 | Moeda retornada/limites inclusivos/seleção temporal; E01/E10/E11. | Faturas/configurações/dia nominal/moedas do ambiente. |

## Contratos e status

- T06/T21 concluídas no frontend com evidência local. T01, T07–13, T16–20 e T22 implementadas com aceitação integrada pendente. T04/05 têm frontend/fixtures entregues e contrato de API pendente. T14/15 preparados sob flag aguardam T02/03. T23/24 consolidam validação local, sem fechamento integrado. T25 não executada.
- [integration-contracts](integration-contracts.md) mantém divergências explícitas; o OpenAPI da API não foi alterado. Arrays operacionais aceitos; current_allocation opcional e não inferido.
- Manutenção de ajustes usa contrato proposto: GET antes de editar, capabilities, localização persistida e campos omitidos preservados. Flag desabilitada até API/regressões aprovadas.
- Snapshot histórico separado do DTO operacional; [fontes de fixtures](../../tests/fixtures/README.md). Nenhum snapshot existente modificado.

Build aprovado não certifica cálculos financeiros, idempotência de importação, autorização de terceiros ou produção. Fechamento de RF/CA depende das evidências da matriz.

## Integração solicitada — 26/10/2026

O mantenedor autorizou incluir todo o trabalho local do frontend na integração. A revisão desta etapa executou 18 testes locais (incluindo as duas visões de caixa), TypeScript e lint de app/components/lib/middleware. O build de produção do mesmo código foi aprovado na implementação do seletor. Os testes financeiros da API também aprovaram os dois modos, mantendo cobranças importadas.

O frontend consulta include_variable_forecasts=false/true conforme o seletor, preservando a escolha na URL, no carregamento inicial e ao trocar mês/base. A API compatível está em origin/develop e ainda depende da promoção por PR para main. A publicação final do frontend deve ocorrer depois dessa API. Não há migration para essa mudança.

As pendências contratuais de ajustes/capabilities continuam em pending-items.md e a flag correspondente permanece desabilitada. Aceitação autenticada e smoke em produção ainda não foram realizados; 18 testes locais não encerram essas pendências.

A develop remota do frontend continha o histórico legado da API, sem ancestral comum com a main do app. A integração preservará os dois históricos e manterá a árvore de arquivos do frontend por um merge de reconciliação, sem force push ou reescrita dos commits publicados.
