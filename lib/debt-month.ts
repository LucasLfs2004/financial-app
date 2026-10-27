export function isMonth(value: unknown): value is string {
  return typeof value === "string" && /^[0-9]{4}-(0[1-9]|1[0-2])$/.test(value) && value.slice(0, 4) !== "0000";
}

export function currentMonth(timezone = "America/Sao_Paulo", now = new Date()) {
  const parts = new Intl.DateTimeFormat("en-US", { timeZone: timezone, year: "numeric", month: "2-digit" }).formatToParts(now);
  return `${parts.find((part) => part.type === "year")!.value}-${parts.find((part) => part.type === "month")!.value}`;
}

function monthIndex(month: string) {
  if (!isMonth(month)) throw new RangeError("Mês inválido.");
  const [year, part] = month.split("-").map(Number);
  return year * 12 + part - 1;
}

export function addMonths(month: string, count: number) {
  if (!Number.isInteger(count)) throw new RangeError("Deslocamento inválido.");
  const index = monthIndex(month) + count;
  const year = Math.floor(index / 12);
  if (year < 1 || year > 9999) throw new RangeError("Mês fora do calendário suportado.");
  return `${String(year).padStart(4, "0")}-${String(index % 12 + 1).padStart(2, "0")}`;
}

export function inclusiveMonthCount(from: string, to: string) {
  return monthIndex(to) - monthIndex(from) + 1;
}

export function monthRange(from: string, to: string, limit: number) {
  const count = inclusiveMonthCount(from, to);
  if (count < 1 || count > limit) throw new RangeError(`Escolha um intervalo de até ${limit} meses.`);
  return Array.from({ length: count }, (_, index) => addMonths(from, index));
}

export function clampMonth(month: string, from: string, to: string) {
  if (!isMonth(month) || inclusiveMonthCount(from, to) < 1) throw new RangeError("Intervalo inválido.");
  return month < from ? from : month > to ? to : month;
}

export function projectionMonths(from: string, plan: { start_month: string; end_month: string }) {
  const start = clampMonth(from, plan.start_month, plan.end_month);
  const end = addMonths(start, Math.min(4, inclusiveMonthCount(start, plan.end_month) - 1));
  return monthRange(start, end, 5);
}

export function selectPeriod<T extends { start_month: string; end_month: string | null }>(periods: T[], month: string): T | undefined {
  if (!isMonth(month)) return undefined;
  return periods.find((period) => period.start_month <= month && (period.end_month === null || period.end_month >= month));
}

export function settlementMonthBounds(debt: { scheduled_start_month: string; scheduled_end_month: string }) {
  const from = debt.scheduled_start_month;
  const to = addMonths(debt.scheduled_end_month, -1);
  return { from, to, available: from <= to };
}

export function formatMonth(month: string) {
  const [year, part] = month.split("-").map(Number);
  const date = new Date(0);
  date.setUTCFullYear(year, part - 1, 1);
  return new Intl.DateTimeFormat("pt-BR", { month: "short", year: "numeric", timeZone: "UTC" }).format(date);
}
