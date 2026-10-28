import type { InvoiceAdjustmentDetail } from "./types";
import { changedFields, centsField, optionalText, textField } from "./requests";

export function adjustmentPatch(detail: InvoiceAdjustmentDetail, form: FormData) {
  if (!detail.capabilities.can_edit || detail.status !== "active" || detail.entry_kind === "payment") throw new Error("Este lançamento não pode ser editado.");
  const values: Record<string, unknown> = {};
  for (const key of detail.capabilities.editable_fields) {
    // Missing metadata is not a user request to clear it.
    if (!Object.hasOwn(detail, key) || !form.has(key === "amount_cents" ? "amount" : key)) continue;
    values[key] = key === "amount_cents" ? centsField(form, "amount") : key === "name" ? textField(form, key) : optionalText(form, key);
  }
  return changedFields(detail, values);
}

export function adjustmentPath(detail: Pick<InvoiceAdjustmentDetail, "id" | "credit_card_id" | "payment_month">) {
  return `/credit-cards/${encodeURIComponent(detail.credit_card_id)}/invoices/${encodeURIComponent(detail.payment_month)}/adjustments/${encodeURIComponent(detail.id)}`;
}
