import { variableForecastPreference } from "@/lib/api/summary-path";
import type { Metadata } from "next";
import { Camera, ChartNoAxesCombined } from "lucide-react";
import { Card, CardContent } from "@lucaslfs2004/luke-ui";
import Link from "next/link";
import { PlanForm } from "@/components/planning/plan-form";
import { SavingsForm } from "@/components/planning/savings-form";
import { PlanningProjection } from "@/components/planning/planning-projection";
import { getCurrentPlan, getMonthlySummary, getOriginalPlan, getPlannedSavings, getProfile, absentOn404 } from "@/lib/api/financial-api";
import { currentMonth, clampMonth, isMonth, projectionMonths } from "@/lib/debt-month";
import { featureFlags } from "@/lib/feature-flags";
import type { MonthlySummary } from "@/lib/api/types";

export const metadata: Metadata = { title: "Planejamento" };

export default async function PlanningPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const includeVariableForecasts = variableForecastPreference(query.include_variable_forecasts);
  const basis = featureFlags.referenceBasis && query.basis === "reference" ? "reference" : "cash";
  const [plan, profile] = await Promise.all([getCurrentPlan().catch(absentOn404), getProfile()]);
  const today = currentMonth(profile.timezone);
  const requestedMonth = isMonth(query.from) ? query.from : today;
  const initialMonth = plan ? clampMonth(requestedMonth, plan.start_month, plan.end_month) : requestedMonth;
  const summariesPromise: Promise<(MonthlySummary | null)[]> = plan
    ? Promise.all(projectionMonths(initialMonth, plan).map((month) => getMonthlySummary(month, basis, includeVariableForecasts)))
    : Promise.resolve([]);
  const [savings, original, summaries] = await Promise.all([plan ? getPlannedSavings().catch(absentOn404) : null, plan ? getOriginalPlan().catch(absentOn404) : null, summariesPromise]);

  return <main className="resource-page-shell planning-page">
    <header className="resource-page-header"><p>Acompanhe a projeção e ajuste seu planejamento.</p></header>
    {requestedMonth !== initialMonth && <p role="status">Mês ajustado ao horizonte do planejamento.</p>}
    {plan ? <PlanningProjection key={`${initialMonth}|${basis}|${includeVariableForecasts}|${summaries.map((item) => `${item?.income_cents}:${item?.commitments_cents}:${item?.planned_savings_cents}:${item?.result_cents}`).join("|")}`} initialMonth={initialMonth} initialSummaries={summaries} initialBasis={basis} initialIncludeVariableForecasts={includeVariableForecasts} plan={plan} /> : <div className="planning-intro"><ChartNoAxesCombined size={24} /><div><strong>Seu gráfico começa aqui</strong><p>Defina o período do planejamento para acompanhar a projeção dos próximos meses.</p></div></div>}
    <div className="planning-settings-heading"><div><span className="section-kicker">CONFIGURAÇÃO</span><h2>Parâmetros do planejamento</h2></div><Link href="/entries">Ver receitas e despesas</Link></div>
    <div className="planning-settings-grid"><PlanForm plan={plan} defaultMonth={today} currencyCode={profile.currency_code} />{plan && <SavingsForm savings={savings} plan={plan} month={initialMonth} />}</div>
    {original && <Card className="snapshot-card planning-snapshot" padding="none"><CardContent><Camera size={20} /><div><strong>Referência original criada</strong><span>Capturada em {new Date(original.captured_at).toLocaleDateString("pt-BR")}. Ela permanece imutável. <Link href="/planning/original">Consultar referência</Link></span></div></CardContent></Card>}
  </main>;
}
