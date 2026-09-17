# Release 2 — Relatório final de qualidade

Data: 17/09/2026

## Resultado

A Release 2 está concluída e validada no ambiente local. Os Marcos A a G e as
tasks R2-T01 a R2-T17 atendem aos critérios registrados em `spec.md`,
`design.md` e `tasks.md`.

## Evidências

| Gate | Resultado |
| --- | --- |
| `go fmt ./...` | aprovado, sem alterações pendentes de formatação |
| `go vet ./...` | aprovado |
| `go build ./...` | aprovado |
| `go test ./... -count=1` com banco/Auth/Data API | todos os pacotes aprovados |
| `go test -race ./... -count=1` | aprovado, nenhuma corrida detectada |
| `supabase db reset --local --yes` | banco recriado e seis migrations aplicadas |
| `supabase test db supabase/tests --local` | 136 testes em 6 arquivos aprovados |
| `supabase db lint --local --level warning` | nenhum erro de schema |
| `govulncheck ./...` | nenhuma vulnerabilidade encontrada |
| Redocly CLI em `api/openapi.yaml` | contrato OpenAPI 3.1 válido |
| Newman — collection Release 2 | 24 requisições e 24 assertions aprovadas |

O lint do OpenAPI manteve quatro warnings não bloqueantes: licença ausente no
objeto `info`, servidor local e ausência de resposta `4XX` nos endpoints
públicos `/health` e `/ready`. Nenhum warning representa inconsistência do
contrato ou incompatibilidade funcional.

## Compatibilidade

- os endpoints da Release 1 permanecem registrados no contrato `/v1`;
- a suíte de integração da Release 1 passou sobre o banco recriado;
- itens sem forma de pagamento explícita continuam usando `direct`;
- os novos campos do resumo são aditivos, exceto a nulabilidade de
  `reference_month` já congelada na R2-T01 antes de consumidores externos;
- snapshots v1 continuam sendo devolvidos sem conversão e novas ativações usam
  schema v2.

## Segurança

O cenário R2-T15 comprovou ownership pela API Go e RLS pelo Data API com dois
usuários reais. Leitura, alteração, vínculo e movimentação cruzados foram
rejeitados. A revisão automatizada dos logs não encontrou tokens, nomes ou
valores financeiros, e o `govulncheck` não encontrou vulnerabilidades
alcançáveis.

## Riscos residuais

- o smoke test no Supabase Cloud está pendente porque o ambiente ainda não foi
  provisionado;
- os quatro warnings documentais do Redocly podem ser tratados em uma task de
  governança do contrato, sem impacto no deploy;
- a política de proteção e approvals de `main`/`develop` depende de
  configuração no provedor Git e permanece responsabilidade do mantenedor.

## Decisão

A Release 2 está apta para homologação. A promoção para produção deve seguir
`Docs/git-release-flow.md`, reutilizar o commit validado e executar o smoke
cloud descrito em `migration-rollback.md` antes da aprovação final.
