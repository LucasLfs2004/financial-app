# Pendências para revisão — 04/10/2026

O código frontend viável foi entregue. As pendências abaixo exigem backend, ambiente autenticado ou publicação. Não há deploy nesta entrega.

## 1. Manutenção de ajustes — FE-T02/03/14/15

**API:** publicar GET `/v1/invoice-adjustments/{adjustment_id}` com localização persistida, metadados e capabilities; corrigir PATCH para distinguir omissão de null e restringir campos pela capacidade real. Verificar ownership, auditoria, relações imutáveis, reserva de identidade, reimportação e restauração da previsão ao arquivar cobrança conciliada. Pagamento importado permanece sem editor.

O frontend contém o editor e testes de payload. `NEXT_PUBLIC_ENABLE_INVOICE_ADJUSTMENT_MAINTENANCE` é false por padrão. Habilitar somente depois de OpenAPI/handler/testes concordarem com o DTO proposto em [design](design.md). Flag desligada não conclui os critérios de manutenção/conciliação.

## 2. Históricos e alocação atual — FE-T04/12

**API:** corrigir documentação de payment-history e invoice-moves para arrays conforme os handlers. Publicar `current_allocation` como extensão aditiva do histórico, com resposta definida para identidade sem ocorrência/alocação. O frontend aceita arrays, mantém eventos e suporta o campo opcional; não infere alocação pelo último evento.

## 3. Contratos históricos — FE-T05

**API/documentação:** formalizar documentos de snapshot v1–v4 sem migrar nem reconstruir snapshots existentes. Fixtures sintéticas e visualizador existem no frontend; origem e diferenças em [tests/fixtures/README](../../tests/fixtures/README.md). Validar contrato formal contra os arquivos históricos quando disponível.

## 4. Aceitação autenticada — FE-T23/24

**Frontend/API:** executar em ambiente de teste com dados apropriados:

- Instituições/cartões: edição, arquivo sem cascata, sobreposição, ausência de configuração e dia nominal inexistente.
- Registros/pagamentos: mudança futura preservando passado, arquivo por competência, troca cartão/direto, falha parcial e recuperação pelo registro já criado.
- Movimentações: dois eventos, volta à origem, atualização das faturas/resumo e alocação fornecida pelo backend.
- CSV/conciliação: ambiguidade, igualdade do título aplicada pela API, backfill sem novos lançamentos, reimportação após edição/arquivo, pagamentos e parcelas projetadas; reserva de identidade e não duplicação.
- Dívida de 180 parcelas: paginação sem alterar contagens, quitação fora da janela, rejeições e evento único; liberações não somadas como renda.
- Planejamento: ativação/original, economia zero/atualização temporal, fim do plano, caixa/competência e fontes sem competência conhecida.
- Sessão/UX: expiração real, erros 400/404/409/422/500, resposta fora de ordem, reload/voltar e acessibilidade de todos os fluxos em mobile/teclado.

Os 14 testes locais e a conferência visual com fixtures não substituem essas evidências. A [matriz de qualidade](quality-report.md) registra o limite por critério.

## 5. Publicação — FE-T25

Definir ambiente, passar gates de versionamento/PR/checks e publicar API/contratos antes do frontend dependente. Depois executar smoke autenticado e registrar URLs/versões/resultados. Nenhuma migração ou alteração no banco foi aplicada nesta entrega.
