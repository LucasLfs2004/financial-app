import { addMonths, clampMonth, inclusiveMonthCount, isMonth } from "./debt-month";

export function queryWindow(query: { from?: unknown; to?: unknown }, initial: string, bounds: { from: string; to: string }, limit: number, size = 12) {
  const fallback = initial > bounds.to ? addMonths(bounds.to, -Math.min(size - 1, inclusiveMonthCount(bounds.from, bounds.to) - 1)) : initial;
  const requestedFrom = isMonth(query.from) ? query.from : fallback;
  const from = clampMonth(requestedFrom, bounds.from, bounds.to);
  const defaultTo = addMonths(from, Math.min(size - 1, inclusiveMonthCount(from, bounds.to) - 1));
  const requestedTo = isMonth(query.to) ? query.to : defaultTo;
  const count = inclusiveMonthCount(from, requestedTo);
  const invalid = (query.from !== undefined && !isMonth(query.from)) || (query.to !== undefined && !isMonth(query.to)) || count < 1 || count > limit;
  const to = invalid ? defaultTo : clampMonth(requestedTo, from, bounds.to);
  return { from, to, adjusted: invalid || from !== requestedFrom || to !== requestedTo || (query.from === undefined && initial !== fallback) };
}
