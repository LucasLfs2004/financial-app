import Link from "next/link";
import type { MonthlySummary } from "@/lib/api/types";

export function FreeBalanceChart({ months, summaries, currencyCode, basis = "cash" }: {
  months: string[]; summaries: (MonthlySummary | null)[]; currencyCode: string; basis?: "cash" | "reference";
}) {
  const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: currencyCode, maximumFractionDigits: 2 });
  const scale = Math.max(1, ...summaries.map((summary) => Math.abs(summary?.result_cents ?? 0)));
  const hasNegative = summaries.some((summary) => summary && summary.result_cents < 0);
  const label = new Intl.DateTimeFormat("pt-BR", { month: "short", timeZone: "UTC" });
  return <div className="free-balance-chart" aria-label={basis === "cash" ? "Dinheiro livre previsto por mês" : "Resultado previsto por competência"}>
    <p className="finance-help">{basis === "cash" ? "Depois dos compromissos e da reserva, com os mesmos filtros do resumo." : "Receitas menos compromissos e reserva na base de competência."}</p>
    <div className={`free-chart-columns${hasNegative ? " has-negative" : ""}`}>
      {months.map((month, index) => {
        const summary = summaries[index];
        const value = summary?.result_cents;
        const negative = value != null && value < 0;
        return <div className={`free-chart-column${negative ? " is-negative" : ""}`} key={month}>
          <strong>{value == null ? "Indisponível" : money.format(value / 100)}</strong>
          <div className="free-chart-track" aria-hidden="true"><span className="free-chart-zero" /><span className="free-chart-bar" style={{ height: `${value == null ? 0 : Math.abs(value) / scale * (hasNegative ? 48 : 100)}%` }} /></div>
          <span>{label.format(new Date(`${month}-02T12:00:00Z`)).replace(".", "")}/{month.slice(2, 4)}</span>
          {value === 0 && <small>Sem sobra</small>}
        </div>;
      })}
    </div>
    <p className="finance-chart-scroll-hint finance-help">Deslize para consultar os próximos meses.</p>
    <Link className="finance-text-link" href="/planning">Ver planejamento completo →</Link>
  </div>;
}
