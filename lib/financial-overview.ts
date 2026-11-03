import type { CardInvoiceComponent, MonthlySummarySource } from "./api/types";

// Presentation groups only. Official balances always come from the API.
export function commitmentGroups(sources: MonthlySummarySource[]) {
  const groups = [
    { id: "cards", label: "Faturas de cartões", sources: [] as MonthlySummarySource[] },
    { id: "other", label: "Outros compromissos", sources: [] as MonthlySummarySource[] },
  ];
  for (const source of sources) {
    if (source.effect !== "commitment") continue;
    const card = source.payment_method === "credit_card" || Boolean(source.credit_card_id);
    groups[card ? 0 : 1].sources.push(source);
  }
  return groups.map((group) => ({ ...group, amount: group.sources.reduce((sum, source) => sum + source.amount_cents, 0) }));
}

export function invoiceComponentLabel(component: CardInvoiceComponent) {
  if (component.source_type === "invoice_payment") return "Pagamento registrado";
  if (component.debt_occurrence_kind === "early_settlement") return "Quitação prevista";
  if (component.is_projected) return "Parcela prevista";
  if (component.source_type === "financial_item_occurrence") return "Compromisso previsto";
  if (component.source_type === "invoice_adjustment" && component.amount_cents < 0) return "Crédito registrado";
  if (component.source_type === "invoice_adjustment") return "Cobrança registrada";
  return "Lançamento";
}

export function invoiceComposition(components: CardInvoiceComponent[]) {
  const groups = [
    { id: "planned", label: "Compromissos previstos", amount: 0, count: 0 },
    { id: "charges", label: "Cobranças registradas", amount: 0, count: 0 },
    { id: "payments", label: "Pagamentos e créditos", amount: 0, count: 0 },
    { id: "other", label: "Outros lançamentos", amount: 0, count: 0 },
  ];
  for (const component of components) {
    // Preserve negative credits and payments; never turn them into expenses.
    const index = component.source_type === "invoice_payment" || component.amount_cents < 0 ? 2
      : component.is_projected || component.source_type === "financial_item_occurrence" ? 0
      : component.source_type === "invoice_adjustment" ? 1 : 3;
    groups[index].amount += component.amount_cents;
    groups[index].count++;
  }
  return groups.filter((group) => group.count > 0);
}
