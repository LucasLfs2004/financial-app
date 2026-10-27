# Frontend 0–3.1 — Tasks de implementação

**Data:** 04/10/2026. **Estado:** escopo frontend viável entregue localmente; aceitação integrada parcial. Evidências em [quality-report](quality-report.md) e próximos passos em [pending-items](pending-items.md).
Fontes: [spec](../../../IA/financial-api/Docs/frontend/spec-cobertura-api-0-3.1.md), [design](design.md) e [contratos](integration-contracts.md).

## 1. Regras de execução

Cada task deve entregar código revisável, evidência dos critérios indicados e atualização do status neste documento. Usar a implementação existente como baseline, preservando mudanças locais do usuário. Uma task pode aproveitar código já existente, mas só é concluída depois da validação. Não gerar testes que apenas repitam a implementação; priorizar contratos, efeitos financeiros, limites e estados de falha.

API: `/Users/lucas/Documents/IA/financial-api`. Frontend: `/Users/lucas/Documents/projecao-financeira-app`. Alterações propostas de contrato começam no OpenAPI/handler e testes correspondentes. Não implementar tabelas financeiras diretamente no frontend. Versionamento e publicação seguem as regras de cada repositório; o backlog não autoriza contornar PR ou checks de produção.

**Status:** `[ ]` pendente; `[~]` em execução; `[x]` concluída com evidência. Dependências abaixo indicam gates técnicos, não necessidade de concluir toda uma fase antes de qualquer trabalho independente.

## 2. Fase A — Contratos e infraestrutura

### [~] FE-T01 — Transporte, erros e estados comuns

- **Avanço em 04/10/2026:** Proxy, clientes e boundary preservam status/code/message/details e distinguem 404 esperado; sessão 401 redireciona ao login, abort não alerta e POST não repete automaticamente. Testes de status, multipart, sessão e falha de rede aprovados. Smoke autenticado e respostas fora de ordem pela interface ainda pendentes.

- **Repositório:** frontend. **Depende:** nenhuma. **RF:** 001, 003.
- Ajustar `lib/api/financial-api.ts`, `browser-api.ts` e `app/api/financial/[...path]/route.ts` para preservar status HTTP e code/message da API. Helpers de JSON atuais permanecem compatíveis.
- Separar 404 esperado de falha de serviço nas páginas; não usar catch genérico como lista vazia. Definir erro, retry, sessão expirada e request abortado.
- Preservar Multipart boundary e query; não repetir POST automaticamente depois de timeout.
- **Pronto quando:** 200/201/400/401/404/409/422/500 atravessam o proxy sem perda da semântica; CA-FE-21; resposta antiga não substitui consulta atual.

### [ ] FE-T02 — Leitura de ajuste por ID e capacidades

- **Repositório:** API. **Depende:** nenhuma. **RF:** 013, 014. **Design:** D-05, seção 5.1.
- Adicionar GET `/v1/invoice-adjustments/{adjustment_id}` com request sem body e response `{data: InvoiceAdjustmentDetail}` conforme design.
- Carregar metadados de importação, parcela inferida, conciliação, contexto e cartão/mês persistidos. Capabilities são derivadas pela API; não aceitar capabilities do request.
- Resolver ID de cobrança mesmo se exibida em fatura movida; verificar ownership, status e relações. Documentar erros 401/404/500.
- **Pronto quando:** OpenAPI e handler concordam; ID alheio retorna 404; nenhuma origem/capacidade depende do nome; DTO não expõe fingerprint/token desnecessários.

### [ ] FE-T03 — PATCH parcial e manutenção consistente de ajustes

- **Repositório:** API. **Depende:** FE-T02. **RF:** 013, 014. **Design:** seção 5.2.
- Corrigir PATCH existente para diferenciar omissão de null, preservar dados omitidos e restringir campos por capabilities.
- Validar cartão/mês/ID da rota, sem realocar ajuste inadvertidamente. Auditoria e alteração na mesma transação; erros documentados e sem mensagens SQL.
- Manter fingerprint/vínculo de importação/assinatura. Validar arquivamento conciliado, reserva de identidade, reimportação após edição e semântica de projeção inferida arquivada.
- Pagamento importado continua sem editor nesta entrega. Não prometer auditoria de reason no arquivamento sem implementá-la/documentá-la.
- **Pronto quando:** testes de integração mostram contexto preservado, vínculo imutável e previsão restaurada; CA-FE-05/06/07/08/10. Request `{amount_cents:11500}` não apaga outros campos.

### [~] FE-T04 — Históricos: contratos atuais e alocação aditiva

- **Avanço em 04/10/2026:** Frontend consome arrays reais de payment-history e invoice-moves, exibe timestamps/contexto e eventos ordenados. current_allocation somente quando fornecido; não deduz a partir do último evento. Atualização do OpenAPI e extensão de alocação pertencem à API.

- **Repositórios:** API e frontend. **Depende:** FE-T01. **RF:** 016, 017. **Design:** D-03, 4.3.
- Documentar GET payment-history como `{data: PaymentMethodPeriod[]}`; adapter frontend produz periods/defaultMethod sem acessar data.periods.
- Documentar GET invoice-moves como `{data: InvoiceOccurrenceMove[]}`. Acrescentar current_allocation no topo preservando o array e resolvendo a lógica efetiva pelo backend.
- Definir resposta para identidade inválida, sem ocorrência e sem alocação de cartão; preservar 404 para ownership inexistente/alheio. Adicionar timestamps aos tipos usados.
- **Pronto quando:** contratos e fixtures cobrem arrays vazios, histórico válido, alocação calculada/movida e ausência de caminho; detalhe de dívida não quebra em paymentHistory.periods.

### [~] FE-T05 — Contrato de snapshot e fixtures históricas

- **Avanço em 04/10/2026:** Fixtures sintéticas v1–v4 e desconhecida criadas a partir do histórico da query da API, com datas integrais e campos opcionais antigos. Adapter e testes preservam o documento. Contrato histórico formal no OpenAPI/API ainda pendente; fontes em tests/fixtures/README.md.

- **Repositório:** API/documentação; tipos no frontend. **Depende:** nenhuma. **RF:** 020. **Design:** D-04, 4.6.
- Documentar documentos v1–v4, com release_2/release_3 e diferença entre DTO operacional e JSON histórico. Não alterar snapshots existentes.
- Criar fixtures sintéticas fiéis por versão a partir das migrations/tests, incluindo datas integrais, campos ausentes e conciliação quando capturada.
- Definir adapter de documento desconhecido com leitura estruturada; sem completar dados por GET atual.
- **Pronto quando:** payloads de fixtures validam o contrato de cada versão; original permanece byte/semanticamente imutável; CA-FE-20.

### [x] FE-T06 — Helpers de mês, moeda e janelas

- **Avanço em 04/10/2026:** Helpers, timezone do perfil e moeda da resposta aplicados aos defaults e consultas. Testes aprovados: virada do ano/timezone, zero, fim do plano, limites inclusivos 24/120 e quitação independente da janela. Evidência em quality-report.md.

- **Repositório:** frontend. **Depende:** nenhuma. **RF:** 002.
- Centralizar mês atual no timezone do perfil, operações de calendário, seleção de vigência e limites inclusivos.
- Janelas: resumo/plano de até 5 meses dentro do horizonte; faturas de até 24 meses dentro do horizonte operacional; cronograma de até 120 meses, padrão 12.
- Formatadores usam moeda da resposta. Registros/dívidas não ficam limitados ao horizonte de consulta do plano.
- **Pronto quando:** testes de borda verificam virada de ano, fim do plano, limite exato 24/120, timezone e zero explícito; CA-FE-19/24.

### [~] FE-T07 — Rotas, query e links entre recursos

- **Avanço em 04/10/2026:** Criadas /entries/[itemId], /cards/[cardId] e /planning/original com loading/error/not-found compartilhados. Filtros, abas, mês e retorno preservados; IDs codificados e all omitido do request. Build aprovado. Navegação autenticada permanece na aceitação integrada.

- **Repositório:** frontend. **Depende:** FE-T01, FE-T06. **RF:** 003 e navegação da spec.
- Criar estrutura das rotas `/entries/[itemId]`, `/cards/[cardId]`, `/planning/original` e query params definidos na seção 2 do design.
- Validar filtros e datas; all vira ausência de filtro. Preservar URL ao voltar de detalhe. IDs codificados; query sem tokens/payloads sensíveis.
- **Pronto quando:** links atuais continuam funcionando, rotas novas têm loading/error/not-found e estado restaurável; CA-FE-21/22.

## 3. Fase B — Cadastros e históricos

### [~] FE-T08 — Gestão de instituições

- **Avanço em 04/10/2026:** Lista com filtros ativa/arquivada, dialogs nativos de nome e arquivamento contextual, limites de texto e refresh; seleções excluem instituições arquivadas. Foco/Tab/Escape e cancelamento conferidos com dados sintéticos em desktop/mobile. POST/PATCH/archive reais pendentes de smoke.

- **Repositório:** frontend. **Depende:** FE-T01, FE-T07. **RF:** 010.
- Implementar lista/filtros e dialogs de editar nome/arquivar em `/cards`; usar GET/POST/PATCH/archive existentes.
- Name <=120; instituição arquivada não recebe novo cartão; sem cascata artificial.
- **Pronto quando:** CA-FE-01/21/22, com confirmação contextual e reload das seleções.

### [~] FE-T09 — Detalhe e configuração temporal de cartão

- **Avanço em 04/10/2026:** Detalhe por ID, nome separado de configuração temporal, end_month/context null explícitos; configuração selecionada pelo mês e ausência explícita, histórico completo e arquivo consultável. Perfil controla mês padrão; retorno preserva filtros. Sobreposição e dia nominal inválido dependem da aceitação contra API.

- **Repositório:** frontend. **Depende:** FE-T06, FE-T07. **RF:** 011.
- GET cartão; formulário de nome separado de POST `/changes`; payload completo explicitando end_month null.
- Mostrar vigências e configuração aplicável a month, não configurations.at(-1). Ausência de configuração tem estado explícito.
- Atualizar faturas/seleções após mudança; manter consulta de cartão arquivado.
- **Pronto quando:** CA-FE-02/03/24 e rejeição de sobreposição/configuração ausente tratadas.

### [~] FE-T10 — Detalhe de registros e valores

- **Avanço em 04/10/2026:** Lista kind/status e detalhe de valores completo, consulta mensal sem limitar cadastro ao horizonte; metadata PATCH só diferenças, assinatura preservada, arquivo contextual com effective_from opcional. Dívidas encaminham à rota especializada. Payloads testados; efeitos reais de arquivo ainda pendentes.

- **Repositório:** frontend. **Depende:** FE-T06, FE-T07. **RF:** 016.
- Listagem com filtros kind/status e detalhe com todos os períodos; valor vigente/futuro distinguido.
- Preservar metadados/assinatura, changes e archive com effective_from opcional. Dívidas encaminham para gestão especializada.
- **Pronto quando:** CA-FE-13/21, incluindo arquivamento de status versus encerramento a partir de competência e valores fora do horizonte.

### [~] FE-T11 — Pagamentos por vigência em itens e dívidas

- **Avanço em 04/10/2026:** Históricos completos em registros e dívidas com default direto, contexto/timestamps e cartão histórico. POST completo com credit_card_id null no direto, cartão obrigatório no crédito; renda não recebe ação de cartão. Testes de payload aprovados; mudança futura real ainda pendente.

- **Repositório:** frontend. **Depende:** FE-T04, FE-T10. **RF:** 016.
- Consumir arrays atuais via adapter; mostrar cartão/método, início/fim, contexto/timestamp e default direct.
- Enviar POST payment-changes completo; cartão obrigatório em credit_card e null em direct. Rendas sem ações de cartão.
- **Pronto quando:** CA-FE-14, trocas futuras preservam passado e histórico de dívida funciona sem assumptions do OpenAPI antigo.

### [~] FE-T12 — Histórico e ações de movimentação

- **Avanço em 04/10/2026:** Painel por item/competência lê array de eventos, ordena por instante e ID, liga faturas origem/destino e grava novo evento sem apagar histórico. URL restaura painel, abort evita atualização obsoleta e lock impede duplicação. Alocação atual depende do contrato aditivo da API; histórico real de duas movimentações ainda pendente.

- **Repositório:** frontend. **Depende:** FE-T04, FE-T07. **RF:** 017.
- GET eventos/alocação, painel por item+competência, cards históricos e links origem/destino.
- Manter POST de nova movimentação; refrescar origem/destino/resumo e permitir volta via novo evento, sem apagar histórico.
- **Pronto quando:** CA-FE-14/21, duas movimentações ordenadas, array vazio válido e alocação atual exibida pelo contrato aditivo.

## 4. Fase C — Faturas, ajustes e conciliação

### [~] FE-T13 — Navegação de faturas por período

- **Avanço em 04/10/2026:** Overview/lista com from/to, janelas <=24 e limites início do plano..fim+12, navegação e retorno do detalhe. Rótulos de pagamento e totais oficiais mantidos. Testes de janelas aprovados; atualização temporal e consulta autenticada pendentes.

- **Repositório:** frontend. **Depende:** FE-T06, FE-T07, FE-T09. **RF:** 012.
- Atualizar `/invoices/[cardId]/all` e overview para from/to; histórico/transbordo e janelas <=24 meses.
- Trocar rótulos de ciclo presumido por mês de pagamento; data nominal inválida não recebe correção automática.
- **Pronto quando:** CA-FE-03/04/24 e resposta vazia distinta de falha de serviço.

### [~] FE-T14 — Manutenção segura de ajustes

- **Avanço em 04/10/2026:** Editor preparado atrás de flag desabilitada por padrão. GET por ID antes de editar, capabilities restringem campos, PATCH só diferenças e rota persistida. Testes provam omissão/contexto e pagamento/arquivado sem editor. Não habilitar antes de T02/03; manutenção real permanece bloqueada.

- **Repositório:** frontend. **Depende:** FE-T01, FE-T02, FE-T03, FE-T13. **RF:** 013.
- Ler GET por ID antes de editar; formulário respeita editable_fields e capabilities. PATCH só campos modificados; contexto ausente não vira null.
- Construir rota de manutenção pelo cartão/mês persistidos, inclusive componente movido. Arquivar com confirmação e efeito informado; refrescar totais oficiais.
- **Pronto quando:** CA-FE-05/06; editor desabilitado para pagamento importado/arquivado, sem falsa promessa de reativação.

### [~] FE-T15 — Conciliação persistida no detalhe

- **Avanço em 04/10/2026:** Detalhe preparado mostra vínculo persistido e competência com link, sem planned_amount inventado; confirmação usa archive_effect das capabilities. Falha parcial do cadastro mantém ID/link ao registro existente após reload. Semântica de conciliação/reserva/arquivo depende de T02/03 e testes da API.

- **Repositório:** frontend. **Depende:** FE-T14, FE-T10. **RF:** 014.
- Mostrar vínculo de assinatura/competência a partir do GET de ajuste e links do registro. Não fabricar planned_amount_cents no detalhe persistido.
- Arquivamento conciliado explica restauração da previsão e reserva da identidade; futura previsão não recebe valor real do mês anterior.
- **Pronto quando:** CA-FE-07/08/11, incluindo falha parcial do cadastro e recuperação pelo registro existente após reload.

### [~] FE-T16 — Revisar importação e projeções de CSV

- **Avanço em 04/10/2026:** DTO CSV alinhado, metadados de parcelas/projeção, métricas oficiais com fallback de pagamento em magnitude, ambiguidade/backfill e preview preservados. Falha de confirmação invalida preview antes de reenviar; aborto ao desmontar. Testes locais de backfill/zero/totais aprovados; reimportação financeira real pendente.

- **Repositório:** frontend; regressões financeiras na API. **Depende:** FE-T01, FE-T15 para aceitação completa. **RF:** 015.
- Aproveitar importador existente; alinhar DTO real completo e metadados opcionais; validar multipart, alocações e token obsoleto.
- Bloqueio de ambiguidades, backfill sem novos lançamentos, métricas de previsão substituída e is_projected.
- **Pronto quando:** CA-FE-09/10/11/12, reimportação não duplica, pagamento recebido não é ignorado e descrição usa igualdade da API.

### [~] FE-T19 — Metadados de parcelas e origens na fatura

- **Avanço em 04/10/2026:** Detalhe da fatura exibe n/N, quitação, projeção CSV e origens; link de dívida somente com debt_id, link do registro quando existente. Valores assinados e totais oficiais preservados. Aceitação com faturas/importações reais pendente.

- **Repositório:** frontend. **Depende:** FE-T13, FE-T15. **RF:** 013, 015, 019.
- Renderizar n/N, early_settlement, is_projected, pagamento/importação e link da dívida somente quando houver debt_id.
- Manter valor assinado e separar compras/pagamentos/saldo sem totalizar novamente.
- **Pronto quando:** CA-FE-12/18/24, parcela CSV não aparece como dívida cadastrada.

## 5. Fase D — Dívidas e explicabilidade

### [~] FE-T17 — Cronograma de dívidas, filtros e liberações

- **Avanço em 04/10/2026:** Filtros as_of/status/projection_status e liberações com range retornado; detalhe de 12 meses, máximo 120, contagens do GET dívida e retorno preservado. Teste de estrutura de 180 parcelas aprovado; dívida longa e alterações reais ainda pendentes.

- **Repositório:** frontend. **Depende:** FE-T06, FE-T07, FE-T11. **RF:** 019.
- Lista/detalhe preservam as_of/status/projection_status; cronograma por janela de 12 meses, configurável até 120.
- Não calcular parcelas restantes pela janela; usar GET dívida. Liberações por período com range real no topo do response e sem somar como renda.
- **Pronto quando:** CA-FE-16, dívida de 180 parcelas funciona e mês de análise sobrevive à navegação.

### [~] FE-T18 — Quitação independente da janela visível

- **Avanço em 04/10/2026:** Input month usa limites estruturais independentes do schedule; 1/1 sem quitação, evento existente somente leitura com motivo/instante. Lock e ausência de retry garantem submissão única local. POST autenticado fora da janela/rejeições reais pendentes.

- **Repositório:** frontend. **Depende:** FE-T17. **RF:** 019.
- Substituir seleção derivada de todo o schedule por input month/validação de faixa estrutural; verificar mês pontual se necessário.
- POST retorna DebtEarlySettlement; reconsultar Debt depois. Evento único readonly com motivo/instante; 1/1 sem quitação antecipada.
- **Pronto quando:** CA-FE-16/17, quitação fora da janela carregada funciona e rejeição não provoca segunda submissão automática.

### [~] FE-T20 — Composição mensal completa

- **Avanço em 04/10/2026:** Composição por receita/compromissos/economia em dashboard e planejamento, total oficial/buckets/fontes, origem/competência/caixa/pagamento/alocação e links. Competência desconhecida não é inventada, categorias não são somadas novamente. Roteamento testado e fixture visual conferida; integração financeira pendente.

- **Repositório:** frontend. **Depende:** FE-T07, FE-T10, FE-T19. **RF:** 018.
- Tipar todos os metadados de MonthlySummarySource; painel por total/bucket/fontes com origens e links corretos.
- Não somar buckets, sources e fatura simultaneamente; fonte sem competência não ganha competência inventada.
- **Pronto quando:** CA-FE-15/23; Caixa padrão e Competência sob flag, com casos de dívida/ajuste/pagamento/assinatura.

### [x] FE-T21 — Visualizador do snapshot original

- **Avanço em 04/10/2026:** Visualizador v1–v4 com seções condicionais, captura/versão, moeda histórica, datas integrais, zero e versão desconhecida estruturada. Usa apenas GET original, sem reconstrução atual; JSON original disponível. Testes de imutabilidade e fixture v4 em browser aprovados.

- **Repositório:** frontend. **Depende:** FE-T05, FE-T07. **RF:** 020.
- Implementar rota e adapters v1–v4, seções condicionais, data/captura/versão e fallback de versão desconhecida.
- Sem consultas de dados atuais para reconstruir documento; sem links obrigatórios para interpretar recursos antigos.
- **Pronto quando:** CA-FE-20 e fixture desconhecida legível; atualizar plano/itens não altera a referência.

### [~] FE-T22 — Planejamento/economia e horizontes

- **Avanço em 04/10/2026:** Dashboard/planejamento limitam projeção a cinco meses dentro do horizonte, perfil/moeda/query e request ID. Economia exibe vigências/zero; plano ativo readonly e original acessível. Composição integrada. Ativação/economia autenticadas pendentes.

- **Repositório:** frontend. **Depende:** FE-T06, FE-T20. **RF:** 002, 016, 018.
- Ajustar dashboard/planning para não consultar além do horizonte; mostrar vigências da economia e zero configurado.
- Preservar ativação com response `{data:{plan,original}}`, rascunho editável/ativo readonly e estado de plano inexistente.
- **Pronto quando:** CA-FE-19/23, ativação não promete rota inexistente de ampliar/arquivar plano ativo.

## 6. Fase E — Aceitação e publicação

### [~] FE-T23 — Validação integrada de interface e acessibilidade

- **Avanço em 04/10/2026:** 14 testes de contratos/helpers aprovados; tipos/lint/build locais aprovados. Conferência de componentes com dados sintéticos em desktop e 390px, foco por Tab, Escape e confirmação contextual sem enviar mutações. Não substitui cenários financeiros da API, todos os fluxos autenticados ou auditoria de acessibilidade completa.

- **Repositórios:** frontend/API. **Depende:** FE-T08 a FE-T22 e contratos necessários.
- Usar fixtures/dados sintéticos para casos de CA-FE-01 a CA-FE-24, incluindo falha parcial, contexto preservado, arrays reais, mobile/teclado e resposta fora de ordem.
- Executar checks pertinentes: frontend tipos/lint/build; API checks e testes de contrato/integração afetados; migração incremental apenas se a implementação realmente exigir.
- **Pronto quando:** evidências individuais, sem apenas testes que espelhem JSX; CA-FE-21/22 e regressões financeiras aprovados.

### [~] FE-T24 — Relatório de qualidade e cobertura final

- **Avanço em 04/10/2026:** quality-report.md e pending-items.md criados com ambiente, matriz CA01–24, status por task, evidências e gates de contrato/produção. Relatório local concluído; aceitação final permanece parcial enquanto T02/03/04/23 não forem validadas.

- **Repositório:** documentação. **Depende:** FE-T23.
- Criar `quality-report.md` com ambiente/versões, matriz CA-FE → teste/evidência, contratos atualizados, limitações e status por task.
- Revisar links/rotas/requests e responses com a versão implementada; atualizar catálogo após correções, sem conservar divergências como se fossem behavior definitivo.
- **Pronto quando:** todos os RF/CA dentro do escopo cobertos; gaps não são marcados como resolvidos apenas por UI desabilitada.

### [ ] FE-T25 — Publicação e smoke test autenticado

- **Repositórios:** API/frontend/documentação. **Depende:** FE-T24; promoção/autorização conforme regras dos repositórios e ambiente de hospedagem definido.
- Publicar API/contratos primeiro, aplicar migrations se houver e só depois frontend dependente; preservar gates de PR e checks.
- Validar no ambiente publicado: sessão, leitura, alteração temporal, ajuste/arquivo conciliado, importação/backfill e dívida longa com dados apropriados de teste.
- **Pronto quando:** URLs/versões/resultados registrados e smoke autenticado aprovado. Health/build local não substituem evidência de produção. Sem resets de banco cloud.

## 7. Rastreabilidade dos critérios

| CA-FE | Tasks principais |
| --- | --- |
| 01 | T08 |
| 02 | T09 |
| 03 | T09, T13 |
| 04 | T13 |
| 05 | T02, T03, T14 |
| 06 | T03, T14 |
| 07 | T03, T15 |
| 08 | T03, T15 |
| 09 | T16 |
| 10 | T03, T16 |
| 11 | T15, T16 |
| 12 | T16, T19 |
| 13 | T10 |
| 14 | T11, T12 |
| 15 | T20 |
| 16 | T17, T18 |
| 17 | T18 |
| 18 | T19 |
| 19 | T06, T22 |
| 20 | T05, T21 |
| 21 | T01, T23 |
| 22 | T07, T23 |
| 23 | T20, T22 |
| 24 | T06, T09, T13, T19 |

FE-T23/24 consolidam todos os critérios; FE-T25 valida publicação. A numeração mantém T17/18 de dívida e T19 de composição da fatura como unidades independentes; a ordem visual das fases não substitui as dependências.

## 8. Pontos de bloqueio explícitos

- FE-T14/15 não podem concluir sem GET por identidade e manutenção segura (T02/03).
- Histórico completo de alocação não pode concluir sem a extensão aditiva e a correção de arrays (T04).
- Fixtures e visualizador de snapshot entregues no frontend (T05/21); formalização dos contratos históricos ainda pertence à API.
- T25 depende de hospedagem e promoção pelas regras do repositório; tasks documentadas não equivalem a deploy realizado.

Nenhum bloqueio impede executar as tasks independentes de navegação, helpers, instituições, cronogramas e estados enquanto os contratos são corrigidos.
