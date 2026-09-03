# Validação do Marco B — Persistência

Data: 01/09/2026

## Resultado

O Marco B da Release 2 foi concluído. O banco agora representa instituições,
cartões, vigências de configuração, meios de pagamento por competência,
ajustes de fatura e eventos de auditoria append-only.

## Entregas

### R2-T03 — Instituições e cartões

- enum `financial_resource_status`;
- tabelas `financial_institutions`, `credit_cards` e `credit_card_periods`;
- ownership garantido por FKs compostas;
- nomes ativos únicos sem diferenciação de caixa e espaços externos;
- instituição emissora imutável e cartão arquivado sem reativação;
- vigências mensais inclusivas sem sobreposição;
- RLS de leitura por usuário e ausência de grants diretos de escrita.

### R2-T04 — Métodos de pagamento

- enum `payment_method_kind`;
- tabela `financial_item_payment_periods`;
- consistência entre método direto/cartão e presença do cartão;
- períodos limitados ao horizonte do planejamento;
- vínculo permitido apenas para itens de despesa;
- ausência de período explícito preservada como fallback `direct`;
- vigências sem sobreposição, RLS e grants mínimos.

### R2-T05 — Ajustes e auditoria

- enums `invoice_adjustment_status` e `card_invoice_event_type`;
- tabelas `card_invoice_adjustments` e `card_invoice_audit_events`;
- valores não negativos e competências normalizadas;
- referência limitada ao plano e pagamento ao horizonte operacional de doze
  meses após o fim do plano;
- eventos de movimentação com formato mínimo obrigatório;
- eventos append-only, inclusive para conexões privilegiadas;
- índices para fatura, competência de referência, ajuste e última movimentação;
- RLS e grants somente de leitura para o cliente autenticado.

## Evidências automatizadas

```text
supabase db reset --local --yes             aprovado
supabase test db supabase/tests --local     136 testes aprovados em 6 arquivos
supabase db lint --local --level warning    nenhum erro de schema
go test ./...                               107 testes aprovados em 13 pacotes
go test -race ./...                         107 testes aprovados em 13 pacotes
go vet ./...                                nenhum problema encontrado
```

O reset reconstruiu o banco exclusivamente pelas migrations e pelo seed local.
Nenhum ambiente Supabase Cloud foi alterado.

## Segurança e integridade verificadas

- um usuário não consegue consultar recursos financeiros de outro usuário;
- cartões, períodos e ajustes não aceitam referências com ownership divergente;
- o papel `authenticated` não possui escrita direta nas novas tabelas;
- nomes ativos duplicados são rejeitados após normalização;
- vigências sobrepostas são rejeitadas no banco;
- renda não pode receber método de pagamento;
- eventos de auditoria não podem sofrer update ou delete.

## Observação de integração

Esta branch foi criada sobre `feature/r2-milestone-a`, pois o Marco A ainda não
estava presente em `origin/develop` no início do trabalho. O PR do Marco B deve
ser revisado depois do Marco A ou ter sua base atualizada após o merge dele.
