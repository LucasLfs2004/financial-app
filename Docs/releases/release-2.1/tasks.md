# Release 2.1 — Tasks — Registros financeiros contínuos

## Convenções

- `[ ]` pendente;
- `[~]` em andamento;
- `[x]` concluída.

## Marco A — Contrato e modelo

### R21-T01 — Congelar spec e design

- [x] definir plano como janela de projeção;
- [x] separar registros user-scoped de configurações plan-scoped;
- [x] definir moeda explícita, snapshots e compatibilidade;
- [x] registrar critérios de aceite e riscos de migração.

### R21-T02 — Atualizar OpenAPI

- [x] remover/depreciar `plan_id` dos registros;
- [x] adicionar `currency_code` ao item;
- [x] ajustar erros e exemplos de vigência;
- [x] validar o contrato.

## Marco B — Banco

### R21-T03 — Migrar ownership dos registros

- [x] adicionar e preencher moeda dos itens;
- [x] criar identidades compostas por usuário;
- [x] migrar FKs de períodos e métodos;
- [x] migrar ajustes e eventos de fatura;
- [x] remover validações e colunas de plano operacionais;
- [x] recriar índices e validar RLS;
- [x] adicionar testes pgTAP de migração e isolamento.

## Marco C — Aplicação

### R21-T04 — Refatorar itens financeiros

- [x] tornar CRUD independente do plano após a criação;
- [x] remover limites do horizonte;
- [x] persistir e retornar moeda;
- [x] atualizar testes unitários e HTTP.

### R21-T05 — Refatorar métodos e eventos

- [x] tornar métodos user/item-scoped;
- [x] tornar ajustes user/card-scoped;
- [x] tornar movimentações user/item/competência-scoped;
- [x] preservar locks, histórico e ownership.

### R21-T06 — Refatorar projeções

- [x] carregar dados de fatura por usuário;
- [x] filtrar itens pela moeda do plano;
- [x] manter limites apenas nas consultas de projeção;
- [x] reconciliar referência, caixa e fatura.

### R21-T07 — Evoluir ativação e snapshot

- [x] validar renda por interseção com o horizonte;
- [x] gerar snapshot schema v3;
- [x] capturar períodos completos dos itens alcançáveis;
- [x] preservar leitura e imutabilidade v1/v2.

## Marco D — Qualidade

### R21-T08 — Cobrir aceite ponta a ponta

- [x] testar plano `2026-09..2026-12` e seguro até `2027-09`;
- [x] testar método de pagamento além do plano;
- [x] testar reutilização em plano futuro;
- [x] testar economia isolada e moedas diferentes;
- [x] testar ownership e RLS.

### R21-T09 — Executar encerramento técnico

- [x] executar `go fmt`, `go vet` e `go test`;
- [x] resetar banco e executar pgTAP/lint;
- [x] validar OpenAPI;
- [x] atualizar mapa de releases e documentação operacional;
- [x] registrar migração, rollback e evidências.

## Dependências

```text
T01 → T02
 └──→ T03 → T04 → T05 → T06 → T07 → T08 → T09
```

## Definição de pronto

- registros financeiros sobrevivem ao planejamento;
- o horizonte limita visualização, não persistência;
- resumo, fatura e snapshot permanecem explicáveis;
- dados existentes são preservados;
- Release 3 não reintroduz `plan_id` nas dívidas;
- checks e critérios de aceite estão verdes.

