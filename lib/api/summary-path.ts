// The API owns forecast filtering and monetary totals. Keep SSR and browser requests aligned.
export function variableForecastPreference(value: unknown): boolean {
  return value === "true";
}

export function monthlySummaryPath(month: string, basis: "cash" | "reference", includeVariableForecasts = false) {
  const query = new URLSearchParams({ basis });
  if (basis === "cash") query.set("include_variable_forecasts", String(includeVariableForecasts));
  return `/plans/current/months/${encodeURIComponent(month)}/summary?${query.toString()}`;
}
