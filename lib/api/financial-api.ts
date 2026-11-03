import { monthlySummaryPath } from "./summary-path";
import "server-only";
import { redirect } from "next/navigation";
import { getAccessToken } from "@/lib/supabase/server";
import { apiErrorDetail } from "@/lib/api/errors";
import { adaptPaymentHistory } from "@/lib/api/payment-history";
import type { ApiEnvelope, CardInvoice, CardInvoiceSummary, CreditCard, Debt, DebtReleaseList, DebtSchedule, FinancialInstitution, FinancialItem, MonthlySummary, OriginalPlanSnapshot, PaymentMethodHistory, PlannedSavings, Plan } from "@/lib/api/types";

export type Profile = {
  id: string;
  email: string;
  display_name: string | null;
  timezone: string;
  currency_code: string;
};

export class FinancialApiError extends Error {
  constructor(public status: number, message: string, public code = "api_request_failed", public details?: unknown) {
    super(message);
    this.name = "FinancialApiError";
  }
}

export async function financialApiResponse(path: string, init?: RequestInit): Promise<Response> {
  const token = await getAccessToken();

  if (!token) throw new FinancialApiError(401, "Sua sessão expirou. Entre novamente.", "unauthorized");

  const baseUrl = process.env.FINANCIAL_API_URL ?? "http://127.0.0.1:8080";
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");
  headers.set("Authorization", `Bearer ${token}`);
  return fetch(`${baseUrl}${path}`, {
    ...init,
    cache: "no-store",
    headers,
    signal: init?.signal ?? AbortSignal.timeout(4000),
  });
}

export async function financialApiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await financialApiResponse(path, init).catch((error: unknown) => { if (error instanceof FinancialApiError && error.status === 401) redirect("/sign-in"); throw error; });
  if (response.status === 401) redirect("/sign-in");
  if (!response.ok) {
    const detail = apiErrorDetail(await response.json().catch(() => null), response.status);
    throw new FinancialApiError(response.status, detail.message, detail.code, detail.details);
  }

  return response.status === 204 ? undefined as T : response.json() as Promise<T>;
}

export async function getProfile() {
  const response = await financialApiFetch<ApiEnvelope<Profile>>("/v1/me");
  return response.data;
}

export async function getCurrentPlan() {
  const response = await financialApiFetch<ApiEnvelope<Plan>>("/v1/plans/current");
  return response.data;
}

export async function getFinancialItems(filters?: { kind?: string; status?: string }) {
  const query = new URLSearchParams();
  if (filters?.kind) query.set("kind", filters.kind);
  if (filters?.status) query.set("status", filters.status);
  const response = await financialApiFetch<{ data: FinancialItem[] }>(`/v1/financial-items?${query}`);
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

export async function getMonthlySummary(month: string, basis: "cash" | "reference" = "cash", includeVariableForecasts = false) {
  const response = await financialApiFetch<ApiEnvelope<MonthlySummary>>(`/v1${monthlySummaryPath(month, basis, includeVariableForecasts)}`);
  return response.data;
}

export async function getInstitutions() { return (await financialApiFetch<{ data: FinancialInstitution[] }>("/v1/financial-institutions")).data; }
export async function getCreditCards() { return (await financialApiFetch<{ data: CreditCard[] }>("/v1/credit-cards")).data; }
export async function getCardInvoices(cardId: string, from: string, to: string) { return (await financialApiFetch<{ data: CardInvoiceSummary[]; range: { from: string; to: string } }>(`/v1/credit-cards/${encodeURIComponent(cardId)}/invoices?from=${from}&to=${to}&include_empty=true`)).data; }
export async function getCardInvoice(cardId: string, month: string) { return (await financialApiFetch<ApiEnvelope<CardInvoice>>(`/v1/credit-cards/${encodeURIComponent(cardId)}/invoices/${month}`)).data; }

export async function getDebts(asOf?: string, filters?: { status?: string; projection_status?: string }) {
  const query = new URLSearchParams();
  if (asOf) query.set("as_of", asOf);
  if (filters?.status) query.set("status", filters.status);
  if (filters?.projection_status) query.set("projection_status", filters.projection_status);
  const response = await financialApiFetch<{ data: Debt[] }>(`/v1/debts?${query}`);
  return response.data;
}
export async function getDebt(id: string, asOf?: string) { return (await financialApiFetch<ApiEnvelope<Debt>>(`/v1/debts/${encodeURIComponent(id)}${asOf ? `?as_of=${encodeURIComponent(asOf)}` : ""}`)).data; }
export async function getDebtSchedule(id: string, from: string, to: string) { return (await financialApiFetch<ApiEnvelope<DebtSchedule>>(`/v1/debts/${encodeURIComponent(id)}/schedule?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`)).data; }
export async function getDebtReleases(from: string, to: string) { return await financialApiFetch<{ data: DebtReleaseList; range: { from: string; to: string } }>(`/v1/debt-releases?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`); }
export async function getPaymentMethodHistory(id: string): Promise<PaymentMethodHistory> {
  const response = await financialApiFetch<ApiEnvelope<PaymentMethodHistory["periods"]>>(`/v1/financial-items/${encodeURIComponent(id)}/payment-history`);
  return adaptPaymentHistory(id, response.data);
}

export function absentOn404(error: unknown): null {
  if (error instanceof FinancialApiError && error.status === 404) return null;
  throw error;
}

export async function getFinancialItem(id: string) { return (await financialApiFetch<ApiEnvelope<FinancialItem>>(`/v1/financial-items/${encodeURIComponent(id)}`)).data; }
export async function getCreditCard(id: string) { return (await financialApiFetch<ApiEnvelope<CreditCard>>(`/v1/credit-cards/${encodeURIComponent(id)}`)).data; }
