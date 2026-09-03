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

- [ ] implementar repositório e casos de uso;
- [ ] criar, listar e editar;
- [ ] arquivar preservando histórico;
- [ ] rejeitar arquivamento com cartão ativo;
- [ ] testar duplicidade e ownership.

Dependências: R2-T01, R2-T03.

### R2-T07 — Implementar cartões e configurações

- [ ] criar cartão e período inicial em transação;
- [ ] listar e consultar cartão;
- [ ] editar nome mantendo a instituição imutável;
- [ ] alterar vencimento/offset a partir de um mês;
- [ ] arquivar sem apagar faturas passadas;
- [ ] impedir novas operações em cartão arquivado;
- [ ] testar dia 31, offsets e mudança futura.

Dependências: R2-T01, R2-T03, R2-T06.

## Marco D — Vínculo e alocação

### R2-T08 — Implementar forma de pagamento por vigência

- [ ] expor criação de mudança de método;
- [ ] expor histórico ordenado;
- [ ] suportar `direct` e `credit_card`;
- [ ] encerrar período anterior em transação;
- [ ] preservar fallback direto anterior ao primeiro período;
- [ ] rejeitar renda, cartão alheio/arquivado e sobreposição;
- [ ] testar troca de cartão e retorno ao direto.

Dependências: R2-T04, R2-T07.

### R2-T09 — Implementar movimentação excepcional

- [ ] resolver ocorrência e alocação atual;
- [ ] mover para cartão/mês de destino;
- [ ] registrar origem, destino, motivo, autor e instante;
- [ ] expor histórico append-only;
- [ ] rejeitar mesmo destino e origem obsoleta;
- [ ] testar duas movimentações sucessivas.

Dependências: R2-T05, R2-T07, R2-T08.

## Marco E — Faturas projetadas

### R2-T10 — Implementar seleção de componentes

- [ ] selecionar ocorrências financeiras por referência;
- [ ] resolver método aplicável;
- [ ] derivar mês padrão pelo cartão;
- [ ] aplicar última movimentação;
- [ ] incluir ajustes ativos;
- [ ] distinguir referência conhecida e ausente;
- [ ] garantir unicidade da ocorrência;
- [ ] testar determinismo, overflow e transbordo.

Dependências: R2-T02, R2-T04, R2-T05, R2-T09.

### R2-T11 — Implementar ajustes consolidados

- [ ] criar ajuste em fatura selecionada;
- [ ] editar com evento de auditoria;
- [ ] arquivar com evento de auditoria;
- [ ] suportar referência opcional;
- [ ] validar horizonte operacional;
- [ ] testar zero, referência ausente e cartão arquivado.

Dependências: R2-T05, R2-T07.

### R2-T12 — Expor listagem e detalhe de faturas

- [ ] listar intervalo inclusivo de até 24 meses;
- [ ] retornar meses com componentes e omitir meses vazios por padrão;
- [ ] detalhar composição e origem;
- [ ] resolver vencimento nominal;
- [ ] garantir total igual à soma dos componentes;
- [ ] suportar fatura de transbordo;
- [ ] testar ordenação, erros e ownership.

Dependências: R2-T07, R2-T10, R2-T11.

## Marco F — Integração financeira

### R2-T13 — Integrar cartões ao resumo mensal

- [ ] preservar seleção da Release 1 para rendas e despesas diretas;
- [ ] excluir ocorrência de cartão do caminho direto em `cash`;
- [ ] incluir componentes de fatura como sources;
- [ ] incluir despesas de cartão normalmente em `reference`;
- [ ] incluir ajustes apenas quando a base permitir;
- [ ] adicionar metadados opcionais de fatura;
- [ ] provar por teste que não há dupla contagem;
- [ ] manter a suíte da Release 1 verde.

Dependências: R2-T10, R2-T11, R2-T12.

### R2-T14 — Evoluir fotografia original para schema v2

- [ ] definir documento v2 determinístico;
- [ ] incluir recursos da R2 alcançáveis pelo plano;
- [ ] usar v2 em novas ativações sem reescrever snapshots existentes;
- [ ] manter leitura de ambas as versões;
- [ ] testar imutabilidade e ordenação;
- [ ] documentar estratégia de evolução.

Dependências: R2-T06 a R2-T11.

## Marco G — Segurança e encerramento

### R2-T15 — Validar isolamento ponta a ponta

- [ ] criar dois usuários com instituições e cartões distintos;
- [ ] tentar leitura e alteração cruzadas;
- [ ] tentar vincular item a cartão alheio;
- [ ] tentar mover ocorrência para cartão alheio;
- [ ] validar RLS pelo Data API;
- [ ] validar ownership pela API Go;
- [ ] revisar logs para ausência de dados financeiros.

Dependências: R2-T06 a R2-T14.

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
