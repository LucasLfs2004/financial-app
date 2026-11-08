import { test } from "node:test";
import assert from "node:assert/strict";
import { load } from "./test-loader.mjs";

test("overview totals include existing lines and separate refunds from payments", () => {
  const { csvTotals, decisionRows, duplicateCount, unsupportedCredits } = load("lib/invoice-import.ts");
  const data = { rows: [
    { kind: "expense", amount_cents: 10000, status: "existing" },
    { kind: "credit", amount_cents: -4514, status: "new" },
    { kind: "payment", amount_cents: -202576, status: "needs_invoice" },
    { kind: "expense", amount_cents: 4159, status: "needs_duplicate_decision", duplicate_candidates: [{ id: "old" }] },
  ] };
  assert.deepEqual(csvTotals(data), { charges: 14159, credits: 4514, payments: 202576 });
  assert.equal(decisionRows(data).length, 2);
  assert.equal(duplicateCount(data), 1);
  assert.equal(unsupportedCredits(data), 0);
  assert.equal(unsupportedCredits({ rows: [{ kind: "unsupported", amount_cents: -4514, status: "skipped" }] }), 1);
  data.rows[3].status = "skipped";
  assert.equal(decisionRows(data).length, 2, "resolved decisions must remain editable");
});
