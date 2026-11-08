import type { InvoiceImportPreview } from "./api/types";

export function decisionRows(data: InvoiceImportPreview) {
  return data.rows.filter((row) => row.kind === "payment" || row.status === "needs_subscription_match" || (row.duplicate_candidates?.length ?? 0) > 0);
}

export function duplicateCount(data: InvoiceImportPreview) {
  return data.needs_duplicate_decision_count ?? data.rows.filter((row) => row.status === "needs_duplicate_decision").length;
}

export function csvTotals(data: InvoiceImportPreview) {
  return data.rows.reduce((totals, row) => {
    if (row.kind === "payment") totals.payments += Math.abs(row.amount_cents);
    else if (row.amount_cents < 0) totals.credits += Math.abs(row.amount_cents);
    else totals.charges += row.amount_cents;
    return totals;
  }, { charges: 0, credits: 0, payments: 0 });
}

export function unsupportedCredits(data: InvoiceImportPreview) {
  return data.rows.filter((row) => row.amount_cents < 0 && row.kind !== "payment" && row.kind !== "credit").length;
}

export function invoiceCount(data: InvoiceImportPreview) {
  return Number.isFinite(data.needs_invoice_count)
    ? data.needs_invoice_count!
    : data.rows.filter((row) => row.status === "needs_invoice").length;
}

export function subscriptionMatchCount(data: InvoiceImportPreview) {
  return data.needs_subscription_match_count ?? data.rows.filter((row) => row.status === "needs_subscription_match").length;
}

export function subscriptionMatches(data: InvoiceImportPreview) {
  return data.rows.filter((row) => row.matched_item_id && row.planned_amount_cents != null);
}

export function ignoredPayments(data: InvoiceImportPreview) {
  return data.rows.filter((row) => row.status === "skipped" && row.reason === "non_expense" && /^pagamento recebido/i.test(row.title.trim())).length;
}

export function newProjectionCount(data: InvoiceImportPreview) {
  return data.new_projected_count ?? data.projected_installments?.filter((row) => row.status === "new" || row.status === "imported").length ?? 0;
}

export function hasChanges(data: InvoiceImportPreview) {
  return data.new_count > 0 || newProjectionCount(data) > 0 || subscriptionMatches(data).length > 0;
}

export function importTotals(data: InvoiceImportPreview, confirmed: boolean) {
  const rows = data.rows.filter((row) => row.status === (confirmed ? "imported" : "new"));
  const rowCredits = rows.reduce((total, row) => total + (row.kind === "credit" ? Math.abs(row.amount_cents) : 0), 0);
  const rowCharges = rows.reduce((total, row) => total + (row.kind === "payment" || row.kind === "credit" ? 0 : row.amount_cents), 0);
  const rowPayments = rows.reduce((total, row) => total + (row.kind === "payment" ? row.amount_cents : 0), 0);
  const credits = data.new_credits_cents ?? rowCredits;
  const net = Number.isFinite(data.new_total_cents) ? data.new_total_cents : rowCharges - credits + rowPayments;
  const charges = Number.isFinite(data.new_charges_cents)
    ? data.new_charges_cents!
    : Number.isFinite(data.new_payments_cents) ? net + data.new_payments_cents! + credits : rowCharges;
  const payments = Number.isFinite(data.new_payments_cents)
    ? data.new_payments_cents!
    : Math.abs(rowPayments);
  return { charges, credits, payments, net };
}
