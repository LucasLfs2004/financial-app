import type { Metadata } from "next";
import { getCreditCards, getCurrentPlan, getDebts, getMonthlySummary, getProfile } from "@/lib/api/financial-api";
import { createClient } from "@/lib/supabase/server";
import { DashboardOverview } from "@/components/dashboard/dashboard-overview";
import type { MonthlySummary } from "@/lib/api/types";

export const metadata: Metadata = { title: "Visão geral" };

function monthOffset(month: string, offset: number) {
  const [year, number] = month.split("-").map(Number);
  const date = new Date(Date.UTC(year, number - 1 + offset, 1));
  return date.toISOString().slice(0, 7);
}

export default async function DashboardPage() {
  const supabase = await createClient();
  const { data: { user } } = await supabase.auth.getUser();
  const dateParts = new Intl.DateTimeFormat("en-US", { timeZone: "America/Sao_Paulo", year: "numeric", month: "2-digit" }).formatToParts(new Date());
  const currentMonth = `${dateParts.find((part) => part.type === "year")?.value}-${dateParts.find((part) => part.type === "month")?.value}`;
  const [profile, plan, cards, debts] = await Promise.all([
    getProfile().catch(() => null),
    getCurrentPlan().catch(() => null),
    getCreditCards().catch(() => null),
    getDebts().catch(() => null),
  ]);
  const months = Array.from({ length: 5 }, (_, index) => monthOffset(currentMonth, index));
  const summaries: (MonthlySummary | null)[] = plan
    ? await Promise.all(months.map((month) => getMonthlySummary(month).catch(() => null)))
    : months.map(() => null);
  const displayName = profile?.display_name || String(user?.user_metadata?.name ?? "").trim() || user?.email?.split("@")[0] || "por aqui";

  return <main className="dashboard-page"><DashboardOverview name={displayName.split(" ")[0]} initialMonth={currentMonth} initialSummaries={summaries} hasPlan={Boolean(plan)} cardCount={cards?.filter((card) => card.status === "active").length ?? null} debtCount={debts?.filter((debt) => debt.status === "active").length ?? null} apiAvailable={Boolean(profile)} /></main>;
}
