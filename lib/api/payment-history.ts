import type { PaymentMethodHistory, PaymentMethodPeriod } from "./types";

export function adaptPaymentHistory(itemId: string, periods: PaymentMethodPeriod[]): PaymentMethodHistory {
  if (!Array.isArray(periods)) throw new Error("Resposta inválida do histórico de pagamentos.");
  return { financial_item_id: itemId, default_method: "direct", periods };
}
