"use client";

import { useState } from "react";
import { browserApiFetch } from "@/lib/api/browser-api";
import type { ApiEnvelope, FinancialItem } from "@/lib/api/types";

type CardPayment = { effective_from: string; end_month: string | null; method: "credit_card"; credit_card_id: string };
type PendingItem = { id: string; name: string; payment: CardPayment };

// Retain the created item when its second request fails, so retrying only links
// the card and cannot create the same subscription twice.
export function useFinancialItemCreation() {
  const [pendingItem, setPendingItem] = useState<PendingItem | null>(null);

  async function saveItem(payload: Record<string, unknown>, payment: CardPayment | null) {
    let pending = pendingItem;
    if (!pending) {
      const response = await browserApiFetch<ApiEnvelope<FinancialItem>>("/financial-items", { method: "POST", body: JSON.stringify(payload), successMessage: false });
      if (!payment) return response.data.name;
      pending = { id: response.data.id, name: response.data.name, payment };
      setPendingItem(pending);
    }
    await browserApiFetch(`/financial-items/${pending.id}/payment-changes`, { method: "POST", body: JSON.stringify(pending.payment), successMessage: false });
    setPendingItem(null);
    return pending.name;
  }

  return { saveItem, pendingItem };
}
