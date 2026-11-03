import type { MonthlySummarySource } from "./api/types";

export function summarySourceLink(source: MonthlySummarySource): string | null {
  if (source.debt_id) return `/debts/${encodeURIComponent(source.debt_id)}${source.reference_month ? `?as_of=${encodeURIComponent(source.reference_month)}` : ""}`;
  if (source.credit_card_id && source.invoice_payment_month) return `/invoices/${encodeURIComponent(source.credit_card_id)}/${encodeURIComponent(source.invoice_payment_month)}`;
  if (source.item_id) return `/entries/${encodeURIComponent(source.item_id)}${source.reference_month ? `?month=${encodeURIComponent(source.reference_month)}` : ""}`;
  if (source.effect === "savings") return "/planning";
  return null;
}
