import { test } from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import { load } from "./test-loader.mjs";

const form = (values) => { const data = new FormData(); for (const [key, value] of Object.entries(values)) data.set(key, String(value)); return data; };

test("payment changes preserve nullable fields, require card only for credit and reject inverted periods", () => {
  const { paymentChange } = load("lib/api/requests.ts");
  assert.deepEqual(paymentChange(form({ effective_from: "2028-01", method: "direct", credit_card_id: "stale-card" })), { effective_from: "2028-01", end_month: null, context: null, method: "direct", credit_card_id: null });
  assert.throws(() => paymentChange(form({ effective_from: "2028-01", method: "credit_card" })));
  assert.throws(() => paymentChange(form({ effective_from: "2028-01", end_month: "2027-12", method: "direct" })));
  assert.equal(paymentChange(form({ effective_from: "2040-01", method: "credit_card", credit_card_id: "card" })).credit_card_id, "card");
});

test("metadata patch omits unchanged fields and preserves explicit clears/zero", () => {
  const { changedFields, centsField } = load("lib/api/requests.ts");
  assert.deepEqual(changedFields({ name: "Assinatura", description: "Antiga" }, { name: "Assinatura", description: null }), { description: null });
  assert.equal(centsField(form({ amount: "0" }), "amount"), 0);
  assert.equal(centsField(form({ amount: "115,01" }), "amount"), 11501);
  assert.throws(() => centsField(form({ amount: "invalid" }), "amount"));
});

test("proposed adjustment editor uses persisted location/capabilities and never clears missing metadata", () => {
  const { adjustmentPatch, adjustmentPath } = load("lib/api/adjustment.ts");
  const detail = { id: "adjustment/id", credit_card_id: "persisted-card", payment_month: "2026-11", name: "Original", amount_cents: 10000, status: "active", entry_kind: "charge", reference_month: "2026-10", context: "Preservar", capabilities: { can_edit: true, editable_fields: ["amount_cents", "context"] } };
  assert.deepEqual(adjustmentPatch(detail, form({ amount: "115", context: "Preservar", name: "forbidden", reference_month: "2026-12" })), { amount_cents: 11500 });
  assert.deepEqual(adjustmentPatch(detail, form({ context: "" })), { context: null });
  const missing = { ...detail }; delete missing.context;
  assert.deepEqual(adjustmentPatch(missing, form({ context: "" })), {});
  assert.equal(adjustmentPath(detail), "/credit-cards/persisted-card/invoices/2026-11/adjustments/adjustment%2Fid");
  assert.throws(() => adjustmentPatch({ ...detail, entry_kind: "payment" }, form({ amount: "1" })));
  assert.throws(() => adjustmentPatch({ ...detail, status: "archived" }, form({ amount: "1" })));
});

test("snapshot versions preserve original document, historical dates and absent collections", () => {
  const { snapshotSections, historicMonth } = load("lib/snapshot.ts");
  for (const version of [1, 2, 3, 4]) {
    const snapshot = JSON.parse(fs.readFileSync(new URL(`./fixtures/snapshot-v${version}.json`, import.meta.url), "utf8"));
    const original = JSON.stringify(snapshot);
    const view = snapshotSections(snapshot.schema_version, snapshot.plan);
    assert.equal(view.known, true);
    assert.equal(view.sections.some((section) => section.title === "Cartões e configurações"), version >= 2);
    assert.equal(view.sections.some((section) => section.title === "Dívidas e quitações"), version === 4);
    assert.equal(snapshot.plan.savings.periods[0].amount_cents, 0);
    assert.equal(JSON.stringify(snapshot), original);
  }
  const unknown = JSON.parse(fs.readFileSync(new URL("./fixtures/snapshot-unknown.json", import.meta.url), "utf8"));
  assert.deepEqual(snapshotSections(99, unknown.plan), { known: false, sections: [{ title: "Documento original", data: unknown.plan }] });
  assert.equal(historicMonth("2026-10-01"), "2026-10");
  assert.equal(historicMonth("2026-10"), "2026-10");
  assert.equal(historicMonth("2026-10-28"), null);
});

test("CSV payment magnitudes/net and backfill-only preview remain distinct", () => {
  const { importTotals, hasChanges, subscriptionMatchCount, ignoredPayments } = load("lib/invoice-import.ts");
  const data = { new_count: 2, new_total_cents: 6000, new_charges_cents: 11000, new_payments_cents: 5000, rows: [{ kind: "expense", amount_cents: 11000, status: "new" }, { kind: "payment", amount_cents: -5000, status: "new" }] };
  assert.deepEqual(importTotals(data, false), { charges: 11000, payments: 5000, net: 6000 });
  assert.equal(importTotals({ ...data, new_charges_cents: undefined }, false).charges, 11000);
  assert.equal(importTotals({ ...data, new_payments_cents: undefined }, false).payments, 5000);
  assert.equal(hasChanges({ new_count: 0, rows: [{ status: "existing", matched_item_id: "item", planned_amount_cents: 0 }] }), true);
  assert.equal(subscriptionMatchCount({ rows: [{ status: "needs_subscription_match" }] }), 1);
  assert.equal(ignoredPayments({ rows: [{ status: "skipped", reason: "non_expense", title: "Pagamento recebido" }] }), 1);
});

test("summary origins choose debt/invoice/item without inventing competence", () => {
  const { summarySourceLink } = load("lib/summary-source.ts");
  assert.equal(summarySourceLink({ debt_id: "debt/id", item_id: "item", reference_month: null }), "/debts/debt%2Fid");
  assert.equal(summarySourceLink({ credit_card_id: "card", invoice_payment_month: "2026-11", item_id: "item", reference_month: null }), "/invoices/card/2026-11");
  assert.equal(summarySourceLink({ item_id: "item", reference_month: null }), "/entries/item");
  assert.equal(summarySourceLink({ effect: "savings" }), "/planning");
  assert.equal(summarySourceLink({ kind: "card_invoice_adjustment", reference_month: null }), null);
});
