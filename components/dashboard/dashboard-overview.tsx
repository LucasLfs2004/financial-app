"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { ArrowDownLeft, ArrowUpRight, ChevronRight, CreditCard, PiggyBank, Plus, WalletCards } from "lucide-react";
import { MonthNavigation } from "@/components/month-navigation";
import { FreeBalanceChart } from "./free-balance-chart";
import { MonthlyComposition } from "./monthly-composition";
import { VariableForecastFilter } from "./variable-forecast-filter";
import { browserApiFetch } from "@/lib/api/browser-api";
import { monthlySummaryPath } from "@/lib/api/summary-path";
import type { MonthlySummary, Plan } from "@/lib/api/types";
import { clampMonth, isMonth, projectionMonths } from "@/lib/debt-month";
import { featureFlags } from "@/lib/feature-flags";
import { commitmentGroups } from "@/lib/financial-overview";
import { summarySourceLink } from "@/lib/summary-source";

type Props = {
  name: string; initialMonth: string; initialSummaries: (MonthlySummary | null)[];
  hasPlan: boolean; plan: Plan | null; initialCategory?: "income" | "commitments" | "savings";
  initialBasis?: "cash" | "reference"; initialIncludeVariableForecasts?: boolean;
  apiAvailable: boolean;
};

export function DashboardOverview({ name, initialMonth, initialSummaries, hasPlan, plan, initialBasis = "cash", initialIncludeVariableForecasts = false, initialCategory, apiAvailable }: Props) {
  const [month, setMonth] = useState(initialMonth);
  const [includeVariableForecasts, setIncludeVariableForecasts] = useState(initialIncludeVariableForecasts);
  const [basis, setBasis] = useState<"cash" | "reference">(initialBasis);
  const [summaries, setSummaries] = useState(initialSummaries);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const requestId = useRef(0);
  const summary = summaries[0];
  const currencyCode = summary?.currency_code ?? plan?.currency_code ?? "BRL";
  const money = (cents: number) => new Intl.NumberFormat("pt-BR", { style: "currency", currency: currencyCode }).format(cents / 100);
  const months = plan ? projectionMonths(month, plan) : [];

  async function update(nextMonth: string, nextBasis = basis, include = includeVariableForecasts) {
    if (!isMonth(nextMonth) || !plan) return;
    nextMonth = clampMonth(nextMonth, plan.start_month, plan.end_month);
    const id = ++requestId.current;
    setMonth(nextMonth); setBasis(nextBasis); setIncludeVariableForecasts(include);
    const url = new URL(window.location.href);
    url.searchParams.set("month", nextMonth);
    url.searchParams.set("basis", nextBasis);
    url.searchParams.set("include_variable_forecasts", String(include));
    window.history.replaceState(null, "", url);
    setLoading(true); setError(""); setSummaries([]);
    const results = await Promise.all(projectionMonths(nextMonth, plan).map(async (value) => {
      try { return (await browserApiFetch<{ data: MonthlySummary }>(monthlySummaryPath(value, nextBasis, include))).data; }
      catch { return null; }
    }));
    if (id !== requestId.current) return;
    setSummaries(results); setLoading(false);
    if (results.some((value) => value === null)) setError("Alguns meses não puderam ser carregados.");
  }

  return <div className="finance-overview">
    <div className="finance-page-heading"><p>Olá, {name}. Confira seu mês.</p><Link className="finance-icon-button" aria-label="Adicionar receita ou despesa" href="/entries#new-item"><Plus size={22} /></Link></div>
    {plan && <MonthNavigation month={month} min={plan.start_month} max={plan.end_month} onChange={(value) => void update(value)} />}
    {!apiAvailable && <p className="finance-notice" role="status">Os dados financeiros estão indisponíveis. Tente carregar a página novamente.</p>}
    {!hasPlan ? <section className="finance-panel finance-empty"><h2>Comece pelo seu planejamento</h2><p>Defina o período e cadastre suas receitas e compromissos para descobrir quanto fica livre.</p><Link className="finance-primary-action" href="/planning">Criar planejamento <ChevronRight size={18} /></Link></section> : <>
      <details className="finance-filters"><summary>Visão: {basis === "reference" ? "Competência" : includeVariableForecasts ? "Caixa com previsões variáveis" : "Caixa sem previsões variáveis"}</summary><div>{basis === "cash" && <VariableForecastFilter include={includeVariableForecasts} onChange={(include) => void update(month, basis, include)} />}{featureFlags.referenceBasis && <label className="native-field"><span>Base da projeção</span><select value={basis} onChange={(event) => void update(month, event.target.value as "cash" | "reference")}><option value="cash">Caixa</option><option value="reference">Competência</option></select></label>}</div></details>
      {loading && <p role="status" className="finance-help">Atualizando seu planejamento…</p>}
      {error && <p role="alert" className="finance-notice">{error} <button type="button" onClick={() => void update(month)}>Tentar novamente</button></p>}
      <div className="finance-dashboard-grid" aria-busy={loading}>
        <div className="finance-dashboard-main">
          <section className={`finance-hero${summary?.is_negative ? " is-negative" : ""}`} aria-label="Resultado previsto do mês">
            <span className="section-kicker">{basis === "cash" ? "LIVRE PREVISTO" : "RESULTADO POR COMPETÊNCIA"}</span><strong>{summary ? money(summary.result_cents) : "—"}</strong><p>Após os compromissos e a reserva planejada.</p>
            <span className="finance-help">{basis === "cash" && !includeVariableForecasts ? "Não desconta estimativas variáveis." : "Inclui as despesas previstas."} Não é saldo bancário.</span>
          </section>
          {summary && <>
            <dl className="finance-arithmetic finance-panel"><div><dt><ArrowDownLeft size={18} />Receitas</dt><dd>{money(summary.income_cents)}</dd></div><div><dt><ArrowUpRight size={18} />Compromissos</dt><dd>{summary.commitments_cents < 0 ? "+" : "−"} {money(Math.abs(summary.commitments_cents))}</dd></div><div><dt><PiggyBank size={18} />Reserva planejada</dt><dd>− {money(summary.planned_savings_cents)}</dd></div></dl>
            <section className="finance-composition"><h2>O que compõe os compromissos?</h2><div className="finance-panel finance-row-list">{commitmentGroups(summary.sources).map((group) => <details className="finance-source-group" key={group.id}><summary>{group.id === "cards" ? <CreditCard size={18} /> : <WalletCards size={18} />}<span>{group.label}</span><strong>{money(group.amount)}</strong><ChevronRight size={16} /></summary><div className="finance-source-list">{group.sources.length === 0 ? <p className="finance-help">Nenhum compromisso neste grupo.</p> : group.sources.map((source, index) => {
              const href = summarySourceLink(source);
              return <div className="finance-source-row" key={`${source.source_id}-${index}`}><span>{href ? <Link href={href}>{source.name}</Link> : source.name}</span><strong>{money(source.amount_cents)}</strong></div>;
            })}{group.id === "cards" && <Link className="finance-text-link" href={`/cards?month=${month}`}>Ver cartões e faturas →</Link>}</div></details>)}</div></section>
            {(summary.excluded_variable_forecasts?.length ?? 0) > 0 && <aside className="finance-notice"><strong>{money(summary.excluded_variable_forecasts_cents ?? 0)} em estimativas fora deste saldo</strong><p>Para incluí-las na sobra prevista, altere a visão acima.</p><button type="button" onClick={() => void update(month, basis, true)}>Incluir previsões variáveis</button></aside>}
            <details className="finance-disclosure" open={initialCategory ? true : undefined}><summary>Consultar todas as origens e estimativas</summary><MonthlyComposition key={`${summary.month}-${summary.basis}-${includeVariableForecasts}`} summary={summary} initialCategory={initialCategory} /></details>
          </>}
        </div>
        <section className="finance-panel finance-forecast"><h2>{basis === "cash" ? "Quanto fica livre?" : "Resultado nos próximos meses"}</h2>{!loading && <FreeBalanceChart months={months} summaries={summaries} currencyCode={currencyCode} basis={basis} />}</section>
      </div>
    </>}
    <Link className="finance-primary-action" href="/entries#new-item"><Plus size={19} /> Adicionar receita ou despesa</Link>
    <div className="finance-secondary-links"><Link href="/debts">Dívidas e parcelas <ChevronRight size={16} /></Link><Link href="/planning">Planejamento e reserva <ChevronRight size={16} /></Link></div>
  </div>;
}
