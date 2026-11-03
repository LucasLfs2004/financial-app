import { test } from "node:test";
import assert from "node:assert/strict";
import { load } from "./test-loader.mjs";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

const { commitmentGroups, invoiceComposition, invoiceComponentLabel } = load("lib/financial-overview.ts");

test("monthly groups preserve signed payments and exclude savings/income", () => {
  const sources = [
    { effect: "income", amount_cents: 500000 },
    { effect: "savings", amount_cents: 100000 },
    { effect: "commitment", amount_cents: 80000, payment_method: "credit_card" },
    { effect: "commitment", amount_cents: -30000, credit_card_id: "card" },
    { effect: "commitment", amount_cents: 20000, payment_method: "direct" },
    { effect: "commitment", amount_cents: 15000 },
  ];
  const original = JSON.stringify(sources);
  const groups = commitmentGroups(sources);
  assert.deepEqual(groups.map((group) => group.amount), [50000, 35000]);
  assert.equal(groups.flatMap((group) => group.sources).length, 4);
  assert.equal(JSON.stringify(sources), original);
  assert.deepEqual(commitmentGroups([]).map((group) => group.amount), [0, 0]);
});

test("planned occurrences are never presented as realized even when is_projected is false", () => {
  const item = { source_type: "financial_item_occurrence", is_projected: false, amount_cents: 60000 };
  const installment = { source_type: "invoice_adjustment", is_projected: true, amount_cents: 40000 };
  const charge = { source_type: "invoice_adjustment", is_projected: false, amount_cents: 45000 };
  const payment = { source_type: "invoice_payment", amount_cents: -25000 };
  const credit = { source_type: "invoice_adjustment", amount_cents: -5000 };
  const unknown = { source_type: "future_contract", amount_cents: 1000 };
  assert.equal(invoiceComponentLabel(item), "Compromisso previsto");
  assert.equal(invoiceComponentLabel(installment), "Parcela prevista");
  assert.equal(invoiceComponentLabel(charge), "Cobrança registrada");
  assert.equal(invoiceComponentLabel(payment), "Pagamento registrado");
  assert.equal(invoiceComponentLabel(credit), "Crédito registrado");
  assert.equal(invoiceComponentLabel(unknown), "Lançamento");
  const groups = invoiceComposition([item, installment, charge, payment, credit, unknown]);
  assert.deepEqual(groups.map((group) => [group.id, group.amount]), [["planned", 100000], ["charges", 45000], ["payments", -30000], ["other", 1000]]);
  assert.equal(groups.reduce((sum, group) => sum + group.amount, 0), 116000);
  assert.deepEqual(invoiceComposition([]), []);
});

test("optional reads preserve authentication redirects and unexpected failures", () => {
  class FinancialApiError extends Error { constructor(status) { super("API"); this.status = status; } }
  const { unavailableRead } = load("lib/api/optional-read.ts", { "./financial-api": { FinancialApiError } });
  assert.equal(unavailableRead(new FinancialApiError(404)), null);
  assert.equal(unavailableRead(new FinancialApiError(500)), null);
  assert.equal(unavailableRead(new DOMException("Timeout", "TimeoutError")), null);
  assert.throws(() => unavailableRead(new FinancialApiError(401)));
  assert.throws(() => unavailableRead(new Error("NEXT_REDIRECT")), /NEXT_REDIRECT/);
  assert.throws(() => unavailableRead(new Error("Unexpected")), /Unexpected/);
});

test("cards distinguish unavailable invoices from real zero and constrain the requested month", async () => {
  class FinancialApiError extends Error { constructor(status) { super("API"); this.status = status; } }
  let missing = false;
  const requested = [];
  const api = {
    FinancialApiError,
    getInstitutions: async () => [],
    getCreditCards: async () => [{ id: "a", status: "active" }, { id: "b", status: "archived" }],
    getProfile: async () => ({ currency_code: "BRL", timezone: "America/Sao_Paulo" }),
    getCurrentPlan: async () => ({ start_month: "2026-11", end_month: "2027-12" }),
    absentOn404: () => null,
    getCardInvoice: async (id, month) => {
      requested.push(month);
      if (missing && id === "b") throw new FinancialApiError(404);
      return { currency_code: "BRL", projected_total_cents: 0 };
    },
  };
  const Page = load("app/(protected)/cards/page.tsx", {
    "@/lib/api/financial-api": api, "./financial-api": api,
    "@/components/cards/institution-list": { InstitutionList: () => null },
    "@/components/cards/cards-hub": { CardsHub: () => null },
    "@/components/month-navigation": { MonthNavigation: () => null },
    "next/link": { default: ({ children, ...props }) => createElement("a", props, children) },
  }).default;
  const render = async () => renderToStaticMarkup(await Page({ searchParams: Promise.resolve({ month: "2025-01" }) }));
  const zero = await render();
  assert.match(zero, /<strong>R\$\s*0,00<\/strong>/);
  assert.deepEqual(requested, ["2026-11", "2026-11"]);
  missing = true;
  const partial = await render();
  assert.match(partial, /<strong>—<\/strong>/);
  assert.match(partial, /indisponíveis não são tratados como zero/);
});

test("free-balance chart labels negative, zero and unavailable months separately", () => {
  const { FreeBalanceChart } = load("components/dashboard/free-balance-chart.tsx", {
    "next/link": { default: ({ children, ...props }) => createElement("a", props, children) },
  });
  const html = renderToStaticMarkup(createElement(FreeBalanceChart, {
    months: ["2026-11", "2026-12", "2027-01"],
    summaries: [{ result_cents: -50000 }, { result_cents: 0 }, null], currencyCode: "BRL",
  }));
  assert.match(html, /has-negative/);
  assert.match(html, /is-negative/);
  assert.match(html, /-R\$\s*500,00/);
  assert.match(html, /Sem sobra/);
  assert.match(html, /Indisponível/);
  assert.match(html, /jan.*27/);
});
