export type SnapshotSection = { title: string; data: unknown };
function object(value: unknown): Record<string, unknown> { return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {}; }

// Read the frozen document only. Never synthesize absent collections from live DTOs.
export function snapshotSections(version: number, document: unknown): { known: boolean; sections: SnapshotSection[] } {
  if (![1, 2, 3, 4].includes(version)) return { known: false, sections: [{ title: "Documento original", data: document }] };
  const plan = object(document);
  const sections: SnapshotSection[] = [{ title: "Plano", data: Object.fromEntries(Object.entries(plan).filter(([key]) => !["items", "savings", "release_2", "release_3"].includes(key))) }];
  if ("items" in plan) sections.push({ title: "Registros e valores", data: plan.items });
  if ("savings" in plan) sections.push({ title: "Economia planejada", data: plan.savings });
  const release2 = object(plan.release_2);
  for (const [key, title] of Object.entries({ financial_institutions: "Instituições", credit_cards: "Cartões e configurações", payment_methods: "Métodos de pagamento", invoice_adjustments: "Ajustes de faturas", invoice_moves: "Movimentações de faturas" })) {
    if (key in release2) sections.push({ title, data: release2[key] });
  }
  const release3 = object(plan.release_3);
  if ("debts" in release3) sections.push({ title: "Dívidas e quitações", data: release3.debts });
  return { known: true, sections };
}

export function historicMonth(value: unknown): string | null {
  if (typeof value !== "string") return null;
  const match = /^(\d{4}-(?:0[1-9]|1[0-2]))(?:-01)?$/.exec(value);
  return match?.[1] ?? null;
}
