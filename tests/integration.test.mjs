import { test } from "node:test";
import assert from "node:assert/strict";
import { load } from "./test-loader.mjs";

const months = load("lib/debt-month.ts");
test("calendar handles rollover, timezone and rejects invalid months", () => {
  assert.equal(months.addMonths("2026-12", 1), "2027-01");
  assert.equal(months.addMonths("2027-01", -1), "2026-12");
  const instant = new Date("2027-01-01T01:00:00Z");
  assert.equal(months.currentMonth("America/Sao_Paulo", instant), "2026-12");
  assert.equal(months.currentMonth("Asia/Tokyo", instant), "2027-01");
  for (const value of ["2026-00", "2026-13", "2026-1", "0000-01"]) {
    assert.equal(months.isMonth(value), false);
    assert.throws(() => months.addMonths(value, 1), RangeError);
  }
});

test("inclusive 24/120 boundaries and projection clipped at plan end", () => {
  assert.equal(months.monthRange("2026-10", "2028-09", 24).length, 24);
  assert.throws(() => months.monthRange("2026-10", "2028-10", 24));
  assert.equal(months.monthRange("2026-10", "2036-09", 120).length, 120);
  assert.throws(() => months.monthRange("2026-10", "2036-10", 120));
  const plan = { start_month: "2026-10", end_month: "2027-01" };
  assert.deepEqual(months.projectionMonths("2026-12", plan), ["2026-12", "2027-01"]);
  assert.deepEqual(months.projectionMonths("2027-12", plan), ["2027-01"]);
  assert.equal(months.selectPeriod([{ start_month: "2026-10", end_month: null, amount_cents: 0 }], "2026-11").amount_cents, 0);
  assert.equal(months.selectPeriod([{ start_month: "2026-12", end_month: null }], "2026-11"), undefined);
});

test("payment-history adapts API arrays, preserving metadata and empty direct default", () => {
  const { adaptPaymentHistory } = load("lib/api/payment-history.ts");
  assert.deepEqual(adaptPaymentHistory("item", []), { financial_item_id: "item", default_method: "direct", periods: [] });
  const period = { id: "p", start_month: "2026-10", end_month: null, method: "credit_card", credit_card_id: "c", context: "Mudança", recorded_at: "2026-10-04T12:00:00Z", created_at: "2026-10-04T12:00:00Z" };
  assert.deepEqual(adaptPaymentHistory("item", [period]).periods, [period]);
  assert.throws(() => adaptPaymentHistory("item", { periods: [] }));
});

test("180 installments open in bounded windows; settlement is independent of visible schedule", () => {
  const { queryWindow } = load("lib/query-window.ts");
  const debt = { scheduled_start_month: "2026-10", scheduled_end_month: "2041-09" };
  const bounds = { from: debt.scheduled_start_month, to: debt.scheduled_end_month };
  assert.deepEqual(queryWindow({}, "2026-10", bounds, 120), { from: "2026-10", to: "2027-09", adjusted: false });
  assert.deepEqual(queryWindow({ from: "2027-10" }, "2026-10", bounds, 120), { from: "2027-10", to: "2028-09", adjusted: false });
  assert.deepEqual(queryWindow({}, "2042-01", bounds, 120), { from: "2040-10", to: "2041-09", adjusted: true });
  const exact = queryWindow({ from: "2026-10", to: "2036-09" }, "2026-10", bounds, 120);
  assert.equal(months.inclusiveMonthCount(exact.from, exact.to), 120);
  assert.equal(queryWindow({ from: "2026-10", to: "2036-10" }, "2026-10", bounds, 120).to, "2027-09");
  assert.equal(queryWindow({ from: "2026-13", to: "2026-01" }, "2026-10", bounds, 120).adjusted, true);
  assert.deepEqual(months.settlementMonthBounds(debt), { from: "2026-10", to: "2041-08", available: true });
  assert.equal(months.settlementMonthBounds({ scheduled_start_month: "2026-10", scheduled_end_month: "2026-10" }).available, false);
});

test("invoice window supports history and overflow without exceeding operational horizon", () => {
  const { queryWindow } = load("lib/query-window.ts");
  const bounds = { from: "2026-01", to: "2028-12" };
  assert.deepEqual(queryWindow({ from: "2026-01", to: "2027-12" }, "2026-10", bounds, 24), { from: "2026-01", to: "2027-12", adjusted: false });
  assert.deepEqual(queryWindow({ from: "2028-10", to: "2029-09" }, "2026-10", bounds, 24), { from: "2028-10", to: "2028-12", adjusted: true });
  assert.equal(queryWindow({ from: "2026-01", to: "2028-01" }, "2026-10", bounds, 24).adjusted, true);
  assert.equal(queryWindow({}, "2029-01", bounds, 24).from, "2028-01");
});

test("proxy preserves status, code/message, query, encoded ID and multipart boundary", async () => {
  const originalFetch = global.fetch;
  const upstreamStatuses = [200, 201, 400, 401, 404, 409, 422, 500];
  const api = load("lib/api/financial-api.ts", { "server-only": {}, "@/lib/supabase/server": { getAccessToken: async () => "session-token" } });
  const route = load("app/api/financial/[...path]/route.ts", {
    "@/lib/api/financial-api": api,
    "next/server": { NextResponse: { json: (body, init) => Response.json(body, init) } },
  });
  try {
    for (const status of upstreamStatuses) {
      const payload = status < 400 ? { data: { id: "created" } } : { error: { code: "period_overlap", message: "Conflito de vigência", details: { month: "2026-10" } } };
      let forwarded;
      global.fetch = async (url, init) => { forwarded = { url, init }; return Response.json(payload, { status }); };
      const request = new Request("http://frontend/api/financial/credit-cards/card/imports?from=2026-10&to=2027-09", { method: "POST", headers: { "Content-Type": "multipart/form-data; boundary=fixture-boundary" }, body: "--fixture-boundary\r\nCSV\r\n" });
      request.nextUrl = new URL(request.url);
      const response = await route.POST(request, { params: Promise.resolve({ path: ["credit-cards", "card/id", "imports"] }) });
      assert.equal(response.status, status);
      assert.deepEqual(await response.json(), payload);
      assert.match(forwarded.url, /\/v1\/credit-cards\/card%2Fid\/imports\?from=2026-10&to=2027-09$/);
      assert.equal(forwarded.init.headers.get("Authorization"), "Bearer session-token");
      assert.equal(forwarded.init.headers.get("Content-Type"), "multipart/form-data; boundary=fixture-boundary");
      assert.equal(new TextDecoder().decode(forwarded.init.body), "--fixture-boundary\r\nCSV\r\n");
    }
    global.fetch = async () => Response.json({ error: { code: "not_found", message: "Ausente" } }, { status: 404 });
    await assert.rejects(api.financialApiFetch("/v1/plans/current"), (error) => error.status === 404 && error.code === "not_found" && error.message === "Ausente");
  } finally { global.fetch = originalFetch; }
});

test("browser preserves errors, Headers and multipart; failed mutations are submitted once", async () => {
  const originalFetch = global.fetch;
  const { browserApiFetch, BrowserApiError } = load("lib/api/browser-api.ts");
  try {
    let calls = 0;
    global.fetch = async () => { calls++; throw new TypeError("connection lost after sending"); };
    await assert.rejects(browserApiFetch("/debts/d/early-settlement", { method: "POST", body: "{}" }), (error) => error instanceof BrowserApiError && error.code === "network_error");
    assert.equal(calls, 1);
    global.fetch = async () => { throw new DOMException("Cancelled", "AbortError"); };
    await assert.rejects(browserApiFetch("/debts", { signal: new AbortController().signal }), (error) => error.name === "AbortError");
    const form = new FormData(); form.set("file", "csv");
    global.fetch = async (_url, init) => {
      assert.equal(init.body, form);
      assert.equal(init.headers.has("Content-Type"), false);
      assert.equal(init.headers.get("X-Test"), "kept");
      return Response.json({ error: { code: "validation_error", message: "Prévia obsoleta", details: { preview: "expired" } } }, { status: 422 });
    };
    await assert.rejects(browserApiFetch("/imports", { method: "POST", body: form, headers: new Headers({ "Content-Type": "multipart/form-data", "X-Test": "kept" }) }), (error) => error.status === 422 && error.code === "validation_error" && error.message === "Prévia obsoleta" && error.details.preview === "expired");
  } finally { global.fetch = originalFetch; }
});


test("expired sessions keep proxy semantics and redirect server reads to sign-in", async () => {
  const originalFetch = global.fetch;
  const redirect = (path) => { throw new Error(`redirect:${path}`); };
  const missingSession = load("lib/api/financial-api.ts", { "server-only": {}, "@/lib/supabase/server": { getAccessToken: async () => null }, "next/navigation": { redirect } });
  const expiredSession = load("lib/api/financial-api.ts", { "server-only": {}, "@/lib/supabase/server": { getAccessToken: async () => "expired" }, "next/navigation": { redirect } });
  try {
    await assert.rejects(missingSession.financialApiResponse("/v1/me"), (error) => error.status === 401 && error.code === "unauthorized");
    await assert.rejects(missingSession.getProfile(), /redirect:\/sign-in/);
    global.fetch = async () => Response.json({ error: { code: "unauthorized", message: "Expired" } }, { status: 401 });
    assert.equal((await expiredSession.financialApiResponse("/v1/me")).status, 401);
    await assert.rejects(expiredSession.getProfile(), /redirect:\/sign-in/);
  } finally { global.fetch = originalFetch; }
});
