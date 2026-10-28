import { variableForecastPreference } from "@/lib/api/summary-path";
import type { Metadata } from "next";
import { getCreditCards, getCurrentPlan, getDebts, getMonthlySummary, getProfile, absentOn404 } from "@/lib/api/financial-api";
import { getAuthenticatedClaims } from "@/lib/supabase/server";
import { DashboardOverview } from "@/components/dashboard/dashboard-overview";
import { currentMonth, clampMonth, isMonth, projectionMonths } from "@/lib/debt-month";
import { featureFlags } from "@/lib/feature-flags";
import type { MonthlySummary } from "@/lib/api/types";

export const metadata: Metadata = { title: "Visão geral" };

export default async function DashboardPage({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const query = await searchParams;
  const includeVariableForecasts = variableForecastPreference(query.include_variable_forecasts);
  const basis = featureFlags.referenceBasis && query.basis === "reference" ? "reference" : "cash";
  const claimsPromise = getAuthenticatedClaims();
  const [profileData, plan] = await Promise.all([getProfile(), getCurrentPlan().catch(absentOn404)]);
  const requestedMonth = isMonth(query.month) ? query.month : currentMonth(profileData.timezone);
  const initialMonth = plan ? clampMonth(requestedMonth, plan.start_month, plan.end_month) : requestedMonth;
  const profilePromise = Promise.resolve(profileData);
  const cardsPromise = getCreditCards();
  const debtsPromise = getDebts(initialMonth);
  const summariesPromise: Promise<(MonthlySummary | null)[]> = plan
    ? Promise.all(projectionMonths(initialMonth, plan).map((month) => getMonthlySummary(month, basis, includeVariableForecasts)))
    : Promise.resolve([]);
  const [profile, cards, debts, summaries, { data }] = await Promise.all([profilePromise, cardsPromise, debtsPromise, summariesPromise, claimsPromise]);
  const displayName = profile?.display_name || String(data?.claims?.user_metadata?.name ?? "").trim() || String(data?.claims?.email ?? "").split("@")[0] || "por aqui";

  return <main className="dashboard-page"><DashboardOverview key={`${initialMonth}-${basis}-${includeVariableForecasts}`} name={displayName.split(" ")[0]} initialMonth={initialMonth} initialSummaries={summaries} initialBasis={basis} initialIncludeVariableForecasts={includeVariableForecasts} initialCategory={query.composition === "income" || query.composition === "commitments" || query.composition === "savings" ? query.composition : undefined} hasPlan={Boolean(plan)} plan={plan} cardCount={cards?.filter((card) => card.status === "active").length ?? null} debtCount={debts?.filter((debt) => debt.status === "active").length ?? null} apiAvailable={Boolean(profile)} /></main>;
}
