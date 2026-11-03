import Link from "next/link";
import type { MonthlySummary } from "@/lib/api/types";

export function VariableForecasts({ summary }: { summary: MonthlySummary }) {
  const sources = summary.excluded_variable_forecasts ?? [];
  if (summary.basis !== "cash" || sources.length === 0) return null;
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: summary.currency_code });
  return <aside className="debt-action-panel" aria-label="Previsões variáveis fora do caixa">
    <h2>Previsões variáveis</h2>
    <p><strong>{money.format((summary.excluded_variable_forecasts_cents ?? 0) / 100)}</strong> em estimativas, sem desconto no saldo de caixa exibido.</p>
    <ul>{sources.map((source) => <li key={source.source_id}>
      {source.item_id ? <Link href={`/entries/${encodeURIComponent(source.item_id)}`}>{source.name}</Link> : source.name}
      {" · "}{money.format(source.amount_cents / 100)}
    </li>)}</ul>
    <p>O restante do resumo continua sendo planejado; esse valor não representa saldo bancário realizado.</p>
  </aside>;
}
