# Relatório de qualidade — Release 1

> Data: 20/08/2026  
> Escopo: Marcos A a F  
> Ambiente validado: Supabase local-first

## Resultado

A Release 1 está funcionalmente validada no ambiente local. O cenário de
aceite, o isolamento entre usuários, o contrato HTTP, as migrations e as
verificações estáticas passaram.

O smoke test cloud não foi executado porque o projeto cloud ainda não foi
provisionado. Isso está de acordo com a decisão
[`decisao-ambiente-supabase-local-first.md`](../../decisao-ambiente-supabase-local-first.md),
segundo a qual nenhuma release funcional depende de cloud durante esta fase.

## Evidências executadas

| Verificação | Resultado |
| --- | --- |
| `go test -race ./...` | passou |
| cenário ponta a ponta sem cache | passou |
| `go vet ./...` | passou |
| `go build ./...` | passou |
| Staticcheck | passou sem achados |
| Govulncheck | passou sem vulnerabilidades alcançáveis |
| `supabase db reset` | banco recriado somente pelas migrations |
| pgTAP | 70 testes passaram |
| `supabase db lint --local --level warning` | nenhum erro de schema |
| Redocly CLI | OpenAPI válido |

O Redocly manteve quatro avisos não bloqueantes:

- licença ainda não definida pelo proprietário do projeto;
- servidor local intencional enquanto o projeto é local-first;
- endpoints técnicos `/health` e `/ready` não possuem respostas `4XX`, pois
  não recebem entrada do consumidor que gere erro de validação.

## Segurança e isolamento

O teste `TestRelease1AcceptanceAndIsolation` cria dois usuários pelo Supabase
Auth e valida:

- planejamento e itens distintos;
- leitura cruzada por ID retornando `404` na API Go;
- alteração cruzada retornando `404` na API Go;
- consultas cruzadas vazias para planos, itens, períodos e economia pela Data
  API com RLS;
- totais distintos e isolados nos resumos mensais.

Os logs HTTP registram método, path, status, duração e request ID. Não registram
token, query string, corpo, nomes dos itens ou valores financeiros. Existe um
teste automatizado específico para essa fronteira.

## Cenário de aceite

O cenário oficial foi executado pela API HTTP:

```text
Renda:          R$ 6.000
Fixas:          R$ 2.100
Gasolina:       R$   700
Guardar:        R$ 1.200
Resultado livre R$ 2.000
```

O detalhamento contém as quatro fontes. Após alterar a renda a partir de agosto,
julho permanece em R$ 2.000 e agosto passa a R$ 2.500. A ativação cria uma única
fotografia original e chamadas repetidas retornam a mesma referência.

## Correções originadas pela revisão

- `github.com/jackc/pgx/v5` atualizado para `v5.9.2`;
- `golang.org/x/text` atualizado para `v0.39.0`;
- toolchain fixado em Go `1.26.6` para incorporar correções da biblioteca
  padrão;
- handlers agora rejeitam campos obrigatórios ausentes, distinguindo ausência,
  zero explícito e `null` conforme o OpenAPI;
- descrições das tags do OpenAPI foram adicionadas.

## Riscos residuais

- o smoke cloud permanece pendente até o provisionamento do ambiente;
- os pacotes de repositório têm cobertura unitária baixa, compensada nesta
  release por pgTAP e testes integrados reais; ampliar testes isolados de falhas
  do Postgres é uma melhoria futura;
- a licença do repositório precisa ser definida antes de publicação pública.
