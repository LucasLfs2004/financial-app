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
- [ ] confirmar Release 2 promovida para `develop` (a árvore R3 segue em
  branches empilhadas; integração final pendente);
- [x] executar a suíte R2 no commit-base R3;
- [x] validar migrations desde banco vazio;
- [ ] criar branches R3 somente desse baseline (exceção histórica das branches
  R3 já publicadas, sem reescrita).

## Preparação arquitetural incremental

- [x] confirmar reuso de `Money`, `YearMonth` e intervalos;
- [x] definir porta de ocorrências consumível por fatura e resumo;
- [x] manter regras de dívida fora do módulo genérico `financialitem`;
- [x] incorporar dívidas ao carregamento `repeatable read` sem N+1;
- [x] registrar no design qualquer mudança de fronteira.

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

- [x] adicionar enum em migration isolada;
- [x] criar `debts` em migration posterior;
- [x] adicionar FKs compostas, índices e constraints;
- [x] validar kind, fórmula, limite estrutural e período mensal;
- [x] adicionar RLS e grants mínimos;
- [x] testar invariantes e ownership.

Dependências: R3-T02.

### R3-T04 — Persistir quitação antecipada

- [x] criar `debt_early_settlements`;
- [x] garantir unicidade por dívida;
- [x] validar mês, valor e proprietário;
- [x] bloquear update/delete;
- [x] adicionar índices, FKs, RLS e grants;
- [x] testar append-only e concorrência.

Dependências: R3-T02, R3-T03.

## Marco C — Cadastro e cronograma

### R3-T05 — Implementar cadastro e consulta

- [x] criar módulo vertical `debt`;
- [x] criar item, período e dívida em transação;
- [x] listar com `as_of` e status;
- [x] detalhar término e quantidade restante;
- [x] editar metadados e arquivar;
- [x] rejeitar criação genérica do kind;
- [x] testar erros e ownership.

Dependências: R3-T01, R3-T03.

### R3-T06 — Implementar mudança de valor

- [x] expor mudança especializada;
- [x] bloquear dívida e períodos;
- [x] fechar período anterior e criar novo até o término;
- [x] preservar passado, numeração e término;
- [x] impedir lacuna, sobreposição e mudança após quitação;
- [x] testar primeira, intermediária e última parcela.

Dependências: R3-T05.

### R3-T07 — Expor cronograma projetado

- [x] listar intervalo inclusivo;
- [x] retornar número, total, término e pagamento;
- [x] retornar cartão/fatura quando aplicável;
- [x] omitir meses após quitação;
- [x] testar ordenação, limites e conclusão.

Dependências: R3-T02, R3-T05, R3-T06.

## Marco D — Pagamento e encerramento

### R3-T08 — Integrar métodos de pagamento

- [x] ampliar domínio e banco para aceitar dívida;
- [x] manter fallback direto;
- [x] permitir cartão por vigência;
- [x] continuar rejeitando rendas;
- [x] testar troca de cartão e retorno ao direto;
- [x] manter testes R2 verdes.

Dependências: R3-T03, R3-T05 e módulo R2 de métodos.

### R3-T09 — Implementar quitação antecipada

- [x] expor criação e consulta;
- [x] validar estado, mês e valor;
- [x] usar lock e unicidade;
- [x] preservar cronograma original;
- [x] substituir o mês e remover projeções posteriores;
- [x] recalcular término e liberação;
- [x] testar pagamento direto e conflitos.

Dependências: R3-T04, R3-T07, R3-T08.

### R3-T10 — Expor valores liberados

- [x] listar liberações por intervalo;
- [x] distinguir conclusão e quitação;
- [x] retornar por dívida e agregar por mês;
- [x] usar soma segura;
- [x] provar que liberação não vira source;
- [x] testar múltiplas dívidas no mesmo mês.

Dependências: R3-T07, R3-T09.

## Marco E — Integração financeira

### R3-T11 — Integrar parcelas às faturas

- [x] carregar dívidas sem N+1;
- [x] converter ocorrências em componentes;
- [x] preservar identidade item/competência;
- [x] aplicar cartão, offset e movimentação R2;
- [x] transportar metadados de parcela;
- [x] suportar quitação em cartão;
- [x] testar duplicidade, transbordo e troca de cartão.

Dependências: R3-T07, R3-T08, R3-T09.

### R3-T12 — Integrar dívidas ao resumo mensal

- [x] incluir parcelas em `reference`;
- [x] incluir pagamento direto em `cash`;
- [x] receber cartão pelo projetor de faturas;
- [x] incluir quitação substitutiva;
- [x] adicionar bucket e metadados;
- [x] reconciliar total e sources;
- [x] provar que liberação não é renda;
- [x] manter suítes R1/R2 verdes.

Dependências: R3-T10, R3-T11.

### R3-T13 — Evoluir snapshot para schema v4

- [x] definir documento determinístico;
- [x] incluir dívidas, períodos, quitações e recursos R2 alcançáveis;
- [x] usar v4 apenas em novas ativações;
- [x] manter leitura v1/v2/v3;
- [x] testar imutabilidade, ordem e idempotência;
- [x] documentar evolução.

Dependências: R3-T05 a R3-T11.

## Marco F — Segurança e encerramento

### R3-T14 — Validar isolamento ponta a ponta

- [x] criar dois usuários com dívidas distintas;
- [x] tentar operações cruzadas;
- [x] tentar quitar dívida alheia;
- [x] tentar usar cartão alheio;
- [x] validar API Go, Data API e RLS;
- [x] revisar logs.

Dependências: R3-T05 a R3-T13.

### R3-T15 — Executar cenário de aceite

- [x] cadastrar 12 parcelas iniciando em `5/12`;
- [x] validar oito ocorrências e virada de ano;
- [x] validar liberação sem source positiva;
- [x] alterar valor preservando passado;
- [x] vincular cartão offset 1;
- [x] validar competência, fatura e caixa;
- [x] mover uma parcela;
- [x] quitar antecipadamente;
- [x] validar transbordo e ausência de duplicidade;
- [x] validar snapshot v4.

Dependências: R3-T12, R3-T13, R3-T14.

### R3-T16 — Qualidade final e documentação

- [x] executar format, vet, testes e race detector;
- [x] resetar banco e executar pgTAP/lint;
- [x] validar OpenAPI e compatibilidade R1/R2;
- [x] atualizar README e mapa;
- [x] criar collection Postman e trilha R3;
- [x] registrar quality report e riscos;
- [x] documentar migration, rollback e smoke cloud.

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
