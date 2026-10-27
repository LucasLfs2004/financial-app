import type { InvoiceImportPreview } from "./api/types";

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
  const rowCharges = rows.reduce((total, row) => total + (row.kind === "payment" ? 0 : row.amount_cents), 0);
  const rowPayments = rows.reduce((total, row) => total + (row.kind === "payment" ? row.amount_cents : 0), 0);
  const net = Number.isFinite(data.new_total_cents) ? data.new_total_cents : rowCharges + rowPayments;
  const charges = Number.isFinite(data.new_charges_cents)
    ? data.new_charges_cents!
    : Number.isFinite(data.new_payments_cents) ? net + data.new_payments_cents! : rowCharges;
  const payments = Number.isFinite(data.new_payments_cents)
    ? data.new_payments_cents!
    : Math.abs(rowPayments);
  return { charges, payments, net };
}
