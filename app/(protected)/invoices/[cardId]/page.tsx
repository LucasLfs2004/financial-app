import { redirect, notFound } from "next/navigation";
import { addMonths, currentMonth } from "@/lib/debt-month";
import { queryWindow } from "@/lib/query-window";
import { getCreditCards, getCurrentPlan, getProfile, absentOn404 } from "@/lib/api/financial-api";
import Link from "next/link";

export default async function InvoicesPage({ params, searchParams }: { params: Promise<{ cardId: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { cardId } = await params;
  const query = await searchParams;
  const [cards, plan, profile] = await Promise.all([getCreditCards(), getCurrentPlan().catch(absentOn404), getProfile()]);
  if (!cards.some((card) => card.id === cardId)) notFound();
  if (!plan) return <main className="resource-page-shell"><p>Crie um planejamento para consultar as faturas.</p><Link href="/planning">Criar planejamento</Link></main>;
  const { from, to } = queryWindow(query, currentMonth(profile.timezone), { from: plan.start_month, to: addMonths(plan.end_month, 12) }, 24);
  redirect(`/invoices/${encodeURIComponent(cardId)}/${from}?${new URLSearchParams({ from, to })}`);
}
