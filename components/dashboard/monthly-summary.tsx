"use client";

import { useState } from "react";
import { Alert, Badge, Spinner } from "@lucaslfs2004/luke-ui";
import { ArrowDownLeft, ArrowUpRight, CalendarRange, PiggyBank, WalletCards } from "lucide-react";
import { browserApiFetch } from "@/lib/api/browser-api";
import Link from "next/link";
import type { MonthlySummary } from "@/lib/api/types";
import { featureFlags } from "@/lib/feature-flags";

const money = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const monthLabel = new Intl.DateTimeFormat("pt-BR", { month: "long", year: "numeric" });

export function MonthlySummaryPanel({ initialMonth, initialSummary }: { initialMonth: string; initialSummary: MonthlySummary | null }) {
  const [month, setMonth] = useState(initialMonth); const [basis, setBasis] = useState<"cash" | "reference">(initialSummary?.basis ?? "cash"); const [summary, setSummary] = useState(initialSummary); const [loading, setLoading] = useState(false); const [error, setError] = useState("");
  async function refresh(nextMonth: string, nextBasis: "cash" | "reference") {
    setLoading(true); setError("");
    try { const response = await browserApiFetch<{ data: MonthlySummary }>(`/plans/current/months/${nextMonth}/summary?basis=${nextBasis}`); setSummary(response.data); }
    catch (caughtError) { setSummary(null); setError(caughtError instanceof Error ? caughtError.message : "Não foi possível carregar o resumo."); }
    finally { setLoading(false); }
  }
  function changeMonth(value: string) { setMonth(value); void refresh(value, basis); }
  function changeBasis(value: "cash" | "reference") { setBasis(value); void refresh(month, value); }
  const result = summary?.result_cents ?? 0;
  const debtSources = summary?.sources.filter((source) => source.debt_id) ?? [];
  return <section className="summary-section"><div className="summary-toolbar"><div><span className="section-kicker">RESUMO MENSAL</span><h2>{summary ? monthLabel.format(new Date(`${summary.month}-02T12:00:00`)) : "Sua projeção"}</h2></div><div className="summary-filters"><label aria-label="Mês"><CalendarRange size={15} /><input type="month" value={month} onChange={(event) => changeMonth(event.target.value)} /></label>{featureFlags.referenceBasis && <select value={basis} onChange={(event) => changeBasis(event.target.value as "cash" | "reference")}><option value="cash">Caixa</option><option value="reference">Competência</option></select>}</div></div>{loading && <div className="summary-loading"><Spinner size="sm" /> Atualizando resumo...</div>}{error && <Alert color="error" title="Resumo indisponível">{error}</Alert>}{summary && <><div className={`summary-result ${summary.is_negative ? "negative-result" : "positive-result"}`}><div><span>{summary.basis === "cash" ? "Dinheiro livre planejado" : "Resultado por competência"}</span><strong>{money.format(result / 100)}</strong></div><Badge color={summary.is_negative ? "destructive" : "forest"} variant="soft">{summary.is_negative ? "Negativo" : "Projetado"}</Badge></div><div className="summary-breakdown"><div><ArrowDownLeft size={16} /><span>Entradas<strong>{money.format(summary.income_cents / 100)}</strong></span></div><div><ArrowUpRight size={16} /><span>Compromissos<strong>{money.format(summary.commitments_cents / 100)}</strong></span></div><div><PiggyBank size={16} /><span>Para guardar<strong>{money.format(summary.planned_savings_cents / 100)}</strong></span></div><div><WalletCards size={16} /><span>Base<strong>{summary.basis === "cash" ? "Caixa" : "Competência"}</strong></span></div></div>{summary.breakdown.debt_installments_cents > 0 && <div className="summary-debt-line"><span>Parcelas de dívidas incluídas nos compromissos</span><strong>{money.format(summary.breakdown.debt_installments_cents / 100)}</strong></div>}{debtSources.length > 0 && <div className="summary-debt-sources">{debtSources.map((source) => <Link key={source.source_id} href={`/debts/${source.debt_id}`}><span>{source.name}{source.installment_number ? ` · ${source.installment_number}/${source.installments_total}` : ""}{source.debt_occurrence_kind === "early_settlement" ? " · quitação projetada" : ""}</span><strong>{money.format(source.amount_cents / 100)}</strong></Link>)}</div>}</>}</section>;
}
