# Validação do Marco B — Persistência

Data: 26/09/2026

## Resultado

O Marco B da Release 3 foi concluído. O banco agora representa dívidas
parceladas como itens financeiros especializados e persiste quitações
antecipadas como eventos únicos e imutáveis.

## Entregas

### R3-T03 — Item financeiro e dívida

- valor `debt_installment` adicionado ao enum em migration isolada;
- tabela `debts` criada em migration posterior;
- identidade pública compartilhada com `financial_items`;
- ownership garantido por FK composta;
- término validado contra início, total e primeira parcela projetada;
- cronograma operacional limitado a 120 competências;
- períodos positivos, mensais, delimitados, contidos e sem lacunas;
- integridade cruzada validada no fim da transação para permitir mudanças
  temporais atômicas;
- RLS de leitura por usuário e ausência de grants diretos de escrita.

### R3-T04 — Quitação antecipada

- tabela `debt_early_settlements` com autor e instante de registro;
- uma única quitação permitida por dívida;
- competência normalizada entre a primeira e a penúltima parcela projetada;
- valor positivo, motivo limitado e autor igual ao proprietário;
- eventos protegidos contra `UPDATE` e `DELETE`, inclusive em conexões
  privilegiadas;
- unicidade usada como arbitragem de tentativas concorrentes;
- índice por usuário e competência, FK composta, RLS e grants mínimos.

## Evidências automatizadas

```text
supabase db reset --local --yes             aprovado
supabase test db supabase/tests --local     167 testes aprovados em 8 arquivos
supabase db lint --local --level warning    nenhum erro de schema
go test ./...                               164 testes aprovados em 38 pacotes
go test -race ./...                         164 testes aprovados em 38 pacotes
go vet ./...                                nenhum problema encontrado
```

O reset reconstruiu o banco exclusivamente pelas migrations e pelo seed local.
Nenhum ambiente Supabase Cloud foi alterado.

## Segurança e integridade verificadas

- itens `debt_installment` e metadados de dívida existem em relação um para um;
- dívida não aceita proprietário diferente do item financeiro;
- períodos inválidos, abertos, fora do cronograma ou com lacunas são rejeitados;
- usuários autenticados consultam somente as próprias dívidas e quitações;
- usuários autenticados não escrevem diretamente nas novas tabelas;
- a segunda quitação da mesma dívida é rejeitada pela restrição única;
- quitações persistidas não podem ser alteradas nem removidas.

## Observação de integração

As branches do Marco B foram empilhadas sobre `feature/r3-t02-debt-projector`,
pois as tasks iniciais da R3 ainda não estão presentes em `develop`. A ordem de
integração é T01, T02, T03 e T04, sem reescrever os commits já publicados.
