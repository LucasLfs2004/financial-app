import { notFound } from "next/navigation";
import { addMonths, isMonth, inclusiveMonthCount } from "@/lib/debt-month";
import type { Metadata } from "next";
import { getCardInvoice, getCreditCards, getCurrentPlan, FinancialApiError, absentOn404 } from "@/lib/api/financial-api";
import { AppHeaderTitle } from "@/components/app-navigation";
import { InvoiceDetail } from "@/components/cards/invoice-detail";

export const metadata: Metadata = { title: "Detalhe da fatura" };

export default async function InvoiceDetailPage({ params, searchParams }: { params: Promise<{ cardId: string; month: string }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const { cardId, month } = await params;
  const query = await searchParams;
  if (!isMonth(month)) notFound();
  const returnQuery = isMonth(query.from) && isMonth(query.to) && inclusiveMonthCount(query.from, query.to) >= 1 && inclusiveMonthCount(query.from, query.to) <= 24 ? new URLSearchParams({ from: query.from, to: query.to }).toString() : "";
  const [invoice, cards, plan] = await Promise.all([
    getCardInvoice(cardId, month).catch((error: unknown) => { if (error instanceof FinancialApiError && error.status === 404) notFound(); throw error; }),
    getCreditCards(), getCurrentPlan().catch(absentOn404),
  ]);
  return <main className="resource-page-shell invoice-detail-page">
    <AppHeaderTitle title={invoice.card_name} />
    <InvoiceDetail invoice={invoice} cards={cards} min={plan?.start_month} max={plan ? addMonths(plan.end_month, 12) : undefined} returnQuery={returnQuery} adjustment={typeof query.adjustment === "string" ? query.adjustment : undefined} moveItem={typeof query.moveItem === "string" ? query.moveItem : undefined} referenceMonth={typeof query.referenceMonth === "string" ? query.referenceMonth : undefined} />
  </main>;
}
