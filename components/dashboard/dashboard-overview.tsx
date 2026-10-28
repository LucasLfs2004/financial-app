"use client";
import { VariableForecastFilter } from "@/components/dashboard/variable-forecast-filter";
import { monthlySummaryPath } from "@/lib/api/summary-path";

import { useRef, useState } from "react";
import Link from "next/link";
import { ArrowDownLeft, ArrowRight, ArrowUpRight, CalendarRange, Check, ChevronDown, CreditCard, Landmark, PiggyBank, Plus, ReceiptText, WalletCards } from "lucide-react";
import { MonthlyComposition } from "@/components/dashboard/monthly-composition";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { MonthlySummary, Plan } from "@/lib/api/types";
import { clampMonth, isMonth, projectionMonths } from "@/lib/debt-month";
import { featureFlags } from "@/lib/feature-flags";

const monthName = new Intl.DateTimeFormat("pt-BR", { month: "short" });

function formatMonth(month: string) {
  return monthName.format(new Date(`${month}-02T12:00:00`)).replace(".", "");
}

type Props = {
  name: string;
  initialMonth: string;
  initialSummaries: (MonthlySummary | null)[];
  hasPlan: boolean;
  plan: Plan | null;
  initialCategory?: "income" | "commitments" | "savings";
  initialBasis?: "cash" | "reference";
  initialIncludeVariableForecasts?: boolean;
  cardCount: number | null;
  debtCount: number | null;
  apiAvailable: boolean;
};

export function DashboardOverview({ name, initialMonth, initialSummaries, hasPlan, plan, initialBasis = "cash", initialIncludeVariableForecasts = false, initialCategory, cardCount, debtCount, apiAvailable }: Props) {
  const [month, setMonth] = useState(initialMonth);
  const [includeVariableForecasts, setIncludeVariableForecasts] = useState(initialIncludeVariableForecasts);
  const [basis, setBasis] = useState<"cash" | "reference">(initialBasis);
  const [summaries, setSummaries] = useState(initialSummaries);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const requestId = useRef(0);
  const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: summaries[0]?.currency_code ?? plan?.currency_code ?? "BRL" });
  const summary = summaries[0];
  const months = plan ? projectionMonths(month, plan) : [];
  const maxValue = Math.max(1, ...summaries.flatMap((item) => item ? [item.income_cents, item.commitments_cents, Math.abs(item.result_cents)] : []));
  const hasChartData = summaries.some((item) => item && (item.income_cents !== 0 || item.commitments_cents !== 0 || item.result_cents !== 0));

  async function update(nextMonth: string, nextBasis: "cash" | "reference", nextIncludeVariableForecasts = includeVariableForecasts) {
    if (!isMonth(nextMonth) || !plan) return;
    nextMonth = clampMonth(nextMonth, plan.start_month, plan.end_month);
    const currentRequest = ++requestId.current;
    setMonth(nextMonth);
    setBasis(nextBasis);
    setIncludeVariableForecasts(nextIncludeVariableForecasts);
    const url = new URL(window.location.href);
    url.searchParams.set("month", nextMonth);
    url.searchParams.set("basis", nextBasis);
    url.searchParams.set("include_variable_forecasts", String(nextIncludeVariableForecasts));
    window.history.replaceState(null, "", url);
    if (!hasPlan) return;
    setSummaries([]);
    setLoading(true);
    setError("");
    const nextMonths = projectionMonths(nextMonth, plan);
    const results = await Promise.all(nextMonths.map(async (value) => {
      try {
        const response = await browserApiFetch<{ data: MonthlySummary }>(monthlySummaryPath(value, nextBasis, nextIncludeVariableForecasts));
        return response.data;
      } catch { return null; }
    }));
    if (currentRequest !== requestId.current) return;
    setSummaries(results);
    if (results.some((item) => item === null)) setError("Não foi possível carregar todos os meses. Tente novamente.");
    setLoading(false);
  }

  const metrics = [
    { label: "Saldo projetado", value: summary?.result_cents, icon: WalletCards, tone: "mint", hint: basis === "cash" ? (includeVariableForecasts ? "Inclui previsões variáveis e reserva" : "Sem previsões variáveis; mantém compromissos e reserva") : "Depois dos compromissos e da reserva" },
    { label: "Receitas", value: summary?.income_cents, icon: ArrowDownLeft, tone: "mint", hint: "Entradas do mês" },
    { label: "Compromissos", value: summary?.commitments_cents, icon: ArrowUpRight, tone: "rose", hint: "Despesas e pagamentos previstos" },
    { label: "Para guardar", value: summary?.planned_savings_cents, icon: PiggyBank, tone: "blue", hint: "Reserva planejada para o mês" },
  ];

  return <>
    <div className="overview-content">
      <div className="overview-intro"><div><p>Olá, {name}. Confira sua projeção.</p></div><details className="new-entry-menu"><summary><Plus size={19} /> Novo cadastro <ChevronDown size={16} /></summary><div><Link href="/debts#new-debt">Adicionar dívida</Link><Link href="/cards#new-card">Cadastrar cartão</Link><Link href="/entries#new-item">Receita ou despesa</Link></div></details></div>
      {!apiAvailable && <div className="overview-notice">Os dados financeiros estão indisponíveis agora. Você ainda pode acessar as áreas de cadastro pelo menu.</div>}
      <section className="overview-shortcuts" aria-label="Cadastros rápidos"><Link href="/debts#new-debt"><span className="shortcut-icon"><Landmark size={23} /></span><span><strong>Adicionar dívida</strong><small>Parcelas e compromissos</small></span><ArrowRight size={19} /></Link><Link href="/cards#new-card"><span className="shortcut-icon"><CreditCard size={23} /></span><span><strong>Cadastrar cartão</strong><small>Vencimento e faturas</small></span><ArrowRight size={19} /></Link><Link href="/entries#new-item"><span className="shortcut-icon"><Plus size={25} /></span><span><strong>Receita ou despesa</strong><small>Organize o seu mês</small></span><ArrowRight size={19} /></Link></section>
      <section className="overview-metrics" aria-label="Resumo financeiro do mês">{metrics.map(({ label, value, icon: Icon, tone, hint }) => <div className={`overview-metric ${tone}`} key={label}><div className="metric-heading"><span>{label}</span><Icon size={21} /></div><strong>{value === null || value === undefined ? "—" : currency.format(value / 100)}</strong><small>{hint}</small></div>)}</section>
      {summary && <MonthlyComposition key={`${summary.month}-${summary.basis}-${includeVariableForecasts}`} summary={summary} initialCategory={initialCategory} />}
      <div className="overview-lower"><section className="projection-panel"><div className="panel-heading"><div><span className="section-kicker">PLANEJAMENTO</span><h2>Projeção dos próximos meses</h2></div><div className="panel-controls"><label className="overview-month"><CalendarRange size={17} /><input aria-label="Mês da projeção" type="month" min={plan?.start_month} max={plan?.end_month} value={month} onChange={(event) => void update(event.target.value, basis)} /></label>{basis === "cash" && <VariableForecastFilter include={includeVariableForecasts} onChange={(include) => void update(month, basis, include)} />}{featureFlags.referenceBasis && <select aria-label="Base da projeção" value={basis} onChange={(event) => void update(month, event.target.value as "cash" | "reference")}><option value="cash">Caixa</option><option value="reference">Competência</option></select>}</div></div>{loading && <p className="chart-message">Atualizando projeção...</p>}{error && <p className="chart-message" role="alert">{error} <button type="button" onClick={() => void update(month, basis)}>Tentar novamente</button></p>}{!hasPlan ? <div className="chart-empty"><span><ReceiptText size={26} /></span><h3>Sua projeção começa com um planejamento</h3><p>Defina o período e adicione suas receitas e despesas para visualizar os próximos meses.</p><Link href="/planning">Criar planejamento <ArrowRight size={16} /></Link></div> : loading ? null : summaries.every((item) => item === null) ? <div className="chart-empty"><h3>Projeção indisponível</h3><p>Confira seu planejamento ou tente selecionar outro mês.</p><Link href="/planning">Ver planejamento <ArrowRight size={16} /></Link></div> : error || loading ? null : !hasChartData ? <div className="chart-empty"><span><ReceiptText size={26} /></span><h3>Adicione valores ao planejamento</h3><p>Cadastre uma receita ou despesa para ver sua projeção ganhar forma.</p><Link href="/entries#new-item">Adicionar receita ou despesa <ArrowRight size={16} /></Link></div> : <><div className="chart-legend"><span><i className="legend-income" /> Receitas</span><span><i className="legend-commitments" /> Compromissos</span><span><i className="legend-result" /> Saldo</span></div><div className="projection-chart" role="img" aria-label="Gráfico de receitas, compromissos e saldo projetado para os próximos cinco meses">{months.map((value, index) => { const item = summaries[index]; return <div className="chart-month" key={value}><div className="chart-bars"><span className="bar income" style={{ height: `${item ? Math.max(3, item.income_cents / maxValue * 100) : 0}%` }} title={`${formatMonth(value)}: receitas ${currency.format((item?.income_cents ?? 0) / 100)}`} /><span className="bar commitments" style={{ height: `${item ? Math.max(3, item.commitments_cents / maxValue * 100) : 0}%` }} title={`${formatMonth(value)}: compromissos ${currency.format((item?.commitments_cents ?? 0) / 100)}`} /><span className={`bar result ${item?.is_negative ? "negative" : ""}`} style={{ height: `${item ? Math.max(3, Math.abs(item.result_cents) / maxValue * 100) : 0}%` }} title={`${formatMonth(value)}: saldo ${currency.format((item?.result_cents ?? 0) / 100)}`} /></div><small>{formatMonth(value)} {value.slice(0, 4)}</small></div>; })}</div></>}</section>
      <section className="next-actions-panel"><span className="section-kicker">PARA SE ORGANIZAR</span><h2>Próximos passos</h2><Link href="/planning"><span className={`step-check ${hasPlan ? "done" : ""}`}>{hasPlan && <Check size={15} />}</span><span><strong>{hasPlan ? "Planejamento criado" : "Crie seu planejamento"}</strong><small>Defina o período da sua projeção</small></span><ArrowRight size={17} /></Link><Link href="/debts#new-debt"><span className={`step-check ${debtCount !== null && debtCount > 0 ? "done" : ""}`}>{debtCount !== null && debtCount > 0 && <Check size={15} />}</span><span><strong>{debtCount !== null && debtCount > 0 ? "Dívidas cadastradas" : "Cadastre suas dívidas"}</strong><small>Veja as parcelas no seu caixa</small></span><ArrowRight size={17} /></Link><Link href="/cards#new-card"><span className={`step-check ${cardCount !== null && cardCount > 0 ? "done" : ""}`}>{cardCount !== null && cardCount > 0 && <Check size={15} />}</span><span><strong>{cardCount !== null && cardCount > 0 ? "Cartões cadastrados" : "Cadastre seus cartões"}</strong><small>Acompanhe vencimentos e faturas</small></span><ArrowRight size={17} /></Link></section></div>
    </div>
  </>;
}
