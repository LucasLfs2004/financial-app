# Validação do Marco C — Cadastro de cartões

Data: 03/09/2026

## Resultado

O Marco C da Release 2 foi concluído. A API agora oferece o ciclo de vida de
instituições financeiras, cartões de crédito e períodos mensais de configuração
com isolamento por usuário e preservação do histórico.

## Entregas

### R2-T06 — Instituições financeiras

- módulo vertical separado em domínio, aplicação, persistência e transporte;
- criação, listagem, edição e arquivamento;
- nomes ativos únicos por usuário;
- bloqueio de arquivamento enquanto houver cartão ativo;
- ownership aplicado em todas as operações.

### R2-T07 — Cartões e configurações

- criação transacional do cartão e do período inicial;
- listagem e consulta com instituição e configurações carregadas em uma única
  consulta, sem N+1;
- edição limitada ao nome, mantendo a instituição emissora imutável;
- mudança futura de vencimento e offset com fechamento do período anterior;
- validação de dia nominal entre 1 e 31 e offset entre 0 e 12;
- arquivamento sem exclusão dos períodos históricos;
- bloqueio de edição e novas mudanças após arquivamento;
- rotas HTTP registradas pelo próprio módulo.

## Evidências automatizadas

```text
supabase db reset --local --yes             aprovado
supabase test db supabase/tests --local     aprovado
supabase db lint --local --level warning    nenhum erro de schema
go test ./...                               aprovado
go test -race ./...                         aprovado
go vet ./...                                nenhum problema encontrado
go build ./cmd/api                          aprovado
docker build -t financial-api:t07 .         aprovado com Go 1.26.6
```

O cenário de integração da T07 comprovou criação atômica, normalização de nome,
duplicidade, ownership, instituição arquivada, edição, mudança a partir de uma
competência futura, preservação dos períodos e bloqueio do cartão arquivado.

## Preparação arquitetural encerrada

- a listagem de itens e períodos financeiros já utiliza consulta agregada sem
  N+1;
- a imagem de build foi alinhada ao toolchain Go 1.26.6 declarado pelo módulo;
- o README passou a refletir o estado real da Release 2 e as rotas do Marco C;
- a avaliação de extração de `Money`, `YearMonth` e `MonthInterval` permanece
  registrada para antes do encerramento da Release 2, conforme decisão de
  evolução incremental.

Nenhum ambiente Supabase Cloud foi alterado.
