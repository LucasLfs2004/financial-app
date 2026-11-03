# Fixtures sintéticas de snapshot

Documentos gerados a partir de `internal/planning/repository.go` da API e das migrations de cada release, com IDs e valores de teste. Não são snapshots exportados de usuários.

- v1: query de d8715b3; itens com plan_id, sem currency_code por item e sem release_2.
- v2: query de 8383726; adiciona release_2; payment_methods/adjustments via to_jsonb contêm plan_id e datas YYYY-MM-DD.
- v3: query de b621ad3; registros pertencem ao usuário, currency_code em item/adjustment, sem plan_id nas coleções desvinculadas.
- v4: query atual (base ad55dc0, metadados de assinatura em 6de54f8); release_3 inclui financial_item_id, períodos integrais, early_settlement com financial_item_id e referência em data integral. Campos recentes de conciliação foram incluídos conforme migration 20261004000100. Snapshots v4 anteriores a essa migration podem omiti-los.
- unknown: forma deliberadamente desconhecida para exercitar fallback, sem busca atual.

Os recursos operacionais e os históricos não compartilham um único DTO. O adapter conserva o documento original e apresenta apenas as seções presentes.
