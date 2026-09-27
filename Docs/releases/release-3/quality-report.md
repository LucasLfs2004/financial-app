# Release 3 — Relatório de qualidade

Validação local: 26–27/09/2026.

## Escopo validado

- cadastro, mudança temporal, cronograma, pagamento direto e cartão;
- quitação substitutiva, valores liberados, faturas e resumo mensal;
- snapshot v4 determinístico, imutável e idempotente, com leitura legada;
- isolamento entre usuários via API Go, Data API e RLS;
- cenário de aceite completo com movimentação e transbordo de fatura.

## Evidências

| Check | Resultado |
| --- | --- |
| `go fmt ./...` | aprovado |
| `go vet ./...` | aprovado |
| `go build ./...` | aprovado |
| `docker build -t financial-api:r3-local .` após integrar `develop` | imagem criada |
| `go test ./...` com Supabase local | 204 testes aprovados em 41 pacotes |
| `go test -race ./...` com Supabase local | 204 testes aprovados em 41 pacotes |
| `supabase db reset --local --yes` | aprovado desde banco vazio |
| `supabase test db supabase/tests --local` | 170 assertions aprovadas em 9 arquivos |
| `supabase db lint --local --level warning` | sem erros |
| Redocly CLI em `api/openapi.yaml` | contrato 3.1 válido; 4 avisos preexistentes de documentação |
| `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | nenhuma vulnerabilidade encontrada |
| Newman na collection R3 | 18 requisições e 25 assertions aprovadas |

Os testes de segurança `TestRelease3EndToEndIsolation` usam dois usuários
reais do Auth: operações cruzadas recebem `404`, cartão alheio é rejeitado,
consultas da Data API retornam `[]` e os logs não contêm token, nomes nem
valores protegidos. `TestRelease3AcceptanceScenario` percorre a jornada por
HTTP, incluindo parcela 5/12, virada de ano, mudança futura, cartão offset 1,
movimentação, quitação e snapshot v4. A suíte R1/R2/2.1 continuou verde.

## Riscos e pendências externas

- O smoke no Supabase Cloud depende da aplicação das migrations R3 e do deploy
  da API desta branch no ambiente provisionado.
- A integração das branches R3 em `develop`, a revisão por PR, a promoção para
  `main`, a tag de produção e o deploy dependem do fluxo do provedor e do
  mantenedor. Esta validação local não equivale a publicação.

Ver [migration e rollback](./migration-rollback.md) para a ordem e os cuidados
de implantação.
