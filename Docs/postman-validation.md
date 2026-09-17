# Validação manual pelo Postman

As collections estão em:

- [`Financial API - Release 1.postman_collection.json`](../postman/Financial%20API%20-%20Release%201.postman_collection.json);
- [`Financial API - Release 2.postman_collection.json`](../postman/Financial%20API%20-%20Release%202.postman_collection.json).

Ambas usam o ambiente
[`Financial API - Local.postman_environment.json`](../postman/Financial%20API%20-%20Local.postman_environment.json).

## Preparar o ambiente

1. Inicie o Supabase local:

   ```bash
   supabase start
   supabase db reset
   ```

2. Copie a chave `anon`/publishable exibida por `supabase status` para a variável `supabasePublishableKey` do ambiente do Postman.

3. Inicie a API:

   ```bash
   go run ./cmd/api
   ```

4. Importe a collection e o environment e selecione `Financial API - Local`.

A collection cria usuários novos automaticamente. Portanto, não é necessário
preencher token, usuário ou IDs manualmente.

## Trilha da Release 1

Execute as pastas na ordem:

1. **01 - Saúde e autenticação** — confirma API viva, dependências prontas,
   cadastro no Supabase Auth e vínculo do token ao perfil.
2. **02 - Planejamento principal** — cria o planejamento único e verifica o
   estado inicial `draft`.
3. **03 - Premissas financeiras** — cadastra salário, fixas, gasolina e
   R$ 1.200 para guardar; a lista deve retornar três itens.
4. **04 - Resumo mensal** — julho deve resultar em R$ 2.000. A chamada
   `reference` usa julho como competência; `cash` demonstra o salário de junho
   recebido em julho.
5. **05 - Ativação e original** — altera o salário a partir de agosto, mantém
   julho em R$ 2.000, calcula agosto em R$ 2.500 e valida o snapshot original
   idempotente.
6. **06 - Isolamento entre usuários** — cria um segundo usuário e confirma que
   a API Go retorna `404` e a Data API retorna `[]` para dados do primeiro.

## Trilha da Release 2

Execute as pastas da collection da Release 2 em ordem:

1. **01 - Preparação** — cria usuário, plano, renda, aluguel, gasolina e
   poupança explícita;
2. **02 - Cartões e vínculos** — cria instituição, cartão offset `1`/dia `6`,
   cartão dia `31`, vínculos e ajuste consolidado;
3. **03 - Fatura e resumo** — valida total de R$ 1.880, competência, caixa,
   ausência de dupla contagem e vencimento inexistente;
4. **04 - Movimentação e transbordo** — move novembro para janeiro, consulta
   histórico e valida a fatura fora do horizonte;
5. **05 - Snapshot v2 e isolamento** — ativa o plano, inspeciona o snapshot v2
   e comprova ownership na API e RLS no Data API.

Execução automatizada opcional:

```bash
npm exec --yes newman -- run \
  'postman/Financial API - Release 2.postman_collection.json' \
  -e 'postman/Financial API - Local.postman_environment.json'
```

Na validação de encerramento foram aprovadas 24 requisições e 24 assertions,
sem falhas.

## Resultado esperado

Todas as requisições devem ficar verdes no Postman. Os cenários deixam
variáveis preenchidas no environment, como `planId`, `cardId`, `fuelId`,
`adjustmentId` e `originalSnapshotId`, para facilitar inspeção manual.

Para limpar os usuários e dados criados, rode `supabase db reset` novamente.
Esse comando é destrutivo para o banco local, mas não afeta qualquer projeto
Supabase Cloud.
