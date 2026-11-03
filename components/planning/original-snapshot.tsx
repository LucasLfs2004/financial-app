import type { OriginalPlanSnapshot } from "@/lib/api/types";
import { snapshotSections, historicMonth } from "@/lib/snapshot";
import { formatMonth } from "@/lib/debt-month";
import { formatDate } from "@/lib/format-date";

const labels: Record<string, string> = { name: "Nome", start_month: "Início", end_month: "Término", currency_code: "Moeda", items: "Registros", savings: "Economia", configured: "Configurada", periods: "Vigências", amount_cents: "Valor", kind: "Tipo", description: "Descrição", context: "Contexto", status: "Status", recurrence: "Recorrência", nominal_due_day: "Vencimento nominal", payment_month_offset: "Deslocamento do pagamento", cash_month_offset: "Deslocamento de caixa", scheduled_start_month: "Primeira parcela", scheduled_end_month: "Última parcela", total_installments: "Total de parcelas", first_projected_installment: "Primeira parcela projetada", original_total_cents: "Valor original", payment_methods: "Métodos de pagamento", early_settlement: "Quitação antecipada", reference_month: "Competência", payment_month: "Mês de pagamento", invoice_match_title: "Descrição na fatura", renewal_day: "Renovação nominal", reason: "Motivo", recorded_at: "Registrado em", created_at: "Criado em", updated_at: "Atualizado em", method: "Método", entry_kind: "Lançamento", import_source: "Origem da importação", transaction_date: "Data da transação", installment_number: "Parcela", installments_total: "Total de parcelas", is_projection: "Projetada", reconciled_reference_month: "Competência conciliada", archived_at: "Arquivado em" };
function FrozenData({ value, field = "", currency }: { value: unknown; field?: string; currency: string | null }) {
  if (value === null) return <span>Não informado</span>;
  if (Array.isArray(value)) return value.length ? <ol className="snapshot-records">{value.map((item, index) => <li key={index}><FrozenData value={item} currency={currency} /></li>)}</ol> : <p>Nenhum registro capturado.</p>;
  if (typeof value === "object") {
    const record = value as Record<string, unknown>;
    const recordCurrency = typeof record.currency_code === "string" && /^[A-Z]{3}$/.test(record.currency_code) ? record.currency_code : currency;
    return <dl className="snapshot-fields">{Object.entries(value as Record<string, unknown>).filter(([key]) => key !== "id" && !key.endsWith("_id") && !["import_fingerprint", "import_occurrence", "installment_series_key", "recorded_by"].includes(key)).map(([key, data]) => <div key={key}><dt>{labels[key] ?? key}</dt><dd><FrozenData value={data} field={key} currency={recordCurrency} /></dd></div>)}</dl>;
  }
  if (typeof value === "number" && field.endsWith("_cents")) return <span>{currency ? new Intl.NumberFormat("pt-BR", { style: "currency", currency }).format(value / 100) : `${value} centavos (moeda não informada)`}</span>;
  if (typeof value === "boolean") return <span>{value ? "Sim" : "Não"}</span>;
  if (typeof value === "string" && field.endsWith("month") && historicMonth(value)) return <span>{formatMonth(historicMonth(value)!)}</span>;
  if (typeof value === "string" && /_at$/.test(field)) return <span>{formatDate(value)}</span>;
  const terms: Record<string, string> = { active: "Ativo", archived: "Arquivado", fixed_expense: "Despesa fixa", variable_expense: "Despesa variável", income: "Receita", debt: "Dívida", monthly: "Mensal", one_time: "Pontual", direct: "Direto", credit_card: "Cartão", charge: "Compra", payment: "Pagamento", nubank_csv: "CSV Nubank" };
  return <span>{terms[String(value)] ?? String(value)}</span>;
}
export function OriginalSnapshot({ snapshot }: { snapshot: OriginalPlanSnapshot }) {
  const view = snapshotSections(snapshot.schema_version, snapshot.plan);
  const document = snapshot.plan && typeof snapshot.plan === "object" ? snapshot.plan as Record<string, unknown> : {};
  const currency = typeof document.currency_code === "string" && /^[A-Z]{3}$/.test(document.currency_code) ? document.currency_code : null;
  return <><p>Capturado em {formatDate(snapshot.captured_at)} · versão {snapshot.schema_version} · somente leitura.</p><p>Esta referência permanece como foi capturada. Alterações nos cadastros atuais não modificam este documento.</p>{!view.known && <p>Esta versão é exibida como documento estruturado.</p>}{view.sections.map((section) => <section className="debt-action-panel" key={section.title}><h2>{section.title}</h2><FrozenData value={section.data} currency={currency} /></section>)}<details className="debt-action-panel"><summary>Ver JSON original</summary><pre className="snapshot-json">{JSON.stringify(snapshot.plan, null, 2)}</pre></details></>;
}
