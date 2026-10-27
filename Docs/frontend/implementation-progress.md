# Entrega frontend — 04/10/2026

O escopo frontend executável com os contratos atuais foi implementado localmente. Manutenção de ajustes preparada e desabilitada até publicação dos contratos da API. Sem alteração de backend ou deploy; aceitação financeira integrada permanece pendente.

## Entregas

| Área | Comportamento |
| --- | --- |
| Transporte/sessão | Proxy preserva status/body/query/multipart; erros estruturados, 401 encaminha ao login, abort e POST sem repetição automática. Boundary com retry e 404 esperado separado de falha de serviço. |
| Instituições/cartões | Filtros, edição/arquivo contextual, detalhe por ID, histórico, configuração aplicável ao mês e ausência explícita. Nome separado de changes; arquivos consultáveis. |
| Registros | Filtros/detalhe/vigências/assinatura, metadata PATCH parcial, arquivo total ou por competência. Dívidas usam gestão especializada. |
| Pagamentos/movimentações | Arrays reais/contexto/timestamps/default direto e campos null explícitos. Eventos ordenados, links origem/destino; alocação somente quando retornada pela API. |
| CSV | DTO/parcelas/projeções, ambiguidade/backfill/zero, pagamentos em magnitude e preview invalidada após falha de confirmação. |
| Ajustes/conciliação | Editor preparado por ID/capabilities/rota persistida/PATCH diferenças, vínculo e efeitos de arquivo. Flag desabilitada por padrão. |
| Dívidas | Filtros/liberações com range real, cronograma 12 até 120 e contagens oficiais. Quitação independente da janela, lock e evento readonly. |
| Faturas | Até 24 meses no horizonte operacional, histórico/transbordo/retorno/origens/n/N e totais oficiais. |
| Dashboard/planejamento | Até 5 meses no plano, perfil/moeda/URL/request ID. Composição oficial e links. Economia com vigências/zero; plano ativo readonly. |
| Original | Rota v1–v4/seções/datas históricas/fallback desconhecido/JSON, sem reconstrução por GET atual. |

Mudanças locais prévias em cards-hub e globals.css preservadas. Fixtures são sintéticas; rota temporária de conferência removida.

## Verificação

14 testes aprovados, tipos/lint/build aprovados. Componentes reais conferidos pelo Next em desktop e 390 px; Tab/Escape/cancelamento e confirmação contextual. Não substitui fluxos autenticados ou testes financeiros da API.

Ambiente e matriz: [quality-report](quality-report.md). Próximos passos: [pending-items](pending-items.md). Status: [tasks](tasks.md).
