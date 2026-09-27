import "server-only";
import { createClient } from "@/lib/supabase/server";
import type { ApiEnvelope, CardInvoice, CardInvoiceSummary, CreditCard, Debt, FinancialInstitution, FinancialItem, MonthlySummary, OriginalPlanSnapshot, PlannedSavings, Plan } from "@/lib/api/types";

export type Profile = {
  id: string;
  email: string;
  display_name: string | null;
  timezone: string;
  currency_code: string;
};

export class FinancialApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "FinancialApiError";
  }
}

export async function financialApiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const supabase = await createClient();
  const { data } = await supabase.auth.getSession();
  const token = data.session?.access_token;

  if (!token) throw new FinancialApiError(401, "Authentication is required");

  const baseUrl = process.env.FINANCIAL_API_URL ?? "http://127.0.0.1:8080";
  const response = await fetch(`${baseUrl}${path}`, {
    ...init,
    cache: "no-store",
    headers: {
      Accept: "application/json",
      ...init?.headers,
      Authorization: `Bearer ${token}`,
    },
    signal: init?.signal ?? AbortSignal.timeout(4000),
  });

  if (!response.ok) {
    throw new FinancialApiError(response.status, `Financial API returned ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export async function getProfile() {
  const response = await financialApiFetch<ApiEnvelope<Profile>>("/v1/me");
  return response.data;
}

export async function getCurrentPlan() {
  const response = await financialApiFetch<ApiEnvelope<Plan>>("/v1/plans/current");
  return response.data;
}

export async function getFinancialItems() {
  const response = await financialApiFetch<{ data: FinancialItem[] }>("/v1/financial-items");
  return response.data;
}

export async function getPlannedSavings() {
  const response = await financialApiFetch<ApiEnvelope<PlannedSavings>>("/v1/plans/current/savings");
  return response.data;
}

export async function getOriginalPlan() {
  const response = await financialApiFetch<ApiEnvelope<OriginalPlanSnapshot>>("/v1/plans/current/original");
  return response.data;
}

export async function getMonthlySummary(month: string, basis: "cash" | "reference" = "cash") {
  const response = await financialApiFetch<ApiEnvelope<MonthlySummary>>(`/v1/plans/current/months/${month}/summary?basis=${basis}`);
  return response.data;
}

export async function getInstitutions() { return (await financialApiFetch<{ data: FinancialInstitution[] }>("/v1/financial-institutions")).data; }
export async function getCreditCards() { return (await financialApiFetch<{ data: CreditCard[] }>("/v1/credit-cards")).data; }
export async function getCardInvoices(cardId: string, from: string, to: string) { return (await financialApiFetch<{ data: CardInvoiceSummary[]; range: { from: string; to: string } }>(`/v1/credit-cards/${cardId}/invoices?from=${from}&to=${to}&include_empty=true`)).data; }
export async function getCardInvoice(cardId: string, month: string) { return (await financialApiFetch<ApiEnvelope<CardInvoice>>(`/v1/credit-cards/${cardId}/invoices/${month}`)).data; }

export async function getDebts() {
  const response = await financialApiFetch<{ data: Debt[] }>("/v1/debts");
  return response.data;
}
