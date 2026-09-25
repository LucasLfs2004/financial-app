# Release 3 — Tasks — Quando essa dívida termina?

## Convenções

Estados:

- `[ ]` pendente;
- `[~]` em andamento;
- `[x]` concluída.

Cada task usa branch `feature/r3-tNN-descricao`, partindo de `develop` com a
Release 2 integral. Código, migration, testes, OpenAPI e documentação
proporcionais ao risco pertencem à própria task.

Enquanto a Release 2.1 ainda não estiver promovida para `develop`, as primeiras
branches R3 podem ser empilhadas sobre `feature/r21-user-owned-financial-records`.
Depois da promoção, devem ser redirecionadas para `develop` sem reescrever
commits já publicados.

## Pré-condição da release

- [x] revisar e congelar a spec funcional da Release 3;
- [ ] confirmar Release 2 promovida para `develop`;
- [ ] executar a suíte R2 no commit-base;
- [ ] validar migrations desde banco vazio;
- [ ] criar branches R3 somente desse baseline.

## Preparação arquitetural incremental

- [x] confirmar reuso de `Money`, `YearMonth` e intervalos;
- [x] definir porta de ocorrências consumível por fatura e resumo;
- [x] manter regras de dívida fora do módulo genérico `financialitem`;
- [ ] incorporar dívidas ao carregamento `repeatable read` sem N+1;
- [ ] registrar no design qualquer mudança de fronteira.

## Marco A — Contrato e domínio

### R3-T01 — Publicar contrato HTTP da Release 3

- [x] adicionar schemas de dívida, cronograma, ocorrência e liberação;
- [x] adicionar quitação antecipada;
- [x] adicionar `debt_installment` aos enums;
- [x] separar o enum aceito pela criação genérica para impedir dívida
  incompleta fora da rota especializada;
- [x] estender sources/componentes com metadados opcionais;
- [x] adicionar `debt_installments_cents` ao breakdown;
- [x] definir filtros, intervalos, erros e exemplos;
- [x] validar compatibilidade e OpenAPI.

Dependências: R2 em `develop` e spec aprovada.

Aceite: contrato congelado antes dos handlers, sem ambiguidade entre projeção
e realizado.

### R3-T02 — Implementar domínio e projetor puro

- [x] criar tipos de dívida, ocorrência, quitação e liberação;
- [x] derivar parcelas restantes e término;
- [x] numerar por competência e selecionar valor temporal;
- [x] aplicar quitação substitutiva;
- [x] calcular término efetivo, liberação e status;
- [x] garantir ordenação e soma segura;
- [x] testar bordas, virada de ano e overflow.

Dependências: R3-T01.

Aceite: nenhuma regra de cronograma depende de HTTP, relógio global ou banco.

## Marco B — Persistência

### R3-T03 — Evoluir item financeiro e persistir dívida

- [ ] adicionar enum em migration isolada;
- [ ] criar `debts` em migration posterior;
- [ ] adicionar FKs compostas, índices e constraints;
- [ ] validar kind, fórmula, limite estrutural e período mensal;
- [ ] adicionar RLS e grants mínimos;
- [ ] testar invariantes e ownership.

Dependências: R3-T02.

### R3-T04 — Persistir quitação antecipada

- [ ] criar `debt_early_settlements`;
- [ ] garantir unicidade por dívida;
- [ ] validar mês, valor e proprietário;
- [ ] bloquear update/delete;
- [ ] adicionar índices, FKs, RLS e grants;
- [ ] testar append-only e concorrência.

Dependências: R3-T02, R3-T03.

## Marco C — Cadastro e cronograma

### R3-T05 — Implementar cadastro e consulta

- [ ] criar módulo vertical `debt`;
- [ ] criar item, período e dívida em transação;
- [ ] listar com `as_of` e status;
- [ ] detalhar término e quantidade restante;
- [ ] editar metadados e arquivar;
- [ ] rejeitar criação genérica do kind;
- [ ] testar erros e ownership.

Dependências: R3-T01, R3-T03.

### R3-T06 — Implementar mudança de valor

- [ ] expor mudança especializada;
- [ ] bloquear dívida e períodos;
- [ ] fechar período anterior e criar novo até o término;
- [ ] preservar passado, numeração e término;
- [ ] impedir lacuna, sobreposição e mudança após quitação;
- [ ] testar primeira, intermediária e última parcela.

Dependências: R3-T05.

### R3-T07 — Expor cronograma projetado

- [ ] listar intervalo inclusivo;
- [ ] retornar número, total, término e pagamento;
- [ ] retornar cartão/fatura quando aplicável;
- [ ] omitir meses após quitação;
- [ ] testar ordenação, limites e conclusão.

Dependências: R3-T02, R3-T05, R3-T06.

## Marco D — Pagamento e encerramento

### R3-T08 — Integrar métodos de pagamento

- [ ] ampliar domínio e banco para aceitar dívida;
- [ ] manter fallback direto;
- [ ] permitir cartão por vigência;
- [ ] continuar rejeitando rendas;
- [ ] testar troca de cartão e retorno ao direto;
- [ ] manter testes R2 verdes.

Dependências: R3-T03, R3-T05 e módulo R2 de métodos.

### R3-T09 — Implementar quitação antecipada

- [ ] expor criação e consulta;
- [ ] validar estado, mês e valor;
- [ ] usar lock e unicidade;
- [ ] preservar cronograma original;
- [ ] substituir o mês e remover projeções posteriores;
- [ ] recalcular término e liberação;
- [ ] testar pagamento direto e conflitos.

Dependências: R3-T04, R3-T07, R3-T08.

### R3-T10 — Expor valores liberados

- [ ] listar liberações por intervalo;
- [ ] distinguir conclusão e quitação;
- [ ] retornar por dívida e agregar por mês;
- [ ] usar soma segura;
- [ ] provar que liberação não vira source;
- [ ] testar múltiplas dívidas no mesmo mês.

Dependências: R3-T07, R3-T09.

## Marco E — Integração financeira

### R3-T11 — Integrar parcelas às faturas

- [ ] carregar dívidas sem N+1;
- [ ] converter ocorrências em componentes;
- [ ] preservar identidade item/competência;
- [ ] aplicar cartão, offset e movimentação R2;
- [ ] transportar metadados de parcela;
- [ ] suportar quitação em cartão;
- [ ] testar duplicidade, transbordo e troca de cartão.

Dependências: R3-T07, R3-T08, R3-T09.

### R3-T12 — Integrar dívidas ao resumo mensal

- [ ] incluir parcelas em `reference`;
- [ ] incluir pagamento direto em `cash`;
- [ ] receber cartão pelo projetor de faturas;
- [ ] incluir quitação substitutiva;
- [ ] adicionar bucket e metadados;
- [ ] reconciliar total e sources;
- [ ] provar que liberação não é renda;
- [ ] manter suítes R1/R2 verdes.

Dependências: R3-T10, R3-T11.

### R3-T13 — Evoluir snapshot para schema v4

- [ ] definir documento determinístico;
- [ ] incluir dívidas, períodos, quitações e recursos R2 alcançáveis;
- [ ] usar v4 apenas em novas ativações;
- [ ] manter leitura v1/v2/v3;
- [ ] testar imutabilidade, ordem e idempotência;
- [ ] documentar evolução.

Dependências: R3-T05 a R3-T11.

## Marco F — Segurança e encerramento

### R3-T14 — Validar isolamento ponta a ponta

- [ ] criar dois usuários com dívidas distintas;
- [ ] tentar operações cruzadas;
- [ ] tentar quitar dívida alheia;
- [ ] tentar usar cartão alheio;
- [ ] validar API Go, Data API e RLS;
- [ ] revisar logs.

Dependências: R3-T05 a R3-T13.

### R3-T15 — Executar cenário de aceite

- [ ] cadastrar 12 parcelas iniciando em `5/12`;
- [ ] validar oito ocorrências e virada de ano;
- [ ] validar liberação sem source positiva;
- [ ] alterar valor preservando passado;
- [ ] vincular cartão offset 1;
- [ ] validar competência, fatura e caixa;
- [ ] mover uma parcela;
- [ ] quitar antecipadamente;
- [ ] validar transbordo e ausência de duplicidade;
- [ ] validar snapshot v4.

Dependências: R3-T12, R3-T13, R3-T14.

### R3-T16 — Qualidade final e documentação

- [ ] executar format, vet, testes e race detector;
- [ ] resetar banco e executar pgTAP/lint;
- [ ] validar OpenAPI e compatibilidade R1/R2;
- [ ] atualizar README e mapa;
- [ ] criar collection Postman e trilha R3;
- [ ] registrar quality report e riscos;
- [ ] documentar migration, rollback e smoke cloud.

Dependências: todas.

## Ordem sugerida

```text
R2 em develop
      ↓
T01 → T02
       ├→ T03 → T05 → T06 → T07 ───────┐
       │    └→ T04 ───────────────→ T09 ├→ T10 ─┐
       │                       T08 ──────┘       │
       └──────────────────────────────→ T11 ────┼→ T12
                                      └─────────┼→ T13
                                                ↓
                                              T14 → T15 → T16
```

R3-T03 e R3-T04 podem avançar em paralelo após T02, respeitando a ordem das
migrations. T08 pode avançar após T05; T11 espera ocorrência, método e
quitação estáveis.

## Definição de pronto da Release 3

- usuário cadastra dívida começando no meio do contrato;
- cronograma explica número, competência, valor e pagamento;
- término e liberação são derivados corretamente;
- mudança de valor preserva passado e numeração;
- pagamento direto/cartão funciona por vigência;
- parcela aparece uma vez na competência, fatura e caixa corretos;
- quitação substitui o mês e remove projeções posteriores;
- liberação nunca infla o resumo;
- faturas e resumos são reconciliáveis;
- snapshot v4 é determinístico e anteriores permanecem intactos;
- RLS e API isolam usuários;
- contrato, migrations, testes e documentação refletem o comportamento real.
