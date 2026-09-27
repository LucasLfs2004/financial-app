# Release 3 — Validação do Marco D

## Escopo

O Marco D conclui pagamento e encerramento nas tasks R3-T08 a R3-T10:

- métodos de pagamento temporais para dívidas;
- quitação antecipada única e append-only;
- valores liberados por dívida e agregados por mês.

## Evidências funcionais

### R3-T08 — Métodos de pagamento

- `debt_installment` é classificado como despesa projetada no domínio;
- o banco aceita períodos `direct` e `credit_card` para dívidas e continua
  rejeitando rendas;
- cadastro com cartão persiste item, cronograma, dívida e método na mesma
  transação;
- ausência de método explícito mantém o fallback direto;
- troca temporal de cartão para direto preserva competências anteriores.

### R3-T09 — Quitação antecipada

- `POST` e `GET /v1/debts/{debt_id}/early-settlement` estão implementados;
- a escrita bloqueia a dívida, valida estado, intervalo e unicidade;
- o evento preserva os períodos originais, substitui a ocorrência do mês e
  suprime as posteriores;
- término efetivo e valor liberado são recalculados pelo projetor puro;
- uma segunda quitação retorna conflito sem alterar o primeiro evento.

### R3-T10 — Valores liberados

- `GET /v1/debt-releases` aceita intervalo mensal inclusivo;
- a resposta distingue `scheduled_completion` de `early_settlement`;
- liberações são ordenadas por dívida e agregadas por mês com `Money` seguro;
- a consulta não persiste item, renda, economia ou qualquer source financeira.

## Validações executadas

- reset completo do banco a partir de todas as migrations;
- 170 assertions pgTAP em 9 arquivos;
- lint do schema Supabase sem warnings;
- suíte Go completa com 183 testes em 41 pacotes;
- testes de integração reais para cartão → direto, quitação substitutiva,
  conflito de segunda quitação e múltiplas liberações no mesmo mês.

## Compatibilidade

- métodos de pagamento da Release 2 permanecem válidos;
- rendas continuam incompatíveis com métodos de pagamento;
- cronograma original e identidade item/competência permanecem preservados;
- valor liberado é somente projeção informativa e não participa dos totais.
