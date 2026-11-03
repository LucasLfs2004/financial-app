import { test } from "node:test";
import assert from "node:assert/strict";
import { load } from "./test-loader.mjs";

test("cash requests exclude variable forecasts while reference preserves the projection", () => {
  const { monthlySummaryPath } = load("lib/api/summary-path.ts");
  const cash = new URL(monthlySummaryPath("2026-10", "cash"), "http://localhost");
  assert.equal(cash.pathname, "/plans/current/months/2026-10/summary");
  assert.equal(cash.searchParams.get("include_variable_forecasts"), "false");
  assert.equal(cash.searchParams.get("basis"), "cash");
  const reference = new URL(monthlySummaryPath("2026-10", "reference"), "http://localhost");
  assert.equal(reference.searchParams.get("basis"), "reference");
  assert.equal(reference.searchParams.has("include_variable_forecasts"), false);
});

test("SSR monthly summary uses the same cash policy and preserves official API amounts", async () => {
  const { getMonthlySummary } = load("lib/api/financial-api.ts", {
    "server-only": {},
    "@/lib/supabase/server": { getAccessToken: async () => "test-token" },
    "next/navigation": { redirect: () => { throw new Error("unexpected redirect"); } },
  });
  const original = global.fetch;
  const paths = [];
  try {
    global.fetch = async (url) => {
      paths.push(new URL(url));
      return Response.json({ data: { commitments_cents: 100000, result_cents: 200000, excluded_variable_forecasts_cents: 50000, excluded_variable_forecasts: [{ name: "Posto", amount_cents: 50000 }] } });
    };
    const cash = await getMonthlySummary("2026-10");
    assert.equal(cash.commitments_cents, 100000);
    assert.equal(cash.result_cents, 200000);
    assert.equal(cash.excluded_variable_forecasts_cents, 50000);
    assert.equal(paths[0].searchParams.get("include_variable_forecasts"), "false");
    await getMonthlySummary("2026-10", "reference");
    assert.equal(paths[1].searchParams.has("include_variable_forecasts"), false);
  } finally { global.fetch = original; }
});


test("explicit cash modes survive URL parsing and do not affect reference queries", () => {
  const { monthlySummaryPath, variableForecastPreference } = load("lib/api/summary-path.ts");
  for (const include of [false, true]) {
    const url = new URL(monthlySummaryPath("2026-11", "cash", include), "http://localhost");
    assert.equal(variableForecastPreference(url.searchParams.get("include_variable_forecasts")), include);
    const nextMonth = new URL(monthlySummaryPath("2026-12", "cash", variableForecastPreference(url.searchParams.get("include_variable_forecasts"))), "http://localhost");
    assert.equal(nextMonth.searchParams.get("include_variable_forecasts"), String(include));
    assert.equal(new URL(monthlySummaryPath("2026-11", "reference", include), "http://localhost").searchParams.has("include_variable_forecasts"), false);
  }
  for (const value of [undefined, null, "invalid", ["true"], "false"]) assert.equal(variableForecastPreference(value), false);
});

test("SSR cash modes request their own official totals including imported charges", async () => {
  const { getMonthlySummary } = load("lib/api/financial-api.ts", {
    "server-only": {},
    "@/lib/supabase/server": { getAccessToken: async () => "test-token" },
    "next/navigation": { redirect: () => { throw new Error("unexpected redirect"); } },
  });
  const original = global.fetch;
  try {
    global.fetch = async (input) => {
      const include = new URL(input).searchParams.get("include_variable_forecasts") === "true";
      return Response.json({ data: { commitments_cents: include ? 100000 : 50000, excluded_variable_forecasts_cents: include ? 0 : 50000 } });
    };
    assert.equal((await getMonthlySummary("2026-10", "cash", true)).commitments_cents, 100000);
    const without = await getMonthlySummary("2026-10", "cash", false);
    assert.equal(without.commitments_cents, 50000);
    assert.equal(without.excluded_variable_forecasts_cents, 50000);
  } finally { global.fetch = original; }
});
