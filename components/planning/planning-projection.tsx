"use client";

import { useRef, useState } from "react";
import { ArrowDownLeft, ArrowRight, ArrowUpRight, CalendarRange, ChartNoAxesCombined, WalletCards } from "lucide-react";
import Link from "next/link";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { MonthlySummary } from "@/lib/api/types";
import { featureFlags } from "@/lib/feature-flags";

const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const monthName = new Intl.DateTimeFormat("pt-BR", { month: "short" });

function monthOffset(month: string, offset: number) {
  const [year, number] = month.split("-").map(Number);
  return new Date(Date.UTC(year, number - 1 + offset, 1)).toISOString().slice(0, 7);
}

function shortMonth(month: string) {
  return monthName.format(new Date(`${month}-02T12:00:00`)).replace(".", "");
}

export function PlanningProjection({ initialMonth, initialSummaries }: { initialMonth: string; initialSummaries: (MonthlySummary | null)[] }) {
  const [month, setMonth] = useState(initialMonth);
  const [basis, setBasis] = useState<"cash" | "reference">("cash");
  const [summaries, setSummaries] = useState(initialSummaries);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const requestId = useRef(0);
  const months = Array.from({ length: 5 }, (_, index) => monthOffset(month, index));
  const current = summaries[0];
  const maxValue = Math.max(1, ...summaries.flatMap((item) => item ? [item.income_cents, item.commitments_cents, Math.abs(item.result_cents)] : []));
  const hasData = summaries.some((item) => item && (item.income_cents || item.commitments_cents || item.result_cents));

  async function update(nextMonth: string, nextBasis: "cash" | "reference") {
    if (!/^\d{4}-\d{2}$/.test(nextMonth)) return;
    const currentRequest = ++requestId.current;
    setMonth(nextMonth);
    setBasis(nextBasis);
    setLoading(true);
    setError("");
    const results = await Promise.all(Array.from({ length: 5 }, async (_, index) => {
      try {
        const response = await browserApiFetch<{ data: MonthlySummary }>(`/plans/current/months/${monthOffset(nextMonth, index)}/summary?basis=${nextBasis}`);
        return response.data;
      } catch { return null; }
    }));
    if (currentRequest !== requestId.current) return;
    setSummaries(results);
    if (results.every((item) => item === null)) setError("Não foi possível carregar a projeção desse período.");
    setLoading(false);
  }

  return <section className="planning-projection" aria-label="Projeção do planejamento">
    <div className="planning-projection-heading"><div><span className="section-kicker">PROJEÇÃO</span><h2>Como os próximos meses se desenham</h2><p>Compare entradas, compromissos e saldo previstos.</p></div><ChartNoAxesCombined size={22} /></div>
    <div className="planning-projection-controls"><label><CalendarRange size={17} /><span>Mês inicial</span><input type="month" value={month} onChange={(event) => void update(event.target.value, basis)} /></label>{featureFlags.referenceBasis && <label><span>Visualização</span><select value={basis} onChange={(event) => void update(month, event.target.value as "cash" | "reference")}><option value="cash">Caixa</option><option value="reference">Competência</option></select></label>}</div>
    <div className="planning-projection-metrics"><div><span><ArrowDownLeft size={17} /> Receitas</span><strong>{current ? money.format(current.income_cents / 100) : "—"}</strong></div><div><span><ArrowUpRight size={17} /> Compromissos</span><strong>{current ? money.format(current.commitments_cents / 100) : "—"}</strong></div><div><span><WalletCards size={17} /> Saldo projetado</span><strong className={current?.is_negative ? "negative" : ""}>{current ? money.format(current.result_cents / 100) : "—"}</strong></div></div>
    {loading && <p className="chart-message" role="status">Atualizando projeção...</p>}
    {error && <p className="chart-message" role="alert">{error}</p>}
    {hasData ? <><div className="chart-legend"><span><i className="legend-income" /> Receitas</span><span><i className="legend-commitments" /> Compromissos</span><span><i className="legend-result" /> Saldo</span></div><div className="projection-chart" role="img" aria-label={`Receitas, compromissos e saldo projetado de ${shortMonth(month)} em diante`}>{months.map((value, index) => { const item = summaries[index]; return <div className="chart-month" key={value}><div className="chart-bars"><span className="bar income" style={{ height: `${item ? item.income_cents / maxValue * 100 : 0}%` }} title={`${shortMonth(value)}: receitas ${money.format((item?.income_cents ?? 0) / 100)}`} /><span className="bar commitments" style={{ height: `${item ? item.commitments_cents / maxValue * 100 : 0}%` }} title={`${shortMonth(value)}: compromissos ${money.format((item?.commitments_cents ?? 0) / 100)}`} /><span className={`bar result ${item?.is_negative ? "negative" : ""}`} style={{ height: `${item ? Math.abs(item.result_cents) / maxValue * 100 : 0}%` }} title={`${shortMonth(value)}: saldo ${money.format((item?.result_cents ?? 0) / 100)}`} /></div><small>{shortMonth(value)} {value.slice(0, 4)}</small></div>; })}</div></> : !loading && <div className="planning-projection-empty"><p>Os gráficos aparecem quando você cadastra receitas e despesas.</p><Link href="/entries#new-item">Cadastrar registro <ArrowRight size={16} /></Link></div>}
  </section>;
}
