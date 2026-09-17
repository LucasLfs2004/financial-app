# Release 2 — Tasks — Por que minha fatura é esse valor?

## Convenções

Estados:

- `[ ]` pendente;
- `[~]` em andamento;
- `[x]` concluída.

Cada task deve ser executada em branch `feature/r2-tNN-descricao`, partindo do
`develop` que já contenha a Release 1. Código, migration, testes, OpenAPI e
documentação proporcionais ao risco fazem parte da própria task.

## Pré-condição da release

- [x] aprovar e congelar a spec funcional da Release 2;
- [x] promover a Release 1 para `develop` por pull request;
- [x] confirmar a suíte da Release 1 verde no commit-base;
- [x] criar as branches da R2 somente a partir desse baseline.

O planejamento atual foi escrito sobre `feature/r1-milestone-f` porque
`develop` ainda aponta para a fundação. Essa dependência deve ser resolvida
antes de R2-T01, sem reescrever o histórico já publicado.

## Preparação arquitetural incremental

- [x] eliminar o N+1 da listagem de itens e períodos financeiros antes do
  Marco C, mantendo uma única consulta e a ordenação determinística;
- [x] alinhar a imagem de build com o toolchain Go declarado no módulo e
  atualizar o estado da R2 no `README.md` até o encerramento do Marco C;
- [ ] avaliar, antes do encerramento da release, a extração de `Money`,
  `YearMonth` e `MonthInterval` para um pacote financeiro compartilhado, sem
  ampliar a refatoração caso não haja benefício líquido nesta release.

Esses ajustes seguem a estratégia incremental: código novo adota o padrão
modular e código estável da Release 1 só é reorganizado quando a task que o
consome já precisa tocá-lo.

## Marco A — Contrato e domínio

### R2-T01 — Publicar contrato HTTP da Release 2

- [x] adicionar schemas de instituição, cartão e períodos de configuração;
- [x] adicionar forma de pagamento e histórico;
- [x] adicionar fatura, componente, ajuste e movimentação;
- [x] estender source do resumo com metadados opcionais;
- [x] tornar `reference_month` anulável somente para ajuste sem referência;
- [x] adicionar `card_invoice_adjustments_cents` ao breakdown;
- [x] definir paginação por intervalo, erros e exemplos;
- [x] validar compatibilidade e OpenAPI.

Dependências: Release 1 promovida para `develop`.

Aceite: todos os endpoints e enums da spec estão congelados antes dos handlers.

Status: concluída em 01/09/2026.

### R2-T02 — Implementar tipos e projetor puro

- [x] criar enums de status, método, componente, alocação e resolução de data;
- [x] validar dia nominal e offset;
- [x] implementar cálculo do mês padrão;
- [x] implementar resolução do vencimento nominal;
- [x] modelar identidade `item + reference_month`;
- [x] implementar soma segura e ordenação determinística;
- [x] testar meses inválidos, virada de ano e overflow.

Dependências: R2-T01.

Aceite: nenhuma regra de projeção depende de HTTP ou Postgres.

Status: concluída em 01/09/2026.

## Marco B — Persistência

### R2-T03 — Criar migration de instituições e cartões

- [x] criar enums compartilhados;
- [x] criar `financial_institutions`;
- [x] criar `credit_cards`;
- [x] criar `credit_card_periods`;
- [x] adicionar FKs compostas, índices e constraints;
- [x] impedir períodos sobrepostos;
- [x] adicionar RLS e grants mínimos;
- [x] testar ownership, nomes ativos e vigências.

Dependências: R2-T02.

Status: concluída em 01/09/2026.

### R2-T04 — Criar migration de métodos de pagamento

- [x] criar `payment_method_kind`;
- [x] criar `financial_item_payment_periods`;
- [x] validar cartão obrigatório ou nulo conforme método;
- [x] impedir vínculo de renda;
- [x] impedir sobreposição;
- [x] adicionar FKs compostas, índices e RLS;
- [x] testar fallback direto sem linha.

Dependências: R2-T02, R2-T03.

Status: concluída em 01/09/2026.

### R2-T05 — Criar migration de ajustes e auditoria

- [x] criar `card_invoice_adjustments`;
- [x] criar `card_invoice_audit_events`;
- [x] adicionar constraints de meses e dinheiro;
- [x] bloquear update/delete de eventos;
- [x] adicionar índices de fatura, referência e última movimentação;
- [x] adicionar RLS e grants mínimos;
- [x] testar append-only e isolamento.

Dependências: R2-T02, R2-T03.

Status: concluída em 01/09/2026.

## Marco C — Cadastro de cartões

### R2-T06 — Implementar instituições financeiras

- [x] implementar repositório e casos de uso;
- [x] organizar domínio, aplicação, persistência e transporte como módulo de
  referência para as novas features da R2;
- [x] criar, listar e editar;
- [x] arquivar preservando histórico;
- [x] rejeitar arquivamento com cartão ativo;
- [x] registrar as rotas do recurso sem ampliar o registro central do servidor;
- [x] testar duplicidade e ownership.

Dependências: R2-T01, R2-T03.

Status: concluída em 03/09/2026. Cenários unitários, de transporte e de
integração PostgreSQL/Auth aprovados.

### R2-T07 — Implementar cartões e configurações

- [x] criar cartão e período inicial em transação;
- [x] manter domínio, aplicação, persistência e transporte separados conforme
  o módulo de referência da R2-T06;
- [x] listar e consultar cartão;
- [x] editar nome mantendo a instituição imutável;
- [x] alterar vencimento/offset a partir de um mês;
- [x] arquivar sem apagar faturas passadas;
- [x] impedir novas operações em cartão arquivado;
- [x] testar dia 31, offsets e mudança futura.

Dependências: R2-T01, R2-T03, R2-T06.

Status: concluída em 03/09/2026. Cenários unitários, de transporte e de
integração PostgreSQL aprovados; validação consolidada em
[`milestone-c-validation.md`](milestone-c-validation.md).

## Marco D — Vínculo e alocação

### R2-T08 — Implementar forma de pagamento por vigência

- [x] expor criação de mudança de método;
- [x] expor histórico ordenado;
- [x] suportar `direct` e `credit_card`;
- [x] encerrar período anterior em transação;
- [x] preservar fallback direto anterior ao primeiro período;
- [x] rejeitar renda, cartão alheio/arquivado e sobreposição;
- [x] testar troca de cartão e retorno ao direto.

Dependências: R2-T04, R2-T07.

Status: concluída em 03/09/2026. Módulo vertical, rotas autenticadas,
transação com lock e testes de domínio validados.

### R2-T09 — Implementar movimentação excepcional

- [x] resolver ocorrência e alocação atual;
- [x] mover para cartão/mês de destino;
- [x] registrar origem, destino, motivo, autor e instante;
- [x] expor histórico append-only;
- [x] rejeitar mesmo destino e origem obsoleta;
- [x] testar duas movimentações sucessivas.

Dependências: R2-T05, R2-T07, R2-T08.

Status: concluída em 03/09/2026. Movimentações usam a auditoria append-only
existente, lock da ocorrência e ownership composto.

## Marco E — Faturas projetadas

### R2-T10 — Implementar seleção de componentes

- [x] dividir o projetor puro de `cardinvoice` em arquivos coesos antes de
  ampliar suas responsabilidades, sem alterar comportamento existente;
- [x] selecionar ocorrências financeiras por referência;
- [x] resolver método aplicável;
- [x] derivar mês padrão pelo cartão;
- [x] aplicar última movimentação;
- [x] incluir ajustes ativos;
- [x] distinguir referência conhecida e ausente;
- [x] garantir unicidade da ocorrência;
- [x] testar determinismo, overflow e transbordo.

Dependências: R2-T02, R2-T04, R2-T05, R2-T09.

Status: concluída em 09/09/2026. O seletor puro expande premissas financeiras,
resolve vigências de pagamento e cartão e reutiliza o projetor para
movimentações, ajustes, ordenação e soma segura. A leitura PostgreSQL usa uma
fotografia `repeatable read`, sem N+1, e o cenário integrado valida vigência,
transbordo, ajuste e ownership.

Exceção histórica: a branch publicada `feature/r2-t10-invoice-components`
implementou o trabalho posteriormente classificado como R2-T11. Seu histórico
foi preservado; a implementação efetiva desta task foi realizada em
`feature/r2-t10-component-selection`.

### R2-T11 — Implementar ajustes consolidados

- [x] criar ajuste em fatura selecionada;
- [x] editar com evento de auditoria;
- [x] arquivar com evento de auditoria;
- [x] suportar referência opcional;
- [x] validar horizonte operacional;
- [x] testar zero, referência ausente e cartão arquivado.

Dependências: R2-T05, R2-T07.

Status: concluída em 03/09/2026. CRUD de ajustes com horizonte operacional,
arquivamento e eventos append-only integrado às rotas autenticadas.

### R2-T12 — Expor listagem e detalhe de faturas

- [x] listar intervalo inclusivo de até 24 meses;
- [x] retornar meses com componentes e omitir meses vazios por padrão;
- [x] detalhar composição e origem;
- [x] resolver vencimento nominal;
- [x] garantir total igual à soma dos componentes;
- [x] suportar fatura de transbordo;
- [x] testar ordenação, erros e ownership.

Dependências: R2-T07, R2-T10, R2-T11.

Status: concluída em 17/09/2026. Listagem e detalhe reutilizam uma única
fotografia das fontes por requisição, respeitam o intervalo inclusivo de até 24
meses, omitem faturas vazias por padrão e retornam composição, vencimento,
instituição e origem de alocação conforme o contrato OpenAPI.

## Marco F — Integração financeira

### R2-T13 — Integrar cartões ao resumo mensal

- [x] usar o value object `Money` nas somas e subtrações do resumo, removendo
  os helpers aritméticos duplicados;
- [x] preservar seleção da Release 1 para rendas e despesas diretas;
- [x] excluir ocorrência de cartão do caminho direto em `cash`;
- [x] incluir componentes de fatura como sources;
- [x] incluir despesas de cartão normalmente em `reference`;
- [x] incluir ajustes apenas quando a base permitir;
- [x] adicionar metadados opcionais de fatura;
- [x] provar por teste que não há dupla contagem;
- [x] manter a suíte da Release 1 verde.

Dependências: R2-T10, R2-T11, R2-T12.

Status: concluída em 17/09/2026. O resumo reutiliza o seletor puro de
componentes, mantém o caminho direto da Release 1, usa `Money` em todos os
totais e diferencia competência e caixa sem duplicar ocorrências de cartão.

### R2-T14 — Evoluir fotografia original para schema v2

- [x] definir documento v2 determinístico;
- [x] incluir recursos da R2 alcançáveis pelo plano;
- [x] usar v2 em novas ativações sem reescrever snapshots existentes;
- [x] manter leitura de ambas as versões;
- [x] testar imutabilidade e ordenação;
- [x] documentar estratégia de evolução.

Dependências: R2-T06 a R2-T11.

Status: concluída em 17/09/2026. Novas ativações gravam schema v2 com recursos
R2 alcançáveis e ordenação estável; a leitura continua orientada por
`schema_version`, sem conversão ou regravação de fotografias v1.

## Marco G — Segurança e encerramento

### R2-T15 — Validar isolamento ponta a ponta

- [x] criar dois usuários com instituições e cartões distintos;
- [x] tentar leitura e alteração cruzadas;
- [x] tentar vincular item a cartão alheio;
- [x] tentar mover ocorrência para cartão alheio;
- [x] validar RLS pelo Data API;
- [x] validar ownership pela API Go;
- [x] revisar logs para ausência de dados financeiros.

Dependências: R2-T06 a R2-T14.

Status: concluída em 17/09/2026. O cenário E2E usa dois usuários reais do
Supabase Auth e comprova isolamento na API Go e no Data API para instituições,
cartões, faturas, vínculos e movimentações. Os logs HTTP não expõem tokens,
nomes ou valores financeiros.

### R2-T16 — Executar cenário de aceite

- [ ] cadastrar instituição e cartão com offset 1/dia 6;
- [ ] vincular aluguel e gasolina;
- [ ] adicionar ajuste consolidado;
- [ ] validar total e composição;
- [ ] validar novembro em `reference` e dezembro em `cash`;
- [ ] provar ausência de dupla contagem;
- [ ] trocar cartão a partir de agosto;
- [ ] mover uma ocorrência e validar histórico;
- [ ] validar dia inexistente e transbordo;
- [ ] ativar rascunho e validar snapshot v2.

Dependências: R2-T13, R2-T14, R2-T15.

### R2-T17 — Qualidade final e documentação operacional

- [ ] executar `go fmt ./...`;
- [ ] executar `go vet ./...`;
- [ ] executar `go test ./...` e testes de corrida;
- [ ] resetar banco do zero e validar migrations/testes SQL;
- [ ] executar lint do banco;
- [ ] validar OpenAPI e compatibilidade com Release 1;
- [ ] atualizar README e mapa de releases;
- [ ] criar collection Postman da Release 2 e trilha de validação;
- [ ] registrar relatório de qualidade e riscos residuais;
- [ ] documentar migration, rollback e smoke cloud pendente.

Dependências: todas.

## Ordem sugerida

```text
R1 em develop
      ↓
T01 → T02
       ├→ T03 → T06 → T07 ──────────────┐
       │     └→ T04 → T08 → T09 ─┐      │
       └────────→ T05 ────────────┼→ T10 ├→ T12 → T13
                                  └→ T11 ┘           │
                                      └────────→ T14 ┤
                                                     ↓
                                                    T15 → T16 → T17
```

R2-T03, R2-T04 e R2-T05 podem ser desenvolvidas em paralelo depois de R2-T02,
desde que as migrations mantenham ordem determinística e sejam integradas por
PRs pequenos.

## Definição de pronto da Release 2

- usuário autenticado cadastra instituição e cartões;
- configuração do cartão e forma de pagamento preservam histórico mensal;
- despesas vinculadas geram faturas projetadas;
- toda fatura explica exatamente seu total;
- ajustes sem compra individual são suportados;
- movimentações excepcionais são auditáveis;
- competência e caixa apresentam a mesma ocorrência no mês correto;
- resumo mensal não duplica fatura e componente;
- faturas de transbordo são consultáveis;
- snapshot v2 é determinístico e snapshots v1 permanecem intactos;
- RLS e API isolam dois usuários;
- contrato, migrations, testes, Postman e documentação correspondem ao
  comportamento real.
