# Release 1 — Tasks — Quanto está livre?

## Convenções

Estados:

- `[ ]` pendente;
- `[~]` em andamento;
- `[x]` concluída.

Uma task somente é concluída com código, migration, testes e documentação
proporcionais ao risco.

## Marco A — Contrato e domínio

### R1-T01 — Publicar contrato HTTP inicial

- [x] definir schemas de plano, item, período, economia e resumo;
- [x] definir erros;
- [x] definir `YearMonth`, dinheiro e `basis`;
- [x] adicionar exemplos;
- [x] validar OpenAPI.

Dependências: Release 0.

Aceite: todos os endpoints da spec existem no contrato antes dos handlers.

Status: concluída em 30/07/2026.

### R1-T02 — Implementar tipos de domínio

- [x] criar `Money`;
- [x] criar `YearMonth`;
- [x] criar enums;
- [x] criar intervalos mensais;
- [x] validar recorrência e offsets;
- [x] testar overflow, zero, meses inválidos e intervalos.

Dependências: R1-T01.

Aceite: nenhuma regra monetária utiliza ponto flutuante.

Status: concluída em 31/07/2026.

## Marco B — Persistência

### R1-T03 — Criar migration do planejamento

- [x] criar enums;
- [x] criar `plans`;
- [x] criar `plan_snapshots`;
- [x] adicionar índices e constraints;
- [x] adicionar RLS;
- [x] testar unicidade por usuário;
- [x] testar snapshot imutável.

Dependências: R1-T02.

Status: concluída em 13/08/2026.

### R1-T04 — Criar migration de itens e vigências

- [x] criar `financial_items`;
- [x] criar `financial_item_periods`;
- [x] criar constraints mensais;
- [x] impedir sobreposição;
- [x] adicionar índices;
- [x] adicionar RLS;
- [x] testar ownership e intervalos.

Dependências: R1-T02.

Status: concluída em 13/08/2026.

### R1-T05 — Criar migration de economia

- [x] criar `saving_periods`;
- [x] impedir sobreposição;
- [x] adicionar RLS;
- [x] testar ausência e zero explícito.

Dependências: R1-T02.

Status: concluída em 13/08/2026.

## Marco C — Planejamento

### R1-T06 — Criar planejamento em rascunho

- [x] implementar repositório;
- [x] implementar caso de uso;
- [x] implementar `POST /v1/plans`;
- [x] implementar `GET /v1/plans/current`;
- [x] implementar edição de rascunho;
- [x] tratar plano já existente;
- [x] testar isolamento.

Dependências: R1-T03.

Status: concluída em 13/08/2026.

### R1-T07 — Ativar planejamento

- [x] validar pré-condições;
- [x] criar snapshot determinístico;
- [x] ativar em transação;
- [x] tornar operação idempotente;
- [x] implementar consulta do original;
- [x] impedir alteração do snapshot;
- [x] testar concorrência.

Dependências: R1-T03, R1-T06, R1-T09, R1-T11.

Status: concluída em 20/08/2026.

## Marco D — Premissas financeiras

### R1-T08 — Cadastrar rendas

- [x] suportar recorrente;
- [x] suportar pontual;
- [x] suportar mês de referência;
- [x] suportar offset de recebimento;
- [x] validar horizonte;
- [x] expor criação e consulta;
- [x] testar renda defasada.

Dependências: R1-T04, R1-T06.

Status: concluída em 20/08/2026.

### R1-T09 — Cadastrar despesas

- [x] suportar fixa;
- [x] suportar variável projetada;
- [x] suportar recorrente e pontual quando permitido;
- [x] suportar pagamento direto;
- [x] validar horizonte;
- [x] expor criação e consulta;
- [x] permitir compromissos superiores à renda (cálculo negativo em R1-T13).

Dependências: R1-T04, R1-T06.

Status: concluída em 20/08/2026.

### R1-T10 — Alterar vigência

- [x] implementar mudança a partir de competência;
- [x] encerrar período anterior;
- [x] rejeitar sobreposição;
- [x] registrar alteração retroativa;
- [x] implementar arquivamento;
- [x] preservar histórico;
- [x] testar julho intacto após mudança em agosto.

Dependências: R1-T08, R1-T09.

Status: concluída em 20/08/2026.

### R1-T11 — Configurar valor para guardar

- [x] criar repositório;
- [x] criar caso de uso;
- [x] implementar PUT e GET;
- [x] distinguir ausência de zero;
- [x] suportar vigência;
- [x] testar alteração futura.

Dependências: R1-T05, R1-T06.

Status: concluída em 20/08/2026.

## Marco E — Resultado mensal

### R1-T12 — Implementar seleção de ocorrências

- [ ] selecionar períodos aplicáveis;
- [ ] expandir recorrência mensal;
- [ ] aplicar offset de caixa;
- [ ] impedir duplicidade;
- [ ] testar bases `reference` e `cash`.

Dependências: R1-T08, R1-T09, R1-T10, R1-T11.

### R1-T13 — Implementar cálculo puro

- [ ] somar rendas;
- [ ] somar compromissos;
- [ ] subtrair economia;
- [ ] suportar negativo;
- [ ] gerar breakdown;
- [ ] gerar sources;
- [ ] verificar consistência entre total e componentes.

Dependências: R1-T02, R1-T12.

### R1-T14 — Expor resumo mensal

- [ ] validar `AAAA-MM`;
- [ ] validar base;
- [ ] validar horizonte;
- [ ] identificar prévia de rascunho;
- [ ] implementar endpoint;
- [ ] documentar exemplos;
- [ ] testar respostas e erros.

Dependências: R1-T13.

## Marco F — Segurança e encerramento

### R1-T15 — Validar isolamento ponta a ponta

- [ ] criar dois usuários de teste;
- [ ] criar planos e itens distintos;
- [ ] tentar leitura cruzada por ID;
- [ ] tentar alteração cruzada;
- [ ] validar RLS pelo Data API;
- [ ] validar ownership pela API Go.

Dependências: R1-T06 a R1-T14.

### R1-T16 — Executar cenário de aceite

- [ ] cadastrar salário de R$ 6.000;
- [ ] cadastrar R$ 2.100 de fixas;
- [ ] cadastrar gasolina projetada de R$ 700;
- [ ] cadastrar R$ 1.200 para guardar;
- [ ] validar livre de R$ 2.000;
- [ ] validar detalhamento;
- [ ] alterar valor a partir de agosto;
- [ ] validar julho preservado;
- [ ] ativar e validar original.

Dependências: R1-T14, R1-T15.

### R1-T17 — Qualidade final

- [ ] executar `go test ./...`;
- [ ] executar testes de corrida;
- [ ] resetar banco do zero;
- [ ] validar migrations;
- [ ] validar OpenAPI;
- [ ] atualizar README;
- [ ] registrar decisões novas;
- [ ] revisar logs para ausência de dados financeiros;
- [ ] executar smoke test no ambiente cloud.

Dependências: todas.

## Ordem sugerida

```text
T01 → T02
       ├→ T03 → T06 ────────────────┐
       ├→ T04 → T08 → T10 ─┐       │
       │       └→ T09 ──────┤       ├→ T07
       └→ T05 → T11 ────────┤       │
                             ↓       │
                            T12 → T13 → T14
                                           ↓
                                          T15 → T16 → T17
```

## Definição de pronto da Release 1

- usuário autenticado cria um planejamento principal;
- itens e períodos ficam isolados por proprietário;
- renda, despesas e economia são cadastráveis;
- resumo mensal funciona em competência e caixa;
- todo total é explicável;
- mudança futura não altera o passado;
- rascunho pode ser ativado;
- original é imutável;
- cenário de aceite passa;
- contrato e documentação correspondem ao comportamento real.
