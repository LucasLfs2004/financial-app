import type { Metadata } from "next";
import { Camera, ChartNoAxesCombined } from "lucide-react";
import { Card, CardContent } from "@lucaslfs2004/luke-ui";
import Link from "next/link";
import { PlanForm } from "@/components/planning/plan-form";
import { SavingsForm } from "@/components/planning/savings-form";
import { PlanningProjection } from "@/components/planning/planning-projection";
import { getCurrentPlan, getMonthlySummary, getOriginalPlan, getPlannedSavings } from "@/lib/api/financial-api";
import type { MonthlySummary } from "@/lib/api/types";

export const metadata: Metadata = { title: "Planejamento" };

function monthOffset(month: string, offset: number) {
  const [year, number] = month.split("-").map(Number);
  return new Date(Date.UTC(year, number - 1 + offset, 1)).toISOString().slice(0, 7);
}

export default async function PlanningPage() {
  const planPromise = getCurrentPlan().catch(() => null);
  const savingsPromise = getPlannedSavings().catch(() => null);
  const originalPromise = getOriginalPlan().catch(() => null);
  const dateParts = new Intl.DateTimeFormat("en-US", { timeZone: "America/Sao_Paulo", year: "numeric", month: "2-digit" }).formatToParts(new Date());
  const currentMonth = `${dateParts.find((part) => part.type === "year")?.value}-${dateParts.find((part) => part.type === "month")?.value}`;
  const plan = await planPromise;
  const initialMonth = plan ? currentMonth < plan.start_month ? plan.start_month : currentMonth > plan.end_month ? plan.end_month : currentMonth : currentMonth;
  const summariesPromise: Promise<(MonthlySummary | null)[]> = plan
    ? Promise.all(Array.from({ length: 5 }, (_, index) => getMonthlySummary(monthOffset(initialMonth, index)).catch(() => null)))
    : Promise.resolve([]);
  const [savings, original, summaries] = await Promise.all([savingsPromise, originalPromise, summariesPromise]);

  return <main className="resource-page-shell planning-page">
    <header className="resource-page-header"><p>Acompanhe a projeção e ajuste seu planejamento.</p></header>
    {plan ? <PlanningProjection key={`${initialMonth}|${summaries.map((item) => `${item?.income_cents}:${item?.commitments_cents}:${item?.planned_savings_cents}:${item?.result_cents}`).join("|")}`} initialMonth={initialMonth} initialSummaries={summaries} /> : <div className="planning-intro"><ChartNoAxesCombined size={24} /><div><strong>Seu gráfico começa aqui</strong><p>Defina o período do planejamento para acompanhar a projeção dos próximos meses.</p></div></div>}
    <div className="planning-settings-heading"><div><span className="section-kicker">CONFIGURAÇÃO</span><h2>Parâmetros do planejamento</h2></div><Link href="/entries">Ver receitas e despesas</Link></div>
    <div className="planning-settings-grid"><PlanForm plan={plan} />{plan && <SavingsForm savings={savings} />}</div>
    {original && <Card className="snapshot-card planning-snapshot" padding="none"><CardContent><Camera size={20} /><div><strong>Referência original criada</strong><span>Capturada em {new Date(original.captured_at).toLocaleDateString("pt-BR")}. Ela permanece imutável.</span></div></CardContent></Card>}
  </main>;
}
